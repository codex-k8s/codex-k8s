---
doc_id: SPR-CK8S-0009
type: sprint-plan
title: "Sprint S9: Mission Control Dashboard and console control plane (Issue #333)"
status: in-review
owner_role: PM
created_at: 2026-03-12
updated_at: 2026-03-14
related_issues: [333, 335, 337, 340, 351, 363, 369, 370, 371, 372, 373, 374, 375]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-12-issue-333-intake"
---

# Sprint S9: Mission Control Dashboard and console control plane (Issue #333)

## TL;DR
- Цель спринта: превратить staff console из набора разрозненных operational screen'ов в единый control-plane UX для active work items, discussion, PR и агентов.
- Sprint S9 стартовал с intake-этапа в Issue `#333`; vision baseline зафиксирован в Issue `#335`, PRD baseline зафиксирован в Issue `#337`, architecture baseline зафиксирован в Issue `#340`.
- Базовые ограничения спринта: GitHub-first MVP, review человека в provider UI, webhook-driven orchestration, active-set UX вместо попытки показывать весь исторический граф.

## Scope спринта
### In scope
- Полная doc-stage цепочка `intake -> vision -> prd -> arch -> design -> plan` для инициативы Mission Control Dashboard.
- Формализация продуктовой модели для:
  - active-set dashboard;
  - discussion-first flow;
  - realtime состояния и side panel;
  - provider sync / webhook reconciliation / dedupe guardrails.
- Создание последовательных follow-up issue без `run:*`-лейблов; после `run:plan` Owner запускает implementation waves по issues `#369..#375`.

### Out of scope
- Кодовая реализация до завершения и утверждения `run:plan`.
- Произвольный dashboard/layout builder, полный исторический graph-view без active-set ограничений и замена human review в GitHub UI.
- GitLab parity в первой волне MVP.
- Преждевременный dependency lock-in для конкретных frontend/realtime библиотек до `run:arch`/`run:design`.

## Рекомендованный launch profile
- Базовый launch profile: `feature`.
- Обязательная эскалация:
  - `vision` обязателен, потому что инициатива меняет продуктовую миссию staff console и вводит новые KPI;
  - `arch` обязателен, потому что затрагиваются service boundaries, realtime contracts, persisted projections и webhook reconciliation.
- Целевая continuity-цепочка:
  `#333 (intake) -> #335 (vision) -> #337 (prd) -> #340 (arch) -> #351 (design) -> #363 (plan) -> dev -> qa -> release -> postdeploy -> ops`.

## План этапов и handover

| Stage | Основной артефакт | Целевая роль | Правило выхода |
|---|---|---|---|
| Intake (`#333`) | Problem/Brief/Scope/Constraints + intake AC | `pm` | Owner review intake-пакета и создана issue следующего этапа |
| Vision (`#335`) | Mission, KPI, persona model, MVP/Post-MVP границы | `pm` | Зафиксирован vision baseline и создана issue `run:prd` |
| PRD (`#337`) | User stories, FR/AC/NFR и wave priorities | `pm` + `sa` | Подтверждён PRD package и создана issue `run:arch` |
| Architecture (`#340`) | C4/ADR/boundary decisions по control-plane UX и realtime | `sa` | Подтверждены сервисные границы и создана issue `#351` для `run:design` |
| Design (`#351`) | API/data/realtime/design package | `sa` + `qa` | Подготовлен implementation-ready design package и создана issue `#363` для `run:plan` |
| Plan (`#363`) | Delivery waves, execution issues, DoR/DoD, quality-gates | `em` + `km` | Сформирован execution package и отдельные implementation issues `#369..#375` |

## Execution handover after `run:plan`

| Wave | Issues | Что открывает |
|---|---|---|
| Wave 1 | `#369` | Foundation schema, repository contracts и rollout guards, без которых запрещён core read/write exposure |
| Wave 2 | `#370` | `control-plane` domain owner для active-set projection и command lifecycle |
| Wave 3 | `#371`, `#372` | `#371` владеет warmup/backfill execution и reconcile/retry path, `#372` владеет только core thin-edge transport/realtime contracts |
| Wave 4 | `#373` | Dashboard shell, board/list toggle и side panel UX на typed contracts |
| Wave 5 | `#374` | Superseded 2026-03-14: отдельная observability/readiness wave не входит в активный Sprint S9 backlog |
| Wave 6 | `#375` | Conditional voice contour; запускается только отдельным owner decision |

## Guardrails спринта
- Dashboard остаётся control plane над существующей stage/label моделью и не создаёт обходов review/audit policy.
- MVP default view обязан работать от active set (`working`, `waiting`, `blocked`, `review`, `recent critical updates`), а не от полного архива.
- GitHub остаётся каноническим provider'ом MVP; GitLab допускается только как future-compatible continuation через provider abstraction.
- Voice intake рассматривается как условная последующая волна: не блокирует core dashboard wave, пока не подтверждены ROI, AI policy и operational readiness.
- Voice-specific OpenAPI/codegen не входят в core transport wave и активируются только через conditional issue `#375`.
- Owner revision 2026-03-14 вывел `#374` из active implementation chain; issue сохранена только как historical superseded artifact.

## Handover
- Текущий stage в review: `run:plan` в Issue `#363`.
- Следующий operational stage: owner-managed `run:dev` по implementation issues `#369..#373`.
- `#374` не считается активным handover issue после owner revision 2026-03-14.
- Conditional continuation: issue `#375` запускается только после подтверждения core Mission Control wave.
