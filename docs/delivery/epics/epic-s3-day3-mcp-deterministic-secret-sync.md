---
doc_id: EPC-CK8S-S3-D3
type: epic
title: "Epic S3 Day 3: MCP deterministic secret sync (GitHub + Kubernetes)"
status: completed
owner_role: EM
created_at: 2026-02-13
updated_at: 2026-02-13
related_issues: [19]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
---

# Epic S3 Day 3: MCP deterministic secret sync (GitHub + Kubernetes)

## TL;DR
- Цель: реализовать безопасный инструмент синхронного создания секрета в GitHub и Kubernetes без раскрытия значения модели.
- MVP-результат: deterministic secret lifecycle по окружению и проекту.

## Priority
- `P0`.

## Scope
### In scope
- MCP tool `secret.sync.github_k8s` с входом: project/repository/environment/secret_name/policy.
- Генерация секрета внутри trusted tool runtime, masking в логах и callback payload.
- Idempotency-key и retry-safe поведение.
- Approval policy и детальный audit trail.
- Вендор-нейтральный approver слой:
  - аппрувером может быть любой HTTP-адаптер, поддерживающий утвержденный контракт;
  - Telegram-адаптер (`telegram-approver`) и `yaml-mcp-server` используются как референсные реализации контракта.

### Out of scope
- Массовая миграция существующих секретов и интеграция с Vault/KMS.

## Критерии приемки
- Повторный вызов не приводит к дрейфу состояния.
- Секретный материал недоступен в model output и user-facing логах.

## Фактический результат (выполнено)
- Канонизирован MCP tool name для секрета:
  - `secret.sync.github_k8s` (основной);
  - `mcp_secret_sync_env` оставлен как legacy alias на transport-уровне для мягкого перехода.
- Расширен вход `SecretSyncEnvInput`:
  - `project_id`, `repository`, `environment`, `github_secret_name`, `kubernetes_secret_name`, `kubernetes_namespace`, `kubernetes_secret_key`;
  - `policy` (`deterministic|random|provided`);
  - `idempotency_key` (опциональный; при отсутствии вычисляется детерминированно).
- Реализована детерминированная генерация секрета (default policy):
  - значение выводится через HMAC-SHA256 от стабильного material (`project/repository/env/secret refs`) и platform seed;
  - при `policy=random` генерируется криптографически случайное значение;
  - при `policy=provided` обязателен `secret_value`.
- Реализован retry-safe idempotency для secret sync:
  - action signature расширена `project/repository/policy/idempotency_key`;
  - добавлен repo-метод `FindLatestBySignature`;
  - если найден уже `requested/approved` — возвращается existing approval request;
  - если найден `applied` — side effects не повторяются, возвращается `reused=true`, `message=idempotent_replay`.
- Секретный материал остаётся hidden:
  - секрет не возвращается в tool output;
  - в аудит/flow events пишутся только безопасные метаданные;
  - в `mcp_action_requests.payload` хранится только encrypted blob.
- Усилен payload apply-path:
  - проверка `project_id`/`repository` из payload против run context перед фактическим apply.

## Data model impact
- Схема БД не менялась.
- Логическая модель `mcp_action_requests` расширена на уровне `target_ref/payload` (JSONB):
  - `project_id`, `repository`, `policy`, `idempotency_key`.

## Проверки
- `go test ./services/internal/control-plane/...` — passed.
- `make lint-go` — passed.
- `make dupl-go` — passed.
