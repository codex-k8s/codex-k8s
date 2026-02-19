---
doc_id: LBL-CK8S-0001
type: labels-policy
title: "codex-k8s — Labels and Trigger Policy"
status: active
owner_role: PM
created_at: 2026-02-11
updated_at: 2026-02-14
related_issues: [1, 19]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-19-full-docset"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-19
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
| `run:intake` | старт проработки идеи/инициативы | active (S3 Day1 trigger path) |
| `run:intake:revise` | ревизия intake артефактов | active (S3 Day1 trigger path) |
| `run:vision` | формирование charter/vision/metrics | active (S3 Day1 trigger path) |
| `run:vision:revise` | ревизия vision | active (S3 Day1 trigger path) |
| `run:prd` | формирование PRD | active (S3 Day1 trigger path) |
| `run:prd:revise` | ревизия PRD | active (S3 Day1 trigger path) |
| `run:arch` | формирование C4/ADR/NFR | active (S3 Day1 trigger path) |
| `run:arch:revise` | ревизия архитектуры | active (S3 Day1 trigger path) |
| `run:design` | detailed design + API/data model | active (S3 Day1 trigger path) |
| `run:design:revise` | ревизия design | active (S3 Day1 trigger path) |
| `run:plan` | delivery plan + epics/stories | active (S3 Day1 trigger path) |
| `run:plan:revise` | ревизия плана | active (S3 Day1 trigger path) |
| `run:dev` | разработка и создание PR | active (S2) |
| `run:dev:revise` | доработка существующего PR | active (S2) |
| `run:doc-audit` | аудит код↔доки↔чек-листы | active (S3 Day1 trigger path) |
| `run:qa` | тест-артефакты и прогоны | active (S3 Day1 trigger path) |
| `run:release` | релиз и release artifacts | active (S3 Day1 trigger path) |
| `run:postdeploy` | post-deploy review / postmortem | active (S3 Day1 trigger path) |
| `run:ops` | эксплуатационные улучшения | active (S3 Day1 trigger path) |
| `run:self-improve` | анализ запусков/комментариев и подготовка улучшений docs/prompts/tools | active (S3 Day1 trigger path; deep logic S3 Day6+) |
| `run:rethink` | переоткрытие этапа и смена траектории | active (S3 Day1 trigger path) |

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
| `need:em` | нужен review/решение Engineering Manager |
| `need:km` | нужен review по документации/трассировке |
| `need:reviewer` | нужен предварительный технический pre-review |

## Диагностические labels

| Label | Назначение |
|---|---|
| `run:debug` | не запускает run сам по себе; при наличии на issue в момент `run:dev`/`run:dev:revise` worker сохраняет namespace/job после завершения и пишет `run.namespace.cleanup_skipped` |
| `mode:discussion` | planned: диалоговый pre-run режим для brainstorming под Issue; сам по себе run не запускает |

## Конфигурационные лейблы модели/рассуждений

Лейблы модели (одновременно активен один):
- `[ai-model-gpt-5.3-codex]`
- `[ai-model-gpt-5.3-codex-spark]`
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
- Для всех `run:*` обязателен review gate перед финальным Owner review:
  - pre-review от системного `reviewer` для технических артефактов;
  - role-specific review через `need:*` labels для профильных артефактов.
- Для control tools (`secret sync`, `database lifecycle`, `owner feedback`) применяется policy-driven approval matrix по связке `agent_key + run_label + action`.
### Diagnostic labels (`run:debug`)
- `run:debug` не запускает workflow/deploy напрямую.
- Если label присутствует на issue при старте `run:dev`/`run:dev:revise`, worker не удаляет run-namespace автоматически.
- Для такого случая пишется событие `run.namespace.cleanup_skipped` с `cleanup_command` для ручного удаления namespace.


### Service (`state:*`, `need:*`)
- Могут ставиться агентом автоматически в рамках политики проекта.
- Не должны запускать workflow/deploy напрямую.
- Обязательна запись в аудит с actor/correlation.
- Для role-specific ревью артефактов используются `need:*` labels (вместе с `state:in-review`).
- Для всех `run:*` при наличии артефактов для проверки Owner ставится `state:in-review`:
  - на PR и на Issue, если run сформировал PR;
  - только на Issue, если run не формирует PR.

### Discussion mode (`mode:discussion`, planned)
- Если `mode:discussion` присутствует на Issue в момент `run:dev`/`run:dev:revise`, запуск работает в режиме обсуждения:
  - агент изучает код/окружение и отвечает комментариями под Issue;
  - PR/commit/push не выполняются;
  - вместо job поднимается отдельный `discussion` pod с `codex-cli` session snapshot.
- `discussion` pod живёт до первого из событий:
  - idle timeout `8h`;
  - закрытие Issue;
  - постановка на Issue любого trigger `run:*`.
- На webhook `issue_comment`:
  - если комментарий оставил не агент, раннер продолжает текущую discussion-сессию и публикует ответ под Issue;
  - служебные комментарии платформы и комментарии агента не считаются пользовательским входом.
- После снятия `mode:discussion` и повторного trigger (`run:dev`/`run:dev:revise`) агент продолжает ту же сессию и выполняет согласованный план реализации.
- Политика вводится как planned-фича следующих спринтов (после стабилизации базового dogfooding контура).

### Model/reasoning labels (`[ai-model-*]`, `[ai-reasoning-*]`)
- Не запускают workflow/deploy напрямую.
- Могут применяться агентом автоматически по policy проекта через MCP.
- Эффективные значения читаются на каждый запуск (`run:dev` и `run:dev:revise`), что позволяет менять модель/reasoning между итерациями ревью.


## Оркестрационный flow для `run:dev` / `run:dev:revise`

- Для каждого запуска MCP выдает только run-scoped список ручек:
  - для `run:dev`/`run:dev:revise` baseline = только label-ручки;
  - недоступные ручки скрываются из `tools/list` и блокируются на `tools/call`.
- На issue одновременно допускается только один активный trigger label из группы `run:*`.
- `run:dev` используется для первичного запуска цикла разработки и создания PR.
- `run:dev:revise` используется только для итерации по уже существующему PR.
- `run:dev:revise` может запускаться:
  - по label `run:dev:revise` на Issue;
  - по webhook `pull_request_review` с `action=submitted` и `review.state=changes_requested`,
    если на PR стоит ровно один stage label из поддержанных пар:
    `run:intake|run:intake:revise`, `run:vision|run:vision:revise`,
    `run:prd|run:prd:revise`, `run:arch|run:arch:revise`,
    `run:design|run:design:revise`, `run:plan|run:plan:revise`,
    `run:dev|run:dev:revise`.
    В этом случае платформа запускает соответствующий `run:<stage>:revise`.
    Если stage labels нет или их несколько, ран не создается.
- Для `run:dev:revise` при отсутствии связанного PR run отклоняется с `failed_precondition` и событием `run.revise.pr_not_found`.
- Label transitions после завершения run должны выполняться через MCP (а не вручную в коде агента), чтобы сохранять единый policy/audit контур.
- Для dev/dev:revise transition выполняется так:
  - снять trigger label с Issue;
  - поставить `state:in-review` на PR и на Issue.
- S2 baseline:
  - pre-review остается обязательным шагом перед финальным Owner review;
  - post-run transitions `run:* -> state:*` фиксируются в Day5/Day6 как отдельные доработки policy и аудита.
- S3 Day1 факт:
- активирован полный stage-контур `run:intake..run:ops` + revise/rethink;
  - активирован trigger path для `run:self-improve` (расширенная бизнес-логика дорабатывается по S3 Day6+).

## Оркестрационный flow для `run:self-improve`

- На входе: issue/pr с лейблом `run:self-improve` и доступным audit trail (`agent_sessions`, `flow_events`, comments, links).
- Это основной и единственный use-case self-improve: анализ качества предыдущих запусков и выпуск PR с улучшениями платформы.
- В run-scoped MCP-каталоге для `run:self-improve` доступны label-ручки и diagnostic self-improve ручки; остальные ручки скрыты и недоступны.
- Агент обязан работать через связку MCP+GitHub CLI:
  - MCP `self_improve_runs_list`: список запусков с пагинацией (по 50, newest-first);
  - MCP `self_improve_run_lookup`: поиск запусков по `issue_number`/`pull_request_number`;
  - MCP `self_improve_session_get`: извлечение `codex-cli` session JSON выбранного run;
  - `gh`: чтение Issue/PR, комментариев и review-диагностики.
- Для анализа session JSON используется временный каталог `/tmp/codex-sessions/<run-id>` (создается до записи).
- Результат оформляется как change-set (docs/prompts/instructions/tooling/image/scripts) в PR с обязательной трассировкой источников.
- Transition по завершению:
  - снять `run:self-improve` с Issue;
  - поставить `state:in-review` на PR и на Issue (для явного owner decision по улучшениям).

## Требования к GitHub variables (labels-as-vars)

- Все workflow условия сравнения label должны использовать `vars.*`, а не строковые литералы.
- В GitHub Variables хранится **полный каталог** `run:*`, `state:*`, `need:*`:
  - для `run:*`: `CODEXK8S_RUN_<STAGE>_LABEL` и `CODEXK8S_RUN_<STAGE>_REVISE_LABEL` (где применимо), плюс `CODEXK8S_RUN_DEBUG_LABEL`, `CODEXK8S_RUN_SELF_IMPROVE_LABEL`, `CODEXK8S_MODE_DISCUSSION_LABEL`,
  - для `state:*`: `CODEXK8S_STATE_*_LABEL`,
  - для `need:*`: `CODEXK8S_NEED_*_LABEL`.
- Для model/reasoning также хранится каталог vars:
  - `CODEXK8S_AI_MODEL_GPT_5_3_CODEX_LABEL`, `CODEXK8S_AI_MODEL_GPT_5_3_CODEX_SPARK_LABEL`, `CODEXK8S_AI_MODEL_GPT_5_2_CODEX_LABEL`, `CODEXK8S_AI_MODEL_GPT_5_1_CODEX_MAX_LABEL`, `CODEXK8S_AI_MODEL_GPT_5_2_LABEL`, `CODEXK8S_AI_MODEL_GPT_5_1_CODEX_MINI_LABEL`,
  - `CODEXK8S_AI_REASONING_LOW_LABEL`, `CODEXK8S_AI_REASONING_MEDIUM_LABEL`, `CODEXK8S_AI_REASONING_HIGH_LABEL`, `CODEXK8S_AI_REASONING_EXTRA_HIGH_LABEL`.
- Для новых `run:*` лейблов vars заводятся заранее до активации соответствующего этапа.
- Bootstrap синхронизация каталога выполняется командой `go run ./cmd/codex-bootstrap github-sync`.

## Аудит и наблюдаемость

- Для каждого изменения label фиксируются:
  - `correlation_id`,
  - `requested_by`/`applied_by`,
  - `approval_state` (если применимо),
  - `source` (human/mcp/system),
  - `timestamp`.
- Источник аудита: `flow_events` + связка с `agent_sessions`.
