.PHONY: help lint lint-go dupl-go test-go fmt-go gen-openapi-go gen-openapi-ts gen-openapi gen-proto-go

help:
	@echo "Targets:"
	@echo "  make lint-go   - golangci-lint ./..."
	@echo "  make dupl-go   - fail on duplicated Go code (dupl -t 50)"
	@echo "  make test-go   - go test ./..."
	@echo "  make fmt-go    - gofmt -w on tracked .go files"
	@echo "  make gen-openapi-go [SVC=services/external/api-gateway] - generate Go transport code from OpenAPI"
	@echo "  make gen-openapi-ts [APP=services/staff/web-console] - generate TS API client from OpenAPI"
	@echo "  make gen-openapi - run Go+TS OpenAPI generators for default services"
	@echo "  make gen-proto-go - generate Go gRPC contracts from proto/**/*.proto"
	@echo "  make lint      - run all linters"

lint: lint-go dupl-go

lint-go:
	@golangci-lint run ./...

dupl-go:
	@tmp="$$(mktemp)"; \
	filtered="$$(mktemp)"; \
	dupl -t 50 -plumbing services libs > "$$tmp"; \
	grep -v '/generated/' "$$tmp" > "$$filtered" || true; \
	if [ -s "$$filtered" ]; then \
		cat "$$filtered"; \
		echo "dupl-go: duplicates found (threshold=50)"; \
		rm -f "$$tmp" "$$filtered"; \
		exit 1; \
	fi; \
	rm -f "$$tmp" "$$filtered"

test-go:
	@go test ./...

fmt-go:
	@git ls-files '*.go' | xargs gofmt -w

gen-openapi-go:
	@svc="$${SVC:-services/external/api-gateway}"; \
	spec="$$svc/api/server/api.yaml"; \
	cfg="tools/codegen/openapi/api-gateway.oapi-codegen.yaml"; \
	out="$$svc/internal/transport/http/generated/openapi.gen.go"; \
	if [ "$$svc" != "services/external/api-gateway" ]; then \
		echo "gen-openapi-go: unsupported SVC=$$svc (currently only services/external/api-gateway)"; \
		exit 1; \
	fi; \
	test -f "$$spec"; \
	test -f "$$cfg"; \
	mkdir -p "$$(dirname "$$out")"; \
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.5.0 -config "$$cfg" "$$spec" > "$$out"

gen-openapi-ts:
	@app="$${APP:-services/staff/web-console}"; \
	if [ "$$app" != "services/staff/web-console" ]; then \
		echo "gen-openapi-ts: unsupported APP=$$app (currently only services/staff/web-console)"; \
		exit 1; \
	fi; \
	npm --prefix "$$app" run gen:openapi

gen-openapi: gen-openapi-go gen-openapi-ts

gen-proto-go:
	@protoc -I proto \
		--go_out=proto/gen/go --go_opt=paths=source_relative \
		--go-grpc_out=proto/gen/go --go-grpc_opt=paths=source_relative \
		proto/codexk8s/controlplane/v1/controlplane.proto
