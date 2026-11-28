package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/catalog"
)

// The two rows are zent00's Synthetic /openai/v1/models response in #7430.
// Infinite is a separate service mode, not another spelling to collapse.
func TestSyntheticFetchedModelsKeepTheirCapabilitiesAndIdentity(t *testing.T) {
	isolate(t)
	isolateFetchHome(t)
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
	body, err := os.ReadFile("testdata/synthetic-models-7430.json")
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
	http.DefaultClient.Transport = rewrite{srv}
	t.Cleanup(func() { http.DefaultClient.Transport = old })
	for _, cat := range []string{"", "synthetic"} {
		t.Run("catalog="+cat, func(t *testing.T) {
			p := Provider{ID: "synthetic", Name: "Synthetic", Chat: "https://api.synthetic.new/openai/v1", Key: "fixture-key", Catalog: cat}
			if _, err := p.Fetch(context.Background()); err != nil {
				t.Fatal(err)
			}
			entries := EntriesOf(p, p.Available())
			if len(entries) != 2 {
				t.Fatalf("entries: %+v", entries)
			}
			for i, id := range []string{"infinite:small:vision", "hf:deepseek-ai/DeepSeek-V4.1-Flash-1017"} {
				e := entries[i]
				if e.ID != "synthetic/"+id || e.Model != id {
					t.Errorf("entry identity: %+v", e)
				}
				if !e.Reasoning || !slices.Equal(e.Efforts, []string{"none", "low", "high", "xhigh", "max"}) {
					t.Errorf("%s reasoning = %v, efforts = %v", id, e.Reasoning, e.Efforts)
				}
				if e.Output != 65536 || !e.Images {
					t.Errorf("%s output = %d, images = %v", id, e.Output, e.Images)
				}
			}
			if entries[0].Name == entries[1].Name || !strings.Contains(entries[0].Name, "Infinite") || entries[1].Name != "DeepSeek V4.1 Flash 10/17" {
				t.Errorf("indistinguishable service modes: %q, %q", entries[0].Name, entries[1].Name)
			}
		})
	}
}
