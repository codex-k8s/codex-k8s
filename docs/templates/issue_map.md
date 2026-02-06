---
doc_id: MAP-XXXX
type: issue-map
title: "Issue ↔ Docs Map"
status: draft
owner_role: KM
created_at: YYYY-MM-DD
updated_at: YYYY-MM-DD
---

# Issue ↔ Docs Map

## TL;DR
Матрица трассируемости: Issue/PR ↔ документы ↔ релизы.

## Матрица
| Issue/PR | DocSet | PRD | Design | ADRs | Test Plan | Release Notes | Postdeploy | Status |
|---|---|---|---|---|---|---|---|---|
| #123 | docset/issues/issue-123.md | PRD-... | DSG-... | ADR-... | TST-... | REL-... | PDR-... | ... |

## Правила
- Если нет обязательного документа — статус “blocked”.
- Ссылки должны быть кликабельны.
