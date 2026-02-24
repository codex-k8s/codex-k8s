---
doc_id: EPC-CK8S-S3-D15
type: epic
title: "Epic S3 Day 15: Prompt context overhaul (docs tree, role matrix, GitHub service messages)"
status: completed
owner_role: EM
created_at: 2026-02-13
updated_at: 2026-02-19
related_issues: [19]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-19-full-docset"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-19
---

# Epic S3 Day 15: Prompt context overhaul (docs tree, role matrix, GitHub service messages)

## TL;DR
- Цель: закрыть базовые пробелы core-flow перед финальным e2e: добавить декларативный docs-context в `services.yaml`, полностью пересобрать prompt templates по ролям/типам задач и унифицировать служебные GitHub-сообщения.
- Результат: агент получает полный и ролево-специфичный контекст (где он находится, что может делать, какие ресурсы доступны, какой порядок действий ожидается), а GitHub thread отражает run lifecycle через новые типизированные шаблоны сообщений.

## Priority
- `P0`.

## Контекст
- Текущий prompt-контур работает, но остаётся слишком универсальным и не покрывает профессиональные режимы (architect/devops/km/qa/reviewer и т.д.) с нужной глубиной.
- В `services.yaml` пока нет декларативного дерева проектной документации с ролями, которое можно безопасно и детерминированно подмешивать в prompt context.
- Служебные GitHub-комментарии (`run status`/conflict/операционные события) нужно привести к единому шаблонному каталогу с расширяемыми типами сообщений.
- Референс шаблонов для структуры и полноты: `/home/s/projects/codexctl/internal/prompt/templates/*.tmpl`.

## Scope
### In scope
- Расширение контракта `services.yaml`:
  - новый раздел project docs tree (`path`, `description`, `roles[]`, optional flags);
  - валидация путей, детект дублей, строгая typed-модель в `libs/go/servicescfg`.
- Prompt context assembler:
  - экспорт docs tree для runtime с фильтрацией по роли агента;
  - экспорт ролевых capabilities (k8s/github/mcp/tools) в структурированном виде.
- Новый каталог prompt templates:
  - role x trigger_kind x template_kind x locale;
  - fallback политика (role-specific -> stage-specific -> global default);
  - отдельные инструкции для developer/architect/devops/qa/reviewer/km/ops.
- Новый каталог GitHub service message templates:
  - run created/started/auth_required/auth_resolved/succeeded/failed;
  - trigger conflict, preflight diagnostics summary, deploy/build status summaries;
  - для slot/full-env запусков в сообщениях обязательно указывать `slot_url` (HTTPS link на домен слота) при наличии публичного host;
  - локализация RU/EN и единая модель данных для рендера.
- Обновление политики prompt templates:
  - актуализация `docs/architecture/prompt_templates_policy.md` и связанных docs с матрицей ролей.

### Out of scope
- Добавление новых ролей в продуктовую модель (используем текущий roster + уже поддержанные custom роли).
- Полный UI-редактор всех prompt templates (достаточно backend/runtime + seed policy + docs).

## Декомпозиция (Stories/Tasks)
- Story-1: `services.yaml` docs tree contract + schema + loader validation.
- Story-2: prompt context расширение (docs tree + role-aware capabilities).
- Story-3: полноразмерные prompt templates по ролям и типам задач (work/revise) + fallback matrix.
- Story-4: GitHub service messages templates v2 и подключение в `runstatus`/связанные use-cases.
  - Task: добавить в шаблоны run-сообщений обязательный блок `Slot URL`/`Ссылка на слот` для `full-env`/slot run, чтобы из issue/PR можно было открыть слот в один клик.
- Story-5: документация и трассируемость (policy + delivery docs + traceability updates).

## Статус выполнения (2026-02-19)
- Story-1: выполнено.
  - `libs/go/servicescfg`: добавлен `spec.projectDocs[]` (`path`, `description`, `roles[]`, `optional`) в typed model + JSON schema + валидация путей/дублей/ролей.
- Story-2: выполнено.
  - `services/internal/control-plane/internal/domain/mcp`: prompt context расширен блоками `role` + `docs` (role-aware filtering).
  - `services/jobs/agent-runner/internal/runner`: role-aware docs context подмешивается в итоговый prompt envelope (с лимитом refs для контроля размера).
- Story-3: выполнено.
  - Добавлена role-aware матрица выбора prompt-seed с fallback: `stage+role -> role -> stage -> default`.
  - Добавлены role templates `work/revise` для поддержанных ролей (`dev/pm/sa/em/reviewer/qa/sre/km`) в локалях `ru/en`.
- Story-4: выполнено.
  - `runstatus` переведен на v2 lifecycle phases (`created`, `started`, `auth_required`, `auth_resolved`, `finished`, `namespace_deleted`).
  - В run-сообщения добавлен `Slot URL` для full-env run (по runtime host или расчету из namespace + `CODEXK8S_AI_DOMAIN`/`CODEXK8S_PRODUCTION_DOMAIN`).
- Story-5: выполнено.
  - Обновлены sprint/traceability/policy связанные документы и статусы выполнения Day15.

## Критерии приемки
- `services.yaml` поддерживает декларативный docs tree с `path`, `description`, `roles[]`; контракт валидируется typed loader и тестами.
- В prompt context присутствуют docs refs и role-aware блоки, которые реально используются в итоговом prompt.
- Для каждой поддержанной роли есть полноценный шаблон (RU/EN) минимум для `work` и `review`.
- GitHub служебные сообщения рендерятся из нового шаблонного каталога и покрывают run lifecycle.
- Для `full-env`/slot run GitHub-сообщения содержат кликабельную ссылку на домен слота (если host успешно резолвлен runtime deploy).
- Обновлены policy/docs с описанием новой матрицы prompt/service messages.

## Риски/зависимости
- Риск слишком большого prompt payload: нужен лимит/обрезка docs refs и приоритизация по роли.
- Риск дрейфа между template data model и runtime payload: нужен typed DTO контракт и golden tests.
- Зависимость от согласования минимально обязательного набора ролей для первой итерации.
