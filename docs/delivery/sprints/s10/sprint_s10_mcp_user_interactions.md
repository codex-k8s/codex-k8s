---
doc_id: SPR-CK8S-0010
type: sprint-plan
title: "Sprint S10: Built-in MCP user interactions and channel-neutral adapter contracts (Issue #360)"
status: in-review
owner_role: PM
created_at: 2026-03-12
updated_at: 2026-03-12
related_issues: [360, 378]
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
- Intake-пакет создаёт continuity issue `#378` для `run:vision`; Telegram и другие channel adapters остаются отдельным последовательным потоком после проработки core platform contracts.

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
  `#360 (intake) -> #378 (vision) -> prd -> arch -> design -> plan -> dev -> qa -> release -> postdeploy -> ops`.

## План этапов и handover

| Stage | Основной артефакт | Целевая роль | Правило выхода |
|---|---|---|---|
| Intake (`#360`) | Problem/Brief/Scope/Constraints + intake AC | `pm` | Owner review intake-пакета и создана issue следующего этапа |
| Vision (`#378`) | Mission, persona outcomes, KPI, MVP/Post-MVP границы | `pm` | Зафиксирован vision baseline и создана issue `run:prd` |
| PRD | User stories, FR/AC/NFR, delivery evidence expectations | `pm` + `sa` | Подтверждён PRD package и создана issue `run:arch` |
| Architecture | Service boundaries, ownership матрица, interaction lifecycle | `sa` | Подтверждены архитектурные границы и создана issue `run:design` |
| Design | API/data/wait-state/adapter contracts и rollout notes | `sa` + `qa` | Подготовлен implementation-ready design package и создана issue `run:plan` |
| Plan | Delivery waves, quality-gates, execution issues, DoR/DoD | `em` + `km` | Сформирован execution package и backlog отдельных implementation issues |

## Guardrails спринта
- Инициатива расширяет существующий built-in MCP server `codex_k8s`; новый runtime server block не вводится.
- Approval flow и user interaction flow остаются отдельными доменными контурами с разными семантиками и wait-state правилами.
- `user.notify` не создаёт paused/waiting state; `user.decision.request` блокирует run только там, где ответ пользователя действительно обязателен.
- Dispatch, retries, idempotency, audit и callback correlation остаются responsibility platform domain, а не agent pod.
- Будущие channel adapters работают поверх typed interaction contract и не переопределяют core semantics.
- Telegram остаётся отдельным follow-up stream после завершения Sprint S10 doc-stage цепочки.

## Handover
- Завершённый текущий stage: `run:intake` в Issue `#360`.
- Следующий stage: `run:vision` в Issue `#378`.
- Trigger-лейбл для Issue `#378` не ставится автоматически и остаётся owner-managed переходом после review intake-пакета.
- Для всех следующих continuity issues Sprint S10 действует обязательный нейминг из
  `docs/delivery/development_process_requirements.md`:
  `S10 Day<день>: <Stage> для <краткая цель>` для doc-stage issue и
  `S10 Day7: Dev — <краткая реализация>` для implementation stage.
