---
doc_id: EPC-CK8S-S2-D3
type: epic
title: "Epic S2 Day 3: Per-issue namespace orchestration and RBAC baseline"
status: planned
owner_role: EM
created_at: 2026-02-10
updated_at: 2026-02-10
related_issues: []
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
---

# Epic S2 Day 3: Per-issue namespace orchestration and RBAC baseline

## TL;DR
- Цель эпика: исполнять dev/revise runs в изолированном namespace с доступом к нужному стеку.
- Ключевая ценность: воспроизводимость, изоляция и управляемость прав.
- MVP-результат: для каждого run создаётся namespace (или выбирается пул), в нём запускается агентный Job.

## Priority
- `P0`.

## Scope
### In scope
- Создание namespace по шаблону имени (например, `codex-issue-<id>` или `codex-run-<run_id>`).
- Создание/применение RBAC для агентного service account (минимально необходимые права).
- Политики ресурсов: quotas/limits (минимальный baseline).
- Запись lifecycle событий namespace/job в БД (audit/flow_events).

### Out of scope
- Продвинутая network policy матрица (будет отдельным hardening эпиком).

## Критерии приемки эпика
- Run исполняется в отдельном namespace.
- Namespace может быть безопасно убран/переиспользован без утечек слотов и объектов.

