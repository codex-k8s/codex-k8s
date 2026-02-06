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
- создаёт `ClusterIssuer` (`codex-k8s-letsencrypt`) и выпускает TLS-сертификат через HTTP-01;
- включает host firewall hardening: с внешней сети доступны только `SSH`, `HTTP`, `HTTPS`;
- запрашивает внешние креды (`GitHub fine-grained token`, `CODEXK8S_OPENAI_API_KEY`), внутренние секреты генерирует автоматически;
- настраивает GitHub repository secrets/variables для staging deploy workflow;
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
- Workflow staging должен запускаться на `runs-on: <CODEXK8S_RUNNER_SCALE_SET_NAME>`.
- `CODEXK8S_LEARNING_MODE_DEFAULT` задаёт default для новых проектов (`true` в шаблоне; пустое значение = выключено).
- В `bootstrap/host/config.env` используйте только переменные с префиксом `CODEXK8S_` для платформенных параметров и секретов.
- `CODEXK8S_STAGING_DOMAIN` и `CODEXK8S_LETSENCRYPT_EMAIL` обязательны.
- Для single-node/bare-metal staging по умолчанию включён `CODEXK8S_INGRESS_HOST_NETWORK=true` (ingress слушает хостовые `:80/:443`).
- При `CODEXK8S_INGRESS_HOST_NETWORK=true` сервис ingress автоматически приводится к `ClusterIP`, чтобы не оставлять внешние `NodePort`.
- Внутренний registry работает без auth по design MVP и слушает только `127.0.0.1:<CODEXK8S_INTERNAL_REGISTRY_PORT>` на node.
- Loopback-режим registry рассчитан на single-node staging; для multi-node нужен отдельный registry-профиль.
- По умолчанию включён firewall hardening (`CODEXK8S_FIREWALL_ENABLED=true`), снаружи открыты только `CODEXK8S_SSH_PORT`, `80`, `443`.
