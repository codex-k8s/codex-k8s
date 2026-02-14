# Bootstrap (staging)

Набор скриптов для первичного развёртывания `codex-k8s` на удалённом сервере Ubuntu 24.04.

## Что делает bootstrap

- запускается с хоста разработчика;
- подключается к удалённому серверу по SSH под `root`;
- создаёт отдельного операционного пользователя;
- ставит k3s и базовые сетевые компоненты;
- проверяет DNS до старта раскатки: `CODEXK8S_STAGING_DOMAIN` должен резолвиться в IP `TARGET_HOST`;
- поднимает внутренний registry без auth в loopback-режиме (`127.0.0.1` на node) и собирает образ через Kaniko;
- автоматически настраивает `/etc/rancher/k3s/registries.yaml` для mirror на локальный registry (`http://127.0.0.1:<port>`);
- разворачивает PostgreSQL и `codex-k8s` в staging namespace;
- запускает декларативный deploy через `codex-bootstrap reconcile` (source of truth: `services.yaml`);
- после первичного bootstrap дальнейший runtime lifecycle task-namespace выполняется платформой (`control-plane + worker`), а не ручным запуском `codex-bootstrap`;
- создаёт `ClusterIssuer` (`codex-k8s-letsencrypt`) и выпускает TLS-сертификат через HTTP-01;
- применяет baseline `NetworkPolicy` (platform namespace + labels для `system/platform` зон);
- включает host firewall hardening: с внешней сети доступны только `SSH`, `HTTP`, `HTTPS`;
- запрашивает внешние креды (`GitHub fine-grained token`, `CODEXK8S_OPENAI_API_KEY`), внутренние секреты генерирует автоматически;
- настраивает GitHub repository secrets/variables для staging deploy workflow;
- создаёт или обновляет GitHub webhook на `https://<CODEXK8S_STAGING_DOMAIN>/api/v1/webhooks/github` и синхронизирует с `CODEXK8S_GITHUB_WEBHOOK_SECRET`;
- устанавливает ARC controller и runner scale set для staging deploy workflow.

## Быстрый запуск

1. Скопируйте пример конфига:

```bash
cp bootstrap/host/config.env.example bootstrap/host/config.env
```

2. Заполните `bootstrap/host/config.env`.

3. Запустите:

```bash
bash bootstrap/host/bootstrap_remote_staging.sh
```

## Примечания

- Скрипты — каркас первого этапа. Перед production обязательны hardening и отдельный runbook.
- Для деплоя через GitHub Actions нужен `CODEXK8S_GITHUB_PAT` (fine-grained) с правами на repository actions/secrets/variables и чтение содержимого репозитория.
- Для staff UI и staff API требуется GitHub OAuth App:
  - создать на `https://github.com/settings/applications/new`;
  - `Homepage URL`: `https://<CODEXK8S_STAGING_DOMAIN>`;
  - `Authorization callback URL` (staging/dev через `oauth2-proxy`): `https://<CODEXK8S_STAGING_DOMAIN>/oauth2/callback`;
  - заполнить `CODEXK8S_GITHUB_OAUTH_CLIENT_ID` и `CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET` в `bootstrap/host/config.env`.
- `CODEXK8S_PUBLIC_BASE_URL` должен совпадать с публичным URL (для staging обычно `https://<CODEXK8S_STAGING_DOMAIN>`).
- `CODEXK8S_BOOTSTRAP_OWNER_EMAIL` задаёт единственный email, которому разрешён первый вход (platform admin). Self-signup запрещён.
- `CODEXK8S_BOOTSTRAP_ALLOWED_EMAILS` (опционально) — дополнительные staff email'ы (через запятую),
  которые будут автоматически добавлены в БД при старте `api-gateway`, чтобы первый вход не упирался в
  `{"code":"forbidden","message":"email is not allowed"}`.
- `CODEXK8S_BOOTSTRAP_PLATFORM_ADMIN_EMAILS` (опционально) — дополнительные platform admin (owners) email'ы (через запятую),
  которые будут автоматически добавлены/обновлены в БД при старте `api-gateway` с `is_platform_admin=true`.
- `CODEXK8S_GITHUB_WEBHOOK_SECRET` используется для валидации `X-Hub-Signature-256`; если переменная пуста, bootstrap генерирует значение автоматически.
- `CODEXK8S_GITHUB_WEBHOOK_URL` (опционально) позволяет переопределить URL webhook; по умолчанию используется `https://<CODEXK8S_STAGING_DOMAIN>/api/v1/webhooks/github`.
- `CODEXK8S_GITHUB_WEBHOOK_EVENTS` задаёт список событий webhook (comma-separated).
- Workflow staging должен запускаться на `runs-on: <CODEXK8S_RUNNER_SCALE_SET_NAME>`.
- Worker-параметры (`CODEXK8S_WORKER_*`) также синхронизируются в GitHub Variables и применяются при deploy.
- `CODEXK8S_LEARNING_MODE_DEFAULT` задаёт default для новых проектов (`true` в шаблоне; пустое значение = выключено).
- В `bootstrap/host/config.env` используйте только переменные с префиксом `CODEXK8S_` для платформенных параметров и секретов.
- `CODEXK8S_STAGING_DOMAIN` и `CODEXK8S_LETSENCRYPT_EMAIL` обязательны.
- Для single-node/bare-metal staging по умолчанию включён `CODEXK8S_INGRESS_HOST_NETWORK=true` (ingress слушает хостовые `:80/:443`).
- При `CODEXK8S_INGRESS_HOST_NETWORK=true` сервис ingress автоматически приводится к `ClusterIP`, чтобы не оставлять внешние `NodePort`.
- Внутренний registry работает без auth по design MVP и слушает только `127.0.0.1:<CODEXK8S_INTERNAL_REGISTRY_PORT>` на node.
- Loopback-режим registry рассчитан на single-node staging; для multi-node нужен отдельный registry-профиль.
- Remote bootstrap устанавливает Go toolchain (`CODEXK8S_GO_VERSION`, default `1.25.6`) для запуска `codex-bootstrap`.
- По умолчанию включён baseline `NetworkPolicy` (`CODEXK8S_NETWORK_POLICY_BASELINE=true`).
- Чтобы worker мог обращаться к Kubernetes API, baseline также разрешает egress на API endpoint
  (для k3s обычно это `nodeIP:6443`). Управляется переменными:
  - `CODEXK8S_K8S_API_CIDR` (рекомендуется `TARGET_HOST/32` для single-node staging);
  - `CODEXK8S_K8S_API_PORT` (по умолчанию `6443`).
- Для новых namespace проектов/агентов используйте `deploy/base/network-policies/project-agent-baseline.yaml.tpl` через `deploy/scripts/apply_network_policy_baseline.sh`.
- По умолчанию включён firewall hardening (`CODEXK8S_FIREWALL_ENABLED=true`), снаружи открыты только `CODEXK8S_SSH_PORT`, `80`, `443`.
