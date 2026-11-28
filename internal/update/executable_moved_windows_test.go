package update

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/yetone/magpie/internal/proc"
	"golang.org/x/sys/windows"
)

func TestQueryProcessExeInvalidHandle(t *testing.T) {
	if path, err := queryProcessExe(windows.Handle(0)); err == nil || path != "" {
		t.Fatalf("invalid process handle returned %q, %v", path, err)
	}
}

func TestExecutableMovedHelper(t *testing.T) {
	if os.Getenv("MAGPIE_UPDATE_MOVED_HELPER") != "1" {
		return
	}
	fmt.Fprintln(os.Stdout, "ready")
	if !bufio.NewScanner(os.Stdin).Scan() {
		t.Fatal("parent did not release helper")
	}
	exe, err := Executable()
	result := struct{ Path, Error string }{Path: exe}
	if err != nil {
		result.Error = err.Error()
	}
	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		t.Fatal(err)
	}
}

// Only a private copy of this test binary is moved, after it has started
// and before it asks the production Executable function for its path
func TestExecutableMovedWhileRunning(t *testing.T) {
	for _, long := range []bool{false, true} {
		name := "short"
		if long {
			name = "long-unicode"
		}
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			start := filepath.Join(root, "download", "helper.exe")
			destDir := filepath.Join(root, "moved")
			if long {
				for range 4 {
					destDir = filepath.Join(destDir, strings.Repeat("路径-sdk-", 12))
				}
			}
			dest := filepath.Join(destDir, "已移动.exe")
			for _, dir := range []string{filepath.Dir(start), destDir} {
				if err := os.MkdirAll(dir, 0o700); err != nil {
					t.Fatal(err)
				}
			}
			self, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			copyExecutableT(t, self, start)
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			cmd := proc.CommandContext(ctx, start, "-test.run=^TestExecutableMovedHelper$")
			cmd.Env = append(os.Environ(), "MAGPIE_UPDATE_MOVED_HELPER=1")
			cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: noWindow}
			input, err := cmd.StdinPipe()
			if err != nil {
				t.Fatal(err)
			}
			output, err := cmd.StdoutPipe()
			if err != nil {
				t.Fatal(err)
			}
			cmd.Stderr = os.Stderr
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				input.Close()
				cmd.Process.Kill()
				cmd.Wait()
			})
			scanner := bufio.NewScanner(output)
			if !scanner.Scan() || scanner.Text() != "ready" {
				t.Fatalf("helper did not become ready: %q, %v", scanner.Text(), scanner.Err())
			}
			if err := os.Rename(start, dest); err != nil {
				t.Fatal(err)
			}
			fmt.Fprintln(input, "resolve")
			if !scanner.Scan() {
				t.Fatalf("helper returned no result: %v", scanner.Err())
			}
			var result struct{ Path, Error string }
			if err := json.Unmarshal(scanner.Bytes(), &result); err != nil {
				t.Fatalf("helper result %q: %v", scanner.Text(), err)
			}
			if result.Error != "" || !strings.EqualFold(filepath.Clean(result.Path), filepath.Clean(dest)) {
				t.Fatalf("Executable after move = %q, error %q; want %q", result.Path, result.Error, dest)
			}
			if _, err := os.Stat(result.Path); err != nil {
				t.Fatalf("resolved executable does not exist: %v", err)
			}
			io.Copy(io.Discard, output)
			if err := cmd.Wait(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func copyExecutableT(t *testing.T, source, dest string) {
	t.Helper()
	in, err := os.Open(source)
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o700)
	if err != nil {
		t.Fatal(err)
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil || closeErr != nil {
		t.Fatalf("copy test executable: %v %v", copyErr, closeErr)
	}
}
