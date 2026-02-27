---
doc_id: SPR-CK8S-INDEX-0001
type: sprint-index
title: "Sprint Index (normalized structure)"
status: active
owner_role: EM
created_at: 2026-02-24
updated_at: 2026-02-27
related_issues: [112, 154, 184, 185, 187, 189, 195, 197, 199, 201, 216]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-24-sprint-index"
---

# Sprint Index

## TL;DR
- Спринты вынесены в отдельную структуру `docs/delivery/sprints/s<номер>/`.
- Для каждого спринта сохранён единый формат: sprint plan + epic catalog + day epics + traceability.
- Источник процесса: `docs/delivery/development_process_requirements.md`.

## Карта спринтов

| Sprint | План | Каталог эпиков | Статус | Комментарий |
|---|---|---|---|---|
| S1 | `docs/delivery/sprints/s1/sprint_s1_mvp_vertical_slice.md` | `docs/delivery/epics/s1/epic_s1.md` | completed | Базовый MVP vertical slice закрыт (Day0..Day7). |
| S2 | `docs/delivery/sprints/s2/sprint_s2_dogfooding.md` | `docs/delivery/epics/s2/epic_s2.md` | completed | Dogfooding + governance baseline закрыты. |
| S3 | `docs/delivery/sprints/s3/sprint_s3_mvp_completion.md` | `docs/delivery/epics/s3/epic_s3.md` | in-progress | Финальный e2e и closeout выполняются по Day20. |
| S4 | `docs/delivery/sprints/s4/sprint_s4_multi_repo_federation.md` | `docs/delivery/epics/s4/epic_s4.md` | completed (day1) | Execution foundation по multi-repo зафиксирован. |
| S5 | `docs/delivery/sprints/s5/sprint_s5_stage_entry_and_label_ux.md` | `docs/delivery/epics/s5/epic_s5.md` | in-progress | UX-упрощение stage/label запуска и deterministic next-step actions (Issues #154/#155/#170/#171). |
| S6 | `docs/delivery/sprints/s6/sprint_s6_agents_prompt_management.md` | `docs/delivery/epics/s6/epic_s6.md` | in-progress | Day1..Day8 (intake/vision/prd/arch/design/plan/dev/qa) синхронизированы; создана issue `#216` для stage `run:release` (trigger-лейбл ставит Owner). |

## Правила структуры
- Sprint-plan: `docs/delivery/sprints/s<номер>/sprint_s<номер>_<name>.md`.
- Epic-catalog: `docs/delivery/epics/s<номер>/epic_s<номер>.md`.
- Day-epic: `docs/delivery/epics/s<номер>/epic-s<номер>-day<день>-<name>.md`.
- Любое изменение статуса спринта синхронно отражается в:
  - `docs/delivery/delivery_plan.md`;
  - `docs/delivery/issue_map.md`;
  - `docs/delivery/requirements_traceability.md`.
