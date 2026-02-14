---
doc_id: BRF-CK8S-0001
type: brief
title: "codex-k8s platform bootstrap"
status: active
owner_role: PM
created_at: 2026-02-06
updated_at: 2026-02-14
related_issues: [1, 19]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-12-project-docs-approval"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-12
---

# Brief: codex-k8s platform bootstrap

## TL;DR (1 абзац)
- **Проблема:** текущая связка `codexctl` + `github.com/codex-k8s/yaml-mcp-server` + ручные практики разнесена по репозиториям и не даёт единого control-plane.
- **Для кого:** Owner и команда, управляющие несколькими проектами и агентами в Kubernetes.
- **Предлагаемое решение:** единый сервис `codex-k8s` (Go + Vue3), webhook-driven, с хранением состояния и знаний в PostgreSQL (`JSONB` + `pgvector`).
- **Почему сейчас:** принято решение консолидировать архитектуру и убрать workflow-first оркестрацию продуктовых процессов.
- **Что считаем успехом:** staging разворачивается одним bootstrap-скриптом, push в `main` обновляет staging, ручные тесты проходят через UI и webhook сценарии.
- **Что считаем успехом (расширено):** кроме базового dogfooding, в MVP работают full stage labels, MCP control tools (secret/db/feedback), staff debug observability и `run:self-improve`.
- **Дополнительная ценность:** при включённом learning mode платформа объясняет важные инженерные решения и компромиссы, чтобы пользователи учились паттернам, а не только получали код.
- **Что НЕ делаем:** поддержку не-Kubernetes оркестраторов и self-signup пользователей.
- **Source of truth требований:** `docs/product/requirements_machine_driven.md`.

## Контекст
- Предыстория: в `project-example` и `codexctl` собран рабочий базис, но он распределён по отдельным компонентам.
- Текущее состояние: новый репозиторий `codex-k8s` создан, структура и гайды перенесены/актуализированы.
- Почему это важно: нужна единая платформа управления агентами, слотами, вебхуками, MCP-инструментами и документами.

## Цель
- Бизнес-цель: сократить time-to-delivery и операционные издержки за счёт единой платформы.
- Техническая цель: реализовать MVP control-plane для Kubernetes + GitHub с расширяемостью под GitLab.

## Пользователи и сценарий (очень кратко)
- Персона/роль: Owner (администратор платформы), инженер проекта.
- Основной сценарий: подключить проект и репозитории, принимать webhook-события, запускать агентные процессы, смотреть статусы/логи в UI.
- Болезнь/боль: разрозненные инструменты, ручная синхронизация, слабая трассируемость процессов.

## Предлагаемое решение (в 3–7 буллетов)
- Реализовать сервисы: `api-gateway`, `control-plane`, `worker`, `web-console`.
- Сделать webhook ingestion как основной вход запуска процессов.
- Хранить конфигурации пользователей/агентов/проектов/слотов/документов в PostgreSQL.
- Реализовать встроенные MCP service-tools в Go внутри платформы.
- Защитить UI через GitHub OAuth с матчингом email.
- Добавить bootstrap-скрипт развёртывания staging по SSH на Ubuntu 24.04.
- Включить CI/CD deploy для самой платформы через self-hosted runner в Kubernetes (staging first).
- Зафиксировать stage-driven delivery модель и label taxonomy (`run:*`, `state:*`, `need:*`) как единый процессный контракт.
- Зафиксировать operating model агентов: базовый штат из 8 ролей (включая `dev` и `reviewer`) + custom-агенты проекта, mixed runtime (`full-env`/`code-only`).
- Зафиксировать review контур: для всех `run:*` выполняется pre-review (`reviewer` и/или профильная роль), затем финальный Owner review.
- Зафиксировать policy шаблонов промптов: seed в репозитории + override в БД (`work/review`).
- Добавить locale-aware prompt policy: `project locale -> system default -> en`, с baseline локалями `ru` и `en`.
- Добавить resumable lifecycle для agent runs: сохранение `codex-cli` session JSON и возобновление после пауз.
- Ввести contract-first OpenAPI для external/staff API с codegen server/client до расширения транспорта внешних клиентов.
- Добавить режим обучения для пользовательских задач:
  - подмешивание в инструкции требований объяснять "почему так";
  - дополнительный post-PR образовательный комментарий по ключевым файлам/строкам.
- Добавить self-improve контур `run:self-improve`:
  - анализ логов запусков + комментариев Owner/бота;
  - автоматическое предложение улучшений docs/prompt templates/instructions/tooling через PR.

## Границы
### In scope (входит)
- Kubernetes-only control-plane.
- GitHub provider (первый), provider interface под GitLab.
- PostgreSQL + `JSONB` + `pgvector`.
- Bootstrap staging + runner setup + deploy pipeline.

### Out of scope (не входит)
- Multi-orchestrator support.
- Полноценный marketplace пользовательских агентов.
- Полный отказ от GitHub Actions для deploy самой платформы на этапе MVP.

## Метрики успеха (первичная версия)
- KPI/OKR: время от чистого Ubuntu 24.04 сервера до готового staging <= 45 минут.
- Продуктовые метрики: не менее 1 проекта и 2 репозиториев подключаются через UI без ручного SQL.
- Технические метрики: 99% webhook событий обрабатываются идемпотентно без дублей; p95 API < 500ms для CRUD настроек.

## Ограничения
- Сроки: MVP с staging bootstrap и базовым deploy-пайплайном в первой итерации.
- Ресурсы: один staging сервер Ubuntu 24.04.
- Платформы/технологии: Go, Vue3, Kubernetes, PostgreSQL.
- Регуляторика/безопасность: запрет self-signup; секреты не логируются; repo токены шифруются.

## Риски и допущения
- Риск: root SSH bootstrap может быть хрупким на нестандартных образах Ubuntu.
- Допущение: доступен GitHub fine-grained token с правами на repo/actions/secrets/variables для runner и webhook-настроек.
- Риск: learning mode может зашумлять PR комментарии при слабой фильтрации "важных мест".

## Решение по deploy workflow (принято)
- Для `staging`: deploy workflow запускается автоматически на push в `main`.
- Для `production`: отдельный deploy workflow запускается вручную (`workflow_dispatch`) и проходит environment approval.
- Bootstrap-скрипт на первом этапе настраивает runner и переменные для `staging`.

## Решения Owner (зафиксировано)
- Storage профиль MVP: упрощённый `local-path`, миграция на Longhorn позже.
- Learning mode default: задаётся переменной в `bootstrap/host/config.env`; в шаблоне значение по умолчанию включено, при пустом значении default считается выключенным.
- Public API на первой поставке: только webhook ingress.
- Staff API auth: short-lived JWT через API gateway.
- GitHub Enterprise/GHE provider: не требуется на этапе MVP.
- Production OpenAI account: подключается сразу.

## Post-MVP идеи и направления
- Управление prompt templates и agent policies через web-console: версия, diff, rollout, rollback, аудит.
- Конструктор custom-агентов: создание роли через UI, выбор runtime mode, tool-policy, quotas и RBAC.
- Управление лейблами/stage-политикой через UI с апрувами изменений.
- Единый контур документации (repo + DB + `pgvector`) с MCP-ручками поиска, анализа влияния и синхронизации.
- Новая полноценная staff-консоль:
  - единый рабочий стол операций по run/approval/docs/agents;
  - компонентный подход на UI-библиотеке для admin-сценариев (кандидат: `Vuetify`, capabilities сверены через Context7).
- A2A swarm концепция: параллельные агенты разных ролей в одном процессе с протоколом координации.
- Периодические автономные циклы: security/dependency/docs drift checks, плановые `run:self-improve`, поиск улучшений по телеметрии.

## Решение от Owner (что нужно утвердить)
- [x] Принять brief как базу и перейти к Vision/Architecture
- [ ] Запросить правки
- [ ] Отклонить/заморозить инициативу

## Ссылки
- Канонические требования: `docs/product/requirements_machine_driven.md`
- Операционная модель агентов: `docs/product/agents_operating_model.md`
- Политика лейблов: `docs/product/labels_and_trigger_policy.md`
- Stage-модель процесса: `docs/product/stage_process_model.md`
- Процесс разработки и документации: `docs/delivery/development_process_requirements.md`
- Issue: #1
- Трассируемость Issue/PR ↔ документы: `docs/delivery/issue_map.md`
- Связанные ADR: `ADR-0001`, `ADR-0002`, `ADR-0003`, `ADR-0004`

## Апрув
- Запрошен: 2026-02-06 (request_id: owner-2026-02-06-mvp)
- Решение: approved
- Комментарий: Уточнения по MVP и bootstrap/deploy модели зафиксированы.
