package gateway

import (
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/catalog"
	"github.com/yetone/magpie/internal/provider"
)

// the start of a WAV file (its RIFF header), as base64
const wavData = "UklGRiQAAABXQVZFZm10IBAAAAABAAEAQB8AAIA+AAACABAAZGF0YQAAAAA="

// A Gemini client's sound (inlineData of audio/wav) reaches a plugin's
// Messages model its source says takes audio (agy-cli's gemini-*) as an
// audio block, the video block's shape under its own type. Every other is
// told of it in words, as before: a plugin's model known to take none or
// one nothing is known of; a provider that isn't a plugin, whatever its
// list says, since Anthropic's own API has no audio block and a vendor or
// relay speaking Messages may turn one away; and a sound given by its
// address or in the model's turn, both of which agy-cli's plugin turns
// away.
func TestAudioReachesAPluginModelThatTakesIt(t *testing.T) {
	fresh(t)
	noVision(t)
	anth, up := videoUp(t)
	if err := provider.Save(provider.Provider{ID: "anth", Name: "Anth", Anthropic: up.URL, Key: "k", Models: []string{"a", "t", "u"}}); err != nil {
		t.Fatal(err)
	}
	if err := catalog.SaveLive("anth", up.URL, []catalog.Model{
		{ID: "a", InputModalities: []string{"text", "audio"}, OutputModalities: []string{"text"}},
		{ID: "t", InputModalities: []string{"text", "image"}, OutputModalities: []string{"text"}},
		{ID: "u"},
	}); err != nil {
		t.Fatal(err)
	}
	const inline = `{"contents":[{"role":"user","parts":[{"inlineData":{"mimeType":"audio/wav","data":"` + wavData + `"}},{"text":"what is said here?"}]}]}`
	const words = "[attachment audio/wav]"
	block := `{"type":"audio","source":{"type":"base64","media_type":"audio/wav","data":"` + wavData + `"}}`

	// the provider's own, through the gateway: no plugin, so words
	s := New()
	for _, model := range []string{"anth/a", "anth/t", "anth/u"} {
		anth.reset()
		rec := serveVideo(s, "/v1beta/models/"+model+":generateContent", inline)
		if bodies, _ := anth.sent(); rec.Code != 200 || len(bodies) != 1 || strings.Contains(bodies[0], `"type":"audio"`) || !strings.Contains(bodies[0], words) {
			t.Errorf("to %s: %d %s; upstream %v; want the sound told of in words", model, rec.Code, rec.Body, bodies)
		}
	}

	// a plugin's, as buildFor builds each try for it
	plugin := provider.PluginForTest("anth", "agy-cli")
	r, err := parseGemini([]byte(inline))
	if err != nil {
		t.Fatal(err)
	}
	for model, want := range map[string]string{"a": block, "t": words, "u": words} {
		b, err := buildFor(plugin, provider.Anthropic, r, model)
		if sent := string(b); err != nil || !strings.Contains(sent, want) || want == words && strings.Contains(sent, `"type":"audio"`) {
			t.Errorf("to the plugin's %s: sent %s (%v); want %s", model, sent, err, want)
		}
	}
	for name, c := range map[string]struct{ body, words string }{
		"by its address": {`{"contents":[{"role":"user","parts":[{"fileData":{"mimeType":"audio/wav","fileUri":"https://generativelanguage.googleapis.com/v1beta/files/abc123"}},{"text":"what is said here?"}]}]}`,
			"[attachment audio/wav: https://generativelanguage.googleapis.com/v1beta/files/abc123]"},
		"in the model's turn": {`{"contents":[{"role":"user","parts":[{"text":"say hello"}]},{"role":"model","parts":[{"inlineData":{"mimeType":"audio/wav","data":"` + wavData + `"}}]},{"role":"user","parts":[{"text":"again"}]}]}`,
			words},
	} {
		r, err := parseGemini([]byte(c.body))
		if err != nil {
			t.Fatal(err)
		}
		if b, err := buildFor(plugin, provider.Anthropic, r, "a"); err != nil || strings.Contains(string(b), `"type":"audio"`) || !strings.Contains(string(b), c.words) {
			t.Errorf("a sound %s to the plugin's a: sent %s (%v); want it told of in words", name, b, err)
		}
	}
}
