package casters

import (
	"github.com/codex-k8s/codex-k8s/libs/go/cast"
	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/models"
	"google.golang.org/protobuf/types/known/wrapperspb"
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
		ID:              item.GetId(),
		CorrelationID:   item.GetCorrelationId(),
		ProjectID:       cast.OptionalTrimmedString(item.ProjectId),
		ProjectSlug:     cast.TrimmedStringValue(item.ProjectSlug),
		ProjectName:     cast.TrimmedStringValue(item.ProjectName),
		IssueNumber:     cast.PositiveInt32Ptr(item.IssueNumber),
		IssueURL:        cast.OptionalTrimmedString(item.IssueUrl),
		PRNumber:        cast.PositiveInt32Ptr(item.PrNumber),
		PRURL:           cast.OptionalTrimmedString(item.PrUrl),
		TriggerKind:     cast.OptionalTrimmedString(item.TriggerKind),
		TriggerLabel:    cast.OptionalTrimmedString(item.TriggerLabel),
		AgentKey:        cast.OptionalTrimmedString(item.AgentKey),
		JobName:         cast.OptionalTrimmedString(item.JobName),
		JobNamespace:    cast.OptionalTrimmedString(item.JobNamespace),
		Namespace:       cast.OptionalTrimmedString(item.Namespace),
		JobExists:       item.GetJobExists(),
		NamespaceExists: item.GetNamespaceExists(),
		WaitState:       cast.OptionalTrimmedString(item.WaitState),
		WaitReason:      cast.OptionalTrimmedString(item.WaitReason),
		WaitSince:       cast.OptionalTimestampRFC3339Nano(item.GetWaitSince()),
		LastHeartbeatAt: cast.OptionalTimestampRFC3339Nano(item.GetLastHeartbeatAt()),
		Status:          item.GetStatus(),
		CreatedAt:       cast.TimestampRFC3339Nano(item.GetCreatedAt()),
		StartedAt:       cast.OptionalTimestampRFC3339Nano(item.GetStartedAt()),
		FinishedAt:      cast.OptionalTimestampRFC3339Nano(item.GetFinishedAt()),
	}
}

func Runs(items []*controlplanev1.Run) []models.Run {
	out := make([]models.Run, 0, len(items))
	for _, item := range items {
		out = append(out, Run(item))
	}
	return out
}

func ApprovalRequest(item *controlplanev1.ApprovalRequest) models.ApprovalRequest {
	if item == nil {
		return models.ApprovalRequest{}
	}
	return models.ApprovalRequest{
		ID:            item.GetId(),
		CorrelationID: item.GetCorrelationId(),
		RunID:         cast.OptionalTrimmedString(item.RunId),
		ProjectID:     cast.OptionalTrimmedString(item.ProjectId),
		ProjectSlug:   cast.OptionalTrimmedString(item.ProjectSlug),
		ProjectName:   cast.OptionalTrimmedString(item.ProjectName),
		IssueNumber:   int32PtrFromWrapper(item.GetIssueNumber()),
		PRNumber:      int32PtrFromWrapper(item.GetPrNumber()),
		TriggerLabel:  cast.OptionalTrimmedString(item.TriggerLabel),
		ToolName:      item.GetToolName(),
		Action:        item.GetAction(),
		ApprovalMode:  item.GetApprovalMode(),
		RequestedBy:   item.GetRequestedBy(),
		CreatedAt:     cast.TimestampRFC3339Nano(item.GetCreatedAt()),
	}
}

func ApprovalRequests(items []*controlplanev1.ApprovalRequest) []models.ApprovalRequest {
	out := make([]models.ApprovalRequest, 0, len(items))
	for _, item := range items {
		out = append(out, ApprovalRequest(item))
	}
	return out
}

func ResolveApprovalDecision(item *controlplanev1.ResolveApprovalDecisionResponse) models.ResolveApprovalDecisionResponse {
	if item == nil {
		return models.ResolveApprovalDecisionResponse{}
	}
	return models.ResolveApprovalDecisionResponse{
		ID:            item.GetId(),
		CorrelationID: item.GetCorrelationId(),
		RunID:         cast.OptionalTrimmedString(item.RunId),
		ToolName:      item.GetToolName(),
		Action:        item.GetAction(),
		ApprovalState: item.GetApprovalState(),
	}
}

func int32PtrFromWrapper(value *wrapperspb.Int32Value) *int32 {
	if value == nil || value.GetValue() <= 0 {
		return nil
	}
	item := value.GetValue()
	return &item
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
		ID:                 item.GetId(),
		ProjectID:          item.GetProjectId(),
		Provider:           item.GetProvider(),
		ExternalID:         item.GetExternalId(),
		Owner:              item.GetOwner(),
		Name:               item.GetName(),
		ServicesYAMLPath:   item.GetServicesYamlPath(),
		BotUsername:        cast.OptionalTrimmedString(item.BotUsername),
		BotEmail:           cast.OptionalTrimmedString(item.BotEmail),
		PreflightUpdatedAt: cast.OptionalTrimmedString(item.PreflightUpdatedAt),
	}
}

func RepositoryBindings(items []*controlplanev1.RepositoryBinding) []models.RepositoryBinding {
	out := make([]models.RepositoryBinding, 0, len(items))
	for _, item := range items {
		out = append(out, RepositoryBinding(item))
	}
	return out
}

func ProjectGitHubTokens(item *controlplanev1.ProjectGitHubTokens) models.ProjectGitHubTokens {
	if item == nil {
		return models.ProjectGitHubTokens{}
	}
	return models.ProjectGitHubTokens{
		ProjectID:        item.GetProjectId(),
		HasPlatformToken: item.GetHasPlatformToken(),
		HasBotToken:      item.GetHasBotToken(),
		BotUsername:      cast.OptionalTrimmedString(item.BotUsername),
		BotEmail:         cast.OptionalTrimmedString(item.BotEmail),
	}
}

func PreflightCheckResult(item *controlplanev1.PreflightCheckResult) models.PreflightCheckResult {
	if item == nil {
		return models.PreflightCheckResult{}
	}
	return models.PreflightCheckResult{
		Name:    item.GetName(),
		Status:  item.GetStatus(),
		Details: cast.OptionalTrimmedString(item.Details),
	}
}

func RunRepositoryPreflightResponse(item *controlplanev1.RunRepositoryPreflightResponse) models.RunRepositoryPreflightResponse {
	if item == nil {
		return models.RunRepositoryPreflightResponse{Checks: []models.PreflightCheckResult{}}
	}
	out := models.RunRepositoryPreflightResponse{
		RepositoryID: item.GetRepositoryId(),
		Status:       item.GetStatus(),
		Checks:       make([]models.PreflightCheckResult, 0, len(item.GetChecks())),
		ReportJSON:   item.GetReportJson(),
		FinishedAt:   cast.TimestampRFC3339Nano(item.GetFinishedAt()),
	}
	for _, check := range item.GetChecks() {
		out.Checks = append(out.Checks, PreflightCheckResult(check))
	}
	return out
}

func ConfigEntry(item *controlplanev1.ConfigEntry) models.ConfigEntry {
	if item == nil {
		return models.ConfigEntry{SyncTargets: []string{}}
	}
	out := models.ConfigEntry{
		ID:           item.GetId(),
		Scope:        item.GetScope(),
		Kind:         item.GetKind(),
		ProjectID:    cast.OptionalTrimmedString(item.ProjectId),
		RepositoryID: cast.OptionalTrimmedString(item.RepositoryId),
		Key:          item.GetKey(),
		Value:        cast.OptionalTrimmedString(item.Value),
		SyncTargets:  make([]string, 0, len(item.GetSyncTargets())),
		Mutability:   item.GetMutability(),
		IsDangerous:  item.GetIsDangerous(),
		UpdatedAt:    cast.OptionalTrimmedString(item.UpdatedAt),
	}
	out.SyncTargets = append(out.SyncTargets, item.GetSyncTargets()...)
	return out
}

func ConfigEntries(items []*controlplanev1.ConfigEntry) []models.ConfigEntry {
	out := make([]models.ConfigEntry, 0, len(items))
	for _, item := range items {
		out = append(out, ConfigEntry(item))
	}
	return out
}

func DocsetGroup(item *controlplanev1.DocsetGroup) models.DocsetGroup {
	if item == nil {
		return models.DocsetGroup{}
	}
	return models.DocsetGroup{
		ID:              item.GetId(),
		Title:           item.GetTitle(),
		Description:     item.GetDescription(),
		DefaultSelected: item.GetDefaultSelected(),
	}
}

func DocsetGroups(items []*controlplanev1.DocsetGroup) []models.DocsetGroup {
	out := make([]models.DocsetGroup, 0, len(items))
	for _, item := range items {
		out = append(out, DocsetGroup(item))
	}
	return out
}

func ImportDocsetResponse(item *controlplanev1.ImportDocsetResponse) models.ImportDocsetResponse {
	if item == nil {
		return models.ImportDocsetResponse{}
	}
	return models.ImportDocsetResponse{
		RepositoryFullName: item.GetRepositoryFullName(),
		PRNumber:           item.GetPrNumber(),
		PRURL:              item.GetPrUrl(),
		Branch:             item.GetBranch(),
		FilesTotal:         item.GetFilesTotal(),
	}
}

func SyncDocsetResponse(item *controlplanev1.SyncDocsetResponse) models.SyncDocsetResponse {
	if item == nil {
		return models.SyncDocsetResponse{}
	}
	return models.SyncDocsetResponse{
		RepositoryFullName: item.GetRepositoryFullName(),
		PRNumber:           item.GetPrNumber(),
		PRURL:              item.GetPrUrl(),
		Branch:             item.GetBranch(),
		FilesUpdated:       item.GetFilesUpdated(),
		FilesDrift:         item.GetFilesDrift(),
	}
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

func RunLogs(item *controlplanev1.RunLogs) models.RunLogs {
	out := models.RunLogs{}
	if item == nil {
		return out
	}
	out.RunID = item.GetRunId()
	out.Status = item.GetStatus()
	out.UpdatedAt = cast.OptionalTimestampRFC3339Nano(item.GetUpdatedAt())
	out.SnapshotJSON = item.GetSnapshotJson()
	out.TailLines = item.GetTailLines()
	return out
}

func RuntimeDeployTaskLog(item *controlplanev1.RuntimeDeployTaskLog) models.RuntimeDeployTaskLog {
	out := models.RuntimeDeployTaskLog{}
	if item == nil {
		return out
	}
	out.Stage = item.GetStage()
	out.Level = item.GetLevel()
	out.Message = item.GetMessage()
	out.CreatedAt = cast.OptionalTimestampRFC3339Nano(item.GetCreatedAt())
	return out
}

func RuntimeDeployTask(item *controlplanev1.RuntimeDeployTask) models.RuntimeDeployTask {
	out := models.RuntimeDeployTask{}
	if item == nil {
		return out
	}
	out.RunID = item.GetRunId()
	out.RuntimeMode = item.GetRuntimeMode()
	out.Namespace = item.GetNamespace()
	out.TargetEnv = item.GetTargetEnv()
	out.SlotNo = item.GetSlotNo()
	out.RepositoryFullName = item.GetRepositoryFullName()
	out.ServicesYAMLPath = item.GetServicesYamlPath()
	out.BuildRef = item.GetBuildRef()
	out.DeployOnly = item.GetDeployOnly()
	out.Status = item.GetStatus()
	out.LeaseOwner = cast.OptionalTrimmedString(item.LeaseOwner)
	out.LeaseUntil = cast.OptionalTimestampRFC3339Nano(item.GetLeaseUntil())
	out.Attempts = item.GetAttempts()
	out.LastError = cast.OptionalTrimmedString(item.LastError)
	out.ResultNamespace = cast.OptionalTrimmedString(item.ResultNamespace)
	out.ResultTargetEnv = cast.OptionalTrimmedString(item.ResultTargetEnv)
	out.CreatedAt = cast.OptionalTimestampRFC3339Nano(item.GetCreatedAt())
	out.UpdatedAt = cast.OptionalTimestampRFC3339Nano(item.GetUpdatedAt())
	out.StartedAt = cast.OptionalTimestampRFC3339Nano(item.GetStartedAt())
	out.FinishedAt = cast.OptionalTimestampRFC3339Nano(item.GetFinishedAt())
	out.Logs = make([]models.RuntimeDeployTaskLog, 0, len(item.GetLogs()))
	for _, logItem := range item.GetLogs() {
		out.Logs = append(out.Logs, RuntimeDeployTaskLog(logItem))
	}
	return out
}

func RuntimeDeployTasks(items []*controlplanev1.RuntimeDeployTask) []models.RuntimeDeployTask {
	out := make([]models.RuntimeDeployTask, 0, len(items))
	for _, item := range items {
		out = append(out, RuntimeDeployTask(item))
	}
	return out
}

func RegistryImageTag(item *controlplanev1.RegistryImageTag) models.RegistryImageTag {
	out := models.RegistryImageTag{}
	if item == nil {
		return out
	}
	out.Tag = item.GetTag()
	out.Digest = item.GetDigest()
	out.CreatedAt = cast.OptionalTimestampRFC3339Nano(item.GetCreatedAt())
	out.ConfigSizeBytes = item.GetConfigSizeBytes()
	return out
}

func RegistryImageRepository(item *controlplanev1.RegistryImageRepository) models.RegistryImageRepository {
	out := models.RegistryImageRepository{}
	if item == nil {
		return out
	}
	out.Repository = item.GetRepository()
	out.TagCount = item.GetTagCount()
	out.Tags = make([]models.RegistryImageTag, 0, len(item.GetTags()))
	for _, tagItem := range item.GetTags() {
		out.Tags = append(out.Tags, RegistryImageTag(tagItem))
	}
	return out
}

func RegistryImageRepositories(items []*controlplanev1.RegistryImageRepository) []models.RegistryImageRepository {
	out := make([]models.RegistryImageRepository, 0, len(items))
	for _, item := range items {
		out = append(out, RegistryImageRepository(item))
	}
	return out
}

func RegistryImageDeleteResult(item *controlplanev1.RegistryImageDeleteResult) models.RegistryImageDeleteResult {
	out := models.RegistryImageDeleteResult{}
	if item == nil {
		return out
	}
	out.Repository = item.GetRepository()
	out.Tag = item.GetTag()
	out.Digest = item.GetDigest()
	out.Deleted = item.GetDeleted()
	return out
}

func RegistryImageCleanupResult(item *controlplanev1.CleanupRegistryImagesResponse) models.CleanupRegistryImagesResponse {
	out := models.CleanupRegistryImagesResponse{}
	if item == nil {
		return out
	}
	out.RepositoriesScanned = item.GetRepositoriesScanned()
	out.TagsDeleted = item.GetTagsDeleted()
	out.TagsSkipped = item.GetTagsSkipped()
	out.Deleted = make([]models.RegistryImageDeleteResult, 0, len(item.GetDeleted()))
	for _, deleteItem := range item.GetDeleted() {
		out.Deleted = append(out.Deleted, RegistryImageDeleteResult(deleteItem))
	}
	out.Skipped = make([]models.RegistryImageDeleteResult, 0, len(item.GetSkipped()))
	for _, skipItem := range item.GetSkipped() {
		out.Skipped = append(out.Skipped, RegistryImageDeleteResult(skipItem))
	}
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
