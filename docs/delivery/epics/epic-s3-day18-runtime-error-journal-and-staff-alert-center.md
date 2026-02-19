---
doc_id: EPC-CK8S-S3-D18
type: epic
title: "Epic S3 Day 18: Runtime error journal and staff alert center"
status: completed
owner_role: EM
created_at: 2026-02-18
updated_at: 2026-02-19
related_issues: [19]
related_prs: [50]
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-19-full-docset"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-19
---

# Epic S3 Day 18: Runtime error journal and staff alert center

## TL;DR
- Цель: сделать управляемый контур ошибок платформы: все ошибки control-plane/jobs пишутся в отдельное хранилище и показываются в staff UI как стековые алерты.
- Результат: у оператора всегда есть видимые свежие ошибки (до 5 одновременно), закрытие алерта помечает ошибку как просмотренную.

## Priority
- `P0`.

## Scope
### In scope
- Новая таблица ошибок платформы (`runtime_errors`/эквивалент) с индексами для active feed.
- Error ingest в control-plane/worker/job pipelines:
  - level, source, correlation/run/task ids, payload, stack/trace snippet, timestamps, viewed_at/viewed_by.
- Staff API:
  - список активных ошибок (stacked feed, top-5 newest),
  - mark-as-viewed endpoint,
  - history/filter endpoint для дальнейшего расширения.
- Staff UI:
  - правый нижний alert stack (до 5),
  - dismiss -> mark viewed,
  - быстрый переход к деталям run/deploy/task.

### Out of scope
- Полная интеграция с внешними алертинг-системами (PagerDuty/Slack/etc.) в этой итерации.

## Декомпозиция
- Story-1: data model + migration + repository/service.
- Story-2: runtime error capture hooks (control-plane/jobs).
- Story-3: staff API + frontend alert center.
- Story-4: observability/docs и правила severity классификации.

## Критерии приемки
- Ошибки из control-plane/jobs записываются в отдельную таблицу с достаточным контекстом для дебага.
- Staff UI показывает стек из 5 свежих ошибок и обновляется без перезагрузки страницы.
- Закрытие алерта помечает запись как viewed и больше не показывает её в активном стеке.
- Для каждой ошибки есть ссылка в связанные сущности (run/deploy task/namespace/job) при наличии.

## Риски/зависимости
- Риск чрезмерного шума: нужна дедупликация/aggregation policy для повторяющихся ошибок.
- Риск утечки секретов в payload/stack: обязательный redaction на ingress error journal.

## Фактический результат (выполнено)
- Добавлена новая таблица `runtime_errors` c индексами active/feed и связями на `agent_runs`, `projects`, `users`.
- Реализованы доменные типы, repository contract и PostgreSQL-репозиторий для runtime error journal.
- Добавлен доменный service `runtimeerror` (нормализация level/payload, safe JSON details, best-effort запись).
- Добавлен ingest-hook в control-plane runtime deploy pipeline:
  - при error-level task logs запись дублируется в `runtime_errors` с контекстом (`run_id`, `source`, `stage`, `namespace/job`).
- Расширен gRPC control-plane API:
  - `ListRuntimeErrors`;
  - `MarkRuntimeErrorViewed`.
- Расширен staff HTTP API (OpenAPI + api-gateway):
  - `GET /api/v1/staff/runtime-errors` (active/viewed/all + фильтры);
  - `POST /api/v1/staff/runtime-errors/{runtime_error_id}/viewed`.
- Реализован staff UI alert center:
  - глобальный stack в правом нижнем углу (до 5 свежих ошибок);
  - polling-refresh без перезагрузки страницы;
  - `dismiss -> mark viewed`;
  - быстрый переход к run/deploy details по `run_id`, если доступен.

## Проверки
- `go test ./services/internal/control-plane/...` — passed.
- `go test ./services/external/api-gateway/...` — passed.
- `npm --prefix services/staff/web-console run build` — passed.
