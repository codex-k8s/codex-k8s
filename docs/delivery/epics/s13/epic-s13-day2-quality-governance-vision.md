---
doc_id: EPC-CK8S-S13-D2-QUALITY-GOVERNANCE
type: epic
title: "Epic S13 Day 2: Vision для quality governance system в agent-scale delivery (Issues #471/#476)"
status: in-review
owner_role: PM
created_at: 2026-03-14
updated_at: 2026-03-14
related_issues: [466, 469, 470, 471, 476]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-14-issue-471-vision"
---

# Epic S13 Day 2: Vision для quality governance system в agent-scale delivery (Issues #471/#476)

## TL;DR
- Для Issue `#471` сформирован vision-package: mission, quality north star, persona outcomes, KPI/guardrails, scope boundaries и risk frame для `Quality Governance System`.
- `Quality Governance System` зафиксирована как product capability proportional change governance: качество определяется измеримыми свойствами изменения, обязательным evidence и risk-based gates, а не «героизмом ревью».
- Явно сохранён sequencing gate `Sprint S13 governance baseline -> Sprint S14 runtime/UI safety loop`: downstream stream `#470` наследует risk/evidence/verification contract и не переоткрывает его implementation-first.
- Создана follow-up issue `#476` для stage `run:prd` без trigger-лейбла; continuity chain остаётся owner-managed.
- Для GitHub continuity и PR-flow повторно подтверждён актуальный non-interactive CLI syntax через Context7 `/websites/cli_github_manual` и локальный `gh issue create --help`.

## Priority
- `P0`.

## Vision charter

### Mission statement
Сделать `Quality Governance System` явным продуктовым контрактом поставки, чтобы Owner/reviewer, delivery roles и platform operator могли масштабировать agent throughput без потери контроля: каждый change получает явный risk tier, proportional evidence package, понятный verification minimum и audit-safe основу для go/no-go решения, а low-risk changes не тонут в бюрократии уровня `high/critical`.

### Цели и ожидаемые результаты
1. Превратить качество из субъективного обсуждения про «внимательность ревью» в управляемые свойства change delivery, которые можно проверить, измерить и аудировать.
2. Зафиксировать proportional governance baseline: защитные слои усиливаются с ростом риска и blast radius, а low-risk changes проходят по облегчённому, но всё ещё traceable path.
3. Дать единый quality contract для owner/reviewer, delivery roles (`pm/em/sa/dev/qa/sre/km`) и platform operator, чтобы evidence expectations, verification minimum и waiver discipline не зависели от локального вкуса.
4. Сохранить Sprint S13 как source of truth для risk/evidence/verification baseline и не позволить Sprint S14 (`#470`) переоткрыть его implementation-first.

### Пользователи и стейкхолдеры
- Основные пользователи:
  - Owner / reviewer, которому нужен быстрый go/no-go decision surface по risk tier, evidence completeness и residual risk, а не ручной поиск «что ещё забыли приложить».
  - Delivery roles (`pm/em/sa/dev/qa/sre/km`), которым нужен общий язык для intent, verification, review и release readiness без постоянного переизобретения quality contract под каждую задачу.
  - Platform operator / Mission Control user, которому нужны понятные governance signals: какой tier у изменения, какие gates обязательны и где именно возник gap или waiver.
- Стейкхолдеры:
  - `services/internal/control-plane` и `services/jobs/worker` как будущие владельцы stage-gate orchestration, audit trail и quality-state transitions;
  - `services/external/api-gateway` и `services/staff/web-console` как downstream visibility surfaces для risk/evidence/decision UX;
  - Sprint S14 stream (`#470`), который должен строить runtime safety loop поверх этого baseline, а не вместо него.
- Владелец решения: Owner.

### Продуктовые принципы и ограничения
- Качество = свойства изменения и поставки, которые можно измерить и аудировать.
- Evidence важнее ощущения «достаточно внимательно посмотрели».
- Governance обязана быть proportional: tier `low` не получает автоматически те же gates, что `critical`.
- High/critical changes не могут опираться на silent waivers и implicit assumptions.
- Stage gate latency является таким же guardrail, как completeness evidence: quality governance не должна деградировать в process bottleneck.
- Sprint S13 фиксирует governance baseline, а Sprint S14 (`#470`) наследует его для runtime/UI safety loop.
- В рамках `run:vision` разрешены только markdown-изменения.

## Scope boundaries

### MVP scope
- Mission statement, quality north star и persona outcomes для `Quality Governance System`.
- Success metrics и guardrails для risk classification, evidence completeness, governance proportionality и safe throughput.
- Явное product framing для следующих constructs:
  - explicit risk tier;
  - mandatory evidence package;
  - verification minimum;
  - review/waiver discipline;
  - stage-gate proportionality.
- Явная граница между Sprint S13 governance baseline и Sprint S14 runtime/UI safety loop.
- Handover в `run:prd` через Issue `#476`.

### Post-MVP / deferred scope
- Выбор конкретного CI/CD, rollout controller, observability stack или Mission Control quality cockpit implementation.
- Transport/data/runtime contracts и storage ownership для quality signals до `run:arch` / `run:design`.
- Автоматическая policy enforcement и UI automation beyond agreed baseline.
- Детальная настройка service-specific thresholds без базового общего quality contract.

### Sequencing gate
- Sprint S14 (`#470`) не должен переоткрывать risk/evidence/verification baseline implementation-first.
- Existing baselines из S6 operational package, Sprint S9 Mission Control и Sprint S12 rate-limit resilience остаются обязательными reference inputs.
- Follow-up issue `#476` является единственным owner-managed входом в следующий stage `run:prd`; trigger-лейбл ставится отдельно.

## Success metrics

### North Star
| ID | Метрика | Определение | Источник | Целевое значение |
|---|---|---|---|---|
| `NSM-471-01` | Quality-governed delivery rate | Доля изменений, которые проходят explicit risk classification, приходят на review/release с полным mandatory evidence package, укладываются в policy lead-time budget своего tier и не требуют emergency remediation из-за пропущенного governance gap | `flow_events`, issue/PR metadata, release/postdeploy evidence, traceability audit | `>= 80%` на pilot-сценариях governance baseline |

### Supporting metrics
| ID | Метрика | Определение/формула | Источник | Цель |
|---|---|---|---|---|
| `GOV-471-01` | Evidence completeness rate | Доля changes, у которых mandatory evidence package заполнен полностью до owner review | issue/PR checklist audit, traceability review | `100%` для `high/critical`, `>= 95%` overall |
| `RISK-471-01` | Late risk reclassification rate | Доля changes, чей tier пришлось повышать после review/release из-за недооценённого blast radius | review findings, release/postdeploy follow-ups | `<= 5%` |
| `FLOW-471-01` | Lead-time budget attainment by tier | Доля changes, завершивших обязательный governance path в пределах budget своего risk tier | `flow_events`, stage timestamps | `>= 85%` для `low/medium`, `>= 70%` для `high/critical` |
| `GOV-471-02` | Low-risk over-governance rate | Доля `low` changes, которые получили evidence/gates уровня `high/critical` без явного обоснования | stage audit, review sampling | `<= 10%` |
| `REL-471-01` | Governance gap escape rate | Доля incidents/remediation cases, где причиной стала missing classification, missing evidence или bypass review/waiver discipline | postdeploy evidence, remediation/self-improve issues | `0%` для известных `high/critical` gaps |

### Guardrails (ранние сигналы)
- `GR-471-01`: если `GOV-471-01 < 100%` для `high/critical`, следующий stage не может ослаблять mandatory evidence и должен приоритизировать clarity contract.
- `GR-471-02`: если `RISK-471-01 > 10%`, `run:prd` и `run:arch` обязаны сначала уточнить classification rules, а не расширять automation/UI scope.
- `GR-471-03`: если `GOV-471-02 > 20%`, следующие стадии обязаны снижать bureaucratic overhead и упрощать low-risk path.
- `GR-471-04`: если `FLOW-471-01` системно не достигается по tier budgets, новые ручные gates нельзя добавлять без отдельного owner-решения.
- `GR-471-05`: если downstream runtime/UI stream пытается ввести release safety или cockpit semantics в обход explicit risk/evidence contract, stage переводится в `need:input`.

## Risks and Product Assumptions
| Тип | ID | Описание | Митигирующее действие | Статус |
|---|---|---|---|---|
| risk | `RSK-471-01` | Scope может расползтись в implementation-first redesign CI/CD, rollout automation или Mission Control UI | Жёстко держать vision на уровне governance contract и sequencing gate относительно Sprint S14 | open |
| risk | `RSK-471-02` | Метрики окажутся слишком абстрактными и не дадут проверяемого audit contract | На `run:prd` описать evidence sources и explicit formulas для core metrics | open |
| risk | `RSK-471-03` | Слишком жёсткий governance baseline убьёт throughput low-risk changes | Держать proportionality и `Low-risk over-governance rate` как обязательный guardrail | open |
| risk | `RSK-471-04` | Quality contract останется «бумагой», если risk tier/evidence/waiver discipline не станут частью stage decisions и operator surfaces | Сохранить downstream ownership split и handover в `run:prd -> run:arch` как blocking requirement | open |
| assumption | `ASM-471-01` | Существующий issue/PR/traceability baseline платформы достаточен для первой версии evidence completeness и governance audit | Подтвердить data/evidence sources на `run:arch` | accepted |
| assumption | `ASM-471-02` | Основная ценность инициативы достигается уже через explicit tiering, evidence contract и proportional gates, без немедленного полного automation layer | Сохранить runtime/UI automation в downstream Sprint S14 | accepted |
| assumption | `ASM-471-03` | Existing baselines из S6, S9 и S12 можно связать в единый governance baseline без переписывания source-of-truth документов с нуля | Проверить traceability и ownership split на `run:prd` / `run:arch` | accepted |

## Readiness criteria для `run:prd`
- [x] Mission, quality north star и persona outcomes сформулированы для owner/reviewer, delivery roles и platform operator.
- [x] KPI/success metrics и guardrails определены как измеримые product/operational сигналы.
- [x] MVP scope, deferred scope и sequencing gate `S13 governance baseline -> S14 runtime/UI stream` разделены явно.
- [x] Подтверждены неподвижные решения vision: explicit risk tier, mandatory evidence package, verification minimum, review/waiver discipline и proportionality baseline.
- [x] Создана отдельная issue следующего этапа `run:prd` (`#476`) без trigger-лейбла.

## Acceptance criteria (Issue #471)
- [x] Mission, quality north star и продуктовые принципы для `Quality Governance System` сформулированы явно.
- [x] KPI/success metrics и guardrails зафиксированы для evidence completeness, risk accuracy, lead-time proportionality, low-risk overhead и governance-gap prevention.
- [x] Персоны, scope boundaries, риски и assumptions описаны без смешения с implementation details.
- [x] Сохранён проверяемый sequencing gate `Sprint S13 governance baseline -> Sprint S14 runtime/UI safety loop`.
- [x] Подготовлен handover в `run:prd` и создана follow-up issue `#476` без trigger-лейбла.

## Handover в следующий этап
- Следующий stage: `run:prd`.
- Follow-up issue: `#476`.
- Trigger-лейбл `run:prd` на issue `#476` ставит Owner.
- На `run:prd` нельзя потерять следующие решения vision:
  - explicit risk tier обязателен для каждого change package;
  - evidence completeness, verification minimum и review/waiver discipline остаются отдельными product constructs;
  - governance должна быть proportional и не перегружать `low` тем же overhead, что `high/critical`;
  - high/critical changes не допускают silent waivers и implicit gates;
  - Sprint S14 (`#470`) наследует baseline S13 и не переоткрывает его implementation-first;
  - PRD stage обязан в конце создать issue для `run:arch` без trigger-лейбла.

## Связанные документы
- `docs/delivery/sprints/s13/sprint_s13_quality_governance_system.md`
- `docs/delivery/epics/s13/epic_s13.md`
- `docs/delivery/epics/s13/epic-s13-day1-quality-governance-intake.md`
- `docs/delivery/traceability/s13_quality_governance_system_history.md`
- `docs/delivery/issue_map.md`
- `docs/delivery/delivery_plan.md`
- `docs/product/requirements_machine_driven.md`
- `docs/product/constraints.md`
- `docs/product/agents_operating_model.md`
- `docs/product/labels_and_trigger_policy.md`
- `docs/product/stage_process_model.md`
- `docs/delivery/development_process_requirements.md`
- `docs/research/src_idea-machine_driven_company_requirements.md`
- `docs/ops/handovers/s6/operational_baseline.md`
- `docs/delivery/traceability/s9_mission_control_dashboard_history.md`
- `docs/delivery/traceability/s12_github_api_rate_limit_resilience_history.md`
- `services/internal/control-plane/README.md`
- `services/jobs/worker/README.md`
- `services/external/api-gateway/README.md`
- `services/staff/web-console/README.md`
