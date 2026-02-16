package runtimedeploy

import "context"

// ApplyNow applies runtime desired state immediately without persisting queue task.
// This path is used by bootstrap/emergency CLI flows before control-plane runtime loop is up.
func (s *Service) ApplyNow(ctx context.Context, params PrepareParams) (PrepareResult, error) {
	params = normalizePrepareParams(params)
	return s.applyDesiredState(ctx, params)
}
