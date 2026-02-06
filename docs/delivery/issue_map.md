---
doc_id: MAP-CK8S-0001
type: issue-map
title: "Issue ↔ Docs Map"
status: draft
owner_role: KM
created_at: 2026-02-06
updated_at: 2026-02-06
---

# Issue ↔ Docs Map

## TL;DR
Матрица трассируемости: Issue/PR ↔ документы ↔ релизы.

## Матрица
| Issue/PR | DocSet | PRD | Design | ADRs | Test Plan | Release Notes | Postdeploy | Status |
|---|---|---|---|---|---|---|---|---|
| #1 | `docs/_docset/issues/issue-0001-codex-k8s-bootstrap.md` | `docs/product/requirements_machine_driven.md` + `docs/product/brief.md` + `docs/product/constraints.md` | `docs/architecture/c4_context.md`, `docs/architecture/c4_container.md`, `docs/architecture/data_model.md`, `docs/architecture/api_contract.md` | `ADR-0001..0004` | learning mode smoke TBD | TBD | TBD | in-progress |

## Требования и трассировка
- Source of truth требований: `docs/product/requirements_machine_driven.md`.
- Матрица трассируемости: `docs/delivery/requirements_traceability.md`.

## Правила
- Если нет обязательного документа — статус `blocked`.
- Ссылки должны быть кликабельны.
- Матрица обновляется при каждом переходе этапа Delivery Plan.
