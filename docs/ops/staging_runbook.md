---
doc_id: OPS-CK8S-STAGING-0001
type: runbook
title: "Staging Runbook (MVP)"
status: draft
owner_role: SRE
created_at: 2026-02-09
updated_at: 2026-02-12
related_issues: [1]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
---

# Staging Runbook (MVP)

Цель: минимальный набор проверок и действий для ежедневного деплоя и ручного smoke/regression на staging.

## Быстрый smoke (на сервере)

Предпосылки:
- есть доступ по SSH на staging host (Ubuntu 24.04);
- на host установлен `kubectl` (k3s) и кластер поднят;
- namespace по умолчанию: `codex-k8s-ai-staging`.

Команда:

```bash
export CODEXK8S_STAGING_NAMESPACE="codex-k8s-ai-staging"
export CODEXK8S_STAGING_DOMAIN="staging.codex-k8s.dev"
bash deploy/scripts/staging_smoke.sh
```

Ожидаемо:
- rollout `codex-k8s-control-plane`, `codex-k8s`, `codex-k8s-worker`, `oauth2-proxy`, `codex-k8s-web-console` успешен;
- последний `codex-k8s-migrate-*` job completed;
- `/healthz`, `/readyz`, `/metrics` доступны через `kubectl port-forward`;
- `codex-k8s-staging-tls` secret существует;
- webhook endpoint отвечает **401** на invalid signature (и не редиректит в OAuth).

Порядок выкладки staging:
- `PostgreSQL -> migrations -> control-plane -> api-gateway -> frontend`.
- Зависимости между сервисами ожидаются через `initContainers` в манифестах.

## Проверка внешних портов (снаружи)

Требование staging (MVP):
- извне доступны только `22`, `80`, `443`.

Проверка с хоста разработчика:

```bash
host="staging.codex-k8s.dev"
for p in 22 80 443 6443 5000 10250 10254 8443; do
  echo -n "$p "
  if timeout 3 bash -lc "</dev/tcp/$host/$p" >/dev/null 2>&1; then echo open; else echo closed; fi
done
```

## Полезные команды kubectl

```bash
ns="codex-k8s-ai-staging"
kubectl -n "$ns" get pods -o wide
kubectl -n "$ns" logs deploy/codex-k8s --tail=200
kubectl -n "$ns" logs deploy/codex-k8s-control-plane --tail=200
kubectl -n "$ns" logs deploy/codex-k8s-worker --tail=200
kubectl -n "$ns" get ingress
kubectl -n "$ns" describe ingress codex-k8s
kubectl -n "$ns" get certificate,order,challenge -A

# Full-env run namespaces (S2 Day3 baseline)
kubectl get ns -l codex-k8s.dev/managed-by=codex-k8s-worker,codex-k8s.dev/namespace-purpose=run
for run_ns in $(kubectl get ns -l codex-k8s.dev/managed-by=codex-k8s-worker,codex-k8s.dev/namespace-purpose=run -o jsonpath='{.items[*].metadata.name}'); do
  echo "=== ${run_ns} ==="
  kubectl -n "${run_ns}" get sa,role,rolebinding,resourcequota,limitrange,job,pod
done

# Day4: проверить env wiring и логи agent-runner job
for run_ns in $(kubectl get ns -l codex-k8s.dev/managed-by=codex-k8s-worker,codex-k8s.dev/namespace-purpose=run -o jsonpath='{.items[*].metadata.name}'); do
  echo "=== ${run_ns} agent jobs ==="
  kubectl -n "${run_ns}" get jobs,pods
  kubectl -n "${run_ns}" get pod -l app.kubernetes.io/name=codex-k8s-run \
    -o jsonpath='{range .items[*].spec.containers[*].env[*]}{.name}{"\n"}{end}' \
    | grep -E 'CODEXK8S_OPENAI_API_KEY|CODEXK8S_GIT_BOT_TOKEN|CODEXK8S_GIT_BOT_USERNAME|CODEXK8S_GIT_BOT_MAIL|CODEXK8S_AGENT_DISPLAY_NAME' || true
done

# Legacy runtime keys must not appear after Day3 rollout
kubectl get ns -o json | grep -E 'codexk8s.io/(managed-by|namespace-purpose|runtime-mode|project-id|run-id|correlation-id)' || true
```

## Типовые проблемы

### Web UI не открывается / "ui upstream unavailable"
- Проверить, что `codex-k8s-web-console` pod Running и port `5173` открыт в cluster.
- Проверить NetworkPolicy baseline (должен быть allow до web-console).

### OAuth2 callback не проходит
- В GitHub OAuth App callback должен быть:
  - `https://<CODEXK8S_STAGING_DOMAIN>/oauth2/callback`

### Webhook не доходит
- Убедиться, что path пропущен без auth:
  - `oauth2-proxy --skip-auth-regex=^/api/v1/webhooks/.*`
- Проверить `CODEXK8S_GITHUB_WEBHOOK_SECRET` совпадает с секретом вебхука в GitHub.
