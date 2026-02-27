---
doc_id: EPC-CK8S-S7-D5-GO-REF
type: epic
title: "Sprint S7 Day5: Plan для рефакторинга Go-сервисов (Issue #223)"
status: in-review
owner_role: EM
created_at: 2026-02-27
updated_at: 2026-02-27
related_issues: [223, 225, 226, 227, 228, 229, 230]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-27-issue-223-plan"
---

# Sprint S7 Day5: Plan для рефакторинга Go-сервисов (Issue #223)

## TL;DR
- Issue `#223` декомпозирован на 6 параллельных implementation-задач по сервисам и библиотекам Go.
- Декомпозиция построена по принципу независимых bounded scopes, чтобы команды могли работать параллельно без пересечений.
- Для каждого потока зафиксированы quality-gates, критерии завершения и риски.

## Контекст
- Запрос stage `run:plan`: подготовить backlog рефакторинга Go-кода без правок production-кода в рамках текущего запуска.
- Исходные сигналы: oversized service/handler файлы, дублирующиеся helper-паттерны, неоднородность работы с PostgreSQL и размещения типов.
- Source of truth:
  - `AGENTS.md`
  - `docs/design-guidelines/AGENTS.md`
  - `docs/design-guidelines/common/check_list.md`
  - `docs/delivery/development_process_requirements.md`

## Параллельные implementation-потоки

| Epic ID | GitHub Issue | Scope | Owner role |
|---|---|---|---|
| S7-E19 | `#225` | control-plane: decomposition больших domain/transport файлов | `dev` |
| S7-E20 | `#226` | api-gateway: cleanup transport handlers и boundary hardening | `dev` |
| S7-E21 | `#227` | worker: decomposition service и cleanup duplication | `dev` |
| S7-E22 | `#228` | agent-runner: normalization helpers и dedup prompt context | `dev` |
| S7-E23 | `#229` | shared libs: pgx alignment + modularization `servicescfg` | `dev` |
| S7-E24 | `#230` | cross-service hygiene closure и residual debt report | `dev` |

## Последовательность исполнения
1. Независимо стартуют `#225`, `#226`, `#227`, `#228`, `#229`.
2. После их завершения выполняется `#230` как consolidating поток.
3. Для каждого потока обязателен self-check по `common/go` чек-листам перед PR.

## Quality-gates

| Gate | Критерий | Применение |
|---|---|---|
| QG-223-01 Scope isolation | Каждый issue изменяет только свой сервис/библиотеку без cross-cutting drift | `#225..#229` |
| QG-223-02 No behavior change | Рефакторинг не меняет функциональное поведение и контракты | `#225..#230` |
| QG-223-03 Structural conformance | Типы/константы/helpers размещены по design-guidelines | `#225..#230` |
| QG-223-04 Verification | Выполнены релевантные `go test` и self-check по чек-листам | `#225..#230` |
| QG-223-05 Consolidation | По завершении потока есть единый residual risk/debt report | `#230` |

## Definition of Done (для execution-потоков)
- [ ] Изменения ограничены согласованным scope конкретного issue.
- [ ] Контракты транспорта и доменные границы не нарушены.
- [ ] Все релевантные тесты проходят.
- [ ] Обновлены `docs/delivery/issue_map.md` и `docs/delivery/requirements_traceability.md`.
- [ ] В PR перечислены риски и owner decisions.

## Блокеры, риски, owner decisions

| Тип | ID | Описание | Статус |
|---|---|---|---|
| blocker | BLK-223-01 | Без owner-аппрува нельзя запускать `run:dev` на созданные implementation issues | open |
| risk | RSK-223-01 | Возможны регрессии при декомпозиции крупных service-файлов | open |
| risk | RSK-223-02 | Неоднородность подходов к БД может потребовать staged migration path | open |
| decision | OD-223-01 | Выполнять рефакторинг строго без изменения продуктового поведения | accepted |
| decision | OD-223-02 | `#230` исполняется только после закрытия `#225..#229` | accepted |

## Next stage handover
- Следующий этап: `run:dev`.
- Handover-пакет: implementation issues `#225..#230`, quality-gates `QG-223-01..QG-223-05`, DoD этого документа.
