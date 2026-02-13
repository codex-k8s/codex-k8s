---
doc_id: EPC-CK8S-S2-D7
type: epic
title: "Epic S2 Day 7: Dogfooding regression gate for MVP readiness"
status: planned
owner_role: EM
created_at: 2026-02-10
updated_at: 2026-02-13
related_issues: [19]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
---

# Epic S2 Day 7: Dogfooding regression gate for MVP readiness

## TL;DR
- Цель эпика: подтвердить, что S2 baseline + Day6 hardening готовы к расширению MVP в Sprint S3.
- Ключевая ценность: снимаем риски регрессий перед включением полного набора stage-flow и `run:self-improve`.
- MVP-результат: воспроизводимый regression bundle, формальный go/no-go и зафиксированный backlog Sprint S3 Day1..Day10.

## Priority
- `P0`.

## Scope
### In scope
- Regression matrix по текущему контуру:
  - `run:dev` -> run -> job -> PR;
  - `run:dev:revise` -> changes -> update PR;
  - отказ при конфликтных `ai-model` / `ai-reasoning` labels;
  - отказ privileged MCP операций без required approval.
- Regression matrix по Day6 control tools:
  - deterministic secret sync (GitHub + Kubernetes) с проверкой idempotency;
  - database create/delete по окружению;
  - owner feedback request (варианты + custom input).
- Проверка staff observability:
  - список running jobs;
  - исторические логи и flow events;
  - wait queue (`waiting_mcp`, `waiting_owner_review`) и причина ожидания.
- Проверка runtime hygiene:
  - отсутствие утечек namespace/job/slot после успешных/ошибочных прогонов;
  - поведение `run:debug` (cleanup skip + audit evidence).
- Документационный gate:
  - синхронизация product/architecture/delivery docs с расширенным MVP scope;
  - готовый Sprint S3 plan с 7-12 эпиками (в этом пакете: Day1..Day10).

### Out of scope
- Полный e2e regression по всем `run:*` стадиям до их реализации в Sprint S3.
- Production rollout; проверка ограничивается staging/dev dogfooding средой.

## Критерии приемки эпика
- Regression matrix и evidence опубликованы и воспроизводимы на staging.
- Нет открытых `P0` блокеров для старта Sprint S3.
- Зафиксирован go/no-go протокол и список рисков/долгов с owner decision.
