---
doc_id: EPC-CK8S-S7-D1
type: epic
title: "Epic S7 Day 1: Intake для закрытия MVP readiness gaps (Issue #212)"
status: in-review
owner_role: PM
created_at: 2026-02-27
updated_at: 2026-02-27
related_issues: [212, 199, 201, 210, 216]
related_prs: [213, 215]
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-27-issue-212-intake"
---

# Epic S7 Day 1: Intake для закрытия MVP readiness gaps (Issue #212)

## TL;DR
- Подтверждён системный разрыв между заявленным MVP stage-flow и фактической готовностью ключевых контуров.
- В staff UI остаются крупные функциональные зоны в статусе `comingSoon` и scaffold/TODO (`services/staff/web-console/src/app/navigation.ts`, страницы `governance/*`, `admin/*`, `configuration/*`).
- На момент первичного intake блокерами были `#199` (dev) и `#201` (qa); по факту на 2026-02-27 они закрыты (`PR #202` merged), а текущая открытая зависимость S6 — Issue `#216` (`run:release`).
- Для `run:doc-audit` есть policy/seed/config-база, но нет подтверждённого сквозного evidence успешного stage-run в текущем delivery-цикле.

## Problem Statement
### As-Is
- Навигация staff UI содержит `comingSoon` для governance/admin/platform разделов (Audit Log, Labels & Stages, Cluster resources, Agents, Docs/Knowledge, MCP Tools).
- В коде UI зафиксированы TODO на подключение реальных backend-контуров (например, `AuditLogPage.vue`, `LabelsStagesPage.vue`, `McpToolsPage.vue`, `DocsKnowledgePage.vue`, `AgentDetailsPage.vue`, `NamespacesPage.vue`, `PodsPage.vue` и др.).
- Реализация S6 Day7 уже влита в `main` (`PR #202` merged), QA-этап `#201` закрыт, и цепочка продолжена в Issue `#216` (`run:release`).
- Запрос PMO (`#210`) добавляет обязательные требования к качеству постановки задач: user-story формулировка и edge cases.

### To-Be
- MVP readiness подтверждён не декларативно, а по evidence: функциональные потоки работают в `main`, stage-цепочка закрыта до `ops/doc-audit`, а критичные UI-разделы либо реализованы, либо явно выведены в post-MVP backlog.
- Каждая крупная задача оформляется в формате user story + acceptance criteria + edge cases.
- Delivery-governance имеет прозрачный backlog с P0/P1/P2 приоритетами и owners по каждому потоку.

## Brief
- Бизнес-ценность: получить предсказуемый MVP, который можно демонстрировать и использовать без «скоро»-зон в критическом workflow.
- Операционная ценность: убрать разрыв между product-policy и фактическим stage execution, особенно в завершающих этапах `qa/release/postdeploy/ops/doc-audit`.
- Delivery-ценность: превратить разрозненные замечания в управляемый, трассируемый execution backlog.

## MVP Scope
### In scope
- MVP-gap backlog с декомпозицией на 18 candidate execution-эпиков (P0/P1/P2).
- Закрытие текущей зависимости S6 (`#216`, `run:release`) как обязательный вход в MVP closeout.
- Формализация критериев готовности для UI readiness, stage reliability и governance quality.
- Включение требования PMO: user story + edge cases для будущих implementation issues.

### Out of scope
- Добавление новых продуктовых направлений вне MVP.
- Пересмотр архитектурных границ сервисов и label taxonomy.
- Ослабление security/policy ограничений ради ускорения разработки.

## Constraints
- Сохраняются Kubernetes-only, webhook-driven и stage/label policy ограничения.
- Для `run:intake` допустимы только markdown-изменения.
- Все переходы по stage выполняются через review gate и owner decision.
- Работа с PR/issue через `gh`, label transitions по завершению run через MCP.

## Acceptance Criteria (Intake stage)
- [x] Подтверждены фактические MVP-gaps по UI scaffold и stage continuity на основе кода/issue/PR состояния.
- [x] Зафиксирован приоритетный backlog потоков P0/P1/P2 с owner-role alignment.
- [x] Зафиксирована актуальная dependency-цепочка S6 (`#199`/`#201` закрыты, открытый блокер — `#216`) для MVP closeout.
- [x] Зафиксированы продуктовые риски и допущения с привязкой к execution-плану.
- [x] Зафиксирован handover в `run:vision` с требованиями к KPI и edge-case coverage.

## Декомпозиция глобальных потоков

| Stream | Priority | Цель | Основные deliverables |
|---|---|---|---|
| S7-W1: S6 closure dependency | P0 | Закрыть открытый release-этап `#216` и подтвердить готовность к полной цепочке `release -> postdeploy -> ops` | release evidence + issue continuity `postdeploy -> ops` |
| S7-W2: UI readiness | P0 | Убрать `comingSoon` из MVP-критичных разделов или формально вывести в post-MVP с owner approval | реализация/декомпозиция governance+platform+admin страниц с typed API |
| S7-W3: Stage reliability | P0 | Подтвердить работоспособность завершающих stage-циклов, включая `run:doc-audit` | stage-run evidence, audit trail, обновлённые runbooks/checklists |
| S7-W4: Backlog quality governance | P1 | Стандартизировать постановку задач по user story + edge cases (Issue `#210`) | обновлённые issue templates/process rules + AC/edge-case checklist |
| S7-W5: Final MVP gate | P1 | Провести финальный e2e regression и зафиксировать go/no-go | consolidated QA/release/postdeploy/ops/doc-audit пакет |
| S7-W6: Post-MVP spillover control | P2 | Отделить то, что не входит в MVP, чтобы не блокировать релиз | owner-approved post-MVP backlog с traceability |
| S7-W7: Documentation governance | P0 | Унифицировать issue/PR и структуру project-docs по ролям/шаблонам | единый стандарт заголовков/body + role template matrix + обновлённый process doc |
| S7-W8: Run reliability hardening | P0 | Закрыть reliability-разрывы review/revise/self-improve контуров | QA revise-label coverage + run status consistency + self-improve session persistence |

## Матрица owner-замечаний PR #213

| Comment ID | Формулировка | Priority group | Статус | Трассировка |
|---|---|---|---|---|
| PRC-01 | Нужен rebase | behavior/data | `fix_required` | `S7-E01` |
| PRC-02 | Убрать runtime-deploy/images и неиспользуемый код | behavior/data | `fix_required` | `S7-E04` |
| PRC-03 | Блок замечаний по Agents (badge, runtime mode, locale, prompt-source, таблица) | behavior/data | `fix_required` | `S7-E05`, `S7-E06`, `S7-E07`, `S7-E08` |
| PRC-04 | Удалить глобальный фильтр и связанный dead code | quality/style | `fix_required` | `S7-E03` |
| PRC-05 | Удалить не-MVP разделы в левом меню и связанный код | behavior/data | `fix_required` | `S7-E02`, `S7-E04` |
| PRC-06 | В Runs убрать тип запуска, в деталях всегда давать delete namespace | behavior/data | `fix_required` | `S7-E09` |
| PRC-07 | В runtime deploy details добавить кнопку cancel/stop | behavior/data | `fix_required` | `S7-E10` |
| PRC-08 | Не работает `mode:discussion` | behavior/data | `fix_required` | `S7-E11` |
| PRC-09 | Добавить revise-петлю для `run:qa` | behavior/data | `fix_required` | `S7-E13` |
| PRC-10 | QA должен проверять новые/изменённые ручки через DNS Kubernetes (без упора в OAuth UI-flow) | behavior/data | `fix_required` | `S7-E14` |
| PRC-11 | В Agents нужна кнопка обновления prompt templates из репозитория с созданием новых версий | behavior/data | `fix_required` | `S7-E15` |
| PRC-12 | `run:intake:revise` (run `398275e1-161f-4bfa-86ac-baf27004dcaa`) отработал по факту, но отмечен как failed | behavior/data | `fix_required` | `S7-E16` |
| PRC-13 | `run:self-improve` не извлёк целевую сессию агента; нужна верификация сохранения/перезаписи session snapshot | behavior/data | `fix_required` | `S7-E17` |
| PRC-14 | Привести заголовки issue/PR к единому стилю по роли и типу задачи | quality/style | `fix_required` | `S7-E18` |
| PRC-15 | Систематизировать документацию по типам документов и их месту | quality/style | `fix_required` | `S7-E18` |
| PRC-16 | Явно закрепить, какие роли какие документы и шаблоны готовят | quality/style | `fix_required` | `S7-E18` |

## Candidate execution backlog (18 эпиков)

| Epic ID | Priority | Краткая задача | Основной вход |
|---|---|---|---|
| S7-E01 | P0 | Rebase/mainline hygiene и merge-conflict policy | PRC-01 |
| S7-E02 | P0 | Удаление не-MVP разделов из sidebar/routes + dead code cleanup | PRC-05 |
| S7-E03 | P0 | Удаление глобального фильтра + удаление зависимого кода | PRC-04 |
| S7-E04 | P0 | Удаление runtime-deploy/images контуров | PRC-02, PRC-05 |
| S7-E05 | P0 | Agents UI cleanup: убрать `Скоро`, переразметить таблицу | PRC-03 |
| S7-E06 | P0 | Agents import defaults: runtime mode policy + locale policy/bulk update | PRC-03 |
| S7-E07 | P0 | Worker prompt source selector (`repo`/`db`) + contract alignment | PRC-03 |
| S7-E08 | P1 | Agents UX hardening и пакет массовых операций | PRC-03 |
| S7-E09 | P0 | Runs UX cleanup: удалить run-type, добавить deterministic namespace delete action | PRC-06 |
| S7-E10 | P0 | Runtime deploy cancel/stop action + safety guardrails | PRC-07 |
| S7-E11 | P0 | `mode:discussion` trigger/review-flow remediation | PRC-08 |
| S7-E12 | P1 | Final readiness gate: consolidated evidence + go/no-go | PRC-01..PRC-08 |
| S7-E13 | P0 | Label taxonomy alignment: добавить `run:qa:revise` в stage/labels policy и review automation | PRC-09 |
| S7-E14 | P0 | QA execution policy: проверка новых/изменённых ручек через K8s DNS path + evidence requirements | PRC-10 |
| S7-E15 | P0 | Agents prompt lifecycle UX: обновление prompt templates из repo с версионированием | PRC-11 |
| S7-E16 | P0 | Run status consistency: устранить false-failed для `run:intake:revise` при фактически успешном completion | PRC-12 |
| S7-E17 | P0 | Self-improve diagnostics hardening: гарантировать доступность и перезапись session snapshot | PRC-13 |
| S7-E18 | P0 | Documentation governance hardening: единый стандарт issue/PR + doc IA + role-template matrix | PRC-14, PRC-15, PRC-16 |

## Предварительный candidate set для stage `run:vision`
1. Перенести `S7-E01..S7-E18` в отдельные issue-кандидаты с user-story формулировкой.
2. Для каждого epic-кандидата дополнить блок edge cases и метрики готовности.
3. Зафиксировать dependency graph (`must-have P0` перед стартом P1).

## Risks and Product Assumptions
### Risks
- `RSK-212-01`: Issue `#216` (`run:release`) остаётся открытой; без закрытия release/postdeploy/ops цепочки MVP-gate не может считаться завершённым.
- `RSK-212-02`: без подтверждения `run:doc-audit` возможен ложный сигнал готовности MVP.
- `RSK-212-03`: объём UI-scaffold задач может превысить окно спринта без жёсткой приоритизации.
- `RSK-212-04`: отсутствие стандарта edge cases в issue продолжит снижать качество QA-приёмки.

### Assumptions
- `ASM-212-01`: технический фундамент P0 по S6 закрыт в `main` (PR `#202` merged), дальнейший риск смещён в release/postdeploy continuity.
- `ASM-212-02`: Owner подтвердит staged-подход без параллельных конфликтующих `run:*`.
- `ASM-212-03`: post-MVP потоки можно отделить без влияния на MVP-go/no-go.

## Stage Handover Instructions
- Следующий этап: `run:vision`.
- Обязательный артефакт vision-этапа: зафиксировать KPI и measurable readiness criteria по backlog-набору `S7-E01..S7-E18`.
- В конце vision-этапа создать follow-up issue для `run:prd` без trigger-лейбла с обязательным блоком edge cases.
