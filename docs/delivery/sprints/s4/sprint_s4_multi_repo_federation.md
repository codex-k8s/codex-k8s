---
doc_id: SPR-CK8S-0004
type: sprint-plan
title: "Sprint S4: Multi-repo runtime and docs federation execution (Issue #100)"
status: completed
owner_role: EM
created_at: 2026-02-23
updated_at: 2026-02-23
related_issues: [100, 106]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-23-issue-100-sprint-plan"
---

# Sprint S4: Multi-repo runtime and docs federation execution (Issue #100)

## TL;DR
- Цель спринта: перевести дизайн multi-repo режима (Issue #100) в исполняемый delivery-контур с прозрачными quality-gates и измеримыми критериями завершения.
- Базовый технический выбор: federated composition с единым `effective services.yaml` на запуск и repo-aware docs federation.
- Результат Day1: owner-ready execution package для `run:dev` без изменения архитектурных границ платформы.

## Scope спринта
### In scope
- Поддержка трех runtime режимов `services.yaml`:
  - монорепо (single root);
  - per-repo manifests;
  - гибридный orchestrator + imports.
- Поддержка трех режимов размещения документации:
  - выделенный docs repo;
  - docs рядом с сервисами;
  - комбинированный режим.
- Декомпозиция implementation backlog по `control-plane`, `worker`, `agent-runner`, staff API/traceability.
- Подготовка quality-gates для contract/runtime/security/regression.

### Out of scope
- Пересмотр базовой архитектурной модели зон `external/internal/jobs`.
- Изменение policy ограничений по RBAC и secret access.

## План эпиков по дням

| День | Эпик | Priority | Документ | Статус |
|---|---|---|---|---|
| Day 1 | Multi-repo composition and docs federation execution foundation | P0 | `docs/delivery/epics/s4/epic-s4-day1-multi-repo-composition-and-docs-federation.md` | completed |

## Daily gate (обязательно)
- Кодовые и документационные изменения синхронизированы в рамках одного PR.
- Обновлены traceability документы (`issue_map`, `requirements_traceability`) и верхнеуровневый delivery план.
- Зафиксированы блокеры/риски/owner decisions для handover в `run:dev`.

## Completion критерии спринта
- [x] Выбранный вариант реализации (federated composition) зафиксирован как целевой.
- [x] Есть execution-plan с stories, quality-gates и acceptance criteria.
- [x] Для всех кейсов A..F из Issue #100 задан детерминированный путь проверки.
- [x] Подготовлен handover пакет для `dev`/`qa`/`sre` без открытых P0 неопределённостей.

## Итог Day 1 (Issue #106)
- Execution-package Sprint S4 Day1 закрыт в формате docs + code.
- Статусы и факты выполнения синхронизированы в delivery и traceability документах.
- Реализованы Story-1 и Story-5 (repository topology + repo-aware docs federation), оставшиеся story зафиксированы в backlog.
- Следующие day-эпики S4 формируются на основе текущего code baseline и quality-gates execution-пакета.

## Handover после закрытия Day1
- `dev`: реализация stories execution-пакета в `run:dev` с runtime evidence.
- `qa`: test-design для кейсов A..F + негативные сценарии конфликтов/import cycles/rate-limits.
- `sre`: оценка capacity impact и операционных метрик multi-repo checkout/reconcile.
- `km`: поддержание актуальной трассируемости и синхронизация linked docs.
