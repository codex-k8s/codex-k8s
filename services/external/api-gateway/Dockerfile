ARG NODE_IMAGE=127.0.0.1:5000/kodex/mirror/node:22-alpine
ARG GOLANG_IMAGE=127.0.0.1:5000/kodex/mirror/golang:1.25.8-alpine

FROM ${NODE_IMAGE} AS web

WORKDIR /web

COPY services/staff/web-console/package.json services/staff/web-console/package-lock.json ./
RUN npm ci

COPY services/staff/web-console/ ./
RUN npm run build

FROM ${GOLANG_IMAGE} AS builder

WORKDIR /workspace

COPY go.mod go.sum ./
COPY libs/go/ libs/go/
COPY proto/ proto/
COPY services/external/api-gateway/ services/external/api-gateway/

RUN go mod download
RUN go install github.com/githubnemo/CompileDaemon@v1.4.0
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/kodex ./services/external/api-gateway/cmd/api-gateway

FROM ${GOLANG_IMAGE}

RUN addgroup -S app && adduser -S app -G app

WORKDIR /workspace

COPY --from=builder /go/pkg/mod /go/pkg/mod
COPY --from=builder /go/bin/CompileDaemon /usr/local/bin/CompileDaemon
COPY --from=builder /out/kodex /usr/local/bin/kodex
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

ENV KODEX_HOT_RELOAD="true"
ENV GOCACHE="/tmp/.cache/go-build"

USER app

EXPOSE 8080

ENTRYPOINT ["sh", "-ec", "if [ \"${KODEX_ENV:-production}\" = \"production\" ] || [ \"${KODEX_HOT_RELOAD:-true}\" = \"false\" ]; then exec /usr/local/bin/kodex; fi; exec CompileDaemon --build='go -C /workspace build -mod=readonly -o /tmp/kodex ./services/external/api-gateway/cmd/api-gateway' --command='/tmp/kodex' --polling --graceful-kill"]
