package agentcallback

import (
	"context"
	"encoding/json"
	"testing"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
)

func TestExtractInteractionResumePayload_ReturnsCompactedPayload(t *testing.T) {
	t.Parallel()

	raw := json.RawMessage(`{
		"interaction_resume_payload": {
			"interaction_id": "interaction-1",
			"tool_name": "user.decision.request"
		}
	}`)

	payload, found, err := extractInteractionResumePayload(raw)
	if err != nil {
		t.Fatalf("extractInteractionResumePayload() error = %v", err)
	}
	if !found {
		t.Fatal("expected interaction resume payload to be found")
	}
	if got, want := string(payload), `{"interaction_id":"interaction-1","tool_name":"user.decision.request"}`; got != want {
		t.Fatalf("payload = %q, want %q", got, want)
	}
}

func TestService_GetRunInteractionResumePayload_UsesRunRepository(t *testing.T) {
	t.Parallel()

	service := &Service{
		runs: fakeInteractionResumeRunRepository{
			run: entitytypes.AgentRun{
				ID: "run-1",
				RunPayload: json.RawMessage(`{
					"interaction_resume_payload": {
						"interaction_id": "interaction-1"
					}
				}`),
			},
			found: true,
		},
	}

	payload, found, err := service.GetRunInteractionResumePayload(context.Background(), "run-1")
	if err != nil {
		t.Fatalf("GetRunInteractionResumePayload() error = %v", err)
	}
	if !found {
		t.Fatal("expected found=true")
	}
	if got, want := string(payload), `{"interaction_id":"interaction-1"}`; got != want {
		t.Fatalf("payload = %q, want %q", got, want)
	}
}

type fakeInteractionResumeRunRepository struct {
	run   entitytypes.AgentRun
	found bool
	err   error
}

func (f fakeInteractionResumeRunRepository) GetByID(context.Context, string) (entitytypes.AgentRun, bool, error) {
	if f.err != nil {
		return entitytypes.AgentRun{}, false, f.err
	}
	return f.run, f.found, nil
}
