package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	addr := os.Getenv("CODEXK8S_HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	mux := http.NewServeMux()

	ready := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}
	live := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("alive"))
	}

	// Backward compatibility with current deployment probes.
	mux.HandleFunc("/readyz", ready)
	mux.HandleFunc("/healthz", live)

	// Forward-compatible endpoints from design guidelines.
	mux.HandleFunc("/health/readyz", ready)
	mux.HandleFunc("/health/livez", live)

	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("codex-k8s bootstrap image"))
	})

	log.Printf("control-plane placeholder listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("http server failed: %v", err)
	}
}
