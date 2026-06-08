package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	if len(cmd) == 0 {
		return 0
	}

	command := exec.Command(cmd[0], cmd[1:]...) //#nosec G204

	envSlice := make([]string, 0, len(os.Environ()))
	for _, e := range os.Environ() {
		key, _, found := strings.Cut(e, "=")
		if !found {
			continue
		}
		if _, exists := env[key]; exists {
			continue
		}
		envSlice = append(envSlice, e)
	}

	for key, val := range env {
		if !val.NeedRemove {
			envSlice = append(envSlice, key+"="+val.Value)
		}
	}

	command.Env = envSlice

	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	if err := command.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode()
		}
		return 1
	}

	return 0
}
