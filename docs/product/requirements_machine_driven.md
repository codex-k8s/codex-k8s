---
doc_id: REQ-CK8S-0001
type: requirements
title: "codex-k8s — Machine-Driven Requirements Baseline"
status: draft
owner_role: PM
created_at: 2026-02-06
updated_at: 2026-02-06
related_issues: [1]
related_prs: []
related_docsets: ["docs/_docset/issues/issue-0001-codex-k8s-bootstrap.md"]
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-06-mvp"
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
| FR-006 | Служебные MCP ручки платформы реализуются в Go внутри `codex-k8s`; `yaml-mcp-server` остаётся пользовательским расширяемым слоем для кастомных ручек. |
| FR-007 | Staff frontend защищён GitHub OAuth. |
| FR-008 | Пользовательские настройки платформы хранятся в БД и редактируются через frontend; системные секреты и deploy-настройки `codex-k8s` берутся из env. |
| FR-009 | Шаблоны инструкций агентов, настройки агентов, сессии и журнал действий хранятся в БД и доступны в минимальном staff UI/API. |
| FR-010 | Набор агентов на MVP фиксирован штатной моделью из `machine_driven_company_requirements` с архитектурным заделом на будущие пользовательские агенты и процессы. |
| FR-011 | У агента есть `name`, `github_nick`, `email`, `token`; токены генерируются через API provider с нужными scope, ротируются платформой и хранятся в БД в зашифрованном виде. |
| FR-012 | Состояние запусков агентов, жизненный цикл pod/namespace и runtime-переходы хранятся в БД и отображаются в минимальном UI/API. |
| FR-013 | Поддерживается многоподовость `codex-k8s`; синхронизация между pod выполняется через БД; архитектура сразу разделяется на сервисы и jobs по зонам (`services/external|staff|internal|jobs|dev`). |
| FR-014 | Система слотов реализуется через БД. |
| FR-015 | Шаблоны документов (по `codexctl/docs/templates/**.md`) хранятся в БД и редактируются через простой markdown editor в staff UI. |
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

## Ссылки
- `docs/product/brief.md`
- `docs/product/constraints.md`
- `docs/architecture/c4_context.md`
- `docs/architecture/c4_container.md`
- `docs/architecture/data_model.md`
- `docs/architecture/api_contract.md`
- `docs/delivery/delivery_plan.md`
- `docs/delivery/issue_map.md`
- `docs/delivery/development_process_requirements.md`
- `docs/design-guidelines/AGENTS.md`

## Апрув
- request_id: owner-2026-02-06-mvp
- Решение: approved
- Комментарий: Канонический baseline требований зафиксирован.
