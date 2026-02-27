---
doc_id: EPC-CK8S-0008
type: epic
title: "Epic Catalog: Sprint S8 (Go refactoring parallelization)"
status: in-progress
owner_role: EM
created_at: 2026-02-27
updated_at: 2026-02-27
related_issues: [223, 225, 226, 227, 228, 229, 230]
related_prs: [231]
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-27-issue-223-plan-revise"
---

# Epic Catalog: Sprint S8 (Go refactoring parallelization)

## TL;DR
- Sprint S8 содержит execution backlog Go-рефакторинга, выделенный из Sprint S7 в отдельный поток.
- Каталог фиксирует независимые bounded scopes для параллельной разработки.

## Execution backlog

| Epic ID | GitHub Issue | Scope |
|---|---|---|
| S8-E01 | `#225` | control-plane: decomposition больших domain/transport файлов |
| S8-E02 | `#226` | api-gateway: cleanup transport handlers и boundary hardening |
| S8-E03 | `#227` | worker: decomposition service и cleanup duplication |
| S8-E04 | `#228` | agent-runner: normalization helpers и dedup prompt context |
| S8-E05 | `#229` | shared libs: pgx alignment + modularization `servicescfg` |
| S8-E06 | `#230` | cross-service hygiene closure и residual debt report |
