package dbmodel

import "time"

// ResponseRecordRow mirrors one interaction_response_records row.
type ResponseRecordRow struct {
	ID               int64     `db:"id"`
	InteractionID    string    `db:"interaction_id"`
	CallbackEventID  int64     `db:"callback_event_id"`
	ResponseKind     string    `db:"response_kind"`
	SelectedOptionID string    `db:"selected_option_id"`
	FreeText         string    `db:"free_text"`
	ResponderRef     string    `db:"responder_ref"`
	Classification   string    `db:"classification"`
	IsEffective      bool      `db:"is_effective"`
	RespondedAt      time.Time `db:"responded_at"`
}
