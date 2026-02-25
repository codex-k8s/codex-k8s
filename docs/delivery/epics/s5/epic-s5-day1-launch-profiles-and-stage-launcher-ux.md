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
  status: pending
  request_id: "owner-2026-02-24-issue-155-day1-vision-prd"
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

### 1. Канонический набор launch profiles и эскалации

| Profile | Обязательная траектория | Когда применять | Детерминированные триггеры эскалации |
|---|---|---|---|
| `quick-fix` | `intake -> plan -> dev -> qa -> release -> postdeploy -> ops` | точечная правка в одном сервисе без изменения контрактов/схемы | любой из триггеров: `cross-service impact`, новая интеграция, миграция БД, изменение RBAC/policy |
| `feature` | `intake -> prd -> design -> plan -> dev -> qa -> release -> postdeploy -> ops` | функциональное изменение в существующих сервисах | при изменении архитектурных границ/NFR добавить `arch`; при изменении продуктовой стратегии добавить `vision` |
| `new-service` | `intake -> vision -> prd -> arch -> design -> plan -> dev -> qa -> release -> postdeploy -> ops` | новый сервис или крупная инициатива со сменой системного контура | сокращение этапов запрещено без явного owner-решения в audit trail |

### 2. Контракт next-step action cards

| Поле карточки | Требование |
|---|---|
| `launch_profile` | Обязателен (`quick-fix`, `feature`, `new-service`) |
| `stage_path` | Краткая строка траектории текущего профиля |
| `primary_action` | Валидированный deep-link в staff web-console |
| `fallback_action` | Копируемая команда label transition |
| `guardrail_note` | Явное правило по ambiguity и policy-gate |

### 3. Fallback command templates
- Pre-check перед ручным transition (обязателен):
  - `gh issue view <ISSUE_NUMBER> --json labels --jq '.labels[].name'` (оператор подтверждает, что `run:<current-stage>` единственный stage-trigger).
- Переход на следующий stage:
  - `gh issue edit <ISSUE_NUMBER> --remove-label "run:<current-stage>" --add-label "run:<next-stage>"`.
- Постановка explicit input-gate при неоднозначности:
  - `gh issue edit <ISSUE_NUMBER> --add-label "need:input"`.
- Переход в review gate после формирования PR:
  - `gh issue edit <ISSUE_NUMBER> --add-label "state:in-review"`;
  - `gh pr edit <PR_NUMBER> --add-label "state:in-review"`.
- Безопасность:
  - команды не содержат секретов/токенов;
  - команды используют только labels из каталога `run:*|state:*|need:*`.
  - при неоднозначном текущем stage (`0` или `>1` labels из `run:*`) transition не выполняется, ставится `need:input`.

### 4. Актуальность fallback-синтаксиса (`gh`)
- Синтаксис `gh issue edit` / `gh pr edit` для `--add-label` и `--remove-label` сверен с актуальным manual GitHub CLI (Context7, источник `/websites/cli_github_manual`).
- Для Sprint S5 принимаем эти флаги как канонический fallback-контракт до отдельного owner-решения.

## Stories (handover в `run:dev`)
- Story-1: Profile Registry + escalation resolver.
- Story-2: Next-step Action Contract + fallback template builder.
- Story-3: Service-message Rendering with profile path and guardrail note.
- Story-4: Governance Guardrails (`need:input`, ambiguity-stop, no silent-skip).
- Story-5: Traceability Sync (`issue_map`, `requirements_traceability`, sprint/epic docs).

## Декомпозиция реализации для `run:dev`

| Инкремент | Priority | Что реализовать | DoD инкремента |
|---|---|---|---|
| I1: Profile resolver core | P0 | Deterministic resolver `quick-fix/feature/new-service` + escalation rules | profile определяется однозначно; ambiguity приводит к `need:input` |
| I2: Next-step card contract | P0 | Рендер `launch_profile`, `stage_path`, `primary_action`, `fallback_action`, `guardrail_note` | service-comment в GitHub содержит полный контракт без ручных дописок |
| I3: Fallback pre-check + transition | P0 | Генератор fallback-команд: pre-check labels + transition | fallback не выполняет best-guess; при конфликте stage публикуется remediation |
| I4: Review-gate transitions | P1 | Унифицированный post-PR переход в `state:in-review` для Issue+PR | после формирования PR обе сущности получают `state:in-review` без дрейфа |
| I5: Traceability sync | P1 | Автообновление `issue_map`/`requirements_traceability` при переходах S5 | связь `issue -> требования -> AC` остаётся актуальной после каждого merge |

## Quality gates
- Planning gate:
  - FR-053/FR-054 отражены в продуктовых документах;
  - launch profile matrix согласована с `stage_process_model`.
- Contract gate:
  - next-step action contract детерминирован и допускает только policy-safe transitions.
- Architecture gate:
  - ownership resolver/escalation закреплён за `services/internal/control-plane`;
  - `external` и `staff` контуры остаются thin adapters без доменной логики переходов.
- UX gate:
  - сценарий broken/dead link не блокирует owner-flow;
  - fallback-путь выполняется детерминированно (`pre-check -> transition`) без best-guess.
- Security gate:
  - fallback-команды не содержат секретов;
  - любой transition сохраняет audit trail.
- Traceability gate:
  - изменения отражены в `issue_map` и `requirements_traceability`.

## Acceptance scenarios (owner-ready)

| ID | Сценарий | Ожидаемый результат |
|---|---|---|
| AC-01 | Profile=`feature`, deep-link доступен | Owner выполняет переход через `primary_action`; stage меняется по профилю и фиксируется в audit |
| AC-02 | Deep-link недоступен (404/timeout), stage однозначен | Owner использует fallback (`pre-check -> transition`), процесс не блокируется |
| AC-03 | Stage ambiguity (`0` или `>1` trigger labels) | Transition блокируется, ставится `need:input`, публикуется remediation-message |
| AC-04 | `quick-fix` с признаком `cross-service impact` | Происходит обязательная эскалация профиля (`feature` или `new-service`) без silent-skip |
| AC-05 | После формирования PR требуется review gate | На PR и Issue устанавливается `state:in-review`; trigger label снимается |
| AC-06 | Security/policy проверка fallback | Команды содержат только `run:*|state:*|need:*` labels и не содержат секретов |

## Acceptance criteria
- [x] Подтвержден канонический набор launch profiles и правила эскалации.
- [x] Подтвержден формат next-step action-карт (`primary deep-link + fallback command`).
- [x] Подготовлены риски и продуктовые допущения для `run:dev`.
- [ ] Получен Owner approval vision/prd пакета для запуска `run:dev`.

## Открытые риски
- Риск UX-перегрузки: слишком много действий в service-comment снижает читаемость.
- Риск policy-drift: profile shortcut может быть использован для обхода обязательных этапов.
- Риск интеграции: fallback-команды могут расходиться с фактической policy, если не централизовать шаблон.
- Риск операционной рассинхронизации: ручной `gh` fallback может быть выполнен с неверным `current-stage` без pre-check.

## Продуктовые допущения
- Launch profiles не заменяют канонический stage-pipeline, а задают управляемые “траектории входа”.
- Primary канал (`web-console`) может временно быть недоступен; fallback обязан быть достаточным для продолжения процесса.
- Для P0/P1 инициатив Owner может принудительно переводить задачу в `new-service` траекторию независимо от исходного профиля.

## Readiness gate для `run:dev`
- [x] FR-053 и FR-054 синхронизированы между product policy и delivery docs.
- [x] Profile matrix, escalation rules и ambiguity handling формализованы.
- [x] Next-step action-card contract и fallback templates зафиксированы.
- [x] Архитектурный контракт и runtime impact/migration path зафиксированы в ADR-0008.
- [x] Риски и продуктовые допущения отражены в Day1 epic.
- [ ] Owner review/approve в этом PR.
