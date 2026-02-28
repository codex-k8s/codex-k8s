package staff

type servicesYAMLPreflightEnvSlot struct {
	Env  string
	Slot int
}

var servicesYAMLPreflightEnvSlots = []servicesYAMLPreflightEnvSlot{
	{Env: "production", Slot: 0},
	{Env: "ai", Slot: 1},
}
