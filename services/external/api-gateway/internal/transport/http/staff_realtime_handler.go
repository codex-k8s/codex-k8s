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
	runRealtimeFetchTimeout       = 8 * time.Second
	runRealtimeWriteTimeout       = 10 * time.Second
	runRealtimePongTimeout        = 60 * time.Second
	runRealtimeUpdateInterval     = 2 * time.Second
	runRealtimePingInterval       = 25 * time.Second
)

var runRealtimeUpgrader = websocket.Upgrader{
	HandshakeTimeout: 10 * time.Second,
	ReadBufferSize:   4 * 1024,
	WriteBufferSize:  4 * 1024,
	CheckOrigin:      allowRunRealtimeOrigin,
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

func allowRunRealtimeOrigin(r *http.Request) bool {
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
	conn, err := runRealtimeUpgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		// Upgrader already wrote response details on error.
		return nil
	}
	defer conn.Close()

	conn.SetReadLimit(64 * 1024)
	_ = conn.SetReadDeadline(time.Now().Add(runRealtimePongTimeout))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(runRealtimePongTimeout))
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

	fetchCtx, cancelFetch := context.WithTimeout(c.Request().Context(), runRealtimeFetchTimeout)
	initialSnapshot, err := h.fetchRunRealtimeSnapshot(fetchCtx, principal, arg)
	cancelFetch()
	if err != nil {
		_ = writeRunRealtimeMessage(conn, runRealtimeErrorMessage(err))
		sendRunRealtimeClose(conn)
		return nil
	}

	if err := writeRunRealtimeMessage(conn, runRealtimeSnapshotMessage(initialSnapshot)); err != nil {
		return nil
	}

	fingerprint := buildRunRealtimeFingerprint(initialSnapshot)

	updateTicker := time.NewTicker(runRealtimeUpdateInterval)
	defer updateTicker.Stop()
	pingTicker := time.NewTicker(runRealtimePingInterval)
	defer pingTicker.Stop()

	for {
		select {
		case <-c.Request().Context().Done():
			sendRunRealtimeClose(conn)
			return nil
		case <-readerDone:
			return nil
		case <-pingTicker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(runRealtimeWriteTimeout))
			if pingErr := conn.WriteMessage(websocket.PingMessage, nil); pingErr != nil {
				return nil
			}
		case <-updateTicker.C:
			nextFetchCtx, nextCancel := context.WithTimeout(c.Request().Context(), runRealtimeFetchTimeout)
			nextSnapshot, fetchErr := h.fetchRunRealtimeSnapshot(nextFetchCtx, principal, arg)
			nextCancel()
			if fetchErr != nil {
				_ = writeRunRealtimeMessage(conn, runRealtimeErrorMessage(fetchErr))
				continue
			}

			nextFingerprint := buildRunRealtimeFingerprint(nextSnapshot)
			if nextFingerprint.run != fingerprint.run {
				if writeErr := writeRunRealtimeMessage(conn, runRealtimeRunMessage(nextSnapshot.Run)); writeErr != nil {
					return nil
				}
			}
			if nextFingerprint.events != fingerprint.events {
				if writeErr := writeRunRealtimeMessage(conn, runRealtimeEventsMessage(nextSnapshot.Events)); writeErr != nil {
					return nil
				}
			}
			if nextSnapshot.Logs != nil && nextFingerprint.logs != fingerprint.logs {
				if writeErr := writeRunRealtimeMessage(conn, runRealtimeLogsMessage(*nextSnapshot.Logs)); writeErr != nil {
					return nil
				}
			}

			fingerprint = nextFingerprint
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

func writeRunRealtimeMessage(conn *websocket.Conn, message models.RunRealtimeMessage) error {
	_ = conn.SetWriteDeadline(time.Now().Add(runRealtimeWriteTimeout))
	return conn.WriteJSON(message)
}

func sendRunRealtimeClose(conn *websocket.Conn) {
	_ = conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "closing"),
		time.Now().Add(runRealtimeWriteTimeout),
	)
}

func buildRunRealtimeFingerprint(snapshot runRealtimeSnapshot) runRealtimeFingerprint {
	fp := runRealtimeFingerprint{
		run:    marshalRunRealtimeFingerprint(snapshot.Run),
		events: marshalRunRealtimeFingerprint(snapshot.Events),
	}
	if snapshot.Logs != nil {
		fp.logs = marshalRunRealtimeFingerprint(snapshot.Logs)
	}
	return fp
}

func marshalRunRealtimeFingerprint(value any) string {
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
