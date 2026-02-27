---
doc_id: EPC-CK8S-S7-D2
type: epic
title: "Epic S7 Day 2: Vision для закрытия MVP readiness gaps (Issue #218)"
status: in-review
owner_role: PM
created_at: 2026-02-27
updated_at: 2026-02-27
related_issues: [212, 218, 220, 216]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-27-issue-218-vision"
---

# Epic S7 Day 2: Vision для закрытия MVP readiness gaps (Issue #218)

## TL;DR
- Зафиксирована vision-рамка закрытия MVP readiness разрывов по всем потокам `S7-E01..S7-E18`.
- Для каждого потока определены измеримые KPI и baseline для входа в `run:prd`: user story, acceptance criteria, edge cases, expected evidence.
- Введено обязательное governance-правило continuity:
  - каждая stage-issue обязана создавать следующую stage-issue без trigger-лейбла;
  - до входа в `run:dev` количество implementation issues должно быть равно количеству утверждённых execution-эпиков (`1:1`).

## Priority
- `P0`.

## Vision charter

### Mission statement
Довести Sprint S7 до проверяемого состояния MVP readiness, при котором критичные продуктовые, stage-flow и governance разрывы закрываются не декларацией, а измеримыми результатами и доказательствами выполнения в delivery-цепочке `vision -> prd -> arch -> design -> plan -> dev -> qa -> release -> postdeploy -> ops -> doc-audit`.

### Цели и expected outcomes
1. Сделать закрытие `S7-E01..S7-E18` управляемым через KPI, quality gates и единый evidence-контур.
2. Обеспечить предсказуемую stage-continuity без разрывов в handover между этапами.
3. Зафиксировать формальные условия входа в `run:dev`, исключающие недодекомпозированный execution scope.

### Пользователи и стейкхолдеры
- Основные пользователи: Owner, PM, EM, Dev, QA, SRE, KM.
- Стейкхолдеры: команды `control-plane`, `api-gateway`, `agent-runner`, `staff/web-console`.
- Владелец решения: Owner.

## Scope boundaries

### In scope (Vision stage)
- KPI/success-metrics и measurable readiness criteria для `S7-E01..S7-E18`.
- Baseline-декомпозиция каждого execution-эпика: user story, AC, edge cases, expected evidence.
- Правила continuity и execution decomposition до входа в `run:dev`.
- Handover в `run:prd` с обязательным шаблоном для создания следующей stage-issue.

### Out of scope
- Кодовые изменения сервисов, frontend и инфраструктуры.
- Изменение архитектурных границ и базовой label taxonomy.
- Исполнение самих implementation-эпиков (это scope стадий `run:dev+`).

## Success metrics

### North Star
| ID | Метрика | Определение | Источник | Целевое значение |
|---|---|---|---|---|
| NSM-S7-01 | MVP readiness stream completion | Доля потоков `S7-E01..S7-E18`, закрытых с подтверждённым evidence и owner-approved статусом | `issue_map`, `requirements_traceability`, stage-issues/PR evidence | `18/18` (100%) до Sprint S7 close |

### Supporting metrics
| ID | Метрика | Формула | Источник | Target |
|---|---|---|---|---|
| SM-S7-01 | Stage continuity compliance | `created_next_stage_issues / completed_stage_issues` | GitHub issues + traceability docs | 100% |
| SM-S7-02 | Execution decomposition parity | `dev_issues_created / approved_execution_epics` | `run:plan` артефакт + issue map | 1.0 |
| SM-S7-03 | P0 stream closure on time | `P0 closed before final readiness gate / total P0` | stage issues + sprint S7 | >= 95% |
| SM-S7-04 | QA evidence completeness | `streams with DNS-path QA evidence / streams requiring API/runtime checks` | QA artifacts + issue map | 100% |
| SM-S7-05 | Stage reliability incidents | Количество false-failed/continuity-break инцидентов по stage-run | run evidence + flow events | 0 |
| SM-S7-06 | Documentation governance compliance | `issues/PR/docs, соответствующие стандарту / total S7 artifacts` | docs review checklist | 100% |

### Readiness KPI по потокам `S7-E01..S7-E18`
| Epic | KPI ID | Измерение | Целевое значение | Expected evidence |
|---|---|---|---|---|
| S7-E01 | KPI-S7-E01 | Доля PR-итераций без конфликтов после rebase | 100% | PR timeline + merge/rebase history |
| S7-E02 | KPI-S7-E02 | Количество не-MVP разделов в sidebar/routes | 0 в MVP scope | UI diff + route inventory |
| S7-E03 | KPI-S7-E03 | Наличие глобального filter-кода вне scope | 0 | code search evidence + UI behavior checks |
| S7-E04 | KPI-S7-E04 | Наличие runtime-deploy/images UI-контуров в MVP | 0 | UI/nav diff + dead-code cleanup report |
| S7-E05 | KPI-S7-E05 | Наличие badge `Скоро` и лишних колонок в Agents | 0 | UI screenshots + acceptance checklist |
| S7-E06 | KPI-S7-E06 | Консистентность runtime mode + locale defaults при import | 100% кейсов | import scenario matrix + QA evidence |
| S7-E07 | KPI-S7-E07 | Наличие selector `repo|db` и policy-консистентность worker | 100% случаев | API/UI contract evidence + worker behavior checks |
| S7-E08 | KPI-S7-E08 | Покрытие массовых операций в Agents UX | 100% заявленного scope | UX сценарии + regression notes |
| S7-E09 | KPI-S7-E09 | Наличие run type колонки и availability namespace delete | run type = 0, delete action = 100% | Runs UI screenshots + action logs |
| S7-E10 | KPI-S7-E10 | Доля зависших deploy tasks без cancel/stop пути | 0 | runtime deploy сценарии + ops evidence |
| S7-E11 | KPI-S7-E11 | Частота некорректного поведения `mode:discussion` | 0 регрессий | orchestration checks + flow evidence |
| S7-E12 | KPI-S7-E12 | Проход финального readiness gate по цепочке stage | 100% required gates | consolidated readiness package |
| S7-E13 | KPI-S7-E13 | Поддержка `run:qa:revise` в policy + transitions | 100% | policy docs + revise scenario evidence |
| S7-E14 | KPI-S7-E14 | Покрытие QA проверок новых/изменённых ручек через DNS path | 100% применимых ручек | QA matrix + DNS-path run logs |
| S7-E15 | KPI-S7-E15 | Наличие repo refresh + version bump для prompt templates | 100% | UI/API evidence + version history |
| S7-E16 | KPI-S7-E16 | Количество false-failed для `run:intake:revise` | 0 | run status audit evidence |
| S7-E17 | KPI-S7-E17 | Доля self-improve запусков с доступным session snapshot | 100% | session retrieval/rewrite evidence |
| S7-E18 | KPI-S7-E18 | Соответствие issue/PR/doc IA + role templates стандарту | 100% | governance checklist + docs diff |

## Baseline для execution-эпиков (`S7-E01..S7-E18`)

| Epic | User story | Acceptance criteria baseline | Edge cases baseline | Expected evidence baseline |
|---|---|---|---|---|
| S7-E01 | Как Dev, я хочу стабильный rebase/mainline процесс, чтобы revise-итерации не ломали merge path. | 1) Описан и применён rebase policy для активных PR-веток. 2) Нет блокирующих merge conflicts к моменту review gate. | Одновременные правки в одних и тех же файлах; force-push после review. | История коммитов/PR, чек rebase policy, отсутствие конфликтов в финальной итерации. |
| S7-E02 | Как Owner, я хочу убрать не-MVP разделы в navigation, чтобы UI отражал только готовый scope. | 1) Не-MVP разделы удалены из sidebar/routes. 2) Связанный dead code удалён. | Deep-link на удалённый маршрут; stale bookmarks пользователей. | Diff navigation/routes, список удалённых страниц, smoke-check навигации. |
| S7-E03 | Как пользователь staff UI, я хочу убрать глобальный фильтр, чтобы избежать ложных ожиданий и лишней сложности. | 1) Глобальный фильтр полностью удалён. 2) Связанные зависимости и UI-следы удалены. | Скрытые зависимости фильтра в дочерних компонентах. | Поиск по коду + UI smoke до/после. |
| S7-E04 | Как Owner, я хочу исключить runtime-deploy/images контуры из MVP UI, чтобы не смешивать scope readiness. | 1) UI-секции runtime-deploy/images удалены из MVP контуров. 2) Связанный код очищен. | Внешние ссылки/кнопки на удалённые разделы. | UI diff, route cleanup report, проверка отсутствия битых переходов. |
| S7-E05 | Как оператор Agents, я хочу чистую таблицу без `Скоро` и лишних колонок, чтобы UI был функционально завершён. | 1) Badge `Скоро` удалён. 2) Таблица пересобрана по утверждённому MVP-составу. | Пустой список агентов; длинные значения полей. | Скриншоты UI, checklist по полям таблицы, smoke-отчёт. |
| S7-E06 | Как оператор, я хочу консистентные defaults runtime mode и locale при import, чтобы избежать ручных исправлений. | 1) Runtime mode defaults формализованы. 2) Locale policy и bulk update правила определены и проверяемы. | Частичный импорт; конфликт owner-locale и repo-locale. | Матрица import-сценариев, policy notes, acceptance evidence. |
| S7-E07 | Как оператор, я хочу выбирать source prompt templates (`repo`/`db`), чтобы контролировать поведение worker без ручных обходов. | 1) Selector `repo|db` формализован в policy/контрактах. 2) Поведение worker согласовано с выбранным source. | Source недоступен; рассинхрон версий между repo и db. | Контрактные тест-кейсы, UI/API сценарии, worker run evidence. |
| S7-E08 | Как Owner, я хочу hardening массовых операций Agents UX, чтобы управлять большим набором агентов без ручной рутины. | 1) Список массовых операций формализован. 2) Для каждой операции есть AC и rollback/guardrails. | Частичный успех batch-операции; конфликт прав доступа. | UX сценарии, лог массовых операций, rollback checks. |
| S7-E09 | Как QA/SRE, я хочу в Runs UI видеть только релевантные поля и всегда иметь delete namespace, чтобы ускорить диагностику. | 1) Колонка run type удалена. 2) Delete namespace action доступен детерминированно по policy. | Namespace уже удалён; namespace в terminating state. | Скриншоты Runs/RunDetails, action evidence, negative-case checks. |
| S7-E10 | Как SRE, я хочу cancel/stop для зависших deploy tasks, чтобы управлять инцидентами без ручных обходов. | 1) Для зависших tasks существует явный cancel/stop path. 2) Guardrails предотвращают опасные действия. | Гонка статусов task в момент отмены; повторный cancel. | Ops-сценарии, task lifecycle logs, guardrail evidence. |
| S7-E11 | Как PM/Owner, я хочу корректный `mode:discussion`, чтобы stage orchestration не ломала обсуждение и handover. | 1) `mode:discussion` ведёт себя согласно policy. 2) Нет ложных trigger/flow переходов. | Одновременный комментарий и label transition; снятие режима в процессе run. | Flow-event evidence, issue timeline, regression сценарии. |
| S7-E12 | Как Owner, я хочу финальный readiness gate с доказательствами, чтобы принять go/no-go по MVP обоснованно. | 1) Собран единый evidence bundle по stage-цепочке. 2) Зафиксировано решение go/no-go. | Частично закрытые P0; незавершённые dependencies. | Consolidated readiness report, release/postdeploy/ops/doc-audit evidence. |
| S7-E13 | Как QA, я хочу revise-петлю `run:qa:revise`, чтобы корректно дорабатывать QA-артефакты по review feedback. | 1) `run:qa:revise` отражён в stage/labels policy. 2) revise loop подтверждён сценариями переходов. | Конфликт stage labels; запуск revise без исходного QA контекста. | Policy diffs, transition checks, QA revise run evidence. |
| S7-E14 | Как QA, я хочу проверять новые/изменённые ручки через Kubernetes DNS path, чтобы acceptance не зависела от UI OAuth-flow. | 1) Для всех применимых ручек определён DNS-path check. 2) Evidence включён в QA artifacts. | Недоступность сервиса в namespace; частичный DNS resolution. | QA matrix, команды/логи DNS-path проверок, traceability ссылки. |
| S7-E15 | Как оператор Agents, я хочу обновлять prompt templates из repo с version bump, чтобы поддерживать актуальность шаблонов. | 1) Есть действие refresh из repo. 2) Каждое обновление создаёт новую версию шаблона. | Невалидный шаблон в repo; конфликт одновременных обновлений. | Version history, diff evidence, negative-case handling notes. |
| S7-E16 | Как Owner, я хочу исключить false-failed статусы для `run:intake:revise`, чтобы статус run отражал фактический результат. | 1) Устранён сценарий false-failed. 2) Статус run консистентен с фактическим completion. | Дубли callback-событий; race при финализации run. | Сравнение run logs/status, regression scenario evidence. |
| S7-E17 | Как KM, я хочу гарантированно получать и перезаписывать session snapshot в self-improve, чтобы диагностика была воспроизводимой. | 1) Session snapshot всегда доступен для self-improve. 2) Перезапись snapshot работает без потери контекста. | Частично повреждённый snapshot; повторная запись с тем же ключом. | Session retrieval/rewrite evidence, diagnostic logs, consistency checks. |
| S7-E18 | Как PM/KM, я хочу единый governance-стандарт для issue/PR/документов, чтобы снизить quality drift в delivery. | 1) Зафиксирован и применён единый формат issue/PR. 2) Документация приведена к утверждённой IA и role-template matrix. | Legacy-документы вне стандарта; смешение типов документов в одном файле. | Governance checklist, docs diff, traceability updates. |

## Continuity и decomposition rules (обязательно)

### Rule C-01: Next-stage continuity
- Каждая stage-issue (`run:intake..run:plan`, `run:qa`, `run:release`, `run:postdeploy`, `run:ops`, `run:doc-audit`) обязана в конце этапа создать отдельную issue следующего этапа.
- Новая issue создаётся **без trigger-лейбла**; trigger ставит Owner после review.

### Rule C-02: Decomposition parity before `run:dev`
- До входа в `run:dev` должно выполняться равенство:
  - `approved_execution_epics_count == created_run_dev_issues_count`.
- Дополнительно:
  - `coverage_ratio = created_run_dev_issues_count / approved_execution_epics_count`;
  - `coverage_ratio` должен быть ровно `1.0`.

### Rule C-03: Gate on mismatch
- Если `coverage_ratio != 1.0`, stage переход в `run:dev` блокируется до устранения рассинхрона.
- Рассинхрон фиксируется как governance-risk и выносится в `need:input` для owner decision.

## Readiness criteria for `run:prd`
- [x] Mission/targets и KPI по `S7-E01..S7-E18` формализованы в vision-артефакте.
- [x] Для каждого execution-эпика зафиксированы user story, AC, edge cases, expected evidence.
- [x] Continuity rules (next-stage issue creation + decomposition parity) зафиксированы как обязательные.
- [x] Подготовлен handover в `run:prd` с шаблоном создания следующей stage-issue.
- [x] Создана follow-up issue `run:prd` без trigger-лейбла и добавлена в traceability.

## Acceptance criteria (Issue #218)
- [x] Обновлён vision-артефакт S7 с миссией, KPI и measurable readiness-метриками для `S7-E01..S7-E18`.
- [x] Для каждого epic-кандидата зафиксирован baseline: user story, acceptance criteria, edge cases, expected evidence.
- [x] Явно зафиксировано правило: до входа в `run:dev` формируется столько dev-issues, сколько execution-эпиков утверждено в scope.
- [x] В конце этапа создана отдельная follow-up issue для `run:prd` (без trigger-лейбла при создании).
- [x] В `run:prd` issue передан handover-блок с шаблоном создания следующей stage-задачи.

## Risks and product assumptions

| Type | ID | Описание | Митигирующее действие | Статус |
|---|---|---|---|---|
| risk | RSK-218-01 | Без decomposition parity возможен недопокрытый execution scope перед `run:dev` | Ввести обязательный parity gate (Rule C-02/C-03) | open |
| risk | RSK-218-02 | При отсутствии stage continuity chain возникает разрыв между stage-артефактами | Обязательное создание next-stage issue на каждом этапе | open |
| risk | RSK-218-03 | Часть KPI может быть не подтверждена evidence в QA/release цикле | Для каждого потока закрепить expected evidence до `run:prd` | open |
| assumption | ASM-218-01 | Intake backlog `S7-E01..S7-E18` покрывает текущие P0/P1 readiness gaps | Проверить и уточнить в `run:prd` при детализации FR/NFR | accepted |
| assumption | ASM-218-02 | Owner подтверждает последовательную stage-модель и ручную постановку trigger-лейблов | Сохранять owner-governed transitions на каждом этапе | accepted |

## Handover в следующий этап
- Следующий stage: `run:prd`.
- Follow-up issue: `#220` (создана без trigger-лейбла).
- В PRD-stage обязательно:
  - формализовать FR/AC/NFR по каждому execution-эпику;
  - уточнить dependency graph и sequencing для перехода в `run:arch`;
  - подготовить и создать issue следующего этапа `run:arch` без trigger-лейбла.

### Шаблон создания следующей stage-задачи (обязателен)
```md
## Контекст
- Родительская issue: #<текущая>
- Артефакт этапа: <ссылка на документ>
- Source of truth: <product/architecture/delivery ссылки>

## Проблема
<какой разрыв остаётся и почему следующий этап обязателен>

## Scope
### In scope
- ...

### Out of scope
- ...

## Acceptance Criteria
- [ ] Артефакт следующего этапа подготовлен и синхронизирован в traceability.
- [ ] Для каждого epic/stream зафиксированы: user story, AC, edge cases, expected evidence.
- [ ] Создана следующая stage-issue без trigger-лейбла.

## Риски и допущения
- ...

## Next stage handover
- Следующий этап: `run:<next-stage>`
- Что обязательно передать: <список>
```

## Связанные документы
- `docs/delivery/epics/s7/epic-s7-day1-mvp-readiness-intake.md`
- `docs/delivery/epics/s7/epic_s7.md`
- `docs/delivery/sprints/s7/sprint_s7_mvp_readiness_gap_closure.md`
- `docs/delivery/delivery_plan.md`
- `docs/delivery/requirements_traceability.md`
- `docs/delivery/issue_map.md`
- `docs/product/requirements_machine_driven.md`
- `docs/product/labels_and_trigger_policy.md`
- `docs/product/stage_process_model.md`
