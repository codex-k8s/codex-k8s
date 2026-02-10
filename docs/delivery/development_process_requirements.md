---
doc_id: PRC-CK8S-0001
type: process-requirements
title: "codex-k8s — Development and Documentation Process Requirements"
status: active
owner_role: EM
created_at: 2026-02-06
updated_at: 2026-02-10
related_issues: [1]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-06-process"
---

# Development and Documentation Process Requirements

## TL;DR
- Этот документ задаёт обязательный weekly-процесс: планирование спринта, ежедневное выполнение, ежедневный deploy на staging, закрытие спринта.
- Требования обязательны для разработки и для ведения документации.
- Любое отклонение от процесса фиксируется явно и согласуется с Owner.

## Нормативные ссылки (source of truth)
- `AGENTS.md`
- `docs/product/requirements_machine_driven.md`
- `docs/product/constraints.md`
- `docs/delivery/delivery_plan.md`
- `docs/delivery/sprint_s1_mvp_vertical_slice.md`
- `docs/delivery/sprint_s2_dogfooding.md`
- `docs/delivery/issue_map.md`
- `docs/delivery/requirements_traceability.md`
- `docs/design-guidelines/**`
- `docs/templates/**`

## Базовые принципы процесса
- Weekly sprint cadence: каждая неделя начинается формальным kickoff и завершается formal close.
- Trunk-based delivery: маленькие инкременты, ежедневные merge в `main`.
- CI/CD discipline: merge только после green pipeline и обязательного deploy в staging.
- Docs-as-code: изменения кода и документации синхронны в одном рабочем цикле.
- Traceability by default: каждое решение привязано к требованиям и артефактам.
- Security by default: секреты не хранятся в репозитории, префикс переменных платформы `CODEXK8S_`.

## Роли и ответственность

| Роль | Ответственность | Основные артефакты |
|---|---|---|
| Owner | Утверждает scope, приоритеты, критические решения, go/no-go | Апрувы в frontmatter, решения по рискам |
| PM | Поддерживает продуктовые требования и ограничения | `docs/product/requirements_machine_driven.md`, `docs/product/brief.md`, `docs/product/constraints.md` |
| EM | Ведёт спринт-план, эпики, daily delivery gate | `docs/delivery/sprint_s*.md`, `docs/delivery/epic_s*.md`, `docs/delivery/epics/epic-s*-day*.md` |
| SA | Архитектурная и data-model консистентность | `docs/architecture/*.md`, миграционная стратегия |
| Dev | Реализация задач и технические проверки | код, тесты, миграции, изменения API/контрактов |
| QA | Ручной smoke/regression на staging, acceptance evidence | test evidence, regression checklist |
| SRE | Bootstrap/deploy/runbook/операционная устойчивость | bootstrap scripts, deploy manifests, runbook |
| KM | Трассируемость документации и docset-актуальность | `docs/delivery/issue_map.md`, docset документы |

## Нейминг артефактов (обязателен)

Цель: чтобы ссылки в документации были стабильными и чтобы каждый спринт имел однозначные файлы.

Правила:
- Sprint plan файл:
  - `docs/delivery/sprint_s<номер>_<краткое-имя>.md`
  - пример: `docs/delivery/sprint_s2_dogfooding.md`
- Epic catalog файл:
  - `docs/delivery/epic_s<номер>.md`
  - пример: `docs/delivery/epic_s2.md`
- Daily epic docs:
  - `docs/delivery/epics/epic-s<номер>-day<день>-<краткое-имя>.md`
  - пример: `docs/delivery/epics/epic-s2-day0-control-plane-extraction.md`

## Еженедельный цикл спринта

### 1. Sprint Start (день начала недели)
- Проверить актуальность требований и ограничений.
- Сформировать/актуализировать план спринта и набор эпиков по дням.
- Для каждого эпика задать priority (`P0/P1/P2`) и ожидаемые артефакты дня.
- Провести DoR-check.

Обязательные артефакты:
- Sprint plan: `docs/delivery/sprint_s<номер>_<краткое-имя>.md` (актуальный sprint-file недели).
- Epic catalog: `docs/delivery/epic_s<номер>.md` и daily epic docs в `docs/delivery/epics/`.
- `docs/delivery/issue_map.md` и `docs/delivery/requirements_traceability.md`.

### 2. Daily Execution (каждый рабочий день спринта)
- Реализовать задачи текущего дневного эпика.
- Выполнить merge в `main`.
- Подтвердить автоматический deploy на staging.
- Выполнить ручной smoke-check и зафиксировать результат.
- Обновить документацию при изменении API/data model/webhook/RBAC/процессов.

Daily gate (must pass):
- PR/merge только при green CI.
- Staging deployment успешен.
- Smoke-check успешен или заведен блокер с решением.
- Документация синхронизирована.

### 3. Mid-Sprint Control (середина недели)
- Перепроверить риски, блокеры, зависимости.
- Разрешается перераспределение `P1/P2`; `P0` меняется только через явное решение Owner.
- Актуализировать эпики и sprint-plan.

### 4. Sprint Close (последний день недели)
- Прогнать regression ключевых сценариев.
- Зафиксировать go/no-go на следующий спринт.
- Закрыть/перенести незавершённые задачи с обоснованием.
- Обновить roadmap/delivery-план.

## Матрица артефактов: кто и когда производит

| Артефакт | Когда | Кто производит (R) | Кто утверждает (A) |
|---|---|---|---|
| Requirements baseline | При изменении scope/решений | PM | Owner |
| Sprint plan | В начале недели и при major reprioritization | EM | Owner |
| Epic docs по дням | До старта дня и при закрытии дня | EM + Dev/SA/SRE | Owner |
| Data model updates | При любом изменении схемы/индексов | SA + Dev | Owner |
| API contract updates | При изменении внешних/внутренних API | SA + Dev | Owner |
| Issue/Doc traceability | Ежедневно после merge | KM + EM | Owner |
| Smoke/Regression evidence | Ежедневно / в конце спринта | QA + Dev | EM |
| Runbook/deploy updates | При изменении bootstrap/deploy/ops поведения | SRE + Dev | Owner |

## Обязательные quality gates
- Planning gate: DoR пройден, приоритеты и артефакты на день назначены.
- Merge gate: green CI + код ревью + синхронная документация.
- Deploy gate: staging deployment success + ручной smoke.
- Close gate: regression pass + согласованный backlog следующего спринта.

## Правило разрешения противоречий
- Если задача противоречит `docs/design-guidelines/**` или source-of-truth требованиям, работа останавливается.
- Предлагаются варианты решения с trade-offs.
- Финальное решение фиксируется в документации и утверждается Owner.

## Апрув
- request_id: owner-2026-02-06-process
- Решение: approved
- Комментарий: Процесс weekly sprint и doc governance утверждён.
