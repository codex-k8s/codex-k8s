---
doc_id: TRH-CK8S-S12-0001
type: traceability-history
title: "Sprint S12 Traceability History"
status: completed
owner_role: KM
created_at: 2026-03-13
updated_at: 2026-03-14
related_issues: [366, 413, 416, 418, 420, 423, 425, 426, 427, 428, 429, 430, 431]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-03-13-traceability-s12-history"
  approved_by: "ai-da-stas"
  approved_at: 2026-03-13
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

## Актуализация по Issue #420 (`run:design`, 2026-03-13)
- Подготовлен design package:
  - `docs/delivery/epics/s12/epic-s12-day5-github-api-rate-limit-design.md`;
  - `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/README.md`;
  - `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/design_doc.md`;
  - `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/api_contract.md`;
  - `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/data_model.md`;
  - `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/migrations_policy.md`.
- Зафиксированы:
  - новый coarse wait-state `waiting_backpressure` и business reason `github_rate_limit` без reuse `waiting_mcp`;
  - persisted model `github_rate_limit_waits` + `github_rate_limit_wait_evidence`, dominant wait election и typed linkage через `agent_runs.wait_target_kind=github_rate_limit_wait`;
  - internal callback contract `ReportGitHubRateLimitSignal`, finite auto-resume policy для primary/secondary limits и typed manual-action guidance без отдельного operator write endpoint;
  - best-effort статус GitHub service-comment как mirror staff/private projection с отдельным retry path при platform contour saturation.
- Для continuity создана follow-up issue `#423` (`run:plan`) без trigger-лейбла.
- Внешний baseline дополнительно сверен:
  - Context7 `/github/docs` использован для актуальной верификации primary/secondary rate-limit semantics, `Retry-After`, response headers и best-practice guidance по pacing/backoff;
  - локально подтверждён non-interactive GitHub flow через `gh issue create --help`, `gh pr create --help`, `gh pr edit --help`.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась по существу: design stage переводит architecture package в implementation-ready contracts и rollout notes, а не добавляет новые канонические требования.

## Актуализация по Issue #423 (`run:plan`, 2026-03-13)
- Подготовлен plan package:
  - `docs/delivery/epics/s12/epic-s12-day6-github-api-rate-limit-plan.md`;
  - `docs/delivery/sprints/s12/sprint_s12_github_api_rate_limit_resilience.md`;
  - `docs/delivery/epics/s12/epic_s12.md`;
  - `docs/delivery/delivery_plan.md`;
  - `docs/delivery/issue_map.md`.
- Зафиксированы:
  - execution waves `#425..#431` по контурам schema foundation, `control-plane`, `worker`, `agent-runner`, `api-gateway`, `web-console` и observability/readiness gate;
  - rollout order `migrations -> control-plane -> worker -> agent-runner -> api-gateway -> web-console`;
  - quality gates, DoR/DoD, blockers, risks и owner decisions для handover в `run:dev`;
  - обязательный gate `#431` перед handover в `run:qa`;
  - frontmatter для sprint package, epic catalog, day6 plan epic и traceability history нормализован в завершённое состояние, чтобы document contour `intake -> vision -> prd -> arch -> design -> plan` не выглядел как продолжающийся implementation stage;
  - канонический Day6 epic для Issue `#423` закреплён как `docs/delivery/epics/s12/epic-s12-day6-github-api-rate-limit-plan.md`; конкурирующий дубль `docs/delivery/epics/s12/epic-s12-day6-github-api-rate-limit-resilience-plan.md` удалён в revise-итерации, чтобы сохранить единый source of truth;
  - факт, что документный контур `intake -> vision -> prd -> arch -> design -> plan` согласован и завершён, а дальнейший handover остаётся owner-managed.
- Для continuity созданы follow-up issues `#425`, `#426`, `#427`, `#428`, `#429`, `#430`, `#431` без trigger-лейблов.
- Внешний baseline дополнительно сверен:
  - Context7 `/github/docs` использован для повторной проверки primary/secondary rate-limit semantics, `Retry-After`, guidance `wait at least one minute` и exponential backoff;
  - локально подтверждён non-interactive GitHub flow через `gh issue create --help`, `gh pr create --help`, `gh pr edit --help`.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась по существу: plan stage декомпозирует delivery waves, evidence и quality gates, но не вводит новые канонические требования.

## Актуализация по Issue #425 (`run:dev`, 2026-03-14)
- Реализован foundation stream `S12-E01`:
  - `services/internal/control-plane/cmd/cli/migrations/20260314110000_day30_github_rate_limit_wait_foundation.sql`;
  - доменные enum/value/entity/query типы и rollout guard в `services/internal/control-plane/internal/domain/{githubratelimit,repository/githubratelimitwait,types/...}`;
  - PostgreSQL repository foundation в `services/internal/control-plane/internal/repository/postgres/githubratelimitwait/*`.
- Зафиксированы:
  - additive schema для `github_rate_limit_waits` и `github_rate_limit_wait_evidence`, partial unique индексы для open wait per contour и dominant wait per run, а также enum/check expansion для `agent_runs.status`, `agent_runs.wait_reason`, `agent_runs.wait_target_kind` и `agent_sessions.wait_state`;
  - transactional `RefreshRunProjection`, который выбирает dominant wait и синхронизирует typed linkage в `agent_runs` / `agent_sessions`, не перетирая чужой wait-context вне `github_rate_limit`;
  - отдельные rollout guards `schema -> domain -> worker -> runner -> transport -> ui`, чтобы последующие волны `#426..#430` не обходили sequencing из Day5/Day6 package;
  - unit coverage для dominant wait election и rollout guard logic, плюс migration guard test на обязательные DDL/index/enum expansion элементы.
- Проверки:
  - `go test ./services/internal/control-plane/internal/domain/githubratelimit ./services/internal/control-plane/internal/repository/postgres/githubratelimitwait ./services/internal/control-plane/cmd/cli/migrations`
  - `go test ./services/internal/control-plane/...`
  - `make lint-go`
  - `make dupl-go`
  - `git diff --check`
- Внешний baseline дополнительно сверен:
  - Context7 `/jackc/pgx` использован для перепроверки idiomatic transaction + row-locking patterns (`BeginTx`, safe `defer Rollback`, `CollectRows`) перед реализацией repository refresh path.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась по существу: Issue `#425` закрывает foundation implementation wave и не вводит новых продуктовых требований.
