---
doc_id: INTAKE-CK8S-0123
type: intake-brief
title: "Issue #123 — Intake/Revise E2E Smoke"
status: draft
owner_role: PM
created_at: 2026-02-24
updated_at: 2026-02-24
related_issues: [123]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-24-issue-123-intake-smoke"
---

# Intake Artifact: Issue #123

## Контекст
- Issue: `#123`.
- Цель smoke-прогона: подтвердить, что контур `run:intake` и `run:intake:revise` корректно отрабатывает минимальный PM-артефакт и переводит задачу в review-gate без выхода за markdown-only policy.
- Ограничение прогона: изменения только в `*.md`.

## Scope (in)
- Формализация минимальных требований для intake smoke.
- Явные acceptance criteria для первичного прогона и revise-итерации.
- Обновление трассируемости `issue -> артефакт -> критерии готовности`.

## Scope (out)
- Изменения кода, YAML/JSON, Dockerfile, скриптов.
- Изменения stage-политик и архитектурных стандартов.
- Расширение функционального покрытия beyond intake smoke.

## Product Requirements (минимум для smoke)
- RQ-123-01: На `run:intake` должен формироваться PM-артефакт с контекстом, scope, AC и рисками.
- RQ-123-02: Артефакт должен быть трассируем до Issue #123 в delivery-документации.
- RQ-123-03: После создания/обновления PR должен применяться transition `run:intake -> state:in-review` для Issue и PR.
- RQ-123-04: На `run:intake:revise` должна поддерживаться точечная доработка артефакта без нарушения markdown-only policy.

## Acceptance Criteria
1. В репозитории есть intake-артефакт для Issue #123 с явными разделами: контекст, scope, требования, AC, риски, допущения.
2. `docs/delivery/issue_map.md` содержит строку для Issue #123 с корректными ссылками на требования и intake-артефакт.
3. `docs/delivery/e2e_mvp_master_plan.md` содержит отдельный smoke-сценарий `run:intake`/`run:intake:revise`.
4. Изменения ограничены markdown-файлами.
5. Для PR применён review-gate: `state:in-review` на PR и Issue #123 после обновления PR.
6. Для `run:intake:revise` каждое открытое review-замечание получает явный результат: `fix_required` или `not_applicable` с фактологическим обоснованием.

## Definition of Done
- AC-1..AC-6 выполнены.
- В PR описана трассируемость `Issue #123 -> RQ-123-01..04 -> Acceptance Criteria`.
- В PR перечислены открытые риски и продуктовые допущения.

## Revise Resolution Log (PR #124, 2026-02-24)

| Источник замечания | Приоритет | Статус | Решение |
|---|---|---|---|
| Review `CHANGES_REQUESTED` от `ai-da-stas` (`E2E revise check for intake: please address requested changes.`) | `high` | `fix_required` | Уточнены критерии intake-revise (AC-6), добавлена явная фиксация обработки review-замечаний в этом артефакте и синхронизирована delivery-трассируемость для Issue #123. |

Пояснение:
- Inline review-треды и незакрытые line-comments в PR #124 отсутствуют (проверено через `gh api graphql` для `reviewThreads` и `gh api pulls/124/comments`).
- Единственное открытое замечание было общим review-level запросом на доработку; эта итерация закрывает его содержательно через усиление критериев и evidence.

## Открытые риски
- RISK-123-01: smoke не покрывает переходы следующих стадий (`vision+`) и может скрыть проблемы stage chaining.
- RISK-123-02: при ручном ревью возможна задержка между обновлением PR и label transition.

## Продуктовые допущения
- ASM-123-01: для smoke достаточно минимального PM-артефакта без добавления новых FR/NFR в baseline.
- ASM-123-02: success-критерий smoke — корректный review-gate и трассируемость, а не полнота бизнес-проработки.
