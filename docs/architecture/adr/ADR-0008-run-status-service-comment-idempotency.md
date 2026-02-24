---
doc_id: ADR-0008
type: adr
title: "Run status service-comment idempotency and anchor strategy"
status: accepted
owner_role: SA
created_at: 2026-02-24
updated_at: 2026-02-24
related_issues: [143]
related_prs: []
supersedes: []
superseded_by: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-24-issue-143-run-status-idempotency"
---

# ADR-0008: Run status service-comment idempotency and anchor strategy

## TL;DR
- Проблема: при ретраях `run_status_report` платформа могла создать дубликаты run service-comment в GitHub и нарушить стабильность ссылки на прогресс.
- Решение: закрепить для каждого run один anchor-комментарий (`comment_id`) и выполнять только idempotent upsert этого же комментария.
- Результат: повторные вызовы становятся retry-safe без дополнительных side effects, а аудит остаётся воспроизводимым.

## Контекст
- `run_status_report` используется агентом как частый progress-feedback сигнал.
- В условиях сетевых сбоев и конкурентных попыток обновления нужен детерминированный путь без дублирования комментариев.
- Для review gate важно сохранять стабильный URL run service-comment, так как он используется в next-step навигации Owner.

## Decision Drivers
- Idempotent поведение при ретраях и partial-failure сценариях.
- Один канонический service-comment на run.
- Полная трассируемость через `flow_events`.
- Минимальные изменения runtime-контракта без миграции внешних API.

## Рассмотренные варианты

### Вариант A: append-only комментарии на каждый статус
- Плюсы: простая реализация.
- Минусы: шум в PR/Issue, нет стабильной точки ссылки, тяжело анализировать прогресс.

### Вариант B: upsert по поиску текста в комментариях
- Плюсы: меньше дубликатов по сравнению с append-only.
- Минусы: недетерминированно при изменении шаблона текста; возможны ложные совпадения.

### Вариант C (выбран): anchor `comment_id` + idempotent upsert
- Плюсы: строгая детерминированность, retry-safe поведение, стабильная ссылка.
- Минусы: требуется хранение и валидация `comment_id` в runtime-состоянии.

## Решение
Выбран **Вариант C**.

### Инварианты
1. На один `run_id` существует максимум один активный run service-comment.
2. После первичного создания `comment_id` становится anchor для всех последующих обновлений.
3. Повторный `run_status_report` не создаёт новый комментарий, если anchor доступен.
4. При конфликте/утрате anchor выполняется controlled recovery с повторной фиксацией нового `comment_id` в аудите.

### Audit-события
- `run.agent.status_reported` — принят очередной статус.
- `run.service_message.updated` — выполнен upsert anchor-комментария.
- `mcp.tool.failed` — зафиксирована ошибка публикации (если есть).

## Влияние на сервисные границы
- `services/jobs/agent-runner`: только отправка короткого статуса через MCP `run_status_report`.
- `services/internal/control-plane`: владелец orchestration и idempotent upsert логики service-comment.
- `services/external/api-gateway`: без изменений доменной логики; только транспорт webhook/staff.

## Миграция и runtime-impact
- Внешние API/DTO не меняются.
- Схема БД не меняется.
- Runtime-impact:
  - снижается риск дублирования комментариев в PR/Issue;
  - стабилизируется ссылка на status-comment для review/revise UX;
  - retry-поведение становится предсказуемым при временных сбоях GitHub API.

## Риски и компенсирующие меры
- Риск: устаревший `comment_id` (удалён вручную).
  - Мера: controlled recovery с перепривязкой anchor и аудитом события.
- Риск: race-condition при конкурентных апдейтах.
  - Мера: сериализация update path на стороне control-plane и idempotent re-read перед записью.

## Связанные документы
- `docs/architecture/api_contract.md`
- `docs/architecture/mcp_approval_and_audit_flow.md`
- `docs/delivery/e2e_mvp_master_plan.md`
- `docs/delivery/issue_map.md`
