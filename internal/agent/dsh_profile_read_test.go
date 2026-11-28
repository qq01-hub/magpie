package agent

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestDshProfileReadSelection(t *testing.T) {
	native := "- id: agent-default-model\n  config:\n    provider: deepseek-official\n    model: deepseek-flash\n    reasoningEffort: high\n"
	webPick := "- id: agent-default-model\n  config:\n    provider: deepseek-official\n    model: deepseek-v4-pro\n    reasoningEffort: low\n"
	for _, tc := range []struct {
		name, web, desktop, homePatch, settings, legacyConfig string
		profile, model, effort                                string
	}{
		{name: "empty web template", web: dshDesktopPatch, desktop: native, profile: "desktop", model: "deepseek-flash", effort: "high"},
		{name: "configured web wins", web: webPick, desktop: native, profile: "web", model: "deepseek-v4-pro", effort: "low"},
		{name: "nonempty web is not an unused template", web: "- id: ui-settings-general\n  config:\n    theme: dark\n", desktop: native, profile: "web"},
		{name: "invalid web is unknown", web: "not a patch list\n", desktop: native, profile: "web"},
		{name: "no configured profile", web: dshDesktopPatch, desktop: dshDesktopPatch, profile: "web"},
		{name: "home model keeps web", web: dshDesktopPatch, desktop: native, homePatch: webPick, profile: "web", model: "deepseek-v4-pro", effort: "low"},
		{name: "legacy settings keep web", web: dshDesktopPatch, desktop: native, settings: "agent-default-model:\n  provider: deepseek-official\n  model: deepseek-v4-pro\n", profile: "web", model: "deepseek-v4-pro"},
		{name: "unused legacy config does not block desktop", web: dshDesktopPatch, desktop: native, legacyConfig: "agent-loop:\n  config:\n    model: old-config-model\n", profile: "desktop", model: "deepseek-flash", effort: "high"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			home, dir, web, desktop := dshProfileReadHome(t, tc.web, tc.desktop)
			if tc.homePatch != "" {
				dshProfileReadWrite(t, filepath.Join(dir, "cordis.patch.yml"), tc.homePatch)
			}
			if tc.settings != "" {
				dshProfileReadWrite(t, filepath.Join(dir, "settings.yaml"), tc.settings)
			}
			if tc.legacyConfig != "" {
				dshProfileReadWrite(t, filepath.Join(dir, "config.yaml"), tc.legacyConfig)
			}
			a := dsh(home)
			if got := a.Field("model").Get(); got != tc.model {
				t.Errorf("model = %q, want %q", got, tc.model)
			}
			if got := a.Field("effort").Get(); got != tc.effort {
				t.Errorf("effort = %q, want %q", got, tc.effort)
			}
			wantPath := web
			if tc.profile == "desktop" {
				wantPath = desktop
			}
			if a.Path != wantPath {
				t.Errorf("config path = %q, want %q", a.Path, wantPath)
			}
			if files := DshPatchFiles(dir); len(files) != 2 || files[0] != web {
				t.Errorf("write targets changed: %v", files)
			}
			for path, want := range map[string]string{web: tc.web, desktop: tc.desktop} {
				got, err := os.ReadFile(path)
				if err != nil || !bytes.Equal(got, []byte(want)) {
					t.Errorf("reading changed %s: %v", path, err)
				}
			}
		})
	}
}

func dshProfileReadHome(t *testing.T, webPatch, desktopPatch string) (home, dir, web, desktop string) {
	t.Helper()
	home = t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("DSH_HOME", "")
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(home, ".cache"))
	dir = filepath.Join(home, ".dsh")
	web = filepath.Join(dir, "profiles", "web", "cordis.patch.yml")
	desktop = filepath.Join(dir, "profiles", "desktop", "cordis.patch.yml")
	dshProfileReadWrite(t, web, webPatch)
	dshProfileReadWrite(t, desktop, desktopPatch)
	return
}

func dshProfileReadWrite(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
}
