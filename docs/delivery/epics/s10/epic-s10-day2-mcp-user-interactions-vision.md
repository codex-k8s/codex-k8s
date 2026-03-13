---
doc_id: EPC-CK8S-S10-D2-MCP-INTERACTIONS
type: epic
title: "Epic S10 Day 2: Vision для built-in MCP user interactions (Issues #378/#383)"
status: in-review
owner_role: PM
created_at: 2026-03-12
updated_at: 2026-03-13
related_issues: [360, 378, 383]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-03-12-issue-378-vision"
  approved_by: "ai-da-stas"
  approved_at: 2026-03-13
---

# Epic S10 Day 2: Vision для built-in MCP user interactions (Issues #378/#383)

## TL;DR
- Для Issue `#378` сформирован vision-пакет: mission, north star, persona outcomes, KPI/guardrails, MVP/Post-MVP границы и риск-рамка для built-in MCP user interactions.
- Built-in MCP user interactions зафиксированы как channel-neutral product capability платформы: `user.notify` и `user.decision.request` становятся каноническим baseline для user-facing completion/decision flow без смешения с approval semantics.
- Создана follow-up issue `#383` для stage `run:prd` без trigger-лейбла; PRD должен формализовать user stories, FR/AC/NFR, edge cases и handover в `run:arch`.

## Priority
- `P0`.

## Vision charter

### Mission statement
Сделать built-in MCP user interactions базовой channel-neutral capability платформы, чтобы owner/product lead, конечный пользователь задачи и platform operator могли завершать и продолжать run через понятные уведомления и typed user decisions без GitHub-comment detours, без смешения с approval flow и без vendor lock-in на первый канал доставки.

### Цели и ожидаемые результаты
1. Дать агенту канонический product-safe способ завершить задачу или запросить решение пользователя через `user.notify` и `user.decision.request`.
2. Сократить время от момента, когда run требует пользовательского решения, до получения валидного typed ответа без ручного поиска комментариев и ad-hoc формулировок.
3. Закрепить product semantics, в которых wait-state допустим только для response-required сценариев, а delivery/retry/correlation/audit остаются responsibility platform domain.
4. Подготовить adapter-friendly baseline, поверх которого Telegram и другие каналы смогут подключаться без переопределения core interaction semantics.

### Пользователи и стейкхолдеры
- Основные пользователи:
  - Owner / product lead, которому нужен понятный task outcome и ясный next step без ручного follow-up в issue comments.
  - Конечный пользователь задачи, который должен быстро ответить агенту кнопками или свободным текстом и не разбираться в GitHub label/process mechanics.
  - Platform operator / product designer, которому нужен единый channel-neutral baseline для следующих adapter integrations и runtime policy решений.
- Стейкхолдеры:
  - `services/internal/control-plane` и `services/jobs/worker` как будущие владельцы interaction lifecycle, audit, retry и wait-state semantics;
  - `services/external/api-gateway` как будущий edge для callback/adapters;
  - `services/jobs/agent-runner` как runtime-контур, который должен получить ясную product contract surface для built-in tools.
- Владелец решения: Owner.

### Продуктовые принципы и ограничения
- User interaction flow и approval/control flow остаются раздельными доменами с разными semantics, состояниями ожидания и критериями завершения.
- `user.notify` является non-blocking инструментом: он завершает информирование и не создаёт paused/waiting state.
- `user.decision.request` является единственным core-MVP инструментом, который может переводить run в wait-state, если ответ пользователя реально нужен для продолжения.
- Core contract остаётся typed и channel-neutral: adapters не могут переопределять смысл `selected_option`, `free_text`, correlation или outcome classification.
- Delivery, retries, idempotency, audit trail и callback correlation принадлежат platform domain, а не agent pod и не channel-specific adapter.
- В рамках `run:vision` разрешены только markdown-изменения.

## Scope boundaries

### MVP scope
- Built-in MCP tool `user.notify` для:
  - завершения задачи;
  - сообщения о результате действия;
  - передачи следующего шага, который понятен человеку без дополнительного контекста из issue comments.
- Built-in MCP tool `user.decision.request` для:
  - запроса решения пользователя с 2-5 вариантами ответа;
  - optional free-text, когда фиксированных вариантов недостаточно;
  - получения typed user response, пригодного для дальнейшего platform processing.
- Channel-neutral interaction semantics:
  - outbound/inbound contract без обязательных Telegram-specific полей;
  - единая модель correlation, response kind и outcome classification;
  - сохранение separation from approval flow.
- Wait-state policy:
  - run ждёт пользователя только в response-required сценариях;
  - notification-only path остаётся non-blocking.
- Handover в `run:prd` через issue `#383`.

### Post-MVP / deferred scope
- Telegram adapter как отдельный follow-up stream после фиксации core interaction contract.
- Voice/STT, richer multi-turn conversation threads и long-lived conversational UX.
- Channel-specific affordances, routing/escalation policies, delivery preferences и advanced reminder logic.
- Любой redesign approval flow под видом user interaction scope.

### Candidate substream с отдельным gate
- Telegram и другие channel adapters остаются отдельным последовательным stream:
  - их value, delivery model и UX решения нужно подтверждать после `run:prd`;
  - они не являются blocking requirement для core Sprint S10;
  - архитектурный lock-in по adapter transport запрещён до `run:arch` и `run:design`.

## Success metrics

### North Star
| ID | Метрика | Определение | Источник | Целевое значение |
|---|---|---|---|---|
| `NSM-378-01` | Actionable interaction completion rate | Доля interaction-сценариев, где built-in tools доводят пользователя до понятного завершения: `user.notify` не требует ручного clarification follow-up, а `user.decision.request` получает валидный typed response без fallback в GitHub comments | interaction audit events + follow-up classification в `flow_events` | `>= 85%` в pilot-сценариях MVP |

### Supporting metrics
| ID | Метрика | Определение/формула | Источник | Цель |
|---|---|---|---|---|
| `PM-378-01` | Next-step clarity rate | Доля `user.notify`, после которых в течение 24 часов не потребовался дополнительный ручной follow-up для объяснения следующего шага | interaction audit + issue/comment follow-up review | `>= 85%` |
| `PM-378-02` | Decision turnaround p75 | p75 времени от отправки `user.decision.request` до получения валидного typed user response | interaction timestamps + callback audit | `<= 15 минут` |
| `GOV-378-01` | Interaction vs approval separation correctness | Доля interaction-сценариев, обработанных без использования approval-only semantics, approval storage path или approval UX shortcuts | product acceptance matrix + audit review | `100%` |
| `REL-378-01` | Correlation and replay correctness | Доля callback/retry сценариев, которые завершаются без duplicate logical response, duplicate state transition или потери correlation | retry/callback audit + incident review | `>= 99.5%` |
| `UX-378-01` | GitHub-comment fallback rate | Доля response-required сценариев, в которых пользователю пришлось уходить в GitHub comment вместо built-in response path | interaction audit + human review sample | `<= 15%` |

### Guardrails (ранние сигналы)
- `GR-378-01`: если `PM-378-01 < 70%`, дальнейшее расширение каналов замораживается до пересмотра message clarity и next-step contract.
- `GR-378-02`: если `PM-378-02 > 30 минут`, PRD обязан приоритизировать UX/contract clarity выше channel expansion.
- `GR-378-03`: если `GOV-378-01 < 100%`, инициативу нельзя переводить в implementation-ready design до устранения смешения с approval flow.
- `GR-378-04`: если `REL-378-01 < 99%`, следующие stage обязаны приоритизировать correlation/retry semantics выше adapter-specific UX.
- `GR-378-05`: если core AC начинают требовать Telegram-specific affordances, stage переводится в `need:input` до повторного owner-решения по scope.

## Risks and Product Assumptions
| Тип | ID | Описание | Митигирующее действие | Статус |
|---|---|---|---|---|
| risk | `RSK-378-01` | Scope может расползтись в Telegram-first инициативу вместо channel-neutral platform baseline | Жёстко держать Telegram в отдельном follow-up stream и не включать его в core MVP AC | open |
| risk | `RSK-378-02` | User interaction flow может быть ошибочно спроектирован как подмножество approval flow | Сохранить separation as non-negotiable guardrail уже на PRD stage | open |
| risk | `RSK-378-03` | Слишком свободный response model приведёт к ad-hoc payloads и неоднозначной интерпретации callback data | Зафиксировать typed response semantics и explicit edge cases на `run:prd` | open |
| risk | `RSK-378-04` | Notification path станет “информационным шумом” без actionable next step | Держать clarity rate и fallback rate как product gates для дальнейшего scope | open |
| assumption | `ASM-378-01` | Existing built-in MCP server `codex_k8s` достаточно расширяем, чтобы принять новые user-facing tools без отдельного runtime server block | Проверить ownership и lifecycle на `run:arch` | accepted |
| assumption | `ASM-378-02` | Пользовательская ценность достигается быстрее через typed buttons/free-text path, чем через GitHub comment-only flow | Подтвердить user stories и expected evidence на `run:prd` | accepted |
| assumption | `ASM-378-03` | Channel-neutral contract удастся сохранить без потери usability на первом канале интеграции | Подтвердить adapter constraints и contract boundaries на `run:arch`/`run:design` | accepted |

## Readiness criteria для `run:prd`
- [x] Mission и north star зафиксированы для трёх ключевых persona/outcome-потоков.
- [x] Метрики успеха и guardrails сформулированы как измеримые product/operational сигналы.
- [x] MVP scope, deferred scope и отдельный adapter follow-up stream явно разделены.
- [x] Неподвижные ограничения инициативы сохранены: built-in server `codex_k8s`, separation from approval flow, wait-state discipline, channel-neutral typed semantics.
- [x] Создана отдельная issue следующего этапа `run:prd` (`#383`) без trigger-лейбла.

## Acceptance criteria (Issue #378)
- [x] Mission, persona outcomes и north star для built-in MCP user interactions сформулированы явно.
- [x] KPI/success metrics и guardrails зафиксированы для next-step clarity, decision turnaround, separation from approval flow, fallback-to-comments и correlation correctness.
- [x] Персоны, MVP/Post-MVP границы, риски и продуктовые принципы описаны для owner/product lead, end user и platform operator flows.
- [x] Подтверждено, что MVP остаётся channel-neutral, built-in, typed и не превращается в Telegram-first или approval-first решение.
- [x] Подготовлен handover в `run:prd` и создана follow-up issue `#383` без trigger-лейбла.

## Handover в следующий этап
- Следующий stage: `run:prd`.
- Follow-up issue: `#383`.
- Trigger-лейбл `run:prd` на issue `#383` ставит Owner.
- Обязательное условие для `#383`: в конце PRD stage создать issue для stage `run:arch` без trigger-лейбла и с continuity-инструкцией сохранить channel-neutral semantics, wait-state discipline и separation from approval flow.
- На `run:prd` нельзя потерять следующие product decisions:
  - расширяется существующий built-in server `codex_k8s`, а не создаётся новый runtime server block;
  - core MVP baseline ограничен `user.notify` и `user.decision.request`;
  - `user.notify` остаётся non-blocking, а wait-state разрешён только для response-required path;
  - interaction lifecycle не смешивается с approval/control lifecycle;
  - Telegram, voice/STT и richer threads не блокируют core Sprint S10.

## Связанные документы
- `docs/delivery/sprints/s10/sprint_s10_mcp_user_interactions.md`
- `docs/delivery/epics/s10/epic_s10.md`
- `docs/delivery/epics/s10/epic-s10-day1-mcp-user-interactions-intake.md`
- `docs/delivery/issue_map.md`
- `docs/delivery/delivery_plan.md`
- `docs/product/requirements_machine_driven.md`
- `docs/product/agents_operating_model.md`
- `docs/product/labels_and_trigger_policy.md`
- `docs/product/stage_process_model.md`
- `docs/research/src_idea-machine_driven_company_requirements.md`
- `services/internal/control-plane/README.md`
- `services/jobs/agent-runner/README.md`
- `services/external/api-gateway/README.md`
