package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v5"

	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/realtime"
)

const (
	realtimeWSReadLimit       = 1 << 20
	realtimeWSPingInterval    = 20 * time.Second
	realtimeWSWriteTimeout    = 10 * time.Second
	realtimeWSPongWait        = 45 * time.Second
	realtimeCatchupBatchLimit = 400
)

type realtimeHandler struct {
	backplane *realtime.Backplane
	repo      *realtime.Repository
	upgrader  websocket.Upgrader
}

type realtimeSession struct {
	principal     *controlplanev1.Principal
	userID        string
	isPlatformOps bool

	mu         sync.RWMutex
	filter     realtime.SubscriptionFilter
	lastEvent  int64
	projectACL map[string]bool
}

type realtimeWSInbound struct {
	Type        string   `json:"type"`
	LastEventID *int64   `json:"last_event_id,omitempty"`
	Topics      []string `json:"topics,omitempty"`
	ProjectID   string   `json:"project_id,omitempty"`
	RunID       string   `json:"run_id,omitempty"`
	TaskID      string   `json:"task_id,omitempty"`
}

type realtimeWSOutbound struct {
	Type  string           `json:"type"`
	Event *realtimeWSEvent `json:"event,omitempty"`
	Meta  *realtimeWSMeta  `json:"meta,omitempty"`
	Error *realtimeWSError `json:"error,omitempty"`
}

type realtimeWSEvent struct {
	ID            int64           `json:"id"`
	Topic         string          `json:"topic"`
	Scope         json.RawMessage `json:"scope"`
	Payload       json.RawMessage `json:"payload"`
	CorrelationID string          `json:"correlation_id"`
	ProjectID     string          `json:"project_id"`
	RunID         string          `json:"run_id"`
	TaskID        string          `json:"task_id"`
	CreatedAt     string          `json:"created_at"`
}

type realtimeWSMeta struct {
	Status      string `json:"status"`
	LastEventID int64  `json:"last_event_id"`
}

type realtimeWSError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func newRealtimeHandler(backplane *realtime.Backplane, repo *realtime.Repository) *realtimeHandler {
	return &realtimeHandler{
		backplane: backplane,
		repo:      repo,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(_ *http.Request) bool {
				return true
			},
		},
	}
}

func (h *realtimeHandler) Connect(c *echo.Context) error {
	if h == nil || h.backplane == nil || h.repo == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"code":    "unavailable",
			"message": "realtime backplane is unavailable",
		})
	}

	principal, err := requirePrincipal(c)
	if err != nil {
		return err
	}

	conn, err := h.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return fmt.Errorf("upgrade realtime websocket: %w", err)
	}
	defer func() { _ = conn.Close() }()

	filter, lastEventID := parseRealtimeQuery(c)
	session := &realtimeSession{
		principal:     principal,
		userID:        strings.TrimSpace(principal.GetUserId()),
		isPlatformOps: principal.GetIsPlatformAdmin() || principal.GetIsPlatformOwner(),
		filter:        filter.Normalize(),
		lastEvent:     lastEventID,
		projectACL:    make(map[string]bool),
	}

	subCh, unsubscribe := h.backplane.Subscribe(128)
	defer unsubscribe()

	if err := h.configureConn(conn); err != nil {
		return err
	}

	if err := h.writeJSON(conn, realtimeWSOutbound{
		Type: "hello",
		Meta: &realtimeWSMeta{
			Status:      "connected",
			LastEventID: session.lastEventID(),
		},
	}); err != nil {
		return err
	}

	if err := h.replay(conn, c.Request().Context(), session, session.lastEventID()); err != nil {
		return err
	}

	writeDone := make(chan struct{})
	go func() {
		defer close(writeDone)
		ticker := time.NewTicker(realtimeWSPingInterval)
		defer ticker.Stop()
		for {
			select {
			case <-c.Request().Context().Done():
				return
			case <-ticker.C:
				if err := conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(realtimeWSWriteTimeout)); err != nil {
					return
				}
			case event, ok := <-subCh:
				if !ok {
					return
				}
				allowed, allowedErr := h.isAllowedEvent(c.Request().Context(), session, event)
				if allowedErr != nil || !allowed {
					continue
				}
				if !session.currentFilter().Matches(event) {
					continue
				}
				if err := h.writeEvent(conn, event); err != nil {
					return
				}
				session.setLastEventID(event.ID)
			}
		}
	}()

	for {
		select {
		case <-writeDone:
			return nil
		default:
		}

		var inbound realtimeWSInbound
		if err := conn.ReadJSON(&inbound); err != nil {
			return nil
		}
		switch strings.TrimSpace(strings.ToLower(inbound.Type)) {
		case "", "ack":
			if inbound.LastEventID != nil && *inbound.LastEventID > 0 {
				session.setLastEventID(*inbound.LastEventID)
			}
		case "subscribe":
			next := realtime.SubscriptionFilter{
				Topics:    inbound.Topics,
				ProjectID: inbound.ProjectID,
				RunID:     inbound.RunID,
				TaskID:    inbound.TaskID,
			}.Normalize()
			session.setFilter(next)
			if inbound.LastEventID != nil && *inbound.LastEventID > 0 {
				session.setLastEventID(*inbound.LastEventID)
			}
			if err := h.replay(conn, c.Request().Context(), session, session.lastEventID()); err != nil {
				return err
			}
			if err := h.writeJSON(conn, realtimeWSOutbound{
				Type: "subscribed",
				Meta: &realtimeWSMeta{
					Status:      "ok",
					LastEventID: session.lastEventID(),
				},
			}); err != nil {
				return err
			}
		default:
			if err := h.writeJSON(conn, realtimeWSOutbound{
				Type: "error",
				Error: &realtimeWSError{
					Code:    "invalid_argument",
					Message: "unsupported message type",
				},
			}); err != nil {
				return err
			}
		}
	}
}

func parseRealtimeQuery(c *echo.Context) (realtime.SubscriptionFilter, int64) {
	topicsRaw := strings.TrimSpace(c.QueryParam("topics"))
	topics := make([]string, 0)
	if topicsRaw != "" {
		for _, token := range strings.Split(topicsRaw, ",") {
			topic := strings.TrimSpace(token)
			if topic != "" {
				topics = append(topics, topic)
			}
		}
	}
	var lastEventID int64
	if raw := strings.TrimSpace(c.QueryParam("last_event_id")); raw != "" {
		if parsed, err := strconv.ParseInt(raw, 10, 64); err == nil && parsed > 0 {
			lastEventID = parsed
		}
	}
	return realtime.SubscriptionFilter{
		Topics:    topics,
		ProjectID: strings.TrimSpace(c.QueryParam("project_id")),
		RunID:     strings.TrimSpace(c.QueryParam("run_id")),
		TaskID:    strings.TrimSpace(c.QueryParam("task_id")),
	}, lastEventID
}

func (h *realtimeHandler) configureConn(conn *websocket.Conn) error {
	if conn == nil {
		return fmt.Errorf("realtime websocket connection is nil")
	}
	conn.SetReadLimit(realtimeWSReadLimit)
	if err := conn.SetReadDeadline(time.Now().Add(realtimeWSPongWait)); err != nil {
		return err
	}
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(realtimeWSPongWait))
	})
	return nil
}

func (h *realtimeHandler) replay(conn *websocket.Conn, ctx context.Context, session *realtimeSession, afterID int64) error {
	cursor := afterID
	for {
		items, err := h.repo.ListAfterID(ctx, cursor, realtimeCatchupBatchLimit)
		if err != nil {
			return fmt.Errorf("realtime replay list after id=%d: %w", cursor, err)
		}
		if len(items) == 0 {
			return nil
		}
		for _, event := range items {
			allowed, allowedErr := h.isAllowedEvent(ctx, session, event)
			if allowedErr != nil || !allowed {
				cursor = event.ID
				continue
			}
			if !session.currentFilter().Matches(event) {
				cursor = event.ID
				continue
			}
			if err := h.writeEvent(conn, event); err != nil {
				return err
			}
			cursor = event.ID
			session.setLastEventID(event.ID)
		}
		if len(items) < realtimeCatchupBatchLimit {
			return nil
		}
	}
}

func (h *realtimeHandler) isAllowedEvent(ctx context.Context, session *realtimeSession, event realtime.Event) (bool, error) {
	if session == nil {
		return false, nil
	}
	if session.isPlatformOps {
		return true, nil
	}
	projectID := strings.TrimSpace(event.ProjectID)
	if projectID == "" {
		return false, nil
	}
	session.mu.RLock()
	allowed, ok := session.projectACL[projectID]
	session.mu.RUnlock()
	if ok {
		return allowed, nil
	}
	allowed, err := h.repo.UserHasProjectAccess(ctx, projectID, session.userID)
	if err != nil {
		return false, err
	}
	session.mu.Lock()
	session.projectACL[projectID] = allowed
	session.mu.Unlock()
	return allowed, nil
}

func (h *realtimeHandler) writeEvent(conn *websocket.Conn, event realtime.Event) error {
	return h.writeJSON(conn, realtimeWSOutbound{
		Type: "event",
		Event: &realtimeWSEvent{
			ID:            event.ID,
			Topic:         event.Topic,
			Scope:         event.ScopeJSON,
			Payload:       event.PayloadJSON,
			CorrelationID: event.CorrelationID,
			ProjectID:     event.ProjectID,
			RunID:         event.RunID,
			TaskID:        event.TaskID,
			CreatedAt:     event.CreatedAt.UTC().Format(time.RFC3339Nano),
		},
	})
}

func (h *realtimeHandler) writeJSON(conn *websocket.Conn, payload realtimeWSOutbound) error {
	if conn == nil {
		return fmt.Errorf("realtime websocket connection is nil")
	}
	if err := conn.SetWriteDeadline(time.Now().Add(realtimeWSWriteTimeout)); err != nil {
		return err
	}
	if err := conn.WriteJSON(payload); err != nil {
		return fmt.Errorf("write realtime websocket payload: %w", err)
	}
	return nil
}

func (s *realtimeSession) currentFilter() realtime.SubscriptionFilter {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.filter
}

func (s *realtimeSession) setFilter(filter realtime.SubscriptionFilter) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.filter = filter
}

func (s *realtimeSession) lastEventID() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastEvent
}

func (s *realtimeSession) setLastEventID(value int64) {
	if value <= 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if value > s.lastEvent {
		s.lastEvent = value
	}
}
