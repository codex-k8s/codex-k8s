---
doc_id: TRH-CK8S-S5-0001
type: traceability-history
title: "Sprint S5 Traceability History"
status: in-review
owner_role: KM
created_at: 2026-03-12
updated_at: 2026-03-12
related_issues: [154, 155, 170, 171, 175, 327]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-12-traceability-s5-history"
---

# Sprint S5 Traceability History

## TL;DR
- Этот файл хранит historical delta для Sprint S5.
- Текущая master-карта связей остаётся в `docs/delivery/issue_map.md`.
- Текущее покрытие FR/NFR остаётся в `docs/delivery/requirements_traceability.md`.

## Актуализация по Issue #155 (`run:plan`, 2026-02-25)
- Для FR-053/FR-054 добавлены execution-governance артефакты Sprint S5 (`epic_s5.md`, обновлённый sprint-plan, issue-map sync).
- Зафиксированы quality-gates QG-01..QG-05 и критерии завершения handover в `run:dev`; QG-05 закрыт после Owner review в PR #166.
- Реестр `BLK-155-*`, `RSK-155-*`, `OD-155-*` синхронизирован между `sprint_s5`, `epic_s5` и Day1 epic; `BLK-155-01..02` закрыты, `OD-155-01..03` утверждены (2026-02-25).

## Актуализация по Issue #170 (`run:plan`, 2026-02-25)
- Добавлен Day2 execution-артефакт `docs/delivery/epics/s5/epic-s5-day2-launch-profiles-dev-execution.md` для single-epic реализации FR-053/FR-054.
- Зафиксированы quality-gates QG-D2-01..QG-D2-05 и DoD-пакет для handover в `run:dev`.
- Создана implementation issue `#171`; связь `#170 -> #171 -> FR-053/FR-054` синхронизирована в `issue_map` и Sprint S5 docs.

## Актуализация по Issue #175 (`run:dev`, 2026-02-25)
- Для FR-026/FR-027 зафиксировано исключение в label policy: `need:reviewer` на PR (`pull_request:labeled`) запускает reviewer-run для ручного pre-review.
- Обновлены связные документы: `README.md`, `docs/product/{requirements_machine_driven,labels_and_trigger_policy,agents_operating_model,stage_process_model}.md`, `docs/architecture/api_contract.md`.

## Актуализация по Issue #171 (`run:dev`, 2026-02-25)
- В `runstatus` реализована typed next-step action matrix:
  `action_kind`, `target_label`, `display_variant`, `url`.
- Добавлен deterministic profile resolver для next-step матрицы (baseline `quick-fix|feature|new-service`) и preview/execute path через staff API.
- Для ambiguity/not-resolved review-stage сценариев добавлен hard-stop remediation:
  автоматическая постановка `need:input` через runstatus/webhook path до публикации warning comment.
- Проверка изменений зафиксирована unit-пакетом:
  `go test ./services/internal/control-plane/internal/domain/runstatus ./services/internal/control-plane/internal/domain/webhook`.
