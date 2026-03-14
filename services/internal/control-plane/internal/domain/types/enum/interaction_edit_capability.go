package enum

// InteractionEditCapability captures what the adapter reported after initial message delivery.
type InteractionEditCapability string

const (
	InteractionEditCapabilityUnknown      InteractionEditCapability = "unknown"
	InteractionEditCapabilityEditable     InteractionEditCapability = "editable"
	InteractionEditCapabilityKeyboardOnly InteractionEditCapability = "keyboard_only"
	InteractionEditCapabilityFollowUpOnly InteractionEditCapability = "follow_up_only"
)
