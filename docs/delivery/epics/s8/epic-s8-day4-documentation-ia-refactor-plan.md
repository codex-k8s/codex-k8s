---
doc_id: EPC-CK8S-S8-D4-DOC-IA
type: epic
title: "Sprint S8 Day 4: Plan для рефакторинга структуры проектной документации (Issue #320)"
status: in-review
owner_role: EM
created_at: 2026-03-11
updated_at: 2026-03-11
related_issues: [318, 320, 254, 281, 282, 309, 312]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-11-issue-320-plan"
---

# Sprint S8 Day 4: Plan для рефакторинга структуры проектной документации (Issue #320)

## TL;DR
- Issue `#320` подготовлен как единый execution backlog item для рефакторинга IA проектной документации без полного re-root дерева `docs/`.
- Исполнение разбито на три последовательные волны: governance baseline -> migration/sync -> issue/link/drift closure.
- План явно синхронизирован с зависимыми открытыми потоками `#281`, `#282`, `#309`, `#312`, `#318` и traceability-документами Sprint S8.

## Контекст
- Stage continuity: `#318 -> #320`.
- В Issue `#318` Owner согласовал вариант 1: сохранить верхний доменный слой `docs/product`, `docs/architecture`, `docs/delivery`, `docs/ops`, `docs/templates`, а инициативные и handover-пакеты переносить во вложенные доменные папки.
- Source of truth для IA/governance уже частично существует в `docs/delivery/development_process_requirements.md`, где зафиксированы:
  - фиксированная карта `product/architecture/delivery/ops/templates`;
  - запрет превращать `docs/templates/` в хранилище фактических проектных индексов;
  - требование синхронно обновлять `issue_map`, `requirements_traceability`, sprint/epic indexes.
- Фактический drift, подтверждённый при анализе репозитория и открытых issues:
  - отсутствует канонический root-index `docs/index.md`;
  - доменные папки `docs/product`, `docs/architecture`, `docs/delivery`, `docs/ops` не имеют собственных `README.md`;
  - `docs/templates/user_story.md` содержит устаревшую ссылку на `docs/03_engineering/definition_of_done.md`;
  - onboarding-потоки `#281` и `#282` всё ещё описывают baseline через `docs/README.md`, что конфликтует с новой root-IA;
  - `services.yaml/spec.projectDocs` и open issues используют жёсткие ссылки на текущие пути, поэтому перенос без migration-map недопустим.

## Принятое execution-решение
- Governance source of truth не выносится в новый конкурирующий policy-документ: execution расширяет `docs/delivery/development_process_requirements.md` и добавляет навигационные индексы.
- Канонический root navigation path: `docs/index.md`.
- Канонический navigation path доменных папок: `docs/<domain>/README.md`.
- `docs/templates/` остаётся только каталогом шаблонов и инструкций по их применению; проектные индексы и фактические handover-пакеты туда не переносятся.
- Execution остаётся в одном implementation issue `#320` с последовательными волнами; дополнительные follow-up issues допускаются только при обнаружении реального blocker'а на Wave 2.

## Execution waves

| Wave | Цель | Основные артефакты | Ключевой результат |
|---|---|---|---|
| Wave 1 | Зафиксировать governance baseline и каноническую навигацию | `docs/delivery/development_process_requirements.md`, `docs/index.md`, `docs/{product,architecture,delivery,ops}/README.md`, `docs/templates/index.md`, `docs/templates/user_story.md` | Утверждены правила размещения, root/domain indexes и template-only роль каталога `docs/templates/` |
| Wave 2 | Выполнить migration по одобренной карте без нарушения source-of-truth путей | candidate-пакеты `docs/architecture/*`, `docs/ops/*`, `docs/delivery/*`, migration-map, `docs/delivery/{delivery_plan,issue_map,requirements_traceability}.md`, `docs/delivery/{sprints,epics}/README.md` | Инициативные/handover документы разложены по доменным подпапкам, а traceability и внутренние markdown-ссылки синхронизированы |
| Wave 3 | Закрыть внешние ссылки, runtime contracts и drift-control | `services.yaml`, open issues `#254`, `#281`, `#282`, `#309`, `#312`, `#318`, repo-local doc-drift check entrypoint/evidence | `spec.projectDocs/spec.roleDocTemplates`, issue/PR ссылки и проверка drift приведены к новой IA |

## Sequencing constraints
- Wave 1 обязателен перед любым переносом файлов: без согласованного root/domain convention нельзя безопасно переписывать links и onboarding baselines.
- Wave 2 не стартует без явного migration-map формата `old path -> new path -> owner_role -> affected links/issues -> migration note`.
- Wave 3 выполняется только после завершения файлового переноса и локальной проверки markdown-ссылок; update open issues раньше приведёт к повторному drift.
- Потоки `#281` и `#282` не должны финализировать docs baseline до merge результата `#320`, иначе onboarding автоматически создаст устаревшую структуру.

## Quality gates

| Gate | Что проверяем | Ожидаемый результат |
|---|---|---|
| `QG-320-01 IA contract` | Канонический root/domain navigation зафиксирован | `docs/index.md` + `docs/<domain>/README.md` согласованы в governance-docs и handover-плане |
| `QG-320-02 Migration map` | Перед переносом подготовлена и reviewed полная migration-map | Нет ad-hoc move без списка затронутых ссылок и owner-role |
| `QG-320-03 Runtime docs sync` | `services.yaml/spec.projectDocs` и `spec.roleDocTemplates` указывают только на существующие пути | Агентный docs context и role-aware templates не ломаются после migration |
| `QG-320-04 External refs sync` | Обновлены открытые issues и artefact references | `#254`, `#281`, `#282`, `#309`, `#312`, `#318` больше не содержат stale doc-path/blob refs |
| `QG-320-05 Reproducible drift check` | Есть один repo-local entrypoint для проверки doc drift | Проверка выявляет broken repo-relative paths, stale blob links и несинхронизированные doc refs |
| `QG-320-06 Traceability closure` | Delivery docs синхронизированы | `delivery_plan`, `issue_map`, `requirements_traceability`, sprint/epic indexes отражают итоговую IA |

## Definition of Ready (`run:dev` на Issue #320)
- [x] Вариант 1 IA из `#318` закреплён как базовый execution direction.
- [x] Определено, что root navigation path = `docs/index.md`, а доменные индексы = `README.md`.
- [x] Зафиксирован affected-issues baseline: `#254`, `#281`, `#282`, `#309`, `#312`, `#318`.
- [x] Execution strategy оставлена в одном issue `#320`, без преждевременного дробления scope.
- [x] Подтверждено, что на этапе `run:plan` достаточно markdown-only изменений; реализация `services.yaml`/issue updates перенесена в `run:dev`.

## Definition of Done (`run:dev` для Issue #320)
- [ ] В репозитории есть `docs/index.md` и доменные `README.md`, отражающие новую IA.
- [ ] Governance-правила размещения и миграции зафиксированы в source-of-truth документе без дублирующего policy-файла.
- [ ] Инициативные/handover-пакеты перемещены по migration-map без несанкционированного re-root доменов.
- [ ] `services.yaml`, `docs/delivery/{issue_map,requirements_traceability,delivery_plan}.md`, `docs/delivery/{sprints,epics}/README.md` и внутренние markdown-ссылки синхронизированы.
- [ ] Открытые issues и PR-артефакты с прямыми ссылками на docs обновлены после migration.
- [ ] Выполнен repo-local doc-drift check, а evidence добавлен в PR.

## Self-check (common checklist, релевантные пункты)
- Scope ограничен delivery/documentation governance; код, YAML, scripts и runtime не менялись в рамках `run:plan`.
- Сохранена карта доменов `product/architecture/delivery/ops/templates`; план не вводит role-based физическую структуру папок.
- Новые внешние зависимости не добавлялись.
- Traceability синхронизируется через `delivery_plan`, `issue_map`, `requirements_traceability`, sprint/epic indexes.
- Секреты и runtime-данные в артефакты не включались.

## Blockers, risks, owner decisions

| Тип | ID | Описание | Статус |
|---|---|---|---|
| blocker | `BLK-320-01` | Без owner-approve нельзя запускать `run:dev` на перенос документов и правки open issues | open |
| blocker | `BLK-320-02` | Если до merge `#320` начнётся реализация `#281/#282` с новым docs baseline, возникнет двойной источник правды по структуре onboarding docs | open |
| risk | `RSK-320-01` | Частичный перенос без полного migration-map сломает внутренние ссылки, prompt context и ссылочную навигацию открытых issues | open |
| risk | `RSK-320-02` | Попытка вынести governance в отдельный новый policy-файл создаст конкурирующие source-of-truth и усложнит поддержку | open |
| risk | `RSK-320-03` | Массовое обновление open issues после migration может оставить часть branch-specific blob links, если не будет автоматизированной проверки | open |
| decision | `OD-320-01` | Канонический root index = `docs/index.md`; доменные индексы = `docs/<domain>/README.md` | accepted |
| decision | `OD-320-02` | Governance source of truth остаётся в `docs/delivery/development_process_requirements.md`; отдельный параллельный policy-файл не создаётся | accepted |
| decision | `OD-320-03` | Execution выполняется в одном issue `#320` по трем волнам; новые follow-up issues создаются только при обнаружении blocker'а во время migration | accepted |

## Context7 verification
- Через Context7 (`/websites/cli_github_manual`) подтверждён актуальный синтаксис неинтерактивных команд:
  `gh issue create`, `gh issue edit`, `gh pr create`, `gh pr edit`, `gh pr view`.
- Новые библиотеки и внешние зависимости для plan-stage не требуются.

## Acceptance criteria (Issue #320, plan-stage)
- [x] Сформирован единый execution package без расползания scope на несколько несвязанных backlog items.
- [x] Зафиксированы волны исполнения, sequencing constraints и зависимость с onboarding streams `#281/#282`.
- [x] Определены quality-gates, DoR/DoD, blockers, risks и owner decisions.
- [x] Перечень затронутых артефактов и open issues явно зафиксирован для handover в `run:dev`.

## Handover в `run:dev`
- Следующий этап: `run:dev` на том же Issue `#320`.
- Порядок исполнения: `Wave 1 -> Wave 2 -> Wave 3`, без параллельного docs-migration PR.
- Обязательные проверки в `run:dev`:
  - repo-local doc-drift check;
  - `git diff --check`;
  - синхронизация `services.yaml` и traceability-доков;
  - update открытых issues после завершения migration.
- Если на Wave 2 обнаружится blocker, который нельзя закрыть в рамках одного PR, агент обязан создать отдельный follow-up issue по контракту `em: ...` и остановить перенос до owner-решения.
