---
doc_id: DS-ISSUE-0001
type: docset-issue
title: "DocSet — Issue #1"
status: draft
owner_role: KM
created_at: 2026-02-06
updated_at: 2026-02-06
issue:
  number: 1
  repo: "codex-k8s/codex-k8s"
  url: ""
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-06-mvp"
---

# DocSet: Issue #1 — codex-k8s bootstrap

## Обязательные артефакты (для этой категории работы)
> Ставьте ✅/❌ и ссылки. Заполняется/обновляется через MCP.
- [x] Brief (BRF-): `docs/product/brief.md`
- [x] Constraints (CST-): `docs/product/constraints.md`
- [ ] PRD (PRD-): TBD
- [x] Architecture (C4): `docs/architecture/c4_context.md`, `docs/architecture/c4_container.md`
- [x] ADR (если нужно): `docs/architecture/adr/ADR-0001-kubernetes-only.md`, `docs/architecture/adr/ADR-0002-webhook-driven-and-deploy-workflows.md`, `docs/architecture/adr/ADR-0003-postgres-jsonb-pgvector.md`, `docs/architecture/adr/ADR-0004-repository-provider-interface.md`
- [ ] Design Doc: TBD
- [x] Delivery Plan: `docs/delivery/delivery_plan.md`
- [ ] DoD: TBD
- [ ] Test Strategy/Plan: TBD
- [ ] Release Plan/Rollback: TBD
- [ ] SLO/Alerts/Runbook (если затрагивает прод): TBD

## Связанные документы (полный список)
- `docs/product/brief.md`
- `docs/product/constraints.md`
- `docs/architecture/c4_context.md`
- `docs/architecture/c4_container.md`
- `docs/architecture/data_model.md`
- `docs/architecture/api_contract.md`
- `docs/architecture/adr/ADR-0001-kubernetes-only.md`
- `docs/architecture/adr/ADR-0002-webhook-driven-and-deploy-workflows.md`
- `docs/architecture/adr/ADR-0003-postgres-jsonb-pgvector.md`
- `docs/architecture/adr/ADR-0004-repository-provider-interface.md`
- `docs/delivery/delivery_plan.md`
- `docs/delivery/roadmap.md`
- `docs/delivery/epic.md`
- `docs/delivery/issue_map.md`

## Связанные PR
- TBD

## Статус и блокеры
- state: in-progress
- blocked: waiting for first staging bootstrap execution and smoke results
- need: staging server access and initial bootstrap run

## Learning mode scope for this issue
- MVP behavior:
  - user включает toggle learning mode в UI;
  - для его задач в prompt/context добавляется explain block (why/tradeoffs);
  - после PR worker публикует образовательный комментарий (summary + optional line-level).
- Acceptance:
  - хотя бы один test PR в staging содержит корректные объяснения и не раскрывает секреты.

## Решения/апрувы
- request_id: owner-2026-02-06-mvp
- Решение: approved
- Комментарий: Owner решения по MVP зафиксированы.
