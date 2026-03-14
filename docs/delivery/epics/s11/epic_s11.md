---
doc_id: EPC-CK8S-0011
type: epic
title: "Epic Catalog: Sprint S11 (Telegram-адаптер взаимодействия с пользователем и первый внешний канал доставки)"
status: in-review
owner_role: PM
created_at: 2026-03-14
updated_at: 2026-03-14
related_issues: [361, 444, 447, 448]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-14-issue-361-intake"
---

# Epic Catalog: Sprint S11 (Telegram-адаптер взаимодействия с пользователем и первый внешний канал доставки)

## TL;DR
- Sprint S11 открывает отдельную product initiative вокруг Telegram как первого внешнего канала поверх нового platform interaction contract.
- Day1 intake (`#361`) фиксирует проблему, MVP scope, sequencing guardrails и continuity в `run:vision`.
- Day2 vision выполняется в issue `#447`: mission, north star, persona outcomes, KPI/guardrails и MVP/Post-MVP границы уже зафиксированы, а issue `#444` остаётся historical handover от intake-stage.
- Создана follow-up issue `#448` для stage `run:prd` без trigger-лейбла; следующий этап должен формализовать user stories, FR/AC/NFR, expected evidence и сохранить continuity-цепочку до `run:dev`.
- До `run:plan` Sprint S11 остаётся markdown-only контуром: код, runtime topology и library/runtime binding decisions начинаются только после owner review последующих stage-пакетов.

## Stage roadmap
- Day 1 (Intake): `docs/delivery/epics/s11/epic-s11-day1-telegram-user-interaction-adapter-intake.md` (Issue `#361`).
- Day 2 (Vision): `docs/delivery/epics/s11/epic-s11-day2-telegram-user-interaction-adapter-vision.md` (Issue `#447`); active stage сохраняет prerequisite `#389 closed` + `#387` как typed contract baseline.
- Day 3 (PRD): follow-up issue `#448`; stage фиксирует user stories, FR/AC/NFR, expected evidence и Telegram-specific edge cases.
- Day 4 (Architecture): continuity issue создаётся на завершении `run:prd`; stage фиксирует adapter/service boundaries, callback ownership и security/correlation lifecycle.
- Day 5 (Design): continuity issue создаётся на завершении `run:arch`; stage фиксирует implementation-ready API/data/runtime contracts.
- Day 6 (Plan): continuity issue создаётся на завершении `run:design`; stage формирует execution package, sequencing-waves и owner-managed handover в `run:dev`.

## Delivery-governance правила
- Sprint S11 не стартует параллельно с незафиксированным platform-core contract из Sprint S10; Telegram остаётся зависимым stream, а не заменой core initiative.
- Проверяемый gate для active vision stage `#447`: S10 plan issue `#389` остаётся closed и не отрывается от design package `#387`, где зафиксирован typed interaction contract.
- Каждый stage создаёт следующую issue без trigger-лейбла; запуск следующего stage остаётся owner-managed.
- До `run:plan` в Sprint S11 не создаются implementation issues и не добавляются новые зависимости в репозиторий.
- Reference repositories `telegram-approver` и `telegram-executor` используются только как UX/stack baseline; `github.com/mymmrac/telego v1.7.0` внесён в каталог зависимостей как planned baseline, а прямое копирование решений запрещено без отдельного stage evidence.
- Telegram-specific UX, webhook ergonomics и inline buttons допустимы только как adapter-layer affordances поверх platform-owned interaction semantics.
- Voice/STT, reminders, richer conversation flows и дополнительные каналы не считаются blocking scope для core Sprint S11 и остаются отдельными follow-up waves.
