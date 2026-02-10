.PHONY: help lint lint-go dupl-go test-go fmt-go

help:
	@echo "Targets:"
	@echo "  make lint-go   - golangci-lint ./..."
	@echo "  make dupl-go   - fail on duplicated Go code (dupl -t 50)"
	@echo "  make test-go   - go test ./..."
	@echo "  make fmt-go    - gofmt -w on tracked .go files"
	@echo "  make lint      - run all linters"

lint: lint-go dupl-go

lint-go:
	@golangci-lint run ./...

dupl-go:
	@tmp="$$(mktemp)"; \
	dupl -t 50 -plumbing services libs > "$$tmp"; \
	if [ -s "$$tmp" ]; then \
		cat "$$tmp"; \
		echo "dupl-go: duplicates found (threshold=50)"; \
		rm -f "$$tmp"; \
		exit 1; \
	fi; \
	rm -f "$$tmp"

test-go:
	@go test ./...

fmt-go:
	@git ls-files '*.go' | xargs gofmt -w
