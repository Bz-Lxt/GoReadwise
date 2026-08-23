package clip_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"goreadwise/internal/clip"
)

func TestRealProviderPreservesBodyReadError(t *testing.T) {
	readErr := errors.New("response body interrupted")
	body := &failingBody{readErr: readErr}
	provider := clip.RealProvider{
		Client: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       body,
				Request:    req,
			}, nil
		})},
		MaxBytes: 1024,
	}

	_, err := provider.Clip(context.Background(), "https://8.8.8.8/article")
	if !errors.Is(err, readErr) {
		t.Fatalf("expected response body error, got %v", err)
	}
	if !body.closed {
		t.Fatal("response body was not closed")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type failingBody struct {
	readErr error
	closed  bool
}

func (b *failingBody) Read([]byte) (int, error) {
	return 0, b.readErr
}

func (b *failingBody) Close() error {
	b.closed = true
	return nil
}
