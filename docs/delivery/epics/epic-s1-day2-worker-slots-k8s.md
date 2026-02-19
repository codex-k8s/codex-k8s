---
doc_id: EPC-CK8S-S1-D2
type: epic
title: "Epic Day 2: Worker run loop, slots, Kubernetes jobs"
status: completed
owner_role: EM
created_at: 2026-02-06
updated_at: 2026-02-09
related_issues: [1]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-19-full-docset"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-19
---

# Epic Day 2: Worker run loop, slots, Kubernetes jobs

## TL;DR
- Цель эпика: реализовать execution loop из `pending` run в реальный Kubernetes workload.
- Ключевая ценность: первый рабочий end-to-end контур исполнения.
- MVP-результат: worker берёт слот, создаёт Job/Pod, обновляет статусы и освобождает слот.

## Priority
- `P0` (ядро исполнения задач).

## Ожидаемые артефакты дня
- Реализация worker run-loop и reconciliation в `services/jobs/worker/**`.
- Kubernetes launcher/adapter через `client-go` в `libs/go/**` и/или `services/internal/control-plane/**`.
- Обновлённые state transitions для `agent_runs` и slot lifecycle в БД.
- E2E smoke сценарий webhook -> run -> worker -> job completion на production.

## Контекст
- Почему эпик нужен: без этого webhook ingress не приводит к фактическому выполнению задач.
- Связь с требованиями: FR-001, FR-012, FR-014, NFR-002.

## Scope
### In scope
- Poll/claim `pending` run из БД.
- Slot lease lifecycle (`free -> leased -> releasing -> free`).
- Создание Kubernetes Job/Pod через `client-go`.
- Обновление `agent_runs.status` и событий в `flow_events`.

### Out of scope
- Сложные scheduling policy beyond slots.
- Расширенная поддержка нескольких кластеров.

## Декомпозиция (Stories/Tasks)
- Story-1: worker polling and locking strategy.
- Story-2: slot lease алгоритм и TTL.
- Story-3: Kubernetes job launcher и reconciliation.
- Story-4: run status transitions + failure handling.

## Data model impact (по шаблону data_model.md)
- Сущности:
  - `slots`: lease поля используются как источник правды.
  - `agent_runs`: статусы `pending/running/succeeded/failed/canceled`.
  - `flow_events`: события жизненного цикла запуска.
- Связи/FK:
  - `slots.project_id -> projects.id`.
  - `agent_runs.project_id -> projects.id`, `agent_runs.agent_id -> agents.id`.
- Индексы и запросы:
  - Проверить/добавить `agent_runs(status, started_at)`.
  - Проверить/добавить `slots(project_id, state)`.
- Миграции:
  - Добавить enum/check для статусов, если ещё не зафиксированы.
- Retention/PII:
  - Хранить только технический runtime context, без секретов.

## Критерии приемки эпика
- Worker обрабатывает `pending` run и переводит его в финальный статус.
- Слот корректно освобождается при успехе и ошибке.
- После merge изменения задеплоены на production и пройден e2e smoke.

## Evidence
- Day2 migration добавлена:
  - `services/internal/control-plane/cmd/cli/migrations/20260207093000_day2_worker_slots_and_status.sql`
- Реализация worker run-loop добавлена:
  - `services/jobs/worker/internal/domain/worker/service.go`
  - `services/jobs/worker/internal/repository/postgres/runqueue/repository.go`
  - `services/jobs/worker/internal/repository/postgres/runqueue/sql/*.sql`
  - `services/jobs/worker/internal/repository/postgres/flowevent/repository.go`
- Kubernetes Job launcher через `client-go` добавлен:
  - `libs/go/k8s/joblauncher/launcher.go`
  - `services/jobs/worker/internal/clients/kubernetes/launcher/adapter.go`
- Worker service wiring добавлен:
  - `services/jobs/worker/internal/app/config.go`
  - `services/jobs/worker/internal/app/app.go`
  - `services/jobs/worker/cmd/worker/main.go`
- Production deploy и bootstrap синхронизированы под worker:
  - `deploy/base/codex-k8s/app.yaml.tpl`
  - `services/internal/control-plane/internal/domain/runtimedeploy/service_defaults.go`
  - `deploy/base/codex-k8s/codegen-check-job.yaml.tpl`
  - `cmd/codex-bootstrap/internal/cli/github_sync.go`
  - `bootstrap/host/bootstrap_remote_production.sh`
  - `bootstrap/host/config.env.example`
- Unit tests:
  - `services/jobs/worker/internal/domain/worker/service_test.go`
  - `services/jobs/worker/internal/app/config_test.go`
- Verification commands:
  - `go test ./...`
  - `go test ./cmd/codex-bootstrap/internal/cli ./services/internal/control-plane/cmd/runtime-deploy`

## Риски/зависимости
- Зависимости: стабильный доступ worker к Kubernetes API.
- Риск: race conditions при параллельных worker pod.

## План релиза (верхний уровень)
- По завершению дня провести ручной прогон минимум 3 запусков подряд на production.

## Апрув
- request_id: owner-2026-02-06-day2
- Решение: approved
- Комментарий: Day 2 scope принят.
