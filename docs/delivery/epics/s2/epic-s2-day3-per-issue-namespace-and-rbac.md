---
doc_id: EPC-CK8S-S2-D3
type: epic
title: "Epic S2 Day 3: Per-issue namespace orchestration and RBAC baseline"
status: completed
owner_role: EM
created_at: 2026-02-10
updated_at: 2026-02-11
related_issues: []
related_prs: [9]
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-19-full-docset"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-19
---

# Epic S2 Day 3: Per-issue namespace orchestration and RBAC baseline

## TL;DR
- Цель эпика: исполнять dev/revise runs в изолированном namespace с доступом к нужному стеку.
- Ключевая ценность: воспроизводимость, изоляция и управляемость прав.
- MVP-результат: для каждого run создаётся namespace (или выбирается пул), в нём запускается агентный Job.

## Priority
- `P0`.

## Scope
### In scope
- Создание namespace по шаблону имени (например, `codex-issue-<id>` или `codex-run-<run_id>`).
- Создание/применение RBAC для агентного service account (минимально необходимые права).
- Поддержка mixed runtime policy:
  - `full-env` для ролей/профилей, где нужен доступ к runtime;
  - `code-only` профили без k8s runtime доступа.
- Политики ресурсов: quotas/limits (минимальный baseline).
- Запись lifecycle событий namespace/job в БД (audit/flow_events).

### Out of scope
- Продвинутая network policy матрица (будет отдельным hardening эпиком).

## Критерии приемки эпика
- Run исполняется в отдельном namespace.
- Namespace может быть безопасно убран/переиспользован без утечек слотов и объектов.

## Прогресс реализации (2026-02-11)
- Реализована runtime-классификация run по режимам:
  - `full-env` для issue-trigger `run:dev`/`run:dev:revise`;
  - `code-only` для остальных run без issue-trigger контекста.
- Для `full-env` реализована подготовка отдельного run namespace:
  - namespace naming: issue-aware шаблон с суффиксом run-id (deterministic, без коллизий);
  - idempotent apply baseline ресурсов:
    - `ServiceAccount`,
    - `Role`,
    - `RoleBinding`,
    - `ResourceQuota`,
    - `LimitRange`.
- Worker запускает Job в целевом namespace и передаёт runtime metadata в env/payload.
- Добавлен cleanup baseline:
  - по завершении `full-env` run namespace удаляется (управляемо через env-флаг cleanup);
  - в S2 baseline поддерживался legacy manual-retention label: cleanup пропускался, namespace сохранялся для отладки и фиксировался в `flow_events` (позже удалено как избыточное поведение).
  - удаляются только managed namespace’ы, промаркированные worker’ом.
- Для runtime metadata закреплён доменный префикс:
  - labels/annotations в namespace/job используют `kodex.works/*`.
- Добавлен audit lifecycle в `flow_events`:
  - `run.namespace.prepared`,
  - `run.namespace.cleaned`,
  - `run.namespace.cleanup_failed`,
  - `run.namespace.cleanup_skipped` (например, при legacy manual-retention режиме в S2).
- Для reconciliation running runs расширено чтение `agent_runs.run_payload`, чтобы namespace/runtime mode определялись детерминированно и после рестартов worker.
- Deploy baseline обновлён:
  - worker получил cluster-scope RBAC для lifecycle namespace и runtime-объектов;
  - добавлены env/vars для namespace policy и quota/limitrange baseline в bootstrap/deploy/CI.

## Evidence
- Runtime namespace orchestration и cleanup:
  - `libs/go/k8s/joblauncher/runtime_namespace.go`
  - `libs/go/k8s/joblauncher/metadata.go`
  - `services/jobs/worker/internal/domain/worker/run_runtime.go`
  - `services/jobs/worker/internal/domain/worker/service.go`
- Worker runtime contracts:
  - `services/jobs/worker/internal/domain/worker/launcher.go`
  - `services/jobs/worker/internal/clients/kubernetes/launcher/adapter.go`
- Runtime policy env wiring:
  - `services/jobs/worker/internal/app/config.go`
  - `deploy/base/kodex/codegen-check-job.yaml.tpl`
  - `services/internal/control-plane/internal/domain/runtimedeploy/service_defaults.go`
- Production runbook checks:
  - `docs/ops/production_runbook.md`

## Verification
- Unit tests:
  - `go test ./libs/go/k8s/joblauncher ./services/jobs/worker/...`
- Static checks:
  - `make lint-go`
  - `make dupl-go`
- Production:
  - `AI Production deploy 🚀` success для `codex/dev` (manual dispatch на целевой SHA).
  - ручной smoke/regression по runbook -> `OK`.

## Апрув
- request_id: owner-2026-02-11-s2-day3
- Решение: approved
- Комментарий: Day 3 scope принят; per-issue namespace/RBAC/resource policy baseline закреплён.
