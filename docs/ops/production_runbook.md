---
doc_id: OPS-CK8S-PRODUCTION-0001
type: runbook
title: "Production Runbook (MVP)"
status: active
owner_role: SRE
created_at: 2026-02-09
updated_at: 2026-02-21
related_issues: [1]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-19-full-docset"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-19
---

# Production Runbook (MVP)

Цель: минимальный набор проверок и действий для ежедневного деплоя и ручного smoke/regression на production.

## Быстрый ручной smoke (на сервере)

Предпосылки:
- есть доступ по SSH на production host (Ubuntu 24.04);
- на host установлен `kubectl` (k3s) и кластер поднят;
- namespace по умолчанию: `codex-k8s-prod`.

Базовые команды:

```bash
export CODEXK8S_PRODUCTION_NAMESPACE="codex-k8s-prod"
export CODEXK8S_PRODUCTION_DOMAIN="platform.codex-k8s.dev"

kubectl -n "$CODEXK8S_PRODUCTION_NAMESPACE" get pods -o wide
kubectl -n "$CODEXK8S_PRODUCTION_NAMESPACE" get deploy,job,ingress
kubectl -n "$CODEXK8S_PRODUCTION_NAMESPACE" logs deploy/codex-k8s --tail=200
kubectl -n "$CODEXK8S_PRODUCTION_NAMESPACE" logs deploy/codex-k8s-worker --tail=200
```

Ожидаемо:
- rollout `codex-k8s-control-plane`, `codex-k8s` (api-gateway + staff UI), `codex-k8s-worker`, `oauth2-proxy` успешен;
- последний `codex-k8s-migrate-*` job completed;
- `/healthz`, `/readyz`, `/metrics` доступны через `kubectl port-forward`;
- `codex-k8s-production-tls` secret существует;
- при включённом TLS reuse в служебном namespace (`codex-k8s-system`) существует `codex-k8s-tls-<hash>` secret;
- webhook endpoint отвечает **401** на invalid signature (и не редиректит в OAuth).

Порядок выкладки production:
- `PostgreSQL -> migrations -> control-plane -> api-gateway -> frontend`.
- Зависимости между сервисами ожидаются через `initContainers` в манифестах.

## Проверка внешних портов (снаружи)

Требование production (MVP):
- извне доступны только `22`, `80`, `443`.

Проверка с хоста разработчика:

```bash
host="platform.codex-k8s.dev"
for p in 22 80 443 6443 5000 10250 10254 8443; do
  echo -n "$p "
  if timeout 3 bash -lc "</dev/tcp/$host/$p" >/dev/null 2>&1; then echo open; else echo closed; fi
done
```

## Полезные команды kubectl

```bash
ns="codex-k8s-prod"
kubectl -n "$ns" get pods -o wide
kubectl -n "$ns" logs deploy/codex-k8s --tail=200
kubectl -n "$ns" logs deploy/codex-k8s-control-plane --tail=200
kubectl -n "$ns" logs deploy/codex-k8s-worker --tail=200
kubectl -n "$ns" get ingress
kubectl -n "$ns" describe ingress codex-k8s
kubectl -n "$ns" get certificate,order,challenge -A

# TLS reuse store (best-effort, может быть пусто в самый первый деплой)
kubectl -n codex-k8s-system get secrets | grep '^codex-k8s-tls-' || true

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

## Registry GC (автоматический)

- В production/non-ai окружениях включён `CronJob` `codex-k8s-registry-gc`.
- Расписание по умолчанию: ежедневно в `03:17 UTC`.
- Job делает `scale deployment/codex-k8s-registry 1 -> 0`, выполняет `registry garbage-collect --delete-untagged`, затем возвращает `replicas=1`.
- Для init-контейнера GC helper по умолчанию используется `alpine/k8s:1.32.2` (можно переопределить через `CODEXK8S_KUBECTL_IMAGE`).

Проверка статуса:

```bash
ns="codex-k8s-prod"
kubectl -n "$ns" get cronjob codex-k8s-registry-gc
kubectl -n "$ns" get jobs -l app.kubernetes.io/name=codex-k8s-registry-gc
kubectl -n "$ns" logs job/<gc_job_name> --tail=200
```

Форсированный запуск вне расписания:

```bash
ns="codex-k8s-prod"
kubectl -n "$ns" create job --from=cronjob/codex-k8s-registry-gc codex-k8s-registry-gc-manual-$(date +%s)
kubectl -n "$ns" get jobs -l app.kubernetes.io/name=codex-k8s-registry-gc
```

## Cleanup heavy JSON payloads (автоматический)

- Control-plane выполняет hourly cleanup heavy JSON-полей для старых записей (по умолчанию `7` дней):
  - `agent_runs.agent_logs_json`;
  - `agent_sessions.session_json`, `agent_sessions.codex_cli_session_json`;
  - `runtime_deploy_tasks.logs_json`.
- Retention настраивается через:
  - `CODEXK8S_RUN_HEAVY_FIELDS_RETENTION_DAYS` (основной ключ);
  - `CODEXK8S_RUN_AGENT_LOGS_RETENTION_DAYS` (legacy fallback).

## Типовые проблемы

### Web UI не открывается / "ui upstream unavailable"
- Проверить, что `codex-k8s-web-console` pod Running и port `5173` открыт в cluster.
- Проверить NetworkPolicy baseline (должен быть allow до web-console).

### OAuth2 callback не проходит
- В GitHub OAuth App callback должен быть:
  - `https://<CODEXK8S_PRODUCTION_DOMAIN>/oauth2/callback`

### Webhook не доходит
- Убедиться, что path пропущен без auth:
  - `oauth2-proxy --skip-auth-regex=^/api/v1/webhooks/.*`
- Проверить `CODEXK8S_GITHUB_WEBHOOK_SECRET` совпадает с секретом вебхука в GitHub.

### TLS не выпускается (HTTP-01) / cert-manager молчит
- Убедиться, что `CODEXK8S_PRODUCTION_DOMAIN` резолвится в production host IP.
- Если это первый выпуск TLS, runtime-deploy использует echo-probe (HTTP) до включения issuer:
  - проверить `kubectl -n "$CODEXK8S_PRODUCTION_NAMESPACE" get deploy,svc,ingress | grep echo-probe`;
  - проверить логи `kubectl -n "$CODEXK8S_PRODUCTION_NAMESPACE" logs deploy/codex-k8s-control-plane --tail=200`.

### Build падает с `MANIFEST_UNKNOWN` при `retrieving image from cache`
- Симптом: Kaniko падает на base image с логом вида `Error while retrieving image from cache ... MANIFEST_UNKNOWN`.
- Причина: в registry мог остаться stale mirror/cache state после cleanup/GC (тег виден, но digest манифест недоступен).
- Текущее безопасное значение по умолчанию: `CODEXK8S_KANIKO_CACHE_ENABLED=false`.
- Если cache включали вручную и снова получили `MANIFEST_UNKNOWN`:
  - переключить `CODEXK8S_KANIKO_CACHE_ENABLED=false` в `codex-k8s-runtime`;
  - убедиться, что `codex-k8s-control-plane` подтянул значение после rollout;
  - повторить deploy.
- Дополнительно:
  - mirror шаг выполняет health-check и ремонтирует stale mirror;
  - при cache-related `MANIFEST_UNKNOWN` build автоматически ретраится без cache.
