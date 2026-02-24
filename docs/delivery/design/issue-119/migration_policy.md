---
doc_id: MGP-CK8S-0119
type: design-migration-policy
title: "Issue #119 — Migration and Runtime Impact Policy"
status: draft
owner_role: SA
created_at: 2026-02-24
updated_at: 2026-02-24
related_issues: [119, 118]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-24-issue-119-design"
---

# Migration Policy: Issue #119

## TL;DR
- Это design-only шаг: миграции схемы, контрактов и runtime manifests не выполняются.
- Влияние на runtime ограничено документированием и проверкой существующего orchestration path.
- Rollback сводится к откату markdown-документов.

## Политика изменений
- Разрешено:
  - обновление markdown-артефактов design/traceability;
  - фиксация архитектурных инвариантов и критериев приемки.
- Запрещено:
  - изменения SQL migrations, кода сервисов, manifest/CI pipeline;
  - любые non-markdown правки.

## Миграции данных/схемы
- PostgreSQL schema migration: not required.
- OpenAPI/proto migration: not required.
- Runtime deploy migration: not required.

## Runtime impact assessment
- Основной риск: процессный (неполный evidence или неполная трассируемость), не технический.
- Mitigation:
  - обязательный evidence schema из `data_model.md`;
  - обновление `issue_map` и master plan в том же PR.
  - явная матрица связей в `traceability_matrix.md`.

## Traceability guardrails
- Каждый design-артефакт должен иметь ссылку на source requirements и evidence target.
- Срез issue #119 в `docs/delivery/issue_map.md` должен включать весь пакет `docs/delivery/design/issue-119/*.md`.
- Для сценариев B1/B2/B3 должен быть проверяемый `expected vs actual` в evidence bundle для Issue #118.

## Rollback plan
1. Отменить PR с design-изменениями.
2. Вернуться к предыдущим версиям markdown-артефактов.
3. Повторно запустить `run:design` после уточнения замечаний Owner.

## Acceptance checklist
- [ ] Подтверждено отсутствие DB/API/runtime миграций.
- [ ] Зафиксирован process-only impact и меры контроля.
- [ ] Rollback-путь описан и воспроизводим.

## Апрув
- request_id: owner-2026-02-24-issue-119-design
- Решение: pending
- Комментарий:
