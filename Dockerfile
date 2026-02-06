FROM golang:1.25.6-alpine AS builder

WORKDIR /src

COPY go.mod ./
COPY services/internal/control-plane/cmd/control-plane/main.go services/internal/control-plane/cmd/control-plane/main.go

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/codex-k8s ./services/internal/control-plane/cmd/control-plane

FROM alpine:3.20

RUN addgroup -S app && adduser -S app -G app
USER app

COPY --from=builder /out/codex-k8s /usr/local/bin/codex-k8s

# Metadata only; runtime listen address is controlled by CODEXK8S_HTTP_ADDR.
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/codex-k8s"]
