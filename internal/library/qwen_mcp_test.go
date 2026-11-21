package library

import (
	"encoding/json"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/tidwall/jsonc"
)

// Qwen Code keeps its user-wide MCP servers in mcpServers of the
// settings.json magpie sets its models in, as `qwen mcp add --scope user`
// writes them, and the library had no place for them: `magpie library mcp
// agents <server> qwen` answered "Qwen Code has no user-wide place for MCP
// servers that magpie knows of", so a server every other agent was given
// had to be added to Qwen Code by hand. The entries are Gemini CLI's
// (Qwen Code is its fork): httpUrl for a streamable server, url for SSE,
// command, args and env for a command, excludeTools for the tools left
// out. The user's own servers and the rest of the settings stay.
func TestQwenMCPServers(t *testing.T) {
	h := sandbox(t)
	cfg := filepath.Join(h, ".qwen", "settings.json")
	write(t, cfg, "{\n  // mine\n  \"model\": {\"name\": \"qwen3.8-max\"},\n  \"mcpServers\": {\n    \"own\": {\"command\": \"own-mcp\", \"timeout\": 30000}\n  }\n}\n")
	tg := targetByID("qwen")
	if tg == nil || tg.MCP == nil || tg.MCP.Path != cfg || tg.MCP.Format != fmtGemini {
		t.Fatalf("Qwen Code's MCP servers: %+v", tg)
	}
	for _, q := range []string{"qwen", "qwen-code", "Qwen Code"} {
		if id, err := Takes(q, "mcp"); err != nil || id != "qwen" {
			t.Fatalf("Takes(%s, mcp): %q %v", q, id, err)
		}
	}

	ok(t)(SaveServer("", Server{Name: "docs", Transport: "http", URL: "https://mcp.example.com/mcp", Headers: map[string]string{"X-Team": "core"}, Agents: []string{"qwen"}}))
	ok(t)(SaveServer("", Server{Name: "events", Transport: "sse", URL: "https://mcp.example.com/sse", Agents: []string{"qwen"}}))
	ok(t)(SaveServer("", Server{Name: "fs", Transport: "stdio", Command: "npx", Args: []string{"-y", "@mcp/fs"}, Env: map[string]string{"ROOT": "/srv"}, Agents: []string{"qwen"}}))
	es, err := tg.MCP.entries()
	if err != nil {
		t.Fatal(err)
	}
	// the user's own server is as it was
	if e := es["own"]; e["command"] != "own-mcp" || e["timeout"] != float64(30000) {
		t.Errorf("the user's own server: %v", e)
	}
	if e := es["docs"]; e["httpUrl"] != "https://mcp.example.com/mcp" || e["url"] != nil || e["command"] != nil {
		t.Errorf("a streamable server: %v", e)
	} else if hd, _ := e["headers"].(map[string]any); hd["X-Team"] != "core" {
		t.Errorf("a streamable server's headers: %v", e)
	}
	if e := es["events"]; e["url"] != "https://mcp.example.com/sse" || e["httpUrl"] != nil {
		t.Errorf("an SSE server: %v", e)
	}
	if e := es["fs"]; e["command"] != "npx" || e["httpUrl"] != nil || e["url"] != nil {
		t.Errorf("a command: %v", e)
	} else {
		args, _ := e["args"].([]any)
		env, _ := e["env"].(map[string]any)
		if len(args) != 2 || args[0] != "-y" || args[1] != "@mcp/fs" || env["ROOT"] != "/srv" {
			t.Errorf("a command's args and env: %v", e)
		}
	}
	// a server of the library's with the name of one the user wrote by hand:
	// what says where the server is becomes the library's, the user's other
	// keys in the entry stay, as in every agent's file
	ok(t)(SaveServer("", Server{Name: "own", Transport: "stdio", Command: "library-mcp", Agents: []string{"qwen"}}))
	if es, err = tg.MCP.entries(); err != nil {
		t.Fatal(err)
	}
	if e := es["own"]; e["command"] != "library-mcp" || e["timeout"] != float64(30000) {
		t.Errorf("the library's server over the user's of the same name: %v", e)
	}
	// the rest of the settings are as they were
	if s := read(t, cfg); !strings.Contains(s, "// mine") || !strings.Contains(s, "qwen3.8-max") {
		t.Errorf("the settings beside the servers:\n%s", s)
	}

	// a tool turned off is written into the server's entry, as Gemini CLI's
	if toolsOffOf(tg) != toolsInEntry {
		t.Fatalf("Qwen Code: no list of tools left out")
	}
	ok(t)(ServerTools("fs", []string{"delete"}))
	if got, _ := offIn(t, tg, "fs"); !slices.Equal(got, []string{"delete"}) {
		t.Errorf("excludeTools: %v, want [delete]", got)
	}
	// the key as Qwen Code reads it, in the file itself
	var onDisk struct {
		MCPServers map[string]struct {
			ExcludeTools []string `json:"excludeTools"`
		} `json:"mcpServers"`
	}
	if err := json.Unmarshal(jsonc.ToJSON([]byte(read(t, cfg))), &onDisk); err != nil {
		t.Fatal(err)
	}
	if got := onDisk.MCPServers["fs"].ExcludeTools; !slices.Equal(got, []string{"delete"}) {
		t.Errorf("mcpServers.fs.excludeTools in settings.json: %v, want [delete]", got)
	}

	// a server taken away leaves the others, the user's own among them
	ok(t)(RemoveServer("docs"))
	if es, err = tg.MCP.entries(); err != nil {
		t.Fatal(err)
	}
	if es["docs"] != nil || es["fs"] == nil || es["events"] == nil || es["own"] == nil {
		t.Errorf("after docs is removed: %v", es)
	}
	if s := read(t, cfg); !strings.Contains(s, "// mine") || !strings.Contains(s, "qwen3.8-max") {
		t.Errorf("the settings beside the servers, after a server is removed:\n%s", s)
	}

	// a variable named in a header is written as $NAME or ${NAME}, which Qwen
	// Code puts in as it loads the file, as Gemini CLI does
	ok(t)(SaveServer("", Server{Name: "gh", Transport: "http", URL: "https://mcp.example.com/gh", Headers: map[string]string{"Authorization": "Bearer ${GITHUB_TOKEN}"}, Agents: []string{"qwen"}}))
	if es, err = tg.MCP.entries(); err != nil {
		t.Fatal(err)
	}
	if hd, _ := es["gh"]["headers"].(map[string]any); hd["Authorization"] != "Bearer ${GITHUB_TOKEN}" {
		t.Errorf("a header that names a variable: %v", es["gh"])
	}

	// $QWEN_HOME moves the file
	q := filepath.Join(h, "qwen-home")
	t.Setenv("QWEN_HOME", q)
	write(t, filepath.Join(q, "settings.json"), "{}\n")
	if tg := targetByID("qwen"); tg == nil || tg.MCP == nil || tg.MCP.Path != filepath.Join(q, "settings.json") {
		t.Fatalf("Qwen Code under QWEN_HOME: %+v", tg)
	}
}
