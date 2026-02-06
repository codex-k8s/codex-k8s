# Bootstrap (staging)

Набор скриптов для первичного развёртывания `codex-k8s` на удалённом сервере Ubuntu 24.04.

## Что делает bootstrap

- запускается с хоста разработчика;
- подключается к удалённому серверу по SSH под `root`;
- создаёт отдельного операционного пользователя;
- ставит k3s и базовые сетевые компоненты;
- проверяет DNS до старта раскатки: `CODEXK8S_STAGING_DOMAIN` должен резолвиться в IP `TARGET_HOST`;
- разворачивает PostgreSQL и `codex-k8s` в staging namespace;
- создаёт `ClusterIssuer` (`codex-k8s-letsencrypt`) и выпускает TLS-сертификат через HTTP-01;
- запрашивает внешние креды (`GitHub PAT`, `CODEXK8S_OPENAI_API_KEY`), внутренние секреты генерирует автоматически;
- создаёт GHCR image pull secret (`ghcr-pull-secret`) из `CODEXK8S_GITHUB_USERNAME` + `CODEXK8S_GITHUB_PAT` (на сервере, без записи этих значений в файлы репозитория);
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
- Для деплоя через GitHub Actions нужен PAT с правами на repository/admin/actions.
- Workflow staging должен запускаться на `runs-on: <CODEXK8S_RUNNER_SCALE_SET_NAME>`.
- `CODEXK8S_LEARNING_MODE_DEFAULT` задаёт default для новых проектов (`true` в шаблоне; пустое значение = выключено).
- В `bootstrap/host/config.env` используйте только переменные с префиксом `CODEXK8S_` для платформенных параметров и секретов.
- `CODEXK8S_STAGING_DOMAIN` и `CODEXK8S_LETSENCRYPT_EMAIL` обязательны.
- Для single-node/bare-metal staging по умолчанию включён `CODEXK8S_INGRESS_HOST_NETWORK=true` (ingress слушает хостовые `:80/:443`).
