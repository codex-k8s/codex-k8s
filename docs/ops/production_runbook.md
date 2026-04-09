---
doc_id: OPS-CK8S-PRODUCTION-0001
type: runbook
title: "Production Runbook (MVP)"
status: active
owner_role: SRE
created_at: 2026-02-09
updated_at: 2026-03-14
related_issues: [1, 256, 395, 461]
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
- namespace по умолчанию: `kodex-prod`.

Базовые команды:

```bash
export KODEX_PRODUCTION_NAMESPACE="kodex-prod"
export KODEX_PRODUCTION_DOMAIN="platform.kodex.works"

kubectl -n "$KODEX_PRODUCTION_NAMESPACE" get pods -o wide
kubectl -n "$KODEX_PRODUCTION_NAMESPACE" get deploy,job,ingress
kubectl -n "$KODEX_PRODUCTION_NAMESPACE" logs deploy/kodex --tail=200
kubectl -n "$KODEX_PRODUCTION_NAMESPACE" logs deploy/kodex-worker --tail=200
```

Ожидаемо:
- rollout `kodex-control-plane`, `kodex` (api-gateway + staff UI), `kodex-worker`, `oauth2-proxy` успешен;
- последний `kodex-migrate-*` job completed;
- `/healthz`, `/readyz`, `/metrics` доступны через `kubectl port-forward`;
- `kodex-production-tls` secret существует;
- при включённом TLS reuse в служебном namespace (`kodex-system`) существует `kodex-tls-<hash>` secret;
- webhook endpoint отвечает **401** на invalid signature (и не редиректит в OAuth).

Порядок выкладки production:
- `PostgreSQL -> migrations -> control-plane -> api-gateway -> frontend`.
- Зависимости между сервисами ожидаются через `initContainers` в манифестах.

## QA acceptance через Kubernetes service DNS (S7-E14)

Для новых или изменённых HTTP-ручек QA acceptance не должна зависеть только от browser/OAuth flow через Ingress.
Базовый путь проверки: обращаться к сервису по Kubernetes DNS внутри namespace и фиксировать отдельный evidence bundle.

Обязательный минимум evidence по каждой применимой ручке:
- namespace и service FQDN;
- точная команда (`getent`/`curl`);
- HTTP status;
- краткий excerpt headers/body;
- timestamp и ссылка на issue/PR/checklist;
- при fail — `kubectl`-диагностика по сервису и pod'ам.

Канонический формат service DNS по Kubernetes:
- короткое имя `<service>` работает только внутри того же namespace и зависит от search path pod'а;
- для QA evidence использовать явный FQDN `<service>.<namespace>.svc.cluster.local`, чтобы было видно, какой namespace/service проверялся.

Минимальный шаблон проверки:

```bash
ns="<runtime-namespace>" # например kodex-dev-1; не подставлять production default
svc="kodex"
fqdn="${svc}.${ns}.svc.cluster.local"
base="http://${fqdn}"
ts="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
issue_or_pr="<issue-or-pr-url>"
checklist_ref="<checklist-url-or-path>"

getent hosts "$fqdn"
curl -sS -o /tmp/health.out -D /tmp/health.headers -w '%{http_code}\n' "$base/healthz"
curl -sS -o /tmp/authme.out -D /tmp/authme.headers -w '%{http_code}\n' "$base/api/v1/auth/me"
curl -sS -o /tmp/webhook.out -D /tmp/webhook.headers -w '%{http_code}\n' -X POST "$base/api/v1/webhooks/github"
printf 'timestamp=%s\nissue_or_pr=%s\nchecklist=%s\n' "$ts" "$issue_or_pr" "$checklist_ref"
```

Интерпретация baseline:
- `GET /healthz` ожидаемо возвращает `200`;
- защищённая ручка без credentials ожидаемо возвращает `401` или `403`;
- invalid webhook request ожидаемо возвращает `400` или `401` по контракту сервиса;
- наличие только browser redirect/OAuth-path не считается достаточным acceptance evidence для ручки.

Если DNS-path проверка падает, добавить базовую диагностику:

```bash
kubectl -n "$ns" get svc,pods -o wide
kubectl -n "$ns" logs deploy/kodex --tail=200
kubectl -n "$ns" get events --sort-by=.lastTimestamp | tail -n 80
```

Примечание:
- Ingress/TLS/browser smoke остаются отдельной проверкой для production;
- для acceptance новых/изменённых HTTP-ручек первичным evidence считается именно service DNS path из runtime namespace.

## MCP interaction observability smoke (S10-E05)

Использовать для candidate/prod contour после rollout built-in MCP user interactions.

Минимальный evidence bundle:
- namespace и service DNS/FQDN;
- `/metrics` excerpt для `control-plane` и `api-gateway`;
- `/metrics` excerpt для `worker` через service DNS, а для hot-reload candidate без materialized `kodex-worker` Service — через `kubectl port-forward`;
- хотя бы один callback probe с ожидаемым `401`/`400`, чтобы подтвердить, что edge-метрики инкрементируются даже на rejected ingress;
- логи `control-plane` и `worker` после rollout без repeated crash/restart pattern.

Команды:

```bash
ns="<runtime-namespace>"
control_plane_fqdn="kodex-control-plane.${ns}.svc.cluster.local"
worker_fqdn="kodex-worker.${ns}.svc.cluster.local"
api_fqdn="kodex.${ns}.svc.cluster.local"
worker_metrics_url="http://${worker_fqdn}:8082/metrics"

if ! kubectl -n "$ns" get svc kodex-worker >/dev/null 2>&1; then
  echo "kodex-worker Service is not materialized in this namespace; using port-forward fallback"
  kubectl -n "$ns" port-forward deploy/kodex-worker 18082:8082 >/tmp/kodex-worker-port-forward.log 2>&1 &
  worker_port_forward_pid=$!
  trap 'kill "${worker_port_forward_pid}" >/dev/null 2>&1 || true' EXIT
  worker_metrics_url="http://127.0.0.1:18082/metrics"
fi

curl -fsS "http://${control_plane_fqdn}:8081/metrics" | grep 'kodex_interaction' || true
curl -fsS "${worker_metrics_url}" | grep 'kodex_interaction_dispatch_' || true
curl -fsS "http://${api_fqdn}/metrics" | grep 'kodex_interaction_callback_' || true

curl -sS -o /tmp/interaction-callback.out -D /tmp/interaction-callback.headers -w '%{http_code}\n' \
  -H 'Content-Type: application/json' \
  -X POST "http://${api_fqdn}/api/v1/mcp/interactions/callback" \
  --data '{"interaction_id":"smoke-probe","callback_kind":"decision_response"}'

curl -fsS "http://${api_fqdn}/metrics" | grep 'kodex_interaction_callback_' || true
kubectl -n "$ns" logs deploy/kodex-control-plane --tail=120
kubectl -n "$ns" logs deploy/kodex-worker --tail=120
```

Интерпретация:
- `control-plane` должен отдавать `/metrics` без 5xx и без `promhttp_metric_handler_errors_total` роста по interaction collector path;
- `worker` должен отдавать `/metrics` и публиковать `kodex_interaction_dispatch_attempt_total` / `kodex_interaction_dispatch_retry_scheduled_total`;
- service DNS остаётся основным evidence path для fully applied namespace/prod; `port-forward` fallback допустим только для hot-reload candidate, где `kodex-worker` Service ещё не materialized;
- `api-gateway` после probe должен показать `kodex_interaction_callback_requests_total{callback_kind="unknown",classification="error"}` и histogram `kodex_interaction_callback_duration_seconds`;
- отсутствие interaction-specific samples в `control-plane` допустимо до первого реального tool/callback traffic, но endpoint и collector registration должны оставаться стабильными;
- repeated restart loops, repeated collector errors или невозможность прочитать `/metrics` считаются rollout blocker до `run:qa`.

## Postdeploy checklist (S6 continuity)

Для postdeploy-контуров (`run:postdeploy` / `run:ops`) используйте минимальный gate:

```bash
ns="${KODEX_PRODUCTION_NAMESPACE:-kodex-prod}"

kubectl -n "$ns" get pods,deploy,job -o wide
kubectl -n "$ns" logs deploy/kodex-control-plane --tail=200
kubectl -n "$ns" logs deploy/kodex-worker --tail=200
kubectl -n "$ns" logs deploy/kodex --tail=200
kubectl -n "$ns" get events --sort-by=.lastTimestamp | tail -n 120
```

Интерпретация:
- единичные `startup probe failed` в первые минуты rollout допустимы, если pod стабильно переходит в `Running/Ready`;
- рост restart count, повторяющиеся probe-failures или `CrashLoopBackOff` считаются деградацией и требуют rollback decision;
- paging принимается только по user-impact сигналам (availability/error/critical latency), noise-сигналы должны быть подавлены через `for`/`keep_firing_for`.

Операционный handover для Sprint S6:
- `docs/ops/handovers/s6/postdeploy_ops_handover.md`;
- `docs/ops/handovers/s6/operational_baseline.md`;
- `docs/delivery/epics/s6/epic-s6-day10-postdeploy-review.md`.

## Ops baseline checklist (S6 Day11)

Использовать после `run:ops` (Issue `#265`) как обязательный gate для production decision.

| Проверка | Условие PASS | Решение при FAIL |
|---|---|---|
| Availability | success-rate не ниже 99.5% в 10m окне | page incident + rollback assessment |
| API latency p95 | не выше 2.0s в 10m окне | mitigation, затем rollback decision при отсутствии снижения |
| 5xx error rate | не выше 3% в 5m окне | немедленный incident triage, ограничение blast radius |
| Worker backlog | нет устойчивого роста >30m | queue/retry analysis, capacity mitigation |
| Postgres health | `pg_isready` стабильно OK | DB incident procedure |

Эскалация:
- `ticket`: burn-rate > `2x` на окнах `1h + 6h`;
- `page`: burn-rate > `6x` на окнах `5m + 1h`;
- anti-noise: для page-сигналов использовать `for >= 5m` и `keep_firing_for >= 5m`.

## Проверка внешних портов (снаружи)

Требование production (MVP):
- извне доступны только `22`, `80`, `443`.

Проверка с хоста разработчика:

```bash
host="platform.kodex.works"
for p in 22 80 443 6443 5000 10250 10254 8443; do
  echo -n "$p "
  if timeout 3 bash -lc "</dev/tcp/$host/$p" >/dev/null 2>&1; then echo open; else echo closed; fi
done
```

## Полезные команды kubectl

```bash
ns="kodex-prod"
kubectl -n "$ns" get pods -o wide
kubectl -n "$ns" logs deploy/kodex --tail=200
kubectl -n "$ns" logs deploy/kodex-control-plane --tail=200
kubectl -n "$ns" logs deploy/kodex-worker --tail=200
kubectl -n "$ns" get ingress
kubectl -n "$ns" describe ingress kodex
kubectl -n "$ns" get certificate,order,challenge -A

# TLS reuse store (best-effort, может быть пусто в самый первый деплой)
kubectl -n kodex-system get secrets | grep '^kodex-tls-' || true

# Full-env run namespaces (S2 Day3 baseline)
kubectl get ns -l kodex.works/managed-by=kodex-worker,kodex.works/namespace-purpose=run
for run_ns in $(kubectl get ns -l kodex.works/managed-by=kodex-worker,kodex.works/namespace-purpose=run -o jsonpath='{.items[*].metadata.name}'); do
  echo "=== ${run_ns} ==="
  kubectl -n "${run_ns}" get sa,role,rolebinding,resourcequota,limitrange,job,pod
done

# Day4: проверить env wiring и логи agent-runner job
for run_ns in $(kubectl get ns -l kodex.works/managed-by=kodex-worker,kodex.works/namespace-purpose=run -o jsonpath='{.items[*].metadata.name}'); do
  echo "=== ${run_ns} agent jobs ==="
  kubectl -n "${run_ns}" get jobs,pods
  kubectl -n "${run_ns}" get pod -l app.kubernetes.io/name=kodex-run \
    -o jsonpath='{range .items[*].spec.containers[*].env[*]}{.name}{"\n"}{end}' \
    | grep -E 'KODEX_OPENAI_API_KEY|KODEX_GIT_BOT_TOKEN|KODEX_GIT_BOT_USERNAME|KODEX_GIT_BOT_MAIL|KODEX_AGENT_DISPLAY_NAME' || true
done

# Legacy runtime keys must not appear after Day3 rollout
kubectl get ns -o json | grep -E 'kodex.io/(managed-by|namespace-purpose|runtime-mode|project-id|run-id|correlation-id)' || true
```

## Namespace cleanup (автоматический)

- В production развёрнут `CronJob` `kodex-worker-namespace-cleanup`.
- Расписание по умолчанию: каждые `15` минут (`*/15 * * * *`, `Etc/UTC`).
- Cleanup работает только для managed runtime namespace с guardrails:
  - labels `kodex.works/managed-by=kodex-worker` и `kodex.works/namespace-purpose=run`;
  - allowlist platform runtime namespace names: `codex-issue*` и slot namespaces `kodex-dev-*`;
  - в БД нет non-terminal run для `kodex.works/run-id`;
  - в namespace нет active `pod/job/cronjob/deployment/statefulset/daemonset/replicaset`.
- Аудит причин пишется в worker logs и `flow_events` (`run.namespace.cleaned`, `run.namespace.cleanup_skipped`, `run.namespace.cleanup_failed`).
- In-band cleanup в worker tick остаётся как best-effort backstop; для полного отключения нужно выключить и tick, и CronJob.

Проверка статуса:

```bash
ns="kodex-prod"

kubectl -n "$ns" get cronjob kodex-worker-namespace-cleanup
kubectl -n "$ns" get jobs -l app.kubernetes.io/component=worker-namespace-cleanup
kubectl -n "$ns" logs job/<cleanup_job_name> --tail=200
kubectl -n "$ns" logs deploy/kodex-worker --tail=200 | grep 'namespace cleanup' || true
kubectl get ns -l kodex.works/managed-by=kodex-worker,kodex.works/namespace-purpose=run
```

Как отключить:

```bash
# 1. Остановить периодический CronJob:
export KODEX_WORKER_NAMESPACE_CLEANUP_CRON_SUSPEND="true"

# 2. Если нужно полностью выключить и in-band cleanup внутри worker:
export KODEX_WORKER_RUN_NAMESPACE_CLEANUP="false"

# 3. Применить обновлённые env и дождаться rollout платформы.
```

## Registry GC (автоматический)

- В production/non-ai окружениях включён `CronJob` `kodex-registry-gc`.
- Расписание по умолчанию: ежедневно в `03:17 UTC`.
- Job делает `scale deployment/kodex-registry 1 -> 0`, выполняет `registry garbage-collect --delete-untagged`, затем возвращает `replicas=1`.
- Для init-контейнера GC helper по умолчанию используется mirrored shell-capable image
  `127.0.0.1:5000/kodex/mirror/alpine-k8s:1.32.2`
  (можно переопределить через `KODEX_KUBECTL_IMAGE`).

Проверка статуса:

```bash
ns="kodex-prod"
kubectl -n "$ns" get cronjob kodex-registry-gc
kubectl -n "$ns" get jobs -l app.kubernetes.io/name=kodex-registry-gc
kubectl -n "$ns" logs job/<gc_job_name> --tail=200
```

Форсированный запуск вне расписания:

```bash
ns="kodex-prod"
kubectl -n "$ns" create job --from=cronjob/kodex-registry-gc kodex-registry-gc-manual-$(date +%s)
kubectl -n "$ns" get jobs -l app.kubernetes.io/name=kodex-registry-gc
```

## Host containerd image GC (автоматический)

- Registry GC удаляет только untagged blobs из internal registry PVC.
- Node-level `containerd` cache/snapshots это отдельный слой хранения; его чистит kubelet image GC и host prune timer.
- Bootstrap настраивает:
  - kubelet thresholds через `/var/lib/rancher/k3s/agent/etc/kubelet.conf.d/10-kodex-image-gc.conf`;
  - timer `kodex-image-prune.timer`, который запускает `k3s crictl --timeout 120s rmi --prune`.

Проверка:

```bash
sudo cat /var/lib/rancher/k3s/agent/etc/kubelet.conf.d/10-kodex-image-gc.conf
sudo systemctl status kodex-image-prune.timer --no-pager
sudo journalctl -u kodex-image-prune.service -n 200 --no-pager
sudo /usr/local/bin/k3s crictl images | head -n 50
```

Форсированный запуск:

```bash
sudo systemctl start kodex-image-prune.service
sudo /usr/local/bin/k3s crictl --timeout 120s rmi --prune
```

## Cleanup heavy JSON payloads (автоматический)

- Control-plane выполняет hourly cleanup heavy JSON-полей для старых записей (по умолчанию `7` дней):
  - `agent_runs.agent_logs_json`;
  - `agent_sessions.session_json`, `agent_sessions.codex_cli_session_json`;
  - `runtime_deploy_tasks.logs_json`.
- Retention настраивается через:
  - `KODEX_RUN_HEAVY_FIELDS_RETENTION_DAYS` (основной ключ);
  - `KODEX_RUN_AGENT_LOGS_RETENTION_DAYS` (legacy fallback).

## Типовые проблемы

### Web UI не открывается / "ui upstream unavailable"
- Проверить, что `kodex-web-console` pod Running и port `5173` открыт в cluster.
- Проверить NetworkPolicy baseline (должен быть allow до web-console).

### OAuth2 callback не проходит
- В GitHub OAuth App callback должен быть:
  - `https://<KODEX_PRODUCTION_DOMAIN>/oauth2/callback`

### Webhook не доходит
- Убедиться, что path пропущен без auth:
  - `oauth2-proxy --skip-auth-regex=^/api/v1/webhooks/.*`
- Проверить `KODEX_GITHUB_WEBHOOK_SECRET` совпадает с секретом вебхука в GitHub.

### TLS не выпускается (HTTP-01) / cert-manager молчит
- Убедиться, что `KODEX_PRODUCTION_DOMAIN` резолвится в production host IP.
- Если это первый выпуск TLS, runtime-deploy использует echo-probe (HTTP) до включения issuer:
  - проверить `kubectl -n "$KODEX_PRODUCTION_NAMESPACE" get deploy,svc,ingress | grep echo-probe`;
  - проверить логи `kubectl -n "$KODEX_PRODUCTION_NAMESPACE" logs deploy/kodex-control-plane --tail=200`.

### Build падает с `MANIFEST_UNKNOWN` при `retrieving image from cache`
- Симптом: Kaniko падает на base image с логом вида `Error while retrieving image from cache ... MANIFEST_UNKNOWN`.
- Причина: в registry мог остаться stale mirror/cache state после cleanup/GC (тег виден, но digest манифест недоступен).
- Текущее безопасное значение по умолчанию: `KODEX_KANIKO_CACHE_ENABLED=false`.
- Если cache включали вручную и снова получили `MANIFEST_UNKNOWN`:
  - переключить `KODEX_KANIKO_CACHE_ENABLED=false` в `kodex-runtime`;
  - убедиться, что `kodex-control-plane` подтянул значение после rollout;
  - повторить deploy.
- Дополнительно:
  - mirror шаг выполняет platform-aware health-check (`--platform linux/amd64`) и ремонтирует stale mirror;
  - mirror выполняется в single-arch режиме (`KODEX_IMAGE_MIRROR_PLATFORM=linux/amd64`), чтобы не оставлять multi-arch index с отсутствующими дочерними манифестами;
  - при cache-related `MANIFEST_UNKNOWN` build автоматически ретраится без cache.
