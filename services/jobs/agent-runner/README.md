# agent-runner

`agent-runner` — job-сервис запуска агентных сессий в Kubernetes: подготавливает runtime-контекст, выполняет run и собирает артефакты.

Prompt seed policy:
- task-body шаблон берётся из встроенного каталога `services/jobs/agent-runner/internal/runner/promptseeds/*.md` (embed) по связке `agent_key + trigger_kind + template_kind + locale`;
- role profile и контракты оформления follow-up Issue / PR / review / discussion рендерятся из
  `services/jobs/agent-runner/internal/runner/templates/prompt_blocks/*.tmpl`;
- поддержаны role-aware шаблоны:
  - `<stage>-<agent_key>-<kind>_<locale>.md` / `<stage>-<agent_key>-<kind>.md`;
  - `role-<agent_key>-<kind>_<locale>.md` / `role-<agent_key>-<kind>.md`;
- fallback chain: `stage+role -> role -> stage -> dev -> default -> embedded runner template`.

## Full-env repo cache

В `full-env` live сервисы и `agent-runner` работают с одним и тем же repo-cache PVC в `/workspace`:
branch/ref туда заранее доставляет runtime deploy (`repo-sync`) до запуска agent pod.

После этого runner больше не делает `git fetch/checkout/reset/clean` по живому дереву и не создаёт
вторую рабочую директорию. Агент работает прямо в уже подготовленном git worktree `/workspace`, а
из служебных файлов runner создаёт только короткоживущие объекты в `/tmp` (например, `git-askpass`
скрипт для git auth и временный каталог для проверки `codex` auth), чтобы не триггерить
hot-reload watcher-ы.

Для `run:*:revise` при переиспользовании живого namespace worker повторно не запускает runtime
prepare/repo-sync и просто стартует нового агента в существующем `/workspace`.

```text
services/jobs/agent-runner/                          runtime исполнитель агентных запусков
├── README.md                                        карта структуры сервиса и run-пайплайна
├── Dockerfile                                       image для выполнения agent-run job
├── cmd/agent-runner/main.go                         точка входа процесса runner
├── internal/
│   ├── app/                                         конфиг и bootstrap runner-приложения
│   ├── controlplane/client.go                       клиент внутренних API control-plane
│   └── runner/                                      основная логика запуска/мониторинга агентной сессии
│       ├── service.go                               orchestration жизненного цикла run
│       ├── helpers.go                               вспомогательные функции подготовки окружения
│       ├── kubectl_helpers.go                       утилиты работы с Kubernetes runtime
│       ├── security_helpers.go                      безопасная обработка токенов/секретных данных
│       ├── structs.go                               типы payload/runtime состояния runner
│       └── templates/                               шаблоны runtime-артефактов и job-конфигураций
└── scripts/                                         вспомогательные скрипты контейнера runner
    ├── bootstrap_tools.sh                           установка обязательного CLI toolchain
    └── entrypoint.sh                                стартовый скрипт контейнера
```
