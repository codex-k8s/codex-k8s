---
doc_id: EPC-CK8S-S16-D2-MISSION-CONTROL-GRAPH
type: epic
title: "Epic S16 Day 2: Vision для Mission Control graph workspace и continuity control plane (Issues #496/#510)"
status: superseded
owner_role: PM
created_at: 2026-03-15
updated_at: 2026-03-25
related_issues: [480, 490, 492, 496, 510, 561, 562, 563]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-15-issue-496-vision"
---

# Epic S16 Day 2: Vision для Mission Control graph workspace и continuity control plane (Issues #496/#510)

## TL;DR
- 2026-03-25 issue `#561` перевела этот vision-package в historical superseded state.
- Зафиксированная здесь vision-модель больше не описывает текущий Mission Control baseline: вместо неё действует reset path `#562` -> `#563`.
- Больше не считать source of truth: lane/column shell, taxonomy `discussion/work_item/run/pull_request` и старую S16 continuity модель.
- Остальной текст файла сохраняется только как historical evidence отклонённого vision baseline.

## Priority
- `P0`.

## Vision charter

### Mission statement
Сделать Mission Control основным graph workspace/control plane платформы, чтобы Owner/product lead, delivery operator и discussion-driven пользователь могли вести несколько инициатив одновременно, видеть полный lineage `discussion/work_item -> run -> PR/follow-up issue -> next run`, понимать безопасный следующий шаг за минуты и при этом не терять GitHub-native review/merge semantics, inventory-backed coverage и audit-safe platform policy.

### Цели и ожидаемые результаты
1. Превратить Mission Control из board/list-экрана над частичной platform evidence в multi-root graph workspace, который действительно становится primary working surface для active delivery.
2. Зафиксировать continuity-first модель: stage artifacts, runs, PR и follow-up issues должны читаться как единая graph-цепочка, а не как набор разрозненных сущностей в разных экранах.
3. Удержать явную hybrid truth matrix: provider inventory из `#480` обязателен как foundation layer, а platform state отвечает за operational graph, launch params, metadata, watermarks и enrichment.
4. Сохранить фокус первой волны на core workspace и time-to-value: voice/STT, dashboard orchestrator agent, richer provider enrichment и отдельная `agent` node taxonomy не должны размывать core Wave 1.

### Пользователи и стейкхолдеры
- Основные пользователи:
  - Owner / product lead, которому нужен быстрый situational awareness по 2-3 инициативам сразу, понятный next step и видимость continuity-цепочки до PR/follow-up issue.
  - Delivery operator / engineer / manager, которому нужен единый control plane для stage progression, run context, linked artifacts и launch configuration без ручного поиска по GitHub.
  - Discussion-driven пользователь, который может начать с discussion, но также должен иметь путь создать любой stage-issue, связать его с родителем и продолжить работу из graph workspace.
- Стейкхолдеры:
  - `services/staff/web-console` как primary UX-контур нового workspace;
  - `services/internal/control-plane` как будущий owner graph truth, continuity state и launch surfaces;
  - `services/jobs/worker` как bounded reconcile/background contour для inventory mirror freshness, enrichment pipelines и lifecycle tasks под policy `control-plane`;
  - `services/external/api-gateway` как thin-edge для transport/UI commands и webhook-driven provider sync;
  - foundation stream `#480`, который поставляет persisted inventory mirror и bounded reconcile baseline;
  - existing baselines Sprint S9 и Sprint S12, которые дают typed projection/runtime groundwork и rate-limit/backpressure constraints.
- Владелец решения: Owner.

### Продуктовые принципы и ограничения
- Workspace first: Mission Control должен быть основным operational workspace, а не декоративным summary-слоем над GitHub.
- Continuity graph важнее isolated card views: граф должен помогать понять lineage и следующий шаг, а не только показывать текущий snapshot.
- Hybrid truth matrix остаётся explicit и typed: platform canonical для graph/metadata/launch/watermarks, GitHub canonical для provider-native issue/pr/comment/review/development state.
- Service-boundary guardrail сохраняется уже на vision stage: `control-plane` остаётся owner graph truth, continuity и launch policy surfaces, а `worker` ограничен background/reconcile execution для provider mirror и lifecycle tasks.
- Inventory-backed provider mirror из `#480` обязателен; live-fetch-only UI и размытый backlog mirror не считаются допустимым shortcut.
- GitHub остаётся местом human review, merge и provider-native collaboration; dashboard не подменяет provider semantics.
- Metadata, summaries, watermarks и launch params должны быть platform-canonical и typed; markdown scraping и LLM-generated watermarks не допускаются как source of truth.
- Voice/STT и dashboard orchestrator agent остаются отдельной later-wave траекторией и не блокируют core Wave 1.
- В рамках `run:vision` разрешены только markdown-изменения.

## Scope boundaries

### MVP scope
- Fullscreen graph workspace shell:
  - detached top toolbar;
  - right drawer/chat/details;
  - graph-first workspace вместо board/list-first landing.
- Filtered multi-root workspace:
  - default filters Wave 1 = `open_only`, `assigned_to_me_or_unassigned`, `active-state presets`;
  - в одном срезе видны несколько инициатив одновременно;
  - graph layout остаётся left-to-right по каждой инициативе;
  - secondary/dimmed semantics используются только для graph integrity.
- Wave 1 node taxonomy:
  - `discussion`;
  - `work_item`;
  - `run`;
  - `pull_request`.
- Hybrid truth matrix и inventory-backed foundation:
  - provider inventory из `#480` = source layer для `all open Issues/PR + bounded recent closed history`;
  - platform state = enrichment layer для relations, run lineage, launch params, metadata, watermarks, sync status и next-step context.
- Typed workspace surfaces:
  - node title and summary;
  - artifact intent and relation hints;
  - platform-generated watermarks;
  - platform-canonical launch params (`model`, `reasoning`, runtime/access profile, defaults + overrides).
- Platform-safe inline actions и continuity surfaces:
  - открыть drawer/details/chat;
  - подготовить/запустить следующий stage;
  - видеть PR/follow-up issue как обязательные downstream artifacts;
  - сохранять continuity rule `PR + follow-up issue` для каждого stage через `run:dev`.
- Handover в `run:prd` через issue `#510`.

### Post-MVP / deferred scope
- Отдельная `agent` node taxonomy на canvas.
- Dashboard orchestrator agent и voice/STT path.
- GitHub Projects / Issue Type / Relationships как primary graph model.
- Full-history archive и “весь мир” как default canvas scope.
- Richer provider enrichment beyond agreed Wave 1 baseline.
- Произвольный graph builder/layout designer и non-essential visual experimentation.

### Sequencing gate
- Day2 vision не переоткрывает intake decisions из `#492`; PRD stage наследует их как locked baseline.
- Foundation stream `#480` не заменяется и не размывается: persisted provider mirror и bounded reconcile остаются обязательным входом для следующего stage.
- Existing baselines Sprint S9 и Sprint S12 считаются prerequisite execution context:
  - Sprint S9 даёт первый Mission Control UX/runtime baseline;
  - Sprint S12 задаёт rate-limit/backpressure discipline для provider inventory и reconcile.
- Если scope начинает включать voice/STT, отдельные `agent` nodes или GitHub-first single-root framing как core AC, stage должен быть остановлен до нового owner-решения.

## Success metrics

### North Star
| ID | Метрика | Определение | Источник | Целевое значение |
|---|---|---|---|---|
| `NSM-496-01` | Continuity workspace adoption rate | Доля pilot-сессий Mission Control, в которых первый continuity action по активной инициативе (`open drawer`, `launch next stage`, `open linked PR/follow-up issue`, `inspect run context`) совершается из graph workspace, а у самой инициативы видна связная цепочка от root до последнего downstream artifact без ручного GitHub label hunting | frontend analytics + `flow_events` + command audit | `>= 70%` в течение 30 дней после pilot rollout |

### Supporting metrics
| ID | Метрика | Определение/формула | Источник | Цель |
|---|---|---|---|---|
| `PM-496-01` | Next-step clarity p75 | p75 времени от открытия Mission Control до первого continuity-focused действия (`open drawer`, `launch stage`, `open linked PR`, `open follow-up issue`) | frontend analytics + action audit | `<= 60 секунд` |
| `OPS-496-01` | Inventory-backed active coverage | Доля in-scope provider сущностей, которые попали в persisted workspace projection в рамках policy `all open Issues/PR + bounded recent closed history` | provider mirror audit + reconcile evidence | `100%` для open scope и `100%` для configured bounded recent window |
| `REL-496-01` | Hybrid truth merge correctness | Доля graph entities, где provider state и platform enrichment объединены без duplicate nodes, missing lineage или ambiguous ownership classification | projection audit + incident review | `>= 99.5%` |
| `UX-496-01` | Multi-initiative usability rate | Доля pilot-сценариев с 2+ одновременными инициативами, где пользователь удерживает их в одном workspace и находит next step без отката к single-root-only view или ручному provider search | UX pilot evidence + product review | `>= 80%` |
| `GOV-496-01` | Continuity artifact completeness | Доля stage transitions этой инициативы до `run:dev` включительно, завершившихся артефактом stage и linked follow-up issue по policy `PR + follow-up issue` | `flow_events` + issue/PR traceability audit | `100%` |

### Guardrails (ранние сигналы)
- `GR-496-01`: если `REL-496-01 < 99%`, следующий stage обязан приоритизировать truth matrix, linking policy и reconcile correctness выше новых UX-волн.
- `GR-496-02`: если `OPS-496-01` падает ниже agreed scope policy, Sprint S16 не может двигаться в implementation-ready design, пока inventory foundation не станет надёжной.
- `GR-496-03`: если `PM-496-01 > 120 секунд` или `UX-496-01 < 70%`, PRD/design обязаны сначала упростить graph density, drawer flows и next-step surfaces, а не расширять node taxonomy.
- `GR-496-04`: если `GOV-496-01 < 100%`, stage progression блокируется до восстановления continuity rule `PR + follow-up issue`.
- `GR-496-05`: если core AC начинают требовать voice/STT, отдельные `agent` nodes или GitHub-first single-root framing, stage переводится в `need:input` до нового owner-решения.

## Risks and Product Assumptions
| Тип | ID | Описание | Митигирующее действие | Статус |
|---|---|---|---|---|
| risk | `RSK-496-01` | Scope может расползтись одновременно в graph UX, provider mirror, voice/orchestrator path, metadata redesign и новую agent taxonomy | Держать core Wave 1 жёстко ограниченным и проверять каждый следующий stage по explicit deferred-scope boundary | open |
| risk | `RSK-496-02` | Multi-root graph без жёстких filters и density rules превратится в визуальный шум | Сохранить exact filter baseline, secondary/dimmed semantics только для graph integrity и list/drawer fallback как обязательные guardrails | open |
| risk | `RSK-496-03` | Platform state, provider inventory и runtime lineage начнут расходиться без явной ownership matrix | Удерживать hybrid truth matrix как non-negotiable product decision и тащить её дальше в PRD/Arch | open |
| risk | `RSK-496-04` | Typed metadata/watermark/launch surfaces могут деградировать в markdown scraping и ad-hoc heuristics | Зафиксировать typed/platform-canonical contract уже на vision stage и проверить его в PRD/design | open |
| assumption | `ASM-496-01` | Existing `control-plane` / `worker` / `api-gateway` / `web-console` baselines можно эволюционно расширить до graph workspace без немедленного выделения нового deployable сервиса | Проверить ownership split на `run:arch` | accepted |
| assumption | `ASM-496-02` | Foundation stream `#480` достаточен, чтобы дать policy-bound provider inventory для core Wave 1 без превращения инициативы в GitHub data warehouse | Подтвердить scope и bounded policies на `run:prd` / `run:arch` | accepted |
| assumption | `ASM-496-03` | Пользовательская ценность workspace достигается уже на core Wave 1, даже если human review/merge остаются в GitHub UI | Сохранить provider-native collaboration outside dashboard и подтвердить pilot metrics на `run:prd` | accepted |

## Readiness criteria для `run:prd`
- [x] Mission, north star и persona outcomes сформулированы для Owner/product lead, delivery operator и discussion-driven workflow.
- [x] KPI/success metrics и guardrails определены как измеримые product/operational сигналы.
- [x] Locked Day1 baseline сохранён явно: filters `open_only`, `assigned_to_me_or_unassigned`, `active-state presets`, secondary/dimmed semantics only for graph integrity, Wave 1 nodes `discussion/work_item/run/pull_request`, hybrid truth matrix, inventory foundation `#480`, typed metadata/watermarks, platform-canonical launch params и continuity rule `PR + follow-up issue`.
- [x] Core Wave 1 и later-wave scope разделены явно; voice/STT, dashboard orchestrator agent и отдельная `agent` node taxonomy не входят в blocking scope.
- [x] Создана отдельная issue следующего этапа `run:prd` (`#510`) без trigger-лейбла.

## Acceptance criteria (Issue #496)
- [x] Mission, north star и продуктовые принципы для Mission Control graph workspace сформулированы явно.
- [x] KPI/success metrics и guardrails зафиксированы для next-step clarity, inventory-backed coverage, hybrid truth merge correctness, multi-initiative usability и continuity completeness.
- [x] Персоны, MVP/Post-MVP границы, риски и assumptions описаны без reopening intake baseline и без смешения с implementation details.
- [x] Явно сохранены как обязательный baseline: Wave 1 filters `open_only`, `assigned_to_me_or_unassigned`, `active-state presets`, secondary/dimmed semantics только для graph integrity, Wave 1 nodes `discussion/work_item/run/pull_request`, hybrid truth matrix, inventory foundation из `#480`, typed metadata/watermarks, platform-canonical launch params и continuity rule `PR + follow-up issue`.
- [x] Подготовлен handover в `run:prd` и создана follow-up issue `#510` без trigger-лейбла.

## Handover в следующий этап
- Следующий stage: `run:prd`.
- Follow-up issue: `#510`.
- Trigger-лейбл `run:prd` на issue `#510` ставит Owner.
- Обязательное continuity-требование для `#510`:
  - в конце PRD stage агент обязан создать issue для `run:arch` без trigger-лейбла;
  - в body этой issue нужно явно повторить требование продолжить цепочку `arch -> design -> plan -> dev` без разрывов.
- На `run:prd` нельзя потерять следующие решения vision:
  - Mission Control остаётся primary multi-root graph workspace/control plane, а не board/list refresh;
  - inventory foundation из `#480` и coverage contract `all open Issues/PR + bounded recent closed history` обязательны;
  - hybrid truth matrix остаётся typed и explicit;
  - Wave 1 filters = `open_only`, `assigned_to_me_or_unassigned`, `active-state presets`;
  - secondary/dimmed semantics используются только для graph integrity;
  - Wave 1 nodes = `discussion/work_item/run/pull_request`, без отдельной `agent` node taxonomy;
  - metadata/watermarks/launch params остаются platform-canonical typed surfaces;
  - human review/merge/provider-native collaboration остаются в GitHub UI;
  - voice/STT и dashboard orchestrator agent остаются later-wave scope;
  - stage continuity до `run:dev` = `PR + linked follow-up issue`.

## Связанные документы
- `docs/delivery/sprints/s16/sprint_s16_mission_control_graph_workspace.md`
- `docs/delivery/epics/s16/epic_s16.md`
- `docs/delivery/epics/s16/epic-s16-day1-mission-control-graph-workspace-intake.md`
- `docs/delivery/traceability/s16_mission_control_graph_workspace_history.md`
- `docs/delivery/issue_map.md`
- `docs/delivery/delivery_plan.md`
- `docs/product/requirements_machine_driven.md`
- `docs/product/agents_operating_model.md`
- `docs/product/labels_and_trigger_policy.md`
- `docs/product/stage_process_model.md`
- `docs/delivery/development_process_requirements.md`
- `docs/research/src_idea-machine_driven_company_requirements.md`
- `docs/delivery/traceability/s9_mission_control_dashboard_history.md`
- `docs/delivery/traceability/s12_github_api_rate_limit_resilience_history.md`
- `services/internal/control-plane/README.md`
- `services/jobs/worker/README.md`
- `services/external/api-gateway/README.md`
- `services/staff/web-console/README.md`
