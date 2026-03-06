---
doc_id: EPC-CK8S-0008
type: epic
title: "Epic Catalog: Sprint S8 (Go refactoring parallelization + repository onboarding automation)"
status: in-progress
owner_role: EM
created_at: 2026-02-27
updated_at: 2026-03-06
related_issues: [223, 225, 226, 227, 228, 229, 230, __ISSUE_EMPTY__, __ISSUE_EXISTING__]
related_prs: [231]
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-27-issue-223-plan-revise"
---

# Epic Catalog: Sprint S8 (Go refactoring parallelization + repository onboarding automation)

## TL;DR
- Sprint S8 содержит execution backlog Go-рефакторинга, выделенный из Sprint S7 в отдельный поток.
- Дополнительно в Sprint S8 добавлены два P0 onboarding-эпика для автоматизации подключения пустых и уже существующих проектных репозиториев.
- Каталог фиксирует независимые bounded scopes для параллельной разработки и repository onboarding.

## Execution backlog

| Epic ID | GitHub Issue | Scope |
|---|---|---|
| S8-E01 | `#225` | control-plane: decomposition больших domain/transport файлов |
| S8-E02 | `#226` | api-gateway: cleanup transport handlers и boundary hardening |
| S8-E03 | `#227` | worker: decomposition service и cleanup duplication |
| S8-E04 | `#228` | agent-runner: normalization helpers и dedup prompt context |
| S8-E05 | `#229` | shared libs: pgx alignment + modularization `servicescfg` |
| S8-E06 | `#230` | cross-service hygiene closure и residual debt report |
| S8-E07 | `#__ISSUE_EMPTY__` | empty repository initialization: default branch + `services.yaml` + docs scaffold |
| S8-E08 | `#__ISSUE_EXISTING__` | existing repository adoption: deterministic scan + bootstrap PR with `services.yaml` and docs |
