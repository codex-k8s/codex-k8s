ARG GOLANG_IMAGE=127.0.0.1:5000/kodex/mirror/golang:1.25.8-alpine

FROM ${GOLANG_IMAGE} AS builder

WORKDIR /workspace

COPY go.mod go.sum ./
COPY libs/go/ libs/go/
COPY proto/ proto/
COPY services/jobs/worker/ services/jobs/worker/

RUN go mod download
RUN go install github.com/githubnemo/CompileDaemon@v1.4.0
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/kodex-worker ./services/jobs/worker/cmd/worker

FROM ${GOLANG_IMAGE}

RUN addgroup -S app && adduser -S app -G app

WORKDIR /workspace

COPY --from=builder /go/pkg/mod /go/pkg/mod
COPY --from=builder /go/bin/CompileDaemon /usr/local/bin/CompileDaemon
COPY --from=builder /out/kodex-worker /usr/local/bin/kodex-worker
COPY --from=builder /root/.cache/go-build /tmp/.cache/go-build

COPY go.mod go.sum ./
COPY libs/go/ libs/go/
COPY proto/ proto/
COPY services/jobs/worker/ services/jobs/worker/
RUN chown -R app:app /tmp/.cache

ENV KODEX_HOT_RELOAD="true"
ENV GOCACHE="/tmp/.cache/go-build"

USER app

ENTRYPOINT ["sh", "-ec", "if [ \"${KODEX_ENV:-production}\" = \"production\" ] || [ \"${KODEX_HOT_RELOAD:-true}\" = \"false\" ]; then exec /usr/local/bin/kodex-worker; fi; exec CompileDaemon --build='go -C /workspace build -mod=readonly -o /tmp/kodex-worker ./services/jobs/worker/cmd/worker' --command='/tmp/kodex-worker' --polling --graceful-kill"]
