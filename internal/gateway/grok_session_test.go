package gateway

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// grokTurn is the headers grok 1.0.89 and 1.0.92 (grok-shell, headless)
// send an OpenAI-compatible endpoint with a request: the ones that name
// things, as they spell them. session and conv are x-grok-session-id and
// x-grok-conv-id, which a turn's requests send alike; the request sent
// before a session exists sends both, the request id and the agent id
// empty.
func grokTurn(session, conv string) http.Header {
	h := http.Header{}
	h.Set("User-Agent", "grok-shell/1.0.89 (macos; aarch64)")
	h.Set("x-grok-client-mode", "headless")
	h.Set("x-grok-client-version", "1.0.89")
	h.Set("x-grok-client-identifier", "grok-shell")
	h.Set("x-grok-conv-group-id", "b6170fa2-4ecd-4f0b-a318-c7d2a12fcef8")
	h.Set("x-grok-conv-id", conv)
	h.Set("x-grok-session-id", session)
	h.Set("x-grok-model-override", "fakeco/fake-1")
	if session == "" {
		h.Set("x-grok-req-id", "")
		h.Set("x-grok-agent-id", "")
		return h
	}
	h.Set("x-grok-req-id", "27d1c5a3-91be-4a70-88fd-0e442724fa88")
	h.Set("x-grok-agent-id", "9e0143bb-ace7-42dd-8163-5fd63461dd92")
	h.Set("x-grok-turn-idx", "1")
	return h
}

// grokOpens is the first user message of a Grok conversation as grok 1.0.92
// sends it (its rules shortened): context alone, the same for every
// conversation started in a folder on one day. What the user typed comes
// in a later user message of the same turn.
const grokOpens = "<user_info>\nOS Version: macos\nShell: /bin/zsh\nWorkspace Path: /Users/me/project\nToday's date: 2026-11-21\n</user_info>\n\n<rules>\nThe rules section has a number of possible rules/memories/context that you should consider.\n</rules>"

// grokAsks is the body of a Grok turn that asks query.
func grokAsks(t *testing.T, query string) []byte {
	t.Helper()
	b, err := json.Marshal(map[string]any{
		"model": "fakeco/fake-1", "stream": true,
		"messages": []map[string]string{
			{"role": "system", "content": "You are Grok."},
			{"role": "user", "content": grokOpens},
			{"role": "user", "content": "<system-reminder>\nMCP server connected:\n- plugin (1 tool)\n</system-reminder>"},
			{"role": "user", "content": "<user_query>\n" + query + "\n</user_query>"},
		},
		"tools": []map[string]any{{"type": "function", "function": map[string]any{"name": "run_terminal_cmd", "parameters": map[string]any{"type": "object"}}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// Grok Build names its session in x-grok-session-id, and magpie read none
// of its headers: a Grok conversation was known only by how it opens, and
// every Grok conversation opens with the same message (grokOpens). So all
// the conversations started in one folder on one day were one conversation
// (one account kept for them all, one id sent on as the conversation), a
// subagent was its lead, and a plugin's fetch was told no session at all.
// The header is read with the others now: a turn's requests are the
// session's, also after the session is resumed, a subagent's are its own.
// The summary Grok asks after a turn carries the session's id with a conv
// id of its own, on the turn's model: it is left as it was, told apart by
// how it opens, so its short answer is not kept as the conversation's. The
// request before a session, which sends the header empty, names none.
func TestGrokBuildNamesItsSession(t *testing.T) {
	const (
		lead     = "01f3b2c8-7d44-4e1a-9c60-5a12edf75ce8"
		other    = "5c9e7a10-22b3-4f6d-8e01-77c4d619a140"
		subagent = "e2a40d6f-0b19-47c3-b5a8-93f1e23a51c2"
		summary  = "turn-summary-13e02e3a-ea1e-46b8-9feb-0d3a5bea73a2"
	)
	turn := grokAsks(t, "fix the bug")
	summed := []byte(`{"model":"fakeco/fake-1","stream":false,"messages":[{"role":"user","content":"<user_message>\n<user_query>\nfix the bug\n</user_query>\n</user_message>\n\n<agent_reply>\ndone\n</agent_reply>"}]}`)

	for _, tc := range []struct {
		name string
		id   func(http.Header, []byte) string
	}{
		{"sessionOf", func(h http.Header, _ []byte) string { return sessionOf(h) }},
		{"conversationID", conversationID},
		{"affinityID", affinityID},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.id(grokTurn(lead, lead), turn); got != lead {
				t.Errorf("a turn of session %s is of %q", lead, got)
			}
			// another session in the same folder, asked something else: every
			// Grok conversation opens with the same message
			if got := tc.id(grokTurn(other, other), grokAsks(t, "write the changelog")); got != other {
				t.Errorf("another session, asked something else, is of %q, want %s", got, other)
			}
			// the session resumed, at its second turn: the same id again
			second := grokTurn(lead, lead)
			second.Set("x-grok-turn-idx", "2")
			if got := tc.id(second, grokAsks(t, "and the tests")); got != lead {
				t.Errorf("a resumed session's second turn is of %q, want %s", got, lead)
			}
			// the title asked with the whole conversation sends both ids alike
			// and no turn index: the session's
			title := grokTurn(lead, lead)
			title.Del("x-grok-turn-idx")
			if got := tc.id(title, turn); got != lead {
				t.Errorf("the title request of %s is of %q", lead, got)
			}
			// a subagent's requests carry an id of their own
			if got := tc.id(grokTurn(subagent, subagent), grokAsks(t, "list the tests")); got != subagent {
				t.Errorf("a subagent's turn is of %q, want its own %s", got, subagent)
			}
		})
	}

	// the summary asked after a turn: the session's id, a conv id of its own.
	// It names no session and is the conversation its opening makes it, as
	// before, so it does not write over what is kept for the session
	sum := grokTurn(lead, summary)
	sum.Del("x-grok-turn-idx")
	if got := sessionOf(sum); got != "" {
		t.Errorf("the summary after a turn names session %q", got)
	}
	if sessionNamed(sum) {
		t.Errorf("the summary after a turn counts as naming a session")
	}
	for name, id := range map[string]func(http.Header, []byte) string{"conversationID": conversationID, "affinityID": affinityID} {
		if got, want := id(sum, summed), conversationID(http.Header{}, summed); got != want || !strings.HasPrefix(got, "magpie-") {
			t.Errorf("the summary after a turn: %s %q, want its opening's %q", name, got, want)
		}
	}

	// the request before a session exists sends the header empty: it names
	// no session, and is told apart by how it opens, as before
	before := grokTurn("", "")
	if got := sessionOf(before); got != "" {
		t.Errorf("an empty x-grok-session-id names session %q", got)
	}
	plain := http.Header{}
	if got, want := conversationID(before, turn), conversationID(plain, turn); got != want || !strings.HasPrefix(got, "magpie-") {
		t.Errorf("an empty x-grok-session-id: conversation %q, want the opening's %q", got, want)
	}

	// an agent's header that came first in the list still does, and the
	// caller's own X-Magpie-Session is still the session
	both := grokTurn(lead, lead)
	both.Set("x-session-id", "ses_1")
	if got := sessionOf(both); got != "ses_1" {
		t.Errorf("x-session-id beside Grok's: session %q, want ses_1", got)
	}
	both.Set(SessionHeader, "mine")
	if got := sessionOf(both); got != "mine" {
		t.Errorf("X-Magpie-Session beside Grok's: session %q, want mine", got)
	}
}
