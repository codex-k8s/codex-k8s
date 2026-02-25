---
doc_id: ADR-0009
type: adr
title: "Prompt templates lifecycle and audit model"
status: proposed
owner_role: SA
created_at: 2026-02-25
updated_at: 2026-02-25
related_issues: [184, 185, 187, 189]
related_prs: []
supersedes: []
superseded_by: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-25-issue-189-arch-adr-0009"
---

# ADR-0009: Prompt templates lifecycle and audit model

## TL;DR
- Проблема: нужен управляемый lifecycle шаблонов промптов с версионированием, diff/preview и audit history.
- Решение: использовать версионирование внутри `prompt_templates` + `flow_events` как audit trail, с optimistic concurrency и явной активацией версии.
- Последствия: проще реализация и совместимость с текущей моделью данных, но требуется контроль роста таблицы и точные индексы.

## Контекст
PRD по Issue #187 требует:
- lifecycle шаблонов `work/revise` с локалями и fallback;
- diff/effective preview;
- историю изменений и аудит.

Ограничения:
- доменная логика и БД остаются в `services/internal/control-plane`;
- audit должен быть append-only и связан с `correlation_id`;
- event-outbox на MVP не вводится.

## Decision drivers
- Auditability и непротиворечивость истории.
- Совместимость с текущей data model (`prompt_templates` уже есть).
- Простота миграции и низкий риск для `run:design`.
- Поддержка rollback и recovery.

## Рассмотренные варианты

### Вариант A: Версионирование внутри `prompt_templates` + audit через `flow_events`
- Плюсы: минимальные изменения схемы, единая таблица, быстрый доступ к активной версии.
- Минусы: рост таблицы, необходимость строгих индексов.
- Риски: конкурирующие правки, частичная запись без audit.
- Стоимость внедрения: средняя.
- Эксплуатация: контроль размера таблицы и индексации.

### Вариант B: Отдельные таблицы `prompt_template_versions` и `prompt_template_audit`
- Плюсы: чистая модель истории, легче фильтровать audit.
- Минусы: сложнее миграции и запросы, больше данных и join-операций.
- Риски: несогласованность между версиями и audit при сбоях.
- Стоимость внедрения: высокая.
- Эксплуатация: сложнее поддержка и индексы.

### Вариант C: Хранение шаблонов в Git (PR-based), БД как кэш
- Плюсы: полный Git audit и diff tooling.
- Минусы: усложнение UX, зависимость от репозиториев, нужен новый workflow.
- Риски: выход за scope, задержка бизнес-ценности.
- Стоимость внедрения: высокая.
- Эксплуатация: зависимость от Git операций и permissions.

## Решение
Выбираем **Вариант A** с минимальными расширениями схемы.

Рекомендованные расширения для design-этапа:
- поля `status` (`draft|active|archived`), `change_reason`, `checksum`;
- optimistic concurrency через `expected_version`;
- явная операция `activate` для смены active версии.

## Обоснование (Rationale)
Вариант A сохраняет совместимость с текущей data model и удовлетворяет audit/NFR без введения новых сложных таблиц или Git-based workflow. Это минимизирует риск и время до `run:design`, сохраняя возможность эволюции в Variant B при росте масштаба.

## Последствия (Consequences)

### Позитивные
- Быстрая реализация без слома текущей модели данных.
- Прозрачная история версий и возможность rollback.

### Негативные / компромиссы
- Рост `prompt_templates` и необходимость строгих индексов.
- Diff вычисляется на лету или требует кэширования.

### Технический долг
- Возможный переход к отдельной audit-таблице при росте объёма истории.
- Введение архивирования версий старше retention-порога.

## План внедрения (минимально)
- Добавить поля и индексы (если утверждены) в `prompt_templates`.
- Обновить доменную логику: create version, activate version, rollback.
- Добавить audit события в `flow_events`.
- Обновить staff API и UI для diff/preview/history.

## План отката/замены
- Вернуться к single-version режиму, сохранив историю как read-only.
- Откат не удаляет исторические версии.

## Связанные документы
- `docs/architecture/agents_prompt_templates_lifecycle_design.md`
- `docs/architecture/prompt_templates_policy.md`
- `docs/architecture/data_model.md`
- `docs/delivery/epics/s6/epic-s6-day3-agents-prompts-prd.md` (PR #190, ожидает merge)
- `docs/delivery/epics/s6/prd-s6-day3-agents-prompts-lifecycle.md` (PR #190, ожидает merge)
