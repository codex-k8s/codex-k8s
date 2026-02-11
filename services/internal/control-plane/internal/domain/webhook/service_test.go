package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	agentrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentrun"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/flowevent"
	projectmemberrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/projectmember"
	repocfgrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/repocfg"
	userrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/user"
)

func TestIngestGitHubWebhook_Dedup(t *testing.T) {
	ctx := context.Background()
	runs := &inMemoryRunRepo{items: map[string]string{}}
	events := &inMemoryEventRepo{}
	svc := NewService(runs, events, nil, nil, nil, nil, false, TriggerLabels{})

	payload := json.RawMessage(`{"action":"opened","repository":{"id":1,"full_name":"codex-k8s/codex-k8s"},"sender":{"id":10,"login":"ai-da-stas"}}`)
	cmd := IngestCommand{
		CorrelationID: "delivery-1",
		DeliveryID:    "delivery-1",
		EventType:     "pull_request",
		ReceivedAt:    time.Now().UTC(),
		Payload:       payload,
	}

	first, err := svc.IngestGitHubWebhook(ctx, cmd)
	if err != nil {
		t.Fatalf("first ingest failed: %v", err)
	}
	if first.Duplicate {
		t.Fatalf("expected first event to be accepted")
	}

	second, err := svc.IngestGitHubWebhook(ctx, cmd)
	if err != nil {
		t.Fatalf("second ingest failed: %v", err)
	}
	if !second.Duplicate {
		t.Fatalf("expected duplicate event on second delivery")
	}

	if len(events.items) != 2 {
		t.Fatalf("expected 2 flow events, got %d", len(events.items))
	}
}

func TestIngestGitHubWebhook_LearningMode_DefaultFallback(t *testing.T) {
	ctx := context.Background()
	runs := &inMemoryRunRepo{items: map[string]string{}}
	events := &inMemoryEventRepo{}
	svc := NewService(runs, events, nil, nil, nil, nil, true, TriggerLabels{})

	payload := json.RawMessage(`{"action":"opened","repository":{"id":1,"full_name":"codex-k8s/codex-k8s"},"sender":{"id":10,"login":"ai-da-stas"}}`)
	cmd := IngestCommand{
		CorrelationID: "delivery-1",
		DeliveryID:    "delivery-1",
		EventType:     "pull_request",
		ReceivedAt:    time.Now().UTC(),
		Payload:       payload,
	}

	if _, err := svc.IngestGitHubWebhook(ctx, cmd); err != nil {
		t.Fatalf("ingest failed: %v", err)
	}
	if !runs.last.LearningMode {
		t.Fatalf("expected learning mode to fallback to default=true")
	}
}

func TestIngestGitHubWebhook_IssueRunDev_CreatesRunForAllowedMember(t *testing.T) {
	ctx := context.Background()
	runs := &inMemoryRunRepo{items: map[string]string{}}
	events := &inMemoryEventRepo{}
	repos := &inMemoryRepoCfgRepo{
		byExternalID: map[int64]repocfgrepo.FindResult{
			42: {
				ProjectID:        "project-1",
				RepositoryID:     "repo-1",
				ServicesYAMLPath: "services.yaml",
			},
		},
	}
	users := &inMemoryUserRepo{
		byLogin: map[string]userrepo.User{
			"member": {
				ID:          "user-1",
				GitHubLogin: "member",
			},
		},
	}
	members := &inMemoryProjectMemberRepo{
		roles: map[string]string{
			"project-1|user-1": "read_write",
		},
	}
	svc := NewService(runs, events, repos, nil, users, members, false, TriggerLabels{})

	payload := json.RawMessage(`{
		"action":"labeled",
		"label":{"name":"run:dev"},
		"issue":{"id":1001,"number":77,"title":"Implement feature","html_url":"https://github.com/codex-k8s/codex-k8s/issues/77","state":"open","user":{"id":55,"login":"owner"}},
		"repository":{"id":42,"full_name":"codex-k8s/codex-k8s","name":"codex-k8s"},
		"sender":{"id":10,"login":"member"}
	}`)
	cmd := IngestCommand{
		CorrelationID: "delivery-77",
		DeliveryID:    "delivery-77",
		EventType:     "issues",
		ReceivedAt:    time.Now().UTC(),
		Payload:       payload,
	}

	got, err := svc.IngestGitHubWebhook(ctx, cmd)
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}
	if got.Status != "accepted" || got.Duplicate {
		t.Fatalf("unexpected result: %+v", got)
	}
	if got.RunID == "" {
		t.Fatalf("expected run id for issue trigger")
	}

	var runPayload map[string]any
	if err := json.Unmarshal(runs.last.RunPayload, &runPayload); err != nil {
		t.Fatalf("unmarshal run payload: %v", err)
	}
	trigger, ok := runPayload["trigger"].(map[string]any)
	if !ok {
		t.Fatalf("expected trigger object in run payload")
	}
	if trigger["kind"] != "dev" {
		t.Fatalf("unexpected trigger kind: %#v", trigger["kind"])
	}
	if trigger["label"] != "run:dev" {
		t.Fatalf("unexpected trigger label: %#v", trigger["label"])
	}
}

func TestIngestGitHubWebhook_IssueRunDev_DeniesUnknownSender(t *testing.T) {
	ctx := context.Background()
	runs := &inMemoryRunRepo{items: map[string]string{}}
	events := &inMemoryEventRepo{}
	repos := &inMemoryRepoCfgRepo{
		byExternalID: map[int64]repocfgrepo.FindResult{
			42: {
				ProjectID:        "project-1",
				RepositoryID:     "repo-1",
				ServicesYAMLPath: "services.yaml",
			},
		},
	}
	svc := NewService(runs, events, repos, nil, &inMemoryUserRepo{}, &inMemoryProjectMemberRepo{}, false, TriggerLabels{})

	payload := json.RawMessage(`{
		"action":"labeled",
		"label":{"name":"run:dev"},
		"issue":{"id":1001,"number":77,"title":"Implement feature","html_url":"https://github.com/codex-k8s/codex-k8s/issues/77","state":"open"},
		"repository":{"id":42,"full_name":"codex-k8s/codex-k8s","name":"codex-k8s"},
		"sender":{"id":10,"login":"unknown"}
	}`)
	cmd := IngestCommand{
		CorrelationID: "delivery-78",
		DeliveryID:    "delivery-78",
		EventType:     "issues",
		ReceivedAt:    time.Now().UTC(),
		Payload:       payload,
	}

	got, err := svc.IngestGitHubWebhook(ctx, cmd)
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}
	if got.Status != "ignored" || got.RunID != "" || got.Duplicate {
		t.Fatalf("unexpected result: %+v", got)
	}
	if len(runs.items) != 0 {
		t.Fatalf("expected no run creation for denied sender")
	}
	if len(events.items) != 1 {
		t.Fatalf("expected one flow event, got %d", len(events.items))
	}
	if events.items[0].EventType != "webhook.ignored" {
		t.Fatalf("unexpected event type: %s", events.items[0].EventType)
	}
}

func TestIngestGitHubWebhook_IssueNonTriggerLabelIgnored(t *testing.T) {
	ctx := context.Background()
	runs := &inMemoryRunRepo{items: map[string]string{}}
	events := &inMemoryEventRepo{}
	svc := NewService(runs, events, nil, nil, nil, nil, false, TriggerLabels{})

	payload := json.RawMessage(`{
		"action":"labeled",
		"label":{"name":"bug"},
		"issue":{"id":1001,"number":77},
		"repository":{"id":42,"full_name":"codex-k8s/codex-k8s","name":"codex-k8s"},
		"sender":{"id":10,"login":"member"}
	}`)
	cmd := IngestCommand{
		CorrelationID: "delivery-79",
		DeliveryID:    "delivery-79",
		EventType:     "issues",
		ReceivedAt:    time.Now().UTC(),
		Payload:       payload,
	}

	got, err := svc.IngestGitHubWebhook(ctx, cmd)
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}
	if got.Status != "ignored" || got.RunID != "" || got.Duplicate {
		t.Fatalf("unexpected result: %+v", got)
	}
	if len(runs.items) != 0 {
		t.Fatalf("expected no run creation for non-trigger issue label")
	}
}

type inMemoryRunRepo struct {
	items map[string]string
	last  agentrunrepo.CreateParams
}

func (r *inMemoryRunRepo) CreatePendingIfAbsent(_ context.Context, params agentrunrepo.CreateParams) (agentrunrepo.CreateResult, error) {
	r.last = params
	if id, ok := r.items[params.CorrelationID]; ok {
		return agentrunrepo.CreateResult{
			RunID:    id,
			Inserted: false,
		}, nil
	}
	id := "run-" + params.CorrelationID
	r.items[params.CorrelationID] = id
	return agentrunrepo.CreateResult{
		RunID:    id,
		Inserted: true,
	}, nil
}

type inMemoryEventRepo struct {
	items []floweventrepo.InsertParams
}

func (r *inMemoryEventRepo) Insert(_ context.Context, params floweventrepo.InsertParams) error {
	r.items = append(r.items, params)
	return nil
}

type inMemoryRepoCfgRepo struct {
	byExternalID map[int64]repocfgrepo.FindResult
}

func (r *inMemoryRepoCfgRepo) ListForProject(_ context.Context, _ string, _ int) ([]repocfgrepo.RepositoryBinding, error) {
	return nil, nil
}

func (r *inMemoryRepoCfgRepo) Upsert(_ context.Context, _ repocfgrepo.UpsertParams) (repocfgrepo.RepositoryBinding, error) {
	return repocfgrepo.RepositoryBinding{}, fmt.Errorf("not implemented")
}

func (r *inMemoryRepoCfgRepo) Delete(_ context.Context, _, _ string) error {
	return nil
}

func (r *inMemoryRepoCfgRepo) FindByProviderExternalID(_ context.Context, _ string, externalID int64) (repocfgrepo.FindResult, bool, error) {
	res, ok := r.byExternalID[externalID]
	if !ok {
		return repocfgrepo.FindResult{}, false, nil
	}
	return res, true, nil
}

func (r *inMemoryRepoCfgRepo) GetTokenEncrypted(_ context.Context, _ string) ([]byte, bool, error) {
	return nil, false, nil
}

type inMemoryUserRepo struct {
	byLogin map[string]userrepo.User
}

func (r *inMemoryUserRepo) EnsureOwner(_ context.Context, _ string) (userrepo.User, error) {
	return userrepo.User{}, nil
}

func (r *inMemoryUserRepo) GetByID(_ context.Context, _ string) (userrepo.User, bool, error) {
	return userrepo.User{}, false, nil
}

func (r *inMemoryUserRepo) GetByEmail(_ context.Context, _ string) (userrepo.User, bool, error) {
	return userrepo.User{}, false, nil
}

func (r *inMemoryUserRepo) GetByGitHubLogin(_ context.Context, githubLogin string) (userrepo.User, bool, error) {
	u, ok := r.byLogin[githubLogin]
	return u, ok, nil
}

func (r *inMemoryUserRepo) UpdateGitHubIdentity(_ context.Context, _ string, _ int64, _ string) error {
	return nil
}

func (r *inMemoryUserRepo) CreateAllowedUser(_ context.Context, _ string, _ bool) (userrepo.User, error) {
	return userrepo.User{}, nil
}

func (r *inMemoryUserRepo) List(_ context.Context, _ int) ([]userrepo.User, error) {
	return nil, nil
}

func (r *inMemoryUserRepo) DeleteByID(_ context.Context, _ string) error {
	return nil
}

type inMemoryProjectMemberRepo struct {
	roles map[string]string
}

func (r *inMemoryProjectMemberRepo) List(_ context.Context, _ string, _ int) ([]projectmemberrepo.Member, error) {
	return nil, nil
}

func (r *inMemoryProjectMemberRepo) Upsert(_ context.Context, _, _, _ string) error {
	return nil
}

func (r *inMemoryProjectMemberRepo) Delete(_ context.Context, _, _ string) error {
	return nil
}

func (r *inMemoryProjectMemberRepo) GetRole(_ context.Context, projectID string, userID string) (string, bool, error) {
	role, ok := r.roles[projectID+"|"+userID]
	return role, ok, nil
}

func (r *inMemoryProjectMemberRepo) SetLearningModeOverride(_ context.Context, _, _ string, _ *bool) error {
	return nil
}

func (r *inMemoryProjectMemberRepo) GetLearningModeOverride(_ context.Context, _, _ string) (*bool, bool, error) {
	return nil, false, nil
}
