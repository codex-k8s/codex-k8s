package realtime

import (
	"encoding/json"
	"strings"
	"time"
)

const (
	defaultTopicRunEvents    = "run.events"
	defaultTopicRunStatus    = "run.status"
	defaultTopicRunLogs      = "run.logs"
	defaultTopicDeployEvents = "deploy.events"
	defaultTopicDeployLogs   = "deploy.logs"
	defaultTopicSystemErrors = "system.errors"
)

var defaultTopics = []string{
	defaultTopicRunEvents,
	defaultTopicRunStatus,
	defaultTopicRunLogs,
	defaultTopicDeployEvents,
	defaultTopicDeployLogs,
	defaultTopicSystemErrors,
}

// Event is one realtime payload delivered to websocket clients.
type Event struct {
	ID            int64
	Topic         string
	ScopeJSON     json.RawMessage
	PayloadJSON   json.RawMessage
	CorrelationID string
	ProjectID     string
	RunID         string
	TaskID        string
	CreatedAt     time.Time
}

// SubscriptionFilter is per-connection topic/scope filter.
type SubscriptionFilter struct {
	Topics    []string
	ProjectID string
	RunID     string
	TaskID    string
}

// Normalize returns a canonical filter shape.
func (f SubscriptionFilter) Normalize() SubscriptionFilter {
	normalized := SubscriptionFilter{
		ProjectID: strings.TrimSpace(f.ProjectID),
		RunID:     strings.TrimSpace(f.RunID),
		TaskID:    strings.TrimSpace(f.TaskID),
	}

	if len(f.Topics) == 0 {
		normalized.Topics = append(normalized.Topics, defaultTopics...)
		return normalized
	}

	seen := make(map[string]struct{}, len(f.Topics))
	for _, raw := range f.Topics {
		topic := strings.TrimSpace(raw)
		if topic == "" {
			continue
		}
		if _, ok := seen[topic]; ok {
			continue
		}
		seen[topic] = struct{}{}
		normalized.Topics = append(normalized.Topics, topic)
	}
	if len(normalized.Topics) == 0 {
		normalized.Topics = append(normalized.Topics, defaultTopics...)
	}
	return normalized
}

// Matches checks if event satisfies topic/scope filter.
func (f SubscriptionFilter) Matches(event Event) bool {
	normalized := f.Normalize()
	if len(normalized.Topics) > 0 {
		matched := false
		for _, topic := range normalized.Topics {
			if topic == event.Topic {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	if normalized.ProjectID != "" && normalized.ProjectID != event.ProjectID {
		return false
	}
	if normalized.RunID != "" && normalized.RunID != event.RunID {
		return false
	}
	if normalized.TaskID != "" && normalized.TaskID != event.TaskID {
		return false
	}
	return true
}
