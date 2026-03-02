---
doc_id: EPC-CK8S-S7-D4
type: epic
title: "Epic S7 Day 4: Architecture для закрытия MVP readiness gaps (Issue #222)"
status: in-review
owner_role: SA
created_at: 2026-03-02
updated_at: 2026-03-02
related_issues: [212, 218, 220, 222, 238]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-02-issue-222-arch"
---

# Epic S7 Day 4: Architecture для закрытия MVP readiness gaps (Issue #222)

## TL;DR
- Подготовлен architecture package для потоков `S7-E01..S7-E18` с явными сервисными границами, ownership и contract/data impact.
- Закреплён architecture sequencing для перехода `run:arch -> run:design -> run:plan`.
- Подтверждён parity-gate перед `run:dev`: `approved_execution_epics_count == created_run_dev_issues_count`.

## Контекст
- Handover из PRD stage: Issue `#220` + артефакт
  `docs/delivery/epics/s7/prd-s7-day3-mvp-readiness-gap-closure.md`.
- Текущий этап: `run:arch` в Issue `#222`.
- Scope этапа: только markdown-изменения (policy-safe).

## Основные архитектурные артефакты
- Архитектурная декомпозиция потоков:
  - `docs/architecture/s7_mvp_readiness_gap_closure_architecture.md`
- C4 overlays:
  - `docs/architecture/c4_context_s7_mvp_readiness_gap_closure.md`
  - `docs/architecture/c4_container_s7_mvp_readiness_gap_closure.md`
- ADR:
  - `docs/architecture/adr/ADR-0010-s7-mvp-readiness-stream-boundaries-and-parity-gate.md`
- Alternatives:
  - `docs/architecture/alternatives/ALT-0002-s7-mvp-readiness-stream-architecture.md`

## Ключевые решения Stage
- Для каждого `S7-E*` зафиксирован owner и сервисная граница:
  - UI cleanup потоки остаются в `web-console`;
  - policy/reliability потоки закреплены за `control-plane`/`worker`;
  - edge (`api-gateway`) остаётся thin-adapter для typed contracts.
- Sequencing wave-модель подтверждена в architecture artifact:
  1. foundation,
  2. UI cleanup,
  3. agents/prompt de-scope,
  4. runtime reliability,
  5. final readiness/governance closeout.
- Parity-gate закреплён как обязательный architecture guard до `run:dev`.

## Migration и runtime impact
- На этапе `run:arch` код, БД-схема и runtime не менялись.
- Зафиксирован design-handover scope:
  - contract-first уточнения для потоков с transport impact;
  - data model/migration policy для потоков с persisted-state impact;
  - rollback notes для runtime state transitions.

## Context7 верификация
- Проверен актуальный синтаксис GitHub CLI для stage handover/PR flow:
  - `/websites/cli_github_manual`.
- Проверен актуальный синтаксис Mermaid C4 для диаграмм архитектурного пакета:
  - `/mermaid-js/mermaid`.
- Новые внешние зависимости в `run:arch` не требуются.

## Acceptance Criteria (Issue #222)
- [x] Архитектурный артефакт S7 подготовлен и включает матрицу `S7-E01..S7-E18` (границы, ownership, contract/data impact, risk mitigation).
- [x] Зафиксирован architecture sequencing/dependency graph для перехода в `run:design`.
- [x] Подтверждён и формализован parity-gate перед `run:dev`.
- [x] Обновлены traceability документы (`issue_map`, `requirements_traceability`, sprint/epic docs, delivery plan).
- [x] Подготовлена follow-up issue для stage `run:design` без trigger-лейбла.

## Handover в `run:design`
- Следующий этап: `run:design`.
- Follow-up issue: `#238` (без trigger-лейбла; trigger ставит Owner).
- В design-stage обязательно:
  - формализовать typed transport contracts по потокам с contract impact;
  - определить migration policy для потоков с data impact;
  - подготовить execution-quality gates для `run:plan`.
