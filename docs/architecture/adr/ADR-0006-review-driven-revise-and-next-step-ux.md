---
doc_id: ADR-0006
type: adr
title: "Review-driven revise resolution and next-step label UX"
status: accepted
owner_role: SA
created_at: 2026-02-20
updated_at: 2026-02-21
related_issues: [90, 95]
related_prs: []
supersedes: []
superseded_by: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-20-issue-90-label-flow-ux"
---

# ADR-0006: Review-driven revise resolution and next-step label UX

## TL;DR
- Проблема: при `changes_requested` Owner вынужден вручную выставлять stage/model/reasoning лейблы перед ревизией.
- Цель: сократить ручные действия и сохранить детерминированный audit-путь.
- Предложение: гибридный resolver (`PR labels -> Issue labels -> last run context`) + stage-aware сервисные сообщения с ссылками на следующие шаги.
- Статус: реализовано в `run:dev` цикле по Issue #95.

## Контекст
- В текущем baseline авто-запуск `run:<stage>:revise` от `pull_request_review` работает только если на PR стоит ровно один stage-лейбл.
- В реальном review-флоу это создаёт ручной overhead:
  - перед `Request changes` нужно вручную синхронизировать stage-лейбл;
  - при необходимости отдельно проставлять model/reasoning override;
  - после ревью вручную ориентироваться, какие next-stage лейблы ставить дальше.
- Одновременно нельзя ослаблять существующие ограничения:
  - label transitions должны оставаться audit-friendly (MCP для run-scoped операций и staff API для owner next-step действий);
  - run trigger остаётся один на Issue;
  - при неоднозначности запуск не должен становиться недетерминированным.

## Decision Drivers
- Минимизировать ручные шаги для Owner в review/revise.
- Сохранить прозрачный и воспроизводимый audit trail.
- Не ломать существующие правила label policy и stage-gate.
- Сохранить возможность ручного override model/reasoning.

## Рассмотренные варианты

### Вариант A: только явный stage-лейбл на PR (как сейчас)
- Плюсы:
  - максимально простая логика;
  - детерминированность без inference.
- Минусы:
  - высокий ручной overhead;
  - частые `changes_requested` без автозапуска revise.

### Вариант B: inference только по активному stage-лейблу Issue
- Плюсы:
  - меньше ручной работы на PR.
- Минусы:
  - если Issue stage неактуален или очищен после предыдущего run, inference ломается;
  - слабая устойчивость к частичным ручным действиям.

### Вариант C: гибридный resolver + stage-aware сервисные сообщения
- Плюсы:
  - минимальный ручной overhead в типовом флоу;
  - детерминированность сохраняется за счёт фиксированной цепочки резолва;
  - Owner получает явные next-step подсказки по текущему stage.
- Минусы:
  - выше сложность orchestration и диагностики.

## Решение
Выбран **Вариант C** как целевая модель.

### 1. Resolver stage для `changes_requested` (целевая цепочка)
1. Ровно один stage-лейбл на PR из поддержанного набора пар `run:<stage>|run:<stage>:revise`.
2. Иначе ровно один stage-лейбл на связанном Issue.
3. Иначе последний успешный/ожидающий review run по связке `(repo, issue, pr)` с `trigger_kind=run:<stage>|run:<stage>:revise`.
4. Иначе последний stage-transition в `flow_events` по Issue.

Если на любом уровне обнаружен конфликт (несколько разных stage) или stage не найден:
- автозапуск revise **не выполняется**;
- выставляется `need:input`;
- публикуется сервисное сообщение с явной инструкцией, какой label оставить.

### 2. Resolver model/reasoning (sticky profile)
При запуске revise используется приоритет:
1. `[ai-model-*]` / `[ai-reasoning-*]` на Issue;
2. те же группы на PR;
3. profile из последнего run по связке `(repo, issue, pr)`;
4. project/agent defaults.

Цель: убрать обязательность ручного повторного выставления profile-лейблов перед каждым `changes_requested`.

### 3. Stage-aware сервисные сообщения (next-step cards)
Платформа обновляет единый service-comment и показывает ссылки/действия по контексту:
- для `intake|vision|prd|arch|design|plan`:
  - `run:<stage>:revise` (доработка текущего артефакта),
  - `run:<next-stage>` (переход по pipeline после принятия).
- для `dev`:
  - `run:dev:revise`,
  - `run:qa`.
- для `qa|release|postdeploy|ops`:
  - revise текущего шага (если применимо),
  - переход к следующему stage.

Ограничение плотности подсказок в GitHub-сообщении (чтобы не перегружать Owner):
- обычно 2 action-подсказки (`revise` + канонический `next-stage`);
- для `design` публикуется третий fast-track вариант `run:dev` (в дополнение к `run:plan`).

Минимальный набор ссылок в сообщении:
- Issue;
- PR;
- run-status/диагностический комментарий;
- явный список рекомендуемых лейблов для следующего шага.

Реализованное расширение UX:
- next-step action-link из GitHub ведёт в staff web-console на экран перехода этапа (`/governance/labels-stages?...`);
- frontend выполняет проверку прав (platform admin guard) и показывает confirm-модалку;
- после подтверждения backend выполняет label transition на Issue (снятие текущих `run:*` + постановка целевого `run:*`) через staff API control-plane.

### 4. Оркестрационные ограничения
- Label transitions для next-step UX выполняются через staff API control-plane с RBAC-проверкой и аудитом.
- На Issue остаётся правило одного активного trigger `run:*`.
- При ambiguous resolve не допускается эвристический “best guess” запуск без явного Owner input.

## Последствия

### Позитивные
- Значительно меньше ручных действий в review/revise.
- Быстрее цикл “Request changes -> revise run”.
- Единообразная и понятная коммуникация платформы о следующих шагах.

### Негативные/компромиссы
- Усложнение логики резолва и audit-событий.
- Требуются дополнительные регрессионные сценарии на конфликтные состояния labels.

## Минимальные требования к аудиту
- `run.review.changes_requested.received`
- `run.revise.stage_resolved`
- `run.revise.stage_ambiguous`
- `run.profile.resolved`
- `run.service_message.updated`

## Статус реализации (Issue #95)
1. Реализован resolver chain в review webhook orchestration:
   - `PR labels -> Issue labels -> last run context -> issue history`.
2. Реализован sticky profile resolver:
   - приоритет `Issue -> PR -> last run context -> defaults`,
   - для review-driven revise (`pull_request_review`) используется issue-first порядок.
3. Обновлены run service messages:
   - добавлены stage-aware next-step подсказки,
   - добавлены прямые ссылки на Issue/PR,
   - добавлен список рекомендуемых label-действий.
4. Добавлены/обновлены тесты:
   - resolver happy path и fallback/ambiguous сценарии,
   - render warning/next-step сценарии,
   - profile resolver и launch audit события.

## Внешние референсы
- GitHub Docs (`pull_request_review`, `review.state=changes_requested`):
  - https://github.com/github/docs/blob/main/content/actions/reference/workflows-and-actions/events-that-trigger-workflows.md
  - https://github.com/github/docs/blob/main/content/rest/using-the-rest-api/github-event-types.md
