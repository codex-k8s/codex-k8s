---
doc_id: EPC-CK8S-S3-D3
type: epic
title: "Epic S3 Day 3: MCP deterministic secret sync (GitHub + Kubernetes)"
status: planned
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
