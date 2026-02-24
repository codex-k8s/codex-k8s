---
doc_id: LBL-CK8S-0001
type: labels-policy
title: "codex-k8s — Labels and Trigger Policy"
status: active
owner_role: PM
created_at: 2026-02-11
updated_at: 2026-02-24
related_issues: [1, 19, 74, 90, 95, 154]
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
- Для review->revise цикла реализован гибридный resolver stage/profile и stage-aware сервисные сообщения (Issue #95, ADR-0006).

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
| `run:ai-repair` | аварийное восстановление инфраструктуры и runtime-потока (production pod, main-direct режим без обязательного PR) | active |
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

### Политика scope изменений для `run:*`
- Для `run:intake|vision|prd|arch|design|plan|doc-audit|qa|release|postdeploy|ops|rethink` разрешены только изменения markdown-документации (`*.md`).
- `run:dev|run:dev:revise` остаются единственными trigger-лейблами для кодовых изменений.
- Роль `reviewer` работает только в существующем PR и оставляет комментарии; изменения репозитория для reviewer запрещены.
- Для `run:self-improve` разрешены только:
  - prompt files (`services/jobs/agent-runner/internal/runner/promptseeds/**`, `services/jobs/agent-runner/internal/runner/templates/prompt_envelope.tmpl`);
  - markdown-инструкции/документация (`*.md`);
  - `services/jobs/agent-runner/Dockerfile`.
  Остальные изменения считаются policy violation.
### Namespace retention policy (`full-env`)
- Для `full-env` namespace сохраняется по role-based TTL из `services.yaml` (default `24h`).
- Для `run:<stage>:revise` lease namespace продлевается (`expires_at = now + role_ttl`).
- Отдельный diagnostic label для manual-retention не используется.


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
    если удаётся детерминированно определить stage по policy резолва.
- Current baseline (S3, Issue #95):
  - stage резолвится по цепочке:
    1. PR stage label (если ровно один),
    2. Issue stage label (если ровно один),
    3. последний run context по связке `(repo, issue, pr)`,
    4. последний stage transition в `flow_events`;
  - при конфликте/неопределённости revise-run не создаётся, выставляется `need:input` и публикуется action-card с remediation.
- Для `run:dev:revise` при отсутствии связанного PR run отклоняется с `failed_precondition` и событием `run.revise.pr_not_found`.
- Для `run:<stage>:revise` в `full-env` worker пытается переиспользовать активный namespace текущей связки `(project, issue, agent_key)` и продлить lease по TTL роли; если namespace отсутствует или уже в `Terminating`, создаётся новый.
- При постановке trigger-лейбла платформа сразу даёт обратную связь в issue:
  - ставит reaction `:eyes:` (если ещё нет);
  - публикует/обновляет единый статус-комментарий в фазе «планируется запуск агента»;
  - дальше обновляет тот же комментарий по фазам `подготовка окружения -> запуск -> завершение`.
- Label transitions после завершения run должны выполняться через MCP (а не вручную в коде агента), чтобы сохранять единый policy/audit контур.
- Для owner next-step action-link в staff web-console используется отдельный staff endpoint перехода этапа (RBAC + аудит), без прямых мутаций лейблов из frontend.
- Для dev/dev:revise transition выполняется так:
  - снять trigger label с Issue;
  - поставить `state:in-review` на PR и на Issue.
- S2 baseline:
  - pre-review остается обязательным шагом перед финальным Owner review;
  - post-run transitions `run:* -> state:*` фиксируются в Day5/Day6 как отдельные доработки policy и аудита.
- S3 Day1 факт:
- активирован полный stage-контур `run:intake..run:ops` + revise/rethink;
  - добавлен аварийный инфраструктурный trigger `run:ai-repair`;
- для `run:ai-repair` используется специальный pod-path в production namespace (не full-env slot), с fallback по образу и main-direct recovery режимом;
- активирован trigger path для `run:self-improve` (расширенная бизнес-логика дорабатывается по S3 Day6+).

### Implemented UX improvements for review/revise (Issue #95)

#### Варианты организации
| Вариант | Суть | Плюсы | Минусы |
|---|---|---|---|
| A | Оставить только PR stage label как триггер auto-revise | простая и прозрачная логика | высокий ручной overhead у Owner |
| B | Резолвить stage только по Issue | меньше ручных действий на PR | ломается при рассинхроне Issue labels |
| C (recommended) | Гибридный resolver + sticky profile + stage-aware сервисные сообщения | лучший баланс UX и детерминированности | выше сложность orchestration |

#### Sticky model/reasoning profile (implemented)
- Для `changes_requested` effective profile резолвится по приоритету:
  1. `[ai-model-*]`/`[ai-reasoning-*]` на Issue;
  2. те же лейблы на PR;
  3. профиль последнего run по связке `(repo, issue, pr)`;
  4. project/agent defaults.
- Цель: убрать обязательность ручного повторного выбора model/reasoning перед каждой revise-итерацией.

#### Stage-aware action cards в service-comment (implemented)
- Платформа обновляет единый service-comment и добавляет stage-aware подсказки:
  - `intake|vision|prd|arch|design|plan`: `run:<stage>:revise` и `run:<next-stage>`;
  - `dev`: `run:dev:revise`, `run:qa`;
  - `qa|release|postdeploy|ops`: revise текущего stage (если применимо) и следующий stage.
- Карточка остаётся компактной: обычно 2 action-подсказки (`revise` + канонический `next-stage`).
- Для `design` публикуется дополнительный fast-track вариант `run:dev` (вместе с каноническим `run:plan`).
- В сообщении всегда остаются ссылки:
  - на Issue;
  - на PR;
  - на актуальный run-status/диагностический комментарий;
  - на явный список рекомендованных label-действий.
- При ambiguous stage resolve:
  - revise-run не стартует;
  - выставляется `need:input`;
  - публикуется remediation-message с конкретным требуемым label action.

#### Next-step deep-link в web-console (implemented)
- Action-link из GitHub service-comment открывает staff web-console для перехода этапа (`/governance/labels-stages?...`).
- На фронте выполняется RBAC-проверка (platform admin guard) и показывается confirm-модалка перехода.
- После подтверждения backend выполняет label transition на Issue: снимает текущие `run:*` и ставит целевой `run:*`.

### Planned UX hardening for launch/transition (Issue #154)
- Проблема текущего UX: ссылки для быстрого добавления лейблов могут быть нерабочими в зависимости от контекста comment/web/session и блокируют owner-flow.
- Норматив:
  - next-step action-card обязан содержать два канала запуска:
    - validated deep-link в staff web-console (primary);
    - fallback-команду в текстовом виде (copy-paste для `gh`/label transition) без открытия UI.
  - сервисные сообщения должны публиковать не только next-step label, но и выбранный launch profile (`quick-fix`, `feature`, `new-service`) с краткой расшифровкой stage-пути.
  - если primary deep-link недоступен (invalid host/RBAC/session), run не блокируется: owner может завершить переход через fallback-команду.
- Ограничения безопасности:
  - fallback-команды не содержат секретов/токенов;
  - любой transition остаётся в policy/audit контуре (`flow_events` + actor + correlation_id).

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
  - для `run:*`: `CODEXK8S_RUN_<STAGE>_LABEL` и `CODEXK8S_RUN_<STAGE>_REVISE_LABEL` (где применимо), плюс `CODEXK8S_RUN_SELF_IMPROVE_LABEL`, `CODEXK8S_MODE_DISCUSSION_LABEL`,
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
