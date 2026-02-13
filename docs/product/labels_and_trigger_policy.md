---
doc_id: LBL-CK8S-0001
type: labels-policy
title: "codex-k8s — Labels and Trigger Policy"
status: draft
owner_role: PM
created_at: 2026-02-11
updated_at: 2026-02-13
related_issues: [1, 19]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
---

# Labels and Trigger Policy

## TL;DR
- Канонический набор лейблов включает классы `run:*`, `state:*`, `need:*` и диагностические labels.
- Trigger/deploy лейблы управляют запуском этапов и требуют апрува Owner при агент-инициации.
- `state:*`, `need:*` и диагностические labels не запускают деплой/исполнение и могут ставиться автоматически по политике.

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
| `run:self-improve` | анализ запусков/комментариев и подготовка улучшений docs/prompts/tools | planned (S3) |
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

## Диагностические labels

| Label | Назначение |
|---|---|
| `run:debug` | не запускает run сам по себе; при наличии на issue в момент `run:dev`/`run:dev:revise` worker сохраняет namespace/job после завершения и пишет `run.namespace.cleanup_skipped` |
| `mode:discussion` | planned: диалоговый pre-run режим для brainstorming под Issue; сам по себе run не запускает |

## Конфигурационные лейблы модели/рассуждений

Лейблы модели (одновременно активен один):
- `[ai-model-gpt-5.3-codex]`
- `[ai-model-gpt-5.2-codex]`
- `[ai-model-gpt-5.1-codex-max]`
- `[ai-model-gpt-5.2]`
- `[ai-model-gpt-5.1-codex-mini]`

Лейблы рассуждений (одновременно активен один):
- `[ai-reasoning-low]`
- `[ai-reasoning-medium]`
- `[ai-reasoning-high]`
- `[ai-reasoning-extra-high]`

Правило при конфликте:
- если на issue выставлено несколько лейблов одной группы (`ai-model` или `ai-reasoning`), run отклоняется как `failed_precondition` с диагностикой в `flow_events`.

## Политика постановки лейблов

### Trigger/deploy (`run:*`)
- Если лейбл инициирует агент, требуется апрув Owner до фактического применения.
- Если лейбл инициирует человек с правами admin/owner, применяется по правам GitHub и политике репозитория.
- Любая операция с `run:*` логируется в `flow_events`.
- Для цикла `run:dev`/`run:dev:revise` перед финальным Owner review обязателен pre-review от системного `reviewer`.
- Для control tools (`secret sync`, `database lifecycle`, `owner feedback`) применяется policy-driven approval matrix по связке `agent_key + run_label + action`.
### Diagnostic labels (`run:debug`)
- `run:debug` не запускает workflow/deploy напрямую.
- Если label присутствует на issue при старте `run:dev`/`run:dev:revise`, worker не удаляет run-namespace автоматически.
- Для такого случая пишется событие `run.namespace.cleanup_skipped` с `cleanup_command` для ручного удаления namespace.


### Service (`state:*`, `need:*`)
- Могут ставиться агентом автоматически в рамках политики проекта.
- Не должны запускать workflow/deploy напрямую.
- Обязательна запись в аудит с actor/correlation.

### Discussion mode (`mode:discussion`, planned)
- Если `mode:discussion` присутствует на Issue в момент `run:dev`/`run:dev:revise`, запуск работает в режиме обсуждения:
  - агент изучает код/окружение и отвечает комментариями под Issue;
  - PR/commit/push не выполняются;
  - сохраняется текущая `codex-cli` session snapshot для продолжения.
- После снятия `mode:discussion` и повторного trigger (`run:dev`/`run:dev:revise`) агент продолжает ту же сессию и выполняет согласованный план реализации.
- Политика вводится как planned-фича следующих спринтов (после стабилизации базового dogfooding контура).

### Model/reasoning labels (`[ai-model-*]`, `[ai-reasoning-*]`)
- Не запускают workflow/deploy напрямую.
- Могут применяться агентом автоматически по policy проекта через MCP.
- Эффективные значения читаются на каждый запуск (`run:dev` и `run:dev:revise`), что позволяет менять модель/reasoning между итерациями ревью.


## Оркестрационный flow для `run:dev` / `run:dev:revise`

- На issue одновременно допускается только один активный trigger label из группы `run:*`.
- `run:dev` используется для первичного запуска цикла разработки и создания PR.
- `run:dev:revise` используется только для итерации по уже существующему PR.
- `run:dev:revise` может запускаться:
  - по label `run:dev:revise` на Issue;
  - по webhook `pull_request_review` с `action=submitted` и `review.state=changes_requested`.
- Для `run:dev:revise` при отсутствии связанного PR run отклоняется с `failed_precondition` и событием `run.revise.pr_not_found`.
- Label transitions после завершения run должны выполняться через MCP (а не вручную в коде агента), чтобы сохранять единый policy/audit контур.
- Для dev/dev:revise transition выполняется так:
  - снять trigger label с Issue;
  - поставить `state:in-review` на PR (не на Issue).
- S2 baseline:
  - pre-review остается обязательным шагом перед финальным Owner review;
  - post-run transitions `run:* -> state:*` фиксируются в Day5/Day6 как отдельные доработки policy и аудита.
- S3 target:
  - активируется полный stage-контур `run:intake..run:ops` + revise/abort/rethink;
  - вводится `run:self-improve` с отдельным post-run transition policy.

## Оркестрационный flow для `run:self-improve`

- На входе: issue/pr с лейблом `run:self-improve` и доступным audit trail (`agent_sessions`, `flow_events`, comments, links).
- Агент собирает источники замечаний (Owner/бот), релевантные логи и артефакты.
- Результат оформляется как change-set (docs/prompts/instructions/tooling) в PR с обязательной трассировкой источников.
- Transition по завершению:
  - снять `run:self-improve` с Issue;
  - поставить `state:in-review` на PR и на Issue (для явного owner decision по улучшениям).

## Требования к GitHub variables (labels-as-vars)

- Все workflow условия сравнения label должны использовать `vars.*`, а не строковые литералы.
- В GitHub Variables хранится **полный каталог** `run:*`, `state:*`, `need:*`:
  - для `run:*`: `RUN_<STAGE>_LABEL` и `RUN_<STAGE>_REVISE_LABEL` (где применимо), плюс `RUN_DEBUG_LABEL`, `RUN_SELF_IMPROVE_LABEL`,
  - для `state:*`: `STATE_*_LABEL`,
  - для `need:*`: `NEED_*_LABEL`.
- Для model/reasoning также хранится каталог vars:
  - `AI_MODEL_GPT_5_3_CODEX_LABEL`, `AI_MODEL_GPT_5_2_CODEX_LABEL`, `AI_MODEL_GPT_5_1_CODEX_MAX_LABEL`, `AI_MODEL_GPT_5_2_LABEL`, `AI_MODEL_GPT_5_1_CODEX_MINI_LABEL`,
  - `AI_REASONING_LOW_LABEL`, `AI_REASONING_MEDIUM_LABEL`, `AI_REASONING_HIGH_LABEL`, `AI_REASONING_EXTRA_HIGH_LABEL`.
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
