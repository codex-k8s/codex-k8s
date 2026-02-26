ARG GOLANG_IMAGE=127.0.0.1:5000/codex-k8s/mirror/golang:1.25.6-alpine

FROM ${GOLANG_IMAGE} AS builder

WORKDIR /workspace

COPY go.mod go.sum ./
COPY libs/go/ libs/go/
COPY proto/ proto/
COPY services/jobs/worker/ services/jobs/worker/

RUN go mod download
RUN go install github.com/githubnemo/CompileDaemon@v1.4.0
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/codex-k8s-worker ./services/jobs/worker/cmd/worker

FROM ${GOLANG_IMAGE}

RUN addgroup -S app && adduser -S app -G app

WORKDIR /workspace

COPY --from=builder /go/pkg/mod /go/pkg/mod
COPY --from=builder /go/bin/CompileDaemon /usr/local/bin/CompileDaemon
COPY --from=builder /out/codex-k8s-worker /usr/local/bin/codex-k8s-worker
COPY --from=builder /root/.cache/go-build /tmp/.cache/go-build

COPY go.mod go.sum ./
COPY libs/go/ libs/go/
COPY proto/ proto/
COPY services/jobs/worker/ services/jobs/worker/
RUN chown -R app:app /tmp/.cache

ENV CODEXK8S_HOT_RELOAD="true"
ENV GOCACHE="/tmp/.cache/go-build"

USER app

ENTRYPOINT ["sh", "-ec", "if [ \"${CODEXK8S_ENV:-production}\" = \"production\" ] || [ \"${CODEXK8S_HOT_RELOAD:-true}\" = \"false\" ]; then exec /usr/local/bin/codex-k8s-worker; fi; exec CompileDaemon --build='go -C /workspace build -mod=readonly -o /tmp/codex-k8s-worker ./services/jobs/worker/cmd/worker' --command='/tmp/codex-k8s-worker' --polling --graceful-kill"]
