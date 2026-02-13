---
doc_id: EPC-CK8S-S2-D4
type: epic
title: "Epic S2 Day 4: Agent job image, git workflow and PR creation"
status: completed
owner_role: EM
created_at: 2026-02-10
updated_at: 2026-02-13
related_issues: []
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-12-s2-day4"
---

# Epic S2 Day 4: Agent job image, git workflow and PR creation

## TL;DR
- Цель эпика: довести run до результата "создан PR" (для dev) и "обновлен PR" (для revise).
- Ключевая ценность: полный dogfooding цикл без ручного вмешательства.
- MVP-результат: агентный Job через MCP-ручки выполняет изменения, фиксирует их в git и ведет PR-цикл.

## Priority
- `P0`.

## Scope
### In scope
- Определить image/entrypoint агентного Job (инструменты: `@openai/codex`, `git`, `jq`, `curl`, `bash`, проектные toolchain `go`/`node`).
- Политика доступа и секретов:
  - агентный pod не получает прямые Kubernetes креды;
  - агентный pod получает отдельный bot-token только для git transport (commit/push в незащищенные ветки);
  - issue/pr/comments/labels и branch-context операции выполняются через MCP ручки (policy + audit);
  - в pod доступен только `CODEXK8S_OPENAI_API_KEY` и технич. параметры run.
- Policy шаблонов промптов:
  - `work`/`review` шаблоны для запуска берутся по приоритету `DB override -> repo seed`;
  - шаблоны выбираются по locale policy `project locale -> system default -> en`;
  - для системных агентов baseline заполняется минимум `ru` и `en`;
  - для run фиксируется effective template source/version/locale в аудит-контуре.
- Policy выбора модели и уровня рассуждения:
  - effective model/reasoning определяется по labels issue (`[ai-model-*]`, `[ai-reasoning-*]`) с fallback к настройкам агента/проекта;
  - для `run:dev:revise` параметры читаются повторно на момент запуска, чтобы Owner мог сменить модель/reasoning между итерациями ревью;
  - effective `model` и `reasoning_effort` пишутся в `agent_sessions`/`flow_events`.
- Resume policy:
  - сохранять `codex-cli` session JSON в `agent_sessions`;
  - при завершении каждого run пересохранять актуальный session snapshot;
  - если локальная session в pod отсутствует, перед `resume` восстановить ее из `agent_sessions.codex_cli_session_json`.
- PR flow:
  - детерминированное имя ветки;
  - создание/обновление PR с ссылкой на Issue;
  - запись PR URL/номер в БД.
- Dev/revise runtime orchestration:
  - `run:dev` запускает `work` prompt и ведет цикл до открытия PR;
  - `run:dev:revise` запускает `review` prompt, применяет фиксы в ту же ветку и обновляет существующий PR;
  - при отсутствии связанного PR в revise-режиме запуск отклоняется с явной диагностикой в `flow_events` (без автосоздания нового PR).

### Dependency from Day3.5
- Day3.5 предоставляет встроенный MCP tool layer (GitHub/Kubernetes) и prompt context assembler.
- Day4 не реализует собственный обходной доступ к GitHub/Kubernetes и использует только подготовленные Day3.5 MCP-контракты.

### Out of scope
- Автоматический final code review вместо Owner (финальный review остается за Owner).

## Критерии приемки эпика
- `run:dev` создает PR.
- `run:dev:revise` обновляет существующий PR; при отсутствии PR запуск отклоняется с диагностикой.
- В `flow_events` есть трасса: issue -> run -> namespace -> job -> pr.
- Агентный pod не содержит Kubernetes секретов и не получает GitHub governance credentials; governance write-действия проходят через MCP-контур.
- Агентный pod содержит только минимальный GitHub bot-token для git push-path; governance действия по GitHub/Kubernetes проходят через MCP-контур.

## Контекст и референсы реализации

Референсы из legacy-подхода (как источники механики, не как финальный дизайн):
- `../codexctl/internal/cli/prompt.go`
- `../codexctl/internal/prompt/templates/dev_issue_ru.tmpl`
- `../codexctl/internal/prompt/templates/dev_review_ru.tmpl`
- `../codexctl/internal/prompt/templates/config_default.toml`
- `../project-example/deploy/codex/Dockerfile`

Актуальные сведения по Codex (через Context7 и upstream docs):
- библиотека: `/openai/codex`
- SDK (`@openai/codex-sdk`) оборачивает CLI-бинарь `codex` (spawn + JSONL events);
- CLI resume/exec:
  - `codex resume --last`
  - `codex exec resume --last "<prompt>"`
- SDK resume:
  - восстановление thread из persisted данных в `~/.codex/sessions` через `resumeThread(...)`.

### Решение Day4: CLI-first, SDK-next

- Для Day4 фиксируется CLI-first подход:
  - текущий worker/control-plane стек написан на Go;
  - `@openai/codex-sdk` требует Node runtime-слой и все равно использует тот же CLI-бинарь и ту же persistence модель.
- SDK рассматривается как следующий шаг для richer event-stream/typed integrations в отдельном подпроекте, когда базовый cycle (`run:dev`/`run:dev:revise`) стабилизирован.

## Проектное решение Day4 (детализация)

### 1. Контур исполнения run

1. `issues.labeled` (`run:dev`/`run:dev:revise`) -> `agent_runs` + `flow_events`.
2. Worker claim -> runtime mode + namespace (из Day3 baseline).
3. Worker запускает agent Job в per-issue namespace.
4. Job:
   - подготавливает `codex` окружение и MCP-коннекторы;
   - резолвит effective prompt/config/model/reasoning;
   - при revise:
     - сначала пытается resume активной локальной session;
     - если локальной session нет, восстанавливает ее из БД snapshot в `~/.codex/sessions` и затем выполняет `resume`;
   - выполняет разработку/ревизию;
   - инициирует git/PR операции через MCP-инструменты GitHub (без прямого `gh auth` в pod).
5. Control-plane фиксирует результаты (PR link, branch, session snapshot refs, audit events).

### 2. Agent job image и entrypoint

Обязательное содержимое image:
- `@openai/codex` CLI;
- `git`, `jq`, `curl`, `bash`;
- базовые toolchains для проекта (`go`, `node`, при необходимости `python3`);
- runtime-конфиг для Codex через `~/.codex/config.toml`.

Требования к entrypoint:
- fail-fast по критическим ошибкам auth/session-restore/MCP connectivity;
- mask секретов в логах;
- структурированный stdout/stderr для последующего audit/парсинга.

### 3. Prompt/config pipeline

Политика источника prompt:
1. `project override` в БД;
2. `global override` в БД;
3. `repo seed` (`docs/product/prompt-seeds/<stage>-work.md`, `docs/product/prompt-seeds/<stage>-review.md`).

Seed usage:
- `prompt-seeds` используются как task-body шаблоны.
- В агентный runtime передается final prompt, собранный из:
  - system policy envelope,
  - runtime context,
  - MCP/tool context,
  - issue/pr context,
  - task-body (override/seed),
  - output contract.

Локаль prompt:
1. locale проекта;
2. system default locale;
3. fallback `en`.

Требования к рендеру:
- даже при ограниченном контексте передаются обязательные инструкции по:
  - source-of-truth документам,
  - правилам обновления документации,
  - требованиям к проверкам и PR.

### 4. Model/reasoning policy

- Источник конфигурации:
  1. labels issue (`[ai-model-*]`, `[ai-reasoning-*]`);
  2. project/agent defaults;
  3. system defaults.
- Для `run:dev:revise` effective параметры перечитываются на каждый запуск.
- В аудит пишутся:
  - source (`issue_label`/`agent_default`/`system_default`);
  - selected `model`;
  - selected `reasoning_effort`.

### 5. Auth и креды в Job

- Для `codex login` используется `CODEXK8S_OPENAI_API_KEY`:
  - `printenv CODEXK8S_OPENAI_API_KEY | codex login --with-api-key`
- Прямые Kubernetes креды в агентный pod не выдаются.
- В pod выдаётся только GitHub bot-token для git transport (`git fetch/pull/push` в незащищённые ветки).
- GitHub governance операции (issue/pr/comments/labels/branch context) и Kubernetes операции делаются через MCP-ручки approver/executor.
- Секреты не логируются и не пишутся в итоговые комментарии/PR body.

### 6. Session/resume стратегия

Обязательные принципы:
- после каждого run сохраняется codex session snapshot в БД (`agent_sessions.codex_cli_session_json` + metadata);
- snapshot обновляется при завершении каждого последующего run по той же issue/PR ветке;
- файловый слой Codex в контейнере (`~/.codex/sessions`) считается runtime-кешем;
- источником восстановления является запись в БД.

Поведение для `run:dev:revise`:
- если есть связанная успешная/активная сессия по текущему PR/issue -> resume;
- если локальной сессии нет, но есть БД snapshot -> восстановить runtime session из JSON и resume;
- если сессии нет, но PR существует -> новый `review` запуск в той же ветке;
- если PR не найден -> отклонить запуск с event `run.revise.pr_not_found`, статусом `failed_precondition` и рекомендацией использовать `run:dev`.

### 7. Branch/PR policy

Детерминированный naming:
- ветка: `codex/issue-<issue-number>` (опционально суффикс run-id при коллизии);
- commit messages: на английском, со ссылкой на Issue.

PR policy:
- `run:dev`: создать PR в `main` и связать с Issue (`Closes #<issue>`).
- `run:dev:revise`: обновить существующий PR в той же ветке.
- операции PR/labels/comments выполняются через MCP GitHub-инструменты.
- в БД/links фиксировать:
  - `issue -> run`,
  - `run -> branch`,
  - `run -> pr`.

## Детализация задач (Stories/Tasks)

### Story-1: Agent execution image
- Добавить отдельный Dockerfile/target для agent job runtime.
- Установить `@openai/codex` и обязательные утилиты.
- Согласовать переменные окружения (`CODEXK8S_OPENAI_API_KEY`, repo slug, issue/pr/run ids, MCP endpoint/token).

### Story-2: Prompt/config/model render and launch
- Реализовать резолв effective template (`work`/`review`, locale fallback) через Day3.5 context assembler.
- Реализовать резолв effective model/reasoning из labels с fallback.
- Рендерить `~/.codex/config.toml` перед запуском.
- Запускать:
  - dev: `codex exec "<work-prompt>" ...`
  - revise: `codex exec resume --last "<review-prompt>"` при наличии сессии.

### Story-3: Git/PR workflow via MCP
- Checkout/cd в рабочий repo.
- Детерминированно создавать/использовать ветку.
- Делать commit/push, создавать/обновлять PR через MCP GitHub ручки.
- Явно использовать policy metadata от Day3.5 (approval-required флаги и ограничения ролей/режимов).
- Писать PR ссылку/номер в БД и `flow_events`.

### Story-4: Session persistence and restore
- Сохранять session metadata и JSON snapshot в `agent_sessions`.
- Привязывать session к run/issue/PR.
- Реализовать восстановление `~/.codex/sessions` из БД snapshot и resume при `run:dev:revise`/перезапуске run.

### Story-5: Observability and audit
- Добавить события:
  - `run.agent.started`,
  - `run.agent.session.restored`,
  - `run.agent.session.saved`,
  - `run.pr.created`,
  - `run.pr.updated`,
  - `run.agent.resume.used`.
- Расширить payload audit-полями (branch, pr_number, session_id/thread_id, template source/locale/version, model/reasoning source/value).

## Тестовый контур приемки (обязательный)

Минимальный e2e сценарий Day4:
1. Создать Issue с задачей.
2. Поставить `run:dev`.
3. Проверить:
   - создана ветка,
   - в ветке появился тестовый/целевой файл с изменением,
   - создан PR, привязанный к Issue.
4. Добавить review-комментарий в PR.
5. Поменять label модели/рассуждений (при необходимости) и поставить `run:dev:revise`.
6. Проверить:
   - в ту же ветку добавлен фикс,
   - PR обновлен,
   - комментарий закрыт/адресован,
   - effective model/reasoning в run отражает актуальные labels.
7. Проверить аудит:
   - trace `issue -> run -> namespace -> job -> pr`,
   - session snapshot сохранен и связан с run;
   - при эмуляции потери локальной session восстановление из БД snapshot отрабатывает корректно.

## Риски и открытые вопросы

- Риск неполного/хрупкого resume при несовместимых версиях формата session snapshot.
- Риск зависимостей от стабильности MCP-контуров GitHub/Kubernetes на раннем этапе.
- Открытый выбор для long-term:
  - оставить CLI-first runtime как baseline;
  - или вынести control loop в отдельный SDK/app-server слой (без потери совместимости с CLI-сессиями).

## Фактическая реализация (2026-02-12)
- Добавлен отдельный runtime image `services/jobs/agent-runner/Dockerfile` и Go-бинарь `services/jobs/agent-runner/cmd/agent-runner`:
  - инструменты runtime устанавливаются проектным bootstrap-скриптом `services/jobs/agent-runner/scripts/bootstrap_tools.sh` (после установки `@openai/codex`);
  - baseline toolchain включает `go`, `protoc` + `protoc-gen-go`/`protoc-gen-go-grpc`, `oapi-codegen`, `openapi-ts`, `golangci-lint`, `dupl`, а также базовые утилиты (`git`, `jq`, `bash`);
  - governance write-действия по GitHub/Kubernetes выполняются через MCP policy/audit контур;
- В `control-plane` добавлены gRPC callback методы для agent-runner:
  - `UpsertAgentSession`,
  - `GetLatestAgentSession`,
  - `InsertRunFlowEvent`.
- Персистентность resumable session расширена до multi-agent сценария:
  - в `agent_sessions` добавлен `agent_key`;
  - latest-session lookup выполняется по `repository_full_name + branch_name + agent_key`;
  - revise-path восстанавливает только сессию соответствующего системного агента.
- В worker runtime-context протащен `agent_key` и `CODEXK8S_CONTROL_PLANE_GRPC_TARGET` в run job env.
- В `flow_events` сохранены события Day4:
  - `run.agent.started`,
  - `run.agent.session.restored`,
  - `run.agent.session.saved`,
  - `run.agent.resume.used`,
  - `run.pr.created`,
  - `run.pr.updated`,
  - `run.revise.pr_not_found`,
  - `run.failed.precondition`.
- В run pod реализована split access model:
  - прямых Kubernetes credentials нет;
  - GitHub/Kubernetes governance-операции выполняются через MCP;
  - для git transport path используется выделенный `CODEXK8S_GIT_BOT_TOKEN`.
- Обновлены CI/deploy/bootstrap pipeline:
  - добавлен компонент сборки `agent-runner`;
  - добавлены image/env/secret/variable (`CODEXK8S_AGENT_RUNNER_IMAGE`, `CODEXK8S_WORKER_RUN_CREDENTIALS_SECRET_NAME`, `CODEXK8S_AGENT_DEFAULT_*`, `CODEXK8S_AGENT_BASE_BRANCH`, `CODEXK8S_GIT_BOT_TOKEN`);
  - `CODEXK8S_WORKER_JOB_IMAGE` по умолчанию переключен на `agent-runner`.

### Актуализация baseline (post-Day4 hardening, 2026-02-12)
- Для agent pod закреплён direct execution path:
  - GitHub операции (issue/PR/comments/review + git push) выполняются через `gh`/`git` с `CODEXK8S_GIT_BOT_TOKEN`;
  - в `full-env` для namespace выдаётся `KUBECONFIG`, runtime-дебаг выполняется через `kubectl`.
- MCP-контур сокращён до label-операций (`github_labels_*`) для детерминированных transitions и аудита.
- Прямой доступ к Kubernetes `secrets` в run namespace запрещён RBAC; future secret-management через MCP+approver вынесен в последующие эпики.

### Label flow (S2 baseline + next)
- S2 baseline: `run:dev` и `run:dev:revise` остаются единственными активными trigger-labels для dev цикла.
- После выполнения run управление label/state выполняется через MCP-ручки, чтобы исключить гонки ручных и агентных изменений.
- Follow-up в плане:
  - Day5: детерминированные post-run label transitions для owner flow (`run:* -> state:*`),
  - Day6: policy/audit hardening для автоматических label transitions и конфликтов.
  - Day6+: вынести `mcp_servers.codex_k8s.tool_timeout_sec` в настраиваемую policy/runtime-конфигурацию (пер-run/пер-agent), чтобы поддержать долгие approver flow (часы) без ручных правок шаблона.
## Критерии приемки эпика — статус
- Выполнено: `run:dev` запускает агентный Job и поддерживает PR-flow.
- Выполнено: `run:dev:revise` работает с resume-path, при отсутствии PR отклоняется с `failed_precondition` и событием `run.revise.pr_not_found`.
- Выполнено: `flow_events` содержит трассу `issue -> run -> namespace -> job -> pr` с Day4 событиями агента.
- Выполнено: split access model введена (direct `gh`/`kubectl` в рамках выданных прав + MCP только для labels).
- Выполнено: для run namespace запрещён прямой доступ к Kubernetes `secrets` (read/write).
