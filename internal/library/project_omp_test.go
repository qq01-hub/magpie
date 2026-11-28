package library

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOmpProjectSkillsUseItsNativeFolder(t *testing.T) {
	h, proj := projectLib(t)
	custom := filepath.Join(h, "custom-omp")
	t.Setenv("PI_CONFIG_DIR", custom)
	user := filepath.Join(custom, "agent", "config.yml")
	write(t, user, "skills:\n  enableAgentsProject: false\n")
	if got := ProjectSkillsDir("omp"); got != ".omp/skills" {
		t.Fatalf("omp's native project skills folder: %q", got)
	}
	write(t, filepath.Join(proj, ".omp", "skills", "mine", "SKILL.md"), "my own skill\n")
	ok(t)(ProjectSkill(proj, "pdf", []string{"omp", "codex"}))
	for _, rel := range []string{".omp/skills/pdf", ".agents/skills/pdf"} {
		p := filepath.Join(proj, rel)
		if !isLink(t, p) || !ours(p, "pdf") {
			t.Errorf("%s isn't linked to the library skill", rel)
		}
		if _, err := os.Stat(filepath.Join(p, "SKILL.md")); err != nil {
			t.Error(err)
		}
	}
	if ex := excluded(t, proj); !strings.Contains(ex, "/.omp/skills/pdf\n") || strings.Contains(ex, "/.omp/skills/mine\n") {
		t.Errorf("project exclusion: %s", ex)
	}
	ok(t)(ProjectSkill(proj, "pdf", []string{"codex"}))
	gone(t, filepath.Join(proj, ".omp", "skills", "pdf"))
	if got := read(t, filepath.Join(proj, ".omp", "skills", "mine", "SKILL.md")); got != "my own skill\n" {
		t.Errorf("the user's native skill changed: %q", got)
	}
	there(t, filepath.Join(proj, ".agents", "skills", "pdf"))
	if got := read(t, user); got != "skills:\n  enableAgentsProject: false\n" {
		t.Errorf("custom user settings changed: %q", got)
	}
	gone(t, filepath.Join(custom, "agent", "skills", "pdf"))
}

func TestOmpProjectMCPUsesItsNativeFile(t *testing.T) {
	h, proj := mcpProject(t)
	if got := ProjectMCPFile("omp"); got != ".omp/mcp.json" {
		t.Fatalf("omp's native project MCP file: %q", got)
	}
	custom := filepath.Join(h, "custom-omp")
	t.Setenv("PI_CONFIG_DIR", custom)
	user := filepath.Join(custom, "agent", "mcp.json")
	write(t, user, "{\"mcpServers\":{\"user\":{\"command\":\"user-server\"}}}\n")
	rootBefore := read(t, filepath.Join(proj, ".mcp.json"))
	path := filepath.Join(proj, ".omp", "mcp.json")
	write(t, path, "{\n // mine\n \"other\": true,\n \"mcpServers\": {\"own\": {\"command\": \"own-server\"}}\n}\n")
	ok(t)(SaveServer("", Server{Name: "events", Transport: "sse", URL: "https://events.example/sse"}))
	for _, name := range []string{"fs", "docs", "events"} {
		r := ok(t)(ProjectServer(proj, name, []string{"omp"}))
		if len(r.Problems) != 0 {
			t.Fatalf("%s: %+v", name, r.Problems)
		}
	}
	es := servers(t, path, "mcpServers")
	if e, _ := es["fs"].(map[string]any); e["command"] != "npx" || e["type"] != nil {
		t.Errorf("omp stdio entry: %v", e)
	}
	for _, e := range []struct{ name, transport, url string }{{"docs", "http", "https://docs.example/mcp"}, {"events", "sse", "https://events.example/sse"}} {
		if got, _ := es[e.name].(map[string]any); got["type"] != e.transport || got["url"] != e.url {
			t.Errorf("omp %s entry: %v", e.name, got)
		}
	}
	if es["own"] == nil || jsonOf(t, path)["other"] != true || !strings.Contains(read(t, path), "// mine") {
		t.Errorf("the user's native config changed: %s", read(t, path))
	}
	if ex := excluded(t, proj); strings.Contains(ex, "/.omp/mcp.json\n") {
		t.Errorf("the user's file was excluded: %s", ex)
	}
	ok(t)(RemoveProject(proj, false))
	if es = servers(t, path, "mcpServers"); len(es) != 1 || es["own"] == nil {
		t.Errorf("native config after removal: %v", es)
	}
	if read(t, filepath.Join(proj, ".mcp.json")) != rootBefore || read(t, user) != "{\"mcpServers\":{\"user\":{\"command\":\"user-server\"}}}\n" {
		t.Error("a root compatibility file or user-wide config changed")
	}
}

func TestOmpProjectPluginHasSkillsAndMCP(t *testing.T) {
	_, proj, other := pluginProjects(t)
	r := ok(t)(ProjectPlugin(proj, "acme-kit", []string{"omp"}))
	if len(r.Problems) != 0 {
		t.Fatalf("omp project plugin: %+v", r.Problems)
	}
	there(t, filepath.Join(proj, ".omp", "skills", "tdd", "SKILL.md"))
	if es := servers(t, filepath.Join(proj, ".omp", "mcp.json"), "mcpServers"); es["acme-db"] == nil {
		t.Errorf("plugin server missing: %v", es)
	}
	gone(t, filepath.Join(other, ".omp"))
	ok(t)(ProjectPlugin(proj, "acme-kit", nil))
	gone(t, filepath.Join(proj, ".omp"))
}
