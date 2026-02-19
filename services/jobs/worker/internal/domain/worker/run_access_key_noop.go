package worker

import "context"

type noopRunAccessKeyIssuer struct{}

func (noopRunAccessKeyIssuer) IssueRunAccessKey(_ context.Context, _ IssueRunAccessKeyParams) (IssuedRunAccessKey, error) {
	return IssuedRunAccessKey{}, nil
}
