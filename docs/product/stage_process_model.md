---
doc_id: STG-CK8S-0001
type: process-model
title: "codex-k8s — Stage Process Model"
status: draft
owner_role: EM
created_at: 2026-02-11
updated_at: 2026-02-13
related_issues: [1, 19]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
---

# Stage Process Model

## TL;DR
- Целевая модель: `intake -> vision -> prd -> arch -> design -> plan -> dev -> qa -> release -> postdeploy -> ops`.
- Для каждого этапа есть `run:*` и `run:*:revise` петля.
- Переход между этапами требует формального подтверждения артефактов и фиксируется в audit.
- Дополнительный служебный цикл `run:self-improve` работает поверх stage-контура и улучшает docs/prompts/tools по итогам запусков.

## Source of truth
- `docs/product/labels_and_trigger_policy.md`
- `docs/product/agents_operating_model.md`
- `docs/delivery/development_process_requirements.md`

## Этапы и обязательные артефакты

| Stage | Trigger labels | Основные артефакты | Основные роли |
|---|---|---|---|
| Intake | `run:intake`, `run:intake:revise` | problem, personas, scope, constraints, brief, traceability bundle | `pm`, `km` |
| Vision | `run:vision`, `run:vision:revise` | charter, success metrics, risk register | `pm`, `em` |
| PRD | `run:prd`, `run:prd:revise` | PRD, acceptance criteria, NFR draft | `pm`, `sa` |
| Architecture | `run:arch`, `run:arch:revise` | C4, ADR backlog/ADR, alternatives | `sa` |
| Design | `run:design`, `run:design:revise` | design doc, API contract, data model, migration policy | `sa`, `qa` |
| Plan | `run:plan`, `run:plan:revise` | delivery plan, epics/stories, DoD | `em`, `km` |
| Development | `run:dev`, `run:dev:revise` | code changes, PR, docs updates | `dev`, `reviewer` |
| QA | `run:qa` | test strategy/plan/matrix, regression result | `qa` |
| Release | `run:release` | release plan/notes, rollback plan | `em`, `sre` |
| Postdeploy | `run:postdeploy` | postdeploy review, postmortem | `qa`, `sre` |
| Ops | `run:ops` | SLO/alerts/runbook improvements | `sre`, `km` |
| Self-Improve | `run:self-improve` | run/session diagnosis (MCP), change-set PR, policy/tooling recommendations | `km`, `dev`, `reviewer` |

## Петли ревизии и переосмысления

- На каждом этапе доступны:
  - `run:<stage>:revise` для доработки артефактов;
  - `run:rethink` для возврата на более ранний этап.
- После `run:rethink` предыдущие версии артефактов маркируются как `state:superseded`.

## Вход/выход этапа

Общие правила входа:
- есть обязательные входные артефакты предыдущего этапа;
- нет блокеров `state:blocked`;
- отсутствует незакрытый `need:input`.

Общие правила выхода:
- артефакты этапа обновлены и связаны с Issue/PR в traceability документах (`issue_map`, sprint/epic docs);
- статус этапа отражён через `state:*` лейблы;
- события перехода записаны в аудит.

### Правило review gate для всех этапов
- Для всех `run:*` выход этапа проходит через review gate перед финальным review Owner:
  - pre-review от `reviewer` (для технических артефактов) и/или профильной роли через `need:*`;
  - финальное решение Owner по принятию артефактов.
- Постановка `state:in-review` выполняется так:
  - на PR и на Issue, если run завершился артефактами в PR;
  - только на Issue, если run завершился без PR.

## Паузы и таймауты в stage execution

- Разрешены paused состояния:
  - `waiting_owner_review`;
  - `waiting_mcp`.
- Для `waiting_mcp` timeout-kill не применяется до завершения ожидания.
- Для длительных пауз run должен оставаться resumable за счёт сохранения `codex-cli` session snapshot.

## Текущий активный контур (S3 Day1)

На текущем этапе реализации активирован полный trigger-контур:
- `run:intake..run:ops`;
- `run:<stage>:revise`;
- `run:rethink`, `run:self-improve`.

Ограничение текущего этапа:
- для части стадий пока активирован базовый orchestration path (routing/audit/policy),
  а специализированная бизнес-логика стадий дорабатывается следующими S3 эпиками.
- для prompt-body используется минимальная stage-matrix seed-шаблонов в `docs/product/prompt-seeds/`
  (по схеме `<stage>-work.md` и `<stage>-review.md` для revise-loop стадий).

## План активации контуров

- S2 baseline: `run:dev` и `run:dev:revise` (completed).
- S2 Day6: approval/audit hardening (completed).
- S2 Day7: regression gate под полный MVP (completed).
- S3 Day1: активация полного stage-flow (`run:intake..run:ops`) и trigger path для `run:self-improve` (completed).
- S3 Day2+ : поэтапное насыщение stage-specific логики и observability.

## Конфигурационные labels для исполнения stage

- Помимо trigger/status labels используются конфигурационные labels:
  - `[ai-model-*]` — выбор модели;
  - `[ai-reasoning-*]` — выбор уровня рассуждений.
- Эти labels не запускают stage сами по себе, но влияют на effective runtime profile.
- Для `run:dev:revise` профиль model/reasoning перечитывается перед каждым запуском.
