---
doc_id: EPC-CK8S-S12-D1-RATE-LIMIT
type: epic
title: "Epic S12 Day 1: Intake для GitHub API rate-limit resilience, wait-state UX и MCP backpressure (Issue #366)"
status: completed
owner_role: PM
created_at: 2026-03-13
updated_at: 2026-03-13
related_issues: [366, 413]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-03-13-issue-366-intake"
  approved_by: "ai-da-stas"
  approved_at: 2026-03-13
---

# Epic S12 Day 1: Intake для GitHub API rate-limit resilience, wait-state UX и MCP backpressure (Issue #366)

## TL;DR
- Платформа уже использует два разных GitHub-контурa (`platform PAT` и `agent bot-token`), но не умеет единообразно и прозрачно пережидать rate-limit exhaustion.
- Issue `#366` фиксирует инициативу как отдельный product stream: controlled wait-state вместо ложного `failed`, явная прозрачность для Owner/операторов и MCP backpressure для agent path.
- Intake-решение: инициатива остаётся GitHub-first, но не ломает provider abstraction; primary/secondary rate limits рассматриваются как разные product signals, а не как один общий countdown.

## Контекст
- Текущий As-Is baseline уже разделяет два operational contour:
  - platform management path использует `KODEX_GITHUB_PAT`;
  - agent runtime path использует `KODEX_GIT_BOT_TOKEN` через `gh`/`git`.
- При исчерпании GitHub API limit сейчас отсутствует единая пользовательская семантика:
  - часть операций выглядит как обычный `failed`;
  - часть может уходить в локальные retries без прозрачного controlled wait;
  - staff/operator не видит, какой именно контур заблокирован и когда ожидание может закончиться.
- Официальная GitHub Docs, проверенная 2026-03-13, различает primary и secondary rate limits и допускает разные recovery signals (`Retry-After`, `X-RateLimit-*`), поэтому продукт не может опираться на один фиксированный числовой threshold как source of truth.

## Рекомендованный launch profile
- Базовый launch profile: `feature`.
- Обязательная эскалация до полного doc-stage контура:
  - `vision` обязателен из-за user-facing transparency, owner notifications и KPI по controlled wait;
  - `arch` обязателен из-за cross-service impact на `control-plane`, `worker`, `agent-runner`, MCP и staff UX.
- Целевая stage-цепочка:
  `run:intake -> run:vision -> run:prd -> run:arch -> run:design -> run:plan`.

## Problem Statement
### As-Is
- При `403/429` GitHub API платформа и/или агент не дают нормализованный status ожидания до `reset`.
- Нет единой модели persisted wait-state, которая связывает token kind, affected operation, retry hints и audit trail.
- Агентный path через `gh` не имеет обязательного backpressure-механизма на product/policy уровне и рискует застрять в локальных retry-loop.
- Staff/owner UX не показывает, какой контур заблокирован (`platform PAT` или `agent bot-token`) и какие run/deploy/admin операции это затрагивает.

### To-Be
- Rate-limit exhaustion трактуется как типизированный controlled wait-state, а не как «непонятный fail».
- Пользователь и оператор видят:
  - какой именно контур упёрся в лимит;
  - почему началось ожидание;
  - когда ожидание ориентировочно закончится;
  - какие объекты и действия затронуты.
- Агент при rate-limit сигнале переходит в controlled wait через MCP/policy path, а не пытается бесконечно ретраить локально.
- После снятия лимита поток продолжает работу автоматически там, где это безопасно и audit-safe.

## Brief
- **Проблема:** GitHub API rate limits уже влияют на orchestration path, но продукт не показывает controlled wait-state и не удерживает единый recovery UX.
- **Для кого:** Owner/reviewer, который ждёт завершения run/stage; оператор платформы в staff UI; агент, который должен корректно реагировать на rate-limit сигнал.
- **Предлагаемое решение:** единая product model для rate-limit resilience: budget-aware wait-state, split visibility `platform PAT` vs `agent bot-token`, MCP backpressure для agent path, owner notifications и безопасный resume.
- **Почему сейчас:** лимиты уже мешают dogfooding и stage-flow; без нормализованного поведения платформа выглядит нестабильной даже там, где нужна не ошибка, а ожидание.
- **Что считаем успехом:** recoverable rate-limit больше не превращается в ложный fail, а пользователь понимает причину ожидания и следующий шаг.
- **Что НЕ делаем:** не превращаем инициативу в общий redesign всех retry/backoff политик и не расширяем scope до произвольного quota-management для всех провайдеров.

## MVP Scope
### In scope
- Разделение продуктового поведения для `platform PAT` и `agent bot-token`.
- Typed controlled wait-state для GitHub rate-limit exhaustion, включая primary/secondary limit distinctions там, где это важно для UX и orchestration.
- Persisted/auditable wait context: token contour, affected operation, recovery hints, вход/выход из ожидания.
- Видимость в UI/service-comment/owner notification path:
  - какой контур заблокирован;
  - сколько ждать;
  - какие run/task/admin действия затронуты.
- Agent guidance и MCP backpressure handoff вместо infinite local retries.
- Resume semantics после снятия лимита там, где продолжение безопасно.

### Out of scope для core wave
- Универсальный quota-management слой для всех провайдеров и всех типов внешних ограничений.
- Автоматическая смена токена, пересоздание credentials или изменение token scope как ответ на rate limit.
- Глубокая аналитика стоимости/бюджетов по всем внешним API помимо GitHub rate-limit resilience.
- Любые code/runtime changes до завершения `run:plan`.

## Constraints
- GitHub остаётся текущим provider baseline, но решения не должны ломать repository-provider abstraction.
- Split `platform PAT` vs `agent bot-token` обязателен на всех следующих stage: их ограничения, прозрачность и recovery-path не смешиваются.
- Controlled wait-state не должен скрывать `bad credentials`, `forbidden by policy` и другие не-rate-limit ошибки.
- Из-за provider-driven secondary limit semantics продукт не должен обещать абсолютную точность countdown, если GitHub не дал достаточных сигналов.
- `run:intake` ограничен markdown-only изменениями.

## Product principles
- Controlled wait лучше ложного `failed`, если проблема recoverable и подтверждена provider signals.
- Пользователь должен видеть, какой контур упёрся в лимит, а не общий статус «GitHub недоступен».
- Агент обязан backpressure upstream через platform policy, а не brute-force ретраить GitHub локально.
- Wait-state должен быть audit-safe: вход, выход и resume фиксируются как часть управляемого процесса.

## Acceptance Criteria (Intake stage)
- [x] Проблема зафиксирована как отдельная cross-cutting initiative, а не как локальный retry-bug в одном сервисе.
- [x] Определены primary actors/outcomes и split `platform PAT` vs `agent bot-token`.
- [x] Зафиксированы MVP границы, non-goals и неподвижные ограничения инициативы.
- [x] Зафиксирована provider-driven неопределённость primary/secondary rate-limit semantics как продуктовое ограничение.
- [x] Рекомендован launch profile `feature` с обязательной эскалацией в `vision` и `arch`.
- [x] Создана follow-up issue `#413` для stage `run:vision`.

## Декомпозиция по этапам (до plan)

| Stage | Фокус | Выходной артефакт |
|---|---|---|
| Vision | Mission, persona outcomes, KPI/guardrails для controlled wait и rate-limit transparency | charter + success metrics + risk frame |
| PRD | User stories, FR/AC/NFR, evidence expectations и MVP/Post-MVP boundaries | PRD + user stories + NFR |
| Arch | Service boundaries, ownership wait-state lifecycle, notification/resume responsibilities | architecture package + ADR/alternatives |
| Design | API/data/runtime/UI contracts и rollout notes | design package + API/data model |
| Plan | Execution waves, quality-gates, implementation issues, DoR/DoD | execution package + linked issues |

## Risks and Product Assumptions
- Риск: инициатива расползётся в общий retry/backoff redesign вместо узкой product capability вокруг GitHub rate-limit resilience.
- Риск: без явного split `platform PAT` vs `agent bot-token` UI и audit будут показывать «среднюю температуру», а не реальную причину блокировки.
- Риск: secondary limit semantics у GitHub останутся частично непрозрачными; без product guardrails это приведёт к ложным обещаниям по countdown/resume.
- Допущение: существующий stage/audit/MCP/runtime baseline достаточно зрелый, чтобы добавить controlled wait как отдельную capability, а не проектировать новый orchestration flow с нуля.
- Допущение: owner/operator value выше всего там, где recoverable wait объяснён и продолжение после reset не требует ручного разбора логов.

## Stage Handover Instructions
- Следующий этап: `run:vision`.
- Созданная issue следующего этапа: `#413`.
- На stage `run:vision` обязательно сохранить и не размыть следующие решения intake:
  - controlled wait вместо ложного failed для recoverable rate-limit;
  - split `platform PAT` vs `agent bot-token`;
  - owner/operator transparency по причине ожидания и affected operations;
  - отсутствие infinite retry-loop на agent path;
  - provider-driven неопределённость secondary limits как design constraint, а не повод скрыть UX.
- После завершения vision stage должна быть создана новая issue для `run:prd` без trigger-лейбла.
