---
doc_id: EPC-CK8S-S3-D19-6
type: epic
title: "Epic S3 Day 19.6: Staff realtime subscriptions and UI integration"
status: completed
owner_role: EM
created_at: 2026-02-18
updated_at: 2026-02-19
related_issues: [19]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-19-full-docset"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-19
---

# Epic S3 Day 19.6: Staff realtime subscriptions and UI integration

## TL;DR
- Цель: подключить staff frontend к новой WS-шине и перевести ключевые экраны на near-realtime обновления.
- Результат: пользователь видит изменения run/deploy/errors/logs/events без ручного refresh, с graceful fallback на polling.

## Priority
- `P0`.

## Scope
### In scope
- Frontend realtime client layer:
  - единый WS transport + reconnect/backoff;
  - resume с `last_event_id`;
  - topic/scope subscriptions (project/run/deploy/system errors/run logs/run events/deploy logs/deploy events).
- Интеграция в критичные экраны:
  - Runs list + Run details (status/events/logs live stream);
  - Build & Deploy list/details (status/events/logs live stream);
  - Alert stack ошибок (из Day18).
- UX правила:
  - индикатор realtime connection state;
  - dedupe по `event_id`;
  - fallback polling при недоступном WS;
  - удаление кнопок `Обновить` в экранах с realtime-подпиской (обновление только автоматически).
- Тестовый контур:
  - manual сценарии multi-tab/multi-reconnect;
  - smoke check на production перед Day20.

### Out of scope
- Полная замена всех текущих polling вызовов в экранах без WS-подписки.
- Realtime для второстепенных/редко используемых экранов.

## Декомпозиция
- Story-1: shared realtime client module + state management.
- Story-2: интеграция в runs/deploy/errors/logs/events views.
- Story-3: удаление ручных кнопок `Обновить` в realtime-экранах + автообновление по событиям.
- Story-4: UX polish (indicators, degraded mode, dedupe).
- Story-5: regression checklist и readiness report к Day20.

## Критерии приемки
- Изменения статусов run/deploy, логи и события отображаются в UI без ручного refresh.
- При обрыве соединения фронт восстанавливается и догружает пропущенные события.
- При отключенном WS интерфейс продолжает работать через polling (без функциональной деградации).
- Кнопки `Обновить` удалены из всех экранов, покрытых realtime подпиской.
- Realtime-интеграция проходит ручной regression перед Day20 e2e.

## Риски/зависимости
- Зависимость от Day19.5 (backend bus + WS endpoint).
- Риск race-condition при смешанном WS + polling режиме: требуется единая стратегия merge обновлений.
