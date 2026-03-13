---
doc_id: EPC-CK8S-S12-D3-RATE-LIMIT
type: epic
title: "Epic S12 Day 3: PRD для GitHub API rate-limit resilience и controlled wait contract (Issues #416/#418)"
status: completed
owner_role: PM
created_at: 2026-03-13
updated_at: 2026-03-13
related_issues: [366, 413, 416, 418]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-03-13-issue-416-prd"
  approved_by: "ai-da-stas"
  approved_at: 2026-03-13
---

# Epic S12 Day 3: PRD для GitHub API rate-limit resilience и controlled wait contract (Issues #416/#418)

## TL;DR
- Подготовлен PRD-пакет Sprint S12 для GitHub API rate-limit resilience: `epic-s12-day3-github-api-rate-limit-prd.md` и `prd-s12-day3-github-api-rate-limit-resilience.md`.
- Зафиксированы user stories, FR/AC/NFR, edge cases, expected evidence и wave priorities для controlled wait-state, contour attribution, transparency surfaces, resume semantics и запрета infinite local retries.
- Принято продуктовое решение: GitHub остаётся текущим provider baseline, `platform PAT` и `agent bot-token` остаются разными operational contour, а predictive budgeting, multi-provider quota governance и adapter-specific notifications не блокируют core MVP.
- Создана follow-up issue `#418` для stage `run:arch` без trigger-лейбла.

## Priority
- `P0`.

## Контекст
- Intake baseline: `#366` (`docs/delivery/epics/s12/epic-s12-day1-github-api-rate-limit-intake.md`).
- Vision baseline: `#413` (`docs/delivery/epics/s12/epic-s12-day2-github-api-rate-limit-vision.md`).
- Текущий этап: `run:prd` в Issue `#416`.
- Следующий этап: `run:arch` в Issue `#418`.

## Scope
### In scope
- Формализация user stories, FR/AC/NFR и edge cases для GitHub API rate-limit resilience.
- Приоритизация волн `core wait-state -> transparency/resume evidence -> deferred expansion`.
- Фиксация product guardrails для split `platform PAT` vs `agent bot-token`, primary vs secondary rate-limit semantics, transparency surfaces и manual-intervention path.
- Явный handover в `run:arch` с перечнем продуктовых решений, которые нельзя потерять.
- Синхронизация traceability (`issue_map`, `delivery_plan`, sprint/epic docs, history package).

### Out of scope
- Кодовая реализация, storage/schema decisions и transport/runtime lock-in.
- Общий quota-management framework для всех провайдеров и всех внешних ограничений.
- Predictive budget analytics, proactive throttling policies и cross-provider quota dashboards.
- Adapter-specific notification semantics, channel policies и расширенный alerting redesign.

## PRD package
- `docs/delivery/epics/s12/epic-s12-day3-github-api-rate-limit-prd.md`
- `docs/delivery/epics/s12/prd-s12-day3-github-api-rate-limit-resilience.md`
- `docs/delivery/traceability/s12_github_api_rate_limit_resilience_history.md`

## Wave priorities

| Wave | Приоритет | Scope | Exit signal |
|---|---|---|---|
| Wave 1 | `P0` | Recoverable controlled wait-state для `platform PAT` и `agent bot-token`, contour attribution, hard-failure separation | Recoverable primary/secondary rate-limit больше не трактуются как ложный `failed`, а owner/operator видит затронутый контур и recovery hint |
| Wave 2 | `P0` | Visibility surfaces, safe resume, manual-intervention path, audit/correlation evidence и backpressure discipline на agent path | Run может детерминированно выйти из wait-state в auto-resume или manual-action-required без infinite local retries |
| Wave 3 | `P1` (deferred) | Predictive budgeting, multi-provider quota governance, adapter-specific notifications/reminders | Stream попадает в roadmap только после подтверждения core architecture и owner-решения вне Sprint S12 MVP |

## Acceptance criteria (Issue #416)
- [x] Подготовлен PRD-артефакт GitHub API rate-limit resilience и синхронизирован в traceability-документах.
- [x] Для core flows зафиксированы user stories, FR/AC/NFR, edge cases и expected evidence.
- [x] Wave priorities сформулированы без смешения core MVP и deferred quota/notification streams.
- [x] Сохранены неподвижные ограничения инициативы: GitHub-first baseline, split `platform PAT` vs `agent bot-token`, controlled wait вместо ложного `failed`, typed recovery hints и запрет infinite local retries.
- [x] Создана follow-up issue `#418` для stage `run:arch` без trigger-лейбла.

## Quality gates
| Gate | Что проверяем | Статус |
|---|---|---|
| QG-S12-D3-01 PRD completeness | User stories, FR/AC/NFR, edge cases и expected evidence покрывают scope Day3 | passed |
| QG-S12-D3-02 Guardrails preserved | Split contours, hard-failure separation, typed recovery hints и no local retry-loop сохранены | passed |
| QG-S12-D3-03 Deferred scope discipline | Predictive budgeting, multi-provider governance и adapter-specific notifications не смешаны с core MVP | passed |
| QG-S12-D3-04 Stage continuity | Создана issue `#418` для `run:arch` без trigger-лейбла | passed |
| QG-S12-D3-05 Policy compliance | Изменены только markdown-артефакты | passed |

## Handover в `run:arch`
- Следующий этап: `run:arch`.
- Follow-up issue: `#418`.
- Trigger-лейбл `run:arch` на issue `#418` ставит Owner после review PRD-пакета.
- Обязательные выходы архитектурного этапа:
  - service boundaries и ownership для rate-limit signal ingestion, wait-state orchestration, visibility surfaces, resume gating и manual-intervention path;
  - alternatives/ADR по ownership `control-plane` / `worker` / `agent-runner` / edge surfaces без потери product contract;
  - фиксация, как сохраняются split `platform PAT` vs `agent bot-token`, typed recovery hints и запрет infinite local retries;
  - отдельная issue на `run:design` без trigger-лейбла после завершения `run:arch`.

## Открытые риски и допущения
| Type | ID | Описание | Статус |
|---|---|---|---|
| risk | `RSK-416-01` | Инициатива может расползтись в общий retry/backoff redesign вместо узкой product capability вокруг GitHub rate-limit resilience | open |
| risk | `RSK-416-02` | Продукт может пообещать точный countdown там, где GitHub secondary limits не дают надёжного сигнала | open |
| risk | `RSK-416-03` | Ownership detection/wait/resume останется размытым между сервисами и вернёт ложные `failed` или local retry-loops | open |
| assumption | `ASM-416-01` | Existing runtime/audit baseline платформы достаточно зрелый, чтобы добавить controlled wait как отдельную capability без redesign orchestration с нуля | accepted |
| assumption | `ASM-416-02` | GitHub provider signals (`Retry-After`, `X-RateLimit-*`, response class) достаточны, чтобы отделять recoverable wait от hard failure на product-уровне | accepted |
