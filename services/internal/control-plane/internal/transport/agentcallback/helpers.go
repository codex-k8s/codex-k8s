package agentcallback

import (
	"fmt"
	"strings"

	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
)

// ParseBearerToken extracts bearer token from Authorization header value.
func ParseBearerToken(value string) string {
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

// ParseEventType validates allowed Day4 agent callback events.
func ParseEventType(value string) (floweventdomain.EventType, error) {
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
