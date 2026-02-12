package internalapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	mcpdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/mcp"
	agentsessionrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentsession"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/flowevent"
)

type mcpTokenVerifier interface {
	VerifyRunToken(ctx context.Context, rawToken string) (mcpdomain.SessionContext, error)
}

// Dependencies configures internal agent HTTP transport.
type Dependencies struct {
	Sessions   agentsessionrepo.Repository
	FlowEvents floweventrepo.Repository
	MCP        mcpTokenVerifier
	Logger     *slog.Logger
}

type handler struct {
	sessions   agentsessionrepo.Repository
	flowEvents floweventrepo.Repository
	mcp        mcpTokenVerifier
	logger     *slog.Logger
	now        func() time.Time
}

// NewHandler builds internal agent callback HTTP handler.
func NewHandler(deps Dependencies) http.Handler {
	h := &handler{
		sessions:   deps.Sessions,
		flowEvents: deps.FlowEvents,
		mcp:        deps.MCP,
		logger:     deps.Logger,
		now:        time.Now,
	}
	if h.logger == nil {
		h.logger = slog.Default()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/internal/agent/session", h.handleSessionUpsert)
	mux.HandleFunc("/internal/agent/session/latest", h.handleSessionLatest)
	mux.HandleFunc("/internal/agent/event", h.handleEventInsert)
	return mux
}

func (h *handler) handleSessionUpsert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}
	if h.sessions == nil {
		http.Error(w, "agent session repository is not configured", http.StatusServiceUnavailable)
		return
	}

	session, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	var req sessionUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	runID := strings.TrimSpace(req.RunID)
	if runID == "" {
		runID = session.RunID
	}
	if runID == "" {
		http.Error(w, "run_id is required", http.StatusBadRequest)
		return
	}
	if runID != session.RunID {
		http.Error(w, "run_id mismatch with token", http.StatusForbidden)
		return
	}

	repositoryFullName := strings.TrimSpace(req.RepositoryFullName)
	if repositoryFullName == "" {
		http.Error(w, "repository_full_name is required", http.StatusBadRequest)
		return
	}
	branchName := strings.TrimSpace(req.BranchName)
	if branchName == "" {
		http.Error(w, "branch_name is required", http.StatusBadRequest)
		return
	}

	startedAt := h.now().UTC()
	if req.StartedAt != nil {
		startedAt = req.StartedAt.UTC()
	}

	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = sessionStatusRunning
	}

	correlationID := strings.TrimSpace(req.CorrelationID)
	if correlationID == "" {
		correlationID = session.CorrelationID
	}
	if correlationID == "" {
		http.Error(w, "correlation_id is required", http.StatusBadRequest)
		return
	}

	projectID := strings.TrimSpace(req.ProjectID)
	if projectID == "" {
		projectID = session.ProjectID
	}

	if err := h.sessions.Upsert(r.Context(), agentsessionrepo.UpsertParams{
		RunID:              runID,
		CorrelationID:      correlationID,
		ProjectID:          projectID,
		RepositoryFullName: repositoryFullName,
		IssueNumber:        req.IssueNumber,
		BranchName:         branchName,
		PRNumber:           req.PRNumber,
		PRURL:              strings.TrimSpace(req.PRURL),
		TriggerKind:        strings.TrimSpace(req.TriggerKind),
		TemplateKind:       strings.TrimSpace(req.TemplateKind),
		TemplateSource:     strings.TrimSpace(req.TemplateSource),
		TemplateLocale:     strings.TrimSpace(req.TemplateLocale),
		Model:              strings.TrimSpace(req.Model),
		ReasoningEffort:    strings.TrimSpace(req.ReasoningEffort),
		Status:             status,
		SessionID:          strings.TrimSpace(req.SessionID),
		SessionJSON:        req.SessionJSON,
		CodexSessionPath:   strings.TrimSpace(req.CodexSessionPath),
		CodexSessionJSON:   req.CodexSessionJSON,
		StartedAt:          startedAt,
		FinishedAt:         normalizeOptionalTime(req.FinishedAt),
	}); err != nil {
		h.logger.Error("upsert agent session failed", "run_id", runID, "err", err)
		http.Error(w, "failed to persist agent session", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, sessionUpsertResponse{
		OK:    true,
		RunID: runID,
	})
}

func (h *handler) handleSessionLatest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}
	if h.sessions == nil {
		http.Error(w, "agent session repository is not configured", http.StatusServiceUnavailable)
		return
	}

	if _, ok := h.authenticate(w, r); !ok {
		return
	}

	repositoryFullName := strings.TrimSpace(r.URL.Query().Get("repository_full_name"))
	branchName := strings.TrimSpace(r.URL.Query().Get("branch_name"))
	if repositoryFullName == "" || branchName == "" {
		http.Error(w, "repository_full_name and branch_name are required", http.StatusBadRequest)
		return
	}

	item, found, err := h.sessions.GetLatestByRepositoryBranch(r.Context(), repositoryFullName, branchName)
	if err != nil {
		h.logger.Error("get latest agent session failed", "repository_full_name", repositoryFullName, "branch_name", branchName, "err", err)
		http.Error(w, "failed to load latest agent session", http.StatusInternalServerError)
		return
	}
	if !found {
		writeJSON(w, http.StatusOK, latestSessionResponse{Found: false})
		return
	}

	resp := latestSessionResponse{
		Found: true,
		Session: &sessionSnapshotDTO{
			RunID:              item.RunID,
			CorrelationID:      item.CorrelationID,
			ProjectID:          item.ProjectID,
			RepositoryFullName: item.RepositoryFullName,
			IssueNumber:        item.IssueNumber,
			BranchName:         item.BranchName,
			PRNumber:           item.PRNumber,
			PRURL:              item.PRURL,
			TriggerKind:        item.TriggerKind,
			TemplateKind:       item.TemplateKind,
			TemplateSource:     item.TemplateSource,
			TemplateLocale:     item.TemplateLocale,
			Model:              item.Model,
			ReasoningEffort:    item.ReasoningEffort,
			Status:             item.Status,
			SessionID:          item.SessionID,
			SessionJSON:        item.SessionJSON,
			CodexSessionPath:   item.CodexSessionPath,
			CodexSessionJSON:   item.CodexSessionJSON,
			StartedAt:          item.StartedAt,
			CreatedAt:          item.CreatedAt,
			UpdatedAt:          item.UpdatedAt,
		},
	}
	if !item.FinishedAt.IsZero() {
		finishedAt := item.FinishedAt
		resp.Session.FinishedAt = &finishedAt
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *handler) handleEventInsert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}
	if h.flowEvents == nil {
		http.Error(w, "flow event repository is not configured", http.StatusServiceUnavailable)
		return
	}

	session, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	var req eventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	runID := strings.TrimSpace(req.RunID)
	if runID == "" {
		runID = session.RunID
	}
	if runID != session.RunID {
		http.Error(w, "run_id mismatch with token", http.StatusForbidden)
		return
	}

	eventType, err := parseAgentEventType(req.EventType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	payload := req.Payload
	if len(payload) == 0 {
		payload = json.RawMessage(`{}`)
	}

	if err := h.flowEvents.Insert(r.Context(), floweventrepo.InsertParams{
		CorrelationID: session.CorrelationID,
		ActorType:     floweventdomain.ActorTypeAgent,
		ActorID:       floweventdomain.ActorIDAgentRunner,
		EventType:     eventType,
		Payload:       payload,
		CreatedAt:     h.now().UTC(),
	}); err != nil {
		h.logger.Error("insert agent flow event failed", "run_id", runID, "event_type", eventType, "err", err)
		http.Error(w, "failed to persist flow event", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, eventInsertResponse{
		OK:        true,
		EventType: string(eventType),
	})
}

func (h *handler) authenticate(w http.ResponseWriter, r *http.Request) (mcpdomain.SessionContext, bool) {
	if h.mcp == nil {
		http.Error(w, "mcp verifier is not configured", http.StatusServiceUnavailable)
		return mcpdomain.SessionContext{}, false
	}

	rawToken := bearerToken(r.Header.Get("Authorization"))
	if rawToken == "" {
		http.Error(w, "missing bearer token", http.StatusUnauthorized)
		return mcpdomain.SessionContext{}, false
	}

	session, err := h.mcp.VerifyRunToken(r.Context(), rawToken)
	if err != nil {
		http.Error(w, "invalid bearer token", http.StatusUnauthorized)
		return mcpdomain.SessionContext{}, false
	}

	return session, true
}

func bearerToken(value string) string {
	token := strings.TrimSpace(value)
	if token == "" {
		return ""
	}
	parts := strings.SplitN(token, " ", 2)
	if len(parts) != 2 {
		return ""
	}
	if !strings.EqualFold(parts[0], "bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func parseAgentEventType(value string) (floweventdomain.EventType, error) {
	eventType := floweventdomain.EventType(strings.TrimSpace(value))
	switch eventType {
	case floweventdomain.EventTypeRunAgentStarted,
		floweventdomain.EventTypeRunAgentSessionRestored,
		floweventdomain.EventTypeRunAgentSessionSaved,
		floweventdomain.EventTypeRunAgentResumeUsed,
		floweventdomain.EventTypeRunPRCreated,
		floweventdomain.EventTypeRunPRUpdated,
		floweventdomain.EventTypeRunRevisePRNotFound,
		floweventdomain.EventTypeRunFailedPrecondition:
		return eventType, nil
	default:
		return "", fmt.Errorf("unsupported event_type %q", value)
	}
}

func normalizeOptionalTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	v := value.UTC()
	return &v
}

func writeMethodNotAllowed(w http.ResponseWriter, allowedMethod string) {
	w.Header().Set("Allow", allowedMethod)
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func writeJSON(w http.ResponseWriter, statusCode int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(value)
}
