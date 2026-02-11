---
doc_id: LBL-CK8S-0001
type: labels-policy
title: "codex-k8s — Labels and Trigger Policy"
status: draft
owner_role: PM
created_at: 2026-02-11
updated_at: 2026-02-11
related_issues: [1]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
---

# Labels and Trigger Policy

## TL;DR
- Канонический набор лейблов включает классы `run:*`, `state:*`, `need:*`.
- Trigger/deploy лейблы управляют запуском этапов и требуют апрува Owner при агент-инициации.
- `state:*` и `need:*` не запускают деплой/исполнение и могут ставиться автоматически по политике.

## Source of truth
- `docs/product/stage_process_model.md`
- `docs/product/agents_operating_model.md`
- `docs/research/src_idea-machine_driven_company_requirements.md`

## Класс `run:*` (trigger/deploy)

| Label | Назначение | Статус в платформе |
|---|---|---|
| `run:intake` | старт проработки идеи/инициативы | planned |
| `run:intake:revise` | ревизия intake артефактов | planned |
| `run:vision` | формирование charter/vision/metrics | planned |
| `run:vision:revise` | ревизия vision | planned |
| `run:prd` | формирование PRD | planned |
| `run:prd:revise` | ревизия PRD | planned |
| `run:arch` | формирование C4/ADR/NFR | planned |
| `run:arch:revise` | ревизия архитектуры | planned |
| `run:design` | detailed design + API/data model | planned |
| `run:design:revise` | ревизия design | planned |
| `run:plan` | delivery plan + epics/stories | planned |
| `run:plan:revise` | ревизия плана | planned |
| `run:dev` | разработка и создание PR | active (S2) |
| `run:dev:revise` | доработка существующего PR | active (S2) |
| `run:doc-audit` | аудит код↔доки↔чек-листы | planned |
| `run:qa` | тест-артефакты и прогоны | planned |
| `run:release` | релиз и release artifacts | planned |
| `run:postdeploy` | post-deploy review / postmortem | planned |
| `run:ops` | эксплуатационные улучшения | planned |
| `run:abort` | остановка/cleanup текущей инициативы | planned |
| `run:rethink` | переоткрытие этапа и смена траектории | planned |

## Класс `state:*` (служебные статусы)

| Label | Назначение |
|---|---|
| `state:blocked` | выполнение заблокировано |
| `state:in-review` | артефакт ожидает ревью Owner |
| `state:approved` | этап/артефакт принят |
| `state:superseded` | артефакт заменён более новым |
| `state:abandoned` | работа остановлена/отменена |

## Класс `need:*` (запросы на участие/вход)

| Label | Назначение |
|---|---|
| `need:input` | нужен ответ/решение Owner |
| `need:pm` | нужно продуктовое уточнение |
| `need:sa` | нужно архитектурное уточнение |
| `need:qa` | нужен QA-вход или тест-дизайн |
| `need:sre` | нужно участие SRE/OPS |

## Политика постановки лейблов

### Trigger/deploy (`run:*`)
- Если лейбл инициирует агент, требуется апрув Owner до фактического применения.
- Если лейбл инициирует человек с правами admin/owner, применяется по правам GitHub и политике репозитория.
- Любая операция с `run:*` логируется в `flow_events`.
- Для цикла `run:dev`/`run:dev:revise` перед финальным Owner review обязателен pre-review от системного `reviewer`.

### Service (`state:*`, `need:*`)
- Могут ставиться агентом автоматически в рамках политики проекта.
- Не должны запускать workflow/deploy напрямую.
- Обязательна запись в аудит с actor/correlation.

## Требования к GitHub variables (labels-as-vars)

- Все workflow условия сравнения label должны использовать `vars.*`, а не строковые литералы.
- В GitHub Variables хранится **полный каталог** `run:*`, `state:*`, `need:*`:
  - для `run:*`: `RUN_<STAGE>_LABEL` и `RUN_<STAGE>_REVISE_LABEL` (где применимо),
  - для `state:*`: `STATE_*_LABEL`,
  - для `need:*`: `NEED_*_LABEL`.
- Для planned `run:*` лейблов vars заводятся заранее, даже если этап ещё не активирован.
- Bootstrap синхронизация каталога выполняется скриптом `bootstrap/remote/45_configure_github_repo_ci.sh`.

## Аудит и наблюдаемость

- Для каждого изменения label фиксируются:
  - `correlation_id`,
  - `requested_by`/`applied_by`,
  - `approval_state` (если применимо),
  - `source` (human/mcp/system),
  - `timestamp`.
- Источник аудита: `flow_events` + связка с `agent_sessions`.
