package cast

import "strings"

// BoolValue is implemented by protobuf wrapper bool types.
type BoolValue interface {
	GetValue() bool
}

// TrimmedStringPtr returns nil for empty strings and pointer for non-empty trimmed values.
func TrimmedStringPtr(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

// OptionalTrimmedString returns nil for nil or empty values, otherwise pointer to trimmed value.
func OptionalTrimmedString(value *string) *string {
	if value == nil {
		return nil
	}
	return TrimmedStringPtr(*value)
}

// TrimmedStringValue returns trimmed string value or empty string for nil.
func TrimmedStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

// PositiveInt32Ptr returns nil for nil/non-positive values.
func PositiveInt32Ptr(value *int32) *int32 {
	if value == nil || *value <= 0 {
		return nil
	}
	v := *value
	return &v
}

// PositiveInt64Ptr returns nil for nil/non-positive values.
func PositiveInt64Ptr(value *int64) *int64 {
	if value == nil || *value <= 0 {
		return nil
	}
	v := *value
	return &v
}

// BoolPtr returns bool pointer for wrapper values.
func BoolPtr(v BoolValue) *bool {
	if v == nil {
		return nil
	}
	value := v.GetValue()
	return &value
}
