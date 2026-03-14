package enum

// InteractionChannelFamily describes which delivery contour owns the active interaction path.
type InteractionChannelFamily string

const (
	InteractionChannelFamilyPlatformOnly InteractionChannelFamily = "platform_only"
	InteractionChannelFamilyTelegram     InteractionChannelFamily = "telegram"
)
