---
doc_id: PRD-CK8S-0119
type: prd
title: "Issue #119 — PRD: E2E A+B core lifecycle and review-revise regression"
status: draft
owner_role: PM
created_at: 2026-02-24
updated_at: 2026-02-24
related_issues: [119, 118]
related_prs: []
related_docsets: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-24-issue-119-prd"
---

# PRD: Issue #119 — E2E A+B core lifecycle and review-revise regression

## TL;DR
- Что строим: формализованный acceptance-пакет для e2e-среза A+B из master plan.
- Для кого: Owner, PM, QA, EM как участники MVP gate.
- Почему: нужен однозначный критерий «готово/не готово» для ядра жизненного цикла и review-driven revise.
- MVP: сценарии A1-A3 и B1-B3 без P0/P1 регрессий с полным evidence.
- Критерии успеха: 100% прохождение 6 сценариев, корректные label transitions, доказуемая трассируемость Issue -> PRD -> evidence.

## Проблема и цель
- Problem statement: текущие артефакты issue #119 зафиксированы на уровне vision, но критерии приемки и декомпозиция для PRD-этапа требуют явной формализации.
- Цели:
  - формализовать scope/priority для A+B e2e-среза;
  - задать проверяемые acceptance criteria в формате Given/When/Then;
  - закрепить обязательный состав evidence для приемки в Issue #118.
- Почему сейчас: issue #119 является частью финального MVP E2E gate и влияет на решение о готовности общего stage-контура.

## Пользователи / Персоны
- Owner: принимает решение по quality gate и финальному статусу.
- PM: подтверждает соответствие продуктовым требованиям и границам scope.
- QA: проводит верификацию сценариев и полноту evidence.
- EM: контролирует прохождение delivery-gates и отсутствие незакрытых блокеров.

## Сценарии/Use Cases (кратко)
- UC-1: как QA, я хочу пройти A1-A3 последовательно, чтобы подтвердить воспроизводимость полного lifecycle.
- UC-2: как PM, я хочу проверить B1-B3, чтобы убедиться в корректности review-driven revise и policy поведения при ambiguity.
- UC-3: как Owner, я хочу получить единый evidence bundle, чтобы принять решение без ручной реконструкции событий.

## Требования (Functional Requirements)
- FR-119-1: Scope issue #119 ограничен сценариями A1, A2, A3, B1, B2, B3 из `docs/delivery/e2e_mvp_master_plan.md`.
- FR-119-2: Для каждого сценария обязателен однозначный результат `pass/fail` с ссылкой на run evidence.
- FR-119-3: Для B2 ambiguity обязателен `need:input` и отсутствие запуска revise-run.
- FR-119-4: Для B3 подтверждается sticky model/reasoning между revise-итерациями по policy resolver.
- FR-119-5: После завершения среза evidence публикуется в Issue #118 в формате, достаточном для воспроизведения проверки.
- FR-119-6: Трассируемость должна быть синхронизирована в `docs/delivery/issue_map.md` и артефактах issue #119.

## Acceptance Criteria (AC)
- AC-1: Полный проход A-среза
  - Given подготовлено рабочее окружение и активны корректные labels
  - When выполняются A1, A2 и A3
  - Then каждый из трёх сценариев завершён в статусе pass без P0/P1 регрессий
- AC-2: Review-driven revise regression
  - Given для B1/B3 есть валидный PR и однозначный stage context
  - When поступает `changes_requested` и выполняется revise flow
  - Then revise-run запускается и завершается корректно, а профиль model/reasoning соответствует policy sticky resolve
- AC-3: Ambiguity handling
  - Given на PR/Issue присутствует неоднозначное stage-состояние
  - When инициируется review-driven revise
  - Then revise-run не стартует, выставляется `need:input`, публикуется remediation-комментарий
- AC-4: Label/state traceability
  - Given сценарии A+B завершены
  - When проверяются `flow_events` и service-комментарии
  - Then ожидаемые `run:*`/`state:*`/`need:*` transitions зафиксированы без пропусков
- AC-5: Evidence completeness
  - Given завершены все 6 сценариев
  - When формируется итоговый отчёт
  - Then в Issue #118 опубликован evidence bundle: `run_id`, transitions, ключевые ссылки на логи, список отклонений и resolution status

## Non-Goals (явно)
- Наборы C, D, E, F из master plan не входят в scope issue #119.
- Изменения кода, инфраструктуры и runtime policy в рамках этого issue не выполняются.
- Расширение stage taxonomy или label classes не входит в задачу.

## Нефункциональные требования (NFR, верхний уровень)
- Надежность: результаты прогона воспроизводимы при повторной проверке.
- Производительность: отсутствуют требования к latency сверх базового run SLA.
- Безопасность: evidence не содержит секретов/токенов/credential material.
- Наблюдаемость: все ключевые события проверяются через `flow_events` и связанные service-комментарии.
- Совместимость: проверки не нарушают текущую label policy и stage process model.
- Локализация: артефакты и PR-коммуникация ведутся на русском языке.

## Аналитика и события (Instrumentation)
- События: `run.*`, `label.transition.*`, `review.revise.*`, `run.revise.pr_not_found` (при наличии), `need.input.*`.
- Атрибуты: `issue_number`, `pr_number`, `run_id`, `stage`, `agent_key`, `correlation_id`.
- Метрики:
  - pass rate A+B;
  - accuracy label transitions;
  - revise success rate (B1/B3);
  - ambiguity handling correctness (B2).
- Дашборды: evidence-сводка в Issue #118 + выборка по `flow_events`.

## Зависимости
- `docs/delivery/e2e_mvp_master_plan.md`.
- `docs/product/labels_and_trigger_policy.md`.
- `docs/product/stage_process_model.md`.
- `docs/architecture/adr/ADR-0006-review-driven-revise-and-next-step-ux.md`.
- Актуальная доступность Issue #118 для публикации evidence.

## Риски и вопросы
- Риски:
  - нестабильность full-env окна прогона;
  - ambiguity labels перед запуском B-сценариев;
  - неполнота audit-следа.
- Открытые вопросы:
  - требуется ли отдельный owner-шаблон комментария для унификации evidence в Issue #118.

## План релиза (черновик)
- Ограничения выката: документарный PRD scope без кодовых изменений.
- Риски релиза: отсутствие кодового риска, но возможен процессный риск неполного evidence.
- Роллбек: откат markdown-изменений через PR.

## Приложения
- Ссылки на design/ADR:
  - `docs/architecture/adr/ADR-0006-review-driven-revise-and-next-step-ux.md`
  - `docs/architecture/mcp_approval_and_audit_flow.md`
- Ссылки на docset:
  - `docs/delivery/e2e_mvp_master_plan.md`
  - `docs/delivery/vision/issue-119/project_charter.md`
  - `docs/delivery/vision/issue-119/success_metrics.md`
  - `docs/delivery/vision/issue-119/risks_register.md`

## Апрув
- request_id: owner-2026-02-24-issue-119-prd
- Решение: pending
- Комментарий:
