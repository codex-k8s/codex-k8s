---
doc_id: STR-CK8S-0119
type: story
title: "Story: Issue #119 — E2E A+B acceptance and evidence"
status: draft
owner_role: PM
created_at: 2026-02-24
updated_at: 2026-02-24
related_issues: [119, 118]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-24-issue-119-prd"
---

# User Story: Issue #119 — E2E A+B acceptance and evidence

## TL;DR
Сформировать и пройти проверяемый acceptance-контур для A+B сценариев, чтобы Owner принял результат MVP gate на основе воспроизводимого evidence.

## История
**Как** Owner/PM,
**я хочу** получить формально проверенный результат A+B e2e-среза,
**чтобы** принять решение по готовности core lifecycle и review-driven revise без ручных догадок.

## Контекст
- Предусловия:
  - есть актуальный `docs/delivery/e2e_mvp_master_plan.md`;
  - доступен Issue #118 для публикации evidence;
  - label policy и stage model действуют без override.
- Ограничения:
  - scope только A1-A3 и B1-B3;
  - только markdown-артефакты в рамках `run:prd`.

## Acceptance Criteria (Given/When/Then)
- AC-1: A lifecycle
  - Given подготовлен сценарный набор A1-A3
  - When выполнен полный прогон `run:intake -> ... -> run:ops`
  - Then подтверждён pass без P0/P1 и с ожидаемыми transitions
- AC-2: Review-driven revise
  - Given есть PR и `changes_requested`
  - When выполняются B1 и B3
  - Then revise workflow работает корректно, включая sticky model/reasoning profile
- AC-3: Ambiguity guard
  - Given создан неоднозначный label-контекст
  - When запускается B2
  - Then revise-run не стартует, выставляется `need:input`, публикуется remediation
- AC-4: Evidence handover
  - Given завершены все A+B сценарии
  - When формируется финальный отчёт
  - Then в Issue #118 опубликован полный evidence bundle с run_id и ссылками на артефакты

## Нефункциональные требования для этой истории
- Результаты воспроизводимы при повторной проверке.
- В evidence нет секретов.
- Трассируемость синхронизирована в issue_map.

## Telemetry / Analytics
- События:
  - `run.lifecycle.*`
  - `label.transition.*`
  - `review.revise.*`
- Метрики:
  - `PassRate_AB`
  - `EvidenceCompleteness`
  - `AuditCompleteness`

## Definition of Done
- Базовый DoD: `docs/delivery/development_process_requirements.md`
- Дополнительно для этой истории:
  - [ ] Все 6 сценариев A+B завершены в pass
  - [ ] `need:input` корректно отработан для B2
  - [ ] Evidence опубликован в Issue #118
  - [ ] `docs/delivery/issue_map.md` обновлён и содержит ссылки на PRD-пакет issue #119
  - [ ] `docs/delivery/design/issue-119/*.md` синхронизирован с PRD/NFR и master plan

## Открытые вопросы
1. Нужен ли отдельный owner-формат комментария для evidence bundle в Issue #118.

## Апрув
- request_id: owner-2026-02-24-issue-119-prd
- Решение: pending
- Комментарий:
