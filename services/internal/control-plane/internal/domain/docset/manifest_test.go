package docset

import "testing"

func TestParseManifest_UnsupportedVersion(t *testing.T) {
	_, err := ParseManifest([]byte(`{"manifest_version":2,"id":"x","groups":[]}`))
	if err == nil {
		t.Fatalf("expected error")
	}
}
