package cast

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// TimestampRFC3339Nano converts protobuf timestamp to UTC RFC3339Nano string.
func TimestampRFC3339Nano(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return ts.AsTime().UTC().Format(time.RFC3339Nano)
}

// OptionalTimestampRFC3339Nano converts protobuf timestamp to optional UTC RFC3339Nano string.
func OptionalTimestampRFC3339Nano(ts *timestamppb.Timestamp) *string {
	if ts == nil {
		return nil
	}
	v := ts.AsTime().UTC().Format(time.RFC3339Nano)
	return &v
}
