---
doc_id: EPC-CK8S-0003
type: epic
title: "Epic Catalog: Sprint S3 (MVP completion)"
status: in-progress
owner_role: EM
created_at: 2026-02-13
updated_at: 2026-02-15
related_issues: [19]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
---

# Epic Catalog: Sprint S3 (MVP completion)

## TL;DR
- Sprint S3 завершает MVP и переносит платформу от узкого `run:dev` контура к полному stage-driven циклу.
- Центральные deliverables: full stage labels, staff debug observability, MCP control tools, `run:self-improve` loop, declarative full-env deploy и новый staff UI контур на Vuetify.

## Эпики Sprint S3
- Day 1: `docs/delivery/epics/epic-s3-day1-full-stage-and-label-activation.md`
- Day 2: `docs/delivery/epics/epic-s3-day2-staff-runtime-debug-console.md`
- Day 3: `docs/delivery/epics/epic-s3-day3-mcp-deterministic-secret-sync.md`
- Day 4: `docs/delivery/epics/epic-s3-day4-mcp-database-lifecycle.md`
- Day 5: `docs/delivery/epics/epic-s3-day5-feedback-and-approver-interfaces.md`
- Day 6: `docs/delivery/epics/epic-s3-day6-self-improve-ingestion-and-diagnostics.md`
- Day 7: `docs/delivery/epics/epic-s3-day7-self-improve-updater-and-pr-flow.md`
- Day 8: `docs/delivery/epics/epic-s3-day8-agent-toolchain-auto-extension.md`
- Day 9: `docs/delivery/epics/epic-s3-day9-declarative-full-env-deploy-and-runtime-parity.md`
- Day 10: `docs/delivery/epics/epic-s3-day10-staff-console-vuetify-redesign.md`
- Day 11: `docs/delivery/epics/epic-s3-day11-full-env-slots-and-subdomains.md`
- Day 12: `docs/delivery/epics/epic-s3-day12-mvp-closeout-and-handover.md`

## Прогресс
- Day 1 (`full stage and label activation`) завершён и согласован Owner.
- Day 2 (`staff runtime debug console`) завершён и согласован Owner.
- Day 3 (`mcp deterministic secret sync`) завершён.
- Day 4 (`mcp database lifecycle`) завершён.
- Day 5 (`owner feedback + HTTP approver/executor`) завершён.
- Day 6 (`run:self-improve` ingestion/diagnostics) завершён.
- Day 7 (`run:self-improve` updater/PR flow) завершён.
- Day 8 (`agent toolchain auto-extension`) завершён.
- Day 9..Day12 остаются в статусе planned.

## Критерий успеха Sprint S3 (выжимка)
- Все MVP-сценарии из Issue #19 покрыты кодом, тестами и эксплуатационной документацией.
- `run:self-improve` работает как управляемый и аудируемый контур улучшений.
- У Owner есть полный evidence bundle для решения о переходе к post-MVP фазе.
