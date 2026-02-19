---
doc_id: EPC-CK8S-S3-D19
type: epic
title: "Epic S3 Day 19: Frontend manual QA hardening loop before full e2e"
status: planned
owner_role: EM
created_at: 2026-02-18
updated_at: 2026-02-19
related_issues: [19]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
---

# Epic S3 Day 19: Frontend manual QA hardening loop before full e2e

## TL;DR
- Цель: до финального e2e закрыть UI/UX дефекты в основном staff flow через управляемый цикл: ручная проверка Owner -> баг-репорты -> быстрые фиксы -> повторная проверка.
- Результат: основные разделы staff-консоли стабилизированы, критичные UI-регрессии закрыты до запуска полного e2e.

## Priority
- `P0`.

## Scope
### In scope
- Test matrix для ручной проверки UI:
  - глобальные фильтры/навигация/хлебные крошки,
  - build&deploy/runs/images/configs/repositories/users/projects,
  - run details/deploy task details/logs/events,
  - i18n и форматирование дат/времени.
- Быстрый triage багов и P0/P1 фиксы.
- UI polish и consistency pass для основных workflow.
- Regression checklist перед допуском к full e2e.

### Out of scope
- Дизайн-переосмысление всей консоли с нуля.
- Полный автотестовый UI набор (допускается ограниченный smoke).

## Декомпозиция
- Story-1: сформировать ручную QA матрицу и acceptance checklist.
- Story-2: пройти цикл баг-репортов и закрыть P0/P1.
- Story-3: выполнить regression pass по критичным экранам.
- Story-4: зафиксировать readiness report для перехода к Day20.

## Критерии приемки
- Все критичные P0/P1 баги из ручного frontend цикла закрыты.
- Основные страницы staff UI проходят regression checklist без блокеров.
- Owner подтверждает readiness к запуску full e2e gate.

## Риски/зависимости
- Риск накопления UX-долга в P2: требуется явный backlog после MVP.
- Зависимость от оперативного обратного цикла с Owner по найденным дефектам.
