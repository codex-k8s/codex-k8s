---
doc_id: CHR-CK8S-0119
type: charter
title: "Issue #119 — E2E A+B Core Lifecycle Charter"
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

# Project Charter: Issue #119 — E2E A+B Core Lifecycle

## TL;DR
- **Миссия:** подтвердить, что A/B ядро e2e master plan стабильно воспроизводимо и не регрессирует review-driven revise.
- **Цели:** пройти наборы A1–A3 и B1–B3 без P0/P1, зафиксировать корректные label/state transitions и собрать evidence.
- **Метрики успеха:** 100% прохождения A/B с ожидаемыми переходами; evidence bundle опубликован в Issue #118.
- **MVP-границы:** только A/B ядро (без C–F), без расширения runtime/deploy.
- **Ключевые риски:** нестабильность окружения, ambiguity labels, неполные audit/evidence.

## Миссия (Mission Statement)
Проверить критический e2e-контур `run:*` стадий и review-driven revise, чтобы закрыть риск регрессий перед финальным MVP gate.

## Цели и ожидаемые результаты
1) Пройти A1/A2/A3 (stage lifecycle + dev/revise + qa/release/postdeploy/ops).
2) Подтвердить B1/B2/B3 (review-driven revise, ambiguity handling, sticky model/reasoning).
3) Зафиксировать audit/evidence и обновить трассируемость (Issue #118 + issue_map).

## Декомпозиция инкрементов
1) Incr-1: PRD-пакет issue #119 (`prd.md`, `nfr.md`, `user_story.md`) с формальными AC и DoD.
2) Incr-2: Прогон A-сценариев (A1-A3) и фиксация pass/fail + переходов в evidence.
3) Incr-3: Прогон B-сценариев (B1-B3), включая ambiguity и sticky profile проверки.
4) Incr-4: Публикация evidence bundle в Issue #118 и синхронизация traceability в delivery-документах.

## Пользователи / Стейкхолдеры
- Основные пользователи: Owner, QA, EM, PM.
- Стейкхолдеры: контроль качества MVP, release gate.
- Владелец решения (Owner): Owner.

## Область (Scope)
### MVP Scope
- Наборы A1–A3 и B1–B3 из `docs/delivery/e2e_mvp_master_plan.md`.
- Проверка ожидаемых label/state/need переходов.
- Evidence bundle в Issue #118.

### Не делаем (Non-goals)
- Наборы C–F (governance tools, runtime/security, multi-repo) для Issue #119.
- Изменения архитектуры, кода, инфраструктуры или процесса deploy.

## Продуктовые принципы/ограничения
- Следовать stage-модели и label policy без ручных обходов.
- Только markdown-изменения (policy scope для `run:vision`).
- Evidence фиксируется публично и воспроизводимо.

## Предположения
- Label resolver и review-driven revise уже реализованы (Issue #95).
- E2E master plan актуален (updated_at: 2026-02-24).
- Доступ к Issue #118 для публикации evidence открыт.

## Риски
- Амбигуитет label-состояния может блокировать revise.
- Нестабильность full-env окружения приведет к ложным отказам.
- Неполный audit trail затруднит приемку.

## Депенденси
- `docs/delivery/e2e_mvp_master_plan.md` как source of truth для наборов A/B.
- ADR-0006 и политика labels/stage.
- Доступ к production-like среде для прогона.

## Требования верхнего уровня (черновик)
- Функциональные: A1–A3, B1–B3, ожидаемые label/state transitions.
- Нефункциональные: отсутствие P0/P1 регрессий, полнота audit/evidence.
- Наблюдаемость/эксплуатация: наличие flow_events и ссылок на run/pr/issue.

## План принятия решений и апрувов
- Артефакты на апрув: этот charter, метрики успеха, risk register, evidence bundle.
- Канал апрува: стандартный review Owner.

## Следующий шаг
- [ ] Подтвердить метрики успеха и risk register.
- [ ] Запустить прогоны A/B и собрать evidence в Issue #118.

## Апрув
- Запрошен: 2026-02-24 (request_id: owner-2026-02-24-issue-119-vision)
- Решение: pending
- Комментарий:
