package main

import (
	"log"

	"github.com/codex-k8s/kodex/services/internal/control-plane/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatalf("control-plane failed: %v", err)
	}
}
