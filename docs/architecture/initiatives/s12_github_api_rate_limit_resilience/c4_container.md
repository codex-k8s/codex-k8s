---
doc_id: ARC-C4N-S12-0001
type: c4-container
title: "Sprint S12 Day 4 — C4 Container overlay for GitHub API rate-limit resilience"
status: approved
owner_role: SA
created_at: 2026-03-13
updated_at: 2026-03-13
related_issues: [366, 413, 416, 418, 420, 423, 425, 426, 427, 428, 429, 430, 431]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-03-13-issue-418-arch"
  approved_by: "ai-da-stas"
  approved_at: 2026-03-13
---

# C4 Container: Sprint S12 Day 4 GitHub API rate-limit resilience

## TL;DR
- Container baseline платформы не меняется: capability реализуется внутри существующих `agent-runner`, `control-plane`, `worker`, `api-gateway`, `web-console`, `postgres`.
- Новая Day4-фиксация касается ownership split для signal handoff, controlled wait aggregate, resume sweeps и typed visibility projections.

## Диаграмма (Mermaid C4Container)
```mermaid
C4Container
title Sprint S12 Day4 - GitHub API rate-limit resilience container overlay

Person(owner, "Owner / reviewer", "Ждёт понятный wait-state и следующий шаг")
Person(operator, "Platform operator", "Диагностирует affected contour и manual action")
System_Ext(github, "GitHub API", "Rate-limit signals and repo operations")

System_Boundary(b0, "codex-k8s") {
  Container(runner, "Agent Runner / agent pod", "Codex CLI job", "Captures raw agent-path evidence and stops local retries after handoff")
  Container(cp, "Control Plane", "Go", "Owns classification, controlled wait aggregate, visibility contract and manual-action decisions")
  Container(worker, "Worker", "Go", "Runs wait scheduling, eligibility sweeps and finite auto-resume attempts")
  Container(gw, "API Gateway", "Go HTTP", "Thin-edge staff/private transport for wait-state visibility")
  Container(web, "Web Console", "Vue 3", "Displays typed wait projections and next-step guidance")
  ContainerDb(db, "PostgreSQL", "Postgres + JSONB + pgvector", "Persisted wait aggregate, audit evidence and sweep state")
}

Rel(owner, web, "Reads contour, hint and next step", "HTTPS")
Rel(operator, web, "Uses diagnostics and manual-action views", "HTTPS")
Rel(runner, github, "gh/git requests with agent bot-token", "HTTPS")
Rel(cp, github, "Platform-managed GitHub calls and signal classification", "HTTPS")
Rel(worker, github, "Safe re-attempts under control-plane policy", "HTTPS")
Rel(runner, cp, "Handoff raw evidence after agent-path signal", "Internal callbacks")
Rel(cp, db, "Reads/Writes wait aggregate, audit and projection state")
Rel(worker, db, "Reads/Writes scheduling and sweep evidence")
Rel(gw, cp, "Reads typed wait projections", "gRPC")
Rel(web, gw, "Reads staff/private wait-state API", "HTTPS")
```

## Container responsibilities in GitHub API rate-limit resilience

| Container | Role |
|---|---|
| `agent-runner` | Видит raw agent-path evidence и прекращает local retry после handoff |
| `control-plane` | Единственный owner classification, wait aggregate, contour attribution и visibility contract |
| `worker` | Планирует wake-up, делает finite auto-resume attempts и эскалирует uncertainty |
| `api-gateway` | Отдаёт typed staff/private visibility contract без доменной логики |
| `web-console` | Показывает typed wait projections и manual-action hints |
| `postgres` | Единая persisted coordination layer между pod |

## Runtime и data boundaries
- `agent-runner` не хранит source-of-truth wait-state внутри pod.
- `worker` не выбирает contour и не переводит hard failure в recoverable wait без решения `control-plane`.
- `api-gateway` и `web-console` не вычисляют countdown или provider classification самостоятельно.
- `postgres` остаётся единственной точкой синхронизации для wait aggregate и resume evidence.

## Continuity after `run:plan`
- Typed raw-evidence handoff, persisted wait aggregate и finite auto-resume orchestration были детализированы на Day5 и разложены на execution waves `#425..#431` в Issue `#423`.
- Container ownership из этой схемы остаётся обязательным guardrail для implementation streams и не может меняться локальными решениями внутри `worker`, `agent-runner`, `api-gateway` или `web-console`.
