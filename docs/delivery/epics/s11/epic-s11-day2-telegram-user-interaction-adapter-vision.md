---
doc_id: EPC-CK8S-S11-D2-TELEGRAM-ADAPTER
type: epic
title: "Epic S11 Day 2: Vision для Telegram-адаптера взаимодействия с пользователем (Issues #447/#448)"
status: in-review
owner_role: PM
created_at: 2026-03-14
updated_at: 2026-03-14
related_issues: [361, 444, 447, 448]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-14-issue-447-vision"
---

# Epic S11 Day 2: Vision для Telegram-адаптера взаимодействия с пользователем (Issues #447/#448)

## TL;DR
- Для Issue `#447` сформирован vision-package: mission, north star, persona outcomes, KPI/guardrails, MVP/Post-MVP границы и risk frame для Telegram-адаптера как первого внешнего channel-specific stream.
- Telegram зафиксирован как первый реальный user-facing adapter path поверх platform-owned interaction contract Sprint S10, а не как источник core semantics или shortcut вокруг approval flow.
- Проверяемый sequencing gate сохранён явно: active vision stage в `#447` может двигаться дальше только пока Sprint S10 держит closed-plan baseline `#389` и design package `#387` как effective typed interaction contract.
- Initial continuity issue `#444` остаётся intake-generated handover artifact только в исторической трассировке; после переноса active stage в `#447` она 2026-03-14 закрыта как `state:superseded`, а следующий активный stage вынесен в follow-up issue `#448` для `run:prd` без trigger-лейбла.

## Priority
- `P0`.

## Vision charter

### Mission statement
Сделать Telegram первым рабочим внешним каналом platform interaction contract, чтобы конечный пользователь, owner/product lead и platform operator могли получать actionable `user.notify` / `user.decision.request`, отвечать в привычном channel UX и при этом не терять typed platform semantics, audit trail и callback/correlation safety.

### Цели и ожидаемые результаты
1. Дать platform interaction-domain первый реальный внешний канал, через который пользователь может получить уведомление или дать решение без обязательного GitHub/comment detour.
2. Сократить время от отправки `user.decision.request` до валидного ответа пользователя за счёт inline callbacks и optional free-text, не превращая MVP в большой conversational product.
3. Подтвердить, что channel-specific UX можно добавить без Telegram-first влияния на core semantics, wait-state policy и separation from approval flow.
4. Подготовить adapter-safe baseline, поверх которого следующие stage смогут формализовать PRD, service boundaries и operability, не копируя 1-в-1 reference repositories.

### Пользователи и стейкхолдеры
- Основные пользователи:
  - конечный пользователь задачи, которому нужен быстрый actionable канал ответа в Telegram;
  - owner / product lead, которому нужен предсказуемый канал решения без ручного follow-up в GitHub comments;
  - platform operator / product designer, которому нужен первый канал с ясными guardrails по callback safety, correlation и operability.
- Стейкхолдеры:
  - `services/internal/control-plane` как owner platform interaction semantics, wait-state и audit/correlation;
  - `services/jobs/worker` как будущий owner dispatch/retry/delivery lifecycle;
  - `services/external/api-gateway` как thin-edge callback ingress;
  - reference repositories `telegram-approver` и `telegram-executor` как UX/stack baseline, но не как source of truth.
- Владелец решения: Owner.

### Продуктовые принципы и ограничения
- Telegram stream остаётся зависимым от typed interaction contract Sprint S10 и не переопределяет core platform semantics.
- Approval flow и user interaction flow остаются раздельными доменами даже внутри Telegram UX.
- MVP Telegram-канала ограничен actionable notify/decision flows и не включает richer conversation product.
- Inline callbacks и optional free-text нужны как быстрый response path, а не как повод утащить domain state в channel-specific transport.
- Delivery, retry, idempotency, audit trail, duplicate/replay classification и correlation остаются platform-owned responsibilities.
- В рамках `run:vision` разрешены только markdown-изменения.

## Scope boundaries

### MVP scope
- Telegram delivery path для `user.notify` как первого user-facing notification channel.
- Telegram delivery path для `user.decision.request` с 2-5 inline options.
- Приём callback-ответов по inline buttons.
- Optional free-text reply как controlled fallback/дополнение к fixed options.
- Базовая webhook/callback safety рамка:
  - correlation и callback authenticity expectations;
  - duplicate/replay/expired guardrails;
  - observability baseline для delivery success и fallback behavior.
- Handover в `run:prd` через issue `#448`.

### Post-MVP / deferred scope
- Voice/STT и любые voice-first сценарии.
- Rich conversation threads, advanced reminders и long-lived conversational UX.
- Multi-chat routing policy, multi-user assignment и escalation rules.
- Дополнительные каналы помимо Telegram.
- Любой redesign platform approval flow под видом Telegram scope.

### Dependency and sequencing gate
- Active vision stage в Issue `#447` допустим только пока Issue `#389` остаётся `closed` и продолжает использовать design package Issue `#387` как effective typed interaction contract baseline.
- Initial continuity issue `#444` сохраняется как исторический intake handover, но не считается источником текущего stage status и 2026-03-14 закрыта как `state:superseded`.
- Если Sprint S10 baseline будет reopened или superseded, Sprint S11 не должен двигаться в `run:prd` без явного обновления prerequisite.

## Success metrics

### North Star
| ID | Метрика | Определение | Источник | Целевое значение |
|---|---|---|---|---|
| `NSM-447-01` | Telegram actionable completion rate | Доля Telegram interaction-сценариев, где пользователь получает сообщение, даёт валидный ответ и не уходит в ручной GitHub fallback, а platform correlation остаётся корректной | interaction audit events + callback outcome review | `>= 80%` в pilot-сценариях MVP |

### Supporting metrics
| ID | Метрика | Определение/формула | Источник | Цель |
|---|---|---|---|---|
| `PM-447-01` | Decision turnaround p75 | p75 времени от отправки `user.decision.request` в Telegram до получения валидного ответа пользователя | interaction timestamps + callback audit | `<= 10 минут` |
| `UX-447-01` | GitHub fallback rate | Доля Telegram-eligible сценариев, где пользователю всё равно пришлось идти в GitHub comment или другой ручной канал | interaction audit + sample review | `<= 20%` |
| `REL-447-01` | Delivery success rate | Доля Telegram delivery attempts для MVP-сценариев, завершившихся без ручного operator вмешательства | delivery audit + retry classification | `>= 98%` |
| `REL-447-02` | Callback safety correctness | Доля duplicate/replay/stale callback scenarios, обработанных без потери correlation и без duplicate logical completion | callback audit + incident review | `>= 99.5%` |
| `GOV-447-01` | Platform semantics purity | Доля сценариев, где Telegram path не использует approval-only semantics и не требует Telegram-specific полей в core contract | product acceptance review + design gate | `100%` |

### Guardrails (ранние сигналы)
- `GR-447-01`: если `UX-447-01 > 25%`, следующий stage обязан приоритизировать clarity/fallback decisions выше channel expansion.
- `GR-447-02`: если `REL-447-01 < 95%`, дальнейшее расширение scope блокируется до пересмотра delivery/retry/operability assumptions.
- `GR-447-03`: если `REL-447-02 < 99%`, `run:arch` и `run:design` не могут уходить в richer UX раньше callback safety baseline.
- `GR-447-04`: если `GOV-447-01 < 100%`, stage переводится в `need:input` до устранения Telegram-first drift.
- `GR-447-05`: если future scope начинает включать voice/STT или rich threads в core MVP, требуется отдельное owner-решение перед переходом дальше.

## Risks and Product Assumptions
| Тип | ID | Описание | Митигирующее действие | Статус |
|---|---|---|---|---|
| risk | `RSK-447-01` | Telegram stream может расползтись в channel-first продукт и начать диктовать core interaction semantics | Держать sequencing gate S10 и purity guardrail обязательными до `run:design` | open |
| risk | `RSK-447-02` | Scope может вырасти в voice/STT и richer conversation flows раньше подтверждения базовой ценности канала | Жёстко отделить MVP от deferred scope уже на vision/PRD | open |
| risk | `RSK-447-03` | Прямое копирование reference repositories принесёт чужие service boundaries и governance assumptions | Использовать reference stack только как baseline и зафиксировать product/domain ownership внутри `codex-k8s` | open |
| risk | `RSK-447-04` | Callback path окажется user-friendly, но операционно хрупким из-за duplicate/replay/expired сценариев | Держать callback safety metrics и guardrails как blocking criteria для следующих stage | open |
| assumption | `ASM-447-01` | Notify + decision request + inline callbacks + optional free-text достаточно, чтобы подтвердить ценность первого внешнего канала | Проверить user stories и expected evidence на `run:prd` | accepted |
| assumption | `ASM-447-02` | Пользовательская ценность Telegram возникает быстрее, чем у GitHub comment-only path, но без потери platform control | Подтвердить turnaround/fallback metrics и product guardrails на `run:prd` | accepted |
| assumption | `ASM-447-03` | Channel-neutral semantics Sprint S10 можно сохранить, даже если первый channel adapter будет именно Telegram | Зафиксировать contract boundaries и adapter ownership на `run:arch`/`run:design` | accepted |

## Readiness criteria для `run:prd`
- [x] Mission, north star и persona outcomes сформулированы для end user, owner/product lead и platform operator.
- [x] KPI/success metrics и guardrails определены как измеримые product/operational сигналы.
- [x] MVP scope, deferred scope и sequencing gate относительно Sprint S10 разделены явно.
- [x] Подтверждено, что Telegram остаётся channel-specific adapter stream, а не Telegram-first перепроектированием platform interaction-domain.
- [x] Создана отдельная issue следующего этапа `run:prd` (`#448`) без trigger-лейбла и с continuity-требованием продолжить цепочку до `run:dev`.

## Acceptance criteria (Issue #447)
- [x] Mission, north star и продуктовые принципы для Telegram-адаптера сформулированы явно.
- [x] KPI/success metrics и guardrails зафиксированы для turnaround, fallback, delivery success, callback safety и purity platform semantics.
- [x] Персоны, MVP/Post-MVP границы, риски и assumptions описаны для Telegram-канала как отдельного stream.
- [x] Сохранён проверяемый sequencing gate `#389 closed -> #387 baseline effective` для active vision stage.
- [x] Подготовлен handover в `run:prd` и создана follow-up issue `#448` без trigger-лейбла.

## Handover в следующий этап
- Следующий stage: `run:prd`.
- Follow-up issue: `#448`.
- Trigger-лейбл `run:prd` на issue `#448` ставит Owner.
- Initial continuity issue `#444` остаётся только historical handover artifact от intake-stage, 2026-03-14 закрыта как `state:superseded` и не заменяет active stage issue `#447`.
- Обязательное continuity-требование для `#448`:
  - в конце PRD stage агент обязан создать issue для `run:arch`;
  - в body этой issue должно быть явно повторено требование продолжить цепочку `arch -> design -> plan -> dev` без разрывов.
- На `run:prd` нельзя потерять следующие решения vision:
  - Telegram остаётся первым внешним каналом поверх platform-owned interaction contract;
  - MVP ограничен notify/decision/callback/free-text path;
  - approval flow и user interaction flow не смешиваются;
  - callback safety и platform semantics purity важнее channel expansion;
  - voice/STT, reminders, rich threads и дополнительные каналы остаются deferred scope.

## Связанные документы
- `docs/delivery/sprints/s11/sprint_s11_telegram_user_interaction_adapter.md`
- `docs/delivery/epics/s11/epic_s11.md`
- `docs/delivery/epics/s11/epic-s11-day1-telegram-user-interaction-adapter-intake.md`
- `docs/delivery/traceability/s11_telegram_user_interaction_adapter_history.md`
- `docs/delivery/issue_map.md`
- `docs/delivery/delivery_plan.md`
- `docs/product/requirements_machine_driven.md`
- `docs/product/constraints.md`
- `docs/product/agents_operating_model.md`
- `docs/product/labels_and_trigger_policy.md`
- `docs/product/stage_process_model.md`
- `docs/architecture/mcp_approval_and_audit_flow.md`
- `docs/architecture/api_contract.md`
- `docs/research/src_idea-machine_driven_company_requirements.md`
- `services/internal/control-plane/README.md`
- `services/jobs/worker/README.md`
- `services/external/api-gateway/README.md`
