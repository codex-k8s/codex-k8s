---
doc_id: ARC-PRM-CK8S-0001
type: prompt-policy
title: "codex-k8s — Prompt Templates Policy"
status: active
owner_role: SA
created_at: 2026-02-11
updated_at: 2026-02-21
related_issues: [1, 19, 100]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-19-full-docset"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-19
---

# Prompt Templates Policy

## TL;DR
- Поддерживаются два класса шаблонов: `work` и `revise`.
- Каноническая модель шаблонов role-specific: отдельный body для каждого `agent_key` в каждой ветке `work/revise`.
- `services.yaml` поддерживает repo-aware `spec.projectDocs[]` (`repository`, `path`, `description`, `roles[]`, `optional`) для role-aware docs context в multi-repo проектах.
- `services.yaml` поддерживает `spec.roleDocTemplates` (role -> list of template paths) для role-aware подсказок по шаблонам артефактов в prompt envelope.
- Источник шаблона определяется по приоритету: project override в БД -> global override в БД -> seed в репозитории.
- Для каждого run фиксируется effective template version/hash для аудита и воспроизводимости.
- Шаблоны хранятся по локалям; выбор языка выполняется по цепочке project locale -> system default locale -> `en`.
- Контур `run:self-improve` использует тот же policy resolve, но дополняется обязательным блоком evidence-trace по источникам улучшений.

## Классы шаблонов

| Kind | Назначение | Пример seed |
|---|---|---|
| `work` | Выполнение задачи (plan/implement/test/doc update) | `services/jobs/agent-runner/internal/runner/promptseeds/dev-work.md`, `services/jobs/agent-runner/internal/runner/promptseeds/plan-work.md` |
| `revise` | Устранение замечаний Owner к существующему PR | `services/jobs/agent-runner/internal/runner/promptseeds/dev-revise.md`, `services/jobs/agent-runner/internal/runner/promptseeds/plan-revise.md` |

Примечание:
- seed-файлы в репозитории задают baseline-структуру и требования;
- effective prompt в рантайме формируется после resolve override в БД и контекстного рендера.

## Каноническая template-матрица

- Ключ шаблона: `(scope, role, kind, locale, version)`.
- Для каждого `agent_key` обязателен отдельный body-шаблон:
  - `kind=work`;
  - `kind=revise`.
- Для каждого `(agent_key, kind)` обязательны минимум локали:
  - `ru`;
  - `en`.
- Использование одного общего body-шаблона для разных ролей не допускается.
- Stage-specific seed-файлы в репозитории являются bootstrap/fallback и не заменяют role-specific модель.

Реализованный seed fallback (runtime):
1. `stage-role-kind_locale` (например: `design-sa-revise_ru.md`);
2. `stage-role-kind`;
3. `role-role-kind_locale` (например: `role-sa-revise_ru.md`);
4. `role-role-kind`;
5. `stage-kind_locale`;
6. `stage-kind`;
7. `dev-kind_locale`;
8. `dev-kind`;
9. `default-kind_locale`;
10. `default-kind`;
11. встроенные fallback templates runner-а.

## Seed vs final prompt

- `services/jobs/agent-runner/internal/runner/promptseeds/*.md` не отправляются агенту "как есть" в изоляции.
- Seed/override — это только task-body слой шаблона.
- В рантайме формируется final prompt, который включает:
  1. system policy envelope (правила безопасности, source-of-truth документы, формат результата);
  2. runtime context block;
  3. MCP capabilities block;
  4. issue/pr context block;
  5. task-body (DB override или repo seed);
  6. output contract block (проверки, PR/audit требования, learning mode при необходимости).

Для output contract block (обязательное правило baseline):
- указывается communication language (из effective locale запуска), обязательный для:
  - PR title/body/comments;
  - issue/PR replies;
  - feedback-инструментов MCP.
- добавляется требование регулярного progress-feedback через MCP tool `run_status_report`
  (минимум 1 вызов после каждых 5-7 вызовов других инструментов, плюс перед долгими операциями/ожиданием).

## Модель источников

Референс подхода к объёму и структуре шаблонов:
- `../codexctl/internal/prompt/templates/*.tmpl` (кроме `env_comment_*.tmpl`).

### Repo seeds
- Базовые stage-specific шаблоны в `services/jobs/agent-runner/internal/runner/promptseeds/*.md`.
- Нейминг baseline: `<stage>-work.md` и `<stage>-revise.md` (для revise-loop стадий).
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
  - repo/issue metadata: primary repo slug, repository graph (aliases + refs), issue number/title/body, labels, PR refs;
  - environment/services: namespace, сервисы проекта, основные endpoints, диагностические команды;
  - MCP catalog: серверы, инструменты, категории (read/write), approval policy;
  - template metadata: source/version/hash/locale, render context version.
- Дополнительно для Day15:
  - role profile (display name + capability areas);
  - role-aware docs refs из `services.yaml/spec.projectDocs[]`;
  - role-aware template refs из `services.yaml/spec.roleDocTemplates`;
  - лимит на количество docs refs в final prompt (для контроля размера prompt payload).
- Формат контекста должен быть версионирован; изменения контракта рендера должны быть обратно совместимы либо сопровождаться миграцией шаблонов.

### Multi-repo docs federation (Issue #100, design)
- В multi-repo проекте каждый docs ref в `spec.projectDocs[]` обязан указывать `repository` (alias из project repositories topology).
- Для monorepo режимов `repository` может быть опущен (используется primary/trigger repository).
- При формировании docs context применяется deterministic priority:
  1. policy/docs repository;
  2. orchestrator repository;
  3. service repositories.
- Для каждой записи сохраняется resolved commit SHA источника docs, чтобы prompt context был воспроизводим.
- При недоступности `optional: true` источника контекст строится без hard-fail; для `optional: false` возвращается `failed_precondition`.

## Переходный профиль Day3.5 -> Day4

- Day3.5 формирует минимально полный runtime context для запуска Day4.
- На Day4 недопустимо убирать обязательные блоки:
  - source-of-truth документы и архитектурные ограничения;
  - runtime metadata (`run/issue/pr/namespace/mode`);
  - MCP catalog и policy flags;
  - требования по тестам/документации/PR flow;
  - правила безопасности (секреты, policy, аудит).
- Допускается отложить только необязательные расширенные поля (например, расширенную observability телеметрию), если это явно зафиксировано в `flow_events` как технический долг.
- Для `run:dev:revise` используется `revise`-класс шаблонов, даже если запуск идет через resume-path.
- Для `run:dev:revise` effective model/reasoning перечитываются из актуальных issue labels перед запуском.
- Долг/план замены:
  - Day5: расширить наблюдаемость effective prompt/session/template metadata в UI;
  - Day6: синхронизировать prompt-контекст с MCP approval-flow и убрать временные упрощения.

## Политика `run:self-improve` для шаблонов

- Для `run:self-improve` используется `work`-класс шаблонов с расширенным output contract:
  - классификация findings (`docs`, `prompts`, `instructions`, `tools`);
  - ссылки на источники (`flow_events`, `agent_sessions`, PR/Issue comments);
  - proposal diff с оценкой риска.
- В `run:self-improve` prompt-body обязан закреплять диагностический MCP-процесс:
  - `self_improve_runs_list` (50/page newest-first);
  - `self_improve_run_lookup` (поиск run по Issue/PR);
  - `self_improve_session_get` (получение `codex-cli` session JSON и путь под `/tmp/codex-sessions/...`).
- Перед анализом session JSON prompt обязан требовать создание целевого каталога `/tmp/codex-sessions/<run-id>`.
- Repo seed baseline для этого контура:
  - `services/jobs/agent-runner/internal/runner/promptseeds/self-improve-work.md`.
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
- `services/jobs/agent-runner/internal/runner/promptseeds/README.md`
- `services/jobs/agent-runner/internal/runner/promptseeds/dev-work.md`
- `services/jobs/agent-runner/internal/runner/promptseeds/dev-revise.md`
- `docs/architecture/data_model.md`
