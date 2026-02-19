---
doc_id: EPC-CK8S-S3-D19-5
type: epic
title: "Epic S3 Day 19.5: Realtime event bus (PostgreSQL LISTEN/NOTIFY) and WebSocket backplane"
status: completed
owner_role: EM
created_at: 2026-02-18
updated_at: 2026-02-19
related_issues: [19]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-19-full-docset"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-19
---

# Epic S3 Day 19.5: Realtime event bus (PostgreSQL LISTEN/NOTIFY) and WebSocket backplane

## TL;DR
- Выбран вариант реализации `#1`: PostgreSQL event log + `LISTEN/NOTIFY` как межсерверная шина для realtime.
- Цель: обеспечить push-апдейты с сервера на клиент при нескольких pod'ах без введения новой инфраструктуры.
- Результат: любой pod может публиковать событие, любой pod `api-gateway` получает его через общую шину и рассылает своим WebSocket-клиентам, включая live-логи и flow events.

## Priority
- `P0`.

## Выбранная архитектура
- Source of truth: таблица `realtime_events` в PostgreSQL.
- Транспорт пробуждения: PostgreSQL `NOTIFY` с `event_id`.
- Consumer-путь:
  1. producer (control-plane/worker/api-gateway) сохраняет событие в `realtime_events`;
  2. producer делает `NOTIFY codex_realtime, <event_id>`;
  3. каждый `api-gateway` pod слушает `LISTEN codex_realtime`, дочитывает событие по `event_id` и fanout'ит его в локальные WS-сессии.
- Recovery-path:
  - клиент хранит `last_event_id`;
  - при reconnect сервер отдает missed events из БД (`id > last_event_id`) до входа в live stream.

## Scope
### In scope
- Data model и migration:
  - таблица `realtime_events` (`id`, `topic`, `scope`, `payload_json`, `created_at`, `correlation_id`, `project_id`, `run_id`, `task_id` ...);
  - индексы для catch-up и topic/scope выборки;
  - TTL/cleanup policy.
- Realtime publisher API в backend:
  - typed publish contract;
  - redaction policy (никаких секретов в payload).
- WebSocket backplane в `api-gateway`:
  - WS endpoint, ping/pong, authz;
  - subscribe filters (project/run/deploy scope);
  - отдельные топики для статусов, логов и событий (`run.logs`, `run.events`, `deploy.logs`, `deploy.events`, `system.errors`);
  - delivery ack + resume by `last_event_id`.
- Multi-server guarantees:
  - at-least-once delivery на уровне шины;
  - idempotency ключ события (`event_id`) на клиенте.
  - последовательность событий внутри одного stream key (`run_id`/`task_id`) сохраняется по возрастанию `event_id`.

### Out of scope
- Отдельный message broker (Redis/NATS) в этой итерации.
- Exactly-once delivery гарантия.

## Декомпозиция
- Story-1: migration + repository/service для `realtime_events`.
- Story-2: publisher hooks из control-plane/worker/runstatus/runtime deploy.
- Story-3: `api-gateway` WS endpoint + listener loop (`LISTEN/NOTIFY` + catch-up).
- Story-4: log/event stream transport (incremental chunks, sequence markers, end-of-stream events).
- Story-5: cleanup policy и observability (метрики lag/drop/reconnect).

## Критерии приемки
- Событие, опубликованное на одном pod, видят WS-клиенты, подключенные к другому pod.
- При reconnect клиент получает пропущенные события через `last_event_id`.
- При кратковременном обрыве `LISTEN`/соединения событие не теряется (recover через БД log).
- Логи run/deploy и flow-events доставляются в WS поток через ту же шину без отдельного polling API.
- Payload не содержит секретов и проходит redaction policy.

## Риски/зависимости
- Риск роста таблицы `realtime_events`: нужен агрессивный cleanup/TTL.
- Риск дублирования сообщений: UI должен обрабатывать idempotency по `event_id`.
- Для high-throughput позже может потребоваться миграция на Redis/NATS (вынесено post-MVP).
