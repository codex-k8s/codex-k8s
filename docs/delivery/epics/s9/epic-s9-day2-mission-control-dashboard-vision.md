---
doc_id: EPC-CK8S-S9-D2-MISSION-CONTROL
type: epic
title: "Epic S9 Day 2: Vision для Mission Control Dashboard и console control plane (Issues #335/#337)"
status: in-review
owner_role: PM
created_at: 2026-03-12
updated_at: 2026-03-12
related_issues: [333, 335, 337]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-12-issue-335-vision"
---

# Epic S9 Day 2: Vision для Mission Control Dashboard и console control plane (Issues #335/#337)

## TL;DR
- Для Issue `#335` сформирован vision-пакет: mission, north star, supporting metrics, persona outcomes, MVP/Post-MVP границы и риск-рамка для Mission Control Dashboard.
- Mission Control Dashboard зафиксирован как основной control plane для active work, а не как общий редизайн staff console; GitHub остаётся каноническим provider'ом MVP и местом human review.
- Создана follow-up issue `#337` для stage `run:prd` без trigger-лейбла; PRD должен формализовать user stories, FR/AC/NFR и handover в `run:arch`.

## Priority
- `P0`.

## Vision charter

### Mission statement
Сделать Mission Control Dashboard основным операционным интерфейсом staff console, чтобы owner/product lead, инженер-оператор и discussion-first пользователь могли за минуты понять active set, выбрать безопасное следующее действие и продолжить работу из одного control-plane UX без разрыва traceability, audit policy и GitHub-first review-модели.

### Цели и ожидаемые результаты
1. Превратить staff console из набора разрозненных operational screen'ов в единый active-set control plane для work items, discussion, PR и агентов.
2. Сделать discussion-first funnel first-class сценарием: идея или discussion должны быстро переходить в formalized task и следующий stage без ручной orchestration через GitHub labels UI.
3. Закрепить provider-safe operational model: действия из console обязаны оставаться audit-safe, корректно коррелироваться с outbound sync и не порождать повторные run/comment/task после webhook echo.
4. Зафиксировать UX-рамку, которая остаётся usable на active set от десятков до сотен сущностей за счёт active-set default, list fallback, фильтров и search.

### Пользователи и стейкхолдеры
- Основные пользователи:
  - Owner / product lead, которому нужен быстрый situational awareness и понятный запуск следующего шага.
  - Инженер / operator, которому нужен единый экран для active work items, PR, агентов, ожиданий и блокировок.
  - Discussion-first пользователь, который начинает работу с идеи, обсуждения или голосового ввода и ожидает безопасную formalization path.
- Стейкхолдеры:
  - `services/staff/web-console` как primary UX-контур;
  - `services/external/api-gateway` и `services/internal/control-plane` как будущие владельцы typed commands, projections и sync contract;
  - PM/EM/SA/QA/SRE роли, которые будут использовать dashboard как source of situational awareness и control actions.
- Владелец решения: Owner.

### Продуктовые принципы и ограничения
- Console становится control plane, а не заменой GitHub как места human review и merge decision.
- Active set важнее полного архива: сначала пользователь должен понять, что важно сейчас, и только затем уходить в deep drilldown и historical context.
- Любое действие из dashboard проходит через typed command/reconciliation path и не создаёт обходов label policy, audit trail и webhook-driven orchestration.
- GitHub-first MVP и provider abstraction остаются неподвижными: GitLab допускается только как future-compatible continuation.
- Voice flow рассматривается как candidate stream: core dashboard wave не зависит от его включения до подтверждения ROI, AI policy и operational readiness.
- В рамках `run:vision` разрешены только markdown-изменения.

## Scope boundaries

### MVP scope
- Active-set dashboard shell:
  - новая landing page staff console;
  - summary strip;
  - board/canvas как primary view;
  - list view как fallback для high-volume режима;
  - правая side panel с details/timeline/chat/actions.
- Typed сущности и связи:
  - Work Item / Discussion;
  - Pull Request;
  - Agent;
  - Comment stream;
  - Relation.
- Core действия:
  - создать task/discussion;
  - открыть details/chat;
  - запустить или продолжить stage;
  - формализовать discussion в task;
  - увидеть provider state и текущего исполнителя.
- Realtime snapshot/delta модель поверх существующего WebSocket/realtime baseline.
- Provider sync и reconciliation для titles, labels, comments, PR status и ключевых active-state updates.
- Command/correlation/dedupe модель для пути `UI command -> outbound sync -> webhook echo`.

### Post-MVP / deferred scope
- Произвольный dashboard/layout builder и полнофункциональный custom canvas designer.
- Default-представление полного исторического архива без active-set ограничения.
- GitLab parity в первой волне инициативы.
- Автономные AI-assisted orchestration loops без явного human decision point.
- Расширенные voice/copilot сценарии, которые выходят за рамки guided discussion/task drafting.

### Candidate substream с отдельным gate
- Guided voice intake остаётся отдельной candidate-wave внутри инициативы:
  - его ценность и формат нужно подтвердить на `run:prd`;
  - он не является blocking requirement для core dashboard wave;
  - архитектурный lock-in по speech/realtime transport запрещён до `run:arch` и `run:design`.

## Success metrics

### North Star
| ID | Метрика | Определение | Источник | Целевое значение |
|---|---|---|---|---|
| `NSM-335-01` | Control-plane adoption rate | Доля tracked staff sessions, в которых первое operational action по active work item/discussion/PR выполняется из Mission Control Dashboard | staff analytics events + `flow_events` | `>= 70%` в течение 30 дней после MVP rollout |

### Supporting metrics
| ID | Метрика | Определение/формула | Источник | Цель |
|---|---|---|---|---|
| `PM-335-01` | Situational awareness latency | p75 времени от открытия dashboard до первого entity-focused действия (`open details`, `open chat`, `continue stage`, `open PR context`) | frontend analytics + action audit | `<= 60 секунд` |
| `PM-335-02` | Discussion-to-task lead time | p75 времени от создания discussion до успешного `Formalize as task` для сессий, где discussion действительно формализуется | command audit + work-item relations | `<= 15 минут` |
| `OPS-335-01` | Console-start coverage | Доля core flows, которые можно инициировать из console без ручной навигации по GitHub label UX | product acceptance matrix + release evidence | `>= 80%` core flows к MVP release |
| `REL-335-01` | Reconciliation correctness | Доля console-инициированных команд, завершившихся без дублей task/run/comment после provider webhook echo | sync/reconciliation audit + incidents register | `>= 99.5%` |
| `GOV-335-01` | Traceability freshness | Доля stage-изменений инициативы, отражённых в `delivery_plan`, `issue_map`, sprint/epic docs в день изменения | docs PR evidence | `100%` |

### Guardrails (ранние сигналы)
- `GR-335-01`: если `PM-335-01 > 120 секунд` на pilot-сценариях, scope следующих волн замораживается до упрощения active-set UX и primary actions.
- `GR-335-02`: если `REL-335-01 < 99%`, следующие этапы обязаны приоритизировать reconciliation path выше UX-расширений.
- `GR-335-03`: если voice candidate stream угрожает срокам core wave или не доказывает measurable contribution в PRD, он переводится в post-MVP/deferred backlog.
- `GR-335-04`: если MVP-рамка начинает включать общий redesign console за пределами control-plane UX, stage переводится в `need:input` до повторного owner-решения по scope.

## Risks and Product Assumptions
| Тип | ID | Описание | Митигирующее действие | Статус |
|---|---|---|---|---|
| risk | `RSK-335-01` | Scope может расползтись в общий redesign staff console вместо control-plane инициативы | Жёстко держать initiative вокруг active work, discussion, PR, agents и operational actions | open |
| risk | `RSK-335-02` | Board/canvas может превратиться в визуальный шум на high-volume сценариях | Закрепить active-set default, list fallback и ограничения на information density уже на PRD stage | open |
| risk | `RSK-335-03` | Без явной command/reconciliation модели dashboard будет создавать дубли после webhook echo | Зафиксировать reconciliation correctness как gate-метрику и перенести детализацию в PRD/Arch | open |
| risk | `RSK-335-04` | Voice intake увеличит scope и time-to-value раньше подтверждения ROI | Оставить voice отдельной candidate-wave и не делать его blocking requirement для core MVP | open |
| assumption | `ASM-335-01` | Текущий realtime baseline (`LISTEN/NOTIFY` + WebSocket backplane) можно расширять, а не проектировать заново | Проверить ownership и contracts на `run:arch` | accepted |
| assumption | `ASM-335-02` | GitHub-first provider model остаётся достаточной для MVP Mission Control Dashboard | Сохранить GitLab только как future-compatible continuation | accepted |
| assumption | `ASM-335-03` | Active-set + filters + list fallback достаточно, чтобы удержать usability на 20..400+ сущностях до более поздних UX-волн | Подтвердить PRD edge cases и acceptance evidence для large active set | accepted |

## Readiness criteria для `run:prd`
- [x] Mission и expected value зафиксированы для трёх ключевых persona/outcome-потоков.
- [x] Метрики успеха и guardrails сформулированы как измеримые product/operational сигналы.
- [x] Границы MVP/Post-MVP и candidate substream по voice intake явно отделены друг от друга.
- [x] Неподвижные ограничения инициативы сохранены: GitHub-first MVP, provider-based human review, active-set default, audit-safe command/reconciliation path.
- [x] Создана отдельная issue следующего этапа `run:prd` (`#337`) без trigger-лейбла.

## Acceptance criteria (Issue #335)
- [x] Mission, target outcomes и north star для Mission Control Dashboard сформулированы явно.
- [x] KPI/success metrics и guardrails зафиксированы для situational awareness, discussion-to-task lead time, console-start coverage и reconciliation quality.
- [x] Персоны, MVP/Post-MVP границы, риски и продуктовые принципы описаны для operator и discussion-first workflows.
- [x] Подтверждено, что MVP остаётся GitHub-first, human review выполняется во внешнем provider UI, а dashboard не ломает webhook-driven orchestration и label policy.
- [x] Подготовлен handover в `run:prd` и создана follow-up issue `#337` без trigger-лейбла.

## Handover в следующий этап
- Следующий stage: `run:prd`.
- Follow-up issue: `#337`.
- Trigger-лейбл `run:prd` на issue `#337` ставит Owner.
- Обязательное условие для `#337`: в конце PRD-stage создать issue для stage `run:arch` без trigger-лейбла с ссылками на `#333`, `#335` и `#337`, а также с инструкцией создать issue следующего этапа (`run:design`) после завершения `run:arch`.
- На `run:prd` нельзя потерять следующие product decisions:
  - Mission Control Dashboard остаётся control plane для active work, а не общим redesign всей console;
  - active-set default и list fallback обязательны уже для core wave;
  - GitHub-first provider model и human review во внешнем UI не меняются;
  - command -> provider sync -> webhook echo reconciliation является обязательным продуктовым guardrail, а не только технической деталью;
  - voice intake не блокирует первую волну core dashboard MVP.

## Связанные документы
- `docs/delivery/sprints/s9/sprint_s9_mission_control_dashboard_control_plane.md`
- `docs/delivery/epics/s9/epic_s9.md`
- `docs/delivery/epics/s9/epic-s9-day1-mission-control-dashboard-intake.md`
- `docs/delivery/issue_map.md`
- `docs/delivery/delivery_plan.md`
- `docs/product/requirements_machine_driven.md`
- `docs/product/agents_operating_model.md`
- `docs/product/labels_and_trigger_policy.md`
- `docs/product/stage_process_model.md`
- `docs/research/src_idea-machine_driven_company_requirements.md`
- `services/staff/web-console/README.md`
