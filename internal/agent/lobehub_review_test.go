package agent

import (
	"testing"

	"github.com/yetone/magpie/internal/usage"
)

func TestLobeHubClientIdentityIsExplicit(t *testing.T) {
	for _, ua := range []string{"lobehub/2.2.20", "LobeHub/2.2.20", "LOBEHUB/2.2.20", "lobehub"} {
		if got := usage.AgentOf(ua); got != "lobehub" {
			t.Errorf("AgentOf(%q) = %q, want lobehub", ua, got)
		}
	}
	for _, ua := range []string{"k/JS 6.49.0", "OpenAI/JS 6.49.0"} {
		if got := usage.AgentOf(ua); got == "lobehub" {
			t.Errorf("unrelated or unnamed SDK %q became LobeHub", ua)
		}
	}
	a, _ := FindIn(ClientsIn(nil), "lobehub")
	if a == nil || a.Name != "LobeHub" || a.Icon != "lobehub-color" {
		t.Fatalf("LobeHub client metadata = %+v", a)
	}
	if a.Bin != "" || a.Path != "" || len(a.Fields) != 0 {
		t.Fatalf("client unexpectedly has an editable agent config: %+v", a)
	}
	if a, _ := FindIn(All(), "lobehub"); a != nil {
		t.Fatal("LobeHub became a configurable agent row")
	}
}
