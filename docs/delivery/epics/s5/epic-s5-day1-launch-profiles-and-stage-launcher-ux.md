---
doc_id: EPC-CK8S-S5-D1
type: epic
title: "Epic S5 Day 1: Launch profiles and deterministic next-step actions (Issues #154/#155)"
status: in-review
owner_role: PM
created_at: 2026-02-24
updated_at: 2026-02-25
related_issues: [154, 155]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-24-issue-155-day1-vision-prd"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-25
---

# Epic S5 Day 1: Launch profiles and deterministic next-step actions (Issues #154/#155)

## TL;DR
- Проблема: текущий label-driven flow неудобен для ручного управления; пользователи забывают порядок этапов, а часть быстрых ссылок не работает.
- Day1 split:
  - Issue #154: intake baseline и фиксация проблемы;
  - Issue #155: vision/prd пакет с каноническими профилями и детерминированным next-step UX.
- Результат Day1: подготовлен owner-ready execution package для входа в `run:dev`; архитектурные границы и контракты закреплены ADR-0008.

## Priority
- `P0`.

## Scope
### In scope
- Ввести launch profiles:
  - `quick-fix` для минимальных исправлений;
  - `feature` для функциональных доработок;
  - `new-service` для полного цикла с архитектурной проработкой.
- Зафиксировать mapping profile -> обязательные стадии -> условия эскалации.
- Формализовать deterministic next-step actions:
  - primary: staff web-console deep-link;
  - fallback: текстовая команда label transition (`gh`/MCP path) для copy-paste.
- Сформировать acceptance matrix для UX и policy поведения.

### Out of scope
- Реализация UI/Backend кода.
- Пересмотр базовых label-классов и security policy.
- Изменения процесса ревью вне связки stage launch / stage transition.

## Vision/PRD package (Issue #155)

- Канонический PRD-артефакт Day1: `docs/delivery/epics/s5/prd-s5-day1-launch-profiles-and-stage-launcher-ux.md` (по шаблону `docs/templates/prd.md`).
- Канонический ADR-артефакт Day1: `docs/architecture/adr/ADR-0008-profile-driven-stage-launch-and-next-step-contract.md` (service boundaries + runtime impact/migration).
- Канонический API-contract sync Day1: `docs/architecture/api_contract.md` (норматив next-step action-card payload + audit/ambiguity guardrails).

### 1. Канонический набор launch profiles и эскалации

| Profile | Обязательная траектория | Когда применять | Детерминированные триггеры эскалации |
|---|---|---|---|
| `quick-fix` | `intake -> plan -> dev -> qa -> release -> postdeploy -> ops` | точечная правка в одном сервисе без изменения контрактов/схемы | любой из триггеров: `cross-service impact`, новая интеграция, миграция БД, изменение RBAC/policy |
| `feature` | `intake -> prd -> design -> plan -> dev -> qa -> release -> postdeploy -> ops` | функциональное изменение в существующих сервисах | при изменении архитектурных границ/NFR добавить `arch`; при изменении продуктовой стратегии добавить `vision` |
| `new-service` | `intake -> vision -> prd -> arch -> design -> plan -> dev -> qa -> release -> postdeploy -> ops` | новый сервис или крупная инициатива со сменой системного контура | сокращение этапов запрещено без явного owner-решения в audit trail |

### 2. Контракт next-step action matrix

| Поле действия | Требование |
|---|---|
| `action_kind` | Тип действия (`issue_stage_transition`, `pull_request_label_add`) |
| `target_label` | Целевой label |
| `display_variant` | Человекочитаемый тип действия (`revise`, `full_flow`, `reviewer` и т.д.) |
| `url` | Deep-link на `/` staff web-console с confirm-modal |

### 3. Preview / execute flow
- GitHub service-message публикует список typed действий.
- Deep-link всегда ведёт на `/` и открывает confirm-модалку.
- Frontend выполняет preview через staff API и показывает `removed_labels`, `added_labels`, `final_labels`.
- После подтверждения execute-path выполняет transition с аудитом.
- При неоднозначном текущем stage (`0` или `>1` labels из `run:*`) transition не выполняется, ставится `need:input`.

### 4. Актуальность контракта
- OpenAPI / proto фиксируют preview/execute endpoints и typed responses.
- Для Sprint S5 канонический next-step UX закреплён как matrix-based contract без raw fallback-команд в GitHub.

## Stories (handover в `run:dev`)
- Story-1: Profile Registry + escalation resolver.
- Story-2: Next-step Action Matrix + preview/execute contract.
- Story-3: Service-message Rendering with typed actions and display variants.
- Story-4: Governance Guardrails (`need:input`, ambiguity-stop, no silent-skip).
- Story-5: Traceability Sync (`issue_map`, `requirements_traceability`, sprint/epic docs).

## Декомпозиция реализации для `run:dev`

| Инкремент | Priority | Что реализовать | DoD инкремента |
|---|---|---|---|
| I1: Profile resolver core | P0 | Deterministic resolver `quick-fix/feature/new-service` + escalation rules | profile определяется однозначно; ambiguity приводит к `need:input` |
| I2: Next-step action matrix | P0 | Рендер `action_kind`, `target_label`, `display_variant`, `url` | service-comment в GitHub содержит полный список typed действий без raw-команд |
| I3: Preview / execute path | P0 | Preview diff labels + execute transition через staff API | transition не выполняет best-guess; при конфликте stage публикуется remediation |
| I4: Review-gate transitions | P1 | Унифицированный post-PR переход в `state:in-review` для Issue+PR | после формирования PR обе сущности получают `state:in-review` без дрейфа |
| I5: Traceability sync | P1 | Автообновление `issue_map`/`requirements_traceability` при переходах S5 | связь `issue -> требования -> AC` остаётся актуальной после каждого merge |

## Quality gates
- Planning gate:
  - FR-053/FR-054 отражены в продуктовых документах;
  - launch profile matrix согласована с `stage_process_model`.
- Contract gate:
  - next-step action matrix детерминирована и допускает только policy-safe transitions.
  - `docs/architecture/api_contract.md` фиксирует обязательные поля typed action и ownership resolver в `control-plane`.
- Architecture gate:
  - ownership resolver/escalation закреплён за `services/internal/control-plane`;
  - `external` и `staff` контуры остаются thin adapters без доменной логики переходов.
- UX gate:
  - confirm-модалка на `/` показывает preview diff перед execute;
  - transition выполняется детерминированно без best-guess.
- Security gate:
  - frontend не делает прямые GitHub mutations;
  - любой transition сохраняет audit trail.
- Traceability gate:
  - изменения отражены в `issue_map` и `requirements_traceability`.

## Acceptance scenarios (owner-ready)

| ID | Сценарий | Ожидаемый результат |
|---|---|---|
| AC-01 | Profile=`feature`, deep-link доступен | Owner открывает confirm-модалку, видит preview diff и выполняет transition; stage меняется по профилю и фиксируется в audit |
| AC-02 | Stage однозначен, available next-step action существует | Owner использует action-link и confirm-modal, процесс не блокируется |
| AC-03 | Stage ambiguity (`0` или `>1` trigger labels) | Transition блокируется, ставится `need:input`, публикуется remediation-message |
| AC-04 | `quick-fix` с признаком `cross-service impact` | Происходит обязательная эскалация профиля (`feature` или `new-service`) без silent-skip |
| AC-05 | После формирования PR требуется review gate | На PR и Issue устанавливается `state:in-review`; trigger label снимается |
| AC-06 | Security/policy проверка fallback | Команды содержат только `run:*|state:*|need:*` labels и не содержат секретов |

## Acceptance criteria
- [x] Подтвержден канонический набор launch profiles и правила эскалации.
- [x] Подтвержден формат next-step action matrix (`action_kind`, `target_label`, `display_variant`, `url`).
- [x] Подготовлены риски и продуктовые допущения для `run:dev`.
- [x] Получен Owner approval vision/prd пакета для запуска `run:dev`.

## Открытые риски
- Риск UX-перегрузки: слишком много действий в service-comment снижает читаемость.
- Риск policy-drift: profile shortcut может быть использован для обхода обязательных этапов.
- Риск интеграции: fallback-команды могут расходиться с фактической policy, если не централизовать шаблон.
- Риск операционной рассинхронизации: ручной `gh` fallback может быть выполнен с неверным `current-stage` без pre-check.

## Блокеры и owner decisions (run:plan handover)

| Тип | ID | Суть | Что требуется от Owner | Статус |
|---|---|---|---|---|
| blocker | BLK-155-01 | Финальное решение по запуску `run:dev` после review этого пакета | Запуск разрешён Owner (`Запуск разрешаю`, PR #166, 2026-02-25) | closed |
| blocker | BLK-155-02 | Требовалась фиксация единой политики fast-track `design -> dev` | Политика зафиксирована как OD-155-01 с guardrails и каноническим fallback | closed |
| owner-decision | OD-155-01 | Fast-track `design -> dev` как опциональный путь | Утверждено: fast-track допускается как дополнительный путь вместе с canonical `design -> plan`; обязательны `pre-check` и audit trail | approved |
| owner-decision | OD-155-02 | Ambiguity hard-stop (`0` или `>1` stage labels) | Утверждено: при неоднозначности обязателен `need:input`, best-guess переходы запрещены | approved |
| owner-decision | OD-155-03 | Единый review-gate для Issue + PR | Утверждено: обязательная синхронизация `state:in-review` на Issue и PR | approved |

## Зафиксированные Owner-решения (2026-02-25)
- `OD-155-01`: fast-track `design -> dev` остаётся optional action и публикуется только вместе с canonical `design -> plan`; ручной fallback выполняется строго как `pre-check -> transition`.
- `OD-155-02`: при ambiguity stage (`0` или `>1` labels `run:*`) transition блокируется, а service-comment публикует только remediation с `need:input`.
- `OD-155-03`: review-gate после формирования PR обязан устанавливать `state:in-review` на PR и Issue одновременно.

## Продуктовые допущения
- Launch profiles не заменяют канонический stage-pipeline, а задают управляемые “траектории входа”.
- Primary канал (`web-console`) может временно быть недоступен; fallback обязан быть достаточным для продолжения процесса.
- Для P0/P1 инициатив Owner может принудительно переводить задачу в `new-service` траекторию независимо от исходного профиля.

## EM execution control loop для `run:dev`

| Этап контроля | Контрольная точка | Артефакт подтверждения |
|---|---|---|
| Pre-start | QG-01..QG-05 и Readiness gate пройдены, блокеры задокументированы | `sprint_s5_stage_entry_and_label_ux.md`, `epic_s5.md` |
| Mid-execution | I1..I3 закрывают deterministic resolver и fallback contract без policy bypass | PR updates + acceptance evidence AC-01..AC-04 |
| Exit to QA | I4..I5 закрывают review-gate и traceability sync | `issue_map`, `requirements_traceability`, PR checklist |

## Readiness gate для `run:dev`
- [x] FR-053 и FR-054 синхронизированы между product policy и delivery docs.
- [x] Profile matrix, escalation rules и ambiguity handling формализованы.
- [x] Next-step action-card contract и fallback templates зафиксированы.
- [x] Архитектурный контракт и runtime impact/migration path зафиксированы в ADR-0008.
- [x] Канонический API-contract (`docs/architecture/api_contract.md`) синхронизирован с FR-053/FR-054.
- [x] Риски и продуктовые допущения отражены в Day1 epic.
- [x] Owner review/approve в этом PR.
