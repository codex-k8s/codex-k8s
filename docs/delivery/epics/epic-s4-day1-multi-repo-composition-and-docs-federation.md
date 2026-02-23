---
doc_id: EPC-CK8S-S4-D1
type: epic
title: "Epic S4 Day 1: Multi-repo composition and docs federation execution foundation (Issue #100)"
status: completed
owner_role: EM
created_at: 2026-02-23
updated_at: 2026-02-23
related_issues: [100, 106]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-23-issue-100-day1"
---

# Epic S4 Day 1: Multi-repo composition and docs federation execution foundation (Issue #100)

## TL;DR
- Проблема: текущая delivery-практика не содержит формального execution-гейта для multi-repo compose/deploy и role-aware docs federation.
- Цель Day1: подготовить детерминированный delivery-план реализации federated composition для всех кейсов из Issue #100.
- Результат Day1: execution foundation завершён; owner-ready пакет для `run:dev` с декомпозицией, quality-gates, критериями приемки, рисками и решениями зафиксирован.

## Priority
- `P0`.

## Scope
### In scope
- Декомпозиция implementation stories по компонентам:
  - `control-plane`: resolver + effective manifest + preview path;
  - `worker`: multi-repo checkout/reconcile;
  - `agent-runner`: repo-aware docs context handover;
  - staff/private API: topology/composition/docs sources visibility.
- Формализация validation-пакета для кейсов A..F.
- Уточнение rollout-порядка (`preview -> enforced`) и quality-gates перехода.

### Out of scope
- Непосредственная реализация кода и миграций в рамках текущего запуска.
- Изменение policy для `run:*` label-flow вне требуемого для Issue #100.
- Поддержка non-Kubernetes оркестраторов.

## Граница результата для Issue #106
- В рамках Issue #106 закрывается только документационный deliverable Day1 (execution package и синхронизация traceability).
- Реализация Story-1..Story-7 переносится в отдельный `run:dev` цикл по Issue #100 и не является частью текущего PR.

## Матрица кейсов Issue #100

| Кейс | Что должно поддерживаться | Execution задача | Evidence на выходе |
|---|---|---|---|
| A. Монорепо + единый `services.yaml` | Legacy-compatible single-root режим | Подтвердить zero-regression при включении federated resolver | Regression-checklist без изменений поведения |
| B. Multi-repo + `services.yaml` в каждом repo | Сборка `effective manifest` из нескольких источников | Определить import graph validation и cycle detection | Preview report + fail-fast taxonomy |
| C. Гибрид (orchestrator repo + service repos) | Центральная компоновка с repo imports | Зафиксировать приоритет `root` и правила pin на commit SHA | Resolve trace + audit events |
| D. Docs только в docs repo | Вынесенная документация | Определить обязательные policy/docs sources для ролей | Role-aware docs refs snapshot |
| E. Docs только в service repos | Распределённая сервисная документация | Нормализовать dedup/priority между repo | Docs graph report до/после dedup |
| F. Комбинированный docs режим | Docs repo + service docs одновременно | Утвердить порядок приоритета источников | Deterministic selection report |

## Декомпозиция stories (handover в `run:dev`)
- Story-1: расширить repository topology контракт (`alias`, `role`, `default_ref`, `services_yaml_path`, `docs_root_path`) и staff DTO.
- Story-2: реализовать `composition/preview` с валидацией imports, cycle detection и resolved commit pinning.
- Story-3: внедрить runtime resolve `effective manifest` в `control-plane` с audit-событиями.
- Story-4: добавить multi-repo checkout/reconcile path в `worker` с идемпотентными ретраями.
- Story-5: реализовать docs federation resolve для prompt-context (role-aware + priority + dedup).
- Story-6: расширить observability (composition latency, imported repos count, docs refs count).
- Story-7: собрать regression пакет по кейсам A..F и негативным сценариям (`not_found`, `conflict`, rate limits).

## Quality gates
- Planning gate:
  - `docs/architecture/multi_repo_mode_design.md` и `ADR-0007` согласованы как source-of-truth.
  - Sprint/Epic/Traceability документы синхронизированы.
- Contract gate:
  - topology/composition/docs-source поля формализованы в data/API модели.
  - backward-compatible поведение монорепо описано явно.
- Resolver gate:
  - детерминированный `effective manifest` для A/B/C;
  - imports разрешены только на зарегистрированные repo aliases;
  - commit pinning выполняется перед reconcile.
- Runtime gate:
  - worker применяет единый execution-plan из `effective manifest`;
  - retries и conflict-handling документированы для reconcile path.
- Docs gate:
  - role-aware docs context корректно собирается для D/E/F;
  - детерминированный порядок приоритета источников зафиксирован.
- Security gate:
  - cross-repo операции ограничены project repository bindings;
  - path traversal исключен в preflight validation.
- Regression gate:
  - есть проверяемый test-matrix по кейсам A..F;
  - негативные сценарии приводят к ожидаемому `failed_precondition`/`conflict`/`not_found`.

## Критерии приемки
- [x] Для каждого кейса A..F есть отдельный test scenario и ожидаемый результат.
- [x] Декомпозиция Story-1..Story-7 принята как обязательный scope следующего `run:dev`.
- [x] Зафиксирован rollout-порядок `preview -> enforced` с условиями перехода.
- [x] Обновлены `issue_map` и `requirements_traceability` с ссылкой на execution-артефакты.
- [x] В PR приложен перечень блокеров, рисков и требуемых owner decisions.

## Итог выполнения (Issue #106)
- Day1 execution foundation закрыт как документационный deliverable.
- Handover пакет для `dev`/`qa`/`sre`/`km` подготовлен и синхронизирован с Sprint S4 документами.
- Реализация Story-1..Story-7 вынесена в следующий `run:dev` цикл по Issue #100.

## Блокеры, риски и owner decisions
### Блокеры
- Подтверждение Owner по приоритету: старт S4 Day1 до/после финального S3 Day20 closeout.
- Утверждение Owner по статусу design-документов (`pending -> approved`) для `multi_repo_mode_design` и `ADR-0007`.

### Риски
- Рост GitHub API нагрузки в multi-repo resolve path и риск rate-limit.
- Ошибки компоновки imports (циклы, конфликтующие refs/path).
- Усложнение диагностики при комбинированном docs режиме.

### Owner decisions (required)
1. Подтвердить federated composition как единственный целевой вариант реализации для Issue #100.
2. Подтвердить приоритет source selection в docs federation: `policy/docs repo -> orchestrator repo -> service repos`.
3. Подтвердить rollout policy: обязательный `preview` этап до `enforced` для production проектов.
4. Подтвердить минимальный acceptance threshold: все кейсы A..F проходят regression без P0/P1 блокеров.

## Техническая валидация (Context7)
- Подтверждены релевантные механики `go-github`:
  - чтение содержимого по `path + ref` через `Repositories.GetContents`;
  - обработка `RateLimitError` и `AbuseRateLimitError` для устойчивого resolve path.
- Подтверждены релевантные механики `client-go`:
  - controller pattern на `shared informer + rate-limited workqueue`;
  - `RetryOnConflict` и ограничение запросов через `rest.Config` (`QPS/Burst`).

## Handover
- `dev`: реализация Story-1..Story-7 с unit/integration/runtime evidence.
- `qa`: regression matrix по A..F + негативные сценарии импортов/refs/rate-limits.
- `sre`: наблюдаемость и capacity baseline по multi-repo execution path.
- `km`: контроль doc traceability и синхронизация delivery/product/architecture ссылок.
