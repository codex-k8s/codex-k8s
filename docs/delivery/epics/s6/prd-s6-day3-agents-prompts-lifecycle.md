---
doc_id: PRD-CK8S-S6-I187
type: prd
title: "Issue #187 — PRD: Agents configuration and prompt templates lifecycle"
status: completed
owner_role: PM
created_at: 2026-02-25
updated_at: 2026-02-25
related_issues: [184, 185, 187, 189]
related_prs: [190]
related_docsets:
  - docs/delivery/sprints/s6/sprint_s6_agents_prompt_management.md
  - docs/delivery/epics/s6/epic_s6.md
  - docs/delivery/issue_map.md
  - docs/delivery/requirements_traceability.md
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-25-issue-187-prd"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-25
---

# PRD: Agents configuration and prompt templates lifecycle (Issue #187)

## TL;DR
- Что строим: управляемый staff-контур для настройки агентов, lifecycle шаблонов промптов и полной истории изменений.
- Для кого: platform owner, PM/SA/EM и staff-команда, управляющая агентами и stage-переходами.
- Почему: текущий раздел `Agents` в staff UI остается scaffold, а contract-first API для agents/templates/audit отсутствует.
- MVP: list/details/settings для агентов + шаблоны `work/revise` (`ru/en`) + diff/effective preview + versioning + audit history.
- Критерии успеха: сценарии AC-01..AC-08 выполняются без нарушения RBAC/policy и с полным audit trail.

## Проблема и цель
- Problem statement:
  - staff-пользователь не может безопасно и прозрачно управлять агентами/шаблонами в production-процессе;
  - изменения шаблонов промптов не имеют полного lifecycle-контроля через API/UI контур.
- Цели:
  - формализовать продуктовые требования к управлению агентами и шаблонами;
  - зафиксировать проверяемые acceptance criteria;
  - подготовить NFR-draft и handover-пакет в архитектурный этап.
- Почему сейчас:
  - после intake (`#184`) и vision (`#185`) сформирован достаточный baseline для детального PRD перед `run:arch`.

## Пользователи / Персоны
- Persona A: Platform Owner, который утверждает policy и контролирует риски rollout.
- Persona B: PM/EM/KM, которым нужна трассируемость изменений и предсказуемость поведения агентов.
- Persona C: Staff-оператор с доступом `read_write/admin`, который редактирует шаблоны и проверяет эффекты изменений.

## Сценарии / Use Cases (кратко)
- UC-1: просмотреть список агентов и текущие настройки по проекту.
- UC-2: обновить параметры агента с проверкой прав и фиксацией аудита.
- UC-3: создать новую версию prompt template (`work`/`revise`) для `ru/en`.
- UC-4: сравнить версии через diff перед публикацией.
- UC-5: получить effective preview с учетом fallback-цепочки.
- UC-6: выполнить rollback на предыдущую версию и получить подтвержденный audit след.

## Требования (Functional Requirements)
- FR-187-01: staff-контур предоставляет list/details/settings для агентов по проекту.
- FR-187-02: поддерживается lifecycle шаблонов `work/revise` по локалям `ru/en`.
- FR-187-03: каждая правка шаблона формирует новую версию с явным статусом (`draft`/`active`/`archived`).
- FR-187-04: перед публикацией доступен diff между выбранными версиями шаблона.
- FR-187-05: доступен effective preview по цепочке `project override -> global override -> repo seed` с указанием источника.
- FR-187-06: для publish/rollback/update-agent-settings ведется полный audit trail в привязке к `flow_events`.
- FR-187-07: все операции защищены проектным RBAC (`read`, `read_write`, `admin`) без bypass path.
- FR-187-08: mutation-операции должны принимать и сохранять reason/comment изменения для последующего review.
- FR-187-09: поддерживается rollback активной версии шаблона на предыдущую с идемпотентным поведением.
- FR-187-10: staff-пользователь может просмотреть историю изменений агента/шаблона в едином журнале.
- FR-187-11: PRD-артефакты должны сохранять stage continuity по цепочке `intake -> vision -> prd -> arch`.
- FR-187-12: после завершения PRD обязательно создается issue следующего этапа для stage `run:arch` без trigger-лейбла (лейбл ставит Owner), с инструкцией создать issue для `run:design`.

## Acceptance Criteria (Given/When/Then)
- AC-01
  - Given staff-пользователь с правом `read`
  - When открывает раздел `Agents`
  - Then видит список агентов и актуальные настройки без mock-данных.
- AC-02
  - Given staff-пользователь с правом `read_write`
  - When редактирует параметры агента и сохраняет изменения
  - Then изменения применяются и фиксируются в audit history с `actor` и `correlation_id`.
- AC-03
  - Given существует активная версия шаблона
  - When пользователь создает новую версию и открывает diff
  - Then система показывает разницу между версиями до публикации.
- AC-04
  - Given для шаблона отсутствует project override
  - When пользователь запрашивает effective preview
  - Then система возвращает preview из следующего источника fallback и указывает origin.
- AC-05
  - Given опубликована новая версия шаблона
  - When пользователь запрашивает историю изменений
  - Then в журнале есть событие публикации с контекстом issue/pr и причиной изменения.
- AC-06
  - Given пользователь без прав `read_write/admin`
  - When пытается выполнить mutation-операцию
  - Then операция отклоняется и отражается в audit как denied access.
- AC-07
  - Given пользователь выполняет rollback шаблона
  - When операция отправлена повторно с тем же request context
  - Then результат идемпотентен и не создает конфликтующих активных версий.
- AC-08
  - Given PRD-этап завершен
  - When проверяется stage continuity
  - Then создана issue `#189` для stage `run:arch` без trigger-лейбла (лейбл ставит Owner) с обязательной инструкцией для следующего этапа `run:design`.

## Non-Goals (явно)
- Реализация backend/frontend кода в рамках текущего PRD-этапа.
- Пересмотр глобальной taxonomy labels и stage-модели платформы.
- Введение новых типов системных ролей агентов вне текущего roster.

## Нефункциональные требования (NFR draft)
- NFR-187-01 (Security): mutation-операции допускаются только при валидном RBAC и не раскрывают секреты в логах/ответах.
- NFR-187-02 (Auditability): 100% mutation-событий по agents/templates имеют `actor`, `timestamp`, `correlation_id`, `issue/pr context`.
- NFR-187-03 (Observability): доступны метрики успех/ошибка и latency по операциям `list`, `publish`, `rollback`, `preview`, `diff`.
- NFR-187-04 (Performance/UX): целевые бюджеты p95 на MVP — list/details <= 2s, diff <= 2s, effective preview <= 3s.
- NFR-187-05 (Reliability): publish/rollback operations устойчивы к retry и не создают расходящихся active-version состояний.
- NFR-187-06 (Localization): поддержка `ru/en` с детерминированным fallback и фиксируемым origin шаблона.

## UX/UI заметки
- В diff/preview режимах пользователю показывается источник версии и статус (`draft`/`active`).
- В history-view отображаются actor, причина изменения и link на связанный issue/pr.
- Все destructive действия (`publish`, `rollback`) подтверждаются и требуют reason/comment.

## Аналитика и события (Instrumentation)
- События:
  - `agents.settings.updated`
  - `prompt_template.version_created`
  - `prompt_template.diff_viewed`
  - `prompt_template.effective_preview_viewed`
  - `prompt_template.published`
  - `prompt_template.rolled_back`
  - `agents.audit_log_viewed`
- Атрибуты:
  - `project_id`, `agent_key`, `template_kind`, `locale`, `version_id`, `actor_role`, `result`.
- Метрики:
  - publish success rate;
  - rollback success rate;
  - p95/p99 latency для diff/preview;
  - число denied-access попыток.

## Зависимости
- Внешние системы: GitHub labels/issue workflow для stage continuity.
- Команды/сервисы: `pm`, `sa`, `em`, `dev`, `qa`, `km`, `sre`.
- Технические зависимости: contract-first API контур и доменная модель `agents/agent_policies/prompt_templates/flow_events`.

## Риски и вопросы
- Риски:
  - конкурирующие изменения одного шаблона могут приводить к конфликтам версий;
  - неполный audit при частичных отказах publish/rollback;
  - рост latency effective preview из-за сложной fallback-цепочки.
- Вопросы для `run:arch`:
  - какой механизм консистентности версий выбрать (optimistic locking vs revision tokens);
  - где проходит граница между history-view и операционным audit журналом;
  - как закрепить latency budget в архитектурных ограничениях без деградации UX.

## План релиза (черновик)
- Ограничения выката: реализация в `run:dev` только после завершения `run:arch` и `run:design`.
- Риски релиза: policy drift между UI и backend при несвоевременной синхронизации контрактов.
- Роллбек: возврат на предыдущую активную версию шаблона с обязательной audit фиксацией.

## Handover в `run:arch`
- Следующий этап: `run:arch` (Issue `#189`).
- Trigger-лейбл `run:arch` на issue `#189` ставит Owner.
- Ожидаемые архитектурные выходы:
  - сервисные границы и ownership data/transport policy;
  - ADR по versioning/locking/audit consistency;
  - архитектурный пакет для следующего этапа `run:design`.
- Обязательное правило continuity:
  - по завершении `run:arch` создать issue на stage `run:design` без trigger-лейбла, с явной инструкцией создать issue следующего этапа `run:plan`.

## Приложения
- `docs/delivery/epics/s6/epic-s6-day1-agents-prompts-intake.md`
- `docs/delivery/epics/s6/epic-s6-day3-agents-prompts-prd.md`
- `docs/delivery/sprints/s6/sprint_s6_agents_prompt_management.md`
- `docs/delivery/issue_map.md`
- `docs/delivery/requirements_traceability.md`

## Апрув
- request_id: owner-2026-02-25-issue-187-prd
- Решение: pending
- Комментарий: PRD пакет подготовлен и передан на review.
