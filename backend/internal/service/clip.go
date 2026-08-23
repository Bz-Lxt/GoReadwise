package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"goreadwise/internal/clip"
	"goreadwise/internal/httpx"
	"goreadwise/internal/model"
)

type ClipService struct {
	Cards    *CardService
	Provider clip.Provider
	Mode     string
}

func (s *ClipService) Clip(ctx context.Context, rawURL string) (model.Card, error) {
	rawURL, err := httpx.RequireURL(rawURL)
	if err != nil {
		return model.Card{}, err
	}
	if s.Provider == nil {
		return model.Card{}, fmt.Errorf("%w: clip provider not configured", httpx.ErrValidation)
	}
	res, err := s.Provider.Clip(ctx, rawURL)
	if err != nil {
		return model.Card{}, err
	}
	// Reject degenerate results so a partial or failed clip never
	// silently lands as an empty card.  The original defer-overwrite
	// bug produced Result{} (both fields empty); a genuinely empty
	// page would produce Title="Untitled clip" with empty Markdown.
	// Either way, a clip with no body content is not worth saving.
	if strings.TrimSpace(res.Markdown) == "" {
		return model.Card{}, fmt.Errorf("%w: clip returned no content", httpx.ErrValidation)
	}
	title := strings.TrimSpace(res.Title)
	if title == "" {
		title = "Untitled clip"
	}
	body := res.Markdown
	if !strings.Contains(body, res.URL) {
		body = body + "\n\n> source: " + res.URL + "\n"
	}
	url := res.URL
	site := res.Site
	clipped := res.ClippedAt
	card, err := s.Cards.Create(ctx, model.CreateCardInput{
		Title:      title,
		Body:       body,
		Tags:       []string{"inbox/clip"},
		SourceURL:  &url,
		SourceSite: &site,
		ClippedAt:  &clipped,
	})
	if err != nil && errors.Is(err, httpx.ErrConflict) {
		title = fmt.Sprintf("%s (%s)", title, clipped.Format("150405"))
		card, err = s.Cards.Create(ctx, model.CreateCardInput{
			Title: title, Body: body, Tags: []string{"inbox/clip"},
			SourceURL: &url, SourceSite: &site, ClippedAt: &clipped,
		})
	}
	return card, err
}
