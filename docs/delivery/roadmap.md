---
doc_id: RDM-CK8S-0001
type: roadmap
title: "codex-k8s — Roadmap"
status: draft
owner_role: PM
created_at: 2026-02-06
updated_at: 2026-02-11
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-06-mvp"
---

# Roadmap: codex-k8s

## TL;DR
- Q1: foundation + core backend + bootstrap staging.
- Q2: dogfooding и stage-driven execution через labels (`run:*`), с observability/audit.
- Q3: production readiness + full stage coverage (`intake..ops`) + GitLab provider onboarding.
- Q4: расширяемость custom-агентов и шаблонов процессов/промптов.

## Принципы приоритизации
- Сначала контроль рисков и deployability.
- Затем продуктовые возможности и масштабирование.
- Избегать расширений, которые ломают webhook-driven core.

## Roadmap (high-level)
| Период | Инициатива | Цель | Метрики | Статус |
|---|---|---|---|---|
| Q1 | MVP core + staging bootstrap | запустить рабочий staging и ручные тесты | one-command bootstrap, green deploy from main | planned |
| Q2 | Dogfooding via labels | довести `run:dev`/`run:dev:revise` до устойчивого E2E + заложить полный label taxonomy | >=95% run:dev запускаются через Issue label без ручного вмешательства | in-progress |
| Q3 | Stage coverage + production readiness | включить этапы `run:intake..run:ops`, усилить release/postdeploy gate | prod runbook + approval gates + full stage traceability | planned |
| Q4 | Extensibility | custom-агенты per project + управляемые prompt templates | configurable custom roles + template policies in UI/API | planned |

## Backlog кандидатов
- Contract-first OpenAPI rollout completion: полное покрытие active external/staff API + строгая CI-проверка codegen.
- Split control-plane по внутренним сервисам при росте нагрузки.
- Vault/KMS интеграция вместо хранения repo token material в БД.
- Расширенная политика workflow approvals.
- Автоматическое управление labels-as-vars через staff UI.
- Квоты и policy packs для custom-агентов по проектам.
- Расширение i18n prompt templates: добавление locale + авто-перевод шаблонов через ИИ.

## Риски roadmap
- Задержка из-за инфраструктурной автоматизации bootstrap.
- Недостаточная зрелость security baseline при раннем масштабировании.

## Апрув
- request_id: owner-2026-02-06-mvp
- Решение: approved
- Комментарий: Дорожная карта этапов MVP утверждена Owner.
