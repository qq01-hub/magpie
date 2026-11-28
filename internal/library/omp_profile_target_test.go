package library

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/agent"
)

func TestOmpProfilesGetTheirOwnLibrary(t *testing.T) {
	h := sandbox(t)
	t.Setenv("PI_CONFIG_DIR", ".omp-work")
	t.Setenv("OMP_PROFILE", "")
	holdCLI(t, h, "omp")
	root := filepath.Join(h, ".omp-work")
	write(t, filepath.Join(root, "agent", "config.yml"), "modelRoles:\n  default: base/model\n")
	profile := filepath.Join(root, "profiles", "work", "agent")
	write(t, filepath.Join(profile, "config.yml"), "modelRoles:\n  default: user/model\n")
	write(t, filepath.Join(profile, "AGENTS.md"), "My own instructions.\n")
	lab := filepath.Join(root, "profiles", "lab", "agent")
	write(t, filepath.Join(lab, "config.yml"), "modelRoles:\n  default: lab/model\n")

	tg := targetByID("omp#work")
	if tg == nil || tg.Instructions != filepath.Join(profile, "AGENTS.md") || tg.MCP == nil || tg.MCP.Path != filepath.Join(profile, "mcp.json") || tg.Skills != filepath.Join(profile, "skills") {
		t.Fatalf("profile target: %+v", tg)
	}
	if tg.SharedToo || tg.readsSharedDir() {
		t.Fatalf("an older profile reads the shared skills folder: %+v", tg)
	}
	for _, kind := range []string{"instructions", "mcp", "skills"} {
		if id, err := Takes("omp#work", kind); id != "omp#work" || err != nil {
			t.Errorf("Takes %s = %q, %v", kind, id, err)
		}
	}

	if _, err := SaveServer("", Server{Name: "fetch", Transport: "stdio", Command: "uvx", Args: []string{"mcp-server-fetch"}, Agents: []string{"omp#work"}}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(read(t, filepath.Join(profile, "mcp.json")), "fetch") {
		t.Fatalf("profile MCP: %s", read(t, filepath.Join(profile, "mcp.json")))
	}
	if strings.Contains(read(t, filepath.Join(lab, "mcp.json")), "fetch") || strings.Contains(read(t, filepath.Join(root, "agent", "mcp.json")), "fetch") {
		t.Fatal("profile MCP server reached the default or another profile")
	}
	shared := "Follow the library."
	ok(t)(SaveInstructions(InstructionsChange{Shared: &shared, Agents: []string{"omp#work"}}))
	if got := read(t, filepath.Join(profile, "AGENTS.md")); !strings.Contains(got, "My own instructions.") || !strings.Contains(got, shared) {
		t.Fatalf("profile instructions did not keep the user's text: %s", got)
	}
	if strings.Contains(read(t, filepath.Join(lab, "AGENTS.md")), shared) || strings.Contains(read(t, filepath.Join(root, "agent", "AGENTS.md")), shared) {
		t.Fatal("profile instructions reached the default or another profile")
	}
	ok(t)(SaveInstructions(InstructionsChange{Agents: []string{}}))
	if got := read(t, filepath.Join(profile, "AGENTS.md")); strings.Contains(got, shared) || !strings.Contains(got, "My own instructions.") {
		t.Fatalf("taking library instructions back changed the user's text: %s", got)
	}

	version := "10.9.9"
	oldVersion := ompVersion
	ompVersion = func(a *agent.Agent) string {
		if a.ID == "omp" && a.Bin == "omp" {
			return version
		}
		return ""
	}
	t.Cleanup(func() { ompVersion = oldVersion })
	src := filepath.Join(h, "src/skills")
	skill(t, filepath.Join(src, "pdf"), "pdf", "Read PDFs")
	sharedSkills := filepath.Join(h, ".agents", "skills")
	skill(t, filepath.Join(sharedSkills, "orch"), "orch", "Orchestrate")
	ok(t)(ImportSkill("orch"))
	ok(t)(SkillAgents("orch", []string{"omp#work"}))
	ok(t)(InstallSkills(src, []string{"pdf"}, []string{"omp#work"}))
	if !ours(filepath.Join(profile, "skills", "pdf"), "pdf") {
		t.Fatal("profile skills folder has no pdf")
	}
	for _, dir := range []string{filepath.Join(root, "agent", "skills"), filepath.Join(lab, "skills")} {
		if _, err := os.Lstat(filepath.Join(dir, "pdf")); !os.IsNotExist(err) {
			t.Fatalf("profile skill reached the default or another profile: %s: %v", dir, err)
		}
	}
	if !ours(filepath.Join(profile, "skills", "orch"), "orch") {
		t.Fatal("an omp before 11.0.0 has no private link to shared skills")
	}
	version = "11.0.0"
	ok(t)(Sync())
	if _, err := os.Lstat(filepath.Join(profile, "skills", "orch")); !os.IsNotExist(err) {
		t.Fatalf("an omp at 11.0.0 has a duplicate private link: %v", err)
	}
	if !slices.Contains(alwaysOf(t, "orch"), "omp#work") {
		t.Fatalf("an omp at 11.0.0 isn't listed as reading shared skills: %v", alwaysOf(t, "orch"))
	}
	version = ""
	ok(t)(Sync())
	if !ours(filepath.Join(profile, "skills", "orch"), "orch") || slices.Contains(alwaysOf(t, "orch"), "omp#work") {
		t.Fatal("an omp of unknown version wasn't kept as before")
	}
}

func TestWSLOmpProfilesGetTheirOwnLibraryTarget(t *testing.T) {
	sandbox(t)
	root := t.TempDir()
	dhome := filepath.Join(root, "home", "me")
	write(t, filepath.Join(dhome, ".omp", "agent", "config.yml"), "{}\n")
	t.Cleanup(agent.FakeWSL(
		map[string]string{"Ubuntu": "home:/home/me\ndir:.omp\nver:omp omp/18.4.8\nprofile:omp:work\n"},
		map[string]string{"Ubuntu": root}),
	)

	tg := targetByID("omp#work@wsl:Ubuntu")
	if tg == nil || tg.MCP == nil || tg.MCP.Path != filepath.Join(dhome, ".omp", "profiles", "work", "agent", "mcp.json") || tg.Skills != filepath.Join(dhome, ".omp", "profiles", "work", "agent", "skills") || !tg.MCP.WSL || tg.MCP.Distro != "Ubuntu" || !tg.SharedToo {
		t.Fatalf("WSL profile target: %+v", tg)
	}
	if _, err := os.Stat(filepath.Join(dhome, ".omp", "profiles", "work", "agent")); !os.IsNotExist(err) {
		t.Fatalf("target inspection created profile files: %v", err)
	}
	if _, err := SaveServer("", Server{Name: "fetch", Transport: "stdio", Command: "uvx", Args: []string{"mcp-server-fetch"}, Agents: []string{"omp#work@wsl:Ubuntu"}}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(read(t, filepath.Join(dhome, ".omp", "profiles", "work", "agent", "mcp.json")), "fetch") || strings.Contains(read(t, filepath.Join(dhome, ".omp", "agent", "mcp.json")), "fetch") {
		t.Fatal("WSL profile MCP server did not stay in its own profile")
	}
	for _, kind := range []string{"instructions", "mcp", "skills"} {
		if id, err := Takes("omp#work@wsl:Ubuntu", kind); id != "omp#work@wsl:Ubuntu" || err != nil {
			t.Errorf("WSL Takes %s = %q, %v", kind, id, err)
		}
	}
	src := filepath.Join(home(), "src", "skills")
	skill(t, filepath.Join(src, "pdf"), "pdf", "Read PDFs")
	ok(t)(InstallSkills(src, []string{"pdf"}, []string{"omp#work@wsl:Ubuntu"}))
	profileSkill := filepath.Join(dhome, ".omp", "profiles", "work", "agent", "skills", "pdf")
	if !ours(profileSkill, "pdf") {
		t.Fatal("WSL profile skill was not installed into its own agent folder")
	}
	if _, err := os.Lstat(filepath.Join(dhome, ".omp", "agent", "skills", "pdf")); !os.IsNotExist(err) {
		t.Fatalf("WSL profile skill reached the default profile: %v", err)
	}
}
