---
doc_id: EPC-CK8S-S3-D17
type: epic
title: "Epic S3 Day 17: Environment-scoped secret overrides and OAuth callback strategy"
status: completed
owner_role: EM
created_at: 2026-02-18
updated_at: 2026-02-19
related_issues: [19]
related_prs: [49]
approvals:
  required: ["Owner"]
  status: approved
  request_id: ""
---

# Epic S3 Day 17: Environment-scoped secret overrides and OAuth callback strategy

## TL;DR
- Цель: сделать переиспользуемую модель environment-scoped секретов для сервисов, чтобы dev/ai/prod могли использовать разные значения без хардкода под конкретный проект.
- Результат: платформа умеет резолвить секреты по политике override/fallback, а OAuth proxy в ai-slot получает корректные callback credentials для отдельного домена.

## Priority
- `P0`.

## Scope
### In scope
- Расширение `services.yaml` для секретов/credentials mapping:
  - возможность указать source key и environment-specific override key(s);
  - поддержка шаблонов имени (например, `*_AI`) как одного из режимов, но без vendor hardcode.
- Runtime/Bootstrap secret resolver:
  - deterministic chain: env override -> environment base -> platform default;
  - единый код резолва в `control-plane` и `codex-bootstrap`.
- OAuth callback strategy:
  - отдельные secrets для production и ai environment;
  - интеграция в `oauth2-proxy` deployment/secret wiring для ai-slot доменов.
- GitHub/Kubernetes sync:
  - environment-aware запись/чтение секретов;
  - обратная синхронизация в config/env при восстановлении значений.

### Out of scope
- Полный dynamic secret provider (Vault/External Secrets) как обязательный runtime.

## Декомпозиция
- Story-1: `services.yaml` secret override contract + schema.
- Story-2: единый resolver package для bootstrap/control-plane.
- Story-3: OAuth proxy env split (production vs ai) + migration path.
- Story-4: tests + docs (операционная инструкция по секретам окружений).

## Критерии приемки
- Для одного и того же логического секрета можно задать разные значения по окружениям без правок кода.
- Ai-slot использует отдельные OAuth callback credentials и не ломает production login flow.
- Secret sync в GitHub/Kubernetes учитывает environment и не перетирает соседние env значения.
- Поведение описано в docs и покрыто тестами на override/fallback цепочку.

## Риски/зависимости
- Риск неправильного fallback и silent misconfiguration: нужен явный audit лог резолва ключа.
- Зависимость от чистого миграционного пути для уже существующих production secrets.

## Фактический результат (выполнено)
- Расширен typed-контракт `services.yaml`:
  - `spec.secretResolution.environmentAliases`;
  - `spec.secretResolution.keyOverrides` (`sourceKey -> overrideKeys{env:key}`);
  - `spec.secretResolution.patterns` (`sourcePrefix/exclude*/environments/overrideTemplate`).
- Добавлен reusable service-scope в `services.yaml`:
  - `spec.services[].scope: environment | infrastructure-singleton`;
  - `oauth2-proxy` переведён в `infrastructure-singleton`, чтобы не деплоиться в каждом AI-слоте.
- AI ingress переведён на shared OAuth (nginx `auth-url`/`auth-signin`) через centralized oauth2-proxy endpoint.
- Добавлена schema-валидация и runtime-валидация для `spec.secretResolution` в `libs/go/servicescfg`.
- Реализован единый `SecretResolver` в `libs/go/servicescfg` и подключен в оба контура:
  - `cmd/codex-bootstrap` (`github-sync`, `sync-secrets`);
  - `control-plane` runtime prerequisites (`runtimedeploy`).
- Реализована детерминированная цепочка резолва:
  - `env override -> environment secret -> shared/platform secret -> base value`.
- OAuth split production/ai доведен до env-aware резолва:
  - `CODEXK8S_GITHUB_OAUTH_CLIENT_ID/SECRET` теперь резолвятся через общий resolver;
  - в task logs пишется источник резолва OAuth ключей (без утечки значений).
- В `services.yaml` codex-k8s добавлены правила для OAuth credentials:
  - explicit `*_AI` override keys;
  - pattern-mode `CODEXK8S_<NAME>_AI`.
- `bootstrap/host/config.env.example` обновлен:
  - задокументированы оба override-режима;
  - добавлены `CODEXK8S_GITHUB_OAUTH_CLIENT_ID_AI`, `CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET_AI`,
    `CODEXK8S_PRODUCTION_GITHUB_OAUTH_CLIENT_ID`, `CODEXK8S_PRODUCTION_GITHUB_OAUTH_CLIENT_SECRET`.

## Проверки
- `go test ./libs/go/servicescfg` — passed.
- `go test ./cmd/codex-bootstrap/...` — passed.
- `go test ./services/internal/control-plane/...` — passed.
