---
doc_id: TRH-CK8S-S10-0001
type: traceability-history
title: "Sprint S10 Traceability History"
status: in-review
owner_role: KM
created_at: 2026-03-12
updated_at: 2026-03-14
related_issues: [360, 378, 383, 385, 387, 389, 391, 392, 393, 394, 395, 402, 437]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-12-traceability-s10-history"
---

# Sprint S10 Traceability History

## TL;DR
- Этот файл хранит historical delta для Sprint S10.
- Текущая master-карта связей остаётся в `docs/delivery/issue_map.md`.
- Текущее покрытие FR/NFR остаётся в `docs/delivery/requirements_traceability.md`.

## Актуализация по Issue #360 (`run:intake`, 2026-03-12)
- Intake зафиксировал built-in MCP user interactions как отдельную product initiative поверх существующего built-in server `codex_k8s`.
- В качестве baseline зафиксированы:
  - MVP tools `user.notify` и `user.decision.request`;
  - channel-neutral interaction-domain;
  - раздельные semantics для approval flow и user interaction flow;
  - wait-state только для response-required сценариев;
  - Telegram как отдельный последовательный follow-up stream.
- Создана continuity issue `#378` для stage `run:vision`.
- Root FR/NFR matrix не менялась: intake stage не обновляет канонический requirements baseline, а фиксирует problem/scope/handover для нового delivery stream.

## Актуализация по Issue #378 (`run:vision`, 2026-03-12)
- Подготовлен vision package:
  - `docs/delivery/epics/s10/epic-s10-day2-mcp-user-interactions-vision.md`.
- Зафиксированы:
  - mission и north star для built-in MCP user interactions как отдельной channel-neutral capability платформы;
  - persona outcomes для owner/product lead, end user и platform operator;
  - KPI/guardrails для actionable notifications, decision turnaround, fallback-to-comments, separation from approval flow и correlation correctness;
  - явное разделение core MVP и deferred streams: Telegram/adapters, voice/STT, richer threads и advanced delivery policies не блокируют core baseline.
- Для continuity создана follow-up issue `#383` (`run:prd`) без trigger-лейбла.
- Попытка использовать Context7 для GitHub CLI manual завершилась ошибкой `Monthly quota exceeded`; неинтерактивный issue/PR flow дополнительно сверен локально по `gh issue create --help`, `gh pr create --help`, `gh pr edit --help`.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась, потому что vision stage уточняет mission, KPI и scope boundaries, но не меняет канонический requirements baseline.

## Актуализация по Issue #383 (`run:prd`, 2026-03-12)
- Подготовлен PRD package:
  - `docs/delivery/epics/s10/epic-s10-day3-mcp-user-interactions-prd.md`;
  - `docs/delivery/epics/s10/prd-s10-day3-mcp-user-interactions.md`.
- Зафиксированы:
  - user stories, FR/AC/NFR и wave priorities для `user.notify`, `user.decision.request`, typed response semantics и adapter-neutral contract;
  - explicit edge cases для stale/duplicate/invalid responses, fallback-to-comments и separation from approval flow;
  - handover decisions, которые нельзя потерять на `run:arch`: built-in `codex_k8s`, non-blocking `user.notify`, wait-state только для `user.decision.request`, platform-owned audit/correlation/retry semantics и deferred scope для Telegram/adapters.
- Для continuity создана follow-up issue `#385` (`run:arch`) без trigger-лейбла.
- Попытка использовать Context7 для GitHub CLI manual снова завершилась ошибкой `Monthly quota exceeded`; для non-interactive GitHub flow использованы локальные `gh issue create --help`, `gh pr create --help`, `gh pr edit --help`.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась по существу: PRD stage уточняет product contract и delivery evidence, а в root-матрице синхронизирована только связь по issue/traceability governance.

## Актуализация по Issue #385 (`run:arch`, 2026-03-12)
- Подготовлен architecture package:
  - `docs/delivery/epics/s10/epic-s10-day4-mcp-user-interactions-arch.md`;
  - `docs/architecture/initiatives/s10_mcp_user_interactions/README.md`;
  - `docs/architecture/initiatives/s10_mcp_user_interactions/architecture.md`;
  - `docs/architecture/initiatives/s10_mcp_user_interactions/c4_context.md`;
  - `docs/architecture/initiatives/s10_mcp_user_interactions/c4_container.md`;
  - `docs/architecture/adr/ADR-0012-built-in-mcp-user-interactions-control-plane-owned-lifecycle.md`;
  - `docs/architecture/alternatives/ALT-0004-built-in-mcp-user-interactions-lifecycle-boundaries.md`.
- Зафиксированы:
  - ownership split между `control-plane`, `worker`, `api-gateway` и future adapters;
  - отдельный interaction-domain без reuse approval-specific semantics как source-of-truth;
  - lifecycle `tool call -> dispatch -> callback -> resume` с platform-owned retry/idempotency/expiry/audit expectations.
- Для continuity создана follow-up issue `#387` (`run:design`) без trigger-лейбла.
- Попытка использовать Context7 для Mermaid/C4 documentation завершилась ошибкой `Monthly quota exceeded`; для пакета использованы существующие Mermaid/C4 conventions репозитория.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась по существу: architecture stage фиксирует service boundaries и handover в design, а не меняет канонический requirements baseline.

## Актуализация по Issue #387 (`run:design`, 2026-03-12)
- Подготовлен design package:
  - `docs/delivery/epics/s10/epic-s10-day5-mcp-user-interactions-design.md`;
  - `docs/architecture/initiatives/s10_mcp_user_interactions/design_doc.md`;
  - `docs/architecture/initiatives/s10_mcp_user_interactions/api_contract.md`;
  - `docs/architecture/initiatives/s10_mcp_user_interactions/data_model.md`;
  - `docs/architecture/initiatives/s10_mcp_user_interactions/migrations_policy.md`.
- Зафиксированы:
  - typed contracts для `user.notify`, `user.decision.request`, outbound adapter envelope и inbound callback family;
  - отдельный persisted interaction-domain: aggregate, delivery attempts, callback evidence, response records;
  - wait-state taxonomy с сохранением coarse runtime state `waiting_mcp`, но с отдельным `wait_reason=interaction_response` и typed wait linkage;
  - resume contract через deterministic `interaction_resume_payload`, использующий existing `agent_sessions` snapshot path без reuse approval tables.
- Для continuity создана follow-up issue `#389` (`run:plan`) без trigger-лейбла.
- Попытка использовать Context7 для `kin-openapi` и `goose` завершилась ошибкой `Monthly quota exceeded`; новые внешние зависимости на этапе `run:design` не добавлялись.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась по существу: design stage конкретизирует API/data/runtime contracts и migration policy, но не меняет канонический product baseline.

## Актуализация по Issue #389 (`run:plan`, 2026-03-13)
- Подготовлен plan package:
  - `docs/delivery/epics/s10/epic-s10-day6-mcp-user-interactions-plan.md`.
- Зафиксированы:
  - execution waves `#391..#395` для `control-plane` foundation, worker dispatch/retry/expiry, contract-first callback ingress, deterministic resume path в `agent-runner` и observability/evidence gate;
  - sequencing `#391 -> #392 -> #393/#394 -> #395` с сохранением rollout order `migrations -> control-plane -> worker -> api-gateway` и отдельным resume gate;
  - DoR/DoD, blockers/risks/owner decisions и запрет на auto-trigger labels для implementation issues;
  - channel-specific adapters, Telegram, reminders и voice/STT оставлены вне core Sprint S10 execution package.
- Для continuity созданы follow-up issues `#391`, `#392`, `#393`, `#394`, `#395` без trigger-лейблов.
- Попытка использовать Context7 для GitHub CLI manual завершилась ошибкой `Monthly quota exceeded`; для non-interactive GitHub flow использованы локальные `gh issue create --help`, `gh pr create --help`, `gh pr edit --help`.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась по существу: plan stage фиксирует execution backlog, quality gates и handover в `run:dev`, а не меняет канонический requirements baseline.

## Актуализация по Issue #391 (`run:dev`, 2026-03-13)
- Реализован foundation package для stream `S10-E01` в `services/internal/control-plane`:
  - миграция `20260313120000_day29_mcp_user_interactions_foundation.sql` добавила таблицы `interaction_requests`, `interaction_delivery_attempts`, `interaction_callback_events`, `interaction_response_records`, а также typed wait columns в `agent_runs` с backfill из legacy wait-state данных;
  - доменные typed models/repositories добавлены для interaction aggregate, delivery attempts, callback evidence и resume payload;
  - built-in MCP orchestration получила foundation для `user.notify` и `user.decision.request`, callback classification/resume helpers и audit events;
  - staff run SQL projections переведены на `agent_runs.wait_reason` с fallback к coarse `agent_sessions.wait_state`, чтобы сохранить совместимость текущих wait filters до следующих волн.
- Rollout boundary сохранён:
  - новые tools зарегистрированы в MCP catalog/transport, но не добавлены в allow-list policy, поэтому остаются скрытыми до delivery waves `#392`, `#393`, `#394`;
  - callback ingress / OpenAPI / gRPC transport на этой волне не открывались.
- Выполнены проверки:
  - `go test ./services/internal/control-plane/...`
  - `git diff --check`
- Для residual `dupl-go` blockers вне scope создан follow-up issue `#402`.
- Попытка использовать Context7 для актуальной документации/verification завершилась ошибкой `Monthly quota exceeded`; новых внешних зависимостей в issue `#391` не добавлялось.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась по существу: issue `#391` реализует approved foundation wave из execution package, не меняя продуктовый baseline.

## Актуализация по Issue #392 (`run:dev`, 2026-03-13)
- Реализован worker lifecycle package для stream `S10-E02` в `services/jobs/worker` и `services/internal/control-plane`:
  - `worker` получил polling loop для expiry и outbound dispatch, configurable limits/timeouts/backoff/max-attempts, flow events `interaction.dispatch.attempted` / `interaction.dispatch.retry_scheduled` и transport client к новым `control-plane` RPC;
  - `control-plane` получил worker-facing domain/repository API для claim/complete/expire interaction delivery attempts, включая reclaim stuck pending attempts, ledger update по `delivery_id`, terminal outcome classification и wait-context cleanup/resume payload publication только после terminal status;
  - gRPC contract `proto/codexk8s/controlplane/v1/controlplane.proto` расширен RPC `ClaimNextInteractionDispatch`, `CompleteInteractionDispatch`, `ExpireNextInteraction`, после чего regenerated Go stubs синхронизированы для `control-plane`, `worker` и `agent-runner`;
  - текущий dispatch adapter оставлен `noop`, потому что channel-specific delivery adapters не входят в core scope issue `#392`; user-facing MCP tools по-прежнему скрыты rollout gate и не открываются до волн `#393/#394`.
- Rollout boundary сохранён:
  - callback ingress / OpenAPI / typed HTTP DTO остаются в scope issue `#393`;
  - deterministic runner resume path и prompt handoff остаются в scope issue `#394`;
  - terminal outcome scheduling выполняется только после terminal request outcome, без переноса callback semantics в `worker`.
- Выполнены проверки:
  - `go test ./services/internal/control-plane/... ./services/jobs/worker/... ./services/jobs/agent-runner/...`
  - `git diff --check`
- Для verification использован Context7:
  - `/grpc/grpc-go` для проверки актуального RPC/metadata usage pattern;
  - `/jackc/pgx` для проверки transaction/query pattern при claim/complete/expiry update path.
- Новых внешних зависимостей в issue `#392` не добавлялось.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась по существу: issue `#392` реализует approved worker lifecycle wave из execution package, не меняя продуктовый baseline.

## Актуализация по Issue #393 (`run:dev`, 2026-03-13)
- Реализован transport package для stream `S10-E03` в `services/external/api-gateway` и `services/internal/control-plane`:
  - OpenAPI source-of-truth расширен endpoint `POST /api/v1/mcp/interactions/callback`, typed DTO `InteractionCallbackEnvelope` / `InteractionCallbackOutcome` и regenerated codegen artifacts для Go/TS;
  - `api-gateway` получил отдельный thin-edge handler, typed HTTP models/casters и route-level per-interaction rate limiting, где идентификатор лимита извлекается из `interaction_id`, а callback bearer token прокидывается в gRPC metadata без локального interaction state;
  - `control-plane` gRPC contract расширен RPC `SubmitInteractionCallback`, transport auth теперь проверяет `token subject == mcp-interaction-callback:<interaction_id>`, а успешная domain classification `accepted` отображается наружу как transport classification `applied`;
  - approval callback family `/api/v1/mcp/approver|executor/callback` сохранён как additive coexistence и не переиспользует interaction DTO/handlers.
- Rollout boundary сохранён:
  - callback endpoint остаётся thin-edge ingress и не реализует wait-state transitions, retry semantics или replay classification локально;
  - channel-specific adapter UX и deterministic resume path по-прежнему находятся вне scope issue `#393` и остаются в waves `#394/#395`.
- Выполнены проверки:
  - `make gen-proto-go`
  - `make gen-openapi`
  - `go test ./services/internal/control-plane/... ./services/external/api-gateway/... ./services/jobs/worker/... ./services/jobs/agent-runner/...`
  - `make lint-go`
  - `make dupl-go`
  - `git diff --check`
- Для verification использован Context7:
  - `/labstack/echo` для проверки актуального `RateLimiterWithConfig`, `IdentifierExtractor` и `DenyHandler` паттерна в Echo v5.
- Новых внешних зависимостей в issue `#393` не добавлялось.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась по существу: issue `#393` реализует approved transport wave из execution package, не меняя продуктовый baseline.

## Актуализация по Issue #394 (`run:dev`, 2026-03-13)
- Реализован runner handoff package для stream `S10-E04` в `services/internal/control-plane`, `services/jobs/worker`, `libs/go/k8s/joblauncher` и `services/jobs/agent-runner`:
  - `control-plane` при scheduling resume-run теперь сохраняет terminal `interaction_resume_payload` в persisted `agent_runs.run_payload`, сохраняя existing session snapshot path в `agent_sessions` как единственный resume-source для `codex exec resume`;
  - `worker` извлекает этот typed payload из run payload и прокидывает его в env нового run job без локального recompute interaction semantics;
  - `agent-runner` валидирует machine-readable payload, требует наличие restored Codex session/session_id и добавляет typed JSON block в начало prompt перед `codex exec resume`, чтобы модель получала deterministic terminal outcome без повторного adapter lookup;
  - `joblauncher` синхронно пробрасывает новый env contract `CODEXK8S_INTERACTION_RESUME_PAYLOAD`, а `services/jobs/agent-runner/README.md` актуализирован под новый resume path.
- Rollout boundary сохранён:
  - interaction source-of-truth остаётся у `control-plane`; `worker` и `agent-runner` только потребляют persisted payload;
  - channel-specific adapters и broader pause/resume engine refactor по-прежнему остаются вне scope issue `#394`.
- Выполнены проверки:
  - `go test ./services/internal/control-plane/internal/domain/mcp ./services/jobs/worker/internal/domain/worker ./services/jobs/agent-runner/internal/runner ./libs/go/k8s/joblauncher`
  - `go test ./services/internal/control-plane/... ./services/jobs/worker/... ./services/jobs/agent-runner/...`
  - `git diff --check`
  - runtime diagnostics:
    - `kubectl get pods,job -n codex-k8s-dev-1 -o wide`
    - `kubectl get events -n codex-k8s-dev-1 --sort-by=.lastTimestamp | tail -n 20`
    - `kubectl logs deploy/codex-k8s-control-plane -n codex-k8s-dev-1 --tail=40`
    - `kubectl logs deploy/codex-k8s-worker -n codex-k8s-dev-1 --tail=40`
- Runtime evidence:
  - candidate namespace `codex-k8s-dev-1` содержит running pods `codex-k8s-control-plane`, `codex-k8s-worker`, `codex-k8s-web-console`, `codex-k8s`, текущий run job и `postgres-0`;
  - recent events/logs показывают ожидаемые hot-reload restarts control-plane/worker и кратковременные probe/dial failures в момент локальной пересборки, но без новых resume-specific ошибок.
- Для verification использован Context7:
  - `/protocolbuffers/protobuf` для проверки additive contract discipline при расширении typed payload-handshake.
- Новых внешних зависимостей в issue `#394` не добавлялось.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась по существу: issue `#394` реализует approved deterministic resume wave из execution package, не меняя продуктовый baseline.

## Актуализация по Issue #437 (`run:dev`, 2026-03-14)
- Реализован hardening follow-up для resume handoff поверх stream `S10-E04`:
  - `worker` и `joblauncher` больше не используют `CODEXK8S_INTERACTION_RESUME_PAYLOAD` и не записывают interaction outcome в Pod env/spec;
  - `control-plane` получил run-scoped gRPC lookup `GetRunInteractionResumePayload`, а `agent-runner` забирает persisted payload уже после старта pod через bearer-аутентифицированный runtime call;
  - callback classification и resume scheduling теперь явно ограничены size-контрактом: `response.free_text <= 8192` UTF-8 bytes, serialized `interaction_resume_payload <= 12288` bytes, overflow классифицируется как `invalid` без постановки resume-run;
  - docs и contract-first артефакты синхронизированы для нового secure carrier path и size guardrails.
- Rollout boundary сохранён:
  - source-of-truth по interaction outcome остаётся в `control-plane`, а `agent-runner` получает только persisted run-scoped projection;
  - новые внешние зависимости не добавлялись; verification для proto contract discipline выполнен через Context7 `/protocolbuffers/protobuf`.
- Выполнены проверки:
  - `make gen-openapi`
  - `go test ./services/internal/control-plane/... ./services/jobs/agent-runner/... ./services/jobs/worker/... ./libs/go/k8s/joblauncher`
  - `make lint-go`
  - `git diff --check`
  - runtime diagnostics:
    - `kubectl config view --minify -o jsonpath='{..namespace}'`
    - `kubectl -n codex-k8s-dev-1 get pods,deploy,job -o wide`
    - `kubectl -n codex-k8s-dev-1 logs deploy/codex-k8s-control-plane --tail=60`
    - `kubectl -n codex-k8s-dev-1 logs deploy/codex-k8s-worker --tail=60`
- Runtime evidence:
  - candidate namespace `codex-k8s-dev-1` содержит ready deployments `codex-k8s-control-plane`, `codex-k8s-worker`, `codex-k8s`, `codex-k8s-web-console`, running run job и healthy `postgres-0`;
  - в логах `control-plane` зафиксирован ожидаемый transient compile failure во время hot-reload до regeneration нового proto-stub, после чего сервис перезапустился штатно;
  - в логах `worker` зафиксирован кратковременный `dial tcp ...:9090: connect: connection refused` во время restart окна `control-plane`, после чего worker восстановился без новых resume-specific ошибок.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась по существу: issue `#437` ужесточает transport/runtime handoff и payload limits внутри approved Sprint S10 scope, не меняя продуктовый baseline.
