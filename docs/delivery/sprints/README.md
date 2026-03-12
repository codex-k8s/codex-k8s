---
doc_id: SPR-CK8S-INDEX-0001
type: sprint-index
title: "Sprint Index (normalized structure)"
status: active
owner_role: EM
created_at: 2026-02-24
updated_at: 2026-03-12
related_issues: [112, 154, 184, 185, 187, 189, 195, 197, 199, 201, 212, 218, 220, 222, 238, 241, 243, 244, 245, 246, 247, 248, 249, 250, 251, 252, 253, 254, 255, 256, 257, 258, 259, 260, 216, 262, 263, 265, 281, 282, 320, 327, 333, 335]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-24-sprint-index"
---

# Sprint Index

## TL;DR
- Спринты ведутся в структуре `docs/delivery/sprints/s<номер>/` (S1..S9).
- Для каждого спринта сохранён единый формат: sprint plan + epic catalog + day epics + traceability.
- Sprint index хранит только каноническую карту спринтов и ссылки на sprint/epic артефакты.
- Исторические issue-specific updates по спринтам размещаются в `docs/delivery/traceability/s<номер>_*.md` и не дублируются в sprint index.
- Источник процесса: `docs/delivery/development_process_requirements.md`.

## Карта спринтов

| Sprint | План | Каталог эпиков | Статус | Комментарий |
|---|---|---|---|---|
| S1 | `docs/delivery/sprints/s1/sprint_s1_mvp_vertical_slice.md` | `docs/delivery/epics/s1/epic_s1.md` | completed | Базовый MVP vertical slice закрыт (Day0..Day7). |
| S2 | `docs/delivery/sprints/s2/sprint_s2_dogfooding.md` | `docs/delivery/epics/s2/epic_s2.md` | completed | Dogfooding + governance baseline закрыты. |
| S3 | `docs/delivery/sprints/s3/sprint_s3_mvp_completion.md` | `docs/delivery/epics/s3/epic_s3.md` | in-progress | Финальный e2e и closeout выполняются по Day20. |
| S4 | `docs/delivery/sprints/s4/sprint_s4_multi_repo_federation.md` | `docs/delivery/epics/s4/epic_s4.md` | completed (day1) | Execution foundation по multi-repo зафиксирован. |
| S5 | `docs/delivery/sprints/s5/sprint_s5_stage_entry_and_label_ux.md` | `docs/delivery/epics/s5/epic_s5.md` | in-progress | UX-упрощение stage/label запуска и deterministic next-step actions (Issues #154/#155/#170/#171). |
| S6 | `docs/delivery/sprints/s6/sprint_s6_agents_prompt_management.md` | `docs/delivery/epics/s6/epic_s6.md` | completed | Day1..Day11 закрыты: release closeout `#262`, postdeploy `#263`, ops closeout `#265`; сформирован baseline runbook/monitoring/alerts/SLO/rollback. |
| S7 | `docs/delivery/sprints/s7/sprint_s7_mvp_readiness_gap_closure.md` | `docs/delivery/epics/s7/epic_s7.md` | in-progress | Day1..Day5 закрыли intake/vision/prd/arch/design пакет (`#212/#218/#220/#222/#238`), Day6 plan (`#241`) сформировал execution package и по owner-уточнению создал 18 implementation issues `#243..#260` (по одному на `S7-E01..S7-E18`) с parity-check `18/18` перед `run:dev`. |
| S8 | `docs/delivery/sprints/s8/sprint_s8_go_refactoring_parallelization.md` | `docs/delivery/epics/s8/epic_s8.md` | in-progress | Параллельный Go-refactor stream расширен onboarding-потоками; Day4 по `#320` уже внедрил `docs/index.md`, domain `README.md`, migration-map и remediation issue refs. |
| S9 | `docs/delivery/sprints/s9/sprint_s9_mission_control_dashboard_control_plane.md` | `docs/delivery/epics/s9/epic_s9.md` | in-review | Новый product stream для Mission Control Dashboard: Day1 intake (`#333`) зафиксировал active-set control-plane UX, GitHub-first ограничения и handover в vision issue `#335`. |

## Правила структуры
- Sprint-plan: `docs/delivery/sprints/s<номер>/sprint_s<номер>_<name>.md`.
- Epic-catalog: `docs/delivery/epics/s<номер>/epic_s<номер>.md`.
- Day-epic: `docs/delivery/epics/s<номер>/epic-s<номер>-day<день>-<name>.md`.
- Любое изменение статуса спринта синхронно отражается в:
  - `docs/delivery/delivery_plan.md`;
  - `docs/delivery/issue_map.md`;
  - `docs/delivery/requirements_traceability.md`.
- Historical delta по sprint-issue evidence синхронно отражается в релевантном файле `docs/delivery/traceability/s<номер>_*.md`.
