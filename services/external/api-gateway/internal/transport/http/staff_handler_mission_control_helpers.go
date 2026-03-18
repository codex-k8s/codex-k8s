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

type missionControlWorkspaceArg struct {
	viewMode    string
	statePreset string
	search      string
	cursor      string
	rootLimit   int32
}

type missionControlNodeArg struct {
	nodeKind     string
	nodePublicID string
}

type missionControlActivityArg struct {
	nodeKind     string
	nodePublicID string
	cursor       string
	limit        int32
}

type missionControlResumeTokenPayload struct {
	SnapshotID  string `json:"snapshot_id"`
	ViewMode    string `json:"view_mode"`
	StatePreset string `json:"state_preset"`
	Search      string `json:"search,omitempty"`
	Cursor      string `json:"cursor,omitempty"`
	RootLimit   int32  `json:"root_limit"`
	IssuedAt    string `json:"issued_at"`
}

func resolveMissionControlWorkspaceArg(c *echo.Context) (missionControlWorkspaceArg, error) {
	limit, err := parseLimit(c, missionControlDefaultLimit)
	if err != nil {
		return missionControlWorkspaceArg{}, err
	}
	return missionControlWorkspaceArg{
		viewMode:    strings.TrimSpace(c.QueryParam("view_mode")),
		statePreset: strings.TrimSpace(c.QueryParam("state_preset")),
		search:      strings.TrimSpace(c.QueryParam("search")),
		cursor:      strings.TrimSpace(c.QueryParam("cursor")),
		rootLimit:   int32(limit),
	}, nil
}

func resolveMissionControlNodeArg(c *echo.Context) (missionControlNodeArg, error) {
	nodeKind, err := requirePathParam(c, "node_kind")
	if err != nil {
		return missionControlNodeArg{}, err
	}
	nodePublicID, err := resolvePathUnescaped("node_public_id")(c)
	if err != nil {
		return missionControlNodeArg{}, err
	}
	return missionControlNodeArg{
		nodeKind:     nodeKind,
		nodePublicID: nodePublicID,
	}, nil
}

func resolveMissionControlActivityArg(c *echo.Context) (missionControlActivityArg, error) {
	nodeArg, err := resolveMissionControlNodeArg(c)
	if err != nil {
		return missionControlActivityArg{}, err
	}
	limit, err := parseLimit(c, missionControlDefaultLimit)
	if err != nil {
		return missionControlActivityArg{}, err
	}
	return missionControlActivityArg{
		nodeKind:     nodeArg.nodeKind,
		nodePublicID: nodeArg.nodePublicID,
		cursor:       strings.TrimSpace(c.QueryParam("cursor")),
		limit:        int32(limit),
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
	if payload.RootLimit <= 0 {
		payload.RootLimit = missionControlDefaultLimit
	}
	return payload, nil
}

func newMissionControlResumeTokenPayload(
	arg missionControlWorkspaceArg,
	snapshotID string,
) missionControlResumeTokenPayload {
	return missionControlResumeTokenPayload{
		SnapshotID:  strings.TrimSpace(snapshotID),
		ViewMode:    strings.TrimSpace(arg.viewMode),
		StatePreset: strings.TrimSpace(arg.statePreset),
		Search:      strings.TrimSpace(arg.search),
		Cursor:      strings.TrimSpace(arg.cursor),
		RootLimit:   arg.rootLimit,
		IssuedAt:    time.Now().UTC().Format(time.RFC3339Nano),
	}
}

func missionControlWorkspaceArgFromResumeToken(token missionControlResumeTokenPayload) missionControlWorkspaceArg {
	limit := token.RootLimit
	if limit <= 0 {
		limit = missionControlDefaultLimit
	}
	return missionControlWorkspaceArg{
		viewMode:    strings.TrimSpace(token.ViewMode),
		statePreset: strings.TrimSpace(token.StatePreset),
		search:      strings.TrimSpace(token.Search),
		cursor:      strings.TrimSpace(token.Cursor),
		rootLimit:   limit,
	}
}
