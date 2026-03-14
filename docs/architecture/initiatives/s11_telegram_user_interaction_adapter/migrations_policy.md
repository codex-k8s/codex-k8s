---
doc_id: MIG-S11-CK8S-0001
type: migrations-policy
title: "Sprint S11 Day 5 — Migrations policy for Telegram user interaction adapter (Issue #454)"
status: approved
owner_role: SA
created_at: 2026-03-14
updated_at: 2026-03-14
related_issues: [361, 444, 447, 448, 452, 454, 456, 458]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-03-14-issue-454-migrations-policy"
  approved_by: "ai-da-stas"
  approved_at: 2026-03-14
---

# DB Migrations Policy: Sprint S11 Telegram user interaction adapter

## TL;DR
- Подход: additive expand-enable with explicit prerequisite on Sprint S10 interaction foundation.
- Владелец схемы/миграций: `services/internal/control-plane`.
- Миграции живут в `services/internal/control-plane/cmd/cli/migrations/*.sql`.
- Rollback ограничен после начала Telegram traffic: уже доставленные сообщения, callback evidence и accepted decisions не откатываются schema rollback-ом.

## Размещение миграций и владелец схемы
- Schema owner: `services/internal/control-plane`.
- Telegram interaction extension does not introduce a new DB owner.
- `api-gateway`, `worker` и Telegram adapter contour не получают своих migration paths.

## Пререквизит Sprint S10
- S11 migrations не могут стартовать до применения Sprint S10 interaction foundation:
  - generic interaction tables;
  - typed wait linkage in `agent_runs`;
  - resume payload persistence path.
- Если S10 foundation не deployed:
  - S11 rollout block;
  - no partial Telegram-only schema branch.

## Потоки и миграционная обязательность
| Stream | Нужна миграция | Политика |
|---|---|---|
| `interaction_channel_bindings` | да | новая таблица |
| `interaction_callback_handles` | да | новая таблица + unique hash index |
| `interaction_requests` extension fields | да | additive columns |
| `interaction_delivery_attempts` extension fields | да | additive columns |
| `interaction_callback_events` extension fields | да | additive columns |
| `interaction_response_records` optional Telegram fields | да | additive columns |
| `agent_runs` | нет | reuse S10 linkage |

## Принципы
- Expand only:
  - Telegram tables/columns add data; existing interaction rows remain valid.
- Handle secrecy:
  - only hashes are persisted; no raw callback handle backfill is needed.
- Writes last:
  - callback ingress and continuation writes enable only after schema + owner services are ready.
- Notify before decision:
  - rollout may expose notify-only Telegram path before decision callbacks/free-text.
- Continuation isolation:
  - edit/follow-up attempts must be disabled independently if Bot API behavior is unstable.

## Процесс миграции (run:dev target)
1. Verify S10 prerequisite:
   - confirm interaction foundation migrations already applied;
   - confirm no unknown wait linkage state remains.
2. Expand schema:
   - create `interaction_channel_bindings`;
   - create `interaction_callback_handles`;
   - add Telegram extension columns to `interaction_requests`;
   - add `delivery_role`, `channel_binding_id`, provider ref snapshot columns to `interaction_delivery_attempts`;
   - add Telegram evidence columns to `interaction_callback_events`;
   - optionally add `channel_binding_id`, `handle_kind` to `interaction_response_records`.
3. Index hardening:
   - unique `handle_hash`;
   - operator visibility index on `interaction_requests`;
   - continuation queue indexes;
   - partial uniqueness for provider message refs where safe.
4. Enable owner writes:
   - rollout `control-plane` with channel binding + handle allocation logic;
   - rollout `worker` with Telegram delivery/continuation logic.
5. Enable callback ingress:
   - rollout `api-gateway`;
   - enable normalized adapter callback admission.
6. Enable Telegram adapter traffic:
   - first notify-only;
   - then decision inline callbacks;
   - then free-text session flow;
   - finally edit-in-place continuation if metrics remain healthy.

## Как выполняются миграции при деплое
- Mandatory order:
  1. stateful dependencies ready
  2. S10 prerequisite confirmed
  3. S11 migration job
  4. `control-plane`
  5. `worker`
  6. `api-gateway`
  7. Telegram adapter contour
- Concurrency control:
  - single migration runner under `goose` advisory lock.
- Failure policy:
  - if migration or S10 prerequisite check fails, Telegram routes and adapter rollout remain disabled.

## Политика backfill
- No raw handle/token backfill:
  - all callback handles are generated only for new Telegram interactions.
- Extension columns backfill:
  - existing S10 interactions default to `channel_family=platform_only`, `operator_state=nominal`.
- Notify-first rollout:
  - no backfill of old interactions into Telegram bindings.
- Restart safety:
  - repeated migration or partial rollout does not create second binding for the same active interaction.

## Политика rollback
- Safe rollback before adapter traffic:
  - keep additive schema;
  - disable Telegram path in service config and MCP exposure.
- Limited rollback after notify-only traffic:
  - stop new Telegram deliveries;
  - preserve `interaction_channel_bindings` and attempt ledger.
- Limited rollback after decision traffic:
  - stop new decision requests and callback ingress;
  - keep callback evidence, effective responses and operator signals;
  - do not attempt to retract already delivered Telegram messages or accepted decisions.
- Continuation-specific rollback:
  - disable `message_edit` attempts while keeping follow-up notify;
  - if follow-up notify also unstable, force `manual_fallback_required`.

## Что нельзя безопасно откатить
- Already accepted user decisions and corresponding run resumes.
- Already delivered Telegram messages and follow-up notifications.
- Callback evidence, operator fallback history and provider refs written for audit.

## Проверки
### Pre-migration checks
- Sprint S10 interaction foundation deployed and healthy.
- No conflicting custom tables/columns with planned S11 names.
- `interaction_requests.channel_family` and `operator_state` defaults validated on staging/candidate data.
- Telegram adapter readiness confirmed for webhook secret token and callback bearer flow.

### Post-migration verification
- Notify-only Telegram interaction creates channel binding and primary delivery attempt without wait-state.
- Decision interaction creates callback handle hashes and valid token metadata.
- Duplicate callback does not create second effective response.
- Expired handle returns `classification=expired` inside grace window.
- `message_edit` failure can schedule `follow_up_notify`.
- Manual fallback queue is queryable from persisted operator state.

## Runtime impact / Migration impact
- Runtime impact (`run:design`): none.
- Migration impact (`run:dev`): moderate, additive extension over S10 foundation with new tables, columns and indexes.

## Operational notes
- If adapter webhook path is unstable, keep `api-gateway` callback endpoint deployed but gated from production traffic until adapter recovers.
- If edit continuation is unstable, keep decision callbacks enabled and downgrade continuation policy to `follow_up_only`.
- If follow-up delivery also degrades, surface `manual_fallback_required` instead of silently retrying forever.

## Continuity after `run:plan`
- [x] Plan package Issue `#456` подтвердил rollout order `migrations -> control-plane -> worker -> api-gateway -> Telegram adapter contour -> observability/evidence gate`.
- [x] S10 interaction foundation остаётся hard prerequisite для execution anchor `#458`.
- [x] Notify-first enablement, `follow_up_only` downgrade и manual fallback сохраняются как допустимые operational toggles, а не как отдельные schema branches.
- [x] Additive migration path остаётся единственным baseline; parallel DB owner или destructive rollback не допускаются.
