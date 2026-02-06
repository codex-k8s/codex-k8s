# Runner / ARC manifests

Шаблоны для staging self-hosted runner в Kubernetes через GitHub Actions Runner Controller (ARC) Helm charts.

## Файлы

- `namespace.yaml` — namespace контроллера ARC (`actions-runner-system`).
- `runner-namespace.yaml.tpl` — namespace runner scale set.
- `values-ai-staging.yaml.tpl` — values для chart `gha-runner-scale-set`.
- `staging-deployer-rbac.yaml.tpl` — RBAC для deploy workflow в staging namespace.

Секрет с GitHub PAT в репозитории не хранится. Он создаётся из env (`CODEXK8S_GITHUB_PAT`) во время bootstrap.

## Применение

```bash
export CODEXK8S_GITHUB_REPO=owner/repo
export CODEXK8S_RUNNER_NAMESPACE=actions-runner-staging
export CODEXK8S_RUNNER_SCALE_SET_NAME=codex-k8s-ai-staging
export CODEXK8S_RUNNER_MIN=1
export CODEXK8S_RUNNER_MAX=2
export CODEXK8S_RUNNER_IMAGE=ghcr.io/actions/actions-runner:latest
export CODEXK8S_STAGING_NAMESPACE=codex-k8s-ai-staging

kubectl apply -f deploy/runner/namespace.yaml
envsubst < deploy/runner/runner-namespace.yaml.tpl | kubectl apply -f -

# PAT comes from env only; do not commit token manifests.
kubectl -n "${CODEXK8S_RUNNER_NAMESPACE}" create secret generic gha-runner-scale-set-secret \
  --from-literal=github_token="${CODEXK8S_GITHUB_PAT}" \
  --dry-run=client -o yaml | kubectl apply -f -

helm upgrade --install gha-rs-controller \
  oci://ghcr.io/actions/actions-runner-controller-charts/gha-runner-scale-set-controller \
  --namespace actions-runner-system --create-namespace

envsubst < deploy/runner/values-ai-staging.yaml.tpl > /tmp/arc-values.yaml
helm upgrade --install "${CODEXK8S_RUNNER_SCALE_SET_NAME}" \
  oci://ghcr.io/actions/actions-runner-controller-charts/gha-runner-scale-set \
  --namespace "${CODEXK8S_RUNNER_NAMESPACE}" --create-namespace -f /tmp/arc-values.yaml

envsubst < deploy/runner/staging-deployer-rbac.yaml.tpl | kubectl apply -f -
```

После установки workflow должен использовать `runs-on: <CODEXK8S_RUNNER_SCALE_SET_NAME>`.
