# agent-runner

`agent-runner` — job-сервис запуска агентных сессий в Kubernetes: подготавливает runtime-контекст, выполняет run и собирает артефакты.

Prompt seed policy:
- task-body шаблон берётся из встроенного каталога `services/jobs/agent-runner/internal/runner/promptseeds/*.md` (embed) по связке `agent_key + trigger_kind + template_kind + locale`;
- role profile и контракты оформления follow-up Issue / PR / review / discussion рендерятся из
  `services/jobs/agent-runner/internal/runner/templates/prompt_blocks/*.tmpl`;
- deterministic resume path для built-in user interactions читает persisted `interaction_resume_payload` через run-bound gRPC lookup в `control-plane` и добавляет typed JSON block в начало resume prompt перед `codex exec resume`;
- GitHub rate-limit handoff path отправляет typed `ReportGitHubRateLimitSignal`, сохраняет coarse session snapshots со статусами `running -> waiting_backpressure`, прекращает local retry-loop и для resume читает persisted `github_rate_limit_resume_payload` через run-bound gRPC lookup вместо повторного derive semantics из stderr/headers;
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

Для `run:*:revise` worker сначала проверяет reusable namespace через persisted runtime fingerprint
(`build_ref` должен быть immutable SHA, fingerprint и rendered manifests должны совпадать, namespace не
должен быть в `Terminating`, а в том же namespace не должно быть активной `runtime_deploy_task`).

Только при положительной проверке worker пропускает runtime prepare/repo-sync и стартует нового агента
в существующем `/workspace`. При любой инвалидации fast-path отключается, control-plane делает обычный
runtime deploy/repo-sync в тот же namespace, после чего runner получает уже обновлённый `/workspace`.

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
