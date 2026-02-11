---
doc_id: EPC-CK8S-S2-D7
type: epic
title: "Epic S2 Day 7: Dogfooding regression and release gate"
status: planned
owner_role: EM
created_at: 2026-02-10
updated_at: 2026-02-12
related_issues: []
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
---

# Epic S2 Day 7: Dogfooding regression and release gate

## TL;DR
- Цель эпика: зафиксировать “работает end-to-end” для dogfooding цикла.
- Ключевая ценность: дальше можно расширять label-набор и переносить больше этапов разработки в платформу.
- MVP-результат: regression матрица пройдена, есть go/no-go и готовый backlog на включение остальных `run:*` этапов.

## Priority
- `P0`.

## Scope
### In scope
- Прогон сценариев:
  - `issues.labeled(run:dev)` -> run -> job -> PR;
  - `issues.labeled(run:dev:revise)` -> revise -> update PR;
  - отказ запуска при отсутствии прав.
- Регрессионный сценарий Day4 E2E:
  - создание ветки и изменения файла агентом;
  - создание PR;
  - добавление комментария на правку;
  - повторный запуск `run:dev:revise` и проверка устранения замечаний в том же PR.
- Проверка утечек слотов/namespaces/job объектов.
- Обновление runbook/smoke checklist.

## Критерии приемки эпика
- End-to-end проходит на staging.
- Нет известных P0 блокеров для продолжения dogfooding.
- Для Day4 сценария есть воспроизводимые evidence (flow_events + links + PR timeline).
