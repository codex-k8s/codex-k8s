---
doc_id: PLN-CK8S-0001
type: delivery-plan
title: "codex-k8s — Delivery Plan"
status: draft
owner_role: EM
created_at: 2026-02-06
updated_at: 2026-02-06
related_issues: [1]
related_prs: []
related_docsets: ["docs/_docset/issues/issue-0001-codex-k8s-bootstrap.md"]
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-06-mvp"
---

# Delivery Plan: codex-k8s

## TL;DR
- Что поставляем: MVP control-plane + staff UI + webhook orchestration + staging bootstrap/deploy loop.
- Когда: поэтапно, с ранним staging для ручных тестов.
- Главные риски: bootstrap automation, security hardening, runner stability.
- Что нужно от Owner: подтверждение deploy-модели и доступов (GitHub PAT/OpenAI key).

## Входные артефакты
- Brief: `docs/product/brief.md`
- Constraints: `docs/product/constraints.md`
- Architecture (C4): `docs/architecture/c4_context.md`, `docs/architecture/c4_container.md`
- ADR: `docs/architecture/adr/ADR-0001-kubernetes-only.md`, `docs/architecture/adr/ADR-0002-webhook-driven-and-deploy-workflows.md`, `docs/architecture/adr/ADR-0003-postgres-jsonb-pgvector.md`, `docs/architecture/adr/ADR-0004-repository-provider-interface.md`
- Data model: `docs/architecture/data_model.md`

## Структура работ (WBS)
### Epic 1: Foundation
- Репо-каркас, гайды, базовые ADR.
- Базовый `services.yaml` и deploy skeleton.

### Epic 2: Core backend MVP
- `api-gateway`, `control-plane`, `worker`.
- База данных + миграции + audit/event model.
- RepositoryProvider (GitHub adapter).

### Epic 3: Staff UI MVP
- OAuth login flow.
- short-lived JWT с ротацией в API gateway.
- Экраны проектов/репозиториев/агентов/запусков/документов.
- Управление learning mode toggle (пер-пользователь, пер-проект).

### Epic 3.1: Learning mode MVP
- Prompt augmentation для user-initiated задач.
- Сохранение объяснений в БД.
- Post-PR образовательные комментарии (summary + line-level при необходимости).

### Epic 4: Staging bootstrap and first deploy loop
- Скрипт запуска с хоста разработчика по SSH root.
- Автоматическое создание пользователя на сервере и hardening базового доступа.
- Автоустановка k3s + сеть + зависимости + postgres + codex-k8s (`local-path` storage profile на MVP).
- Настройка self-hosted runner:
  - локально: 1 persistent runner (long polling);
  - на staging/ai-staging/prod при наличии домена: autoscaled runner set.
- Workflow deploy `main -> staging`.

### Epic 5: Manual staging tests
- Smoke/functional сценарии вручную на staging.
- Фиксация дефектов и стабилизация.

## Зависимости
- Внутренние: Core backend до полноценного UI управления.
- Внешние: GitHub PAT с нужными правами, рабочий staging сервер Ubuntu 24.04.

## План сред/окружений
- Dev slots: локальный/кластерный dev для компонентов.
- Staging: обязателен до расширения функционала.
- Prod: после стабилизации staging и security review.

## Специальный этап bootstrap staging (обязательный)

Цель этапа: когда уже есть что тестировать вручную, запускать один скрипт с машины разработчика и автоматически поднимать staging окружение.

Ожидаемое поведение скрипта:
- запускается на машине разработчика (текущей) и подключается по SSH к серверу как `root`;
- создаёт отдельного пользователя (sudo + ssh key auth), отключает дальнейший root-password вход;
- ставит k3s и сетевой baseline (ingress, cert-manager, network policy baseline);
- ставит зависимости платформы;
- разворачивает PostgreSQL и `codex-k8s`;
- спрашивает внешние креды (`GitHub PAT`, `OPENAI_API_KEY`), внутренние секреты генерирует сам;
- передаёт default `learning_mode` из `bootstrap/host/config.env` (по умолчанию включён, пустое значение = выключен);
- настраивает GitHub Actions runner в Kubernetes для staging (ARC или эквивалент);
- подготавливает deploy workflow и необходимые repo secrets/variables.

## Чек-листы готовности
### Definition of Ready (DoR)
- [ ] Brief/Constraints/Architecture/ADR согласованы.
- [ ] Server access для staging подтверждён.
- [ ] GitHub PAT и OpenAI ключ доступны.

### Definition of Done (DoD)
- [ ] Staging bootstrap выполняется одной командой.
- [ ] Runner в k8s онлайн и принимает staging deploy job.
- [ ] Push в `main` обновляет staging.
- [ ] Manual smoke tests проходят.
- [ ] Learning mode проверен на тестовом PR: есть объяснение why/tradeoffs в результате и post-PR комментарий.

## Риски и буферы
- Риск: нестабильная сеть/доступы при bootstrap.
- Буфер: fallback runbook ручной установки.

## План релиза (верхний уровень)
- Релизные окна:
  - staging continuous (auto deploy on push to `main`);
  - production gated (manual dispatch + environment approval).
- Rollback: возвращение на предыдущий контейнерный тег + DB migration rollback policy.

## Решения Owner
- Runner scale policy утверждена:
  - локальные запуски — один persistent runner;
  - серверные окружения с доменом — autoscaled set.
- Storage policy утверждена: на MVP используем `local-path`, Longhorn переносим на следующий этап.
- Read replica policy утверждена: минимум одна async streaming replica на MVP, далее эволюция до 2+ и sync/quorum без изменений приложения.

## Апрув
- request_id: owner-2026-02-06-mvp
- Решение: approved
- Комментарий: План поставки и условия bootstrap/staging утверждены.
