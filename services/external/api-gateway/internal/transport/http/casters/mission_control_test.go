package casters

import (
	"encoding/json"
	"testing"
	"time"

	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/generated"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestMissionControlCommandRequest_DiscussionCreatePayload(t *testing.T) {
	t.Parallel()

	parentKind := generated.DiscussionCreatePayloadParentEntityKindWorkItem
	targetKind := generated.MissionControlDiscussionCreateCommandRequestTargetEntityKindDiscussion
	parentPublicID := " discussion-42 "
	bodyMarkdown := "  body  "

	body := generated.MissionControlCommandRequest{}
	if err := body.FromMissionControlDiscussionCreateCommandRequest(generated.MissionControlDiscussionCreateCommandRequest{
		BusinessIntentKey: "intent-1",
		CommandKind:       generated.DiscussionCreate,
		ProjectId:         "project-1",
		TargetEntityKind:  &targetKind,
		Payload: generated.DiscussionCreatePayload{
			Title:                "Open thread",
			BodyMarkdown:         &bodyMarkdown,
			ParentEntityKind:     &parentKind,
			ParentEntityPublicId: &parentPublicID,
		},
	}); err != nil {
		t.Fatalf("FromMissionControlDiscussionCreateCommandRequest() error = %v", err)
	}

	req, err := MissionControlCommandRequest(body, "corr-1", time.Unix(123, 0).UTC())
	if err != nil {
		t.Fatalf("MissionControlCommandRequest() error = %v", err)
	}
	if req.GetCommandKind() != "discussion.create" {
		t.Fatalf("expected command kind %q, got %q", "discussion.create", req.GetCommandKind())
	}
	if req.GetBusinessIntentKey() != "intent-1" {
		t.Fatalf("expected business intent key %q, got %q", "intent-1", req.GetBusinessIntentKey())
	}
	if req.GetCorrelationId() != "corr-1" {
		t.Fatalf("expected correlation id %q, got %q", "corr-1", req.GetCorrelationId())
	}
	if req.GetRequestedAt() == nil || !req.GetRequestedAt().AsTime().Equal(time.Unix(123, 0).UTC()) {
		t.Fatalf("expected requested_at to be preserved, got %v", req.GetRequestedAt())
	}
	if req.GetTargetEntityKind() != "discussion" {
		t.Fatalf("expected target entity kind %q, got %q", "discussion", req.GetTargetEntityKind())
	}

	payload := req.GetDiscussionCreate()
	if payload == nil {
		t.Fatal("expected discussion.create payload")
	}
	if payload.GetTitle() != "Open thread" {
		t.Fatalf("expected title %q, got %q", "Open thread", payload.GetTitle())
	}
	if payload.GetBodyMarkdown() != "body" {
		t.Fatalf("expected trimmed body markdown %q, got %q", "body", payload.GetBodyMarkdown())
	}
	if payload.GetParentEntityKind() != "work_item" {
		t.Fatalf("expected parent entity kind %q, got %q", "work_item", payload.GetParentEntityKind())
	}
	if payload.GetParentEntityPublicId() != "discussion-42" {
		t.Fatalf("expected trimmed parent public id %q, got %q", "discussion-42", payload.GetParentEntityPublicId())
	}
}

func TestMissionControlEntityDetails_CastsDiscussionPayload(t *testing.T) {
	t.Parallel()

	details, err := MissionControlEntityDetails(&controlplanev1.MissionControlEntityDetails{
		Entity: &controlplanev1.MissionControlEntityCard{
			EntityKind:     "discussion",
			EntityPublicId: "discussion-42",
			Title:          "Discussion title",
			State:          "working",
			SyncStatus:     "synced",
			ProviderReference: &controlplanev1.MissionControlProviderReference{
				Provider:   "github",
				ExternalId: "123",
			},
		},
		DetailPayload: &controlplanev1.MissionControlEntityDetails_Discussion{
			Discussion: &controlplanev1.MissionControlDiscussionDetailsPayload{
				DiscussionKind:       stringPtr("issue_comment_thread"),
				Status:               stringPtr("open"),
				Author:               stringPtr("octocat"),
				ParticipantCount:     3,
				LatestCommentExcerpt: stringPtr("Latest update"),
				FormalizationTarget:  stringPtr("work_item"),
			},
		},
		TimelinePreview: []*controlplanev1.MissionControlTimelineEntry{
			{
				EntryId:        "entry-1",
				EntityKind:     "discussion",
				EntityPublicId: "discussion-42",
				SourceKind:     "provider",
				SourceRef:      "comment:1",
				OccurredAt:     timestamppb.New(time.Unix(456, 0).UTC()),
				Summary:        "First comment",
			},
		},
	})
	if err != nil {
		t.Fatalf("MissionControlEntityDetails() error = %v", err)
	}

	payload, err := details.DetailPayload.AsDiscussionDetailsPayload()
	if err != nil {
		t.Fatalf("AsDiscussionDetailsPayload() error = %v", err)
	}
	if payload.DiscussionKind != "issue_comment_thread" {
		t.Fatalf("expected discussion kind %q, got %q", "issue_comment_thread", payload.DiscussionKind)
	}
	if payload.ParticipantCount == nil || *payload.ParticipantCount != 3 {
		t.Fatalf("expected participant_count 3, got %v", payload.ParticipantCount)
	}
	if len(details.TimelinePreview) != 1 {
		t.Fatalf("expected one timeline preview item, got %d", len(details.TimelinePreview))
	}
}

func TestMissionControlCommandState_RequiredCollectionsMarshalAsArrays(t *testing.T) {
	t.Parallel()

	raw, err := json.Marshal(MissionControlCommandState(&controlplanev1.MissionControlCommandState{
		CommandId:         "cmd-1",
		CommandKind:       "discussion.create",
		Status:            "accepted",
		CorrelationId:     "corr-1",
		BusinessIntentKey: "intent-1",
		UpdatedAt:         timestamppb.New(time.Unix(789, 0).UTC()),
	}))
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if _, ok := decoded["entity_refs"].([]any); !ok {
		t.Fatalf("expected entity_refs to marshal as array, got %T", decoded["entity_refs"])
	}
	if _, ok := decoded["provider_delivery_ids"].([]any); !ok {
		t.Fatalf("expected provider_delivery_ids to marshal as array, got %T", decoded["provider_delivery_ids"])
	}
}

func stringPtr(value string) *string {
	return &value
}
