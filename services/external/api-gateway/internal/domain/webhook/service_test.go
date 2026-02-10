package webhook

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	agentrunrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/agentrun"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/flowevent"
)

func TestIngestGitHubWebhook_Dedup(t *testing.T) {
	ctx := context.Background()
	runs := &inMemoryRunRepo{items: map[string]string{}}
	events := &inMemoryEventRepo{}
	svc := NewService(runs, events, nil, nil, nil, nil, false)

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
	svc := NewService(runs, events, nil, nil, nil, nil, true)

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
