package agent

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yetone/magpie/internal/proc"
)

func TestOpenCodeBrowserDirectoryAloneIsNotOpenCode(t *testing.T) {
	for _, files := range [][]string{{"browser"}, {"browser", "skills"}} {
		t.Run(filepath.Join(files...), func(t *testing.T) {
			home, _ := cindyDetectHome(t)
			t.Setenv("OPENCODE_CONFIG_DIR", "")
			configDir := filepath.Join(home, ".config", "opencode")
			for _, name := range files {
				if err := os.MkdirAll(filepath.Join(configDir, name), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			proc.Forget()
			a := opencode(home, filepath.Join(home, ".config"))
			if a.Detected() {
				t.Fatalf("OpenCode detected with only %v under its config directory", files)
			}

			writeAt(t, filepath.Join(configDir, "opencode.json"), `{}`)
			if !opencode(home, filepath.Join(home, ".config")).Detected() {
				t.Fatal("OpenCode config was not detected beside the browser directory")
			}
		})
	}
}
