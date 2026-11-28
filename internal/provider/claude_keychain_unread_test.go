package provider

// A saved Claude account whose keychain item can't be read (the login
// keychain out of reach from an ssh session: `security` exits 36,
// errSecInteractionNotAllowed) is unknown. The file Claude Code emptied
// beside it once it kept the sign-in in the keychain is not read in its
// place, so the account is not marked refused, and no file is written
// with the saved sign-in to become a second holder of its refresh token.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/testenv"
)

func fakeClaudeKeychain(t *testing.T, exit int, output string) {
	t.Helper()
	bin := t.TempDir()
	if runtime.GOOS == "windows" {
		script := "@echo off\r\n"
		if output != "" {
			script += "@echo " + output + "\r\n"
		}
		script += "exit /b " + strconv.Itoa(exit) + "\r\n"
		if err := os.WriteFile(filepath.Join(bin, "security.cmd"), []byte(script), 0o700); err != nil {
			t.Fatal(err)
		}
	} else {
		script := "#!/bin/sh\n"
		if output != "" {
			script += "printf '%s' '" + output + "'\n"
		}
		script += "exit " + strconv.Itoa(exit) + "\n"
		testenv.Program(t, filepath.Join(bin, "security"), script)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	old := claudeKeychain
	claudeKeychain = true
	t.Cleanup(func() { claudeKeychain = old })
	return
}

func writeClaudeDirCredentials(t *testing.T, user, access, refresh string) string {
	t.Helper()
	dir := claudeAccountDir(user)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	creds := filepath.Join(dir, ".credentials.json")
	writeFile(t, creds, map[string]any{"claudeAiOauth": map[string]any{
		"accessToken": access, "refreshToken": refresh, "expiresAt": 0, "subscriptionType": "max",
	}})
	return creds
}

func emptiedClaudeFileBeside(t *testing.T, exit int) (Provider, savedLogin, string) {
	t.Helper()
	p, side := claudeAuthFixture(t, "")
	creds := writeClaudeDirCredentials(t, side.User, "", "")
	fakeClaudeKeychain(t, exit, "")
	return p, side, creds
}

func TestClaudeUnreadKeychainIsNotARefusedSignIn(t *testing.T) {
	for _, exit := range []int{36, 51, 128} {
		t.Run(strconv.Itoa(exit), func(t *testing.T) {
			p, side, creds := emptiedClaudeFileBeside(t, exit)
			_, err := sideToken(t, p, side.User)
			if err == nil || !strings.Contains(err.Error(), "keychain") {
				t.Fatalf("run on an unreadable keychain: %v", err)
			}
			if ClaudeSignInRequired(err.Error()) {
				t.Fatalf("an unreadable keychain reads as a sign-in to make again: %v", err)
			}
			if l := claudeLapseOf(side.User); l != "" {
				t.Fatalf("marked lapsed for a keychain it couldn't read: %q", l)
			}
			for _, l := range readLogins() {
				if strings.EqualFold(l.User, side.User) && l.Refused != "" {
					t.Fatalf("refused kept: %q", l.Refused)
				}
			}
			if b, _ := os.ReadFile(creds); strings.Contains(string(b), "old-refresh") {
				t.Fatal("the saved sign-in was written beside the keychain: a second holder of its refresh token")
			}
		})
	}
}

func TestClaudeMalformedKeychainIsNotARefusedSignIn(t *testing.T) {
	p, side := claudeAuthFixture(t, "")
	creds := writeClaudeDirCredentials(t, side.User, "", "")
	fakeClaudeKeychain(t, 0, "not-json")

	_, err := sideToken(t, p, side.User)
	if err == nil || !strings.Contains(err.Error(), "keychain") || ClaudeSignInRequired(err.Error()) {
		t.Fatalf("malformed keychain read: %v", err)
	}
	if claudeLapseOf(side.User) != "" {
		t.Fatalf("malformed keychain marked lapsed: %q", claudeLapseOf(side.User))
	}
	if b, _ := os.ReadFile(creds); strings.Contains(string(b), "old-refresh") {
		t.Fatal("the saved sign-in was written beside the unreadable keychain")
	}
}

func TestClaudeClearedKeychainIsKnownEmptyBeforeTheFile(t *testing.T) {
	p, side := claudeAuthFixture(t, "")
	creds := writeClaudeDirCredentials(t, side.User, "", "")
	cleared, err := json.Marshal(map[string]any{"claudeAiOauth": map[string]any{
		"accessToken": "", "refreshToken": "", "expiresAt": 0, "subscriptionType": "max",
	}})
	if err != nil {
		t.Fatal(err)
	}
	fakeClaudeKeychain(t, 0, string(cleared))

	_, err = sideToken(t, p, side.User)
	if err == nil || !ClaudeSignInRequired(err.Error()) {
		t.Fatalf("an emptied keychain item wasn't treated as the known sign-in: %v", err)
	}
	if claudeLapseOf(side.User) != claudeAuthLapse {
		t.Fatalf("cleared keychain lapse: %q", claudeLapseOf(side.User))
	}
	if b, _ := os.ReadFile(creds); strings.Contains(string(b), "old-refresh") {
		t.Fatal("the stale file was used beside the emptied keychain")
	}
}

// An item that isn't there (exit 44) still leaves the file as Claude
// Code's word, emptied or not: what it read before stands.
func TestClaudeMissingKeychainItemStillReadsTheFile(t *testing.T) {
	p, side, _ := emptiedClaudeFileBeside(t, 44)
	if _, err := sideToken(t, p, side.User); err == nil {
		t.Fatal("ran on a sign-in Claude Code emptied")
	}
	if claudeLapseOf(side.User) != claudeAuthLapse {
		t.Fatalf("lapse: %q", claudeLapseOf(side.User))
	}
}

func TestClaudeMissingKeychainItemUsesTheFileWhenItHasSignIn(t *testing.T) {
	_, side := claudeAuthFixture(t, "")
	writeClaudeDirCredentials(t, side.User, "file-access", "file-refresh")
	fakeClaudeKeychain(t, 44, "")

	c, ok := readClaudeDir(claudeAccountDir(side.User))
	if !ok || c.OAuth.AccessToken != "file-access" || c.OAuth.RefreshToken != "file-refresh" {
		t.Fatalf("file after missing keychain item: %v %+v", ok, c.OAuth)
	}
}

func TestClaudeReadableKeychainTakesPrecedenceOverTheFile(t *testing.T) {
	_, side := claudeAuthFixture(t, "")
	writeClaudeDirCredentials(t, side.User, "file-access", "file-refresh")
	item, err := json.Marshal(map[string]any{"claudeAiOauth": map[string]any{
		"accessToken": "keychain-access", "refreshToken": "keychain-refresh", "expiresAt": 0, "subscriptionType": "max",
	}})
	if err != nil {
		t.Fatal(err)
	}
	fakeClaudeKeychain(t, 0, string(item))

	c, ok := readClaudeDir(claudeAccountDir(side.User))
	if !ok || c.OAuth.AccessToken != "keychain-access" || c.OAuth.RefreshToken != "keychain-refresh" {
		t.Fatalf("keychain ahead of file: %v %+v", ok, c.OAuth)
	}
}
