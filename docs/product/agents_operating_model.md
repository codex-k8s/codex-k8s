---
doc_id: AGM-CK8S-0001
type: operating-model
title: "codex-k8s — Agents Operating Model"
status: active
owner_role: PM
created_at: 2026-02-11
updated_at: 2026-02-15
related_issues: [1, 19]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-19-full-docset"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-19
---

# Agents Operating Model

## TL;DR
- Базовый штат платформы: 8 системных агентных ролей (`pm`, `sa`, `em`, `dev`, `reviewer`, `qa`, `sre`, `km`) + человек `Owner`.
- Для каждого проекта допускаются расширяемые custom-агенты без нарушения базовых ролей.
- Режим исполнения смешанный: часть ролей работает в `full-env`, часть в `code-only`.
- Шаблоны промптов ведутся в role-specific матрице: для каждого `agent_key` отдельные body-шаблоны `work/revise`, с override в БД и seed fallback.
- Для каждого системного агента шаблоны поддерживаются минимум в `ru` и `en`; язык выбирается по locale с fallback до `en`.
- Управление агентами и шаблонами промптов проектируется как staff UI/API контур: `Agents` (настройки) и `Prompt templates` (diff/preview/effective preview), с использованием Monaco Editor для markdown.
- Для `run:self-improve` применяется совместный контур `km + dev + reviewer` с обязательной трассировкой источников улучшений.

## Source of truth
- `docs/product/requirements_machine_driven.md`
- `docs/product/labels_and_trigger_policy.md`
- `docs/product/stage_process_model.md`
- `docs/research/src_idea-machine_driven_company_requirements.md`

## Базовый штат (system agents)

| agent_key | Роль | Основной результат | Режим по умолчанию | Базовый лимит параллельных run на проект |
|---|---|---|---|---:|
| `pm` | Product Manager / BA | brief/PRD/scope/метрики | `code-only` | 1 |
| `sa` | Solution Architect | C4/ADR/NFR/design decisions | `full-env` (read-only) | 1 |
| `em` | Engineering Manager | delivery plan/epics/DoR-DoD | `full-env` (read-only) | 1 |
| `dev` | Software Engineer | реализация `run:dev`/`run:dev:revise`, код + тесты + docs update | `full-env` | 2 |
| `reviewer` | Pre-review Engineer | комментарии в существующем PR: inline findings + summary для Owner (без изменений репозитория) | `full-env` (read-mostly) | 2 |
| `qa` | QA Lead | markdown test strategy/plan/matrix/regression evidence | `full-env` | 2 |
| `sre` | SRE/OPS | markdown runbook/SLO/alerts/postdeploy improvements | `full-env` | 1 |
| `km` | Doc/KM | issue↔docs traceability, self-improve диагностика, prompts/instructions updates | `code-only` | 2 |

Примечания:
- `Owner` не является агентом, но остаётся финальным апрувером решений и trigger/deploy действий.
- Базовый профиль `run:dev` закреплён за системным агентом `dev`; custom-агенты могут расширять его поведение в рамках project policy.
- Роль `reviewer` заменяет прежний ad-hoc `auditor` режим: выполняет pre-review для всех `run:*`, где формируются артефакты на проверку Owner.

## Расширяемые custom-агенты проекта

Для каждого проекта допускается добавление custom-агентов.

Обязательные правила:
- custom-агент привязан к конкретному проекту и не меняет обязанности базовых системных ролей;
- для custom-агента обязательно задаются: ответственность, execution mode, RBAC-права, лимит параллелизма, шаблоны `work/revise`;
- custom-агент проходит тот же аудит событий (`flow_events`, `agent_sessions`, `token_usage`) и policy апрувов;
- custom-агент не может обходить политику trigger/deploy labels.

## Режимы исполнения

### `full-env`
- Запуск в issue/run namespace рядом со стеком.
- Доступ: логи, events, сервисы, метрики, DB/cache в рамках namespace, `exec` в pod'ы namespace.
- В pod передаются минимальные runtime-секреты (`CODEXK8S_OPENAI_API_KEY`, `CODEXK8S_GIT_BOT_TOKEN`) и формируется namespaced `KUBECONFIG`.
- GitHub операции (issue/PR/comments/review + git push) выполняются напрямую через `gh`/`git` с bot-token.
- Для PR-flow запрещено использовать `CODEXK8S_GITHUB_PAT`; допустим только `CODEXK8S_GIT_BOT_TOKEN`.
- Для локальных ручных GitHub операций оператор использует `CODEXK8S_GIT_BOT_TOKEN` из `bootstrap/host/config.env`.
- Kubernetes runtime-дебаг и изменения в своём namespace выполняются напрямую через `kubectl`.
- Исключение: прямой доступ к `secrets` (read/write) запрещён RBAC.
- MCP в MVP baseline используется для label-операций и control tools (`secret sync`, `database lifecycle`, `owner feedback`) по approval policy.
- Используется для ролей, где нужно подтверждать решения по фактическому состоянию окружения.

### Канал апрувов и уточнений
- Для операций через MCP используются HTTP approver/executor контракты.
- Telegram (`github.com/codex-k8s/telegram-approver`, `github.com/codex-k8s/telegram-executor`) является первым адаптером, но модель должна поддерживать и другие интеграции без изменений core-кода платформы.

### `code-only`
- Доступ только к репозиторию и API контексту без прямого Kubernetes runtime доступа.
- Используется для продуктовых/документационных задач без необходимости runtime-диагностики.

## Роли в цикле разработки и ревью

- `dev`:
  - реализует задачу, обновляет тесты и проектную документацию, в процессе выполняет дебаг решений;
  - формирует PR и прикладывает evidence проверок.
- `reviewer`:
  - выполняет предварительное ревью до финального ревью Owner для всех `run:*`;
  - проверяет соответствие задаче, проектной документации и `docs/design-guidelines/**`;
  - оставляет inline-комментарии в PR (если PR есть) и публикует summary для Owner;
  - не изменяет файлы репозитория и не создает коммиты/новые PR.
- `Owner`:
  - выполняет финальный review/approve после прохождения pre-review.
- `km`:
  - ведёт цикл `run:self-improve`: анализирует повторяющиеся замечания/сбои, формирует и вносит улучшения в docs/prompt templates;
  - в `run:self-improve` changeset ограничен scope: markdown-инструкции, prompt files, `services/jobs/agent-runner/Dockerfile`;
  - обязан использовать MCP-диагностику запусков:
    - `self_improve_runs_list` (пагинация истории запусков);
    - `self_improve_run_lookup` (поиск запусков по Issue/PR);
    - `self_improve_session_get` (извлечение `codex-cli` session JSON для анализа).
- `dev` (в self-improve контуре):
  - дорабатывает toolchain/agent image, если self-improve выявил отсутствие инструментов или runtime-gap.

## Политика шаблонов промптов (work/revise)

Классы шаблонов:
- `work` — шаблон на выполнение задачи;
- `revise` — шаблон на устранение замечаний Owner к существующему PR.

Источник и приоритет:
1. project-level override в БД;
2. global override в БД;
3. seed-файл в репозитории.

Seed-файлы:
- каталог `services/jobs/agent-runner/internal/runner/promptseeds/*.md` (embed в `agent-runner`) как bootstrap/fallback слой
  (минимальная stage-матрица, включая `dev-work`/`dev-revise`).

Обязательная целевая модель:
- для каждого `agent_key` поддерживаются отдельные body-шаблоны `work` и `revise`;
- для каждого `(agent_key, kind)` поддерживаются минимум локали `ru` и `en`;
- резолв в runtime выполняется по цепочке `project override -> global override -> repo seed`, с language fallback `project locale -> system default -> en`.

Текущий переходный профиль:
- stage-specific seed-шаблоны в репозитории остаются bootstrap/fallback источником до полного заполнения role-specific матрицы в БД;
- по мере наполнения role-specific матрицы stage seed остаются резервным fallback и не отменяют требования к отдельному body по роли.

Требования:
- изменение seed/override должно быть трассируемо через `flow_events`;
- в рантайме сохраняется effective template version/hash;
- шаблон не должен содержать секреты и не должен ослаблять policy безопасности.

### Локализация шаблонов
- Базовая загрузка в БД для системных агентов включает как минимум локали `ru` и `en` для `work` и `revise`.
- Выбор языка шаблона выполняется по цепочке:
  1. locale проекта;
  2. default locale системы;
  3. fallback `en`.
- Добавление новой локали планируется как отдельная фича (TODO следующего спринта):
  - оператор добавляет locale через staff UI (`System settings -> Locales`) или эквивалентный staff API;
  - платформа готовит версионируемые шаблоны на новую локаль (копия/инициализация от базовой локали);
  - опционально платформа генерирует первичный перевод шаблонов через ИИ;
  - перевод сохраняется как новая версия шаблона и может быть вручную скорректирован в staff UI (Monaco Editor).

### Контекстный рендер шаблонов
- Effective prompt рендерится с runtime-контекстом (окружение, namespace/slot, доступные MCP-сервера и инструменты, project/services context, issue/pr/run identifiers).
- Контракт контекстного рендера должен быть стабильным между версиями рантайма, чтобы избежать несовместимости seed/override шаблонов.
- В output contract рендера обязательно фиксируются:
  - communication language текущего запуска (для PR/комментариев/feedback-инструментов);
  - cadence для `run_status_report` (регулярный progress-feedback каждые 5-7 инструментальных вызовов).

### Role-specific prompt template matrix
- Для каждого `agent_key` используются отдельные шаблоны `work/revise`:
  - `pm`, `sa`, `em`, `dev`, `reviewer`, `qa`, `sre`, `km`;
  - отдельные шаблоны для специализированных режимов: `run:self-improve`, `mode:discussion`.
- Для каждого `(agent_key, kind)` обязательны минимум локали `ru` и `en`.
- Использование единого общего body-шаблона для всех ролей запрещено.
- Для `mode:discussion` шаблон обязан:
  - явно запрещать commit/push/PR;
  - требовать работу только комментариями под Issue;
  - требовать обработку новых пользовательских комментариев как продолжение той же сессии.
- Для стадий с артефактным ревью шаблон должен завершать run переходом в `state:in-review` и постановкой role-specific `need:*` labels.

### Модель и степень рассуждения
- По умолчанию профиль модели/рассуждений задается в настройках агента/проекта.
- На конкретном issue профиль может быть переопределен конфигурационными лейблами:
  - `[ai-model-*]` для модели;
  - `[ai-reasoning-*]` для уровня рассуждений.
- Для `run:dev:revise` effective model/reasoning перечитываются на каждый запуск, чтобы Owner мог поменять профиль между итерациями revise.
- При конфликтующих лейблах одной группы запуск отклоняется как `failed_precondition` и требует ручного исправления labels.

## Инфраструктурная модель и capacity baseline

- По умолчанию одновременно активен максимум 1 `run:dev` на issue.
- Для `run:self-improve` допускается не более 1 активного запуска на issue/PR связку, чтобы избежать конфликтующих улучшений.
- `run:self-improve` выполняется как управляемый диагностический цикл:
  - извлечение run/session evidence;
  - анализ причин;
  - PR с улучшениями prompt/docs/guidelines/toolchain.
- Лимиты параллелизма на проект задаются в настройках проекта и применяются worker-очередью.
- Для `full-env` запусков обязателен отдельный namespace на run/issue с контролем cleanup.
- MCP policy для инструментов/ресурсов определяется по связке:
  - `agent_key` (базовый профиль роли);
  - `run:*` label/тип задачи (task override).

## Resume и timeout-policy

- Во время ожидания review от Owner run может быть приостановлен и возобновлён.
- Во время ожидания ответа MCP (`wait_state=mcp`) pod/run не должен завершаться по timeout.
- Для resumable выполнения в сессии сохраняется `codex-cli` session JSON, чтобы продолжать с того же места.

## Planned: discussion-mode before implementation

- Для `run:dev`/`run:dev:revise` планируется режим `mode:discussion`:
  - запускается отдельный `discussion` pod (не job) с сохранением `codex-cli` session snapshot;
  - агент работает только в комментариях Issue (brainstorming/уточнения), commit/push/PR не выполняются;
  - pod живёт до первого события: idle timeout `8h`, закрытие Issue или постановка любого `run:*` label;
  - на webhook `issue_comment`, если автор комментария не агент, сессия продолжается и агент публикует ответ под Issue;
  - после снятия `mode:discussion` следующий trigger продолжает ту же session snapshot и переходит к реализации.

## Управление изменениями модели

- Изменения состава базового штата, execution modes или policy шаблонов фиксируются в:
  - `docs/product/requirements_machine_driven.md`,
  - `docs/architecture/data_model.md`,
  - `docs/architecture/agent_runtime_rbac.md`,
  - `docs/delivery/requirements_traceability.md`.
