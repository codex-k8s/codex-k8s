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
- Что нужно от Owner: подтверждение deploy-модели и доступов (GitHub fine-grained token/OpenAI key).

## Входные артефакты
- Requirements baseline: `docs/product/requirements_machine_driven.md`
- Brief: `docs/product/brief.md`
- Constraints: `docs/product/constraints.md`
- Architecture (C4): `docs/architecture/c4_context.md`, `docs/architecture/c4_container.md`
- ADR: `docs/architecture/adr/ADR-0001-kubernetes-only.md`, `docs/architecture/adr/ADR-0002-webhook-driven-and-deploy-workflows.md`, `docs/architecture/adr/ADR-0003-postgres-jsonb-pgvector.md`, `docs/architecture/adr/ADR-0004-repository-provider-interface.md`
- Data model: `docs/architecture/data_model.md`
- Sprint plan: `docs/delivery/sprint_s1_mvp_vertical_slice.md`
- Epic catalog: `docs/delivery/epic_s1.md`
- Process requirements: `docs/delivery/development_process_requirements.md`

## Структура работ (WBS)
### Sprint S1: Day 0 + Day 1..7 (8 эпиков)
- Day 0 (completed): `docs/delivery/epics/epic-s1-day0-bootstrap-baseline.md`
- Day 1: `docs/delivery/epics/epic-s1-day1-webhook-idempotency.md`
- Day 2: `docs/delivery/epics/epic-s1-day2-worker-slots-k8s.md`
- Day 3: `docs/delivery/epics/epic-s1-day3-auth-rbac-ui.md`
- Day 4: `docs/delivery/epics/epic-s1-day4-repository-provider.md`
- Day 5: `docs/delivery/epics/epic-s1-day5-learning-mode.md`
- Day 6: `docs/delivery/epics/epic-s1-day6-hardening-observability.md`
- Day 7: `docs/delivery/epics/epic-s1-day7-stabilization-gate.md`

### Daily delivery contract (обязательный)
- Каждый день задачи дня влиты в `main`.
- Каждый день изменения автоматически задеплоены на staging.
- Каждый день выполнен ручной smoke-check.
- Каждый день актуализированы документы при изменениях API/data model/webhook/RBAC.
- Для каждого эпика заполнен `Data model impact` по структуре `docs/templates/data_model.md`.
- Правила спринт-процесса и ownership артефактов выполняются по `docs/delivery/development_process_requirements.md`.

## Зависимости
- Внутренние: Core backend до полноценного UI управления.
- Внешние: GitHub fine-grained token с нужными правами, рабочий staging сервер Ubuntu 24.04.

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
- поднимает внутренний registry (`ClusterIP`, без auth на уровне registry) и Kaniko pipeline для сборки образа в кластере;
- разворачивает PostgreSQL и `codex-k8s`;
- спрашивает внешние креды (`GitHub fine-grained token`, `CODEXK8S_OPENAI_API_KEY`), внутренние секреты генерирует сам;
- передаёт default `learning_mode` из `bootstrap/host/config.env` (по умолчанию включён, пустое значение = выключен);
- настраивает GitHub Actions runner в Kubernetes для staging (ARC или эквивалент);
- подготавливает deploy workflow и необходимые repo secrets/variables.

## Чек-листы готовности
### Definition of Ready (DoR)
- [ ] Brief/Constraints/Architecture/ADR согласованы.
- [ ] Server access для staging подтверждён.
- [ ] GitHub fine-grained token и OpenAI ключ доступны.

### Definition of Done (DoD)
- [x] Day 0 baseline bootstrap выполнен.
- [ ] Для Day 1..7: каждый эпик закрыт по своим acceptance criteria.
- [ ] Для Day 1..7: ежедневный merge -> auto deploy -> smoke check выполнен.
- [ ] Webhook -> run -> worker -> k8s -> UI цепочка проходит regression.
- [ ] Learning mode проверен на staging в on/off режимах.

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
