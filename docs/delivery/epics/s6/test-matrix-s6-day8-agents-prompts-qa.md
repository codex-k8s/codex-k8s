---
doc_id: TST-MTX-S6-D8
type: test-matrix
title: "S6 Day8 — Test Matrix для QA lifecycle agents/prompt templates (Issue #201)"
status: in-review
owner_role: QA
created_at: 2026-02-27
updated_at: 2026-02-27
related_issues: [199, 201, 216]
related_prs: [202]
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-27-issue-201-test-matrix"
---

# Test Matrix: S6 Day8 acceptance/regression

## TL;DR
- Критичные сценарии: contracts (`agents/templates/audit`), migration invariants, typed UI flow, runtime readiness.
- Минимальный набор для перехода в release: backend tests + frontend build + runtime rollout + DB schema checks.

## Матрица покрытия
| Фича/Сценарий | Unit | Integration | E2E | Manual | Perf | Security | Notes |
|---|---:|---:|---:|---:|---:|---:|---|
| Staff OpenAPI endpoints для `agents/templates/audit` | 0 | 1 | 0 | 1 | 0 | 0 | `api.yaml` содержит `/staff/agents*`, `/staff/prompt-templates*`, `/staff/audit/prompt-templates` |
| gRPC contracts `ListAgents..ListPromptTemplateAuditEvents` | 0 | 1 | 0 | 1 | 0 | 0 | `controlplane.proto` содержит полный RPC набор lifecycle |
| Backend regression (`control-plane`) | 1 | 1 | 0 | 0 | 0 | 0 | `go test ./services/internal/control-plane/...` PASS |
| Backend regression (`api-gateway`) | 1 | 1 | 0 | 0 | 0 | 0 | `go test ./services/external/api-gateway/...` PASS |
| Frontend typed integration compile/build | 0 | 1 | 0 | 0 | 0 | 0 | `npm --prefix services/staff/web-console run build` PASS |
| Migration schema `prompt_templates` + `agents.settings_version` | 0 | 1 | 1 | 1 | 0 | 0 | `psql \\d+ prompt_templates`, `psql \\d+ agents` PASS |
| Invariants: unique active version/check constraints | 0 | 1 | 1 | 1 | 0 | 0 | индексы/constraints присутствуют в runtime DB |
| Seed bootstrap `ru/en` + `work/revise` | 0 | 1 | 1 | 1 | 0 | 0 | `SELECT ... FROM prompt_templates` показывает active seed rows |
| Runtime readiness (`control-plane`, `worker`, `web-console`) | 0 | 1 | 1 | 0 | 0 | 1 | `kubectl rollout status` PASS, deployments Ready |
| Go lint / duplicate check | 1 | 0 | 0 | 0 | 0 | 0 | `make lint-go` PASS, `make dupl-go` FAIL (known risk) |
| Browser manual flow (`Agents` UI clickthrough) | 0 | 0 | 0 | 0 | 0 | 0 | Не выполнялось в этом run (explicitly out of scope) |

## Дефекты/риски по зонам
- `RSK-201-01` (P1): `make dupl-go` сообщает дубли в репозитории (включая S6 Day7 области); не блокирует функциональную готовность, но требует контроля в release cycle.
- `RSK-201-02` (P2): отсутствует manual browser e2e для UX.
- `RSK-201-03` (P2): нет perf baseline для `preview/diff`.

## Апрув
- request_id: owner-2026-02-27-issue-201-test-matrix
- Решение: pending
- Комментарий: матрица готова для Owner review.
