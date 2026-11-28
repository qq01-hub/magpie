package library

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestHooksDshEditedMarkedInsertStaysTheUsers(t *testing.T) {
	for _, layout := range []string{"marked", "reformatted"} {
		for _, change := range []string{"path", "setting", "sibling", "package"} {
			t.Run(layout+"/"+change, func(t *testing.T) {
				dir := t.TempDir()
				file := filepath.Join(dir, "magpie-hooks.json")
				patch := filepath.Join(dir, "cordis.patch.yml")
				row := "- insert: # magpie\n    - id: magpie-hooks\n      name: " + dshScalar(dshHooksPkg) + "\n      config:\n        configPath: " + dshScalar(file) + "\n"
				switch change {
				case "path":
					row = strings.Replace(row, dshScalar(file), dshScalar(filepath.Join(dir, "own.json")), 1)
				case "setting":
					row += "        ownSetting: true\n"
				case "sibling":
					row += "    - id: own-plugin\n      name: other-plugin\n"
				case "package":
					row = strings.Replace(row, dshHooksPkg, "my-hooks-plugin", 1)
				}
				if layout == "reformatted" {
					row = strings.Replace(row, "- insert: # magpie\n", "- insert:\n    # magpie\n", 1)
				}
				write(t, patch, row)
				kept := 0
				keep := func(string) error { kept++; return nil }
				_ = dshPutHooksRow(patch, file, keep)
				if err := dshDelHooksRow(patch, file, keep); err != nil {
					t.Fatal(err)
				}
				if got := read(t, patch); got != row || kept != 0 {
					t.Fatalf("a user-edited insert was rewritten or removed, backups %d:\n%s", kept, got)
				}
			})
		}
	}
}
