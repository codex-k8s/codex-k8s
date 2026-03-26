---
doc_id: PRD-CK8S-S16-D3-MISSION-CONTROL-GRAPH
type: prd
title: "Mission Control graph workspace and continuity control plane — PRD Sprint S16 Day 3"
status: superseded
owner_role: PM
created_at: 2026-03-16
updated_at: 2026-03-25
related_issues: [480, 490, 492, 496, 510, 516, 561, 562, 563]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-16-issue-510-prd"
---

# PRD: Mission Control graph workspace and continuity control plane

## TL;DR
- 2026-03-25 issue `#561` пометила этот PRD как historical superseded artifact.
- Зафиксированный здесь MVP больше не является текущим Mission Control baseline: его заменяет reset path `#562` (frontend-first UX) и `#563` (backend rebuild).
- Больше не считать source of truth taxonomy `discussion/work_item/run/pull_request`, lane/column shell и старую freshness/continuity модель Sprint S16.
- Остальной текст файла сохраняется как historical evidence отклонённого product contract.

## Проблема и цель
- Problem statement:
  - текущий Mission Control остаётся board/list-first слоем над частичной platform evidence и не стал основным рабочим местом Owner для нескольких инициатив одновременно;
  - provider inventory и platform enrichment не сведены в явную hybrid truth matrix, поэтому пользователь не доверяет completeness/freshness workspace и продолжает вручную проверять GitHub;
  - continuity между `discussion/work_item`, `run`, `pull_request`, follow-up issue и следующим запуском не выражена как единая graph-модель;
  - без чёткого PRD команда легко смешает core graph workspace, inventory foundation, voice/orchestrator path и future enrichments в один размытый scope.
- Цели:
  - закрепить Mission Control как primary multi-root graph workspace/control plane;
  - формализовать trusted hybrid truth matrix между provider inventory и platform state;
  - зафиксировать typed metadata/watermarks и platform-canonical launch params как обязательные product surfaces;
  - сделать continuity rule `PR + linked follow-up issue` частью продукта, а не внутренней привычкой процесса.
- Почему сейчас:
  - intake `#492` и vision `#496` уже зафиксировали product contour и locked baseline;
  - без PRD следующий stage начнёт спорить об архитектуре без проверяемого продукта;
  - Sprint S16 идёт по launch profile `new-service`, где `run:prd` обязателен перед `run:arch`.

## Зафиксированные продуктовые решения
- `D-510-01`: Mission Control остаётся primary multi-root graph workspace/control plane, а не board/list-only refresh Sprint S9.
- `D-510-02`: issue `#480` является обязательным foundation layer; coverage contract provider mirror = `all open Issues/PR + bounded recent closed history`.
- `D-510-03`: exact Wave 1 filters `open_only`, `assigned_to_me_or_unassigned`, `active-state presets` остаются canonical default baseline.
- `D-510-04`: secondary/dimmed semantics используются только для graph integrity и не заменяют primary filter baseline.
- `D-510-05`: Wave 1 canvas nodes ограничены `discussion`, `work_item`, `run`, `pull_request`; comments/chat/summaries/approvals живут в drawer/timeline, а `agent` не становится core node taxonomy.
- `D-510-06`: platform canonical для relations, run lineage, launch params, metadata, sync state и watermarks; GitHub canonical для issue/pr/comment/review/development state.
- `D-510-07`: metadata/watermarks/launch params остаются typed и platform-canonical; markdown scraping и LLM-generated watermarks не допускаются как source of truth.
- `D-510-08`: human review, merge и provider-native collaboration остаются в GitHub UI; workspace даёт platform-safe next-step actions и переход в provider context, но не заменяет provider semantics.
- `D-510-09`: stage continuity через `run:dev` обязана сохранять правило `PR + linked follow-up issue`.
- `D-510-10`: voice/STT, dashboard orchestrator agent, отдельная `agent` node taxonomy, full-history/archive и richer provider enrichment остаются deferred later-wave scope.

## Scope boundaries
### In scope
- Fullscreen graph workspace shell:
  - detached top toolbar;
  - right drawer/chat/details;
  - graph-first workspace, а не board/list-first landing.
- Filtered multi-root workspace:
  - default filters Wave 1 = `open_only`, `assigned_to_me_or_unassigned`, `active-state presets`;
  - в одном срезе видны несколько инициатив одновременно;
  - каждая инициатива раскладывается слева направо: root discussion/work item слева, downstream runs и artifacts справа;
  - узлы вне primary selection, но нужные для связности, показываются secondary/dimmed.
- Hybrid truth matrix и inventory-backed foundation:
  - persisted provider mirror обязателен;
  - coverage contract = все open Issues/PR + bounded recent closed history;
  - workspace строится на persisted projection с явными freshness/scope watermarks.
- Typed workspace surfaces:
  - `display_title`;
  - `display_summary`;
  - `artifact_intent`;
  - `suggested_relations[]`;
  - `ui_badges[]`;
  - platform-generated watermarks;
  - platform-canonical launch params (`model`, `reasoning`, runtime/access profile, defaults + overrides).
- Platform-safe continuity actions:
  - открыть drawer/context;
  - inspect run context;
  - prepare/launch next allowed stage;
  - открыть linked PR или follow-up issue;
  - показать continuity gap, если PR или follow-up issue отсутствуют.
- Direct and discussion-driven entry:
  - пользователь может стартовать с discussion;
  - пользователь может стартовать с любого stage issue и выбрать родительскую связь;
  - continuity graph обязан оставаться когерентным в обоих путях.

### Out of scope
- Отдельная `agent` node taxonomy в core Wave 1.
- Voice/STT и dashboard orchestrator agent как blocking requirement.
- Full-history/archive как default canvas scope.
- GitHub Projects / Issue Type / Relationships как primary graph model.
- Live-fetch-only dashboard без persisted provider mirror.
- Перенос human review/merge/provider-native collaboration в dashboard.

## Пользователи / персоны

| Persona | Основная работа | Что считает успехом |
|---|---|---|
| Owner / product lead | Понять ситуацию по нескольким инициативам, быстро определить приоритет и запустить следующий безопасный шаг | Один workspace показывает lineage, freshness и downstream artifacts без ручного GitHub label hunting |
| Delivery operator / engineer / manager | Видеть stage progression, run context, PR/follow-up issue и launch configuration в одном месте | Platform state и provider state объединены без duplicate/missing links, а next-step action понятен и policy-safe |
| Discussion-driven пользователь | Начать с discussion либо сразу со stage issue и не потерять continuity | Можно сохранить parent-child связи и дойти до PR/follow-up issue без разрыва графа |
| Reviewer / approver | Проверить, какой artifact сейчас является источником продолжения и где остаётся human decision | Workspace показывает linked PR/follow-up issue и переводит в GitHub для review/merge вместо скрытого обхода процесса |

## User stories и wave priorities

| Story ID | История | Wave | Приоритет |
|---|---|---|---|
| `S16-US-01` | Как Owner/product lead, я хочу открыть Mission Control и сразу увидеть 2-3 активные инициативы с их lineage и следующим безопасным шагом, чтобы не собирать картину вручную по GitHub и runtime-экранам | Wave 1 | `P0` |
| `S16-US-02` | Как delivery operator, я хочу видеть в одном graph workspace run context, linked PR/follow-up issue, freshness watermarks и launch params, чтобы доверять continuity control plane | Wave 1 | `P0` |
| `S16-US-03` | Как discussion-driven пользователь, я хочу начать либо с discussion, либо с прямого stage issue и сохранить parent relation, чтобы workflow не был завязан только на discussion path | Wave 1 | `P0` |
| `S16-US-04` | Как operator/reviewer, я хочу понимать, где заканчивается platform truth и начинается GitHub truth, чтобы не спорить с merge/review semantics и не получать duplicate/missing nodes | Wave 1 | `P0` |
| `S16-US-05` | Как пользователь графа, я хочу видеть узлы вне primary filter как secondary/dimmed только когда они нужны для связности, чтобы граф оставался понятным и не терял continuity | Wave 1 | `P0` |
| `S16-US-06` | Как operator, я хочу получить optional richer density/focus surfaces после core rollout, чтобы workspace оставался usable при росте active set без reopening truth model | Wave 2 | `P1` |
| `S16-US-07` | Как voice-first пользователь, я хочу в будущем опционально запускать assistant/voice path поверх уже безопасных command surfaces, чтобы не смешивать эту ценность с core Wave 1 | Wave 3 | `P1` |

### Wave sequencing
- Wave 1 (`P0`):
  - fullscreen graph workspace;
  - filtered multi-root baseline;
  - inventory-backed foundation;
  - typed metadata/watermarks/launch params;
  - platform-safe next-step actions;
  - continuity rule `PR + linked follow-up issue`.
- Wave 2 (`P1`):
  - usability hardening для большего active set;
  - richer continuity inspection и non-blocking enrichments внутри зафиксированной truth model.
- Wave 3 (`P1`, deferred):
  - voice/STT;
  - dashboard orchestrator agent;
  - отдельная `agent` node taxonomy;
  - full-history/archive;
  - richer provider enrichment beyond agreed Wave 1 contract.

## Functional Requirements

| ID | Требование |
|---|---|
| `FR-510-01` | Mission Control должен быть primary fullscreen graph workspace/control plane для active delivery, а не board/list-only dashboard. |
| `FR-510-02` | По умолчанию workspace должен показывать несколько инициатив одновременно по exact Wave 1 filters `open_only`, `assigned_to_me_or_unassigned`, `active-state presets`, а не single-root-only view. |
| `FR-510-03` | Для каждой инициативы граф должен сохранять continuity layout слева направо: `discussion/work_item -> run -> pull_request/follow-up issue -> next run`. |
| `FR-510-04` | Узлы, не прошедшие primary filter, но необходимые для graph integrity, должны отображаться secondary/dimmed, а не исчезать полностью. |
| `FR-510-05` | Core Wave 1 canvas nodes ограничены `discussion`, `work_item`, `run`, `pull_request`; comments/chat/summaries/approvals остаются drawer/timeline entities, а `agent` не входит в Wave 1 taxonomy. |
| `FR-510-06` | Workspace обязан использовать persisted provider inventory foundation с coverage contract `all open Issues/PR + bounded recent closed history`; page-load live fetch не может быть primary graph source. |
| `FR-510-07` | Hybrid truth matrix должна быть explicit: platform canonical для relations, run lineage, launch params, metadata, sync state и watermarks; GitHub canonical для issue/pr/comment/review/development state. |
| `FR-510-08` | Каждый node/drawer surface должен показывать typed metadata: `display_title`, `display_summary`, `artifact_intent`, `suggested_relations[]`, `ui_badges[]` и platform-generated watermarks. |
| `FR-510-09` | Platform-canonical launch params (`model`, `reasoning`, runtime/access profile, defaults + overrides) должны быть видимы и связаны с конкретным next-step action. |
| `FR-510-10` | Platform-safe inline actions должны поддерживать минимум: открыть drawer/context, inspect run context, prepare/launch next allowed stage, открыть linked PR и открыть linked follow-up issue. |
| `FR-510-11` | Workspace не должен обходить approval/label/review policy; при недопустимом действии пользователь должен увидеть явную блокировку или переход в GitHub, а не скрытый side effect. |
| `FR-510-12` | Stage continuity через `run:dev` обязана соблюдать правило `stage artifact = PR + linked follow-up issue`; отсутствие одного из артефактов трактуется как continuity gap. |
| `FR-510-13` | Пользователь должен иметь path начать continuity как из discussion, так и из напрямую созданного stage issue с родительской связью; graph model не может быть discussion-only. |
| `FR-510-14` | Workspace должен помогать пользователю понять следующий safe step по 2-3 инициативам в одном сеансе без ручного GitHub label hunting. |
| `FR-510-15` | GitHub остаётся местом human review, merge decision и provider-native collaboration; workspace открывает provider context, но не заменяет его. |
| `FR-510-16` | Voice/STT, dashboard orchestrator agent, отдельная `agent` taxonomy, full-history/archive и richer provider enrichment остаются deferred и не блокируют core Wave 1 release. |

## Acceptance Criteria (Given/When/Then)

### `AC-510-01` Multi-root workspace и primary filters
- Given в системе есть несколько активных инициатив,
- When пользователь открывает Mission Control,
- Then он видит multi-root graph workspace по filters `open_only`, `assigned_to_me_or_unassigned`, `active-state presets`, а не single-root-only экран.
- Expected evidence: product walkthrough по 2-3 инициативам + UX review trace.

### `AC-510-02` Continuity graph и dimmed semantics
- Given часть узлов нужна для связности графа, но не проходит primary filter,
- When workspace рендерит lineage инициативы,
- Then такие узлы остаются secondary/dimmed только для graph integrity и не ломают continuity `discussion/work_item -> run -> pull_request/follow-up issue -> next run`.
- Expected evidence: acceptance scenario с primary + supporting nodes и визуальной проверкой dimmed behavior.

### `AC-510-03` Hybrid truth и inventory coverage
- Given provider inventory содержит open Issues/PR и configured bounded recent closed history,
- When projection synchronised в workspace,
- Then каждая in-scope сущность появляется один раз с корректным split между provider state и platform enrichment, а coverage/freshness отражены typed watermarks.
- Expected evidence: inventory audit, projection diff и coverage evidence по policy `all open Issues/PR + bounded recent closed history`.

### `AC-510-04` Typed metadata и node taxonomy
- Given пользователь открывает любую Wave 1 сущность,
- When он смотрит node card или drawer,
- Then он видит typed metadata, watermarks и launch params, а comments/chat/summaries остаются в drawer/timeline и не превращаются в отдельные canvas nodes.
- Expected evidence: acceptance walkthrough по `discussion`, `work_item`, `run`, `pull_request`.

### `AC-510-05` Next-step actions и continuity contract
- Given stage artifact уже существует или должен быть продолжен,
- When пользователь инспектирует следующий шаг,
- Then workspace показывает linked PR и linked follow-up issue, разрешённые platform-safe actions и отмечает отсутствие любого из этих артефактов как continuity gap.
- Expected evidence: scenario matrix по stage transitions и traceability audit `PR + linked follow-up issue`.

### `AC-510-06` Discussion path и direct stage issue path
- Given пользователь начинает либо с discussion, либо с прямого stage issue,
- When он связывает downstream artifacts,
- Then continuity graph сохраняет parent-child relations и остаётся когерентным без принудительного discussion-only funnel.
- Expected evidence: walkthrough для обоих путей со сравнением получившегося lineage.

### `AC-510-07` GitHub-native human review preserved
- Given текущий следующий шаг требует human review или merge decision,
- When пользователь действует из workspace,
- Then workspace переводит его в GitHub context и не пытается выполнить provider-native review/merge внутри dashboard.
- Expected evidence: product walkthrough + policy review по action surfaces.

### `AC-510-08` Deferred later-wave scope does not block core release
- Given voice/STT, dashboard orchestrator agent, отдельная `agent` taxonomy или full-history/archive недоступны,
- When оценивается готовность core Wave 1,
- Then Mission Control остаётся релизуемым как graph workspace/control plane без этих contours.
- Expected evidence: scope review + wave-based release readiness checklist.

## Scenario matrix

| ID | Сценарий | Обязательное поведение | Expected evidence |
|---|---|---|---|
| `SC-510-01` | Owner triages 3 active initiatives | Workspace показывает все инициативы в одном graph view, keeps exact filters, позволяет быстро открыть нужный node/drawer и понять next step | Pilot walkthrough + product review note |
| `SC-510-02` | Delivery operator проверяет run -> PR -> follow-up continuity | Для выбранного branch of work видны run context, linked PR, linked follow-up issue и launch params для следующего stage | Traceability walkthrough + action audit mock |
| `SC-510-03` | Discussion переходит в stage issue | Discussion может иметь downstream stage issue без потери lineage; graph остаётся когерентным и не делает discussion единственным допустимым entrypoint | Scenario walkthrough `discussion -> work_item -> run` |
| `SC-510-04` | Persisted inventory подтягивает open + bounded recent closed history | Workspace отражает in-scope provider entities, не теряет recent closed context, а freshness/scope показываются watermarks | Inventory coverage audit + projection sample |
| `SC-510-05` | Связующий узел не прошёл primary filter | Узел остаётся dimmed, чтобы не потерять continuity, но не захламляет primary active workspace | Visual acceptance capture |
| `SC-510-06` | Human review/merge остаются в GitHub | Workspace показывает downstream PR и redirect в provider context вместо hidden review/merge semantics inside dashboard | Policy walkthrough + operator review |
| `SC-510-07` | Voice/STT contour не включён | Core Wave 1 остаётся complete и releaseable без assistant/voice path | Wave scope review |

## Edge cases и non-happy paths

| ID | Сценарий | Ожидаемое поведение | Evidence |
|---|---|---|---|
| `EC-510-01` | В continuity branch есть stage artifact без linked follow-up issue | Workspace показывает continuity gap и не считает branch завершённым | Traceability audit |
| `EC-510-02` | Provider mirror устарел или bounded recent closed history не включает более старый artifact | Пользователь видит freshness/scope watermark и понимает, что entity вне agreed coverage window | Coverage/freshness walkthrough |
| `EC-510-03` | Ветвь содержит supporting node вне primary filter | Узел остаётся dimmed, чтобы не потерять graph integrity | Visual acceptance check |
| `EC-510-04` | Пользователь пытается выполнить human review/merge в workspace | Workspace блокирует это действие или переводит в GitHub context по policy-safe path | Policy check |
| `EC-510-05` | Deferred contours недоступны | Core Wave 1 остаётся usable и не показывает их как blocked prerequisite | Scope review |

## Non-Goals
- Переносить весь provider collaboration lifecycle в dashboard.
- Делать voice/STT или dashboard orchestrator agent обязательной частью core Wave 1.
- Добавлять `agent` как обязательный canvas node type в первой волне.
- Делать полный архив или “весь мир” дефолтным canvas scope.
- Использовать markdown scraping или LLM-generated watermarks как canonical metadata source.

## NFR draft для handover в architecture

| ID | Требование | Как измеряем / проверяем |
|---|---|---|
| `NFR-510-01` | p75 времени от открытия workspace до первого continuity-focused действия должно быть `<= 60 секунд` | frontend analytics + product pilot |
| `NFR-510-02` | Inventory-backed coverage для scope `all open Issues/PR + bounded recent closed history` должна быть `100%` в пределах agreed policy window | provider mirror audit + reconcile evidence |
| `NFR-510-03` | Hybrid truth merge correctness должна быть `>= 99.5%` без duplicate nodes, missing lineage или ambiguous ownership | projection audit + incident review |
| `NFR-510-04` | Continuity artifact completeness по правилу `PR + linked follow-up issue` до `run:dev` включительно должна быть `100%` | `flow_events` + issue/PR traceability audit |
| `NFR-510-05` | Workspace должен оставаться usable для 2-3 одновременных инициатив без отката к single-root-only workflow | UX pilot evidence + product review |
| `NFR-510-06` | Пользователь всегда видит typed freshness/scope watermarks для graph surfaces, зависящих от provider inventory | acceptance review + design checklist |
| `NFR-510-07` | Workspace не должен обходить label/approval/review policy и provider-native merge semantics | policy review + architecture gate |

## Analytics и product evidence
- События:
  - `mission_control_workspace_opened`
  - `mission_control_filters_applied`
  - `mission_control_node_opened`
  - `mission_control_next_step_opened`
  - `mission_control_provider_context_opened`
  - `mission_control_launch_params_viewed`
  - `mission_control_continuity_gap_seen`
  - `mission_control_watermark_seen`
- Метрики:
  - `NSM-496-01` Continuity workspace adoption rate
  - `PM-496-01` Next-step clarity p75
  - `OPS-496-01` Inventory-backed active coverage
  - `REL-496-01` Hybrid truth merge correctness
  - `UX-496-01` Multi-initiative usability rate
  - `GOV-496-01` Continuity artifact completeness
- Expected evidence:
  - product walkthrough по 2-3 инициативам;
  - traceability audit по `PR + linked follow-up issue`;
  - inventory coverage evidence against agreed provider policy;
  - UX review по primary filters, dimmed semantics и next-step clarity;
  - policy walkthrough по GitHub-native review/merge boundary.

## Handover в `run:arch`
- Follow-up issue: `#516`.
- На архитектурном этапе нельзя потерять:
  - Mission Control как primary multi-root graph workspace/control plane;
  - issue `#480` как mandatory inventory foundation с coverage contract `all open Issues/PR + bounded recent closed history`;
  - exact Wave 1 filters `open_only`, `assigned_to_me_or_unassigned`, `active-state presets`;
  - secondary/dimmed semantics только для graph integrity;
  - Wave 1 nodes `discussion`, `work_item`, `run`, `pull_request`;
  - typed metadata/watermarks и platform-canonical launch params;
  - platform-safe inline actions и continuity rule `PR + linked follow-up issue`;
  - GitHub-native human review/merge boundary;
  - deferred статус voice/STT, dashboard orchestrator agent, отдельной `agent` taxonomy, full-history/archive и richer provider enrichment.
- Архитектурный этап обязан определить:
  - ownership split для `control-plane`, `worker`, `api-gateway`, `web-console`;
  - как живут persisted projections, relations и hybrid truth merge path;
  - какие transport/data surfaces нужны для typed metadata, watermarks, launch params и next-step actions;
  - какую issue на `run:design` нужно создать без trigger-лейбла, сохранив continuity `design -> plan -> dev`.
