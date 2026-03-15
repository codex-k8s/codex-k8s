---
doc_id: EPC-CK8S-S16-D1-MISSION-CONTROL-GRAPH
type: epic
title: "Epic S16 Day 1: Intake для Mission Control graph workspace и continuity control plane (Issue #492)"
status: in-review
owner_role: PM
created_at: 2026-03-15
updated_at: 2026-03-15
related_issues: [480, 490, 492, 496]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-15-issue-492-intake"
---

# Epic S16 Day 1: Intake для Mission Control graph workspace и continuity control plane (Issue #492)

## TL;DR
- Текущий Mission Control не соответствует ожиданиям Owner: нужен не улучшенный board/list, а новый primary workspace платформы с fullscreen canvas, detached top toolbar, правым drawer/chat и явным lineage `discussion/work_item -> run -> PR/follow-up issue -> next run`.
- Sprint S16 поглощает foundation issue `#480`: persisted GitHub inventory mirror и bounded reconcile остаются обязательным нижним слоем, при этом foundation coverage contract сохраняется как `all open Issues/PR + bounded recent closed history`, а не заменяется расплывчатым backlog mirror.
- Intake фиксирует hybrid truth matrix, filtered multi-root workspace с точным Wave 1 filter baseline `open_only`, `assigned_to_me_or_unassigned` и active-state presets, закрытый Wave 1 node set `discussion/work_item/run/pull_request`, typed metadata/watermark contract, platform-canonical launch params и continuity-rule `PR + follow-up issue` для каждого stage до `run:dev` включительно.
- Создана continuity issue `#496` для stage `run:vision`; на следующем этапе нельзя возвращаться к GitHub-first single-root board framing или превращать voice path в blocking scope.

## Контекст
- Sprint S9 (`#333`, `#335`, `#337`, `#340`, `#351`, `#363`, `#369..#375`) уже построил первый Mission Control baseline: active-set dashboard, typed projection, core transport/UI foundations и execution backlog.
- Issue `#480` затем зафиксировал product gap: текущий Mission Control строится в основном из platform execution evidence и поэтому не покрывает GitHub inventory как полный active backlog; был предложен persisted provider mirror как обязательный foundation stream с concrete coverage contract `all open Issues/PR + bounded recent closed history`.
- Owner discussion в `#490` отменила trajectory «дополировать текущий dashboard» и явно потребовала новый рабочий контур:
  - fullscreen canvas вместо board/list-first страницы;
  - detached top toolbar и правый drawer/chat;
  - граф связей между `discussion`, stage-issues, runs, PR и follow-up issues;
  - возможность работать не только через discussion, но и создавать любой stage-issue с любым родителем;
  - launch params (`model`, `reasoning`, runtime profile) задаются перед запуском и хранятся platform-canonically.
- Issue `#492` консолидировал эти решения в intake baseline и зафиксировал, что S16 не заменяет `#480`, а поглощает его как foundation layer.

## Рекомендованный launch profile
- Базовый launch profile: `new-service`.
- Причины:
  - инициатива меняет product contour Mission Control, ownership expectations и continuity contract между несколькими stage;
  - затрагиваются `control-plane`, `worker`, `api-gateway`, `web-console`, webhook/reconcile semantics и provider/platform truth split;
  - обязательны `vision`, `arch` и `design`, иначе граф-модель, metadata contract и ownership boundaries снова разъедутся.
- Целевая stage-цепочка:
  `run:intake -> run:vision -> run:prd -> run:arch -> run:design -> run:plan`.

## Problem Statement
### As-Is
- Текущий Mission Control остаётся board/list представлением поверх частичной platform evidence и не является для Owner первичным рабочим местом по нескольким инициативам сразу.
- В продукте отсутствует first-class graph model, которая одновременно выражает:
  - multi-root workspace по текущим owner filters;
  - lineage `discussion/work_item -> run -> PR/follow-up issue -> next run`;
  - split source-of-truth между platform state и provider-native state;
  - typed metadata, launch params и dashboard-specific watermarks.
- Persisted provider mirror как foundation из `#480` ещё не превращён в часть новой product shape, поэтому текущий dashboard может быть технически живым, но продуктово пустым.
- Если в первую волну одновременно включить core graph workspace, provider mirror, voice/STT и dashboard orchestrator agent, инициатива потеряет фокус и time-to-value.

### To-Be
- Mission Control становится fullscreen graph workspace/control plane, а не incremental улучшением существующей dashboard-страницы.
- Workspace по умолчанию показывает все сущности, прошедшие точный Wave 1 filter baseline `open_only`, `assigned_to_me_or_unassigned` и active-state presets, и позволяет одновременно видеть несколько инициатив; graph layout остаётся left-to-right по каждой инициативе, а узлы, нужные для graph integrity, но не прошедшие основной фильтр, остаются secondary/dimmed вместо полного скрытия.
- Platform и GitHub получают явный hybrid truth split:
  - platform canonical: operational graph, relations, run nodes, launch params, dashboard metadata, sync state, platform-generated watermarks;
  - GitHub canonical: issue/pr/comment/review/provider-native state и development links.
- Wave 1 ограничен core node set `discussion`, `work_item`, `run`, `pull_request`; comments/chat/summaries живут в drawer/timeline, а `agent` остаётся enrichment surface, а не canvas node.

## Brief
- **Проблема:** Mission Control не стал основным рабочим местом Owner и platform operator; текущий результат не поддерживает multi-root graph continuity и не выражает реальный lifecycle `discussion -> stage issue -> run -> PR/follow-up`.
- **Для кого:** для Owner/operator, которому нужен единый workspace нескольких инициатив; для discussion-driven пользователя; для delivery-ролей, которым нужен typed continuity graph и безопасный запуск следующих stage.
- **Предлагаемое решение:** Sprint S16 оформляет новый graph workspace как отдельную инициативу с intake/vision/prd/arch/design/plan chain и поглощает `#480` как обязательный inventory-backed foundation layer.
- **Почему сейчас:** S9 дал execution baseline, `#480` показал data-source gap, а `#490` дал owner-approved redesign target; дальше откладывать нельзя, иначе Mission Control будет продолжать накапливать drift между UX, provider inventory и platform semantics.
- **Что считаем успехом:** intake stage фиксирует новый product contour, scope boundaries, acceptance baseline и continuity issue `#496` без потери owner decisions.
- **Что не делаем на этой стадии:** не выбираем конкретные transport/data/UI implementations, не добавляем новые зависимости, не обещаем voice/STT в core Wave 1 и не делаем live-fetch UI поверх GitHub без persisted mirror.

## MVP Scope
### In scope
- Полное перепроектирование Mission Control как primary graph workspace/control plane staff console.
- Fullscreen canvas с detached top toolbar и правым drawer/chat/details panel.
- Filtered multi-root workspace:
  - по умолчанию отображаются все сущности, прошедшие точные Wave 1 filters `open_only`, `assigned_to_me_or_unassigned` и active-state presets;
  - в одном срезе видно несколько инициатив одновременно;
  - каждая инициатива должна уметь визуально раскладываться слева направо: discussion/root nodes слева, затем runs и последующие артефакты вправо;
  - узлы, которые нужны только для связности графа и не проходят основной фильтр, показываются secondary/dimmed, а не исчезают.
- Закрытый Wave 1 node set:
  - `discussion`
  - `work_item`
  - `run`
  - `pull_request`
- Agent presence в Wave 1 только как enrichment surface:
  - badges/chips на карточках;
  - filters по agent/ownership;
  - details/timeline context в drawer;
  - отдельная `agent`-node допустима только как later-wave option.
- Hybrid truth matrix:
  - platform canonical для operational graph, relations, run-produced artifacts, launch params, dashboard metadata, sync state и watermarks;
  - GitHub canonical для provider-native issue/pr/comment/review state и development links;
  - GitHub Projects / Issue Type / Relationships допускаются только как mirror/enrichment, но не как primary graph model.
- Inventory-backed foundation из `#480`:
  - persisted provider mirror обязателен;
  - coverage contract foundation stream = все открытые Issues, все открытые PR и bounded recent closed history по Issues/PR;
  - backfill/rebuild baseline = `hybrid warmup + lazy repair`;
  - dashboard summary/graph должны строиться по persisted projection, а не по page-load live fetch.
- Typed metadata contract и будущие MCP/control-plane surfaces для:
  - `display_title`
  - `display_summary`
  - `artifact_intent`
  - `suggested_relations[]`
  - `ui_badges[]`
- Platform-generated watermarks и platform-canonical launch params (`model`, `reasoning`, runtime/access profile) с defaults + overrides.
- Continuity contract:
  - каждый doc-stage `run:intake -> run:vision -> run:prd -> run:arch -> run:design -> run:plan` обязан выпускать PR/markdown artifact и отдельную linked follow-up issue без trigger-лейбла;
  - `run:plan` обязан создать follow-up issue для `run:dev`;
  - `run:dev` для этой инициативы считается завершённым только при наличии PR с кодом/тестами/docs и следующей linked follow-up issue для downstream execution step.

### Out of scope для core wave
- Отдельная `agent` node taxonomy в Wave 1.
- Попытка показывать «весь мир» или весь архив по умолчанию без active filters.
- GitHub Projects / Issue Type / Relationships как primary source of truth для operational graph.
- LLM-generated system watermarks.
- Markdown scraping как канонический источник dashboard metadata.
- Перенос human review/merge/provider-native collaboration из GitHub UI в dashboard.
- STT/voice как blocking requirement core Wave 1.
- Отказ от persisted provider mirror foundation в пользу live-fetch-only UX.

## Constraints
- Sprint S16 обязан использовать existing baselines, а не проектироваться «с чистого листа»:
  - Sprint S9 как baseline core Mission Control execution/package;
  - issue `#480` как inventory-backed foundation stream;
  - Sprint S12 как rate-limit/wait/backpressure baseline для GitHub sync.
- Low-level service ownership, DTO and storage design не выбираются на intake stage; это работа `run:arch` и `run:design`.
- Dashboard не имеет права обходить stage/label policy, approval rules, provider review model и audit trail.
- Comments, replies, summaries, approval events и chat остаются timeline/drawer сущностями, а не отдельными canvas nodes.
- Stage continuity до `run:dev` включительно фиксируется как обязательный процессный контракт и не может считаться optional artifact.

## Product principles
- Workspace first, а не incremental board/list polishing.
- Continuity graph важнее локального «красивого» UI-среза.
- Hybrid truth matrix должна быть явной и typed, а не скрытой смесью GitHub полей, markdown scraping и эвристик.
- Persisted provider mirror обязателен как foundation, но сам по себе не считается продуктовым финалом.
- Voice/orchestrator experience ценен, но не должен размывать первую волну core workspace.

## Candidate product waves
| Wave | Фокус | Exit signal |
|---|---|---|
| Wave 1 | Graph workspace, hybrid truth matrix, inventory foundation, run nodes, drawer/chat/details, typed metadata, launch params, platform-safe inline actions | Owner может вести 2-3 инициативы в одном workspace и понимать lineage/next step без возврата к board/list-only модели |
| Wave 2 | Density/fidelity improvements, richer graph continuity, additional provider enrichment and filters | Workspace устойчив к росту объёма сущностей без визуального шума и split-brain |
| Wave 3 | Dashboard orchestrator agent и voice conversation path поверх platform-safe commands | Голосовой/assistant path доказывает ценность, не ломая core policy и control-plane semantics |

## Acceptance Criteria (Intake stage)
- [x] Формально зафиксировано, что Sprint S16 — это полное перепроектирование Mission Control в graph workspace/control plane, а не incremental UX-tuning текущего dashboard.
- [x] Зафиксирован filtered multi-root workspace baseline с точными Wave 1 filters `open_only`, `assigned_to_me_or_unassigned`, active-state presets, left-to-right layout и secondary/dimmed handling для связующих узлов вне primary selection.
- [x] Зафиксирован закрытый Wave 1 node set `discussion + work_item + run + pull_request`, а `agent` не входит в Wave 1 canvas nodes.
- [x] Зафиксирована hybrid truth matrix по ownership boundaries между platform state и GitHub state.
- [x] Зафиксировано, что inventory-backed provider mirror из `#480` является обязательным foundation с coverage contract `all open Issues/PR + bounded recent closed history`, а не альтернативным backlog.
- [x] Зафиксирован typed metadata/watermark contract baseline: platform-generated watermark, no markdown scraping as canonical source.
- [x] Зафиксировано platform-canonical хранение launch params с defaults + overrides.
- [x] Зафиксировано, что voice/STT не блокирует core Wave 1 и выносится в later wave через dashboard orchestrator path.
- [x] Зафиксировано continuity-требование: каждый stage до `run:dev` включительно обязан завершаться не только PR, но и linked follow-up issue для следующего шага.
- [x] Подготовлена continuity issue `#496` для stage `run:vision`.

## Risks and Product Assumptions
- Риск: scope explosion, если в одну волну смешать graph UX, provider mirror, voice, dashboard orchestrator и новые agent surfaces.
- Риск: split-brain между provider mirror, platform graph и runtime lineages, если ownership boundaries и truth matrix будут оставлены «на усмотрение реализации».
- Риск: multi-root graph без density rules превратится в визуальный шум; это ограничивается filters, dimmed semantics, grouping rules и fallback views.
- Допущение: существующие `control-plane`/`worker`/`api-gateway`/`web-console` bounded contexts могут эволюционно расшириться до graph workspace без немедленного выделения нового deployable сервиса.
- Допущение: dashboard orchestrator agent должен строиться поверх platform-safe command surfaces и текущего approval/process baseline, а не обходить его.

## Stage Handover Instructions
- Следующий этап: `run:vision`.
- Созданная issue следующего этапа: `#496`.
- На stage `run:vision` обязательно сохранить и не переоткрывать без owner-решения следующие intake-decisions:
  - Sprint S16 поглощает issue `#480` как mandatory foundation stream;
  - foundation coverage contract из `#480` сохраняется как `all open Issues/PR + bounded recent closed history`;
  - multi-root filtered workspace удерживает точные Wave 1 filters `open_only`, `assigned_to_me_or_unassigned` и active-state presets;
  - multi-root filtered workspace является default baseline;
  - Wave 1 nodes = `discussion`, `work_item`, `run`, `pull_request`, без `agent` node;
  - comments/chat/summaries остаются drawer/timeline entities, а не canvas nodes;
  - secondary/dimmed handling используется только для graph integrity, а не как замена primary filter baseline;
  - hybrid truth matrix между platform и GitHub обязательна;
  - typed metadata contract, platform-generated watermarks и platform-canonical launch params обязательны;
  - voice/STT и dashboard orchestrator agent не входят в blocking scope core Wave 1;
  - continuity requirement до `run:dev`: результат stage = PR + linked follow-up issue.
- После завершения vision stage должна быть создана новая issue для `run:prd` без trigger-лейбла.
