---
doc_id: EPC-CK8S-S3-D9
type: epic
title: "Epic S3 Day 9: Declarative full-env deploy and runtime parity"
status: planned
owner_role: EM
created_at: 2026-02-13
updated_at: 2026-02-14
related_issues: [19]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "approved-day9-rework"
---

# Epic S3 Day 9: Declarative full-env deploy and runtime parity

## TL;DR
- Цель: ввести универсальный контракт `services.yaml` (любой стек, любой проект) и общий движок рендера/планирования для `control-plane` и bootstrap binary.
- Ключевая ценность: единый source of truth для full-env runtime, dogfooding `codex-k8s`, prompt context и первичного развертывания платформы.
- MVP-результат: webhook-driven `codex-k8s` разворачивает окружения по typed execution-plan из `services.yaml`; shell-скрипты остаются thin-wrapper.

## Priority
- `P0`.

## Контекст
- В `codexctl` orchestration строился вокруг `services.yaml` и GitHub workflows; в `codex-k8s` source-of-truth остаётся `services.yaml`, но запуск теперь webhook-driven через `control-plane`.
- Для MVP требуется одинаковая логика:
  - для любых внешних проектов (любой стек);
  - для dogfooding проекта `codex-k8s` (саморазвёртывание в ai-slot namespace).
- Нужен одноразовый bootstrap binary для первичной инициализации чистого Ubuntu 24.04 сервера: подготовка Kubernetes/зависимостей/секретов/репозитория, деплой `codex-k8s`, передача управления `control-plane`.

## Принятые решения по результатам брейншторма
- `R1` (выбрано `1-1`): в `services.yaml` используется enum-поле `codeUpdateStrategy` со значениями `hot-reload | rebuild | restart` (вместо bool-флага).
- `R2` (выбрано `2-1`): изоляция dogfooding выполняется namespace-уровнем + policy/валидацией cluster-scope ресурсов (без `vcluster`/nested cluster в MVP).
- `R3` (выбрано `3-1`): переиспользование блоков реализуется через собственные `imports + components + deep-merge` в typed render engine (без перехода на Helm как базовый движок).
- `R4`: для проекта `codex-k8s` окружение `ai-staging` в `services.yaml` задаётся шаблоном `{{ .Project }}-ai-staging`; для `project=codex-k8s` это даёт namespace `codex-k8s-ai-staging`.

## Scope
### In scope
- Контракт `services.yaml v2`:
  - универсальная модель окружений/инфраструктуры/сервисов/образов/хуков/политик обновления кода;
  - поддержка `imports`, `components`, детерминированного merge и schema-validation;
  - запрет `any`-подобных невалидируемых runtime-секций в core-контракте.
- Общая библиотека `services.yaml`:
  - единый loader/parser/validator/renderer/planner;
  - единый API для `control-plane` и bootstrap binary.
- Детерминированный порядок развёртывания:
  - `stateful dependencies -> migrations -> internal domain services -> edge services -> frontend`.
- Runtime parity для non-prod (`dev`, `ai-staging`, `ai-slot`):
  - `codeUpdateStrategy=hot-reload` поддерживается на уровне Dockerfile/manifests/entrypoints;
  - стратегии `rebuild` и `restart` поддерживаются в execution-plan и prompt context.
- Dogfooding без конфликтов:
  - `codex-k8s` в ai-slot разворачивает изолированную копию себя в отдельном namespace;
  - для `codex-k8s` `ai-staging` задаётся шаблоном `{{ .Project }}-ai-staging`.
- Bootstrap binary (одноразовый):
  - настройка чистого Ubuntu 24.04: зависимости, Kubernetes, базовые скрипты/секреты/env;
  - подготовка целевого GitHub-репозитория проекта;
  - деплой `codex-k8s` и handoff в webhook-driven `control-plane`.
- E2E маршрут на отдельном чистом VPS через `bootstrap/host/config-e2e-test.env`.

### Out of scope
- Внедрение `vcluster`/nested cluster в MVP.
- Полная деактивация всех shell scripts за пределами day9-объёма.
- FinOps/production performance tuning.

## Целевой контракт `services.yaml v2` (MVP-срез)
- Верхний уровень:
  - `apiVersion`, `kind`, `metadata`, `spec`.
- Обязательные блоки в `spec`:
  - `environments` (inheritance через `from`, namespace template, runtime flags);
  - `images` (external/build/mirror policy);
  - `infrastructure` и `services` (typed deploy units, dependencies, hooks);
  - `orchestration` (deploy order, readiness strategy, cleanup/ttl policy).
- Новые обязательные поля:
  - `services[].codeUpdateStrategy` enum: `hot-reload | rebuild | restart`;
  - `imports[]` и `components[]` для переиспользования;
  - `instanceScope`/эквивалентный runtime marker для anti-conflict policy в dogfooding.
- Правило для `codex-k8s`:
  - `environments.ai-staging.namespaceTemplate` задаётся как `{{ .Project }}-ai-staging`.

## Детерминированный render pipeline
1. Load root config.
2. Resolve `imports` (с детектом циклов/дубликатов).
3. Построить итоговый AST через `components + deep-merge`.
4. Schema validation.
5. Resolve environment inheritance (`from`) и defaults.
6. Resolve template context (`project/env/slot/namespace/vars`).
7. Resolve namespace and image refs.
8. Build deploy graph (infra/services/hooks/migrations/dependencies).
9. Validate graph (cycles, unknown refs, forbidden cluster-scope for slot mode).
10. Emit typed execution plan для runtime/bootstrap.

## Декомпозиция (Stories/Tasks)
- Story-1: Спецификация `services.yaml v2` + JSON Schema + миграционные правила совместимости.
- Story-2: Общая Go-библиотека для `services.yaml` (`load/validate/render/plan`) в `libs/go/*`.
- Story-3: Render engine с `imports/components/deep-merge`, детектом циклов и строгими ошибками precondition.
- Story-4: Интеграция execution-plan в `control-plane`/`worker` (webhook-driven путь, без workflow-first допущений).
- Story-5: Интеграция prompt context: экспорт `codeUpdateStrategy`, runtime hints и resolved service inventory.
- Story-6: Bootstrap binary для первичной установки Ubuntu 24.04 + Kubernetes + deploy `codex-k8s` + handoff.
- Story-7: Dogfooding safeguards:
  - namespace isolation для ai-slot;
  - шаблон для `codex-k8s ai-staging`: `{{ .Project }}-ai-staging`;
  - блокировка конфликтующих cluster-scope ресурсов в slot-профиле.
- Story-8: Full E2E на новом чистом VPS:
  - входной конфиг: `bootstrap/host/config-e2e-test.env`;
  - сценарий: установка зависимостей на Ubuntu 24.04, поднятие Kubernetes, деплой `codex-k8s`, проверка webhook-driven lifecycle;
  - отдельный пустой GitHub repo проекта-примера подключается в e2e и проходит provisioning/deploy smoke.

## Критерии приемки
- Для минимум двух проектов (`project-example` и `codex-k8s`) full-env поднимается из `services.yaml` через typed execution-plan.
- `services.yaml v2` поддерживает `imports/components/deep-merge`, имеет schema-validation и покрыт unit/integration тестами.
- `codeUpdateStrategy` (enum) присутствует в контракте, учитывается в runtime orchestration и попадает в prompt context.
- Для `codex-k8s` подтверждено правило шаблона: `ai-staging` задаётся как `{{ .Project }}-ai-staging` и для `project=codex-k8s` резолвится в `codex-k8s-ai-staging`.
- Для ai-slot dogfooding подтверждено отсутствие конфликтов со staging/prod платформой (namespace и runtime resources).
- Bootstrap binary успешно отрабатывает e2e на чистом Ubuntu 24.04 с входом `bootstrap/host/config-e2e-test.env`.
- По e2e опубликован evidence bundle:
  - команды/логи bootstrap;
  - состояние k8s ресурсов;
  - webhook ingest + run lifecycle evidence;
  - ссылка на тестовый репозиторий и итоговый smoke-check.

## Риски/зависимости
- Зависимость от готовности отдельного чистого VPS и заполненного `bootstrap/host/config-e2e-test.env`.
- Зависимость от тестового пустого GitHub repository для project-example e2e контура.
- Риск регрессий при миграции со shell-first на execution-plan путь; нужен dual-run/feature-flag rollout.
- Риск несогласованности namespace policy и текущих staging манифестов; нужен отдельный preflight check.

## План релиза (верхний уровень)
- Wave-1: спецификация и библиотека `services.yaml v2`.
- Wave-2: runtime интеграция (`control-plane`/`worker`) + prompt context.
- Wave-3: bootstrap binary + preflight checks.
- Wave-4: dogfooding policy + e2e на новом VPS + evidence.

## Апрув
- request_id: approved-day9-rework
- Решение: approved
- Комментарий: включает выбранные решения `1-1`, `2-1`, `3-1`, обязательный Story-8 e2e и namespace правило для `codex-k8s ai-staging`.
