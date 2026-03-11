---
doc_id: EPC-CK8S-S6-D10
type: epic
title: "Epic S6 Day 10: Postdeploy review для lifecycle управления агентами и шаблонами промптов (Issue #263)"
status: in-review
owner_role: SRE
created_at: 2026-03-02
updated_at: 2026-03-02
related_issues: [184, 185, 187, 189, 195, 197, 199, 201, 216, 262, 263, 265]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-02-issue-263-postdeploy"
---

# Epic S6 Day 10: Postdeploy review для lifecycle управления агентами и шаблонами промптов (Issue #263)

## TL;DR
- Postdeploy-проверка runtime в текущем full-env контуре (`codex-k8s-dev-1`) прошла без blocker-инцидентов.
- Критичные сервисы (`api-gateway`, `control-plane`, `worker`, `web-console`, `postgres`) в `Running`, restart count = `0`.
- Подтверждена readiness базовых health-checks (`/healthz`, `/health/readyz`, `pg_isready`) и отсутствие деградации в логах запуска.
- Для следующего этапа `run:ops` подготовлен отдельный операционный handover: runbook + monitoring + alerts + SLO + rollback checks.

## Контекст
- Вход из `run:release` (`#262`): release closeout и continuity rule для обязательного postdeploy stage.
- Stage continuity Sprint S6:
  `#184 -> #185 -> #187 -> #189 -> #195 -> #197 -> #199 -> #201 -> #216 -> #262 -> #263`.
- Scope `run:postdeploy` ограничен markdown-only политикой (`*.md`).

## Scope
### In scope
- Runtime-диагностика в namespace активного запуска (`codex-k8s-dev-1`).
- Фиксация postdeploy evidence (health/logs/events/services/jobs).
- Уточнение операционных рисков и handover в `run:ops`.
- Синхронизация delivery/traceability документов Sprint S6.

### Out of scope
- Изменения кода, Kubernetes manifests, scripts, docker artifacts.
- Изменения архитектурных границ и transport contracts.

## Postdeploy evidence (2026-03-02)

| Проверка | Команда | Факт | Статус |
|---|---|---|---|
| Состояние workload | `kubectl -n codex-k8s-dev-1 get pods,deploy,job -o wide` | Все основные pod/deploy в `Running`, migration и mirror jobs в `Completed`, активный run-job выполняется | passed |
| Логи control-plane | `kubectl -n codex-k8s-dev-1 logs deploy/codex-k8s-control-plane --tail=200` | Успешный старт HTTP/gRPC, runtime deploy loop активен, ошибок startup после readiness нет | passed |
| Логи worker | `kubectl -n codex-k8s-dev-1 logs deploy/codex-k8s-worker --tail=200` | Worker стартовал штатно, без crash/retry | passed |
| Логи api-gateway | `kubectl -n codex-k8s-dev-1 logs deploy/codex-k8s --tail=200` | OpenAPI request validation включена, сервис поднят на `:8080` | passed |
| Сетевой контур | `kubectl -n codex-k8s-dev-1 get svc,ingress -o wide` | Сервисы и ingress присутствуют, DNS-host выдан для окружения | passed |
| Health API | `kubectl exec deploy/codex-k8s -- wget -qO- http://127.0.0.1:8080/healthz` | `alive` | passed |
| Readiness control-plane | `kubectl exec deploy/codex-k8s-control-plane -- wget -qO- http://127.0.0.1:8081/health/readyz` | `ok` | passed |
| Доступность PostgreSQL | `kubectl exec statefulset/postgres -- pg_isready` | `accepting connections` | passed |
| События namespace | `kubectl -n codex-k8s-dev-1 get events --sort-by=.lastTimestamp` | Зафиксированы единичные startup probe warnings на этапе старта; далее контейнеры перешли в stable `Running` | passed |

## Monitoring / Alerts / SLO / Rollback (handover)
- Операционный handover оформлен в `docs/ops/handovers/s6/postdeploy_ops_handover.md` и включает:
  - runbook-паттерн диагностики для postdeploy;
  - мониторинг ключевых сигналов (`availability`, `latency`, `errors`, `saturation`);
  - каталог алертов с anti-noise мерами (`for`, `keep_firing_for`);
  - SLO-impact оценку по результатам postdeploy;
  - rollback-процедуру и критерии активации.

## Риски и решения
| Тип | ID | Описание | Статус |
|---|---|---|---|
| risk | RSK-263-01 | На bootstrap-этапе зафиксированы transient startup probe warnings, что типично для холодного старта | monitoring |
| risk | RSK-263-02 | Возможна скрытая деградация под длительной production-нагрузкой вне текущего окна наблюдения | monitoring |
| decision | OD-263-01 | Postdeploy в пределах текущего scope признан успешным; переход в `run:ops` обязателен для операционного hardening | proposed |

## Context7 validation
- Через Context7 (`/websites/kubernetes_io`) подтверждено применение `startupProbe` для slow-start контейнеров и корректная семантика переключения на liveness/readiness после успешного старта.
- Через Context7 (`/prometheus/docs`, `/websites/prometheus_io`) подтверждены практики снижения noise для alerting (`for`, `keep_firing_for`) и привязка paging к user-impact метрикам.

## Handover в `run:ops`
- Следующий stage: `run:ops` через follow-up issue `#265` (создана без trigger-лейбла).
- Обязательные deliverables `run:ops`:
  - закрыть `RSK-263-01/02` через операционные улучшения наблюдаемости;
  - подтвердить стабильность SLO в окне postdeploy наблюдения;
  - синхронизировать ops-процедуры с production runbook.

## Связанные документы
- `docs/delivery/epics/s6/epic-s6-day9-release-closeout.md`
- `docs/delivery/epics/s6/epic_s6.md`
- `docs/delivery/sprints/s6/sprint_s6_agents_prompt_management.md`
- `docs/delivery/delivery_plan.md`
- `docs/ops/production_runbook.md`
- `docs/ops/handovers/s6/postdeploy_ops_handover.md`
