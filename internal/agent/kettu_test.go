package agent

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/yetone/magpie/internal/usage"
)

// Kettu runs the user's own `codex app-server` and gives it its name in
// clientInfo, which Codex makes its originator and User-Agent, so a Codex
// turn of Kettu's says "Kettu/0.169.0 (...) (Kettu; 0.1.10)" and was drawn
// as the cube, an unknown client. The User-Agent is the one codex-cli 0.169.0
// sent for an initialize naming "Kettu".
func TestKettuIsAClient(t *testing.T) {
	for ua, want := range map[string]string{
		"Kettu/0.169.0 (Mac OS 27.0.1; arm64) unknown (Kettu; 0.1.10)": "kettu",
		"Kettu/0.1.10": "kettu",
		"codex_cli_rs/0.169.0 (Mac OS 27.0.1; arm64) unknown (codex_cli_rs; 0.1.10)": "codex",
	} {
		if got := usage.AgentOf(ua); got != want {
			t.Errorf("%q: %q, want %q", ua, got, want)
		}
	}
	if !slices.ContainsFunc(Clients(), func(a *Agent) bool { return a.ID == "kettu" && a.Name == "Kettu" && a.Icon == "kettu" }) {
		t.Error("kettu isn't a client the Routing page and Usage draw")
	}
	if _, err := os.Stat(filepath.Join("..", "gui", "assets", "icons", "kettu.png")); err != nil {
		t.Errorf("kettu has no icon file: %v", err)
	}
}
