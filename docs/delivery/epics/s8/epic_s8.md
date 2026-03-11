---
doc_id: EPC-CK8S-0008
type: epic
title: "Epic Catalog: Sprint S8 (Go refactoring parallelization + repository onboarding automation)"
status: in-progress
owner_role: EM
created_at: 2026-02-27
updated_at: 2026-03-11
related_issues: [223, 225, 226, 227, 228, 229, 230, 281, 282, 320]
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
- В текущей ревизии добавлен S8-E09: единый documentation-governance поток для выравнивания IA, migration-map и ссылочной синхронизации.
- Каталог фиксирует независимые bounded scopes для параллельной разработки, repository onboarding и documentation governance.

## Execution backlog

| Epic ID | GitHub Issue | Scope |
|---|---|---|
| S8-E01 | `#225` | control-plane: decomposition больших domain/transport файлов |
| S8-E02 | `#226` | api-gateway: cleanup transport handlers и boundary hardening |
| S8-E03 | `#227` | worker: decomposition service и cleanup duplication |
| S8-E04 | `#228` | agent-runner: normalization helpers и dedup prompt context |
| S8-E05 | `#229` | shared libs: pgx alignment + modularization `servicescfg` |
| S8-E06 | `#230` | cross-service hygiene closure и residual debt report |
| S8-E07 | `#281` | empty repository initialization: default branch + `services.yaml` + docs scaffold |
| S8-E08 | `#282` | existing repository adoption: deterministic scan + bootstrap PR with `services.yaml` and docs |
| S8-E09 | `#320` | documentation IA refactor: canonical `docs/index.md`, domain `README.md`, migration-map, services.yaml sync, open-issue link remediation |

## Cross-stream dependency
- `S8-E09/#320` должен завершить каноническую docs IA до финализации execution-потоков `#281/#282`, иначе onboarding будет генерировать устаревший docs baseline.
- Governance source of truth для `S8-E09` остаётся в `docs/delivery/development_process_requirements.md`; отдельный competing policy-doc не вводится.
