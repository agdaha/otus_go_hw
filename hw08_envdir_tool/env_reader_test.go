package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadDir(t *testing.T) {
	t.Run("read simple variable", testSimpleVariable)
	t.Run("empty file marks variable for removal", testEmptyFileMarksRemoval)
	t.Run("only first line is used", testOnlyFirstLine)
	t.Run("spaces and tabs are trimmed", testTrailingWhitespaceTrimmed)
	t.Run("null bytes replaced with newlines", testNullBytesReplacedWithNewlines)
	t.Run("files with '=' are skipped", testSkipsNamesWithEquals)
	t.Run("subdirectories are skipped", testSkipsSubdirectories)
	t.Run("nonexistent dir returns error", testNonexistentDirReturnsError)
	t.Run("multiple variables", testMultipleVariables)
}

func makeEnvDir(t *testing.T, files map[string][]byte) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), content, 0o644); err != nil {
			t.Fatalf("makeEnvDir: write %s: %v", name, err)
		}
	}
	return dir
}

func readDirOrFatal(t *testing.T, dir string) Environment {
	t.Helper()
	env, err := ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir(%q): unexpected error: %v", dir, err)
	}
	return env
}

func testSimpleVariable(t *testing.T) {
	dir := makeEnvDir(t, map[string][]byte{"FOO": []byte("bar")})
	env := readDirOrFatal(t, dir)

	got, ok := env["FOO"]
	if !ok {
		t.Fatal("FOO not found in env")
	}
	if got.Value != "bar" || got.NeedRemove {
		t.Errorf("got %+v, want {Value:bar NeedRemove:false}", got)
	}
}

func testEmptyFileMarksRemoval(t *testing.T) {
	dir := makeEnvDir(t, map[string][]byte{"EMPTY": {}})
	env := readDirOrFatal(t, dir)

	got, ok := env["EMPTY"]
	if !ok {
		t.Fatal("EMPTY not found in env")
	}
	if !got.NeedRemove {
		t.Error("expected NeedRemove=true for empty file")
	}
}

func testOnlyFirstLine(t *testing.T) {
	dir := makeEnvDir(t, map[string][]byte{"MULTILINE": []byte("first\nsecond\nthird")})
	env := readDirOrFatal(t, dir)

	if env["MULTILINE"].Value != "first" {
		t.Errorf("got %q, want %q", env["MULTILINE"].Value, "first")
	}
}

func testTrailingWhitespaceTrimmed(t *testing.T) {
	dir := makeEnvDir(t, map[string][]byte{"SPACES": []byte("hello   \t  ")})
	env := readDirOrFatal(t, dir)

	if env["SPACES"].Value != "hello" {
		t.Errorf("got %q, want %q", env["SPACES"].Value, "hello")
	}
}

func testNullBytesReplacedWithNewlines(t *testing.T) {
	dir := makeEnvDir(t, map[string][]byte{"NULLS": []byte("a\x00b\x00c")})
	env := readDirOrFatal(t, dir)

	if env["NULLS"].Value != "a\nb\nc" {
		t.Errorf("got %q, want %q", env["NULLS"].Value, "a\nb\nc")
	}
}

func testSkipsNamesWithEquals(t *testing.T) {
	dir := makeEnvDir(t, map[string][]byte{
		"BAD=NAME": []byte("value"),
		"GOOD":     []byte("ok"),
	})
	env := readDirOrFatal(t, dir)

	if _, ok := env["BAD=NAME"]; ok {
		t.Error("expected BAD=NAME to be skipped")
	}
	if env["GOOD"].Value != "ok" {
		t.Errorf("GOOD = %q, want %q", env["GOOD"].Value, "ok")
	}
}

func testSkipsSubdirectories(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "subdir"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "VAR"), []byte("val"), 0o644); err != nil {
		t.Fatal(err)
	}
	env := readDirOrFatal(t, dir)

	if _, ok := env["subdir"]; ok {
		t.Error("subdirectory should not appear in env")
	}
	if env["VAR"].Value != "val" {
		t.Errorf("VAR = %q, want %q", env["VAR"].Value, "val")
	}
}

func testNonexistentDirReturnsError(t *testing.T) {
	_, err := ReadDir("/nonexistent/path/xyz")
	if err == nil {
		t.Error("expected error for nonexistent dir, got nil")
	}
}

func testMultipleVariables(t *testing.T) {
	dir := makeEnvDir(t, map[string][]byte{
		"A": []byte("1"),
		"B": []byte("2"),
		"C": []byte("3"),
	})
	env := readDirOrFatal(t, dir)

	for _, tc := range []struct{ key, val string }{
		{"A", "1"}, {"B", "2"}, {"C", "3"},
	} {
		if env[tc.key].Value != tc.val {
			t.Errorf("%s = %q, want %q", tc.key, env[tc.key].Value, tc.val)
		}
	}
}
