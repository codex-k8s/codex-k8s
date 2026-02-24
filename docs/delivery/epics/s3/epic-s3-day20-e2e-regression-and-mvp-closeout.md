---
doc_id: EPC-CK8S-S3-D20
type: epic
title: "Epic S3 Day 20: Full e2e regression gate and MVP closeout"
status: planned
owner_role: EM
created_at: 2026-02-18
updated_at: 2026-02-24
related_issues: [19, 112, 128]
related_prs: [129]
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-19-full-docset"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-19
---

# Epic S3 Day 20: Full e2e regression gate and MVP closeout

## TL;DR
- Цель: выполнить финальный full e2e после закрытия core-flow и frontend hardening, собрать evidence bundle и завершить MVP formal handover.
- Результат: подтверждён end-to-end контур (bootstrap -> deploy -> run lifecycle -> governance/UI), сформирован go/no-go пакет для Owner.
- Детальный сценарный план и матрица label coverage вынесены в `docs/delivery/e2e_mvp_master_plan.md`.

## Priority
- `P0`.

## Scope
### In scope
- Full e2e на чистом VPS (Ubuntu 24.04):
  - bootstrap с `bootstrap/host/config-e2e-test.env`,
  - проверка self-deploy контура,
  - cross-project сценарий (`project-example` + `codex-k8s`).
- E2E regression по ключевым MVP сценариям:
  - stage labels flow,
  - intake/revise matrix checkpoint (`run:intake` + `run:intake:revise`) по Issue #128,
  - self-improve loop,
  - MCP governance tools,
  - repo onboarding + docset/config governance,
  - runtime deploy/build/image management.
- Security/reliability checks:
  - secret leakage safeguards,
  - approvals/RBAC,
  - retries/idempotency/cleanup/resume.
- Consolidated closeout:
  - evidence bundle,
  - release notes + runbook updates,
  - formal owner sign-off (go/no-go).

### Out of scope
- Post-MVP feature implementation.
- Полноценный external pentest.

## Декомпозиция
- Story-1: подготовка e2e окружения и preflight.
- Story-2: прогон e2e smoke + regression сценариев.
- Story-3: security/reliability checks + устранение найденных блокеров.
- Story-4: финальный handover пакет и owner sign-off.

## Execution package
- Канонический test-plan: `docs/delivery/e2e_mvp_master_plan.md`.
- Label coverage: `run:*`, `run:*:revise`, `state:*`, `need:*`, `[ai-model-*]`, `[ai-reasoning-*]`.
- Обязательные зоны: review-driven revise, MCP governance tools, runtime TTL/revise reuse, security/RBAC, docs/traceability consistency.

## Критерии приемки
- Full e2e проходит без P0 блокеров.
- E2E-checkpoint по Issue #128 для `run:intake` и `run:intake:revise` подтверждён и отражён в traceability docs.
- Собран и опубликован evidence bundle (команды, логи, ссылки на run/deploy artifacts).
- Owner получает go/no-go пакет и подтверждает завершение MVP фазы.

## Execution evidence (Issue #128 intake revise)
- PR: `https://github.com/codex-k8s/codex-k8s/pull/129`.
- Run `aa46a919-d6c4-4512-a7e1-17e088a3a5c2`: intake baseline (Issue #128).
- Run `d2fe5bdb-6c02-460a-92d4-7583b3ad4787`: intake revise loop (Issue #128).
- Документы-трассировки: `docs/delivery/e2e_mvp_master_plan.md`, `docs/delivery/issue_map.md`.

## Риски/зависимости
- Зависимость от готовности чистого VPS и валидного e2e config.
- Риск обнаружения late-stage блокеров: требуется резерв по времени на hotfix цикл.
