---
doc_id: EPC-CK8S-S8-D3-EXISTING-REPO
type: epic
title: "Epic S8 Day 3: Existing repository adoption on onboarding (scan -> docs/services.yaml PR) (Issue 282)"
status: planned
owner_role: EM
created_at: 2026-03-06
updated_at: 2026-03-06
related_issues: [282]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-06-s8-existing-repo-adoption"
---

# Epic S8 Day 3: Existing repository adoption on onboarding (scan -> docs/services.yaml PR)

## TL;DR
- Проблема: платформа умеет подключить существующий репозиторий и проверить доступы, но не умеет перевести уже живой кодовый репозиторий в состояние `codex-k8s-ready`.
- Цель: сделать guided adoption flow для существующего репозитория с кодом, который ещё не содержит `services.yaml`, role-aware docs baseline и metadata для платформы.
- Результат: controlled onboarding pipeline `scan -> onboarding plan -> dedicated agent/task -> PR`, чтобы человек принимал результат через review, а не настраивал репозиторий вручную.

## Priority
- `P0`.

## Контекст и проблема
- FR-020 допускает per-repo `services.yaml`, а FR-050 требует role-aware docs context для агентов.
- S3 Day14 решил только preflight attach/checks, но не решил adoption существующего кода.
- Для уже живого репозитория direct write в default branch недопустим:
  - там уже есть код и потенциальная эксплуатационная ценность;
  - изменение `services.yaml`/docs должно проходить через reviewable PR.
- Простого запуска generic `run:dev` недостаточно:
  - нужен детерминированный onboarding context;
  - нужен безопасный scope (без произвольной переделки приложения);
  - нужен отчёт, какие сигналы были считаны из репозитория и почему предложен именно такой `services.yaml`.

## Scope
### In scope
- Детект сценария `existing repository not onboarded`:
  - в репозитории есть код/директории/история;
  - отсутствует `services.yaml` или platform docs baseline;
  - либо эти артефакты присутствуют, но невалидны/явно не соответствуют typed contract.
- Adoption pipeline из двух фаз:
  1. deterministic scanner;
  2. dedicated onboarding generation task с фиксированным prompt/policy.
- Scanner должен собрать onboarding context:
  - языки/стек;
  - признаки deployable components;
  - наличие Dockerfile, manifests, CI, env examples;
  - текущее состояние docs;
  - список candidate services/components для `services.yaml`.
- Generation task должен:
  - подготовить draft `services.yaml`;
  - добавить или актуализировать полный template-catalog `docs/templates/*.md`;
  - подготовить docs baseline/расширение текущей документации;
  - не переписывать прикладной код, если это не требуется для bootstrap платформы;
  - открыть PR с explanation bundle.
- Template-catalog, который должен добавляться в onboarding PR при его отсутствии, обязан включать весь текущий набор `docs/templates/*.md` платформы:
  - индекс: `index.md`;
  - product templates: `problem.md`, `brief.md`, `constraints.md`, `scope_mvp.md`, `personas.md`, `risks_register.md`, `project_charter.md`, `success_metrics.md`, `prd.md`, `nfr.md`, `user_story.md`;
  - architecture templates: `c4_context.md`, `c4_container.md`, `adr.md`, `alternatives.md`, `design_doc.md`, `api_contract.md`, `data_model.md`, `migrations_policy.md`;
  - delivery templates: `delivery_plan.md`, `epic.md`, `definition_of_done.md`, `issue_map.md`, `roadmap.md`, `docset_issue.md`, `docset_pr.md`;
  - quality/release/ops templates: `test_strategy.md`, `test_plan.md`, `test_matrix.md`, `regression_checklist.md`, `release_plan.md`, `release_notes.md`, `rollback_plan.md`, `postdeploy_review.md`, `runbook.md`, `monitoring.md`, `alerts.md`, `slo.md`, `incident_playbook.md`, `incident_postmortem.md`.
- Стартовый docs baseline, который должен формироваться в PR поверх template-catalog, включает как минимум:
  - `docs/README.md`;
  - `docs/product/problem.md`;
  - `docs/product/brief.md`;
  - `docs/product/constraints.md`;
  - `docs/product/scope_mvp.md`;
  - `docs/product/personas.md`;
  - `docs/product/risks_register.md`;
  - `docs/architecture/README.md`;
  - `docs/delivery/README.md`;
  - `docs/ops/README.md`.
- Explanation bundle в PR должен включать:
  - что платформа обнаружила в репозитории;
  - какие файлы добавила/изменила;
  - какие места требуют ручной проверки человеком;
  - какие допущения были сделаны при генерации.
- UI/API:
  - отдельный action `Adopt existing repository`;
  - просмотр scan report до запуска generation;
  - статус adoption pipeline и ссылка на PR.
- Idempotency и repair:
  - повторный запуск на том же ref обновляет существующую onboarding-ветку/PR, а не плодит новые сущности;
  - при смене target ref формируется новый adoption cycle.

### Out of scope
- Полный reverse-engineering архитектуры проекта до production-grade качества без участия человека.
- Автоматическое исправление runtime/deploy ошибок приложения в момент onboarding.
- Полная миграция legacy docs в структуру платформы без review.
- Поддержка non-GitHub провайдеров.

## Выбранный execution-подход
1. Репозиторий проходит attach + preflight.
2. Платформа детектирует, что репозиторий непустой, но не onboarded.
3. Оператор запускает `Adopt existing repository`.
4. Control-plane выполняет deterministic scan и сохраняет adoption report.
5. После scan запускается специализированная onboarding-task с жёстким scope:
  - разрешено добавлять `services.yaml`, docs baseline и platform metadata;
  - запрещено делать произвольный рефакторинг прикладного кода.
6. Task работает по фиксированному prompt-пакету и открывает PR в репозитории.
7. Человек review/approve/merge этот PR; только после merge репозиторий получает статус `onboarded`.

## Декомпозиция (Stories/Tasks)
- Story-1: Repository classification и onboarding state machine.
  - Различать `already_onboarded`, `empty_repository`, `existing_requires_adoption`, `conflict_manual_review`.
  - Учитывать `services.yaml`, docs baseline и scanner signals вместе, а не по одному признаку.
- Story-2: Deterministic repository scanner.
  - Сбор deploy/docs/code signals без ИИ на первом шаге.
  - Нормализованный adoption report с evidence path'ами.
  - Ограничение scan scope по ref и path policy.
- Story-3: Dedicated onboarding generation task.
  - Специализированный prompt/инструкции.
  - Фиксированный writable scope: `services.yaml`, `docs/templates/*`, docs baseline, onboarding metadata.
  - PR с structured explanation bundle.
- Story-4: Conflict policy.
  - Если в репозитории уже есть `services.yaml`/docs, но они расходятся с ожиданиями, task не должна silently overwrite всё подряд.
  - Нужно уметь переводить поток в `manual_review_required` с отчётом, что именно конфликтует.
- Story-5: Staff UX.
  - Просмотр scan report до генерации PR.
  - Статусы adoption pipeline: `scanning`, `ready_for_generation`, `pr_opened`, `merged`, `failed`, `manual_review_required`.
- Story-6: Traceability.
  - Adoption report в БД;
  - ссылка на PR;
  - flow events и GitHub service message/issue comment с кратким summary.
- Story-7: Tests.
  - Unit: classification/scanner/prompt-scope enforcement.
  - Integration: PR creation/update path.
  - E2E smoke: существующий репозиторий без `services.yaml` переводится в PR-based onboarding flow.

## Quality gates
- QG-S8-D3-01 Classification gate:
  - существующий repo не должен попадать в empty-repo path.
- QG-S8-D3-02 Scanner gate:
  - adoption report воспроизводим и содержит path-based evidence.
- QG-S8-D3-03 Safe-generation gate:
  - onboarding-task ограничена bootstrap-артефактами и не переписывает код вне заявленного scope.
- QG-S8-D3-04 PR gate:
  - adoption всегда идёт через reviewable PR и explanation bundle.
- QG-S8-D3-05 Repair gate:
  - rerun обновляет текущий onboarding PR идемпотентно либо создаёт новый cycle только при смене target ref.

## Критерии приёмки
- Для существующего репозитория без `services.yaml` платформа показывает статус `existing_requires_adoption`.
- Оператор может запустить scan и увидеть structured adoption report до генерации PR.
- Платформа создаёт PR, в котором присутствуют:
  - draft `services.yaml`;
  - полный `docs/templates/*` каталог или его актуализация;
  - docs baseline/дополнение;
  - explanation bundle с допущениями и зонами ручной проверки.
- PR-based adoption не меняет прикладной код вне bootstrap-scope без явного justification в отчёте.
- После merge PR репозиторий получает статус `onboarded`, а повторный preflight отражает новый статус.

## Риски и допущения
- Scanner не сможет полностью понять архитектуру сложного legacy-репозитория без ручной валидации; именно поэтому нужен обязательный PR-review.
- Нужно жёстко зафиксировать writable scope onboarding-task, иначе репозиторий будет получать небезопасные “умные” правки.
- Для mono/multi-repo проектов adoption нужно увязать с уже существующим ADR-0007 по composition/docs federation.

## Handover
- Next stage: `run:dev`.
- Execution issue: `282`.
- Handover package:
  - этот epic;
  - `docs/delivery/epics/s3/epic-s3-day14-repository-onboarding-preflight.md`;
  - `docs/delivery/epics/s3/epic-s3-day12-docset-import-and-safe-sync.md`;
  - `docs/delivery/epics/s4/epic-s4-day1-multi-repo-composition-and-docs-federation.md`;
  - `docs/architecture/prompt_templates_policy.md`;
  - `docs/architecture/adr/ADR-0007-multi-repo-composition-and-docs-federation.md`.
