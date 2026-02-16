---
doc_id: EPC-CK8S-S1-D5
type: epic
title: "Epic Day 5: Learning mode MVP"
status: completed
owner_role: EM
created_at: 2026-02-06
updated_at: 2026-02-11
related_issues: [1]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-06-day5"
---

# Epic Day 5: Learning mode MVP

## TL;DR
- Цель эпика: добавить образовательный слой для user-initiated задач.
- Ключевая ценность: пользователи не только получают результат, но и понимают инженерные решения.
- MVP-результат: explain-instructions augmentation + хранение feedback в БД.

## Priority
- `P1` (важно для product value, но после core execution/auth).

## Ожидаемые артефакты дня
- Реализация effective learning mode resolution и prompt augmentation в pipeline.
- Схема и persistence для `learning_feedback` + индексы.
- Минимальный staff UI/API просмотр образовательных объяснений.
- Проверка on/off сценариев learning mode на production.

## Контекст
- Почему эпик нужен: learning mode является частью продуктовой идеи платформы.
- Связь с требованиями: FR-023.

## Scope
### In scope
- Learning mode toggle применение к user/project context.
- Prompt augmentation блок (`почему/зачем/компромиссы/альтернативы`).
- Сохранение образовательных объяснений в `learning_feedback`.
- Отображение базового списка learning feedback в staff UI/API.

### Out of scope
- Продвинутая оценка качества объяснений.
- Полная автоматизация line-level PR comments (можно частично, если успевает в день).

## Декомпозиция (Stories/Tasks)
- Story-1: effective learning mode resolution (project + member override).
- Story-2: prompt/context augmentation в pipeline запуска.
- Story-3: persistence `learning_feedback`.
- Story-4: минимальный просмотр feedback в UI/API.

## Data model impact (по шаблону data_model.md)
- Сущности:
  - `agent_runs.learning_mode` как effective flag run-level.
  - `learning_feedback` для объяснений (inline/post_pr).
- Связи/FK:
  - `learning_feedback.run_id -> agent_runs.id`.
  - `learning_feedback.repository_id -> repositories.id` (optional).
- Индексы и запросы:
  - Индекс `learning_feedback(run_id, created_at)`.
  - Индекс `agent_runs(project_id, learning_mode, started_at)`.
- Миграции:
  - Создать таблицу `learning_feedback`, если ещё отсутствует.
- Retention/PII:
  - Исключить запись секретов и чувствительных токенов в explanation text.

## Критерии приемки эпика
- Для включенного режима в результатах задач есть объяснение why/tradeoffs.
- Для выключенного режима augmentation отсутствует.
- Изменения задеплоены и проверены на production в день реализации.

## Риски/зависимости
- Зависимости: устойчивый контур запуска и storage.
- Риск: избыточный шум в ответах без quality filter.

## План релиза (верхний уровень)
- Deploy в production + ручной сценарий с включением/выключением learning mode.

## Апрув
- request_id: owner-2026-02-06-day5
- Решение: approved
- Комментарий: Day 5 scope принят.
