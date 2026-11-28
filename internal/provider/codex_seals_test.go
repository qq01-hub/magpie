package provider

import (
	"encoding/json"
	"testing"

	"github.com/yetone/magpie/internal/catalog"
	"github.com/yetone/magpie/internal/codexcat"
)

// A model seals a V2 lead's subagent tasks when its provider is a no-sign-in
// Responses relay that can seal them (#1109), or a group has such a member.
// OpenCode Go uses Responses for its own models but does not seal tasks; a
// chat-only key and another account (Copilot) seal none (brethhbbm on X).
func TestCodexSeals(t *testing.T) {
	relay := Provider{ID: "relay", Chat: "https://relay.example/v1", Responses: "https://relay.example/v1"}
	chat := Provider{ID: "ds", Chat: "https://api.deepseek.example/v1"}
	copilot := Provider{ID: "copilot", Responses: "https://api.githubcopilot.example", Account: &Account{Agent: "copilot"}}
	opencodeGo := Provider{ID: "opencode-go", Preset: "opencode-go", Chat: OpenCodeGoChat, Responses: OpenCodeGoChat}
	opencodeGoCustom := Provider{ID: "custom-go", Chat: OpenCodeGoChat, Responses: OpenCodeGoChat}
	members := map[string][]Member{
		"group/mixed":     {{Provider: chat, Model: "deepseek-v4"}, {Provider: relay, Model: "gpt-6-astra"}},
		"group/plain":     {{Provider: chat, Model: "deepseek-v4"}, {Provider: copilot, Model: "gpt-6.1-sol"}},
		"group/go":        {{Provider: chat, Model: "deepseek-v4"}, {Provider: opencodeGo, Model: "deepseek-v4.1-flash"}},
		"group/go-custom": {{Provider: chat, Model: "deepseek-v4"}, {Provider: opencodeGoCustom, Model: "deepseek-v4.1-flash"}},
		"group/go-sealed": {{Provider: opencodeGo, Model: "deepseek-v4.1-flash"}, {Provider: relay, Model: "gpt-6-astra"}},
	}
	of := func(id string) []Member { return members[id] }
	for _, c := range []struct {
		e    Entry
		want bool
	}{
		{Entry{ID: "relay/gpt-6-astra", Provider: &relay, Model: "gpt-6-astra"}, true},
		{Entry{ID: "ds/deepseek-v4", Provider: &chat, Model: "deepseek-v4"}, false},
		{Entry{ID: "copilot/gpt-6.1-sol", Provider: &copilot, Model: "gpt-6.1-sol"}, false},
		{Entry{ID: "opencode-go/deepseek-v4.1-flash", Provider: &opencodeGo, Model: "deepseek-v4.1-flash"}, false},
		{Entry{ID: "custom-go/deepseek-v4.1-flash", Provider: &opencodeGoCustom, Model: "deepseek-v4.1-flash"}, false},
		{Entry{ID: "group/mixed", Group: "mixed"}, true},
		{Entry{ID: "group/plain", Group: "plain"}, false},
		{Entry{ID: "group/go", Group: "go"}, false},
		{Entry{ID: "group/go-custom", Group: "go-custom"}, false},
		{Entry{ID: "group/go-sealed", Group: "go-sealed"}, true},
	} {
		var m catalog.Model
		CodexSeals(&m, c.e, of)
		if m.Seals != c.want {
			t.Errorf("%s: Seals %v, want %v", c.e.ID, m.Seals, c.want)
		}
	}
}

// Codex's actual model catalog must offer Ultra on OpenCode Go models and
// groups, while keeping it off a group containing a ChatGPT-backend relay.
func TestOpenCodeGoCodexCatalogOffersUltra(t *testing.T) {
	goProvider := Provider{ID: "opencode-go", Preset: "opencode-go", Chat: OpenCodeGoChat, Responses: OpenCodeGoChat}
	customGo := Provider{ID: "custom-go", Chat: OpenCodeGoChat, Responses: OpenCodeGoChat}
	relay := Provider{ID: "relay", Chat: "https://relay.example/v1", Responses: "https://relay.example/v1"}
	members := map[string][]Member{
		"group/go":        {{Provider: goProvider, Model: "deepseek-v4.1-flash"}},
		"group/mixed":     {{Provider: goProvider, Model: "deepseek-v4.1-flash"}, {Provider: relay, Model: "gpt-6-astra"}},
		"group/custom-go": {{Provider: customGo, Model: "deepseek-v4.1-flash"}},
	}
	find := func(id string) []Member { return members[id] }
	entries := []Entry{
		{ID: "opencode-go/deepseek-v4.1-flash", Provider: &goProvider, Model: "deepseek-v4.1-flash"},
		{ID: "group/go", Group: "go"},
		{ID: "group/mixed", Group: "mixed"},
		{ID: "custom-go/deepseek-v4.1-flash", Provider: &customGo, Model: "deepseek-v4.1-flash"},
		{ID: "group/custom-go", Group: "custom-go"},
	}
	models := make([]catalog.Model, 0, len(entries))
	for _, e := range entries {
		m := catalog.Model{ID: e.ID, Name: e.ID, Efforts: []string{"low", "high", "max"}}
		CodexSeals(&m, e, find)
		models = append(models, m)
	}
	var got struct {
		Models []struct {
			Slug   string `json:"slug"`
			Levels []struct {
				Effort string `json:"effort"`
			} `json:"supported_reasoning_levels"`
		} `json:"models"`
	}
	if err := json.Unmarshal(codexcat.CatalogAt(models, ""), &got); err != nil {
		t.Fatal(err)
	}
	wantUltra := map[string]bool{
		"opencode-go/deepseek-v4.1-flash": true,
		"group/go":                        true,
		"group/mixed":                     false,
		"custom-go/deepseek-v4.1-flash":   true,
		"group/custom-go":                 true,
	}
	for _, m := range got.Models {
		hasUltra := len(m.Levels) > 0 && m.Levels[len(m.Levels)-1].Effort == "ultra"
		if hasUltra != wantUltra[m.Slug] {
			t.Errorf("%s: Ultra %v, want %v; levels=%+v", m.Slug, hasUltra, wantUltra[m.Slug], m.Levels)
		}
	}
}
