---
doc_id: EPC-CK8S-S12-D5-GITHUB-RATE-LIMIT
type: epic
title: "Epic S12 Day 5: Design для GitHub API rate-limit resilience и controlled wait contracts (Issue #420)"
status: completed
owner_role: SA
created_at: 2026-03-13
updated_at: 2026-03-13
related_issues: [366, 413, 416, 418, 420, 423]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-03-13-issue-420-design-epic"
  approved_by: "ai-da-stas"
  approved_at: 2026-03-13
---

# Epic S12 Day 5: Design для GitHub API rate-limit resilience и controlled wait contracts (Issue #420)

## TL;DR
- Подготовлен Day5 design package Sprint S12: `design_doc`, `api_contract`, `data_model`, `migrations_policy`.
- Зафиксированы отдельный coarse wait-state `waiting_backpressure`, typed signal handoff `agent-runner -> control-plane`, persisted wait/evidence model и finite auto-resume policy для primary/secondary GitHub limits.
- Manual path остаётся typed visibility contract без dedicated operator write endpoint в первой волне MVP; handover materialized в plan package `#423` и execution waves `#425..#431`.

## Priority
- `P0`.

## Контекст
- Intake baseline: `#366` (`docs/delivery/epics/s12/epic-s12-day1-github-api-rate-limit-intake.md`).
- Vision baseline: `#413` (`docs/delivery/epics/s12/epic-s12-day2-github-api-rate-limit-vision.md`).
- PRD baseline: `#416` (`docs/delivery/epics/s12/epic-s12-day3-github-api-rate-limit-prd.md`, `docs/delivery/epics/s12/prd-s12-day3-github-api-rate-limit-resilience.md`).
- Architecture baseline: `#418` (`docs/delivery/epics/s12/epic-s12-day4-github-api-rate-limit-arch.md` + architecture package).
- Текущий этап: `run:design` в Issue `#420`.
- Scope этапа: только markdown-изменения.

## Design package
- `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/README.md`
- `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/design_doc.md`
- `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/api_contract.md`
- `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/data_model.md`
- `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/migrations_policy.md`
- `docs/delivery/traceability/s12_github_api_rate_limit_resilience_history.md`

## Ключевые design-решения
- Wait taxonomy:
  - выбран новый coarse runtime state `waiting_backpressure` вместо reuse `waiting_mcp`;
  - business reason фиксирован как `github_rate_limit`.
- Persistence:
  - schema owner остаётся `control-plane`;
  - вводятся `github_rate_limit_waits` и `github_rate_limit_wait_evidence`;
  - `agent_runs` держит dominant wait pointer, а related waits остаются видимыми на staff surfaces.
- Transport contract:
  - существующие staff routes `/api/v1/staff/runs/{run_id}`, `/api/v1/staff/runs/waits` и realtime stream расширяются additively;
  - для `agent-runner` вводится internal callback RPC `ReportGitHubRateLimitSignal`.
- Resume policy:
  - primary limit uses deterministic `reset_at + guard`;
  - secondary limit uses `Retry-After` or bounded exponential backoff;
  - every path has finite auto-resume budget and typed manual-action escalation.
- Visibility:
  - staff/private surfaces are canonical;
  - GitHub service-comment stays best-effort mirror and gets explicit retry path if blocked by platform contour.

## Context7 и внешний baseline
- Через Context7 `/github/docs` подтверждены:
  - primary vs secondary rate-limit semantics;
  - приоритет response headers над `GET /rate_limit`;
  - guidance `wait at least one minute` + exponential backoff when `Retry-After` отсутствует;
  - avoidance of concurrency bursts.
- Новые внешние библиотеки Day5 не выбирались.

## Acceptance Criteria (Issue #420)
- [x] Подготовлен design package (`design_doc`, `api_contract`, `data_model`, `migrations_policy`).
- [x] Явно описаны typed wait-state и recovery-hint contracts, persisted ownership model, rollout/rollback notes и observability evidence.
- [x] Сохранены архитектурные инварианты: split contours, hard-failure separation, no infinite local retries, GitHub-first baseline.
- [x] Подготовлена follow-up issue `#423` для stage `run:plan` без trigger-лейбла.

## Quality gates
| Gate | Что проверяем | Статус |
|---|---|---|
| `QG-S12-D5-01` Contract completeness | Есть полный API/data/migrations/design package | passed |
| `QG-S12-D5-02` Wait taxonomy integrity | GitHub wait не смешан с approval/interaction semantics | passed |
| `QG-S12-D5-03` Finite auto-resume | Для primary/secondary path описаны bounded retries и manual escalation | passed |
| `QG-S12-D5-04` Visibility fallback | Staff surfaces каноничны, service-comment описан как best-effort mirror | passed |
| `QG-S12-D5-05` Stage continuity | Создана issue `#423` на `run:plan` без trigger-лейбла | passed |

## Handover в `run:plan`
- Следующий этап: `run:plan`.
- Follow-up issue: `#423`.
- Trigger-лейбл на новую issue ставит Owner после review design package.
- На plan-stage обязательно:
  - декомпозировать execution waves по schema, domain, worker, runner, edge/UI и observability;
  - зафиксировать DoR/DoD, acceptance evidence и rollout gates;
  - подготовить continuity issue для `run:dev`.
