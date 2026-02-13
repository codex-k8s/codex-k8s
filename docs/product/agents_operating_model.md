---
doc_id: AGM-CK8S-0001
type: operating-model
title: "codex-k8s — Agents Operating Model"
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

# Agents Operating Model

## TL;DR
- Базовый штат платформы: 8 системных агентных ролей (`pm`, `sa`, `em`, `dev`, `reviewer`, `qa`, `sre`, `km`) + человек `Owner`.
- Для каждого проекта допускаются расширяемые custom-агенты без нарушения базовых ролей.
- Режим исполнения смешанный: часть ролей работает в `full-env`, часть в `code-only`.
- Шаблоны промптов для работы и ревью имеют seed в репозитории и override в БД.
- Для каждого системного агента шаблоны поддерживаются минимум в `ru` и `en`; язык выбирается по locale с fallback до `en`.
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
| `reviewer` | Pre-review Engineer | предварительное ревью PR: inline findings + summary для Owner | `full-env` (read-mostly) | 2 |
| `qa` | QA Lead | test strategy/plan/matrix/regression | `full-env` | 2 |
| `sre` | SRE/OPS | runbook/SLO/alerts/postdeploy | `full-env` | 1 |
| `km` | Doc/KM | issue↔docs traceability, индексы, self-improve диагностика и обновление знаний | `code-only` | 2 |

Примечания:
- `Owner` не является агентом, но остаётся финальным апрувером решений и trigger/deploy действий.
- Базовый профиль `run:dev` закреплён за системным агентом `dev`; custom-агенты могут расширять его поведение в рамках project policy.
- Роль `reviewer` заменяет прежний ad-hoc `auditor` режим для PR-precheck: сначала замечания получает `dev`-агент, затем Owner делает финальное code review.

## Расширяемые custom-агенты проекта

Для каждого проекта допускается добавление custom-агентов.

Обязательные правила:
- custom-агент привязан к конкретному проекту и не меняет обязанности базовых системных ролей;
- для custom-агента обязательно задаются: ответственность, execution mode, RBAC-права, лимит параллелизма, шаблоны `work/review`;
- custom-агент проходит тот же аудит событий (`flow_events`, `agent_sessions`, `token_usage`) и policy апрувов;
- custom-агент не может обходить политику trigger/deploy labels.

## Режимы исполнения

### `full-env`
- Запуск в issue/run namespace рядом со стеком.
- Доступ: логи, events, сервисы, метрики, DB/cache в рамках namespace, `exec` в pod'ы namespace.
- В pod передаются минимальные runtime-секреты (`CODEXK8S_OPENAI_API_KEY`, `CODEXK8S_GIT_BOT_TOKEN`) и формируется namespaced `KUBECONFIG`.
- GitHub операции (issue/PR/comments/review + git push) выполняются напрямую через `gh`/`git` с bot-token.
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
  - выполняет предварительное ревью до финального ревью Owner;
  - проверяет соответствие задаче, проектной документации и `docs/design-guidelines/**`;
  - оставляет inline-комментарии в PR для `dev`-агента и публикует summary для Owner.
- `Owner`:
  - выполняет финальный review/approve после прохождения pre-review.
- `km`:
  - ведёт цикл `run:self-improve`: анализирует повторяющиеся замечания/сбои, формирует и вносит улучшения в docs/prompt templates.
- `dev` (в self-improve контуре):
  - дорабатывает toolchain/agent image, если self-improve выявил отсутствие инструментов или runtime-gap.

## Политика шаблонов промптов (work/review)

Классы шаблонов:
- `work` — шаблон на выполнение задачи;
- `review` — шаблон на аудит/ревью результата.

Источник и приоритет:
1. project-level override в БД;
2. global override в БД;
3. seed-файл в репозитории.

Seed-файлы:
- `docs/product/prompt-seeds/dev-work.md`
- `docs/product/prompt-seeds/dev-review.md`

Текущий baseline (S3 Day1):
- в runtime активно используются `dev-work` и `dev-review`;
- для остальных ролей и специализированных режимов используется planned matrix ниже.

Требования:
- изменение seed/override должно быть трассируемо через `flow_events`;
- в рантайме сохраняется effective template version/hash;
- шаблон не должен содержать секреты и не должен ослаблять policy безопасности.

### Локализация шаблонов
- Базовая загрузка в БД для системных агентов включает как минимум локали `ru` и `en` для `work` и `review`.
- Выбор языка шаблона выполняется по цепочке:
  1. locale проекта;
  2. default locale системы;
  3. fallback `en`.
- Добавление новой локали планируется как отдельная фича:
  - оператор добавляет locale в систему;
  - платформа генерирует перевод шаблонов через ИИ;
  - перевод сохраняется как новая версия шаблона и может быть вручную скорректирован.

### Контекстный рендер шаблонов
- Effective prompt рендерится с runtime-контекстом (окружение, namespace/slot, доступные MCP-сервера и инструменты, project/services context, issue/pr/run identifiers).
- Контракт контекстного рендера должен быть стабильным между версиями рантайма, чтобы избежать несовместимости seed/override шаблонов.

### Planned: role-specific prompt template matrix
- Для каждого `agent_key` вводятся отдельные шаблоны `work/review`:
  - `pm`, `sa`, `em`, `dev`, `reviewer`, `qa`, `sre`, `km`;
  - отдельные шаблоны для специализированных режимов: `run:self-improve`, `mode:discussion`.
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
- Для `run:dev:revise` effective model/reasoning перечитываются на каждый запуск, чтобы Owner мог поменять профиль между итерациями review.
- При конфликтующих лейблах одной группы запуск отклоняется как `failed_precondition` и требует ручного исправления labels.

## Инфраструктурная модель и capacity baseline

- По умолчанию одновременно активен максимум 1 `run:dev` на issue.
- Для `run:self-improve` допускается не более 1 активного запуска на issue/PR связку, чтобы избежать конфликтующих улучшений.
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
