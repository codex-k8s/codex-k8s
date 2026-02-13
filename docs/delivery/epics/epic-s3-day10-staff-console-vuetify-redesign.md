---
doc_id: EPC-CK8S-S3-D10
type: epic
title: "Epic S3 Day 10: Staff console full redesign on Vuetify"
status: planned
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

# Epic S3 Day 10: Staff console full redesign on Vuetify

## TL;DR
- Цель: полностью переделать текущую staff-консоль под `Vuetify` и закрыть операционные UX-пробелы MVP.
- MVP-результат: единый admin UI-контур для jobs/logs/waits/approvals/feedback с переиспользуемыми компонентами и предсказуемым UX.

## Priority
- `P0`.

## Scope
### In scope
- Полная миграция staff-консоли на `Vuetify` (Vue 3):
  - layout/navigation/table/form patterns для админ-сценариев;
  - унифицированные компоненты для run list, run details, wait queue, approvals.
- UI/UX паритет с `telegram-executor` для feedback/approval сценариев:
  - вариантные ответы + custom input;
  - поддержка voice/STT маршрута для отказа/комментария через соответствующий HTTP-аппрувер.
- Операционные экраны:
  - список запущенных job и их статусов;
  - live/historical logs;
  - список ожидающих run и причины ожидания;
  - pending approvals и итоги approve/deny.
- UI governance:
  - локализация системных сообщений/ошибок;
  - единые шаблоны диагностики и действий для Owner/оператора.

### Out of scope
- Расширение staff-консоли за рамки MVP-функций (template editor 2.0, agent constructor, analytics studio).

## Критерии приемки
- Ключевые MVP-сценарии staff-консоли работают на новой Vuetify-базе без регрессии по API.
- Сценарии feedback/approve/deny/voice-STT воспроизводимы и документированы.
- UX-паттерны консоли единообразны и готовы к последующему масштабированию post-MVP.
