---
doc_id: EPC-CK8S-S12-D4-RATE-LIMIT
type: epic
title: "Epic S12 Day 4: Architecture для GitHub API rate-limit resilience и controlled wait ownership (Issues #418/#420)"
status: in-review
owner_role: SA
created_at: 2026-03-13
updated_at: 2026-03-13
related_issues: [366, 413, 416, 418, 420]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-13-issue-418-arch"
---

# Epic S12 Day 4: Architecture для GitHub API rate-limit resilience и controlled wait ownership (Issues #418/#420)

## TL;DR
- Подготовлен architecture package Sprint S12 для GitHub API rate-limit resilience: architecture decomposition, C4 overlays, ADR-0013 и alternatives по wait-state ownership и contour boundaries.
- Зафиксирован ownership split для provider signal normalization, agent-path handoff, controlled wait aggregate, visibility surfaces, finite auto-resume orchestration и manual-intervention path.
- Подготовлен handover в `run:design` без premature transport/schema lock-in.

## Priority
- `P0`.

## Контекст
- Intake baseline: `#366` (`docs/delivery/epics/s12/epic-s12-day1-github-api-rate-limit-intake.md`).
- Vision baseline: `#413` (`docs/delivery/epics/s12/epic-s12-day2-github-api-rate-limit-vision.md`).
- PRD baseline: `#416` (`docs/delivery/epics/s12/epic-s12-day3-github-api-rate-limit-prd.md`, `docs/delivery/epics/s12/prd-s12-day3-github-api-rate-limit-resilience.md`).
- Текущий этап: `run:arch` в Issue `#418`.
- Scope этапа: только markdown-изменения.

## Architecture package
- `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/README.md`
- `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/architecture.md`
- `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/c4_context.md`
- `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/c4_container.md`
- `docs/architecture/adr/ADR-0013-github-rate-limit-controlled-wait-ownership.md`
- `docs/architecture/alternatives/ALT-0005-github-rate-limit-wait-state-boundaries.md`
- `docs/delivery/traceability/s12_github_api_rate_limit_resilience_history.md`

## Ключевые решения Stage
- `control-plane` остаётся единственным owner для recoverable/hard-failure classification, controlled wait aggregate, contour attribution, recovery hints и visibility contract.
- `worker` закреплён за time-based wait scheduling, finite auto-resume attempts и escalation в manual-action-required; `agent-runner` не может продолжать local retry loop после handoff.
- `api-gateway` и `web-console` остаются thin visibility surfaces и не вычисляют countdown или GitHub classification из сырых сигналов.
- Day4 вводит разделение `contour_kind` и `signal_origin`, чтобы сохранить two-contour PRD contract без дублирования доменной логики.

## Context7 и внешний baseline
- Проверен актуальный non-interactive GitHub CLI flow через Context7:
  - `/websites/cli_github_manual`.
- Проверен актуальный Mermaid C4 syntax через Context7:
  - `/mermaid-js/mermaid`.
- Официальные GitHub Docs `Rate limits for the REST API` и `Best practices for using the REST API` просмотрены 2026-03-13 и использованы как внешний baseline для primary/secondary semantics.

## Acceptance Criteria (Issue #418)
- [x] Подготовлен architecture package с service boundaries, ownership, C4 overlays, ADR и alternatives для GitHub API rate-limit resilience.
- [x] Для core flows определены owner-сервисы и границы ответственности: signal detection, classification, wait orchestration, visibility surfaces, resume/manual path и agent backpressure handoff.
- [x] Зафиксированы architecture-level trade-offs по ownership и contour model без premature transport/storage lock-in.
- [x] Обновлены delivery/traceability документы и package indexes.
- [x] Подготовлена follow-up issue `#420` для stage `run:design` без trigger-лейбла.

## Quality gates
| Gate | Что проверяем | Статус |
|---|---|---|
| `QG-S12-D4-01` Architecture completeness | Есть package `architecture + C4 + ADR + alternatives` | passed |
| `QG-S12-D4-02` Boundary integrity | Ownership за `control-plane` / `worker` / `agent-runner` / visibility surfaces зафиксирован явно | passed |
| `QG-S12-D4-03` Contour fidelity | Split `platform PAT` vs `agent bot-token` сохранён без UI/service drift | passed |
| `QG-S12-D4-04` Agent backpressure discipline | `agent-runner` не остаётся владельцем wait-state и не retry'ит локально после handoff | passed |
| `QG-S12-D4-05` Stage continuity | Подготовлена issue `#420` на `run:design` без trigger-лейбла | passed |

## Handover в `run:design`
- Следующий этап: `run:design`.
- Follow-up issue: `#420`.
- Trigger-лейбл на новую issue ставит Owner после review architecture package.
- На design-stage обязательно:
  - выпустить `design_doc`, `api_contract`, `data_model`, `migrations_policy`;
  - определить typed signal/handoff/visibility DTO и wait-state persisted model;
  - зафиксировать rollout/rollback notes и finite auto-resume policy.
