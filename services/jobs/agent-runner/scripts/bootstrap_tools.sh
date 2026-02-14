#!/usr/bin/env bash
set -euo pipefail

export DEBIAN_FRONTEND=noninteractive
export PATH="/usr/local/go/bin:/usr/local/bin:${PATH:-/usr/bin:/bin}"
export GOBIN="${GOBIN:-/usr/local/bin}"

apt-get update -y
apt-get install -y --no-install-recommends \
  ca-certificates curl \
  git jq gh kubernetes-client bash openssh-client make python3 \
  unzip zip ripgrep
rm -rf /var/lib/apt/lists/*

: "${PROTOC_VERSION:=32.1}"
install_protoc_from_apt() {
  apt-get update -y
  apt-get install -y --no-install-recommends protobuf-compiler
  rm -rf /var/lib/apt/lists/*
}

install_protoc_from_release() {
  local tmp_dir
  tmp_dir="$(mktemp -d)"

  if ! curl -fL --retry 5 --retry-all-errors --retry-delay 2 --retry-max-time 120 \
    -o "${tmp_dir}/protoc.zip" \
    "https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip"; then
    rm -rf "${tmp_dir}"
    return 1
  fi

  if ! unzip -tq "${tmp_dir}/protoc.zip" >/dev/null; then
    rm -rf "${tmp_dir}"
    return 1
  fi

  unzip -qo "${tmp_dir}/protoc.zip" -d "${tmp_dir}"
  install -m 0755 "${tmp_dir}/bin/protoc" /usr/local/bin/protoc
  cp -r "${tmp_dir}/include/." /usr/local/include/
  rm -rf "${tmp_dir}"
}

if ! install_protoc_from_release; then
  install_protoc_from_apt
fi

if [[ -x /usr/local/go/bin/go && ! -e /usr/local/bin/go ]]; then
  ln -s /usr/local/go/bin/go /usr/local/bin/go
fi
if [[ -x /usr/local/go/bin/gofmt && ! -e /usr/local/bin/gofmt ]]; then
  ln -s /usr/local/go/bin/gofmt /usr/local/bin/gofmt
fi

: "${PROTOC_GEN_GO_VERSION:=v1.36.10}"
: "${PROTOC_GEN_GO_GRPC_VERSION:=v1.5.1}"
: "${OAPI_CODEGEN_VERSION:=v2.5.0}"
: "${GOLANGCI_LINT_VERSION:=v1.64.8}"
: "${DUPL_VERSION:=v1.0.0}"

GO111MODULE=on GOBIN="${GOBIN}" go install "google.golang.org/protobuf/cmd/protoc-gen-go@${PROTOC_GEN_GO_VERSION}"
GO111MODULE=on GOBIN="${GOBIN}" go install "google.golang.org/grpc/cmd/protoc-gen-go-grpc@${PROTOC_GEN_GO_GRPC_VERSION}"
GO111MODULE=on GOBIN="${GOBIN}" go install "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@${OAPI_CODEGEN_VERSION}"
GO111MODULE=on GOBIN="${GOBIN}" go install "github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}"
GO111MODULE=on GOBIN="${GOBIN}" go install "github.com/mibk/dupl@${DUPL_VERSION}"

npm install -g @hey-api/openapi-ts
