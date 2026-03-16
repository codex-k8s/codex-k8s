---
doc_id: TRH-CK8S-S12-0001
type: traceability-history
title: "Sprint S12 Traceability History"
status: completed
owner_role: KM
created_at: 2026-03-13
updated_at: 2026-03-15
related_issues: [366, 413, 416, 418, 420, 423, 425, 426, 427, 428, 429, 430, 431, 500]
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

## Актуализация по Issue #426 (`run:dev`, 2026-03-14)
- Реализован domain stream `S12-E02`:
  - новый service layer в `services/internal/control-plane/internal/domain/githubratelimit/{contract.go,model.go,service.go,service_classification.go,service_projection.go,service_report.go,service_resume_payload.go,service_templates.go,templates/messages_ru.tmpl}`;
  - новые domain value/enum типы в `services/internal/control-plane/internal/domain/types/{enum/github_rate_limit_visibility.go,value/github_rate_limit_signal.go,value/github_rate_limit_visibility.go,value/github_rate_limit_resume_payload.go}`.
- Зафиксированы:
  - canonical `GitHubRateLimitSignal` normalization/classification для primary limit, secondary limit с `Retry-After`, secondary/abuse path без countdown и hard-failure separation без создания wait aggregate;
  - wait aggregate lifecycle поверх foundation repository: idempotent signal dedupe по `signal_id`, upsert open wait per contour, evidence append (`signal_detected`, `classified`), `RefreshRunProjection` и `run.wait.paused` flow-event payload с `github_rate_limit.wait.entered` / `github_rate_limit.manual_action_required`;
  - typed visibility projection с dominant/related waits, comment mirror state, recovery hints, manual-action guidance и best-effort comment render context, который строится только из persisted projection, а не из raw headers/logs;
  - deterministic agent-session resume payload builder для future runner/worker waves без повторного derive semantics from raw stderr/headers.
- Проверки:
  - `go test ./services/internal/control-plane/internal/domain/githubratelimit`
  - `go test ./services/internal/control-plane/...`
  - `make lint-go`
  - `make dupl-go`
  - `git diff --check`
- Внешний baseline дополнительно сверен:
  - Context7 `/github/docs` использован для повторной проверки GitHub guidance по primary/secondary rate limits, `Retry-After`, `x-ratelimit-reset` и backoff discipline перед реализацией classification policy.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась по существу: Issue `#426` закрывает domain semantics/control-plane ownership wave и не добавляет новых продуктовых требований.

## Актуализация по Issue #427 (`run:dev`, 2026-03-14)
- Реализован worker stream `S12-E03`:
  - `services/internal/control-plane/internal/domain/githubratelimit/service_worker.go`;
  - `services/internal/control-plane/internal/domain/runstatus/github_rate_limit_retry.go`;
  - `services/internal/control-plane/internal/domain/staff/github_rate_limit_replay.go`;
  - `services/internal/control-plane/internal/transport/grpc/server_github_rate_limit_worker_methods.go`;
  - `services/jobs/worker/internal/domain/worker/github_rate_limit.go`;
  - `proto/codexk8s/controlplane/v1/controlplane.proto`.
- Зафиксированы:
  - `control-plane` теперь владеет worker-facing `ProcessNextGitHubRateLimitWait` RPC, claim/resume lifecycle, resolved/manual-action evidence append и `run.wait.resumed` flow-event для deterministic replay outcome;
  - `worker` получил bounded sweep loop с отдельным feature flag `CODEXK8S_GITHUB_RATE_LIMIT_WAIT_ENABLED` и лимитом `CODEXK8S_WORKER_GITHUB_RATE_LIMIT_SWEEP_LIMIT`, который обрабатывает due waits до exhaustion/empty queue без собственной domain classification;
  - `run_status_comment_retry`, `platform_github_call_replay` и `agent_session_resume` теперь исполняются через typed replay payloads; agent path создаёт pending resume run с persisted `github_rate_limit_resume_payload`, а replay failure reschedule’ится по finite budget и эскалируется в `manual_action_required` только после реального исчерпания safe attempts;
  - `platform_github_call_replay` для issue stage transition получил CAS snapshot `expected_current_run_labels` и request metadata, поэтому delayed replay не удаляет более новый `run:*` label-set вслепую, а уходит в conflict/manual review path;
  - config/deploy/codegen синхронизированы: proto regenerated, control-plane/worker wiring добавлен в app/grpc/client layers, production manifest и bootstrap example получили новые env.
- Проверки:
  - `make gen-proto-go`
  - `go test ./services/internal/control-plane/internal/domain/githubratelimit ./services/internal/control-plane/internal/domain/runstatus ./services/internal/control-plane/internal/domain/staff`
  - `go test ./services/jobs/worker/internal/domain/worker ./services/jobs/worker/internal/controlplane`
  - `go test ./services/internal/control-plane/internal/app ./services/internal/control-plane/internal/transport/grpc ./services/internal/control-plane/internal/repository/postgres/githubratelimitwait`
  - `go test ./services/jobs/worker/internal/app`
  - `git diff --check`
- Внешний baseline дополнительно сверен:
  - Context7 `/github/docs` повторно использован как source of truth для приоритета `Retry-After`, fallback к `x-ratelimit-reset`, ожидания не меньше минуты и bounded backoff при secondary limit.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась по существу: Issue `#427` закрывает worker orchestration wave и не добавляет новых продуктовых требований.

## Актуализация по Issue #428 (`run:dev`, 2026-03-14)
- Реализован runner stream `S12-E04`:
  - `services/jobs/agent-runner/internal/runner/{service.go,helpers_github_rate_limit_handoff.go,helpers_github_rate_limit_resume_prompt.go,helpers_prompt_templates.go}`;
  - `services/jobs/agent-runner/internal/controlplane/client.go`;
  - `services/internal/control-plane/internal/domain/agentcallback/{interaction_resume_payload.go,service.go}`;
  - `services/internal/control-plane/internal/transport/grpc/{server.go,server_github_rate_limit_runtime_methods.go}`;
  - `proto/codexk8s/controlplane/v1/controlplane.proto`.
- Зафиксированы:
  - `agent-runner` теперь детектирует GitHub rate-limit по stderr/stdout, сохраняет coarse session snapshots `running -> waiting_backpressure`, передаёт typed `ReportGitHubRateLimitSignal` и прекращает local retry-loop после подтверждённого handoff;
  - `control-plane` получил run-bound runtime RPC для runner path: `ReportGitHubRateLimitSignal` маппит hard-failure в `failed_precondition`, а `GetRunGitHubRateLimitResumePayload` отдаёт компактный deterministic JSON из persisted `run_payload`;
  - requeued runner resume path распознаёт correlation prefix `github-rate-limit-resume:*`, требует persisted `github_rate_limit_resume_payload`, восстанавливает последнюю codex session без PR-precondition и prepend'ит typed wait outcome в resume prompt вместо повторного derive semantics из stderr/headers;
  - rollout wiring синхронизирован: `RunnerReady` теперь следует `CODEXK8S_GITHUB_RATE_LIMIT_WAIT_ENABLED`, proto/go codegen обновлён, unit coverage добавлена для handoff detection, resume payload parsing и runtime gRPC callbacks.
- Проверки:
  - `make gen-proto-go`
  - `go test ./services/jobs/agent-runner/internal/runner ./services/jobs/agent-runner/internal/controlplane`
  - `go test ./services/internal/control-plane/internal/domain/agentcallback ./services/internal/control-plane/internal/transport/grpc`
  - `go test ./services/internal/control-plane/internal/app ./services/internal/control-plane/internal/domain/githubratelimit`
  - `go test ./services/jobs/agent-runner/internal/app`
- Внешний baseline дополнительно сверен:
  - Context7 `/github/docs` повторно использован как source of truth для primary/secondary rate-limit semantics, приоритета `Retry-After`, fallback к `x-ratelimit-reset` и запрета на локальный retry ownership после handoff.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась по существу: Issue `#428` закрывает runner handoff/resume wave и не добавляет новых продуктовых требований.

## Актуализация по Issue #429 (`run:dev`, 2026-03-14)
- Реализован transport stream `S12-E05`:
  - `services/external/api-gateway/api/server/api.yaml`;
  - `services/external/api-gateway/internal/transport/http/{models/staff.go,models/realtime.go,casters/controlplane.go,staff_realtime_handler.go}`;
  - `services/internal/control-plane/internal/transport/grpc/{server.go,server_staff_methods.go,server_staff_run_wait_projection.go}`;
  - `proto/codexk8s/controlplane/v1/controlplane.proto`.
- Зафиксированы:
  - contract-first расширение staff visibility contracts: `Run.wait_projection` и typed модели `dominant_wait` / `related_waits`, recovery hint и manual action guidance;
  - additive realtime envelope в `api-gateway`: `wait_entered`, `wait_updated`, `wait_resolved`, `wait_manual_action_required` без доменных решений inside handlers;
  - `control-plane` gRPC transport публикует persisted wait projection в `Run` через domain-owned `GetRunProjection`, сохраняя ownership classification/recovery logic в `control-plane` domain слое;
  - синхронно обновлены generated artifacts `proto` + OpenAPI (`Go` и `TS` codegen) для handover в wave `#430`.
- Проверки:
  - `make gen-proto-go SVC=services/internal/control-plane`
  - `make gen-openapi`
  - `go test ./services/internal/control-plane/internal/transport/grpc`
  - `go test ./services/external/api-gateway/internal/transport/http`
  - `npm --prefix services/staff/web-console run build`
  - `git diff --check`
- Внешний baseline дополнительно сверен:
  - Context7 `/github/docs` был использован в предшествующих S12 waves как source of truth для rate-limit semantics; Wave `#429` не добавляет новых provider assumptions и реализует только transport exposure уже утверждённых typed contracts.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась по существу: Issue `#429` закрывает transport visibility wave и не добавляет новых продуктовых требований.

## Актуализация по Issue #430 (`run:dev`, 2026-03-15)
- Реализован frontend stream `S12-E06`:
  - `services/staff/web-console/src/features/runs/{types.ts,realtime.ts,store.ts,wait-presenters.ts}`;
  - `services/staff/web-console/src/pages/{RunDetailsPage.vue,operations/WaitQueuePage.vue}`;
  - `services/staff/web-console/src/shared/lib/run-waits.test.ts`;
  - `services/staff/web-console/src/i18n/messages/{ru.ts,en.ts}`.
- Зафиксированы:
  - wait queue теперь рендерит typed dominant wait, related waits, contour attribution, comment-mirror state и next-step guidance из `Run.wait_projection`, сохраняя fallback для legacy wait surfaces без projection;
  - run details получили отдельную wait visibility card с dominant/related waits, recovery hint source, attempts budget и manual-action guidance, построенную только из typed DTO без parse raw logs/service-comment;
  - realtime stream `wait_entered|wait_updated|wait_resolved|wait_manual_action_required` теперь отображается в staff UI как отдельная activity feed, а frontend store хранит последние wait envelopes без переноса domain classification из `control-plane`;
  - unit coverage добавлена для wait projection presenters и realtime envelope parsing/builders, чтобы handover в `run:qa` имел machine-checked evidence по UI-мэппингу typed contracts.
- Проверки:
  - `npm --prefix services/staff/web-console run test:unit`
  - `npm --prefix services/staff/web-console run build`
  - `git diff --check`
- Внешний baseline дополнительно сверен:
  - Context7 `/vuetifyjs/vuetify` использован для проверки slot/custom-cell patterns `VDataTable` перед обновлением wait queue layout; provider semantics и contract ownership по-прежнему берутся из уже утверждённых S12 design/API docs.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась по существу: Issue `#430` закрывает frontend visibility wave и не добавляет новых продуктовых требований.

## Актуализация по Issue #431 (`run:doc-audit`, 2026-03-15)
- Подготовлен readiness bundle:
  - `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/observability_readiness.md`;
  - обновлены `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/README.md`, `docs/delivery/delivery_plan.md`, `docs/delivery/issue_map.md`.
- Зафиксированы:
  - canonical readiness gate перед `run:qa`: rollout order `migrations -> control-plane -> worker -> agent-runner -> api-gateway -> web-console -> evidence gate`, typed evidence surfaces (`flow_events`, `github_rate_limit_wait_evidence`, `Run.wait_projection`, realtime wait envelopes, runner/worker logs) и rollback notes собраны в одном source-of-truth;
  - candidate runtime фактами подтверждён namespace `codex-k8s-dev-1`: основные deployments готовы, `codex-k8s-migrate`/kaniko/repo-sync jobs завершены, а текущий agent run job активен в том же rollout lineage;
  - в текущем candidate rollout feature gate остаётся default-disabled: `kubectl get deploy ... env` показал пустые `CODEXK8S_GITHUB_RATE_LIMIT_WAIT_ENABLED` и `CODEXK8S_WORKER_GITHUB_RATE_LIMIT_SWEEP_LIMIT`, а defaults в `services/internal/control-plane/internal/app/config.go` и `services/jobs/worker/internal/app/config.go` оставляют live wait-path неактивным без явного owner rollout decision;
  - по последним 120 строкам `control-plane` логов live GitHub rate-limit events не обнаружены, поэтому readiness bundle явно отделяет документированный runtime baseline от ещё не выполненного synthetic/live smoke.
- Проверки:
  - `kubectl config view --minify -o jsonpath='{..namespace}'`
  - `kubectl get deploy,pods,job -n codex-k8s-dev-1 -o wide`
  - `kubectl logs -n codex-k8s-dev-1 deploy/codex-k8s-control-plane --tail=120 | rg 'github rate-limit|wait.entered|wait.resumed|manual_action_required|waiting_backpressure'`
  - `kubectl get deploy -n codex-k8s-dev-1 codex-k8s-control-plane -o jsonpath='{range .spec.template.spec.containers[0].env[*]}{.name}={.value}{"\n"}{end}' | rg '^CODEXK8S_GITHUB_RATE_LIMIT'`
  - `kubectl get deploy -n codex-k8s-dev-1 codex-k8s-worker -o jsonpath='{range .spec.template.spec.containers[0].env[*]}{.name}={.value}{"\n"}{end}' | rg '^CODEXK8S_GITHUB_RATE_LIMIT|^CODEXK8S_WORKER_GITHUB_RATE_LIMIT'`
  - `git diff --check`
- Внешний baseline дополнительно сверен:
  - Context7 `/github/docs` использован для повторной проверки guidance по primary/secondary rate limits, `Retry-After`, `x-ratelimit-reset` и bounded retry discipline, чтобы readiness-пакет не расходился с provider semantics.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась по существу: Issue `#431` синхронизирует readiness evidence и traceability, но не вводит новые канонические требования.

## Актуализация по Issue #500 (`run:dev`, 2026-03-15)
- Реализован platform settings stream для live runtime switches:
  - `services/internal/control-plane/cmd/cli/migrations/20260315120000_day31_system_settings_foundation.sql`;
  - `services/internal/control-plane/internal/domain/{repository/systemsetting,systemsettings,staff/service_system_settings.go}`;
  - `services/internal/control-plane/internal/repository/postgres/systemsetting`;
  - `services/internal/control-plane/internal/transport/grpc/server_staff_system_settings.go`;
  - `services/jobs/worker/internal/domain/systemsettings/service.go`;
  - `services/jobs/worker/internal/repository/postgres/systemsetting`;
  - `libs/go/systemsettings/systemsettings.go`;
  - `services/external/api-gateway/api/server/{api.yaml,asyncapi.yaml}`;
  - `services/external/api-gateway/internal/transport/http/{casters/system_settings.go,staff_handler_system_settings.go}`;
  - `services/staff/web-console/src/pages/configuration/SystemSettingsPage.vue`.
- Зафиксированы:
  - `system_settings` и `system_setting_changes` стали реальным control-plane owned contour с durable versioning, audit trail и seeded catalog entry `github_rate_limit_wait_enabled=false`;
  - `CODEXK8S_GITHUB_RATE_LIMIT_WAIT_ENABLED` удалён из `control-plane`/`worker` app config и production/bootstrap wiring; effective rollout state GitHub rate-limit wait path теперь читается из DB-backed platform setting с hot-reload через PostgreSQL `LISTEN/NOTIFY` и reconnect-safe reload from durable tables;
  - staff/private contract-first surface добавлен end-to-end: новые gRPC/OpenAPI/AsyncAPI контракты, typed API/client codegen, staff admin routes `list/get/update/reset/realtime` и рабочая `System settings` page вместо scaffold;
  - policy для future work синхронизирована в common design guidelines: product/runtime switches, которые должны меняться на лету, больше не вводятся как env-only flags и должны жить в typed platform settings catalog.
- Проверки:
  - `make gen-proto-go`
  - `make gen-openapi`
  - `go test ./services/internal/control-plane/internal/domain/systemsettings ./services/internal/control-plane/internal/repository/postgres/systemsetting ./services/internal/control-plane/internal/domain/staff ./services/internal/control-plane/internal/domain/githubratelimit ./services/internal/control-plane/internal/app`
  - `go test ./services/jobs/worker/internal/domain/systemsettings ./services/jobs/worker/internal/repository/postgres/systemsetting ./services/jobs/worker/internal/domain/worker ./services/jobs/worker/internal/app`
  - `go test ./services/internal/control-plane/internal/transport/grpc ./services/external/api-gateway/internal/transport/http ./services/external/api-gateway/internal/controlplane`
  - `go test ./services/internal/control-plane/internal/... ./services/external/api-gateway/internal/... ./services/jobs/worker/internal/...`
  - `npm --prefix services/staff/web-console run build`
  - `git diff --check`
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` уточнена по существу для `FR-008`: current-state traceability теперь явно закрепляет typed platform settings catalog и DB-backed runtime switches как канонический способ управления live product behavior.
