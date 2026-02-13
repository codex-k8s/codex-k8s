---
doc_id: REG-CK8S-S2-0001
type: regression
title: "Sprint S2 Regression Gate (staging)"
status: completed
owner_role: QA
created_at: 2026-02-13
updated_at: 2026-02-13
related_issues: [19]
related_prs: [20, 22, 23]
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-13-s2-day7"
---

# Sprint S2 Regression Gate (staging)

Цель: зафиксировать воспроизводимый regression bundle по dogfooding baseline S2 перед стартом Sprint S3.

## Preconditions
- Актуальный deploy на `main` завершился успешно:
  - workflow run: `21985095587`
  - job: `Deploy codex-k8s to ai-staging`
  - conclusion: `success`.
- В namespace staging все ключевые deploy в состоянии `READY 1/1`:
  - `codex-k8s`, `codex-k8s-control-plane`, `codex-k8s-worker`, `codex-k8s-web-console`, `oauth2-proxy`.

## Regression matrix (S2 Day7)

| Сценарий | Evidence | Результат |
|---|---|---|
| `run:dev` -> run -> job -> PR | `agent_runs`: `run:dev succeeded=9`; `flow_events`: `run.pr.created=5`; примеры: run `8867c62e-ff01-4ada-b0a2-d5cbd15111df`, PR `#22` | pass |
| `run:dev:revise` -> changes -> update PR | `agent_runs`: `run:dev:revise succeeded=3`; `flow_events`: `run.pr.updated=3`; примеры: run `536beaea-a17d-4843-a8a6-1698addccb37`, PR `#20` | pass |
| Конфликтные `ai-model`/`ai-reasoning` labels отклоняются | Unit regression: `TestResolveModelFromLabels_ConflictingLabels`, `TestResolveRunAgentContext_ConflictingPullRequestLabelsFail` | pass |
| `run:debug` сохраняет namespace и фиксирует аудит | `flow_events`: `run.namespace.cleanup_skipped=5`, payload reason=`debug_label_present` | pass |
| Runtime hygiene (утечки namespace/job) | На момент gate: `kubectl get ns -l codex-k8s.dev/namespace-purpose=run` возвращает пусто | pass |
| Staff observability baseline (runs/events/waits) | `flow_events` и `agent_runs` консистентны; `agent_sessions.wait_state`: только `null` в текущем окне, pending approvals не зависли | pass |
| Day6 approval queue consistency | `mcp_action_requests` не содержит зависших `requested`; `pending_approvals=0` | pass |

## Команды проверки (фактически выполненные)

```bash
# GitHub workflow status
gh run view -R "$CODEXK8S_GITHUB_REPO" 21985095587 --json workflowName,status,conclusion,jobs,url

# Staging health
kubectl -n "$CODEXK8S_STAGING_NAMESPACE" get deploy,pods,jobs
kubectl -n "$CODEXK8S_STAGING_NAMESPACE" logs deploy/codex-k8s-worker --tail=80
kubectl -n "$CODEXK8S_STAGING_NAMESPACE" logs deploy/codex-k8s-control-plane --tail=80

# Run namespace leaks
kubectl get ns -l codex-k8s.dev/managed-by=codex-k8s-worker,codex-k8s.dev/namespace-purpose=run

# DB evidence snapshot
kubectl -n "$CODEXK8S_STAGING_NAMESPACE" exec -i postgres-0 -- \
  psql -U codex_k8s -d codex_k8s

# Label conflict regression tests
go test ./services/jobs/worker/internal/domain/worker \
  -run 'TestResolveModelFromLabels_ConflictingLabels|TestResolveRunAgentContext_ConflictingPullRequestLabelsFail|TestResolveRunAgentContext_ConfigLabelsPullRequestOverrideIssue|TestResolveRunAgentContext_ReasoningExtraHighLabel|TestResolveRunAgentContext_UsesPullRequestHintsForRevise'
```

## Go/No-Go
- Решение: **Go** для старта Sprint S3.
- Основание: нет открытых P0 блокеров в S2 runtime контуре, deploy/staging стабильны, dogfooding цикл `run:dev`/`run:dev:revise` воспроизводим.

## Residual risks / follow-ups
- Полный runtime e2e для Day6 control tools (`secret.sync.github_k8s`, `database.lifecycle`, `owner.feedback.request`) в production-режиме approval/deny закреплён в Sprint S3 Day3..Day5.
- Исторические неуспешные S2 прогоны (RBAC/escalate и ранние PR edge-cases) закрыты как известные инциденты и не воспроизводятся в актуальном контуре.
