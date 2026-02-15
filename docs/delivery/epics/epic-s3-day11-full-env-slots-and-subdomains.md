---
doc_id: EPC-CK8S-S3-D11
type: epic
title: "Epic S3 Day 11: Full-env slot namespace + subdomain templating (TLS) + agent run"
status: planned
owner_role: EM
created_at: 2026-02-13
updated_at: 2026-02-15
related_issues: [19]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
---

# Epic S3 Day 11: Full-env slot namespace + subdomain templating (TLS) + agent run

## TL;DR
- Цель: сделать full-env слоты полностью самодостаточными и пригодными для manual QA.
- MVP-результат: по webhook-triggered full-env деплою поднимается изолированный namespace слота (вся инфра/сервисы/БД), запускается агент внутри этого namespace, а слот доступен по отдельному HTTPS поддомену с валидным сертификатом.

## Priority
- `P0`.

## Контекст
- В Day9 введен typed `services.yaml` и persisted reconcile-контур `runtime_deploy_tasks` для full-env.
- Чтобы полноценно использовать full-env режим для dogfooding и внешних проектов, нужен:
  - запуск agent-run внутри slot namespace (как часть full-env исполнения),
  - детерминированный рендер и выдача public URL для слота (поддомен + TLS),
  - отсутствие cluster-scope конфликтов между prod/ai-staging и ai-слотами.

## Scope
### In scope
- Full-env runtime: в slot namespace разворачиваются все `infrastructure` и `services` из `services.yaml` в правильном порядке (stateful -> migrations -> internal -> edge -> frontend).
- Agent-run внутри slot namespace:
  - после readiness сервисов создается `Job`/`Pod` с `agent-runner` (или эквивалентный runtime workload);
  - агент использует те же политики и аудит, что и обычный `run:*` контур;
  - результат работы сохраняется в БД и виден через staff UI.
- Шаблонизация поддоменов для full-env slot namespaces:
  - контракт `services.yaml` расширяется полем уровня окружения (MVP): `environments.<env>.domainTemplate`.
  - шаблон использует контекст рендера: `.Project`, `.Env`, `.Slot`, `.Namespace`.
  - runtime deploy резолвит host в `CODEXK8S_STAGING_DOMAIN` и `CODEXK8S_PUBLIC_BASE_URL` перед рендером манифестов.
  - ingress манифесты используют резолвленный host, а oauth2-proxy redirect URL всегда соответствует этому host.
- TLS и cert-manager:
  - `ClusterIssuer` (Let’s Encrypt) считается bootstrap-only и не применяется в runtime deploy.
  - для слотов создается только namespaced `Ingress` + `Certificate` (через аннотацию), используя общий `ClusterIssuer`.
- Manual QA маршрут:
  - после деплоя слот доступен по URL вида, полученного из `domainTemplate` (HTTPS);
  - staff UI и runbook содержат команды проверки (ingress/cert/job/logs).

### Out of scope
- Автоматизация DNS провайдера (создание wildcard-записей) и DNS01-валидация.
- Переход на wildcard-сертификаты для всех слотов.
- `vcluster`/nested clusters.

## Критерии приемки
- Full-env деплой по webhook создает/обновляет slot namespace и выводит его public URL (host) детерминированно.
- В slot namespace:
  - инфраструктура и сервисы развернуты и готовы;
  - агент запущен и пишет артефакты/логи/статусы;
  - повторный reconcile идемпотентен.
- Поддомен слота:
  - соответствует `domainTemplate`;
  - резолвится в ingress (предусловие: wildcard DNS настроен);
  - cert-manager выпускает сертификат и `kubectl get certificate` показывает `Ready=True`;
  - слот открывается в браузере и доступен для manual QA.
- Slot mode не пытается создавать/менять cluster-scoped ресурсы (в т.ч. `ClusterIssuer`), чтобы исключить конфликты с ai-staging/prod.
