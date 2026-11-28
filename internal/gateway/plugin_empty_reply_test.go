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
const pluginEmptyBody = `{"type":"error","error":{"type":"upstream_error","message":"the upstream answered an empty reply","diagnostics":{"kind":"empty","bytes":1350,"lastEventType":"message_stop","stopReason":"end_turn"}}}`

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
			name:  "in-stream error frame",
			code:  200,
			ctype: "text/event-stream",
			body:  "event: error\ndata: " + pluginEmptyBody + "\n\n",
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
