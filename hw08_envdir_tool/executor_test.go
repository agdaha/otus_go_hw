package main

import (
	"os"
	"testing"
)

func TestRunCmd(t *testing.T) {
	t.Run("returns exit code 0 on success", func(t *testing.T) {
		code := RunCmd([]string{"true"}, Environment{})
		if code != 0 {
			t.Errorf("got %d, want 0", code)
		}
	})

	t.Run("returns non-zero exit code on failure", func(t *testing.T) {
		code := RunCmd([]string{"false"}, Environment{})
		if code == 0 {
			t.Error("expected non-zero exit code")
		}
	})

	t.Run("sets env variable for child process", func(t *testing.T) {
		env := Environment{
			"MY_TEST_VAR": {Value: "hello123"},
		}
		// sh -c exits 0 only if the variable matches
		code := RunCmd([]string{"sh", "-c", `[ "$MY_TEST_VAR" = "hello123" ]`}, env)
		if code != 0 {
			t.Error("env variable was not passed correctly to child")
		}
	})

	t.Run("removes env variable when NeedRemove is true", func(t *testing.T) {
		// Set the var in current process so child would inherit it
		os.Setenv("REMOVE_ME_TEST", "should_disappear")
		defer os.Unsetenv("REMOVE_ME_TEST")

		env := Environment{
			"REMOVE_ME_TEST": {NeedRemove: true},
		}
		// sh -c exits 1 if the variable is set (empty or non-empty), 0 if unset
		code := RunCmd([]string{"sh", "-c", `[ -z "${REMOVE_ME_TEST+x}" ]`}, env)
		if code != 0 {
			t.Error("env variable was not removed from child process")
		}
	})

	t.Run("passes args to command", func(t *testing.T) {
		code := RunCmd([]string{"sh", "-c", `[ "$1" = "hello" ]`, "--", "hello"}, Environment{})
		if code != 0 {
			t.Error("args were not passed correctly")
		}
	})

	t.Run("empty command returns 0", func(t *testing.T) {
		code := RunCmd([]string{}, Environment{})
		if code != 0 {
			t.Errorf("expected 0 for empty cmd, got %d", code)
		}
	})

	t.Run("inherits existing env variables", func(t *testing.T) {
		os.Setenv("INHERITED_VAR", "inherited_value")
		defer os.Unsetenv("INHERITED_VAR")

		code := RunCmd([]string{"sh", "-c", `[ "$INHERITED_VAR" = "inherited_value" ]`}, Environment{})
		if code != 0 {
			t.Error("inherited env variable was not passed to child")
		}
	})

	t.Run("returns specific exit code", func(t *testing.T) {
		code := RunCmd([]string{"sh", "-c", "exit 42"}, Environment{})
		if code != 42 {
			t.Errorf("got exit code %d, want 42", code)
		}
	})
}
