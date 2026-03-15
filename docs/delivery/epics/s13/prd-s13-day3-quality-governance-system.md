---
doc_id: PRD-CK8S-S13-I476
type: prd
title: "Quality Governance System — PRD Sprint S13 Day 3"
status: in-review
owner_role: PM
created_at: 2026-03-15
updated_at: 2026-03-15
related_issues: [466, 469, 470, 471, 476, 484]
related_prs: []
related_docsets:
  - docs/delivery/sprints/s13/sprint_s13_quality_governance_system.md
  - docs/delivery/epics/s13/epic_s13.md
  - docs/delivery/issue_map.md
  - docs/delivery/requirements_traceability.md
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-15-issue-476-prd"
---

# PRD: Quality Governance System

## TL;DR
- Что строим: risk-based product contract для качества agent-scale delivery, где каждый change получает explicit risk tier, обязательный evidence package, verification minimum и понятный review/waiver path.
- Для кого: owner/reviewer, delivery roles (`pm/em/sa/dev/qa/sre/km`) и platform operator.
- Почему: сейчас quality expectations легко расползаются между review comments, локальными привычками и implementation-first исключениями без единого decision surface.
- MVP: explicit tiering `low / medium / high / critical`, tier-aware evidence completeness, verification minimum, review/waiver discipline, proportional stage-gates, governance-gap feedback loop и publication policy `internal working draft -> semantic wave map -> published waves`.
- Критерии успеха: change-governance decisions становятся воспроизводимыми и proportional, а Sprint S14 наследует этот contract вместо повторного выбора risk/evidence policy.

## Проблема и цель
- Problem statement:
  - качество агентной поставки пока легко сводится к субъективному «насколько внимательно посмотрели», а не к воспроизводимым свойствам изменения и его evidence;
  - без единого contract для risk tier, evidence completeness, verification minimum и review/waiver discipline owner/reviewer, delivery roles и operator принимают решения на разных основаниях;
  - low-risk changes рискуют утонуть в governance overhead уровня `high/critical`, а high/critical changes могут пройти с implicit assumptions и missing evidence;
  - не нормирован bridge между внутренним поиском решения и тем, что вообще допустимо публиковать в review/merge поток: raw working draft, semantic decomposition и критерии допустимого PR смешиваются;
  - без явного product baseline downstream runtime/UI stream Sprint S14 (`#470`) быстро начнёт решать policy implementation-first и откроет повторный спор о самих правилах.
- Цели:
  - зафиксировать core-MVP contract для explicit risk tiering, mandatory evidence package, verification minimum, review/waiver discipline и proportional stage-gates;
  - описать пользовательские сценарии, FR/AC/NFR и edge cases до architecture/design этапов;
  - явно отделить hidden internal working draft от review-ready артефакта и определить publication path только через semantic waves;
  - сохранить границу `Sprint S13 governance baseline -> Sprint S14 runtime/UI stream`;
  - подготовить architecture handover с явными invariants для canonical governance state, residual-risk framing, semantic decomposition и operator visibility.
- Почему сейчас:
  - Sprint S13 уже закрепил initiative baseline на intake и vision этапах;
  - без PRD stage архитектура будет спорить о сервисных границах и surfaces без общего продукта;
  - сейчас лучшее окно зафиксировать policy baseline, пока реализация ещё не ушла в tool/UI-first компромиссы.

## Зафиксированные продуктовые решения
- `D-476-01`: Каждый change package обязан получить explicit risk tier `low / medium / high / critical` до owner review и release-go/no-go решения.
- `D-476-02`: `evidence completeness`, `verification minimum` и `review/waiver discipline` остаются отдельными product constructs; один сильный сигнал не заменяет остальные.
- `D-476-03`: Governance обязана быть proportional: `low` path не получает автоматически evidence/gates уровня `high/critical` без явного обоснования.
- `D-476-04`: Для `high/critical` changes silent waivers запрещены; любой gap требует explicit waiver и residual-risk framing.
- `D-476-05`: Launch profile и minimum stage-gates определяются сочетанием risk tier и change characteristics, а не локальной интуицией исполнителя.
- `D-476-06`: Postdeploy/remediation feedback обязан подсвечивать governance gaps, если risk tier, evidence completeness или review discipline оказались неверными.
- `D-476-07`: Sprint S13 задаёт governance baseline; Sprint S14 (`#470`) реализует runtime/UI surfaces поверх него и не переоткрывает policy baseline implementation-first.
- `D-476-08`: `Internal working draft` допустим только как скрытый внутренний артефакт поиска решения и никогда не считается merge-candidate, review artifact или published change package.
- `D-476-09`: Перед публикацией change обязан пройти путь `internal working draft -> semantic wave map -> published waves`; первой опубликованной единицей является semantic wave, а не raw draft.
- `D-476-10`: Большой PR допустим только для behaviour-neutral mechanical changes в одном bounded scope (`move/rename/split/codegen/format/refactor-only`) при наличии mechanical verification и отсутствии смешения с business-behavior change.
- `D-476-11`: Качество PR оценивается по semantic intent и independent verifiability, а не по LOC; маленький diff не считается автоматически хорошим, если он смешивает разные классы изменений.

## Scope boundaries
### In scope
- Explicit risk tiering для каждого change package с базовыми драйверами решения:
  - blast radius;
  - contracts/data impact;
  - security/policy impact;
  - runtime/release impact.
- Mandatory evidence package и tier-aware evidence completeness rules:
  - intent / contract evidence;
  - verification evidence;
  - review / waiver evidence;
  - release / postdeploy readiness evidence.
- Verification minimum по tier и правила представления residual risk.
- Review/waiver discipline:
  - когда change считается review-ready;
  - что можно waiver'ить и в каком виде;
  - что запрещено скрывать для `high/critical`.
- Proportional stage-gates и launch-profile expectations по tier.
- Draft-to-wave publication discipline:
  - `internal working draft` остаётся hidden exploration artifact;
  - перед review/change publication обязателен semantic wave map;
  - published unit должен иметь один основной semantic intent и отдельную verification surface.
- Governance-gap feedback loop: как postdeploy/remediation результаты возвращаются в quality contract.

### Out of scope
- Конкретные storage/schema, transport, service ownership и runtime automation decisions до `run:arch`.
- Выбор quality cockpit, rollout controller, alerting stack или operator UX implementation.
- Полная policy automation и per-service tuning до подтверждения architecture/design package.
- Подмена governance baseline ручными reviewer-only договорённостями без traceable evidence contract.

## Пользователи / персоны

| Persona | Основная задача | Что считает успехом |
|---|---|---|
| Owner / reviewer | Принять go/no-go решение по change без ручного поиска «что ещё забыли приложить» | Risk tier, evidence completeness, verification result, gaps/waivers и residual risk видны сразу |
| Delivery roles | Готовить change package по понятным правилам и не спорить каждый раз о базовых ожиданиях | Есть единый contract: какой tier, какой minimum evidence, какая verification и какой review path требуется |
| Platform operator | Видеть governance state и понимать, где возник gap, waiver или risk escalation | Signals объяснимы и не зависят от чтения сырого чата или локальных договорённостей |
| Downstream runtime/UI stream | Реализовать surfaces и automation без переоткрытия policy baseline | Product contract стабилен, а runtime/UI строится поверх него |

## User stories и wave priorities

| Story ID | История | Wave | Приоритет |
|---|---|---|---|
| `S13-US-01` | Как owner/reviewer, я хочу видеть explicit risk tier и rationale по change package, чтобы быстро понять его blast radius | Wave 1 | `P0` |
| `S13-US-02` | Как delivery role, я хочу видеть mandatory evidence package по tier, чтобы готовить change без угадывания обязательных артефактов | Wave 1 | `P0` |
| `S13-US-03` | Как owner/reviewer, я хочу видеть verification minimum и фактический статус его выполнения, чтобы отличать review-ready package от incomplete package | Wave 1 | `P0` |
| `S13-US-04` | Как owner/reviewer, я хочу explicit waiver/residual-risk path, чтобы gaps не превращались в silent assumptions | Wave 2 | `P0` |
| `S13-US-05` | Как delivery role, я хочу proportional governance для `low` changes, чтобы не тратить время на избыточные gates без явной причины | Wave 2 | `P0` |
| `S13-US-06` | Как platform operator, я хочу видеть governance-gap feedback из postdeploy/remediation, чтобы policy можно было усиливать на основе фактов | Wave 2 | `P0` |
| `S13-US-07` | Как delivery role, я хочу иметь право на hidden internal working draft, но публиковать наружу только semantic waves, чтобы внутренний поиск решения не становился merge-candidate по умолчанию | Wave 1 | `P0` |
| `S13-US-08` | Как owner/reviewer, я хочу видеть, когда большой PR допустим как mechanical bundle, а когда даже маленький diff обязан быть разложен по semantic intent, чтобы оценка качества не сводилась к LOC | Wave 2 | `P0` |
| `S13-US-09` | Как downstream runtime/UI stream, я хочу получить стабильный policy baseline, чтобы строить surfaces и automation без переоткрытия risk/evidence contract | Wave 3 | `P1` |

### Wave sequencing
- Wave 1 (`core contract`, `P0`):
  - explicit risk tier;
  - mandatory evidence package;
  - verification minimum;
  - tier-aware completeness rules;
  - `internal working draft` как non-mergeable hidden artifact;
  - обязательный переход к semantic wave map перед публикацией.
- Wave 2 (`decision discipline`, `P0`):
  - waiver/residual-risk path;
  - proportional low-risk path;
  - governance-gap feedback loop;
  - role-specific decision surfaces;
  - критерии допустимости large mechanical PR;
  - запрет считать small-but-semantically-mixed diff автоматически хорошим.
- Wave 3 (`deferred automation`, `P1`):
  - runtime/UI automation;
  - quality cockpit;
  - service-specific tuning;
  - advanced operator workflows.

### Draft-to-wave publication policy
- `Internal working draft`:
  - может существовать как локальный черновик, используемый для поиска решения, прототипирования или end-to-end проверки гипотезы;
  - не публикуется в review/merge поток и не служит baseline для owner decision.
- `Semantic wave map`:
  - обязателен перед первой внешней публикацией change package;
  - раскладывает draft на последовательность reviewable units с одним semantic intent на волну.
- `Published waves`:
  - каждая wave должна быть independently reviewable и иметь свой verification surface;
  - bundle допускается только если wave behaviour-neutral, bounded-scope и mechanical по природе.

## Functional Requirements

| ID | Требование |
|---|---|
| `FR-476-01` | Каждый change package должен получить explicit risk tier `low / medium / high / critical` до owner review и release decision. |
| `FR-476-02` | Risk tier должен сопровождаться кратким rationale по blast radius, security/policy impact, contracts/data impact и runtime/release impact. |
| `FR-476-03` | Для каждого tier должен существовать mandatory evidence package с tier-aware completeness rules; evidence completeness оценивается отдельно от verification result. |
| `FR-476-04` | Verification minimum должен быть определён по tier и change characteristics; incomplete verification не может маскироваться под complete evidence. |
| `FR-476-05` | Review-ready decision surface должен явно показывать risk tier, evidence completeness status, verification status, open gaps, waiver state и residual risk. |
| `FR-476-06` | Waiver path должен быть explicit: кто разрешил отклонение, что именно waived, какой residual risk принят и какой follow-up требуется. |
| `FR-476-07` | Для `high/critical` changes silent waivers и implicit gates запрещены; missing evidence должен оставаться видимым до owner decision. |
| `FR-476-08` | `Low` changes должны проходить proportional governance path и не получать overhead уровня `high/critical` без явного обоснования или risk escalation. |
| `FR-476-09` | Product contract должен определять minimum stage-gates и launch-profile expectations по сочетанию risk tier и change characteristics. |
| `FR-476-10` | Governance-gap feedback из review, release, postdeploy или remediation должен быть способен поднять late risk reclassification, missing evidence и bypassed discipline как отдельные product outcomes. |
| `FR-476-11` | Owner/reviewer, delivery roles и operator surfaces должны использовать один и тот же canonical vocabulary change governance, а не role-local трактовки. |
| `FR-476-12` | Sprint S13 должен оставаться source-of-truth для policy baseline; Sprint S14 (`#470`) реализует runtime/UI surfaces, не изменяя core contract без отдельного product цикла. |
| `FR-476-13` | Для core governance flows должны быть определены expected product evidence и telemetry, достаточные для acceptance walkthrough и architecture handover. |
| `FR-476-14` | `Internal working draft` не может считаться review-ready или merge-ready артефактом; он всегда остаётся hidden internal artifact. |
| `FR-476-15` | Перед первой публикацией change должен быть разложен в `semantic wave map`; наружу публикуются только semantic waves с явным intent и verification surface. |
| `FR-476-16` | Large PR допустим только когда change behaviour-neutral, mechanical и bounded-scope; допустимые классы изменений должны быть явно перечислены и отдельно верифицируемы. |
| `FR-476-17` | Размер diff не является самостоятельным quality signal: semantically mixed change package обязан быть разложен, даже если он мал по LOC. |

## Acceptance Criteria (Given/When/Then)

### `AC-476-01` Explicit risk tier assignment
- Given change package входит в governance flow,
- When выполняется initial classification,
- Then package получает explicit risk tier и rationale до owner review.
- Expected evidence: walkthrough `change opened -> tier assigned -> rationale visible`.

### `AC-476-02` Mandatory evidence completeness
- Given risk tier уже определён,
- When команда готовит change package,
- Then mandatory evidence package по этому tier либо заполнен полностью, либо явные gaps отмечены отдельно.
- Expected evidence: completeness matrix по tier с visible status каждого обязательного блока.

### `AC-476-03` Verification minimum cannot hide behind completeness
- Given change package имеет собранные intent/review artifacts,
- When verification minimum по tier не выполнен,
- Then package не считается fully review-ready только за счёт формальной completeness остальных блоков.
- Expected evidence: negative scenario `evidence partly complete / verification incomplete`.

### `AC-476-04` Explicit waiver and residual risk
- Given change отклоняется от mandatory evidence или verification minimum,
- When owner/reviewer принимает решение продолжать,
- Then waiver фиксируется явно вместе с residual risk и обязательным follow-up.
- Expected evidence: walkthrough `gap detected -> waiver requested -> residual risk stated`.

### `AC-476-05` Proportional low-risk path
- Given change отнесён к `low`,
- When выбирается governance path,
- Then package не получает evidence/gates уровня `high/critical` без явного основания или risk escalation.
- Expected evidence: comparison matrix `low` vs `high/critical`.

### `AC-476-06` High/Critical no-silent-waiver policy
- Given change отнесён к `high` или `critical`,
- When в package отсутствует обязательный evidence block или verification proof,
- Then gap остаётся видимым и не может быть скрыт как implicit assumption.
- Expected evidence: high/critical readiness review with visible gap state.

### `AC-476-07` Stage-gate and launch-profile mapping
- Given change имеет risk tier и понятные characteristics,
- When нужно определить minimum path до `run:dev`,
- Then minimum stages/gates и required evidence остаются traceable и повторяемыми для этой комбинации.
- Expected evidence: mapping `tier + change characteristics -> gates/evidence`.

### `AC-476-08` Governance-gap feedback loop
- Given review, release или postdeploy находит under-classification, missing evidence или bypassed waiver discipline,
- When outcome фиксируется,
- Then governance gap попадает в explicit feedback path, а не теряется в narrative comments.
- Expected evidence: lifecycle walkthrough `gap detected -> governance feedback recorded`.

### `AC-476-09` S13 to S14 boundary
- Given downstream runtime/UI stream начинает проектировать surfaces или automation,
- When он использует governance baseline,
- Then Sprint S14 наследует PRD contract из Sprint S13 и не переопределяет его implementation-first.
- Expected evidence: handover package и deferred-scope decision record.

### `AC-476-10` Internal working draft is never published as review artifact
- Given delivery role собрал локальный working draft для поиска решения,
- When change готовится к первой внешней публикации,
- Then raw draft не становится review/merge package и должен быть преобразован в semantic waves.
- Expected evidence: negative path `working draft exists / raw draft not published`.

### `AC-476-11` Semantic wave map before publication
- Given найден рабочий end-to-end путь,
- When change переходит из внутреннего черновика в review stream,
- Then существует semantic wave map, который разбивает draft на последовательность independently reviewable waves.
- Expected evidence: walkthrough `working draft -> semantic wave map -> wave 1..N`.

### `AC-476-12` Large mechanical PR is explicitly constrained
- Given change публикуется большим bundle,
- When owner/reviewer оценивает допустимость такого bundle,
- Then bundle должен быть behaviour-neutral, bounded-scope, mechanical по природе и иметь mechanical verification.
- Expected evidence: admissibility matrix `large mechanical bundle -> allowed classes / required verification`.

### `AC-476-13` Small semantically mixed diff still fails publication quality
- Given diff мал по LOC,
- When он смешивает несколько semantic classes изменения,
- Then такой package не считается автоматически хорошим и требует decomposition или явного policy exception.
- Expected evidence: negative scenario `small diff / mixed intent / decomposition required`.

## Edge cases и non-happy paths

| ID | Сценарий | Ожидаемое поведение | Evidence |
|---|---|---|---|
| `EC-476-01` | Несколько локальных `low`-changes вместе дают существенный blast radius | Допускается escalation tier; proportional path не должен маскировать accumulated risk | reclassification scenario |
| `EC-476-02` | Change выглядит безопасным, но затрагивает security/policy surface | Tier определяется по actual impact, а не только по размеру diff | security-impact scenario |
| `EC-476-03` | Tests пройдены, но отсутствует rollback/residual-risk framing для `high` change | Verification не заменяет release/readiness evidence; gap остаётся видимым | incomplete-readiness scenario |
| `EC-476-04` | Для `low` change кто-то требует gates уровня `critical` «на всякий случай» | Over-governance считается отклонением и требует явного обоснования | over-governance scenario |
| `EC-476-05` | Waiver запрашивается для `critical` change перед release gate | Waiver остаётся explicit и сопровождается residual risk; silent acceptance запрещён | critical-waiver scenario |
| `EC-476-06` | Postdeploy incident показывает, что tier был занижен | Governance-gap feedback фиксирует late reclassification и возвращает его в policy loop | postdeploy-gap scenario |
| `EC-476-07` | Sprint S14 пытается добавить runtime/UI gate, которого нет в baseline | Новый gate не становится source-of-truth без отдельного product цикла | downstream-boundary scenario |
| `EC-476-08` | Агент собрал рабочий end-to-end draft, который меняет migration + domain logic + transport + UI | Draft остаётся внутренним; наружу публикуется только semantic wave map с последовательной декомпозицией | draft-to-wave scenario |
| `EC-476-09` | Большой diff состоит только из move/rename/codegen/format/refactor-only в одном bounded scope | Bundle допустим как один PR, если поведение не меняется и есть mechanical verification | large-mechanical scenario |
| `EC-476-10` | Diff небольшой, но одновременно меняет migration, logic и transport/UI | Size не спасает package: требуется semantic decomposition или explicit exception | small-mixed scenario |

## Non-Goals
- Выбор конкретных service boundaries, storage schema и transport contracts в рамках PRD.
- Выбор quality cockpit, rollout controller, alerting/observability implementation и release-safety tooling.
- Автоматическое enforcement всего governance contract без подтверждённой architecture/design модели.
- Унификация всех будущих compliance/policy инициатив внутри одного Sprint S13 baseline.

## NFR draft для handover в architecture

| ID | Требование | Как измеряем / проверяем |
|---|---|---|
| `NFR-476-01` | `Evidence completeness rate` должна оставаться `100%` для `high/critical` и `>= 95%` overall | traceability audit + `GOV-471-01` |
| `NFR-476-02` | `Late risk reclassification rate` должна целиться в `<= 5%` | review/release/postdeploy feedback + `RISK-471-01` |
| `NFR-476-03` | `Low-risk over-governance rate` должна оставаться `<= 10%` | stage audit + `GOV-471-02` |
| `NFR-476-04` | Silent-waiver rate для `high/critical` должна оставаться `0%` | review/waiver audit |
| `NFR-476-05` | `Lead-time budget attainment by tier` должна целиться в `>= 85%` для `low/medium` и `>= 70%` для `high/critical` | stage timestamps + `FLOW-471-01` |
| `NFR-476-06` | `Governance gap escape rate` для известных `high/critical` gaps должна оставаться `0%` | postdeploy/remediation evidence + `REL-471-01` |
| `NFR-476-07` | Каждый governance package должен иметь actor/correlation evidence для tier, completeness, waiver и residual-risk decision (`100%`) | audit review |

## Analytics и product evidence
- События:
  - `change_risk_tier_assigned`
  - `change_risk_tier_escalated`
  - `evidence_package_evaluated`
  - `verification_minimum_evaluated`
  - `review_waiver_requested`
  - `review_waiver_resolved`
  - `residual_risk_recorded`
  - `governance_gate_entered`
  - `governance_gate_exited`
  - `governance_gap_detected`
  - `semantic_wave_map_created`
  - `semantic_wave_published`
  - `raw_working_draft_rejected_for_publication`
- Метрики:
  - `NSM-471-01` Quality-governed delivery rate
  - `GOV-471-01` Evidence completeness rate
  - `RISK-471-01` Late risk reclassification rate
  - `FLOW-471-01` Lead-time budget attainment by tier
  - `GOV-471-02` Low-risk over-governance rate
  - `REL-471-01` Governance gap escape rate
- Expected evidence:
  - acceptance walkthrough по `low`, `medium`, `high`, `critical` governance paths;
  - walkthrough `internal working draft -> semantic wave map -> published waves`;
  - negative matrix по silent waivers, over-governance, late reclassification и small-but-semantically-mixed diffs;
  - explicit trace о том, что Sprint S14 не переопределяет policy baseline Sprint S13.

## Риски и допущения
| Type | ID | Описание | Митигирующее действие | Статус |
|---|---|---|---|---|
| risk | `RSK-476-01` | Инициатива расползётся в общий process/UI redesign вместо change-governance contract | Держать focus на risk/evidence decisions и deferred automation boundary | open |
| risk | `RSK-476-02` | Governance станет слишком тяжёлой для `low` path | Сохранять proportionality и метрику over-governance как quality gate | open |
| risk | `RSK-476-03` | Evidence completeness станет формальным чекбоксом без влияния на decision-making | Передать в `run:arch` обязательный decision surface и ownership matrix | open |
| risk | `RSK-476-04` | Downstream runtime/UI stream переоткроет policy baseline под видом implementation constraints | Держать boundary `S13 -> S14` как non-negotiable handover rule | open |
| assumption | `ASM-476-01` | Existing issue/PR/traceability surfaces достаточно зрелы для первой версии governance contract | accepted |
| assumption | `ASM-476-02` | Delivery roles готовы работать с unified vocabulary risk/evidence/waiver без новой отдельной governance role | accepted |
| assumption | `ASM-476-03` | Основная ценность достигается до полной automation/UI layer, если policy baseline зафиксирован качественно | accepted |

## Открытые вопросы для `run:arch`
- Где проходит ownership boundary для canonical change-governance aggregate, evidence-state lifecycle и residual-risk decisions?
- Какие сервисы публикуют canonical governance state и какие surfaces только отображают его без policy ownership?
- Как разложить proportional stage-gate enforcement и launch-profile mapping, не дублируя policy в нескольких местах?
- Где живут canonical `semantic wave map` и publication-status signals: в change aggregate, audit trail или отдельном planning surface?
- Как downstream runtime/UI stream Sprint S14 должен наследовать contract, не становясь источником правды по policy semantics?

## Handover в `run:arch`
- Follow-up issue: `#484`.
- На архитектурном этапе нельзя потерять:
  - explicit risk tier для каждого change package;
  - separate constructs `evidence completeness / verification minimum / review-waiver discipline`;
  - proportional low-risk path;
  - запрет silent waivers для `high/critical`;
  - `internal working draft` как hidden non-mergeable artifact;
  - publication path `internal working draft -> semantic wave map -> published waves`;
  - критерий `large PR allowed only for behaviour-neutral mechanical bounded-scope changes`;
  - правило `PR quality judged by semantic intent and independent verifiability, not LOC`;
  - governance-gap feedback loop;
  - boundary `Sprint S13 governance baseline -> Sprint S14 runtime/UI stream`.
- Архитектурный этап обязан определить:
  - service boundaries и ownership matrix;
  - lifecycle governance state и audit path;
  - ownership/publication surfaces для `semantic wave map` и draft/publication status;
  - visibility responsibilities для owner/reviewer/operator;
  - список ADR/alternatives и issue для `run:design`.
