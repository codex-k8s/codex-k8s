# Bootstrap (staging)

Набор скриптов для первичного развёртывания `codex-k8s` на удалённом сервере Ubuntu 24.04.

## Что делает bootstrap

- запускается с хоста разработчика;
- подключается к удалённому серверу по SSH под `root`;
- создаёт отдельного операционного пользователя;
- ставит k3s и базовые сетевые компоненты;
- разворачивает PostgreSQL и `codex-k8s` в staging namespace;
- запрашивает внешние креды (`GitHub PAT`, `OPENAI_API_KEY`), внутренние секреты генерирует автоматически;
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
- Workflow staging должен запускаться на `runs-on: <RUNNER_SCALE_SET_NAME>`.
- `LEARNING_MODE_DEFAULT` задаёт default для новых проектов (`true` в шаблоне; пустое значение = выключено).
