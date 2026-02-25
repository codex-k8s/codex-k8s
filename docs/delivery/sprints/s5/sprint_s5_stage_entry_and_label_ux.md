---
doc_id: SPR-CK8S-0005
type: sprint-plan
title: "Sprint S5: Stage entry and label UX orchestration (Issues #154/#155)"
status: in-progress
owner_role: EM
created_at: 2026-02-24
updated_at: 2026-02-25
related_issues: [154, 155]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-24-issue-155-sprint-plan"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-25
---

# Sprint S5: Stage entry and label UX orchestration (Issues #154/#155)

## TL;DR
- Цель спринта: убрать ручную сложность управления `run:*` лейблами и дать Owner детерминированный UX переходов по этапам.
- Основной результат: profile-driven stage launch (`quick-fix`, `feature`, `new-service`) + рабочие next-step action paths с fallback без зависимости от web-link.
- Day1 split: Issue #154 закрыл intake baseline, Issue #155 в `run:plan` сформировал handover-пакет в `run:dev` с quality-gates, критериями завершения и реестром блокеров.

## Scope спринта
### In scope
- Формализация и реализация launch profiles с явной матрицей этапов.
- Обновление service-message UX:
  - profile-aware next-step cards;
  - fallback-команды для запуска без web-console.
- Валидация и guardrails для переходов между профилями (эскалация в full pipeline).
- Полная трассируемость `issue -> requirements -> stage policy -> acceptance`.

### Out of scope
- Изменение базовой taxonomy `run:* / state:* / need:*`.
- Изменение RBAC-модели доступа к production runtime.
- Изменение multi-repo architecture из Sprint S4.

## План эпиков по дням

| День | Эпик | Priority | Документ | Статус |
|---|---|---|---|---|
| Day 1 | Launch profiles и deterministic next-step actions | P0 | `docs/delivery/epics/s5/epic-s5-day1-launch-profiles-and-stage-launcher-ux.md` + `docs/delivery/epics/s5/prd-s5-day1-launch-profiles-and-stage-launcher-ux.md` + `docs/architecture/adr/ADR-0008-profile-driven-stage-launch-and-next-step-contract.md` | in-review (`run:plan`, Issue #155) |

## Daily gate (обязательно)
- Любой переход stage должен иметь audit запись и детерминированный fallback.
- Изменения policy/docs синхронно отражаются в `requirements_traceability` и `issue_map`.
- Для каждого profile зафиксированы критерии входа/выхода и условия эскалации.

## Quality-gates для handover в `run:dev` (Issue #155)

| Gate | Что проверяем | Статус |
|---|---|---|
| QG-01 Planning | Канонические launch profiles + escalation rules синхронизированы между PRD, stage policy и epic Day1 | passed |
| QG-02 Contract | Next-step contract фиксирует `primary_action + fallback_action` и ambiguity guardrails | passed |
| QG-03 Governance | Fallback path использует только labels из каталога и не содержит секретов | passed |
| QG-04 Traceability | Обновлены `issue_map` и `requirements_traceability`; связь `Issue -> FR-053/FR-054 -> epic` сохранена | passed |
| QG-05 Review readiness | Owner decision package (блокеры, риски, решения) подготовлен и подтверждён в review | passed (Owner approved, 2026-02-25) |

## Completion критерии спринта
- [x] Launch profiles закреплены как продуктовый стандарт и связаны с stage policy.
- [x] Next-step action UX не зависит от единственного web-link канала.
- [x] Для каждого profile определены acceptance scenarios и failure modes.
- [x] Подготовлен handover в `run:dev` с приоритетами реализации и рисками.
- [x] Owner review/approval vision-prd пакета по Issue #155.

## Критерии приемки `run:plan` по Issue #155
- [x] Подтвержден канонический набор launch profiles и правила эскалации.
- [x] Подтвержден формат next-step action-карт и fallback-команд.
- [x] Подготовлен owner-facing пакет quality-gates и критериев завершения перед `run:dev`.
- [x] Получено финальное Owner approval на запуск `run:dev`.

## Блокеры, риски и owner decisions (run:plan)

| Тип | ID | Описание | Что требуется | Статус |
|---|---|---|---|---|
| blocker | BLK-155-01 | Owner approval пакета Day1 для запуска `run:dev` | Approve получен в PR #166 (review comments от 2026-02-25) | closed |
| blocker | BLK-155-02 | Требовалась фиксация политики fast-track `design -> dev` как fallback-пути | Решение зафиксировано как OD-155-01 с guardrails и audit-требованиями | closed |
| risk | RSK-155-01 | Перегрузка service-comment снижает читаемость и может привести к ошибкам перехода | Подтвердить лимит action-card (1 primary + 1 fallback) | monitoring |
| risk | RSK-155-02 | Ручной fallback без pre-check может вызвать некорректный transition | Сохранить hard requirement `pre-check -> transition` | monitoring |
| owner-decision | OD-155-01 | Политика fast-track для `design -> dev` в S5 | Утверждено: optional fast-track сохраняется только вместе с canonical `design -> plan`; обязательны `pre-check -> transition` и audit trail | approved |
| owner-decision | OD-155-02 | Hard-stop при ambiguity (`0` или `>1` stage-labels) | Утверждено: обязательная постановка `need:input`, best-guess переходы запрещены | approved |
| owner-decision | OD-155-03 | Единый review-gate на Issue + PR | Утверждено: обязательная синхронизация `state:in-review` на обе сущности | approved |

## Handover после завершения Sprint S5 Day1
- `dev`: реализовать profile resolver, service-message fallback и policy проверки.
- `qa`: подготовить сценарии UX-валидации (happy-path + broken-link + wrong-profile).
- `sre`: проверить аудит/наблюдаемость и отсутствие обходов governance.
- `km`: поддерживать трассируемость и синхронизацию продуктовой документации.

## Приоритеты входа в `run:dev` (декомпозиция Day1)
- `P0`: deterministic profile resolver (`quick-fix|feature|new-service`) и блокировка ambiguity через `need:input`.
- `P0`: контракт next-step action-card + fallback path `pre-check -> transition` на основе `gh issue/pr edit`.
- `P1`: унифицированный review-gate переход (`state:in-review` на Issue и PR после формирования PR).
- `P1`: автоматическая синхронизация traceability-артефактов (`issue_map`, `requirements_traceability`) после stage transition.

## Факт по Issue #155
- Подтверждён канонический набор launch profiles и deterministic escalation rules.
- Зафиксирован контракт next-step action-card (`primary deep-link + fallback command`) с guardrails на ambiguity и обязательным pre-check перед ручным transition.
- Fallback-синтаксис `gh issue edit` / `gh pr edit` для `--add-label`/`--remove-label` дополнительно сверен по Context7 (`/websites/cli_github_manual`).
- Подготовлен отдельный PRD-документ по шаблону `docs/templates/prd.md`: `docs/delivery/epics/s5/prd-s5-day1-launch-profiles-and-stage-launcher-ux.md`.
- Подготовлен ADR-артефакт для реализации в `run:dev`: `docs/architecture/adr/ADR-0008-profile-driven-stage-launch-and-next-step-contract.md`.
- Пакет требований и governance-checklist готов к запуску `run:dev`; Owner approval зафиксирован.
