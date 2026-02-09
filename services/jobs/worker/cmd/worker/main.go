package main

import (
	"log"

	"github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatalf("worker failed: %v", err)
	}
}
