---
doc_id: RSK-CK8S-0119
type: risks
title: "Issue #119 — Risk Register (E2E A+B)"
status: draft
owner_role: PM
created_at: 2026-02-24
updated_at: 2026-02-24
related_issues: [119]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-24-issue-119-vision"
---

# Risk Register: Issue #119 — E2E A+B

## TL;DR
Топ-3 риска: нестабильность full-env окружения, ambiguity labels, неполный evidence/audit.

## Реестр рисков
| ID | Риск | Причина | Вероятность | Влияние | Митигирующие действия | Триггер/сигнал | Владелец | Статус |
|---|---|---|---|---|---|---|---|---|
| R1 | Нестабильность full-env слота | Зависимость от runtime deploy и состояния кластера | M | H | Прогонять в стабильном окне; фиксировать namespace/лог evidence | Ошибки readiness/job failures | SRE/QA | open |
| R2 | Ambiguity labels ломает revise | Конфликтующие stage labels на Issue/PR | M | H | Жёстко следовать policy, проверка labels до запуска B-сценариев | `need:input` без ожидаемого revise | PM/QA | open |
| R3 | Неполный audit trail | Недостаточная фиксация событий/логов | M | H | Чек-лист evidence + сверка flow_events | Отсутствие ключевых `run.*`/`approval.*` событий | PM/KM | open |
| R4 | Sticky model/reasoning не сохраняется | Профиль переопределяется или сбрасывается | L | M | Явно фиксировать labels и проверять profile resolver в evidence | Несоответствие профиля между итерациями | PM | open |
| R5 | Несогласованность traceability | issue_map/e2e plan не обновлены | L | M | Обновлять docs синхронно с прогоном | Отсутствуют ссылки на evidence | KM | open |

## Риски безопасности (если есть)
- Утечки секретов в логах/комментариях при сборе evidence.

## Технические долги/компромиссы
- Ограничение Issue #119 только наборами A/B оставляет риски C–F на следующих прогонах.

## Апрув
- request_id: owner-2026-02-24-issue-119-vision
- Решение: pending
- Комментарий:
