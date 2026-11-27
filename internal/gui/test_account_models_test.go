package gui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/provider"
)

// Image and video models on a key-based provider are returned in drawIds and
// filmIds rather than models, and /api/provider/accountmodels accepts them so
// AccountServes lets that key serve image and video generation (#7642).
func TestProviderAccountModelsAcceptsImageAndVideoModels(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	p := provider.Provider{
		ID:          "relay",
		Name:        "Relay",
		Chat:        "https://example.invalid/v1",
		Icon:        "generic",
		Key:         "sk-a",
		KeyName:     "first",
		KeyProtocol: provider.Chat,
		Models:      []string{"chat-1", "Qwen-Image-2.1", "Seedance-1.0"},
	}
	if err := provider.Save(p); err != nil {
		t.Fatal(err)
	}
	if err := provider.SetModelKind("relay/Qwen-Image-2.1", provider.KindImage); err != nil {
		t.Fatal(err)
	}
	if err := provider.SetModelKind("relay/Seedance-1.0", provider.KindVideo); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	providerRoutes(mux, nil)

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/providers", nil))
	if w.Code != 200 {
		t.Fatalf("GET /api/providers: %d %s", w.Code, w.Body)
	}
	var state providersJSON
	if err := json.Unmarshal(w.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	var got *providerJSON
	for i := range state.Providers {
		if state.Providers[i].ID == "relay" {
			got = &state.Providers[i]
			break
		}
	}
	if got == nil {
		t.Fatalf("relay missing from /api/providers: %+v", state.Providers)
	}
	if !slices.Contains(got.DrawIDs, "Qwen-Image-2.1") {
		t.Fatalf("drawIds = %v, want Qwen-Image-2.1", got.DrawIDs)
	}
	if !slices.Contains(got.FilmIDs, "Seedance-1.0") {
		t.Fatalf("filmIds = %v, want Seedance-1.0", got.FilmIDs)
	}

	saved, err := provider.Find("relay")
	if err != nil {
		t.Fatal(err)
	}
	refs := saved.AccountRefs()
	if len(refs) == 0 {
		t.Fatalf("AccountRefs() empty")
	}
	ref := refs[0]

	body, _ := json.Marshal(map[string]any{
		"id":      "relay",
		"account": ref,
		"allow":   []string{"chat-1", "Qwen-Image-2.1", "Seedance-1.0"},
	})
	pw := httptest.NewRecorder()
	mux.ServeHTTP(pw, httptest.NewRequest("POST", "/api/provider/accountmodels", strings.NewReader(string(body))))
	if pw.Code != 200 {
		t.Fatalf("POST /api/provider/accountmodels: %d %s", pw.Code, pw.Body)
	}

	updated, err := provider.Find("relay")
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"chat-1", "Qwen-Image-2.1", "Seedance-1.0"} {
		if !updated.AccountServes(ref, id) {
			t.Fatalf("AccountServes(%q, %q) = false, want true", ref, id)
		}
	}
	if updated.AccountServes(ref, "other-model") {
		t.Fatalf("AccountServes(%q, other-model) = true, want false", ref)
	}
}
