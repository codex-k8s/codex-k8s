---
doc_id: EPC-CK8S-0003
type: epic
title: "Epic Catalog: Sprint S3 (MVP completion)"
status: in-progress
owner_role: EM
created_at: 2026-02-13
updated_at: 2026-02-20
related_issues: [19, 45, 74]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-19-full-docset"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-19
---

# Epic Catalog: Sprint S3 (MVP completion)

## TL;DR
- Sprint S3 завершает MVP и переводит платформу в устойчивый full stage-driven контур.
- Центральные deliverables: full stage labels, staff debug observability, MCP control tools, `run:self-improve` loop, declarative full-env deploy, docset import/sync, unified config/secrets governance, onboarding preflight.
- Дополнительный фокус финальной части S3: закрыть core-flow недоделки (prompt/docs context, env-scoped secrets, runtime error journal, frontend hardening) до полного e2e gate.
- Добавлен P0 эпик Issue #74: role-based retention для `full-env` namespace и reuse/lease-extension на `run:*:revise`.

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
- Day 19: `docs/delivery/epics/epic-s3-day19-frontend-manual-qa-hardening-loop.md`
- Day 19.5: `docs/delivery/epics/epic-s3-day19.5-realtime-event-bus-and-websocket-backplane.md`
- Day 19.6: `docs/delivery/epics/epic-s3-day19.6-staff-realtime-subscriptions-and-ui.md`
- Day 19.7: `docs/delivery/epics/epic-s3-day19.7-run-namespace-ttl-and-revise-reuse.md`
- Day 20: `docs/delivery/epics/epic-s3-day20-e2e-regression-and-mvp-closeout.md`

## Прогресс
- Day 1 (`full stage and label activation`) завершён и согласован Owner.
- Day 2 (`staff runtime debug console`) завершён и согласован Owner.
- Day 3 (`mcp deterministic secret sync`) завершён.
- Day 4 (`mcp database lifecycle`) завершён.
- Day 5 (`owner feedback + HTTP approver/executor`) завершён.
- Day 6 (`run:self-improve` ingestion/diagnostics) завершён.
- Day 7 (`run:self-improve` updater/PR flow) завершён.
- Day 8 (`agent toolchain auto-extension`) завершён.
- Day 9 (`declarative full-env deploy and runtime parity`) завершён; финальный e2e контур вынесен в Day20.
- Day 10 (`staff console redesign on Vuetify`) завершён.
- Day 11 (`full-env slots + subdomains + TLS`) завершён.
- Day 12 (`docset import + safe sync`) завершён.
- Day 13 (`unified config/secrets governance + GitHub creds fallback`) завершён.
- Day 14 (`repository onboarding preflight`) завершён.
- Day 15 (`prompt context overhaul: docs tree + role matrix + GitHub service messages`) завершён и согласован Owner.
- Day 16 (`gRPC transport boundary hardening`) завершён как refactoring-hygiene эпик по Issue #45.
- Day 17 (`environment-scoped secret overrides + OAuth callback strategy`) завершён и согласован Owner (PR #49).
- Day 18 (`runtime error journal + staff alert center`) завершён и согласован Owner (PR #50).
- В работе остаются Day19/Day19.5/Day19.6/Day19.7 (frontend hardening + realtime + namespace retention), после них финальный Day20 full e2e gate.

## Порядок закрытия остатка S3
1. Day19: manual frontend QA hardening loop.
2. Day19.5: PostgreSQL LISTEN/NOTIFY realtime bus + WS backplane (multi-server).
3. Day19.6: staff realtime subscriptions and UI integration.
4. Day19.7: namespace TTL retention + revise reuse/lease extension (Issue #74).
5. Day20: full e2e regression gate + MVP closeout.

## Критерий успеха Sprint S3 (выжимка)
- Все MVP-сценарии из Issue #19 покрыты кодом, тестами и эксплуатационной документацией.
- Runtime lifecycle сценарий Issue #74 закрывает review-friendly retention без ослабления RBAC/security policy.
- `run:self-improve` работает как управляемый и аудируемый контур улучшений.
- У Owner есть полный evidence bundle для решения о переходе к post-MVP фазе.
