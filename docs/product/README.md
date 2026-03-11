---
doc_id: IDX-CK8S-PRODUCT-0001
type: domain-index
title: "Product Documentation Index"
status: in-review
owner_role: PM
created_at: 2026-03-11
updated_at: 2026-03-11
related_issues: [320]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-11-product-index"
---

# Product Documentation

## TL;DR
- `docs/product/` хранит продуктовый source of truth для модели платформы.
- Здесь не размещаются delivery-планы, sprint day-эпики и runbooks.

## Канонические документы
- `docs/product/requirements_machine_driven.md` — полный baseline требований.
- `docs/product/brief.md` — краткая формулировка продукта и ценности.
- `docs/product/constraints.md` — продуктовые и эксплуатационные ограничения.
- `docs/product/agents_operating_model.md` — роли агентов и operating model.
- `docs/product/labels_and_trigger_policy.md` — правила лейблов и trigger flow.
- `docs/product/stage_process_model.md` — stage model и порядок прохождения этапов.

## Правила
- Product docs описывают цели, ограничения и policy, но не дублируют sprint execution.
- Если документ начинает описывать delivery sequencing или реализацию сервиса, ссылка должна вести в `docs/delivery/` или `docs/architecture/`.
