package manifesttpl

import (
	"strings"
	"testing"
)

func TestRender_TrimPrefixTemplateHelperUsesPrefixFirstOrder(t *testing.T) {
	t.Parallel()

	raw := []byte(strings.TrimSpace(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: demo
data:
  tag: '{{ trimPrefix "v" "v1.32.2" }}'
`) + "\n")

	out, err := Render("trim-prefix", raw, nil)
	if err != nil {
		t.Fatalf("render template: %v", err)
	}

	rendered := string(out)
	if !strings.Contains(rendered, "tag: '1.32.2'") {
		t.Fatalf("unexpected rendered template:\n%s", rendered)
	}
}
