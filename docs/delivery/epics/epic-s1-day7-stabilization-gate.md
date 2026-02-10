---
doc_id: EPC-CK8S-S1-D7
type: epic
title: "Epic Day 7: Stabilization, regression and release gate"
status: planned
owner_role: EM
created_at: 2026-02-06
updated_at: 2026-02-06
related_issues: [1]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-06-day7"
---

# Epic Day 7: Stabilization, regression and release gate

## TL;DR
- Цель эпика: закрыть спринт с проверяемым состоянием staging и подтверждённым quality gate.
- Ключевая ценность: предсказуемый переход к следующему спринту без накопленных регрессий.
- MVP-результат: regression checklist, runbook обновлён, backlog следующего спринта сформирован.

## Priority
- `P0` (релизный gate перед следующим спринтом).

## Ожидаемые артефакты дня
- Regression report и итоговый статус ключевых сценариев на staging.
- Список дефектов и решение `go/no-go` для Sprint S2.
- Обновлённый backlog/план следующего спринта в delivery документах.
- Актуализированные runbook/rollback notes при найденных проблемах.

## Контекст
- Почему эпик нужен: без stabilizing day рост дефектов сорвёт темп ежедневных деплоев.
- Связь с требованиями: NFR-001, NFR-002, NFR-007.

## Scope
### In scope
- Прогон e2e regression сценариев на staging.
- Проверка webhook -> run -> worker -> k8s -> UI цепочки.
- Фиксация известных дефектов и решений (go/no-go для Sprint S2).
- Обновление runbook и delivery артефактов.

### Out of scope
- Новые крупные фичи вне стабилизации.
- Production rollout.

## Декомпозиция (Stories/Tasks)
- Story-1: regression matrix и ручной прогон.
- Story-2: bug triage, приоритизация и фиксы критических дефектов.
- Story-3: финальный staging deployment report.
- Story-4: подготовка плана Sprint S2.

## Data model impact (по шаблону data_model.md)
- Сущности:
  - Проверка консистентности `agent_runs`, `slots`, `flow_events`, `learning_feedback`.
- Связи/FK:
  - Валидация отсутствия нарушений FK и orphan rows.
- Индексы и запросы:
  - Профилирование основных запросов и фиксация hot spots.
- Миграции:
  - Стабилизационные миграции только по критичным дефектам.
- Retention/PII:
  - Проверка, что логи/feedback не содержат секретов.

## Критерии приемки эпика
- Все критические сценарии regression пройдены.
- Нет блокирующих дефектов P0/P1 для продолжения daily deploy.
- Подготовлен и согласован backlog Sprint S2.
- Финальные изменения дня задеплоены на staging и проверены.

## Риски/зависимости
- Зависимости: завершённость Day 1..6.
- Риск: выявление критичных дефектов в конце спринта.

## План релиза (верхний уровень)
- Sprint S1 закрывается отчётом о стабильности staging и списком next actions.

## Апрув
- request_id: owner-2026-02-06-day7
- Решение: approved
- Комментарий: Day 7 scope принят.
