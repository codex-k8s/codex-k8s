---
doc_id: ADR-0010
type: adr
title: "S7 MVP readiness: stream boundaries and parity-gate before run:dev"
status: accepted
owner_role: SA
created_at: 2026-03-02
updated_at: 2026-03-02
related_issues: [220, 222, 238]
related_prs: []
supersedes: []
superseded_by: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-02-issue-222-arch"
---

# ADR-0010: S7 MVP readiness — stream boundaries and parity-gate before run:dev

## TL;DR
- Контекст: PRD Stage S7 (`#220`) определил 18 execution-streams (`S7-E01..S7-E18`), но без архитектурной фиксации ownership риск scope drift остаётся высоким.
- Решение: применяем bounded-stream архитектуру по сервисным зонам + обязательный parity-gate до `run:dev`.
- Последствия: снижается риск хаотичного implementation scope, но повышается строгость stage-governance и контроль handover.

## Контекст
- Проблема: без архитектурной матрицы `stream -> boundary -> owner -> contract/data impact` design/dev этапы получат конфликтующие интерпретации scope.
- Ограничения:
  - markdown-only change set для `run:arch`;
  - нельзя менять базовую label taxonomy;
  - edge (`external|staff`) должен оставаться thin-layer.
- Связанные требования:
  - FR-026, FR-028, FR-033, FR-053, FR-054;
  - NFR-010, NFR-018.
- Что ломается без решения:
  - параллельные dev-issues без эквивалентности approved epics;
  - смешение доменной логики в edge/UI;
  - rework при переходе из design в dev.

## Decision Drivers
- Детерминизм stage-переходов.
- Явный ownership по сервисным границам.
- Минимизация rework между `run:design` и `run:dev`.
- Сохранение существующих архитектурных ограничений платформы.

## Рассмотренные варианты

### Вариант A: Centralized mega-implementation issue
- Плюсы:
  - минимум governance артефактов.
- Минусы:
  - высокий риск scope drift и конфликтов работ.
- Риски:
  - невозможность прозрачно валидировать parity coverage.
- Стоимость внедрения:
  - низкая в short-term, высокая в long-term (rework).
- Эксплуатация:
  - сложный аудит и ретроспектива.

### Вариант B: Полная децентрализация без parity-gate
- Плюсы:
  - высокая скорость параллельного старта.
- Минусы:
  - отсутствует формальный контроль соответствия PRD scope.
- Риски:
  - пропуск P0-потоков, неконсистентные handover-цепочки.
- Стоимость внедрения:
  - средняя.
- Эксплуатация:
  - непредсказуемая нагрузка на review/qa.

### Вариант C (выбран): Bounded-stream ownership + parity-gate
- Плюсы:
  - однозначные сервисные границы и ответственность;
  - проверяемый контроль перед `run:dev`.
- Минусы:
  - больше документационного governance.
- Риски:
  - необходимость строгого соблюдения traceability discipline.
- Стоимость внедрения:
  - средняя.
- Эксплуатация:
  - высокая воспроизводимость решений и упрощённый аудит.

## Решение
Мы выбираем: **вариант C — bounded-stream ownership + parity-gate**.

## Обоснование (Rationale)
- Вариант C лучше всего соответствует текущей stage-модели и требованиям FR-053/FR-054 к детерминизму действий.
- Позволяет сохранить thin-edge и доменную централизацию в `control-plane`/`worker`.
- Уменьшает риск незавершённого execution scope перед стартом кодовых работ.

## Последствия (Consequences)

### Позитивные
- Для каждого `S7-E*` есть однозначный owner и boundary.
- Переход в `run:dev` становится измеримым (coverage ratio = 1.0).
- Проще проверять `design -> plan -> dev` handover на полноту.

### Негативные / компромиссы
- Увеличивается объём governance-обновлений в delivery docs.
- Выше дисциплинарная нагрузка на stage continuity.

### Технический долг
- Что откладываем:
  - автоматизированный parity-check в runtime policy уровне.
- Когда вернуться:
  - в потоках `run:design`/`run:plan` Sprint S7.

## План внедрения (минимально)
- Изменения в коде:
  - отсутствуют на `run:arch`.
- Изменения в инфраструктуре:
  - отсутствуют.
- Миграции/совместимость:
  - не применимо в текущем этапе.
- Наблюдаемость:
  - на `run:design` уточнить обязательные audit-events для потоков reliability/policy.

## План отката/замены
- Условия отката:
  - если Owner отклоняет parity-gate как обязательный.
- Как откатываем:
  - перевод ADR в `superseded` и возврат к варианту B с явным owner-decision в docs/audit.

## Ссылки
- PRD:
  - `docs/delivery/epics/s7/prd-s7-day3-mvp-readiness-gap-closure.md`
- Design doc:
  - `docs/architecture/s7_mvp_readiness_gap_closure_architecture.md`
- Alternatives:
  - `docs/architecture/alternatives/ALT-0002-s7-mvp-readiness-stream-architecture.md`
