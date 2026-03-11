---
doc_id: ARC-S7-0001
type: architecture-design
title: "Sprint S7 Day 4 — MVP readiness gap closure architecture (Issue #222)"
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

# Sprint S7 Day 4 — MVP readiness gap closure architecture (Issue #222)

## TL;DR
- Зафиксирована архитектурная декомпозиция потоков `S7-E01..S7-E18` по сервисным границам, ownership и контрактному impact.
- Подтверждён architecture-level sequencing для перехода `run:arch -> run:design -> run:plan` без нарушения bounded-context границ.
- Закреплён parity-gate контракт перед `run:dev`: `approved_execution_epics_count == created_run_dev_issues_count`.

## Контекст и входные артефакты
- Родительская цепочка: `#212 (intake) -> #218 (vision) -> #220 (prd) -> #222 (arch)`.
- Source of truth:
  - `docs/product/requirements_machine_driven.md`
  - `docs/product/labels_and_trigger_policy.md`
  - `docs/product/stage_process_model.md`
  - `docs/delivery/epics/s7/prd-s7-day3-mvp-readiness-gap-closure.md`
  - `docs/delivery/development_process_requirements.md`

## Цели архитектурного этапа
- Разложить `S7-E01..S7-E18` на bounded изменения без смешения зон `external|internal|jobs|staff`.
- Зафиксировать owner для каждого потока и запретить «скрытую» доменную логику в edge/frontend.
- Подготовить handover-контракт для `run:design`: API/data/migration/quality-gates.

## Service Boundaries And Ownership Matrix

| Stream | Сервисные границы (основной scope) | Owner роли | Contract impact | Data/runtime impact | Риски и mitigation |
|---|---|---|---|---|---|
| S7-E01 Rebase/mainline hygiene | Delivery/governance-only; runtime сервисы не меняются | EM + Dev | Нет transport-изменений | Нет schema/runtime изменений | Риск конфликтов веток; mitigation: обязательный rebase gate до merge/review |
| S7-E02 Sidebar cleanup | `services/staff/web-console` (UI маршрутизация) | Dev (frontend) + SA review | Возможен remove unused staff routes; public API не меняется | Нет БД-изменений | Риск битых deep-link; mitigation: маршрутный smoke + явный route inventory |
| S7-E03 Global filter removal | `services/staff/web-console` | Dev (frontend) + SA review | Нет backend contracts | Нет БД-изменений | Риск скрытых зависимостей фильтра; mitigation: code search + cleanup checklist |
| S7-E04 Runtime-deploy/images UI de-scope | `services/staff/web-console` (+ при необходимости thin-edge staff DTO cleanup) | Dev (frontend/backend transport) | Возможное удаление неиспользуемых staff DTO endpoints | Нет БД-изменений | Риск удаления используемого пути; mitigation: inventory endpoint->screen mapping |
| S7-E05 Agents table cleanup | `services/staff/web-console`, опционально staff DTO projection | Dev + SA | Возможен UI-level reshape DTO usage без изменения домена | Нет БД-изменений | Риск рассинхрона UI/DTO; mitigation: typed API snapshot checks |
| S7-E06 Agents settings de-scope (runtime mode/locale) | `services/staff/web-console` + `services/external/api-gateway` + `services/internal/control-plane` (staff use-case boundary) | SA + Dev | Требуется contract cleanup staff API (убрать non-MVP поля) | Вероятен data migration policy на soft-ignore legacy fields | Риск backward несовместимости UI; mitigation: stage-gated contract-first change |
| S7-E07 Prompt source repo-only policy | `services/internal/control-plane` + `services/jobs/worker` + `services/external/api-gateway` + `services/staff/web-console` | SA (domain) + Dev | Contract change: исключить `repo|db` selector в MVP | Возможен data-state normalization для source flags | Риск дрейфа между policy и runtime; mitigation: единый policy source + flow_events audit |
| S7-E08 Agents UX de-scope hardening | `services/staff/web-console` (MVP-only actions) | Dev + SA | Без новых контрактов, только remove non-MVP actions | Нет БД-изменений | Риск скрытого non-MVP action path; mitigation: UX action matrix |
| S7-E09 Runs UX cleanup + namespace delete | `services/staff/web-console` + `services/external/api-gateway` (если надо typed action DTO) + `services/internal/control-plane` (existing action semantics) | SA + Dev + SRE consult | Возможна корректировка staff runtime action DTO | Нет новых таблиц; runtime policy check обязателен | Риск destructive action без guardrails; mitigation: confirm + audit + RBAC check |
| S7-E10 Runtime deploy cancel/stop | `services/internal/control-plane` + `services/jobs/worker` + thin-edge adapters | SA (domain) + SRE + Dev | Новый/уточнённый staff action contract для cancel/stop | Изменение runtime task state-machine, без cross-service shared DB | Риск гонок состояния; mitigation: idempotent transition + lease-aware guard |
| S7-E11 mode:discussion reliability | `services/internal/control-plane` webhook/runstatus domain + `services/jobs/worker` orchestration | SA + Dev | Contract impact минимальный (policy/flow), external API без расширений | Нет БД-схемы на arch этапе | Риск false trigger; mitigation: deterministic label resolver + ambiguity stop |
| S7-E12 Final readiness gate | Delivery/QA/OPS docs governance | EM + QA + SRE + KM | Нет transport-изменений | Нет runtime-схемы, только evidence process | Риск неполного evidence; mitigation: обязательный gate checklist |
| S7-E13 late-stage revise policy coverage | Product policy + `control-plane` stage resolver | SA + QA + Dev | Contract impact в label/stage policy path | Runtime impact: добавить revise routes для `doc-audit|qa|release|postdeploy|ops|self-improve` | Риск конфликтов label set; mitigation: resolver precedence + `need:input` fallback |
| S7-E14 QA DNS-path policy | QA process + runbook/docs | QA + SRE + SA review | Нет API изменений | Нет БД-изменений | Риск неподтверждённых проверок; mitigation: обязательный DNS evidence bundle |
| S7-E15 Prompt templates repo-only workflow | Product/prompt policy + control-plane prompt resolver boundary | SA + KM + Dev | Уточнение policy contracts, без нового API surface | Возможна очистка legacy toggles | Риск bypass через UI; mitigation: убрать UI controls + enforce repo workflow |
| S7-E16 `run:intake:revise` false-failed fix | `services/internal/control-plane` runstatus + `services/jobs/worker` finish path | SA + Dev | Нет внешнего API расширения | Runtime logic consistency; без новой схемы на arch этапе | Риск race при финализации; mitigation: terminal-state idempotency rules |
| S7-E17 Self-improve snapshot reliability | `services/jobs/agent-runner` + `services/internal/control-plane` session callbacks + repository layer | SA + KM + Dev | Нет public API изменений | Возможен data access policy review для `agent_sessions` upsert | Риск missing snapshot; mitigation: upsert semantics + retry-safe write path |
| S7-E18 Documentation governance | `docs/delivery/*`, `docs/product/*`, process standards | KM + EM + SA review | Нет runtime contracts | Нет runtime/schema impact | Риск process drift; mitigation: mandatory templates + traceability sync |

## Architecture Sequencing (S7)

### Wave model
1. Foundation wave: `S7-E01`, `S7-E11`, `S7-E13`.
2. UI cleanup wave: `S7-E02`, `S7-E03`, `S7-E04`, `S7-E05`.
3. Agents/prompt MVP de-scope wave: `S7-E06`, `S7-E07`, `S7-E08`, `S7-E15`, `S7-E17`.
4. Runtime operations reliability wave: `S7-E09`, `S7-E10`, `S7-E16`.
5. Acceptance/governance closeout wave: `S7-E14`, `S7-E18`, `S7-E12`.

### Dependency constraints
- `S7-E06` стартует только после UI baseline cleanup (`S7-E05`).
- `S7-E07` стартует только после фиксации решения по `S7-E06` (чтобы не держать dual-source).
- `S7-E10` стартует после `S7-E09` (единый runtime operations UX contract).
- `S7-E12` разрешён только при закрытых P0 потоках и readiness evidence.

## Parity-Gate Architecture Contract (before run:dev)
- Формула:
  - `approved_execution_epics_count == created_run_dev_issues_count`
  - `coverage_ratio = created_run_dev_issues_count / approved_execution_epics_count`
- Требование:
  - `coverage_ratio` должен быть ровно `1.0`.
- Граница ответственности:
  - governance фиксация — `docs/delivery/*` (EM/KM);
  - policy validation и transition guard — `control-plane` (domain/webhook/runstatus);
  - при mismatch — `need:input` и блок перехода в `run:dev`.

## C4 Overlays For S7 Day4
- Context view: `docs/architecture/c4_context_s7_mvp_readiness_gap_closure.md`.
- Container view: `docs/architecture/c4_container_s7_mvp_readiness_gap_closure.md`.

## Migration And Runtime Impact
- На этапе `run:arch` изменения ограничены markdown-артефактами.
- Схема БД, код сервисов, deploy manifests и runtime поведение не менялись.
- Обязательные design-stage выходы:
  - contract-first пакет (staff API/gRPC, где применимо);
  - data model + migration policy для потоков, где будут изменяться persisted поля/состояния;
  - deterministic rollback notes по runtime state transitions.
- Rollout order (для будущего run:dev):
  - `migrations -> internal domain services -> edge services -> frontend`.

## Context7 Baseline (актуализация)
- Подтверждён актуальный CLI-синтаксис для handover automation:
  - Context7 library: `/websites/cli_github_manual` (`gh issue create`, `gh pr create`, `gh pr edit`).
- Подтверждён актуальный синтаксис C4-диаграмм в Mermaid:
  - Context7 library: `/mermaid-js/mermaid` (`C4Context`, `C4Container`, `Person/System/System_Ext/Rel`).
- Новые внешние зависимости на этапе `run:arch` не требуются.

## Handover In run:design
- Следующий этап: `run:design`.
- Follow-up issue: `#238` (создаётся без trigger-лейбла).
- Обязательные артефакты design-этапа:
  - typed transport contracts по потокам с contract impact (`S7-E06`, `S7-E07`, `S7-E09`, `S7-E10`, `S7-E13`, `S7-E16`, `S7-E17`);
  - data model/migration policy там, где меняются persisted состояния;
  - проверяемый quality-gate набор для parity-gate и sequencing.
