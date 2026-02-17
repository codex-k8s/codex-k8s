package docset

import "testing"

func TestParseManifest_UnsupportedVersion(t *testing.T) {
	_, err := ParseManifest([]byte(`{"manifest_version":2,"docset":{"id":"x"},"groups":[],"items":[]}`))
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestParseManifest_NormalizesSHA256Prefix(t *testing.T) {
	manifest, err := ParseManifest([]byte(`{
  "manifest_version": 1,
  "docset": { "id": "x" },
  "groups": [
    {
      "id": "core",
      "title": { "ru": "x", "en": "x" },
      "description": { "ru": "", "en": "" },
      "default_selected": true,
      "items": ["file:a"]
    }
  ],
  "items": [
    {
      "id": "file:a",
      "kind": "file",
      "category": "core",
      "common": true,
      "stack_tags": [],
      "import_path": "docs/a.md",
      "source_paths": { "ru": "docs/a_ru.md", "en": "docs/a_en.md" },
      "sha256": { "ru": "sha256:abc", "en": "sha256:def" },
      "title": { "ru": "", "en": "" },
      "description": { "ru": "", "en": "" }
    }
  ]
}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(manifest.Items) != 1 {
		t.Fatalf("unexpected items: %d", len(manifest.Items))
	}
	if got := manifest.Items[0].SHA256.RU; got != "abc" {
		t.Fatalf("RU sha=%q, want %q", got, "abc")
	}
	if got := manifest.Items[0].SHA256.EN; got != "def" {
		t.Fatalf("EN sha=%q, want %q", got, "def")
	}
}
