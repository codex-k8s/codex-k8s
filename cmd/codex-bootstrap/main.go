package main

import (
	"os"

	"github.com/codex-k8s/codex-k8s/cmd/codex-bootstrap/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
