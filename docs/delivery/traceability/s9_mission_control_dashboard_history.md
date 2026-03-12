---
doc_id: TRH-CK8S-S9-0001
type: traceability-history
title: "Sprint S9 Traceability History"
status: in-review
owner_role: KM
created_at: 2026-03-12
updated_at: 2026-03-12
related_issues: [333, 335, 337, 340]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-12-traceability-s9-history"
---

# Sprint S9 Traceability History

## TL;DR
- Этот файл хранит historical delta для Sprint S9.
- Текущая master-карта связей остаётся в `docs/delivery/issue_map.md`.
- Текущее покрытие FR/NFR остаётся в `docs/delivery/requirements_traceability.md`.

## Актуализация по Issue #333 (`run:intake`, 2026-03-12)
- Intake зафиксировал Mission Control Dashboard как отдельную product initiative, а не как локальный refactor staff console.
- В качестве baseline зафиксированы:
  - active-set control-plane UX;
  - discussion-first workflow;
  - provider-safe actions и command/reconciliation framing;
  - GitHub-first MVP и external human review.
- Создана follow-up issue `#335` для stage `run:vision`.

## Актуализация по Issue #335 (`run:vision`, 2026-03-12)
- Vision stage закрепил mission, north star, persona outcomes, KPI и guardrails для Mission Control Dashboard.
- Явно отделены MVP scope, Post-MVP scope и conditional voice candidate stream.
- Зафиксирована обязательная continuity-инструкция: на PRD stage сохранить active-set default, GitHub-first MVP, external review и command/reconciliation guardrails.
- Создана follow-up issue `#337` для stage `run:prd`.

## Актуализация по Issue #337 (`run:prd`, 2026-03-12)
- Подготовлен PRD package:
  - `docs/delivery/epics/s9/epic-s9-day3-mission-control-dashboard-prd.md`;
  - `docs/delivery/epics/s9/prd-s9-day3-mission-control-dashboard.md`.
- Зафиксированы:
  - user stories `S9-US-01..S9-US-05`;
  - product waves `Wave 1 -> Wave 2 -> Wave 3`;
  - FR/AC/NFR, edge cases и expected evidence для active-set dashboard, discussion formalization, provider-safe commands, dedupe/reconciliation и degraded fallback.
- Принято продуктовое решение: voice intake остаётся conditional `Wave 3` и не блокирует core MVP.
- Для continuity создана follow-up issue `#340` (`run:arch`) без trigger-лейбла.
- Для GitHub automation перед созданием follow-up issue и будущего PR-flow через Context7 подтверждён актуальный CLI-синтаксис `gh issue create`, `gh pr create`, `gh pr edit` (`/websites/cli_github_manual`).
- Проверки stage-пакета: markdown-only self-check, traceability sync, `git diff --check`.

