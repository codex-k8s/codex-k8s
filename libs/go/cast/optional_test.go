package cast

import (
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestTrimmedStringPtr(t *testing.T) {
	if got := TrimmedStringPtr("   "); got != nil {
		t.Fatalf("expected nil, got %q", *got)
	}
	got := TrimmedStringPtr("  hello  ")
	if got == nil || *got != "hello" {
		t.Fatalf("expected hello, got %#v", got)
	}
}

func TestOptionalTrimmedString(t *testing.T) {
	if got := OptionalTrimmedString(nil); got != nil {
		t.Fatalf("expected nil for nil input, got %q", *got)
	}
	empty := "   "
	if got := OptionalTrimmedString(&empty); got != nil {
		t.Fatalf("expected nil for empty input, got %q", *got)
	}
	value := "  world "
	got := OptionalTrimmedString(&value)
	if got == nil || *got != "world" {
		t.Fatalf("expected world, got %#v", got)
	}
}

func TestTrimmedStringValue(t *testing.T) {
	if got := TrimmedStringValue(nil); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
	value := "  foo "
	if got := TrimmedStringValue(&value); got != "foo" {
		t.Fatalf("expected foo, got %q", got)
	}
}

func TestPositiveIntPointers(t *testing.T) {
	if got := PositiveInt32Ptr(nil); got != nil {
		t.Fatalf("expected nil for nil int32")
	}
	zero32 := int32(0)
	if got := PositiveInt32Ptr(&zero32); got != nil {
		t.Fatalf("expected nil for zero int32")
	}
	one32 := int32(1)
	if got := PositiveInt32Ptr(&one32); got == nil || *got != 1 {
		t.Fatalf("expected 1 for int32, got %#v", got)
	}

	if got := PositiveInt64Ptr(nil); got != nil {
		t.Fatalf("expected nil for nil int64")
	}
	zero64 := int64(0)
	if got := PositiveInt64Ptr(&zero64); got != nil {
		t.Fatalf("expected nil for zero int64")
	}
	one64 := int64(1)
	if got := PositiveInt64Ptr(&one64); got == nil || *got != 1 {
		t.Fatalf("expected 1 for int64, got %#v", got)
	}
}

func TestBoolPtr(t *testing.T) {
	if got := BoolPtr(nil); got != nil {
		t.Fatalf("expected nil for nil wrapper")
	}
	got := BoolPtr(wrapperspb.Bool(true))
	if got == nil || *got != true {
		t.Fatalf("expected true, got %#v", got)
	}
}

func TestTimestamps(t *testing.T) {
	ts := timestamppb.New(time.Date(2026, time.January, 2, 3, 4, 5, 123000000, time.FixedZone("X", 3*3600)))
	got := TimestampRFC3339Nano(ts)
	if got != "2026-01-02T00:04:05.123Z" {
		t.Fatalf("unexpected timestamp string: %q", got)
	}

	optional := OptionalTimestampRFC3339Nano(ts)
	if optional == nil || *optional != got {
		t.Fatalf("unexpected optional timestamp: %#v", optional)
	}
	if OptionalTimestampRFC3339Nano(nil) != nil {
		t.Fatalf("expected nil for nil timestamp")
	}
}
