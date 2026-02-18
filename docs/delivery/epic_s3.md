---
doc_id: EPC-CK8S-0003
type: epic
title: "Epic Catalog: Sprint S3 (MVP completion)"
status: in-progress
owner_role: EM
created_at: 2026-02-13
updated_at: 2026-02-18
related_issues: [19, 45]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
---

# Epic Catalog: Sprint S3 (MVP completion)

## TL;DR
- Sprint S3 завершает MVP и переводит платформу в устойчивый full stage-driven контур.
- Центральные deliverables: full stage labels, staff debug observability, MCP control tools, `run:self-improve` loop, declarative full-env deploy, docset import/sync, unified config/secrets governance, onboarding preflight.
- Дополнительный фокус финальной части S3: закрыть core-flow недоделки (prompt/docs context, env-scoped secrets, runtime error journal, OAuth bypass key, frontend hardening) до полного e2e gate.

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
- Day 12: `docs/delivery/epics/epic-s3-day12-docset-import-and-safe-sync.md`
- Day 13: `docs/delivery/epics/epic-s3-day13-config-and-credentials-governance.md`
- Day 14: `docs/delivery/epics/epic-s3-day14-repository-onboarding-preflight.md`
- Day 15: `docs/delivery/epics/epic-s3-day15-mvp-closeout-and-handover.md`
- Day 16: `docs/delivery/epics/epic-s3-day16-grpc-transport-boundary-hardening.md`
- Day 17: `docs/delivery/epics/epic-s3-day17-environment-scoped-secret-overrides-and-oauth-callbacks.md`
- Day 18: `docs/delivery/epics/epic-s3-day18-runtime-error-journal-and-staff-alert-center.md`
- Day 19: `docs/delivery/epics/epic-s3-day19-run-access-key-and-oauth-bypass.md`
- Day 20: `docs/delivery/epics/epic-s3-day20-frontend-manual-qa-hardening-loop.md`
- Day 21: `docs/delivery/epics/epic-s3-day21-e2e-regression-and-mvp-closeout.md`

## Прогресс
- Day 1 (`full stage and label activation`) завершён и согласован Owner.
- Day 2 (`staff runtime debug console`) завершён и согласован Owner.
- Day 3 (`mcp deterministic secret sync`) завершён.
- Day 4 (`mcp database lifecycle`) завершён.
- Day 5 (`owner feedback + HTTP approver/executor`) завершён.
- Day 6 (`run:self-improve` ingestion/diagnostics) завершён.
- Day 7 (`run:self-improve` updater/PR flow) завершён.
- Day 8 (`agent toolchain auto-extension`) завершён.
- Day 9 (`declarative full-env deploy and runtime parity`) завершён; финальный e2e контур вынесен в Day21.
- Day 10 (`staff console redesign on Vuetify`) завершён.
- Day 11 (`full-env slots + subdomains + TLS`) завершён.
- Day 12 (`docset import + safe sync`) завершён.
- Day 13 (`unified config/secrets governance + GitHub creds fallback`) завершён.
- Day 14 (`repository onboarding preflight`) завершён.
- Day 16 (`gRPC transport boundary hardening`) завершён как refactoring-hygiene эпик по Issue #45.
- В работе остаются Day15/Day17/Day18/Day19/Day20 (core-flow + frontend hardening), после них финальный Day21 full e2e gate.

## Порядок закрытия остатка S3
1. Day15: docs tree + role prompt matrix + GitHub service templates.
2. Day17: environment-scoped secret overrides + OAuth callback strategy.
3. Day18: runtime error journal + staff alert center.
4. Day19: run access key + OAuth bypass.
5. Day20: manual frontend QA hardening loop.
6. Day21: full e2e regression gate + MVP closeout.

## Критерий успеха Sprint S3 (выжимка)
- Все MVP-сценарии из Issue #19 покрыты кодом, тестами и эксплуатационной документацией.
- `run:self-improve` работает как управляемый и аудируемый контур улучшений.
- У Owner есть полный evidence bundle для решения о переходе к post-MVP фазе.
