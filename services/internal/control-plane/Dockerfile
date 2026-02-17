ARG GOLANG_IMAGE=127.0.0.1:5000/codex-k8s/mirror/golang:1.25.6-alpine

FROM ${GOLANG_IMAGE} AS builder

WORKDIR /workspace

COPY go.mod go.sum ./
COPY libs/go/ libs/go/
COPY proto/ proto/
COPY services/internal/control-plane/ services/internal/control-plane/

RUN go mod download
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.26.0
RUN go install github.com/githubnemo/CompileDaemon@v1.4.0
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/codex-k8s-control-plane ./services/internal/control-plane/cmd/control-plane

FROM ${GOLANG_IMAGE}

RUN addgroup -S app && adduser -S app -G app

WORKDIR /workspace

COPY --from=builder /go/pkg/mod /go/pkg/mod
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY --from=builder /go/bin/CompileDaemon /usr/local/bin/CompileDaemon
COPY --from=builder /out/codex-k8s-control-plane /usr/local/bin/codex-k8s-control-plane
COPY --from=builder /root/.cache/go-build /tmp/.cache/go-build
COPY services.yaml /app/services.yaml

COPY go.mod go.sum ./
COPY libs/go/ libs/go/
COPY proto/ proto/
COPY deploy/ deploy/
COPY services/internal/control-plane/ services/internal/control-plane/
RUN chown -R app:app /tmp/.cache

ENV CODEXK8S_HOT_RELOAD="true"
ENV GOCACHE="/tmp/.cache/go-build"

USER app

EXPOSE 8081
EXPOSE 9090

ENTRYPOINT ["sh", "-ec", "if [ \"${CODEXK8S_ENV:-production}\" = \"production\" ] || [ \"${CODEXK8S_HOT_RELOAD:-true}\" = \"false\" ]; then exec /usr/local/bin/codex-k8s-control-plane; fi; exec CompileDaemon --build='go build -mod=readonly -o /tmp/codex-k8s-control-plane ./services/internal/control-plane/cmd/control-plane' --command='/tmp/codex-k8s-control-plane' --polling --graceful-kill"]
