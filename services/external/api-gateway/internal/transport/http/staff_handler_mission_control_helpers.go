package http

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
)

const (
	missionControlDefaultLimit = 50
)

type missionControlDashboardArg struct {
	viewMode     string
	activeFilter string
	search       string
	cursor       string
	limit        int32
}

type missionControlEntityArg struct {
	entityKind     string
	entityPublicID string
	timelineLimit  int32
}

type missionControlTimelineArg struct {
	entityKind     string
	entityPublicID string
	cursor         string
	limit          int32
}

type missionControlResumeTokenPayload struct {
	SnapshotID   string `json:"snapshot_id"`
	ViewMode     string `json:"view_mode"`
	ActiveFilter string `json:"active_filter"`
	Search       string `json:"search,omitempty"`
	Cursor       string `json:"cursor,omitempty"`
	Limit        int32  `json:"limit"`
	IssuedAt     string `json:"issued_at"`
}

func resolveMissionControlDashboardArg(c *echo.Context) (missionControlDashboardArg, error) {
	limit, err := parseLimit(c, missionControlDefaultLimit)
	if err != nil {
		return missionControlDashboardArg{}, err
	}
	return missionControlDashboardArg{
		viewMode:     strings.TrimSpace(c.QueryParam("view_mode")),
		activeFilter: strings.TrimSpace(c.QueryParam("active_filter")),
		search:       strings.TrimSpace(c.QueryParam("search")),
		cursor:       strings.TrimSpace(c.QueryParam("cursor")),
		limit:        int32(limit),
	}, nil
}

func resolveMissionControlEntityArg(c *echo.Context) (missionControlEntityArg, error) {
	entityKind, err := requirePathParam(c, "entity_kind")
	if err != nil {
		return missionControlEntityArg{}, err
	}
	entityPublicID, err := resolvePathUnescaped("entity_public_id")(c)
	if err != nil {
		return missionControlEntityArg{}, err
	}
	limit, err := parseLimit(c, missionControlDefaultLimit)
	if err != nil {
		return missionControlEntityArg{}, err
	}
	return missionControlEntityArg{
		entityKind:     entityKind,
		entityPublicID: entityPublicID,
		timelineLimit:  int32(limit),
	}, nil
}

func resolveMissionControlTimelineArg(c *echo.Context) (missionControlTimelineArg, error) {
	entityArg, err := resolveMissionControlEntityArg(c)
	if err != nil {
		return missionControlTimelineArg{}, err
	}
	return missionControlTimelineArg{
		entityKind:     entityArg.entityKind,
		entityPublicID: entityArg.entityPublicID,
		cursor:         strings.TrimSpace(c.QueryParam("cursor")),
		limit:          entityArg.timelineLimit,
	}, nil
}

func resolveMissionControlRealtimeArg(c *echo.Context) (missionControlResumeTokenPayload, error) {
	return decodeMissionControlResumeToken(c.QueryParam("resume_token"))
}

func missionControlCorrelationID(c *echo.Context) string {
	if c == nil {
		return uuid.NewString()
	}
	if value := strings.TrimSpace(c.Request().Header.Get("X-Correlation-ID")); value != "" {
		return value
	}
	return uuid.NewString()
}

func setMissionControlCorrelationHeader(c *echo.Context, correlationID string) {
	if c == nil {
		return
	}
	correlationID = strings.TrimSpace(correlationID)
	if correlationID == "" {
		return
	}
	c.Response().Header().Set("X-Correlation-ID", correlationID)
}

func encodeMissionControlResumeToken(payload missionControlResumeTokenPayload) (string, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func decodeMissionControlResumeToken(raw string) (missionControlResumeTokenPayload, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return missionControlResumeTokenPayload{}, errs.Validation{Field: "resume_token", Msg: "is required"}
	}
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return missionControlResumeTokenPayload{}, errs.Validation{Field: "resume_token", Msg: "must be a valid opaque token"}
	}
	var payload missionControlResumeTokenPayload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return missionControlResumeTokenPayload{}, errs.Validation{Field: "resume_token", Msg: "must be a valid opaque token"}
	}
	if strings.TrimSpace(payload.SnapshotID) == "" {
		return missionControlResumeTokenPayload{}, errs.Validation{Field: "resume_token", Msg: "snapshot scope is missing"}
	}
	if payload.Limit <= 0 {
		payload.Limit = missionControlDefaultLimit
	}
	return payload, nil
}

func newMissionControlResumeTokenPayload(
	arg missionControlDashboardArg,
	snapshotID string,
) missionControlResumeTokenPayload {
	return missionControlResumeTokenPayload{
		SnapshotID:   strings.TrimSpace(snapshotID),
		ViewMode:     strings.TrimSpace(arg.viewMode),
		ActiveFilter: strings.TrimSpace(arg.activeFilter),
		Search:       strings.TrimSpace(arg.search),
		Cursor:       strings.TrimSpace(arg.cursor),
		Limit:        arg.limit,
		IssuedAt:     time.Now().UTC().Format(time.RFC3339Nano),
	}
}

func missionControlDashboardArgFromResumeToken(token missionControlResumeTokenPayload) missionControlDashboardArg {
	limit := token.Limit
	if limit <= 0 {
		limit = missionControlDefaultLimit
	}
	return missionControlDashboardArg{
		viewMode:     strings.TrimSpace(token.ViewMode),
		activeFilter: strings.TrimSpace(token.ActiveFilter),
		search:       strings.TrimSpace(token.Search),
		cursor:       strings.TrimSpace(token.Cursor),
		limit:        limit,
	}
}
