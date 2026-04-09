<div align="center">
  <img src="docs/media/logo.png" alt="kodex logo" width="120" height="120" />
  <h1>kodex</h1>
  <p>🧠 Webhook-driven платформа управления AI-агентами в Kubernetes для полного цикла: от идеи и документации до кода, PR, деплоя и эксплуатации.</p>
  <p><b>Что дает платформа:</b> агенты работают в полноценном, но изолированном runtime-окружении Kubernetes; процесс разработки ведется через документацию и stage-модель; весь delivery-цикл закрывается в одной системе — от discovery и проектирования до реализации, ревью, релиза и ops.</p>
</div>

![Go Version](https://img.shields.io/github/go-mod/go-version/codex-k8s/kodex)
[![Go Reference](https://pkg.go.dev/badge/github.com/codex-k8s/kodex.svg)](https://pkg.go.dev/github.com/codex-k8s/kodex)

## 🚧 Статус Публикации

Платформа находится в финальной стадии доработки. Пока репозиторий опубликован без права форка.

Когда будет готов MVP, здесь появятся:
- подробные инструкции по запуску и использованию;
- видео-демонстрации ключевых сценариев;
- анонсы в Telegram-канале: https://t.me/ai_da_stas;
- открытие возможности форкать репозиторий.

Для понимания текущих возможностей рекомендую посмотреть следующие этапы и связанные артефакты:

| Этап | Issue | PR |
|---|---|---|
| S6 Day1 - Intake | [#184 Проработка раздела редактирования параметров агентов и их промптов](https://github.com/codex-k8s/kodex/issues/184) | [#186 Issue #184: intake-документация и трассируемость для S6 Agents/Prompt lifecycle](https://github.com/codex-k8s/kodex/pull/186) |
| S6 Day2 - Vision | [#185 S6 Day2: Vision для lifecycle управления агентами и шаблонами промптов](https://github.com/codex-k8s/kodex/issues/185) | [#188 S6 Day2 Vision: lifecycle управления агентами и prompt templates](https://github.com/codex-k8s/kodex/pull/188) |
| S6 Day3 - PRD | [#187 S6 Day3: PRD для lifecycle управления агентами и шаблонами промптов](https://github.com/codex-k8s/kodex/issues/187) | [#190 Issue #187: PRD-пакет S6 Day3 и handover в run:arch](https://github.com/codex-k8s/kodex/pull/190) |
| S6 Day4 - Architecture | [#189 S6 Day4: Architecture для lifecycle управления агентами и шаблонами промптов](https://github.com/codex-k8s/kodex/issues/189) | [#196 S6 Day4: архитектурный пакет и handover для lifecycle агентов и prompt templates](https://github.com/codex-k8s/kodex/pull/196) |
| S6 Day5 - Design | [#195 S6 Day5: Design для lifecycle управления агентами и шаблонами промптов](https://github.com/codex-k8s/kodex/issues/195) | [#198 S6 Day5: Design package для lifecycle управления агентами и шаблонами](https://github.com/codex-k8s/kodex/pull/198) |
| S6 Day6 - Plan | [#197 S6 Day6: Plan для реализации lifecycle управления агентами и шаблонами промптов](https://github.com/codex-k8s/kodex/issues/197) | [#200 S6 Day6: execution package и handover в run:dev для lifecycle agents/prompts](https://github.com/codex-k8s/kodex/pull/200) |
| S6 Day7 - Development | [#199 S6 Day7: Реализация lifecycle управления агентами и шаблонами промптов](https://github.com/codex-k8s/kodex/issues/199) | [#202 S6 Day7: реализация lifecycle управления агентами и шаблонами промптов (#199)](https://github.com/codex-k8s/kodex/pull/202) |

По факту этапы выше были выполнены менее чем за сутки. Далее идут этапы `qa -> release -> postdeploy -> ops`: каждый запускается своим label на задаче, и каждый выполняется отдельным агентом со своей ролью и инструкциями.
Для late-stage delivery действует единая runtime-семантика: `run:dev -> run:qa -> run:release` в `full-env` продолжают один candidate namespace/build lineage до merge, а `run:postdeploy -> run:ops` работают уже против production namespace с read-only RBAC.

### 🖼️ Текущий Вид Платформы
> UI/UX на финальном этапе MVP будет дорабатываться и полироваться.

![Текущий скриншот платформы](docs/media/screenshot.png)

`kodex` запускает агентные роли по GitHub-лейблам (`run:*`) и ведет задачу по stage-модели:
`intake -> vision -> prd -> arch -> design -> plan -> dev -> qa -> release -> postdeploy -> ops`.

Ключевые функции платформы:
- 🚀 запуск агентных job в Kubernetes (code-only/full-env);
- 🧾 управление stage-процессом через issue labels;
- 🛡️ review gate (`state:*`, `need:*`) и аудит `flow_events`;
- 🔧 сборка/деплой и наблюдаемость через staff web-console;
- 📦 работа с конфигами/секретами, repo preflight, runtime diagnostics.

## 👥 Роли агентов и как запускать

Системные роли:
- `pm` — продуктовая проработка и формализация требований;
- `sa` — архитектура и контракты;
- `em` — планирование, delivery и quality-gates;
- `dev` — реализация и PR;
- `reviewer` — pre-review и технические замечания;
- `qa` — тестовые сценарии и регресс;
- `sre` — эксплуатация, надежность, postdeploy/ops;
- `km` — документация, трассируемость, self-improve.

Запуск выполняется постановкой `run:*` лейбла на Issue:
- `run:intake`, `run:vision`, `run:prd`, `run:arch`, `run:design`, `run:plan`, `run:dev`, `run:qa`, `run:release`, `run:postdeploy`, `run:ops`, `run:self-improve`, `run:rethink`.
- Для доработок по замечаниям Owner используйте `run:*:revise` (например, `run:dev:revise`).

Служебные лейблы:
- `state:*` — статус этапа (`state:in-review`, `state:approved` и т.д.);
- `need:*` — запрос участия роли (`need:qa`, `need:sa`, `need:reviewer` и т.д.);
- `mode:discussion` — lightweight discussion-run под Issue: поднимает long-lived comment-only pod в отдельном lightweight namespace без PR/commit/push.

## 🏷️ Полный процесс по лейблам (Issue + PR)

### 1. Какие лейблы реально запускают ран
- Базовый запуск ран делает класс `run:*`.
- Исключение: `need:reviewer` на PR (событие `pull_request:labeled`) запускает pre-review ран роли `reviewer`.
- `mode:discussion` на Issue сам по себе запускает lightweight long-lived discussion-run.
- `state:*`, остальные `need:*`, `[ai-model-*]`, `[ai-reasoning-*]` сами по себе ран не запускают.

### 2. Базовый запуск по Issue
1. Вешаете на Issue один trigger-лейбл `run:<stage>`.
2. Платформа создает run и стартует роль, соответствующую stage.
3. По завершению stage вы переходите к следующему `run:<next-stage>` или к `run:<stage>:revise` при замечаниях.
4. После завершения run платформа обновляет единый GitHub service-comment и публикует матрицу `Следующие шаги`:
   - каждая ссылка ведёт на `/` staff web-console;
   - на фронте открывается confirm-модалка с preview diff лейблов (`removed / added / final`);
   - после подтверждения transition применяется через staff API и аудитится в `flow_events`.

Поддержанные stage:
- `run:intake`, `run:vision`, `run:prd`, `run:arch`, `run:design`, `run:plan`, `run:dev`, `run:doc-audit`, `run:qa`, `run:release`, `run:postdeploy`, `run:ops`, `run:self-improve`, `run:rethink`.

Late-stage routing:
- `run:dev` создаёт candidate runtime или продолжает уже существующий candidate lineage той же Issue/PR.
- `run:qa` и `run:release` используют только существующий candidate lineage; при его отсутствии платформа не делает silent fallback на default branch и запрашивает `need:input`.
- `run:postdeploy` и `run:ops` таргетят `production` и получают `production-readonly` профиль без `exec`, `port-forward`, `secrets` и mutating операций.

### 3. Revise-запуски
- Ручной revise: ставите `run:<stage>:revise` на Issue.
- Автоматический revise по PR review (`changes_requested`):
  - webhook `pull_request_review` запускает revise только если на PR стоит ровно один stage из пар:
    - `run:intake`/`run:intake:revise`
    - `run:vision`/`run:vision:revise`
    - `run:prd`/`run:prd:revise`
    - `run:arch`/`run:arch:revise`
    - `run:design`/`run:design:revise`
    - `run:plan`/`run:plan:revise`
    - `run:dev`/`run:dev:revise`
    - `run:doc-audit`/`run:doc-audit:revise`
    - `run:qa`/`run:qa:revise`
    - `run:release`/`run:release:revise`
    - `run:postdeploy`/`run:postdeploy:revise`
    - `run:ops`/`run:ops:revise`
    - `run:self-improve`/`run:self-improve:revise`
  - результат: платформа запускает соответствующий `run:<stage>:revise`.
  - если на PR нет stage-лейбла или stage-лейблов несколько, ран не создается.

### 3.1 Pre-review по PR
- Для ручного запуска ревьюера на конкретном PR достаточно поставить лейбл `need:reviewer` на PR.
- Платформа создаёт ран с ролью `reviewer` в контексте текущего PR.
- Роль `reviewer` работает только комментариями в существующем PR (без коммитов/пуша).

### 4. Что можно вешать параллельно
Можно одновременно (без конфликта запуска):
- `1 x run:*` + `state:*` + `need:*`
- `1 x run:*` + `1 x [ai-model-*]` + `1 x [ai-reasoning-*]`
- `state:*` + `need:*` без `run:*` (чисто workflow-координация людей)

Практический смысл:
- для `full-env` namespace живёт по role-based TTL policy; для `run:<stage>:revise` lease продлевается.
- `need:*` сигнализирует, чье участие требуется до следующего запуска.
- `state:*` фиксирует прогресс процесса (например, `state:in-review`).

### 5. Что нельзя вешать параллельно
Нельзя:
- несколько `run:*` trigger-лейблов одновременно на одной Issue;
- несколько stage-лейблов из списка выше одновременно на одном PR, если ждете авто-revise от `changes_requested`;
- несколько `[ai-model-*]` одновременно;
- несколько `[ai-reasoning-*]` одновременно.

Что будет при нарушении:
- для конфликтных `run:*` на Issue ран не стартует;
- для конфликтных stage-лейблов на PR при `changes_requested` ран не стартует;
- при конфликте model/reasoning run завершается `failed_precondition`.

### 6. Как не конфликтовать агентами в ветке
- На одну рабочую ветку одновременно запускайте только один активный `run:*`.
- Для новой параллельной задачи используйте отдельную Issue + отдельную ветку.
- Для PR-итераций держите на PR один stage-лейбл (обычно `run:dev` или `run:dev:revise`).
- Если нужно параллельно двигать документацию и код:
  - делайте разные PR в разные ветки;
  - не запускайте два `run:dev*` на один и тот же PR одновременно.

## 🧭 Практические флоу по ролям (7 сценариев)

Ниже — рабочие сценарии, сверенные с текущими prompt seeds (`services/jobs/agent-runner/internal/runner/promptseeds/*.md`),
stage-моделью и политикой лейблов.

### Карта этапов: роли и шаблоны артефактов

| Stage | Ключевые роли | Что обычно обновляется (templates) |
|---|---|---|
| `run:intake` | `pm`, `km` | `problem.md`, `brief.md`, `scope_mvp.md`, `personas.md`, `constraints.md`, `risks_register.md` |
| `run:vision` | `pm`, `em` | `project_charter.md`, `success_metrics.md`, `risks_register.md` |
| `run:prd` | `pm`, `sa` | `prd.md`, `nfr.md`, `user_story.md` |
| `run:arch` | `sa` | `c4_context.md`, `c4_container.md`, `adr.md`, `alternatives.md` |
| `run:design` | `sa`, `qa` | `design_doc.md`, `api_contract.md`, `data_model.md`, `migrations_policy.md` |
| `run:plan` | `em`, `km` | `delivery_plan.md`, `epic.md`, `definition_of_done.md`, `issue_map.md` |
| `run:dev` | `dev`, `reviewer` | код + PR + синхронные docs updates |
| `run:doc-audit` | `km` | аудит и синхронизация docs/traceability |
| `run:qa` | `qa` | `test_strategy.md`, `test_plan.md`, `test_matrix.md`, `regression_checklist.md` |
| `run:release` | `em`, `sre` | `release_plan.md`, `release_notes.md`, `rollback_plan.md` |
| `run:postdeploy` | `qa`, `sre` | `postdeploy_review.md`, `incident_postmortem.md` |
| `run:ops` | `sre`, `km` | `runbook.md`, `slo.md`, `alerts.md`, `monitoring.md` |
| `run:self-improve` | `km`, `dev`, `reviewer` | улучшения prompts/docs/tooling по run-evidence |

### 1) Новая фича в существующей системе (новый сервис + доработка фронта)

1. Создайте Issue с целью, AC, рамками backend/frontend и критериями проверки.
2. Для крупной фичи идите полной цепочкой: `run:intake -> run:vision -> run:prd -> run:arch -> run:design -> run:plan`.
3. На `run:design` зафиксируйте transport/data/UI:
   - для `services/external/*` и `services/staff/*` — сначала OpenAPI (`api/server/api.yaml`), потом codegen;
   - для backend — сервисные границы, владельца схемы и миграции.
4. Запустите `run:dev` для реализации и PR.
5. Дождитесь `state:in-review`, пройдите pre-review (`reviewer`) и комментарии Owner.
6. На замечания запускайте `run:dev:revise` до полного закрытия threads.
7. После этого: `run:qa -> run:release -> run:postdeploy -> run:ops`.

### 2) Доработка существующего функционала (пример: RBAC для UI и webhook-flow)

1. Если impact только локальный — можно стартовать с `run:prd`; если impact сквозной (как RBAC), лучше с `run:intake`.
2. В PRD зафиксируйте:
   - матрицу ролей и разрешений;
   - как RBAC влияет на UI;
   - как роль пользователя влияет на webhook/label-trigger поведение.
3. На `run:arch` разделите ответственность по границам:
   - `services/external/*` — валидация/authn/authz на edge;
   - доменная авторизация и правила переходов — во внутренних сервисах (`services/internal/*`).
4. На `run:design` проверьте typed DTO/contract-first и сценарии отказов.
5. На `run:qa` прогоняйте role-matrix (позитив/негатив) и регресс webhook-процесса.
6. На замечания используйте `run:dev:revise`.

### 3) Новый проект с нуля

1. Создайте проект/репозиторий в staff UI и выполните `Repository preflight`.
2. Создайте issue-инициативу (problem, users, metrics, constraints).
3. Пройдите базовую stage-цепочку: `run:intake -> run:vision -> run:prd -> run:arch -> run:design -> run:plan`.
4. На каждом этапе держите review gate: `state:in-review` + профильный `need:*`, при необходимости — `run:<stage>:revise`.
5. После `run:plan` запускайте `run:dev`, затем `run:qa`, `run:release`, `run:postdeploy`, `run:ops`.
6. При смене направления используйте `run:rethink`, а устаревшие решения фиксируйте как superseded.

### 4) Отладка и устранение замечаний (runtime + revise)

1. Для расследования runtime-проблем запускайте `run:dev` или `run:dev:revise` в `full-env`.
2. Агент в `full-env` проверяет логи/события/ресурсы namespace через `kubectl` (к `secrets` доступа нет по RBAC).
3. Namespace сохраняется по TTL policy; для `run:<stage>:revise` платформа переиспользует namespace связки `(project, issue, agent_key)` и продлевает lease.
4. Для ревизии PR используйте `run:dev:revise`:
   - собрать все открытые review-комментарии;
   - ответить на каждый (исправлено или обоснованно отклонено);
   - запушить правки в ту же PR-ветку.
5. Для авто-revise по `changes_requested` на PR держите ровно один поддержанный stage label (`run:<stage>` или `run:<stage>:revise`).

### 5) Проверка соответствия гайдам/документации и создание эпика на устранение нарушений

1. Создайте issue на аудит и поставьте `run:doc-audit`.
2. `km` фиксирует расхождения: код ↔ docs ↔ checklists ↔ требования.
3. Если нарушения системные, запускайте `run:plan` и оформляйте remediation-эпик:
   - epic + декомпозиция задач;
   - приоритизация (`P0/P1/P2`);
   - traceability в `issue_map` и `requirements_traceability`.
4. Реализацию remediation-задач ведите через `run:dev`/`run:dev:revise`.
5. Для завершения цикла добавьте `run:qa` и при необходимости `run:ops` (если затронута эксплуатация).

### 6) Обсуждение идеи в режиме диалога под Issue

Как это работает:
1. Ставите на Issue `mode:discussion`.
2. Платформа создает long-lived `code-only` pod в отдельном lightweight namespace; PR/commit/push не используются.
3. Агент отвечает пользователю комментариями под Issue через `gh issue comment`.
4. Пока `mode:discussion` висит на Issue, тот же pod продолжает одну и ту же discussion-сессию и периодически перечитывает Issue/comments.
5. Каждый новый пользовательский `issue_comment` не создает новый run, если discussion-pod уже активен: текущая сессия должна увидеть новый комментарий и ответить на него.
6. Служебные комментарии платформы и комментарии GitHub-бота новый discussion-run не запускают.
7. Если на Issue дополнительно ставится любой `run:*`, discussion-контекст останавливается, discussion namespace удаляется и запускается обычный stage-run.
8. Если снять `mode:discussion`, закрыть или удалить Issue, discussion namespace и pod удаляются.

### 7) Self-improve по завершённым задачам

1. Создайте issue на улучшение и поставьте `run:self-improve`.
2. Агент обязан собрать evidence:
   - MCP: `self_improve_runs_list`, `self_improve_run_lookup`, `self_improve_session_get`;
   - GitHub: issue/PR/comments/review threads.
3. Session JSON сохраняется в `/tmp/codex-sessions/<run-id>/`.
4. По результату выходит PR с трассировкой:
   - `источник (run/session/comment)` -> `диагноз` -> `изменение`.
5. Обновления обычно затрагивают prompts, docs/guidelines, иногда toolchain/скрипты.
6. После review Owner изменения вливаются как очередной цикл улучшения платформы.
