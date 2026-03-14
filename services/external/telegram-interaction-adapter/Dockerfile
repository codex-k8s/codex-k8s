# syntax=docker/dockerfile:1.7

ARG GOLANG_IMAGE=127.0.0.1:5000/codex-k8s/mirror/golang:1.25.8-bookworm

FROM ${GOLANG_IMAGE} AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/telegram-interaction-adapter ./services/external/telegram-interaction-adapter/cmd/telegram-interaction-adapter

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app

COPY --from=builder /out/telegram-interaction-adapter /usr/local/bin/telegram-interaction-adapter

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/telegram-interaction-adapter"]
