package agent

import (
	"path/filepath"
	"runtime"
	"testing"
)

// VS Code's cache target uses Node's platform/architecture spellings
func hostSDKPlatformT(t *testing.T) string {
	t.Helper()
	platform := runtime.GOOS
	if platform == "windows" {
		platform = "win32"
	}
	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "x64"
	}
	return platform + "-" + arch
}

func hostSDKRootT(home string) string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support")
	case "windows":
		return filepath.Join(home, "AppData", "Roaming")
	default:
		return filepath.Join(home, ".config")
	}
}

// Without a CLI or extension, a complete current-platform SDK still reads
// the user's agent config. Other platforms and interrupted downloads do not
func TestAgentHostSDKKeepsAgentConfigActive(t *testing.T) {
	judgeOffWindows(t)
	otherPlatform := "linux-arm64"
	if hostSDKPlatformT(t) == otherPlatform {
		otherPlatform = "win32-x64"
	}
	for _, id := range []string{"claude", "codex"} {
		for _, c := range []struct {
			name, platform string
			complete       bool
			missing        bool
		}{
			{"complete", hostSDKPlatformT(t), true, false},
			{"partial", hostSDKPlatformT(t), false, true},
			{"other-platform", otherPlatform, true, true},
		} {
			t.Run(id+"/"+c.name, func(t *testing.T) {
				home := t.TempDir()
				t.Setenv("HOME", home)
				t.Setenv("USERPROFILE", home)
				t.Setenv("APPDATA", filepath.Join(home, "AppData", "Roaming"))
				t.Setenv("LOCALAPPDATA", filepath.Join(home, "AppData", "Local"))
				t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
				t.Setenv("XDG_CACHE_HOME", filepath.Join(home, ".cache"))
				t.Setenv("PATH", t.TempDir())
				noApps(t)
				dir := filepath.Join(home, "."+id)
				writeT(t, filepath.Join(dir, "settings.json"), "{}")
				cache := filepath.Join(hostSDKRootT(home), "Code", "agent-host", "sdk-cache", id, "0.3.281", c.platform)
				mkdirT(t, cache)
				if c.complete {
					writeT(t, filepath.Join(cache, ".complete"), "")
				}
				a := &Agent{ID: id, Name: id, Bin: "magpie-test-host-sdk-" + id, Dir: dir, Path: filepath.Join(dir, "settings.json")}
				if got := a.CLIMissing(); got != c.missing {
					t.Fatalf("%s %s: CLI missing = %t, want %t", id, c.name, got, c.missing)
				}
			})
		}
	}
}

func TestAgentHostSDKBesideCindyIsDetected(t *testing.T) {
	for _, id := range []string{"claude", "codex"} {
		for _, complete := range []bool{false, true} {
			t.Run(id+map[bool]string{false: "/partial", true: "/complete"}[complete], func(t *testing.T) {
				home, cindyDir := cindyDetectHome(t)
				t.Setenv("CODEX_HOME", "")
				var a *Agent
				if id == "claude" {
					writeT(t, filepath.Join(home, ".claude", "settings.json"), "{}")
					writeT(t, filepath.Join(cindyDir, "claude-code", "2.1.321", exe(id)), "")
					a = claude(home)
				} else {
					writeT(t, filepath.Join(home, ".codex", "config.toml"), "")
					writeT(t, filepath.Join(cindyDir, "codex-package", "0.130.0", "bin", exe(id)), "")
					a = codex(home)
				}
				cache := agentHostCacheDir(home, id, "0.3.281", hostSDKPlatformT(t))
				mkdirT(t, cache)
				if complete {
					writeT(t, filepath.Join(cache, ".complete"), "")
				}
				if got := a.Detected(); got != complete {
					t.Fatalf("Cindy beside SDK complete=%t: %s detected=%t", complete, id, got)
				}
			})
		}
	}
}
