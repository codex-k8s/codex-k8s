#!/usr/bin/env bash
set -euo pipefail

export DEBIAN_FRONTEND=noninteractive

apt-get update -y
apt-get install -y --no-install-recommends \
  ca-certificates curl \
  git jq bash openssh-client make python3 \
  unzip zip ripgrep
rm -rf /var/lib/apt/lists/*

: "${PROTOC_VERSION:=32.1}"
TMP_PROTOC_DIR="$(mktemp -d)"
curl -sSL -o "${TMP_PROTOC_DIR}/protoc.zip" \
  "https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip"
unzip -qo "${TMP_PROTOC_DIR}/protoc.zip" -d "${TMP_PROTOC_DIR}"
install -m 0755 "${TMP_PROTOC_DIR}/bin/protoc" /usr/local/bin/protoc
cp -r "${TMP_PROTOC_DIR}/include/." /usr/local/include/
rm -rf "${TMP_PROTOC_DIR}"

: "${PROTOC_GEN_GO_VERSION:=v1.36.10}"
: "${PROTOC_GEN_GO_GRPC_VERSION:=v1.5.1}"
: "${OAPI_CODEGEN_VERSION:=v2.5.0}"
: "${GOLANGCI_LINT_VERSION:=v1.64.8}"
: "${DUPL_VERSION:=v1.0.0}"

GO111MODULE=on GOBIN=/usr/local/bin go install "google.golang.org/protobuf/cmd/protoc-gen-go@${PROTOC_GEN_GO_VERSION}"
GO111MODULE=on GOBIN=/usr/local/bin go install "google.golang.org/grpc/cmd/protoc-gen-go-grpc@${PROTOC_GEN_GO_GRPC_VERSION}"
GO111MODULE=on GOBIN=/usr/local/bin go install "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@${OAPI_CODEGEN_VERSION}"
GO111MODULE=on GOBIN=/usr/local/bin go install "github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}"
GO111MODULE=on GOBIN=/usr/local/bin go install "github.com/mibk/dupl@${DUPL_VERSION}"

npm install -g @hey-api/openapi-ts
