package http

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v5"

	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/casters"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/models"
)

const (
	runRealtimeDefaultEventsLimit = 200
	runRealtimeDefaultTailLines   = 200
	realtimeFetchTimeout          = 8 * time.Second
	realtimeWriteTimeout          = 10 * time.Second
	realtimePongTimeout           = 60 * time.Second
	realtimeUpdateInterval        = 2 * time.Second
	realtimePingInterval          = 25 * time.Second
	runWaitResumedFlowEventType   = "run.wait.resumed"
	defaultWaitResolutionKind     = "auto_resumed"
	defaultWaitContourKind        = "agent_bot_token"
)

var staffRealtimeUpgrader = websocket.Upgrader{
	HandshakeTimeout: 10 * time.Second,
	ReadBufferSize:   4 * 1024,
	WriteBufferSize:  4 * 1024,
	CheckOrigin:      allowStaffRealtimeOrigin,
}

type runRealtimeSnapshot struct {
	Run    models.Run
	Events []models.FlowEvent
	Logs   *models.RunLogs
}

type runRealtimeFingerprint struct {
	run    string
	events string
	logs   string
}

type runWaitResolvedEventPayload struct {
	WaitID         string `json:"wait_id"`
	ContourKind    string `json:"contour_kind"`
	ResolutionKind string `json:"resolution_kind"`
}

func allowStaffRealtimeOrigin(r *http.Request) bool {
	originRaw := strings.TrimSpace(r.Header.Get("Origin"))
	if originRaw == "" {
		return true
	}
	originURL, err := url.Parse(originRaw)
	if err != nil {
		return false
	}

	originHost := strings.TrimSpace(originURL.Hostname())
	if originHost == "" {
		return false
	}

	requestHost := strings.TrimSpace(r.Host)
	if host, _, splitErr := net.SplitHostPort(requestHost); splitErr == nil {
		requestHost = strings.TrimSpace(host)
	}
	if requestHost == "" {
		requestHost = strings.TrimSpace(r.URL.Hostname())
	}
	if requestHost == "" {
		return false
	}

	return strings.EqualFold(originHost, requestHost)
}

func resolveRunRealtimeArg(defaultEventsLimit int, defaultTailLines int) func(c *echo.Context) (runRealtimeArg, error) {
	return func(c *echo.Context) (runRealtimeArg, error) {
		logsArg, err := resolveRunLogsArg(defaultTailLines)(c)
		if err != nil {
			return runRealtimeArg{}, err
		}

		eventsLimit, err := parseLimit(c, defaultEventsLimit)
		if err != nil {
			return runRealtimeArg{}, err
		}

		includeLogs, err := parseOptionalBool(c.QueryParam("include_logs"), "include_logs")
		if err != nil {
			return runRealtimeArg{}, err
		}

		return runRealtimeArg{
			runID:       logsArg.runID,
			eventsLimit: int32(eventsLimit),
			includeLogs: includeLogs,
			tailLines:   logsArg.tailLines,
		}, nil
	}
}

// RunRealtime opens one authenticated websocket stream for run updates/events/logs.
func (h *staffHandler) RunRealtime(c *echo.Context) error {
	return withPrincipalAndResolved(c, resolveRunRealtimeArg(runRealtimeDefaultEventsLimit, runRealtimeDefaultTailLines), func(principal *controlplanev1.Principal, arg runRealtimeArg) error {
		return h.streamRunRealtime(c, principal, arg)
	})
}

func (h *staffHandler) streamRunRealtime(c *echo.Context, principal *controlplanev1.Principal, arg runRealtimeArg) error {
	conn, err := staffRealtimeUpgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		// Upgrader already wrote response details on error.
		return nil
	}
	defer conn.Close()

	conn.SetReadLimit(64 * 1024)
	_ = conn.SetReadDeadline(time.Now().Add(realtimePongTimeout))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(realtimePongTimeout))
	})

	readerDone := make(chan struct{})
	go func() {
		defer close(readerDone)
		for {
			if _, _, readErr := conn.ReadMessage(); readErr != nil {
				return
			}
		}
	}()

	fetchCtx, cancelFetch := context.WithTimeout(c.Request().Context(), realtimeFetchTimeout)
	initialSnapshot, err := h.fetchRunRealtimeSnapshot(fetchCtx, principal, arg)
	cancelFetch()
	if err != nil {
		_ = writeRealtimeJSONMessage(conn, runRealtimeErrorMessage(err))
		sendRealtimeClose(conn)
		return nil
	}

	if err := writeRealtimeJSONMessage(conn, runRealtimeSnapshotMessage(initialSnapshot)); err != nil {
		return nil
	}

	fingerprint := buildRunRealtimeFingerprint(initialSnapshot)
	currentSnapshot := initialSnapshot

	updateTicker := time.NewTicker(realtimeUpdateInterval)
	defer updateTicker.Stop()
	pingTicker := time.NewTicker(realtimePingInterval)
	defer pingTicker.Stop()

	for {
		select {
		case <-c.Request().Context().Done():
			sendRealtimeClose(conn)
			return nil
		case <-readerDone:
			return nil
		case <-pingTicker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(realtimeWriteTimeout))
			if pingErr := conn.WriteMessage(websocket.PingMessage, nil); pingErr != nil {
				return nil
			}
		case <-updateTicker.C:
			nextFetchCtx, nextCancel := context.WithTimeout(c.Request().Context(), realtimeFetchTimeout)
			nextSnapshot, fetchErr := h.fetchRunRealtimeSnapshot(nextFetchCtx, principal, arg)
			nextCancel()
			if fetchErr != nil {
				_ = writeRealtimeJSONMessage(conn, runRealtimeErrorMessage(fetchErr))
				continue
			}

			nextFingerprint := buildRunRealtimeFingerprint(nextSnapshot)
			waitMessages := buildRunWaitRealtimeMessages(currentSnapshot, nextSnapshot)
			for _, waitMessage := range waitMessages {
				if writeErr := writeRealtimeJSONMessage(conn, waitMessage); writeErr != nil {
					return nil
				}
			}
			if nextFingerprint.run != fingerprint.run {
				if writeErr := writeRealtimeJSONMessage(conn, runRealtimeRunMessage(nextSnapshot.Run)); writeErr != nil {
					return nil
				}
			}
			if nextFingerprint.events != fingerprint.events {
				if writeErr := writeRealtimeJSONMessage(conn, runRealtimeEventsMessage(nextSnapshot.Events)); writeErr != nil {
					return nil
				}
			}
			if nextSnapshot.Logs != nil && nextFingerprint.logs != fingerprint.logs {
				if writeErr := writeRealtimeJSONMessage(conn, runRealtimeLogsMessage(*nextSnapshot.Logs)); writeErr != nil {
					return nil
				}
			}

			fingerprint = nextFingerprint
			currentSnapshot = nextSnapshot
		}
	}
}

func (h *staffHandler) fetchRunRealtimeSnapshot(ctx context.Context, principal *controlplanev1.Principal, arg runRealtimeArg) (runRealtimeSnapshot, error) {
	runItem, err := h.getRunCall(ctx, principal, arg.runID)
	if err != nil {
		return runRealtimeSnapshot{}, err
	}
	eventsResp, err := h.listRunEventsCall(ctx, principal, runEventsArg{
		runID:          arg.runID,
		limit:          arg.eventsLimit,
		includePayload: true,
	})
	if err != nil {
		return runRealtimeSnapshot{}, err
	}

	out := runRealtimeSnapshot{
		Run:    casters.Run(runItem),
		Events: casters.FlowEvents(eventsResp.GetItems()),
	}

	if arg.includeLogs {
		logsItem, logsErr := h.getRunLogsCall(ctx, principal, runLogsArg{
			runID:           arg.runID,
			tailLines:       arg.tailLines,
			includeSnapshot: false,
		})
		if logsErr != nil {
			return runRealtimeSnapshot{}, logsErr
		}
		logs := casters.RunLogs(logsItem)
		out.Logs = &logs
	}

	return out, nil
}

func writeRealtimeJSONMessage(conn *websocket.Conn, message any) error {
	_ = conn.SetWriteDeadline(time.Now().Add(realtimeWriteTimeout))
	return conn.WriteJSON(message)
}

func sendRealtimeClose(conn *websocket.Conn) {
	_ = conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "closing"),
		time.Now().Add(realtimeWriteTimeout),
	)
}

func buildRunRealtimeFingerprint(snapshot runRealtimeSnapshot) runRealtimeFingerprint {
	fp := runRealtimeFingerprint{
		run:    marshalRealtimeFingerprint(snapshot.Run),
		events: marshalRealtimeFingerprint(snapshot.Events),
	}
	if snapshot.Logs != nil {
		fp.logs = marshalRealtimeFingerprint(snapshot.Logs)
	}
	return fp
}

func marshalRealtimeFingerprint(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(raw)
}

func runRealtimeSnapshotMessage(snapshot runRealtimeSnapshot) models.RunRealtimeMessage {
	msg := models.RunRealtimeMessage{
		Type:   models.RunRealtimeMessageTypeSnapshot,
		Run:    &snapshot.Run,
		Events: snapshot.Events,
		SentAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
	if snapshot.Logs != nil {
		msg.Logs = snapshot.Logs
	}
	return msg
}

func runRealtimeRunMessage(run models.Run) models.RunRealtimeMessage {
	return models.RunRealtimeMessage{
		Type:   models.RunRealtimeMessageTypeRun,
		Run:    &run,
		SentAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
}

func runRealtimeEventsMessage(events []models.FlowEvent) models.RunRealtimeMessage {
	return models.RunRealtimeMessage{
		Type:   models.RunRealtimeMessageTypeEvents,
		Events: events,
		SentAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
}

func runRealtimeLogsMessage(logs models.RunLogs) models.RunRealtimeMessage {
	return models.RunRealtimeMessage{
		Type:   models.RunRealtimeMessageTypeLogs,
		Logs:   &logs,
		SentAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
}

func buildRunWaitRealtimeMessages(previous runRealtimeSnapshot, current runRealtimeSnapshot) []models.RunRealtimeMessage {
	prevProjection := previous.Run.WaitProjection
	nextProjection := current.Run.WaitProjection
	if prevProjection == nil && nextProjection == nil {
		return nil
	}

	switch {
	case prevProjection == nil && nextProjection != nil:
		return buildRunWaitLifecycleMessages(nil, nextProjection, current.Events)
	case prevProjection != nil && nextProjection == nil:
		return []models.RunRealtimeMessage{runRealtimeWaitResolvedMessage(buildRunWaitResolution(*prevProjection, current.Events))}
	default:
		if !runWaitProjectionChanged(prevProjection, nextProjection) {
			return nil
		}
		if dominantWaitChanged(prevProjection, nextProjection) {
			messages := []models.RunRealtimeMessage{
				runRealtimeWaitResolvedMessage(buildRunWaitResolution(*prevProjection, current.Events)),
			}
			return append(messages, buildRunWaitLifecycleMessages(prevProjection, nextProjection, current.Events)...)
		}
		return buildRunWaitLifecycleMessages(prevProjection, nextProjection, current.Events)
	}
}

func runWaitProjectionChanged(left *models.RunWaitProjection, right *models.RunWaitProjection) bool {
	return marshalRealtimeFingerprint(left) != marshalRealtimeFingerprint(right)
}

func buildRunWaitLifecycleMessages(previous *models.RunWaitProjection, current *models.RunWaitProjection, events []models.FlowEvent) []models.RunRealtimeMessage {
	if current == nil {
		return nil
	}
	if current.DominantWait.ManualAction != nil && shouldEmitWaitManualAction(previous, current) {
		return []models.RunRealtimeMessage{
			runRealtimeWaitManualActionRequiredMessage(
				current.DominantWait.WaitID,
				*current.DominantWait.ManualAction,
				latestFlowEventCreatedAt(events),
			),
		}
	}
	if previous == nil || !waitProjectionContainsWaitID(previous, current.DominantWait.WaitID) {
		return []models.RunRealtimeMessage{runRealtimeWaitEnteredMessage(*current)}
	}
	return []models.RunRealtimeMessage{runRealtimeWaitUpdatedMessage(*current)}
}

func shouldEmitWaitManualAction(previous *models.RunWaitProjection, current *models.RunWaitProjection) bool {
	if current == nil || current.DominantWait.ManualAction == nil {
		return false
	}
	if previous == nil {
		return true
	}
	if dominantWaitChanged(previous, current) {
		return true
	}
	if previous.DominantWait.ManualAction == nil {
		return true
	}
	return strings.TrimSpace(previous.DominantWait.ManualAction.Kind) != strings.TrimSpace(current.DominantWait.ManualAction.Kind)
}

func dominantWaitChanged(previous *models.RunWaitProjection, current *models.RunWaitProjection) bool {
	if previous == nil || current == nil {
		return previous != current
	}
	return strings.TrimSpace(previous.DominantWait.WaitID) != strings.TrimSpace(current.DominantWait.WaitID)
}

func waitProjectionContainsWaitID(projection *models.RunWaitProjection, waitID string) bool {
	waitID = strings.TrimSpace(waitID)
	if projection == nil || waitID == "" {
		return false
	}
	if strings.TrimSpace(projection.DominantWait.WaitID) == waitID {
		return true
	}
	for _, related := range projection.RelatedWaits {
		if strings.TrimSpace(related.WaitID) == waitID {
			return true
		}
	}
	return false
}

func latestFlowEventCreatedAt(events []models.FlowEvent) string {
	if len(events) == 0 {
		return time.Now().UTC().Format(time.RFC3339Nano)
	}
	createdAt := strings.TrimSpace(events[0].CreatedAt)
	if createdAt == "" {
		return time.Now().UTC().Format(time.RFC3339Nano)
	}
	return createdAt
}

func buildRunWaitResolution(previous models.RunWaitProjection, events []models.FlowEvent) models.RunWaitResolution {
	resolution := models.RunWaitResolution{
		WaitID:         previous.DominantWait.WaitID,
		ContourKind:    previous.DominantWait.ContourKind,
		ResolutionKind: defaultWaitResolutionKind,
		ResolvedAt:     time.Now().UTC().Format(time.RFC3339Nano),
	}
	previousWaitID := strings.TrimSpace(previous.DominantWait.WaitID)
	var fallback runWaitResolvedEventPayload
	var fallbackResolvedAt string
	var hasFallback bool
	for _, item := range events {
		if !strings.EqualFold(strings.TrimSpace(item.EventType), runWaitResumedFlowEventType) {
			continue
		}
		var payload runWaitResolvedEventPayload
		if err := json.Unmarshal([]byte(item.PayloadJSON), &payload); err != nil {
			continue
		}
		waitID := strings.TrimSpace(payload.WaitID)
		if waitID == "" {
			waitID = resolution.WaitID
		}
		resolvedAt := strings.TrimSpace(item.CreatedAt)
		if previousWaitID != "" && waitID != previousWaitID {
			if !hasFallback {
				payload.WaitID = waitID
				fallback = payload
				fallbackResolvedAt = resolvedAt
				hasFallback = true
			}
			continue
		}
		applyRunWaitResolutionPayload(&resolution, payload, resolvedAt)
		return resolution
	}
	if hasFallback {
		applyRunWaitResolutionPayload(&resolution, fallback, fallbackResolvedAt)
	}
	if strings.TrimSpace(resolution.ContourKind) == "" {
		resolution.ContourKind = defaultWaitContourKind
	}
	return resolution
}

func applyRunWaitResolutionPayload(resolution *models.RunWaitResolution, payload runWaitResolvedEventPayload, resolvedAt string) {
	if resolution == nil {
		return
	}
	if waitID := strings.TrimSpace(payload.WaitID); waitID != "" {
		resolution.WaitID = waitID
	}
	if contourKind := strings.TrimSpace(payload.ContourKind); contourKind != "" {
		resolution.ContourKind = contourKind
	}
	if resolutionKind := strings.TrimSpace(payload.ResolutionKind); resolutionKind != "" {
		resolution.ResolutionKind = resolutionKind
	}
	if resolvedAt = strings.TrimSpace(resolvedAt); resolvedAt != "" {
		resolution.ResolvedAt = resolvedAt
	}
}

func runRealtimeWaitEnteredMessage(waitProjection models.RunWaitProjection) models.RunRealtimeMessage {
	return models.RunRealtimeMessage{
		Type:           models.RunRealtimeMessageTypeWaitEntered,
		WaitProjection: &waitProjection,
		SentAt:         time.Now().UTC().Format(time.RFC3339Nano),
	}
}

func runRealtimeWaitUpdatedMessage(waitProjection models.RunWaitProjection) models.RunRealtimeMessage {
	return models.RunRealtimeMessage{
		Type:           models.RunRealtimeMessageTypeWaitUpdated,
		WaitProjection: &waitProjection,
		SentAt:         time.Now().UTC().Format(time.RFC3339Nano),
	}
}

func runRealtimeWaitResolvedMessage(resolution models.RunWaitResolution) models.RunRealtimeMessage {
	return models.RunRealtimeMessage{
		Type:           models.RunRealtimeMessageTypeWaitResolved,
		WaitResolution: &resolution,
		SentAt:         time.Now().UTC().Format(time.RFC3339Nano),
	}
}

func runRealtimeWaitManualActionRequiredMessage(waitID string, manualAction models.GitHubRateLimitManualAction, updatedAt string) models.RunRealtimeMessage {
	return models.RunRealtimeMessage{
		Type: models.RunRealtimeMessageTypeWaitManualActionRequired,
		WaitManualAction: &models.RunWaitManualActionEvent{
			WaitID:       strings.TrimSpace(waitID),
			ManualAction: manualAction,
			UpdatedAt:    strings.TrimSpace(updatedAt),
		},
		SentAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
}

func runRealtimeErrorMessage(err error) models.RunRealtimeMessage {
	text := strings.TrimSpace(err.Error())
	if text == "" {
		text = "realtime stream fetch failed"
	}
	return models.RunRealtimeMessage{
		Type:    models.RunRealtimeMessageTypeError,
		Message: &text,
		SentAt:  time.Now().UTC().Format(time.RFC3339Nano),
	}
}
