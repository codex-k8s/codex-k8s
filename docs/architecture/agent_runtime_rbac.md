---
doc_id: ARC-RBAC-CK8S-0001
type: runtime-rbac
title: "codex-k8s — Agent Runtime and RBAC Model"
status: draft
owner_role: SA
created_at: 2026-02-11
updated_at: 2026-02-12
related_issues: [1]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
---

# Agent Runtime and RBAC Model

## TL;DR
- Поддерживаются два режима исполнения: `full-env` и `code-only`.
- Права назначаются по роли агента и окружению запуска.
- Для `full-env` обязательно изолированное namespace-исполнение и аудит всех write-операций.

## Режимы исполнения

### `full-env`
- Агент запускается в отдельном run/issue namespace.
- Доступны логи, events, сервисы, метрики; write в Kubernetes ограничен ролью и policy.
- Используется для ролей, где решение зависит от состояния окружения.

### `code-only`
- Агент работает с репозиторием и API без прямого доступа к Kubernetes runtime.
- Используется для продуктовых, документационных и части ревизионных задач.

## RBAC-матрица (baseline)

| Роль | Default mode | K8s read | K8s write | DB/cache access | Secrets |
|---|---|---|---|---|---|
| `pm` | `code-only` | optional | no | no direct | no |
| `sa` | `full-env` | yes | no | schema/read-only via API | no |
| `em` | `full-env` | yes | limited (slot orchestration only) | no direct | no |
| `dev` | `full-env` | yes | via MCP tools + approval policy | read/write in run namespace scope | no direct |
| `reviewer` | `full-env` | yes | no direct (only diagnostic MCP calls) | read-only in run namespace scope | no |
| `qa` | `full-env` | yes | limited (test jobs) | read-only test scope | no |
| `sre` | `full-env` | yes | yes (via policy + approval) | diagnostic read-only | via controlled tools |
| `km` | `code-only` | optional read | no | docs/meta via API | no |
| `custom` | policy-defined | policy-defined | policy-defined | policy-defined | policy-defined |

## Namespace и ресурсная изоляция

- Для `run:dev`/`run:dev:revise` создаётся отдельный namespace по шаблону run/issue.
- На namespace применяются:
  - `ResourceQuota`/`LimitRange`,
  - service account per role/profile,
  - network policy baseline.
- Cleanup обязателен после завершения run (или по `run:abort`).

Текущий baseline реализации (S2 Day3):
- Worker создаёт namespace idempotent, применяет `ServiceAccount + Role + RoleBinding + ResourceQuota + LimitRange`.
- В `flow_events` пишутся lifecycle события `run.namespace.prepared|cleaned|cleanup_failed`.
- Runtime metadata namespace/job унифицированы через labels/annotations с префиксом `codex-k8s.dev/*`.
- Cleanup удаляет только managed namespaces с `codex-k8s.dev/managed-by=codex-k8s-worker` и `codex-k8s.dev/namespace-purpose=run`.

## Права `full-env` в рамках namespace

- Разрешено:
  - читать логи/события/метрики;
  - выполнять диагностический `exec` в pod'ы namespace;
  - обращаться к DB/cache сервисам проекта в границах namespace policy.
- Запрещено прямое изменение runtime без policy:
  - `kubectl apply/delete`, rollout/restart, создание/удаление workload выполняются только через MCP-инструменты.
- Для write-операций через MCP обязателен approver flow и аудит (`approval.requested/approved/denied`, `label.applied`, `run.wait.*`).

## Timeout и возобновление сессий

- Для paused wait-state `owner_review` run может иметь длительную паузу и возобновляться по решению Owner.
- Для wait-state `mcp` timeout-kill запрещён до получения ответа MCP.
- `codex-cli` session JSON сохраняется в `agent_sessions` и используется для resumable продолжения работы с того же места.

## Контроль доступа к данным и секретам

- Repo tokens хранятся в БД в шифрованном виде и не логируются.
- Agent pod получает только минимально необходимый секрет на время run.
- Прямой доступ агента к cluster secrets запрещён, кроме управляемых MCP-инструментов с policy.

## Аудит

- Каждая runtime-операция должна быть связана с `correlation_id`.
- Обязательные события:
  - namespace created/cleaned,
  - job started/finished,
  - privileged action requested/approved/applied.
- Источник аудита: `flow_events` + `agent_sessions` + `links`.

## Связанные документы
- `docs/product/agents_operating_model.md`
- `docs/product/labels_and_trigger_policy.md`
- `docs/architecture/mcp_approval_and_audit_flow.md`
- `docs/architecture/data_model.md`
