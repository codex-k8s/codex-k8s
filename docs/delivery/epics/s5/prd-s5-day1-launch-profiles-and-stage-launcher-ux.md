---
doc_id: PRD-CK8S-S5-I155
type: prd
title: "Issue #155 — PRD: Launch profiles and deterministic next-step transitions"
status: in-review
owner_role: PM
created_at: 2026-02-25
updated_at: 2026-02-25
related_issues: [155, 154]
related_prs: [158]
related_docsets:
  - docs/delivery/issue_map.md
  - docs/delivery/requirements_traceability.md
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-25-issue-155-prd"
---

# PRD: Launch profiles and deterministic next-step transitions (Issue #155)

## TL;DR
- Что строим: profile-driven контракт переходов по stage (`quick-fix`, `feature`, `new-service`) с обязательным fallback для next-step действий.
- Для кого: Owner и команды, которые управляют pipeline через labels и service-message подсказки.
- Почему: текущий flow зависит от web-ссылок и ручного выбора этапов "по памяти", что создаёт блокировки и ошибки переходов.
- MVP: deterministic profile resolver + action-card contract (`primary_action` + `fallback_action`) + ambiguity guardrails.
- Критерии успеха: сценарии AC-01..AC-06 проходят без bypass policy и без блокировки при недоступности UI deep-link.

## Проблема и цель
- Problem statement: при review/revise и stage-переходах owner-flow может блокироваться из-за недоступного deep-link или неоднозначного stage контекста.
- Цели:
  - зафиксировать канонические launch profiles и правила эскалации;
  - сделать next-step действия двухканальными (`web-console deep-link` + `text fallback command`);
  - запретить best-guess transition при ambiguity и требовать `need:input`.
- Почему сейчас: Sprint S5 Day1 формирует owner-ready пакет для перехода в `run:dev` по FR-053/FR-054.

## Пользователи / Персоны
- Persona A: Owner, который выполняет stage transitions в issue/pr review-контуре.
- Persona B: PM/EM/KM, которые поддерживают трассируемость и валидируют корректность переходов.
- Persona C: Dev/QA/SRE, которые получают детерминированный handover path в `run:dev` и downstream этапы.

## Сценарии/Use Cases (кратко)
- UC-1: Owner запускает следующий stage через `primary_action` в staff UI.
- UC-2: При недоступности UI owner использует fallback-команду и продолжает процесс без блокировки.
- UC-3: При неоднозначности stage система блокирует transition, выставляет `need:input` и публикует remediation.

## Требования (Functional Requirements)
- FR-155-01: Поддерживаются три launch profile: `quick-fix`, `feature`, `new-service`.
- FR-155-02: Для каждого profile фиксируется обязательная stage trajectory и детерминированные escalation rules.
- FR-155-03: Каждая next-step карточка содержит `launch_profile`, `stage_path`, `primary_action`, `fallback_action`, `guardrail_note`.
- FR-155-04: Fallback contract использует pre-check (`gh issue view ... labels`) перед transition.
- FR-155-05: При ambiguity stage (`0` или `>1` trigger labels) transition запрещён, ставится `need:input`.
- FR-155-06: Review gate после формирования PR синхронизирует `state:in-review` на Issue и PR.

## Acceptance Criteria (AC)
- AC-01
  - Given profile=`feature` и доступный deep-link
  - When owner выбирает `primary_action`
  - Then stage transition выполняется по профилю и фиксируется в audit.
- AC-02
  - Given deep-link недоступен и текущий stage однозначен
  - When owner выполняет fallback (`pre-check -> transition`)
  - Then процесс не блокируется, переход выполняется детерминированно.
- AC-03
  - Given ambiguity stage (`0` или `>1` trigger labels)
  - When выполняется next-step transition
  - Then переход блокируется, ставится `need:input`, публикуется remediation-message.
- AC-04
  - Given profile=`quick-fix` и обнаружен `cross-service impact`
  - When выполняется оценка риска
  - Then происходит обязательная эскалация в `feature` или `new-service`.
- AC-05
  - Given сформирован PR
  - When run завершает этап
  - Then `state:in-review` установлен на PR и Issue, а trigger-label снят с Issue.
- AC-06
  - Given fallback-команда опубликована в service-comment
  - When проверяется security-policy
  - Then команда не содержит секретов и использует только labels из `run:*|state:*|need:*`.

## Non-Goals (явно)
- Реализация backend/frontend кода stage launcher в рамках этого PRD-цикла.
- Пересмотр taxonomy labels вне FR-053/FR-054.
- Изменение RBAC-модели и runtime execution modes.

## Нефункциональные требования (NFR, верхний уровень)
- Надежность: owner-flow не блокируется при недоступности deep-link.
- Производительность: next-step подсказка формируется без ручной донастройки в каждом run.
- Безопасность: fallback-команды не содержат секретов и не допускают переходов вне каталога labels.
- Наблюдаемость: transition и guardrail-решения фиксируются в audit trail (`flow_events`).
- Совместимость: контракт должен быть применим и для `run:prd:revise`, и для downstream `run:dev`.
- Локализация: шаблоны подсказок поддерживают русскую коммуникацию для текущего run locale.

## UX/UI заметки (если применимо)
- Карточка next-step должна оставаться компактной: один primary deep-link + один fallback command.
- При ambiguity карточка не предлагает target-stage переход, только remediation path (`need:input`).
- Текст fallback должен быть готов к копированию без ручной подстановки дополнительных параметров, кроме `<ISSUE_NUMBER>`/`<PR_NUMBER>`.

## Аналитика и события (Instrumentation)
- События:
  - `stage.launch_profile.resolved`
  - `stage.next_step.card_rendered`
  - `stage.next_step.fallback_used`
  - `stage.next_step.blocked_need_input`
- Атрибуты:
  - `issue_number`, `pr_number`, `launch_profile`, `current_stage`, `next_stage`, `reason`.
- Метрики:
  - доля переходов через primary/fallback;
  - число blocked transitions из-за ambiguity;
  - доля успешных review-gate transitions (`Issue + PR`).
- Дашборды:
  - use-case граф "next-step success/blocked";
  - trend blocked ambiguity для оценки drift в label-политике.

## Зависимости
- Внешние системы: GitHub labels/PR APIs и staff web-console deep-link path.
- Команды/сервисы: `pm`, `em`, `dev`, `qa`, `km`, `sre`.
- Лицензии/ограничения: в рамках текущего CI/CD governance и policy каталога labels.

## Риски и вопросы
- Риски:
  - UX перегрузка service-comment при расширении количества actions;
  - рассинхронизация fallback шаблонов и runtime policy;
  - ложная уверенность в stage-контексте без pre-check.
- Открытые вопросы:
  - Нужны ли отдельные fallback-шаблоны для fast-track (`design -> dev`) при сохранении canonical path?
  - Требуется ли отдельная визуальная маркировка для forced escalation Owner-решением?

## План релиза (черновик)
- Ограничения выката: реализация в `run:dev` по инкрементам I1..I5 из Day1 epic.
- Риски релиза: policy-drift при ручных label transitions.
- Роллбек: возврат к текущему stage-aware сервисному сообщению без profile-specific fallback логики.

## Приложения
- Ссылки на design/ADR:
  - `docs/architecture/adr/ADR-0006-review-driven-revise-and-next-step-ux.md`
  - `docs/product/labels_and_trigger_policy.md`
  - `docs/product/stage_process_model.md`
- Ссылки на docset:
  - `docs/delivery/epics/s5/epic-s5-day1-launch-profiles-and-stage-launcher-ux.md`
  - `docs/delivery/sprints/s5/sprint_s5_stage_entry_and_label_ux.md`
  - `docs/delivery/issue_map.md`
  - `docs/delivery/requirements_traceability.md`

## Апрув
- request_id: owner-2026-02-25-issue-155-prd
- Решение: pending
- Комментарий: PRD подготовлен в рамках `run:prd:revise`, ожидается Owner review.
