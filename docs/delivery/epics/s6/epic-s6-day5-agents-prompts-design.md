---
doc_id: EPC-CK8S-S6-D5
type: epic
title: "Epic S6 Day 5: Design для lifecycle управления агентами и шаблонами промптов (Issue #195)"
status: in-review
owner_role: SA
created_at: 2026-02-25
updated_at: 2026-02-25
related_issues: [184, 185, 187, 189, 195, 197]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-25-issue-195-design-epic"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-25
---

# Epic S6 Day 5: Design для lifecycle управления агентами и шаблонами промптов (Issue #195)

## TL;DR
- Подготовлен полный design package для handover в `run:plan`:
  - detailed design,
  - API contracts,
  - data model,
  - migrations policy.
- Зафиксированы transport boundaries `api-gateway <-> control-plane`, error/validator/concurrency контракты и UI state-flow для `diff/preview/history`.
- На этапе `run:design` runtime/code не изменялись (markdown-only).

## Контекст
- Stage continuity: `#184 -> #185 -> #187 -> #189 -> #195 -> #197`.
- Входной baseline: архитектурный пакет Day4 (`ADR-0009`, `ALT-0001`, `agents_prompt_templates_lifecycle_design.md`).
- Цель Day5: превратить архитектурный baseline в implementation-ready contract package.

## Артефакты Day5
- `docs/architecture/agents_prompt_templates_lifecycle_design_doc.md`
- `docs/architecture/agents_prompt_templates_lifecycle_api_contract.md`
- `docs/architecture/agents_prompt_templates_lifecycle_data_model.md`
- `docs/architecture/agents_prompt_templates_lifecycle_migrations_policy.md`

## Ключевые design-решения
- API boundary:
  - staff HTTP контур фиксирован через contract-first OpenAPI (обновление в `run:dev`);
  - internal gRPC контур фиксирован через typed RPC/DTO/casters.
- Seed/fallback strategy:
  - текущие embed markdown seeds остаются baseline/fallback источником;
  - перенос в БД выполняется управляемо через `dry-run/apply` seed sync без перезаписи project overrides.
- Concurrency:
  - все mutating template operations используют `expected_version`;
  - 409 `conflict` возвращает `actual_version` + `latest_checksum`.
- Data model:
  - сохраняется ADR-0009 (история в `prompt_templates`, audit в `flow_events`);
  - вводятся поля lifecycle-state/checksum/audit metadata и partial unique index на active version.
- UI flow:
  - list -> details -> draft edit -> preview/diff -> activate -> history.

## Dependency baseline (Context7)
- Проверен `kin-openapi` (`/getkin/kin-openapi`): request/response validation path достаточен для staff contract-first.
- Проверен `monaco-editor` (`/microsoft/monaco-editor`): `DiffEditor` полностью покрывает compare versions use-case.
- Новые внешние зависимости не требуются.

## Runtime и migration impact
- Runtime impact: отсутствует на этапе Day5 (markdown-only change set).
- Migration impact (для `run:dev`):
  - DDL расширения `prompt_templates`;
  - backfill + индексация;
  - rollout order: `migrations -> internal -> edge -> frontend`.

## Acceptance criteria status (Issue #195)
- [x] Подготовлен design package (`design_doc`, `api_contract`, `data_model`, `migrations_policy`).
- [x] Зафиксированы контракты ошибок/валидаторов/конкурентности.
- [x] Описаны runtime/migration impacts.
- [x] Обновлена трассируемость (`issue_map`, `requirements_traceability`).

## Handover в `run:plan`
- Следующий этап должен декомпозировать реализацию на execution issues:
  - OpenAPI/proto updates + codegen;
  - control-plane domain/repository implementation;
  - api-gateway transport updates;
  - web-console state/UI integration;
  - test/observability gates.
- Follow-up issue `#197` создана без trigger-лейбла, запуск stage выполняется Owner.
- По завершении `run:plan` обязательно создать issue следующего этапа `run:dev`.
