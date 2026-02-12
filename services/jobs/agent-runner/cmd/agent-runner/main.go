package main

import (
	"fmt"
	"os"

	"github.com/codex-k8s/codex-k8s/services/jobs/agent-runner/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		exitCode := 1
		if appErr, ok := app.AsExitError(err); ok {
			exitCode = appErr.ExitCode
		}
		os.Exit(exitCode)
	}
}
