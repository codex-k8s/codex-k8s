package app

import (
	"reflect"
	"testing"
)

func TestNormalizeBootstrapEmails(t *testing.T) {
	got, err := normalizeBootstrapEmails([]string{" Owner@example.com ", "owner@example.com", "", "USER@example.com"}, "TEST_ENV")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"owner@example.com", "user@example.com"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected result: got=%v want=%v", got, want)
	}
}

func TestNormalizeBootstrapEmails_Invalid(t *testing.T) {
	_, err := normalizeBootstrapEmails([]string{"invalid-email"}, "TEST_ENV")
	if err == nil {
		t.Fatalf("expected validation error")
	}
}
