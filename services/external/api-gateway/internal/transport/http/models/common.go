package models

// ItemsResponse is a standard JSON envelope for list endpoints.
type ItemsResponse[T any] struct {
	Items []T `json:"items"`
}
