---
doc_id: EPC-CK8S-0001
type: epic
title: "Epic Catalog: Sprint S1 (MVP vertical slice, Day 0..7)"
status: active
owner_role: EM
created_at: 2026-02-06
updated_at: 2026-02-11
related_issues: [1]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-06-sprint-s1"
---

# Epic Catalog: Sprint S1 (MVP vertical slice, Day 0..7)

## TL;DR
- Каталог фиксирует 8 эпиков спринта: `Day 0..7` выполнены и приняты.
- Цель каталога: один источник правды по ежедневному объёму и критериям приёмки.
- Рабочий режим: каждый день merge в `main` и deploy на `staging`.
- Процессное управление: `docs/delivery/development_process_requirements.md`.

## Контекст
- Почему нужен: команда ведёт ежедневную поставку и должна синхронно идти по документированному плану.
- Связь с требованиями: `docs/product/requirements_machine_driven.md`.

## Scope
### In scope
- Описание эпиков `Day 0..7`.
- Ссылки на критерии приемки каждого дня.
- Единый daily delivery gate.

### Out of scope
- Подробные технические задачи внутри PR (это уровень story/task в рамках каждого эпика).

## Декомпозиция (Stories/Tasks)
- Day 0: `docs/delivery/epics/epic-s1-day0-bootstrap-baseline.md`
- Day 1: `docs/delivery/epics/epic-s1-day1-webhook-idempotency.md`
- Day 2: `docs/delivery/epics/epic-s1-day2-worker-slots-k8s.md`
- Day 3: `docs/delivery/epics/epic-s1-day3-auth-rbac-ui.md`
- Day 4: `docs/delivery/epics/epic-s1-day4-repository-provider.md`
- Day 5: `docs/delivery/epics/epic-s1-day5-learning-mode.md`
- Day 6: `docs/delivery/epics/epic-s1-day6-hardening-observability.md`
- Day 7: `docs/delivery/epics/epic-s1-day7-stabilization-gate.md`

## Приоритеты Sprint S1
- `P0`: Day 0, Day 1, Day 2, Day 3, Day 4, Day 7.
- `P1`: Day 5, Day 6.
- `P2`: не запланирован в Sprint S1.

## Критерии приемки эпика
- Для каждого эпика заполнен блок `Data model impact` в структуре шаблона `docs/templates/data_model.md`.
- По завершению каждого дня изменения влиты в `main` и раскатаны на `staging`.
- Для каждого дня выполнен ручной smoke-check.

## Риски/зависимости
- Зависимости: доступный staging, валидные `CODEXK8S_*` секреты, стабильный CI.
- Риски: ежедневные слияния могут накапливать регрессии без строгого gate.

## План релиза (верхний уровень)
- Day 0 фиксируется как baseline.
- Day 1..7 выпускаются последовательно с ежедневным staging deploy.
- По Day 7 формируется go/no-go для следующего спринта.

## Апрув
- request_id: owner-2026-02-06-sprint-s1
- Решение: approved
- Комментарий: Каталог эпиков Sprint S1 утверждён.
