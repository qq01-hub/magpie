package catalog

import (
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
)

func TestFetchSyntheticMetadataPrecedenceAndFallback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":[
		  {"id":"infinite:small:vision","display_name":"DeepSeek","supported_reasoning_levels":["low"],"reasoning_parameters":{"efforts":["none","high","xhigh","high"]},"max_output_tokens":1000,"max_output_length":65536},
		  {"id":"switch","supported_features":["tools","reasoning"]},
		  {"id":"legacy","supported_reasoning_levels":["low","high"],"max_output_tokens":2000},
		  {"id":"odd","reasoning_parameters":false,"supported_features":17,"max_output_length":"lots","supported_reasoning_levels":["medium"],"max_output_tokens":3000},
		  {"id":"empty","reasoning_parameters":{"efforts":[]},"max_output_length":-1,"supported_reasoning_levels":["low"],"max_output_tokens":4000},
		  {"id":"infinite:named","display_name":"DeepSeek Infinite"}
		]}`))
	}))
	defer srv.Close()
	old := http.DefaultClient.Transport
	http.DefaultClient.Transport = toServer{srv}
	t.Cleanup(func() { http.DefaultClient.Transport = old })
	for _, tc := range []struct {
		base string
		own  bool
	}{
		{"https://api.synthetic.new/openai/v1", true},
		{"https://api.synthetic.new/anthropic", true},
		{"https://API.SYNTHETIC.NEW/openai/v1", true},
		{"https://relay.example/openai/v1", false},
		{"https://api.synthetic.new.example/openai/v1", false},
	} {
		t.Run(tc.base, func(t *testing.T) {
			ms, _, err := FetchAt(context.Background(), tc.base, "fixture-key", false, nil)
			if err != nil || len(ms) != 6 {
				t.Fatalf("models: %+v, %v", ms, err)
			}
			wantLevels, wantOutput, wantName := []string{"low"}, 1000, "DeepSeek"
			if tc.own {
				wantLevels, wantOutput, wantName = []string{"none", "high", "xhigh"}, 65536, "DeepSeek (Infinite)"
			}
			if !slices.Equal(ms[0].Efforts, wantLevels) || ms[0].Output != wantOutput || ms[0].Name != wantName {
				t.Errorf("preferred metadata: %+v", ms[0])
			}
			if ms[1].Reasoning != tc.own || len(ms[1].Efforts) != 0 {
				t.Errorf("reasoning switch: %+v", ms[1])
			}
			for i, want := range []struct {
				levels []string
				output int
			}{
				{[]string{"low", "high"}, 2000},
				{[]string{"medium"}, 3000},
				{[]string{"low"}, 4000},
			} {
				if m := ms[i+2]; !slices.Equal(m.Efforts, want.levels) || m.Output != want.output {
					t.Errorf("fallback: %+v", m)
				}
			}
			if ms[5].Name != "DeepSeek Infinite" {
				t.Errorf("already named mode changed: %+v", ms[5])
			}
		})
	}
}

func TestDecorateKeepsLiveReasoningLevels(t *testing.T) {
	known := []Model{{ID: "model", Efforts: []string{"low", "high"}, Provider: "vendor"}}
	live := []Model{{ID: "model", Efforts: []string{"none", "xhigh", "max"}}}
	got := Decorate(live, known)
	if !slices.Equal(got[0].Efforts, []string{"none", "xhigh", "max"}) {
		t.Errorf("live levels overwritten: %v", got[0].Efforts)
	}
	live[0].Efforts = nil
	if got := Decorate(live, known); !slices.Equal(got[0].Efforts, []string{"low", "high"}) {
		t.Errorf("missing levels not filled: %v", got[0].Efforts)
	}
}
