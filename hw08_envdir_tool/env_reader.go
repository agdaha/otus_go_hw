package main

import (
	"os"
	"path/filepath"
	"strings"
)

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	env := make(Environment)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.Contains(name, "=") {
			continue
		}

		filePath := filepath.Join(dir, name)
		fileData, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		if len(fileData) == 0 {
			env[name] = EnvValue{NeedRemove: true}
			continue
		}

		value := string(fileData)
		if idx := strings.IndexByte(value, '\n'); idx >= 0 {
			value = value[:idx]
		}

		value = strings.ReplaceAll(value, "\x00", "\n")
		value = strings.TrimRight(value, " \t")

		env[name] = EnvValue{Value: value}
	}

	return env, nil
}
