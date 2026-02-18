package staff

import (
	"context"
	"strings"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

// ResolveStaffByEmail resolves one allowed staff user by email and optionally refreshes GitHub login.
func (s *Service) ResolveStaffByEmail(ctx context.Context, params querytypes.StaffResolveByEmailParams) (entitytypes.User, error) {
	email := strings.TrimSpace(params.Email)
	if email == "" {
		return entitytypes.User{}, errs.Validation{Field: "email", Msg: "is required"}
	}

	user, ok, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return entitytypes.User{}, err
	}
	if !ok {
		return entitytypes.User{}, errs.Forbidden{Msg: "email is not allowed"}
	}

	login := strings.TrimSpace(params.GitHubLogin)
	if login != "" && !strings.EqualFold(user.GitHubLogin, login) {
		if err := s.users.UpdateGitHubIdentity(ctx, user.ID, user.GitHubUserID, login); err != nil {
			return entitytypes.User{}, err
		}
		user.GitHubLogin = login
	}

	return user, nil
}

// AuthorizeOAuthUser authorizes OAuth identity for an allowed user.
func (s *Service) AuthorizeOAuthUser(ctx context.Context, params querytypes.StaffAuthorizeOAuthUserParams) (entitytypes.User, error) {
	email := strings.TrimSpace(params.Email)
	if email == "" {
		return entitytypes.User{}, errs.Validation{Field: "email", Msg: "is required"}
	}
	login := strings.TrimSpace(params.GitHubLogin)
	if login == "" {
		return entitytypes.User{}, errs.Validation{Field: "github_login", Msg: "is required"}
	}
	if params.GitHubUserID <= 0 {
		return entitytypes.User{}, errs.Validation{Field: "github_user_id", Msg: "is required"}
	}

	user, ok, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return entitytypes.User{}, err
	}
	if !ok {
		return entitytypes.User{}, errs.Forbidden{Msg: "email is not allowed"}
	}

	if err := s.users.UpdateGitHubIdentity(ctx, user.ID, params.GitHubUserID, login); err != nil {
		return entitytypes.User{}, err
	}
	user.GitHubUserID = params.GitHubUserID
	user.GitHubLogin = login

	return user, nil
}
