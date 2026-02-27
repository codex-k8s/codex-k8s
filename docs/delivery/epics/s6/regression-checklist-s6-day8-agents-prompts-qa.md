---
doc_id: REG-S6-D8
type: regression-checklist
title: "S6 Day8 — Regression Checklist для lifecycle agents/prompt templates (Issue #201)"
status: in-review
owner_role: QA
created_at: 2026-02-27
updated_at: 2026-02-27
related_issues: [199, 201, 216]
related_prs: [202]
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-27-issue-201-regression"
---

# Regression Checklist: S6 Day8

## TL;DR
Минимальный регресс для решения о переходе из `run:qa` в `run:release` выполнен; зафиксирован один известный риск качества (`dupl-go`).

## Чек-лист
### Smoke
- [x] Сервисы в namespace `codex-k8s-dev-1` в статусе `Running/Ready`
- [x] `kubectl rollout status` для `codex-k8s-control-plane`, `codex-k8s-worker`, `codex-k8s-web-console` успешно
- [x] Логи `control-plane/worker` без критичных ошибок старта

### Основные сценарии
- [x] `go test ./services/internal/control-plane/...`
- [x] `go test ./services/external/api-gateway/...`
- [x] `npm --prefix services/staff/web-console run build`
- [x] Проверена контрактная полнота `api.yaml` и `controlplane.proto` для `agents/templates/audit`
- [x] Проверены миграционные инварианты в runtime DB (`prompt_templates`, `agents.settings_version`)
- [x] Проверен seeded baseline (`ru/en`, `work/revise`) в `prompt_templates`

### Негативные сценарии
- [x] `make dupl-go` выполнен и зафиксирован non-green результат как риск `RSK-201-01`
- [x] Ограничение RBAC учтено: cluster-scope `kubectl get pods -A` запрещён (expected behavior)

### Наблюдаемость
- [x] Workloads видимы через `kubectl -n codex-k8s-dev-1 get pods,deploy,job -o wide`
- [x] Логи `deploy/codex-k8s-control-plane` и `deploy/codex-k8s-worker` доступны

## What was tested
- Контракты staff HTTP/gRPC (`agents/templates/audit`).
- Автотесты backend и build frontend.
- Runtime rollout + smoke checks в текущем namespace.
- Фактическая схема и инварианты `prompt_templates`/`agents` в PostgreSQL.

## What was not tested
- Browser-based manual UI сценарии (`Agents` clickthrough, diff/preview UX руками).
- Нагрузочные/perf и security penetration сценарии.
- Post-release/postdeploy сценарии (вне scope текущего этапа).

## Результаты прогонов
- Дата: `2026-02-27`
- Окружение: `full-env`, namespace `codex-k8s-dev-1`, run `534af179-7c1f-459b-8c7c-258f9d6c6835`
- Итог: `PASS with known risk`
- Ссылки на логи/отчеты:
  - `go test ./services/internal/control-plane/...` — PASS
  - `go test ./services/external/api-gateway/...` — PASS
  - `npm --prefix services/staff/web-console run build` — PASS
  - `make lint-go` — PASS
  - `make dupl-go` — FAIL (`RSK-201-01`, duplicates)
  - `kubectl -n codex-k8s-dev-1 get pods,deploy,job -o wide` — PASS
  - `kubectl -n codex-k8s-dev-1 rollout status deployment/<name>` — PASS
  - `psql \\d+ prompt_templates`, `psql \\d+ agents` — PASS

## Апрув
- request_id: owner-2026-02-27-issue-201-regression
- Решение: pending
- Комментарий: чек-лист закрыт, требуется Owner review решения readiness.
