---
doc_id: EPC-CK8S-S2-D35
type: epic
title: "Epic S2 Day 3.5: MCP GitHub/K8s tools and prompt context assembly"
status: completed
owner_role: EM
created_at: 2026-02-12
updated_at: 2026-02-13
related_issues: []
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-19-full-docset"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-19
---

# Epic S2 Day 3.5: MCP GitHub/K8s tools and prompt context assembly

## TL;DR
- Цель эпика: ввести MCP-first execution слой до Day4, чтобы агент не получал прямые GitHub/Kubernetes секреты.
- Ключевая ценность: единый policy/audit-контур для всех write-операций агента.
- MVP-результат: встроенный MCP-сервер в `control-plane` с GitHub и namespaced Kubernetes ручками + подготовка runtime-контекста для рендера final prompt.

## Priority
- `P0` (dependency для Day4).

## Scope
### In scope
- Реализовать встроенный MCP-сервер платформы в `services/internal/control-plane`:
  - authn/authz per-run (короткоживущий токен/контекст, привязанный к `run_id`/`project_id`/`namespace`);
  - централизованный аудит MCP-вызовов (`flow_events`, correlation).
- Реализовать MCP-ручки GitHub (минимум для Day4 цикла):
  - read: issue/PR/comments/labels/branches;
  - write: branch sync, push, PR create/update, comment/reply, label apply/remove (по policy).
- Реализовать MCP-ручки Kubernetes (в пределах namespace текущего run):
  - read-only диагностические: pods, logs, events, describe, exec (diagnostic);
  - write-ручки только в рамках policy и namespace scope.
- Формализовать policy ручек:
  - какие ручки требуют approval;
  - какие разрешены без approval;
  - какие полностью запрещены для роли/режима.
- Подготовить prompt runtime context assembler:
  - единый объект контекста для рендера final prompt;
  - включить metadata по окружению, сервисам и MCP-ручкам.
- Обновить документацию по контракту prompt render:
  - seed как baseline body;
  - runtime envelope + context blocks как обязательная надстройка.

### Out of scope
- Полная поддержка внешних MCP-серверов сторонних вендоров (Slack/Jira/Mattermost) кроме базового контрактного слоя.

## Data model impact
- Расширение `flow_events.payload` полями MCP-вызовов:
  - `mcp.server`, `mcp.tool`, `mcp.action`, `mcp.approval_state`, `mcp.result`.
- Расширение `agent_sessions.session_json`:
  - effective MCP tool catalog snapshot;
  - prompt render context metadata/version.

## Критерии приемки эпика
- Агентный pod выполняет GitHub/Kubernetes write-действия только через MCP-инструменты.
- Прямые GitHub/Kubernetes секреты отсутствуют в env агентного pod.
- Все MCP вызовы трассируются в audit-контуре с `correlation_id`.
- Для каждого run формируется детерминированный prompt render context, содержащий:
  - environment/runtime metadata;
  - services overview;
  - MCP server/tool catalog + approval flags.

## Зависимости и handoff
- Input from Day3: per-issue namespace и RBAC baseline.
- Output to Day4:
  - готовый MCP tool layer для git/PR/debug операций;
  - готовый prompt context assembler для рендера `work/review` шаблонов.

## Фактическая реализация (2026-02-12)
- В `services/internal/control-plane` добавлен встроенный MCP transport (`/mcp`, StreamableHTTP) на `github.com/modelcontextprotocol/go-sdk v1.3.0` с bearer auth и run-bound валидацией токена.
- В `control-plane` реализована доменная служба MCP:
  - выпуск и проверка short-lived run токенов;
  - deterministic tool catalog;
  - GitHub read/write ручки (issue/pr/comments/labels/branches/ensure/upsert/comment);
  - для `github_issue_comments_list` введена фильтрация служебных комментариев владельца token по умолчанию с опциональным override (`include_token_owner_comments=true`);
  - Kubernetes ручки:
    - namespaced diagnostics/read: `pods`, `events`, `deployments`, `daemonsets`, `statefulsets`, `replicasets`, `replicationcontrollers`, `jobs`, `cronjobs`, `configmaps`, `secrets`, `resourcequotas`, `hpa`, `services`, `endpoints`, `ingresses`, `networkpolicies`, `pvcs`;
    - cluster-scope read: `ingressclasses`, `pvs`, `storageclasses`;
    - policy-gated write: `pod port-forward`, `manifest apply/delete` -> `approval_required`.
- В `flow_events` добавлены audit-события MCP:
  - `run.mcp.token.issued`,
  - `prompt.context.assembled`,
  - `mcp.tool.called|succeeded|failed|approval_pending`.
- В `control-plane` gRPC контракт добавлен RPC `IssueRunMCPToken`; worker получает run-bound MCP токен перед запуском job.
- В job env run pod теперь передаются только:
  - `CODEXK8S_MCP_BASE_URL`,
  - `CODEXK8S_MCP_BEARER_TOKEN`,
  без прямых GitHub/Kubernetes write-секретов.
- Добавлен MCP prompt context resource `codex://prompt/context` и одноимённый tool для рендера final prompt в Day4.
- Обновлены deploy/bootstrap/workflow переменные и секреты:
  - `CODEXK8S_CONTROL_PLANE_MCP_BASE_URL`,
  - `CODEXK8S_MCP_TOKEN_SIGNING_KEY`,
  - `CODEXK8S_MCP_TOKEN_TTL` (default `24h`, не меньше baseline lifetime агентного контейнера).

### Актуализация после S2 Day4 hardening (2026-02-12)
- Текущий runtime baseline упрощён: из MCP удалены все non-label ручки.
- Активный MCP catalog:
  - `github_labels_list`,
  - `github_labels_add`,
  - `github_labels_remove`,
  - `github_labels_transition`.
- GitHub issue/PR/comments и Kubernetes runtime операции выполняются агентом напрямую через `gh`/`kubectl` в рамках выданных прав.
- Prompt context больше не поставляется отдельным MCP resource/tool в runtime-контуре.

## Следующий шаг по policy (handoff в Day6)
- Вынести effective policy управления MCP-ручками/ресурсами в платформенную модель:
  - связка `agent_key + run label`;
  - action/scope матрица (`read/write/approval_required`) по категориям сущностей.
- Добавить composite tools roadmap:
  - комбинированные ручки (например синхронизация секретов GitHub+Kubernetes) с отдельными approval правилами и аудитом.

## Критерии приемки эпика — статус
- Выполнено: write-действия для Day4 вынесены в MCP tool layer, прямой write-path в агентный pod не закладывается.
- Выполнено: run pod не требует прямые GitHub/Kubernetes секреты для MCP-операций.
- Выполнено: MCP вызовы и token issue трассируются в `flow_events` с `correlation_id`.
- Выполнено: prompt runtime context формируется детерминированно и доступен через MCP tool/resource.
