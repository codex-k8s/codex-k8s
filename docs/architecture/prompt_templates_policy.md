---
doc_id: ARC-PRM-CK8S-0001
type: prompt-policy
title: "codex-k8s — Prompt Templates Policy"
status: draft
owner_role: SA
created_at: 2026-02-11
updated_at: 2026-02-13
related_issues: [1, 19]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
---

# Prompt Templates Policy

## TL;DR
- Поддерживаются два класса шаблонов: `work` и `review`.
- Каноническая модель шаблонов role-specific: отдельный body для каждого `agent_key` в каждой ветке `work/review`.
- Источник шаблона определяется по приоритету: project override в БД -> global override в БД -> seed в репозитории.
- Для каждого run фиксируется effective template version/hash для аудита и воспроизводимости.
- Шаблоны хранятся по локалям; выбор языка выполняется по цепочке project locale -> system default locale -> `en`.
- Контур `run:self-improve` использует тот же policy resolve, но дополняется обязательным блоком evidence-trace по источникам улучшений.

## Классы шаблонов

| Kind | Назначение | Пример seed |
|---|---|---|
| `work` | Выполнение задачи (plan/implement/test/doc update) | `docs/product/prompt-seeds/dev-work.md`, `docs/product/prompt-seeds/plan-work.md` |
| `review` | Ревизия/аудит изменений | `docs/product/prompt-seeds/dev-review.md`, `docs/product/prompt-seeds/plan-review.md` |

Примечание:
- seed-файлы в репозитории задают baseline-структуру и требования;
- effective prompt в рантайме формируется после resolve override в БД и контекстного рендера.

## Каноническая template-матрица

- Ключ шаблона: `(scope, role, kind, locale, version)`.
- Для каждого `agent_key` обязателен отдельный body-шаблон:
  - `kind=work`;
  - `kind=review`.
- Для каждого `(agent_key, kind)` обязательны минимум локали:
  - `ru`;
  - `en`.
- Использование одного общего body-шаблона для разных ролей не допускается.
- Stage-specific seed-файлы в репозитории являются bootstrap/fallback и не заменяют role-specific модель.

## Seed vs final prompt

- `docs/product/prompt-seeds/*.md` не отправляются агенту "как есть" в изоляции.
- Seed/override — это только task-body слой шаблона.
- В рантайме формируется final prompt, который включает:
  1. system policy envelope (правила безопасности, source-of-truth документы, формат результата);
  2. runtime context block;
  3. MCP capabilities block;
  4. issue/pr context block;
  5. task-body (DB override или repo seed);
  6. output contract block (проверки, PR/audit требования, learning mode при необходимости).

## Модель источников

Референс подхода к объёму и структуре шаблонов:
- `../codexctl/internal/prompt/templates/*.tmpl` (кроме `env_comment_*.tmpl`).

### Repo seeds
- Базовые stage-specific шаблоны в `docs/product/prompt-seeds/*.md`.
- Нейминг baseline: `<stage>-work.md` и `<stage>-review.md` (для revise-loop стадий).
- Используются как fallback при отсутствии override в БД по `(role, kind, locale)`.
- В seed-файлах хранится только prompt body (инструкции агенту).
- В seed-файлах не допускаются документные секции и мета-описания.

### DB global overrides
- Шаблоны уровня платформы.
- Применяются ко всем проектам, где нет project override.

### DB project overrides
- Шаблоны конкретного проекта.
- Имеют максимальный приоритет для соответствующего проекта/роли.

## Локали шаблонов

- Минимальный обязательный набор локалей для seed/bootstrapping: `ru` и `en`.
- Шаблон адресуется ключом:
  - `(scope, role, kind, locale, version)`.
- Правила fallback языка:
  1. locale проекта;
  2. default locale системы;
  3. `en`.

Планируемая эволюция:
- поддержка добавления новых локалей через staff UI/API;
- авто-перевод шаблонов через ИИ с сохранением как новой версии и возможностью ручной правки.

## Правила резолва effective template

1. Определить effective locale по policy fallback.
2. Найти активный project override по `(project, role, kind, locale)`.
3. Если нет, использовать активный global override по `(role, kind, locale)`.
4. Если нет, использовать repo seed для locale.
5. Если locale отсутствует в seed, использовать `en` seed.
6. В `agent_sessions.session_json` записать:
   - source (`project_override`/`global_override`/`repo_seed`),
   - template version/hash,
   - role/kind/locale.

## Контекстный рендер

- Перед выполнением шаблон рендерится с runtime-контекстом, включающим:
  - environment (`env`, namespace, slot, run/issue/pr identifiers),
  - доступные MCP servers/tools,
  - policy metadata по MCP-ручкам (approval-required, allowed scopes, denied actions),
  - project context и services overview,
  - effective model/reasoning profile (включая источник: issue labels или defaults),
  - режим исполнения агента (`full-env`/`code-only`) и feature flags.
- Минимальный обязательный набор runtime context полей:
  - run metadata: `run_id`, `correlation_id`, `project_id`, `agent_id`, `role_key`, `mode`;
  - repo/issue metadata: repo slug, issue number/title/body, labels, PR refs;
  - environment/services: namespace, сервисы проекта, основные endpoints, диагностические команды;
  - MCP catalog: серверы, инструменты, категории (read/write), approval policy;
  - template metadata: source/version/hash/locale, render context version.
- Формат контекста должен быть версионирован; изменения контракта рендера должны быть обратно совместимы либо сопровождаться миграцией шаблонов.

## Переходный профиль Day3.5 -> Day4

- Day3.5 формирует минимально полный runtime context для запуска Day4.
- На Day4 недопустимо убирать обязательные блоки:
  - source-of-truth документы и архитектурные ограничения;
  - runtime metadata (`run/issue/pr/namespace/mode`);
  - MCP catalog и policy flags;
  - требования по тестам/документации/PR flow;
  - правила безопасности (секреты, policy, аудит).
- Допускается отложить только необязательные расширенные поля (например, расширенную observability телеметрию), если это явно зафиксировано в `flow_events` как технический долг.
- Для `run:dev:revise` используется `review`-класс шаблонов, даже если запуск идет через resume-path.
- Для `run:dev:revise` effective model/reasoning перечитываются из актуальных issue labels перед запуском.
- Долг/план замены:
  - Day5: расширить наблюдаемость effective prompt/session/template metadata в UI;
  - Day6: синхронизировать prompt-контекст с MCP approval-flow и убрать временные упрощения.

## Политика `run:self-improve` для шаблонов

- Для `run:self-improve` используется `review`-класс шаблонов с расширенным output contract:
  - классификация findings (`docs`, `prompts`, `instructions`, `tools`);
  - ссылки на источники (`flow_events`, `agent_sessions`, PR/Issue comments);
  - proposal diff с оценкой риска.
- Repo seed baseline для этого контура:
  - `docs/product/prompt-seeds/self-improve-work.md`;
  - `docs/product/prompt-seeds/self-improve-review.md`.
- Изменения seed/override, внесённые через self-improve, проходят стандартный PR/review цикл.
- Для предотвращения drift:
  - каждый self-improve diff должен содержать traceable rationale;
  - шаблон не может ослаблять security/policy блоки final prompt.

## Требования безопасности и качества

- В шаблонах запрещены секреты, токены, приватные ключи и прямые credential-инструкции.
- Шаблон не должен обходить policy апрувов или ослаблять security ограничения.
- Изменения шаблонов проходят аудит и должны иметь трассировку в `links` и `flow_events`.

## Связанные документы
- `docs/product/agents_operating_model.md`
- `docs/product/prompt-seeds/README.md`
- `docs/product/prompt-seeds/dev-work.md`
- `docs/product/prompt-seeds/dev-review.md`
- `docs/architecture/data_model.md`
