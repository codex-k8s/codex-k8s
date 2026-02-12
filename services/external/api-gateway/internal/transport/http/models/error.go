package models

// ErrorResponse is a normalized HTTP error payload for API gateway endpoints.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}
