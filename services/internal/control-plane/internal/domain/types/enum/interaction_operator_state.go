package enum

// InteractionOperatorState is an operator-facing projection of current interaction health.
type InteractionOperatorState string

const (
	InteractionOperatorStateNominal                InteractionOperatorState = "nominal"
	InteractionOperatorStateWatch                  InteractionOperatorState = "watch"
	InteractionOperatorStateManualFallbackRequired InteractionOperatorState = "manual_fallback_required"
	InteractionOperatorStateResolved               InteractionOperatorState = "resolved"
)
