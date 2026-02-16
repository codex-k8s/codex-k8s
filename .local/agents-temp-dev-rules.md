# Временные правила разработки и дебага (до внедрения dogfooding через run:dev)

Назначение: зафиксировать текущий “ручной” контур разработки `codex-k8s`, пока платформа ещё не запускает разработку сама через Issue/лейблы.

## Базовый рабочий цикл

1. Работаем в ветке `codex/dev`.
2. Любые правки (код/доки/манифесты) пушим в `codex/dev` в GitHub.
3. Когда готово: создаём PR в `main`.
4. После merge в `main` деплой на production запускается самой платформой через webhook (push в `main` -> deploy-only run -> runtime-deploy).
5. Проверяем состояние деплоя через `kubectl` и логи/объекты в Kubernetes на production.
6. Дебажим/фиксим/перепроверяем соответствие `docs/design-guidelines/**`, обновляем проектную документацию.
7. Дальше Owner оставляет замечания в GitHub (inline comments/общие комментарии) от себя или другого пользователя.
8. После merge ветка `codex/dev` пересоздаётся от нового `main`, и цикл повторяется.

## Источники кредов и переменных

Все временные “ручные” доступы и параметры берём из локального файла:
- `bootstrap/host/config.env` (он в `.gitignore`)

Шаблон и имена переменных см.:
- `bootstrap/host/config.env.example`

Ключевые группы переменных:
- GitHub: `CODEXK8S_GITHUB_REPO`, `CODEXK8S_GITHUB_PAT`
- Production SSH: `TARGET_HOST`, `TARGET_PORT`, `TARGET_ROOT_USER`, `TARGET_ROOT_SSH_KEY`, `OPERATOR_USER`
- Production namespace/domain: `CODEXK8S_PRODUCTION_NAMESPACE`, `CODEXK8S_PRODUCTION_DOMAIN`, `CODEXK8S_AI_DOMAIN`
- Прочее: OAuth/JWT/webhook параметры под префиксом `CODEXK8S_*`

## GitHub: как смотреть PR и конфигурацию

```bash
source bootstrap/host/config.env
export GH_TOKEN="$CODEXK8S_GITHUB_PAT"

gh pr view -R "$CODEXK8S_GITHUB_REPO" <pr_number> --comments

# Repo-level metadata (webhook/labels).
gh api -X GET "repos/${CODEXK8S_GITHUB_REPO}/hooks" --jq '.[].config.url'
gh label list -R "$CODEXK8S_GITHUB_REPO" --limit 200

# Environment-level config (Actions Environments).
gh secret list -R "$CODEXK8S_GITHUB_REPO" -e production
gh variable list -R "$CODEXK8S_GITHUB_REPO" -e production

gh secret list -R "$CODEXK8S_GITHUB_REPO" -e ai
gh variable list -R "$CODEXK8S_GITHUB_REPO" -e ai
```

## Production: как смотреть Kubernetes и логи

```bash
source bootstrap/host/config.env
ssh -i "$TARGET_ROOT_SSH_KEY" -p "$TARGET_PORT" "${TARGET_ROOT_USER}@${TARGET_HOST}"

kubectl -n "$CODEXK8S_PRODUCTION_NAMESPACE" get pods -o wide
kubectl -n "$CODEXK8S_PRODUCTION_NAMESPACE" get deploy,job,ingress
kubectl -n "$CODEXK8S_PRODUCTION_NAMESPACE" logs deploy/codex-k8s --tail=200
kubectl -n "$CODEXK8S_PRODUCTION_NAMESPACE" logs deploy/codex-k8s-worker --tail=200
```

Примечание: после bootstrap kubeconfig кладётся оператору в `/home/${OPERATOR_USER}/.kube/config`, поэтому можно выполнять `kubectl` без `sudo` под `OPERATOR_USER`.

## Smoke/проверки

Smoke для production выполняется вручную через `kubectl` проверки из этого файла и `docs/ops/production_runbook.md`.

## Важные текущие технические договорённости

- Миграции (goose) лежат в держателе схемы:
  - `services/internal/control-plane/cmd/cli/migrations/*.sql`
- Внешние изменения всегда сверяем с:
  - `AGENTS.md`
  - `docs/design-guidelines/**`
  - `docs/product/requirements_machine_driven.md`
