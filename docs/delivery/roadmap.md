---
doc_id: RDM-CK8S-0001
type: roadmap
title: "codex-k8s — Roadmap"
status: draft
owner_role: PM
created_at: 2026-02-06
updated_at: 2026-02-13
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-06-mvp"
---

# Roadmap: codex-k8s

## TL;DR
- Q1: foundation + core backend + bootstrap staging.
- Q2: dogfooding, approval/audit hardening и MVP completion (`run:self-improve`, MCP control tools, runtime debug UI).
- Q3: production readiness + масштабирование governance и knowledge-платформы.
- Q4: расширяемость custom-агентов, A2A swarm и периодические автономные циклы улучшения.

## Принципы приоритизации
- Сначала контроль рисков и deployability.
- Затем продуктовые возможности и масштабирование.
- Избегать расширений, которые ломают webhook-driven core.

## Roadmap (high-level)
| Период | Инициатива | Цель | Метрики | Статус |
|---|---|---|---|---|
| Q1 | MVP core + staging bootstrap | запустить рабочий staging и ручные тесты | one-command bootstrap, green deploy from main | planned |
| Q2 | Dogfooding + MVP completion | довести `run:*` контур до полного stage-цикла, добавить MCP control tools и `run:self-improve` | >=95% run:dev и >=85% self-improve проходят без ручного обхода policy | in-progress |
| Q3 | Stage coverage + production readiness | усилить release/postdeploy gate, стабилизировать governance, подготовить prod hardening | prod runbook + approval latency SLO + full stage traceability | planned |
| Q4 | Extensibility and autonomy | custom-агенты, A2A swarm, периодические автозапуски улучшений/проверок | configurable agent factory + scheduled autonomous runs | planned |

## Backlog кандидатов
- Contract-first OpenAPI rollout completion: полное покрытие active external/staff API + строгая CI-проверка codegen.
- Split control-plane по внутренним сервисам при росте нагрузки.
- Vault/KMS интеграция вместо хранения repo token material в БД.
- Расширенная политика workflow approvals.
- Автоматическое управление labels-as-vars через staff UI.
- Квоты и policy packs для custom-агентов по проектам.
- Расширение i18n prompt templates: добавление locale + авто-перевод шаблонов через ИИ.
- Система управления prompt templates/agent parameters через UI (versioning, preview, rollback, diff).
- Конструктор агентов в web-консоли (роль, права, режим, лимиты, набор инструментов).
- Управление label taxonomy и stage policy через UI с change audit.
- Систематизированное хранение проектной документации (repo + DB) с операциями через MCP.
- Индексация документации в `pgvector` и MCP ручки для semantic retrieval/change impact analysis.
- Полноценная web-console на компонентной UI-библиотеке (кандидат baseline: `Vuetify` для Vue 3 admin-паттернов, проверено по Context7).
- Рой агентов с A2A шиной для параллельной работы над одной задачей и синхронизации решений.
- Периодические автозапуски (`security scan`, `dependency freshness`, `ops drift`, `self-improve` cadence).

## Риски roadmap
- Задержка из-за инфраструктурной автоматизации bootstrap.
- Недостаточная зрелость security baseline при раннем масштабировании.

## Апрув
- request_id: owner-2026-02-06-mvp
- Решение: approved
- Комментарий: Дорожная карта этапов MVP утверждена Owner.
