---
doc_id: MIG-S18-MISSION-CONTROL-0001
type: migrations-policy
title: "Mission Control frontend-first canvas prototype — Migrations policy Sprint S18 Day 5"
status: in-review
owner_role: SA
created_at: 2026-04-01
updated_at: 2026-04-01
related_issues: [480, 561, 562, 563, 565, 567, 571, 573, 579]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-04-01-issue-573-migrations-policy"
---

# DB Migrations Policy: Mission Control frontend-first canvas prototype

## TL;DR
- Подход Sprint S18: no-op for DB/runtime migrations, because the prototype is frontend-only and fixture-backed.
- Инструменты миграций: не требуются в `run:dev`; `goose` under `control-plane` remains relevant only for deferred backend rebuild `#563`.
- Где лежат миграции и кто владелец схемы: новых owner/schemas нет; текущий schema owner `services/internal/control-plane` не получает Sprint S18 changes.
- Политика откатов: rollback only by reverting the `web-console` bundle.

## Размещение миграций и владелец схемы
- Sprint S18 does not add:
  - SQL migrations;
  - temp tables;
  - OpenAPI/proto changes;
  - generated DTO updates.
- Existing ownership rule remains:
  - future persisted Mission Control schema may be introduced only by `services/internal/control-plane`;
  - if `#563` needs migrations, they must live in `services/internal/control-plane/cmd/cli/migrations/*.sql`.
- Shared or temporary DB ownership for fake-data prototype is forbidden.

## Принципы
- Frontend-first isolation:
  - no backend schema or transport is introduced just to support the fake-data walkthrough.
- No shadow persistence:
  - browser memory and bundle fixtures are sufficient for Sprint S18; no temp JSONB rows or cache tables are allowed.
- No env-only product switches:
  - prototype enablement must happen by route implementation, not by runtime flags.
- Replacement seam preserved:
  - all future migration work stays explicitly deferred to `#563`.

## Процесс миграции (шаги)
1. No schema expand step for Sprint S18.
2. Implement frontend-only prototype source, store and presentational components in `web-console`.
3. Verify that OpenAPI/proto/generated transport artifacts remain unchanged.
4. Validate candidate walkthrough and keep backend migration scope deferred to `#563`.

## Как выполняются миграции при деплое
- Production/Prod strategy for Sprint S18:
  - no migration job is executed.
- Гарантия отсутствия параллельных миграций:
  - not applicable; `control-plane` migration runner is untouched.
- Поведение при ошибке миграции:
  - not applicable for Sprint S18.

## Политика backfill
- Как выполняем:
  - no backfill.
- Ограничение по скорости:
  - not applicable.
- Мониторинг прогресса:
  - not applicable.

## Политика rollback
- Когда можно rollback:
  - at any time, by reverting frontend code or the `web-console` image.
- Что нельзя откатить:
  - there is no persisted prototype data to roll back.
- План отката:
  - restore previous `MissionControlPage` implementation and remove prototype source wiring if the owner rejects the walkthrough.

## Проверки
### Pre-migration checks
- confirm that `run:dev` scope remains limited to `services/staff/web-console`;
- confirm that `api/server/api.yaml`, `proto/` and generated DTO are unchanged;
- confirm that no temp persistence is introduced in `control-plane`, `worker` or browser storage.

### Post-migration verification
- Mission Control route renders from bundle-local fixtures only;
- workflow preview remains read-only and references repo seeds;
- no migration job, DB access path or backend deploy dependency was added.

## Runtime impact / Migration impact
- Runtime impact at `run:design`: none.
- Expected runtime impact at `run:dev`: `web-console` only, no backend/service-order change.
- Migration impact: none for Sprint S18.

## Deferred rollout order for `#563`
- If backend rebuild later introduces persisted truth, rollout order must still stay:
  - `schema -> control-plane -> worker -> api-gateway -> web-console`
- That order is explicitly out of Sprint S18 scope and must not leak back as a hidden prerequisite now.

## Открытые вопросы
- Не требуется.

## Апрув
- request_id: `owner-2026-04-01-issue-573-migrations-policy`
- Решение: pending
- Комментарий: требуется owner review no-migration decision и strict deferral boundary к `#563`.
