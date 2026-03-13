---
doc_id: EPC-CK8S-S10-D1-MCP-INTERACTIONS
type: epic
title: "Epic S10 Day 1: Intake для встроенных MCP-взаимодействий с пользователем (Issue #360)"
status: in-review
owner_role: PM
created_at: 2026-03-12
updated_at: 2026-03-13
related_issues: [360, 378]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-03-12-issue-360-intake"
  approved_by: "ai-da-stas"
  approved_at: 2026-03-13
---

# Epic S10 Day 1: Intake для встроенных MCP-взаимодействий с пользователем (Issue #360)

## TL;DR
- В текущем baseline у платформы уже есть built-in MCP server `codex_k8s`, approval/control tools и callback path, но нет отдельного product/domain contour для user-facing interaction tools.
- Issue `#360` формализует новую инициативу: встроенные MCP-взаимодействия с пользователем должны стать отдельной platform capability, а не расширением approval flow.
- Intake-решение: MVP строится вокруг `user.notify` и `user.decision.request`, interaction-domain остаётся channel-neutral и typed, а Telegram выносится в отдельный последовательный stream; continuity issue `#378` подготовлена для `run:vision`.

## Контекст
- В platform baseline уже существуют `run_status_report`, label tools, `owner.feedback.request` и callback endpoints `POST /api/v1/mcp/approver/callback` / `POST /api/v1/mcp/executor/callback`, но они обслуживают approval/control path, а не обычное взаимодействие агента с пользователем.
- По обсуждению в Issue `#360` подтверждено, что новый scope должен расширять существующий built-in server `codex_k8s`; отдельный MCP server entry в runtime `config.toml` не нужен.
- Исследовательский baseline и продуктовые правила требуют:
  - обязательный Owner review на каждом stage transition;
  - typed и auditable контракты;
  - сохранение доменных границ между platform core, adapters и agent runtime.
- Telegram и другие внешние каналы рассматриваются как адаптеры поверх общего interaction contract, а не как драйвер core semantics.

## Рекомендованный launch profile
- Базовый launch profile: `feature`.
- Обязательная эскалация до полного doc-stage контура:
  - `vision` обязателен, потому что инициатива вводит новую user-facing product capability и требует measurable KPI;
  - `arch` обязателен, потому что scope затрагивает callback contracts, wait-state lifecycle, audit/correlation и распределение ответственности между `control-plane`, `worker` и edge transport.
- Целевая stage-цепочка:
  `run:intake -> run:vision -> run:prd -> run:arch -> run:design -> run:plan`.

## Problem Statement
### As-Is
- У платформы нет отдельного доменного контура для уведомлений и запросов решения пользователя поверх built-in MCP server.
- Попытка использовать `owner.feedback.request` для обычных user-dialog сценариев смешивает approval semantics и user interaction semantics.
- Текущий callback path недостаточно формализован для typed user response: не закреплены `selected_option`, `free_text`, `response_kind`, `message correlation` и другие поля, нужные будущим адаптерам и runtime UX.
- Без отдельного intake platform core, future adapters и delivery flow будут проектироваться с разными предположениями о wait-state, audit и retry behavior.

### To-Be
- Платформа предоставляет встроенные user-facing MCP tools для уведомления пользователя и запроса решения без переиспользования approval flow.
- Interaction-domain описан как typed, auditable и channel-neutral, поэтому Telegram и другие каналы подключаются как адаптеры поверх одного platform contract.
- Wait-state допускается только для сценариев, где действительно требуется ответ пользователя, а notification-path остаётся non-blocking.

## Brief
- **Проблема:** у платформы нет отдельного продукта и доменной модели для user-facing interaction tools; approval path не подходит как универсальная замена.
- **Для кого:** для owner/product lead, конечного пользователя задачи и platform operator, который должен безопасно подключать каналы поверх одного core contract.
- **Предлагаемое решение:** выделить built-in interaction capability с MVP baseline `user.notify` + `user.decision.request`, typed callbacks и channel-neutral semantics.
- **Почему сейчас:** без intake дальше будет сложно корректно пройти `vision -> prd -> arch -> design` и не смешать platform core с Telegram-specific решениями.
- **Что считаем успехом:** platform-stage артефакты фиксируют отдельный interaction-domain, явные scope boundaries и handover в `run:vision` без потери ключевых guardrails.
- **Что не делаем на этой стадии:** не проектируем schema/API implementation, не выбираем library/runtime детали, не включаем Telegram/STT в core Sprint S10.

## MVP Scope
### In scope
- Built-in MCP tool `user.notify` для завершения задачи, результата действия и рекомендаций по следующему шагу.
- Built-in MCP tool `user.decision.request` для получения решения пользователя с 2-5 вариантами ответа и optional free-text.
- Typed outbound/inbound interaction contracts, пригодные для будущих adapter integrations.
- Отдельный interaction-domain, не смешанный с `mcp_action_requests` и approval semantics.
- Wait-state policy только для сценариев, где действительно нужен ответ пользователя.
- Handover в `run:vision` с continuity issue `#378`.

### Out of scope для core wave
- Telegram adapter, голос/STT, channel-specific UX affordances и richer conversation threads.
- Новый MCP server entry в runtime `config.toml`.
- Схема БД, migration design, transport schema и runtime ownership detail decisions до следующих stage'ов.
- Любой redesign approval flow под видом user interaction scope.

## Constraints
- Инициатива расширяет существующий built-in server `codex_k8s`, а не добавляет новый серверный контур.
- Approval flow и user interaction flow остаются отдельными доменами с разными моделями состояния и ожидания.
- Контракты должны оставаться typed, adapter-friendly и пригодными для contract-first дальнейшей проработки.
- Dispatch, retries, idempotency, audit и callback correlation должны жить в platform domain, а не в agent pod.
- Telegram и другие каналы остаются follow-up инициативами поверх общего platform interaction contract.
- Intake stage остаётся markdown-only и не фиксирует implementation details раньше `run:prd` / `run:arch`.

## Product principles
- User notification и user decision request являются первоклассными product capabilities, а не побочными случаями approval flow.
- Channel neutrality важнее скорости интеграции с первым конкретным каналом.
- Typed response и явная correlation model важнее удобных, но неформальных callback payloads.
- Wait-state допустим только там, где пользовательский ответ действительно влияет на продолжение run.
- Любой adapter path обязан сохранять одну и ту же platform semantics, audit trail и retry discipline.

## Candidate product waves

| Wave | Фокус | Exit signal |
|---|---|---|
| Wave 1 | Core interaction baseline: `user.notify`, `user.decision.request`, channel-neutral semantics, typed response model | Платформа умеет завершать user-facing interaction flow без смешения с approval flow |
| Wave 2 | Staff/runtime UX, delivery semantics, callback correlation, retries и audit hardening | Interaction lifecycle становится воспроизводимым и platform-safe для разных adapter paths |
| Wave 3 | Telegram adapter и voice/STT расширения | Отдельный follow-up stream подключает первый канал поверх уже зафиксированного core contract |

## Acceptance Criteria (Intake stage)
- [x] Проблема зафиксирована как отсутствие отдельного interaction-domain, а не как частный пробел approval flow.
- [x] Явно определены MVP baseline tools: `user.notify` и `user.decision.request`.
- [x] Зафиксированы обязательные границы: channel-neutral platform core, typed contracts, wait-state discipline, separation from approval semantics.
- [x] Telegram и другие channel adapters вынесены из core Sprint S10 scope и обозначены как последовательный follow-up stream.
- [x] Подготовлена continuity issue `#378` для stage `run:vision` без trigger-лейбла.

## Декомпозиция по этапам (до plan)

| Stage | Фокус | Выходной артефакт |
|---|---|---|
| Vision | Mission, north star, persona outcomes, KPI и guardrails | charter + success metrics + risk frame |
| PRD | User stories, FR/AC/NFR, edge cases, expected evidence | PRD + user stories + NFR |
| Arch | Service boundaries, callback ownership, wait-state lifecycle, adapter responsibilities | architecture package + ADR/alternatives |
| Design | API/data/wait-state contracts, rollout/rollback notes, adapter integration contract | design package + API/data model |
| Plan | Execution waves, quality-gates, implementation issues и follow-up stream split | execution package + linked issues |

## Risks and Product Assumptions
- Риск: scope расползётся в Telegram-first инициативу и потеряет channel-neutral baseline.
- Риск: без явной interaction model следующие stage'ы зашьют ad-hoc callback payloads и неоднозначные wait-state semantics.
- Риск: повторное использование approval flow затянет в core scope ненужные approve/deny assumptions и неверные audit expectations.
- Допущение: existing built-in server `codex_k8s` достаточно расширяем, чтобы принять новые user-facing tools без отдельного runtime server block.
- Допущение: platform core может безопасно держать interaction correlation/retry/audit semantics отдельно от channel-specific UX.

## Stage Handover Instructions
- Следующий этап: `run:vision`.
- Созданная issue следующего этапа: `#378`.
- На stage `run:vision` обязательно сохранить и не размыть следующие решения intake:
  - расширяется существующий built-in server `codex_k8s`;
  - approval flow и user interaction flow остаются раздельными;
  - MVP baseline ограничен `user.notify` и `user.decision.request`;
  - wait-state создаётся только для response-required сценариев;
  - Telegram остаётся отдельной следующей инициативой, а не частью core Sprint S10.
