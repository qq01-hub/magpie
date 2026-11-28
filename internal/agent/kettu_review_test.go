package agent

import (
	"testing"

	"github.com/yetone/magpie/internal/usage"
)

func TestKettuIsOnlyARequestClient(t *testing.T) {
	for _, ua := range []string{"kettu", "Kettu", "KETTU/0.1.10"} {
		if got := usage.AgentOf(ua); got != "kettu" {
			t.Errorf("AgentOf(%q) = %q, want kettu", ua, got)
		}
	}
	a, _ := FindIn(ClientsIn(nil), "kettu")
	if a == nil || a.Name != "Kettu" || a.Icon != "kettu" {
		t.Fatalf("Kettu request-client metadata = %+v", a)
	}
	if a.Bin != "" || a.Path != "" || len(a.Fields) != 0 {
		t.Fatalf("Kettu unexpectedly has editable agent configuration: %+v", a)
	}
	if a, _ := FindIn(All(), "kettu"); a != nil {
		t.Fatal("Kettu became a configurable agent row")
	}
	if got := usage.AgentOf("codex_cli_rs/0.169.0"); got != "codex" {
		t.Fatalf("Codex itself became %q", got)
	}
}
