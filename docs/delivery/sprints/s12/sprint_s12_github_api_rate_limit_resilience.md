---
doc_id: SPR-CK8S-0012
type: sprint-plan
title: "Sprint S12: GitHub API rate-limit resilience, wait-state UX and MCP backpressure (Issue #366)"
status: in-review
owner_role: PM
created_at: 2026-03-13
updated_at: 2026-03-13
related_issues: [366, 413, 416, 418, 420]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-13-issue-366-intake"
---

# Sprint S12: GitHub API rate-limit resilience, wait-state UX and MCP backpressure (Issue #366)

## TL;DR
- Цель спринта: убрать непрозрачные `403/429`-провалы при исчерпании GitHub API budget и заменить их на управляемый, audit-safe controlled wait-state для platform path и agent path.
- Sprint S12 стартовал с intake-этапа в Issue `#366`; vision-пакет оформлен в Issue `#413`, PRD-пакет оформлен в Issue `#416`, architecture continuity зафиксирована в Issue `#418`, а handover в `run:design` подготовлен в Issue `#420`.
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
- Создание последовательных follow-up issue без `run:*`-лейблов; после `run:plan` Owner управляет переходом в execution waves.

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
  `#366 (intake) -> #413 (vision) -> #416 (prd) -> #418 (arch) -> #420 (design) -> plan -> dev -> qa -> release -> postdeploy -> ops`.

## План этапов и handover

| Stage | Основной артефакт | Целевая роль | Правило выхода |
|---|---|---|---|
| Intake (`#366`) | Problem/Brief/Scope/Constraints + intake AC | `pm` | Owner review intake-пакета и создана issue следующего этапа |
| Vision (`#413`) | Mission, north star, persona outcomes, KPI/guardrails | `pm` | Зафиксирован vision baseline и создана issue `#416` для `run:prd` |
| PRD (`#416`) | User stories, FR/AC/NFR, edge cases и expected evidence | `pm` + `sa` | Подтверждён PRD package и создана issue `#418` для `run:arch` |
| Architecture (`#418`) | Service boundaries, ownership matrix, wait-state lifecycle | `sa` | Подтверждены архитектурные границы и создана issue `#420` для `run:design` |
| Design (`#420`) | API/data/notification/runtime contracts и rollout notes | `sa` + `qa` | Подготовлен implementation-ready design package и создана issue `run:plan` |
| Plan | Delivery waves, execution issues, DoR/DoD, quality-gates | `em` + `km` | Сформирован execution package и отдельные implementation issues |

## Guardrails спринта
- `platform PAT` и `agent bot-token` остаются разными operational contour; UI и audit должны показывать, какой именно контур упёрся в лимит.
- Controlled wait-state предпочтительнее ложного `failed`, но он не должен скрывать реальные permission/auth/token-invalid ошибки.
- GitHub Docs на 2026-03-13 подтверждают отдельные primary и secondary rate-limit semantics; продукт не должен зависеть от одного фиксированного численного порога как invariant.
- Agent path обязан использовать backpressure/controlled wait вместо бесконечных локальных retry при `429`/secondary limit.
- Notification path, owner feedback и UI transparency рассматриваются как часть одной инициативы только в пределах GitHub rate-limit resilience, а не как общий redesign alerting платформы.
- До `run:plan` инициатива остаётся markdown-only контуром.

## Handover
- Текущий stage в review: `run:arch` в Issue `#418`.
- Следующий stage: `run:design` в Issue `#420`.
- На `run:design` нельзя размывать архитектурные решения Day4:
  - controlled wait важнее ложного `failed`;
  - split `platform PAT` vs `agent bot-token` обязателен;
  - `control-plane` остаётся owner classification и visibility contract, а `worker` владеет finite auto-resume orchestration;
  - owner/operator должен видеть причину ожидания, affected contour, recovery hint и следующий шаг;
  - hard-failure path не может маскироваться под controlled wait;
  - агент не должен brute-force ретраить GitHub локально после typed rate-limit detection;
  - predictive budgeting, multi-provider governance и adapter-specific notifications не блокируют core Sprint S12.
