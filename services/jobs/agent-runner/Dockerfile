FROM mirror.gcr.io/library/golang:1.25.6-bookworm AS go-builder

WORKDIR /workspace

COPY go.mod go.sum ./
COPY libs ./libs
COPY proto ./proto
COPY tools ./tools
COPY services/jobs/agent-runner ./services/jobs/agent-runner

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /out/codex-k8s-agent-runner ./services/jobs/agent-runner/cmd/agent-runner

FROM mirror.gcr.io/library/node:22-bookworm

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
COPY --from=go-builder /out/codex-k8s-agent-runner /usr/local/bin/codex-k8s-agent-runner
COPY services/jobs/agent-runner/scripts/bootstrap_tools.sh /tmp/bootstrap_tools.sh

RUN chmod +x /usr/local/bin/codex-k8s-agent-runner /tmp/bootstrap_tools.sh \
  && npm install -g @openai/codex \
  && /tmp/bootstrap_tools.sh \
  && rm -f /tmp/bootstrap_tools.sh

ENTRYPOINT ["/usr/local/bin/codex-k8s-agent-runner"]
