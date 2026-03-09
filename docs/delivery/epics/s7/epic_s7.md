---
doc_id: EPC-CK8S-0007
type: epic
title: "Epic Catalog: Sprint S7 (MVP readiness gap closure)"
status: in-progress
owner_role: PM
created_at: 2026-02-27
updated_at: 2026-03-05
related_issues: [212, 218, 220, 222, 238, 241, 243, 244, 245, 246, 247, 248, 249, 250, 251, 252, 253, 254, 255, 256, 257, 258, 259, 260, 274, 199, 201, 210, 216]
related_prs: [213, 215]
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-27-issue-212-intake"
---

# Epic Catalog: Sprint S7 (MVP readiness gap closure)

## TL;DR
- Sprint S7 консолидирует незакрытые MVP-разрывы из UI, stage-flow и delivery-governance в единый execution backlog.
- Day1 intake (`#212`) зафиксировал P0/P1/P2-потоки и актуализировал S6 dependency-chain: `#199/#201` закрыты, открытый блокер — `#216` (`run:release`).
- Цель каталога: дать однозначную stage-декомпозицию и candidate backlog на 18 эпиков + post-plan `S7-E19` до полного readiness цикла `dev -> qa -> release -> postdeploy -> ops -> doc-audit`.

## Stage roadmap
- Day 1 (Intake): `docs/delivery/epics/s7/epic-s7-day1-mvp-readiness-intake.md` (Issue `#212`).
- Day 2 (Vision): `docs/delivery/epics/s7/epic-s7-day2-mvp-readiness-vision.md` (Issue `#218`).
- Day 3 (PRD): `docs/delivery/epics/s7/epic-s7-day3-mvp-readiness-prd.md` + `docs/delivery/epics/s7/prd-s7-day3-mvp-readiness-gap-closure.md` (Issue `#220`).
- Day 4 (Architecture): `docs/delivery/epics/s7/epic-s7-day4-mvp-readiness-arch.md` (Issue `#222`).
- Day 5 (Design): `docs/delivery/epics/s7/epic-s7-day5-mvp-readiness-design.md` (Issue `#238`).
- Day 6 (Plan): `docs/delivery/epics/s7/epic-s7-day6-mvp-readiness-plan.md` (Issue `#241`).
- Day 7+ (Execution): реализация и приемка `run:dev -> run:qa -> run:release -> run:postdeploy -> run:ops -> run:doc-audit` по implementation issues `#243..#260`.

## Day 2 vision fact
- В Issue `#218` зафиксированы mission, KPI/success metrics и measurable readiness criteria для потоков `S7-E01..S7-E18`.
- Для каждого execution-эпика оформлен baseline: user story, AC, edge cases, expected evidence.
- Зафиксировано обязательное правило decomposition parity перед входом в `run:dev`:
  `approved_execution_epics_count == created_run_dev_issues_count` (coverage ratio = `1.0`).
- Создана continuity issue `#220` для этапа `run:prd` без trigger-лейбла.

## Day 3 PRD fact
- В Issue `#220` подготовлен PRD-пакет Sprint S7:
  - `docs/delivery/epics/s7/epic-s7-day3-mvp-readiness-prd.md`;
  - `docs/delivery/epics/s7/prd-s7-day3-mvp-readiness-gap-closure.md`.
- Для каждого execution-эпика `S7-E01..S7-E18` формализованы `user story`, `FR`, `AC`, `NFR`, `edge cases`, `expected evidence`.
- Зафиксированы dependency graph и sequencing-waves для перехода `run:prd -> run:arch -> run:design -> run:plan`.
- Подтверждено parity-правило перед `run:dev`:
  `approved_execution_epics_count == created_run_dev_issues_count`.
- Зафиксирована owner policy для MVP: custom agents/prompt lifecycle выведены из scope, prompt templates изменяются через repo workflow.
- Создана continuity issue `#222` для этапа `run:arch` без trigger-лейбла.

## Day 4 architecture fact
- В Issue `#222` выпущен architecture package Sprint S7:
  - `docs/delivery/epics/s7/epic-s7-day4-mvp-readiness-arch.md`;
  - `docs/architecture/s7_mvp_readiness_gap_closure_architecture.md`;
  - `docs/architecture/c4_context_s7_mvp_readiness_gap_closure.md`;
  - `docs/architecture/c4_container_s7_mvp_readiness_gap_closure.md`;
  - `docs/architecture/adr/ADR-0010-s7-mvp-readiness-stream-boundaries-and-parity-gate.md`;
  - `docs/architecture/alternatives/ALT-0002-s7-mvp-readiness-stream-architecture.md`.
- Для `S7-E01..S7-E18` зафиксированы сервисные границы, ownership и contract/data impact matrix.
- Подтверждены wave-sequencing ограничения и architecture guard перед `run:dev`:
  `approved_execution_epics_count == created_run_dev_issues_count`.
- Создана continuity issue `#238` для этапа `run:design` без trigger-лейбла.

## Day 5 design fact
- В Issue `#238` выпущен design package Sprint S7:
  - `docs/delivery/epics/s7/epic-s7-day5-mvp-readiness-design.md`;
  - `docs/architecture/s7_mvp_readiness_gap_closure_design_doc.md`;
  - `docs/architecture/s7_mvp_readiness_gap_closure_api_contract.md`;
  - `docs/architecture/s7_mvp_readiness_gap_closure_data_model.md`;
  - `docs/architecture/s7_mvp_readiness_gap_closure_migrations_policy.md`.
- Для потоков `S7-E06/S7-E07/S7-E09/S7-E10/S7-E13/S7-E16/S7-E17` зафиксированы typed contract decisions и risk-mitigation.
- Для persisted-state потоков определены migration/rollback правила и rollout order `migrations -> internal -> edge -> frontend`.
- Создана continuity issue `#241` для этапа `run:plan` без trigger-лейбла.

## Day 6 plan fact
- В Issue `#241` выпущен plan package Sprint S7:
  - `docs/delivery/epics/s7/epic-s7-day6-mvp-readiness-plan.md`.
- По owner-уточнению создана отдельная implementation issue на каждый execution-поток `S7-E01..S7-E18`:
  - `#243..#260`.
- Зафиксирован parity-gate перед входом в `run:dev`:
  `approved_execution_epics_count == created_run_dev_issues_count` (`18 == 18`).
- Зафиксирован launch-order по waves и правило запуска:
  trigger `run:dev` на implementation issues ставит Owner в порядке wave-sequencing.
- Post-plan добавление: issue `#274` (`S7-E19`) для backend cleanup Agents/Configs/Secrets (Owner request из PR #272).

## Day 7 execution fact (`S7-E01`)
- В Issue `#243` реализован foundation stream `S7-E01`:
  зафиксирован единый deterministic rebase/mainline процесс для revise-итераций в `run:dev`.
- Process baseline закреплён в `docs/delivery/development_process_requirements.md`:
  обязательный `rebase -> conflict-marker check -> checks -> force-with-lease` и PR checklist для команды.
- В traceability добавлена отдельная запись по issue `#243`; remaining backlog нормализован как `#245..#260` после реализации `#244`.

## Day 7 execution fact (`S7-E02`)
- В Issue `#244` реализован UI cleanup stream `S7-E02`:
  из sidebar удалены non-MVP секции `governance`/`admin`, non-MVP пункты `configuration/docs`, `configuration/mcp-tools`,
  `configuration/agents`, `configuration/config-entries`, а также operations-контуры `runtime-deploy/images` и `running-jobs`.
- В роутинге staff UI удалены соответствующие non-MVP маршруты и добавлен fallback redirect на `projects` для stale deep-links.
- Удалён связанный dead code non-MVP страниц, UI-контур `config-entries` и platform-tokens scaffold в `system-settings`.
- В traceability добавлены обновления по issue `#244`; remaining backlog нормализован как `#245..#260`, а post-plan issue `#274` переведён в `in-review`.

## Day 7 execution fact (`S7-E03`)
- В Issue `#245` реализован stream удаления глобального фильтра:
  из `App.vue` удалён global filter entry с зависимым summary/reset UI-контуром.
- В `features/ui-context/store.ts` удалено глобальное состояние `env/namespace` и связанный cookie-persistence код;
  сохранён только MVP-нужный selected project context.
- В `pages/operations/RuntimeDeployTasksPage.vue` удалена зависимость загрузки списка от `uiContext.env`:
  данные больше не отфильтровываются global env-фильтром.
- Удалён неиспользуемый компонент `shared/ui/AdminClusterContextBar.vue` и очищены i18n-ключи глобального фильтра.
- В traceability добавлены обновления по issue `#245`; remaining backlog нормализован как `#246..#260` + post-plan `#274`.

## Day 7 execution fact (`S7-E04`)
- В Issue `#246` финализирован stream удаления `runtime-deploy/images` из MVP без нового redirect-кода:
  owner-review подтвердил, что после базового cleanup `#244` отдельный redirect для `/runtime-deploy/images*`
  не нужен.
- Stale deeplink продолжает закрываться уже существующим catch-all route
  `/:pathMatch(.*)* -> projects`, поэтому `runtime-deploy/images` не возвращается в MVP-контур.
- В traceability добавлены updates по issue `#246`; remaining backlog нормализован как `#247..#260` + post-plan `#274`.

## Candidate execution backlog (19 эпиков)

| Epic ID | Priority | Scope | Источник замечаний |
|---|---|---|---|
| S7-E01 | P0 | Rebase/mainline hygiene и merge-conflict policy для PR-итераций | PRC-01 |
| S7-E02 | P0 | Удаление не-MVP разделов (включая Agents, Configs/Secrets, Registry images, Running jobs) и связанного dead code | PRC-05 |
| S7-E03 | P0 | Удаление глобального frontend-фильтра и связанного неиспользуемого кода (выполнено в `#245`) | PRC-04 |
| S7-E04 | P0 | Удаление runtime-deploy/images контуров; stale deeplinks остаются на общем fallback route (in-review `#246`) | PRC-02, PRC-05 |
| S7-E05 | P0 | Agents UI cleanup: убрать badge `Скоро`, пересобрать таблицу (без role/project-id) | PRC-03 |
| S7-E06 | P0 | Agents MVP de-scope: убрать runtime mode/locale настройки, оставить фиксированные platform defaults | PRC-03 |
| S7-E07 | P0 | Prompt source MVP contract: удалить selector `repo|db`, закрепить `repo-only` policy | PRC-03 |
| S7-E08 | P1 | Agents UX de-scope hardening: удалить non-MVP массовые операции и cleanup зависимого UX | PRC-03 |
| S7-E09 | P0 | Runs UX: удалить колонку типа запуска и гарантировать namespace delete из run details | PRC-06 |
| S7-E10 | P0 | Runtime deploy UX: кнопка cancel/stop для зависших deploy tasks + guardrails | PRC-07 |
| S7-E11 | P0 | Label orchestration reliability: исправить `mode:discussion` trigger-поведение | PRC-08 |
| S7-E12 | P1 | Final MVP readiness gate: e2e evidence bundle + go/no-go для release chain | PRC-01..PRC-08 |
| S7-E13 | P0 | Label policy alignment: добавить `run:qa:revise` и покрыть revise-loop QA-stage | PRC-09 |
| S7-E14 | P0 | QA execution contract: проверка новых/изменённых ручек через Kubernetes DNS path + evidence | PRC-10 |
| S7-E15 | P0 | Prompt templates MVP policy: изменения только через repo commit workflow, без UI refresh/versioning | PRC-11 |
| S7-E16 | P0 | Run status reliability: устранить false-failed для фактически успешных `run:intake:revise` | PRC-12 |
| S7-E17 | P0 | Self-improve reliability: доступность и корректная перезапись `agent_sessions` snapshot | PRC-13 |
| S7-E18 | P0 | Documentation governance: единый стандарт issue/PR + doc IA + role-template matrix | PRC-14, PRC-15, PRC-16 |
| S7-E19 | P1 | Backend cleanup: удалить non-MVP контуры Agents/Configs/Secrets + registry images + running jobs | Owner request (PR #272) |

## Delivery-governance правила
- Каждая следующая stage-issue создаётся отдельной задачей и без trigger-лейбла.
- Trigger-лейбл на запуск этапа ставит Owner после review предыдущего артефакта.
- Для каждого execution-эпика обязательно фиксируются: priority, user story, AC, edge cases, dependency и expected evidence.
- MVP-closeout не считается завершённым без явного доказательства работоспособности `run:doc-audit`.
