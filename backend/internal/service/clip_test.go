package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"goreadwise/internal/clip"
	"goreadwise/internal/httpx"
)

// failingProvider lets configured tests exercise the provider error path
// without needing a database.
type failingProvider struct{}

func (failingProvider) Clip(context.Context, string) (clip.Result, error) {
	return clip.Result{}, httpx.ErrValidation
}

// TestClipWithoutProviderReturnsValidationNotPanic is the regression test for
// the nil-provider crash. With no provider configured, Clip must hand the caller
// a handleable ErrValidation ("provider not configured") instead of installing
// a nil *MockProvider into the interface and panicking on the value-receiver
// dispatch.
func TestClipWithoutProviderReturnsValidationNotPanic(t *testing.T) {
	svc := &ClipService{Cards: nil, Provider: nil}

	var (
		gotPanic bool
		err      error
	)
	func() {
		defer func() {
			if r := recover(); r != nil {
				gotPanic = true
			}
		}()
		_, err = svc.Clip(context.Background(), "https://example.com")
	}()

	if gotPanic {
		t.Fatalf("Clip panicked instead of returning a handleable error")
	}
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errors.Is(err, httpx.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "provider not configured") {
		t.Fatalf("expected 'provider not configured' in message, got %q", err.Error())
	}
	if code := httpx.ErrorCodeOf(err); code != "VALIDATION" {
		t.Fatalf("expected VALIDATION code, got %s", code)
	}
}

// TestClipWithConfiguredProviderInvokesProvider ensures the nil-guard does
// not short-circuit configured environments (mock or real).
func TestClipWithConfiguredProviderInvokesProvider(t *testing.T) {
	svc := &ClipService{Cards: nil, Provider: failingProvider{}}

	_, err := svc.Clip(context.Background(), "https://example.com")
	if err == nil {
		t.Fatal("expected provider error to surface")
	}
	if !errors.Is(err, httpx.ErrValidation) {
		t.Fatalf("expected ErrValidation from provider, got %v", err)
	}
	if strings.Contains(err.Error(), "provider not configured") {
		t.Fatalf("provider was configured but got not-configured error: %v", err)
	}
}
