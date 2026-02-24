---
doc_id: E2E-CK8S-0001
type: test-plan
title: "MVP Full E2E Master Plan (labels + runtime + governance)"
status: active
owner_role: QA
created_at: 2026-02-24
updated_at: 2026-02-24
related_issues: [19, 74, 95, 100, 112, 139]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-24-e2e-master-plan"
---

# MVP Full E2E Master Plan

## TL;DR
- Документ задаёт единый e2e-план для финальной проверки MVP перед публикацией.
- Покрытие включает полный label lifecycle (`run:*`, `state:*`, `need:*`, `[ai-model-*]`, `[ai-reasoning-*]`) и ключевые продуктовые контуры.
- План применим как к ручному прогону, так и к управляемому execution циклу агента `qa`/`km`.

## Контекст и цель
- Issue: `#112`.
- Дата фиксации плана: `2026-02-24`.
- Цель: подтвердить, что MVP-контур `webhook -> run -> review/revise -> release/postdeploy/ops` воспроизводим и безопасен.

## Границы покрытия

### In scope
- Trigger labels и stage transitions (`run:intake..run:ops`, `run:*:revise`, `run:rethink`, `run:self-improve`, `run:doc-audit`, `run:ai-repair`).
- Review-driven revise automation (Issue #95).
- MCP governance path (labels, `run_status_report`, control tools).
- Runtime режимы `code-only` и `full-env`, включая namespace TTL/revise reuse (Issue #74).
- Repo onboarding, docs federation и delivery traceability.
- Security/RBAC ограничения, включая запрет доступа к `secrets` в runtime namespace.

### Out of scope
- Пост-MVP расширения (custom-agent factory, A2A swarm, full autonomous schedules).
- Внешний пентест и нагрузочное тестирование beyond MVP gate.

## E2E среда

| Контур | Назначение | Обязательные проверки |
|---|---|---|
| `production-like full-env` | финальный go/no-go | full stage flow, MCP controls, TTL/revise reuse, audit completeness |
| `code-only` | документационные и policy stage run | scope enforcement (только markdown), traceability updates |
| `cross-project` (`codex-k8s` + `project-example`) | проверка multi-project изоляции | labels isolation, repo token isolation, docs mapping correctness |

## Матрица label-покрытия

### 1. `run:*` и `run:*:revise`

| Группа | Label | Минимальный сценарий | Ожидаемый результат |
|---|---|---|---|
| Product stages | `run:intake`, `run:vision`, `run:vision:revise`, `run:prd` | запуск stage + выпуск/ревизия docs артефакта | PR/Issue traceability обновлена, `state:in-review` установлен |
| Architecture/design stages | `run:arch`, `run:design`, `run:plan` | stage run + проверка связей в delivery docs | day/sprint/issue_map синхронизированы |
| Implementation | `run:dev`, `run:dev:revise` | создание PR + итерация по review | PR обновлён, revise flow без ambiguity |
| Verification/release | `run:qa`, `run:release`, `run:postdeploy`, `run:ops` | выпуск тест/release/ops артефактов | chain completion без нарушения policy |
| Special | `run:doc-audit`, `run:self-improve`, `run:rethink`, `run:ai-repair` | узкоспециализированные run и scope-проверка | ограничение файлового scope соблюдено, audit events полные |

### 2. `state:*`

| Label | Сценарий | Ожидаемый результат |
|---|---|---|
| `state:in-review` | завершение stage с артефактами | поставлен на Issue; на PR тоже, если PR создан |
| `state:approved` | owner approve | stage закрыт, traceability обновлена |
| `state:blocked` | искусственный блокер в середине stage | run не продолжает next-step до снятия блокера |
| `state:superseded` | `run:rethink` с новым артефактом | предыдущая версия явно помечена superseded |
| `state:abandoned` | отмена ветки сценария | статус финализирован без side effects на другие stage |

### 3. `need:*`

| Label | Сценарий | Ожидаемый результат |
|---|---|---|
| `need:input` | ambiguous stage resolve | revise-run не стартует, опубликована remediation подсказка |
| `need:reviewer` | обязательный pre-review | без review финальное owner-решение не закрывает цикл |
| `need:pm|sa|qa|sre|em|km` | role-specific gate | требуемый артефакт/комментарий подтверждён в issue/pr flow |

### 4. Конфигурационные labels

| Группа | Сценарий | Ожидаемый результат |
|---|---|---|
| `[ai-model-*]` | одиночный label | model profile применяется к следующему run |
| `[ai-model-*]` conflict | 2+ labels одновременно | run отклонён с `failed_precondition`, событие в `flow_events` |
| `[ai-reasoning-*]` | одиночный label | reasoning profile применён |
| `[ai-reasoning-*]` conflict | 2+ labels одновременно | run отклонён с диагностикой |

## Функциональные e2e-наборы

### Набор A. Stage lifecycle (core)
- A1: `run:intake -> run:vision -> run:prd -> run:arch -> run:design -> run:plan`.
- A2: `run:dev` создаёт PR, `run:dev:revise` отрабатывает review.
- A3: `run:qa -> run:release -> run:postdeploy -> run:ops`.
- Gate: ни один next-stage не проходит без актуального `state:*` и traceability.

### Набор B. Review-driven revise (Issue #95)
- B1: `changes_requested` при одном stage label на PR.
- B2: ambiguous labels -> `need:input` без старта revise.
- B3: sticky model/reasoning profile между revise-итерациями.

### Набор C. MCP governance tools
- C1: `github_labels_list|add|remove|transition` с audit trail.
- C2: `run_status_report` cadence (каждые 5-7 tool calls).
- C3: `secret.sync.github_k8s`, `database.lifecycle`, `owner.feedback.request` (approve/deny paths).

### Набор D. Runtime and infra
- D1: run namespace isolation, TTL lease и cleanup sweep.
- D2: revise namespace reuse и lease extension (`run:*:revise`).
- D3: `waiting_mcp` pause/resume без timeout-kill.

### Набор E. Security and RBAC
- E1: запрет `secrets` read/write для agent pod.
- E2: edge service остаётся thin-edge, домен только во internal.
- E3: secret-safe logs (нет токенов/credential material).

### Набор F. Multi-repo and docs governance
- F1: multi-repo docs federation (Issue #100 day1 baseline).
- F2: issue_map и requirements_traceability обновляются синхронно.
- F3: doc-audit run соблюдает markdown-only scope.

### Набор G. Vision revise + status comment idempotency (Issue #139)
- G1: `run:vision` формирует vision-артефакт и ставит `state:in-review` на Issue.
- G2: `run:vision:revise` после комментариев Owner обновляет тот же набор артефактов без потери traceability.
- G3: service status-comment остаётся единым (idempotent upsert): повторный запуск/повторный webhook обновляет существующий комментарий, не создавая дубликаты.
- G4: после цикла `run:vision -> run:vision:revise` в Issue отсутствуют необработанные комментарии по vision-артефактам.

## Порядок прогона
1. Подготовка окружения и preflight (tokens, namespace health, webhook availability).
2. Прогон core lifecycle (A + B + G).
3. Прогон governance/control tools (C).
4. Прогон runtime/security (D + E).
5. Прогон multi-repo/docs governance (F).
6. Сбор evidence bundle + owner go/no-go.

## Формат evidence bundle
- `run_id`, `issue`, `pr`, список label transitions.
- Ссылки на логи control-plane/worker и ключевые job logs.
- SQL/операционные срезы по `agent_runs`, `flow_events`, `mcp_action_requests`.
- Список отклонений (`expected vs actual`) и resolution plan.

## Критерии завершения (Go)
- Нет открытых P0/P1 блокеров.
- Матрица label coverage пройдена полностью.
- Security/RBAC проверки пройдены без критичных нарушений.
- Документация и traceability синхронизированы по факту прогона.
- Final gate Issue #139 пройден: подтверждены `run:vision`, `run:vision:revise` и идемпотентность status-comment.

## Открытые риски и продуктовые допущения
- Риск: если concurrent webhook-события приходят вне гарантированного порядка, возможны race-condition при обновлении service comment.
- Риск: при ручном удалении service-comment в Issue идемпотентный update-path требует корректного recreate-поведения и повторной привязки.
- Допущение: для acceptance-gate Issue #139 достаточно проверить idempotent поведение в рамках одной связки `(repo, issue, run)` без multi-repo конкуренции.
- Допущение: `state:in-review` остаётся обязательным выходным состоянием Vision/Vision:Revise до финального Owner approve.

## Источники фактов (актуализировано на 2026-02-24 через Context7)
- Kubernetes rollout/status checks: https://kubernetes.io/docs/concepts/workloads/controllers/deployment/
- Kubernetes deployment operations: https://kubernetes.io/docs/concepts/cluster-administration/manage-deployment/
- GitHub CLI JSON formatting: https://cli.github.com/manual/gh_help_formatting
- GitHub CLI PR create flow: https://cli.github.com/manual/gh_pr_create
- PostgreSQL LISTEN/NOTIFY: https://www.postgresql.org/docs/current/sql-notify.html
