---
doc_id: TRH-CK8S-S7-0001
type: traceability-history
title: "Sprint S7 Traceability History"
status: in-review
owner_role: KM
created_at: 2026-03-12
updated_at: 2026-03-12
related_issues: [210, 212, 218, 220, 222, 238, 241, 243, 244, 245, 246, 247, 248, 249, 250, 251, 252, 255, 256, 257, 258, 259, 260, 274, 327]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-12-traceability-s7-history"
---

# Sprint S7 Traceability History

## TL;DR
- Этот файл хранит historical delta для Sprint S7.
- Текущая master-карта связей остаётся в `docs/delivery/issue_map.md`.
- Текущее покрытие FR/NFR остаётся в `docs/delivery/requirements_traceability.md`.

## Актуализация по Issue #212 (`run:intake`, 2026-02-27)
- Для FR-026/FR-028/FR-033/FR-036/FR-040/FR-043/FR-045 и NFR-010/NFR-013/NFR-017/NFR-018 добавлен Sprint S7 intake traceability пакет:
  `docs/delivery/sprints/s7/sprint_s7_mvp_readiness_gap_closure.md`,
  `docs/delivery/epics/s7/epic_s7.md`,
  `docs/delivery/epics/s7/epic-s7-day1-mvp-readiness-intake.md`,
  `docs/delivery/development_process_requirements.md`,
  `docs/product/labels_and_trigger_policy.md`,
  `docs/product/stage_process_model.md`.
- Intake зафиксировал фактические MVP gaps:
  - `comingSoon`/scaffold контур в staff UI (`navigation.ts` + профильные TODO-страницы);
  - S6 dependency-chain: `#199/#201` закрыты, release closeout выполнен в Issue `#262`, активный continuity-блокер перенесён в `#263` (`run:postdeploy`);
  - отсутствие подтверждённого run-evidence для `run:doc-audit` в текущем delivery-цикле.
- Для всех открытых owner-замечаний PR #213 выставлен статус `fix_required`; замечания сгруппированы по приоритету `behavior/data -> quality/style`.
- В backlog S7 добавлены 18 candidate execution-эпиков (`S7-E01..S7-E18`) с owner-aligned handover в `run:vision`:
  rebase/mainline hygiene, UI cleanup (navigation/sections/filter), agents de-scope + repo-only prompt policy для MVP, runs/deploy UX, `mode:discussion` reliability, late-stage `run:<stage>:revise` coverage, QA DNS acceptance-policy, `run:intake:revise` status consistency, `run:self-improve` session reliability, финальный readiness gate.
- Для стандартизации качества backlog зафиксировано требование PMO из Issue `#210`:
  формулировка задач в формате user story и обязательный блок edge cases для QA-ready acceptance.
- Для процессного governance добавлен единый стандарт:
  - заголовков и body для Issue/PR по stage/role;
  - информационной архитектуры проектной документации (каталоги `product/architecture/delivery/ops/templates`);
  - ролевой матрицы обязательных шаблонов документации.

## Актуализация по Issue #218 (`run:vision`, 2026-02-27)
- Для FR-026/FR-028/FR-033/FR-045/FR-052/FR-053/FR-054 и NFR-010/NFR-018 добавлен vision traceability пакет Sprint S7:
  `docs/delivery/epics/s7/epic-s7-day2-mvp-readiness-vision.md`,
  `docs/delivery/epics/s7/epic_s7.md`,
  `docs/delivery/sprints/s7/sprint_s7_mvp_readiness_gap_closure.md`,
  `docs/delivery/delivery_plan.md`,
  `docs/delivery/issue_map.md`.
- Vision-stage формализовал measurable KPI для всех execution-потоков `S7-E01..S7-E18` и зафиксировал baseline по каждому потоку:
  `user story + acceptance criteria + edge cases + expected evidence`.
- В vision baseline добавлена owner policy для MVP: custom agents/prompt lifecycle вынесены в post-MVP, prompt templates обслуживаются по repo workflow.
- Введено обязательное governance-правило decomposition parity перед `run:dev`:
  `approved_execution_epics_count == created_run_dev_issues_count` (coverage ratio = `1.0`).
- Для stage continuity создана follow-up issue `#220` (`run:prd`) без trigger-лейбла; в issue передан обязательный шаблон создания следующей stage-задачи (`run:arch`).
- Scope этапа сохранён policy-safe: markdown-only изменения без модификации code/runtime артефактов.

## Актуализация по Issue #220 (`run:prd`, 2026-02-27)
- Для FR-026/FR-028/FR-033/FR-045/FR-052/FR-053/FR-054 и NFR-010/NFR-018 добавлен PRD traceability пакет Sprint S7:
  `docs/delivery/epics/s7/epic-s7-day3-mvp-readiness-prd.md`,
  `docs/delivery/epics/s7/prd-s7-day3-mvp-readiness-gap-closure.md`,
  `docs/delivery/epics/s7/epic_s7.md`,
  `docs/delivery/sprints/s7/sprint_s7_mvp_readiness_gap_closure.md`,
  `docs/delivery/delivery_plan.md`,
  `docs/delivery/issue_map.md`.
- PRD-stage формализовал stream-level execution contract для `S7-E01..S7-E18`:
  `user story + FR + AC + NFR + edge cases + expected evidence + dependencies`.
- Зафиксированы deterministic sequencing и dependency graph для перехода `run:prd -> run:arch -> run:design -> run:plan`.
- В PRD явным контуром зафиксирован `repo-only` policy для prompt templates на MVP и de-scope custom agents/prompt lifecycle.
- Подтверждено governance-правило decomposition parity перед `run:dev`:
  `approved_execution_epics_count == created_run_dev_issues_count` (coverage ratio = `1.0`, блокировка при mismatch).
- Для stage continuity создана follow-up issue `#222` (`run:arch`) без trigger-лейбла; в handover переданы PRD-пакет, sequencing-ограничения и parity-gate правила.
- Scope этапа сохранён policy-safe: markdown-only изменения без модификации code/runtime артефактов.

## Актуализация по Issue #222 (`run:arch`, 2026-03-02)
- Для FR-026/FR-028/FR-033/FR-053/FR-054 и NFR-010/NFR-018 добавлен architecture traceability пакет Sprint S7:
  `docs/delivery/epics/s7/epic-s7-day4-mvp-readiness-arch.md`,
  `docs/architecture/initiatives/s7_mvp_readiness_gap_closure/architecture.md`,
  `docs/architecture/initiatives/s7_mvp_readiness_gap_closure/c4_context.md`,
  `docs/architecture/initiatives/s7_mvp_readiness_gap_closure/c4_container.md`,
  `docs/architecture/adr/ADR-0010-s7-mvp-readiness-stream-boundaries-and-parity-gate.md`,
  `docs/architecture/alternatives/ALT-0002-s7-mvp-readiness-stream-architecture.md`,
  `docs/delivery/epics/s7/epic_s7.md`,
  `docs/delivery/sprints/s7/sprint_s7_mvp_readiness_gap_closure.md`,
  `docs/delivery/delivery_plan.md`,
  `docs/delivery/issue_map.md`.
- На architecture-stage зафиксированы:
  - ownership matrix и сервисные границы по `S7-E01..S7-E18`;
  - deterministic wave-sequencing для перехода `run:arch -> run:design -> run:plan`;
  - parity-gate перед `run:dev`: `approved_execution_epics_count == created_run_dev_issues_count`.
- Для stage continuity создана follow-up issue `#238` (`run:design`) без trigger-лейбла с обязательным handover на подготовку `design_doc/api_contract/data_model/migrations_policy`.
- Через Context7 подтверждён baseline для инструментов stage-handover и C4-артефактов:
  `/websites/cli_github_manual` (актуальный `gh issue/pr` синтаксис) и `/mermaid-js/mermaid` (валидный C4 синтаксис).
- Scope этапа сохранён policy-safe: markdown-only изменения без модификации code/runtime артефактов.

## Актуализация по Issue #238 (`run:design`, 2026-03-02)
- Для FR-026/FR-028/FR-033/FR-053/FR-054 и NFR-010/NFR-018 добавлен design traceability пакет Sprint S7:
  `docs/delivery/epics/s7/epic-s7-day5-mvp-readiness-design.md`,
  `docs/architecture/initiatives/s7_mvp_readiness_gap_closure/design_doc.md`,
  `docs/architecture/initiatives/s7_mvp_readiness_gap_closure/api_contract.md`,
  `docs/architecture/initiatives/s7_mvp_readiness_gap_closure/data_model.md`,
  `docs/architecture/initiatives/s7_mvp_readiness_gap_closure/migrations_policy.md`,
  `docs/delivery/epics/s7/epic_s7.md`,
  `docs/delivery/sprints/s7/sprint_s7_mvp_readiness_gap_closure.md`,
  `docs/delivery/delivery_plan.md`,
  `docs/delivery/issue_map.md`.
- На design-stage зафиксированы typed contract decisions для потоков:
  `S7-E06`, `S7-E07`, `S7-E09`, `S7-E10`, `S7-E13`, `S7-E16`, `S7-E17`.
- Зафиксированы persisted-state изменения и migration/rollback политика:
  `runtime_deploy_tasks`, `agent_runs`, `agent_sessions` (+ flow-events payload hardening).
- Через Context7 подтверждён dependency baseline и актуальная документация:
  `/getkin/kin-openapi` (OpenAPI request/response validation path),
  `/microsoft/monaco-editor` (DiffEditor API `createDiffEditor`/`setModel`).
- Новые внешние зависимости не добавлялись; каталог зависимостей не требует обновления.
- Для stage continuity создана follow-up issue `#241` (`run:plan`) без trigger-лейбла.
- Scope этапа сохранён policy-safe: markdown-only изменения без модификации code/runtime артефактов.

## Актуализация по Issue #241 (`run:plan`, 2026-03-02)
- Для FR-026/FR-028/FR-033/FR-053/FR-054 и NFR-010/NFR-018 добавлен plan traceability пакет Sprint S7:
  `docs/delivery/epics/s7/epic-s7-day6-mvp-readiness-plan.md`,
  `docs/delivery/epics/s7/epic_s7.md`,
  `docs/delivery/sprints/s7/sprint_s7_mvp_readiness_gap_closure.md`,
  `docs/delivery/delivery_plan.md`,
  `docs/delivery/issue_map.md`.
- На plan-stage сформирован execution package для `S7-E01..S7-E18` с wave-sequencing и quality-gates перед `run:dev`.
- По owner-уточнению в Issue `#241` создана отдельная implementation issue на каждый поток:
  `#243..#260` (один issue на `S7-E01..S7-E18`), без trigger-лейблов `run:*`.
- Подтверждено decomposition parity-правило перед входом в `run:dev`:
  `approved_execution_epics_count == created_run_dev_issues_count` (`18 == 18`).
- Через Context7 (`/websites/cli_github_manual`) подтверждён актуальный неинтерактивный синтаксис `gh issue create` / `gh pr create` / `gh pr edit`; новые внешние зависимости не добавлялись.
- Scope этапа сохранён policy-safe: markdown-only изменения без модификации code/runtime артефактов.

## Актуализация по Issue #243 (`run:dev`, 2026-03-02)
- Для FR-026/FR-028/FR-033 и NFR-010/NFR-018 реализован foundation stream `S7-E01`:
  зафиксирован единый deterministic rebase/mainline процесс для revise-итераций в `run:dev`.
- Обновлён process source-of-truth:
  `docs/delivery/development_process_requirements.md` закрепил обязательный порядок
  `git fetch -> git rebase origin/main -> conflict-marker check -> checks -> git push --force-with-lease`,
  запрет `git merge origin/main` для revise-веток и обязательный PR rebase-checklist.
- Актуализирована traceability-матрица Sprint S7:
  `docs/delivery/issue_map.md` (выделен отдельный статус issue `#243`, остаток backlog перенесён в диапазон `#245..#260` после закрытия `#244`).
- Через Context7 (`/websites/git-scm`) подтверждены актуальные команды `git rebase --continue|--abort`
  и безопасный push-path `git push --force-with-lease` для rebased PR-веток.
- Новые внешние зависимости не добавлялись; изменения ограничены markdown/process governance контуром.

## Актуализация по Issue #244 (`run:dev`, 2026-03-05)
- Для FR-026/FR-028/FR-033 и NFR-010/NFR-018 реализован stream `S7-E02`:
  из staff sidebar удалены non-MVP navigation entries (`governance`, `admin`, `configuration/docs`, `configuration/mcp-tools`, `configuration/agents`, `configuration/config-entries`).
- В `services/staff/web-console/src/router/routes.ts` удалены связанные non-MVP маршруты и добавлен fallback redirect на `projects` для stale deep-links,
  чтобы после cleanup не возникало broken transitions.
- Удалён связанный dead code:
  страницы `pages/governance/*`, `pages/admin/*`, `pages/configuration/{DocsKnowledgePage,McpToolsPage}.vue`,
  UI-контур `config-entries` (страница + feature-слой) и platform-tokens scaffold в `System settings`.
- Актуализирована Sprint S7 traceability:
  `docs/delivery/issue_map.md`, `docs/delivery/sprints/s7/sprint_s7_mvp_readiness_gap_closure.md`, `docs/delivery/epics/s7/epic_s7.md`.
- По owner request оформлен follow-up issue `#274` (`S7-E19`) на backend cleanup Agents/Configs/Secrets,
  синхронизированы sprint/epic/plan документы.
- Проверки по scope:
  `npm --prefix services/staff/web-console run build`, route inventory search (`rg -n` по удалённым route names),
  smoke-check MVP навигации (`projects`, `project-repositories`, `project-members`, `runs`, `runtime-deploy/tasks`,
  `wait-queue`, `approvals`, `system-settings`, `users`) и отсутствие broken links на удалённых
  `/runtime-deploy/images` и `/running-jobs`.

## Актуализация по Issue #245 (`run:dev`, 2026-03-05)
- Для FR-026/FR-028/FR-033 и NFR-010/NFR-018 реализован stream `S7-E03`:
  удалён глобальный filter-entry в app shell вместе с зависимым UI summary/reset контуром.
- В `services/staff/web-console/src/features/ui-context/store.ts` удалено глобальное состояние `env/namespace`
  и связанный cookie-persistence path; сохранён только selected project context, нужный для MVP-навигации.
- В `services/staff/web-console/src/pages/operations/RuntimeDeployTasksPage.vue`
  загрузка списка отвязана от `uiContext.env`, чтобы глобальный фильтр больше не влиял на list/load поведение.
- Удалён связанный неиспользуемый компонент `services/staff/web-console/src/shared/ui/AdminClusterContextBar.vue`,
  очищены i18n-ключи глобального фильтра в `services/staff/web-console/src/i18n/messages/{en,ru}.ts`.
- Актуализирована Sprint S7 traceability:
  `docs/delivery/issue_map.md`, `docs/delivery/sprints/s7/sprint_s7_mvp_readiness_gap_closure.md`, `docs/delivery/epics/s7/epic_s7.md`;
  remaining backlog нормализован как `#246..#260` + post-plan `#274`.
- Проверки по scope:
  cleanup-поиск `rg -n` по `Global filter`/`uiContext.env`/`uiContext.namespace`,
  `npm --prefix services/staff/web-console run build`.

## Актуализация по Issue #246 (`run:dev`, 2026-03-09)
- Для FR-E04-1/FR-E04-2 и NFR-E04-1 stream `S7-E04` финализирован без нового redirect-кода:
  owner-review подтвердил, что после удаления UI-контура `runtime-deploy/images`
  отдельный redirect для `/runtime-deploy/images*` не нужен.
- В `services/staff/web-console/src/router/routes.ts` сохраняется только общий fallback
  `/:pathMatch(.*)* -> projects`; stale URL `/runtime-deploy/images*` попадает в него
  после cleanup `#244`, поэтому удалённый раздел не возвращается в MVP navigation.
- Нормализована Sprint S7 traceability:
  `docs/delivery/issue_map.md`, `docs/delivery/sprints/s7/sprint_s7_mvp_readiness_gap_closure.md`, `docs/delivery/epics/s7/epic_s7.md`;
  remaining backlog нормализован как `#247..#260` + post-plan `#274`.
- Через Context7 (`/vuejs/router`) подтверждён актуальный синтаксис Vue Router 4 для catch-all path
  `:pathMatch(.*)*`, которым закрываются удалённые пути без отдельного route record.
- Проверки по scope:
  `rg -n "pathMatch" services/staff/web-console/src/router/routes.ts`,
  `npm --prefix services/staff/web-console run build`.

## Актуализация по Issues #247 / #248 / #249 (`run:dev`, 2026-03-09)
- Для FR-009/FR-029/FR-030/FR-035/FR-037 и NFR-010 выполнен combined cleanup/doco-sync pass по потокам `S7-E05`, `S7-E06`, `S7-E07`.
- Подтверждено фактическое MVP-состояние после ранее выполненных `#244` и `#274`:
  - UI-раздел `Agents` больше не входит в MVP navigation;
  - runtime mode/locale agent settings не редактируются через staff UI/API;
  - prompt templates работают только по repo-seed policy без selector `repo|db`.
- В коде удалены остаточные stale references:
  - agent-related i18n scaffold keys;
  - тест `resolvePathUnescaped`, привязанный к удаленному `/staff/prompt-templates/*` path;
  - мертвые proto messages и HTTP DTO/caster модели старого `Agents/PromptTemplates` staff API.
- В `worker` добавлен unit-test, фиксирующий инвариант:
  - `PromptTemplateSource == repo_seed`;
  - `PromptTemplateLocale` берется из platform default worker config.
- Обновлены source-of-truth документы:
  `docs/product/agents_operating_model.md`,
  `docs/product/requirements_machine_driven.md`,
  `docs/product/constraints.md`,
  `docs/product/brief.md`,
  `docs/architecture/prompt_templates_policy.md`,
  `docs/architecture/api_contract.md`,
  `docs/architecture/data_model.md`,
  а также Sprint S7 traceability docs.
- Combined cleanup `#247/#248/#249` вместе с backend cleanup `#274` также поглотили standalone streams `S7-E08/#250` и `S7-E15/#257`: отдельные `run:dev` для них больше не требуются.
- Remaining standalone backlog Sprint S7 после этой актуализации нормализован как `#251..#256`, `#258..#260`.

## Актуализация по Issue #251 (`run:dev`, 2026-03-10)
- Для FR-012/FR-040 и NFR-010 реализован stream `S7-E09`:
  в `services/staff/web-console/src/pages/RunsPage.vue` удалена колонка `run type`,
  чтобы список запусков показывал только релевантные поля операционной диагностики.
- В `services/staff/web-console/src/pages/RunDetailsPage.vue`
  delete namespace action больше не зависит от `job_exists`;
  UI показывает action при наличии известного namespace и переиспользует существующий typed endpoint
  `DELETE /api/v1/staff/runs/{run_id}/namespace`, который уже идемпотентно обрабатывает повторный delete (`already_deleted=true`).
- В `services/staff/web-console/src/pages/operations/WaitQueuePage.vue` и
  `services/staff/web-console/src/i18n/messages/{en,ru}.ts`
  пользовательские подписи нормализованы как `trigger kind` / `вид триггера`,
  чтобы в MVP UI не оставался термин `run type`.
- Актуализирована Sprint S7 traceability:
  `docs/delivery/issue_map.md`, `docs/delivery/sprints/s7/sprint_s7_mvp_readiness_gap_closure.md`, `docs/delivery/epics/s7/epic_s7.md`;
  remaining backlog нормализован как `#252..#256`, `#258..#260`.
- Проверки по scope:
  cleanup-поиск `rg -n "Run type|Тип запуска" services/staff/web-console/src`,
  `npm --prefix services/staff/web-console run build`.

## Актуализация по Issue #252 (`run:dev`, 2026-03-10)
- Для FR-012/FR-038/FR-040 и NFR-010/NFR-018 реализован stream `S7-E10`:
  contract-first расширены `services/external/api-gateway/api/server/api.yaml` и
  `proto/codexk8s/controlplane/v1/controlplane.proto`, добавлены typed HTTP/gRPC actions
  `cancel/stop` и синхронно обновлены backend/frontend codegen артефакты.
- В control-plane расширена persisted-модель `runtime_deploy_tasks`:
  добавлены control/audit поля `cancel_requested_*`, `stop_requested_*`,
  `terminal_status_source`, `terminal_event_seq`, новая миграция
  `services/internal/control-plane/cmd/cli/migrations/20260310110000_day26_runtime_deploy_task_controls.sql`
  и идемпотентный repository path `RequestAction`.
- В доменном слое `runtimedeploy` и staff use-case добавлены guardrails:
  повторный `cancel/stop` возвращает идемпотентный результат,
  `stop` разрешён только для `running` задачи с активным lease,
  операторские actions пишут audit events в `flow_events`,
  а reconcile-loop быстрее прерывает текущий deploy flow после terminal cancel.
- В staff UI на странице деталей deploy task добавлены кнопки `cancel/stop`,
  confirm dialog с optional reason, success/error feedback, отображение новых audit-полей
  и обработка `failed_precondition` для операторского сценария.
- Актуализированы traceability документы Sprint S7:
  `docs/delivery/issue_map.md`,
  `docs/delivery/sprints/s7/sprint_s7_mvp_readiness_gap_closure.md`,
  `docs/delivery/epics/s7/epic_s7.md`;
  remaining backlog нормализован как `#254..#256`, `#258..#260`.
- Проверки по scope:
  `make gen-proto-go`,
  `make gen-openapi`,
  `go test ./services/internal/control-plane/... ./services/external/api-gateway/...`,
  `npm --prefix services/staff/web-console run build`.

## Актуализация по Issue #255 (`run:dev`, 2026-03-11)
- Для FR-026/FR-028/FR-030/FR-033/FR-052 и NFR-010/NFR-011/NFR-018 реализован stream `S7-E13`:
  revise-loop для недостающих late-stage labels `run:doc-audit|qa|release|postdeploy|ops|self-improve:revise`
  доведён до фактического рабочего состояния в code path.
- В label/runtime source-of-truth добавлены отсутствовавшие revise-элементы:
  typed trigger kinds `doc_audit_revise`, `qa_revise`, `release_revise`, `postdeploy_revise`, `ops_revise`, `self_improve_revise` в `libs/go/domain/webhook`,
  env labels `CODEXK8S_RUN_*_REVISE_LABEL` в `services/internal/control-plane/internal/app/config.go`,
  runtime defaults в `services/internal/control-plane/internal/domain/runtimedeploy/service_defaults.go`,
  а также revise-поля в `TriggerLabels` / next-step label catalog path.
- В control-plane закрыт resolver/runtime gap:
  `services/internal/control-plane/internal/domain/webhook/pull_request_review_resolver.go`
  теперь резолвит PR review `changes_requested` из `run:doc-audit|qa|release|postdeploy|ops|self-improve`
  в детерминированный trigger `run:<stage>:revise`;
  `resolveRunAgentKey` направляет такой run в корректную stage-role;
  `services/internal/control-plane/internal/domain/runstatus/next_step_actions.go`
  строит next-step матрицу для этих revise trigger kinds без потери stage context.
- В `agent-runner` multi-stage revise синхронизирован с policy исполнения:
  `prompt_seed_mapping.go` сохраняет stage mapping для всех новых revise seed'ов,
  `helpers_prompt_docs.go` оставляет full-env prompt docs env = `ai` для `ops_revise`,
  `write_scope_policy.go` сохраняет markdown-only scope для doc-stage revise и restricted scope для `self_improve_revise`.
- Добавлены unit-тесты на ключевые сценарии:
  PR review с label `run:<stage>` создаёт run с trigger `run:<stage>:revise` и корректным агентом для late-stage revise-loop,
  `resolveRunAgentKey` для новых revise trigger kinds,
  mapping next-step stage descriptor,
  runner prompt/docs/write-scope policy.
- Через Context7 (`/caarlos0/env`) перепроверен актуальный baseline для env-tag конфигурации `env`/`envDefault`;
  новые внешние зависимости не добавлялись.

## Актуализация по Issue #256 (`run:dev`, 2026-03-11)
- Для FR-028/FR-033 и NFR-010 реализован stream `S7-E14`:
  `docs/ops/production_runbook.md` закрепил QA policy, по которой новые и изменённые HTTP-ручки
  проверяются через Kubernetes service DNS path (`<service>.<namespace>.svc.cluster.local`),
  а browser/OAuth flow больше не считается единственным acceptance gate.
- Синхронно обновлены QA role templates:
  `docs/templates/test_strategy.md`,
  `docs/templates/test_plan.md`,
  `docs/templates/test_matrix.md`,
  `docs/templates/regression_checklist.md`.
  Теперь они требуют DNS evidence bundle:
  namespace, service FQDN, точную команду, HTTP status, excerpt ответа и `kubectl`-диагностику при fail.
- Full-env verification выполнен в namespace `codex-k8s-dev-1`:
  `kubectl config view --minify -o jsonpath='{..namespace}'`,
  `kubectl get svc -o wide`,
  `getent hosts codex-k8s.codex-k8s-dev-1.svc.cluster.local`,
  `curl -sS ... /healthz` (`200`),
  `curl -sS ... /api/v1/auth/me` (`401`),
  `curl -sS -X POST ... /api/v1/webhooks/github` (`400`).
- Актуализированы traceability документы Sprint S7:
  `docs/delivery/issue_map.md`,
  `docs/delivery/delivery_plan.md`,
  `docs/delivery/sprints/s7/sprint_s7_mvp_readiness_gap_closure.md`,
  `docs/delivery/epics/s7/epic_s7.md`;
  remaining backlog нормализован как `#254`, `#258..#260`.
- Проверки по scope:
  `kubectl config view --minify -o jsonpath='{..namespace}'`,
  `kubectl get svc -o wide`,
  `getent hosts codex-k8s.codex-k8s-dev-1.svc.cluster.local`,
  `curl -sS -o /tmp/codex-health.out -D /tmp/codex-health.headers -w '%{http_code}\n' http://codex-k8s.codex-k8s-dev-1.svc.cluster.local/healthz`,
  `curl -sS -o /tmp/codex-authme.out -D /tmp/codex-authme.headers -w '%{http_code}\n' http://codex-k8s.codex-k8s-dev-1.svc.cluster.local/api/v1/auth/me`,
  `curl -sS -o /tmp/codex-webhook.out -D /tmp/codex-webhook.headers -w '%{http_code}\n' -X POST http://codex-k8s.codex-k8s-dev-1.svc.cluster.local/api/v1/webhooks/github`.

## Актуализация по Issue #258 (`run:dev`, 2026-03-11)
- Для FR-026/FR-028/FR-033 и NFR-010/NFR-018 реализован stream `S7-E16`:
  в `services/internal/control-plane/internal/domain/runstatus`
  нормализована логика выбора и слияния run-status comment state,
  чтобы поздний duplicate/stale update не мог понизить фактически успешный terminal status.
- Добавлен приоритет terminal-state для service-message:
  `succeeded > failed > running/pending`, а при равном status rank
  сохраняется более поздняя lifecycle phase и только затем fallback к `comment_id`.
- Dedupe run-status comments теперь выбирает канонический comment по state precedence,
  а не только по максимальному GitHub `comment_id`,
  что устраняет false-failed при дублирующих terminal updates.
- Добавлено регрессионное unit-покрытие для:
  - merge terminal states (`succeeded` не понижается до `failed`);
  - защиты terminal state от позднего non-terminal update;
  - выбора canonical comment при duplicate terminal comments.
- Актуализирована Sprint S7 traceability:
  `docs/delivery/issue_map.md`,
  `docs/delivery/sprints/s7/sprint_s7_mvp_readiness_gap_closure.md`,
  `docs/delivery/epics/s7/epic_s7.md`;
  remaining backlog нормализован как `#254`, `#255`, `#259..#260`.
- Проверки по scope:
  `go test ./services/internal/control-plane/internal/domain/runstatus/...`.

## Актуализация по Issue #259 (`run:dev`, 2026-03-11)
- Для FR-009/FR-036 и NFR-013/NFR-017 реализован stream `S7-E17`:
  persisted snapshot в `agent_sessions` переведён на versioned CAS-like persistence
  с полями `snapshot_version`, `snapshot_checksum`, `snapshot_updated_at`.
- В `services/internal/control-plane/internal/repository/postgres/agentsession`
  write-path больше не перетирает последний non-empty `codex_cli_session_json` пустым retry/update,
  одинаковый replay становится идемпотентным, а stale rewrite возвращает conflict с `actual_snapshot_version`.
- Внутренний gRPC callback `UpsertAgentSession`/`GetLatestAgentSession`
  расширен version/checksum metadata; conflict возвращается через typed gRPC status details,
  а `agent-runner` держит `snapshot_version` текущего run и публикует enriched `run.agent.session.saved`.
- Добавлены migration/runtime артефакты:
  `services/internal/control-plane/cmd/cli/migrations/20260311143000_day28_agent_session_snapshot_versioning.sql`,
  `libs/go/domain/agent/session_snapshot_checksum.go`,
  обновлённые `proto/codexk8s/controlplane/v1/controlplane.proto` и generated Go contracts.
- Актуализированы traceability документы Sprint S7:
  `docs/delivery/issue_map.md`,
  `docs/delivery/delivery_plan.md`,
  `docs/delivery/sprints/s7/sprint_s7_mvp_readiness_gap_closure.md`,
  `docs/delivery/epics/s7/epic_s7.md`;
  remaining backlog нормализован как `#254`, `#255`, `#260`.
- Проверки по scope:
  `make gen-proto-go`,
  `go test ./libs/go/domain/agent ./services/internal/control-plane/... ./services/jobs/agent-runner/...`,
  `make lint-go`,
  `make dupl-go`.

## Актуализация по Issue #260 (`run:dev`, 2026-03-11)
- Для FR-028/FR-033/FR-045 и NFR-010 реализован stream `S7-E18`:
  documentation governance runtime-проекция в `services.yaml`
  синхронизирована с `docs/delivery/development_process_requirements.md`.
- `services.yaml/spec.roleDocTemplates` теперь соответствует role-template matrix:
  - `em` получает полный release/delivery template set;
  - `dev` ограничен `user_story.md` + `definition_of_done.md`;
  - `qa` дополнен `postdeploy_review.md`;
  - `sre` переведён на ops/incident templates без drift;
  - `km` получает delivery/roadmap/docset traceability templates;
  - `reviewer` больше не получает лишние doc templates.
- `services.yaml/spec.projectDocs` теперь гарантирует доступ runtime prompt к
  `docs/delivery` для `dev/qa/sre` и к `docs/ops` для `qa`,
  чтобы role-aware prompt envelope ссылался на релевантные source-of-truth каталоги,
  а не на устаревший урезанный набор.
- В `docs/delivery/development_process_requirements.md`
  явно закреплено, что `services.yaml/spec.roleDocTemplates` и `services.yaml/spec.projectDocs`
  обязаны оставаться синхронными с role-template matrix и doc IA;
  тот же governance contract продублирован в
  `services/jobs/agent-runner/internal/runner/promptseeds/README.md`.
- Добавлены regression-тесты agent-runner,
  которые проверяют реальный repository `services.yaml` на соответствие governance-матрице
  и наличие delivery/ops doc coverage для нужных ролей.
- Актуализированы traceability документы Sprint S7:
  `docs/delivery/issue_map.md`,
  `docs/delivery/delivery_plan.md`,
  `docs/delivery/development_process_requirements.md`,
  `docs/delivery/sprints/s7/sprint_s7_mvp_readiness_gap_closure.md`,
  `docs/delivery/epics/s7/epic_s7.md`;
  remaining standalone backlog нормализован как `#254`.
- Проверки по scope:
  `go test ./services/jobs/agent-runner/internal/runner/...`,
  `go test ./services/jobs/agent-runner/...`.

## Актуализация по Issue #274 (`run:dev`, 2026-03-05)
- Для FR-026/FR-028/FR-033 и NFR-010/NFR-018 реализован stream `S7-E19`:
  выполнен backend cleanup non-MVP контуров `agents`, `prompt templates`, `config entries`,
  `runtime-deploy registry images` и `running jobs`.
- В OpenAPI удалены staff endpoint-ы:
  `/staff/agents*`, `/staff/prompt-templates*`, `/staff/config-entries*`,
  `/staff/runs/jobs`, `/staff/runtime-deploy/images*`; синхронно обновлены backend/frontend codegen артефакты.
- В `control-plane` удалены соответствующие non-MVP RPC из `ControlPlaneService` и
  runtime/domain-реализации staff use-cases (transport + domain + app wiring).
- В control-plane удалены неиспользуемые SQL/repository слои:
  `internal/repository/postgres/prompttemplate/*`,
  `internal/repository/postgres/configentry/*`,
  staff-only части `internal/repository/postgres/agent/*` (list/get/update settings),
  а также связанные domain types/repository contracts.
- Добавлена миграция
  `services/internal/control-plane/cmd/cli/migrations/20260305170000_day27_staff_non_mvp_schema_cleanup.sql`
  для удаления non-MVP схемы:
  `drop table prompt_templates`, `drop table config_entries`,
  `alter table agents drop columns settings/settings_version`.
- В `api-gateway` удалены связанные transport handlers/call-builders;
  для `web-console` очищен dead scaffold (`features/agents`, `configuration/Agents*`).
- Обновлены архитектурные и delivery-артефакты трассировки:
  `docs/architecture/api_contract.md`, `docs/delivery/issue_map.md`, `docs/delivery/requirements_traceability.md`.
- Проверки по scope:
  `go test ./services/external/api-gateway/... ./services/internal/control-plane/...`,
  `go mod tidy`,
  `make lint-go`, `make dupl-go`.
