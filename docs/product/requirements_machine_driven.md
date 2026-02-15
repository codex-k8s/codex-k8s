---
doc_id: REQ-CK8S-0001
type: requirements
title: "codex-k8s — Machine-Driven Requirements Baseline"
status: active
owner_role: PM
created_at: 2026-02-06
updated_at: 2026-02-15
related_issues: [1]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-12-project-docs-approval"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-12
---

# Requirements Baseline: codex-k8s

## TL;DR
- Этот документ фиксирует канонический набор требований для `codex-k8s` на основе решений Owner.
- Приоритет: требования здесь + обязательные стандарты `docs/design-guidelines/**`.
- При расхождениях другие продуктовые документы приводятся в соответствие с этим файлом.

## Правило приоритета
1. `docs/product/requirements_machine_driven.md` (этот документ) и явные решения Owner.
2. Технические стандарты `docs/design-guidelines/**` (как обязательные инженерные ограничения реализации).
3. Остальные продуктовые/архитектурные документы (`brief`, `constraints`, `delivery`, `c4`, `api_contract`, `data_model`).

## Functional Requirements (FR)

| ID | Требование |
|---|---|
| FR-001 | Платформа поддерживает только Kubernetes и работает через Go SDK (`client-go`), без поддержки других оркестраторов. |
| FR-002 | Интеграции с репозиториями реализуются через provider interface (`RepositoryProvider`): MVP с GitHub, с заделом на GitLab без перелома доменной логики. |
| FR-003 | Продуктовые процессы webhook-driven: бизнес-процессы запускаются webhook-событиями, а не workflow-first подходом. |
| FR-004 | Основное хранилище платформы: PostgreSQL с `JSONB` и `pgvector` для chunk-хранилища и векторного поиска. |
| FR-005 | `codex-k8s` и его PostgreSQL разворачиваются в Kubernetes. |
| FR-006 | Служебные MCP ручки платформы реализуются в Go внутри `codex-k8s`; `github.com/codex-k8s/yaml-mcp-server` остаётся пользовательским расширяемым слоем для кастомных ручек. |
| FR-007 | Staff frontend защищён GitHub OAuth. |
| FR-008 | Пользовательские настройки платформы хранятся в БД и редактируются через frontend; системные секреты и deploy-настройки `codex-k8s` берутся из env. |
| FR-009 | Шаблоны инструкций агентов, настройки агентов, сессии и журнал действий хранятся в БД и доступны в минимальном staff UI/API. |
| FR-010 | Набор агентов на MVP фиксирован штатной моделью из `machine_driven_company_requirements` с архитектурным заделом на будущие пользовательские агенты и процессы. |
| FR-011 | У агента есть `name`, `github_nick`, `email`, `token`; токены генерируются через API provider с нужными scope, ротируются платформой и хранятся в БД в зашифрованном виде. |
| FR-012 | Состояние запусков агентов, жизненный цикл pod/namespace и runtime-переходы хранятся в БД и отображаются в минимальном UI/API. |
| FR-013 | Поддерживается многоподовость `codex-k8s`; синхронизация между pod выполняется через БД; архитектура сразу разделяется на сервисы и jobs по зонам (`services/external|staff|internal|jobs|dev`). |
| FR-014 | Система слотов реализуется через БД. |
| FR-015 | Шаблоны документов (по `codexctl/docs/templates/**.md`) хранятся в БД и редактируются через markdown editor в staff UI (Monaco Editor). |
| FR-016 | Bootstrap поддерживает 2 режима: (a) deploy в уже существующий Kubernetes по kubeconfig; (b) установка k3s при отсутствии кластера (включая создание отдельного пользователя и базовый hardening). |
| FR-017 | Поддерживается любое количество проектов и базовая проектная RBAC-модель: `read`, `read_write`, `admin` (включая право удаления проекта). |
| FR-018 | Self-signup запрещён: пользователь допускается по email, заранее разрешённому администратором, и матчится при первом GitHub OAuth входе. |
| FR-019 | Добавление новых пользователей выполняется через staff UI по email и назначению доступов к проектам. |
| FR-020 | Каждый проект поддерживает несколько репозиториев; в каждом репозитории может быть свой `services.yaml`. |
| FR-021 | Доступ к каждому репозиторию задаётся отдельным токеном, который хранится в БД в зашифрованном виде; интеграция проектируется через интерфейсы для будущего перехода на Vault/JWT/KMS-подход без хранения токен-материала в БД. |
| FR-022 | Сам `codex-k8s` ведётся как проект с монорепозиторием и собственным `services.yaml`. |
| FR-023 | Learning mode: при user-initiated задачах в инструкции подмешивается блок объяснений (`почему так`, `что это даёт`, `какие альтернативы и почему хуже`), плюс после PR возможны образовательные комментарии по ключевым файлам/строкам. |
| FR-024 | Имена env/secrets/CI variables платформы используют префикс `CODEXK8S_` (исключения только для внешних контрактов). |
| FR-025 | На MVP public API ограничен webhook ingress; staff/private API используется для управления платформой. |
| FR-026 | В платформе фиксируется канонический каталог лейблов классов `run:*`, `state:*`, `need:*`, с поддержкой label-driven stage pipeline. |
| FR-027 | Для агент-инициированных trigger/deploy лейблов (`run:*`) обязателен апрув Owner до применения; `state:*` и `need:*` допускают автоустановку по политике проекта. |
| FR-028 | Процесс поставки фиксируется stage-моделью `intake -> vision -> prd -> arch -> design -> plan -> dev -> qa -> release -> postdeploy -> ops` с поддержкой `*:revise`, `run:rethink`. |
| FR-029 | Модель ролей агентов: базовый штат из 8 системных ролей (`pm`, `sa`, `em`, `dev`, `reviewer`, `qa`, `sre`, `km`) + расширяемые custom-агенты на проект (с явными RBAC, execution mode и quota). |
| FR-030 | Для агентных инструкций поддерживается role-specific матрица шаблонов: для каждого `agent_key` отдельные body-шаблоны `work/review`, с приоритетом `DB project override -> DB global override -> repo seed`. |
| FR-031 | Для агентных ролей поддерживается смешанный режим исполнения `full-env`/`code-only`, с политикой выбора режима по роли. |
| FR-032 | В БД как обязательный контур аудита и учета включены сущности `agent_sessions`, `token_usage`, `links` (в дополнение к `agent_runs` и `flow_events`). |
| FR-033 | Traceability поддерживается для всех этапов: связи Issue/PR ↔ docs ↔ stage status обновляются синхронно с выполнением этапов. |
| FR-034 | Шаблоны промптов рендерятся с runtime-контекстом (env/namespace/slot, project context, MCP servers/tools, issue/pr/run context). |
| FR-035 | Шаблоны промптов поддерживают локали; язык выбирается по цепочке `project locale -> system default locale -> en`, с базовой загрузкой `ru` и `en`. |
| FR-036 | Для каждой agent session сохраняется JSON-снимок `codex-cli` сессии, чтобы возобновлять выполнение с того же места после паузы/перезапуска. |
| FR-037 | Сущность `agent` остаётся центральной точкой настроек/политик выполнения (runtime mode, timeout policy, approval policy, prompt policy). |
| FR-038 | Для внешнего/staff HTTP API применяется contract-first OpenAPI: единая спецификация, runtime валидация запросов/ответов, codegen server DTO/backend stubs и frontend API client. |
| FR-039 | Approver/executor интеграции стандартизованы через HTTP-контракты MCP: платформа поддерживает встроенные MCP-ручки и внешний расширяемый слой (например, `github.com/codex-k8s/yaml-mcp-server` с Telegram/Slack/Mattermost/Jira адаптерами). |
| FR-040 | Staff UI на MVP предоставляет runtime-debug контур: список активных jobs, live/historical логи агентов, очередь ожидающих запусков с причинами ожидания (`waiting_mcp`, `waiting_owner_review`). |
| FR-041 | На MVP реализован минимальный набор MCP control tools: детерминированный secret sync (GitHub + Kubernetes) по окружению, database create/delete по окружению, owner feedback handle с вариантами ответа и custom input. |
| FR-042 | Для MCP control tools обязателен policy-driven approval matrix и audit trail с `correlation_id`, `approval_state`, `requested_by`, `applied_by`. |
| FR-043 | В stage taxonomy включён `run:self-improve` как основной контур самоулучшения: агент через MCP (`self_improve_runs_list`, `self_improve_run_lookup`, `self_improve_session_get`) анализирует run/session evidence, комментарии и артефакты по Issue/PR. |
| FR-044 | `run:self-improve` поддерживает управляемое применение результатов: change-set публикуется через PR с явной трассировкой `run/session source -> diagnosis -> change`, включая улучшения prompts/docs/guidelines/toolchain. |
| FR-045 | Для full MVP этапов поддерживается исполняемый контур полного stage-flow (`run:intake..run:ops`, `run:*:revise`, `run:rethink`) с traceability и audit-событиями. |
| FR-046 | Post-MVP roadmap фиксирует расширяемость платформы: управление prompt templates/агентами/лейблами через UI, knowledge lifecycle в `pgvector`, A2A swarm, периодические автономные run-циклы. |

## Non-Functional Requirements (NFR)

| ID | Требование |
|---|---|
| NFR-001 | Безопасность: секреты не логируются, repo токены хранятся в шифрованном виде, регистрация отключена, доступы через OAuth + RBAC. |
| NFR-002 | Масштабируемость: многоподовость `codex-k8s` с синхронизацией через PostgreSQL без конфликтов исполнения. |
| NFR-003 | Надёжность данных: `agent_runs` + `flow_events` являются базовым event/state контуром на MVP (без отдельного event_outbox). |
| NFR-004 | Производительность поиска знаний: базовый размер эмбеддинга `vector(3072)` в `pgvector`. |
| NFR-005 | Готовность к росту чтения: минимум одна asynchronous streaming read replica на MVP, с архитектурным заделом на 2+ replica и sync/quorum без изменений приложения. |
| NFR-006 | Развёртывание staging: выполняется bootstrap-скриптом с хоста разработчика по SSH на Ubuntu 24.04, включая настройку зависимостей и окружения. |
| NFR-007 | CI/CD для платформы: staging deploy автоматический на push в `main`; production deploy отдельным вручную запускаемым workflow с approval gate. |
| NFR-008 | Storage профиль MVP: `local-path`; переход на Longhorn отложен на следующий этап. |
| NFR-009 | Для agent-runs применяются управляемые лимиты параллелизма по ролям/проектам, чтобы избежать деградации кластера и гонок ресурсов. |
| NFR-010 | Любое stage/label действие трассируется с actor/correlation/approval state и доступно для аудита через БД и staff UI/API. |
| NFR-011 | Workflow-условия по лейблам используют repository variables (`vars.*`) вместо строковых литералов для предсказуемой переконфигурации. |
| NFR-012 | Пока агент ожидает ответ от MCP-сервера, timeout-kill pod/run не применяется; таймер выполнения должен быть paused для этого wait-state. |
| NFR-013 | Снимок `codex-cli` сессии для resumable run должен храниться надёжно и быть доступен для восстановления после перезапуска worker/pod. |
| NFR-014 | OpenAPI codegen должен быть воспроизводимым в CI: изменения спецификаций сопровождаются регенерацией backend/frontend артефактов и проверкой их актуальности. |
| NFR-015 | Для live/historical логов и wait-очереди в staff UI latency обновления должна быть достаточно низкой для операционной диагностики (целевой p95 UI refresh <= 5s на staging). |
| NFR-016 | MCP control tools должны быть идемпотентны и безопасны при retries/duplicate callbacks; secret material не должен попадать в model-visible output. |
| NFR-017 | Контур self-improve должен быть воспроизводим: одинаковый входной набор (логи/комментарии/артефакты) приводит к детерминированному diff-предложению в пределах версии шаблонов и policy. |
| NFR-018 | Полный stage-flow должен сохранять консистентность переходов и запрет недопустимых шагов даже при конкурирующих label-событиях. |

## Зафиксированные решения Owner (2026-02-06)

| Topic | Decision |
|---|---|
| Audit/log/chunks data scope | Отдельный логический БД-контур для audit/log/chunks в рамках PostgreSQL кластера MVP. |
| Read replica | Минимум одна async streaming replica на MVP, с последующим масштабированием без изменений приложения. |
| Staff auth | Short-lived JWT через API gateway. |
| Public API in first delivery | Только webhook ingress. |
| GitHub Enterprise/GHE provider in MVP | Не требуется. |
| OpenAI account mode | Production account подключается сразу. |
| Embedding size | `3072`. |
| Event outbox | Не вводится на MVP. |
| Runner scale | Локально: 1 persistent runner; staging/prod при наличии домена: autoscaled set. |
| Storage during bootstrap | `local-path` на MVP, Longhorn позже. |
| Learning mode default | Управляется через `bootstrap/host/config.env`; в шаблоне включён по умолчанию, пустое значение трактуется как выключено. |
| MVP completion scope | В MVP входят S2 Day6/Day7 + Sprint S3 Day1..Day12 (full stage labels, MCP control tools, `run:self-improve`, staff debug observability, declarative full-env deploy и Vuetify staff-console redesign). |

## Post-MVP направления (декомпозиция идей)
- Управление prompt templates и параметрами агентов через UI: версионирование, diff, rollout policy, rollback.
- Конструктор custom-агентов через web-console: role template, runtime mode, RBAC, quota, policy pack.
- Управление label taxonomy и stage policies через UI с governance approvals.
- Централизованный lifecycle документации (repo + DB + `pgvector`) с MCP-инструментами для поиска, impact-analysis и автообновления связей.
- Полноценная web-консоль нового поколения:
  - единый операционный workspace (runs, approvals, docs, agents, labels, metrics);
  - Vuetify app-shell + навигационный scaffold (Operations / Platform / Governance / Admin / Configuration).
  - Operations UX:
    - run timeline/stepper;
    - logs viewer (tail/search/copy/download);
    - approvals как “центр” (единый inbox + история решений).
  - Admin / Cluster (планируемая функциональность, к которой в MVP закладывается scaffold):
    - ресурсы: namespaces/configmaps/secrets/deployments/pods+logs/jobs+logs/pvc;
    - YAML view/edit через Monaco Editor;
    - safety guardrails:
      - в `ai-staging`/`prod` платформенные ресурсы помечаются `app.kubernetes.io/part-of=codex-k8s` (критерий для UI/guardrails и backend policy);
      - в `ai` (ai-slots) при dogfooding платформа может разворачиваться без `app.kubernetes.io/part-of=codex-k8s`, чтобы UI позволял тестировать действия над ресурсами самой платформы (в т.ч. destructive через dry-run);
      - ресурсы с `app.kubernetes.io/part-of=codex-k8s` нельзя удалять (UI и backend policy);
      - `ai-staging`/`prod` — строго view-only для ресурсов с `app.kubernetes.io/part-of=codex-k8s`;
      - ai-slots — destructive действия только dry-run (кнопки есть для dogfooding/debug, реальное действие не выполняется).
  - Agents + prompt templates:
    - управление настройками агента;
    - prompt templates `work/review` минимум в `ru/en` с diff/preview и preview effective template (редактор на Monaco).
  - System settings:
    - управление локалями (add locale + default locale);
    - базовые UI prefs (density, debug hints).
  - Обратная связь пользователю: алерты/снеки + notifications menu.
- A2A swarm контур: несколько агентов разных ролей работают параллельно в общем контексте задачи с protocol-level coordination.
- Периодические автономные run-циклы:
  - dependency freshness;
  - proactive security checks;
  - quality/doc drift detection;
  - scheduled `run:self-improve`.

## Ссылки
- `docs/product/brief.md`
- `docs/product/constraints.md`
- `docs/product/agents_operating_model.md`
- `docs/product/labels_and_trigger_policy.md`
- `docs/product/stage_process_model.md`
- `docs/architecture/c4_context.md`
- `docs/architecture/c4_container.md`
- `docs/architecture/data_model.md`
- `docs/architecture/api_contract.md`
- `docs/architecture/agent_runtime_rbac.md`
- `docs/architecture/mcp_approval_and_audit_flow.md`
- `docs/architecture/prompt_templates_policy.md`
- `docs/delivery/delivery_plan.md`
- `docs/delivery/sprint_s3_mvp_completion.md`
- `docs/delivery/epic_s3.md`
- `docs/delivery/issue_map.md`
- `docs/delivery/development_process_requirements.md`
- `docs/design-guidelines/AGENTS.md`

## Апрув
- request_id: owner-2026-02-06-mvp
- Решение: approved
- Комментарий: Канонический baseline требований зафиксирован.
