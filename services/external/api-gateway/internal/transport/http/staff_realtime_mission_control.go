package http

import (
	"context"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v5"

	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/generated"
)

const (
	missionControlRealtimeFresh    = "fresh"
	missionControlRealtimeStale    = "stale"
	missionControlRealtimeDegraded = "degraded"
)

func (h *staffHandler) MissionControlRealtime(c *echo.Context) error {
	return withPrincipalAndResolved(c, resolveMissionControlRealtimeArg, func(principal *controlplanev1.Principal, token missionControlResumeTokenPayload) error {
		return h.streamMissionControlRealtime(c, principal, token)
	})
}

func (h *staffHandler) streamMissionControlRealtime(
	c *echo.Context,
	principal *controlplanev1.Principal,
	token missionControlResumeTokenPayload,
) error {
	conn, err := staffRealtimeUpgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
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

	arg := missionControlWorkspaceArgFromResumeToken(token)
	fetchCtx, cancelFetch := context.WithTimeout(c.Request().Context(), realtimeFetchTimeout)
	currentSnapshot, currentResumeToken, err := h.fetchMissionControlWorkspaceSnapshot(fetchCtx, principal, arg)
	cancelFetch()
	if err != nil {
		_ = h.writeMissionControlRealtimeEnvelope(
			conn,
			missionControlErrorRealtimeEnvelope(token.SnapshotID, "", "snapshot_fetch_failed", "unable to load mission control snapshot", true),
		)
		sendRealtimeClose(conn)
		return nil
	}

	if currentSnapshot.GetSnapshotId() != token.SnapshotID {
		_ = h.writeMissionControlRealtimeEnvelope(
			conn,
			missionControlResyncRequiredRealtimeEnvelope(currentSnapshot.GetSnapshotId(), currentResumeToken, "resume_token_outdated", 1),
		)
		sendRealtimeClose(conn)
		return nil
	}

	if err := h.writeMissionControlRealtimeEnvelope(
		conn,
		missionControlConnectedRealtimeEnvelope(
			currentSnapshot.GetSnapshotId(),
			currentResumeToken,
			missionControlWorkspaceFreshnessStatus(currentSnapshot),
		),
	); err != nil {
		return nil
	}

	currentFreshness := missionControlFreshnessStatus(missionControlWorkspaceFreshnessStatus(currentSnapshot))

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
			if err := h.writeMissionControlRealtimeEnvelope(
				conn,
				missionControlHeartbeatRealtimeEnvelope(currentSnapshot.GetSnapshotId(), currentResumeToken),
			); err != nil {
				return nil
			}
		case <-updateTicker.C:
			nextFetchCtx, nextCancel := context.WithTimeout(c.Request().Context(), realtimeFetchTimeout)
			nextSnapshot, nextResumeToken, fetchErr := h.fetchMissionControlWorkspaceSnapshot(nextFetchCtx, principal, arg)
			nextCancel()
			if fetchErr != nil {
				_ = h.writeMissionControlRealtimeEnvelope(
					conn,
					missionControlErrorRealtimeEnvelope(currentSnapshot.GetSnapshotId(), currentResumeToken, "snapshot_fetch_failed", "unable to refresh mission control snapshot", true),
				)
				continue
			}

			if nextSnapshot.GetSnapshotId() != currentSnapshot.GetSnapshotId() {
				if err := h.writeMissionControlRealtimeEnvelope(
					conn,
					missionControlInvalidateRealtimeEnvelope(nextSnapshot, nextResumeToken, "snapshot_changed", "workspace_snapshot"),
				); err != nil {
					return nil
				}
				currentSnapshot = nextSnapshot
				currentResumeToken = nextResumeToken
				currentFreshness = missionControlFreshnessStatus(missionControlWorkspaceFreshnessStatus(nextSnapshot))
				continue
			}

			nextFreshness := missionControlFreshnessStatus(missionControlWorkspaceFreshnessStatus(nextSnapshot))
			if nextFreshness == currentFreshness {
				continue
			}

			switch nextFreshness {
			case missionControlRealtimeStale:
				if err := h.writeMissionControlRealtimeEnvelope(
					conn,
					missionControlStaleRealtimeEnvelope(nextSnapshot, nextResumeToken, "snapshot_stale", "refresh_workspace"),
				); err != nil {
					return nil
				}
			case missionControlRealtimeDegraded:
				if err := h.writeMissionControlRealtimeEnvelope(
					conn,
					missionControlDegradedRealtimeEnvelope(nextSnapshot, nextResumeToken, "snapshot_degraded", "explicit_refresh", []string{"realtime_delta"}),
				); err != nil {
					return nil
				}
			case missionControlRealtimeFresh:
				if err := h.writeMissionControlRealtimeEnvelope(
					conn,
					missionControlInvalidateRealtimeEnvelope(nextSnapshot, nextResumeToken, "freshness_recovered", "workspace_snapshot"),
				); err != nil {
					return nil
				}
			}

			currentSnapshot = nextSnapshot
			currentResumeToken = nextResumeToken
			currentFreshness = nextFreshness
		}
	}
}

func (h *staffHandler) writeMissionControlRealtimeEnvelope(
	conn *websocket.Conn,
	build func() (generated.MissionControlRealtimeEnvelope, error),
) error {
	envelope, err := build()
	if err != nil {
		return fmt.Errorf("build mission control realtime envelope: %w", err)
	}
	return writeRealtimeJSONMessage(conn, envelope)
}

func missionControlConnectedRealtimeEnvelope(
	snapshotID string,
	resumeToken string,
	freshnessStatus string,
) func() (generated.MissionControlRealtimeEnvelope, error) {
	return missionControlRealtimeEnvelopeBuilder(
		generated.MissionControlRealtimeEnvelopeEventKindConnected,
		snapshotID,
		resumeToken,
		func(payload *generated.MissionControlRealtimePayload) error {
			return payload.FromMissionControlConnectedRealtimePayload(generated.MissionControlConnectedRealtimePayload{
				ServerCursor:            snapshotID,
				SnapshotFreshnessStatus: generated.MissionControlConnectedRealtimePayloadSnapshotFreshnessStatus(missionControlFreshnessStatus(freshnessStatus)),
			})
		},
	)
}

func missionControlInvalidateRealtimeEnvelope(
	snapshot *controlplanev1.MissionControlWorkspaceSnapshot,
	resumeToken string,
	reason string,
	refreshScope string,
) func() (generated.MissionControlRealtimeEnvelope, error) {
	return missionControlRealtimeEnvelopeBuilder(
		generated.MissionControlRealtimeEnvelopeEventKindInvalidate,
		snapshot.GetSnapshotId(),
		resumeToken,
		func(payload *generated.MissionControlRealtimePayload) error {
			return payload.FromMissionControlInvalidateRealtimePayload(generated.MissionControlInvalidateRealtimePayload{
				AffectedEntityRefs: missionControlRealtimeEntityRefs(snapshot),
				Reason:             reason,
				RefreshScope:       refreshScope,
			})
		},
	)
}

func missionControlStaleRealtimeEnvelope(
	snapshot *controlplanev1.MissionControlWorkspaceSnapshot,
	resumeToken string,
	reason string,
	suggestedRefresh string,
) func() (generated.MissionControlRealtimeEnvelope, error) {
	return missionControlRealtimeEnvelopeBuilder(
		generated.MissionControlRealtimeEnvelopeEventKindStale,
		snapshot.GetSnapshotId(),
		resumeToken,
		func(payload *generated.MissionControlRealtimePayload) error {
			return payload.FromMissionControlStaleRealtimePayload(generated.MissionControlStaleRealtimePayload{
				Reason:           reason,
				StaleSince:       missionControlStaleSince(snapshot),
				SuggestedRefresh: suggestedRefresh,
			})
		},
	)
}

func missionControlDegradedRealtimeEnvelope(
	snapshot *controlplanev1.MissionControlWorkspaceSnapshot,
	resumeToken string,
	reason string,
	fallbackMode string,
	affectedCapabilities []string,
) func() (generated.MissionControlRealtimeEnvelope, error) {
	return missionControlRealtimeEnvelopeBuilder(
		generated.MissionControlRealtimeEnvelopeEventKindDegraded,
		snapshot.GetSnapshotId(),
		resumeToken,
		func(payload *generated.MissionControlRealtimePayload) error {
			return payload.FromMissionControlDegradedRealtimePayload(generated.MissionControlDegradedRealtimePayload{
				AffectedCapabilities: append([]string{}, affectedCapabilities...),
				FallbackMode:         fallbackMode,
				Reason:               reason,
			})
		},
	)
}

func missionControlResyncRequiredRealtimeEnvelope(
	requiredSnapshotID string,
	resumeToken string,
	reason string,
	droppedEventCount int32,
) func() (generated.MissionControlRealtimeEnvelope, error) {
	return missionControlRealtimeEnvelopeBuilder(
		generated.MissionControlRealtimeEnvelopeEventKindResyncRequired,
		requiredSnapshotID,
		resumeToken,
		func(payload *generated.MissionControlRealtimePayload) error {
			return payload.FromMissionControlResyncRequiredRealtimePayload(generated.MissionControlResyncRequiredRealtimePayload{
				DroppedEventCount:  droppedEventCount,
				Reason:             reason,
				RequiredSnapshotId: requiredSnapshotID,
			})
		},
	)
}

func missionControlHeartbeatRealtimeEnvelope(
	snapshotID string,
	resumeToken string,
) func() (generated.MissionControlRealtimeEnvelope, error) {
	return missionControlRealtimeEnvelopeBuilder(
		generated.MissionControlRealtimeEnvelopeEventKindHeartbeat,
		snapshotID,
		resumeToken,
		func(payload *generated.MissionControlRealtimePayload) error {
			return payload.FromMissionControlHeartbeatRealtimePayload(generated.MissionControlHeartbeatRealtimePayload{
				ServerTime: time.Now().UTC(),
				SnapshotId: snapshotID,
			})
		},
	)
}

func missionControlErrorRealtimeEnvelope(
	snapshotID string,
	resumeToken string,
	code string,
	message string,
	retryable bool,
) func() (generated.MissionControlRealtimeEnvelope, error) {
	return missionControlRealtimeEnvelopeBuilder(
		generated.MissionControlRealtimeEnvelopeEventKindError,
		snapshotID,
		resumeToken,
		func(payload *generated.MissionControlRealtimePayload) error {
			return payload.FromMissionControlErrorRealtimePayload(generated.MissionControlErrorRealtimePayload{
				Code:      code,
				Message:   message,
				Retryable: retryable,
			})
		},
	)
}

func missionControlRealtimeEnvelopeBuilder(
	eventKind generated.MissionControlRealtimeEnvelopeEventKind,
	snapshotID string,
	resumeToken string,
	populate func(payload *generated.MissionControlRealtimePayload) error,
) func() (generated.MissionControlRealtimeEnvelope, error) {
	return func() (generated.MissionControlRealtimeEnvelope, error) {
		payload := generated.MissionControlRealtimePayload{}
		if err := populate(&payload); err != nil {
			return generated.MissionControlRealtimeEnvelope{}, err
		}
		return missionControlRealtimeEnvelope(eventKind, snapshotID, resumeToken, payload), nil
	}
}

func missionControlRealtimeEnvelope(
	eventKind generated.MissionControlRealtimeEnvelopeEventKind,
	snapshotID string,
	resumeToken string,
	payload generated.MissionControlRealtimePayload,
) generated.MissionControlRealtimeEnvelope {
	return generated.MissionControlRealtimeEnvelope{
		EventKind:   eventKind,
		SnapshotId:  snapshotID,
		ResumeToken: resumeToken,
		OccurredAt:  time.Now().UTC(),
		Payload:     payload,
	}
}

func missionControlRealtimeEntityRefs(snapshot *controlplanev1.MissionControlWorkspaceSnapshot) []generated.MissionControlEntityRef {
	nodes := snapshot.GetNodes()
	out := make([]generated.MissionControlEntityRef, 0, len(nodes))
	for _, node := range nodes {
		if node == nil {
			continue
		}
		switch node.GetNodeKind() {
		case "discussion", "pull_request", "work_item":
		default:
			continue
		}
		out = append(out, generated.MissionControlEntityRef{
			EntityKind:     generated.MissionControlEntityRefEntityKind(node.GetNodeKind()),
			EntityPublicId: node.GetNodePublicId(),
		})
	}
	return out
}

func missionControlFreshnessStatus(value string) string {
	switch value {
	case missionControlRealtimeStale, missionControlRealtimeDegraded:
		return value
	default:
		return missionControlRealtimeFresh
	}
}

func missionControlWorkspaceFreshnessStatus(snapshot *controlplanev1.MissionControlWorkspaceSnapshot) string {
	if snapshot == nil {
		return missionControlRealtimeFresh
	}

	status := missionControlRealtimeFresh
	for _, watermark := range snapshot.GetWorkspaceWatermarks() {
		if watermark == nil {
			continue
		}
		switch watermark.GetStatus() {
		case missionControlRealtimeDegraded:
			return missionControlRealtimeDegraded
		case missionControlRealtimeStale:
			status = missionControlRealtimeStale
		}
	}
	return status
}

func missionControlStaleSince(snapshot *controlplanev1.MissionControlWorkspaceSnapshot) time.Time {
	if snapshot == nil {
		return time.Now().UTC()
	}

	latest := time.Time{}
	for _, watermark := range snapshot.GetWorkspaceWatermarks() {
		if watermark == nil {
			continue
		}
		status := watermark.GetStatus()
		if status != missionControlRealtimeStale && status != missionControlRealtimeDegraded {
			continue
		}

		candidate := time.Time{}
		if watermark.GetWindowEndedAt() != nil {
			candidate = watermark.GetWindowEndedAt().AsTime().UTC()
		} else if watermark.GetObservedAt() != nil {
			candidate = watermark.GetObservedAt().AsTime().UTC()
		}
		if candidate.After(latest) {
			latest = candidate
		}
	}
	if !latest.IsZero() {
		return latest
	}
	return time.Now().UTC()
}
