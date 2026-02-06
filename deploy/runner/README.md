# Runner / ARC manifests

Шаблоны для staging self-hosted runner в Kubernetes через GitHub Actions Runner Controller (ARC) Helm charts.

## Файлы

- `namespace.yaml` — namespace контроллера ARC (`actions-runner-system`).
- `runner-namespace.yaml.tpl` — namespace runner scale set.
- `values-ai-staging.yaml.tpl` — values для chart `gha-runner-scale-set`.

Секрет с GitHub PAT в репозитории не хранится. Он создаётся из env (`GITHUB_PAT`) во время bootstrap.

## Применение

```bash
export GITHUB_REPO=owner/repo
export RUNNER_NAMESPACE=actions-runner-staging
export RUNNER_SCALE_SET_NAME=codex-k8s-ai-staging
export RUNNER_MIN=0
export RUNNER_MAX=2
export RUNNER_IMAGE=ghcr.io/actions/actions-runner:latest

kubectl apply -f deploy/runner/namespace.yaml
envsubst < deploy/runner/runner-namespace.yaml.tpl | kubectl apply -f -

# PAT comes from env only; do not commit token manifests.
kubectl -n "${RUNNER_NAMESPACE}" create secret generic gha-runner-scale-set-secret \
  --from-literal=github_token="${GITHUB_PAT}" \
  --dry-run=client -o yaml | kubectl apply -f -

helm upgrade --install gha-rs-controller \
  oci://ghcr.io/actions/actions-runner-controller-charts/gha-runner-scale-set-controller \
  --namespace actions-runner-system --create-namespace

envsubst < deploy/runner/values-ai-staging.yaml.tpl > /tmp/arc-values.yaml
helm upgrade --install "${RUNNER_SCALE_SET_NAME}" \
  oci://ghcr.io/actions/actions-runner-controller-charts/gha-runner-scale-set \
  --namespace "${RUNNER_NAMESPACE}" --create-namespace -f /tmp/arc-values.yaml
```

После установки workflow должен использовать `runs-on: <RUNNER_SCALE_SET_NAME>`.
