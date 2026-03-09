---
doc_id: ADR-0008
type: adr
title: "Profile-driven stage launch and deterministic next-step contract"
status: accepted
owner_role: SA
created_at: 2026-02-25
updated_at: 2026-03-09
related_issues: [154, 155]
related_prs: []
supersedes: []
superseded_by: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-25-issue-155-arch-adr"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-25
---

# ADR-0008: Profile-driven stage launch and deterministic next-step contract

## TL;DR
- Проблема: переходы по этапам не должны зависеть от устаревающего raw-контракта в GitHub-комментарии и ручного выбора следующего `run:*` label.
- Решение: сохранить profile-driven resolver внутри `control-plane`, а наружу публиковать typed next-step action matrix с preview/execute через staff API и confirm-modal на `/`.
- Архитектурный принцип: бизнес-решения о профиле, эскалации и ambiguity принимаются только внутри `control-plane` домена; edge и UI остаются thin adapters.

## Контекст

Issue #154 зафиксировал intake baseline для profile-driven UX, а Issue #155 оформил vision/PRD пакет для FR-053/FR-054. Перед входом в `run:dev` требуется единое архитектурное решение, которое:
- сохраняет текущие сервисные границы;
- формализует transport/data контракты next-step карточек;
- задаёт детерминированные правила перехода и эскалации без обхода policy.

Ограничения:
- архитектурные границы `external -> internal -> jobs` не меняются;
- прямые GitHub label-команды не должны быть частью публичного service-message контракта;
- при ambiguity запрещён best-guess transition.

## Decision drivers

- Детерминированность: одинаковый контекст должен давать одинаковый next-step outcome.
- Непрерывность owner-flow: переход не должен блокироваться при недоступности web-console.
- Governance safety: policy и audit должны быть едиными для primary/fallback путей.
- Тонкие границы: доменная логика не переносится в edge/frontend.
- Трассируемость: FR-053/FR-054 должны быть связаны с архитектурным артефактом, а не только с product/delivery docs.

## Рассмотренные варианты

### Вариант A: только deep-link переходы через web-console

Плюсы:
- один операционный канал;
- минимальная реализация на стороне service-message.

Минусы:
- owner-flow блокируется при 404/timeout/degraded UI;
- нет безопасного fallback path для FR-054.

### Вариант B: dual-path (`deep-link + raw fallback command`)

Плюсы:
- отказоустойчивость при degraded UI.

Минусы:
- повышенный риск drift между UI, comment payload и реальной stage-политикой;
- ручные label-команды размывают единый audit/UX контур.

### Вариант C (выбран): matrix-based contract с preview/execute

Суть:
- service-message публикует список typed действий;
- deep-link всегда ведёт на `/` staff web-console и открывает confirm-modal;
- preview/execute выполняются через staff API, а решение о профиле/эскалации/ambiguity централизовано в `control-plane`.

Плюсы:
- выполняет FR-053/FR-054 одновременно;
- устраняет drift между comment payload и реальным transition path;
- сохраняет единый governance и audit.

Минусы:
- выше сложность матрицы действий и тестовых сценариев;
- нужен строгий typed preview/execute contract.

## Решение

Выбран **Вариант C** с следующими архитектурными контрактами.

### 1. Границы сервисов и ответственность

- `services/external/api-gateway`:
  - только transport validation/auth/routing;
  - без profile resolver и без stage transition business rules.
- `services/internal/control-plane`:
  - единственный источник истины для `launch_profile`, `stage_path`, escalation и ambiguity rules;
  - формирует typed payload next-step action matrix;
  - применяет policy-safe transitions и пишет audit events.
- `services/jobs/worker`:
  - публикует/обновляет service-message на основе результата `control-plane`;
  - выполняет idempotent orchestration без собственной бизнес-логики переходов.
- `services/staff/web-console`:
  - исполняет только UX/confirm-path;
  - transition выполняет через staff API control-plane, а не напрямую через GitHub label mutation.

### 2. Канонический next-step контракт

Service-message публикует список typed действий, где каждое действие содержит:
- `action_kind` (`issue_stage_transition`, `pull_request_label_add`);
- `target_label`;
- `display_variant`;
- `url` (deep-link на `/` с query для confirm-modal).

Обязательное поведение:
- raw resolver state (`launch_profile`, `stage_path`, guardrails) остаётся внутренней логикой `control-plane`;
- при ambiguity публикуется только remediation (`need:input`), без transition action;
- финальное изменение лейблов выполняется только через preview/execute staff API.

### 3. Resolver и эскалации

Resolver в `control-plane` определяет профиль и следующий этап детерминированно.

Базовые правила:
- `quick-fix -> feature -> new-service` только в сторону эскалации;
- обратный переход профиля допускается только через явное owner-решение с аудитом;
- risk signals (`cross-service impact`, миграции, RBAC/policy изменения, новая интеграция) инициируют обязательную эскалацию.

### 4. Governance и аудит

Все действия из матрицы обязаны сходиться в общем audit path:
- `run.profile.resolved`;
- `run.stage.escalated`;
- `run.next_step.matrix_rendered`;
- `run.next_step.preview_requested`;
- `run.next_step.executed`;
- `run.next_step.blocked_need_input`.

Любая transition-операция должна сохранять `correlation_id`, actor и source (`ui`) в `flow_events`.

## Runtime impact и миграция

Изменения runtime-контуров:
- required: расширение service-message payload списком typed next-step действий;
- required: новая доменная логика resolver/escalation в `control-plane`;
- required: preview/execute endpoints и confirm-modal на `/`;
- optional (phase 2): дополнительные агрегированные метрики использования next-step matrix.

Миграционная стратегия для `run:dev`:
1. Внедрить resolver и контракт DTO для next-step карточек.
2. Подключить preview/execute contract и confirm-modal.
3. Добавить ambiguity-gate (`need:input`) и блокировку best-guess.
4. Синхронизировать review-gate transition (`state:in-review` на Issue+PR).
5. Зафиксировать regression-сценарии AC-01..AC-06.

Откат:
- feature-flag отключает расширенную матрицу действий и возвращает stage-aware baseline из ADR-0006;
- audit события сохраняются для postmortem и повторного rollout.

## Последствия

### Позитивные
- Owner получает отказоустойчивый transition path без потери governance.
- Архитектурные границы остаются чистыми: business rules централизованы в домене.
- FR-053/FR-054 получают прямое архитектурное покрытие.

### Негативные/компромиссы
- Повышается сложность тестирования и контракта service-message.
- Требуется дисциплина версионирования fallback шаблонов, чтобы избежать drift.

## Связанные документы

- `docs/product/requirements_machine_driven.md` (FR-053, FR-054)
- `docs/product/labels_and_trigger_policy.md`
- `docs/product/stage_process_model.md`
- `docs/architecture/adr/ADR-0006-review-driven-revise-and-next-step-ux.md`
- `docs/delivery/epics/s5/epic-s5-day1-launch-profiles-and-stage-launcher-ux.md`
- `docs/delivery/epics/s5/prd-s5-day1-launch-profiles-and-stage-launcher-ux.md`
