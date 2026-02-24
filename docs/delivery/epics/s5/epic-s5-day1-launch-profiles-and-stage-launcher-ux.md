---
doc_id: EPC-CK8S-S5-D1
type: epic
title: "Epic S5 Day 1: Launch profiles and deterministic next-step actions (Issue #154)"
status: planned
owner_role: PM
created_at: 2026-02-24
updated_at: 2026-02-24
related_issues: [154]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-24-issue-154-day1"
---

# Epic S5 Day 1: Launch profiles and deterministic next-step actions (Issue #154)

## TL;DR
- Проблема: текущий label-driven flow неудобен для ручного управления; пользователи забывают порядок этапов, а часть быстрых ссылок не работает.
- Цель Day1: формализовать profile-driven процесс запуска и задать приемочные критерии для стабильного next-step UX.
- Результат Day1: подготовлен owner-ready execution package для `run:dev` без изменения архитектурных границ.

## Priority
- `P0`.

## Scope
### In scope
- Ввести launch profiles:
  - `quick-fix` для минимальных исправлений;
  - `feature` для функциональных доработок;
  - `new-service` для полного цикла с архитектурной проработкой.
- Зафиксировать mapping profile -> обязательные стадии -> условия эскалации.
- Формализовать deterministic next-step actions:
  - primary: staff web-console deep-link;
  - fallback: текстовая команда label transition (`gh`/MCP path) для copy-paste.
- Сформировать acceptance matrix для UX и policy поведения.

### Out of scope
- Реализация UI/Backend кода.
- Пересмотр базовых label-классов и security policy.
- Изменения процесса ревью вне связки stage launch / stage transition.

## Stories (handover в `run:dev`)
- Story-1: Profile Registry
  - хранение и выдача канонических профилей;
  - валидация допустимых переходов между profile-режимами.
- Story-2: Next-step Action Contract
  - единый contract для action cards с primary/fallback каналами;
  - обязательная диагностика при невалидном deep-link.
- Story-3: Service-message Rendering
  - profile-aware описание шага и следующего действия;
  - явный fallback-блок с безопасной командой без секретов.
- Story-4: Governance Guardrails
  - запрещение silent-skip этапов;
  - автоматическая постановка `need:input` при ambiguity profile/stage.
- Story-5: Traceability Sync
  - синхронизация `issue_map`, `requirements_traceability`, sprint/epic артефактов.

## Quality gates
- Planning gate:
  - FR-053/FR-054 отражены в продуктовых документах;
  - launch profile matrix согласована с `stage_process_model`.
- Contract gate:
  - next-step action contract детерминирован и допускает только policy-safe transitions.
- UX gate:
  - сценарий broken/dead link не блокирует owner-flow;
  - fallback-путь выполняется в один шаг (copy-paste).
- Security gate:
  - fallback-команды не содержат секретов;
  - любой transition сохраняет audit trail.
- Traceability gate:
  - изменения отражены в `issue_map` и `requirements_traceability`.

## Acceptance criteria
- [ ] Для каждого профиля (`quick-fix`, `feature`, `new-service`) определены вход, выход и эскалация.
- [ ] Каждая next-step подсказка включает `primary + fallback` и не блокирует переход при отказе primary.
- [ ] В ambiguous-сценариях платформа не делает best-guess переход, а требует `need:input`.
- [ ] Подготовлен список рисков и продуктовых допущений для `run:dev` этапа.

## Открытые риски
- Риск UX-перегрузки: слишком много действий в service-comment снижает читаемость.
- Риск policy-drift: profile shortcut может быть использован для обхода обязательных этапов.
- Риск интеграции: fallback-команды могут расходиться с фактической policy, если не централизовать шаблон.

## Продуктовые допущения
- Launch profiles не заменяют канонический stage-pipeline, а задают управляемые “траектории входа”.
- Primary канал (`web-console`) может временно быть недоступен; fallback обязан быть достаточным для продолжения процесса.
- Для P0/P1 инициатив Owner может принудительно переводить задачу в `new-service` траекторию независимо от исходного профиля.
