package entity

import "time"

// RuntimeError is one persisted runtime failure journal entry.
type RuntimeError struct {
	ID            string
	Source        string
	Level         string
	Message       string
	DetailsJSON   []byte
	StackTrace    string
	CorrelationID string
	RunID         string
	ProjectID     string
	Namespace     string
	JobName       string
	ViewedAt      *time.Time
	ViewedBy      string
	CreatedAt     time.Time
}
