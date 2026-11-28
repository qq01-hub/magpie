package gui

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/gateway"
	"github.com/yetone/magpie/internal/library"
	"github.com/yetone/magpie/internal/plugin"
	"github.com/yetone/magpie/internal/provider"
)

func TestLibraryPluginIcons(t *testing.T) {
	for _, custom := range []bool{false, true} {
		name := "plugin manifest"
		if custom {
			name = "Library override"
		}
		t.Run(name, func(t *testing.T) {
			cfg := pluginsSandbox(t)
			const spec = "@example/search"
			const customIcon = "https://example.org/assigned.png"
			dir := plugin.Target(spec)
			if err := os.MkdirAll(dir, 0o700); err != nil {
				t.Fatal(err)
			}
			var imageData bytes.Buffer
			if err := png.Encode(&imageData, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
				t.Fatal(err)
			}
			url := strings.TrimRight(gateway.URL(), "/") + gateway.ToolsMCPOne + gateway.ToolPrefix(spec)
			fixtures := map[string]any{
				filepath.Join(dir, "package.json"): map[string]any{
					"magpie": map[string]string{"icon": "data:image/png;base64," + base64.StdEncoding.EncodeToString(imageData.Bytes())},
				},
				filepath.Join(cfg, "plugins.json"): map[string]any{"plugins": []map[string]string{{"spec": spec}}},
			}
			data := map[string]any{
				"pluginsOffered": []string{spec},
				"mcp": []map[string]string{
					{"name": "search", "transport": "http", "url": url, "plugin": spec},
					{"name": "ordinary", "transport": "http", "url": "https://example.org/mcp"},
				},
			}
			if custom {
				data["icons"] = map[string]string{strings.ToLower(url): customIcon}
			}
			fixtures[filepath.Join(cfg, "library.json")] = data
			for path, fixture := range fixtures {
				b, err := json.Marshal(fixture)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, b, 0o600); err != nil {
					t.Fatal(err)
				}
			}

			view, err := libraryRead(nil, true)
			if err != nil {
				t.Fatal(err)
			}
			icons := map[string]string{}
			for _, server := range view.Servers {
				icons[server.Name] = server.Icon
			}
			if len(icons) != 2 || icons["ordinary"] != "" {
				t.Fatalf("unexpected icons for Library servers: %v", icons)
			}
			// Saving an MCP setting replaces the page's servers with this
			// partial response; it must keep the same picture and overrides.
			partial, err := mcpView(&library.Result{})
			if err != nil {
				t.Fatal(err)
			}
			if partial.MCP == nil || len(partial.MCP.Servers) != 2 {
				t.Fatalf("partial MCP response: %+v", partial)
			}
			for _, s := range partial.MCP.Servers {
				if s.Icon != icons[s.Name] {
					t.Errorf("MCP-only update lost %s icon: %q, want %q", s.Name, s.Icon, icons[s.Name])
				}
			}
			if custom {
				if icons["search"] != customIcon {
					t.Fatalf("icon assigned by the Library = %q, want %q", icons["search"], customIcon)
				}
				return
			}
			file, ok := strings.CutPrefix(icons["search"], "file:")
			if !ok {
				t.Fatalf("plugin server icon = %q, want a cached file icon", icons["search"])
			}
			b, err := os.ReadFile(provider.IconFile(file))
			if err != nil || !bytes.Equal(b, imageData.Bytes()) {
				t.Fatalf("reading cached plugin icon: %v, image matches: %v", err, bytes.Equal(b, imageData.Bytes()))
			}
		})
	}
}

func TestLibraryPluginIconsCacheBoundary(t *testing.T) {
	cfg := pluginsSandbox(t)
	var data bytes.Buffer
	if err := png.Encode(&data, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	icon, err := provider.StoreIcon(data.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	const private = "private configuration outside the icon cache"
	if err := os.WriteFile(filepath.Join(cfg, "providers.json"), []byte(private), 0o600); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	providerRoutes(mux, nil)
	get := func(name string) *httptest.ResponseRecorder {
		t.Helper()
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/icons/"+url.PathEscape(name), nil))
		return w
	}
	if w := get(strings.TrimPrefix(icon, "file:")); w.Code != http.StatusOK ||
		w.Header().Get("Content-Type") != "image/png" || !bytes.Equal(w.Body.Bytes(), data.Bytes()) {
		t.Fatalf("cached Library picture: status=%d, type=%q, body=%q", w.Code, w.Header().Get("Content-Type"), w.Body.Bytes())
	}
	for _, name := range []string{"../providers.json", `..\providers.json`, "../../providers.json", "providers.json", "0123456789abcdef.exe"} {
		if w := get(name); w.Code != http.StatusNotFound || strings.Contains(w.Body.String(), private) {
			t.Errorf("unsafe Library picture %q: status=%d body=%q", name, w.Code, w.Body.String())
		}
	}
}
