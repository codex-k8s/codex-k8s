---
doc_id: EPC-CK8S-S7-D6
type: epic
title: "Epic S7 Day 6: Plan для закрытия MVP readiness gaps (Issue #241)"
status: in-review
owner_role: EM
created_at: 2026-03-02
updated_at: 2026-03-05
related_issues: [212, 218, 220, 222, 238, 241, 243, 244, 245, 246, 247, 248, 249, 250, 251, 252, 253, 254, 255, 256, 257, 258, 259, 260, 274, 216]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-02-issue-241-plan"
---

# Epic S7 Day 6: Plan для закрытия MVP readiness gaps (Issue #241)

## TL;DR
- Подготовлен execution package для перехода в `run:dev` по Sprint S7.
- По требованию Owner создана отдельная implementation issue на каждый поток `S7-E01..S7-E18`: `#243..#260`.
- Зафиксированы sequencing-waves, quality-gates, DoR/DoD и parity-check (`18/18`) перед запуском dev-этапа.
- Post-plan дополнение: создан issue `#274` (`S7-E19`) для backend cleanup Agents/Configs/Secrets + registry images + running jobs.

## Контекст
- Stage continuity: `#212 -> #218 -> #220 -> #222 -> #238 -> #241`.
- Входной baseline: design package Day5 (`design_doc`, `api_contract`, `data_model`, `migrations_policy`).
- Дополнительное owner-уточнение в Issue `#241`: вместо одной stage-issue создать отдельные implementation issues по каждому `S7-E*` потоку.

## Execution package (S7-E01..S7-E18)

| Epic | Implementation issue | Wave | Priority | Краткий scope |
|---|---:|---|---|---|
| `S7-E01` | #243 | Wave 1 | P0 | Rebase/mainline hygiene для revise-итераций |
| `S7-E11` | #253 | Wave 1 | P0 | Reliability для `mode:discussion` |
| `S7-E13` | #255 | Wave 1 | P0 | Добавление revise-петли `run:qa:revise` |
| `S7-E02` | #244 | Wave 2 | P0 | Sidebar cleanup (удаление не-MVP разделов) |
| `S7-E03` | #245 | Wave 2 | P0 | Удаление глобального фильтра |
| `S7-E04` | #246 | Wave 2 | P0 | Удаление runtime-deploy/images UI контуров |
| `S7-E05` | #247 | Wave 2 | P0 | Cleanup residual references после удаления `Agents` из MVP UI |
| `S7-E06` | #248 | Wave 3 | P0 | Зафиксировать фактический de-scope agent settings (`runtime mode/locale`) |
| `S7-E07` | #249 | Wave 3 | P0 | Зафиксировать фактический prompt source `repo-only` и убрать residual stale traces |
| `S7-E08` | #250 | Wave 3 | P1 | Agents UX de-scope hardening |
| `S7-E15` | #257 | Wave 3 | P0 | Prompt templates только через repo commit workflow |
| `S7-E17` | #259 | Wave 3 | P0 | Self-improve session snapshot reliability |
| `S7-E09` | #251 | Wave 4 | P0 | Runs UX cleanup + deterministic namespace delete |
| `S7-E10` | #252 | Wave 4 | P0 | Cancel/stop для зависших runtime deploy tasks |
| `S7-E16` | #258 | Wave 4 | P0 | Устранение false-failed для `run:intake:revise` |
| `S7-E14` | #256 | Wave 5 | P0 | QA acceptance policy через Kubernetes DNS path |
| `S7-E18` | #260 | Wave 5 | P0 | Documentation governance standardization |
| `S7-E12` | #254 | Wave 5 | P1 | Финальный MVP readiness gate + go/no-go пакет |

## Post-plan additions
- `S7-E19` (`#274`): backend cleanup Agents/Configs/Secrets + registry images + running jobs (Owner request после plan; не входит в parity `18/18`).

## Sequencing constraints
- Wave 1 (`#243`, `#253`, `#255`) — foundation для стабильного dev/revise контура.
- Wave 2 (`#244..#247`) — UI cleanup до перехода в de-scope потоков.
- Wave 3 (`#248`, `#249`, `#250`, `#257`, `#259`) — agents/prompt policy и self-improve reliability.
- Wave 4 (`#251`, `#252`, `#258`) — runtime/run reliability.
- Wave 5 (`#256`, `#260`, `#254`) — QA/governance closeout и финальный readiness gate.

## Актуализация после cleanup `#244` и `#274`
- Первоначальные формулировки `S7-E05/S7-E06/S7-E07` предполагали существование MVP-контуров `Agents` и `Prompt templates`.
- После выполнения `#244` и `#274` эти контуры выведены из MVP, поэтому:
  - `#247` закрывает residual frontend/test/doc/transport references;
  - `#248` фиксирует source-of-truth контракт без agent settings UI/API;
  - `#249` фиксирует source-of-truth контракт `repo-only prompt seeds`.
- После closure `#247/#248/#249` remaining Sprint S7 backlog нормализуется как `#250..#260`.

## Quality gates (`run:plan`)

| Gate | Что проверяем | Статус |
|---|---|---|
| `QG-S7-D6-01` | На каждый поток `S7-E01..S7-E18` создана отдельная implementation issue | passed |
| `QG-S7-D6-02` | Нумерация и sequencing-waves зафиксированы в planning-артефактах | passed |
| `QG-S7-D6-03` | Parity-gate: `approved_execution_epics_count == created_run_dev_issues_count` | passed (`18 == 18`) |
| `QG-S7-D6-04` | Для implementation issues не выставлены trigger-лейблы `run:*` | passed |
| `QG-S7-D6-05` | Traceability синхронизирована (`issue_map`, `requirements_traceability`, sprint/epic docs, delivery plan) | passed |
| `QG-S7-D6-06` | Scope этапа ограничен markdown-only изменениями | passed |

## DoR/DoD для перехода в `run:dev`

### Definition of Ready (`run:dev` launch)
- [x] Design package Day5 подтверждён (`#238`) и доступен как source of truth.
- [x] Implementation backlog создан по одному issue на поток (`#243..#260`).
- [x] Sequencing и зависимости зафиксированы wave-моделью.
- [x] Parity-gate `18/18` зафиксирован в документации.
- [x] Trigger-лейблы на implementation issues не выставлены (ставит Owner по мере запуска).

### Definition of Done (`run:plan` stage)
- [x] Выпущен plan-эпик Day6 с execution package.
- [x] Обновлены sprint/epic каталоги, delivery plan и traceability документы.
- [x] Подготовлен handover в `run:dev` c owner-facing списком implementation issues.

## Self-check (common checklist)
- Проверен scope изменений: только markdown-документы (`*.md`), без code/runtime правок.
- Проверены архитектурные границы и stage-policy: trigger-лейблы на implementation issues не выставлялись.
- Проверена traceability-синхронизация: `issue_map`, `requirements_traceability`, sprint/epic indexes, `delivery_plan`.
- Проверено отсутствие новых внешних зависимостей и отсутствие секретов в артефактах.

## Blockers, risks, owner decisions

| Тип | ID | Описание | Статус |
|---|---|---|---|
| blocker | BLK-S7-D6-01 | Открытая S6 release-зависимость `#216` остаётся внешним блокером общего MVP closeout | open |
| risk | RSK-S7-D6-01 | Параллельный запуск нескольких `run:dev` потоков может нарушить wave-sequencing и увеличить rework | open |
| risk | RSK-S7-D6-02 | Пропуск документационного обновления по отдельным implementation issues приведёт к traceability drift | open |
| decision | OD-S7-D6-01 | Owner-request выполнен: для каждого `S7-E*` создана отдельная issue (`#243..#260`) | accepted |
| decision | OD-S7-D6-02 | Launch policy: trigger `run:dev` ставится Owner по волнам, без массового параллельного старта всех 18 issues | accepted |

## Context7 verification
- Использован Context7 источник `/websites/cli_github_manual` для актуальной проверки неинтерактивных команд `gh issue create`, `gh pr create`, `gh pr edit`.
- Новые внешние зависимости на этапе Day6 не добавлялись.

## Acceptance criteria (Issue #241)
- [x] Подготовлен plan package c execution breakdown для потоков `S7-E01..S7-E18`.
- [x] Зафиксированы sequencing constraints и зависимости (`foundation -> UI cleanup -> policy/runtime reliability -> closeout`).
- [x] Определены quality-gates (DoR/DoD) и release criteria для входа в `run:dev`.
- [x] Синхронизированы `issue_map`, `requirements_traceability`, sprint/epic docs и `delivery_plan`.
- [x] Выполнено owner-уточнение: созданы отдельные implementation issues `#243..#260` с явной нумерацией.

## Handover в `run:dev`
- Следующий этап: `run:dev`.
- Implementation issues для запуска по waves: `#243..#260`.
- Trigger-лейблы `run:dev` на этих issue ставит Owner, соблюдая sequencing-waves.
- Для каждого потока в `run:dev` обязательны:
  - PR с проверками и evidence;
  - обновление traceability документов;
  - переход в `state:in-review` после завершения итерации.
