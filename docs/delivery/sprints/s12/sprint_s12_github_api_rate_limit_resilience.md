---
doc_id: SPR-CK8S-0012
type: sprint-plan
title: "Sprint S12: GitHub API rate-limit resilience, wait-state UX and MCP backpressure (Issue #366)"
status: completed
owner_role: PM
created_at: 2026-03-13
updated_at: 2026-03-13
related_issues: [366, 413, 416, 418, 420, 423, 425, 426, 427, 428, 429, 430, 431]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-03-13-issue-366-intake"
  approved_by: "ai-da-stas"
  approved_at: 2026-03-13
---

# Sprint S12: GitHub API rate-limit resilience, wait-state UX and MCP backpressure (Issue #366)

## TL;DR
- Цель спринта: убрать непрозрачные `403/429`-провалы при исчерпании GitHub API budget и заменить их на управляемый, audit-safe controlled wait-state для platform path и agent path.
- Sprint S12 стартовал с intake-этапа в Issue `#366`; vision-пакет оформлен в Issue `#413`, PRD-пакет оформлен в Issue `#416`, architecture continuity зафиксирована в Issue `#418`, design package оформлен в Issue `#420`, а execution package `run:plan` подготовлен в Issue `#423` с handover issues `#425..#431`.
- Документный контур `intake -> vision -> prd -> arch -> design -> plan` согласован и завершён; дальнейший handover в реализацию идёт через owner-managed execution waves `#425..#431`.
- Базовые ограничения спринта: сохраняется split `platform PAT` vs `agent bot-token`, GitHub остаётся текущим provider baseline, controlled wait не должен маскировать реальные auth/authz ошибки, а агент не должен уходить в бесконечные локальные retry-loop.

## Scope спринта
### In scope
- Полная doc-stage цепочка `intake -> vision -> prd -> arch -> design -> plan` для инициативы GitHub API rate-limit resilience.
- Формализация продуктовой модели для:
  - budget-aware wait-state по двум контурам (`platform PAT`, `agent bot-token`);
  - различения primary vs secondary rate-limit ситуаций на уровне product semantics;
  - owner/operator transparency в UI/service-comment/notifications;
  - MCP backpressure handoff для agent path;
  - безопасного resume после снятия лимита.
- Создание последовательных follow-up issue без `run:*`-лейблов; после `run:plan` Owner управляет переходом в execution waves `#425..#431`.

### Out of scope
- Кодовая реализация до завершения и утверждения `run:plan`.
- Обобщённый rate-limit framework для всех внешних провайдеров до подтверждения GitHub-first baseline.
- Редизайн всех retry/backoff политик платформы вне rate-limit сценариев.
- Изменение security-модели токенов, секретов или RBAC вне того, что нужно для формализации требований.

## Рекомендованный launch profile
- Базовый launch profile: `feature`.
- Обязательная эскалация:
  - `vision` обязателен, потому что инициатива меняет пользовательский опыт ожидания и операционную прозрачность;
  - `arch` обязателен, потому что затрагиваются service boundaries, persisted wait-state, MCP/runtime semantics, notifications и staff UX.
- Целевая continuity-цепочка:
  `#366 (intake) -> #413 (vision) -> #416 (prd) -> #418 (arch) -> #420 (design) -> #423 (plan) -> #425/#426/#427/#428/#429/#430/#431 (dev waves) -> qa -> release -> postdeploy -> ops`.

## План этапов и handover

| Stage | Основной артефакт | Целевая роль | Правило выхода |
|---|---|---|---|
| Intake (`#366`) | Problem/Brief/Scope/Constraints + intake AC | `pm` | Owner review intake-пакета и создана issue следующего этапа |
| Vision (`#413`) | Mission, north star, persona outcomes, KPI/guardrails | `pm` | Зафиксирован vision baseline и создана issue `#416` для `run:prd` |
| PRD (`#416`) | User stories, FR/AC/NFR, edge cases и expected evidence | `pm` + `sa` | Подтверждён PRD package и создана issue `#418` для `run:arch` |
| Architecture (`#418`) | Service boundaries, ownership matrix, wait-state lifecycle | `sa` | Подтверждены архитектурные границы и создана issue `#420` для `run:design` |
| Design (`#420`) | API/data/notification/runtime contracts и rollout notes | `sa` + `qa` | Подготовлен implementation-ready design package и создана issue `#423` для `run:plan` |
| Plan (`#423`) | Delivery waves, execution issues, DoR/DoD, quality-gates | `em` + `km` | Сформирован execution package `#425..#431` и owner-managed handover в `run:dev` |

## Execution package (S12-E01..S12-E07)

| Stream | Implementation issue | Wave | Scope |
|---|---|---|---|
| `S12-E01` | `#425` | Wave 1 | Wait-state persistence, additive schema и dominant wait linkage |
| `S12-E02` | `#426` | Wave 2 | `control-plane` classification, visibility projection и resume policy |
| `S12-E03` | `#427` | Wave 3 | `worker` auto-resume sweeps, bounded retries и escalation |
| `S12-E04` | `#428` | Wave 4 | `agent-runner` handoff и deterministic resume payload |
| `S12-E05` | `#429` | Wave 5 | `api-gateway` typed visibility transport и additive contracts |
| `S12-E06` | `#430` | Wave 6 | `web-console` wait queue, run visibility и contour attribution |
| `S12-E07` | `#431` | Wave 7 | Observability, rollout gates и readiness evidence перед `run:qa` |

## Guardrails спринта
- `platform PAT` и `agent bot-token` остаются разными operational contour; UI и audit должны показывать, какой именно контур упёрся в лимит.
- Controlled wait-state предпочтительнее ложного `failed`, но он не должен скрывать реальные permission/auth/token-invalid ошибки.
- GitHub Docs на 2026-03-13 подтверждают отдельные primary и secondary rate-limit semantics; продукт не должен зависеть от одного фиксированного численного порога как invariant.
- Agent path обязан использовать backpressure/controlled wait вместо бесконечных локальных retry при `429`/secondary limit.
- Notification path, owner feedback и UI transparency рассматриваются как часть одной инициативы только в пределах GitHub rate-limit resilience, а не как общий redesign alerting платформы.
- Doc-stage контур Sprint S12 завершён markdown-only пакетом; кодовая реализация начинается только по owner-managed waves `#425..#431`.

## Handover
- Текущий owner review касается plan-package в Issue `#423`, но operational handover уже декомпозирован в execution waves `#425 -> #426 -> #427 -> #428 -> #429 -> #430 -> #431`.
- Документный контур `intake -> vision -> prd -> arch -> design -> plan` согласован и завершён в Issue `#423`.
- Следующий operational handover: owner-managed execution waves `#425 -> #426 -> #427 -> #428 -> #429 -> #430 -> #431`.
- Переход в `run:qa` допускается только после закрытия `#431` и подтверждённого observability/readiness evidence.
- На execution waves нельзя терять Day5 design decisions:
  - `waiting_backpressure` остаётся отдельным coarse wait-state для GitHub rate-limit;
  - `control-plane` сохраняет ownership wait aggregate, visibility projection и dominant wait election;
  - `worker` остаётся владельцем finite auto-resume sweeps и manual escalation;
  - `agent-runner` работает только через typed signal handoff и deterministic resume payload;
  - staff/private surfaces остаются каноничными, а GitHub service-comment остаётся best-effort mirror;
  - `#431` остаётся обязательным observability/readiness gate перед `run:qa`;
  - predictive budgeting, multi-provider governance и adapter-specific notifications не блокируют core Sprint S12.
