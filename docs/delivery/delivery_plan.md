---
doc_id: PLN-CK8S-0001
type: delivery-plan
title: "codex-k8s — Delivery Plan"
status: active
owner_role: EM
created_at: 2026-02-06
updated_at: 2026-02-15
related_issues: [1, 19]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-12-project-docs-approval"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-11
---

# Delivery Plan: codex-k8s

## TL;DR
- Что поставляем: MVP control-plane + staff UI + webhook orchestration + MCP governance + self-improve loop + production bootstrap/deploy loop.
- Когда: поэтапно, с ранним production для ручных тестов.
- Главные риски: bootstrap automation, security/governance hardening, runner stability.
- Что нужно от Owner: подтверждение deploy-модели и доступов (GitHub fine-grained token/OpenAI key).

## Входные артефакты
- Requirements baseline: `docs/product/requirements_machine_driven.md`
- Brief: `docs/product/brief.md`
- Constraints: `docs/product/constraints.md`
- Agents operating model: `docs/product/agents_operating_model.md`
- Labels policy: `docs/product/labels_and_trigger_policy.md`
- Stage process model: `docs/product/stage_process_model.md`
- Architecture (C4): `docs/architecture/c4_context.md`, `docs/architecture/c4_container.md`
- ADR: `docs/architecture/adr/ADR-0001-kubernetes-only.md`, `docs/architecture/adr/ADR-0002-webhook-driven-and-deploy-workflows.md`, `docs/architecture/adr/ADR-0003-postgres-jsonb-pgvector.md`, `docs/architecture/adr/ADR-0004-repository-provider-interface.md`
- Data model: `docs/architecture/data_model.md`
- Runtime/RBAC model: `docs/architecture/agent_runtime_rbac.md`
- MCP approval/audit flow: `docs/architecture/mcp_approval_and_audit_flow.md`
- Prompt templates policy: `docs/architecture/prompt_templates_policy.md`
- Sprint plan: `docs/delivery/sprint_s1_mvp_vertical_slice.md`
- Epic catalog: `docs/delivery/epic_s1.md`
- Sprint S2 plan: `docs/delivery/sprint_s2_dogfooding.md`
- Epic S2 catalog: `docs/delivery/epic_s2.md`
- Sprint S3 plan: `docs/delivery/sprint_s3_mvp_completion.md`
- Epic S3 catalog: `docs/delivery/epic_s3.md`
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

### Sprint S2: Dogfooding baseline + hardening (Day 0..7)
- Day 0..4 (completed): архитектурное выравнивание, label triggers, namespace/RBAC, MCP prompt context, agent PR flow.
- Day 4.5 (completed): pgx/db-model refactor.
- Day 5 (completed): staff UI observability baseline.
- Day 6 (completed): approval matrix + MCP control tools + audit hardening.
- Day 7 (completed): MVP readiness regression gate + Sprint S3 kickoff package (`docs/delivery/regression_s2_gate.md`).

### Sprint S3: MVP completion (Day 1..12)
- Day 1: full stage/label activation.
- Day 2: staff runtime debug console.
- Day 3: deterministic secret sync (GitHub + Kubernetes).
- Day 4: database lifecycle MCP tools.
- Day 5: owner feedback handle + HTTP approver/executor + Telegram adapter.
- Day 6..7: `run:self-improve` ingestion + updater + PR flow.
- Day 8: agent toolchain auto-extension safeguards.
- Day 9: declarative full-env deploy, `services.yaml` orchestration, runtime parity/hot-reload.
- Day 10: полный redesign staff-консоли на Vuetify.
- Day 11: full-env slots + agent-run + subdomain templating (TLS) для manual QA.
- Day 12: MVP regression/security gate + closeout/handover и переход к post-MVP roadmap.

### Daily delivery contract (обязательный)
- Каждый день задачи дня влиты в `main`.
- Каждый день изменения автоматически задеплоены на production.
- Каждый день выполнен ручной smoke-check.
- Каждый день актуализированы документы при изменениях API/data model/webhook/RBAC.
- Для каждого эпика заполнен `Data model impact` по структуре `docs/templates/data_model.md`.
- Правила спринт-процесса и ownership артефактов выполняются по `docs/delivery/development_process_requirements.md`.

## Зависимости
- Внутренние: Core backend до полноценного UI управления.
- Внешние: GitHub fine-grained token с нужными правами, рабочий production сервер Ubuntu 24.04.

## План сред/окружений
- Dev slots: локальный/кластерный dev для компонентов.
- Production: обязателен до расширения функционала.
- Prod: после стабилизации production и security review.

## Специальный этап bootstrap production (обязательный)

Цель этапа: когда уже есть что тестировать вручную, запускать один скрипт с машины разработчика и автоматически поднимать production окружение.

Ожидаемое поведение скрипта:
- запускается на машине разработчика (текущей) и подключается по SSH к серверу как `root`;
- создаёт отдельного пользователя (sudo + ssh key auth), отключает дальнейший root-password вход;
- ставит k3s и сетевой baseline (ingress, cert-manager, network policy baseline);
- ставит зависимости платформы;
- поднимает внутренний registry (`ClusterIP`, без auth на уровне registry) и Kaniko pipeline для сборки образа в кластере;
- разворачивает PostgreSQL и `codex-k8s`;
- спрашивает внешние креды (`GitHub fine-grained token`, `CODEXK8S_OPENAI_API_KEY`), внутренние секреты генерирует сам;
- передаёт default `learning_mode` из `bootstrap/host/config.env` (по умолчанию включён, пустое значение = выключен);
- настраивает GitHub webhooks/labels/secrets/variables через API без GitHub Actions runner;
- запускает self-deploy через control-plane runtime deploy job (build/mirror/apply/cleanup).

## Чек-листы готовности
### Definition of Ready (DoR)
- [ ] Brief/Constraints/Architecture/ADR согласованы.
- [ ] Server access для production подтверждён.
- [ ] GitHub fine-grained token и OpenAI ключ доступны.

### Definition of Done (DoD)
- [x] Day 0 baseline bootstrap выполнен.
- [ ] Для активного спринта: каждый эпик закрыт по своим acceptance criteria.
- [ ] Для активного спринта: ежедневный merge -> auto deploy -> smoke check выполнен.
- [ ] Webhook -> run -> worker -> k8s -> UI цепочка проходит regression.
- [ ] Learning mode и self-improve mode проверены на production.
- [ ] MCP governance tools (secret/db/feedback) прошли approve/deny regression.

## Риски и буферы
- Риск: нестабильная сеть/доступы при bootstrap.
- Буфер: fallback runbook ручной установки.

## План релиза (верхний уровень)
- Релизные окна:
  - production continuous (auto deploy on push to `main`);
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
- Комментарий: План поставки и условия bootstrap/production утверждены.
