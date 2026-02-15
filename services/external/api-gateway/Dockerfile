FROM mirror.gcr.io/library/node:22-alpine AS web

WORKDIR /web

COPY services/staff/web-console/package.json services/staff/web-console/package-lock.json ./
RUN npm ci

COPY services/staff/web-console/ ./
RUN npm run build

FROM mirror.gcr.io/library/golang:1.25.6-alpine AS builder

WORKDIR /workspace

COPY go.mod go.sum ./
COPY libs/go/ libs/go/
COPY proto/ proto/
COPY services/external/api-gateway/ services/external/api-gateway/

RUN go mod download
RUN go install github.com/githubnemo/CompileDaemon@v1.4.0
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/codex-k8s ./services/external/api-gateway/cmd/api-gateway

FROM mirror.gcr.io/library/golang:1.25.6-alpine

RUN addgroup -S app && adduser -S app -G app

WORKDIR /workspace

COPY --from=builder /go/pkg/mod /go/pkg/mod
COPY --from=builder /go/bin/CompileDaemon /usr/local/bin/CompileDaemon
COPY --from=builder /out/codex-k8s /usr/local/bin/codex-k8s
COPY --from=builder /root/.cache/go-build /tmp/.cache/go-build
COPY --from=web /web/dist /app/web
COPY services/external/api-gateway/api/server/api.yaml /app/api/server/api.yaml

# Copy hot-reload sources from the builder stage to guarantee runtime parity
# even when context-only COPY cache is stale.
COPY --from=builder /workspace/go.mod /workspace/go.sum ./
COPY --from=builder /workspace/libs/go/ libs/go/
COPY --from=builder /workspace/proto/ proto/
COPY --from=builder /workspace/services/external/api-gateway/ services/external/api-gateway/
RUN chown -R app:app /tmp/.cache

ENV CODEXK8S_HOT_RELOAD="true"
ENV GOCACHE="/tmp/.cache/go-build"

USER app

EXPOSE 8080

ENTRYPOINT ["sh", "-ec", "if [ \"${CODEXK8S_ENV:-ai-staging}\" = \"production\" ] || [ \"${CODEXK8S_HOT_RELOAD:-true}\" = \"false\" ]; then exec /usr/local/bin/codex-k8s; fi; exec CompileDaemon --build='go build -mod=readonly -o /tmp/codex-k8s ./services/external/api-gateway/cmd/api-gateway' --command='/tmp/codex-k8s' --polling --graceful-kill"]
