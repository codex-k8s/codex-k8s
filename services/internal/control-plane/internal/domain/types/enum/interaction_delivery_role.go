package enum

// InteractionDeliveryRole distinguishes primary dispatch from post-callback continuation attempts.
type InteractionDeliveryRole string

const (
	InteractionDeliveryRolePrimaryDispatch InteractionDeliveryRole = "primary_dispatch"
	InteractionDeliveryRoleMessageEdit     InteractionDeliveryRole = "message_edit"
	InteractionDeliveryRoleFollowUpNotify  InteractionDeliveryRole = "follow_up_notify"
)
