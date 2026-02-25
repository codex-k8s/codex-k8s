---
doc_id: PLN-CK8S-0001
type: delivery-plan
title: "codex-k8s — Delivery Plan"
status: active
owner_role: EM
created_at: 2026-02-06
updated_at: 2026-02-25
related_issues: [1, 19, 74, 100, 106, 112, 154, 155, 170, 171, 184, 185, 187]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-19-full-docset"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-19
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
- Sprint plan: `docs/delivery/sprints/s1/sprint_s1_mvp_vertical_slice.md`
- Epic catalog: `docs/delivery/epics/s1/epic_s1.md`
- Sprint S2 plan: `docs/delivery/sprints/s2/sprint_s2_dogfooding.md`
- Epic S2 catalog: `docs/delivery/epics/s2/epic_s2.md`
- Sprint S3 plan: `docs/delivery/sprints/s3/sprint_s3_mvp_completion.md`
- Epic S3 catalog: `docs/delivery/epics/s3/epic_s3.md`
- Sprint S4 plan: `docs/delivery/sprints/s4/sprint_s4_multi_repo_federation.md`
- Epic S4 catalog: `docs/delivery/epics/s4/epic_s4.md`
- Sprint S5 plan: `docs/delivery/sprints/s5/sprint_s5_stage_entry_and_label_ux.md`
- Epic S5 catalog: `docs/delivery/epics/s5/epic_s5.md`
- Sprint S6 plan: `docs/delivery/sprints/s6/sprint_s6_agents_prompt_management.md`
- Epic S6 catalog: `docs/delivery/epics/s6/epic_s6.md`
- Sprint index: `docs/delivery/sprints/README.md`
- Epic index: `docs/delivery/epics/README.md`
- E2E master plan: `docs/delivery/e2e_mvp_master_plan.md`
- Process requirements: `docs/delivery/development_process_requirements.md`

## Структура работ (WBS)
### Sprint S1: Day 0 + Day 1..7 (8 эпиков)
- Day 0 (completed): `docs/delivery/epics/s1/epic-s1-day0-bootstrap-baseline.md`
- Day 1: `docs/delivery/epics/s1/epic-s1-day1-webhook-idempotency.md`
- Day 2: `docs/delivery/epics/s1/epic-s1-day2-worker-slots-k8s.md`
- Day 3: `docs/delivery/epics/s1/epic-s1-day3-auth-rbac-ui.md`
- Day 4: `docs/delivery/epics/s1/epic-s1-day4-repository-provider.md`
- Day 5: `docs/delivery/epics/s1/epic-s1-day5-learning-mode.md`
- Day 6: `docs/delivery/epics/s1/epic-s1-day6-hardening-observability.md`
- Day 7: `docs/delivery/epics/s1/epic-s1-day7-stabilization-gate.md`

### Sprint S2: Dogfooding baseline + hardening (Day 0..7)
- Day 0..4 (completed): архитектурное выравнивание, label triggers, namespace/RBAC, MCP prompt context, agent PR flow.
- Day 4.5 (completed): pgx/db-model refactor.
- Day 5 (completed): staff UI observability baseline.
- Day 6 (completed): approval matrix + MCP control tools + audit hardening.
- Day 7 (completed): MVP readiness regression gate + Sprint S3 kickoff package (`docs/delivery/regression_s2_gate.md`).

### Sprint S3: MVP completion (Day 1..21)
- Day 1: full stage/label activation.
- Day 2: staff runtime debug console.
- Day 3: deterministic secret sync (GitHub + Kubernetes).
- Day 4: database lifecycle MCP tools.
- Day 5: owner feedback handle + HTTP approver/executor + Telegram adapter.
- Day 6..7: `run:self-improve` ingestion + updater + PR flow.
- Day 8: agent toolchain auto-extension safeguards.
- Day 9: declarative full-env deploy, `services.yaml` orchestration, runtime parity/hot-reload.
- Day 10 (completed): полный redesign staff-консоли на Vuetify.
- Day 11 (completed): full-env slots + agent-run + subdomain templating (TLS) для manual QA.
- Day 12 (completed): docset import + safe sync (`agent-knowledge-base` -> projects).
- Day 13 (completed): unified config/secrets governance (platform/project/repo) + GitHub creds fallback.
- Day 14 (completed): repository onboarding preflight (token scopes + GitHub ops + domain resolution) + bot params per repo.
- Day 16 (completed): gRPC transport boundary hardening (transport -> service -> repository) по Issue #45.
- Day 15: prompt context overhaul (`services.yaml` docs tree + role prompt matrix + GitHub service messages templates).
- Day 17: environment-scoped secret overrides + OAuth callback strategy (без project-specific hardcode).
- Day 18: runtime error journal + staff alert center (stacked alerts, mark-as-viewed).
- Day 19: frontend manual QA hardening loop (Owner-driven bug cycle до full e2e).
- Day 19.5: realtime шина на PostgreSQL (`event log + LISTEN/NOTIFY`) + multi-server WebSocket backplane.
- Day 19.6: интеграция realtime подписок в staff UI (runs/deploy/errors/logs/events), удаление кнопок `Обновить` в realtime-экранах, fallback polling.
- Day 19.7: retention full-env namespace по role-based TTL + lease extension/reuse на `run:*:revise` (Issue #74).
- Day 20: full e2e regression/security gate + MVP closeout/handover и переход к post-MVP roadmap (подробности в `docs/delivery/e2e_mvp_master_plan.md`).

### Sprint S4: Multi-repo runtime and docs federation (Issue #100)
- Day 1 (completed): execution foundation для federated multi-repo composition и docs federation (`docs/delivery/epics/s4/epic-s4-day1-multi-repo-composition-and-docs-federation.md`).
- Результат Day 1: формальный execution-plan (stories + quality-gates + owner decisions) для перехода в `run:dev`.
- Следующие day-эпики S4 формируются после Owner review Day 1 и закрытия зависимостей по S3 Day20.

### Sprint S5: Stage entry and label UX orchestration (Issues #154/#155/#170/#171)
- Day 1 (in-review): launch profiles + deterministic next-step actions (`docs/delivery/epics/s5/epic-s5-day1-launch-profiles-and-stage-launcher-ux.md`).
- Результат Day 1 (факт): owner-ready vision/prd + architecture execution package для входа в `run:dev` подготовлен в Issue #155 (включая ADR-0008); Owner approval получен (PR #166, 2026-02-25).
- Day 2 (in-review): single-epic execution package для реализации FR-053/FR-054 (`docs/delivery/epics/s5/epic-s5-day2-launch-profiles-dev-execution.md`).
- Результат Day 2 (факт): в Issue #170 зафиксирован delivery governance пакет (QG-D2-01..QG-D2-05, DoD, handover), создана implementation issue #171 для выполнения одним эпиком.

### Sprint S6: Agents configuration and prompt templates lifecycle (Issue #184)
- Day 1 (in-review): intake baseline по разделу `Agents` (`docs/delivery/epics/s6/epic-s6-day1-agents-prompts-intake.md`).
- Результат Day 1 (факт): подтвержден разрыв между scaffold UI и отсутствием staff API контрактов для agents/templates/audit; зафиксирована полная stage-траектория до `run:doc-audit` и требование создавать follow-up issue на каждом этапе.
- Day 2 (in-review): vision stage (`docs/delivery/epics/s6/epic-s6-day2-agents-prompts-vision.md`, issue #185).
- Результат Day 2 (факт): сформирован vision-пакет (charter + success metrics + risk frame + MVP/Post-MVP boundaries), создана follow-up issue `run:prd` #187 с инструкцией создать issue `run:arch` после PRD.
- Day 3 (planned): PRD stage в отдельной issue #187.
- Следующие day-эпики S6 формируются строго последовательно по stage-цепочке:
  `prd -> arch -> design -> plan -> dev -> doc-audit` с отдельной issue на каждый этап.

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
- [ ] Для `full-env` подтверждены role-based TTL retention namespace и lease extension на `run:*:revise` (Issue #74).
- [x] Для Issue #100 зафиксирован delivery execution-plan Sprint S4 (federated composition + multi-repo docs federation) и подготовлен handover в `run:dev`.
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
