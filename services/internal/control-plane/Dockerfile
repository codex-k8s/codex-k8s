FROM golang:1.25.6-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
COPY libs/go/ libs/go/
COPY proto/ proto/
COPY services/internal/control-plane/ services/internal/control-plane/

RUN go mod download
RUN go install github.com/pressly/goose/v3/cmd/goose@latest
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/codex-k8s-control-plane ./services/internal/control-plane/cmd/control-plane

FROM alpine:3.20

RUN addgroup -S app && adduser -S app -G app
USER app

COPY --from=builder /out/codex-k8s-control-plane /usr/local/bin/codex-k8s-control-plane
COPY --from=builder /go/bin/goose /usr/local/bin/goose

EXPOSE 8081
EXPOSE 9090

ENTRYPOINT ["/usr/local/bin/codex-k8s-control-plane"]
