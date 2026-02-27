---
doc_id: EPC-CK8S-INDEX-0001
type: epic-index
title: "Epic Index (grouped by sprint)"
status: active
owner_role: EM
created_at: 2026-02-24
updated_at: 2026-02-27
related_issues: [112, 154, 184, 185, 187, 189, 195, 197, 199, 201, 212]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-24-epic-index"
---

# Epic Index

## TL;DR
- Все day-эпики сгруппированы по папкам спринтов (`s1`, `s2`, `s3`, `s4`, `s5`, `s6`, `s7`).
- Каталог эпиков каждого спринта размещён в той же папке.
- Это устраняет смешение day-эпиков разных спринтов в одном каталоге.

## Структура

| Папка | Содержимое | Основной документ |
|---|---|---|
| `docs/delivery/epics/s1/` | Day0..Day7 + каталог Sprint S1 | `docs/delivery/epics/s1/epic_s1.md` |
| `docs/delivery/epics/s2/` | Day0..Day7 (+ Day3.5/Day4.5) + каталог Sprint S2 | `docs/delivery/epics/s2/epic_s2.md` |
| `docs/delivery/epics/s3/` | Day1..Day20 + каталог Sprint S3 | `docs/delivery/epics/s3/epic_s3.md` |
| `docs/delivery/epics/s4/` | Day1 + каталог Sprint S4 | `docs/delivery/epics/s4/epic_s4.md` |
| `docs/delivery/epics/s5/` | Day1 + каталог Sprint S5 | `docs/delivery/epics/s5/epic_s5.md` |
| `docs/delivery/epics/s6/` | Day1..Day6 (intake + vision + PRD package + architecture + design + plan) + каталог Sprint S6 | `docs/delivery/epics/s6/epic_s6.md` |
| `docs/delivery/epics/s7/` | Day1 intake + каталог Sprint S7 (MVP readiness gap closure) + candidate backlog `S7-E01..S7-E12` | `docs/delivery/epics/s7/epic_s7.md` |

## Проверка консистентности
- Для каждого `s<номер>` должны существовать:
  - `epic_s<номер>.md`;
  - минимум один `epic-s<номер>-day*.md`.
- Ссылки на day-эпики не должны указывать на корень `docs/delivery/epics/` без `s<номер>/`.
