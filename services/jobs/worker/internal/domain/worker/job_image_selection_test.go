package worker

import (
	"context"
	"errors"
	"testing"
)

type fakeJobImageChecker struct {
	availability map[string]bool
	errs         map[string]error
	previousBy   map[string]string
}

func (f *fakeJobImageChecker) IsImageAvailable(_ context.Context, imageRef string) (bool, error) {
	if err, ok := f.errs[imageRef]; ok {
		return false, err
	}
	return f.availability[imageRef], nil
}

func (f *fakeJobImageChecker) ResolvePreviousImage(_ context.Context, imageRef string) (string, bool, error) {
	if err, ok := f.errs["previous:"+imageRef]; ok {
		return "", false, err
	}
	image, ok := f.previousBy[imageRef]
	if !ok {
		return "", false, nil
	}
	return image, true, nil
}

func TestResolveRunJobImage_PrimaryUncheckedWithoutChecker(t *testing.T) {
	t.Parallel()

	svc := &Service{
		image: JobImageSelectionPolicy{
			Primary: "127.0.0.1:5000/codex-k8s/agent-runner:ai-1.0.0",
		},
	}

	got := svc.resolveRunJobImage(context.Background())
	if got.SelectedImage != "127.0.0.1:5000/codex-k8s/agent-runner:ai-1.0.0" {
		t.Fatalf("unexpected selected image: %q", got.SelectedImage)
	}
	if got.SelectedSource != jobImageSourcePrimaryUnchecked {
		t.Fatalf("unexpected selected source: %q", got.SelectedSource)
	}
	if got.EmitEvent {
		t.Fatalf("expected EmitEvent=false when checker is disabled")
	}
}

func TestResolveRunJobImage_FallbackWhenPrimaryMissing(t *testing.T) {
	t.Parallel()

	primary := "127.0.0.1:5000/codex-k8s/agent-runner:ai-1.0.1"
	fallback := "127.0.0.1:5000/codex-k8s/agent-runner:ai-1.0.0"
	svc := &Service{
		image: JobImageSelectionPolicy{
			Primary:  primary,
			Fallback: fallback,
			Checker: &fakeJobImageChecker{
				availability: map[string]bool{
					primary:  false,
					fallback: true,
				},
			},
		},
	}

	got := svc.resolveRunJobImage(context.Background())
	if got.SelectedImage != fallback {
		t.Fatalf("unexpected selected image: %q", got.SelectedImage)
	}
	if got.SelectedSource != jobImageSourceFallback {
		t.Fatalf("unexpected selected source: %q", got.SelectedSource)
	}
	if !got.EmitEvent {
		t.Fatalf("expected EmitEvent=true when checker is enabled")
	}
}

func TestResolveRunJobImage_PrimaryCheckError(t *testing.T) {
	t.Parallel()

	primary := "127.0.0.1:5000/codex-k8s/agent-runner:ai-1.0.2"
	svc := &Service{
		image: JobImageSelectionPolicy{
			Primary: primary,
			Checker: &fakeJobImageChecker{
				errs: map[string]error{
					primary: errors.New("registry timeout"),
				},
			},
		},
	}

	got := svc.resolveRunJobImage(context.Background())
	if got.SelectedImage != primary {
		t.Fatalf("unexpected selected image: %q", got.SelectedImage)
	}
	if got.SelectedSource != jobImageSourcePrimaryCheckError {
		t.Fatalf("unexpected selected source: %q", got.SelectedSource)
	}
	if got.CheckError == "" {
		t.Fatalf("expected check error to be set")
	}
}

func TestResolveRunJobImage_AutoFallbackFromRegistryHistory(t *testing.T) {
	t.Parallel()

	primary := "127.0.0.1:5000/codex-k8s/agent-runner:ai-1.0.3"
	autoFallback := "127.0.0.1:5000/codex-k8s/agent-runner:ai-1.0.2"
	svc := &Service{
		image: JobImageSelectionPolicy{
			Primary: primary,
			Checker: &fakeJobImageChecker{
				availability: map[string]bool{
					primary: false,
				},
				previousBy: map[string]string{
					primary: autoFallback,
				},
			},
		},
	}

	got := svc.resolveRunJobImage(context.Background())
	if got.SelectedImage != autoFallback {
		t.Fatalf("unexpected selected image: %q", got.SelectedImage)
	}
	if got.SelectedSource != jobImageSourceFallbackAutoPrevious {
		t.Fatalf("unexpected selected source: %q", got.SelectedSource)
	}
	if got.FallbackImage != autoFallback {
		t.Fatalf("unexpected fallback image: %q", got.FallbackImage)
	}
	if !got.FallbackAvailable {
		t.Fatalf("expected fallback_available=true")
	}
}
