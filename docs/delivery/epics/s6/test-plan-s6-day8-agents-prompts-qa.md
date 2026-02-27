---
doc_id: TST-PLN-S6-D8
type: test-plan
title: "S6 Day8 — Test Plan для QA lifecycle agents/prompt templates (Issue #201)"
status: in-review
owner_role: QA
created_at: 2026-02-27
updated_at: 2026-02-27
related_issues: [184, 185, 187, 189, 195, 197, 199, 201, 216]
related_prs: [202]
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-27-issue-201-test-plan"
---

# Test Plan: S6 Day8 QA acceptance/regression

## 1. Идентификатор тест-плана
- ID: `TST-PLN-S6-D8`
- Версия: `1.0`
- Дата: `2026-02-27`

## 2. Ссылки/референсы
- PRD: `docs/delivery/epics/s6/prd-s6-day3-agents-prompts-lifecycle.md`
- Design:
  - `docs/architecture/agents_prompt_templates_lifecycle_design_doc.md`
  - `docs/architecture/agents_prompt_templates_lifecycle_api_contract.md`
  - `docs/architecture/agents_prompt_templates_lifecycle_data_model.md`
  - `docs/architecture/agents_prompt_templates_lifecycle_migrations_policy.md`
- ADR: `docs/architecture/adr/ADR-0009-prompt-templates-lifecycle-and-audit.md`
- DoD/Gates: `docs/delivery/epics/s6/epic-s6-day6-agents-prompts-plan.md`
- Dev handover: GitHub issue `#199`, PR `#202`

## 3. Введение
- Цель тестирования: подтвердить готовность Day7 реализации к переходу в `run:release`.
- Объект тестирования: lifecycle `agents/templates/audit` (contracts + migrations + UI typed flow + regression).

## 4. Test items (что тестируем)
- Компоненты/модули:
  - `services/internal/control-plane`
  - `services/external/api-gateway`
  - `services/staff/web-console`
  - PostgreSQL schema (`agents`, `prompt_templates`)
- API/фичи:
  - staff endpoints `/api/v1/staff/agents*`, `/api/v1/staff/prompt-templates*`, `/api/v1/staff/audit/prompt-templates`
  - gRPC methods `ListAgents..ListPromptTemplateAuditEvents`
  - UI flow `Agents` list/details/settings/templates/diff/preview/history

## 5. Риски (software risk issues)
- `RSK-201-01`: `dupl-go` reports duplicates (non-green quality signal).
- `RSK-201-02`: manual UI behavior не проверен end-to-end через браузер.
- `RSK-201-03`: performance profile `preview/diff` для больших payload не измерен.

## 6. Features to be tested
- Контрактная полнота OpenAPI/proto.
- Схема БД: наличие таблицы `prompt_templates`, индексов и constraints.
- Seed bootstrap: наличие активных `work/revise` шаблонов `ru/en`.
- Backend regressions via `go test`.
- Frontend compile/build via `npm run build`.
- Runtime readiness via `kubectl get/rollout/logs`.

## 7. Features not to be tested
- Browser-driven manual exploratory testing.
- Нагрузочные и fault-injection тесты.
- Penetration/security hardening beyond standard policy checks.

## 8. Подход (Approach)
- Стратегия: risk-based acceptance + regression evidence.
- Уровни тестов:
  - static docs/contracts checks;
  - automated tests/build/lint;
  - runtime smoke + DB schema introspection.
- Автоматизация:
  - `go test ./services/internal/control-plane/...`
  - `go test ./services/external/api-gateway/...`
  - `npm --prefix services/staff/web-console run build`
  - `make lint-go`, `make dupl-go`
- Ручные проверки:
  - анализ runtime логов `control-plane/worker`;
  - проверки rollout и статуса workload в namespace.

## 9. Критерии pass/fail
- `PASS`:
  - обязательные AC issue `#201` закрыты evidence;
  - backend tests + frontend build + lint-go зелёные;
  - migration/invariant checks подтверждены в runtime DB.
- `FAIL`:
  - критичный регресс в contracts/migration/runtime;
  - блокирующие ошибки в backend tests/build.
- Условный риск (non-blocking, фиксируется явно):
  - `dupl-go` failure при отсутствии функциональных регрессий.

## 10. Suspension criteria & resumption
- Когда останавливаем тестирование:
  - потеря доступа к namespace/runtime;
  - критичный fail в базовых тестах/сборке.
- Когда возобновляем:
  - восстановлен доступ/окружение;
  - исправлены блокирующие ошибки.

## 11. Test deliverables
- Отчеты:
  - `docs/delivery/epics/s6/epic-s6-day8-agents-prompts-qa.md`
  - `docs/delivery/epics/s6/test-strategy-s6-day8-agents-prompts-qa.md`
  - `docs/delivery/epics/s6/test-matrix-s6-day8-agents-prompts-qa.md`
  - `docs/delivery/epics/s6/regression-checklist-s6-day8-agents-prompts-qa.md`
- Логи:
  - команды `go test`, `npm build`, `kubectl`, `psql` (сведены в markdown evidence).
- Баг-репорты:
  - не заведены (блокирующих дефектов не выявлено); риск `dupl-go` зафиксирован.

## 12. Remaining test tasks
- Manual UI acceptance с browser interactions на этапе `run:release`.
- Проверка release-window сценария migration-first rollout.
- Regression после релизного cut-over (stage `run:postdeploy`).

## 13. Environmental needs
- Окружения: `full-env` namespace `codex-k8s-dev-1`.
- Доступы: `gh`, `kubectl`, доступ к pod logs и `exec` в namespace.
- Данные: seeded templates в `prompt_templates`.

## 14. Роли и ответственность
- QA: формирование стратегии/плана/матрицы/checklist и readiness decision.
- Dev: устранение потенциальных замечаний `run:dev:revise` при наличии блокеров.
- SRE: release/postdeploy validation и эксплуатационные проверки.

## 15. Расписание
- `2026-02-27`: complete QA acceptance + regression evidence + release handover issue.

## 16. Риски планирования и контингенси
- Риск unavailable tooling/deps (решено установкой npm deps перед build).
- Риск неполного покрытия UI (компенсация через explicit not-tested и follow-up в release stage).

## 17. Апрувы
- request_id: owner-2026-02-27-issue-201-test-plan
- Решение: pending
- Комментарий: ожидается подтверждение readiness к `run:release`.
