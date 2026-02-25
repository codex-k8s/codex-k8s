---
doc_id: ALT-0001
type: alternatives
title: "Agents prompt templates lifecycle — Alternatives & Trade-offs"
status: accepted
owner_role: SA
created_at: 2026-02-25
updated_at: 2026-02-25
related_issues: [184, 185, 187, 189, 195]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-25-issue-189-arch-alt"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-25
---

# Alternatives & Trade-offs: Agents prompt templates lifecycle

## TL;DR
- Рассмотрели: A) версии в `prompt_templates` + audit через `flow_events`, B) отдельные таблицы версий/аудита, C) Git-based шаблоны.
- Рекомендуем: Вариант A.
- Почему: минимальный риск, совместимость с текущей моделью данных, быстрее handover в `run:design`.

## Контекст
- Требуются версии, diff, preview, audit history и rollback.
- Нельзя ломать текущие границы `external -> internal -> jobs`.
- Audit должен быть append-only и связан с `correlation_id`.

## Вариант A: Версионирование внутри `prompt_templates` + `flow_events`
- Описание: каждая правка создаёт новую версию строки; активная версия отмечается флагом.
- Плюсы: минимальные изменения, простая миграция.
- Минусы: рост таблицы, нужен контроль индексации.
- Риски: конфликтные правки без optimistic concurrency.
- Стоимость/сложность: средняя.

## Вариант B: Отдельные таблицы `prompt_template_versions` + `prompt_template_audit`
- Описание: выделяем версии и audit в отдельные сущности.
- Плюсы: чистая история и простые audit-запросы.
- Минусы: сложнее миграции и join-потоки, выше стоимость.
- Риски: согласованность между версиями и audit при сбоях.
- Стоимость/сложность: высокая.

## Вариант C: Git-based workflow (PR на шаблоны)
- Описание: шаблоны живут в репозитории, изменения через PR, БД как кэш.
- Плюсы: сильный аудит и diff tooling.
- Минусы: сильно усложняет UX и процессы, выходит за scope.
- Риски: блокировки релизов при Git-фейлах.
- Стоимость/сложность: высокая.

## Сравнение
| Критерий | A | B | C |
|---|---:|---:|---:|
| Срок | короткий | средний | длинный |
| Риск | низкий | средний | высокий |
| Стоимость | средняя | высокая | высокая |
| Производительность | высокая | средняя | средняя |
| Эксплуатация | средняя | сложная | сложная |

## Рекомендация
- Выбор: Вариант A.
- Обоснование: минимальный риск при сохранении audit-качества.
- Что теряем: идеальную нормализацию audit-данных.
- Что выигрываем: скорость и совместимость с текущей моделью.

## Проверка зависимостей (Context7)
- `kin-openapi` (`/getkin/kin-openapi`): покрывает runtime request/response validation для contract-first API без дополнительных библиотек.
- `monaco-editor` (`/microsoft/monaco-editor`): встроенный `DiffEditor` закрывает сценарий сравнения версий шаблонов.
- Вывод: для `run:design` не требуется вводить новые внешние зависимости, достаточно существующего стека.

## Решение Owner
- request_id: `owner-2026-02-25-issue-189-arch-alt`
- Решение: принят Вариант A (версии в `prompt_templates` + audit через `flow_events`)
- Комментарий: отдельный журнал изменений не требуется, diff вычисляется на лету, soft-lock не вводится.
