---
doc_id: TRH-CK8S-S9-0001
type: traceability-history
title: "Sprint S9 Traceability History"
status: in-review
owner_role: KM
created_at: 2026-03-12
updated_at: 2026-03-12
related_issues: [333, 335, 337, 340, 351, 363, 369, 370, 371, 372, 373, 374, 375]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-12-traceability-s9-history"
---

# Sprint S9 Traceability History

## TL;DR
- Этот файл хранит historical delta для Sprint S9.
- Текущая master-карта связей остаётся в `docs/delivery/issue_map.md`.
- Текущее покрытие FR/NFR остаётся в `docs/delivery/requirements_traceability.md`.

## Актуализация по Issue #333 (`run:intake`, 2026-03-12)
- Intake зафиксировал Mission Control Dashboard как отдельную product initiative, а не как локальный refactor staff console.
- В качестве baseline зафиксированы:
  - active-set control-plane UX;
  - discussion-first workflow;
  - provider-safe actions и command/reconciliation framing;
  - GitHub-first MVP и external human review.
- Создана follow-up issue `#335` для stage `run:vision`.

## Актуализация по Issue #335 (`run:vision`, 2026-03-12)
- Vision stage закрепил mission, north star, persona outcomes, KPI и guardrails для Mission Control Dashboard.
- Явно отделены MVP scope, Post-MVP scope и conditional voice candidate stream.
- Зафиксирована обязательная continuity-инструкция: на PRD stage сохранить active-set default, GitHub-first MVP, external review и command/reconciliation guardrails.
- Создана follow-up issue `#337` для stage `run:prd`.

## Актуализация по Issue #337 (`run:prd`, 2026-03-12)
- Подготовлен PRD package:
  - `docs/delivery/epics/s9/epic-s9-day3-mission-control-dashboard-prd.md`;
  - `docs/delivery/epics/s9/prd-s9-day3-mission-control-dashboard.md`.
- Зафиксированы:
  - user stories `S9-US-01..S9-US-05`;
  - product waves `Wave 1 -> Wave 2 -> Wave 3`;
  - FR/AC/NFR, edge cases и expected evidence для active-set dashboard, discussion formalization, provider-safe commands, dedupe/reconciliation и degraded fallback.
- Принято продуктовое решение: voice intake остаётся conditional `Wave 3` и не блокирует core MVP.
- Для continuity создана follow-up issue `#340` (`run:arch`) без trigger-лейбла.
- Для GitHub automation перед созданием follow-up issue и будущего PR-flow через Context7 подтверждён актуальный CLI-синтаксис `gh issue create`, `gh pr create`, `gh pr edit` (`/websites/cli_github_manual`).
- Проверки stage-пакета: markdown-only self-check, traceability sync, `git diff --check`.

## Актуализация по Issue #340 (`run:arch`, 2026-03-12)
- Подготовлен architecture package:
  - `docs/delivery/epics/s9/epic-s9-day4-mission-control-dashboard-arch.md`;
  - `docs/architecture/initiatives/s9_mission_control_dashboard/architecture.md`;
  - `docs/architecture/initiatives/s9_mission_control_dashboard/c4_context.md`;
  - `docs/architecture/initiatives/s9_mission_control_dashboard/c4_container.md`;
  - `docs/architecture/adr/ADR-0011-mission-control-dashboard-active-set-projection-and-command-reconciliation.md`;
  - `docs/architecture/alternatives/ALT-0003-mission-control-dashboard-projection-and-realtime-trade-offs.md`.
- Зафиксированы:
  - ownership split для active-set projection, relations, timeline/comments mirror и command lifecycle;
  - worker-owned provider sync/retries/reconciliation path;
  - snapshot-first / delta-second realtime baseline с degraded mode через explicit refresh и list fallback;
  - isolated voice candidate stream без blocking impact на core MVP.
- Для continuity подготовлена follow-up issue `#351` (`run:design`) без trigger-лейбла.
- Через Context7 повторно подтверждён актуальный Mermaid C4 syntax (`/mermaid-js/mermaid`) и GitHub CLI syntax для issue/PR handover (`/websites/cli_github_manual`).
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась, потому что architecture stage не изменяет канонический requirements baseline, а уточняет service boundaries и trade-offs.

## Актуализация по Issue #351 (`run:design`, 2026-03-12)
- Подготовлен Day5 design package:
  - `docs/delivery/epics/s9/epic-s9-day5-mission-control-dashboard-design.md`;
  - `docs/architecture/initiatives/s9_mission_control_dashboard/design_doc.md`;
  - `docs/architecture/initiatives/s9_mission_control_dashboard/api_contract.md`;
  - `docs/architecture/initiatives/s9_mission_control_dashboard/data_model.md`;
  - `docs/architecture/initiatives/s9_mission_control_dashboard/migrations_policy.md`.
- Зафиксированы:
  - typed HTTP/gRPC/realtime contracts для snapshot, entity details, command admission/status и optional voice candidate flow;
  - hybrid persisted projection model (`entities + relations + timeline + commands + optional voice_candidates` under `control-plane` ownership);
  - inline write-path только для provider-safe typed commands, при этом PR review/merge/comment editing остаются provider deep-link-only;
  - rollout discipline `migrations -> control-plane -> worker -> api-gateway -> web-console` и limited rollback после provider side effects.
- Через Context7 подтверждён актуальный GitHub CLI syntax для continuity issue и PR flow (`/websites/cli_github_manual`).
- Для continuity подготовлена follow-up issue `#363` (`run:plan`) без trigger-лейбла.

## Актуализация по Issue #363 (`run:plan`, 2026-03-12)
- Подготовлен Day6 plan package:
  - `docs/delivery/epics/s9/epic-s9-day6-mission-control-dashboard-plan.md`.
- Зафиксированы:
  - execution streams `S9-E01..S9-E07` с явным split `schema foundation -> domain -> worker warmup/reconcile -> core transport -> UI -> observability -> conditional voice transport`;
  - handover issues `#369..#375` без trigger-лейблов и с owner-managed wave sequencing;
  - правило, что `#371` владеет фактическим warmup/backfill execution gate, `#372` ограничен core transport paths, а voice-specific OpenAPI/codegen вынесены в `#375`;
  - правило, что `#374` закрывает observability/rollout-readiness gate перед `run:qa`, а `#375` остаётся optional continuation и не блокирует core MVP rollout.
- Попытка использовать Context7 для GitHub CLI manual завершилась ошибкой `Monthly quota exceeded`; неинтерактивный issue/PR flow дополнительно сверен локально по `gh issue create --help`, `gh pr create --help`, `gh pr edit --help`.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась, потому что plan-stage обновляет delivery governance и handover backlog, а не канонический product requirements baseline.

## Актуализация по Issue #369 (`run:dev`, 2026-03-13)
- Реализован foundation stream `S9-E01` в `services/internal/control-plane`:
  - additive goose migration `20260313133000_day29_mission_control_foundation.sql` с таблицами `mission_control_entities`, `mission_control_relations`, `mission_control_timeline_entries`, `mission_control_commands` и индексами active-set / timeline / command lookup;
  - доменные типы `internal/domain/types/{entity,enum,query,value}/mission_control.go` и repository contract `internal/domain/repository/missioncontrol/repository.go`;
  - rollout guard + worker warmup entry contract в `internal/domain/missioncontrol/{rollout_guard.go,contract.go}`;
  - PostgreSQL foundation repository `internal/repository/postgres/missioncontrol/**` с `projection_version` CAS update, relation replace, timeline upsert, command ledger access и warmup summary query.
- Зафиксированы guardrails:
  - core Mission Control path остаётся закрытым без `CODEXK8S_MISSION_CONTROL_ENABLED`;
  - read-path требует `schema + domain + warmup verification`;
  - realtime требует `read-path`;
  - write-path и optional voice path не открываются до прохождения предыдущих gates.
- Через Context7 подтверждён актуальный pgx v5 baseline для scan JSONB/arrays/timestamptz (`/jackc/pgx`), после чего repository foundation собран на `pgx.RowToStructByName`, `pgtype.*` и typed `json.RawMessage` payloads.
- Проверки:
  - `go test ./services/internal/control-plane/...`
  - `go test ./services/internal/control-plane/internal/domain/missioncontrol ./services/internal/control-plane/internal/repository/postgres/missioncontrol ./services/internal/control-plane/cmd/cli/migrations`
  - `go mod tidy`
  - `make lint-go`
  - `make dupl-go`
  - `git diff --check`
- `goose` dry-run/apply не выполнялись, потому что CLI отсутствует в runtime image; миграция проверялась через compile/test path и SQL self-check.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась, потому что `#369` реализует foundation schema/repository layer без изменения канонического product baseline.
