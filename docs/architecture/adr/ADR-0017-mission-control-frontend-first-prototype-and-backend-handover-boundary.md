---
doc_id: ADR-0017
type: adr
title: "Mission Control frontend-first prototype: web-console-owned fake-data slice with explicit backend handover boundary"
status: proposed
owner_role: SA
created_at: 2026-03-27
updated_at: 2026-03-27
related_issues: [480, 561, 562, 563, 565, 567, 571, 573]
related_prs: []
supersedes: []
superseded_by: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-27-issue-571-arch"
---

# ADR-0017: Mission Control frontend-first prototype — web-console-owned fake-data slice with explicit backend handover boundary

## TL;DR
- Контекст: Sprint S18 требует сначала доказать новый Mission Control UX на fake data и только потом запускать backend rebuild `#563`.
- Решение: выбираем `web-console` как единственного owner fake-data scenario/model/view-state для Sprint S18, а `api-gateway`, `control-plane`, `worker` и `PostgreSQL` оставляем thin/deferred boundaries без новой Mission Control truth responsibility.
- Последствия: UX baseline можно довести до `run:dev` без backend lock-in, но design-stage должен явно зафиксировать replacement seam к будущему persisted provider/data foundation.

## Контекст
- Проблема:
  - без закреплённого owner для Sprint S18 fake-data prototype команда снова смешает UX, backend rebuild `#563`, старые S16 assumptions и live provider semantics;
  - попытка опереться на текущие API или прежнюю graph/data model сделает frontend-first sprint зависимым от ещё неутверждённого backend path;
  - workflow editor рискует расползтись в prompt editor или live mutation console, если Day4 не зафиксирует границу ownership.
- Ограничения:
  - `run:arch` остаётся markdown-only;
  - Wave 1 baseline нельзя переоткрывать: fullscreen canvas, taxonomy `Issue/PR/Run`, compact nodes, explicit relations, drawer, toolbar и fake-data workflow preview обязательны;
  - repo-seed prompts остаются source of truth;
  - DB prompt editor, live provider mutation path и backend rebuild `#563` не входят в Sprint S18 scope.
- Связанные требования:
  - `docs/delivery/epics/s18/prd-s18-day3-mission-control-frontend-first-canvas.md`
  - `FR-009`, `FR-026`, `FR-027`
  - `NFR-014`, `NFR-018`
- Что ломается без решения:
  - `run:design` и `run:plan` начнут заново спорить, чей слой владеет fake-data state и workflow preview semantics;
  - Sprint S18 потеряет frontend-first характер и вернётся к backend-first sequencing;
  - Owner review не сможет отделить UX baseline от неготового provider/data foundation.

## Decision Drivers
- Сохранить locked baseline Day1-Day3 без reopening S16 assumptions.
- Удержать thin-edge правила для `api-gateway` и отсутствие доменной логики в UI beyond fake-data view-state.
- Сделать backend rebuild `#563` явным downstream flow, а не скрытой зависимостью.
- Сохранить repo-seed prompts как source of truth и не допустить prompt editor semantics.
- Минимизировать runtime/code lock-in до owner approval Sprint S18.

## Рассмотренные варианты

### Вариант A: Backend-first rebuild сразу в рамках Sprint S18
- Плюсы:
  - быстрее начать будущий persisted provider/data foundation.
- Минусы:
  - UX снова начинает зависеть от спорной backend model;
  - смешиваются Sprint S18 и follow-up `#563`.
- Риски:
  - возврат отвергнутых S16 assumptions;
  - потеря frontend-first sequencing.
- Стоимость внедрения:
  - высокая уже на Day4.
- Эксплуатация:
  - появляется неутверждённый runtime path до owner approval UX.

### Вариант B: Prototype поверх текущих API и частичного live sync
- Плюсы:
  - кажется быстрым компромиссом между UX и existing platform state.
- Минусы:
  - создаёт ложное ощущение canonical data path;
  - тяжело удержать guardrail `platform-safe actions only`.
- Риски:
  - UI начнёт нормализовать live provider truth и live mutations;
  - fake-data boundary станет неявной.
- Стоимость внедрения:
  - средняя, но с высоким architectural debt.
- Эксплуатация:
  - сложно объяснить пользователю, что из увиденного является реальным, а что прототипным.

### Вариант C (выбран): `web-console` owns isolated fake-data slice, backend rebuild lives in separate follow-up
- Плюсы:
  - сохраняет чистый frontend-first sequencing;
  - не нарушает thin-edge/domain boundaries;
  - делает handover к `#563` явным и проверяемым.
- Минусы:
  - появляется необходимость отдельно документировать replacement seam от fake data к persisted truth;
  - часть design artefacts описывает future boundary, а не текущий runtime code.
- Риски:
  - если seam будет описан слишком абстрактно, backend rebuild потом потеряет continuity.
- Стоимость внедрения:
  - средняя и управляемая в markdown-only scope.
- Эксплуатация:
  - предсказуемая: prototype остаётся честно изолированным и не притворяется live system.

## Решение
Мы выбираем: **вариант C — `web-console` owns the Sprint S18 fake-data prototype, while backend truth/model work remains a separate downstream flow in issue `#563`**.

## Обоснование (Rationale)
- Вариант C лучше всего соответствует product sequencing из `#561`, `#562`, `#565` и `#567`.
- Он сохраняет принятые архитектурные границы monorepo:
  - `web-console` отвечает только за UI/view-state;
  - `api-gateway` остаётся thin-edge;
  - `control-plane` и `worker` не получают premature Mission Control truth ownership в этом спринте;
  - prompt policy остаётся за repo seeds и platform docs.
- Это решение позволяет Owner оценить UX baseline отдельно от backend complexity и не создавать ложное впечатление, что fake-data prototype уже является canonical Mission Control.

## Последствия (Consequences)

### Позитивные
- Sprint S18 можно довести до `run:dev` без новых API, миграций и background sync path.
- Backend rebuild `#563` получает явный downstream seam вместо скрытого наследования временных UI решений.
- Workflow preview остаётся policy-shaping UX, а не prompt editor или live action console.

### Негативные / компромиссы
- Design-stage должен описать replacement seam между fake-data model и будущим persisted truth.
- Часть потенциальных transport/data решений откладывается и не может быть "проверена кодом" на этапе Sprint S18.

### Технический долг
- Что откладываем:
  - persisted provider mirror и reconcile;
  - structured workflow definitions/instances;
  - true stale/freshness semantics;
  - live provider mutation path и DB prompt editor.
- Когда вернуться:
  - в отдельном flow `#563` после owner approval UX baseline Sprint S18.

## План внедрения (минимально)
- Изменения в коде:
  - отсутствуют на `run:arch`.
- Изменения в инфраструктуре:
  - отсутствуют.
- Миграции/совместимость:
  - Sprint S18 prototype не требует миграций; migration scope остаётся за `#563`.
- Наблюдаемость:
  - `run:design` должен зафиксировать, какие UI evidence и safe-action surfaces нужны для owner walkthrough без live backend dependency.

## План отката/замены
- Условия отката:
  - если на `run:design` выяснится, что Sprint S18 не может сохранить UX baseline без backend prerequisite.
- Как откатываем:
  - ADR переводится в `superseded`, а owner отдельно решает, нужно ли эскалировать инициативу в backend-first flow до `run:dev`.

## Ссылки
- PRD:
  - `docs/delivery/epics/s18/prd-s18-day3-mission-control-frontend-first-canvas.md`
- Architecture:
  - `docs/architecture/initiatives/s18_mission_control_frontend_first_canvas/architecture.md`
- Alternatives:
  - `docs/architecture/alternatives/ALT-0009-mission-control-frontend-first-prototype-boundaries.md`
- Related issues:
  - `#561`, `#563`, `#571`, `#573`, `#480`
