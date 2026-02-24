---
doc_id: SPR-CK8S-0005
type: sprint-plan
title: "Sprint S5: Stage entry and label UX orchestration (Issue #154)"
status: planned
owner_role: EM
created_at: 2026-02-24
updated_at: 2026-02-24
related_issues: [154]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-24-issue-154-sprint-plan"
---

# Sprint S5: Stage entry and label UX orchestration (Issue #154)

## TL;DR
- Цель спринта: убрать ручную сложность управления `run:*` лейблами и дать Owner детерминированный UX переходов по этапам.
- Основной результат: profile-driven stage launch (`quick-fix`, `feature`, `new-service`) + рабочие next-step action paths с fallback без зависимости от web-link.
- Sprint S5 стартует после review intake-артефактов по Issue #154.

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
| Day 1 | Launch profiles и deterministic next-step actions | P0 | `docs/delivery/epics/s5/epic-s5-day1-launch-profiles-and-stage-launcher-ux.md` | planned |

## Daily gate (обязательно)
- Любой переход stage должен иметь audit запись и детерминированный fallback.
- Изменения policy/docs синхронно отражаются в `requirements_traceability` и `issue_map`.
- Для каждого profile зафиксированы критерии входа/выхода и условия эскалации.

## Completion критерии спринта
- [ ] Launch profiles закреплены как продуктовый стандарт и связаны с stage policy.
- [ ] Next-step action UX не зависит от единственного web-link канала.
- [ ] Для каждого profile определены acceptance scenarios и failure modes.
- [ ] Подготовлен handover в `run:dev` с приоритетами реализации и рисками.

## Handover после завершения Sprint S5 Day1
- `dev`: реализовать profile resolver, service-message fallback и policy проверки.
- `qa`: подготовить сценарии UX-валидации (happy-path + broken-link + wrong-profile).
- `sre`: проверить аудит/наблюдаемость и отсутствие обходов governance.
- `km`: поддерживать трассируемость и синхронизацию продуктовой документации.
