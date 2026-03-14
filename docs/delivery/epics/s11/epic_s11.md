---
doc_id: EPC-CK8S-0011
type: epic
title: "Epic Catalog: Sprint S11 (Telegram-адаптер взаимодействия с пользователем и первый внешний канал доставки)"
status: completed
owner_role: PM
created_at: 2026-03-14
updated_at: 2026-03-14
related_issues: [361, 444, 447, 448, 452, 454, 456, 458]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-03-14-issue-361-intake"
  approved_by: "ai-da-stas"
  approved_at: 2026-03-14
---

# Epic Catalog: Sprint S11 (Telegram-адаптер взаимодействия с пользователем и первый внешний канал доставки)

## TL;DR
- Sprint S11 открывает отдельную product initiative вокруг Telegram как первого внешнего канала поверх нового platform interaction contract.
- Day1 intake (`#361`) фиксирует проблему, MVP scope, sequencing guardrails и continuity в `run:vision`.
- Day2 vision выполнен в issue `#447`: mission, north star, persona outcomes, KPI/guardrails и MVP/Post-MVP границы зафиксированы, а issue `#444` сохранена только как historical handover от intake-stage и 2026-03-14 закрыта как `state:superseded`.
- Day3 PRD выполнен в issue `#448`: подготовлены user stories, FR/AC/NFR, expected evidence, callback/webhook guardrails и continuity issue `#452` для `run:arch`.
- Day4 architecture выполнен в issue `#452`: выпущен architecture package с C4 overlays, ADR/alternatives, ownership split и follow-up issue `#454` для `run:design`.
- Day5 design выполнен в issue `#454`: выпущен implementation-ready package `design_doc/api_contract/data_model/migrations_policy` и follow-up issue `#456` для `run:plan`.
- Day6 plan выполнен в issue `#456`: выпущен execution package с sequencing-waves, quality-gates, DoR/DoD и follow-up issue `#458` для `run:dev`.
- Документный контур Sprint S11 `intake -> vision -> prd -> arch -> design -> plan` завершён; дальнейшая кодовая реализация идёт только через owner-managed issue `#458`.

## Stage roadmap
- Day 1 (Intake): `docs/delivery/epics/s11/epic-s11-day1-telegram-user-interaction-adapter-intake.md` (Issue `#361`).
- Day 2 (Vision): `docs/delivery/epics/s11/epic-s11-day2-telegram-user-interaction-adapter-vision.md` (Issue `#447`); stage зафиксировал prerequisite `#389 closed` + `#387` как typed contract baseline.
- Day 3 (PRD): `docs/delivery/epics/s11/epic-s11-day3-telegram-user-interaction-adapter-prd.md` + `docs/delivery/epics/s11/prd-s11-day3-telegram-user-interaction-adapter.md` (Issue `#448`).
- Day 4 (Architecture): `docs/delivery/epics/s11/epic-s11-day4-telegram-user-interaction-adapter-arch.md` + `docs/architecture/initiatives/s11_telegram_user_interaction_adapter/{README.md,architecture.md,c4_context.md,c4_container.md}` (Issue `#452`).
- Day 5 (Design): `docs/delivery/epics/s11/epic-s11-day5-telegram-user-interaction-adapter-design.md` + `docs/architecture/initiatives/s11_telegram_user_interaction_adapter/{README.md,design_doc.md,api_contract.md,data_model.md,migrations_policy.md}` (Issue `#454`).
- Day 6 (Plan): `docs/delivery/epics/s11/epic-s11-day6-telegram-user-interaction-adapter-plan.md` (Issue `#456`) + handover issue `#458` для owner-managed запуска `run:dev`.

## Delivery-governance правила
- Sprint S11 не стартует параллельно с незафиксированным platform-core contract из Sprint S10; Telegram остаётся зависимым stream, а не заменой core initiative.
- Проверяемый gate для active vision stage `#447`: S10 plan issue `#389` остаётся closed и не отрывается от design package `#387`, где зафиксирован typed interaction contract.
- Каждый stage создаёт следующую issue без trigger-лейбла; запуск следующего stage остаётся owner-managed.
- До запуска `run:dev` в Issue `#458` Sprint S11 сохраняет markdown-only контур; code/runtime implementation не начинается вне этого execution anchor.
- Reference repositories `telegram-approver` и `telegram-executor` используются только как UX/stack baseline; `github.com/mymmrac/telego v1.7.0` внесён в каталог зависимостей как planned baseline, а прямое копирование решений запрещено без отдельного stage evidence.
- Telegram-specific UX, webhook ergonomics и inline buttons допустимы только как adapter-layer affordances поверх platform-owned interaction semantics.
- Voice/STT, reminders, richer conversation flows и дополнительные каналы не считаются blocking scope для core Sprint S11 и остаются отдельными follow-up waves.
