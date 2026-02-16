---
doc_id: REG-CK8S-S1-0001
type: regression
title: "Sprint S1 Regression Gate (production)"
status: active
owner_role: QA
created_at: 2026-02-09
updated_at: 2026-02-14
related_issues: [1]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-12-project-docs-approval"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-12
---

# Sprint S1 Regression Gate (production)

Цель: единый список критических сценариев для go/no-go на следующий спринт.

## Preconditions
- Production домен резолвится на production IP.
- TLS выдан (cert-manager ClusterIssuer `codex-k8s-letsencrypt`).
- Последний deploy зелёный, migrate job completed.

## P0 scenarios (must pass)

1. Webhook ingress (public)
   - `POST https://<domain>/api/v1/webhooks/github` доступен без OAuth (нет 302 на GitHub login).
   - Invalid signature => `401`.
   - First valid request => `202 accepted`.
   - Replay delivery id => `200 duplicate` (idempotency).
   - Проверка выполняется вручную по `docs/ops/production_runbook.md` (invalid signature + idempotency).

2. Worker run loop
   - pending -> running -> succeeded/failed статусы фиксируются в БД.
   - slot lease корректно освобождается.
   - Проверка выполняется вручную (запуск через `issues:labeled` + `run:dev`, финальный статус + отсутствие slot leak).

3. Staff access control
   - OAuth login успешен через `oauth2-proxy`.
   - Неразрешённый email получает `{"code":"forbidden","message":"email is not allowed"}`.
   - Проверка выполняется вручную через `/api/v1/auth/me` и allowlist.

4. Staff UI базовые страницы
   - (manual) `/` Projects загружается.
   - (manual) `/runs` Runs загружается.
   - (manual) `/users` Users загружается для platform admin.

## P1 scenarios (should pass for Sprint S2 readiness)

1. Projects lifecycle
   - platform admin создаёт проект через UI (slug+name).

2. Project repositories lifecycle (Day4)
   - к одному проекту подключаются 2+ GitHub репозитория (репо-токены сохраняются в БД в шифрованном виде).
   - при attach создаётся/обновляется webhook на каждый репозиторий.

3. Learning mode (Day5)
   - effective learning mode резолвится как:
     - project default (projects.settings.learning_mode_default)
     - + member override (project_members.learning_mode_override) при наличии.
   - при `learning_mode=true` создаётся минимум 1 запись в `learning_feedback` после завершения run.
   - staff UI `/runs/:id` показывает learning feedback.

## Evidence capture
- Сохранить результаты ручного smoke/regression в заметки релиза (или issue comment).
- Зафиксировать run_id и screenshots для:
  - Projects -> Repos attach,
  - Runs -> Details (events + feedback).

## Go/No-Go criteria
- Go: все P0 зелёные, нет P0/P1 багов-блокеров.
- No-Go: любой P0 красный, либо есть регрессии security boundary (webhook не проходит без OAuth / наружу открыт k8s api).
