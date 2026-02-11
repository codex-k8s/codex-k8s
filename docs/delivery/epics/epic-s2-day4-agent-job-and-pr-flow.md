---
doc_id: EPC-CK8S-S2-D4
type: epic
title: "Epic S2 Day 4: Agent job image, git workflow and PR creation"
status: planned
owner_role: EM
created_at: 2026-02-10
updated_at: 2026-02-12
related_issues: []
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
---

# Epic S2 Day 4: Agent job image, git workflow and PR creation

## TL;DR
- Цель эпика: довести run до результата “создан PR” (для dev) и “обновлён PR” (для revise).
- Ключевая ценность: полный dogfooding цикл без ручного вмешательства.
- MVP-результат: агентный Job клонирует repo, вносит изменения, пушит ветку и открывает PR.

## Priority
- `P0`.

## Scope
### In scope
- Определить image/entrypoint агентного Job (инструменты: `git`, `gh`, `kubectl`, `@openai/codex`, `go`/`node` по проектным потребностям).
- Политика кредов:
  - repo token берётся из БД (шифрованно), расшифровывается в control-plane и прокидывается в Job безопасно;
  - исключить попадание токенов в логи.
- Policy шаблонов промптов:
  - `work`/`review` шаблоны для запуска берутся по приоритету `DB override -> repo seed`;
  - шаблоны выбираются по locale policy `project locale -> system default -> en`;
  - для системных агентов baseline заполняется минимум `ru` и `en`;
  - для run фиксируется effective template source/version/locale в аудит-контуре.
- Resume policy:
  - сохранять `codex-cli` session JSON в `agent_sessions`;
  - при перезапуске/возобновлении run восстанавливать сессию с того же места.
- PR flow:
  - детерминированное имя ветки;
  - создание PR с ссылкой на Issue;
  - запись PR URL/номер в БД.
- Dev/revise runtime orchestration:
  - `run:dev` запускает `work` prompt и ведёт цикл до открытия PR;
  - `run:dev:revise` запускает `review` prompt, применяет фиксы в ту же ветку и обновляет существующий PR;
  - при отсутствии связанного PR в revise-режиме запуск отклоняется с явной диагностикой в `flow_events` (без автосоздания нового PR).

### Out of scope
- Автоматический code review (финальный ревью остаётся за Owner).
- Полный MCP-driven контроль всех runtime write-операций (переносится в Day6).

## Критерии приемки эпика
- `run:dev` создаёт PR.
- `run:dev:revise` обновляет существующий PR; при отсутствии PR запуск отклоняется с диагностикой.
- В `flow_events` есть трасса: issue -> run -> namespace -> job -> pr.

## Контекст и референсы реализации

Референсы из legacy-подхода (как источники механики, не как финальный дизайн):
- `../codexctl/internal/cli/prompt.go`
- `../codexctl/internal/prompt/templates/dev_issue_ru.tmpl`
- `../codexctl/internal/prompt/templates/dev_review_ru.tmpl`
- `../codexctl/internal/prompt/templates/config_default.toml`
- `../project-example/deploy/codex/Dockerfile`

Актуальные сведения по Codex (через Context7):
- библиотека: `/openai/codex`
- CLI resume/exec:
  - `codex resume --last`
  - `codex exec resume --last "<prompt>"`
- SDK resume:
  - восстановление thread из persisted данных в `~/.codex/sessions` через `resumeThread(...)`.

## Проектное решение Day4 (детализация)

### 1. Контур исполнения run

1. `issues.labeled` (`run:dev`/`run:dev:revise`) -> `agent_runs` + `flow_events`.
2. Worker claim -> runtime mode + namespace (из Day3 baseline).
3. Worker запускает agent Job в per-issue namespace.
4. Job:
   - подготавливает `git`/`gh`/`codex` окружение;
   - рендерит effective prompt и `~/.codex/config.toml`;
   - выполняет `codex exec ...` (для dev) или `codex exec resume --last ...` (для revise при наличии сессии);
   - выполняет commit/push/PR операции через `gh`.
5. Control-plane фиксирует результаты (PR link, branch, session snapshot refs, audit events).

### 2. Agent job image и entrypoint

Обязательное содержимое image:
- `@openai/codex` CLI;
- `git`, `gh`, `kubectl`, `jq`, `curl`, `bash`;
- базовые toolchains для проекта (`go`, `node`, при необходимости `python3`);
- runtime-конфиг для Codex через `~/.codex/config.toml`.

Требования к entrypoint:
- fail-fast по критическим ошибкам auth/checkout/push;
- mask секретов в логах;
- структурированный stdout/stderr для последующего audit/парсинга.

### 3. Prompt/config pipeline

Политика источника prompt:
1. `project override` в БД;
2. `global override` в БД;
3. `repo seed` (`docs/product/prompt-seeds/dev-work.md`, `docs/product/prompt-seeds/dev-review.md`).

Локаль prompt:
1. locale проекта;
2. system default locale;
3. fallback `en`.

Требования к рендеру:
- даже при ограниченном контексте передаются обязательные инструкции по:
  - source-of-truth документам,
  - правилам обновления документации,
  - требованиям к проверкам и PR.

### 4. Auth и креды в Job

- Для `codex login` используется `CODEXK8S_OPENAI_API_KEY`:
  - `printenv CODEXK8S_OPENAI_API_KEY | codex login --with-api-key`
- Для GitHub операций используется расшифрованный repo token из БД:
  - auth через `gh auth login --with-token`.
- Секреты не логируются и не пишутся в итоговые комментарии/PR body.

### 5. Session/resume стратегия

Обязательные принципы:
- после каждого run сохраняется codex session snapshot в БД (`agent_sessions.codex_cli_session_json` + metadata);
- для resume используется persisted session/thread identity;
- файловый слой Codex в контейнере (`~/.codex/sessions`) рассматривается как runtime-source, но источником восстановления в платформе является запись в БД.

Поведение для `run:dev:revise`:
- если есть связанная успешная/активная сессия по текущему PR/issue -> resume;
- если сессии нет, но PR существует -> новый `review` запуск в той же ветке;
- если PR не найден -> отклонить запуск с event `run.revise.pr_not_found`, статусом `failed_precondition` и рекомендацией использовать `run:dev`.

### 6. Branch/PR policy

Детерминированный naming:
- ветка: `codex/issue-<issue-number>` (опционально суффикс run-id при коллизии);
- commit messages: на английском, со ссылкой на Issue.

PR policy:
- `run:dev`: создать PR в `main` и связать с Issue (`Closes #<issue>`).
- `run:dev:revise`: обновить существующий PR в той же ветке.
- в БД/links фиксировать:
  - `issue -> run`,
  - `run -> branch`,
  - `run -> pr`.

## Временные решения Day4 и план замены

Временное решение (Day4):
- до внедрения MCP approval-flow агентный контейнер содержит `gh` и `kubectl` для самостоятельного дебага/доработок в границах своего namespace.
- write-операции выполняются напрямую в рамках runtime job.

План замены:
- Day6: перевести привилегированные runtime write-операции на MCP approver/executor flow, ограничив прямые write-действия агента.
- Day7: закрепить регрессией, что direct-write path выключен/ограничен согласно policy.

## Детализация задач (Stories/Tasks)

### Story-1: Agent execution image
- Добавить отдельный Dockerfile/target для agent job runtime.
- Установить `@openai/codex`, `gh`, `kubectl` и обязательные утилиты.
- Согласовать переменные окружения (`CODEXK8S_OPENAI_API_KEY`, repo token, repo slug, issue/pr/run ids).

### Story-2: Prompt/config render and launch
- Реализовать резолв effective template (`work`/`review`, locale fallback).
- Рендерить `~/.codex/config.toml` перед запуском.
- Запускать:
  - dev: `codex exec "<work-prompt>" ...`
  - revise: `codex exec resume --last "<review-prompt>"` при наличии сессии.

### Story-3: Git/PR workflow
- Checkout/cd в рабочий repo.
- Детерминированно создавать/использовать ветку.
- Делать commit/push, создавать/обновлять PR через `gh`.
- Писать PR ссылку/номер в БД и `flow_events`.

### Story-4: Session persistence
- Сохранять session metadata и JSON snapshot в `agent_sessions`.
- Привязывать session к run/issue/PR.
- Реализовать восстановление при `run:dev:revise`/перезапуске run.

### Story-5: Observability and audit
- Добавить события:
  - `run.agent.started`,
  - `run.agent.session.saved`,
  - `run.pr.created`,
  - `run.pr.updated`,
  - `run.agent.resume.used`.
- Расширить payload audit-полями (branch, pr_number, session_id/thread_id, template source/locale/version).

## Тестовый контур приемки (обязательный)

Минимальный e2e сценарий Day4:
1. Создать Issue с задачей.
2. Поставить `run:dev`.
3. Проверить:
   - создана ветка,
   - в ветке появился тестовый/целевой файл с изменением,
   - создан PR, привязанный к Issue.
4. Добавить review-комментарий в PR.
5. Поставить `run:dev:revise`.
6. Проверить:
   - в ту же ветку добавлен фикс,
   - PR обновлён,
   - комментарий закрыт/адресован.
7. Проверить аудит:
   - trace `issue -> run -> namespace -> job -> pr`,
   - session snapshot сохранён и связан с run.

## Риски и открытые вопросы

- Риск несовместимости prompt-структуры и реального runtime-контекста на раннем этапе.
- Риск неполного/хрупкого resume при рестартах pod.
- Риск избыточных прав в Day4 (временное решение до Day6).
- Открытый выбор для long-term:
  - оставить CLI-first путь;
  - или перейти на SDK/app-server control loop (спайк/решение фиксируется отдельно после Day4).
