package service

import (
	"context"
	"errors"
	"testing"

	"goreadwise/internal/httpx"
)

func TestClipWithoutProviderReturnsValidationError(t *testing.T) {
	svc := &ClipService{}

	_, err := svc.Clip(context.Background(), "https://example.com/article")
	if !errors.Is(err, httpx.ErrValidation) {
		t.Fatalf("Clip() error = %v, want validation error", err)
	}
}
