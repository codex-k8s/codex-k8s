apiVersion: batch/v1
kind: Job
metadata:
  name: {{ envOr "CODEXK8S_CODEGEN_CHECK_JOB_NAME" "codex-k8s-codegen-check" }}
  namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: codex-k8s
    app.kubernetes.io/component: codegen-check
spec:
  backoffLimit: 0
  ttlSecondsAfterFinished: {{ envOr "CODEXK8S_CODEGEN_CHECK_TTL_SECONDS" "3600" }}
  template:
    metadata:
      labels:
        app.kubernetes.io/name: codex-k8s
        app.kubernetes.io/component: codegen-check
    spec:
      restartPolicy: Never
      containers:
        - name: codegen-check
          image: {{ envOr "CODEXK8S_CODEGEN_CHECK_IMAGE" "golang:1.24-bookworm" }}
          imagePullPolicy: IfNotPresent
          env:
            - name: GIT_TOKEN
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-git-token
                  key: token
            - name: CODEXK8S_GITHUB_REPO
              value: '{{ envOr "CODEXK8S_GITHUB_REPO" "" }}'
            - name: CODEXK8S_BUILD_REF
              value: '{{ envOr "CODEXK8S_BUILD_REF" "main" }}'
            - name: CODEXK8S_NODE_VERSION
              value: '{{ envOr "CODEXK8S_CODEGEN_CHECK_NODE_VERSION" "22.14.0" }}'
          command:
            - bash
            - -ec
            - |
              set -euo pipefail
              apt-get update
              apt-get install -y --no-install-recommends ca-certificates curl git make xz-utils

              curl -fsSL "https://nodejs.org/dist/v${CODEXK8S_NODE_VERSION}/node-v${CODEXK8S_NODE_VERSION}-linux-x64.tar.xz" -o /tmp/node.tar.xz
              tar -xJf /tmp/node.tar.xz -C /tmp
              cp -R "/tmp/node-v${CODEXK8S_NODE_VERSION}-linux-x64/." /usr/local/

              git clone "https://x-access-token:${GIT_TOKEN}@github.com/${CODEXK8S_GITHUB_REPO}.git" /workspace
              cd /workspace
              checkout_ref="$CODEXK8S_BUILD_REF"
              if git rev-parse --verify -q "origin/$CODEXK8S_BUILD_REF^{commit}" >/dev/null 2>&1; then
                checkout_ref="origin/$CODEXK8S_BUILD_REF"
              fi
              git checkout --detach "$checkout_ref"

              npm --prefix services/staff/web-console ci
              make gen-openapi

              git diff --exit-code -- \
                services/external/api-gateway/internal/transport/http/generated/openapi.gen.go \
                services/staff/web-console/src/shared/api/generated
