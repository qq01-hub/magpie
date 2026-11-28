package agent

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/claudecode"
	"github.com/yetone/magpie/internal/proc"
)

func TestManagedClaudeVersionFallback(t *testing.T) {
	home, cache, bin := t.TempDir(), t.TempDir(), t.TempDir()
	for name, value := range map[string]string{"HOME": home, "USERPROFILE": home, "XDG_CACHE_HOME": cache, "LOCALAPPDATA": cache, "APPDATA": cache, "PATH": bin} {
		t.Setenv(name, value)
	}
	proc.Forget()
	t.Cleanup(proc.Forget)
	oldRun := runVersion
	runVersion = func(path string) string {
		switch {
		case strings.Contains(path, "2.1.330"):
			return "2.1.330 (Claude Code)"
		case strings.Contains(path, "2.1.99"):
			return "2.1.99 (Claude Code)"
		case filepath.Dir(path) == bin:
			return "2.1.319 (Claude Code)"
		}
		t.Errorf("unexpected version probe %q", path)
		return ""
	}
	t.Cleanup(func() { runVersion = oldRun; versions = memo{} })
	for _, version := range []string{"2.1.99", "2.1.330"} {
		file(t, filepath.Join(claudecode.Root(), version, exe("claude")), "MZ")
	}
	want := filepath.Join(claudecode.Root(), "2.1.330", exe("claude"))
	a := &Agent{ID: "claude", Name: "Claude Code", Bin: "claude"}
	got, ok := a.CLI()
	if !ok || got.Version != "2.1.330" || got.Path != want || got.Update || got.Command != "" {
		t.Fatalf("managed Claude version missing or offers an ordinary updater: %+v, %v", got, ok)
	}
	if got := a.InstalledVersion(); got != "2.1.330" {
		t.Errorf("managed installed version %q", got)
	}
	normal := file(t, filepath.Join(bin, exe("claude")), "MZ")
	proc.Forget()
	if got, ok := a.CLI(); !ok || got.Path != normal || got.Version != "2.1.319" {
		t.Errorf("ordinary installation does not take precedence: %+v, %v", got, ok)
	}
}

func TestAgyVersionHasNoInstaller(t *testing.T) {
	bin := t.TempDir()
	t.Setenv("PATH", bin)
	proc.Forget()
	t.Cleanup(proc.Forget)
	want := file(t, filepath.Join(bin, exe("agy")), "MZ")
	oldRun := runVersion
	runVersion = func(path string) string {
		if path != want {
			t.Errorf("version probe for %q, want %q", path, want)
		}
		return "1.5.15"
	}
	t.Cleanup(func() { runVersion = oldRun; versions = memo{} })
	a := &Agent{ID: "agy", Name: "Antigravity CLI", Bin: "agy"}
	if got, ok := a.CLI(); !ok || got.Version != "1.5.15" || got.Path != want || got.Update || got.Command != "" {
		t.Errorf("agy version missing or updater offered: %+v, %v", got, ok)
	}
	if got := a.InstalledVersion(); got != "1.5.15" {
		t.Errorf("agy installed version %q", got)
	}
	if got := installCommands("agy", "windows", true); len(got) != 0 {
		t.Errorf("agy has an installer: %+v", got)
	}
}
