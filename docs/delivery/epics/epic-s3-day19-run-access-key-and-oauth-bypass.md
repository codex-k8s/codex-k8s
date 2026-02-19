---
doc_id: EPC-CK8S-S3-D19
type: epic
title: "Epic S3 Day 19: Run access key and OAuth bypass flow"
status: completed
owner_role: EM
created_at: 2026-02-18
updated_at: 2026-02-19
related_issues: [19]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
---

# Epic S3 Day 19: Run access key and OAuth bypass flow

## TL;DR
- Цель: добавить controlled bypass для OAuth в рамках run lifecycle, чтобы агент и оператор могли продолжать работу в критичных сценариях.
- Результат: на запуске генерируется временный access key, который может использоваться для authorized bypass маршрута; агент получает ключ в env и инструкцию в prompt.

## Priority
- `P0`.

## Scope
### In scope
- Run-scoped access key модель:
  - генерация, TTL, статус, revocation;
  - привязка к run/project/environment.
- Backend validation middleware для bypass-режима:
  - строгое ограничение по аудитории/namespace/ttl;
  - audit trail всех bypass-действий.
- Интеграция в run lifecycle:
  - key issue при старте,
  - проброс ключа агенту через env,
  - отображение в prompt context как разрешённой возможности.
- UI/операционный контур:
  - отображение статуса bypass key в run details,
  - revoke/regenerate (минимальный staff control).

### Out of scope
- Полная замена OAuth или постоянные machine tokens для всех пользовательских сценариев.

## Декомпозиция
- Story-1: data model + crypto generation + secure storage.
- Story-2: auth middleware + bypass endpoint contract.
- Story-3: run orchestration + prompt/env integration.
- Story-4: staff UI controls + audit events.

## Критерии приемки
- Для каждого нового run может быть выпущен run-scoped access key с TTL.
- Bypass доступ возможен только с валидным ключом и только в рамках разрешённого контекста.
- Агент получает ключ и видит инструкцию о допустимом использовании в prompt.
- Все операции bypass фиксируются в audit и доступны в staff observability.

## Риски/зависимости
- Высокий security-risk: нужен строгий scope, TTL, rotation/revocation.
- Риск неправильного UX: требуется явное разграничение обычного OAuth и временного bypass режима.

## Фактический результат (выполнено)
- Добавлена run-scoped модель доступа `run_access_keys`:
  - миграция БД;
  - доменные типы и repository contract;
  - PostgreSQL-репозиторий с upsert/revoke/touch-last-used.
- Реализован domain service `runaccess`:
  - issue/regenerate/revoke/get-status;
  - authorize-by-key с проверкой scope (`run_id`, `project_id`, `namespace`, `target_env`, `runtime_mode`) и TTL;
  - audit событий в flow events (`issued`, `regenerated`, `revoked`, `authorized`, `denied`).
- Интеграция в run lifecycle:
  - worker выпускает key при старте run;
  - key прокидывается в job env как `CODEXK8S_RUN_ACCESS_KEY`;
  - agent-runner передаёт key в prompt envelope как ограниченный bypass-механизм.
- Расширен gRPC transport control-plane:
  - staff operations: `GetRunAccessKeyStatus`, `RegenerateRunAccessKey`, `RevokeRunAccessKey`;
  - bypass operations: `GetRunByAccessKey`, `ListRunEventsByAccessKey`, `GetRunLogsByAccessKey`;
  - proto-контракт и generated-код синхронизированы.
- Расширен HTTP/API gateway:
  - staff endpoints для rotate/revoke/status;
  - public bypass endpoints (`/api/v1/runs/{run_id}/bypass*`) с обязательным `access_key`;
  - OpenAPI contract и generated SDK синхронизированы.
- Реализован staff UI (Run Details):
  - отображение статуса и метаданных bypass key;
  - действия regenerate/revoke;
  - показ plaintext key только сразу после выпуска/ротации.

## Проверки
- `go test ./services/internal/control-plane/... ./services/jobs/worker/... ./services/jobs/agent-runner/... ./services/external/api-gateway/... ./libs/go/k8s/joblauncher/...` — passed.
- `npm --prefix services/staff/web-console run build` — passed.
