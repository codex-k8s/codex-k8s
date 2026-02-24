---
doc_id: STG-CK8S-0001
type: process-model
title: "codex-k8s — Stage Process Model"
status: active
owner_role: EM
created_at: 2026-02-11
updated_at: 2026-02-24
related_issues: [1, 19, 90, 95, 139]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-19-full-docset"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-19
---

# Stage Process Model

## TL;DR
- Целевая модель: `intake -> vision -> prd -> arch -> design -> plan -> dev -> qa -> release -> postdeploy -> ops`.
- Для каждого этапа есть `run:*` и `run:*:revise` петля.
- Переход между этапами требует формального подтверждения артефактов и фиксируется в audit.
- Дополнительный служебный цикл `run:self-improve` работает поверх stage-контура и улучшает docs/prompts/tools по итогам запусков.
- Операционная видимость стадий/апрувов/логов предоставляется через staff web-console (разделы `Operations` и `Approvals`).

## Source of truth
- `docs/product/labels_and_trigger_policy.md`
- `docs/product/agents_operating_model.md`
- `docs/delivery/development_process_requirements.md`

## Этапы и обязательные артефакты

| Stage | Trigger labels | Основные артефакты | Основные роли |
|---|---|---|---|
| Intake | `run:intake`, `run:intake:revise` | problem, personas, scope, constraints, brief, traceability bundle | `pm`, `km` |
| Vision | `run:vision`, `run:vision:revise` | charter, success metrics, risk register | `pm`, `em` |
| PRD | `run:prd`, `run:prd:revise` | PRD, acceptance criteria, NFR draft | `pm`, `sa` |
| Architecture | `run:arch`, `run:arch:revise` | C4, ADR backlog/ADR, alternatives | `sa` |
| Design | `run:design`, `run:design:revise` | markdown design doc package (design/API/data model/migration policy notes) | `sa`, `qa` |
| Plan | `run:plan`, `run:plan:revise` | delivery plan, epics/stories, DoD | `em`, `km` |
| Development | `run:dev`, `run:dev:revise` | code changes, PR, docs updates | `dev`, `reviewer` |
| QA | `run:qa` | markdown test strategy/plan/matrix + regression evidence | `qa` |
| Release | `run:release` | release plan/notes, rollback plan | `em`, `sre` |
| Postdeploy | `run:postdeploy` | postdeploy review, postmortem | `qa`, `sre` |
| Ops | `run:ops` | markdown SLO/alerts/runbook improvements | `sre`, `km` |
| AI Repair | `run:ai-repair` | emergency infra recovery, stabilization fix, incident handover | `sre` |
| Self-Improve | `run:self-improve` | run/session diagnosis (MCP), PR with prompt/instruction updates and/or agent-runner Dockerfile changes | `km`, `dev`, `reviewer` |

## Петли ревизии и переосмысления

- На каждом этапе доступны:
  - `run:<stage>:revise` для доработки артефактов;
  - `run:rethink` для возврата на более ранний этап.
- После `run:rethink` предыдущие версии артефактов маркируются как `state:superseded`.

### Review-driven revise automation (implemented, Issue #95)
- При `pull_request_review` с `review.state=changes_requested` платформа автоматически запускает `run:<stage>:revise` при успешном stage-resolve.
- Resolver stage детерминирован и идёт по цепочке:
  1. stage label на PR (если ровно один);
  2. stage label на Issue (если ровно один);
  3. последний run context по связке `(repo, issue, pr)`;
  4. последний stage transition в `flow_events` по Issue.
- При конфликте stage labels или отсутствии резолва:
  - revise-run не стартует;
  - выставляется `need:input`;
  - публикуется service-comment с remediation.
- Коммуникация в review gate становится stage-aware:
  - для doc/design этапов подсказки включают `run:<stage>:revise` и `run:<next-stage>`;
  - для dev цикла — `run:dev:revise` и `run:qa`.
- Подсказки в GitHub-комментарии intentionally compact:
  - обычно 2 действия (`revise` + канонический `next-stage`);
  - для `design` дополнительно публикуется fast-track `run:dev` вместе с `run:plan`.
- Реализованный UX: next-step action-link открывает staff web-console, где frontend проверяет RBAC и подтверждает переход через модалку, а backend выполняет label transition на Issue.

## Вход/выход этапа

Общие правила входа:
- есть обязательные входные артефакты предыдущего этапа;
- нет блокеров `state:blocked`;
- отсутствует незакрытый `need:input`.

Общие правила выхода:
- артефакты этапа обновлены и связаны с Issue/PR в traceability документах (`issue_map`, sprint/epic docs);
- статус этапа отражён через `state:*` лейблы;
- события перехода записаны в аудит.

### Критерии приемки для Vision / Vision:Revise (E2E finalize, Issue #139)
- Для `run:vision` артефакты этапа обязаны содержать обновлённые:
  - `charter` (цель и границы результата),
  - `success metrics` (измеримые критерии успеха),
  - `risk register` (ключевые продуктовые риски и mitigation).
- Для `run:vision:revise` обязательно обработать все нерешённые комментарии Owner по vision-артефактам:
  - каждый комментарий получает ответ (внесено изменение или обоснованное отклонение),
  - unresolved threads не остаются без статуса.
- Для `run:vision` и `run:vision:revise` сервисный status-comment в Issue должен быть идемпотентным:
  - платформа поддерживает один активный status-comment на run,
  - повторные webhook/reconcile обновляют существующий комментарий, а не создают дубликаты.
- После завершения run обновляется traceability bundle:
  - записи в `docs/delivery/issue_map.md` и релевантных sprint/epic документах,
  - `state:in-review` установлен на Issue (и на PR при наличии PR-артефакта).

### Политика scope изменений
- Для `run:intake|vision|prd|arch|design|plan|doc-audit|qa|release|postdeploy|ops|rethink` разрешены только изменения markdown-документации (`*.md`).
- `run:dev|run:dev:revise` остаются единственными trigger-этапами для кодовых изменений.
- Для роли `reviewer` repository-write запрещён: только комментарии в существующем PR.
- Для `run:self-improve` разрешены только изменения:
  - prompt files (`services/jobs/agent-runner/internal/runner/promptseeds/**`, `services/jobs/agent-runner/internal/runner/templates/prompt_envelope.tmpl`);
  - markdown-инструкции/документация (`*.md`);
  - `services/jobs/agent-runner/Dockerfile`.

### Правило review gate для всех этапов
- Для всех `run:*` выход этапа проходит через review gate перед финальным review Owner:
  - pre-review от `reviewer` (для технических артефактов) и/или профильной роли через `need:*`;
  - финальное решение Owner по принятию артефактов.
- Постановка `state:in-review` выполняется так:
  - на PR и на Issue, если run завершился артефактами в PR;
  - только на Issue, если run завершился без PR.

## Паузы и таймауты в stage execution

- Разрешены paused состояния:
  - `waiting_owner_review`;
  - `waiting_mcp`.
- Для `waiting_mcp` timeout-kill не применяется до завершения ожидания.
- Для длительных пауз run должен оставаться resumable за счёт сохранения `codex-cli` session snapshot.

## Текущий активный контур (S3 Day1)

На текущем этапе реализации активирован полный trigger-контур:
- `run:intake..run:ops`;
- `run:ai-repair` (аварийный инфраструктурный контур);
- `run:<stage>:revise`;
- `run:rethink`, `run:self-improve`.

Ограничение текущего этапа:
- для части стадий пока активирован базовый orchestration path (routing/audit/policy),
  а специализированная бизнес-логика стадий дорабатывается следующими S3 эпиками.
- для prompt-body используется минимальная stage-matrix seed-шаблонов в `services/jobs/agent-runner/internal/runner/promptseeds/`
  (по схеме `<stage>-work.md` и `<stage>-revise.md` для revise-loop стадий).

## План активации контуров

- S2 baseline: `run:dev` и `run:dev:revise` (completed).
- S2 Day6: approval/audit hardening (completed).
- S2 Day7: regression gate под полный MVP (completed).
- S3 Day1: активация полного stage-flow (`run:intake..run:ops`) и trigger path для `run:self-improve` (completed).
- Day21: добавлен trigger `run:ai-repair` для аварийного восстановления инфраструктуры (production pod-path, fallback image strategy, main-direct recovery режим).
- S3 Day2+ : поэтапное насыщение stage-specific логики и observability.

## Конфигурационные labels для исполнения stage

- Помимо trigger/status labels используются конфигурационные labels:
  - `[ai-model-*]` — выбор модели;
  - `[ai-reasoning-*]` — выбор уровня рассуждений.
- Эти labels не запускают stage сами по себе, но влияют на effective runtime profile.
- Для `run:dev:revise` профиль model/reasoning перечитывается перед каждым запуском.
