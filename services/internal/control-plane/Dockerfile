ARG GOLANG_IMAGE=127.0.0.1:5000/kodex/mirror/golang:1.25.8-alpine

FROM ${GOLANG_IMAGE} AS builder

WORKDIR /workspace

COPY go.mod go.sum ./
COPY libs/go/ libs/go/
COPY proto/ proto/
COPY services/internal/control-plane/ services/internal/control-plane/

RUN go mod download
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.26.0
RUN go install github.com/githubnemo/CompileDaemon@v1.4.0
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/kodex-control-plane ./services/internal/control-plane/cmd/control-plane

FROM ${GOLANG_IMAGE}

RUN addgroup -S app && adduser -S app -G app

WORKDIR /workspace

COPY --from=builder /go/pkg/mod /go/pkg/mod
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY --from=builder /go/bin/CompileDaemon /usr/local/bin/CompileDaemon
COPY --from=builder /out/kodex-control-plane /usr/local/bin/kodex-control-plane
COPY --from=builder /root/.cache/go-build /tmp/.cache/go-build
COPY services.yaml /app/services.yaml

COPY go.mod go.sum ./
COPY libs/go/ libs/go/
COPY proto/ proto/
COPY deploy/ deploy/
COPY services/internal/control-plane/ services/internal/control-plane/
RUN chown -R app:app /tmp/.cache

ENV KODEX_HOT_RELOAD="true"
ENV GOCACHE="/tmp/.cache/go-build"

USER app

EXPOSE 8081
EXPOSE 9090

ENTRYPOINT ["sh", "-ec", "if [ \"${KODEX_ENV:-production}\" = \"production\" ] || [ \"${KODEX_HOT_RELOAD:-true}\" = \"false\" ]; then exec /usr/local/bin/kodex-control-plane; fi; exec CompileDaemon --build='go -C /workspace build -mod=readonly -o /tmp/kodex-control-plane ./services/internal/control-plane/cmd/control-plane' --command='/tmp/kodex-control-plane' --polling --graceful-kill"]
