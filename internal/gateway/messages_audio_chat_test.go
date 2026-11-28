package gateway

import (
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/catalog"
	"github.com/yetone/magpie/internal/provider"
)

func TestChatAudioReachesAPluginMessagesModelThatTakesIt(t *testing.T) {
	fresh(t)
	noVision(t)
	_, up := videoUp(t)
	if err := provider.Save(provider.Provider{ID: "anth", Name: "Anth", Anthropic: up.URL, Key: "k", Models: []string{"a"}}); err != nil {
		t.Fatal(err)
	}
	if err := catalog.SaveLive("anth", up.URL, []catalog.Model{
		{ID: "a", InputModalities: []string{"text", "audio"}, OutputModalities: []string{"text"}},
	}); err != nil {
		t.Fatal(err)
	}
	const audio = "UklGRiQAAABXQVZFZm10IBAAAAABAAEAQB8AAIA+AAACABAAZGF0YQAAAAA="
	r, err := parseChat([]byte(`{"model":"anth/a","messages":[{"role":"user","content":[{"type":"input_audio","input_audio":{"data":"` + audio + `","format":"wav"}},{"type":"text","text":"what is said here?"}]}]}`))
	if err != nil {
		t.Fatal(err)
	}
	p := provider.PluginForTest("anth", "agy-cli")
	b, err := buildFor(p, provider.Anthropic, r, "a")
	want := `{"type":"audio","source":{"type":"base64","media_type":"audio/wav","data":"` + audio + `"}}`
	if err != nil || !strings.Contains(string(b), want) {
		t.Fatalf("Chat input_audio sent to plugin Messages as %s (%v), want %s", b, err, want)
	}
}
