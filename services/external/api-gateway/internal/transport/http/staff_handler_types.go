package http

// pathLimit keeps parsed path id and list limit together for staff list handlers.
type pathLimit struct {
	id    string
	limit int
}

// idLimitArg is a helper DTO for gRPC request builders with id+limit.
type idLimitArg struct {
	id    string
	limit int32
}
