package http

import "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/controlplane"

// staffHandler implements staff/private JSON endpoints protected by JWT.
type staffHandler struct {
	cp *controlplane.Client
}

func newStaffHandler(cp *controlplane.Client) *staffHandler {
	return &staffHandler{cp: cp}
}
