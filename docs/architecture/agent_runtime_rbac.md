---
doc_id: ARC-RBAC-CK8S-0001
type: runtime-rbac
title: "codex-k8s — Agent Runtime and RBAC Model"
status: active
owner_role: SA
created_at: 2026-02-11
updated_at: 2026-02-14
related_issues: [1, 19]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-19-full-docset"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-19
---

# Agent Runtime and RBAC Model

## TL;DR
- Поддерживаются два режима исполнения: `full-env` и `code-only`.
- Права назначаются по роли агента и окружению запуска.
- Для `full-env` обязательно изолированное namespace-исполнение; agent pod получает прямой `kubectl` доступ в свой namespace, кроме `secrets`.
- Привилегированные операции с секретами/БД выполняются через MCP control tools с approval policy, а не прямым `kubectl`/SQL доступом.

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
| `dev` | `full-env` | yes | yes (почти все namespaced операции, кроме secrets) | read/write in run namespace scope | no direct secrets access |
| `reviewer` | `full-env` | yes | yes (диагностика/дебаг в namespace, кроме secrets) | read-only in run namespace scope | no direct secrets access |
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
- Cleanup обязателен после завершения run. Если на issue присутствует `run:debug`, cleanup пропускается и namespace сохраняется для отладки.
- При сохранении namespace в debug-сценарии статус-комментарий run обязан явно отмечать, что namespace не удалён, и давать ссылку на run details для ручного удаления.

Текущий baseline реализации (S2 Day3):
- Worker создаёт namespace idempotent, применяет `ServiceAccount + Role + RoleBinding + ResourceQuota + LimitRange`.
- В `flow_events` пишутся lifecycle события `run.namespace.prepared|cleaned|cleanup_failed|cleanup_skipped`.
- Runtime metadata namespace/job унифицированы через labels/annotations с префиксом `codex-k8s.dev/*`.
- Cleanup удаляет только managed namespaces с `codex-k8s.dev/managed-by=codex-k8s-worker` и `codex-k8s.dev/namespace-purpose=run`.

## Права `full-env` в рамках namespace

- Разрешено:
  - читать логи/события/метрики;
  - выполнять диагностический `exec` в pod'ы namespace;
  - выполнять через `kubectl` namespaced операции для runtime-сущностей (`pods`, `deployments`, `statefulsets`, `daemonsets`, `replicasets`, `jobs`, `cronjobs`, `services`, `ingresses`, `networkpolicies`, `configmaps`, `pvcs`, `resourcequotas`, `limitranges`, `events`);
  - обращаться к DB/cache сервисам проекта в границах namespace policy.
- Запрещено:
  - прямое чтение/запись `secrets`;
  - выход за пределы своего namespace и cluster-scope операции.
- MCP в MVP baseline используется для label-операций и control tools (`secret sync`, `database lifecycle`, `owner feedback`) с approval/audit контуром.

Эволюция policy (Day6+):
- effective MCP права вычисляются по связке `agent_key + run label`;
- для cluster-scope сущностей (`ingressclasses`, `pvs`, `storageclasses`) применяются отдельные ограничения;
- комбинированные MCP-ручки (например GitHub+Kubernetes) имеют отдельные policy-профили и approvals.

## Timeout и возобновление сессий

- Для paused wait-state `owner_review` run может иметь длительную паузу и возобновляться по решению Owner.
- Для wait-state `mcp` timeout-kill запрещён до получения ответа MCP.
- `codex-cli` session JSON сохраняется в `agent_sessions` и используется для resumable продолжения работы с того же места.

## Контроль доступа к данным и секретам

- Repo tokens хранятся в БД в шифрованном виде и не логируются.
- Platform/bot GitHub токены хранятся в singleton таблице `platform_github_tokens`
  (поля `platform_token_encrypted`, `bot_token_encrypted`) и синхронизируются из env на старте control-plane.
- Agent pod получает минимально необходимые runtime-секреты на время run:
  - `CODEXK8S_OPENAI_API_KEY` для codex auth;
  - `CODEXK8S_GIT_BOT_TOKEN` для git transport path.
- Для `full-env` pod формируется `KUBECONFIG` из namespaced ServiceAccount.
- Прямой доступ агента к Kubernetes `secrets` запрещён RBAC (read/write).
- Создание/обновление секретов с генерацией значений и approver-политикой выполняется через MCP control tools.

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
