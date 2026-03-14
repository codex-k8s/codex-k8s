package main

import (
	"log"

	"github.com/codex-k8s/codex-k8s/services/external/telegram-interaction-adapter/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
