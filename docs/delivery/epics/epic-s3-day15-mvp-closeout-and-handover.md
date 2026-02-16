---
doc_id: EPC-CK8S-S3-D15
type: epic
title: "Epic S3 Day 15: MVP regression/security gate + closeout and handover"
status: planned
owner_role: EM
created_at: 2026-02-13
updated_at: 2026-02-16
related_issues: [19]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
---

# Epic S3 Day 15: MVP regression/security gate + closeout and handover

## TL;DR
- Цель: провести финальный regression/security прогон по всем MVP-функциям и формально закрыть MVP.
- MVP-результат: финальный evidence bundle + go/no-go протокол + handover пакет для post-MVP roadmap.

## Priority
- `P0`.

## Scope
### In scope
- E2E regression по full stage labels + self-improve + MCP control tools.
- E2E regression по docset import/sync, onboarding preflight и централизованной конфигурации/секретам.
- Security проверки: отсутствие secret leakage, корректность approvals, RBAC-границы.
- Reliability проверки: cleanup, retries, idempotency, resume поведения.
- Проверка staff UI debug и feedback/approval сценариев на Vuetify-контуре.
- Проверка декларативного full-env deploy пути (`services.yaml`) и runtime parity.
- Consolidated MVP report: что реализовано, что отложено, какие риски остались.
- Обновление roadmap/product docs с post-MVP инициативами и их приоритизацией.
- Handover пакеты: runbook updates, operations checklist, governance policy snapshot.
- Формальный owner sign-off `go/no-go` (decision log + критерии).

### Out of scope
- Реализация post-MVP инициатив в коде.
- Полноценный production penetration testing.

## Критерии приемки
- Все P0 сценарии проходят, блокеров нет, риски классифицированы и подтверждены Owner.
- Regression evidence bundle включает проверки stage-flow, MCP approvals, self-improve и full-env deploy.
- Owner получает полный пакет артефактов и подтверждает завершение MVP фазы.
