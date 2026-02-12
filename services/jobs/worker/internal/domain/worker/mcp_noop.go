package worker

import "context"

type noopMCPTokenIssuer struct{}

func (noopMCPTokenIssuer) IssueRunMCPToken(_ context.Context, _ IssueMCPTokenParams) (IssuedMCPToken, error) {
	return IssuedMCPToken{}, nil
}
