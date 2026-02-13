---
doc_id: SPR-CK8S-0003
type: sprint-plan
title: "Sprint S3: MVP completion (full stage flow, MCP control tools, self-improve, declarative full-env)"
status: in-progress
owner_role: EM
created_at: 2026-02-13
updated_at: 2026-02-13
related_issues: [19]
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
- Дополнительная цель: закрепить декларативный full-env deploy на `services.yaml` и полностью обновить staff-консоль на Vuetify до финального regression gate.
- Результат спринта: платформа поддерживает полный цикл разработки в GitHub, с проверяемым аудитом и формализованным go/no-go.

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
- Self-improve loop:
  - trigger `run:self-improve`;
  - анализ логов/комментариев/артефактов;
  - обновление docs/prompt templates/instructions/tooling по результатам.

## План эпиков по дням

| День | Эпик | Priority | Документ | Статус |
|---|---|---|---|---|
| Day 1 | Full stage and label activation | P0 | `docs/delivery/epics/epic-s3-day1-full-stage-and-label-activation.md` | completed (approved) |
| Day 2 | Staff runtime debug console (jobs/logs/waits) | P0 | `docs/delivery/epics/epic-s3-day2-staff-runtime-debug-console.md` | completed (awaiting Owner approval) |
| Day 3 | MCP deterministic secret sync (GitHub + K8s) | P0 | `docs/delivery/epics/epic-s3-day3-mcp-deterministic-secret-sync.md` | planned |
| Day 4 | MCP database lifecycle (create/delete per env) | P0 | `docs/delivery/epics/epic-s3-day4-mcp-database-lifecycle.md` | planned |
| Day 5 | Owner feedback handle + HTTP approver/executor + Telegram adapter | P0 | `docs/delivery/epics/epic-s3-day5-feedback-and-approver-interfaces.md` | planned |
| Day 6 | `run:self-improve`: ingestion and diagnostics | P0 | `docs/delivery/epics/epic-s3-day6-self-improve-ingestion-and-diagnostics.md` | planned |
| Day 7 | `run:self-improve`: docs/prompt/instruction updater + minimal stage prompt matrix | P0 | `docs/delivery/epics/epic-s3-day7-self-improve-updater-and-pr-flow.md` | planned |
| Day 8 | Agent toolchain auto-extension and policy safeguards | P1 | `docs/delivery/epics/epic-s3-day8-agent-toolchain-auto-extension.md` | planned |
| Day 9 | Declarative full-env deploy and runtime parity | P0 | `docs/delivery/epics/epic-s3-day9-declarative-full-env-deploy-and-runtime-parity.md` | planned |
| Day 10 | Staff console full redesign on Vuetify | P0 | `docs/delivery/epics/epic-s3-day10-staff-console-vuetify-redesign.md` | planned |
| Day 11 | MVP regression and security gate | P0 | `docs/delivery/epics/epic-s3-day11-mvp-regression-and-security-gate.md` | planned |
| Day 12 | MVP closeout, evidence bundle and handover | P0 | `docs/delivery/epics/epic-s3-day12-mvp-closeout-and-handover.md` | planned |

## Daily gate (обязательно)
- Green CI + успешный deploy на staging.
- Smoke + targeted regression для эпика дня.
- Синхронное обновление docs (`product`, `architecture`, `delivery`, `traceability`).
- Публикация evidence: flow events, links, UI screenshots/log excerpts, PR/Issue refs.

## Completion критерии спринта
- MVP-функции из Issue #19 реализованы и проверены на staging.
- Полный label/stage контур формально документирован и подтверждён regression evidence.
- Для `run:self-improve` есть минимум один воспроизводимый цикл с улучшениями в docs/prompt/tools.
- Owner принимает итоговый go/no-go протокол для перехода к post-MVP roadmap.
