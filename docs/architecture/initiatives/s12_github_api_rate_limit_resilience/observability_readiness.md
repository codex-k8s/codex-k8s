---
doc_id: OBS-S12-CK8S-0001
type: readiness-handover
title: "Sprint S12 Day 7 — Observability and rollout readiness for GitHub API rate-limit resilience (Issue #431)"
status: in-review
owner_role: KM
created_at: 2026-03-15
updated_at: 2026-03-15
related_issues: [366, 413, 416, 418, 420, 423, 425, 426, 427, 428, 429, 430, 431, 500]
related_prs: []
related_adrs: ["ADR-0013"]
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-15-issue-431-observability-readiness"
---

# Observability and Rollout Readiness: Sprint S12 GitHub API rate-limit resilience

## TL;DR
- Wave `S12-E07` формализует readiness gate перед `run:qa`: rollout order, typed evidence surfaces, candidate runtime checks и rollback discipline собраны в один handover-артефакт.
- В текущем candidate namespace `kodex-dev-1` подтверждены готовые deploy/job resources для rollout order `migrations -> control-plane -> worker -> agent-runner -> api-gateway -> web-console`.
- Live GitHub rate-limit path в этом doc-audit run не воспроизводился: исторический baseline Issue `#431` фиксировал disabled rollout через пустой env-gate; Issue `#500` позже перенёс source of truth в DB-backed platform setting `github_rate_limit_wait_enabled` и убрал deploy env wiring.

## Scope

### Что покрыто этим readiness-пакетом
- canonical audit/evidence keys для GitHub rate-limit wait lifecycle;
- runtime commands для candidate validation через `kubectl`, `/metrics`, логи и SQL;
- rollout order и rollback notes для уже реализованных waves `#425..#430`;
- явные ограничения текущего readiness baseline перед `run:qa`.

### Что осознанно не утверждается этим документом
- synthetic trigger GitHub primary/secondary rate-limit в candidate namespace;
- отдельные `kodex_github_rate_limit_*` Prometheus series: по repo audit и candidate runtime такие метрики не обнаружены;
- production alert rules, dashboards и manifest-изменения сверх уже принятых волн `#425..#430`.

## Canonical Evidence Surfaces

| Поверхность | Что проверяем | Source of truth |
|---|---|---|
| Wait-state в run status | `waiting_backpressure`, `wait_reason=github_rate_limit`, `wait_target_kind=github_rate_limit_wait` | `agent_runs`, `agent_sessions`, typed `Run.wait_projection` |
| Domain evidence ledger | `signal_detected`, `classified`, `resume_scheduled`, `resume_attempted`, `resume_failed`, `resolved`, `manual_action_required`, `comment_mirror_failed` | `github_rate_limit_wait_evidence` |
| Flow events | `run.wait.paused`, `run.wait.resumed` с event keys `github_rate_limit.wait.entered`, `github_rate_limit.manual_action_required`, `github_rate_limit.resume_succeeded` | `flow_events` |
| Runner logs | typed handoff/resume markers без повторного derive semantics из stderr/headers | `agent-runner` logs |
| Worker logs | bounded wait processing и escalation path | `worker` logs |
| Transport/UI visibility | realtime envelopes `wait_entered`, `wait_updated`, `wait_resolved`, `wait_manual_action_required`; wait queue и run details | `api-gateway` + `web-console` |

## Candidate Findings (2026-03-15)
- `kubectl config view --minify -o jsonpath='{..namespace}'` вернул `kodex-dev-1`.
- `kubectl get deploy,pods,job -n kodex-dev-1 -o wide` подтвердил:
  - `deployment/kodex-control-plane`, `deployment/kodex-worker`, `deployment/kodex`, `deployment/kodex-web-console` готовы;
  - `job/kodex-migrate` завершён;
  - kaniko build jobs и `repo-sync` завершены;
  - agent run job для текущего doc-audit остаётся активным.
- `kubectl logs -n kodex-dev-1 deploy/kodex-control-plane --tail=120 | rg 'github rate-limit|wait.entered|wait.resumed|manual_action_required|waiting_backpressure'` не вернул совпадений в пределах последних 120 строк.
- Исторический baseline Issue `#431` по env-переменным сохраняется как evidence для doc-audit run; после Issue `#500` этот check больше не является current source of truth, потому что effective gate хранится в `system_settings.github_rate_limit_wait_enabled`, а `KODEX_GITHUB_RATE_LIMIT_WAIT_ENABLED` удалён из bootstrap/deploy wiring.

## Runtime Diagnostics

### `kubectl` / health / metrics
1. Проверить rollout resources:
   - `kubectl -n <candidate-namespace> get deploy,pods,job -o wide`
2. Проверить availability service endpoints:
   - `kubectl -n <candidate-namespace> port-forward deploy/kodex-control-plane 18081:8081`
   - `curl -sf http://127.0.0.1:18081/health/readyz`
   - `curl -sf http://127.0.0.1:18081/metrics >/dev/null`
   - `kubectl -n <candidate-namespace> port-forward deploy/kodex-worker 18082:8082`
   - `curl -sf http://127.0.0.1:18082/health/readyz`
   - `curl -sf http://127.0.0.1:18082/metrics >/dev/null`
   - `kubectl -n <candidate-namespace> port-forward deploy/kodex 18080:8080`
   - `curl -sf http://127.0.0.1:18080/metrics >/dev/null`
3. Важное ограничение:
   - readiness gate сейчас проверяет доступность `/metrics`, а не наличие отдельных `github_rate_limit` series; выделенные Prometheus-метрики для этого capability в коде не найдены.

### Логи
- `kubectl -n <candidate-namespace> logs deploy/kodex-control-plane --tail=200 | rg 'github rate-limit|run.wait.(paused|resumed)|manual_action_required'`
- `kubectl -n <candidate-namespace> logs deploy/kodex-worker --tail=200 | rg 'github rate-limit wait processed|manual_action_required|resume'`
- `kubectl -n <candidate-namespace> logs job/<agent-runner-job> --tail=200 | rg 'github rate-limit handoff|waiting_backpressure|resume payload'`

### SQL / psql evidence
- Открытые waits и dominant election:

```sql
SELECT
    id,
    run_id,
    contour_kind,
    state,
    dominant_for_run,
    resume_not_before,
    auto_resume_attempt_no,
    auto_resume_budget
FROM github_rate_limit_waits
WHERE state IN ('open', 'auto_resume_scheduled', 'auto_resume_in_progress', 'manual_action_required')
ORDER BY updated_at DESC;
```

- Evidence trail по wait:

```sql
SELECT
    wait_id,
    event_kind,
    signal_origin,
    observed_at
FROM github_rate_limit_wait_evidence
ORDER BY observed_at DESC, id DESC
LIMIT 50;
```

- Runs в coarse wait-state:

```sql
SELECT
    id,
    status,
    wait_reason,
    wait_target_kind,
    started_at,
    finished_at
FROM agent_runs
WHERE status = 'waiting_backpressure'
   OR wait_reason = 'github_rate_limit'
ORDER BY started_at DESC;
```

## Rollout Discipline

### Обязательная последовательность
1. `migrations`
2. `control-plane`
3. `worker`
4. `agent-runner`
5. `api-gateway`
6. `web-console`
7. readiness evidence gate

### Что проверяем после каждого шага
- После `migrations`:
  - существуют таблицы `github_rate_limit_waits`, `github_rate_limit_wait_evidence`;
  - additive enum/check expansion для `agent_runs` и `agent_sessions` применена.
- После `control-plane`:
  - работают classification/read-projection paths;
  - flow events `run.wait.paused` / `run.wait.resumed` имеют корректные event keys;
  - feature flag wiring не противоречит rollout guard.
- После `worker`:
  - sweep loop доступен и читает persisted wait aggregate;
  - escalation остаётся в `manual_action_required`, а не в silent retry.
- После `agent-runner`:
  - лог path фиксирует typed handoff и `waiting_backpressure`;
  - resume path использует persisted `github_rate_limit_resume_payload`.
- После `api-gateway`:
  - typed `wait_projection` и realtime envelope доступны без переноса domain semantics в handlers.
- После `web-console`:
  - wait queue и run details рендерят dominant/related waits и manual guidance из typed DTO.
- Перед `run:qa`:
  - есть синхронизированная traceability;
  - принято owner-решение по включению platform setting `github_rate_limit_wait_enabled` для live smoke или явному сохранению disabled-mode.

## Rollback / Mitigation
- First stop:
  - не включать `system_settings.github_rate_limit_wait_enabled`, пока rollout order и read surfaces не подтверждены.
- Rollback order:
  1. `web-console`
  2. `api-gateway`
  3. `agent-runner`
  4. `worker`
  5. `control-plane`
- Schema policy:
  - additive DDL из `#425` не откатывается destructive-операциями; исправления только forward-fix.
- Feature gate:
  - при runtime drift достаточно вернуть `system_settings.github_rate_limit_wait_enabled=false`, сохранив read compatibility и исторический evidence trail.

## Acceptance Evidence Before `run:qa`
- [x] Rollout order `migrations -> control-plane -> worker -> agent-runner -> api-gateway -> web-console -> evidence gate` собран в одном source-of-truth документе.
- [x] Typed evidence surfaces перечислены: wait projection, flow events, evidence table, runner/worker logs, realtime envelopes.
- [x] Candidate namespace проверен на наличие готовых deploy/job ресурсов для rollout baseline.
- [ ] Live GitHub rate-limit smoke в candidate namespace выполнен.
  Причина: в текущем doc-audit run feature gate остаётся выключенным по default config, synthetic trigger не запускался.
- [ ] Отдельные GitHub rate-limit Prometheus series подтверждены.
  Причина: по repo audit и candidate runtime отдельные `github_rate_limit` метрики не найдены; текущий readiness gate опирается на `/metrics` availability, flow events, SQL и typed logs/UI evidence.

## Related Documents
- `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/design_doc.md`
- `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/api_contract.md`
- `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/data_model.md`
- `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/migrations_policy.md`
- `docs/delivery/epics/s12/epic-s12-day6-github-api-rate-limit-plan.md`
- `docs/delivery/sprints/s12/sprint_s12_github_api_rate_limit_resilience.md`
