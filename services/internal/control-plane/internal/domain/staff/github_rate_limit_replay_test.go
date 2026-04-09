package staff

import (
	"context"
	"testing"

	"github.com/codex-k8s/kodex/libs/go/crypto/tokencrypt"
	nextstepdomain "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/nextstep"
	projecttokenrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/projecttoken"
	repocfgrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/repocfg"
	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
	querytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/query"
	valuetypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/value"
)

const replayTestTokenCryptKey = "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff"

type stubReplayRepoConfig struct {
	findResult        querytypes.RepositoryBindingFindResult
	tokenEncrypted    []byte
	botTokenEncrypted []byte
}

func (s *stubReplayRepoConfig) ListForProject(context.Context, string, int) ([]repocfgrepo.RepositoryBinding, error) {
	return nil, nil
}

func (s *stubReplayRepoConfig) GetByID(context.Context, string) (repocfgrepo.RepositoryBinding, bool, error) {
	return repocfgrepo.RepositoryBinding{}, false, nil
}

func (s *stubReplayRepoConfig) Upsert(context.Context, repocfgrepo.UpsertParams) (repocfgrepo.RepositoryBinding, error) {
	return repocfgrepo.RepositoryBinding{}, nil
}

func (s *stubReplayRepoConfig) Delete(context.Context, string, string) error {
	return nil
}

func (s *stubReplayRepoConfig) FindByProviderExternalID(context.Context, string, int64) (repocfgrepo.FindResult, bool, error) {
	return repocfgrepo.FindResult{}, false, nil
}

func (s *stubReplayRepoConfig) FindByProviderOwnerName(context.Context, string, string, string) (repocfgrepo.FindResult, bool, error) {
	return repocfgrepo.FindResult(s.findResult), true, nil
}

func (s *stubReplayRepoConfig) GetTokenEncrypted(context.Context, string) ([]byte, bool, error) {
	return append([]byte(nil), s.tokenEncrypted...), len(s.tokenEncrypted) > 0, nil
}

func (s *stubReplayRepoConfig) GetBotTokenEncrypted(context.Context, string) ([]byte, bool, error) {
	return append([]byte(nil), s.botTokenEncrypted...), len(s.botTokenEncrypted) > 0, nil
}

func (s *stubReplayRepoConfig) UpsertBotParams(context.Context, querytypes.RepositoryBotParamsUpsertParams) error {
	return nil
}

func (s *stubReplayRepoConfig) UpsertPreflightReport(context.Context, querytypes.RepositoryPreflightReportUpsertParams) error {
	return nil
}

func (s *stubReplayRepoConfig) AcquirePreflightLock(context.Context, querytypes.RepositoryPreflightLockAcquireParams) (string, bool, error) {
	return "", false, nil
}

func (s *stubReplayRepoConfig) ReleasePreflightLock(context.Context, string, string) error {
	return nil
}

func (s *stubReplayRepoConfig) SetTokenEncryptedForAll(context.Context, []byte) (int64, error) {
	return 0, nil
}

var _ repocfgrepo.Repository = (*stubReplayRepoConfig)(nil)

type noopProjectTokens struct{}

func (noopProjectTokens) GetByProjectID(context.Context, string) (projecttokenrepo.ProjectGitHubTokens, bool, error) {
	return projecttokenrepo.ProjectGitHubTokens{}, false, nil
}

func (noopProjectTokens) GetEncryptedByProjectID(context.Context, string) ([]byte, []byte, string, string, bool, error) {
	return nil, nil, "", "", false, nil
}

func (noopProjectTokens) Upsert(context.Context, projecttokenrepo.UpsertParams) error {
	return nil
}

func (noopProjectTokens) DeleteByProjectID(context.Context, string) error {
	return nil
}

var _ projecttokenrepo.Repository = noopProjectTokens{}

func TestReplayGitHubRateLimitPlatformCallRequiresExpectedRunLabels(t *testing.T) {
	t.Parallel()

	service := newReplayTestService(t, []string{"run:dev:revise"})

	err := service.ReplayGitHubRateLimitPlatformCall(context.Background(), valuetypes.GitHubRateLimitPlatformCallReplayPayload{
		OperationKind:      enumtypes.GitHubRateLimitPlatformReplayOperationKindIssueStageTransition,
		RepositoryFullName: "codex-k8s/kodex",
		IssueNumber:        427,
		TargetLabel:        "run:qa",
	})
	if err == nil {
		t.Fatal("expected validation error when expected_current_run_labels are missing")
	}
}

func TestReplayGitHubRateLimitPlatformCallAppliesOnlyMatchingCASSnapshot(t *testing.T) {
	t.Parallel()

	service := newReplayTestService(t, []string{"run:dev:revise", "state:in-review"})

	err := service.ReplayGitHubRateLimitPlatformCall(context.Background(), valuetypes.GitHubRateLimitPlatformCallReplayPayload{
		OperationKind:            enumtypes.GitHubRateLimitPlatformReplayOperationKindIssueStageTransition,
		RepositoryFullName:       "codex-k8s/kodex",
		IssueNumber:              427,
		TargetLabel:              "run:qa",
		RequestFingerprint:       "issue:427:run:dev:revise->run:qa",
		CorrelationID:            "corr-427",
		ExpectedCurrentRunLabels: []string{"run:dev:revise"},
	})
	if err != nil {
		t.Fatalf("ReplayGitHubRateLimitPlatformCall() error = %v", err)
	}

	mgmt := service.githubMgmt.(*stubNextStepGitHubMgmt)
	if got, want := mgmt.removedLabels, []string{"run:dev:revise"}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("removed labels = %#v, want %#v", got, want)
	}
	if got := len(mgmt.addedLabels); got != 1 {
		t.Fatalf("added label operations = %d, want 1", got)
	}
	if got := mgmt.addedLabels[0][0]; got != "run:qa" {
		t.Fatalf("added label = %q, want %q", got, "run:qa")
	}
}

func TestReplayGitHubRateLimitPlatformCallRejectsDriftedRunLabels(t *testing.T) {
	t.Parallel()

	service := newReplayTestService(t, []string{"run:release", "state:in-review"})

	err := service.ReplayGitHubRateLimitPlatformCall(context.Background(), valuetypes.GitHubRateLimitPlatformCallReplayPayload{
		OperationKind:            enumtypes.GitHubRateLimitPlatformReplayOperationKindIssueStageTransition,
		RepositoryFullName:       "codex-k8s/kodex",
		IssueNumber:              427,
		TargetLabel:              "run:qa",
		RequestFingerprint:       "issue:427:run:dev:revise->run:qa",
		CorrelationID:            "corr-427",
		ExpectedCurrentRunLabels: []string{"run:dev:revise"},
	})
	if err == nil {
		t.Fatal("expected conflict when run labels drifted")
	}

	mgmt := service.githubMgmt.(*stubNextStepGitHubMgmt)
	if len(mgmt.removedLabels) != 0 {
		t.Fatalf("unexpected removed labels = %#v", mgmt.removedLabels)
	}
	if len(mgmt.addedLabels) != 0 {
		t.Fatalf("unexpected added labels = %#v", mgmt.addedLabels)
	}
}

func newReplayTestService(t *testing.T, labels []string) *Service {
	t.Helper()

	crypt, err := tokencrypt.NewService(replayTestTokenCryptKey)
	if err != nil {
		t.Fatalf("tokencrypt.NewService() error = %v", err)
	}
	platformTokenEncrypted, err := crypt.EncryptString("platform-token")
	if err != nil {
		t.Fatalf("EncryptString(platform-token) error = %v", err)
	}
	botTokenEncrypted, err := crypt.EncryptString("bot-token")
	if err != nil {
		t.Fatalf("EncryptString(bot-token) error = %v", err)
	}

	return &Service{
		cfg: Config{NextStepLabels: nextstepdomain.DefaultLabels()},
		repos: &stubReplayRepoConfig{
			findResult: querytypes.RepositoryBindingFindResult{
				ProjectID:    "project-1",
				RepositoryID: "repo-1",
			},
			tokenEncrypted:    platformTokenEncrypted,
			botTokenEncrypted: botTokenEncrypted,
		},
		projectTokens: noopProjectTokens{},
		tokencrypt:    crypt,
		githubMgmt: &stubNextStepGitHubMgmt{
			labels: append([]string(nil), labels...),
		},
	}
}
