---
doc_id: TRH-CK8S-S12-0001
type: traceability-history
title: "Sprint S12 Traceability History"
status: in-review
owner_role: KM
created_at: 2026-03-13
updated_at: 2026-03-13
related_issues: [366, 413, 416, 418, 420]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-13-traceability-s12-history"
---

# Sprint S12 Traceability History

## TL;DR
- Этот файл хранит historical delta для Sprint S12.
- Текущая master-карта связей остаётся в `docs/delivery/issue_map.md`.
- Текущее покрытие FR/NFR остаётся в `docs/delivery/requirements_traceability.md`.

## Актуализация по Issue #366 (`run:intake`, 2026-03-13)
- Intake зафиксировал GitHub API rate-limit resilience как отдельную cross-cutting initiative, а не как локальный retry-баг.
- В качестве baseline зафиксированы:
  - controlled wait-state вместо ложного `failed`;
  - split `platform PAT` vs `agent bot-token`;
  - owner/operator transparency;
  - MCP backpressure semantics на agent path;
  - provider-driven неопределённость primary/secondary rate-limit semantics.
- Создана continuity issue `#413` для stage `run:vision`.
- Root FR/NFR matrix не менялась: intake stage не обновляет канонический requirements baseline, а фиксирует problem/scope/handover для нового delivery stream.

## Актуализация по Issue #413 (`run:vision`, 2026-03-13)
- Подготовлен vision package:
  - `docs/delivery/epics/s12/epic-s12-day2-github-api-rate-limit-vision.md`.
- Зафиксированы:
  - mission и north star для GitHub-first controlled wait capability;
  - persona outcomes для owner/reviewer, operator и agent path;
  - KPI/guardrails для clarity, false-failed prevention, contour attribution и запрета local retry-loop;
  - явное разделение core MVP и deferred streams: notification/adapters и multi-provider governance не блокируют core baseline.
- Для continuity создана follow-up issue `#416` (`run:prd`) без trigger-лейбла.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась, потому что vision stage уточняет mission, KPI и scope boundaries, но не меняет канонический requirements baseline.

## Актуализация по Issue #416 (`run:prd`, 2026-03-13)
- Подготовлен PRD package:
  - `docs/delivery/epics/s12/epic-s12-day3-github-api-rate-limit-prd.md`;
  - `docs/delivery/epics/s12/prd-s12-day3-github-api-rate-limit-resilience.md`.
- Зафиксированы:
  - user stories, FR/AC/NFR и wave priorities для controlled wait-state, contour attribution, transparency surfaces и safe resume/manual-intervention path;
  - explicit edge cases для hard-failure separation, dual-contour wait, secondary-limit ambiguity и запрета infinite local retries;
  - handover decisions, которые нельзя потерять на `run:arch`: GitHub-first baseline, split `platform PAT` vs `agent bot-token`, typed recovery hints, hard-failure separation и deferred scope для predictive budgeting/multi-provider governance.
- Для continuity создана follow-up issue `#418` (`run:arch`) без trigger-лейбла.
- Внешний baseline дополнительно сверен:
  - официальные GitHub Docs `Rate limits for the REST API` и `Best practices for using the REST API` просмотрены 2026-03-13;
  - non-interactive GitHub flow дополнительно подтверждён через Context7 (`/websites/cli_github_manual`) и локальные `gh issue create --help`, `gh pr create --help`, `gh pr edit --help`.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась по существу: PRD stage уточняет product contract и delivery evidence, а в root-матрице синхронизирована только traceability governance и historical package.

## Актуализация по Issue #418 (`run:arch`, 2026-03-13)
- Подготовлен architecture package:
  - `docs/delivery/epics/s12/epic-s12-day4-github-api-rate-limit-arch.md`;
  - `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/README.md`;
  - `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/architecture.md`;
  - `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/c4_context.md`;
  - `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/c4_container.md`;
  - `docs/architecture/adr/ADR-0013-github-rate-limit-controlled-wait-ownership.md`;
  - `docs/architecture/alternatives/ALT-0005-github-rate-limit-wait-state-boundaries.md`.
- Зафиксированы:
  - `control-plane` как owner для classification, controlled wait aggregate, contour attribution, recovery hints и visibility contract;
  - `worker` как owner для time-based wait scheduling, finite auto-resume attempts и escalation в manual-action-required;
  - `agent-runner` как raw-evidence emitter с обязательным stop local retry после handoff;
  - разделение `contour_kind` и `signal_origin`, чтобы сохранить PRD split `platform PAT` vs `agent bot-token` без дублирования доменной логики.
- Для continuity создана follow-up issue `#420` (`run:design`) без trigger-лейбла.
- Внешний baseline дополнительно сверен:
  - официальные GitHub Docs `Rate limits for the REST API` и `Best practices for using the REST API` просмотрены 2026-03-13;
  - Context7 использован для `/websites/cli_github_manual` и `/mermaid-js/mermaid`.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась по существу: architecture stage закрепляет ownership и trade-offs, а не вводит новые канонические требования.
