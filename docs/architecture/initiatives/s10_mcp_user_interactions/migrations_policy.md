---
doc_id: MIG-S10-CK8S-0001
type: migrations-policy
title: "Sprint S10 Day 5 — Migrations policy for built-in MCP user interactions (Issue #387)"
status: in-review
owner_role: SA
created_at: 2026-03-12
updated_at: 2026-03-12
related_issues: [360, 378, 383, 385, 387, 389]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-12-issue-387-migrations-policy"
---

# DB Migrations Policy: Sprint S10 built-in MCP user interactions

## TL;DR
- Подход: staged expand-contract с additive schema и explicit wait-taxonomy backfill.
- Инструменты миграций: `goose` в owner service `services/internal/control-plane`.
- Где лежат миграции и кто владелец схемы: `services/internal/control-plane/cmd/cli/migrations/*.sql`, schema owner остаётся `control-plane`.
- Политика откатов: write-path rollback и disable tool exposure; additive tables/columns по умолчанию не удаляются.

## Размещение миграций и владелец схемы
- Schema owner: `services/internal/control-plane`.
- Все SQL-миграции для interaction-domain размещаются только внутри owner service.
- `api-gateway`, `worker` и adapters не создают свои migration paths и не владеют interaction schema.

## Потоки и миграционная обязательность
| Stream | Нужна миграция | Политика |
|---|---|---|
| `interaction_requests` aggregate | да | новая таблица |
| `interaction_delivery_attempts` | да | новая таблица + retry indexes |
| `interaction_callback_events` | да | новая таблица + dedupe unique index |
| `interaction_response_records` | да | новая таблица + partial unique effective-response index |
| `agent_runs` wait linkage | да | additive columns + wait_reason backfill |
| `agent_sessions` | нет (обязательно) | reuse existing snapshot storage без DDL |
| `flow_events` payload schema | нет | event vocabulary обновляется в коде, без schema rewrite |

## Принципы
- Expand first: новые interaction tables и wait linkage добавляются до переключения write paths.
- Approval-safe rollout: approval flow не должен ломаться из-за backfill `wait_reason`.
- Zero-downtime target: без destructive lock-heavy DDL в рабочем контуре.
- Rollout order неизменен: `migrations -> control-plane -> worker -> api-gateway -> adapters`.

## Процесс миграции (run:dev target)
1. Expand schema:
   - создать `interaction_requests`;
   - создать `interaction_delivery_attempts`;
   - создать `interaction_callback_events`;
   - создать `interaction_response_records`;
   - добавить в `agent_runs` поля `wait_target_kind`, `wait_target_ref`, `wait_deadline_at`.
2. Backfill wait taxonomy:
   - заменить legacy `wait_reason='mcp'` на `approval_pending`;
   - для open approval waits заполнить `wait_target_kind=approval_request`, `wait_target_ref=<approval id>`, если source доступен;
   - decision interactions backfill не требуют, так как это новый traffic path.
3. Index hardening:
   - partial index для open decision waits;
   - unique dedupe index для callback events;
   - partial unique index effective response.
4. Switch writes:
   - включить `control-plane` write path interaction aggregate;
   - включить `worker` dispatch/retry/expiry;
   - затем открыть callback ingress в `api-gateway`.
5. Contract cleanup:
   - после стабилизации удалить internal assumptions о generic `wait_reason=mcp`;
   - approval flow остаётся отдельным callback family.

## Как выполняются миграции при деплое
- Production/prod стратегия:
  1. stateful dependencies готовы;
  2. migration job (`goose`) под advisory lock;
  3. rollout нового `control-plane`;
  4. rollout `worker`;
  5. rollout `api-gateway`;
  6. активация adapter traffic.
- При ошибке миграции rollout останавливается до старта новых pod версий.

## Политика backfill
- Backfill нужен только для `agent_runs.wait_reason` и typed wait linkage approval path.
- Batch strategy: 200-500 rows per transaction для минимизации contention.
- Backfill restart-safe:
  - повторный прогон не создаёт drift;
  - rows уже переведённые в `approval_pending` пропускаются.
- `agent_sessions` и approval records не переписываются.

## Политика rollback
- Разрешено откатывать:
  - exposure новых MCP tools;
  - callback ingress path в `api-gateway`;
  - worker dispatch/retry execution path.
- Не рекомендуется удалять:
  - additive interaction tables и columns;
  - callback evidence, уже записанный в audit trail.
- Стратегия rollback:
  1. выключить новые tools в effective MCP catalog;
  2. прекратить adapter dispatch и callback admission для interaction family;
  3. сохранить additive schema и выполнить corrective forward fix при необходимости.

## Проверки
### Pre-migration checks
- Нет неизвестных значений `agent_runs.wait_reason` вне `owner_review|mcp|null`.
- Открытые `waiting_mcp` runs относятся к approval path и могут быть backfill-нуты в `approval_pending`.
- В БД отсутствуют потенциальные legacy interaction tables с конфликтующим именованием.

### Post-migration verification
- `user.notify` не создаёт wait target.
- `user.decision.request` создаёт `interaction_request` + `agent_runs.wait_target_kind=interaction_request`.
- Duplicate callback не создаёт второй effective response.
- Expiry path переводит interaction в terminal state и schedule-ит resume без повторного logical completion.
- Approval flow продолжает работать с `wait_reason=approval_pending`.

## Runtime impact / Migration impact
- Runtime impact (`run:design`): отсутствует.
- Migration impact (`run:dev`): moderate, additive schema + wait-taxonomy backfill + new indexes.

## Context7 baseline
- Попытка использовать Context7 для `goose` и `kin-openapi` завершилась `Monthly quota exceeded`.
- Дополнительные migration/dependency инструменты не требуются.

## Апрув
- request_id: owner-2026-03-12-issue-387-migrations-policy
- Решение: pending
- Комментарий: Ожидается review rollout/backfill/rollback policy перед `run:plan`.
