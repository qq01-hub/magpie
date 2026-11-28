package gateway

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/yetone/magpie/internal/provider"
)

const copilotReasoningText = "The model is checking the request."

func copilotReasoningTextAPI(t *testing.T, whole bool, sent *[]string) {
	t.Helper()
	var mu sync.Mutex
	copilotAPIFake(t, map[string]string{"gemini-4-hydrogen": `["/chat/completions"]`}, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		if r.URL.Path == "/models" {
			io.WriteString(w, `{"data":[]}`)
			return
		}
		if r.URL.Path != "/chat/completions" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		mu.Lock()
		*sent = append(*sent, string(b))
		mu.Unlock()
		if whole {
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, `{"id":"c1","object":"chat.completion","model":"gemini-4-hydrogen","choices":[{"index":0,"message":{"role":"assistant","content":"answer","reasoning_text":"`+copilotReasoningText+`","reasoning_opaque":"opaque-fixture"},"finish_reason":"stop"}],"usage":{"prompt_tokens":5,"completion_tokens":8,"reasoning_tokens":4}}`)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		io.WriteString(w, sse(
			`data: {"id":"c1","model":"gemini-4-hydrogen","choices":[{"index":0,"delta":{"role":"assistant","reasoning_text":"`+copilotReasoningText+`","reasoning_opaque":"opaque-fixture"}}]}`,
			`data: {"id":"c1","model":"gemini-4-hydrogen","choices":[{"index":0,"delta":{"content":"answer"}}]}`,
			`data: {"id":"c1","model":"gemini-4-hydrogen","choices":[{"index":0,"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":5,"completion_tokens":8,"reasoning_tokens":4}}`,
			`data: [DONE]`))
	})
}

// GitHub Copilot's Gemini 4 writes visible thoughts as reasoning_text. Both
// translated client protocols should receive them, while direct Chat keeps
// Copilot's fields intact; a whole upstream reply has the same behavior.
func TestCopilotGeminiReasoningTextThroughHTTP(t *testing.T) {
	for _, c := range []struct {
		name, path, streamBody, wholeBody, streamEvent string
	}{
		{
			name:        "Responses",
			path:        "/v1/responses",
			streamBody:  `{"model":"copilot/gemini-4-hydrogen","stream":true,"input":"check it"}`,
			wholeBody:   `{"model":"copilot/gemini-4-hydrogen","stream":false,"input":"check it"}`,
			streamEvent: `response.reasoning_summary_text.delta`,
		},
		{
			name:        "Messages",
			path:        "/v1/messages",
			streamBody:  `{"model":"copilot/gemini-4-hydrogen","stream":true,"max_tokens":1000,"messages":[{"role":"user","content":"check it"}]}`,
			wholeBody:   `{"model":"copilot/gemini-4-hydrogen","stream":false,"max_tokens":1000,"messages":[{"role":"user","content":"check it"}]}`,
			streamEvent: `"thinking_delta"`,
		},
	} {
		t.Run(c.name+" stream", func(t *testing.T) {
			var sent []string
			copilotReasoningTextAPI(t, false, &sent)
			code, body := post(t, c.path, c.streamBody)
			if code != http.StatusOK || !strings.Contains(body, "answer") || !strings.Contains(body, copilotReasoningText) || !strings.Contains(body, c.streamEvent) {
				t.Fatalf("status %d, translated stream lost reasoning or answer: %s", code, body)
			}
			if countReasoningDelta(c.path, body) != 1 {
				t.Errorf("reasoning was duplicated in the client stream: %s", body)
			}
			if len(sent) != 1 || !strings.Contains(sent[0], `"model":"gemini-4-hydrogen"`) {
				t.Fatalf("Copilot requests: %v", sent)
			}
		})
		t.Run(c.name+" whole client", func(t *testing.T) {
			var sent []string
			copilotReasoningTextAPI(t, false, &sent)
			code, body := post(t, c.path, c.wholeBody)
			if code != http.StatusOK || !strings.Contains(body, "answer") || !strings.Contains(body, copilotReasoningText) {
				t.Fatalf("status %d, translated whole reply lost reasoning or answer: %s", code, body)
			}
		})
	}

	for _, stream := range []bool{true, false} {
		name := "Chat whole"
		streamValue := "false"
		if stream {
			name, streamValue = "Chat stream", "true"
		}
		t.Run(name, func(t *testing.T) {
			var sent []string
			copilotReasoningTextAPI(t, !stream, &sent)
			body := `{"model":"copilot/gemini-4-hydrogen","stream":` + streamValue + `,"messages":[{"role":"user","content":"check it"}]}`
			code, got := post(t, "/v1/chat/completions", body)
			if code != http.StatusOK || !strings.Contains(got, "answer") || !strings.Contains(got, copilotReasoningText) {
				t.Fatalf("status %d, Copilot's chat fields did not reach the caller: %s", code, got)
			}
			if stream && (!strings.Contains(got, `"reasoning_text"`) || !strings.Contains(got, `"reasoning_opaque":"opaque-fixture"`)) {
				t.Errorf("Copilot's Chat fields were changed: %s", got)
			}
			if !stream && !strings.Contains(got, `"reasoning_opaque":"opaque-fixture"`) {
				t.Errorf("whole Copilot Chat fields were changed: %s", got)
			}
		})
	}
}

func TestCopilotGeminiReasoningTextIsThinkingForMetricsAndContinuation(t *testing.T) {
	chunk := []byte(`data: {"choices":[{"delta":{"reasoning_text":"` + copilotReasoningText + `"}}]}`)
	if content, text, reasoning, _ := timingKind(chunk); !content || text || !reasoning {
		t.Errorf("reasoning_text timing = content %v, text %v, reasoning %v", content, text, reasoning)
	}
	var tail chatTail
	tail.see([]byte(`{"choices":[{"delta":{"reasoning_text":"` + copilotReasoningText + `"}}]}`))
	if tail.void() || tail.said() {
		t.Errorf("reasoning-only chunk void=%v said=%v", tail.void(), tail.said())
	}
}

func TestCopilotGeminiReasoningTextWholeChatDecode(t *testing.T) {
	body := []byte(`{"id":"c1","object":"chat.completion","model":"gemini-4-hydrogen","choices":[{"index":0,"message":{"role":"assistant","content":"answer","reasoning_text":"` + copilotReasoningText + `"},"finish_reason":"stop"}]}`)
	whole := wholeEvents(provider.Chat, body)
	if len(whole) == 0 {
		t.Fatal("whole Chat reply was not recognized")
	}
	var got []Event
	d := &chatDecoder{}
	for _, e := range whole {
		b, err := json.Marshal(e.data)
		if err != nil {
			t.Fatal(err)
		}
		if err := d.decode(string(b), func(ev Event) { got = append(got, ev) }); err != nil {
			t.Fatal(err)
		}
	}
	var thought, text string
	for _, ev := range got {
		switch ev.Kind {
		case KThink:
			thought += ev.Text
		case KText:
			text += ev.Text
		}
	}
	if thought != copilotReasoningText || text != "answer" {
		t.Fatalf("whole Chat decoded thinking %q and text %q", thought, text)
	}
}

func countReasoningDelta(path, body string) int {
	n := 0
	for _, ev := range events(body) {
		switch path {
		case "/v1/responses":
			if ev["type"] == "response.reasoning_summary_text.delta" && ev["delta"] == copilotReasoningText {
				n++
			}
		case "/v1/messages":
			if ev["type"] == "content_block_delta" {
				delta, _ := ev["delta"].(map[string]any)
				if delta["type"] == "thinking_delta" && delta["thinking"] == copilotReasoningText {
					n++
				}
			}
		}
	}
	return n
}

// Copilot Gemini 4 spends its reply budget on thinking by default, and its
// API does not accept reasoning_effort none. Exercise the three client APIs
// through the gateway and inspect the request the local Copilot API receives.
func TestCopilotGemini4ThinkingRoomAndThinkOffOnWire(t *testing.T) {
	const model = "gemini-4-hydrogen"
	msg := `"messages":[{"role":"user","content":"check it"}]`
	cases := []struct {
		name, path, body, wantEffort string
		wantMax, userMax             int
	}{
		{"Responses default", "/v1/responses", `{"model":"copilot/gemini-4-hydrogen","stream":true,"max_output_tokens":1000,"input":"check it"}`, "", 25000, 0},
		{"Responses none", "/v1/responses", `{"model":"copilot/gemini-4-hydrogen","stream":true,"max_output_tokens":1000,"reasoning":{"effort":"none"},"input":"check it"}`, "low", 5096, 0},
		{"Responses high", "/v1/responses", `{"model":"copilot/gemini-4-hydrogen","stream":true,"max_output_tokens":1000,"reasoning":{"effort":"high"},"input":"check it"}`, "high", 1000, 0},
		{"Messages disabled", "/v1/messages", `{"model":"copilot/gemini-4-hydrogen","stream":true,"max_tokens":1000,"thinking":{"type":"disabled"},` + msg + `}`, "low", 5096, 0},
		{"Messages disabled overrides effort", "/v1/messages", `{"model":"copilot/gemini-4-hydrogen","stream":true,"max_tokens":1000,"thinking":{"type":"disabled"},"output_config":{"effort":"high"},` + msg + `}`, "low", 5096, 0},
		{"Messages adaptive high", "/v1/messages", `{"model":"copilot/gemini-4-hydrogen","stream":true,"max_tokens":1000,"thinking":{"type":"adaptive"},"output_config":{"effort":"high"},` + msg + `}`, "high", 1000, 0},
		{"Chat default", "/v1/chat/completions", `{"model":"copilot/gemini-4-hydrogen","stream":true,"max_tokens":1000,` + msg + `}`, "", 25000, 0},
		{"Chat none", "/v1/chat/completions", `{"model":"copilot/gemini-4-hydrogen","stream":true,"max_tokens":1000,"reasoning_effort":"none",` + msg + `}`, "low", 5096, 0},
		{"Chat high", "/v1/chat/completions", `{"model":"copilot/gemini-4-hydrogen","stream":true,"max_tokens":1000,"reasoning_effort":"high","thinking":{"type":"disabled"},` + msg + `}`, "high", 1000, 0},
		{"Chat output config high", "/v1/chat/completions", `{"model":"copilot/gemini-4-hydrogen","stream":true,"max_tokens":1000,"output_config":{"effort":"high"},` + msg + `}`, "", 1000, 0},
		{"Chat user max", "/v1/chat/completions", `{"model":"copilot/gemini-4-hydrogen","stream":true,"max_tokens":2000,` + msg + `}`, "", 1500, 1500},
		{"Chat user max with off", "/v1/chat/completions", `{"model":"copilot/gemini-4-hydrogen","stream":true,"max_tokens":2000,"reasoning_effort":"none",` + msg + `}`, "low", 1500, 1500},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var mu sync.Mutex
			var sent []string
			copilotAPIFake(t, map[string]string{model: `["/chat/completions"]`}, func(w http.ResponseWriter, r *http.Request) {
				b, _ := io.ReadAll(r.Body)
				if r.URL.Path == "/models" {
					io.WriteString(w, `{"data":[]}`)
					return
				}
				if r.URL.Path != "/chat/completions" {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				mu.Lock()
				sent = append(sent, string(b))
				mu.Unlock()
				w.Header().Set("Content-Type", "text/event-stream")
				io.WriteString(w, sse(
					`data: {"id":"budget","model":"gemini-4-hydrogen","choices":[{"index":0,"delta":{"role":"assistant","content":"answer"}}]}`,
					`data: {"id":"budget","model":"gemini-4-hydrogen","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}`,
					`data: [DONE]`))
			})
			if tc.userMax > 0 {
				if err := provider.SetModelOutput("copilot/"+model, tc.userMax); err != nil {
					t.Fatal(err)
				}
			}
			code, reply := post(t, tc.path, tc.body)
			if code != http.StatusOK || !strings.Contains(reply, "answer") {
				t.Fatalf("status %d, reply %s", code, reply)
			}
			mu.Lock()
			defer mu.Unlock()
			if len(sent) != 1 {
				t.Fatalf("Copilot received %d requests: %v", len(sent), sent)
			}
			var wire map[string]json.RawMessage
			if err := json.Unmarshal([]byte(sent[0]), &wire); err != nil {
				t.Fatal(err)
			}
			var gotMax int
			if err := json.Unmarshal(wire["max_tokens"], &gotMax); err != nil {
				t.Fatalf("Copilot request has no Chat max_tokens: %s", sent[0])
			}
			var gotEffort string
			_ = json.Unmarshal(wire["reasoning_effort"], &gotEffort)
			if gotMax != tc.wantMax || gotEffort != tc.wantEffort {
				t.Errorf("Copilot received max_tokens=%d effort=%q, want max_tokens=%d effort=%q: %s", gotMax, gotEffort, tc.wantMax, tc.wantEffort, sent[0])
			}
			if tc.name == "Chat output config high" && !strings.Contains(string(wire["output_config"]), `"effort":"high"`) {
				t.Errorf("Copilot's Chat request lost the caller's output_config effort: %s", sent[0])
			}
		})
	}
}

func TestCopilotGemini4ThinkingRoomStaysScoped(t *testing.T) {
	for _, tc := range []struct {
		name  string
		p     provider.Provider
		model string
	}{
		{"relay Gemini 4", provider.Provider{ID: "relay", Chat: "https://relay.example/v1"}, "gemini-4-hydrogen"},
		{"OpenRouter Gemini 4", provider.Provider{ID: "openrouter", Preset: "openrouter"}, "gemini-4-hydrogen"},
		{"Copilot Gemini 3", provider.Provider{ID: "copilot", Account: &provider.Account{Agent: "copilot"}}, "gemini-3.8-flash"},
		{"Copilot GPT", provider.Provider{ID: "copilot", Account: &provider.Account{Agent: "copilot"}}, "gpt-6-luna"},
		{"Copilot Gemini 40", provider.Provider{ID: "copilot", Account: &provider.Account{Agent: "copilot"}}, "gemini-40-hydrogen"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := []byte(`{"model":"gemini-4-hydrogen","max_tokens":1000,"messages":[{"role":"user","content":"check it"}]}`)
			if got := withThinkingRoom(tc.p, tc.model, provider.Chat, body); string(got) != string(body) {
				t.Fatalf("out-of-scope model's reply limit changed: %s", got)
			}
		})
	}
}
