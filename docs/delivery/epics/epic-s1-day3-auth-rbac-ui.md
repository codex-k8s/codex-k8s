---
doc_id: EPC-CK8S-S1-D3
type: epic
title: "Epic Day 3: GitHub OAuth, JWT, project RBAC, minimal staff UI"
status: completed
owner_role: EM
created_at: 2026-02-06
updated_at: 2026-02-11
related_issues: [1]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-06-day3"
---

# Epic Day 3: GitHub OAuth, JWT, project RBAC, minimal staff UI

## TL;DR
- Цель эпика: включить безопасный вход и базовый контроль доступа в staff контуре.
- Ключевая ценность: управляемый доступ к проектам без self-signup.
- MVP-результат (production/dev): вход через GitHub OAuth (через `oauth2-proxy`), RBAC на уровне БД, minimal UI для проектов и запусков.

## Priority
- `P0` (обязательная безопасность и управление доступом).

## Ожидаемые артефакты дня
- GitHub OAuth для staff-контура через `oauth2-proxy` (production/dev) + allowlist по email в БД.
- RBAC middleware в backend сервисах (roles `read/read_write/admin`).
- Минимальные staff UI экраны в `services/staff/web-console/**`.
- Acceptance evidence по ролям `read/read_write/admin` на production.

## Контекст
- Почему эпик нужен: без auth/RBAC невозможно безопасно использовать платформу.
- Связь с требованиями: FR-007, FR-017, FR-018, FR-019.

## Scope
### In scope
- GitHub OAuth flow через `oauth2-proxy` на входе в `Ingress` (production/dev).
  - Требуется GitHub OAuth App с callback URL `https://<domain>/oauth2/callback`.
  - Identity прокидывается в `api-gateway` через `X-Auth-Request-*` / `X-Forwarded-*` headers.
- Email matching (allowlist) с разрешёнными пользователями в БД (регистрация запрещена).
- Project RBAC (`read`, `read_write`, `admin`).
- Минимальные UI страницы: проекты, запуски, события.

### Out of scope
- Полный UI функционал для всех настроек платформы.
- SSO кроме GitHub OAuth.
 - Production-режим раздачи собранного UI бандла (на production будет `nginx` + bundle).

## Декомпозиция (Stories/Tasks)
- Story-1: oauth2-proxy (GitHub OAuth) на ingress + identity headers до `api-gateway`.
- Story-2: allowlist по email + staff principal resolution (DB-backed).
- Story-3: RBAC middleware для staff/private API.
- Story-4: минимальные Vue3 views для проектов и run-листинга.

## Data model impact (по шаблону data_model.md)
- Сущности:
  - `users`: актуализация `email`, `github_login`.
  - `project_members`: role matrix и learning mode override.
- Связи/FK:
  - `project_members.user_id -> users.id`.
  - `project_members.project_id -> projects.id`.
- Индексы и запросы:
  - Проверить/добавить уникальность `(project_id, user_id)`.
  - Проверить индекс `users(email)` (unique).
- Миграции:
  - Добавить поля/ограничения только при отсутствии в текущей схеме.
- Retention/PII:
  - Email как минимально допустимый PII, без лишних персональных атрибутов.

## Критерии приемки эпика
- Неразрешённый email не получает доступ.
- Пользователь видит только проекты по своей роли.
- Изменения задеплоены на production в день реализации и проверены вручную.

## Риски/зависимости
- Зависимости: корректные OAuth credentials и callback URL.
- Риск: рассинхрон UI и RBAC-прав при кэше.

## План релиза (верхний уровень)
- После merge провести smoke: login, RBAC read/write/admin на production.

## Апрув
- request_id: owner-2026-02-06-day3
- Решение: approved
- Комментарий: Day 3 scope принят.
