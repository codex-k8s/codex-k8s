package main

import (
	"log"

	"github.com/codex-k8s/kodex/services/external/api-gateway/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatalf("api-gateway failed: %v", err)
	}
}
