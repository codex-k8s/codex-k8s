---
doc_id: MIG-S7-CK8S-0001
type: migrations-policy
title: "Sprint S7 Day 5 — Migrations policy for MVP readiness gap closure (Issue #238)"
status: in-review
owner_role: SA
created_at: 2026-03-02
updated_at: 2026-03-02
related_issues: [212, 218, 220, 222, 238, 241]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-02-issue-238-migrations-policy"
---

# DB Migrations Policy: Sprint S7 MVP readiness gap closure

## TL;DR
- Подход: staged expand-contract с акцентом на idempotent runtime transitions.
- Владелец схемы: `services/internal/control-plane`.
- Путь миграций: `services/internal/control-plane/cmd/cli/migrations/*.sql`.
- Rollback: write-path rollback + сохранение additive схемы; destructive down-migration не является default.

## Размещение миграций и владелец схемы
- Schema owner: `services/internal/control-plane`.
- Все SQL-миграции для потоков S7 размещаются только в owner service.
- Shared DB ownership не меняется: внешний/staff edge не создаёт собственных migration paths.

## Потоки и миграционная обязательность
| Stream | Нужна миграция | Политика |
|---|---|---|
| `S7-E06` | нет (обязательно) | де-сcope выполняется contract/domain слоем без DDL |
| `S7-E07` | optional | при необходимости добавить soft-normalization `prompt_templates.source` |
| `S7-E09` | нет | UI-only |
| `S7-E10` | да | additive columns/indexes в `runtime_deploy_tasks` |
| `S7-E13` | нет | payload schema в `flow_events` без DDL |
| `S7-E16` | да | additive columns в `agent_runs` для terminalization metadata |
| `S7-E17` | да | additive columns/indexes в `agent_sessions` для snapshot version/checksum |

## Принципы
- Expand first: только additive DDL до подтверждения runtime stability.
- Idempotency first: миграции не должны ломать повторные action/callback сценарии.
- Zero-downtime target: без длительных table locks в production path.
- Rollout order фиксирован: `migrations -> internal -> edge -> frontend`.

## Процесс миграции (run:dev target)
1. Expand schema:
   - `runtime_deploy_tasks`: добавить action/finalization metadata поля.
   - `agent_runs`: добавить terminal metadata поля.
   - `agent_sessions`: добавить snapshot version/checksum поля.
2. Backfill:
   - заполнить default значения (`terminal_event_seq=0`, `snapshot_version=1` для существующих строк);
   - нормализовать `snapshot_checksum` для непустых snapshot payload.
3. Index hardening:
   - индексы для latest snapshot lookup и terminalization checks.
4. Switch writes:
   - включить новый write-path (cancel/stop actions, normalized finalization, versioned snapshot upsert).
5. Contract cleanup (optional):
   - при подтверждённой стабильности убрать legacy assumptions в code-path.

## Как выполняются миграции при деплое
- Production/prod порядок:
  1. stateful dependencies готовы;
  2. migration job (`goose`) под advisory lock;
  3. запуск обновлённых internal services;
  4. edge/frontend rollout.
- При ошибке миграции rollout останавливается до старта новых pod версий.

## Политика backfill
- Batch strategy: 500-1000 rows per transaction для снижения lock contention.
- Backfill идемпотентный и restart-safe.
- Прогресс фиксируется в migration logs и flow events.

## Политика rollback
- Разрешено откатывать:
  - новые write paths (feature-flag/config);
  - non-destructive индексы при необходимости.
- Не рекомендуется удалять:
  - additive columns с уже записанными audit/terminal metadata.
- Стратегия rollback:
  1. выключить новый write-path;
  2. сохранить данные и вернуть старый execution path;
  3. выполнить corrective forward migration при необходимости.

## Проверки
### Pre-migration checks
- Нет дублирующихся active runtime deploy tasks на `run_id`.
- Нет нарушений enum/status значений в `agent_runs`.
- Snapshot payload format валиден для checksum backfill.

### Post-migration verification
- Повторный `cancel/stop` возвращает idempotent result.
- false-failed regression по `run:intake:revise` не воспроизводится.
- Snapshot read/write проходит с version/checksum consistency.
- Записи `flow_events` создаются для action/finalization/snapshot событий.

## Runtime impact / Migration impact
- Runtime impact (`run:design`): отсутствует.
- Migration impact (`run:dev`): moderate, additive updates в трёх core таблицах + index hardening.

## Context7 baseline
- `/getkin/kin-openapi`: подтверждён baseline runtime validation path для transport contracts.
- `/microsoft/monaco-editor`: подтверждён baseline diff editor API для UI verification flows.
- Новых migration/dependency инструментов не требуется.

## Апрув
- request_id: owner-2026-03-02-issue-238-migrations-policy
- Решение: pending
- Комментарий: Ожидается review migration/rollback правил перед `run:plan`.
