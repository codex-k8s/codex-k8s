---
doc_id: PRC-CK8S-0001
type: process-requirements
title: "codex-k8s — Development and Documentation Process Requirements"
status: active
owner_role: EM
created_at: 2026-02-06
updated_at: 2026-03-02
related_issues: [1, 112, 210, 212, 241, 243]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-19-full-docset"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-19
---

# Development and Documentation Process Requirements

## TL;DR
- Этот документ задаёт обязательный weekly-процесс: планирование спринта, ежедневное выполнение, ежедневный deploy на production, закрытие спринта.
- Требования обязательны для разработки и для ведения документации.
- Любое отклонение от процесса фиксируется явно и согласуется с Owner.
- Для Issue/PR действует единый стандарт заголовков и содержимого, зависящий от stage и роли агента.
- Проектная документация ведётся по фиксированной карте размещения (`product/architecture/delivery/ops/templates`) без смешения типов документов.

## Нормативные ссылки (source of truth)
- `AGENTS.md`
- `docs/product/requirements_machine_driven.md`
- `docs/product/constraints.md`
- `docs/product/agents_operating_model.md`
- `docs/product/labels_and_trigger_policy.md`
- `docs/product/stage_process_model.md`
- `docs/delivery/delivery_plan.md`
- `docs/delivery/sprints/s1/sprint_s1_mvp_vertical_slice.md`
- `docs/delivery/sprints/s2/sprint_s2_dogfooding.md`
- `docs/delivery/sprints/README.md`
- `docs/delivery/epics/README.md`
- `docs/delivery/e2e_mvp_master_plan.md`
- `docs/delivery/issue_map.md`
- `docs/delivery/requirements_traceability.md`
- `docs/design-guidelines/**`
- `docs/templates/**`

## Базовые принципы процесса
- Weekly sprint cadence: каждая неделя начинается формальным kickoff и завершается formal close.
- Trunk-based delivery: маленькие инкременты, ежедневные merge в `main`.
- CI/CD discipline: merge только после green pipeline и обязательного deploy в production.
- Docs-as-code: изменения кода и документации синхронны в одном рабочем цикле.
- Traceability by default: каждое решение привязано к требованиям и артефактам.
- Security by default: секреты не хранятся в репозитории, префикс переменных платформы `CODEXK8S_`.

## Роли и ответственность

| Роль | Ответственность | Основные артефакты |
|---|---|---|
| Owner | Утверждает scope, приоритеты, критические решения, go/no-go | Апрувы в frontmatter, решения по рискам |
| PM | Поддерживает продуктовые требования и ограничения | `docs/product/requirements_machine_driven.md`, `docs/product/brief.md`, `docs/product/constraints.md` |
| EM | Ведёт спринт-план, эпики, daily delivery gate | `docs/delivery/sprints/s*/sprint_s*.md`, `docs/delivery/epics/s*/epic_s*.md`, `docs/delivery/epics/s*/epic-s*-day*.md` |
| SA | Архитектурная и data-model консистентность | `docs/architecture/*.md`, миграционная стратегия |
| Dev | Реализация задач и технические проверки | код, тесты, миграции, изменения API/контрактов |
| Reviewer | Предварительное ревью PR до Owner | inline findings в PR + summary для Owner |
| QA | Ручной smoke/regression на production, acceptance evidence | test evidence, regression checklist |
| SRE | Bootstrap/deploy/runbook/операционная устойчивость | bootstrap scripts, deploy manifests, runbook |
| KM | Трассируемость документации и актуальность карты связей | `docs/delivery/issue_map.md`, `docs/delivery/requirements_traceability.md` |

## Нейминг артефактов (обязателен)

Цель: чтобы ссылки в документации были стабильными и чтобы каждый спринт имел однозначные файлы.

Правила:
- Sprint plan файл:
  - `docs/delivery/sprints/s<номер>/sprint_s<номер>_<краткое-имя>.md`
  - пример: `docs/delivery/sprints/s2/sprint_s2_dogfooding.md`
- Epic catalog файл:
  - `docs/delivery/epics/s<номер>/epic_s<номер>.md`
  - пример: `docs/delivery/epics/s2/epic_s2.md`
- Daily epic docs:
  - `docs/delivery/epics/s<номер>/epic-s<номер>-day<день>-<краткое-имя>.md`
  - пример: `docs/delivery/epics/s2/epic-s2-day0-control-plane-extraction.md`

## Стандарт заголовков и содержимого Issue/PR (обязателен)

Цель: заголовок и body должны сразу показывать stage, роль и тип результата.

### Заголовки

| Тип | Формат | Пример |
|---|---|---|
| Stage issue (doc stages) | `S<спринт> Day<день>: <Stage> для <краткая цель>` | `S7 Day1: Intake для закрытия MVP readiness gaps` |
| Stage issue (`run:dev`) | `S<спринт> Day<день>: Dev — <краткая реализация>` | `S6 Day7: Dev — lifecycle управления агентами и prompt templates` |
| PR по stage-документации | `Issue #<номер>: <stage>-пакет <краткая цель>` | `Issue #212: intake-пакет Sprint S7 для закрытия MVP readiness gaps` |
| PR по `run:dev` | `Issue #<номер>: <краткая реализация> (#<номер>)` | `Issue #199: реализация lifecycle управления агентами и шаблонами промптов (#199)` |

Обязательные правила для заголовков:
- заголовок всегда одной строкой и без точки в конце;
- сначала фиксируется предмет/результат, потом уточнение;
- расплывчатые формулировки (`misc`, `update`, `fixes`, `wip`, `changes`) без предмета запрещены;
- если артефакт связан с существующей задачей, заголовок обязан сохранять прямую связь с `Issue #<номер>`.

### Follow-up Issue body

Для follow-up Issue использовать exact `##`-секции:
- `## Контекст`
- `## Проблема`
- `## Почему это важно сейчас`
- `## Что требуется сделать`
- `## Acceptance Criteria`
- `## Риски и ограничения`
- `## Связанные артефакты`

Дополнительно для QA-gap/failure issue:
- `## Сценарии воспроизведения / тестовые шаги`

Правила:
- если раздел не применим, писать `Не требуется`;
- в `Связанные артефакты` указывать issue/PR/doc/run/environment ссылки;
- для follow-up issue роль должна быть видна уже из заголовка (`Dev follow-up`, `QA gap`, `SRE remediation`, `Review follow-up` или эквивалентная role-prefix форма).

Пример follow-up Issue title:
- `Dev follow-up: добавить шаблонный контракт оформления PR`

### PR body: базовый контракт

| Артефакт | Обязательные блоки |
|---|---|
| Issue (любой stage) | `Контекст`, `Проблема`, `Scope (in/out)`, `Acceptance Criteria`, `Риски/допущения`, `Next stage handover` |
| PR для doc-stage (`run:intake..run:plan`, `run:qa`, `run:release`, `run:postdeploy`, `run:ops`, `run:doc-audit`) | `Контекст stage/role`, `Что обновлено`, `Traceability`, `Проверки`, `Риски`, `Next action` |
| PR для `run:dev` | `Контекст`, `Что изменено (code+docs)`, `Проверки (tests/lint/build)`, `Риски/миграции`, `Traceability` |

Общие правила для любого PR body:
- body должен быть валидным Markdown;
- использовать exact `##`-секции в порядке, определённом role/stage prompt contract;
- в каждом разделе писать факты: команды, тесты, запросы, логи, файлы, окружения, а не общие формулировки;
- если проверка, просмотр логов, работа с БД или чек-лист не выполнялись, это должно быть указано явно;
- в разделе закрытия последней строкой должен стоять exact directive:
  `Closes https://github.com/<owner>/<repo>/issues/<номер>`;
- short form `Closes #<номер>` допустим только если платформа работает в том же репозитории, но каноническим форматом для cloud agents считается полная ссылка.

### PR body: role-specific детализация

Для `run:dev`:
- `## Контекст`
- `## Основание для изменений`
- `## Основной смысл правок`
- `## Что сделано`
- `## Логи и runtime-диагностика`
- `## БД и миграции`
- `## Проверки`
- `## Проверенные чек-листы`
- `## Риски и ограничения`
- `## Рекомендации / следующий шаг`
- `## Закрытие`

Пример PR title для `run:dev`:
- `Issue #253: вынести prompt-контракты в шаблоны (#253)`

Для `run:dev:revise` и других `run:*:revise`:
- существующие PR title и body из исходной implementation-итерации сохраняются; revise-цикл не должен переименовывать или перетирать уже опубликованный контекст, трассируемость и историю проверок;
- перед обновлением PR title/body нужно сначала получить текущее название и описание существующего PR и только после этого встраивать revise-заметки поверх них;
- заголовок существующего PR сохраняется без изменения; допустима только минимальная правка, если заголовок потерял связь с `Issue #<номер>` или стал фактически некорректным;
- секции ниже описывают append-блок revise-итерации, который добавляется в конец текущего body перед финальным `## Закрытие`, а не заменяет весь body целиком;
- `## Контекст ревизии`
- `## Основание для изменений`
- `## Какие замечания обработаны`
- `## Какие замечания отклонены и почему`
- `## Что изменено`
- `## Повторные проверки`
- `## Проверенные чек-листы`
- `## Риски и ограничения`
- `## Рекомендации / следующий шаг`
- `## Закрытие`
- если в текущем body уже есть `## Закрытие`, финальный блок нужно перенести/обновить так, чтобы directive `Closes https://github.com/<owner>/<repo>/issues/<номер>` снова был последней строкой PR body;

Пример поведения для revise-итерации:
- сохранить существующий заголовок PR и дополнить только body новым revise-блоком

Для `run:qa`:
- `## Контекст`
- `## Основание для изменений`
- `## Что тестировалось`
- `## Тестовые сценарии и запросы`
- `## Где проверялось`
- `## Логи и evidence`
- `## Проверки`
- `## Проверенные чек-листы`
- `## Риски и ограничения`
- `## Рекомендации / следующий шаг`
- `## Закрытие`

Для `run:ops`, `run:release`, `run:postdeploy`, `run:ai-repair`:
- `## Контекст`
- `## Основание для изменений`
- `## Что диагностировано`
- `## Логи и инфраструктурная диагностика`
- `## Изменения в инфраструктуре`
- `## Rollback / mitigation`
- `## Проверки`
- `## Проверенные чек-листы`
- `## Риски и ограничения`
- `## Рекомендации / следующий шаг`
- `## Закрытие`

Для doc-stage ролей (`pm`, `sa`, `em`, `km`):
- `## Контекст stage/role`
- `## Основание для изменений`
- `## Что обновлено в артефактах`
- `## Traceability`
- `## Проверки`
- `## Проверенные чек-листы`
- `## Риски и ограничения`
- `## Рекомендации / следующий шаг`
- `## Закрытие`

Для reviewer-run:
- новый PR не создаётся;
- review summary оформляется в существующем PR с exact секциями:
  - `## Findings`
  - `## Evidence`
  - `## Проверенные guides/checklists`
  - `## Blocking / non-blocking`

## Информационная архитектура документации (обязательна)

| Каталог | Что хранится | Запрещено смешивать |
|---|---|---|
| `docs/product/` | продуктовые требования, роли, labels/stage policy, charter/brief/constraints | delivery-планы, day-эпики, runbooks |
| `docs/architecture/` | C4, ADR, API/data model, RBAC/prompt policy, design alternatives | спринт-план и release backlog |
| `docs/delivery/` | delivery plan, sprint/epic документы, issue map, traceability, process requirements | дублирование product source-of-truth и архитектурных контрактов |
| `docs/ops/` | production runbooks и эксплуатационные инструкции | stage-планирование и product scope |
| `docs/templates/` | канонические markdown-шаблоны документов | фактические артефакты этапов вместо шаблонов |

Правила:
- Один документ описывает одну цель и один уровень абстракции.
- Если документ описывает процесс/политику, он должен ссылаться на source-of-truth документ более высокого уровня.
- Канонический root navigation path проекта: `docs/index.md`.
- В каждом доменном каталоге `docs/<domain>/` обязателен `README.md` с картой source-of-truth документов и вложенных подпапок.
- Инициативные/stage-specific пакеты и handover-наборы размещаются в специализированных подпапках, а не на корне домена:
  - для архитектурных инициатив использовать `docs/architecture/initiatives/<slug>/`;
  - для эксплуатационных handover-пакетов использовать `docs/ops/handovers/<slug>/`;
  - для delivery уже закреплены специализированные подпапки `docs/delivery/sprints/` и `docs/delivery/epics/`.
- `docs/templates/` содержит только шаблоны и инструкции по их применению; проектные индексы, handover-пакеты и фактические source-of-truth документы туда не помещаются.
- Любой перенос документов выполняется только по явной migration-map с форматом
  `old path -> new path -> owner_role -> affected links/issues -> migration note`.
- В том же PR, где меняются пути документов, обязательно синхронно обновляются:
  - `services.yaml/spec.projectDocs`;
  - `services.yaml/spec.roleDocTemplates` (если затронуты шаблоны);
  - внутренние markdown-ссылки;
  - `docs/delivery/issue_map.md`;
  - `docs/delivery/requirements_traceability.md`;
  - открытые GitHub issues/PR, где использовались старые doc-path или branch-specific blob links.
- Любой новый документ сразу добавляется в релевантный индекс (`delivery/sprints/README.md`, `delivery/epics/README.md`, `issue_map`, `requirements_traceability`).

## Ролевая матрица шаблонов документов (обязательна)

Эта матрица обязательна не только для людей, но и для runtime-конфигурации:
- `services.yaml/spec.roleDocTemplates` должен оставаться синхронным с таблицей ниже;
- `services.yaml/spec.projectDocs` должен давать роли доступ к релевантным source-of-truth каталогам;
- любые расхождения считаются documentation governance drift и исправляются в том же PR.

| Роль | Основные stage | Обязательные шаблоны (`docs/templates/*.md`) |
|---|---|---|
| PM | `run:intake`, `run:vision`, `run:prd` | `problem.md`, `scope_mvp.md`, `constraints.md`, `brief.md`, `project_charter.md`, `success_metrics.md`, `prd.md`, `nfr.md`, `user_story.md` |
| EM | `run:plan`, `run:release`, `run:release:revise` | `delivery_plan.md`, `epic.md`, `definition_of_done.md`, `release_plan.md`, `release_notes.md`, `rollback_plan.md` |
| SA | `run:arch`, `run:design` | `c4_context.md`, `c4_container.md`, `adr.md`, `alternatives.md`, `api_contract.md`, `data_model.md`, `design_doc.md`, `migrations_policy.md` |
| Dev | `run:dev` | `user_story.md`, `definition_of_done.md` + обязательная синхронизация traceability-доков |
| QA | `run:qa`, `run:qa:revise`, `run:postdeploy`, `run:postdeploy:revise` | `test_strategy.md`, `test_plan.md`, `test_matrix.md`, `regression_checklist.md`, `postdeploy_review.md` |
| SRE | `run:ops`, `run:ops:revise`, `run:release`, `run:release:revise`, `run:postdeploy`, `run:postdeploy:revise`, `run:ai-repair` | `runbook.md`, `monitoring.md`, `alerts.md`, `slo.md`, `incident_playbook.md`, `incident_postmortem.md` |
| KM | cross-stage traceability, `run:doc-audit`, `run:doc-audit:revise`, `run:self-improve`, `run:self-improve:revise` | `issue_map.md`, `delivery_plan.md`, `roadmap.md`, `docset_issue.md`, `docset_pr.md` |
| Reviewer | pre-review (`need:reviewer`) | шаблоны не генерирует; публикует findings/commentary в PR |

## Еженедельный цикл спринта

### 1. Sprint Start (день начала недели)
- Проверить актуальность требований и ограничений.
- Сформировать/актуализировать план спринта и набор эпиков по дням.
- Для каждого эпика задать priority (`P0/P1/P2`) и ожидаемые артефакты дня.
- Провести DoR-check.

Обязательные артефакты:
- Sprint plan: `docs/delivery/sprints/s<номер>/sprint_s<номер>_<краткое-имя>.md` (актуальный sprint-file недели).
- Epic catalog: `docs/delivery/epics/s<номер>/epic_s<номер>.md` и daily epic docs в `docs/delivery/epics/s<номер>/`.
- `docs/delivery/issue_map.md` и `docs/delivery/requirements_traceability.md`.

### 2. Daily Execution (каждый рабочий день спринта)
- Реализовать задачи текущего дневного эпика.
- Выполнить merge в `main`.
- Подтвердить автоматический deploy на production.
- Выполнить ручной smoke-check и зафиксировать результат.
- Обновить документацию при изменении API/data model/webhook/RBAC/процессов.
- Для `run:qa` и `run:qa:revise` проверять новые/изменённые ручки через DNS Kubernetes namespace (service-to-service path), без блокировки на интерактивный OAuth browser-flow.

Daily gate (must pass):
- PR/merge только при green CI.
- Production deployment успешен.
- Smoke-check успешен или заведен блокер с решением.
- Документация синхронизирована.

### 2.1 Mainline Hygiene для `run:dev` и `run:dev:revise` (обязательно)
- Цель: детерминированно синхронизировать рабочую PR-ветку с `main` перед каждой revise-итерацией и исключить скрытые conflict-markers.
- Область применения:
  - все `run:dev` implementation PR;
  - обязательно для Sprint S7 execution issue-потоков `#243..#260`.
- Перед каждым push в открытую PR-ветку выполняется единый порядок:
  1. `git fetch origin --prune`
  2. `git checkout <pr-branch>`
  3. `git rebase origin/main`
  4. При конфликте: разрешить конфликт, `git add <file>`, `git rebase --continue`; при невозможности корректно разрешить конфликт — `git rebase --abort`, зафиксировать блокер и не пушить частично конфликтное состояние.
  5. Проверить отсутствие conflict-markers в рабочем дереве: `rg -n '^(<<<<<<<|=======|>>>>>>>)' .`
  6. Выполнить релевантные проверки (tests/lint/build/doc checks) в рамках scope PR.
  7. Публиковать rewritten историю только через `git push --force-with-lease origin <pr-branch>`.
- Запрещено:
  - делать `git merge origin/main` в рабочую PR-ветку для revise-итераций;
  - использовать `git push --force` вместо `--force-with-lease`;
  - открывать/обновлять review gate при найденных conflict-markers.
- Обязательный rebase-checklist для PR body:
  - [ ] Ветка синхронизирована с актуальным `origin/main` через `git rebase origin/main`.
  - [ ] Проверка `rg -n '^(<<<<<<<|=======|>>>>>>>)' .` не нашла conflict-markers.
  - [ ] После rebase повторно выполнены релевантные проверки по scope PR.
  - [ ] Публикация обновлённой истории выполнена через `git push --force-with-lease`.

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
- Mainline hygiene gate: рабочая PR-ветка ребейзнута на актуальный `origin/main`, conflict-markers отсутствуют, rewritten history опубликована только через `--force-with-lease`.
- Merge gate: green CI + pre-review (`reviewer`) + финальное ревью Owner + синхронная документация.
- Deploy gate: production deployment success + ручной smoke.
- Close gate: regression pass + согласованный backlog следующего спринта.

## Правило разрешения противоречий
- Если задача противоречит `docs/design-guidelines/**` или source-of-truth требованиям, работа останавливается.
- Предлагаются варианты решения с trade-offs.
- Финальное решение фиксируется в документации и утверждается Owner.

## Апрув
- request_id: owner-2026-02-06-process
- Решение: approved
- Комментарий: Процесс weekly sprint и doc governance утверждён.
