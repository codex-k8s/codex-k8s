---
doc_id: SPR-CK8S-0001
type: sprint-plan
title: "Sprint S1: MVP vertical slice (Day 0..7)"
status: completed
owner_role: EM
created_at: 2026-02-06
updated_at: 2026-02-24
related_issues: [1]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-19-full-docset"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-19
---

# Sprint S1: MVP vertical slice (Day 0..7)

## TL;DR
- Спринт фиксирует 8 эпиков и закрыт по результатам поставки `Day 0..7`.
- Режим работы: каждый день закрываем задачи, сливаем в `main`, автоматически деплоим в `production`, делаем ручной smoke.
- Source of truth требований: `docs/product/requirements_machine_driven.md`.
- Source of truth процесса: `docs/delivery/development_process_requirements.md`.

## Цель спринта
- Довести платформу до рабочего MVP vertical slice: webhook -> run -> worker -> k8s execution -> status в UI.
- Удержать ежедневную поставку на текущий production без пересоздания окружения.
- Сохранить возможность полного bootstrap с чистого Ubuntu 24.04 как подтверждённый `Day 0`.

## План эпиков по дням

| День | Эпик | Priority | Ожидаемые артефакты дня | Документ | Статус |
|---|---|---|---|---|---|
| Day 0 | Baseline bootstrap complete | P0 | Bootstrap scripts + deploy baseline + подтвержденный production bootstrap | `docs/delivery/epics/s1/epic-s1-day0-bootstrap-baseline.md` | completed |
| Day 1 | Webhook ingress + idempotency | P0 | Webhook endpoint + signature verify + dedup + production smoke evidence | `docs/delivery/epics/s1/epic-s1-day1-webhook-idempotency.md` | completed |
| Day 2 | Worker run loop + slots + k8s jobs | P0 | Worker execution loop + slot leasing + run status transitions | `docs/delivery/epics/s1/epic-s1-day2-worker-slots-k8s.md` | completed |
| Day 3 | OAuth/JWT + project RBAC + minimal staff UI | P0 | OAuth/JWT auth + RBAC middleware + minimal UI screens | `docs/delivery/epics/s1/epic-s1-day3-auth-rbac-ui.md` | completed |
| Day 4 | Repository provider + project repositories lifecycle | P0 | RepositoryProvider + GitHub adapter + repository CRUD | `docs/delivery/epics/s1/epic-s1-day4-repository-provider.md` | completed |
| Day 5 | Learning mode MVP (prompt augmentation + storage) | P1 | Learning toggle + augmentation + feedback persistence | `docs/delivery/epics/s1/epic-s1-day5-learning-mode.md` | completed |
| Day 6 | Security/Network/Observability hardening for production | P1 | DNS/TLS/firewall checks + observability baseline | `docs/delivery/epics/s1/epic-s1-day6-hardening-observability.md` | completed |
| Day 7 | Stabilization, regression, release gate for next sprint | P0 | Regression report + go/no-go + Sprint S2 backlog draft | `docs/delivery/epics/s1/epic-s1-day7-stabilization-gate.md` | completed |

## Ежедневный delivery-гейт (обязательно)
- Изменения дня влиты в `main`.
- CI pipeline зеленый.
- Production автообновился и зафиксирован `deployed revision`.
- Выполнен ручной smoke-check по runbook.
- Обновлены документы (если менялись API, data model, webhook flow, RBAC, `services.yaml`, MCP контракты).

## Data model governance для спринта
- Для каждого эпика обязательно заполняется раздел `Data model impact` по структуре шаблона `docs/templates/data_model.md`:
  - сущности/инварианты;
  - связи/FK;
  - критичные индексы и запросы;
  - миграции;
  - retention/PII.
- Любая миграция должна иметь rollback-подход и связь с DoD эпика.

## Риски спринта
- Нестабильность production из-за ежедневных инкрементов.
- Регрессии webhook-run pipeline при быстром темпе слияний.
- Расхождение документации и фактической схемы БД при частых миграциях.

## Апрув
- request_id: owner-2026-02-06-sprint-s1
- Решение: approved
- Комментарий: Спринт S1 утверждён, режим ежедневного деплоя на production обязателен.
