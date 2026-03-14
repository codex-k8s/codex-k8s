---
doc_id: EPC-CK8S-S13-D1-QUALITY-GOVERNANCE
type: epic
title: "Epic S13 Day 1: Intake для quality governance system в agent-scale delivery (Issue #469)"
status: in-review
owner_role: PM
created_at: 2026-03-14
updated_at: 2026-03-14
related_issues: [466, 469, 470, 471]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-14-issue-469-intake"
---

# Epic S13 Day 1: Intake для quality governance system в agent-scale delivery (Issue #469)

## TL;DR
- При росте agent-scale throughput качество нельзя описывать как «героизм ревью»: нужен product-level governance baseline, который определяет свойства изменения, уровень риска и обязательное evidence.
- Sprint S13 выделяет `Quality Governance System` в отдельную инициативу и фиксирует, что именно она задаёт: quality north star, risk tiers, evidence taxonomy, verification minimum, review contract и draft stage-gate rules.
- Intake отделяет governance-baseline от runtime/UI слоя Sprint S14 (Issue `#470`): canary, rollback automation, observability contract и quality cockpit не входят в implementation scope Sprint S13, а должны наследовать его решения.
- В качестве draft quality metrics baseline зафиксированы `lead time for change`, `change failure rate`, `mean time to restore`, `evidence completeness rate` и `stage gate latency`.
- Создана continuity issue `#471` для stage `run:vision`; следующий этап должен закрепить mission statement, persona outcomes и measurable guardrails без потери решений Day1.
- Для GitHub continuity и PR-flow повторно подтверждён актуальный non-interactive CLI syntax через Context7 `/websites/cli_github_manual`.

## Контекст
- Owner discussion в Issue `#466` разложила quality-повестку на два последовательных sprint-stream:
  - Sprint `S13` про governance baseline;
  - Sprint `S14` про release safety, observability contract и quality cockpit.
- Текущий репозиторий уже содержит важные reference baselines, которые нельзя переизобретать:
  - `docs/ops/handovers/s6/operational_baseline.md` для SLO/burn-rate/rollback discipline;
  - Sprint S9 Mission Control (`docs/delivery/epics/s9/epic_s9.md`, `docs/delivery/traceability/s9_mission_control_dashboard_history.md`) для quality surfaces и operator decision UX;
  - Sprint S12 (`docs/delivery/traceability/s12_github_api_rate_limit_resilience_history.md`) для controlled wait, evidence discipline и прозрачности state transitions.
- Stage-model и continuity policy уже фиксируют mandatory review gate и follow-up issue contract; Sprint S13 должен не переписать их, а описать, какие quality свойства и risk expectations стоят за этими gate.
- На момент intake задача ограничена markdown-only scope: никакие runtime/library/controller решения не выбираются и не реализуются.

## Рекомендованный launch profile
- Базовый launch profile: `new-service`.
- Причины:
  - инициатива меняет operating model поставки и пересекает несколько ролей и stage-gates;
  - упрощённые launch profile нельзя считать достаточными до фиксации proportional risk policy;
  - `vision` и `arch` обязательны, потому что Sprint S13 становится source-of-truth для downstream release-safety и quality-surface streams.
- Целевая stage-цепочка:
  `run:intake -> run:vision -> run:prd -> run:arch -> run:design -> run:plan`.

## Problem Statement
### As-Is
- Качество обсуждается как смесь code review discipline, тестов, release experience и operational evidence, но без единой формализованной модели.
- Нет общего языка для связки `risk tier -> verification minimum -> review contract -> stage gates -> evidence package`.
- Из-за этого масштабирование agent throughput рискует привести к субъективным решениям: low-risk changes могут быть перегружены, а high/critical changes могут пройти с неполным evidence.
- Runtime/UI инициативы вроде progressive delivery или quality cockpit могут стартовать без утверждённой governance-базы и позже открыть те же policy-вопросы заново.

### To-Be
- `Quality Governance System` описана как отдельный продуктовый baseline, который задаёт measurable quality outcomes, risk taxonomy и обязательное evidence для разных классов изменений.
- Ревью трактуется как часть evidence contract, а не как единственный механизм качества.
- Downstream streams (`release safety`, `observability contract`, `quality cockpit`) получают устойчивую baseline-модель, а не начинают с implementation-first предположений.

## Brief
- **Проблема:** скорость агентной поставки растёт быстрее формализованной модели качества, поэтому blast radius и completeness of evidence начинают зависеть от локальных решений, а не от общей policy.
- **Для кого:** для Owner и reviewer, которым нужна предсказуемая change-governance модель; для delivery-ролей (`pm/em/sa/dev/qa/sre/km`), которым нужен единый quality contract; для platform operator, который должен понимать, какие gates и evidence обязательны перед merge/release.
- **Предлагаемое решение:** выделить Sprint S13 как governance-baseline stream и пройти полный doc-flow до execution-ready package.
- **Почему сейчас:** именно сейчас появляются параллельные streams, где release safety, observability и quality surfaces уже требуют общей policy-модели, иначе каждый следующий спринт будет переоткрывать одни и те же вопросы.
- **Что считаем успехом:** intake stage закрепляет единый problem statement, baseline quality stack и handover в `run:vision` без смешения с runtime implementation.
- **Что не делаем на этой стадии:** не выбираем rollout controller, не строим cockpit UI, не вводим новый CI/CD стек и не подменяем architecture/design этапы деталями реализации.

## MVP Scope
### In scope
- Quality north star и supporting metrics как draft baseline.
- Risk taxonomy `low / medium / high / critical`.
- Draft-список изменений, которые по умолчанию относятся к `high/critical`.
- Evidence taxonomy для doc/dev/release/postdeploy контуров.
- Verification minimum и review contract как risk-based policy baseline.
- Draft mapping `risk tier -> mandatory stages/gates -> required evidence`.
- Continuity rule: каждый doc-stage до `run:dev` создаёт следующую follow-up issue без trigger-лейбла.

### Out of scope для core wave
- Выбор конкретного canary/feature-flag/rollback implementation path.
- Детальная observability architecture и Mission Control UX mechanics.
- Любая попытка заменить quality governance «ещё одним reviewer guide» без системного evidence/verification framing.
- Преждевременное обещание одинаковых gates для всех изменений независимо от risk tier.

## Constraints
- Sprint S13 остаётся governance-baseline и не должен уходить в implementation-first детализацию до `run:arch` / `run:design`, а runtime/UI decisions целиком не входят в Day1 scope.
- Sprint S14 (Issue `#470`) считается downstream инициативой и не должен переоткрывать risk/evidence baseline, который утвердит Sprint S13.
- Existing baselines из S6/S9/S12 обязаны использоваться как reference inputs; дублировать их без явного сравнения запрещено.
- Low-risk changes нельзя автоматически перегружать тем же governance overhead, что `critical`.
- Каждый следующий doc-stage обязан выпускать следующую follow-up issue без trigger-лейбла; `run:plan` создаёт handover issue для `run:dev`, а trigger-лейбл ставит Owner отдельно.

## Product principles
- Качество = свойства изменения и поставки, которые можно проверить и аудировать.
- Evidence важнее субъективного ощущения «достаточно хорошо посмотрели».
- Governance должна быть proportional: защитные слои усиливаются с ростом риска и blast radius.
- Release safety, observability и quality cockpit проектируются поверх governance baseline, а не вместо него.
- Existing operating evidence важнее «чистого листа»: зрелость системы растёт через расширение уже зафиксированных baseline-документов.

## Baseline quality stack

### Quality metrics baseline
| Метрика | Зачем нужна | Что решим на vision |
|---|---|---|
| `Lead time for change` | показывает скорость прохождения безопасного change flow | target/segmentation по risk tiers |
| `Change failure rate` | показывает цену изменения и качество release decisions | пороги и что считать failure |
| `Mean time to restore` | показывает зрелость recovery/remediation loop | windows, data source и guardrails |
| `Evidence completeness rate` | показывает, насколько change package закрывает обязательный evidence minimum | target completeness и breach policy |
| `Stage gate latency` | guardrail против бюрократического оверхеда | допустимые latency budget и escalation rules |

### Draft risk taxonomy
| Tier | Описание | Как трактуем на intake |
|---|---|---|
| `low` | локальное изменение с минимальным blast radius | может использовать сокращённый path только при отсутствии high/critical признаков |
| `medium` | изменение существующего поведения с ограниченным cross-surface impact | требует базового product + verification + review evidence |
| `high` | заметный blast radius или изменение policy/contract/data/runtime semantics | требует усиленных gates и explicit release/readiness evidence |
| `critical` | системно чувствительное изменение, где ошибка бьёт по безопасности, данным или platform-wide availability | требует полного evidence package и owner-governed stop/go framing |

### Draft high/critical change list
- DB migrations, schema ownership changes, destructive backfill/cleanup.
- Authn/authz/RBAC, approval policy, secret/token/credential handling.
- Webhook/callback security и внешние ingress paths.
- Cross-service contracts, shared typed state transitions и provider integrations с quota/billing-like consequences.
- Build/deploy/runtime orchestration, rollback policy и production safety mechanics.

### Evidence taxonomy
| Evidence layer | Минимальное содержание |
|---|---|
| Intent / contract | problem statement, scope, AC/NFR, open assumptions |
| Verification | automated/manual checks, regression scope, failures/waivers |
| Review | review summary, residual risks, approval decision, unresolved comments |
| Release readiness | rollout prerequisites, rollback notes, observability minimum |
| Runtime / postdeploy | health signals, incidents, postdeploy findings, remediation triggers |
| Audit / traceability | links `issue -> PR -> docs -> labels -> run`, gate decisions, service-comments |

### Verification minimum и review contract
| Tier | Verification minimum | Review contract |
|---|---|---|
| `low` | targeted checks, `git diff --check`, change summary | self-check + owner review по intent/evidence |
| `medium` | typed AC + automated tests + regression note | owner/reviewer проверяют completeness evidence и residual risks |
| `high` | integration/contract/regression package + rollback note | reviewer + owner review без missing evidence, explicit risk handling обязателен |
| `critical` | full readiness package, release/postdeploy evidence и manual stop criteria | multi-role gate (`reviewer`/`qa`/`sre` + Owner) без скрытых waivers |

### Draft mapping `risk tier -> stages/gates -> evidence`
| Tier | Минимальная stage-траектория | Обязательные gates/evidence |
|---|---|---|
| `low` | short path допускается по launch profile при сохранении traceability | problem statement, targeted checks, rollback note, owner review |
| `medium` | минимум `feature`-контур `intake -> prd -> design -> plan -> dev -> qa -> release -> postdeploy -> ops` | AC/NFR, verification evidence, review summary, QA/release evidence |
| `high` | `feature` + обязательный `arch`; пропуск stage только по owner decision | architecture/design evidence, regression gate, rollback/readiness notes, postdeploy follow-up |
| `critical` | полный `new-service`-контур без silent сокращения | full doc-flow, explicit risk framing, release safety package и operational evidence |

## Acceptance Criteria (Intake stage)
- [x] Есть единый problem statement и границы инициативы `Quality Governance System` для Sprint `S13`.
- [x] Зафиксирован baseline quality stack: draft quality metrics, risk tiers, high/critical change list, evidence taxonomy, verification minimum и review contract.
- [x] Явно зафиксирована draft-связка `risk tier -> mandatory stages/gates -> required evidence`.
- [x] Разделены governance-baseline Sprint `S13` и downstream runtime/UI stream Sprint `S14`.
- [x] Continuity rule зафиксирован как обязательный для всех doc-stage до `run:dev`.
- [x] Подготовлена continuity issue `#471` для stage `run:vision`.

## Stage Handover Instructions
- Следующий этап: `run:vision`.
- Созданная issue следующего этапа: `#471`.
- На stage `run:vision` обязательно сохранить и не размыть следующие решения intake:
  - quality north star должен описывать свойства change delivery и safe throughput;
  - risk tiers `low / medium / high / critical` остаются обязательным baseline;
  - список high/critical changes используется как вход для proportional governance, а не как исчерпывающий final classifier;
  - evidence taxonomy, verification minimum и review contract остаются отдельными продуктными сущностями, а не «деталями QA»;
  - Sprint S14 остаётся downstream и не получает права переоткрывать baseline implementation-first;
  - следующий doc-stage после vision обязан создать новую follow-up issue для `run:prd` без trigger-лейбла.
