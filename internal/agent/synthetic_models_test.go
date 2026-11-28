package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/catalog"
	"github.com/yetone/magpie/internal/gateway"
	"github.com/yetone/magpie/internal/provider"
)

// Exercise both reported Gateway formats with the actual OpenCode entry
// builder registered by this package, rather than substituting its callback.
func TestSyntheticLiveModelsReachGatewayFormats(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(home, ".cache"))
	t.Setenv("PATH", home)
	if err := os.MkdirAll(filepath.Dir(catalog.CachePath()), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(catalog.CachePath(), []byte(`{"synthetic":{"models":{
	  "hf:deepseek-ai/DeepSeek-V4.1-Flash-1017":{"id":"hf:deepseek-ai/DeepSeek-V4.1-Flash-1017","name":"DeepSeek V4.1 Flash 10/17","limit":{"context":1000000,"output":384000},"reasoning_options":[{"type":"effort","values":["none","low","high","max"]}]}
	}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	catalog.Reset()
	t.Cleanup(catalog.Reset)
	body, err := os.ReadFile("../provider/testdata/synthetic-models-7430.json")
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/openai/v1/models" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	}))
	defer srv.Close()
	old := http.DefaultClient.Transport
	http.DefaultClient.Transport = syntheticModelsTransport{srv}
	t.Cleanup(func() { http.DefaultClient.Transport = old })
	const base = "https://api.synthetic.new/openai/v1"
	p := provider.Provider{ID: "synthetic", Name: "Synthetic", Chat: base, Key: "fixture-key",
		Models: []string{"infinite:small:vision", "hf:deepseek-ai/DeepSeek-V4.1-Flash-1017"}}
	if err := provider.Save(p); err != nil {
		t.Fatal(err)
	}
	if _, err := p.Fetch(context.Background()); err != nil {
		t.Fatal(err)
	}
	levels := []string{"none", "low", "high", "xhigh", "max"}
	h := gateway.New().Handler()
	get := func(path string, out any) {
		t.Helper()
		r := httptest.NewRequest(http.MethodGet, path, nil)
		r.Header.Set("Authorization", "Bearer "+gateway.TokenFor("opencode"))
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("%s: %d %s", path, w.Code, w.Body)
		}
		if err := json.Unmarshal(w.Body.Bytes(), out); err != nil {
			t.Fatalf("%s: %v: %s", path, err, w.Body)
		}
	}
	var plain struct {
		Data []struct {
			ID        string                    `json:"id"`
			Label     string                    `json:"magpie_label"`
			Reasoning bool                      `json:"reasoning"`
			Output    int                       `json:"max_output_tokens"`
			Levels    []struct{ Effort string } `json:"supported_reasoning_levels"`
		} `json:"data"`
	}
	get("/v1/models", &plain)
	seen := map[string]string{}
	for _, m := range plain.Data {
		if !strings.HasPrefix(m.ID, "synthetic/") {
			continue
		}
		seen[m.ID] = m.Label
		var got []string
		for _, lv := range m.Levels {
			got = append(got, lv.Effort)
		}
		if !m.Reasoning || m.Output != 65536 || !slices.Equal(got, levels) {
			t.Errorf("plain model: %+v", m)
		}
	}
	checkSyntheticGatewayNames(t, seen)
	var legacy struct {
		Models []struct {
			ID    string `json:"id"`
			Entry struct {
				Limit map[string]int `json:"limit"`
			} `json:"entry"`
		} `json:"models"`
	}
	get("/v1/models?format=opencode", &legacy)
	for _, m := range legacy.Models {
		if !strings.HasPrefix(m.ID, "synthetic/") {
			continue
		}
		context, hasContext := m.Entry.Limit["context"]
		if !hasContext || m.Entry.Limit["output"] != 65536 ||
			(m.ID == "synthetic/infinite:small:vision" && context != 0) {
			t.Errorf("OpenCode 1 model: %+v", m)
		}
	}
	var native struct {
		Models []struct {
			ID    string `json:"id"`
			Entry struct {
				Name     string               `json:"name"`
				Limit    struct{ Output int } `json:"limit"`
				Variants []struct {
					ID       string                           `json:"id"`
					Settings struct{ ReasoningEffort string } `json:"settings"`
				} `json:"variants"`
			} `json:"entry"`
		} `json:"models"`
	}
	get("/v1/models?format=opencode2", &native)
	seen = map[string]string{}
	for _, m := range native.Models {
		if !strings.HasPrefix(m.ID, "synthetic/") {
			continue
		}
		seen[m.ID] = m.Entry.Name
		var got []string
		for _, v := range m.Entry.Variants {
			got = append(got, v.ID)
			if v.Settings.ReasoningEffort != v.ID {
				t.Errorf("variant setting: %+v", v)
			}
		}
		if m.Entry.Limit.Output != 65536 || !slices.Equal(got, levels) {
			t.Errorf("OpenCode 2 model: %+v", m)
		}
	}
	checkSyntheticGatewayNames(t, seen)
}

func checkSyntheticGatewayNames(t *testing.T, seen map[string]string) {
	t.Helper()
	a, aok := seen["synthetic/infinite:small:vision"]
	b, bok := seen["synthetic/hf:deepseek-ai/DeepSeek-V4.1-Flash-1017"]
	if len(seen) != 2 || !aok || !bok || a == b || !strings.Contains(a, "Infinite") ||
		!strings.Contains(b, "DeepSeek V4.1 Flash 10/17") || strings.Contains(b, "Infinite") {
		t.Errorf("Synthetic identities and labels: %v", seen)
	}
}

type syntheticModelsTransport struct{ srv *httptest.Server }

func (s syntheticModelsTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	u := *r.URL
	u.Scheme, u.Host = "http", s.srv.Listener.Addr().String()
	r.URL = &u
	return http.DefaultTransport.RoundTrip(r)
}
