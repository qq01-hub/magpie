package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDshProfileReadExplicitLead(t *testing.T) {
	_, dir, web, desktop := dshProfileReadHome(t, dshDesktopPatch, "- id: agent-default-model\n  config:\n    provider: deepseek-official\n    model: desktop-model\n")
	if got := dshLead(dir, ""); got != desktop {
		t.Fatalf("implicit lead = %q, want Desktop %q", got, desktop)
	}
	if got := dshLead(dir, "web"); got != web {
		t.Fatalf("explicit Web lead = %q, want %q", got, web)
	}
	if got := dshLead(dir, "desktop"); got != desktop {
		t.Fatalf("explicit Desktop lead = %q, want %q", got, desktop)
	}
}

func TestDshProfileReadUnreadableWeb(t *testing.T) {
	home, _, web, _ := dshProfileReadHome(t, dshDesktopPatch, "- id: agent-default-model\n  config:\n    provider: deepseek-official\n    model: deepseek-flash\n")
	if err := os.Remove(web); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(web, 0o700); err != nil {
		t.Fatal(err)
	}
	a := dsh(home)
	if a.Path != web || a.Field("model").Get() != "" || a.Field("effort").Get() != "" {
		t.Fatalf("an unreadable web profile was treated as empty: path %q, values %v", a.Path, a.Values())
	}
}

func TestDshProfileReadJoinedAndCheck(t *testing.T) {
	home, dir, web := dshRouteHome(t)
	dshKeyFixture(dir)
	dshProfileReadWrite(t, web, dshProfileTemplate)
	desktop := filepath.Join(dir, "profiles", "desktop", "cordis.patch.yml")
	dshProfileReadWrite(t, desktop, dshRouteFixture("deepseek/pro", "deepseek/flash")+
		"- id: agent-default-model\n  config:\n    provider: magpie\n    model: deepseek/flash\n    reasoningEffort: high\n")
	apps := t.TempDir()
	was := appFolders
	appFolders = func() []string { return []string{apps} }
	t.Cleanup(func() { appFolders = was })
	installApp(t, apps, "DeepSeek Harness.app", dshAppPlist)
	a := dsh(home)
	if a.Path != desktop || a.Field("model").Get() != "magpie/deepseek/flash" || a.Field("effort").Get() != "high" {
		t.Fatalf("desktop read source: path %q values %v", a.Path, a.Values())
	}
	if !a.Joined() {
		t.Fatal("the configured Desktop profile's magpie route was not recognized")
	}
	if got := a.Check(); !strings.Contains(got, web) || !strings.Contains(got, "has none of magpie's models") {
		t.Fatalf("Check did not diagnose the non-selected Web profile: %q", got)
	}
}
