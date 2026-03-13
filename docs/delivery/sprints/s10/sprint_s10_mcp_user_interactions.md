---
doc_id: SPR-CK8S-0010
type: sprint-plan
title: "Sprint S10: Built-in MCP user interactions and channel-neutral adapter contracts (Issue #360)"
status: in-review
owner_role: PM
created_at: 2026-03-12
updated_at: 2026-03-13
related_issues: [360, 378, 383, 385, 387, 389, 391, 392, 393, 394, 395]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-12-issue-360-intake"
---

# Sprint S10: Built-in MCP user interactions and channel-neutral adapter contracts (Issue #360)

## TL;DR
- Sprint S10 открывает отдельный product stream для встроенных MCP-взаимодействий с пользователем поверх уже существующего built-in сервера `codex_k8s`.
- Issue `#360` фиксирует intake baseline: MVP строится вокруг `user.notify` и `user.decision.request`, approval semantics не смешиваются с user-dialog semantics, а core interaction-domain остаётся channel-neutral.
- Vision-пакет по Issue `#378` фиксирует mission, KPI/guardrails, persona outcomes и continuity issue `#383` для `run:prd`.
- PRD-пакет по Issue `#383` фиксирует user stories, FR/AC/NFR, wave priorities, wait-state/correlation guardrails и создаёт architecture continuity issue `#385`; Telegram и другие channel adapters остаются отдельным последовательным потоком после проработки core platform contracts.
- Architecture package по Issue `#385` закрепляет ownership split между `control-plane`, `worker`, `api-gateway` и future adapters, отделяет interaction-domain от approval flow и создаёт design continuity issue `#387`.
- Design package по Issue `#387` фиксирует typed tool/callback contracts, interaction aggregate, wait-state taxonomy, deterministic resume payload и migration policy, после чего создаёт plan continuity issue `#389`.
- Plan package по Issue `#389` декомпозирует execution backlog на issues `#391..#395`, фиксирует wave-sequencing, DoR/DoD и quality gates для перехода в `run:dev`.

## Scope спринта
### In scope
- Полная doc-stage цепочка `intake -> vision -> prd -> arch -> design -> plan` для инициативы built-in MCP user interactions.
- Формализация продуктовой модели для:
  - `user.notify` как non-blocking user notification tool;
  - `user.decision.request` как typed request/response tool для сценариев, где действительно нужен ответ пользователя;
  - interaction-domain, который отделён от approval/control domain;
  - channel-neutral outbound/inbound contracts для будущих adapter integrations;
  - wait-state, correlation, retry, idempotency и audit guardrails.
- Создание последовательных follow-up issue без автоматической постановки `run:*`-лейблов.
- Формирование execution backlog `run:dev` по сервисным границам `control-plane`, `worker`, `api-gateway`, `agent-runner` и observability.

### Out of scope
- Кодовая реализация до завершения и утверждения `run:plan`.
- Telegram adapter, голос/STT, channel-specific UX affordances и расширенные conversation threads в рамках Sprint S10.
- Добавление нового MCP server entry в runtime `config.toml` вместо расширения существующего `mcp_servers.codex_k8s`.
- Попытка использовать `owner.feedback.request` как замену отдельному interaction-domain.

## Рекомендованный launch profile
- Базовый launch profile: `feature`.
- Обязательная эскалация:
  - `vision` обязателен, потому что инициатива вводит новую user-facing capability и требует measurable KPI;
  - `arch` обязателен, потому что затрагиваются service boundaries, callback contracts, wait-state lifecycle, audit и adapter responsibilities.
- Целевая continuity-цепочка:
  `#360 (intake) -> #378 (vision) -> #383 (prd) -> #385 (arch) -> #387 (design) -> #389 (plan) -> #391/#392/#393/#394/#395 (dev waves) -> qa -> release -> postdeploy -> ops`.

## План этапов и handover

| Stage | Основной артефакт | Целевая роль | Правило выхода |
|---|---|---|---|
| Intake (`#360`) | Problem/Brief/Scope/Constraints + intake AC | `pm` | Owner review intake-пакета и создана issue следующего этапа |
| Vision (`#378`) | Mission, persona outcomes, KPI, MVP/Post-MVP границы | `pm` | Зафиксирован vision baseline и создана issue `#383` (`run:prd`) |
| PRD (`#383`) | User stories, FR/AC/NFR, delivery evidence expectations | `pm` + `sa` | Подтверждён PRD package и создана issue `#385` (`run:arch`) |
| Architecture (`#385`) | Service boundaries, ownership матрица, interaction lifecycle | `sa` | Подтверждены архитектурные границы и создана issue `#387` (`run:design`) |
| Design (`#387`) | API/data/wait-state/adapter contracts и rollout notes | `sa` + `qa` | Подготовлен implementation-ready design package и создана issue `#389` для `run:plan` |
| Plan (`#389`) | Delivery waves, quality-gates, execution issues, DoR/DoD | `em` + `km` | Сформирован execution package `#391..#395` и owner-managed handover в `run:dev` |

## Guardrails спринта
- Инициатива расширяет существующий built-in MCP server `codex_k8s`; новый runtime server block не вводится.
- Approval flow и user interaction flow остаются отдельными доменными контурами с разными семантиками и wait-state правилами.
- `user.notify` не создаёт paused/waiting state; `user.decision.request` блокирует run только там, где ответ пользователя действительно обязателен.
- Dispatch, retries, idempotency, audit и callback correlation остаются responsibility platform domain, а не agent pod.
- Будущие channel adapters работают поверх typed interaction contract и не переопределяют core semantics.
- Telegram остаётся отдельным follow-up stream после завершения Sprint S10 doc-stage цепочки.

## Handover
- Текущий stage in-review: `run:plan` в Issue `#389`.
- Design package:
  - `docs/delivery/epics/s10/epic-s10-day5-mcp-user-interactions-design.md`;
  - `docs/architecture/initiatives/s10_mcp_user_interactions/README.md`;
  - `docs/architecture/initiatives/s10_mcp_user_interactions/design_doc.md`;
  - `docs/architecture/initiatives/s10_mcp_user_interactions/api_contract.md`;
  - `docs/architecture/initiatives/s10_mcp_user_interactions/data_model.md`;
  - `docs/architecture/initiatives/s10_mcp_user_interactions/migrations_policy.md`.
- Plan package:
  - `docs/delivery/epics/s10/epic-s10-day6-mcp-user-interactions-plan.md`;
  - execution issues `#391`, `#392`, `#393`, `#394`, `#395`.
- Следующий stage: `run:dev` по waves `#391 -> #392 -> #393/#394 -> #395`.
- Trigger-лейблы для execution issues не ставятся автоматически и остаются owner-managed переходом после review plan package.
