---
doc_id: EPC-CK8S-S1-D4
type: epic
title: "Epic Day 4: Repository provider and project repositories lifecycle"
status: planned
owner_role: EM
created_at: 2026-02-06
updated_at: 2026-02-06
related_issues: [1]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-06-day4"
---

# Epic Day 4: Repository provider and project repositories lifecycle

## TL;DR
- Цель эпика: включить управляемую работу с репозиториями через provider interface.
- Ключевая ценность: проект может иметь несколько репозиториев и собственные `services.yaml`.
- MVP-результат: GitHub adapter и CRUD для project repositories в staff API/UI.

## Priority
- `P0` (критично для multi-repo модели проекта).

## Ожидаемые артефакты дня
- Provider контракты и GitHub adapter в `services/internal/control-plane/**` и/или `libs/go/**`.
- Staff/private API и UI для CRUD project repositories.
- Поддержка per-repo `services.yaml` path и шифрование токенов.
- Smoke evidence: подключение 2+ репозиториев к одному проекту на staging.

## Контекст
- Почему эпик нужен: core-domain не завершён без реальных repo интеграций.
- Связь с требованиями: FR-002, FR-020, FR-021, FR-022.

## Scope
### In scope
- `RepositoryProvider` контракты и GitHub implementation.
- CRUD операций для подключённых репозиториев проекта.
- Хранение per-repo token в зашифрованном виде.
- Поддержка per-repo `services.yaml` path.

### Out of scope
- GitLab provider реализация.
- Переход на Vault/KMS.

## Декомпозиция (Stories/Tasks)
- Story-1: provider interfaces в доменном слое.
- Story-2: GitHub adapter для repo validation и webhook setup.
- Story-3: API/UI для добавления/обновления/удаления repositories.
- Story-4: шифрование/дешифрование token material в приложении.

## Data model impact (по шаблону data_model.md)
- Сущности:
  - `repositories`: `provider`, `owner`, `name`, `token_encrypted`, `services_yaml_path`.
- Связи/FK:
  - `repositories.project_id -> projects.id`.
- Индексы и запросы:
  - Индекс/уникальность `(project_id, provider, owner, name)`.
  - Индекс `repositories(project_id)`.
- Миграции:
  - Добавить недостающие ограничения уникальности и not-null.
- Retention/PII:
  - Token хранится только в шифрованном виде, без вывода в логи.

## Критерии приемки эпика
- В проект можно добавить несколько GitHub репозиториев.
- Для каждого репо сохраняется индивидуальный `services.yaml` path.
- Изменения задеплоены и проверены на staging в день реализации.

## Риски/зависимости
- Зависимости: корректные scope у `CODEXK8S_GITHUB_PAT`.
- Риск: ошибки ротации ключей шифрования токенов.

## План релиза (верхний уровень)
- Deploy на staging + ручной тест CRUD репозиториев и webhook wiring.

## Апрув
- request_id: owner-2026-02-06-day4
- Решение: approved
- Комментарий: Day 4 scope принят.
