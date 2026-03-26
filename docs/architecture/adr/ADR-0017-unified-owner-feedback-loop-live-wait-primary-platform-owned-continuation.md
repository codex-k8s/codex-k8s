---
doc_id: ADR-0017
type: adr
title: "Unified owner feedback loop: live wait primary with platform-owned continuation truth"
status: proposed
owner_role: SA
created_at: 2026-03-26
updated_at: 2026-03-26
related_issues: [541, 554, 557, 559, 568]
related_prs: []
supersedes: []
superseded_by: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-26-issue-559-arch"
---

# ADR-0017: Unified owner feedback loop — live wait primary with platform-owned continuation truth

## TL;DR
- Контекст: Sprint S17 требует unified owner feedback loop, где owner отвечает из Telegram или staff-console, а primary happy-path остаётся same live pod / same `codex` session.
- Решение: выбираем live wait primary model, в которой `control-plane` владеет request truth и continuation policy, `worker` удерживает dispatch/reconcile/lease side effects, `agent-runner` удерживает только live session и recovery snapshot, а channel surfaces остаются thin adapters.
- Последствия: сохраняются same-session trust, единый persisted source of truth и channel parity, но design stage обязан детализировать typed contracts, data model и rollout/migration notes.

## Контекст
- Проблема:
  - если detached resume-run нормализовать как основной path, Sprint S17 потеряет locked baseline same-session continuation и long-lived wait перестанет быть реальным live contract;
  - если Telegram или staff-console получат собственный lifecycle owner, dual-channel inbox деградирует в split-brain и потеряет deterministic binding/correlation;
  - если выделить новый dedicated service уже на Day4, появится premature DB owner и лишний service boundary до фиксации design contracts.
- Ограничения:
  - `run:arch` остаётся markdown-only;
  - max timeout/TTL built-in `codex_k8s` MCP wait path не ниже owner wait window;
  - snapshot-resume только recovery fallback;
  - `api-gateway` и `staff web-console` должны оставаться thin surfaces;
  - `run:self-improve` не входит в owner-facing wait contract.
- Связанные требования:
  - `docs/delivery/epics/s17/prd-s17-day3-unified-user-interaction-waits-and-owner-feedback-inbox.md`
  - `FR-012`, `FR-039`, `FR-040`, `FR-041`, `FR-043`
  - `NFR-012`, `NFR-013`, `NFR-015`, `NFR-016`, `NFR-018`
- Что ломается без решения:
  - owner не сможет доверять тому, что агент действительно ждёт его ответ в той же задаче;
  - channel parity станет недоказуемой;
  - design stage будет переоткрывать service ownership вместо детализации contracts/data.

## Decision Drivers (что важно)
- Same-session continuation как primary happy-path.
- Единый persisted owner feedback truth для Telegram и staff-console.
- Deterministic binding/correlation для text/voice/callback replies.
- Явная visibility model для overdue / expired / manual-fallback states.
- Отсутствие premature service split и сохранение текущих bounded contexts.

## Рассмотренные варианты
### Вариант A: Detached resume-run и channel-owned inbox projections как нормальный path
- Плюсы:
  - меньше стоимость live runtime retention;
  - проще initial implementation для reply handling.
- Минусы:
  - ломает Day1-Day3 baseline same-session happy-path;
  - делает tool timeout/TTL фактическим driver UX вместо owner wait window.
- Риски:
  - скрытая normalisation synthetic resume;
  - split-brain между каналами и run/session truth;
  - падение доверия owner к pending inbox.
- Стоимость внедрения:
  - низкая на старте, высокая при исправлении semantic drift.
- Эксплуатация:
  - сложно объяснить, почему continuation относится к той же задаче, а не к новому detached run.

### Вариант B (выбран): Live wait primary, platform-owned truth, thin channel surfaces
- Плюсы:
  - удерживает same live pod / same `codex` session как primary path;
  - даёт один semantic owner для lifecycle, deadlines, parity и degraded classification;
  - сохраняет Telegram и staff-console как replaceable surfaces поверх одного persisted contract.
- Минусы:
  - design stage должен аккуратно оформить typed contracts и runtime retention rules;
  - `control-plane` получает дополнительную доменную ответственность.
- Риски:
  - нужен аккуратный mixed-version rollout для long-lived waits и recovery linkage.
- Стоимость внедрения:
  - средняя.
- Эксплуатация:
  - предсказуемая при условии единого audit/correlation contour и lease keepalive.

### Вариант C: Dedicated owner-feedback coordinator service уже на Day4
- Плюсы:
  - отдельный bounded context;
  - потенциально явный future scaling path.
- Минусы:
  - новый DB owner и новый rollout contour до design stage;
  - выше coordination cost между `control-plane`, `worker`, `agent-runner` и новым сервисом.
- Риски:
  - задержка `run:design` и `run:plan`;
  - premature lock-in в topology вместо проверки contracts.
- Стоимость внедрения:
  - высокая.
- Эксплуатация:
  - ещё один service/runtime contour без доказанной необходимости на MVP.

## Решение
Мы выбираем: **вариант B — live wait primary with platform-owned continuation truth and thin channel surfaces**.

## Обоснование (Rationale)
- Вариант B лучше всего сохраняет locked baseline Sprint S17 и уже утверждённые Sprint S10/S11 service boundaries.
- `control-plane` уже владеет built-in MCP semantics, run/session policy и audit lifecycle, поэтому естественно становится owner request truth, accepted-response winner и degraded classification.
- `worker` уже отвечает за async delivery/reconcile и namespace lease keepalive, поэтому long-lived wait retention и overdue/expired/manual-fallback detection остаются в правильном background contour.
- `agent-runner` остаётся владельцем только live execution and recovery snapshot mechanics, а не persisted request truth.
- Telegram и staff-console не получают права переопределять platform semantics, что удерживает channel neutrality и trust.

## Последствия (Consequences)
### Позитивные
- Same-session continuation остаётся проверяемым primary happy-path.
- Один persisted backend truth обеспечивает parity Telegram и staff-console.
- Recovery-only snapshot-resume, duplicate/stale/expired handling и degraded visibility получают один semantic owner.

### Негативные / компромиссы
- Design stage обязан отдельно зафиксировать typed contracts, wait-state linkage и rollout notes для long-lived waits.
- `control-plane` расширяется дополнительной domain responsibility вокруг feedback request truth.

### Технический долг
- Что откладываем:
  - отдельный owner-feedback coordinator service;
  - additional channels, reminders/escalations, attachments и generalized conversation UX;
  - richer policy automation beyond canonical degraded-state visibility.
- Когда вернуться:
  - после `run:design` / `run:plan`, если появятся доказанные scale, coupling или operability signals.

## План внедрения (минимально)
- Изменения в коде:
  - отсутствуют на `run:arch`.
- Изменения в инфраструктуре:
  - отсутствуют.
- Миграции/совместимость:
  - design stage должен определить schema ownership, mixed-version rollout и rollback path для request truth, projections и recovery linkage.
- Наблюдаемость:
  - design stage должен зафиксировать event set для `delivery_pending`, `delivery_accepted`, `waiting`, `response_bound`, `continuation_live`, `recovery_resume`, `overdue`, `expired`, `manual_fallback`.

## План отката/замены
- Условия отката:
  - если на `run:design` выяснится, что live wait budget нельзя удержать без unacceptable coupling, cost или operability regressions.
- Как откатываем:
  - ADR переводится в `superseded`, а инициатива переходит либо к dedicated service, либо к другому continuation model через новый owner decision.

## Ссылки
- PRD:
  - `docs/delivery/epics/s17/prd-s17-day3-unified-user-interaction-waits-and-owner-feedback-inbox.md`
- Architecture:
  - `docs/architecture/initiatives/s17_unified_owner_feedback_loop/architecture.md`
- Alternatives:
  - `docs/architecture/alternatives/ALT-0009-unified-owner-feedback-loop-live-wait-and-channel-ownership.md`
- Related baseline:
  - `docs/architecture/initiatives/s10_mcp_user_interactions/architecture.md`
  - `docs/architecture/initiatives/s11_telegram_user_interaction_adapter/architecture.md`
  - `docs/architecture/agent_runtime_rbac.md`
  - `docs/architecture/mcp_approval_and_audit_flow.md`
