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
  status: pending
  request_id: "owner-2026-02-24-issue-155-sprint-plan"
---

# Sprint S5: Stage entry and label UX orchestration (Issues #154/#155)

## TL;DR
- Цель спринта: убрать ручную сложность управления `run:*` лейблами и дать Owner детерминированный UX переходов по этапам.
- Основной результат: profile-driven stage launch (`quick-fix`, `feature`, `new-service`) + рабочие next-step action paths с fallback без зависимости от web-link.
- Day1 split: Issue #154 закрыл intake baseline, Issue #155 фиксирует vision/prd пакет перед входом в `run:dev`.

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
| Day 1 | Launch profiles и deterministic next-step actions | P0 | `docs/delivery/epics/s5/epic-s5-day1-launch-profiles-and-stage-launcher-ux.md` + `docs/delivery/epics/s5/prd-s5-day1-launch-profiles-and-stage-launcher-ux.md` | in-review (`run:vision`/`run:prd`, Issue #155) |

## Daily gate (обязательно)
- Любой переход stage должен иметь audit запись и детерминированный fallback.
- Изменения policy/docs синхронно отражаются в `requirements_traceability` и `issue_map`.
- Для каждого profile зафиксированы критерии входа/выхода и условия эскалации.

## Completion критерии спринта
- [x] Launch profiles закреплены как продуктовый стандарт и связаны с stage policy.
- [x] Next-step action UX не зависит от единственного web-link канала.
- [x] Для каждого profile определены acceptance scenarios и failure modes.
- [x] Подготовлен handover в `run:dev` с приоритетами реализации и рисками.
- [ ] Owner review/approval vision-prd пакета по Issue #155.

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
- Подготовлен отдельный PRD-документ по шаблону `docs/templates/prd.md`: `docs/delivery/epics/s5/prd-s5-day1-launch-profiles-and-stage-launcher-ux.md`.
- Пакет требований готов к запуску `run:dev` после Owner approval.
