package gateway

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/provider"
)

// geminiReply is a generateContent reply's first candidate and its usage.
type geminiReply struct {
	Candidates []struct {
		Content struct {
			Role  string `json:"role"`
			Parts []struct {
				Text    string `json:"text"`
				Thought bool   `json:"thought"`
			} `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	UsageMetadata map[string]any `json:"usageMetadata"`
}

// Google's API gives a reply's thought parts only to a request whose
// thinkingConfig says includeThoughts: true. A Gemini client that thinks
// without asking for them (a budget alone, or no thinkingConfig at all) is
// given the whole reply's answer alone, with the thoughts still counted in
// usageMetadata; one that asks gets them. That holds however the thinking
// came: a Chat upstream's reasoning_content, or a Messages upstream's
// thinking block, which comes signed.
func TestGeminiThoughtsOnlyWhenIncluded(t *testing.T) {
	for _, up := range []struct {
		proto   provider.Protocol
		reply   string
		counted any // usageMetadata's thoughtsTokenCount
	}{
		{provider.Chat, sse(
			`data: {"id":"c1","model":"m1","choices":[{"delta":{"role":"assistant","reasoning_content":"think"}}]}`,
			`data: {"id":"c1","choices":[{"delta":{"content":"Sure"}}]}`,
			`data: {"id":"c1","choices":[{"delta":{},"finish_reason":"stop"}]}`,
			`data: {"id":"c1","choices":[],"usage":{"prompt_tokens":10,"completion_tokens":5,"completion_tokens_details":{"reasoning_tokens":2}}}`,
			`data: [DONE]`), float64(2)},
		{provider.Anthropic, sse(
			`event: message_start`+"\n"+`data: {"type":"message_start","message":{"id":"msg_1","model":"m1","usage":{"input_tokens":10}}}`,
			`event: content_block_start`+"\n"+`data: {"type":"content_block_start","index":0,"content_block":{"type":"thinking","thinking":"","signature":""}}`,
			`event: content_block_delta`+"\n"+`data: {"type":"content_block_delta","index":0,"delta":{"type":"thinking_delta","thinking":"think"}}`,
			`event: content_block_delta`+"\n"+`data: {"type":"content_block_delta","index":0,"delta":{"type":"signature_delta","signature":"CAQSpQ8KEAgTGAI4AUIIdGhp"}}`,
			`event: content_block_stop`+"\n"+`data: {"type":"content_block_stop","index":0}`,
			`event: content_block_start`+"\n"+`data: {"type":"content_block_start","index":1,"content_block":{"type":"text","text":""}}`,
			`event: content_block_delta`+"\n"+`data: {"type":"content_block_delta","index":1,"delta":{"type":"text_delta","text":"Sure"}}`,
			`event: content_block_stop`+"\n"+`data: {"type":"content_block_stop","index":1}`,
			`event: message_delta`+"\n"+`data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":5,"output_tokens_details":{"thinking_tokens":2}}}`,
			`event: message_stop`+"\n"+`data: {"type":"message_stop"}`), float64(2)},
	} {
		setup(t, up.proto, &fake{t: t, reply: up.reply})
		for _, c := range []struct {
			name, gen string
			thoughts  bool
		}{
			{"no generationConfig", "", false},
			{"budget alone", `,"generationConfig":{"thinkingConfig":{"thinkingBudget":-1}}`, false},
			{"includeThoughts false", `,"generationConfig":{"thinkingConfig":{"thinkingBudget":1024,"includeThoughts":false}}`, false},
			{"includeThoughts true", `,"generationConfig":{"thinkingConfig":{"thinkingBudget":-1,"includeThoughts":true}}`, true},
		} {
			code, body := post(t, "/v1beta/models/fake/m1:generateContent", `{"contents":[{"role":"user","parts":[{"text":"hi"}]}]`+c.gen+`}`)
			var reply geminiReply
			if err := json.Unmarshal([]byte(body), &reply); err != nil || code != 200 || len(reply.Candidates) != 1 {
				t.Fatalf("%s, %s: %d %s (%v)", up.proto, c.name, code, body, err)
			}
			var thought, text string
			for _, p := range reply.Candidates[0].Content.Parts {
				if p.Thought {
					thought += p.Text
				} else {
					text += p.Text
				}
			}
			if text != "Sure" {
				t.Errorf("%s, %s: answer %q, want Sure: %s", up.proto, c.name, text, body)
			}
			if want := map[bool]string{true: "think"}[c.thoughts]; thought != want {
				t.Errorf("%s, %s: thought parts %q, want %q: %s", up.proto, c.name, thought, want, body)
			}
			if got := reply.UsageMetadata["thoughtsTokenCount"]; got != up.counted {
				t.Errorf("%s, %s: usageMetadata %v, want thoughtsTokenCount %v", up.proto, c.name, reply.UsageMetadata, up.counted)
			}
		}
	}
}

// A reply that only thought (its whole say in reasoning, no text) reaches
// a Gemini client that didn't ask for its thoughts as Google's own does:
// one candidate of the model's, with no parts to show and its finish, the
// thoughts counted in usageMetadata. Never an error, and no thought part.
func TestGeminiThoughtsOnlyReplyKeepsItsCandidate(t *testing.T) {
	setup(t, provider.Chat, &fake{t: t, reply: sse(
		`data: {"id":"c1","model":"m1","choices":[{"delta":{"role":"assistant","reasoning_content":"think"}}]}`,
		`data: {"id":"c1","choices":[{"delta":{},"finish_reason":"stop"}]}`,
		`data: {"id":"c1","choices":[],"usage":{"prompt_tokens":10,"completion_tokens":2,"completion_tokens_details":{"reasoning_tokens":2}}}`,
		`data: [DONE]`)})
	code, body := post(t, "/v1beta/models/fake/m1:generateContent", `{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}`)
	var reply geminiReply
	if err := json.Unmarshal([]byte(body), &reply); err != nil || code != 200 || len(reply.Candidates) != 1 {
		t.Fatalf("%d %s (%v)", code, body, err)
	}
	if c := reply.Candidates[0]; c.Content.Role != "model" || len(c.Content.Parts) != 0 || c.FinishReason != "STOP" ||
		reply.UsageMetadata["thoughtsTokenCount"] != float64(2) {
		t.Errorf("a reply that only thought: %s", body)
	}
}

// A stream (streamGenerateContent) still carries a reply's thought parts
// to a client that didn't ask for them, which Google's doesn't: the loop
// guard reads a stream as the client is sent it (holdWriter.Write), and
// with them left out a looping model's reasoning would go unread
// (Request.hideThoughts). Leaving them out of a stream needs the guard
// fed before they are, a change of its own; until then this says so.
func TestGeminiStreamStillCarriesThoughts(t *testing.T) {
	setup(t, provider.Chat, &fake{t: t, reply: sse(
		`data: {"id":"c1","model":"m1","choices":[{"delta":{"role":"assistant","reasoning_content":"think"}}]}`,
		`data: {"id":"c1","choices":[{"delta":{"content":"Sure"}}]}`,
		`data: {"id":"c1","choices":[{"delta":{},"finish_reason":"stop"}]}`,
		`data: [DONE]`)})
	code, body := post(t, "/v1beta/models/fake/m1:streamGenerateContent?alt=sse", `{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}`)
	if code != 200 || !strings.Contains(body, `{"text":"think","thought":true}`) || !strings.Contains(body, `{"text":"Sure"}`) {
		t.Errorf("the stream: %d %s", code, body)
	}
}
