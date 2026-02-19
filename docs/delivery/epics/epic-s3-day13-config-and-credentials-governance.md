---
doc_id: EPC-CK8S-S3-D13
type: epic
title: "Epic S3 Day 13: Unified config/secrets governance (platform/project/repo) + GitHub credentials fallback"
status: completed
owner_role: EM
created_at: 2026-02-16
updated_at: 2026-02-16
related_issues: [19]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-19-full-docset"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-19
---

# Epic S3 Day 13: Unified config/secrets governance (platform/project/repo) + GitHub credentials fallback

## TL;DR
- Цель: убрать путаницу env/secrets и вынести конфигурацию платформы, проектов и репозиториев в централизованный admin UI, сохраняя секреты в БД в зашифрованном виде и синхронизируя их в GitHub и Kubernetes.
- Ключевая ценность: управляемые и воспроизводимые настройки без ручного редактирования `bootstrap/host/config.env`, с безопасной политикой обновлений и предсказуемым fallback по кредам.
- MVP-результат: модель конфигов по scope (platform/project/repo) + UI/API для редактирования + sync/reconcile в GitHub/K8s + предупреждения о рисках замены секретов.

## Priority
- `P0`.

## Контекст
- Сейчас полный список переменных задан в `bootstrap/host/config.env.example`, а на практике значения часто попадают в `bootstrap/host/config.env`.
- Конфиги смешаны:
  - platform-level (нужны до старта и/или влияют на всю систему);
  - project-level (относятся к конкретному проекту);
  - repository-level (относятся к конкретному репозиторию проекта).
- Есть конфиги, которые теоретически можно менять в рантайме (например, лимиты worker), и есть “опасные” секреты, которые нельзя просто перезаписать без внешних действий (например, пароль БД, если сама БД не обновлена).
- GitHub креды должны быть в нескольких уровнях:
  - платформенные (для `codex-k8s` repo и служебных repos, включая доксет и внешние модули);
  - проектные (как дефолт для репо проекта);
  - репозиторные (override конкретного репо).
- Требуется fallback-цепочка кредов: repo -> project -> platform.

## Scope
### In scope (MVP)
- Классификация конфигов и целевая модель:
  - scope: `platform | project | repository`;
  - kind: `secret | variable`;
  - mutability: `startup_required | runtime_mutable`;
  - sync targets: `github_secret`, `github_variable`, `kubernetes_secret`, `kubernetes_configmap` (минимум: GitHub secrets + K8s secrets).
- GitHub credentials model:
  - два типа кредов: `platform token` (management path) и `bot token` (agent-run path);
  - параметры бота: `username`, `email` (и где применяются);
  - fallback: repo creds -> project creds -> platform creds.
- Хранение:
  - секреты хранятся в PostgreSQL в зашифрованном виде;
  - plaintext секретов не должен появляться в логах/DTO/flow events.
- UI/API:
  - admin UI для платформенных настроек;
  - UI на уровне проекта для project-level настроек;
  - UI на уровне репозитория для repo overrides (включая bot-параметры).
- Sync/reconcile контур:
  - при создании/изменении конфигов выполнять идемпотентную синхронизацию в GitHub и Kubernetes;
  - не переписывать “опасные” секреты без явного подтверждения пользователя и предупреждения о последствиях;
  - поддержать режим “create-if-missing” для чувствительных секретов (дефолт).
- Политика предупреждений:
  - для ключей класса “опасные” UI обязан показывать risk-warning;
  - операции update для опасных ключей требуют явного подтверждения (и в идеале Owner approval, если действие затрагивает production/prod).
- Трассируемость:
  - audit trail: кто и когда изменил конфиг, какой scope, куда синхронизировано (без раскрытия значения).

### Out of scope
- Интеграция с Vault/KMS и full secret rotation workflow.
- Поддержка GitLab credentials и multi-provider policy (оставить задел интерфейсами).
- Автоматическое обновление пароля внутри самой БД при смене секретов (только предупреждения и runbook).

## Декомпозиция (Stories/Tasks)
- Story-1: Inventory конфигов:
  - выписать ключи из `bootstrap/host/config.env.example`;
  - разнести по scope и sync targets;
  - определить список “опасных” ключей (минимум: DB creds, OAuth secrets, webhook secret).
- Story-2: Data model и API для конфигов:
  - сущности для platform/project/repository config entries;
  - encrypted storage, versioning и timestamps;
  - endpoints для list/get/upsert, без утечек secret values.
- Story-3: Effective config resolver:
  - алгоритм fallback repo -> project -> platform;
  - отдельный resolver для GitHub credentials (platform-token и bot-token).
- Story-4: Sync engine (GitHub + Kubernetes):
  - idempotent apply;
  - policy по overwrite (default: create-if-missing, update только при явном флаге);
  - dry-run/preview diff для UI.
- Story-5: UI:
  - формы редактирования с пометкой scope;
  - warning UX для опасных ключей;
  - отображение статуса синхронизации и drift (если обнаружен).
- Story-6: Governance:
  - минимальные правила “кто может менять что” (platform admin vs project admin);
  - фиксация изменений в audit контуре.

## Критерии приемки
- Все конфиги, относящиеся к проектам и репозиториям, вводятся через UI и хранятся в БД зашифрованно (без требования редактировать `bootstrap/host/config.env`).
- Для GitHub кредов работает fallback:
  - если repo creds не заданы, используются project creds;
  - если project creds не заданы, используются platform creds.
- Синхронизация в GitHub/Kubernetes идемпотентна и не перезаписывает “опасные” секреты по умолчанию.
- UI предупреждает о рисках при update опасных секретов и требует явного подтверждения действия.

## Риски/зависимости
- Нужна аккуратная политика “что можно перезаписывать” и явный список исключений, иначе высок риск поломать production доступы.
- Требуется строгий запрет на утечки секретов в логи, ответы API и UI.
- Epic зависит от корректного разделения “platform repo” и “project repos” в текущем bootstrap/runtime-deploy контуре.

## Фактический результат (выполнено)
- Реализована единая модель `ConfigEntry` со scope `platform|project|repository`, kind `variable|secret`, mutability `startup_required|runtime_mutable`.
- Реализованы staff API + UI для:
  - list/upsert/delete конфигов;
  - warning/confirmation для “опасных” ключей (запрещено обновление без явного подтверждения).
- Реализована синхронизация (idempotent apply) по sync targets:
  - GitHub environment secrets: `github_env_secret:<env>`;
  - GitHub environment variables: `github_env_var:<env>`;
  - Kubernetes secrets: `k8s_secret:<namespace>/<name>`.
- Реализован effective resolver GitHub credentials с fallback:
  - repository -> project -> platform (для platform token и bot token).
- Усилен governance слой:
  - RBAC для project/repository scope (нужно право записи в проект, а не только platform admin);
  - исправлена уникальность config entries в PostgreSQL (partial unique индексы, чтобы исключить дубли при `NULL`).

## Data model impact
- Исправлены уникальные ограничения для config entries (частичные unique индексы по scope).
- Добавлен `GetByID` для config entries (для корректного RBAC/delete и проверки dangerous-confirmation).

## Проверки
- `go test ./...` — passed.
