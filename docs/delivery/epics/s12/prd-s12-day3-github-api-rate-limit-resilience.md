---
doc_id: PRD-CK8S-S12-I416
type: prd
title: "GitHub API rate-limit resilience — PRD Sprint S12 Day 3"
status: in-review
owner_role: PM
created_at: 2026-03-13
updated_at: 2026-03-13
related_issues: [366, 413, 416, 418]
related_prs: []
related_docsets:
  - docs/delivery/sprints/s12/sprint_s12_github_api_rate_limit_resilience.md
  - docs/delivery/epics/s12/epic_s12.md
  - docs/delivery/issue_map.md
  - docs/delivery/requirements_traceability.md
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-13-issue-416-prd"
---

# PRD: GitHub API rate-limit resilience

## TL;DR
- Что строим: GitHub-first controlled wait capability платформы для recoverable primary/secondary rate-limit по двум operational contour: `platform PAT` и `agent bot-token`.
- Для кого: owner/reviewer, platform operator и agent runtime.
- Почему: сейчас recoverable rate-limit выглядит как ложный `failed`, непрозрачный локальный retry или неясная пауза без contour attribution и следующего шага.
- MVP: typed controlled wait-state, contour attribution, transparency surfaces, safe resume/manual intervention path, audit-safe recovery hints и backpressure discipline для agent path.
- Критерии успеха: recoverable rate-limit переводится в понятный wait-state без infinite local retries, а owner/operator видит причину ожидания, affected contour и следующий допустимый шаг.

## Проблема и цель
- Problem statement:
  - GitHub API budget exhaustion уже затрагивает orchestration path платформы и agent runtime path, но продукт пока не даёт канонического controlled wait-state;
  - без отдельного rate-limit contract пользователь и оператор видят либо обычный `failed`, либо неясную задержку без объяснения, какой контур упёрся в лимит;
  - агентный path рискует продолжать локальные GitHub-вызовы после typed rate-limit signal, если backpressure не закреплён как продуктовый инвариант;
  - без явной product semantics для primary vs secondary rate limits следующие stages быстро уйдут в retry-first или transport-first решения.
- Цели:
  - зафиксировать core-MVP contract для controlled wait-state, contour attribution, transparency и resume semantics;
  - описать пользовательские сценарии, FR/AC/NFR и edge cases до architecture/design этапов;
  - сохранить split `platform PAT` vs `agent bot-token` и hard-failure separation как неподвижные ограничения;
  - подготовить architecture handover с явными invariants для typed recovery hints, backpressure и manual-intervention path.
- Почему сейчас:
  - Sprint S12 уже подтвердил отдельную инициативу на intake и vision этапах;
  - без PRD stage архитектура будет спорить о service boundaries и storage/runtime деталях без общего product baseline;
  - GitHub Docs, проверенные 2026-03-13, отдельно описывают primary и secondary rate limits и recovery hints, поэтому продукт нужно зафиксировать до архитектурного выбора реализации.

## Зафиксированные продуктовые решения
- `D-416-01`: Controlled wait-state предпочтительнее ложного `failed`, если provider signal указывает на recoverable rate-limit.
- `D-416-02`: `platform PAT` и `agent bot-token` остаются разными operational contour; продукт не усредняет их в один статус "GitHub недоступен".
- `D-416-03`: Primary и secondary rate limits остаются разными product signals; UX опирается на typed recovery hints, а не на один фиксированный countdown.
- `D-416-04`: Agent path обязан backpressure upstream через platform-managed wait, а не brute-force ретраить GitHub локально.
- `D-416-05`: `bad credentials`, invalid token scope, permission/policy failures и другие non-rate-limit ошибки не могут быть замаскированы под controlled wait.
- `D-416-06`: Predictive budgeting, multi-provider quota governance и adapter-specific notifications остаются deferred scope и не блокируют core Sprint S12 MVP.

## Scope boundaries
### In scope
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

### Out of scope
- Универсальный quota-management framework для всех внешних провайдеров.
- Автоматическая смена credentials, token rotation или token-scope remediation как стандартный ответ на rate-limit.
- Predictive budget analytics, proactive throttling и global quota dashboards.
- Channel-specific notifications, reminders и adapter-specific escalation rules.
- Storage schema, transport protocol и runtime implementation details до `run:arch` и `run:design`.

## Пользователи / персоны

| Persona | Основная задача | Что считает успехом |
|---|---|---|
| Owner / reviewer | Понять, почему run поставлен на паузу, и какой следующий шаг допустим | По первой видимой поверхности понятно, какой контур упёрся в лимит, когда возможен recovery и нужен ли ручной action |
| Platform operator | Диагностировать affected contour и не перепутать recoverable wait с hard failure | Wait-state объясним, имеет audit/correlation evidence и не требует вскрывать сырые логи для базового решения |
| Agent runtime | Корректно переждать GitHub rate-limit без local retry-loop | Typed signal переводится в platform-managed wait/backpressure и завершается resume или manual-action-required path |

## User stories и wave priorities

| Story ID | История | Wave | Приоритет |
|---|---|---|---|
| `S12-US-01` | Как owner/reviewer, я хочу видеть recoverable GitHub rate-limit как controlled wait-state, чтобы не принимать его за ложный `failed` | Wave 1 | `P0` |
| `S12-US-02` | Как owner/operator, я хочу видеть affected contour (`platform PAT` или `agent bot-token`) и recovery hint, чтобы понимать, что именно заблокировано и что делать дальше | Wave 1 | `P0` |
| `S12-US-03` | Как агент, я хочу передавать GitHub rate-limit в platform-managed wait/backpressure, чтобы не делать бесконечные локальные retries | Wave 1 | `P0` |
| `S12-US-04` | Как operator, я хочу различать recoverable rate-limit и hard failure, чтобы не терять время на ложные auto-recovery ожидания | Wave 2 | `P0` |
| `S12-US-05` | Как owner, я хочу безопасный выход из wait-state в auto-resume или manual-action-required, чтобы не угадывать завершение recovery window | Wave 2 | `P0` |
| `S12-US-06` | Как продуктовая команда, я хочу оставить predictive budgeting, multi-provider governance и adapter-specific notifications вне core MVP, чтобы не блокировать основную capability | Wave 3 | `P1` |

### Wave sequencing
- Wave 1 (`core MVP`, `P0`):
  - controlled wait-state для двух contour;
  - contour attribution;
  - typed recovery hints;
  - backpressure discipline на agent path.
- Wave 2 (`transparency + resume evidence`, `P0`):
  - owner/operator visibility surfaces;
  - hard-failure separation;
  - safe auto-resume/manual-action-required path;
  - audit/correlation evidence.
- Wave 3 (`deferred`, `P1`):
  - predictive budgeting;
  - multi-provider quota governance;
  - adapter-specific notifications/reminders.

## Functional Requirements

| ID | Требование |
|---|---|
| `FR-416-01` | Recoverable GitHub primary/secondary rate-limit на `platform PAT` path должен переводить затронутый run/task в typed controlled wait-state, а не в ложный `failed`. |
| `FR-416-02` | Recoverable GitHub primary/secondary rate-limit на `agent bot-token` path должен завершать локальный retry-path агента и передаваться в platform-managed backpressure/controlled wait. |
| `FR-416-03` | Product contract должен сохранять contour attribution (`platform PAT` vs `agent bot-token`) и affected operation class для каждого wait-state кейса. |
| `FR-416-04` | Пользовательские поверхности (минимум service-comment и staff UI) должны показывать причину ожидания, affected contour, recovery hint и следующий допустимый шаг без обязательного просмотра сырых логов. |
| `FR-416-05` | Product contract должен явно различать recoverable rate-limit wait и hard failures (`bad credentials`, invalid token scope, policy/permission failure, non-rate-limit 403/429 variants). |
| `FR-416-06` | Primary и secondary rate-limit semantics должны рассматриваться как разные product signals, если это влияет на recovery hint, UX и manual-intervention path. |
| `FR-416-07` | Wait-state contract должен использовать typed recovery hints (`Retry-After`, `X-RateLimit-*`, response class и derived status), а не обещать один фиксированный countdown как invariant. |
| `FR-416-08` | Auto-resume допускается только когда recovery signal и состояние потока делают продолжение безопасным; иначе run должен переходить в explicit manual-action-required path. |
| `FR-416-09` | Каждое вхождение и завершение wait-state должно сохранять audit-safe evidence: contour, operation class, provider signal source, recovery hint, exit reason и correlation. |
| `FR-416-10` | Agent path после typed rate-limit detection не может продолжать infinite local retry-loop; локальные попытки после handoff считаются policy violation. |
| `FR-416-11` | GitHub остаётся текущим provider baseline для core MVP; predictive budgeting, multi-provider governance и adapter-specific notifications остаются deferred scope и не блокируют Sprint S12. |
| `FR-416-12` | Для core flows должны быть определены expected product evidence и telemetry, достаточные для acceptance walkthrough и architecture handover. |

## Acceptance Criteria (Given/When/Then)

### `AC-416-01` Controlled wait для `platform PAT`
- Given platform management path упёрся в recoverable GitHub rate-limit,
- When система классифицирует сигнал как recoverable wait,
- Then run/task получает typed controlled wait-state вместо ложного `failed`, а owner/operator видит affected contour, recovery hint и допустимый следующий шаг.
- Expected evidence: acceptance walkthrough `signal detected -> wait entered -> visibility shown`.

### `AC-416-02` Backpressure для `agent bot-token`
- Given агент получил typed GitHub rate-limit signal на `agent bot-token` path,
- When сигнал признан recoverable,
- Then локальные GitHub retries прекращаются, ожидание передаётся в platform-managed wait/backpressure, а wait-state фиксирует contour attribution и exit condition.
- Expected evidence: audit trail `rate_limit_detected -> local retries stopped -> wait entered`.

### `AC-416-03` Hard-failure separation
- Given GitHub ответ относится к `bad credentials`, invalid token scope, permission/policy failure или другому non-rate-limit hard failure,
- When система классифицирует ответ,
- Then кейс не попадает в controlled wait-state и остаётся ошибкой/blocked-path с объяснимой причиной.
- Expected evidence: negative acceptance matrix по hard-failure сценариям.

### `AC-416-04` Typed recovery hints без ложных обещаний
- Given GitHub secondary limit не даёт точного countdown,
- When система показывает пользователю recovery hint,
- Then UX использует typed hint и не обещает фиксированное время восстановления как гарантированный факт.
- Expected evidence: visibility review для secondary-limit сценария.

### `AC-416-05` Safe resume path
- Given wait-state достиг момента recovery eligibility,
- When recovery hint указывает, что продолжение допустимо,
- Then run выходит из controlled wait в auto-resume либо в manual-action-required path, сохраняя понятный exit reason и audit evidence.
- Expected evidence: lifecycle walkthrough `wait entered -> recovery eligible -> resumed/manual action`.

### `AC-416-06` Manual-intervention path
- Given auto-resume небезопасен или provider signals недостаточны,
- When wait-state не может завершиться автоматически,
- Then owner/operator получает явный manual-intervention path без бесконечного ожидания и без скрытых локальных retry.
- Expected evidence: acceptance scenario `wait entered -> manual action required`.

### `AC-416-07` Deferred scope discipline
- Given команда обсуждает predictive budgeting, multi-provider governance или adapter-specific notifications,
- When принимается решение по core MVP,
- Then эти потоки остаются deferred scope и не блокируют release core controlled wait capability.
- Expected evidence: PRD handover package и deferred-scope decision record.

## Edge cases и non-happy paths

| ID | Сценарий | Ожидаемое поведение | Evidence |
|---|---|---|---|
| `EC-416-01` | `platform PAT` и `agent bot-token` одновременно упираются в лимит | Контуры остаются различимыми; visibility и audit не схлопывают их в один общий статус | dual-contour scenario |
| `EC-416-02` | Приходит secondary limit без надёжного countdown | UX показывает typed recovery hint и manual-intervention fallback вместо точного countdown promise | secondary-hint scenario |
| `EC-416-03` | После handoff агент продолжает локально вызывать GitHub | Это считается policy violation и не входит в допустимый happy path | retry-loop negative scenario |
| `EC-416-04` | GitHub возвращает 403, но причина не rate-limit | Кейс остаётся hard failure и не превращается в controlled wait | hard-failure classification scenario |
| `EC-416-05` | Recovery window наступил, но контекст run уже небезопасен для auto-resume | Flow переходит в explicit manual-action-required path | unsafe-resume scenario |
| `EC-416-06` | Wait-state затрагивает owner-facing visibility, но staff/operator surface недоступна | Service-comment остаётся минимальной обязательной поверхностью объяснения | visibility fallback scenario |
| `EC-416-07` | Команда пытается расширить MVP до multi-provider governance до `run:arch` | Scope считается deferred и не блокирует core Sprint S12 | scope-discipline scenario |

## Non-Goals
- Универсальный quota-management framework для всех внешних API.
- Автоматическая смена токена/credentials как стандартный recovery path.
- Precise countdown promises для secondary limits без достаточного provider signal.
- Adapter-specific notification UX как часть core MVP.
- Выбор storage/transport/runtime реализации в рамках PRD.

## NFR draft для handover в architecture

| ID | Требование | Как измеряем / проверяем |
|---|---|---|
| `NFR-416-01` | Controlled wait recovery rate для recoverable GitHub rate-limit должна целиться в `>= 85%` | product telemetry + acceptance walkthrough против `NSM-413-01` |
| `NFR-416-02` | Wait-state clarity rate для affected contour, причины и next step должна целиться в `>= 90%` | service-comment/staff evidence + owner feedback против `PM-413-01` |
| `NFR-416-03` | False-failed rate для recoverable rate-limit должна целиться в `<= 5%` | run outcome audit против `OPS-413-01` |
| `NFR-416-04` | Agent local retry escape rate после typed detection должна оставаться `0%` | `agent_sessions.session_json` + audit review против `OPS-413-02` |
| `NFR-416-05` | p75 времени от recovery eligibility до выхода из wait-state должна целиться в `<= 5 минут` | wait-state timestamps + run status transitions против `REL-413-01` |
| `NFR-416-06` | Каждая wait-state запись должна иметь contour attribution, provider signal source, recovery hint, exit reason и correlation (`100%`) | audit evidence + `GOV-413-01` |
| `NFR-416-07` | Product surfaces не должны обещать fixed countdown там, где provider signals недостаточны | UX/policy review для secondary-limit scenarios |

## Analytics и product evidence
- События:
  - `github_rate_limit_detected`
  - `github_rate_limit_classified`
  - `run_wait_entered`
  - `run_wait_visibility_published`
  - `run_wait_resume_eligible`
  - `run_wait_resumed`
  - `run_wait_manual_action_required`
  - `agent_local_retry_blocked`
- Метрики:
  - `NSM-413-01` Controlled wait recovery rate
  - `PM-413-01` Wait-state clarity rate
  - `OPS-413-01` False-failed rate
  - `OPS-413-02` Agent local retry escape rate
  - `REL-413-01` Wait-exit resume latency p75
  - `GOV-413-01` Contour attribution completeness
- Expected evidence:
  - acceptance walkthrough по `platform PAT` и `agent bot-token` wait scenarios;
  - negative matrix по hard-failure separation и retry-loop prevention;
  - explicit review trace о том, что secondary-limit UX использует typed hints, а не fixed countdown promises.

## Риски и допущения
| Type | ID | Описание | Митигирующее действие | Статус |
|---|---|---|---|---|
| risk | `RSK-416-01` | Инициатива расползётся в общий retry/backoff redesign | Жёстко держать GitHub-first controlled wait capability как core story | open |
| risk | `RSK-416-02` | Product surfaces начнут обещать недостоверный countdown для secondary limits | Сохранять typed recovery hints и manual-intervention path | open |
| risk | `RSK-416-03` | Ownership wait/detect/resume останется размытым между сервисами | Передать в `run:arch` обязательный ownership matrix и alternatives backlog | open |
| risk | `RSK-416-04` | Contour attribution потеряется и снова появится абстрактный статус "GitHub недоступен" | Держать split contour как non-negotiable quality gate | open |
| assumption | `ASM-416-01` | Existing runtime/audit baseline позволяет добавить controlled wait как отдельную capability без redesign orchestration с нуля | accepted |
| assumption | `ASM-416-02` | GitHub provider signals достаточны, чтобы различать recoverable wait и hard failure на product-уровне | accepted |
| assumption | `ASM-416-03` | Основная пользовательская ценность достигается уже на уровне clarity + controlled wait + safe resume без первой волны predictive budgeting | accepted |

## Открытые вопросы для `run:arch`
- Где проходит ownership boundary между `control-plane`, `worker`, `agent-runner` и visibility surfaces для signal classification, wait orchestration и resume gating?
- Как сохранить split `platform PAT` vs `agent bot-token` на уровне service boundaries, не дублируя доменную логику?
- Где должна жить derived product semantics для typed recovery hints и manual-intervention path без premature storage/transport lock-in?
- Какие architectural alternatives нужно зафиксировать отдельно, чтобы `run:design` не переоткрывал product decisions Day3?

## Handover в `run:arch`
- Follow-up issue: `#418`.
- На архитектурном этапе нельзя потерять:
  - controlled wait вместо ложного `failed` для recoverable GitHub rate-limit;
  - split `platform PAT` и `agent bot-token`;
  - typed recovery hints вместо fixed countdown promises;
  - hard-failure separation;
  - запрет infinite local retries на agent path;
  - deferred scope для predictive budgeting, multi-provider governance и adapter-specific notifications.
- Архитектурный этап обязан определить:
  - service boundaries и ownership matrix;
  - lifecycle wait-state и resume path;
  - audit/correlation responsibilities;
  - список ADR/alternatives и issue для `run:design`.
