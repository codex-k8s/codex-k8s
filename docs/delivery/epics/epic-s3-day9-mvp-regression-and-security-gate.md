---
doc_id: EPC-CK8S-S3-D9
type: epic
title: "Epic S3 Day 9: MVP regression and security gate"
status: planned
owner_role: EM
created_at: 2026-02-13
updated_at: 2026-02-13
related_issues: [19]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
---

# Epic S3 Day 9: MVP regression and security gate

## TL;DR
- Цель: провести финальный regression/security прогон по всем MVP-функциям.
- MVP-результат: формальный набор evidence для решения о завершении MVP.

## Priority
- `P0`.

## Scope
### In scope
- E2E regression по full stage labels + self-improve + MCP control tools.
- Security проверки: отсутствие secret leakage, корректность approvals, RBAC-границы.
- Reliability проверки: cleanup, retries, idempotency, resume поведения.
- Проверка staff UI debug сценариев на нагрузочном срезе MVP.

### Out of scope
- Полноценный production penetration testing.

## Критерии приемки
- Все P0 сценарии проходят, блокеров нет, риски классифицированы и подтверждены Owner.
