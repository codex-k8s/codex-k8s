---
doc_id: ADR-0007
type: adr
title: "Multi-repo composition and docs federation"
status: proposed
owner_role: SA
created_at: 2026-02-21
updated_at: 2026-02-21
related_issues: [100]
related_prs: []
supersedes: []
superseded_by: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-21-multi-repo-adr"
---

# ADR-0007: Multi-repo composition and docs federation

## TL;DR
- Проблема: текущая документация описывает FR-020 на уровне per-repo `services.yaml`, но не фиксирует единый runtime-подход для multi-repo deploy и docs federation.
- Решение: вводим федеративную модель с одним `effective services.yaml` на запуск, формируемым в `control-plane` из repo-aware источников.
- Последствия: поддерживаются все режимы (монорепо, per-repo, гибрид), при этом сохраняются детерминизм, аудит и текущие сервисные границы.

## Контекст

Issue #100 требует:
- разворачивать в dev-slot сервисы из разных репозиториев;
- поддерживать `services.yaml` как в одном repo, так и в каждом repo отдельно;
- поддерживать docs в выделенном docs repo, в service repos и в комбинированном режиме.

Текущий пробел:
- отсутствует зафиксированный механизм компоновки multi-repo `services.yaml`;
- не определён единый алгоритм role-aware docs контекста из нескольких repo.

## Decision drivers

- Детерминированный runtime execution-plan.
- Поддержка всех topologies без отдельной продуктовой ветки логики.
- Минимальное влияние на архитектурные зоны (`external`/`internal`/`jobs`).
- Auditability и policy-governed поведение.

## Рассмотренные варианты

### Вариант A: Только monorepo root (`services.yaml` в одном repo)

Плюсы:
- минимальная сложность реализации.

Минусы:
- не закрывает целевой multi-repo сценарий;
- противоречит intent FR-020.

### Вариант B: Полная автономия repo (каждый deploy только из локального `services.yaml`)

Плюсы:
- независимость команд по репозиториям.

Минусы:
- нет единого compose для cross-repo slot deploy;
- сложно гарантировать согласованный deploy order и ownership зависимостей.

### Вариант C (выбран): Federated composition

Суть:
- на каждый запуск формируется один `effective services.yaml`;
- источники могут быть распределены по repo (root/imports/virtual root);
- docs graph также repo-aware и role-aware.

Плюсы:
- закрывает все кейсы из Issue #100;
- сохраняет текущий execution model (control-plane resolve, worker reconcile);
- предсказуемо для audit/debug.

Минусы:
- выше сложность resolve/validation;
- требуется расширение metadata repositories и staff API.

## Решение

Выбираем **Вариант C**:
- федеративная компоновка runtime и docs с repo-aware контрактом;
- единый `effective manifest` обязателен для каждого запуска;
- deploy и docs-context резолвятся в `control-plane` с idempotent исполнением в `worker`.

Подробный дизайн:
- `docs/architecture/multi_repo_mode_design.md`.

## Последствия

### Позитивные
- Единая модель для monorepo/multi-repo/hybrid.
- Управляемая трассировка и validation до фактического deploy.
- Прозрачная эволюция к cross-repo docs governance.

### Негативные/компромиссы
- Дополнительные сущности и поля в data model.
- Дополнительные staff API для composition preview/docs sources.
- Рост количества GitHub API вызовов в resolve path.

## Миграция

1. Расширить metadata repositories (alias/role/default_ref/docs root).
2. Добавить resolver + preview API (audit-only режим).
3. Подключить worker multi-repo checkout/reconcile.
4. Включить docs federation в prompt context.

## План отката/замены

Условия отката:
- высокий процент unresolved/conflict ошибок в production dogfooding.

Стратегия отката:
- feature-flag rollback `multi_repo_mode=enforced -> preview|off`;
- fallback в single-root режим для affected projects.

## Связанные документы

- `docs/product/requirements_machine_driven.md` (FR-020, FR-022)
- `docs/architecture/data_model.md`
- `docs/architecture/api_contract.md`
- `docs/architecture/prompt_templates_policy.md`
- `docs/architecture/multi_repo_mode_design.md`
