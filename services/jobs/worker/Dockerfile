FROM golang:1.25.6-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
COPY libs/go/ libs/go/
COPY proto/ proto/
COPY services/jobs/worker/ services/jobs/worker/

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/codex-k8s-worker ./services/jobs/worker/cmd/worker

FROM alpine:3.20

RUN addgroup -S app && adduser -S app -G app
USER app

COPY --from=builder /out/codex-k8s-worker /usr/local/bin/codex-k8s-worker

ENTRYPOINT ["/usr/local/bin/codex-k8s-worker"]
