ARG GOLANG_IMAGE=127.0.0.1:5000/kodex/mirror/golang:1.25.8-bookworm
ARG NODE_IMAGE=127.0.0.1:5000/kodex/mirror/node:22-bookworm

FROM ${GOLANG_IMAGE} AS go-builder

WORKDIR /workspace

COPY go.mod go.sum ./
COPY libs ./libs
COPY proto ./proto
COPY tools ./tools
COPY services/jobs/agent-runner ./services/jobs/agent-runner

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /out/kodex-agent-runner ./services/jobs/agent-runner/cmd/agent-runner

FROM ${NODE_IMAGE}

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

ENV DEBIAN_FRONTEND=noninteractive \
    TZ=UTC \
    PATH=/usr/local/go/bin:/usr/local/bin:/usr/bin:/bin

RUN apt-get update -y \
  && apt-get install -y --no-install-recommends \
      ca-certificates curl git jq bash \
      openssh-client make python3 ripgrep \
  && rm -rf /var/lib/apt/lists/*

COPY --from=go-builder /usr/local/go /usr/local/go
COPY --from=go-builder /out/kodex-agent-runner /usr/local/bin/kodex-agent-runner
COPY services/jobs/agent-runner/scripts/bootstrap_tools.sh /tmp/bootstrap_tools.sh

RUN chmod +x /usr/local/bin/kodex-agent-runner /tmp/bootstrap_tools.sh \
  && npm install -g @openai/codex \
  && /tmp/bootstrap_tools.sh \
  && rm -f /tmp/bootstrap_tools.sh

ENTRYPOINT ["/usr/local/bin/kodex-agent-runner"]
