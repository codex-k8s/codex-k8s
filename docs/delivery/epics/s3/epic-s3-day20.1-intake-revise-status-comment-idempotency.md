---
doc_id: EPC-CK8S-S3-D20.1
type: epic
title: "Epic S3 Day 20.1: Intake/revise final e2e check for status-comment idempotency"
status: in-progress
owner_role: PM
created_at: 2026-02-24
updated_at: 2026-02-24
related_issues: [137, 19]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-24-issue-137-intake-e2e"
---

# Epic S3 Day 20.1: Intake/revise final e2e check for status-comment idempotency

## TL;DR
- Цель: формально закрыть финальную e2e-проверку для `run:intake` и `run:intake:revise` после hardening идемпотентности статус-комментариев.
- Результат: подтверждён детерминированный lifecycle service-comment без дублей и с корректными next-step подсказками для Owner.
- Трассируемость: изменения закрепляются в `issue_map`, `requirements_traceability` и policy-документах по лейблам/этапам.

## Priority
- `P0`.

## Scope
### In scope
- Проверка `run:intake` на одном Issue:
  - единый service-comment создаётся/обновляется по фазам `планируется запуск -> подготовка окружения -> запуск -> завершение`;
  - повторные обработчики webhook не создают дубликаты комментариев.
- Проверка `run:intake:revise` на том же Issue:
  - lifecycle service-comment переиспользуется и обновляется детерминированно;
  - после завершения корректно публикуются stage-aware next-step действия (`run:intake:revise`, `run:vision`).
- Проверка label transition policy:
  - trigger `run:intake`/`run:intake:revise` снимается с Issue;
  - на Issue устанавливается `state:in-review` для Owner decision.
- Синхронизация delivery traceability:
  - `issue #137 -> требования (FR-051/FR-052) -> критерии готовности Day20.1`.

### Out of scope
- Изменения dev-контуров `run:dev`/`run:dev:revise`.
- Изменение stage taxonomy, RBAC policy и MCP approval matrix.

## Acceptance criteria
- Для одного запуска создаётся не более одного активного service-comment статуса; повторные события только обновляют существующий комментарий.
- После `run:intake` и `run:intake:revise` в Issue остаётся финальный статус с корректным stage-aware next-step без конфликтующих рекомендаций.
- Для повторного revise-цикла на том же Issue не возникает дублирования статуса и не теряется ссылка на текущий run-context.
- `issue_map` и `requirements_traceability` содержат прямые ссылки на артефакт Day20.1 и Issue #137.
- На этапе handover нет открытых `need:input`, если все AC выполнены и установлен `state:in-review`.

## Product risks
- Риск ложноположительной "идемпотентности" при редких гонках webhook-событий.
- Риск UX-деградации, если stage-aware next-step карточка станет неоднозначной для Owner.

## Product assumptions
- Для `run:intake` и `run:intake:revise` используется единая policy-модель service-comment из `docs/product/labels_and_trigger_policy.md`.
- Проверка выполняется в `code-only` режиме и опирается на аудит GitHub/flow событий без runtime-зависимостей full-env.
