package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s /path/to/env/dir command [arg1 arg2 ...]", os.Args[0])
		os.Exit(1)
	}

	envDir := os.Args[1]
	cmd := os.Args[2:]

	env, err := ReadDir(envDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading env dir: %v\n", err)
		os.Exit(1)
	}

	exitCode := RunCmd(cmd, env)

	os.Exit(exitCode)
}
