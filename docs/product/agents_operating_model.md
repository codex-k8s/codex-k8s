---
doc_id: AGM-CK8S-0001
type: operating-model
title: "codex-k8s — Agents Operating Model"
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

# Agents Operating Model

## TL;DR
- Базовый штат платформы: 7 системных агентных ролей (`pm`, `sa`, `em`, `qa`, `sre`, `km`, `auditor`) + человек `Owner`.
- Для каждого проекта допускаются расширяемые custom-агенты без нарушения базовых ролей.
- Режим исполнения смешанный: часть ролей работает в `full-env`, часть в `code-only`.
- Шаблоны промптов для работы и ревью имеют seed в репозитории и override в БД.
- Для каждого системного агента шаблоны поддерживаются минимум в `ru` и `en`; язык выбирается по locale с fallback до `en`.

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
| `qa` | QA Lead | test strategy/plan/matrix/regression | `full-env` | 2 |
| `sre` | SRE/OPS | runbook/SLO/alerts/postdeploy | `full-env` | 1 |
| `km` | DocSet/KM | docset/index/traceability links | `code-only` | 2 |
| `auditor` | Doc/Checklist Auditor | аудит соответствия код↔доки | `full-env` (read-only) | 2 |

Примечания:
- `Owner` не является агентом, но остаётся финальным апрувером решений и trigger/deploy действий.
- Технический `run:dev` execution profile может быть реализован как системный или custom-агент проекта, но не заменяет базовый штат.

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
- Доступ: логи, events, сервисы, метрики; запись в Kubernetes только по разрешённой роли и политике.
- Используется для ролей, где нужно подтверждать решения по фактическому состоянию окружения.

### `code-only`
- Доступ только к репозиторию и API контексту без прямого Kubernetes runtime доступа.
- Используется для продуктовых/документационных задач без необходимости runtime-диагностики.

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

## Инфраструктурная модель и capacity baseline

- По умолчанию одновременно активен максимум 1 `run:dev` на issue.
- Лимиты параллелизма на проект задаются в настройках проекта и применяются worker-очередью.
- Для `full-env` запусков обязателен отдельный namespace на run/issue с контролем cleanup.

## Resume и timeout-policy

- Во время ожидания review от Owner run может быть приостановлен и возобновлён.
- Во время ожидания ответа MCP (`wait_state=mcp`) pod/run не должен завершаться по timeout.
- Для resumable выполнения в сессии сохраняется `codex-cli` session JSON, чтобы продолжать с того же места.

## Управление изменениями модели

- Изменения состава базового штата, execution modes или policy шаблонов фиксируются в:
  - `docs/product/requirements_machine_driven.md`,
  - `docs/architecture/data_model.md`,
  - `docs/architecture/agent_runtime_rbac.md`,
  - `docs/delivery/requirements_traceability.md`.
