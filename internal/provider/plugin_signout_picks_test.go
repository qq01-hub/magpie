package provider

import (
	"context"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/yetone/magpie/internal/plugin"
	"github.com/yetone/magpie/internal/testenv"
)

// A plugin whose keys list different models, as a new-api relay's plugin
// does (one key with every model, another with fake-1 alone): removing
// the key that listed fake-claude and fake-gemini takes them out of the
// picks, set aside until a list has them again, and a pick typed in by
// hand stays. The listing after the sign-out was short of them already,
// so Fetch models and the hourly relist found nothing gone, and agents
// were offered models no key served. An editor opened before, saved with
// the picks it held, doesn't put them back.
func TestRemovedKeysModelsLeaveThePicks(t *testing.T) {
	bun := testenv.Bun(t)
	claudeHome(t)
	t.Setenv("MAGPIE_BUN", bun)
	t.Cleanup(plugin.Settle)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	abs, _ := filepath.Abs("../plugin/testdata/fake/index.js")
	if _, err := plugin.Add(ctx, abs); err != nil {
		t.Fatal(err)
	}
	// "few" is told fake-1 alone, "full" every model
	for _, key := range []string{"few", "full"} {
		if _, err := PluginAPIKey(ctx, "fakeco", 0, nil, key); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := plugin.Providers(ctx); err != nil {
		t.Fatal(err)
	}
	plugin.Refreshed()
	id := PluginID("fakeco")
	listed := func() map[string]bool { return pluginListedIDs(plugin.Cached())[id] }
	if l := listed(); !l["fake-1"] || !l["fake-claude"] || !l["fake-gemini"] {
		t.Fatalf("the two keys list %v, want fake-1, fake-claude and fake-gemini", l)
	}
	picks := []string{"fake-1", "fake-claude", "fake-gemini", "mine-typed"}
	if err := Save(Provider{ID: id, Models: picks}); err != nil {
		t.Fatal(err)
	}

	// the full key removed, as the editor's remove does
	var pp plugin.Provider
	for _, p := range heldPlugins() {
		if p.ID == "fakeco" {
			pp = p
		}
	}
	auths := plugin.Auths("fakeco")
	user := ""
	for _, l := range pluginLogins(pp) {
		if auths[l.acct.Key]["key"] == "full" {
			user = l.User
		}
	}
	if user == "" {
		t.Fatalf("no account holds the full key: %v", pluginLogins(pp))
	}
	if err := ForgetLogin(pluginAgent(pp), user); err != nil {
		t.Fatal(err)
	}
	// the listing in the background after it, as the window's next look
	// starts it
	for deadline := time.Now().Add(30 * time.Second); listed()["fake-claude"]; {
		if time.Now().After(deadline) {
			t.Fatalf("still listed after the full key went: %v", listed())
		}
		time.Sleep(20 * time.Millisecond)
	}
	plugin.Refreshed()

	got := func(when string) {
		t.Helper()
		p, err := Find(id)
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(p.Models, []string{"fake-1", "mine-typed"}) {
			t.Fatalf("%s: picks %v, want fake-1 and mine-typed: fake-claude and fake-gemini left with the key", when, p.Models)
		}
		if !slices.Contains(p.Delisted, "fake-claude") || !slices.Contains(p.Delisted, "fake-gemini") {
			t.Fatalf("%s: set aside %v, want fake-claude and fake-gemini", when, p.Delisted)
		}
	}
	got("after the key was removed")

	// an editor opened before the key went saves the picks it held
	if err := Save(Provider{ID: id, Models: picks}); err != nil {
		t.Fatal(err)
	}
	got("after an editor opened before saved")

	// the key added again lists them again: they come back
	if _, err := PluginAPIKey(ctx, "fakeco", 0, nil, "full"); err != nil {
		t.Fatal(err)
	}
	for deadline := time.Now().Add(30 * time.Second); !listed()["fake-claude"]; {
		if time.Now().After(deadline) {
			t.Fatalf("not listed after the full key came back: %v", listed())
		}
		time.Sleep(20 * time.Millisecond)
	}
	plugin.Refreshed()
	p, err := Find(id)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range picks {
		if !slices.Contains(p.Models, m) {
			t.Fatalf("picks %v after the full key came back, want %v", p.Models, picks)
		}
	}
}
