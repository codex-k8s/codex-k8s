---
doc_id: SPR-CK8S-0003
type: sprint-plan
title: "Sprint S3: MVP completion (full stage flow, MCP control tools, self-improve, declarative full-env)"
status: in-progress
owner_role: EM
created_at: 2026-02-13
updated_at: 2026-02-18
related_issues: [19, 45]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
---

# Sprint S3: MVP completion (full stage flow, MCP control tools, self-improve, declarative full-env)

## TL;DR
- Sprint S3 завершает MVP после S2 Day6/Day7 hardening.
- Главная цель: включить полный stage/label контур, расширить staff debug observability, добавить обязательные MCP control tools и автоматический `run:self-improve` цикл.
- Дополнительная цель: закрыть core-flow пробелы до запуска полного e2e (prompt/docs context, env-scoped secret governance, runtime error journal, OAuth bypass key, frontend hardening).
- Дополнительная цель: добавить realtime контур (multi-server WebSocket updates через PostgreSQL LISTEN/NOTIFY шину) до запуска полного e2e.
- Финальный шаг спринта: full e2e regression gate и formal MVP closeout.

## Scope спринта
- Полный run-label контур: `run:intake -> run:vision -> run:prd -> run:arch -> run:design -> run:plan -> run:dev -> run:qa -> run:release -> run:postdeploy -> run:ops` + revise/rethink.
- Staff UI debug baseline:
  - running jobs;
  - live/historical logs;
  - wait queue по `waiting_mcp` и `waiting_owner_review`.
- MCP control tools MVP:
  - deterministic secret sync GitHub + Kubernetes;
  - database create/delete по окружениям;
  - owner feedback handle (варианты + custom answer);
  - HTTP approver/executor contracts + Telegram adapter.
- Declarative full-env deploy:
  - typed `services.yaml` contract + execution plan;
  - repo-sync/runtime parity;
  - namespace-level isolation для ai-slot.
- Core-flow completion перед e2e:
  - docs tree/roles в `services.yaml` для prompt context;
  - role-aware prompt templates + GitHub service message templates;
  - environment-scoped secret overrides и OAuth callback strategy;
  - runtime error journal + staff alert stack;
  - run-scoped access key для controlled OAuth bypass;
  - frontend manual QA hardening loop.
- Realtime transport before e2e:
  - backend event bus: PostgreSQL event log + `LISTEN/NOTIFY`;
  - `api-gateway` WebSocket backplane с catch-up через `last_event_id`;
  - frontend realtime subscriptions (runs/deploy/errors) с fallback polling.

## План эпиков по дням

| День | Эпик | Priority | Документ | Статус |
|---|---|---|---|---|
| Day 1 | Full stage and label activation | P0 | `docs/delivery/epics/epic-s3-day1-full-stage-and-label-activation.md` | completed (approved) |
| Day 2 | Staff runtime debug console (jobs/logs/waits) | P0 | `docs/delivery/epics/epic-s3-day2-staff-runtime-debug-console.md` | completed (approved) |
| Day 3 | MCP deterministic secret sync (GitHub + K8s) | P0 | `docs/delivery/epics/epic-s3-day3-mcp-deterministic-secret-sync.md` | completed |
| Day 4 | MCP database lifecycle (create/delete/describe per env) | P0 | `docs/delivery/epics/epic-s3-day4-mcp-database-lifecycle.md` | completed |
| Day 5 | Owner feedback handle + HTTP approver/executor + Telegram adapter | P0 | `docs/delivery/epics/epic-s3-day5-feedback-and-approver-interfaces.md` | completed |
| Day 6 | `run:self-improve`: ingestion and diagnostics | P0 | `docs/delivery/epics/epic-s3-day6-self-improve-ingestion-and-diagnostics.md` | completed |
| Day 7 | `run:self-improve`: docs/prompt/instruction updater + minimal stage prompt matrix | P0 | `docs/delivery/epics/epic-s3-day7-self-improve-updater-and-pr-flow.md` | completed |
| Day 8 | Agent toolchain auto-extension and policy safeguards | P1 | `docs/delivery/epics/epic-s3-day8-agent-toolchain-auto-extension.md` | completed |
| Day 9 | Declarative full-env deploy and runtime parity | P0 | `docs/delivery/epics/epic-s3-day9-declarative-full-env-deploy-and-runtime-parity.md` | completed |
| Day 10 | Staff console full redesign on Vuetify | P0 | `docs/delivery/epics/epic-s3-day10-staff-console-vuetify-redesign.md` | completed |
| Day 11 | Full-env slot namespace + subdomain templating (TLS) + agent run | P0 | `docs/delivery/epics/epic-s3-day11-full-env-slots-and-subdomains.md` | completed |
| Day 12 | Docset import + safe sync (agent-knowledge-base -> projects) | P0 | `docs/delivery/epics/epic-s3-day12-docset-import-and-safe-sync.md` | completed |
| Day 13 | Unified config/secrets governance + GitHub credentials fallback | P0 | `docs/delivery/epics/epic-s3-day13-config-and-credentials-governance.md` | completed |
| Day 14 | Repository onboarding preflight + bot params per repo | P0 | `docs/delivery/epics/epic-s3-day14-repository-onboarding-preflight.md` | completed |
| Day 15 | Prompt context overhaul (docs tree, role matrix, GitHub service messages) | P0 | `docs/delivery/epics/epic-s3-day15-mvp-closeout-and-handover.md` | planned |
| Day 16 | gRPC transport boundary hardening (исключить прямые вызовы repository из handlers) | P0 | `docs/delivery/epics/epic-s3-day16-grpc-transport-boundary-hardening.md` | completed |
| Day 17 | Environment-scoped secret overrides and OAuth callback strategy | P0 | `docs/delivery/epics/epic-s3-day17-environment-scoped-secret-overrides-and-oauth-callbacks.md` | planned |
| Day 18 | Runtime error journal and staff alert center | P0 | `docs/delivery/epics/epic-s3-day18-runtime-error-journal-and-staff-alert-center.md` | planned |
| Day 19 | Run access key and OAuth bypass flow | P0 | `docs/delivery/epics/epic-s3-day19-run-access-key-and-oauth-bypass.md` | planned |
| Day 20 | Frontend manual QA hardening loop | P0 | `docs/delivery/epics/epic-s3-day20-frontend-manual-qa-hardening-loop.md` | planned |
| Day 20.5 | Realtime event bus (PostgreSQL LISTEN/NOTIFY) and WebSocket backplane | P0 | `docs/delivery/epics/epic-s3-day20.5-realtime-event-bus-and-websocket-backplane.md` | planned |
| Day 20.6 | Staff realtime subscriptions and UI integration | P0 | `docs/delivery/epics/epic-s3-day20.6-staff-realtime-subscriptions-and-ui.md` | planned |
| Day 21 | Full e2e regression gate + MVP closeout | P0 | `docs/delivery/epics/epic-s3-day21-e2e-regression-and-mvp-closeout.md` | planned |

## Daily gate (обязательно)
- Green CI + успешный deploy на production.
- Smoke + targeted regression для эпика дня.
- Синхронное обновление docs (`product`, `architecture`, `delivery`, `traceability`).
- Публикация evidence: flow events, links, UI screenshots/log excerpts, PR/Issue refs.

## Completion критерии спринта
- MVP-функции из Issue #19 реализованы и проверены на production.
- Полный label/stage контур формально документирован и подтверждён regression evidence.
- Для `run:self-improve` есть минимум один воспроизводимый цикл с улучшениями в docs/prompt/tools.
- Core-flow недоделки закрыты: prompt/docs context, env secrets, runtime error alerts, OAuth bypass, frontend hardening.
- Realtime контур закрыт: multi-server backend bus + frontend WS subscriptions + fallback mode.
- Финальный Day21 e2e проходит без P0 блокеров и формирует owner-ready closeout пакет.
