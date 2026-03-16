package systemsettings

import "strings"

const (
	VisibilityStaffVisible      = "staff_visible"
	VisibilityInternalOnly      = "internal_only"
	VisibilitySecretForbiddenWS = "secret_forbidden_in_ws"
)

type ExposureSurface string

const (
	ExposureSurfaceStaff    ExposureSurface = "staff"
	ExposureSurfaceRealtime ExposureSurface = "realtime"
)

// IsVisibleOnSurface applies shared visibility semantics for staff/read-model consumers.
func IsVisibleOnSurface(visibility string, surface ExposureSurface) bool {
	switch strings.TrimSpace(visibility) {
	case VisibilityStaffVisible:
		return true
	case VisibilityInternalOnly:
		return false
	case VisibilitySecretForbiddenWS:
		return surface != ExposureSurfaceRealtime
	default:
		return false
	}
}
