---
doc_id: ALT-0002
type: alternatives
title: "S7 MVP readiness streams — alternatives and trade-offs"
status: in-review
owner_role: SA
created_at: 2026-03-02
updated_at: 2026-03-02
related_issues: [220, 222, 238]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-02-issue-222-arch"
---

# Alternatives & Trade-offs: S7 MVP readiness streams

## TL;DR
- Рассмотрели: централизованный mega-stream / децентрализованный старт / bounded-stream + parity-gate.
- Рекомендуем: bounded-stream + parity-gate.
- Почему: лучший баланс предсказуемости, traceability и controllable handover в `run:design -> run:plan -> run:dev`.

## Контекст
- PRD зафиксировал 18 потоков `S7-E01..S7-E18` и dependency graph.
- Без архитектурного выбора растёт риск пересечения сервисных границ и несогласованного dev backlog.
- Ограничения:
  - markdown-only на `run:arch`;
  - без изменения label taxonomy;
  - thin-edge правило обязательно.

## Вариант A: Один большой implementation stream
- Описание:
  - консолидировать все S7-E* в один run:dev поток.
- Плюсы:
  - меньше административных задач.
- Минусы:
  - плохая локализация рисков и ownership.
- Риски:
  - высокий rework на review/qa.
- Стоимость/сложность:
  - низкий старт, высокий общий риск.

## Вариант B: Полная параллель без parity-gate
- Описание:
  - создавать implementation issues независимо от approved epic coverage.
- Плюсы:
  - максимально быстрый старт.
- Минусы:
  - не гарантируется покрытие всех approved потоков.
- Риски:
  - execution drift и конфликтующие приоритеты P0/P1.
- Стоимость/сложность:
  - средняя, но дорого в стабилизации.

## Вариант C: Bounded-stream + parity-gate (recommended)
- Описание:
  - каждый S7-E* получает явный boundary/owner;
  - перед `run:dev` действует gate `approved_execution_epics == created_run_dev_issues`.
- Плюсы:
  - предсказуемость и auditability;
  - проще управлять sequencing и quality-gates.
- Минусы:
  - выше дисциплина и объём docs синхронизации.
- Риски:
  - задержка старта dev при плохой traceability-гигиене.
- Стоимость/сложность:
  - средняя.

## Сравнение
| Критерий | A | B | C |
|---|---:|---:|---:|
| Срок старта | 5 | 5 | 4 |
| Риск scope drift | 1 | 1 | 5 |
| Стоимость rework | 1 | 2 | 4 |
| Управляемость review/qa | 2 | 2 | 5 |
| Эксплуатационная предсказуемость | 2 | 2 | 5 |

## Рекомендация
- Выбор: **Вариант C**.
- Обоснование:
  - соответствует stage-policy и архитектурным ограничениям;
  - создаёт формальный gate перед кодовым этапом.
- Что теряем:
  - часть скорости на старте implementation.
- Что выигрываем:
  - детерминизм, качество handover и снижение rework.

## Нужен апрув от Owner
- [x] Выбор варианта C.
- [x] Разрешение компромисса: больше governance-работы в обмен на меньшее количество execution-ошибок.
