package gui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/yetone/magpie/internal/plugin"
	"github.com/yetone/magpie/internal/provider"
	"github.com/yetone/magpie/internal/testenv"
)

// A plugin's editor, open while one of its keys was removed, held the
// picks of the models only that key listed; the listing after the
// sign-out set them aside, so Fetch models dropped nothing and the editor
// kept them. Fetch models names every pick set aside, as a remote
// magpie's does, and the editor takes them out.
func TestPluginFetchNamesPicksSetAsideBefore(t *testing.T) {
	bun := testenv.Bun(t)
	dir := t.TempDir()
	testenv.SetHome(t, dir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(dir, "cache"))
	t.Setenv("MAGPIE_BUN", bun)
	t.Setenv("MAGPIE_PLUGIN_MARKET", "off")
	t.Cleanup(plugin.Settle)
	abs, _ := filepath.Abs("../plugin/testdata/fake/index.js")
	if _, err := plugin.Add(t.Context(), abs); err != nil {
		t.Fatal(err)
	}
	// "few" is told fake-1 alone, "full" every model
	for _, key := range []string{"few", "full"} {
		if _, err := provider.PluginAPIKey(t.Context(), "fakeco", 0, nil, key); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := plugin.Providers(t.Context()); err != nil {
		t.Fatal(err)
	}
	plugin.Refreshed()
	id := provider.PluginID("fakeco")
	if err := provider.Save(provider.Provider{ID: id, Models: []string{"fake-1", "fake-claude", "fake-gemini"}}); err != nil {
		t.Fatal(err)
	}
	// the full key, signed in second, is the second account
	ls := provider.Logins("plugin:fakeco")
	if len(ls) != 2 {
		t.Fatalf("accounts %+v, want the two keys", ls)
	}
	if err := provider.ForgetLogin("plugin:fakeco", ls[1].User); err != nil {
		t.Fatal(err)
	}
	listed := func() bool {
		for _, p := range plugin.Cached() {
			if p.ID == "fakeco" && slices.ContainsFunc(p.Models, func(m plugin.Model) bool { return m.ID == "fake-claude" }) {
				return true
			}
		}
		return false
	}
	for deadline := time.Now().Add(30 * time.Second); listed(); {
		if time.Now().After(deadline) {
			t.Fatal("fake-claude still listed after the full key went")
		}
		time.Sleep(20 * time.Millisecond)
	}
	plugin.Refreshed()

	mux := http.NewServeMux()
	providerRoutes(mux, nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("POST", "/api/provider/models", strings.NewReader(`{"id":"`+id+`"}`)))
	if w.Code != 200 {
		t.Fatalf("Fetch models: %d %s", w.Code, w.Body)
	}
	var r struct {
		Dropped []string `json:"dropped"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &r); err != nil {
		t.Fatal(err)
	}
	slices.Sort(r.Dropped)
	if !slices.Equal(r.Dropped, []string{"fake-claude", "fake-gemini"}) {
		t.Fatalf("Fetch models after the full key went names %v, so the open editor keeps fake-claude and fake-gemini", r.Dropped)
	}
}
