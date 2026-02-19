package mcp

import "testing"

func TestBuildRunStatusReportInputSchema(t *testing.T) {
	t.Parallel()

	schema := buildRunStatusReportInputSchema()
	if schema == nil {
		t.Fatal("expected non-nil schema")
	}
	if schema.Properties == nil {
		t.Fatal("expected schema properties")
	}

	statusSchema, ok := schema.Properties["status"]
	if !ok || statusSchema == nil {
		t.Fatal("expected status property schema")
	}
	if statusSchema.MaxLength == nil || *statusSchema.MaxLength != 100 {
		t.Fatalf("status maxLength = %v, want 100", statusSchema.MaxLength)
	}
	if statusSchema.MinLength == nil || *statusSchema.MinLength != 1 {
		t.Fatalf("status minLength = %v, want 1", statusSchema.MinLength)
	}
}
