# Bootstrap (production)

Набор скриптов для первичного развёртывания `codex-k8s` на удалённом сервере Ubuntu 24.04.

## Что делает bootstrap

- запускается с хоста разработчика;
- подключается к удалённому серверу по SSH под `root`;
- упаковывает текущий локальный snapshot репозитория и передаёт его на сервер в `/opt/codex-k8s`;
- создаёт отдельного операционного пользователя;
- ставит k3s; базовые сетевые компоненты (ingress-nginx/cert-manager) применяются через Go runtime deploy prerequisites;
- проверяет DNS до старта раскатки: `CODEXK8S_PRODUCTION_DOMAIN` должен резолвиться в IP `TARGET_HOST`;
- поднимает внутренний registry без auth в loopback-режиме (`127.0.0.1` на node) и собирает образ через Kaniko;
- автоматически настраивает `/etc/rancher/k3s/registries.yaml` для mirror на локальный registry (`http://127.0.0.1:<port>`);
- разворачивает PostgreSQL и `codex-k8s` в production namespace;
- TLS для `CODEXK8S_PRODUCTION_DOMAIN` управляется в runtime-deploy:
  - перед выпуском сертификата выполняется HTTP echo-probe (проверка доступности домена на этот кластер);
  - после выпуска TLS secret сохраняется в служебный namespace (`CODEXK8S_TLS_SYSTEM_NAMESPACE`) и переиспользуется в следующих деплоях;
- применяет baseline `NetworkPolicy` (platform namespace + labels для `system/platform` зон);
- включает host firewall hardening: с внешней сети доступны только `SSH`, `HTTP`, `HTTPS`;
- запрашивает внешние креды (`GitHub fine-grained token`, `CODEXK8S_OPENAI_API_KEY`), внутренние секреты генерирует автоматически;
- настраивает GitHub Environments (`production`, `ai`) и синхронизирует env-level secrets/variables для platform repo (`CODEXK8S_GITHUB_REPO`);
- создаёт или обновляет GitHub webhook и каталог labels в platform repo (`CODEXK8S_GITHUB_REPO`) и, если задан отдельный `CODEXK8S_FIRST_PROJECT_GITHUB_REPO`, дополнительно синхронизирует webhook/labels там;
- при старте `control-plane` автоматически создаёт/обновляет записи Project/Repositories в БД для `CODEXK8S_GITHUB_REPO`
  (и опционально для `CODEXK8S_FIRST_PROJECT_GITHUB_REPO`); platform project защищён от удаления через staff UI/API;
- разворачивает platform stack через Kubernetes API без зависимости от GitHub Actions workflows.

## Быстрый запуск

1. Скопируйте пример конфига:

```bash
cp bootstrap/host/config.env.example bootstrap/host/config.env
```

2. Заполните `bootstrap/host/config.env`.

3. Для полного bootstrap + deploy запускайте:

```bash
go run ./cmd/codex-bootstrap validate \
  --config services.yaml \
  --env production

go run ./cmd/codex-bootstrap preflight \
  --env-file bootstrap/host/config.env

go run ./cmd/codex-bootstrap bootstrap \
  --config services.yaml \
  --env-file bootstrap/host/config.env
```

Команда `bootstrap` после host provisioning автоматически запускает удалённый pipeline:
`runtime-deploy --prerequisites-only` -> `sync-secrets` -> `github-sync` -> `runtime-deploy`.

4. Низкоуровневый host-only скрипт (без post-provision deploy pipeline) оставлен для диагностики:

```bash
bash bootstrap/host/bootstrap_remote_production.sh
```

Для отдельного e2e контура:

```bash
go run ./cmd/codex-bootstrap bootstrap \
  --config services.yaml \
  --env-file bootstrap/host/config-e2e-test.env \
  --dry-run
```

Опции `preflight`:
- `--skip-ssh` — пропустить проверку SSH-доступа к target host.
- `--skip-github` — пропустить проверку доступа к GitHub API (repo/webhook/labels).
- `--timeout=30s` — увеличить timeout для сетевых проверок.

## Примечания

- Скрипты — каркас первого этапа. Перед production обязательны hardening и отдельный runbook.
- `bootstrap/host/bootstrap_remote_production.sh` может читать env из кастомного файла через `CODEXK8S_BOOTSTRAP_CONFIG_FILE`; по умолчанию используется `bootstrap/host/config.env`.
- `CODEXK8S_GITHUB_REPO` — platform repo (репозиторий с кодом `codex-k8s` и bootstrap/runtime metadata).
- `CODEXK8S_FIRST_PROJECT_GITHUB_REPO` (опционально) — отдельный репозиторий первого подключаемого проекта, где bootstrap дополнительно создаёт webhook и каталог labels; если пусто, используется только `CODEXK8S_GITHUB_REPO` (dogfooding).
- Platform secrets/variables (`CODEXK8S_*`) записываются в `CODEXK8S_GITHUB_REPO` на уровне GitHub Environments (`production`, `ai`).
  В `CODEXK8S_FIRST_PROJECT_GITHUB_REPO` bootstrap не записывает platform secrets (там только webhook/labels).
- Для bootstrap нужен `CODEXK8S_GITHUB_PAT` (fine-grained) с правами на `administration` (webhooks/labels), `secrets` и `variables`.
- Для staff UI и staff API требуется GitHub OAuth App:
  - создать на `https://github.com/settings/applications/new`;
  - `Homepage URL`: `https://<CODEXK8S_PRODUCTION_DOMAIN>`;
  - `Authorization callback URL` (production/dev через `oauth2-proxy`): `https://<CODEXK8S_PRODUCTION_DOMAIN>/oauth2/callback`;
  - заполнить `CODEXK8S_GITHUB_OAUTH_CLIENT_ID` и `CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET` в `bootstrap/host/config.env`.
- `CODEXK8S_PUBLIC_BASE_URL` должен совпадать с публичным URL (обычно `https://<CODEXK8S_PRODUCTION_DOMAIN>`).
- `CODEXK8S_BOOTSTRAP_OWNER_EMAIL` задаёт единственный email, которому разрешён первый вход (platform admin). Self-signup запрещён.
- `CODEXK8S_BOOTSTRAP_ALLOWED_EMAILS` (опционально) — дополнительные staff email'ы (через запятую),
  которые будут автоматически добавлены в БД при старте `api-gateway`, чтобы первый вход не упирался в
  `{"code":"forbidden","message":"email is not allowed"}`.
- `CODEXK8S_BOOTSTRAP_PLATFORM_ADMIN_EMAILS` (опционально) — дополнительные platform admin (owners) email'ы (через запятую),
  которые будут автоматически добавлены/обновлены в БД при старте `api-gateway` с `is_platform_admin=true`.
- `CODEXK8S_GITHUB_WEBHOOK_SECRET` используется для валидации `X-Hub-Signature-256`; если переменная пуста, bootstrap генерирует значение автоматически.
- `CODEXK8S_GITHUB_WEBHOOK_URL` (опционально) позволяет переопределить URL webhook; по умолчанию используется `https://<CODEXK8S_PRODUCTION_DOMAIN>/api/v1/webhooks/github`.
- `CODEXK8S_GITHUB_WEBHOOK_EVENTS` задаёт список событий webhook (comma-separated).
- `CODEXK8S_PLATFORM_DEPLOYMENT_REPLICAS` управляет replicas для platform `Deployment`-объектов (кроме PostgreSQL); для `production` по умолчанию `2`.
- Worker-параметры (`CODEXK8S_WORKER_*`) также синхронизируются в GitHub Variables и применяются при deploy.
- `CODEXK8S_LEARNING_MODE_DEFAULT` задаёт default для новых проектов (`true` в шаблоне; пустое значение = выключено).
- В `bootstrap/host/config.env` используйте только переменные с префиксом `CODEXK8S_` для платформенных параметров и секретов.
- `CODEXK8S_PRODUCTION_DOMAIN`, `CODEXK8S_AI_DOMAIN` и `CODEXK8S_LETSENCRYPT_EMAIL` обязательны.
- Для single-node/bare-metal production по умолчанию включён `CODEXK8S_INGRESS_HOST_NETWORK=true` (ingress слушает хостовые `:80/:443`).
- При `CODEXK8S_INGRESS_HOST_NETWORK=true` сервис ingress автоматически приводится к `ClusterIP`, чтобы не оставлять внешние `NodePort`.
- Внутренний registry работает без auth по design MVP и слушает только `127.0.0.1:<CODEXK8S_INTERNAL_REGISTRY_PORT>` на node.
- Loopback-режим registry рассчитан на single-node production; для multi-node нужен отдельный registry-профиль.
- По умолчанию включён baseline `NetworkPolicy` (`CODEXK8S_NETWORK_POLICY_BASELINE=true`).
- Чтобы worker мог обращаться к Kubernetes API, baseline также разрешает egress на API endpoint
  (для k3s обычно это `nodeIP:6443`). Управляется переменными:
  - `CODEXK8S_K8S_API_CIDR` (рекомендуется `TARGET_HOST/32` для single-node production);
  - `CODEXK8S_K8S_API_PORT` (по умолчанию `6443`).
- Для новых namespace проектов/агентов используйте `deploy/base/network-policies/project-agent-baseline.yaml.tpl`
  через runtime deploy (`services/internal/control-plane/internal/domain/runtimedeploy`) или вручную через
  `go run ./cmd/codex-bootstrap render-manifest --template deploy/base/network-policies/project-agent-baseline.yaml.tpl`.
- По умолчанию включён firewall hardening (`CODEXK8S_FIREWALL_ENABLED=true`), снаружи открыты только `CODEXK8S_SSH_PORT`, `80`, `443`.
