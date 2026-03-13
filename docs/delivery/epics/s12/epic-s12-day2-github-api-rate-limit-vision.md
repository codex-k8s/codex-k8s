---
doc_id: EPC-CK8S-S12-D2-RATE-LIMIT
type: epic
title: "Epic S12 Day 2: Vision для GitHub API rate-limit resilience и controlled wait UX (Issues #413/#416)"
status: in-review
owner_role: PM
created_at: 2026-03-13
updated_at: 2026-03-13
related_issues: [366, 413, 416]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-13-issue-413-vision"
---

# Epic S12 Day 2: Vision для GitHub API rate-limit resilience и controlled wait UX (Issues #413/#416)

## TL;DR
- Для Issue `#413` сформирован vision-пакет: mission, north star, persona outcomes, KPI/guardrails, MVP/Post-MVP границы и риск-рамка для GitHub API rate-limit resilience.
- GitHub rate-limit resilience зафиксирован как GitHub-first product capability платформы: recoverable rate-limit переводит run в controlled wait-state с прозрачным объяснением причины, affected contour и следующего шага, а не в ложный `failed`.
- Создана follow-up issue `#416` для stage `run:prd` без trigger-лейбла; PRD должен формализовать user stories, FR/AC/NFR, edge cases и handover в `run:arch`.

## Priority
- `P0`.

## Vision charter

### Mission statement
Сделать GitHub API rate-limit resilience управляемой продуктовой capability платформы, чтобы Owner/reviewer, platform operator и агент могли пережидать recoverable primary/secondary rate-limit без ложного `failed`, без ручного лог-триажа и без brute-force retries, сохраняя audit-safe controlled wait-state, различимость контуров `platform PAT` и `agent bot-token`, а также понятный следующий шаг.

### Цели и ожидаемые результаты
1. Превратить recoverable GitHub rate-limit из непрозрачного сбоя в типизированный controlled wait-state с ясной причиной, affected operations и recovery hint.
2. Зафиксировать раздельный пользовательский опыт и операционную семантику для двух контуров: `platform PAT` и `agent bot-token`.
3. Сделать owner/operator transparency first-class требованием: причина ожидания, affected contour, expected recovery path и next step должны быть понятны без чтения сырых логов.
4. Закрепить agent path как backpressure-driven flow: при rate-limit сигнале агент не уходит в бесконечный локальный retry-loop, а передаёт ожидание в platform-managed controlled wait.
5. Удержать инициативу в границах GitHub-first MVP и не раздувать её в общий quota-management/retry framework для всех провайдеров.

### Пользователи и стейкхолдеры
- Основные пользователи:
  - Owner / reviewer, который ждёт завершения run/stage и должен понимать, почему работа поставлена на паузу и что произойдёт дальше.
  - Platform operator / staff user, которому нужен быстрый operational диагноз: какой контур уткнулся в лимит, какие действия затронуты и нужен ли ручной вход.
  - Агент, который должен корректно отреагировать на rate-limit сигнал через platform-managed wait/backpressure, а не пытаться добить GitHub локальными вызовами.
- Стейкхолдеры:
  - `services/internal/control-plane` и `services/jobs/worker` как будущие владельцы wait-state lifecycle, audit trail и resume orchestration;
  - `services/jobs/agent-runner` как runtime-контур, который должен перейти от локальных retries к controlled wait/backpressure;
  - `services/external/api-gateway` и `services/staff/web-console` как поверхности пользовательской прозрачности и диагностики.
- Владелец решения: Owner.

### Продуктовые принципы и ограничения
- Controlled wait-state предпочтительнее ложного `failed`, если провайдерный сигнал указывает на recoverable rate-limit.
- `platform PAT` и `agent bot-token` остаются разными operational contour; продукт не усредняет их в один абстрактный статус "GitHub недоступен".
- GitHub Docs, проверенные 2026-03-13, различают primary и secondary rate limits и используют разные recovery hints (`Retry-After`, `X-RateLimit-*`); продукт не должен обещать один фиксированный countdown как source of truth.
- Wait-state не должен скрывать `bad credentials`, `forbidden by policy`, invalid token scope и другие не-rate-limit ошибки.
- Agent path обязан backpressure upstream через platform policy, а не brute-force ретраить GitHub локально после явного rate-limit сигнала.
- В рамках `run:vision` разрешены только markdown-изменения.

## Scope boundaries

### MVP scope
- GitHub-first controlled wait capability для двух operational contour:
  - `platform PAT` path;
  - `agent bot-token` path.
- Typed user-facing transparency:
  - причина ожидания;
  - affected contour;
  - affected operation class;
  - recovery hint;
  - следующий допустимый шаг.
- Product semantics для primary и secondary rate-limit как разных сигналов, если это влияет на UX, resume policy и expected evidence.
- Agent backpressure handoff и запрет infinite local retry-loop после typed rate-limit detection.
- Resume semantics после снятия лимита:
  - когда допускается auto-resume;
  - когда нужен ручной action;
  - какие состояния и evidence должны быть сохранены.

### Post-MVP / deferred scope
- Универсальный quota-management framework для всех внешних провайдеров.
- Predictive budget analytics, proactive throttling policies и cross-provider quota dashboards.
- Автоматическая смена credentials, token rotation или token-scope remediation как стандартный ответ на rate-limit.
- Расширенные reminder/escalation policies и сложный alerting redesign за пределами GitHub-first wait UX.

### Candidate substream с отдельным gate
- Notification/adapters и richer owner feedback остаются отдельным последующим stream:
  - они не являются blocking requirement для core Sprint S12;
  - их value нужно подтверждать после `run:prd`;
  - lock-in на channel-specific semantics запрещён до `run:arch` и `run:design`.

## Success metrics

### North Star
| ID | Метрика | Определение | Источник | Целевое значение |
|---|---|---|---|---|
| `NSM-413-01` | Controlled wait recovery rate | Доля recoverable GitHub rate-limit инцидентов, которые переводятся в typed controlled wait-state и затем завершаются resume/completion/manual-next-step outcome без ложного `failed` и без ручного лог-триажа | `flow_events`, `agent_runs`, wait-state audit records, owner feedback classification | `>= 85%` на pilot-сценариях MVP |

### Supporting metrics
| ID | Метрика | Определение/формула | Источник | Цель |
|---|---|---|---|---|
| `PM-413-01` | Wait-state clarity rate | Доля wait-state кейсов, где Owner/operator может по первой видимой поверхности определить affected contour, причину ожидания и next step без открытия сырых логов | service-comment review, staff UI evidence, owner feedback | `>= 90%` |
| `OPS-413-01` | False-failed rate | Доля recoverable rate-limit инцидентов, которые заканчиваются `failed` до наступления recovery window | `flow_events` + run outcome audit | `<= 5%` |
| `OPS-413-02` | Agent local retry escape rate | Доля agent-path инцидентов, где после typed rate-limit detection агент продолжает делать локальные GitHub вызовы вместо controlled wait/backpressure | `agent_sessions.session_json` + audit review | `0%` |
| `REL-413-01` | Wait-exit resume latency p75 | p75 времени от достижения recovery-eligible момента (`Retry-After` elapsed или `x-ratelimit-reset` reached) до выхода run из controlled wait в resume/completed/manual-action-required state | wait-state timestamps + run status transitions | `<= 5 минут` |
| `GOV-413-01` | Contour attribution completeness | Доля wait-state записей, где сохранены contour, affected operation, provider signal source, recovery hint и audit correlation | wait-state records + `flow_events` | `100%` |

### Guardrails (ранние сигналы)
- `GR-413-01`: если `OPS-413-01 > 10%`, дальнейшая детализация UX/adapter scope замораживается до исправления false-failed semantics.
- `GR-413-02`: если `OPS-413-02 > 0%`, PRD и Architecture обязаны приоритизировать backpressure/agent behavior выше notification expansion.
- `GR-413-03`: если `GOV-413-01 < 100%`, stage нельзя переводить в implementation-ready design.
- `GR-413-04`: если `PM-413-01 < 75%`, следующие стадии обязаны упрощать clarity contract, а не расширять число surface/adapters.
- `GR-413-05`: если инициатива начинает требовать общий provider-agnostic quota governance для MVP, stage переводится в `need:input` до повторного owner-решения по scope.

## Risks and Product Assumptions
| Тип | ID | Описание | Митигирующее действие | Статус |
|---|---|---|---|---|
| risk | `RSK-413-01` | Scope может расползтись в общий retry/backoff redesign или multi-provider quota management | Жёстко держать инициативу вокруг GitHub-first controlled wait capability | open |
| risk | `RSK-413-02` | Продукт начнёт обещать точный countdown там, где GitHub secondary limits не дают надёжного сигнала | Опираться на typed recovery hints, а не на fixed threshold promises | open |
| risk | `RSK-413-03` | Wait-state скроет auth/authz/policy ошибки и ухудшит диагностику | Зафиксировать отдельную продуктовую семантику для non-rate-limit failures уже на PRD stage | open |
| risk | `RSK-413-04` | Без жёсткого split `platform PAT` vs `agent bot-token` UI и audit снова покажут усреднённый статус без actionable clarity | Держать contour attribution как обязательный KPI/guardrail | open |
| assumption | `ASM-413-01` | Существующий stage/audit/runtime baseline платформы достаточно зрелый, чтобы добавить controlled wait как отдельную capability без redesign orchestration с нуля | Проверить ownership и state model на `run:arch` | accepted |
| assumption | `ASM-413-02` | GitHub provider signals (`Retry-After`, `X-RateLimit-*`, response class) достаточны, чтобы на product-уровне различать recoverable wait и hard failure | Подтвердить contract детализацию на `run:prd` | accepted |
| assumption | `ASM-413-03` | Основная пользовательская ценность достигается уже на уровне clarity + controlled wait + safe resume без обязательной первой волны adapter-specific notifications | Сохранить notification/adapters как deferred stream до отдельного owner-решения | accepted |

## Readiness criteria для `run:prd`
- [x] Mission и north star сформулированы для трёх ключевых persona/outcome-потоков.
- [x] Метрики успеха и guardrails зафиксированы как измеримые product/operational сигналы.
- [x] MVP scope, deferred scope и candidate adapter stream явно разделены.
- [x] Неподвижные ограничения инициативы сохранены: GitHub-first baseline, split `platform PAT` vs `agent bot-token`, audit-safe controlled wait, provider-driven recovery hints, no infinite local retries.
- [x] Создана отдельная issue следующего этапа `run:prd` (`#416`) без trigger-лейбла.

## Acceptance criteria (Issue #413)
- [x] Mission, persona outcomes и north star для GitHub API rate-limit resilience сформулированы явно.
- [x] KPI/success metrics и guardrails зафиксированы для controlled wait recovery, clarity, false-failed prevention, contour attribution и запрета local retry-loop.
- [x] Персоны, MVP/Post-MVP границы, риски и продуктовые принципы описаны для owner/reviewer, operator и agent-path flows.
- [x] Подтверждено, что MVP остаётся GitHub-first capability и не превращается в общий redesign quota-management, retry framework или adapter-first initiative.
- [x] Подготовлен handover в `run:prd` и создана follow-up issue `#416` без trigger-лейбла.

## Handover в следующий этап
- Следующий stage: `run:prd`.
- Follow-up issue: `#416`.
- Trigger-лейбл `run:prd` на issue `#416` ставит Owner.
- Обязательное условие для `#416`: в конце PRD stage создать issue для stage `run:arch` без trigger-лейбла и с continuity-инструкцией сохранить split `platform PAT` vs `agent bot-token`, wait-state discipline, provider-driven recovery hints и запрет infinite local retries.
- На `run:prd` нельзя потерять следующие product decisions:
  - controlled wait важнее ложного `failed`, если provider signal указывает на recoverable rate-limit;
  - `platform PAT` и `agent bot-token` остаются разными operational contour;
  - продукт опирается на typed recovery hints и не обещает fixed countdown как invariant;
  - owner/operator должен видеть причину ожидания, affected contour, affected operations и следующий шаг;
  - agent path обязан backpressure upstream через platform policy, а не ретраить GitHub локально;
  - notification/adapters и multi-provider quota governance не блокируют core Sprint S12.

## Связанные документы
- `docs/delivery/sprints/s12/sprint_s12_github_api_rate_limit_resilience.md`
- `docs/delivery/epics/s12/epic_s12.md`
- `docs/delivery/epics/s12/epic-s12-day1-github-api-rate-limit-intake.md`
- `docs/delivery/issue_map.md`
- `docs/delivery/delivery_plan.md`
- `docs/product/requirements_machine_driven.md`
- `docs/product/agents_operating_model.md`
- `docs/product/labels_and_trigger_policy.md`
- `docs/product/stage_process_model.md`
- `docs/research/src_idea-machine_driven_company_requirements.md`
- `services/internal/control-plane/README.md`
- `services/jobs/agent-runner/README.md`
- `services/jobs/worker/README.md`
- `services/external/api-gateway/README.md`
- `services/staff/web-console/README.md`
