package main

import (
	"context"
	"log"
	"os"

	"github.com/codex-k8s/codex-k8s/bin/codex-bootstrap/internal/cli"
)

func main() {
	if err := cli.ExecuteContext(context.Background()); err != nil {
		log.Printf("codex-bootstrap failed: %v", err)
		os.Exit(1)
	}
}
