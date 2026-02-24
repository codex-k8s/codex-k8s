---
doc_id: APC-CK8S-0119
type: design-api-contract
title: "Issue #119 — Design API Contract and Label Transition Rules"
status: draft
owner_role: SA
created_at: 2026-02-24
updated_at: 2026-02-24
related_issues: [119, 118]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-24-issue-119-design"
---

# API/Label Contract: Issue #119

## TL;DR
- Issue #119 не вводит новые HTTP/gRPC контракты.
- Контракт design-этапа состоит из корректного использования существующих label transitions и audit-событий.
- Для B2 ambiguity обязательный результат: `need:input`, без запуска revise-run.

## Input assumptions
- Stage resolver и next-step UX реализуются по `ADR-0006`.
- Label classes `run:*`, `state:*`, `need:*` обрабатываются по `labels_and_trigger_policy`.
- Evidence публикуется в рамках E2E handover для Issue #118.

## Контрактные границы
- Public boundary: только `POST /api/v1/webhooks/github`.
- Staff/private boundary: существующие endpoints run/approvals/logs без изменений DTO.
- Internal boundary: существующие gRPC/MCP контракты (без новых RPC).

## Операционный контракт по сценариям

| Сценарий | Входной сигнал | Обязательный выход | Запрещенный выход |
|---|---|---|---|
| A1/A2/A3 | trigger `run:*` по stage-модели | соответствующий stage transition + audit trail | несколько trigger labels одновременно |
| B1 | `changes_requested` + однозначный stage | `run:<stage>:revise` на Issue | `need:input` при отсутствии ambiguity |
| B2 | `changes_requested` + ambiguity stage | `need:input` + service remediation message | запуск revise-run |
| B3 | `changes_requested` между revise-итерациями | sticky profile resolve для model/reasoning | silent reset профиля на defaults |

## Контрактные инварианты
- `run:*` запускается только при детерминированном stage resolve.
- `state:*`/`need:*` не запускают run сами по себе.
- Label transitions выполняются через MCP policy path для единого аудита.
- Пользовательская коммуникация и PR-тексты в этом запуске ведутся на русском языке.

## Verification evidence (A+B)

| Scenario | Проверка | Артефакт подтверждения |
|---|---|---|
| A1/A2/A3 | последовательность stage transitions без конфликтов trigger labels | `flow_events` + service comments |
| B1 | revise стартует только при однозначном stage | transition `run:<stage>:revise` |
| B2 | ambiguity приводит к `need:input` без revise-run | label snapshot + remediation comment |
| B3 | model/reasoning не сбрасываются на defaults | `agent_sessions` profile snapshot |

## Обязательные evidence-поля по контракту
- `issue_number`, `pr_number`, `run_id`, `correlation_id`, `stage`.
- список фактических transitions `run:*`/`state:*`/`need:*`.
- ссылка на ключевой service-comment (или комментарий с remediation для B2).
- статус `pass/fail` на сценарий с коротким обоснованием.

## Связь с трассируемостью
- Для проверки покрытия FR/NFR и AC использовать:
  - `docs/delivery/design/issue-119/traceability_matrix.md`;
  - `docs/delivery/requirements_traceability.md`.

## Изменения OpenAPI/proto
- OpenAPI: без изменений.
- gRPC/proto: без изменений.
- AsyncAPI: без изменений.

## Runtime impact
- Нагрузочный профиль не меняется: повторное использование существующих маршрутов и обработчиков.
- Риск ограничен корректностью label-orchestration и audit completeness.

## Апрув
- request_id: owner-2026-02-24-issue-119-design
- Решение: pending
- Комментарий:
