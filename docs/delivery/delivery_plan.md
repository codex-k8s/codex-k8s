---
doc_id: PLN-CK8S-0001
type: delivery-plan
title: "codex-k8s — Delivery Plan"
status: active
owner_role: EM
created_at: 2026-02-06
updated_at: 2026-03-25
related_issues: [1, 19, 74, 100, 106, 112, 154, 155, 170, 171, 184, 185, 187, 189, 195, 197, 199, 201, 210, 212, 216, 218, 220, 222, 223, 225, 226, 227, 228, 229, 230, 238, 241, 243, 244, 245, 246, 247, 248, 249, 250, 251, 252, 253, 254, 255, 256, 257, 258, 259, 260, 262, 263, 265, 274, 281, 282, 320, 333, 335, 337, 340, 351, 360, 361, 363, 366, 369, 370, 371, 372, 373, 374, 375, 378, 383, 385, 387, 389, 391, 392, 393, 394, 395, 413, 416, 418, 420, 423, 425, 426, 427, 428, 429, 430, 431, 444, 447, 448, 452, 454, 456, 458, 469, 471, 476, 480, 484, 490, 492, 494, 496, 500, 510, 512, 516, 519, 521, 522, 523, 524, 525, 537, 541, 542, 543, 544, 545, 546, 547, 554, 557]

related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-19-full-docset"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-19
---

# Delivery Plan: codex-k8s

## TL;DR
- Что поставляем: MVP control-plane + staff UI + webhook orchestration + MCP governance + self-improve loop + production bootstrap/deploy loop.
- Когда: поэтапно, с ранним production для ручных тестов.
- Главные риски: bootstrap automation, security/governance hardening, runner stability.
- Что нужно от Owner: подтверждение deploy-модели и доступов (GitHub fine-grained token/OpenAI key).

## Входные артефакты
- Requirements baseline: `docs/product/requirements_machine_driven.md`
- Brief: `docs/product/brief.md`
- Constraints: `docs/product/constraints.md`
- Agents operating model: `docs/product/agents_operating_model.md`
- Labels policy: `docs/product/labels_and_trigger_policy.md`
- Stage process model: `docs/product/stage_process_model.md`
- Architecture (C4): `docs/architecture/c4_context.md`, `docs/architecture/c4_container.md`
- ADR: `docs/architecture/adr/ADR-0001-kubernetes-only.md`, `docs/architecture/adr/ADR-0002-webhook-driven-and-deploy-workflows.md`, `docs/architecture/adr/ADR-0003-postgres-jsonb-pgvector.md`, `docs/architecture/adr/ADR-0004-repository-provider-interface.md`
- Data model: `docs/architecture/data_model.md`
- Runtime/RBAC model: `docs/architecture/agent_runtime_rbac.md`
- MCP approval/audit flow: `docs/architecture/mcp_approval_and_audit_flow.md`
- Prompt templates policy: `docs/architecture/prompt_templates_policy.md`
- Sprint plan: `docs/delivery/sprints/s1/sprint_s1_mvp_vertical_slice.md`
- Epic catalog: `docs/delivery/epics/s1/epic_s1.md`
- Sprint S2 plan: `docs/delivery/sprints/s2/sprint_s2_dogfooding.md`
- Epic S2 catalog: `docs/delivery/epics/s2/epic_s2.md`
- Sprint S3 plan: `docs/delivery/sprints/s3/sprint_s3_mvp_completion.md`
- Epic S3 catalog: `docs/delivery/epics/s3/epic_s3.md`
- Sprint S4 plan: `docs/delivery/sprints/s4/sprint_s4_multi_repo_federation.md`
- Epic S4 catalog: `docs/delivery/epics/s4/epic_s4.md`
- Sprint S5 plan: `docs/delivery/sprints/s5/sprint_s5_stage_entry_and_label_ux.md`
- Epic S5 catalog: `docs/delivery/epics/s5/epic_s5.md`
- Sprint S6 plan: `docs/delivery/sprints/s6/sprint_s6_agents_prompt_management.md`
- Epic S6 catalog: `docs/delivery/epics/s6/epic_s6.md`
- Sprint S7 plan: `docs/delivery/sprints/s7/sprint_s7_mvp_readiness_gap_closure.md`
- Epic S7 catalog: `docs/delivery/epics/s7/epic_s7.md`
- Sprint S8 plan: `docs/delivery/sprints/s8/sprint_s8_go_refactoring_parallelization.md`
- Epic S8 catalog: `docs/delivery/epics/s8/epic_s8.md`
- Sprint index: `docs/delivery/sprints/README.md`
- Epic index: `docs/delivery/epics/README.md`
- E2E master plan: `docs/delivery/e2e_mvp_master_plan.md`
- Process requirements: `docs/delivery/development_process_requirements.md`

## Структура работ (WBS)
### Sprint S1: Day 0 + Day 1..7 (8 эпиков)
- Day 0 (completed): `docs/delivery/epics/s1/epic-s1-day0-bootstrap-baseline.md`
- Day 1: `docs/delivery/epics/s1/epic-s1-day1-webhook-idempotency.md`
- Day 2: `docs/delivery/epics/s1/epic-s1-day2-worker-slots-k8s.md`
- Day 3: `docs/delivery/epics/s1/epic-s1-day3-auth-rbac-ui.md`
- Day 4: `docs/delivery/epics/s1/epic-s1-day4-repository-provider.md`
- Day 5: `docs/delivery/epics/s1/epic-s1-day5-learning-mode.md`
- Day 6: `docs/delivery/epics/s1/epic-s1-day6-hardening-observability.md`
- Day 7: `docs/delivery/epics/s1/epic-s1-day7-stabilization-gate.md`

### Sprint S2: Dogfooding baseline + hardening (Day 0..7)
- Day 0..4 (completed): архитектурное выравнивание, label triggers, namespace/RBAC, MCP prompt context, agent PR flow.
- Day 4.5 (completed): pgx/db-model refactor.
- Day 5 (completed): staff UI observability baseline.
- Day 6 (completed): approval matrix + MCP control tools + audit hardening.
- Day 7 (completed): MVP readiness regression gate + Sprint S3 kickoff package (`docs/delivery/regression_s2_gate.md`).

### Sprint S3: MVP completion (Day 1..21)
- Day 1: full stage/label activation.
- Day 2: staff runtime debug console.
- Day 3: deterministic secret sync (Kubernetes).
- Day 4: database lifecycle MCP tools.
- Day 5: owner feedback handle + HTTP approver/executor + Telegram adapter.
- Day 6..7: `run:self-improve` ingestion + updater + PR flow.
- Day 8: agent toolchain auto-extension safeguards.
- Day 9: declarative full-env deploy, `services.yaml` orchestration, runtime parity/hot-reload.
- Day 10 (completed): полный redesign staff-консоли на Vuetify.
- Day 11 (completed): full-env slots + agent-run + subdomain templating (TLS) для manual QA.
- Day 12 (completed): docset import + safe sync (`agent-knowledge-base` -> projects).
- Day 13 (completed): unified config/secrets governance (platform/project/repo) + GitHub creds fallback.
- Day 14 (completed): repository onboarding preflight (token scopes + GitHub ops + domain resolution) + bot params per repo.
- Day 16 (completed): gRPC transport boundary hardening (transport -> service -> repository) по Issue #45.
- Day 15: prompt context overhaul (`services.yaml` docs tree + role prompt matrix + GitHub service messages templates).
- Day 17: environment-scoped secret overrides + OAuth callback strategy (без project-specific hardcode).
- Day 18: runtime error journal + staff alert center (stacked alerts, mark-as-viewed).
- Day 19: frontend manual QA hardening loop (Owner-driven bug cycle до full e2e).
- Day 19.5: realtime шина на PostgreSQL (`event log + LISTEN/NOTIFY`) + multi-server WebSocket backplane.
- Day 19.6: интеграция realtime подписок в staff UI (runs/deploy/errors/logs/events), удаление кнопок `Обновить` в realtime-экранах, fallback polling.
- Day 19.7: retention full-env namespace по role-based TTL + lease extension/reuse на `run:*:revise` (Issue #74).
- Day 20: full e2e regression/security gate + MVP closeout/handover и переход к post-MVP roadmap (подробности в `docs/delivery/e2e_mvp_master_plan.md`).

### Sprint S4: Multi-repo runtime and docs federation (Issue #100)
- Day 1 (completed): execution foundation для federated multi-repo composition и docs federation (`docs/delivery/epics/s4/epic-s4-day1-multi-repo-composition-and-docs-federation.md`).
- Результат Day 1: формальный execution-plan (stories + quality-gates + owner decisions) для перехода в `run:dev`.
- Следующие day-эпики S4 формируются после Owner review Day 1 и закрытия зависимостей по S3 Day20.

### Sprint S5: Stage entry and label UX orchestration (Issues #154/#155/#170/#171)
- Day 1 (in-review): launch profiles + deterministic next-step actions (`docs/delivery/epics/s5/epic-s5-day1-launch-profiles-and-stage-launcher-ux.md`).
- Результат Day 1 (факт): owner-ready vision/prd + architecture execution package для входа в `run:dev` подготовлен в Issue #155 (включая ADR-0008); Owner approval получен (PR #166, 2026-02-25).
- Day 2 (in-review): single-epic execution package для реализации FR-053/FR-054 (`docs/delivery/epics/s5/epic-s5-day2-launch-profiles-dev-execution.md`).
- Результат Day 2 (факт): в Issue #170 зафиксирован delivery governance пакет (QG-D2-01..QG-D2-05, DoD, handover), создана implementation issue #171 для выполнения одним эпиком.

### Sprint S6: Agents configuration and prompt templates lifecycle (Issue #184)
- Day 1 (in-review): intake baseline по разделу `Agents` (`docs/delivery/epics/s6/epic-s6-day1-agents-prompts-intake.md`).
- Результат Day 1 (факт): подтвержден разрыв между scaffold UI и отсутствием staff API контрактов для agents/templates/audit; зафиксирована полная stage-траектория до `run:doc-audit` и требование создавать follow-up issue на каждом этапе без постановки `run:*`-лейбла при создании (trigger-лейбл ставит Owner).
- Day 2 (in-review): vision baseline в issue #185 с зафиксированными mission/KPI, границами MVP/Post-MVP и риск-рамкой.
- Day 3 (in-review): PRD stage в issue #187:
  - `docs/delivery/epics/s6/epic-s6-day3-agents-prompts-prd.md`
  - `docs/delivery/epics/s6/prd-s6-day3-agents-prompts-lifecycle.md`
- Результат Day 3 (факт): формализованы FR/AC/NFR-draft для `agents settings + prompt lifecycle + audit/history`; создана issue #189 для stage `run:arch` без постановки trigger-лейбла (лейбл ставит Owner) и с обязательной инструкцией создать issue `run:design`.
- Day 4 (in-review): architecture stage в issue #189 (`docs/delivery/epics/s6/epic-s6-day4-agents-prompts-arch.md`).
- Результат Day 4 (факт): зафиксированы архитектурные границы и ADR-0009, создана issue #195 для stage `run:design`.
- Day 5 (in-review): design stage в issue #195 (`docs/delivery/epics/s6/epic-s6-day5-agents-prompts-design.md`).
- Результат Day 5 (факт): зафиксирован implementation-ready package (`design_doc`, `api_contract`, `data_model`, `migrations_policy`), создана issue #197 для stage `run:plan`.
- Day 6 (in-review): plan stage в issue #197 (`docs/delivery/epics/s6/epic-s6-day6-agents-prompts-plan.md`).
- Результат Day 6 (факт): сформирован execution package `run:dev` (W1..W7, QG-S6-D6-01..QG-S6-D6-07, DoR/DoD, blockers/risks/owner decisions), создана issue #199 для stage `run:dev` без trigger-лейбла.
- Day 7 (completed): dev stage в issue #199 (contract-first/migrations/staff transport/UI integration).
- Результат Day 7 (факт): реализация `agents/templates/audit` завершена в PR #202 (merged), сформирован regression evidence package и создана issue #201 для stage `run:qa`.
- Day 8 (completed): QA stage в issue #201 закрыт с решением GO в `run:release`; создана issue #216 для следующего этапа release-continuity.
- Day 9 (completed): release closeout в issue #262 с фиксацией release-governance пакета (`quality-gates`, DoD, release notes, rollback strategy).
- Day 10 (in-review): postdeploy review в issue #263 с runtime evidence, обновлением ops handover и проверкой rollback readiness.
- Результат Day 10 (факт): сформирована follow-up issue `#265` для stage `run:ops` (без trigger-лейбла, лейбл ставит Owner).
- Day 11 (in-review): ops closeout в issue #265 с фиксацией production baseline по runbook/monitoring/alerts/SLO/rollback.
- Результат Day 11 (факт): операционный хвост S6 закрыт, traceability синхронизирована, следующий continuity-шаг переведён в `run:doc-audit` issue flow.
- Следующий continuity-контур S6: `ops -> doc-audit` с отдельной issue на каждый этап.

### Sprint S7: MVP readiness gap closure (Issue #212)
- Day 1 (in-review): intake пакет по незакрытым MVP-разрывам (`docs/delivery/epics/s7/epic-s7-day1-mvp-readiness-intake.md`).
- Результат Day 1 (факт): подтверждены P0/P1/P2-потоки и dependency-блокеры:
  - release-зависимость S6 закрыта (`#262`), активный continuity-блокер перенесён в postdeploy issue `#263`;
  - крупный UI-scaffold контур с `comingSoon`/TODO в staff web-console;
  - отсутствие подтверждённого run-evidence для `run:doc-audit` в текущем delivery-цикле.
- Дополнительно по owner-review комментариям сформирована candidate-декомпозиция на 18 execution-эпиков (`S7-E01..S7-E18`) + post-plan `S7-E19` с приоритетами и трассировкой в `docs/delivery/epics/s7/epic_s7.md`.
- Добавлены отдельные P0-потоки для:
  - coverage недостающих revise-петель `run:doc-audit|qa|release|postdeploy|ops|self-improve:revise` в stage/labels policy;
  - QA acceptance-проверок через Kubernetes DNS path для новых/изменённых ручек;
  - reliability-контуров (`run:intake:revise` false-failed, `run:self-improve` session snapshot persistence);
  - документационного governance (единый issue/PR стандарт + doc IA + role-template matrix).
- Day 2 (in-review): vision-пакет в Issue `#218` (`docs/delivery/epics/s7/epic-s7-day2-mvp-readiness-vision.md`).
- Результат Day 2 (факт):
  - зафиксированы mission, KPI/success metrics и measurable readiness criteria по `S7-E01..S7-E18`;
  - для каждого execution-эпика оформлен baseline (`user story`, `AC`, `edge cases`, `expected evidence`);
  - закреплено governance-правило decomposition parity перед `run:dev`:
    `approved_execution_epics_count == created_run_dev_issues_count`;
  - создана follow-up issue `#220` для stage `run:prd` без trigger-лейбла.
- Day 3 (in-review): PRD-пакет в Issue `#220`:
  - `docs/delivery/epics/s7/epic-s7-day3-mvp-readiness-prd.md`;
  - `docs/delivery/epics/s7/prd-s7-day3-mvp-readiness-gap-closure.md`.
- Результат Day 3 (факт):
  - по всем потокам `S7-E01..S7-E18` формализованы `user story`, `FR/AC/NFR`, `edge cases`, `expected evidence`;
  - зафиксированы dependency graph и sequencing-waves для перехода `run:prd -> run:arch -> run:design -> run:plan`;
  - закреплён owner policy для MVP: custom agents/prompt lifecycle выведены в post-MVP, prompt templates меняются через repo workflow;
  - подтверждено parity-правило перед `run:dev`: `approved_execution_epics_count == created_run_dev_issues_count`;
  - создана follow-up issue `#222` для stage `run:arch` без trigger-лейбла.
- Day 4 (in-review): architecture stage в issue `#222`:
  - `docs/delivery/epics/s7/epic-s7-day4-mvp-readiness-arch.md`;
  - `docs/architecture/initiatives/s7_mvp_readiness_gap_closure/architecture.md`;
  - `docs/architecture/initiatives/s7_mvp_readiness_gap_closure/c4_context.md`;
  - `docs/architecture/initiatives/s7_mvp_readiness_gap_closure/c4_container.md`;
  - `docs/architecture/adr/ADR-0010-s7-mvp-readiness-stream-boundaries-and-parity-gate.md`;
  - `docs/architecture/alternatives/ALT-0002-s7-mvp-readiness-stream-architecture.md`.
- Результат Day 4 (факт):
  - зафиксированы service boundaries/ownership matrix по `S7-E01..S7-E18`;
  - подтверждены wave-sequencing ограничения и architecture parity-gate перед `run:dev`;
  - создана follow-up issue `#238` для stage `run:design` без trigger-лейбла.
- Day 5 (in-review): design stage в issue `#238`:
  - `docs/delivery/epics/s7/epic-s7-day5-mvp-readiness-design.md`;
  - `docs/architecture/initiatives/s7_mvp_readiness_gap_closure/design_doc.md`;
  - `docs/architecture/initiatives/s7_mvp_readiness_gap_closure/api_contract.md`;
  - `docs/architecture/initiatives/s7_mvp_readiness_gap_closure/data_model.md`;
  - `docs/architecture/initiatives/s7_mvp_readiness_gap_closure/migrations_policy.md`.
- Результат Day 5 (факт):
  - зафиксированы typed contract decisions для потоков `S7-E06/S7-E07/S7-E09/S7-E10/S7-E13/S7-E16/S7-E17`;
  - формализованы data/migration/rollback правила для persisted-state потоков;
  - создана follow-up issue `#241` для stage `run:plan` без trigger-лейбла.
- Day 6 (in-review): plan stage в issue `#241`:
  - `docs/delivery/epics/s7/epic-s7-day6-mvp-readiness-plan.md`.
- Результат Day 6 (факт):
  - по owner-уточнению создана отдельная implementation issue на каждый execution-поток `S7-E01..S7-E18`;
  - сформирован execution issue package `#243..#260` без trigger-лейблов с wave-sequencing;
  - parity-гейт перед `run:dev` подтверждён: `approved_execution_epics_count == created_run_dev_issues_count` (`18 == 18`).
  - post-plan добавление: issue `#274` (`S7-E19`) на backend cleanup Agents/Configs/Secrets + registry images + running jobs.
- Day 7+ (in-progress): `dev -> qa -> release -> postdeploy -> ops -> doc-audit` по implementation issues `#243..#260`, `#274` и owner-governed trigger-лейблам.
  - На текущий момент `#243` и `#244` owner-approved; `#245`, `#246`, `#247/#248/#249`, `#251`, `#252`, `#253`, `#255`, `#256`, `#258`, `#259`, `#260` и `#274` реализованы в execution streams.
  - Standalone issues `#250` и `#257` закрываются doc-actualization pass как уже поглощённые cleanup-потоками.
  - Remaining standalone backlog Sprint S7 после актуализации `#260`: `#254`.

### Sprint S8: Go refactoring parallelization + repository onboarding automation
- Day 1 (in-review): plan-пакет по параллельному Go-рефакторингу (`docs/delivery/epics/s8/epic-s8-day1-go-refactoring-plan.md`).
- Результат Day 1 (факт):
  - execution-поток выделен из Sprint S7 для исключения конфликтов с параллельными задачами MVP readiness;
  - сохранены 6 независимых implementation issues `#225..#230` в bounded scopes;
  - quality-gates `QG-223-01..QG-223-05` и handover в `run:dev` зафиксированы в Sprint S8.
- Day 2 (planned): empty repository initialization (`docs/delivery/epics/s8/epic-s8-day2-empty-repository-initialization.md`, Issue `#281`).
  - Цель: автоматизировать bootstrap пустого GitHub-репозитория при attach в платформу.
  - Ожидаемый результат: default branch + initial commit + `services.yaml` + docs scaffold + onboarding summary issue.
- Day 3 (planned): existing repository adoption (`docs/delivery/epics/s8/epic-s8-day3-existing-repository-adoption.md`, Issue `#282`).
  - Цель: перевести существующий кодовый репозиторий без `services.yaml`/docs baseline в управляемый PR-based onboarding flow.
  - Ожидаемый результат: deterministic scan report + специализированная onboarding-task + PR с draft `services.yaml` и docs baseline.
- Day 4 (in-review): documentation IA refactor (`docs/delivery/epics/s8/epic-s8-day4-documentation-ia-refactor-plan.md`, Issue `#320`).
  - Цель: привести репозиторий к канонической docs IA без re-root доменов и без drift между `docs/`, `services.yaml` и открытыми issues.
  - Результат Day 4 (факт):
    - добавлены `docs/index.md`, доменные `README.md` и `docs/delivery/documentation_ia_migration_map.md`;
    - initiative/handover пакеты перенесены в `docs/architecture/initiatives/` и `docs/ops/handovers/`;
    - синхронизированы `services.yaml`, `docs/templates/*`, delivery traceability-документы и индексы;
    - синхронизированы repo-local path refs, а issue bodies `#281`, `#282`, `#312` очищены от same-repo blob links и branch-specific doc refs.

### Sprint S9: Mission Control Dashboard and console control plane
- Day 1 (in-review): intake-пакет для Mission Control Dashboard (`docs/delivery/epics/s9/epic-s9-day1-mission-control-dashboard-intake.md`, Issue `#333`).
- Результат Day 1 (факт):
  - Mission Control Dashboard зафиксирован как отдельная product initiative, а не как локальный UI-refactor staff console;
  - принят active-set control-plane baseline: work items, discussion, PR, agents, side panel, realtime updates и provider-safe быстрые действия;
  - закреплены неподвижные ограничения: GitHub-first MVP, human review в provider UI, webhook-driven orchestration, contract-first API и audit-safe command/reconciliation path;
  - рекомендован launch profile `feature` с обязательной эскалацией в `vision` и `arch`;
  - создана follow-up issue `#335` для stage `run:vision` без trigger-лейбла.
- Day 2 (in-review): vision-пакет для Mission Control Dashboard (`docs/delivery/epics/s9/epic-s9-day2-mission-control-dashboard-vision.md`, Issue `#335`).
- Результат Day 2 (факт):
  - зафиксированы mission statement, persona outcomes и north star для Mission Control Dashboard как primary control plane staff console;
  - определены measurable KPI/guardrails по situational awareness, discussion-to-task lead time, console-start coverage и reconciliation correctness;
  - подтверждены границы первой волны MVP: active-set dashboard shell, typed entities/relations, provider-safe commands и realtime baseline; voice оставлен отдельным candidate stream;
  - сохранены неподвижные ограничения инициативы: GitHub-first MVP, human review во внешнем provider UI, webhook-driven orchestration и active-set default;
  - создана follow-up issue `#337` для stage `run:prd` без trigger-лейбла.
- Day 3 (in-review): PRD-пакет для Mission Control Dashboard (`docs/delivery/epics/s9/epic-s9-day3-mission-control-dashboard-prd.md`, `docs/delivery/epics/s9/prd-s9-day3-mission-control-dashboard.md`, Issue `#337`).
- Результат Day 3 (факт):
  - зафиксированы user stories `S9-US-01..S9-US-05`, FR/AC/NFR, edge cases и expected evidence для Mission Control Dashboard;
  - wave priorities разложены как `Wave 1 pilot -> Wave 2 MVP release -> Wave 3 conditional voice stream`;
  - подтверждены product guardrails: active-set default, list fallback, provider-safe typed commands, degraded realtime fallback и external human review;
  - voice intake явно вынесен из blocking scope core MVP и оставлен условной следующей волной;
  - создана follow-up issue `#340` для stage `run:arch` без trigger-лейбла.
- Day 4 (in-review): architecture-пакет для Mission Control Dashboard (`docs/delivery/epics/s9/epic-s9-day4-mission-control-dashboard-arch.md`, `docs/architecture/initiatives/s9_mission_control_dashboard/architecture.md`, `docs/architecture/adr/ADR-0011-mission-control-dashboard-active-set-projection-and-command-reconciliation.md`, Issue `#340`).
- Результат Day 4 (факт):
  - зафиксирован ownership split: `control-plane` владеет active-set projection, relations, timeline mirror и command lifecycle, `worker` владеет provider sync/retries/reconciliation;
  - подтверждён snapshot-first / delta-second realtime baseline с обязательным degraded mode через HTTP snapshot, explicit refresh и list fallback;
  - voice intake изолирован как optional candidate stream и не входит в core MVP contracts;
  - подготовлена follow-up issue `#351` для stage `run:design` без trigger-лейбла.
- Day 5 (in-review): design-пакет для Mission Control Dashboard (`docs/delivery/epics/s9/epic-s9-day5-mission-control-dashboard-design.md`, `docs/architecture/initiatives/s9_mission_control_dashboard/design_doc.md`, `docs/architecture/initiatives/s9_mission_control_dashboard/api_contract.md`, `docs/architecture/initiatives/s9_mission_control_dashboard/data_model.md`, `docs/architecture/initiatives/s9_mission_control_dashboard/migrations_policy.md`, Issue `#351`).
- Результат Day 5 (факт):
  - зафиксирован implementation-ready package по snapshot/details/commands/realtime/voice candidate contracts;
  - выбран hybrid persisted projection model с typed tables + JSONB payload fragments под ownership `control-plane`;
  - inline write-path ограничен provider-safe typed commands, а provider review/merge/comment editing оставлены deep-link-only;
  - зафиксирован rollout order `migrations -> control-plane -> worker -> api-gateway -> web-console` и limited rollback после provider side effects;
  - создана follow-up issue `#363` для stage `run:plan` без trigger-лейбла.
- Day 6 (in-review): plan-пакет для Mission Control Dashboard (`docs/delivery/epics/s9/epic-s9-day6-mission-control-dashboard-plan.md`, Issue `#363`).
- Результат Day 6 (факт):
  - execution backlog декомпозирован на issues `#369..#375` с wave-sequencing и owner-managed handover в `run:dev`;
  - foundation/backend/transport/UI/observability были разнесены по отдельным implementation streams, чтобы не смешивать schema, domain, worker warmup execution, edge и UX scope;
  - `#371` закреплён как owner warmup/backfill execution gate, а `#372` ограничен core transport paths без voice-specific OpenAPI/codegen;
  - owner revision 2026-03-14 перевёл `#374` / `S9-E06` в superseded historical artifact: отдельная observability/readiness wave и PR `#463` не входят в active Sprint S9 backlog; `#375` по-прежнему остаётся conditional voice continuation;
  - quality-gates, DoR/DoD, blockers/risks/owner decisions синхронизированы в delivery traceability.
- Day 7+ (planned): `run:dev -> qa -> release -> postdeploy -> ops` по active issues `#369..#373`; issue `#374` сохранена только как superseded history item, issue `#375` запускается отдельным owner decision после core waves.

### Sprint S10: Built-in MCP user interactions
- Day 1 (in-review): intake-пакет для built-in MCP user interactions (`docs/delivery/epics/s10/epic-s10-day1-mcp-user-interactions-intake.md`, Issue `#360`).
- Результат Day 1 (факт):
  - инициатива зафиксирована как отдельный platform stream поверх существующего built-in server `codex_k8s`, а не как расширение approval flow;
  - MVP baseline ограничен `user.notify` и `user.decision.request` с channel-neutral semantics и typed response contract;
  - закреплены неподвижные ограничения: отдельный interaction-domain, wait-state только для response-required сценариев, platform-owned retry/idempotency/audit/correlation, Telegram как отдельный follow-up stream;
  - создана follow-up issue `#378` для stage `run:vision` без trigger-лейбла.
- Day 2 (in-review): vision-пакет для built-in MCP user interactions (`docs/delivery/epics/s10/epic-s10-day2-mcp-user-interactions-vision.md`, Issue `#378`).
- Результат Day 2 (факт):
  - built-in MCP user interactions зафиксированы как channel-neutral user-facing capability платформы, а не как расширение approval flow;
  - mission, north star, persona outcomes и KPI/guardrails определены для actionable notifications, typed user decisions, wait-state discipline и adapter readiness;
  - подтверждены неподвижные ограничения: `user.notify` остаётся non-blocking, wait-state допускается только для `user.decision.request`, delivery/retry/correlation/audit принадлежат platform domain, Telegram остаётся отдельным follow-up stream;
  - создана follow-up issue `#383` для stage `run:prd` без trigger-лейбла.
- Day 3 (in-review): PRD-пакет для built-in MCP user interactions (`docs/delivery/epics/s10/epic-s10-day3-mcp-user-interactions-prd.md`, `docs/delivery/epics/s10/prd-s10-day3-mcp-user-interactions.md`, Issue `#383`).
- Результат Day 3 (факт):
  - зафиксированы user stories, FR/AC/NFR, wave priorities и expected evidence для `user.notify`, `user.decision.request`, typed response semantics и adapter-neutral interaction contract;
  - подтверждены product guardrails: interaction flow не смешивается с approval flow, `user.notify` остаётся non-blocking, wait-state разрешён только для `user.decision.request`, delivery/retry/idempotency/correlation/audit принадлежат platform domain;
  - deferred scope явно отделён от core MVP: Telegram/adapters, reminder policies, richer threads и voice/STT не блокируют Sprint S10 core baseline;
  - создана follow-up issue `#385` для stage `run:arch` без trigger-лейбла.
- Day 4 (in-review): architecture-пакет для built-in MCP user interactions (Issue `#385`):
  - `docs/delivery/epics/s10/epic-s10-day4-mcp-user-interactions-arch.md`;
  - `docs/architecture/initiatives/s10_mcp_user_interactions/README.md`;
  - `docs/architecture/initiatives/s10_mcp_user_interactions/architecture.md`;
  - `docs/architecture/initiatives/s10_mcp_user_interactions/c4_context.md`;
  - `docs/architecture/initiatives/s10_mcp_user_interactions/c4_container.md`;
  - `docs/architecture/adr/ADR-0012-built-in-mcp-user-interactions-control-plane-owned-lifecycle.md`;
  - `docs/architecture/alternatives/ALT-0004-built-in-mcp-user-interactions-lifecycle-boundaries.md`.
- Результат Day 4 (факт):
  - зафиксирован ownership split: `control-plane` владеет interaction aggregate, wait-state transitions и audit/correlation; `worker` закреплён за dispatch/retries/expiry; `api-gateway` остаётся thin-edge callback ingress;
  - подтверждена архитектурная граница между interaction flow и approval/control flow: approval-specific semantics не становятся source-of-truth для обычных user interactions;
  - создана follow-up issue `#387` для stage `run:design` без trigger-лейбла.
- Day 5 (in-review): design-пакет для built-in MCP user interactions (Issue `#387`).
- Результат Day 5 (факт):
  - зафиксированы typed contracts для `user.notify`, `user.decision.request`, outbound adapter envelope, inbound callback family и deterministic resume payload;
  - выбрана отдельная persisted interaction-domain модель: aggregate, delivery attempts, callback evidence, response records и wait linkage к `agent_runs`/`agent_sessions`;
  - подтверждён rollout order `migrations -> control-plane -> worker -> api-gateway` и additive coexistence с approval callback family;
  - создана follow-up issue `#389` для stage `run:plan` без trigger-лейбла.
- Day 6 (in-review): plan-пакет для built-in MCP user interactions (`docs/delivery/epics/s10/epic-s10-day6-mcp-user-interactions-plan.md`, Issue `#389`).
- Результат Day 6 (факт):
  - execution backlog декомпозирован на issues `#391..#395` с wave-sequencing и owner-managed handover в `run:dev`;
  - `#391` закреплён за `control-plane` foundation, `#392` за worker dispatch/retry/expiry, `#393` за contract-first callback ingress, `#394` за deterministic resume path, `#395` за observability/evidence gate;
  - replay/idempotency/resume correctness зафиксированы как обязательный gate перед `run:qa`, а channel-specific adapters оставлены вне core Sprint S10 execution package;
  - quality-gates, DoR/DoD, blockers/risks/owner decisions синхронизированы в delivery traceability.
- Day 7+ (planned): `run:dev -> qa -> release -> postdeploy -> ops` по issues `#391..#395` с owner-managed wave launch.

### Sprint S11: Telegram-адаптер взаимодействия с пользователем
- Day 1 (in-review): intake-пакет для Telegram-адаптера как первого внешнего channel path (`docs/delivery/epics/s11/epic-s11-day1-telegram-user-interaction-adapter-intake.md`, Issue `#361`).
- Результат Day 1 (факт):
  - Telegram зафиксирован как отдельный последовательный stream после platform-core initiative Sprint S10, а не как параллельная или заменяющая её ветка;
  - MVP scope ограничен сценариями `user.notify`, `user.decision.request`, inline callbacks и optional free-text reply;
  - проверяемый readiness gate выражен явно: active vision stage должен выполняться только пока S10 plan issue `#389` остаётся closed и сохраняет design package `#387` как baseline typed interaction contract;
  - reference repositories `telegram-approver` / `telegram-executor` и planned baseline `github.com/mymmrac/telego v1.7.0` признаны ориентиром, но не source-of-truth для service boundaries и product contract;
  - сохранены неподвижные ограничения: typed platform contract, separation from approval flow, deferred scope для voice/STT, reminders и richer conversations;
  - создана follow-up issue `#444` для stage `run:vision` без trigger-лейбла.
- Day 2 (in-review): vision-package для Telegram-адаптера как первого внешнего channel-specific stream (`docs/delivery/epics/s11/epic-s11-day2-telegram-user-interaction-adapter-vision.md`, Issue `#447`).
- Результат Day 2 (факт):
  - зафиксированы mission, north star, persona outcomes и product principles для Telegram-канала как первого реального user-facing adapter path;
  - KPI/success metrics и guardrails оформлены для turnaround, fallback, delivery success, callback safety и purity platform semantics;
  - initial continuity issue `#444` сохранена только как historical handover artifact от intake-stage, 2026-03-14 закрыта как `state:superseded`, а active vision stage ведётся в Issue `#447`;
  - подтверждён и повторно зафиксирован sequencing gate: `#447` может двигаться дальше только пока `#389` остаётся closed и сохраняет `#387` как effective typed interaction contract baseline;
  - создана follow-up issue `#448` для stage `run:prd` без trigger-лейбла и с continuity-требованием продолжить цепочку до `run:dev`.
- Day 3 (in-review): PRD-package для Telegram-адаптера как первого внешнего channel-specific stream (`docs/delivery/epics/s11/epic-s11-day3-telegram-user-interaction-adapter-prd.md`, `docs/delivery/epics/s11/prd-s11-day3-telegram-user-interaction-adapter.md`, Issue `#448`).
- Результат Day 3 (факт):
  - зафиксированы user stories, FR/AC/NFR, wave priorities и expected evidence для `user.notify`, `user.decision.request`, inline callbacks и optional free-text;
  - product guardrails дополнены callback acknowledgement, duplicate/replay/expired handling, webhook authenticity expectations и fallback clarity без transport-first lock-in;
  - через Context7 по `/mymmrac/telego` и `go list -m -json github.com/mymmrac/telego@latest` на `2026-03-14` подтверждён latest stable baseline `v1.7.0`, а официальные Telegram Bot API facts (callback acknowledgement, webhook/polling exclusivity, update retention) зафиксированы как product-level constraints;
  - создана follow-up issue `#452` для stage `run:arch` без trigger-лейбла и с повторным continuity-требованием продолжить цепочку до `run:dev`.
- Day 4 (in-review): architecture package для Telegram-адаптера как первого внешнего channel-specific stream (`docs/delivery/epics/s11/epic-s11-day4-telegram-user-interaction-adapter-arch.md`, `docs/architecture/initiatives/s11_telegram_user_interaction_adapter/{README.md,architecture.md,c4_context.md,c4_container.md}`, Issue `#452`).
- Результат Day 4 (факт):
  - `control-plane` закреплён как owner interaction semantics, correlation, replay/expiry classification и operator-visible outcome; `worker` закреплён за delivery/retry/expiry и post-callback edit/follow-up continuation;
  - raw Telegram webhooks, secret-token verification и callback query acknowledgement оставлены во внешнем Telegram adapter contour, а `api-gateway` сохранён как thin callback bridge для normalized adapter callbacks;
  - callback payload direction зафиксирован как opaque/server-side lookup strategy, а message edit vs follow-up notify — как async platform-owned fallback policy под контролем `worker`;
  - подготовлены ADR-0014 и ALT-0006, а также follow-up issue `#454` для stage `run:design` без trigger-лейбла с continuity-требованием сохранить цепочку `design -> plan -> dev`.
- Day 5 (in-review): design package для Telegram-адаптера (`docs/delivery/epics/s11/epic-s11-day5-telegram-user-interaction-adapter-design.md`, `docs/architecture/initiatives/s11_telegram_user_interaction_adapter/{design_doc.md,api_contract.md,data_model.md,migrations_policy.md}`, Issue `#454`).
- Результат Day 5 (факт):
  - зафиксированы implementation-ready contracts для Telegram delivery/callback path, opaque callback handles и callback token grace;
  - выбрана additive data-model extension поверх Sprint S10 interaction foundation с `interaction_channel_bindings`, `interaction_callback_handles` и operator visibility state;
  - закреплён rollout order `S10 prerequisite -> migrations -> control-plane -> worker -> api-gateway -> Telegram adapter contour` и continuation policy `edit -> follow-up -> manual fallback`;
  - создана follow-up issue `#456` для stage `run:plan` без trigger-лейбла.
- Day 6 (in-review): plan package для Telegram-адаптера (`docs/delivery/epics/s11/epic-s11-day6-telegram-user-interaction-adapter-plan.md`, Issue `#456`).
- Результат Day 6 (факт):
  - execution package декомпозирован на waves `S11-E01..S11-E06` по schema foundation, domain/use-case, worker continuation, thin-edge bridge, Telegram adapter contour и observability/evidence gate;
  - создана follow-up issue `#458` как единый execution anchor для `run:dev` с явным continuity-требованием сохранить цепочку `#361 -> #447 -> #448 -> #452 -> #454 -> #456 -> #458` без разрывов;
  - зафиксированы quality-gates, DoR/DoD, blockers, risks и owner decisions для rollout order `migrations -> control-plane -> worker -> api-gateway -> Telegram adapter contour -> observability/evidence gate`;
  - сохранены platform-owned semantics, separation from approval flow и dependency gate на Sprint S10 interaction foundation.
- Day 7+ (in-review): `run:dev` выполнен в issue `#458` и перевёл Sprint S11 из doc-only baseline в implementation path:
  - `control-plane` получил additive schema foundation, callback handle/binding persistence, operator projections и typed Telegram delivery envelope;
  - `worker` получил HTTP bridge к внешнему Telegram adapter contour, transport-aware retry/fallback metadata и callback token propagation;
  - `api-gateway`/gRPC contracts и generated artifacts синхронно обновлены под normalized callback family `delivery_receipt|option_selected|free_text_received|transport_failure`;
  - follow-up issue `#473` закрывает оставшийся runtime gap: in-repo `telegram-interaction-adapter` materializes raw webhook/auth, callback acknowledgement, Bot API mediation и deploy wiring вместо внешнего placeholder bridge;
  - dev-итерация закрыта сервисными тестами/codegen и готова к handover `run:qa -> release -> postdeploy -> ops` после review.

### Sprint S12: GitHub API rate-limit resilience
- Day 1 (in-review): intake-пакет для GitHub API rate-limit resilience (`docs/delivery/epics/s12/epic-s12-day1-github-api-rate-limit-intake.md`, Issue `#366`).
- Результат Day 1 (факт):
  - инициатива зафиксирована как отдельный cross-cutting stream для GitHub-first rate-limit resilience, а не как локальный retry-bug в одном сервисе;
  - закреплены продуктовые инварианты: controlled wait-state вместо ложного failed, split `platform PAT` vs `agent bot-token`, owner/operator transparency и MCP backpressure на agent path;
  - зафиксировано ограничение: GitHub primary и secondary rate-limit semantics провайдер-управляемы и не сводятся к одному фиксированному countdown, поэтому UX должен опираться на typed recovery hints, а не на жёстко зашитый threshold;
  - создана follow-up issue `#413` для stage `run:vision` без trigger-лейбла.
- Day 2 (in-review): vision-пакет для GitHub API rate-limit resilience (`docs/delivery/epics/s12/epic-s12-day2-github-api-rate-limit-vision.md`, Issue `#413`).
- Результат Day 2 (факт):
  - инициатива зафиксирована как GitHub-first product capability вокруг controlled wait-state, а не как общий redesign quota-management или retry framework;
  - сформированы mission, north star, persona outcomes, KPI/guardrails и risk frame для owner/reviewer, operator и agent-path flows;
  - подтверждены MVP/Post-MVP границы: clarity, contour attribution, backpressure и safe resume входят в core wave, а notification/adapters и multi-provider governance остаются deferred;
  - создана follow-up issue `#416` для stage `run:prd` без trigger-лейбла.
- Day 3 (in-review): PRD-пакет для GitHub API rate-limit resilience (`docs/delivery/epics/s12/epic-s12-day3-github-api-rate-limit-prd.md`, `docs/delivery/epics/s12/prd-s12-day3-github-api-rate-limit-resilience.md`, Issue `#416`).
- Результат Day 3 (факт):
  - user stories, FR/AC/NFR и edge cases для controlled wait-state, rate-limit transparency и resume semantics;
  - продуктовый контракт для split `platform PAT` vs `agent bot-token`, provider-driven recovery hints, hard-failure separation и запрета infinite local retries;
  - проверочные evidence и wave priorities разделены между core MVP и deferred scope;
  - создана continuity issue `#418` для `run:arch` без trigger-лейбла.
- Day 4 (in-review): architecture package для GitHub API rate-limit resilience (`docs/delivery/epics/s12/epic-s12-day4-github-api-rate-limit-arch.md`, Issue `#418`).
- Результат Day 4 (факт):
  - architecture package зафиксировал ownership split для `control-plane`, `worker`, `agent-runner`, `api-gateway` и `web-console`, а также lifecycle `detect -> classify -> wait -> resume/manual action`;
  - `control-plane` выбран owner для classification, controlled wait aggregate, contour attribution и visibility contract, `worker` закреплён за finite auto-resume orchestration, а `agent-runner` переведён в handoff-only path без infinite local retries;
  - созданы initiative package `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/*`, `ADR-0013`, `ALT-0005` и follow-up issue `#420` для stage `run:design` без trigger-лейбла.
- Day 5 (in-review): design package для GitHub API rate-limit resilience (`docs/delivery/epics/s12/epic-s12-day5-github-api-rate-limit-design.md`, Issue `#420`).
- Результат Day 5 (факт):
  - зафиксированы typed contracts для signal handoff, dominant/related wait visibility, persisted wait aggregate/evidence и rollout/rollback notes;
  - выбран отдельный coarse wait-state `waiting_backpressure`, finite auto-resume policy для primary/secondary limits и best-effort GitHub service-comment mirror;
  - создана follow-up issue `#423` для stage `run:plan` без trigger-лейбла.
- Day 6 (in-review): plan-пакет для GitHub API rate-limit resilience (`docs/delivery/epics/s12/epic-s12-day6-github-api-rate-limit-plan.md`, Issue `#423`).
- Результат Day 6 (факт):
  - execution backlog декомпозирован на issues `#425..#431` без trigger-лейблов, с wave-sequencing и owner-managed handover в `run:dev`;
  - `#425` закреплён за schema foundation, `#426` за `control-plane` classification/projection, `#427` за worker auto-resume, `#428` за `agent-runner` handoff, `#429` за `api-gateway` transport, `#430` за `web-console` visibility, `#431` за observability/readiness gate;
  - зафиксирован sequencing order `#425 -> #426 -> #427 -> #428 -> #429 -> #430 -> #431` и rollout `migrations -> control-plane -> worker -> agent-runner -> api-gateway -> web-console -> evidence gate`;
  - документный контур `intake -> vision -> prd -> arch -> design -> plan` согласован и завершён, quality-gates/DoR/DoD/blockers/risks/owner decisions синхронизированы в delivery traceability, а predictive budgeting/multi-provider governance оставлены за пределами core Sprint S12 execution package;
- Day 7+ (in-review): `run:dev -> qa -> release -> postdeploy -> ops` по issues `#425..#431` с owner-managed wave launch и обязательным evidence gate `#431` перед `run:qa`.
  - Wave 1 / Issue `#425` переведён в `in-review`: добавлены schema foundation tables `github_rate_limit_waits` / `github_rate_limit_wait_evidence`, enum/check expansion для `agent_runs` / `agent_sessions`, postgres repository foundation и rollout guards для последующих волн.
  - Wave 2 / Issue `#426` переведён в `in-review`: `control-plane` получил canonical `GitHubRateLimitSignal` classification, typed wait projection/comment context, evidence append и deterministic agent resume payload builder для дальнейших волн `#427` / `#428` / `#429`.
  - Wave 3 / Issue `#427` переведён в `in-review`: `worker` получил due-wait sweep через новый `ProcessNextGitHubRateLimitWait` RPC, bounded replay/resume loop, manual escalation path и env/codegen wiring для дальнейшей волны `#428`.
  - Wave 4 / Issue `#428` переведён в `in-review`: `agent-runner` получил typed handoff `ReportGitHubRateLimitSignal`, coarse session snapshots `running -> waiting_backpressure`, dedicated resume payload lookup и stop-local-retry discipline для дальнейшей волны `#429`.
  - Wave 5 / Issue `#429` переведён в `in-review`: `api-gateway` получил contract-first wait visibility (`wait_projection`, dominant/related waits, typed realtime wait envelopes), codegen синхронизацию `OpenAPI+proto` и thin-edge mapping без доменной классификации в transport handlers.
  - Wave 6 / Issue `#430` переведён в `in-review`: `web-console` получил typed wait queue / run details surfaces для dominant/related waits, contour attribution, manual-action guidance и realtime wait activity без дублирования recovery/classification logic из `control-plane`.
  - Wave 7 / Issue `#431` переведён в `in-review`: подготовлен readiness bundle `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/observability_readiness.md` с rollout order, typed evidence surfaces, candidate `kubectl`/SQL diagnostics и rollback notes; текущий candidate namespace `codex-k8s-dev-1` подтверждает готовность deploy/job resources, но live rate-limit smoke пока не выполнялся, а `CODEXK8S_GITHUB_RATE_LIMIT_WAIT_ENABLED` остаётся на default-disabled wiring.

### Sprint S13: Quality governance system for agent-scale delivery (Issue #469)
- Day 1 (in-review): intake-пакет для `Quality Governance System` (`docs/delivery/epics/s13/epic-s13-day1-quality-governance-intake.md`, Issue `#469`).
- Результат Day 1 (факт):
  - зафиксирован отдельный governance stream для качества агентной поставки, а не локальный reviewer/process tweak;
  - сформирован draft quality stack: quality metrics baseline, risk tiers `low / medium / high / critical`, список high/critical changes, evidence taxonomy, verification minimum и review contract;
  - зафиксирована draft-связка `risk tier -> mandatory stages/gates -> required evidence`;
  - оформлена зависимость на downstream runtime/UI stream Sprint `S14` (`#470`): release safety, observability contract и quality cockpit не должны стартовать implementation-first без решений S13;
  - continuity rule закреплён как обязательный до `run:dev`: каждый doc-stage создаёт следующую follow-up issue без trigger-лейбла.
- Day 2 (in-review): vision-пакет для `Quality Governance System` (`docs/delivery/epics/s13/epic-s13-day2-quality-governance-vision.md`, Issue `#471`).
- Результат Day 2 (факт):
  - `Quality Governance System` закреплена как proportional governance capability: quality north star, persona outcomes, KPI/guardrails и product principles определены для owner/reviewer, delivery roles и platform operator;
  - сохранён sequencing gate `Sprint S13 governance baseline -> Sprint S14 runtime/UI stream` (`#470`) без implementation-first drift;
  - зафиксированы non-negotiables для следующих stage: explicit risk tier, separate constructs `evidence completeness / verification minimum / review-waiver discipline`, proportional governance и запрет silent waivers для `high/critical`;
  - создана follow-up issue `#476` для stage `run:prd` без trigger-лейбла.
- Day 3 (in-review): PRD-пакет для `Quality Governance System` (`docs/delivery/epics/s13/epic-s13-day3-quality-governance-prd.md`, `docs/delivery/epics/s13/prd-s13-day3-quality-governance-system.md`, Issue `#476`).
- Результат Day 3 (факт):
  - user stories, FR/AC/NFR и edge cases зафиксированы для explicit risk tiering, mandatory evidence package, verification minimum, review/waiver discipline и governance-gap feedback loop;
  - product contract закрепил proportional low-risk path, запрет silent waivers для `high/critical`, role-specific decision surfaces и boundary `Sprint S13 governance baseline -> Sprint S14 runtime/UI stream`;
  - publication policy закрепила путь `internal working draft -> semantic wave map -> published waves`; raw draft не считается review/merge artifact;
  - expected evidence и wave priorities разделены между core governance baseline и deferred runtime/UI automation scope, а large PR допустим только для behaviour-neutral mechanical bounded-scope changes;
  - создана continuity issue `#484` для `run:arch` без trigger-лейбла.
- Day 4 (in-review): architecture package для `Quality Governance System` (`docs/delivery/epics/s13/epic-s13-day4-quality-governance-arch.md`, Issue `#484`).
- Результат Day 4 (факт):
  - `control-plane` закреплён как owner canonical aggregate, publication gate, decision surface и projection refresh path;
  - `worker` закреплён как owner asynchronous reconciliation и governance-gap sweeps под policy `control-plane`;
  - publication path `internal working draft -> semantic wave map -> published waves` переведён из product baseline в architecture baseline;
  - создана continuity issue `#494` для `run:design` без trigger-лейбла.
- Day 5 (in-review): design package для `Quality Governance System` (`docs/delivery/epics/s13/epic-s13-day5-quality-governance-design.md`, Issue `#494`).
- Результат Day 5 (факт):
  - зафиксированы typed contracts для hidden draft handoff, semantic wave map, evidence blocks, decision ledger, governance feedback и projection families;
  - bounded historical backfill сохранён как отдельный execution concern без фабрикации hidden drafts, waivers или release decisions;
  - rollout order `migrations -> control-plane -> worker -> api-gateway -> web-console` закреплён как обязательный baseline для `run:dev`;
  - создана continuity issue `#512` для `run:plan` без trigger-лейбла.
- Day 6 (in-review): plan package для `Quality Governance System` (`docs/delivery/epics/s13/epic-s13-day6-quality-governance-plan.md`, Issue `#512`).
- Результат Day 6 (факт):
  - execution package разложен на issues `#521..#525` с sequencing `foundation -> worker feedback/backfill -> transport/mirror -> web-console -> readiness gate`;
  - зафиксированы quality-gates, DoR/DoD и owner-managed launch policy: `run:dev` triggers ставятся только по wave-sequencing;
  - сохранены non-negotiables Sprint S13: hidden draft internal-only, `semantic wave map` mandatory, no silent waivers for `high/critical`, `worker` without canonical semantics;
  - boundary `Sprint S13 -> Sprint S14` удержан: runtime/UI invention остаётся вне Day6 execution package.
- Day 7 (planned): owner-managed `run:dev` execution waves через issues `#521..#525`.
  - Цель: поэтапно реализовать foundation, worker lifecycle, transport/mirror, UI visibility и readiness evidence без нарушения design guardrails.
  - Ожидаемый результат: PR-потоки по waves с обязательным traceability sync и переходом в `state:in-review`.

### Sprint S16: Mission Control graph workspace and continuity control plane (Issue #492)
- Day 1 (in-review): intake-пакет для полного redesign Mission Control (`docs/delivery/epics/s16/epic-s16-day1-mission-control-graph-workspace-intake.md`, Issue `#492`).
- Результат Day 1 (факт):
  - Mission Control зафиксирован как новый graph workspace/control plane, а не как incremental refresh Sprint S9 dashboard;
  - issue `#480` поглощена как mandatory foundation layer для persisted GitHub inventory mirror, bounded reconcile, coverage contract `all open Issues/PR + bounded recent closed history` и future hybrid warmup/lazy repair;
  - hybrid truth matrix, filtered multi-root workspace с точными Wave 1 filters `open_only`, `assigned_to_me_or_unassigned`, active-state presets, left-to-right graph baseline и Wave 1 nodes `discussion/work_item/run/pull_request` закреплены как intake baseline;
  - typed metadata/watermark contract, platform-canonical launch params и continuity rule `PR + linked follow-up issue` через `run:dev` закреплены как обязательный product contract;
  - создана continuity issue `#496` для stage `run:vision` без trigger-лейбла.
- Day 2 (in-review): vision-пакет для Mission Control graph workspace (`docs/delivery/epics/s16/epic-s16-day2-mission-control-graph-workspace-vision.md`, Issue `#496`).
- Результат Day 2 (факт):
  - Mission Control закреплён как primary multi-root graph workspace/control plane для continuity `discussion/work_item -> run -> pull_request/follow-up issue -> next run`, а не как board/list-first refresh;
  - mission, north star, persona outcomes, KPI/guardrails и wave boundaries определены без reopening Day1 baseline;
  - inventory foundation из `#480`, exact Wave 1 filters `open_only`, `assigned_to_me_or_unassigned`, `active-state presets`, secondary/dimmed semantics только для graph integrity, Wave 1 nodes `discussion/work_item/run/pull_request`, typed metadata/watermarks, platform-canonical launch params и continuity rule `PR + follow-up issue` повторно зафиксированы как non-negotiable;
  - voice/STT, dashboard orchestrator agent и отдельная `agent` node taxonomy оставлены в later-wave scope и не блокируют core Wave 1;
  - создана follow-up issue `#510` для stage `run:prd` без trigger-лейбла.
- Day 3 (in-review): PRD package для Mission Control graph workspace (`docs/delivery/epics/s16/epic-s16-day3-mission-control-graph-workspace-prd.md`, `docs/delivery/epics/s16/prd-s16-day3-mission-control-graph-workspace.md`, Issue `#510`).
- Результат Day 3 (факт):
  - user stories, FR/AC/NFR, scenario matrix и expected evidence зафиксированы для fullscreen graph workspace, filtered multi-root continuity, inventory-backed foundation, typed metadata/watermarks, platform-canonical launch params и platform-safe inline actions;
  - explicit continuity contract сохранён: stage through `run:dev` считается complete только при наличии `PR + linked follow-up issue`, а later-wave contours voice/STT, dashboard orchestrator agent, отдельная `agent` node taxonomy, full-history/archive и richer provider enrichment не блокируют core Wave 1;
  - создана continuity issue `#516` для stage `run:arch` без trigger-лейбла.
- Day 4 (in-review): architecture issue `#516` (`docs/delivery/epics/s16/epic-s16-day4-mission-control-graph-workspace-arch.md`, `docs/architecture/initiatives/s16_mission_control_graph_workspace/{README.md,architecture.md,c4_context.md,c4_container.md}`, `docs/architecture/adr/ADR-0016-mission-control-graph-workspace-hybrid-truth-and-continuity-ownership.md`, `docs/architecture/alternatives/ALT-0008-mission-control-graph-workspace-hybrid-truth-boundaries.md`).
- Результат Day 4 (факт):
  - `control-plane` закреплён как owner canonical graph truth, continuity state, typed metadata/watermarks и launch surfaces;
  - `worker` закреплён за bounded provider inventory freshness, recent-closed-history backfill и reconcile execution без ownership graph semantics;
  - hybrid truth lifecycle `provider mirror -> graph truth -> workspace projection` и boundary `thin-edge UI/API` сохранены без reopening Sprint S16 baseline;
  - создана continuity issue `#519` для stage `run:design` без trigger-лейбла.
- Day 5 (in-review): design package для Mission Control graph workspace (`docs/delivery/epics/s16/epic-s16-day5-mission-control-graph-workspace-design.md`, `docs/architecture/initiatives/s16_mission_control_graph_workspace/{design_doc.md,api_contract.md,data_model.md,migrations_policy.md}`, Issue `#519`).
- Результат Day 5 (факт):
  - зафиксирован graph-first interaction model поверх existing Mission Control bounded context без отдельного сервиса и без второго command path;
  - typed transport baseline определён как `workspace -> node details -> activity -> launch preview -> existing command ledger`, а Sprint S9 dashboard contract переводится в superseded state без отдельного long-lived namespace;
  - persisted continuity gaps и workspace watermarks оформлены как отдельные domain constructs `control-plane`, а `run` node закреплён как Wave 1 canvas kind вместо `agent`;
  - rollout path зафиксирован как `expand schema -> shadow backfill -> read switch -> preview exposure -> cleanup last`, сохранив order `migrations -> control-plane -> worker -> api-gateway -> web-console`;
  - через `gh issue create` создана continuity issue `#537` для stage `run:plan` без trigger-лейбла.
- Day 6 (in-review): plan package для Mission Control graph workspace (`docs/delivery/epics/s16/epic-s16-day6-mission-control-graph-workspace-plan.md`, Issue `#537`).
- Результат Day 6 (факт):
  - execution package разложен на issues `#542..#547` с sequencing `schema/backfill foundation -> control-plane graph truth -> worker reconcile/freshness -> transport/preview -> web-console graph workspace -> readiness gate`;
  - зафиксированы quality-gates, DoR/DoD и owner-managed launch policy: `run:dev` triggers ставятся только по wave-sequencing;
  - сохранены non-negotiables Sprint S16: issue `#480`, exact Wave 1 filters/nodes, secondary/dimmed only for graph integrity, read-only launch preview и запрет на новый deployable сервис;
  - удержан boundary относительно Sprint S9: dashboard-first model не возвращается, voice/STT и richer provider enrichment остаются вне Day6 execution package.
- Day 7 (planned): owner-managed `run:dev` execution waves через issues `#542..#547`.
  - Цель: поэтапно реализовать foundation, graph truth, reconcile, transport, UI visibility и readiness evidence без нарушения design guardrails.
  - Ожидаемый результат: PR-потоки по waves с обязательным traceability sync и переходом в `state:in-review`.

### Sprint S17: Unified long-lived user interaction waits and owner feedback inbox (Issue #541)
- Day 1 (in-review): intake package для unified owner feedback loop (`docs/delivery/epics/s17/epic-s17-day1-unified-user-interaction-waits-and-owner-feedback-inbox-intake.md`, Issue `#541`).
- Результат Day 1 (факт):
  - зафиксирован отдельный cross-cutting product stream вокруг long-lived human-wait contract, а не локальный Telegram/runtime bugfix;
  - сравнение execution models закрепило recommended baseline: same live pod / same `codex` session как happy-path, snapshot-resume только как recovery fallback;
  - long human-wait target `>=24h`, lifecycle `created -> delivery pending -> delivery accepted -> waiting -> response -> continuation`, Telegram pending inbox и staff-console fallback закреплены как intake baseline;
  - persisted text/voice binding и deterministic continuation после inline/text/voice reply включены в core Wave 1;
  - `run:self-improve` явно выведен из human-wait contract;
  - создана continuity issue `#554` для stage `run:vision` без trigger-лейбла.
- Day 2 (in-review): vision package для owner feedback loop (`docs/delivery/epics/s17/epic-s17-day2-unified-user-interaction-waits-and-owner-feedback-inbox-vision.md`, Issue `#554`).
- Результат Day 2 (факт):
  - unified owner feedback loop закреплён как platform capability: owner отвечает в Telegram или staff-console, видит pending request и получает детерминированное продолжение той же задачи без GitHub-comment detour;
  - mission, north star, persona outcomes, KPI/guardrails и wave boundaries определены для owner/product lead path, same-session runtime path и staff/operator fallback path;
  - повторно зафиксирован locked baseline: same live pod / same `codex` session как primary happy-path, snapshot-resume как recovery-only fallback, long human-wait target `>=24h`, delivery-before-wait lifecycle, Telegram pending inbox, staff-console fallback, deterministic text/voice binding и `run:self-improve` exclusion;
  - отдельно закреплён product guardrail: built-in `codex_k8s` MCP wait path обязан иметь максимальный timeout/TTL не ниже owner wait window, чтобы happy-path оставался реальным live wait, а synthetic resume с подложенным tool result не нормализовался как основная модель;
  - дополнительные каналы, reminders/escalations, attachments, multi-party routing, richer conversation UX и detached resume-run как равноправный happy-path оставлены в later-wave scope и не блокируют core MVP;
  - создана follow-up issue `#557` для stage `run:prd` без trigger-лейбла.
- Day 3 (planned): PRD package для owner feedback loop (Issue `#557`).
  - Цель: формализовать user stories, FR/AC/NFR, scenario matrix и expected evidence для same-session continuity, dual-channel inbox, lifecycle transparency и max timeout/TTL baseline built-in `codex_k8s` MCP wait path.
  - Ожидаемый результат: follow-up issue для `run:arch` и continuity-требование сохранить цепочку `arch -> design -> plan -> dev` без разрывов.

### Daily delivery contract (обязательный)
- Каждый день задачи дня влиты в `main`.
- Каждый день изменения автоматически задеплоены на production.
- Каждый день выполнен ручной smoke-check.
- Каждый день актуализированы документы при изменениях API/data model/webhook/RBAC.
- Для каждого эпика заполнен `Data model impact` по структуре `docs/templates/data_model.md`.
- Правила спринт-процесса и ownership артефактов выполняются по `docs/delivery/development_process_requirements.md`.

## Зависимости
- Внутренние: Core backend до полноценного UI управления.
- Внешние: GitHub fine-grained token с нужными правами, рабочий production сервер Ubuntu 24.04.

## План сред/окружений
- Dev slots: локальный/кластерный dev для компонентов.
- Production: обязателен до расширения функционала.
- Prod: после стабилизации production и security review.

## Специальный этап bootstrap production (обязательный)

Цель этапа: когда уже есть что тестировать вручную, запускать один скрипт с машины разработчика и автоматически поднимать production окружение.

Ожидаемое поведение скрипта:
- запускается на машине разработчика (текущей) и подключается по SSH к серверу как `root`;
- создаёт отдельного пользователя (sudo + ssh key auth), отключает дальнейший root-password вход;
- ставит k3s и сетевой baseline (ingress, cert-manager, network policy baseline);
- ставит зависимости платформы;
- поднимает внутренний registry (`ClusterIP`, без auth на уровне registry) и Kaniko pipeline для сборки образа в кластере;
- разворачивает PostgreSQL и `codex-k8s`;
- спрашивает внешние креды (`GitHub fine-grained token`, `CODEXK8S_OPENAI_API_KEY`), внутренние секреты генерирует сам;
- передаёт default `learning_mode` из `bootstrap/host/config.env` (по умолчанию включён, пустое значение = выключен);
- настраивает GitHub webhook/labels через API без GitHub Actions runner и хранит runtime config/secrets только в Kubernetes;
- запускает self-deploy через control-plane runtime deploy job (build/mirror/apply/cleanup).

## Чек-листы готовности
### Definition of Ready (DoR)
- [ ] Brief/Constraints/Architecture/ADR согласованы.
- [ ] Server access для production подтверждён.
- [ ] GitHub fine-grained token и OpenAI ключ доступны.

### Definition of Done (DoD)
- [x] Day 0 baseline bootstrap выполнен.
- [ ] Для активного спринта: каждый эпик закрыт по своим acceptance criteria.
- [ ] Для активного спринта: ежедневный merge -> auto deploy -> smoke check выполнен.
- [ ] Webhook -> run -> worker -> k8s -> UI цепочка проходит regression.
- [ ] Для `full-env` подтверждены role-based TTL retention namespace и lease extension на `run:*:revise` (Issue #74).
- [x] Для Issue #100 зафиксирован delivery execution-plan Sprint S4 (federated composition + multi-repo docs federation) и подготовлен handover в `run:dev`.
- [ ] Learning mode и self-improve mode проверены на production.
- [ ] MCP governance tools (secret/db/feedback) прошли approve/deny regression.

## Риски и буферы
- Риск: нестабильная сеть/доступы при bootstrap.
- Буфер: fallback runbook ручной установки.

## План релиза (верхний уровень)
- Релизные окна:
  - production continuous (auto deploy on push to `main`);
  - production gated (manual dispatch + environment approval).
- Rollback: возвращение на предыдущий контейнерный тег + DB migration rollback policy.

## Решения Owner
- Runner scale policy утверждена:
  - локальные запуски — один persistent runner;
  - серверные окружения с доменом — autoscaled set.
- Storage policy утверждена: на MVP используем `local-path`, Longhorn переносим на следующий этап.
- Read replica policy утверждена: минимум одна async streaming replica на MVP, далее эволюция до 2+ и sync/quorum без изменений приложения.

## Апрув
- request_id: owner-2026-02-06-mvp
- Решение: approved
- Комментарий: План поставки и условия bootstrap/production утверждены.
