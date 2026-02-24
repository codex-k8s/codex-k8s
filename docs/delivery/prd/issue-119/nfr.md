---
doc_id: NFR-CK8S-0119
type: nfr
title: "Issue #119 — NFR: reliability/security/observability for E2E A+B"
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

# NFR: E2E A+B — reliability/security/observability

## TL;DR
- Требование: A+B e2e-срез должен быть воспроизводимым, безопасным и наблюдаемым.
- Как измеряем: 4 измеримых NFR-метрики (pass-rate, evidence completeness, audit completeness, secret-safety).
- Как проверяем: по evidence bundle в Issue #118 и выборке `flow_events`.
- Где мониторим: `flow_events`, service comments, issue/PR timeline.

## Контекст
Issue #119 закрывает критический MVP gate для core lifecycle и review-driven revise. Для приемки нужны не только функциональные pass/fail, но и минимальные NFR-гарантии по качеству evidence и безопасности.

## Требования (формализовано)
- NFR-119-1 (Reliability): полный набор A1-A3/B1-B3 повторяем при повторном прогоне с тем же входным label-контекстом без P0/P1 регрессий.
- NFR-119-2 (Observability): каждый обязательный шаг A+B имеет подтверждающий audit/event след, доступный для проверки.
- NFR-119-3 (Security): в комментариях, логах и evidence отсутствуют секреты и токен-материалы.
- NFR-119-4 (Traceability): для каждого сценария есть явная связь `scenario -> run_id -> label transitions -> evidence item`.

## Метрики/SLI/SLO
- SLI-1: `PassRate_AB` = passed scenarios / 6.
- SLI-2: `EvidenceCompleteness` = заполненные обязательные поля evidence / обязательные поля.
- SLI-3: `AuditCompleteness` = ожидаемые transitions в flow_events / ожидаемые transitions.
- SLI-4: `SecretSafety` = число найденных утечек секрета (целевое 0).

SLO:
- `PassRate_AB = 100%`
- `EvidenceCompleteness = 100%`
- `AuditCompleteness = 100%`
- `SecretSafety = 0 incidents`

## Архитектурные и инженерные меры
- Как достигаем:
  - единая структура evidence bundle;
  - формальная сверка с policy labels/stage;
  - запрет публикации секретов в любом артефакте.
- Компромиссы:
  - issue #119 ограничен A+B, поэтому C-F NFR-риски остаются для следующих прогонов.
- Ссылки на ADR:
  - `docs/architecture/adr/ADR-0006-review-driven-revise-and-next-step-ux.md`.

## Тестирование и верификация
- Как проверяем:
  - сверка 6 сценариев по чек-листу A+B;
  - валидация expected transitions в `flow_events`;
  - ручная проверка evidence на отсутствие секретов.
- Критерии прохода:
  - все SLO достигнуты;
  - отсутствуют незакрытые P0/P1 дефекты.

## Эксплуатация
- Какие runbooks нужны:
  - действующие production/log verification шаги из `AGENTS.md`.
- Что логируем:
  - run lifecycle events, label transitions, decisions по ambiguity.
- Что трассируем:
  - связь между issue/PR/run/evidence.

## Открытые вопросы
- Нужно ли фиксировать отдельный шаблон SQL-среза для evidence в рамках issue #119.

## Апрув
- request_id: owner-2026-02-24-issue-119-prd
- Решение: pending
- Комментарий:
