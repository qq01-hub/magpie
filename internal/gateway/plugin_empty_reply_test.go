package gateway

import (
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/provider"
)

// A provider plugin can say a structurally complete stream had no content
// block with error.diagnostics.kind "empty", whatever its wording. The
// gateway must read it as its own empty reply: asked again once, not with
// a busy vendor's whole retry budget, and nobody rests.
const pluginEmptyBody = `{"type":"error","error":{"type":"upstream_error","message":"No content block was returned","diagnostics":{"kind":"empty","bytes":1350,"lastEventType":"message_stop","stopReason":"end_turn"}}}`

const pluginEmptyWhitespaceBody = `{ "type" : "error", "error" : { "type" : "upstream_error", "message" : "No content block was returned", "diagnostics" : { "kind" : "empty", "bytes" : 1350 } } }`

func TestPluginEmptyReplyIsAnEmptyReply(t *testing.T) {
	for _, x := range []struct {
		name, ctype, body, ask string
		code                   int
	}{
		{
			name:  "whole HTTP error",
			code:  502,
			ctype: "application/json",
			body:  pluginEmptyBody,
			ask:   `{"model":"a/m","messages":[{"role":"user","content":"hi"}]}`,
		},
		{
			name:  "whole HTTP error with JSON whitespace",
			code:  502,
			ctype: "application/json",
			body:  pluginEmptyWhitespaceBody,
			ask:   `{"model":"a/m","messages":[{"role":"user","content":"hi"}]}`,
		},
		{
			name:  "in-stream error frame",
			code:  200,
			ctype: "text/event-stream",
			body:  "event: error\ndata: " + pluginEmptyBody + "\n\n",
			ask:   `{"model":"a/m","stream":true,"messages":[{"role":"user","content":"hi"}]}`,
		},
		{
			name:  "in-stream error data without an event name",
			code:  200,
			ctype: "text/event-stream",
			body:  "data: " + pluginEmptyBody + "\n\n",
			ask:   `{"model":"a/m","stream":true,"messages":[{"role":"user","content":"hi"}]}`,
		},
		{
			name:  "in-stream error frame with JSON whitespace",
			code:  200,
			ctype: "text/event-stream",
			body:  "event: error\ndata: " + pluginEmptyWhitespaceBody + "\n\n",
			ask:   `{"model":"a/m","stream":true,"messages":[{"role":"user","content":"hi"}]}`,
		},
	} {
		t.Run(x.name, func(t *testing.T) {
			fresh(t)
			a := &scripted{replies: []reply{{x.code, x.ctype, x.body}}}
			scriptedOn(t, "a", provider.Chat, a)
			s := New()
			postAs(t, s, "", x.ask)
			if a.n != 2 {
				t.Fatalf("asked %d times, want 2: one retry, not a busy vendor's budget", a.n)
			}
			r := s.trace.routes[len(s.trace.routes)-1]
			if len(r.Tries) != 2 || r.Tries[0].Fail != failOther || r.Tries[0].Rest != nil ||
				!strings.Contains(r.Tries[0].Error, emptyReply) {
				t.Fatalf("tries: %+v", r.Tries)
			}
		})
	}
}

// An empty marker elsewhere in JSON doesn't classify the plugin's error.
// The diagnostics.kind value under error alone determines this retry budget.
func TestPluginOtherDiagnosticKindKeepsItsRetries(t *testing.T) {
	for _, x := range []struct {
		name, ctype, body string
		code              int
	}{
		{
			name:  "HTTP metadata marker",
			code:  502,
			ctype: "application/json",
			body:  `{"type":"error","error":{"message":"provider failed","diagnostics":{"kind":"busy"}},"metadata":{"diagnostics":{"kind":"empty"}}}`,
		},
		{
			name:  "SSE metadata marker",
			code:  200,
			ctype: "text/event-stream",
			body:  "event: error\ndata: {\"type\":\"error\",\"error\":{\"message\":\"provider failed\",\"diagnostics\":{\"kind\":\"busy\"}},\"metadata\":{\"diagnostics\":{\"kind\":\"empty\"}}}\n\n",
		},
		{
			name:  "SSE root diagnostics marker",
			code:  200,
			ctype: "text/event-stream",
			body:  "event: error\ndata: {\"type\":\"error\",\"diagnostics\":{\"kind\":\"empty\"},\"error\":{\"message\":\"provider failed\"}}\n\n",
		},
		{
			name:  "HTTP message text marker",
			code:  502,
			ctype: "application/json",
			body:  `{"type":"error","error":{"message":"text mentions diagnostics {\"kind\":\"empty\"}"}}`,
		},
		{
			name:  "HTTP root diagnostics marker",
			code:  502,
			ctype: "application/json",
			body:  `{"type":"error","diagnostics":{"kind":"empty"},"error":{"message":"provider failed"}}`,
		},
	} {
		t.Run(x.name, func(t *testing.T) {
			fresh(t)
			a := &scripted{replies: []reply{{x.code, x.ctype, x.body}}}
			scriptedOn(t, "a", provider.Chat, a)
			s := New()
			postAs(t, s, "", `{"model":"a/m","stream":true,"messages":[{"role":"user","content":"hi"}]}`)
			if a.n < 3 {
				t.Fatalf("asked %d times, want generic 5xx retries: %s", a.n, x.body)
			}
			r := s.trace.routes[len(s.trace.routes)-1]
			if strings.Contains(r.Tries[0].Error, emptyReply) {
				t.Fatalf("outside kind was read as error.diagnostics.kind: %+v", r.Tries[0])
			}
		})
	}
}

// A plain 502 keeps the generic retry budget: only the structured empty
// reply is narrowed to one retry.
func TestPlainBadGatewayKeepsItsRetries(t *testing.T) {
	fresh(t)
	a := &scripted{replies: []reply{{502, "", `{"error":{"message":"upstream failed"}}`}}}
	scriptedOn(t, "a", provider.Chat, a)
	s := New()
	postAs(t, s, "", `{"model":"a/m","messages":[{"role":"user","content":"hi"}]}`)
	if a.n < 3 {
		t.Fatalf("asked %d times, want at least 3: a plain 502 keeps lastRetries", a.n)
	}
	r := s.trace.routes[len(s.trace.routes)-1]
	if strings.Contains(r.Tries[0].Error, emptyReply) {
		t.Fatalf("a plain 502 read as an empty reply: %+v", r.Tries[0])
	}
}

// The plugin marker doesn't override a non-busy failure such as a rejected key.
func TestPluginEmptyMarkerKeepsAuthenticationFailure(t *testing.T) {
	fresh(t)
	a := &scripted{replies: []reply{{401, "application/json", `{"type":"error","error":{"type":"authentication_error","message":"invalid api key","diagnostics":{"kind":"empty"}}}`}}}
	scriptedOn(t, "a", provider.Chat, a)
	s := New()
	code, body := postAs(t, s, "", `{"model":"a/m","messages":[{"role":"user","content":"hi"}]}`)
	if code != 401 || a.n != 1 {
		t.Fatalf("status %d, asked %d times, want the original 401 once: %s", code, a.n, body)
	}
	r := s.trace.routes[len(s.trace.routes)-1]
	if len(r.Tries) != 1 || r.Tries[0].Fail != failOther || strings.Contains(r.Tries[0].Error, emptyReply) || !strings.Contains(r.Tries[0].Error, "invalid api key") {
		t.Fatalf("plugin marker overrode authentication failure: %+v", r.Tries)
	}
}

// A group still asks its next member. Only the last member's structured
// empty reply is asked once more in place, as the core's emptyReply is.
func TestPluginEmptyReplyStillFailsOverInAGroup(t *testing.T) {
	fresh(t)
	a := &scripted{replies: []reply{{502, "", pluginEmptyBody}}}
	b := &scripted{replies: []reply{{502, "", pluginEmptyBody}}}
	scriptedOn(t, "a", provider.Chat, a)
	scriptedOn(t, "b", provider.Chat, b)
	if err := provider.SaveGroup(provider.Group{Name: "G", Members: []string{"a/m", "b/m"}, Routing: provider.Ordered}); err != nil {
		t.Fatal(err)
	}
	s := New()
	postAs(t, s, "", `{"model":"group/g","messages":[{"role":"user","content":"hi"}]}`)
	if a.n != 1 || b.n != 2 {
		t.Fatalf("a asked %d, b asked %d, want 1 and 2: the next member, then the last once more", a.n, b.n)
	}
	r := s.trace.routes[len(s.trace.routes)-1]
	if len(r.Tries) != 3 {
		t.Fatalf("tries: %+v", r.Tries)
	}
	for _, try := range r.Tries {
		if try.Fail != failOther || try.Rest != nil || !strings.Contains(try.Error, emptyReply) {
			t.Fatalf("tries: %+v", r.Tries)
		}
	}
}
