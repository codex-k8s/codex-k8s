---
doc_id: MIG-APT-CK8S-0001
type: migrations-policy
title: "codex-k8s — DB migrations policy: prompt templates lifecycle"
status: in-review
owner_role: SA
created_at: 2026-02-25
updated_at: 2026-02-25
related_issues: [184, 185, 187, 189, 195, 197]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-25-issue-195-migrations-policy"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-25
---

# DB Migrations Policy: prompt templates lifecycle

## TL;DR
- Подход: staged expand-contract c явной проверкой инвариантов.
- Владелец схемы/миграций: `services/internal/control-plane`.
- Миграции размещаются в `services/internal/control-plane/cmd/cli/migrations/*.sql`.
- Rollback: ограниченный rollback с сохранением исторических данных, без destructive delete.

## Размещение миграций и владелец схемы
- Schema owner: `services/internal/control-plane`.
- Все DDL в owner service migrations directory:
  - `services/internal/control-plane/cmd/cli/migrations/*.sql`.
- Shared DB без owner не допускается.

## Принципы
- Expand first: сначала добавляем поля/индексы и backfill, затем переключаем writes.
- Zero-downtime target: DDL совместим с online rollout.
- Backward compatibility не является целевой гарантией инициативы, но rollout должен быть согласованным (`migrations -> internal -> edge -> frontend`).

## Процесс миграции (план для `run:dev`)
1. Expand:
   - добавить в `prompt_templates` поля `status`, `checksum`, `change_reason`, `supersedes_version`, `updated_by`, `updated_at`, `activated_at`, `metadata`.
2. Backfill:
   - заполнить `status`/`checksum`/`updated_at` для исторических строк;
   - нормализовать `is_active <-> status`.
3. Index hardening:
   - создать partial unique index для active версии;
   - создать индекс по key+version для version list и conflict checks.
4. Switch writes:
   - включить write-path с `expected_version` и `status` transitions.
5. Contract cleanup (optional post-stabilization):
   - удалить legacy assumptions в коде, при необходимости скорректировать `is_active` usage.

## Политика seed bootstrap из репозитория (embed)
- Seed-файлы `services/jobs/agent-runner/internal/runner/promptseeds/*.md` не удаляются и не мигрируются "вместо БД"; они остаются baseline/fallback слоем.
- Bootstrap seed->DB выполняется как отдельный шаг после DDL/backfill (не SQL-миграция):
  1. `dry-run`: вычислить diff между embed seeds и текущими DB-записями.
  2. `apply`: создать только отсутствующие baseline записи (обычно global scope), не перезаписывая project overrides.
  3. Зафиксировать `source`/`checksum` и аудит-событие загрузки.
- Если bootstrap не выполнен или выполнен частично, runtime продолжает работать через fallback на embed seeds.
- Любой rollout запрещает destructive действие по seed-слою: удаление/обнуление seed-файлов не допускается.

## Как выполняются миграции при деплое
- Порядок production deploy:
  1. stateful dependencies ready;
  2. migration job;
  3. internal services;
  4. edge services;
  5. frontend.
- Concurrency control:
  - одиночный migration runner + advisory lock (`goose` baseline).
- Failure policy:
  - при ошибке migration rollout останавливается до старта новых service pods.

## Политика backfill
- Batch execution small-chunk (например, 500-1000 rows per transaction) для контроля lock-time.
- Backfill идемпотентный и restart-safe.
- Progress tracking: migration logs + row counters.

## Политика rollback
- Можно откатывать:
  - индексы и неиспользуемые новые constraints (до switch writes).
- Нельзя безопасно откатывать:
  - исторические версии/аудитные записи после перехода на новый write-path.
- Rollback strategy:
  1. выключить новый write-path feature flag;
  2. оставить read-path на validated active versions;
  3. выполнить corrective migration только additive way;
  4. при проблемах seed bootstrap использовать fallback на embed seeds без удаления исторических DB-версий.

## Проверки
### Pre-migration checks
- Проверка дубликатов active версии на template key.
- Проверка null/invalid locale/template_kind/status значений.
- Проверка доступности rollback flag в config.

### Post-migration verification
- Инвариант active uniqueness соблюден.
- Backfill completed 100%.
- Smoke query latency в пределах design target.
- Ошибки `conflict`/`failed_precondition` не выходят за baseline.

## Runtime impact / Migration impact
- Runtime impact (`run:design`): отсутствует.
- Migration impact (`run:dev`): moderate, затрагивает существующую таблицу `prompt_templates` и индексацию `flow_events` queries.

## Context7 dependency check
- Дополнительные migration tooling библиотеки не требуются.
- Existing stack coverage подтверждена через Context7:
  - `kin-openapi` для contract validation;
  - `monaco-editor` для diff UX.

## Апрув
- request_id: owner-2026-02-25-issue-195-migrations-policy
- Решение: approved
- Комментарий: Миграционная политика готова к реализации в `run:dev`.
