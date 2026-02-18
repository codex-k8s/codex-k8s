---
doc_id: EPC-CK8S-S3-D19
type: epic
title: "Epic S3 Day 19: Run access key and OAuth bypass flow"
status: planned
owner_role: EM
created_at: 2026-02-18
updated_at: 2026-02-18
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
