---
doc_id: EPC-CK8S-S6-D9
type: epic
title: "Epic S6 Day 9: Release closeout для lifecycle управления агентами и шаблонами промптов (Issues #262/#263)"
status: in-review
owner_role: EM
created_at: 2026-03-02
updated_at: 2026-03-02
related_issues: [184, 185, 187, 189, 195, 197, 199, 201, 216, 262, 263]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-02-issue-262-release-closeout"
---

# Epic S6 Day 9: Release closeout для lifecycle управления агентами и шаблонами промптов (Issues #262/#263)

## TL;DR
- Issue `#262` закрывает release-governance контур Sprint S6 и фиксирует, что поставка Day7/Day8 находится в production.
- Синхронизированы release-артефакты: quality-gates, definition of done, release notes и rollback/mitigation план.
- Для continuity создана follow-up issue `#263` в stage `run:postdeploy` (без trigger-лейбла).

## Контекст
- Stage continuity Sprint S6:
  `#184 -> #185 -> #187 -> #189 -> #195 -> #197 -> #199 -> #201 -> #216 -> #262 -> #263`.
- Входные артефакты:
  - `docs/delivery/sprints/s6/sprint_s6_agents_prompt_management.md`
  - `docs/delivery/epics/s6/epic_s6.md`
  - `docs/delivery/epics/s6/epic-s6-day6-agents-prompts-plan.md`
  - `docs/ops/production_runbook.md`
  - `docs/delivery/delivery_plan.md`

## Scope
### In scope
- Формальное закрытие `run:release` для Sprint S6 и фиксация release-readiness.
- Проверка delivery-governance: quality-gates, DoD, traceability и policy compliance.
- Подготовка handover в `run:postdeploy` через отдельную issue.

### Out of scope
- Кодовые изменения сервисов (в `run:release` разрешены только markdown-обновления).
- Выполнение самого этапа `run:postdeploy` и `run:ops`.
- Изменение архитектурных границ, API-контрактов и миграций.

## Quality gates (`run:release`)
| Gate | Что проверяем | Статус |
|---|---|---|
| QG-S6-RLS-01 Stage continuity | Цепочка stage-issues S6 закрыта без пропусков и с owner-governed trigger policy | passed (`#184 -> ... -> #262`) |
| QG-S6-RLS-02 QA readiness | QA дал решение GO, release инициирован без открытых P0-blockers | passed (`#201`) |
| QG-S6-RLS-03 Traceability sync | `delivery_plan`, Sprint/Epic S6, `issue_map`, `requirements_traceability` синхронизированы | passed |
| QG-S6-RLS-04 Rollout policy | Зафиксирован порядок выкладки `migrations -> internal -> edge -> frontend` | passed |
| QG-S6-RLS-05 Rollback readiness | Есть явный rollback/mitigation план с критериями активации | passed |
| QG-S6-RLS-06 Security/policy compliance | Scope ограничен markdown, без ослабления security/policy требований | passed |
| QG-S6-RLS-07 Next-stage handover | Создана issue `run:postdeploy` без trigger-лейбла и с обязательной инструкцией continuity | passed (`#263`) |

## Definition of Done package (release closeout)
- [x] Release governance пакет оформлен в markdown и привязан к issue `#262`.
- [x] Обновлены status/факты Sprint S6 в delivery-документах.
- [x] Зафиксированы release notes по итогам S6 (scope, качество, остаточные риски).
- [x] Зафиксирован rollback/mitigation план для post-release контроля.
- [x] Создана follow-up issue `#263` для `run:postdeploy` с инструкцией создать следующую issue `run:ops`.

## Release notes (S6 closeout summary)
- В production закреплён lifecycle `agents/templates/audit`, реализованный в Day7 (`#199`, PR `#202`).
- QA Day8 (`#201`) подтвердил readiness к release по контрактам, миграциям, UI flow и regression evidence.
- Release closeout Day9 (`#262`) завершил документационный и governance контур Sprint S6.
- Остаточный технический риск: pre-existing duplicate debt (`dupl-go`) остаётся под контролем и не блокирует postdeploy-проверку.

## Rollback and mitigation plan
1. При деградации после release применять rollback по политике: предыдущий стабильный image-tag + проверка migration-совместимости.
2. При росте ошибок/latency использовать staged mitigation:
   - ограничение нагрузки на affected endpoints;
   - откат edge/frontend при стабильном internal;
   - полный rollback при невыполнении SLO.
3. Верификация после rollback: health endpoints, ключевые user-flow `Agents`, логи `control-plane`/`worker`, audit trail в `flow_events`.

## Blockers, риски и owner decisions
| Тип | ID | Описание | Статус |
|---|---|---|---|
| blocker | BLK-262-01 | Требуется Owner-постановка trigger-лейбла `run:postdeploy` на issue `#263` | open |
| risk | RSK-262-01 | Остаточный duplicate debt (`dupl-go`) требует контроля в postdeploy/ops | monitoring |
| risk | RSK-262-02 | Возможна скрытая деградация при runtime-нагрузке вне QA-профиля | monitoring |
| owner-decision | OD-262-01 | Sprint S6 считается успешно завершённым по release-governance критериям | proposed |
| owner-decision | OD-262-02 | Postdeploy выполняется отдельным этапом по issue `#263` без расширения scope `#262` | proposed |

## Context7 validation
- Через Context7 (`/websites/cli_github_manual`) подтверждён актуальный синтаксис `gh issue edit` / `gh pr create` / `gh pr edit` для release PR-flow и label-transition fallback команд.
- Новые внешние зависимости в рамках `run:release` не добавлялись.

## Handover в `run:postdeploy`
- Follow-up issue: `#263` (создана без trigger-лейбла).
- Обязательное правило continuity:
  - после `run:postdeploy` создать issue следующего этапа `run:ops`;
  - приложить postdeploy evidence и owner decision по закрытию S6 operational tail.

## Связанные документы
- `docs/delivery/sprints/s6/sprint_s6_agents_prompt_management.md`
- `docs/delivery/epics/s6/epic_s6.md`
- `docs/delivery/delivery_plan.md`
- `docs/delivery/issue_map.md`
- `docs/delivery/requirements_traceability.md`
- `docs/ops/production_runbook.md`
