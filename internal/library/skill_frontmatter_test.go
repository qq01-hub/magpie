package library

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

// An unquoted colon in a plain description is invalid YAML. Keep the skill
// visible by its folder name and expose the parse failure in the Library
// view, while a neighboring hand-authored valid skill keeps its metadata.
func TestMalformedSkillFrontMatterStaysVisible(t *testing.T) {
	sandbox(t)
	src := filepath.Join(t.TempDir(), "skills")
	write(t, filepath.Join(src, "broken", "SKILL.md"), "---\nname: vendor-name\ndisable-model-invocation: true\ndescription: Use it when: something\n---\n\nInstructions remain available.\n")
	write(t, filepath.Join(src, "space skill", "SKILL.md"), "---\nname: vendor-space-name\ndescription: Use it when: something\n---\n\nInstructions remain available.\n")
	write(t, filepath.Join(src, "valid", "SKILL.md"), "---\nname: valid-skill\ndescription: 'Use it when: something'\ndisable-model-invocation: true\n---\n\nHand-written instructions.\n")

	p, err := ProbeSkills(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Candidates) != 3 {
		t.Fatalf("candidates: %+v", p.Candidates)
	}
	candidates := map[string]Candidate{}
	for _, candidate := range p.Candidates {
		candidates[candidate.Path] = candidate
	}
	if malformed := candidates["broken"]; malformed.Name != "broken" || malformed.DisplayName != "broken" || malformed.Description != "" || malformed.FrontmatterError == "" {
		t.Fatalf("malformed repository candidate = %+v", malformed)
	}
	if spaced := candidates["space skill"]; spaced.Name != "space-skill" || spaced.DisplayName != "space skill" || spaced.FrontmatterError == "" {
		t.Fatalf("malformed candidate with a spaced folder name = %+v", spaced)
	}
	if valid := candidates["valid"]; valid.Name != "valid-skill" || valid.Description != "Use it when: something" || valid.FrontmatterError != "" {
		t.Fatalf("valid repository candidate = %+v", valid)
	}
	if err := func() error {
		_, err := InstallSkills(src, []string{"broken", "space skill", "valid"}, []string{"claude"})
		return err
	}(); err != nil {
		t.Fatal(err)
	}

	v, err := Read(nil)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]SkillView{}
	for _, skill := range v.Skills {
		got[skill.Name] = skill
	}
	broken, ok := got["broken"]
	if !ok {
		t.Fatalf("malformed skill was not installed under its folder name: %+v", v.Skills)
	}
	if broken.Description != "" || broken.Manual {
		t.Errorf("malformed metadata was partly trusted: %+v", broken)
	}
	valid, ok := got["valid-skill"]
	if !ok || valid.Description != "Use it when: something" || !valid.Manual {
		t.Errorf("valid neighboring skill changed: %+v", valid)
	}
	spaced, ok := got["space-skill"]
	if !ok || spaced.FrontmatterError == "" || spaced.Description != "" || spaced.Manual {
		t.Errorf("malformed skill from a spaced folder lost its sanitized install name or warning: %+v", spaced)
	}

	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"frontmatterError":`) || !strings.Contains(string(b), "mapping values are not allowed") {
		t.Errorf("Library view does not explain the malformed front matter: %s", b)
	}

	// An agent's own manually written skill follows the same visible fallback.
	write(t, filepath.Join(home(), ".claude", "skills", "own-broken", "SKILL.md"), "---\nname: other-name\ndisable-model-invocation: true\ndescription: Use it when: something\n---\n")
	v, err = Read(nil)
	if err != nil {
		t.Fatal(err)
	}
	b, err = json.Marshal(v.FoundSkills)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"name":"own-broken"`) || !strings.Contains(string(b), `"frontmatterError":`) || !strings.Contains(string(b), "mapping values are not allowed") {
		t.Errorf("agent's own malformed skill is not named and explained: %s", b)
	}
}
