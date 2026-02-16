---
doc_id: EPC-CK8S-S3-D14
type: epic
title: "Epic S3 Day 14: Repository onboarding preflight (token scopes, GitHub ops, domain resolution) + bot params per repo"
status: planned
owner_role: EM
created_at: 2026-02-16
updated_at: 2026-02-16
related_issues: [19]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
---

# Epic S3 Day 14: Repository onboarding preflight (token scopes, GitHub ops, domain resolution) + bot params per repo

## TL;DR
- Цель: сделать добавление репозитория в проект предсказуемым и self-validated: проверять токены (platform и bot), проверять реальные GitHub операции и проверять, что домены проекта резолвятся на кластер для ai/staging и ai slots.
- Ключевая ценность: меньше “сломалось потом в рантайме”, больше ранней диагностики и понятных ошибок в UI.
- MVP-результат: onboarding preflight report + UI для bot-параметров на уровне репо + автоматические проверки и cleanup тестовых артефактов.

## Priority
- `P0`.

## Зависимости
- Требует базовой модели конфигов и fallback кредов из `docs/delivery/epics/epic-s3-day13-config-and-credentials-governance.md`.
- Domain preflight должен быть согласован с full-env `domainTemplate` из `docs/delivery/epics/epic-s3-day11-full-env-slots-and-subdomains.md`.

## Scope
### In scope (MVP)
- Проверки токенов:
  - platform token: наличие прав на management операции репозитория;
  - bot token: наличие прав на агентные операции (issue/pr/comments + git push).
- GitHub preflight (через реальный API, с cleanup):
  - создать и удалить webhook;
  - создать/проверить label (и удалить при возможности);
  - создать Issue от имени бота и оставить comment, затем закрыть/пометить как test;
  - создать временную ветку, сделать минимальный коммит, push;
  - открыть PR и оставить comment, затем закрыть PR и удалить ветку.
- Domain preflight при создании проекта и при настройке full-env:
  - проверить, что домены, используемые для `ai-staging` и ai slots, резолвятся на ingress кластера;
  - проверять как минимум `CODEXK8S_PRODUCTION_DOMAIN`, `CODEXK8S_AI_DOMAIN` и шаблоны, которые используются в `services.yaml` (`domainTemplate`).
- UI/DB:
  - для каждого репозитория хранить не только platform token, но и параметры бота (token + username/email);
  - показывать preflight status и список прошедших/проваленных проверок с подсказками по исправлению.
- Безопасность:
  - preflight не должен логировать secret material;
  - тестовые артефакты в GitHub должны быть помечены префиксом (например `codex-k8s-preflight-*`) и удаляться/закрываться автоматически.

### Out of scope
- Полная поддержка GitHub App installation flow.
- Поддержка GitLab и multi-provider preflight.
- Автоматическое управление DNS провайдером (создание wildcard записей).

## Декомпозиция (Stories/Tasks)
- Story-1: Спецификация preflight checks:
  - список операций;
  - критерии pass/fail;
  - политика cleanup и ограничения по rate limits.
- Story-2: Backend preflight runner:
  - выполнение шагов с таймаутами;
  - сбор отчёта (структурированный DTO без секретов);
  - idempotency и запрет параллельных preflight на один репозиторий.
- Story-3: GitHub operations adapters:
  - webhook create/delete;
  - labels ensure;
  - issues ensure + comment;
  - branches/commits/push;
  - PR open/close + comment.
- Story-4: Domain resolution checker:
  - DNS lookup + сравнение с ожидаемым ingress endpoint;
  - ошибки должны быть user-actionable (что настроить в DNS).
- Story-5: UI:
  - форма добавления репозитория с вводом platform/bot creds (или выбором из fallback);
  - кнопка "Run preflight" и вывод отчёта;
  - статусные бейджи на списке репозиториев проекта.
- Story-6: Тесты:
  - unit: планировщик шагов, интерпретация ошибок GitHub API, нормализация результата;
  - integration (опционально, под флагом): выполнение preflight на тестовом репо.

## Критерии приемки
- При добавлении репозитория платформа может выполнить preflight и показать результат в UI.
- Отчёт preflight включает:
  - какие токены использовались (scope: repo/project/platform, без значений);
  - какие проверки прошли и какие нет;
  - список созданных тестовых артефактов и подтверждение cleanup.
- Ошибки формулируются так, чтобы пользователь понимал, какие права/настройки отсутствуют.
- Проверка доменов явно сообщает, какие hostnames не резолвятся на кластер и что нужно поправить в DNS.

## Риски/зависимости
- GitHub rate limits и права на удаление некоторых артефактов (например label) могут отличаться; нужен fallback cleanup (закрыть Issue/PR, оставить метку "test").
- Domain check зависит от того, откуда выполняется DNS lookup (pod/host) и какие resolver настроены; нужен предсказуемый execution context.
