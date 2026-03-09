---
doc_id: EPC-CK8S-S8-D2-EMPTY-REPO
type: epic
title: "Epic S8 Day 2: Empty repository initialization on onboarding (services.yaml + docs scaffold) (Issue 281)"
status: planned
owner_role: EM
created_at: 2026-03-06
updated_at: 2026-03-06
related_issues: [281]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-06-s8-empty-repo-init"
---

# Epic S8 Day 2: Empty repository initialization on onboarding (services.yaml + docs scaffold)

## TL;DR
- Проблема: текущий onboarding умеет проверить репозиторий (`preflight`), но не умеет довести пустой GitHub-репозиторий до состояния, в котором его уже можно вести через `codex-k8s`.
- Цель: при добавлении пустого репозитория на платформу автоматически инициализировать его минимальным bootstrap-пакетом `services.yaml + docs`, чтобы дальше можно было запускать полноценный stage-flow.
- Результат: deterministic empty-repo bootstrap с audit trail, статусом в staff UI и безопасной идемпотентной повторной инициализацией.

## Priority
- `P0`.

## Контекст и проблема
- FR-020 допускает несколько репозиториев на проект, но текущий продуктовый контур закрывает только attach/preflight существующего репозитория.
- FR-049 покрывает проверку доступов и GitHub-операций, но не покрывает фактическую инициализацию содержимого репозитория.
- Для пустого репозитория нельзя опереться на стандартный PR-flow как на обязательный механизм инициализации:
  - у пустого GitHub-репозитория может отсутствовать default branch и база для PR;
  - человеку всё равно нужен детерминированный bootstrap, а не ручное создание `services.yaml` и начального набора docs.
- В итоге новый проект нельзя быстро завести в платформу “с нуля” без ручной подготовки репозитория вне `codex-k8s`.

## Scope
### In scope
- Детект пустого репозитория:
  - нет ни одного commit;
  - отсутствует default branch;
  - либо default branch существует, но tree пустой и в репозитории нет ни `services.yaml`, ни project docs baseline.
- Новый onboarding-state `empty_repository` / `requires_initialization` в staff API/UI.
- Initial bootstrap pack для пустого репозитория:
  - создание default branch (по policy проекта, default `main`);
  - initial commit в репозиторий;
  - минимальный `services.yaml`, совместимый с typed contract платформы;
  - полный template-catalog `docs/templates/*`;
  - стартовый docs baseline, сгенерированный из шаблонов платформы и снабжённый понятными TODO/next-step маркерами.
- Template-catalog, который должен появиться в репозитории при bootstrap, обязан включать весь текущий набор `docs/templates/*.md` платформы:
  - индекс: `index.md`;
  - product templates: `problem.md`, `brief.md`, `constraints.md`, `scope_mvp.md`, `personas.md`, `risks_register.md`, `project_charter.md`, `success_metrics.md`, `prd.md`, `nfr.md`, `user_story.md`;
  - architecture templates: `c4_context.md`, `c4_container.md`, `adr.md`, `alternatives.md`, `design_doc.md`, `api_contract.md`, `data_model.md`, `migrations_policy.md`;
  - delivery templates: `delivery_plan.md`, `epic.md`, `definition_of_done.md`, `issue_map.md`, `roadmap.md`, `docset_issue.md`, `docset_pr.md`;
  - quality/release/ops templates: `test_strategy.md`, `test_plan.md`, `test_matrix.md`, `regression_checklist.md`, `release_plan.md`, `release_notes.md`, `rollback_plan.md`, `postdeploy_review.md`, `runbook.md`, `monitoring.md`, `alerts.md`, `slo.md`, `incident_playbook.md`, `incident_postmortem.md`.
- Стартовый docs baseline, который генерируется поверх template-catalog, должен включать как минимум:
  - `docs/README.md` с описанием структуры и следующего шага;
  - `docs/product/problem.md`;
  - `docs/product/brief.md`;
  - `docs/product/constraints.md`;
  - `docs/product/scope_mvp.md`;
  - `docs/product/personas.md`;
  - `docs/product/risks_register.md`;
  - `docs/architecture/README.md`;
  - `docs/delivery/README.md`;
  - `docs/ops/README.md`.
- UI flow:
  - preview будущего bootstrap-пакета;
  - явная команда `Initialize repository`;
  - статус `not_initialized -> initializing -> initialized -> failed`.
- Audit/traceability:
  - фиксировать источник шаблонов, commit SHA bootstrap-коммита и actor;
  - сохранять structured result в БД и показывать его в staff UI.
- Idempotent repair path:
  - повторный запуск не должен дублировать файлы;
  - при уже существующих bootstrap-файлах платформа должна сравнить expected vs actual и перейти либо в `initialized`, либо в `drift_detected`.
- Human follow-up:
  - после успешной инициализации платформа создаёт onboarding summary issue в репозитории с описанием, что было создано и какой следующий stage рекомендуется (`run:intake`/`run:plan`).

### Out of scope
- Генерация production-ready кода сервисов или полноценного application scaffold.
- Автоматический запуск `run:dev` сразу после инициализации.
- Автоматическое создание инфраструктурных манифестов под конкретный стек проекта.
- Импорт внешнего docset в момент инициализации (используется отдельный flow Day12).

## Выбранный execution-подход
1. Репозиторий проходит attach + preflight.
2. Платформа детектирует, что репозиторий пустой.
3. Оператор в staff UI запускает `Initialize repository`.
4. Control-plane собирает deterministic bootstrap bundle из platform templates.
5. Для пустого репозитория initial commit выполняется напрямую в default branch, потому что PR-base ещё не существует.
6. После initial commit платформа создаёт summary issue в самом репозитории, чтобы человек видел результат и мог сразу продолжить stage-flow.
7. Репозиторий получает статус `initialized`, и дальнейшая работа уже идёт через стандартные Issue/label процессы.

## Декомпозиция (Stories/Tasks)
- Story-1: Empty-repository detector и onboarding-state machine.
  - Явно различать `empty`, `already_initialized`, `drifted_bootstrap`, `non_empty_repository`.
  - Не считать репозиторий пустым, если в нём уже есть пользовательский код или docs baseline.
- Story-2: Bootstrap bundle generator.
  - Сгенерировать минимальный `services.yaml`.
  - Скопировать в репозиторий полный template-catalog `docs/templates/*.md`.
  - Сгенерировать стартовый docs baseline из `docs/templates/**`/репозиторных seed-шаблонов платформы.
  - Все generated файлы должны содержать traceable header/markers, по которым можно понять, что это bootstrap-слой.
- Story-3: GitHub write path для empty repo.
  - Создание default branch и initial commit.
  - Безопасная обработка race-condition, если ветка появилась между проверкой и записью.
  - Повторный запуск должен быть идемпотентным.
- Story-4: Staff UX.
  - Отдельный CTA для инициализации пустого репозитория.
  - Preview списка создаваемых файлов и статуса операции.
  - Отчёт об ошибке с actionable текстом.
- Story-5: Traceability и audit.
  - БД-статус bootstrap;
  - summary issue в GitHub;
  - flow events для старта/успеха/ошибки/repair.
- Story-6: Tests.
  - Unit: detector, template rendering, idempotency logic.
  - Integration: GitHub branch/init path на тестовом репозитории.

## Quality gates
- QG-S8-D2-01 Detection gate:
  - empty repo детектируется детерминированно и не конфликтует с existing-repo onboarding.
- QG-S8-D2-02 Bootstrap gate:
  - generated `services.yaml` и docs pack валидны и повторяемы.
- QG-S8-D2-03 GitHub gate:
  - initial commit path работает без PR-base и не ломается при повторном запуске.
- QG-S8-D2-04 UX gate:
  - оператор видит, почему репозиторий требует именно initialization, а не normal adoption.
- QG-S8-D2-05 Traceability gate:
  - bootstrap SHA, generated files и follow-up issue фиксируются в БД и UI.

## Критерии приёмки
- При добавлении пустого репозитория платформа явно показывает статус `requires_initialization`.
- После запуска initialization в репозитории появляются:
  - default branch;
  - `services.yaml`;
  - полный каталог `docs/templates/*.md`;
  - docs pack baseline.
- Bootstrap является идемпотентным:
  - повторный запуск не создаёт дубли;
  - drift детектируется и показывается как отдельный статус.
- Платформа создаёт summary issue c описанием bootstrap-результата и следующего рекомендуемого шага.
- Оператор может увидеть в UI commit SHA initial bootstrap и список созданных файлов.

## Риски и допущения
- Пустой GitHub-репозиторий не поддерживает обычный PR-review flow до появления первой ветки/коммита; это осознанное исключение, поэтому bootstrap идёт напрямую в default branch.
- Нужно жёстко контролировать состав bootstrap-пакета, чтобы не засорять новый репозиторий избыточными файлами.
- При смене шаблонов платформы требуется policy, как repair path работает для уже инициализированных репозиториев.

## Handover
- Next stage: `run:dev`.
- Execution issue: `281`.
- Handover package:
  - этот epic;
  - `docs/delivery/epics/s3/epic-s3-day14-repository-onboarding-preflight.md`;
  - `docs/delivery/epics/s3/epic-s3-day12-docset-import-and-safe-sync.md`;
  - `docs/delivery/epics/s3/epic-s3-day9-declarative-full-env-deploy-and-runtime-parity.md`;
  - `docs/architecture/prompt_templates_policy.md`;
  - `docs/architecture/adr/ADR-0007-multi-repo-composition-and-docs-federation.md`.
