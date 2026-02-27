---
doc_id: PRD-CK8S-S7-I220
type: prd
title: "Issue #220 — PRD: Sprint S7 MVP readiness gap closure"
status: in-review
owner_role: PM
created_at: 2026-02-27
updated_at: 2026-02-27
related_issues: [212, 218, 220, 222, 216]
related_prs: []
related_docsets:
  - docs/delivery/sprints/s7/sprint_s7_mvp_readiness_gap_closure.md
  - docs/delivery/epics/s7/epic_s7.md
  - docs/delivery/issue_map.md
  - docs/delivery/requirements_traceability.md
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-27-issue-220-prd"
---

# PRD: Sprint S7 MVP readiness gap closure streams (Issue #220)

## TL;DR
- Что строим: execution-ready PRD-декомпозицию Sprint S7 по потокам `S7-E01..S7-E18`.
- Для кого: Owner, PM, EM, SA, Dev, QA, SRE, KM.
- Почему: без FR/AC/NFR и sequencing на stream-уровне переход в `run:arch` и далее в `run:dev` приводит к scope drift.
- MVP: полное покрытие потоков `user story + FR + AC + NFR + edge cases + expected evidence + dependencies`.
- Критерии успеха: parity-gate перед `run:dev`, формализованный dependency graph и созданная follow-up issue `#222` для `run:arch`.

## Проблема и цель
- Problem statement:
  - Vision stage (`#218`) зафиксировал KPI/baseline, но execution-flow остаётся недодетализированным для архитектурного handover.
- Цели:
  - формализовать stream-level требования и критерии приемки;
  - зафиксировать deterministic sequencing и dependency graph;
  - исключить старт `run:dev` без полной decomposition parity.
- Почему сейчас:
  - Sprint S7 закрывает MVP readiness gaps и требует строгой последовательности `prd -> arch -> design -> plan -> dev`.

## Пользователи / Персоны
- Owner: принимает решения по readiness и go/no-go.
- PM/EM: управляют scope, приоритизацией, traceability и quality gates.
- SA: переводит PRD-пакет в архитектурные решения и ownership boundaries.
- Dev/QA/SRE/KM: используют stream-level AC/NFR как контракт исполнения и проверки.

## Scope
### In scope
- `S7-E01..S7-E18` с полной PRD-декомпозицией.
- Sequencing/dependency rules и parity-gate перед `run:dev`.
- Handover в `run:arch` через issue `#222`.

### Out of scope
- Кодовые изменения.
- Изменение архитектурных границ сервисов в рамках данного этапа.
- Переопределение label taxonomy и stage-модели.

## Cross-stream functional requirements
- FR-220-01: Для каждого execution-эпика должен быть зафиксирован пакет `user story + FR + AC + NFR + edge cases + expected evidence`.
- FR-220-02: Sequencing исполнения фиксируется dependency graph и wave-моделью.
- FR-220-03: Перед `run:dev` действует parity-gate `approved_execution_epics == run:dev implementation issues`.
- FR-220-04: При `coverage_ratio != 1.0` переход в `run:dev` блокируется до owner decision.
- FR-220-05: Для `S7-E13` обязательно покрытие revise-loop `run:qa:revise` в stage/labels policy.
- FR-220-06: Для `S7-E14` QA acceptance для новых/изменённых ручек выполняется через Kubernetes DNS path.
- FR-220-07: Для reliability-потоков (`S7-E11`, `S7-E16`, `S7-E17`) обязательны evidence регрессий/устойчивости.
- FR-220-08: Все stage-переходы оформляются через continuity-rule с созданием follow-up issue без trigger-лейбла.

## Stream-level decomposition (`S7-E01..S7-E18`)

| Epic | User story | FR (stream) | AC (Given/When/Then) | NFR (stream) | Edge cases | Expected evidence | Dependencies |
|---|---|---|---|---|---|---|---|
| S7-E01 | Как Dev, я хочу deterministic rebase/mainline process, чтобы revise-итерации не ломали merge path. | FR-E01-1: формализовать rebase policy для активных PR-веток; FR-E01-2: исключить merge-conflict surprises перед review gate. | Given активная PR-ветка; When запускается revise-итерация; Then ветка синхронизируется с `main` без блокирующих конфликтов. | NFR-E01-1: 100% revise-итераций должны завершаться без незакрытых conflict markers. | Одновременные правки тех же файлов; force-push после review. | Git history, PR timeline, checklist rebase policy. | foundation для всех stream execution. |
| S7-E02 | Как Owner, я хочу убрать не-MVP разделы в sidebar, чтобы UI отражал только готовый scope. | FR-E02-1: удалить не-MVP navigation entries; FR-E02-2: удалить связанный dead code/routes. | Given пользователь в staff UI; When открывает sidebar и роутинг; Then доступны только MVP-разделы без скрытых не-MVP переходов. | NFR-E02-1: не допускаются broken links после cleanup. | Сохранённые deep-links у пользователей; stale bookmarks. | Navigation diff, route inventory, smoke-check навигации. | after S7-E01. |
| S7-E03 | Как пользователь, я хочу убрать глобальный фильтр, чтобы не было ложных ожиданий и лишней сложности UI. | FR-E03-1: удалить глобальный filter UI; FR-E03-2: удалить все связанные зависимости в коде. | Given экран с прежним filter-entry; When пользователь открывает разделы; Then filter-контур отсутствует и поведение консистентно. | NFR-E03-1: отсутствие regressions в list/load состояниях. | Скрытые зависимости в дочерних компонентах. | Code search report, UI smoke до/после. | after S7-E01. |
| S7-E04 | Как Owner, я хочу убрать runtime-deploy/images UI-контуры из MVP, чтобы не размывать scope. | FR-E04-1: удалить runtime-deploy/images разделы из MVP навигации; FR-E04-2: очистить связанный UI code path. | Given MVP navigation; When пользователь проходит сценарии; Then runtime-deploy/images разделы недоступны в MVP scope. | NFR-E04-1: zero broken transitions по удалённым путям. | Внешние ссылки на удалённые страницы. | UI/nav diff, dead-code cleanup report. | after S7-E01. |
| S7-E05 | Как оператор Agents, я хочу таблицу без badge `Скоро` и лишних колонок, чтобы интерфейс был production-ready. | FR-E05-1: убрать badge `Скоро`; FR-E05-2: пересобрать таблицу на MVP-состав полей. | Given страница Agents; When пользователь открывает таблицу; Then отображаются только утверждённые MVP-поля и без placeholder badge. | NFR-E05-1: таблица должна корректно рендериться на пустом и полном наборе данных. | Пустой список; длинные значения колонок. | Скриншоты UI, acceptance checklist по колонкам. | after S7-E02/S7-E03/S7-E04. |
| S7-E06 | Как оператор, я хочу консистентные defaults runtime mode и locale при import, чтобы избегать ручных исправлений. | FR-E06-1: зафиксировать runtime mode defaults; FR-E06-2: зафиксировать locale policy и bulk-update правила. | Given импорт агентов; When применяется импорт-конфигурация; Then runtime mode/locale defaults назначаются детерминированно. | NFR-E06-1: 100% импорт-кейсов применяют policy без drift. | Частичный импорт; конфликт owner-locale и repo-locale. | Import scenario matrix, policy notes, QA evidence. | after S7-E05. |
| S7-E07 | Как оператор, я хочу selector `repo|db` для source prompt templates, чтобы управлять поведением worker. | FR-E07-1: добавить policy-контур выбора `repo|db`; FR-E07-2: синхронизировать source выбор с worker behavior. | Given settings агента; When пользователь выбирает source prompt templates; Then worker использует выбранный source по контракту. | NFR-E07-1: source switch должен быть идемпотентным и audit-traceable. | Source недоступен; рассинхрон версий repo/db. | API/UI contract evidence, worker run checks. | after S7-E06. |
| S7-E08 | Как Owner, я хочу hardening массовых операций Agents, чтобы управлять конфигурацией без ручной рутины. | FR-E08-1: зафиксировать набор batch-операций; FR-E08-2: для каждой операции определить guardrails и rollback expectations. | Given несколько выбранных агентов; When запускается batch-операция; Then операция выполняется согласно policy и отражает partial failures. | NFR-E08-1: p95 выполнения batch-операции стабилен и наблюдаем. | Частичный успех; конфликт прав доступа. | UX scenarios, operation logs, rollback checks. | after S7-E07. |
| S7-E09 | Как QA/SRE, я хочу убрать колонку run type и иметь deterministic delete namespace, чтобы ускорить диагностику. | FR-E09-1: убрать run type column; FR-E09-2: обеспечить доступный delete namespace action по policy. | Given RunDetails; When оператор открывает действия; Then видит только релевантные поля и рабочий delete namespace path. | NFR-E09-1: delete action должен иметь подтверждение и audit trail. | Namespace уже удалён; namespace в `Terminating`. | Runs UI screenshots, action logs, negative-case checks. | after S7-E01. |
| S7-E10 | Как SRE, я хочу cancel/stop для зависших deploy tasks, чтобы безопасно останавливать инцидентные операции. | FR-E10-1: добавить cancel/stop control для зависших задач; FR-E10-2: закрепить safety guardrails. | Given deploy task в hanging state; When оператор нажимает cancel/stop; Then task завершается контролируемо и фиксируется в аудите. | NFR-E10-1: cancel operation должна быть идемпотентной при повторе. | Race статусов при отмене; повторная отмена. | Ops scenarios, task lifecycle logs, guardrail evidence. | after S7-E09. |
| S7-E11 | Как PM/Owner, я хочу корректный `mode:discussion`, чтобы orchestration не ломала обсуждения и stage continuity. | FR-E11-1: обеспечить policy-correct behavior `mode:discussion`; FR-E11-2: исключить ложные trigger/flow transitions. | Given issue с `mode:discussion`; When происходят комментарии и label transitions; Then discussion flow не провоцирует некорректный stage launch. | NFR-E11-1: zero false triggers в regression наборе. | Одновременный comment + label change; снятие mode в активной сессии. | Flow-event evidence, issue timeline checks. | foundation для S7-E13/S7-E16. |
| S7-E12 | Как Owner, я хочу финальный readiness gate с evidence bundle, чтобы принять go/no-go обоснованно. | FR-E12-1: собрать единый readiness пакет по stage chain; FR-E12-2: зафиксировать формальное решение go/no-go. | Given завершены P0/P1 потоки; When выполняется final gate; Then доступно консолидированное evidence и принято решение. | NFR-E12-1: полнота evidence = 100% обязательных gate-артефактов. | Частично закрытые P0; незавершённые dependencies. | Consolidated readiness report, release/postdeploy/ops/doc-audit evidence. | after all P0 streams and S7-E14/S7-E18. |
| S7-E13 | Как QA, я хочу revise-loop `run:qa:revise`, чтобы корректно отрабатывать review feedback на QA-этапе. | FR-E13-1: добавить `run:qa:revise` в stage/labels policy; FR-E13-2: покрыть transition сценарии revise-loop для QA. | Given QA stage в review; When получены changes requested; Then запускается deterministic `run:qa:revise` path без ambiguity. | NFR-E13-1: revise resolver должен быть детерминированным. | Конфликт stage labels; revise без QA context. | Policy diffs, transition checks, QA revise evidence. | after S7-E11. |
| S7-E14 | Как QA, я хочу проверять новые/изменённые ручки через Kubernetes DNS path, чтобы acceptance не блокировалась OAuth-browser flow. | FR-E14-1: формализовать DNS-path check policy; FR-E14-2: включить DNS evidence в обязательные QA artifacts. | Given изменённая ручка; When QA запускает acceptance; Then проверка выполняется через service DNS path и логируется как evidence. | NFR-E14-1: 100% применимых ручек покрыты DNS-path checks. | Недоступность сервиса; частичный DNS resolution. | QA matrix, command/log evidence, traceability links. | after S7-E13. |
| S7-E15 | Как оператор Agents, я хочу refresh prompt templates из repo с version bump, чтобы поддерживать актуальность шаблонов. | FR-E15-1: добавить refresh action из repo; FR-E15-2: каждое обновление создаёт новую версию шаблона. | Given доступен repo source; When пользователь запускает refresh; Then шаблоны обновляются и version history фиксирует bump. | NFR-E15-1: refresh не должен нарушать текущий active template. | Невалидный шаблон; одновременное обновление. | Version history, diff evidence, negative-case handling notes. | after S7-E07. |
| S7-E16 | Как Owner, я хочу исключить false-failed для `run:intake:revise`, чтобы статус отражал фактический результат. | FR-E16-1: устранить false-failed сценарий; FR-E16-2: выровнять completion/status signals. | Given `run:intake:revise` завершён успешно; When фиксируется финальный статус; Then run marked successful без ложного failed. | NFR-E16-1: zero false-failed incidents в regression period. | Duplicate callbacks; race при финализации. | Run logs/status comparison, regression evidence. | after S7-E11. |
| S7-E17 | Как KM, я хочу гарантированно читать/перезаписывать self-improve session snapshot, чтобы диагностика была воспроизводимой. | FR-E17-1: обеспечить доступность session snapshot; FR-E17-2: обеспечить корректную перезапись без потери контекста. | Given run:self-improve; When агент читает/пишет snapshot; Then snapshot доступен и консистентен после перезаписи. | NFR-E17-1: snapshot operations устойчивы к retry/duplicate write. | Повреждённый snapshot; повторная запись с тем же ключом. | Session retrieval/rewrite logs, consistency checks. | after S7-E15. |
| S7-E18 | Как PM/KM, я хочу единый issue/PR/doc governance standard, чтобы снизить quality drift delivery-артефактов. | FR-E18-1: закрепить единый стандарт issue/PR структуры; FR-E18-2: закрепить doc IA и role-template matrix как обязательные контракты. | Given stage issue/PR; When артефакт готовится к review; Then структура соответствует стандарту и проверяется по checklist. | NFR-E18-1: governance compliance 100% для S7 артефактов. | Legacy документы вне стандарта; смешение типов документов. | Governance checklist, docs diff, traceability sync. | after S7-E13, before S7-E12 gate. |

## Dependency graph и sequencing constraints
- Foundation prerequisites: `S7-E01`, `S7-E11`, `S7-E13`.
- UI cleanup block: `S7-E02..S7-E05`.
- Agents lifecycle block: `S7-E06..S7-E08` + `S7-E15` + `S7-E17`.
- Runs/deploy reliability block: `S7-E09`, `S7-E10`, `S7-E16`.
- Closeout block: `S7-E14`, `S7-E18`, затем `S7-E12`.
- Constraint C-220-1: любой `run:dev` issue создаётся только после owner-approved execution-epic.
- Constraint C-220-2: `coverage_ratio = run_dev_issues / approved_execution_epics` обязан быть `1.0`.

## Acceptance criteria (Issue #220)
- [x] PRD-пакет по `S7-E01..S7-E18` зафиксирован в markdown-артефактах.
- [x] Для каждого stream задокументированы `user story`, `FR`, `AC`, `NFR`, `edge cases`, `expected evidence`.
- [x] Зафиксированы dependency graph и sequencing constraints для handover в `run:arch`.
- [x] Зафиксировано parity-правило и блокирующие условия до `run:dev`.
- [x] Создана follow-up issue `#222` для `run:arch` без trigger-лейбла.

## Non-goals
- Реализация stream-эпиков в коде.
- Ревизия глобальной архитектуры вне контекста S7 readiness gaps.
- Изменение security-policy и базовой taxonomy labels.

## Риски и допущения
| Type | ID | Описание | Митигирующее действие | Статус |
|---|---|---|---|---|
| risk | RSK-220-A | Недостаточная детализация ownership в `run:arch` приведёт к rework на `run:design` | Зафиксировать ownership matrix в architecture stage | open |
| risk | RSK-220-B | Нарушение sequencing создаст конкурирующие implementation paths | Использовать wave execution и dependency gate | open |
| risk | RSK-220-C | Нарушение parity-гейта приведёт к неполному dev scope | Блокировать `run:dev` при `coverage_ratio != 1.0` | open |
| assumption | ASM-220-A | Baseline KPI из `#218` достаточно для architecture decomposition | Проверить в `run:arch` и при необходимости эскалировать `need:input` | accepted |
| assumption | ASM-220-B | Owner сохраняет последовательный stage-flow без параллельных конфликтующих `run:*` | Контролировать через issue labels policy | accepted |

## Handover в `run:arch`
- Следующий этап: `run:arch`.
- Follow-up issue: `#222`.
- Trigger-лейбл ставит Owner после review PRD-пакета.
- Что обязательно передать:
  - stream-level PRD-декомпозицию `S7-E01..S7-E18`;
  - dependency graph и sequencing constraints;
  - parity-gate policy и условия блокировки перехода в `run:dev`.
