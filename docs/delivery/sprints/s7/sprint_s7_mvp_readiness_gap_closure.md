---
doc_id: SPR-CK8S-0007
type: sprint-plan
title: "Sprint S7: MVP readiness gap closure (Issue #212)"
status: in-progress
owner_role: PM
created_at: 2026-02-27
updated_at: 2026-02-27
related_issues: [212, 218, 220, 223, 225, 226, 227, 228, 229, 230, 199, 201, 210, 216]
related_prs: [213, 215]
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-27-issue-212-intake"
---

# Sprint S7: MVP readiness gap closure (Issue #212)

## TL;DR
- Цель спринта: закрыть продуктовые и операционные разрывы, из-за которых MVP сейчас выглядит и работает как незавершённый.
- Фактический сигнал из intake (#212): в staff UI остаётся большое количество `comingSoon`-разделов, а stage-контур `run:doc-audit` не подтверждён как рабочий в реальном цикле.
- Sprint S7 фокусируется на завершении S6 release-continuity: `#199`/`#201` уже закрыты, текущая открытая зависимость — `#216` (`run:release`).

## Scope спринта
### In scope
- Формализация и приоритизация глобальных MVP-gaps (product + delivery + stage-flow).
- Закрытие открытого S6-блокера (`run:release`, Issue `#216`) как обязательной зависимости.
- Декомпозиция owner-замечаний в candidate backlog на 18 execution-эпиков (`S7-E01..S7-E18`).
- Подготовка release-ready цепочки `qa -> release -> postdeploy -> ops -> doc-audit`.

### Out of scope
- Post-MVP расширения (A2A swarm, custom-agent factory, marketplace шаблонов).
- Изменение базовой taxonomy labels и архитектурных ограничений платформы.
- Изменение security-policy (RBAC, approval-gates, secret governance).

## План эпиков по дням

| День | Эпик | Priority | Документ | Статус |
|---|---|---|---|---|
| Day 1 | Intake по MVP readiness gaps | P0 | `docs/delivery/epics/s7/epic-s7-day1-mvp-readiness-intake.md` | in-review (`#212`) |
| Day 2 | Vision: целевая картина MVP closeout, KPI и decomposition baseline | P0 | `docs/delivery/epics/s7/epic-s7-day2-mvp-readiness-vision.md` | in-review (`#218`) |
| Day 3 | PRD: FR/AC/NFR и sequencing для gap-closure streams | P0 | Issue `#220` (`run:prd`) | planned |
| Day 4 | Architecture: границы и ownership по stream'ам | P0 | TBD (`run:arch`) | planned |
| Day 5 | Design + Plan: execution package и quality gates | P0 | `docs/delivery/epics/s8/epic-s8-day1-go-refactoring-plan.md` (`run:plan`, Issue `#223`) | in-review |
| Day 6+ | Dev/QA/Release/Postdeploy/Ops/Doc-Audit | P0/P1 | TBD (`run:dev..run:doc-audit`) | planned |

## Candidate execution-эпики (`S7-E01..S8-E06`)

| Epic | Priority | Scope | Блокер/зависимость |
|---|---|---|---|
| S7-E01 | P0 | Rebase/mainline hygiene для PR revise-итераций | required before merge |
| S7-E02 | P0 | Sidebar cleanup: удаление не-MVP разделов и dead code | UI readiness gate |
| S7-E03 | P0 | Удаление глобального фильтра и зависимого кода | UI readiness gate |
| S7-E04 | P0 | Удаление runtime-deploy/images секции и связанного фронтенд-кода | UI readiness gate |
| S7-E05 | P0 | Agents table cleanup + removal of `Скоро` badge | depends on S6 baseline |
| S7-E06 | P0 | Agents import defaults: runtime mode + locale policy + bulk update | depends on S6 baseline |
| S7-E07 | P0 | Worker prompt source selector (`repo`/`db`) в agents settings | depends on API/worker contracts |
| S7-E08 | P1 | Agents UX hardening и массовые операции | after S7-E05..E07 |
| S7-E09 | P0 | Runs UX: убрать run type + гарантировать delete namespace | release-blocking UX |
| S7-E10 | P0 | Runtime deploy task cancel/stop control | release-blocking ops UX |
| S7-E11 | P0 | Исправление поведения `mode:discussion` в label orchestration | stage reliability |
| S7-E12 | P1 | Финальный readiness gate (`qa -> release -> postdeploy -> ops -> doc-audit`) | requires S7-E01..E11 |
| S7-E13 | P0 | Добавить revise-петлю `run:qa:revise` в stage/labels policy | review/revise reliability |
| S7-E14 | P0 | QA policy: проверка новых/изменённых ручек через Kubernetes DNS path | QA acceptance gate |
| S7-E15 | P0 | Agents: обновление prompt templates из repo с version bump | agents lifecycle readiness |
| S7-E16 | P0 | Run status reliability: false-failed для `run:intake:revise` | stage reliability |
| S7-E17 | P0 | Self-improve: доступность и перезапись session snapshot | self-improve reliability |
| S7-E18 | P0 | Documentation governance: issue/PR standard + doc IA + role-template matrix | backlog quality gate |
| S8-E01 | P0 | Control-plane Go refactor: decomposition oversized files | parallel stream (`#225`) |
| S8-E02 | P0 | API-gateway Go refactor: transport cleanup | parallel stream (`#226`) |
| S8-E03 | P0 | Worker Go refactor: service decomposition | parallel stream (`#227`) |
| S8-E04 | P0 | Agent-runner Go refactor: helper normalization | parallel stream (`#228`) |
| S8-E05 | P0 | Shared libs Go refactor: pgx/servicescfg | parallel stream (`#229`) |
| S8-E06 | P1 | Cross-service Go hygiene closure report | after `S8-E01..S8-E05` (`#230`) |

## Quality gates (S7 governance)

| Gate | Что проверяем | Статус |
|---|---|---|
| QG-S7-01 Intake completeness | Проблема, scope, ограничения, AC и backlog streams формализованы на фактах | passed (`#212`) |
| QG-S7-02 Dependency visibility | Зафиксирована актуальная цепочка зависимостей S6: `#199/#201` закрыты, открытый release-блокер — `#216` | passed |
| QG-S7-03 Traceability | Обновлены `issue_map`, `requirements_traceability`, sprint/epic indexes и delivery plan | passed |
| QG-S7-04 Stage continuity | Для Day2 создана follow-up issue `#220` в `run:prd` (без trigger-лейбла) | passed |
| QG-S7-05 Owner comments coverage | Каждое открытое замечание PR #213 классифицировано и сопоставлено с `S7-E*` | passed |
| QG-S7-06 Decomposition parity rule | Перед `run:dev` зафиксировано правило `approved_execution_epics == implementation issues` | passed |
| QG-S7-07 Parallelization readiness | Для Go-refactor stream созданы независимые implementation issues `#225..#230` с явным scope | passed (`#223`) |

## Completion критерии спринта
- [ ] Закрыт открытый P0-блокер S6 (`#216`, `run:release`) и подтверждён переход в `run:postdeploy`.
- [ ] Разделы staff UI с `comingSoon`, попадающие в MVP-сценарии, либо реализованы, либо переведены в явный post-MVP backlog с owner-approved статусом.
- [ ] Stage-контур `run:doc-audit` подтверждён в реальном delivery-цикле с evidence.
- [ ] Выполнен полный e2e проход `run:intake -> ... -> run:ops` для целевого MVP-гейта.
- [ ] Обновлены release/postdeploy/ops артефакты с итоговым go/no-go решением.

## Риски и допущения

| Тип | ID | Описание | Статус |
|---|---|---|---|
| risk | RSK-212-01 | Issue `#216` (`run:release`) остаётся открытой; без release/postdeploy continuity нельзя фиксировать MVP go/no-go | open |
| risk | RSK-212-02 | `run:doc-audit` описан в policy, но без подтверждённого сквозного run-evidence в текущем цикле | open |
| risk | RSK-212-03 | Большой объём UI-scaffold задач может размыть срок MVP closeout без жёсткой P0/P1 декомпозиции | open |
| risk | RSK-223-01 | Рефакторинг Go-кода может дать регрессии без строгого запрета на функциональные изменения | open |
| assumption | ASM-212-01 | Базовые backend-контракты для закрытия P0 уже в `main` (PR `#202` merged) | accepted |
| assumption | ASM-212-02 | Owner подтверждает последовательное закрытие stage-цепочки без параллельных конфликтующих `run:*` | accepted |

## Handover в следующий этап
- Следующий этап: `run:dev` для execution-stream `S8-E01..S8-E06`.
- Для continuity созданы implementation issues `#225..#230` (без trigger-лейблов; лейбл запуска ставит Owner).
- Для `#225..#230` переданы quality-gates и DoD в `docs/delivery/epics/s8/epic-s8-day1-go-refactoring-plan.md`.
