---
doc_id: AGM-CK8S-0001
type: operating-model
title: "codex-k8s — Agents Operating Model"
status: active
owner_role: PM
created_at: 2026-02-11
updated_at: 2026-03-09
related_issues: [1, 19, 74, 175, 247, 248, 249]
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
- MVP использует фиксированный штат из 8 системных ролей: `pm`, `sa`, `em`, `dev`, `reviewer`, `qa`, `sre`, `km`.
- `Owner` не является агентом и остается финальным человеком-апрувером stage-переходов, PR и deploy-решений.
- Режимы исполнения смешанные: часть ролей работает в `full-env`, часть в `code-only`.
- Prompt templates в текущем MVP берутся только из repo seeds (`services/jobs/agent-runner/internal/runner/promptseeds/*.md`), без DB overrides и без staff UI редактирования.
- Locale для prompt templates в текущем MVP задается platform default (`CODEXK8S_AGENT_DEFAULT_LOCALE`, fallback `ru`); при unsupported locale рендер нормализует значение к `en`.
- Контуры `Agents UI`, `Prompt templates UI`, кастомные runtime settings и custom-agent factory выведены из MVP и остаются post-MVP направлением.

## Source of truth
- `docs/product/requirements_machine_driven.md`
- `docs/product/labels_and_trigger_policy.md`
- `docs/product/stage_process_model.md`
- `docs/architecture/agent_runtime_rbac.md`
- `docs/architecture/prompt_templates_policy.md`

## Базовый штат системных агентов

| agent_key | Роль | Основной результат | Режим по умолчанию | Базовый лимит параллельных run на проект |
|---|---|---|---|---:|
| `pm` | Product Manager / BA | brief/PRD/scope/метрики | `code-only` | 1 |
| `sa` | Solution Architect | C4/ADR/NFR/design decisions | `full-env` (read-only) | 1 |
| `em` | Engineering Manager | delivery plan/epics/DoR-DoD | `full-env` (read-only) | 1 |
| `dev` | Software Engineer | реализация `run:dev`/`run:dev:revise`, код + тесты + docs update | `full-env` | 2 |
| `reviewer` | Pre-review Engineer | замечания в PR и summary для Owner | `full-env` (read-mostly) | 2 |
| `qa` | QA Lead | test strategy/plan/matrix/regression evidence | `full-env` | 2 |
| `sre` | SRE / OPS | runbook/SLO/alerts/postdeploy/ops, emergency recovery (`run:ai-repair`) | `code-only` / special production pod | 1 |
| `km` | Knowledge Manager | traceability, docs governance, `run:self-improve` | `code-only` | 2 |

Примечания:
- `dev` остается единственной системной ролью, которая готовит production code changes и PR.
- `reviewer` не изменяет репозиторий и не создает коммиты: его зона ответственности ограничена review feedback.
- `sre` использует отдельный аварийный контур `run:ai-repair`, работающий рядом с production namespace.

## Execution modes

### `full-env`
- Запуск выполняется в отдельном issue/run namespace рядом со стеком проекта.
- Агент имеет доступ к логам, events, pod/deploy/service runtime и диагностике через `kubectl`.
- Прямой доступ к `secrets` запрещен RBAC.
- Для `run:*:revise` namespace переиспользуется и TTL lease продлевается.
- GitHub операции выполняются напрямую через `gh`/`git` с `CODEXK8S_GIT_BOT_TOKEN`.

### `code-only`
- Агент работает только с репозиторием, сервисными API и документами без runtime-доступа к Kubernetes namespace задачи.
- Используется для продуктовых, документационных и governance-задач, которым не нужен live runtime-debug.

## Роли в delivery-цикле

- `pm`:
  - формализует проблему, scope, KPI, PRD и acceptance criteria;
  - синхронизирует продуктовые документы с delivery traceability.
- `sa`:
  - проектирует сервисные границы, контракты и ADR;
  - проверяет архитектурную консистентность решений.
- `em`:
  - управляет decomposition, quality gates и handover между stage.
- `dev`:
  - реализует изменения в коде;
  - обновляет тесты и docs, если меняется поведение.
- `reviewer`:
  - ищет баги, риски, регрессии и пробелы в tests/docs;
  - работает только review-комментариями в существующем PR.
- `qa`:
  - готовит acceptance/regression evidence;
  - проводит stage `run:qa` и revise-итерации QA.
- `sre`:
  - закрывает release/postdeploy/ops и аварийные recovery-сценарии.
- `km`:
  - ведет `run:doc-audit` и `run:self-improve`;
  - синхронизирует traceability и docs governance.

## Prompt templates: текущая MVP-модель

### Классы шаблонов
- `work` — выполнение задачи.
- `revise` — устранение замечаний по существующему PR/артефакту.

### Источник шаблонов
- В текущем MVP шаблоны берутся только из embed seed-каталога:
  `services/jobs/agent-runner/internal/runner/promptseeds/*.md`.
- Источник effective template в runtime/audit фиксируется как `repo_seed`.
- Поверх task-body рендерятся встроенные prompt-блоки из
  `services/jobs/agent-runner/internal/runner/templates/prompt_blocks/*.tmpl`:
  - role profile;
  - follow-up Issue contract;
  - PR/review/discussion contract.
- DB overrides, versioned prompt lifecycle в БД и UI-редактор шаблонов в MVP отсутствуют.

### Locale policy
- Worker определяет locale из platform default:
  `CODEXK8S_AGENT_DEFAULT_LOCALE`.
- Если значение пустое, используется `ru`.
- При рендере prompt envelope неподдерживаемое locale нормализуется к `en`.
- Базовый набор seed-локалей для системных ролей: `ru` и `en`.

### Что считается источником правды для prompt behavior
- repo seeds;
- `services.yaml/spec.projectDocs[]` для role-aware docs context;
- `services.yaml/spec.roleDocTemplates` для role-aware refs к шаблонам артефактов;
- `docs/architecture/prompt_templates_policy.md`.

## Review и revise loop

- Основной цикл артефактных stage:
  `run:<stage>` -> `state:in-review` -> review/comments -> `run:<stage>:revise` при необходимости.
- Для PR pre-review применяется отдельный trigger `need:reviewer` на PR.
- Для review-driven revise stage определяется детерминированно по policy resolver:
  `PR labels -> Issue labels -> run context -> flow_events`.
- Конфликтующие stage/model/reasoning labels трактуются как `failed_precondition`.

## Модель и reasoning profile

- Platform baseline:
  - model: `gpt-5.4`
  - reasoning: `high`
- На Issue/PR профиль может быть переопределен через:
  - `[ai-model-*]`
  - `[ai-reasoning-*]`
- Для revise-run effective profile перечитывается заново на каждом запуске.

## Что не входит в текущий MVP

- Staff UI/API для управления агентами.
- Staff UI/API для prompt templates.
- DB lifecycle prompt templates (`project override`, `global override`, version history).
- Runtime mode/locale settings per agent через UI/API.
- Массовые Agents UX operations.
- Custom-agent factory и self-service создание пользовательских ролей.

## Post-MVP направления

- custom-agent factory на проект;
- UI lifecycle prompt templates;
- richer locale management;
- project-scoped execution policies для пользовательских ролей;
- расширенные governance controls для roles/prompts/policies.

## Управление изменениями operating model

- Любое изменение roster/mode/prompt policy должно синхронно обновлять:
  - `docs/product/requirements_machine_driven.md`
  - `docs/architecture/data_model.md`
  - `docs/architecture/agent_runtime_rbac.md`
  - `docs/architecture/prompt_templates_policy.md`
  - `docs/delivery/requirements_traceability.md`
