FROM mirror.gcr.io/library/node:22-alpine AS web

WORKDIR /web

COPY services/staff/web-console/package.json services/staff/web-console/package-lock.json ./
RUN npm ci

COPY services/staff/web-console/ ./
RUN npm run build

FROM mirror.gcr.io/library/golang:1.25.6-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
COPY libs/go/ libs/go/
COPY proto/ proto/
COPY services/external/api-gateway/ services/external/api-gateway/

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/codex-k8s ./services/external/api-gateway/cmd/api-gateway

FROM mirror.gcr.io/library/alpine:3.20

RUN addgroup -S app && adduser -S app -G app
USER app

COPY --from=builder /out/codex-k8s /usr/local/bin/codex-k8s
COPY --from=web /web/dist /app/web
COPY services/external/api-gateway/api/server/api.yaml /app/api/server/api.yaml

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/codex-k8s"]
