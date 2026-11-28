package gateway

import (
	"context"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/provider"
)

// Riablo's four groups share every model of the last, Review (#7221).
// The request still routes through Review, and its trace must show that
// group and the member's chosen place rather than its first sibling's.
func TestTraceKeepsACompletelySharedRuleGroup(t *testing.T) {
	fresh(t)
	for _, p := range []struct {
		id     string
		models []string
	}{
		{"claude-relay", []string{"claude-sonnet-5-5", "claude-haiku-5-5", "claude-opus-5-5"}},
		{"codex-relay", []string{"gpt-6.2-sol", "gpt-6.1-terra", "gpt-6.2-astra"}},
		{"deepseek", []string{"deepseek-flash"}},
		{"openrouter", []string{"deepseek/deepseek-v4.1-flash-1017"}},
	} {
		serveOn(t, p.id, "k"+p.id, p.models, &keyed{})
	}
	classifier := &clsUp{answer: "4"}
	up := httptest.NewServer(classifier)
	t.Cleanup(up.Close)
	if err := provider.Save(provider.Provider{ID: "classifier", Name: "Classifier", Key: "k", Models: []string{"cls"}, Chat: up.URL + "/v1"}); err != nil {
		t.Fatal(err)
	}
	for _, g := range []provider.Group{
		{ID: "task-mid", Name: "Task mid", Members: []string{"claude-relay/claude-sonnet-5-5", "codex-relay/gpt-6.2-sol", "deepseek/deepseek-flash:max"}, Routing: provider.Ordered},
		{ID: "task-low", Name: "Task low", Members: []string{"deepseek/deepseek-flash", "claude-relay/claude-haiku-5-5", "codex-relay/gpt-6.1-terra", "openrouter/deepseek/deepseek-v4.1-flash-1017"}, Routing: provider.Ordered},
		{ID: "task-high", Name: "Task high", Members: []string{"claude-relay/claude-opus-5-5", "codex-relay/gpt-6.2-astra", "deepseek/deepseek-flash:max"}, Routing: provider.Ordered},
		{ID: "review", Name: "Review", Members: []string{"codex-relay/gpt-6.2-sol", "deepseek/deepseek-flash:max", "claude-relay/claude-opus-5-5"}, Routing: provider.Ordered},
		{ID: "task", Name: "Task", Members: []string{"group/task-mid", "group/task-low", "group/task-high", "group/review"}, Routing: provider.Ordered, Classifier: "classifier/cls", Rules: []provider.Rule{
			{Use: "group/task-mid", Intent: "coding"},
			{Use: "group/task-low", Intent: "simple questions"},
			{Use: "group/task-high", Intent: "hard problems"},
			{Use: "group/review", Intent: "code review"},
		}},
	} {
		if err := provider.SaveGroup(g); err != nil {
			t.Fatal(err)
		}
	}
	s := New()
	code, body := postAs(t, s, "review", `{"model":"group/task","messages":[{"role":"user","content":"review this code"}]}`)
	if code != 200 || !strings.Contains(body, "from kcodex-relay") {
		t.Fatalf("%d %s", code, body)
	}
	state := s.Trace(context.Background(), 0, 0)
	var route *Route
	for i := range state.Routes {
		if state.Routes[i].Model == "group/task" {
			route = &state.Routes[i]
		}
	}
	if route == nil || route.Group == nil || route.Rule == nil || route.Rule.N != 4 || route.Rule.Use != "group/review" {
		t.Fatalf("route %+v", route)
	}
	var subs []string
	for _, sub := range route.Group.Subs {
		subs = append(subs, sub.ID+"@"+sub.In)
	}
	if want := []string{"task-mid@task", "task-low@task", "task-high@task", "review@task"}; !slices.Equal(subs, want) {
		t.Errorf("subgroups %v, want %v", subs, want)
	}
	if len(route.Group.Members) != 9 || len(route.Order) != 9 {
		t.Fatalf("shared models duplicated: members=%v order=%+v", route.Group.Members, route.Order)
	}
	wantLead := []string{"codex-relay/gpt-6.2-sol", "deepseek/deepseek-flash", "claude-relay/claude-opus-5-5"}
	for i, want := range wantLead {
		w := route.Order[i]
		if got := w.Provider + "/" + w.Model; got != want || !slices.Equal(w.Via, []string{"review"}) {
			t.Errorf("order %d: %s via %v, want %s via review", i, got, w.Via, want)
		}
	}
}

func TestTraceSharedGroupKeepsItsNestedRulePath(t *testing.T) {
	fresh(t)
	for _, p := range []struct{ id, model string }{{"a", "alpha"}, {"b", "beta"}, {"c", "gamma"}} {
		serveOn(t, p.id, "k"+p.id, []string{p.model}, &keyed{})
	}
	for _, g := range []provider.Group{
		{ID: "first", Name: "First", Members: []string{"a/alpha", "b/beta"}, Routing: provider.Ordered},
		{ID: "inner", Name: "Inner", Members: []string{"b/beta", "a/alpha"}, Routing: provider.Ordered, Rules: []provider.Rule{{Use: "a/alpha", Tokens: 1}}},
		{ID: "review", Name: "Review", Members: []string{"group/inner"}, Routing: provider.Ordered, Rules: []provider.Rule{{Use: "group/inner", Tokens: 1}}},
		{ID: "outer", Name: "Outer", Members: []string{"group/first", "group/review", "c/gamma"}, Routing: provider.Ordered, Rules: []provider.Rule{{Use: "group/review", Tokens: 1}}},
	} {
		if err := provider.SaveGroup(g); err != nil {
			t.Fatal(err)
		}
	}
	s := New()
	code, body := postAs(t, s, "nested", `{"model":"group/outer","messages":[{"role":"user","content":"review this code"}]}`)
	if code != 200 || !strings.Contains(body, "from ka") {
		t.Fatalf("%d %s", code, body)
	}
	r := lastRoute(s)
	if len(r.Nested) != 2 || r.Nested[0].Group != "review" || r.Nested[1].Group != "inner" {
		t.Fatalf("nested %+v", r.Nested)
	}
	var subs []string
	for _, sub := range r.Group.Subs {
		subs = append(subs, sub.ID+"@"+sub.In)
	}
	if want := []string{"first@outer", "review@outer", "inner@review"}; !slices.Equal(subs, want) {
		t.Errorf("subgroups %v, want %v", subs, want)
	}
	for i, want := range []string{"a", "b", "c"} {
		w := r.Order[i]
		via := []string(nil)
		if i < 2 {
			via = []string{"review", "inner"}
		}
		if w.Provider != want || !slices.Equal(w.Via, via) {
			t.Errorf("order %d: %s via %v, want %s via %v", i, w.Provider, w.Via, want, via)
		}
	}
}
