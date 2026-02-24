---
doc_id: ADR-0008
type: adr
title: "Run service-comment idempotency for progress and next-step updates"
status: accepted
owner_role: SA
created_at: 2026-02-24
updated_at: 2026-02-24
related_issues: [145]
related_prs: []
supersedes: []
superseded_by: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-24-issue-145-run-service-comment-idempotency"
---

# ADR-0008: Run service-comment idempotency for progress and next-step updates

## TL;DR
- Проблема: при повторных webhook/retry и параллельных update-path платформа может создавать дубли service-comment для одного `run_id`.
- Решение: закрепить singleton-модель service-comment на `(repository, issue_number, run_id)` и обновлять его через `PATCH`, а `POST` использовать только для первого создания.
- Влияние: UX в Issue становится детерминированным, а traceability по `run_status_report` и stage next-step не дробится на несколько комментариев.

## Контекст
- Для `run:*` платформа публикует и обновляет service-comment с фазами выполнения и stage-aware подсказками.
- `run_status_report` обязан публиковаться регулярно и влияет на содержимое service-comment.
- В E2E циклах `run:design` и `run:design:revise` возможны:
  - повторные события одной фазы;
  - retry после сетевых сбоев;
  - конкурентные попытки обновления service-comment из разных шагов orchestration.

Без явной idempotency-стратегии это приводит к созданию нескольких комментариев с одним и тем же операционным смыслом.

## Decision Drivers
- Единый видимый источник статуса для одного `run_id`.
- Retry-safe поведение при сбоях GitHub API и повторных webhook событиях.
- Минимальный API surface без усложнения MCP контрактов.
- Полная трассируемость через существующий audit-контур.

## Решение

### 1. Singleton-ключ service-comment
- Канонический ключ: `(repository_full_name, issue_number, run_id)`.
- В body комментария обязательно хранится marker:
  - `<!-- codex-k8s:run-status {...} -->`
  - marker содержит как минимум `run_id`, `issue_number`, `repository_full_name`, `phase`.

### 2. Алгоритм upsert
1. Получить список issue comments и найти комментарий с marker текущего `run_id`.
2. Если найден один комментарий:
  - обновить его через `PATCH /repos/{owner}/{repo}/issues/comments/{comment_id}`.
3. Если не найден:
  - создать комментарий через `POST /repos/{owner}/{repo}/issues/{issue_number}/comments`.
4. Если найдено больше одного (legacy drift):
  - выбрать канонический (последний обновлённый),
  - обновлять только его,
  - остальные считать техническим долгом на cleanup-цикл.

### 3. Идемпотентность содержимого
- Перед `PATCH` рассчитывается `payload_hash` нормализованного markdown-тела.
- Если `payload_hash` совпадает с текущим телом комментария, обновление пропускается как `idempotent_reuse`.
- Повторный приход одинакового `run_status_report` не должен приводить к новому комментарию и не должен менять `updated_at` без изменений контента.

### 4. Аудит
- Для каждой попытки публикации service-comment фиксируются события:
  - `run.service_message.updated` (контент изменён);
  - `run.service_message.reused` (изменений нет, idempotent replay);
  - `run.service_message.duplicate_detected` (обнаружено >1 комментария на `run_id`).

## Последствия

### Позитивные
- Один run = один service-comment в GitHub.
- Повторные вызовы `run_status_report` становятся retry-safe.
- Stage-aware подсказки не «размножаются» в Issue.

### Компромиссы
- Нужен дополнительный read шаг перед update (list/find comment).
- Требуется policy на очистку старых дублей (out-of-band task).

## Влияние на runtime и миграцию
- Изменений схемы БД не требуется.
- Обратная совместимость сохранена:
  - существующие комментарии с marker продолжают использоваться;
  - для legacy issues возможны временные дубли до cleanup.
- Rollout:
  1. включить idempotent upsert в run-status/service-message path;
  2. провести e2e проверку `run:design` и `run:design:revise` с retry сценариями;
  3. отдельно запланировать cleanup legacy дублей.

## Связанные документы
- `docs/architecture/api_contract.md`
- `docs/architecture/mcp_approval_and_audit_flow.md`
- `docs/product/labels_and_trigger_policy.md`
- `docs/architecture/adr/ADR-0006-review-driven-revise-and-next-step-ux.md`
