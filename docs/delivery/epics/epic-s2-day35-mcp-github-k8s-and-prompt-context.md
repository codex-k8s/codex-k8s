---
doc_id: EPC-CK8S-S2-D35
type: epic
title: "Epic S2 Day 3.5: MCP GitHub/K8s tools and prompt context assembly"
status: planned
owner_role: EM
created_at: 2026-02-12
updated_at: 2026-02-12
related_issues: []
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
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
