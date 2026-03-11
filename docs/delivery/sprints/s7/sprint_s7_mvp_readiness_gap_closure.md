---
doc_id: SPR-CK8S-0007
type: sprint-plan
title: "Sprint S7: MVP readiness gap closure (Issue #212)"
status: in-progress
owner_role: PM
created_at: 2026-02-27
updated_at: 2026-03-11
related_issues: [212, 218, 220, 222, 238, 241, 243, 244, 245, 246, 247, 248, 249, 250, 251, 252, 253, 254, 255, 256, 257, 258, 259, 260, 274, 199, 201, 210, 216]
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
| Day 3 | PRD: FR/AC/NFR и sequencing для gap-closure streams | P0 | `docs/delivery/epics/s7/epic-s7-day3-mvp-readiness-prd.md` + `docs/delivery/epics/s7/prd-s7-day3-mvp-readiness-gap-closure.md` | in-review (`#220`) |
| Day 4 | Architecture: границы и ownership по stream'ам | P0 | `docs/delivery/epics/s7/epic-s7-day4-mvp-readiness-arch.md` | in-review (`#222`) |
| Day 5 | Design: execution-ready contracts/data/migrations package | P0 | `docs/delivery/epics/s7/epic-s7-day5-mvp-readiness-design.md` (`#238`) | in-review (`#238`) |
| Day 6 | Plan: execution package и quality gates | P0 | `docs/delivery/epics/s7/epic-s7-day6-mvp-readiness-plan.md` (`#241`) | in-review (`#241`) |
| Day 7+ | Dev/QA/Release/Postdeploy/Ops/Doc-Audit | P0/P1 | implementation issues `#243..#260`, `#274` (`run:dev`) | in-progress (`#243` и `#244` completed + owner-approved; `#245`, `#246`, `#247/#248/#249`, `#251`, `#252`, `#253`, `#255`, `#256`, `#258` и `#274` реализованы в execution streams; `#250/#257` закрываются doc-actualization pass как уже поглощённые cleanup-потоками; remaining standalone backlog: `#254`, `#259..#260`) |

## Candidate execution-эпики (`S7-E01..S7-E18`)

| Epic | Priority | Scope | Блокер/зависимость |
|---|---|---|---|
| S7-E01 | P0 | Rebase/mainline hygiene для PR revise-итераций | done (owner-approved, `#243`) |
| S7-E02 | P0 | Sidebar cleanup: удаление не-MVP разделов (включая Agents, Configs/Secrets, Registry images, Running jobs) и dead code | done (owner-approved, `#244`) |
| S7-E03 | P0 | Удаление глобального фильтра и зависимого кода | in-review (`#245`) |
| S7-E04 | P0 | Удаление runtime-deploy/images секции; stale deeplinks закрываются общим fallback route | in-review (`#246`) |
| S7-E05 | P0 | Финальный cleanup residual references после удаления `Agents` из MVP UI | closes via `#247` |
| S7-E06 | P0 | Зафиксировать фактический MVP de-scope: без runtime mode/locale settings UI/API | closes via `#248` |
| S7-E07 | P0 | Зафиксировать фактический repo-only prompt contract и удалить residual stale traces | closes via `#249` |
| S7-E08 | P1 | Agents UX de-scope hardening: standalone issue больше не нужна после `S7-E05..S7-E07` + `S7-E19`; `#250` закрывается doc-actualization pass | absorbed by S7-E05..E07 + S7-E19 |
| S7-E09 | P0 | Runs UX: убрать run type + гарантировать delete namespace | passed (in-review `#251`) |
| S7-E10 | P0 | Runtime deploy task cancel/stop control | passed (in-review `#252`) |
| S7-E11 | P0 | Исправление поведения `mode:discussion` в label orchestration | implemented |
| S7-E12 | P1 | Финальный readiness gate (`qa -> release -> postdeploy -> ops -> doc-audit`) | requires S7-E01..E11 |
| S7-E13 | P0 | Добавить недостающие revise-петли `run:doc-audit|qa|release|postdeploy|ops|self-improve:revise` в stage/labels policy | passed (in-review `#255`) |
| S7-E14 | P0 | QA policy: проверка новых/изменённых ручек через Kubernetes DNS path | passed (in-review `#256`) |
| S7-E15 | P0 | Prompt templates MVP policy: standalone issue больше не нужна после `S7-E07` + `S7-E19`; `#257` закрывается doc-actualization pass | absorbed by S7-E07 + S7-E19 |
| S7-E16 | P0 | Run status reliability: false-failed для `run:intake:revise` | passed (in-review `#258`) |
| S7-E17 | P0 | Self-improve: доступность и перезапись session snapshot | self-improve reliability |
| S7-E18 | P0 | Documentation governance: issue/PR standard + doc IA + role-template matrix | backlog quality gate |
| S7-E19 | P1 | Backend cleanup: удалить non-MVP контуры Agents/Configs/Secrets + registry images + running jobs | after S7-E02 |

## Quality gates (S7 governance)

| Gate | Что проверяем | Статус |
|---|---|---|
| QG-S7-01 Intake completeness | Проблема, scope, ограничения, AC и backlog streams формализованы на фактах | passed (`#212`) |
| QG-S7-02 Dependency visibility | Зафиксирована актуальная цепочка зависимостей S6: `#199/#201` закрыты, открытый release-блокер — `#216` | passed |
| QG-S7-03 Traceability | Обновлены `issue_map`, `requirements_traceability`, sprint/epic indexes и delivery plan | passed |
| QG-S7-04 Stage continuity | Для Day2 создана follow-up issue `#220` в `run:prd` (без trigger-лейбла) | passed |
| QG-S7-05 Owner comments coverage | Каждое открытое замечание PR #213 классифицировано и сопоставлено с `S7-E*` | passed |
| QG-S7-06 Decomposition parity rule | Перед `run:dev` зафиксировано правило `approved_execution_epics == implementation issues` | passed |
| QG-S7-07 PRD completion | Для Day3 выпущен PRD-пакет (`epic + prd`) и создана follow-up issue `#222` в `run:arch` | passed |
| QG-S7-08 Architecture completion | Для Day4 выпущен architecture-пакет (ownership matrix + C4 overlays + ADR-0010) и создана follow-up issue `#238` в `run:design` | passed |
| QG-S7-09 Design completion | Для Day5 выпущен design package (`design_doc`, `api_contract`, `data_model`, `migrations_policy`) и создана follow-up issue `#241` в `run:plan` | passed |
| QG-S7-10 Plan completion | Для Day6 выпущен execution package, создано 18 implementation issues `#243..#260`, parity-check `18/18` подтверждён | passed |
| QG-S7-11 Foundation stream S7-E01 | Для issue `#243` зафиксирован единый rebase/mainline process и обязательный PR checklist для revise-итераций | passed (owner-approved) |
| QG-S7-12 UI stream S7-E02 | Для issue `#244` удалены non-MVP sidebar/routes (включая Agents, Configs/Secrets, Registry images, Running jobs) и выполнен навигационный smoke-check без broken transitions | passed (owner-approved) |
| QG-S7-13 UI stream S7-E03 | Для issue `#245` удалены global filter UI/state зависимости и подтверждён list-load без env-фильтра в `runtime-deploy/tasks` | passed (in-review `#245`) |
| QG-S7-14 UI stream S7-E04 | Для issue `#246` подтверждено, что stale `/runtime-deploy/images*` закрывается существующим catch-all route `/:pathMatch(.*)* -> projects`, dedicated redirect не добавляется и traceability синхронизирована | passed (in-review `#246`) |
| QG-S7-15 UI stream S7-E09 | Для issue `#251` колонка `run type` удалена, delete namespace path больше не зависит от `job_exists`, negative-case path подтверждён идемпотентным typed endpoint | passed (in-review `#251`) |
| QG-S7-16 Deploy control stream S7-E10 | Для issue `#252` добавлены typed cancel/stop actions, persisted audit/control state, lease-aware/idempotent guardrails и staff UI controls с error/success feedback | passed (in-review `#252`) |
| QG-S7-17 Multi-stage revise stream S7-E13 | Для issue `#255` недостающие revise-петли `run:doc-audit|qa|release|postdeploy|ops|self-improve:revise` доведены до рабочего revise-loop: добавлены typed trigger kinds, resolver path для PR review `changes_requested`, agent routing, next-step/env-label mapping и runner policy | passed (in-review `#255`) |
| QG-S7-18 QA DNS policy stream S7-E14 | Для issue `#256` QA runbook и QA templates синхронно требуют Kubernetes service DNS path и DNS evidence bundle для новых/изменённых HTTP-ручек | passed (in-review `#256`) |

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
| assumption | ASM-212-01 | Базовые backend-контракты для закрытия P0 уже в `main` (PR `#202` merged) | accepted |
| assumption | ASM-212-02 | Owner подтверждает последовательное закрытие stage-цепочки без параллельных конфликтующих `run:*` | accepted |

## Handover в следующий этап
- Следующий этап: `run:dev`.
- По owner-уточнению вместо одной stage-issue подготовлен execution-пакет из 18 implementation issues:
  - Wave 1: `#243`, `#253`, `#255`;
  - Wave 2: `#244`, `#245`, `#246`, `#247`;
  - Wave 3 (исторический plan baseline): `#248`, `#249`, `#250`, `#257`, `#259`;
- Wave 4: `#251`, `#252`, `#258`;
- Wave 5: `#256`, `#260`, `#254`.
- Дополнительно (post-plan): `#274` (`S7-E19`, backend cleanup Agents/Configs/Secrets).
- После combined cleanup `#247/#248/#249/#274` standalone issues `#250` и `#257` закрываются без отдельного `run:dev`; после реализации `#252`, `#253`, `#255`, `#256` и `#258` фактический remaining backlog нормализован как `#254`, `#259..#260`.
- Trigger-лейбл `run:dev` на implementation issues ставит Owner, сохраняя wave-sequencing.
- Обязательные артефакты handover:
  - `docs/delivery/epics/s7/epic-s7-day6-mvp-readiness-plan.md`;
  - синхронизированные `issue_map` и `requirements_traceability`;
  - parity evidence: `approved_execution_epics_count == created_run_dev_issues_count` (`18 == 18`).

## Актуализация фактического состояния по `#247/#248/#249`
- После cleanup `#244` и backend cleanup `#274` страницы и staff API для `Agents`/`Prompt templates` больше не входят в MVP.
- Issue `#247/#248/#249` закрываются не через возврат удалённых экранов, а через:
  - cleanup residual dead code/reference layers;
  - фиксацию реального MVP контракта в source-of-truth документах;
  - явный unit-test для `repo_seed + default locale` поведения worker-а.
- `S7-E08` / Issue `#250` больше не требуют отдельного `run:dev`: после удаления `Agents` UI/API в `#244` и backend cleanup `#274` в MVP не осталось отдельного Agents UX контура, который нужно было бы дополнительно harden'ить.
- `S7-E15` / Issue `#257` больше не требуют отдельного `run:dev`: repo-only prompt policy уже зафиксирован combined closure pass `#247/#248/#249`, а UI/API контуры refresh/versioning отсутствуют после cleanup `#274`.
- Remaining Sprint S7 standalone execution backlog после этой актуализации: `#254`, `#259..#260`.
