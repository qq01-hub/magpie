package agent

import (
	"path/filepath"
	"testing"
)

func TestClaudeLatestCaskCLIUpdate(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	brewHome := filepath.Join(home, "homebrew")
	caskBin := filepath.Join(brewHome, "Caskroom", "claude-code@latest", "2.1.292")
	file(t, filepath.Join(caskBin, exe("claude")), "")
	file(t, filepath.Join(brewHome, "bin", "brew"), "")
	t.Setenv("PATH", caskBin)

	oldRunVersion, oldBrewLatest := runVersion, brewLatest
	versions, latests = memo{}, memo{}
	runVersion = func(string) string { return "Claude Code 2.1.292" }
	brewLatest = func(u *updater) (string, error) {
		if u.pkg != "claude-code@latest" || !u.cask {
			t.Fatalf("latest query = %q, cask %v", u.pkg, u.cask)
		}
		return "2.1.302", nil
	}
	t.Cleanup(func() {
		runVersion, brewLatest = oldRunVersion, oldBrewLatest
		versions, latests = memo{}, memo{}
	})

	a := &Agent{ID: "claude", Bin: exe("claude")}
	c, ok := a.CLI()
	if !ok || c.Version != "2.1.292" || c.Latest != "2.1.302" || !c.Update || c.Via != "brew" || c.Command != "brew upgrade --cask claude-code@latest" {
		t.Fatalf("Claude Code latest cask: %+v, %v", c, ok)
	}
}
