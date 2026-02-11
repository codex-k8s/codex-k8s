---
doc_id: ARC-PRM-CK8S-0001
type: prompt-policy
title: "codex-k8s — Prompt Templates Policy"
status: draft
owner_role: SA
created_at: 2026-02-11
updated_at: 2026-02-11
related_issues: [1]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
---

# Prompt Templates Policy

## TL;DR
- Поддерживаются два класса шаблонов: `work` и `review`.
- Источник шаблона определяется по приоритету: project override в БД -> global override в БД -> seed в репозитории.
- Для каждого run фиксируется effective template version/hash для аудита и воспроизводимости.
- Шаблоны хранятся по локалям; выбор языка выполняется по цепочке project locale -> system default locale -> `en`.

## Классы шаблонов

| Kind | Назначение | Пример seed |
|---|---|---|
| `work` | Выполнение задачи (plan/implement/test/doc update) | `docs/product/prompt-seeds/dev-work.md` |
| `review` | Ревизия/аудит изменений | `docs/product/prompt-seeds/dev-review.md` |

## Модель источников

Референс подхода к объёму и структуре шаблонов:
- `../codexctl/internal/prompt/templates/*.tmpl` (кроме `env_comment_*.tmpl`).

### Repo seeds
- Базовые шаблоны в репозитории.
- Используются как fallback при отсутствии override в БД.

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
  - project context и services overview,
  - режим исполнения агента (`full-env`/`code-only`) и feature flags.
- Формат контекста должен быть версионирован; изменения контракта рендера должны быть обратно совместимы либо сопровождаться миграцией шаблонов.

## Требования безопасности и качества

- В шаблонах запрещены секреты, токены, приватные ключи и прямые credential-инструкции.
- Шаблон не должен обходить policy апрувов или ослаблять security ограничения.
- Изменения шаблонов проходят аудит и должны иметь трассировку в `links` и `flow_events`.

## Связанные документы
- `docs/product/agents_operating_model.md`
- `docs/product/prompt-seeds/dev-work.md`
- `docs/product/prompt-seeds/dev-review.md`
- `docs/architecture/data_model.md`
