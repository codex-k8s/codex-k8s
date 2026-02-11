package casters

import (
	"strings"
	"time"

	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/models"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Project(item *controlplanev1.Project) models.Project {
	if item == nil {
		return models.Project{}
	}
	return models.Project{
		ID:   item.GetId(),
		Slug: item.GetSlug(),
		Name: item.GetName(),
		Role: item.GetRole(),
	}
}

func Projects(items []*controlplanev1.Project) []models.Project {
	out := make([]models.Project, 0, len(items))
	for _, item := range items {
		out = append(out, Project(item))
	}
	return out
}

func Run(item *controlplanev1.Run) models.Run {
	if item == nil {
		return models.Run{}
	}
	return models.Run{
		ID:            item.GetId(),
		CorrelationID: item.GetCorrelationId(),
		ProjectID:     optionalString(item.ProjectId),
		ProjectSlug:   optionalStringValue(item.ProjectSlug),
		ProjectName:   optionalStringValue(item.ProjectName),
		Status:        item.GetStatus(),
		CreatedAt:     timestampString(item.GetCreatedAt()),
		StartedAt:     optionalTimestampString(item.GetStartedAt()),
		FinishedAt:    optionalTimestampString(item.GetFinishedAt()),
	}
}

func Runs(items []*controlplanev1.Run) []models.Run {
	out := make([]models.Run, 0, len(items))
	for _, item := range items {
		out = append(out, Run(item))
	}
	return out
}

func FlowEvent(item *controlplanev1.FlowEvent) models.FlowEvent {
	if item == nil {
		return models.FlowEvent{}
	}
	return models.FlowEvent{
		CorrelationID: item.GetCorrelationId(),
		EventType:     item.GetEventType(),
		CreatedAt:     timestampString(item.GetCreatedAt()),
		PayloadJSON:   item.GetPayloadJson(),
	}
}

func FlowEvents(items []*controlplanev1.FlowEvent) []models.FlowEvent {
	out := make([]models.FlowEvent, 0, len(items))
	for _, item := range items {
		out = append(out, FlowEvent(item))
	}
	return out
}

func LearningFeedback(item *controlplanev1.LearningFeedback) models.LearningFeedback {
	if item == nil {
		return models.LearningFeedback{}
	}
	return models.LearningFeedback{
		ID:           item.GetId(),
		RunID:        item.GetRunId(),
		RepositoryID: optionalString(item.RepositoryId),
		PRNumber:     optionalInt32(item.PrNumber),
		FilePath:     optionalString(item.FilePath),
		Line:         optionalInt32(item.Line),
		Kind:         item.GetKind(),
		Explanation:  item.GetExplanation(),
		CreatedAt:    timestampString(item.GetCreatedAt()),
	}
}

func LearningFeedbackList(items []*controlplanev1.LearningFeedback) []models.LearningFeedback {
	out := make([]models.LearningFeedback, 0, len(items))
	for _, item := range items {
		out = append(out, LearningFeedback(item))
	}
	return out
}

func User(item *controlplanev1.User) models.User {
	if item == nil {
		return models.User{}
	}
	return models.User{
		ID:              item.GetId(),
		Email:           item.GetEmail(),
		GitHubUserID:    optionalInt64(item.GithubUserId),
		GitHubLogin:     optionalString(item.GithubLogin),
		IsPlatformAdmin: item.GetIsPlatformAdmin(),
		IsPlatformOwner: item.GetIsPlatformOwner(),
	}
}

func Users(items []*controlplanev1.User) []models.User {
	out := make([]models.User, 0, len(items))
	for _, item := range items {
		out = append(out, User(item))
	}
	return out
}

func ProjectMember(item *controlplanev1.ProjectMember) models.ProjectMember {
	if item == nil {
		return models.ProjectMember{}
	}
	return models.ProjectMember{
		ProjectID:            item.GetProjectId(),
		UserID:               item.GetUserId(),
		Email:                item.GetEmail(),
		Role:                 item.GetRole(),
		LearningModeOverride: optionalBool(item.GetLearningModeOverride()),
	}
}

func ProjectMembers(items []*controlplanev1.ProjectMember) []models.ProjectMember {
	out := make([]models.ProjectMember, 0, len(items))
	for _, item := range items {
		out = append(out, ProjectMember(item))
	}
	return out
}

func RepositoryBinding(item *controlplanev1.RepositoryBinding) models.RepositoryBinding {
	if item == nil {
		return models.RepositoryBinding{}
	}
	return models.RepositoryBinding{
		ID:               item.GetId(),
		ProjectID:        item.GetProjectId(),
		Provider:         item.GetProvider(),
		ExternalID:       item.GetExternalId(),
		Owner:            item.GetOwner(),
		Name:             item.GetName(),
		ServicesYAMLPath: item.GetServicesYamlPath(),
	}
}

func RepositoryBindings(items []*controlplanev1.RepositoryBinding) []models.RepositoryBinding {
	out := make([]models.RepositoryBinding, 0, len(items))
	for _, item := range items {
		out = append(out, RepositoryBinding(item))
	}
	return out
}

func Me(principal *controlplanev1.Principal) models.MeResponse {
	return models.MeResponse{
		User: models.MeUser{
			ID:              principal.GetUserId(),
			Email:           principal.GetEmail(),
			GitHubLogin:     principal.GetGithubLogin(),
			IsPlatformAdmin: principal.GetIsPlatformAdmin(),
			IsPlatformOwner: principal.GetIsPlatformOwner(),
		},
	}
}

func IngestGitHubWebhook(item *controlplanev1.IngestGitHubWebhookResponse) models.IngestGitHubWebhookResponse {
	out := models.IngestGitHubWebhookResponse{}
	if item == nil {
		return out
	}
	out.CorrelationID = item.GetCorrelationId()
	out.RunID = item.GetRunId()
	out.Status = item.GetStatus()
	out.Duplicate = item.GetDuplicate()
	return out
}

func timestampString(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return ts.AsTime().UTC().Format(time.RFC3339Nano)
}

func optionalTimestampString(ts *timestamppb.Timestamp) *string {
	if ts == nil {
		return nil
	}
	v := ts.AsTime().UTC().Format(time.RFC3339Nano)
	return &v
}

func nullableString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func optionalString(value *string) *string {
	if value == nil {
		return nil
	}
	return nullableString(*value)
}

func optionalStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func optionalInt32(value *int32) *int32 {
	if value == nil {
		return nil
	}
	if *value <= 0 {
		return nil
	}
	v := *value
	return &v
}

func optionalInt64(value *int64) *int64 {
	if value == nil {
		return nil
	}
	if *value <= 0 {
		return nil
	}
	v := *value
	return &v
}

type boolValue interface {
	GetValue() bool
}

func optionalBool(v boolValue) *bool {
	if v == nil {
		return nil
	}
	value := v.GetValue()
	return &value
}
