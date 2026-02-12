package casters

import (
	"github.com/codex-k8s/codex-k8s/libs/go/cast"
	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/models"
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
		ProjectID:     cast.OptionalTrimmedString(item.ProjectId),
		ProjectSlug:   cast.TrimmedStringValue(item.ProjectSlug),
		ProjectName:   cast.TrimmedStringValue(item.ProjectName),
		Status:        item.GetStatus(),
		CreatedAt:     cast.TimestampRFC3339Nano(item.GetCreatedAt()),
		StartedAt:     cast.OptionalTimestampRFC3339Nano(item.GetStartedAt()),
		FinishedAt:    cast.OptionalTimestampRFC3339Nano(item.GetFinishedAt()),
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
		CreatedAt:     cast.TimestampRFC3339Nano(item.GetCreatedAt()),
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
		RepositoryID: cast.OptionalTrimmedString(item.RepositoryId),
		PRNumber:     cast.PositiveInt32Ptr(item.PrNumber),
		FilePath:     cast.OptionalTrimmedString(item.FilePath),
		Line:         cast.PositiveInt32Ptr(item.Line),
		Kind:         item.GetKind(),
		Explanation:  item.GetExplanation(),
		CreatedAt:    cast.TimestampRFC3339Nano(item.GetCreatedAt()),
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
		GitHubUserID:    cast.PositiveInt64Ptr(item.GithubUserId),
		GitHubLogin:     cast.OptionalTrimmedString(item.GithubLogin),
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
		LearningModeOverride: cast.BoolPtr(item.GetLearningModeOverride()),
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

func RunNamespaceDelete(item *controlplanev1.DeleteRunNamespaceResponse) models.RunNamespaceCleanupResponse {
	out := models.RunNamespaceCleanupResponse{}
	if item == nil {
		return out
	}
	out.RunID = item.GetRunId()
	out.Namespace = item.GetNamespace()
	out.Deleted = item.GetDeleted()
	out.AlreadyDeleted = item.GetAlreadyDeleted()
	out.CommentURL = item.GetCommentUrl()
	return out
}
