---
doc_id: EPC-CK8S-S9-D1-MISSION-CONTROL
type: epic
title: "Epic S9 Day 1: Intake для Mission Control Dashboard и console control plane (Issue #333)"
status: in-review
owner_role: PM
created_at: 2026-03-12
updated_at: 2026-03-12
related_issues: [333, 335]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-12-issue-333-intake"
---

# Epic S9 Day 1: Intake для Mission Control Dashboard и console control plane (Issue #333)

## TL;DR
- `services/staff/web-console` уже умеет показывать отдельные runtime/debug контуры, но не даёт единый control-plane экран, который отвечает на вопрос «что сейчас происходит и что делать дальше».
- Issue `#333` формализует новую инициативу Mission Control Dashboard: active-set landing page для work items, discussion, PR и агентов с быстрыми действиями, side panel и realtime-состоянием.
- Intake-решение: инициатива остаётся GitHub-first, active-set oriented и policy-safe; voice intake и GitLab parity не блокируют core dashboard wave до vision/PRD.

## Контекст
- Текущий As-Is baseline платформы уже содержит нужные фрагменты, но не объединяет их в один продуктовый контур:
  - staff console умеет показывать runtime pages и live-состояния по отдельным потокам;
  - `mode:discussion`, stage labels и review/revise policy уже являются канонической operational model;
  - realtime foundation (`PostgreSQL LISTEN/NOTIFY` + WebSocket backplane) уже существует, но не описана как единая модель dashboard;
  - GitHub остаётся каноническим provider'ом MVP, а review человека выполняется во внешнем provider UI.
- По итогам обсуждения в Issue `#333` подтверждён целевой продуктовый сдвиг:
  - пользователь должен запускать и продолжать работу в первую очередь из нашей консоли;
  - GitHub остаётся интеграционным и review-слоем, а не основным операционным центром;
  - console должна одинаково хорошо работать для active set из десятков объектов и не ломаться при росте до сотен сущностей.

## Рекомендованный launch profile
- Базовый launch profile: `feature`.
- Обязательная эскалация до полного doc-stage контура:
  - `vision` обязателен из-за новой mission/KPI модели staff console;
  - `arch` обязателен из-за cross-service impact на projections, realtime contracts, provider sync и persisted state.
- Целевая stage-цепочка:
  `run:intake -> run:vision -> run:prd -> run:arch -> run:design -> run:plan`.

## Problem Statement
### As-Is
- Пользователь переключается между несколькими страницами staff console, GitHub issue/PR UI и runtime-экранами, чтобы понять текущее состояние работы.
- В продукте нет first-class модели, которая связывает discussion -> formal task -> PR -> agent -> comments/timeline.
- Сценарий `идея -> discussion -> formalized task -> запуск stage` не собран в единый UX и не приоритизирован как отдельный funnel.
- Действия, инициированные из console, пока не описаны как полный lifecycle `internal command -> outbound provider sync -> webhook echo reconciliation`, поэтому сохраняется риск дублей и повторных запусков.

### To-Be
- Staff console открывается на Mission Control Dashboard и за 5-10 секунд показывает active set: work items, discussion, PR, агентов, блокировки и ожидания.
- Dashboard оперирует typed сущностями и явными связями, а не набором разрозненных списков.
- Discussion-first путь становится штатным продуктовым сценарием: создать discussion, обсудить, формализовать в task, запустить следующий stage.
- Любое действие из console остаётся audit-safe и provider-safe: review человека не переносится из GitHub UI, а webhook echo не порождает дублей.

## Brief
- **Проблема:** у платформы нет единого control-plane UX для активной работы; пользователь всё ещё мыслит через GitHub UI и разрозненные staff pages.
- **Для кого:** owner/product lead, инженер/оператор и discussion-first пользователь, который начинает работу с идеи или голосового ввода.
- **Предлагаемое решение:** active-set dashboard как новая landing page, которая объединяет work items, PR, discussion, agents, side panel, быстрые действия и realtime updates.
- **Почему сейчас:** текущий функциональный baseline достаточно зрелый, и главный продуктовый разрыв сместился из области «чего не умеет платформа» в область «насколько быстро и понятно можно управлять всем потоком из одного места».
- **Что считаем успехом:** dashboard сокращает время на situational awareness и на переход от идеи к formalized task без обхода существующих policy и review guardrails.
- **Что не делаем на этой стадии:** не превращаем задачу в общий редизайн всей консоли, не заменяем GitHub как место human review и не фиксируем инженерные реализации раньше `run:arch`/`run:design`.

## MVP Scope
### In scope
- Новая `DashboardPage` как default landing page для active work items.
- Summary strip, board/canvas view как primary UX и list view как fallback для high-volume режима.
- Правая side panel с деталями, timeline, chat/comments и action surface.
- Typed модель сущностей и связей:
  - Work Item / Discussion;
  - Pull Request;
  - Agent;
  - Comment stream;
  - Relation.
- Быстрые действия на сущностях:
  - создать task/discussion;
  - открыть детали;
  - перейти в чат/комментарии;
  - запустить/продолжить stage;
  - формализовать discussion в task.
- Realtime snapshot/delta модель поверх существующего WebSocket backplane.
- GitHub-first sync для titles, labels, comments, PR status и state reconciliation.
- Command/correlation/dedupe модель для пути `UI command -> outbound sync -> webhook echo`.

### Out of scope для core wave
- Произвольный dashboard/layout builder и полная настройка canvas-представления пользователем.
- Попытка показать весь исторический архив на дефолтном canvas без active-set фильтра.
- Перенос human review и merge decision из GitHub UI в console.
- GitLab parity в первой волне MVP.
- Voice intake как blocking requirement для запуска core dashboard wave:
  он остаётся отдельным candidate stream внутри инициативы и требует отдельной оценки на vision/PRD.

## Constraints
- MVP остаётся GitHub-first по provider semantics; GitLab допускается только как future-compatible continuation через provider abstraction.
- Stage/label policy, `mode:discussion`, review/revise loop и webhook-driven orchestration остаются source of truth и не переопределяются dashboard UX.
- Active set, search и filters обязательны; продукт не должен оптимизироваться под «рисуем всё подряд на одном графе».
- Dashboard actions должны проходить через typed commands, audit trail и reconciliation, а не через скрытые side-effect операции.
- Для staff/external API сохраняется contract-first OpenAPI и typed DTO baseline.
- Visual baseline для primary view фиксируется как board/canvas UX; выбор конкретной frontend dependency (включая обсуждавшийся Vue Flow path) подтверждается только на `run:arch`/`run:design`.
- Voice сценарии возможны только при наличии platform AI policy и соответствующих env-capabilities; они не должны ломать основной funnel без voice.

## Product principles
- Console становится control plane, а не очередной страницей-наблюдашкой.
- Active set важнее полного архива: сначала ответить «что важно сейчас», потом уже давать deep drilldown.
- Provider collaboration, а не provider replacement: GitHub остаётся местом human review и частью синхронизации.
- Любой запуск или update из console должен быть детерминированно коррелируемым и безопасным к webhook echo.

## Candidate product waves

| Wave | Фокус | Exit signal |
|---|---|---|
| Wave 1 | Active-set dashboard shell: landing page, summary, board/list toggle, side panel, typed entities и базовые действия | Пользователь видит единый control-plane экран и может создать/продолжить активную работу без переключения по разрозненным страницам |
| Wave 2 | Discussion formalization, provider sync hardening, comment/chat projection и webhook-echo dedupe | Discussion-first сценарий работает end-to-end без дублей и без потери traceability |
| Wave 3 | Voice intake и AI-assisted draft structuring | Голосовой путь доказал ценность и не ломает policy/ops baseline core dashboard |

## Acceptance Criteria (Intake stage)
- [x] Проблема зафиксирована как отсутствие unified control-plane UX, а не как общий «редизайн интерфейса».
- [x] Зафиксированы primary users, core entities и active-set принцип отображения.
- [x] Определены MVP границы, non-goals и неизменяемые ограничения инициативы.
- [x] Рекомендован launch profile `feature` с обязательной эскалацией в `vision` и `arch`.
- [x] Зафиксированы основные продуктовые риски и допущения: visual noise, scope explosion, webhook dedupe/correlation, voice dependency, GitLab deferral.
- [x] Создана follow-up issue `#335` для stage `run:vision` без trigger-лейбла.

## Декомпозиция по этапам (до plan)

| Stage | Фокус | Выходной артефакт |
|---|---|---|
| Vision | Mission, KPI, persona outcomes, MVP/Post-MVP границы | charter + success metrics + risk frame |
| PRD | User stories, FR/AC/NFR, candidate waves и edge cases | PRD + user stories + NFR |
| Arch | Service boundaries, projection ownership, provider sync / reconciliation ownership, realtime responsibilities | architecture package + ADR/alternatives |
| Design | API/data/UI/realtime contracts, migration and rollout notes | design package + API/data model |
| Plan | Execution waves, quality-gates, implementation issues, DoR/DoD | execution package + linked issues |

## Risks and Product Assumptions
- Риск: dashboard превратится в визуальный шум, если не зафиксировать active-set default и ограничения на graph density.
- Риск: инициатива расползётся в «переделать всю консоль», если не удерживать её вокруг control-plane UX и active work flows.
- Риск: без явной command/reconciliation модели console будет порождать дубли issue/run/comment после webhook echo.
- Риск: voice intake увеличит scope и time-to-value, если станет обязательной частью core wave.
- Допущение: существующий realtime substrate и текущие provider/webhook контуры достаточно зрелые, чтобы расширяться, а не проектироваться с нуля.
- Допущение: новую Work Item / Relation модель можно встроить без слома существующей label/stage policy.

## Stage Handover Instructions
- Следующий этап: `run:vision`.
- Созданная issue следующего этапа: `#335`.
- На stage `run:vision` обязательно сохранить и не размыть следующие решения intake:
  - GitHub-first MVP;
  - human review остаётся во внешнем provider UI;
  - active-set default важнее полного исторического canvas;
  - voice intake не является blocking scope для core dashboard wave;
  - dashboard не создаёт обходов label/audit policy и webhook-driven orchestration.
- После завершения vision stage должна быть создана новая issue для `run:prd` без trigger-лейбла.
