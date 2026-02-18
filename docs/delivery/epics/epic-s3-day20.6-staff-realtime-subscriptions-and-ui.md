---
doc_id: EPC-CK8S-S3-D20-6
type: epic
title: "Epic S3 Day 20.6: Staff realtime subscriptions and UI integration"
status: planned
owner_role: EM
created_at: 2026-02-18
updated_at: 2026-02-18
related_issues: [19]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
---

# Epic S3 Day 20.6: Staff realtime subscriptions and UI integration

## TL;DR
- Цель: подключить staff frontend к новой WS-шине и перевести ключевые экраны на near-realtime обновления.
- Результат: пользователь видит изменения run/deploy/errors/alers без ручного refresh, с graceful fallback на polling.

## Priority
- `P0`.

## Scope
### In scope
- Frontend realtime client layer:
  - единый WS transport + reconnect/backoff;
  - resume с `last_event_id`;
  - topic/scope subscriptions (project/run/deploy/system errors).
- Интеграция в критичные экраны:
  - Runs list + Run details (status/events/log markers);
  - Build & Deploy list/details;
  - Alert stack ошибок (из Day18).
- UX правила:
  - индикатор realtime connection state;
  - dedupe по `event_id`;
  - fallback polling при недоступном WS.
- Тестовый контур:
  - manual сценарии multi-tab/multi-reconnect;
  - smoke check на production перед Day21.

### Out of scope
- Полная замена всех текущих polling вызовов.
- Realtime для второстепенных/редко используемых экранов.

## Декомпозиция
- Story-1: shared realtime client module + state management.
- Story-2: интеграция в runs/deploy/errors views.
- Story-3: UX polish (indicators, degraded mode, dedupe).
- Story-4: regression checklist и readiness report к Day21.

## Критерии приемки
- Изменения статусов run/deploy отображаются в UI без ручного refresh.
- При обрыве соединения фронт восстанавливается и догружает пропущенные события.
- При отключенном WS интерфейс продолжает работать через polling (без функциональной деградации).
- Realtime-интеграция проходит ручной regression перед Day21 e2e.

## Риски/зависимости
- Зависимость от Day20.5 (backend bus + WS endpoint).
- Риск race-condition при смешанном WS + polling режиме: требуется единая стратегия merge обновлений.
