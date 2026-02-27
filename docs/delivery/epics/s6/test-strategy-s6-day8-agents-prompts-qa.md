---
doc_id: TST-STR-S6-D8
type: test-strategy
title: "S6 Day8 — Test Strategy для QA lifecycle agents/prompt templates (Issue #201)"
status: in-review
owner_role: QA
created_at: 2026-02-27
updated_at: 2026-02-27
related_issues: [184, 185, 187, 189, 195, 197, 199, 201, 216]
related_prs: [202]
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-27-issue-201-qa-strategy"
---

# Test Strategy: S6 Day8 QA для lifecycle agents/prompt templates

## TL;DR
- Подход: risk-based QA с фокусом на AC issue `#201` (contracts, migrations/invariants, UI flow, regression evidence).
- Уровни тестирования: static contract review + automated backend/frontend + runtime smoke в `full-env`.
- Автоматизация: `go test`, `npm build`, `make lint-go`, Kubernetes rollout/runtime checks, SQL schema verification.
- Ключевой риск: `make dupl-go` падает на дубли (включая затронутые в Day7 зоны), требуется фиксация в release risk register.

## Цели тестирования
- Подтвердить корректность staff HTTP/gRPC контрактов `agents/templates/audit` после merge PR `#202`.
- Подтвердить применение миграции `prompt_templates` и инвариантов (`unique active version`, check constraints).
- Подтвердить, что UI-контур `Agents` работает через typed API вместо mock/scaffold логики.
- Сформировать evidence-пакет и решение readiness для перехода в `run:release`.

## Область тестирования
### Что тестируем
- Contract-first слой:
  - OpenAPI endpoints для `staff/agents`, `prompt-templates`, `audit`.
  - gRPC RPC-контракты `ListAgents/GetAgent/UpdateAgentSettings/.../ListPromptTemplateAuditEvents`.
- Data/migration слой:
  - DDL миграции `20260225120000_day22_prompt_templates_lifecycle.sql`.
  - Наличие таблицы/индексов/check constraints в runtime PostgreSQL.
  - Факт bootstrap seeds в `prompt_templates` (global `ru/en`, `work/revise`).
- Backend/runtime:
  - `go test` по `control-plane` и `api-gateway`.
  - `make lint-go`.
  - Runtime smoke: `kubectl get`, `kubectl rollout status`, логи `control-plane/worker`.
- Frontend:
  - `npm --prefix services/staff/web-console run build`.
  - Статическая проверка typed API интеграции в `features/agents` и `AgentDetailsPage`.

### Что не тестируем
- Полноценный ручной UI-clickthrough через браузер (auth/session/manual interaction не запускались).
- Нагрузочные/перфоманс тесты `preview/diff` для больших шаблонов.
- Security/pentest beyond стандартных policy и RBAC проверок.
- Cross-project regression за пределами scope `S6 Day8`.

## Уровни и виды тестирования
- Unit: покрытие существующими Go unit/integration тестами в пакетах сервисов.
- Integration: compile/build и runtime совместимость `control-plane <-> api-gateway <-> web-console`.
- E2E: ограниченный smoke на уровне namespace (rollout/health/logs/schema).
- Performance: не выполнялось (out of scope этапа `run:qa` для этой задачи).
- Security: проверка policy/RBAC ограничений и отсутствия секретов в evidence.
- Chaos: не применялось.

## Окружения
- Dev slot: `full-env`, namespace `codex-k8s-dev-1`, run `534af179-7c1f-459b-8c7c-258f9d6c6835`.
- Production: не использовалось в этом QA run.
- Prod: не использовалось.

## Тестовые данные
- Генерация: bootstrap seeded templates (`ru/en`, `work/revise`) в `prompt_templates`.
- Маскирование PII: персональные/секретные данные не извлекались, значения секретов в отчёт не включались.

## Инструменты и CI
- Линтеры: `make lint-go`, `make dupl-go` (diagnostic risk signal).
- Тест раннеры: `go test ./services/internal/control-plane/...`, `go test ./services/external/api-gateway/...`.
- Отчеты: markdown evidence в `epic-s6-day8-agents-prompts-qa.md` + artifacts этого пакета.
- Нормативные команды сверены через Context7:
  - `/websites/cli_github_manual` (gh issue/pr/json usage).
  - `/websites/kubernetes_io` (rollout status/log checks).

## Критерии входа/выхода
### Entry criteria
- [x] PR `#202` merged в `main`.
- [x] QA issue `#201` активна и содержит AC.
- [x] Full-env namespace доступен для runtime smoke.

### Exit criteria
- [x] Обязательные QA-артефакты сформированы (`strategy/plan/matrix/checklist`).
- [x] AC по contracts/migrations/UI/regression покрыты evidence.
- [x] Readiness decision для `run:release` зафиксирован.
- [x] Создана follow-up issue `#216` для `run:release` (без trigger label) с continuity-инструкцией про `run:postdeploy`.

## Риски и меры
- Риск `RSK-201-01`: `make dupl-go` падает на дубли (в т.ч. в S6 Day7 затронутых файлах).
  - Мера: зафиксировать как известный технический долг и контролировать в `run:release`.
- Риск `RSK-201-02`: отсутствие ручного UI clickthrough может скрыть UX-дефекты.
  - Мера: выполнить manual acceptance на этапе `run:release`.

## Апрув
- request_id: owner-2026-02-27-issue-201-qa-strategy
- Решение: pending
- Комментарий: ожидается review Owner по readiness и рискам.
