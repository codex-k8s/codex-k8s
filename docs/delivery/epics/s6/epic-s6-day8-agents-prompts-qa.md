---
doc_id: EPC-CK8S-S6-D8
type: epic
title: "Epic S6 Day 8: QA для lifecycle управления агентами и шаблонами промптов (Issue #201)"
status: in-review
owner_role: QA
created_at: 2026-02-27
updated_at: 2026-02-27
related_issues: [184, 185, 187, 189, 195, 197, 199, 201, 216]
related_prs: [202]
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-27-issue-201-qa-epic"
---

# Epic S6 Day 8: QA для lifecycle agents/prompt templates

## TL;DR
- Выполнена QA-приёмка реализации Day7 (`PR #202`) по AC issue `#201`.
- Contracts/migrations/runtime regression подтверждены; сформирован полный QA-пакет (`strategy/plan/matrix/checklist`).
- Readiness decision: `GO` в `run:release` с зафиксированным риском `RSK-201-01` (`dupl-go` non-green).

## Контекст
- Stage continuity: `#184 -> #185 -> #187 -> #189 -> #195 -> #197 -> #199 -> #201`.
- Вход:
  - GitHub issue `#199` (dev handover)
  - GitHub PR `#202` (merged в `main`)
  - design/data/migration документы Day5
- Роль этапа: `qa` (markdown-only scope).

## План выполнения и критерии приемки
### План выполнения
1. Проверить contracts `staff API + gRPC` для `agents/templates/audit`.
2. Проверить migration/schema invariants `prompt_templates` и `agents.settings_version`.
3. Проверить отсутствие regressions по backend/frontend сборке и lint.
4. Выполнить runtime smoke в `full-env` namespace.
5. Зафиксировать тестовые артефакты и readiness decision.

### Критерии приемки (Issue #201)
- [x] Корректность staff API/gRPC контрактов `agents/templates/audit` подтверждена.
- [x] Миграции и инварианты `prompt_templates` подтверждены.
- [x] UI flow `Agents` подтвержден на уровне typed integration (без mock в API layer).
- [x] Regression evidence собран, решение readiness зафиксировано.

## Выполненные проверки
### Автотесты и сборка
- `go test ./services/internal/control-plane/...` — PASS
- `go test ./services/external/api-gateway/...` — PASS
- `npm --prefix services/staff/web-console run build` — PASS
- `make lint-go` — PASS
- `make dupl-go` — FAIL (см. `RSK-201-01`)

### Runtime checks (`codex-k8s-dev-1`)
- `kubectl -n codex-k8s-dev-1 get pods,deploy,job -o wide` — PASS
- `kubectl -n codex-k8s-dev-1 rollout status deployment/codex-k8s-control-plane --timeout=120s` — PASS
- `kubectl -n codex-k8s-dev-1 rollout status deployment/codex-k8s-worker --timeout=120s` — PASS
- `kubectl -n codex-k8s-dev-1 rollout status deployment/codex-k8s-web-console --timeout=120s` — PASS
- `kubectl -n codex-k8s-dev-1 logs deploy/codex-k8s-control-plane --tail=80` — PASS (critical errors not found)
- `kubectl -n codex-k8s-dev-1 logs deploy/codex-k8s-worker --tail=80` — PASS (critical errors not found)

### Contract/schema validation
- OpenAPI endpoints присутствуют:
  - `/api/v1/staff/agents*`
  - `/api/v1/staff/prompt-templates*`
  - `/api/v1/staff/audit/prompt-templates`
- gRPC RPC присутствуют:
  - `ListAgents`, `GetAgent`, `UpdateAgentSettings`
  - `ListPromptTemplateKeys/Versions`, `Create/ActivatePromptTemplateVersion`
  - `SyncPromptTemplateSeeds`, `PreviewPromptTemplate`, `DiffPromptTemplateVersions`, `ListPromptTemplateAuditEvents`
- PostgreSQL runtime schema:
  - `prompt_templates` table и индексы `uq_prompt_templates_scope_version`, `uq_prompt_templates_active_version`, `idx_prompt_templates_scope_version_desc` присутствуют
  - `agents.settings_version` присутствует
  - seeded active templates (`global`, `ru/en`, `work/revise`) присутствуют

## QA-артефакты Day8
- `docs/delivery/epics/s6/test-strategy-s6-day8-agents-prompts-qa.md`
- `docs/delivery/epics/s6/test-plan-s6-day8-agents-prompts-qa.md`
- `docs/delivery/epics/s6/test-matrix-s6-day8-agents-prompts-qa.md`
- `docs/delivery/epics/s6/regression-checklist-s6-day8-agents-prompts-qa.md`

## What was tested / What was not tested
### What was tested
- Contracts (OpenAPI/proto), migration invariants, runtime schema.
- Backend tests, frontend build, lint checks.
- Namespace runtime readiness and logs.

### What was not tested
- Manual browser clickthrough UI acceptance.
- Performance/load testing `preview/diff`.
- Security penetration beyond baseline policy/RBAC checks.

## Риски и замечания
| Тип | ID | Описание | Статус |
|---|---|---|---|
| risk | RSK-201-01 | `make dupl-go` фиксирует дубли (включая зоны S6 Day7). | accepted-for-release |
| risk | RSK-201-02 | Не выполнен manual browser acceptance для UI flow. | carry-to-release |
| risk | RSK-201-03 | Нет perf baseline `preview/diff` на больших шаблонах. | carry-to-release |

## Readiness decision
- Решение QA: `GO` в `run:release` при условии явной фиксации рисков `RSK-201-01..03` в release-stage артефактах.
- Follow-up issue `run:release`: `#216` (создана без trigger label, лейбл ставит Owner).
- Continuity-инструкция для следующего этапа: по завершении `run:release` обязательно создать issue для `run:postdeploy`.

## Context7 verification notes
- Проверены актуальные команды:
  - GitHub CLI (`/websites/cli_github_manual`) для issue/pr json/create flows.
  - Kubernetes docs (`/websites/kubernetes_io`) для rollout/status/log patterns.
- Новые внешние зависимости для Day8 не требуются.
