package worker

import (
	"context"
	"strings"
)

const (
	jobImageSourceLauncherDefault      = "launcher_default"
	jobImageSourcePrimaryUnchecked     = "primary_unchecked"
	jobImageSourcePrimaryCheckError    = "primary_check_error"
	jobImageSourcePrimary              = "primary"
	jobImageSourceFallback             = "fallback"
	jobImageSourceFallbackAutoPrevious = "fallback_auto_previous"
	jobImageSourceFallbackCheckError   = "fallback_check_error"
	jobImageSourcePrimaryMissing       = "primary_missing_no_fallback"
	jobImageSourceFallbackMissing      = "fallback_missing_use_primary"
)

type resolvedJobImage struct {
	SelectedImage     string
	SelectedSource    string
	PrimaryImage      string
	FallbackImage     string
	PrimaryAvailable  bool
	FallbackAvailable bool
	CheckError        string
	EmitEvent         bool
}

func (s *Service) resolveRunJobImage(ctx context.Context) resolvedJobImage {
	primary := strings.TrimSpace(s.image.Primary)
	fallback := strings.TrimSpace(s.image.Fallback)
	checker := s.image.Checker

	if primary == "" {
		return resolvedJobImage{
			SelectedSource: jobImageSourceLauncherDefault,
			PrimaryImage:   primary,
			FallbackImage:  fallback,
		}
	}
	if checker == nil {
		return resolvedJobImage{
			SelectedImage:  primary,
			SelectedSource: jobImageSourcePrimaryUnchecked,
			PrimaryImage:   primary,
			FallbackImage:  fallback,
		}
	}

	result := resolvedJobImage{
		SelectedImage: primary,
		PrimaryImage:  primary,
		FallbackImage: fallback,
		EmitEvent:     true,
	}

	primaryAvailable, primaryErr := checker.IsImageAvailable(ctx, primary)
	if primaryErr != nil {
		result.SelectedSource = jobImageSourcePrimaryCheckError
		result.CheckError = primaryErr.Error()
		return result
	}
	result.PrimaryAvailable = primaryAvailable
	if primaryAvailable {
		result.SelectedSource = jobImageSourcePrimary
		return result
	}

	if fallback == "" {
		autoFallback, ok, autoErr := checker.ResolvePreviousImage(ctx, primary)
		if autoErr != nil {
			result.SelectedSource = jobImageSourceFallbackCheckError
			result.CheckError = autoErr.Error()
			return result
		}
		if ok {
			result.FallbackImage = autoFallback
			result.FallbackAvailable = true
			result.SelectedImage = autoFallback
			result.SelectedSource = jobImageSourceFallbackAutoPrevious
			return result
		}
		result.SelectedSource = jobImageSourcePrimaryMissing
		return result
	}

	fallbackAvailable, fallbackErr := checker.IsImageAvailable(ctx, fallback)
	if fallbackErr != nil {
		result.SelectedSource = jobImageSourceFallbackCheckError
		result.CheckError = fallbackErr.Error()
		return result
	}
	result.FallbackAvailable = fallbackAvailable
	if fallbackAvailable {
		result.SelectedImage = fallback
		result.SelectedSource = jobImageSourceFallback
		return result
	}

	result.SelectedSource = jobImageSourceFallbackMissing
	return result
}
