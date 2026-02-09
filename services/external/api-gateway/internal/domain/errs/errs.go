package errs

import "fmt"

// Validation indicates a domain validation failure for incoming data.
type Validation struct {
	// Field is a logical input field name, if applicable.
	Field string
	// Msg is a human-readable validation failure reason.
	Msg string
}

// Error returns a stable validation error message.
func (e Validation) Error() string {
	if e.Field == "" {
		return fmt.Sprintf("validation failed: %s", e.Msg)
	}
	return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Msg)
}

// Unauthorized indicates authentication or authorization failure.
type Unauthorized struct {
	// Msg overrides the default unauthorized message when provided.
	Msg string
}

// Error returns a stable unauthorized error message.
func (e Unauthorized) Error() string {
	if e.Msg == "" {
		return "unauthorized"
	}
	return e.Msg
}

// Conflict indicates a state conflict such as duplicate or race condition.
type Conflict struct {
	// Msg overrides the default conflict message when provided.
	Msg string
}

// Error returns a stable conflict error message.
func (e Conflict) Error() string {
	if e.Msg == "" {
		return "conflict"
	}
	return e.Msg
}
