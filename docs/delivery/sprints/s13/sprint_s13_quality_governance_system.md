---
doc_id: SPR-CK8S-0013
type: sprint-plan
title: "Sprint S13: Quality governance system для agent-scale delivery (Issue #469)"
status: in-review
owner_role: PM
created_at: 2026-03-14
updated_at: 2026-03-15
related_issues: [469, 471, 476, 484]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-14-issue-469-intake"
---

# Sprint S13: Quality governance system для agent-scale delivery (Issue #469)

## TL;DR
- Sprint S13 открывает отдельную cross-cutting инициативу `Quality Governance System`: качество агентной поставки должно определяться измеримыми свойствами изменения и обязательным evidence, а не субъективной «внимательностью ревью».
- Intake stage в Issue `#469` зафиксировал baseline quality stack: draft quality metrics, risk tiers `low / medium / high / critical`, список high/critical changes, evidence taxonomy, verification minimum и review contract.
- Vision stage в Issue `#471` зафиксировал mission, quality north star, persona outcomes, KPI/guardrails и proportional governance baseline для owner/reviewer, delivery roles и platform operator.
- PRD stage в Issue `#476` перевёл vision baseline в user stories, FR/AC/NFR, edge cases, expected evidence и explicit contract для risk/evidence/waiver decisions; создана follow-up issue `#484` для `run:arch`.
- Sprint S13 не выбирает implementation-first решения по rollout controller, cockpit UI или runtime automation: этот runtime/UI слой выделен в отдельный Sprint S14 (Issue `#470`) и должен наследовать governance-baseline, а не переоткрывать его.
- После PRD создана follow-up issue `#484` на stage `run:arch`; trigger-лейбл следующего этапа остаётся owner-managed.
- На `2026-03-15` через Context7 (`/websites/cli_github_manual`) повторно подтверждён актуальный non-interactive GitHub CLI flow для `gh issue create`, `gh pr create` и `gh pr edit`, чтобы continuity issue и PR-flow не расходились с текущим automation-путём.

## Scope спринта
### In scope
- Полная doc-stage цепочка `intake -> vision -> prd -> arch -> design -> plan` для инициативы `Quality Governance System`.
- Формализация quality governance baseline:
  - quality north star и supporting metrics;
  - risk taxonomy и признаки high/critical change;
  - evidence taxonomy;
  - verification minimum;
  - review contract;
  - draft mapping `risk tier -> mandatory stages/gates -> required evidence`.
- Явная продуктовая граница между governance-baseline Sprint S13 и downstream runtime/UI stream Sprint S14.
- Создание последовательных follow-up issue без автоматической постановки `run:*`-лейблов.

### Out of scope
- Кодовая реализация до завершения и утверждения `run:plan`.
- Выбор конкретного rollout controller, canary engine, observability stack implementation или Mission Control UI mechanics.
- Попытка решить quality governance только через manual review без системного evidence/verification контура.
- Унификация всех будущих runtime safety контуров внутри Sprint S13 без отдельной инициативы Sprint S14.

## Рекомендованный launch profile
- Базовый launch profile: `new-service`.
- Обоснование:
  - инициатива меняет operating model поставки и затрагивает несколько ролей и stage-gates;
  - сокращённые траектории нельзя считать безопасными до фиксации proportional risk policy;
  - `vision` и `arch` обязательны, потому что Sprint S13 становится source-of-truth для downstream release-safety и quality-surface streams.
- Целевая continuity-цепочка:
  `#469 (intake) -> #471 (vision) -> #476 (prd) -> arch -> design -> plan -> dev`.

## Governance baseline, зафиксированный на intake

### Quality metrics baseline
- `Lead time for change` как ориентир скорости безопасной поставки.
- `Change failure rate` как индикатор цены изменения.
- `Mean time to restore` / recovery time как индикатор зрелости recovery loop.
- `Evidence completeness rate` как доля изменений, прошедших через обязательный evidence package без пробелов.
- `Stage gate latency` как вспомогательный guardrail, чтобы governance не деградировала в бюрократический bottleneck.

### Draft risk taxonomy
| Tier | Смысл | Типичные признаки |
|---|---|---|
| `low` | Локальное изменение без расширения blast radius | markdown-only правки, локальный bug-fix без контракта/данных, безопасные refactor-only изменения |
| `medium` | Изменение существующего поведения с ограниченным blast radius | фича в одном bounded context, локальные UI/API изменения без schema/security/deploy impact |
| `high` | Изменение с заметным blast radius и требованием усиленного governance | migrations, cross-service contracts, webhook/callback security, authn/authz/RBAC, deploy/runtime policy |
| `critical` | Изменение, которое может повлиять на системную безопасность или на широкий production contour | destructive schema/data ops, secret/token/credential handling, release/rollback mechanics, platform-wide policy or infra ownership changes |

### Изменения, которые по умолчанию относятся к `high/critical`
- Любые миграции БД, schema ownership changes и data backfill/cleanup с production impact.
- Изменения в `authn/authz`, RBAC, approval policy, secret/token handling и webhook/callback security.
- Изменения release/deploy/runtime orchestration, rollback policy и build pipeline.
- Cross-service transport/data contracts и external-provider integrations, влияющие на state, quota или billing-like контуры.
- Любые операции с высоким destructive potential в production и platform-wide policy changes.

### Evidence taxonomy
| Слой evidence | Что подтверждает |
|---|---|
| Intent / contract evidence | problem statement, scope boundaries, AC/NFR, ADR/design decisions |
| Verification evidence | unit/integration/contract/regression checks и их результаты |
| Review evidence | review contract, unresolved risks, owner decisions, waivers |
| Release readiness evidence | rollout prerequisites, rollback notes, observability minimum, open blockers |
| Runtime / postdeploy evidence | health signals, incident traces, postdeploy findings, remediation triggers |
| Audit / traceability evidence | links `issue -> PR -> docs -> labels -> run`, quality gates и stage transitions |

### Verification minimum и review contract
| Tier | Verification minimum | Review contract |
|---|---|---|
| `low` | targeted checks + `git diff --check` + change summary | evidence-based self-check и owner review без лишней stage-эскалации |
| `medium` | typed AC + targeted automated tests + regression note | обязательный review по change intent, checks и residual risks |
| `high` | contract/integration coverage + regression plan + rollback note | reviewer/owner review с проверкой evidence completeness и explicit risk handling |
| `critical` | полный readiness package, release/postdeploy evidence и manual stop criteria | multi-role review gate (`reviewer`/`qa`/`sre` + Owner) без права скрывать missing evidence |

### Draft mapping `risk tier -> stages/gates -> evidence`
| Tier | Минимальная stage-траектория | Обязательные gates/evidence |
|---|---|---|
| `low` | short path допускается по launch profile при сохранении traceability | problem statement, targeted checks, rollback note, owner review |
| `medium` | минимум `feature`-контур `intake -> prd -> design -> plan -> dev -> qa -> release -> postdeploy -> ops` | AC/NFR, verification evidence, review summary, QA/release evidence |
| `high` | `feature` + обязательный `arch`; пропуск stage только по owner decision | architecture/design evidence, regression gate, rollback/readiness notes, postdeploy follow-up |
| `critical` | полный `new-service`-контур без silent сокращения | full doc-flow, explicit risk framing, release safety package и operational evidence |

## План этапов и handover

| Stage | Основной артефакт | Целевая роль | Правило выхода |
|---|---|---|---|
| Intake (`#469`) | Problem/Brief/Scope/Constraints + intake AC | `pm` | Owner review intake-пакета и создана issue следующего этапа |
| Vision (`#471`) | Mission, quality north star, persona outcomes, success metrics, guardrails | `pm` | Зафиксирован vision baseline и создана issue `#476` для `run:prd` |
| PRD (`#476`) | User stories, FR/AC/NFR, risk/evidence scenarios, expected evidence и proportional stage-gate contract | `pm` + `sa` | Подтверждён PRD package и создана issue `#484` для `run:arch` |
| Architecture (`#484`) | Ownership matrix, service/rule boundaries, governance data surfaces | `sa` | Подтверждены архитектурные границы и создана issue для `run:design` |
| Design (`TBD`) | Typed contracts для quality signals, evidence package и stage-gate orchestration | `sa` + `qa` | Подготовлен implementation-ready design package и создана issue для `run:plan` |
| Plan (`TBD`) | Delivery waves, quality-gates, execution decomposition, DoR/DoD | `em` + `km` | Сформирован execution package и создана issue для owner-managed `run:dev` |

## Guardrails спринта
- Качество определяется как свойства изменения и поставки, а не как требование «читать код дольше».
- Governance должна быть risk-based и proportional: low-risk changes нельзя автоматически обременять тем же контуром, что `critical`.
- Sprint S13 не выбирает финальные rollout/canary/alerting implementations; этот runtime/UI слой остаётся downstream scope Sprint S14.
- Existing baselines из S6, S9 и S12 расширяются и переиспользуются, а не переписываются как будто их не существовало.
- Каждый doc-stage до `run:dev` обязан выпускать следующую follow-up issue без trigger-лейбла; `run:plan` создаёт handover issue для `run:dev`, а trigger запускает Owner отдельно.

## Handover
- Day1/Day2 package:
  - `docs/delivery/sprints/s13/sprint_s13_quality_governance_system.md`;
  - `docs/delivery/epics/s13/epic_s13.md`;
  - `docs/delivery/epics/s13/epic-s13-day1-quality-governance-intake.md`;
  - `docs/delivery/epics/s13/epic-s13-day2-quality-governance-vision.md`;
  - `docs/delivery/epics/s13/epic-s13-day3-quality-governance-prd.md`;
  - `docs/delivery/epics/s13/prd-s13-day3-quality-governance-system.md`;
  - `docs/delivery/traceability/s13_quality_governance_system_history.md`.
- Следующий stage: `run:arch` в Issue `#484`.
- Sprint S14 (Issue `#470`) остаётся downstream инициативой и не должен стартовать implementation-first без решений S13 по risk/evidence/verification baseline.
- На `run:arch` нельзя потерять следующие решения intake + vision + PRD:
  - quality north star должен описывать свойства change delivery и safe throughput, а не «внимательность ревью»;
  - explicit risk tier обязателен для каждого change package;
  - risk tiers `low / medium / high / critical` остаются обязательным baseline;
  - список high/critical changes считается source input для дальнейшей proportional governance, а не исчерпывающим final classifier;
  - evidence completeness, verification minimum и review/waiver discipline остаются отдельными product constructs, а не «деталями QA» или только `reviewer` stage;
  - proportional governance обязательна: low-risk changes нельзя автоматически обременять overhead уровня `high/critical`;
  - high/critical changes не допускают silent waivers и implicit gates;
  - governance-gap feedback loop должен уметь поднимать late reclassification и missing evidence как отдельные outcomes;
  - Sprint S14 (`#470`) наследует baseline S13 и не становится источником правды по policy semantics;
  - каждый следующий doc-stage должен создавать следующую follow-up issue без trigger-лейбла.
