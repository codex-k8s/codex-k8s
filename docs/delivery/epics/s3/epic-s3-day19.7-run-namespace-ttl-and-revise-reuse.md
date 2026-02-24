---
doc_id: EPC-CK8S-S3-D19.7
type: epic
title: "Epic S3 Day 19.7: Run namespace TTL retention and revise namespace reuse (Issue #74)"
status: planned
owner_role: EM
created_at: 2026-02-20
updated_at: 2026-02-20
related_issues: [74]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-20-issue-74-plan"
---

# Epic S3 Day 19.7: Run namespace TTL retention and revise namespace reuse (Issue #74)

## TL;DR
- Проблема: после завершения run namespace удаляется сразу, из-за чего теряется среда для ревью и диагностики.
- Цель: закрепить delivery execution-plan для реализации role-based TTL retention (`default 24h`) и lease extension/reuse на `run:*:revise`.
- Результат: `full-env` namespace сохраняется на время ревью, cleanup остаётся управляемым, security/RBAC требования не ослабляются.

## Priority
- `P0`.

## Scope
### In scope
- Внедрение policy в runtime orchestration:
  - `spec.webhookRuntime.defaultNamespaceTTL`;
  - `spec.webhookRuntime.namespaceTTLByRole`;
  - fallback `24h`, если policy не задана.
- Namespace lifecycle в worker:
  - lease scheduling при create;
  - lease extension при `run:<stage>:revise`;
  - cleanup sweep по `expires_at` для managed run namespaces.
- Reuse policy для revise:
  - ключ `(project, issue_number, agent_key)`;
  - reuse существующего активного namespace;
  - recreate, если namespace отсутствует или `Terminating`.
- Runtime/audit observability:
  - события `run.namespace.ttl_scheduled`, `run.namespace.ttl_extended`, `run.namespace.cleaned`, `run.namespace.cleanup_failed`;
  - статусные сообщения run с `expires_at` и признаком reuse.
- Обновление traceability и evidence bundle в рамках `run:dev`.

### Out of scope
- Введение отдельного debug-label для manual-retention.
- Новая UI-функциональность policy-редактирования TTL (вынесено в техдолг после MVP).
- Изменение RBAC модели доступа к Kubernetes `secrets` (доступ остаётся запрещённым).

## Декомпозиция
- Story-1: расширить typed contract/loader `services.yaml` полями role-based TTL.
- Story-2: реализовать lease scheduling/extension в namespace lifecycle worker.
- Story-3: реализовать revise reuse по ключу `(project, issue_number, agent_key)` с guardrails.
- Story-4: добавить audit/status evidence для TTL/reuse/cleanup событий.
- Story-5: выполнить regression-пакет и зафиксировать результаты в PR.

## Quality gates
- Planning gate:
  - ADR-0005 согласован как source of truth для решения.
  - Sprint/epic/traceability документы синхронизированы.
- Contract gate:
  - `services.yaml` валидирует TTL длительности и fallback.
  - Для `codex-k8s` проставлены `24h` для всех системных ролей.
- Runtime gate:
  - после `run:dev` namespace не удаляется immediately;
  - в metadata видны `namespace_lease_ttl` и `namespace_lease_expires_at`.
- Revise gate:
  - `run:*:revise` продлевает lease от текущего времени;
  - при валидном активном namespace выполняется reuse, иначе create-new.
- Cleanup gate:
  - sweep удаляет только managed run namespace (`codex-k8s.dev/managed-by=codex-k8s-worker`, `codex-k8s.dev/namespace-purpose=run`);
  - expired namespace удаляется с audit-событием.
- Security gate:
  - RBAC ограничения сохраняются: `secrets` read/write в run namespace недоступны.

## Критерии приемки
- В `services.yaml` есть `defaultNamespaceTTL: 24h` и per-role `namespaceTTLByRole` для всех системных ролей.
- Для `full-env` run namespace живёт не менее заданного TTL и доступен для manual review.
- Для `run:*:revise` подтверждён reuse того же namespace с продлением `expires_at`.
- Cleanup не затрагивает не-managed namespace и не удаляет активные namespace до expiry.
- В `flow_events` присутствуют события TTL schedule/extend/cleanup.
- Набор проверок (`tests/lint/runtime checks`) приложен в PR evidence.

## Блокеры, риски и owner decisions
### Блокеры
- На этапе планирования блокеров нет; dev-реализация зависит от подтверждения приоритета Day19.7 перед Day20 e2e.

### Риски
- Рост потребления ресурсов кластера при увеличении числа активных namespace.
- Гонки lease extension при конкурентных revise-run.
- Риск "зависших" namespace при сбоях cleanup sweep.

### Owner decisions (required)
1. Подтвердить policy `24h` как дефолт и как стартовое per-role значение для всех системных ролей.
2. Подтвердить отказ от manual-retention label в пользу единой TTL lease-policy.
3. Подтвердить operational target: cleanup expired namespace выполняется в пределах одного sweep-цикла без ручного вмешательства.

## Handover
- `dev`: реализация Story-1..5 в одном `run:dev` цикле с runtime evidence.
- `qa`: targeted regression по retention/revise/cleanup сценариям.
- `sre`: проверка влияния на capacity и observability в production namespace.
