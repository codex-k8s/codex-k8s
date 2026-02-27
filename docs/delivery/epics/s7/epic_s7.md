---
doc_id: EPC-CK8S-0007
type: epic
title: "Epic Catalog: Sprint S7 (MVP readiness gap closure)"
status: in-progress
owner_role: PM
created_at: 2026-02-27
updated_at: 2026-02-27
related_issues: [212, 218, 220, 199, 201, 210, 216]
related_prs: [213, 215]
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-27-issue-212-intake"
---

# Epic Catalog: Sprint S7 (MVP readiness gap closure)

## TL;DR
- Sprint S7 консолидирует незакрытые MVP-разрывы из UI, stage-flow и delivery-governance в единый execution backlog.
- Day1 intake (`#212`) зафиксировал P0/P1/P2-потоки и актуализировал S6 dependency-chain: `#199/#201` закрыты, открытый блокер — `#216` (`run:release`).
- Цель каталога: дать однозначную stage-декомпозицию и candidate backlog на 18 эпиков до полного readiness цикла `dev -> qa -> release -> postdeploy -> ops -> doc-audit`.

## Stage roadmap
- Day 1 (Intake): `docs/delivery/epics/s7/epic-s7-day1-mvp-readiness-intake.md` (Issue `#212`).
- Day 2 (Vision): `docs/delivery/epics/s7/epic-s7-day2-mvp-readiness-vision.md` (Issue `#218`).
- Day 3 (PRD): формализовать FR/AC/NFR + edge cases по каждому epic-кандидату (`run:prd`, Issue `#220`).
- Day 4 (Architecture): проверить сервисные границы и контракты для implementation-пакетов (`run:arch`).
- Day 5 (Design/Plan): утвердить execution-sequence, quality gates, DoR/DoD (`run:design`, `run:plan`).
- Day 6+ (Execution): реализация и приемка `run:dev -> run:qa -> run:release -> run:postdeploy -> run:ops -> run:doc-audit`.

## Day 2 vision fact
- В Issue `#218` зафиксированы mission, KPI/success metrics и measurable readiness criteria для потоков `S7-E01..S7-E18`.
- Для каждого execution-эпика оформлен baseline: user story, AC, edge cases, expected evidence.
- Зафиксировано обязательное правило decomposition parity перед входом в `run:dev`:
  `approved_execution_epics_count == created_run_dev_issues_count` (coverage ratio = `1.0`).
- Создана continuity issue `#220` для этапа `run:prd` без trigger-лейбла.

## Candidate execution backlog (18 эпиков)

| Epic ID | Priority | Scope | Источник замечаний |
|---|---|---|---|
| S7-E01 | P0 | Rebase/mainline hygiene и merge-conflict policy для PR-итераций | PRC-01 |
| S7-E02 | P0 | Удаление не-MVP разделов и связанного dead code из sidebar/routes | PRC-05 |
| S7-E03 | P0 | Удаление глобального frontend-фильтра и связанного неиспользуемого кода | PRC-04 |
| S7-E04 | P0 | Удаление runtime-deploy/images контуров и cleanup связанных страниц | PRC-02, PRC-05 |
| S7-E05 | P0 | Agents UI cleanup: убрать badge `Скоро`, пересобрать таблицу (без role/project-id) | PRC-03 |
| S7-E06 | P0 | Agents import defaults: runtime mode + locale policy (owner locale + bulk update) | PRC-03 |
| S7-E07 | P0 | Prompt source selector для worker (`repo` vs `db`) и policy в UI | PRC-03 |
| S7-E08 | P1 | Agents UX hardening: массовые операции и консистентность конфигурации | PRC-03 |
| S7-E09 | P0 | Runs UX: удалить колонку типа запуска и гарантировать namespace delete из run details | PRC-06 |
| S7-E10 | P0 | Runtime deploy UX: кнопка cancel/stop для зависших deploy tasks + guardrails | PRC-07 |
| S7-E11 | P0 | Label orchestration reliability: исправить `mode:discussion` trigger-поведение | PRC-08 |
| S7-E12 | P1 | Final MVP readiness gate: e2e evidence bundle + go/no-go для release chain | PRC-01..PRC-08 |
| S7-E13 | P0 | Label policy alignment: добавить `run:qa:revise` и покрыть revise-loop QA-stage | PRC-09 |
| S7-E14 | P0 | QA execution contract: проверка новых/изменённых ручек через Kubernetes DNS path + evidence | PRC-10 |
| S7-E15 | P0 | Agents prompt lifecycle UX: кнопка обновления prompt templates из repo с версионированием | PRC-11 |
| S7-E16 | P0 | Run status reliability: устранить false-failed для фактически успешных `run:intake:revise` | PRC-12 |
| S7-E17 | P0 | Self-improve reliability: доступность и корректная перезапись `agent_sessions` snapshot | PRC-13 |
| S7-E18 | P0 | Documentation governance: единый стандарт issue/PR + doc IA + role-template matrix | PRC-14, PRC-15, PRC-16 |

## Delivery-governance правила
- Каждая следующая stage-issue создаётся отдельной задачей и без trigger-лейбла.
- Trigger-лейбл на запуск этапа ставит Owner после review предыдущего артефакта.
- Для каждого execution-эпика обязательно фиксируются: priority, user story, AC, edge cases, dependency и expected evidence.
- MVP-closeout не считается завершённым без явного доказательства работоспособности `run:doc-audit`.
