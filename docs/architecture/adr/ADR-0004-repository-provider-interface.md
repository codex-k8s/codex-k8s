---
doc_id: ADR-0004
type: adr
title: "Repository provider abstraction (GitHub first, GitLab-ready)"
status: accepted
owner_role: SA
created_at: 2026-02-06
updated_at: 2026-02-06
related_issues: [1]
related_prs: []
supersedes: []
superseded_by: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-19-full-docset"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-19
---

# ADR-0004: Repository provider abstraction (GitHub first, GitLab-ready)

## TL;DR
- Контекст: сейчас нужен GitHub, позже нужен GitLab.
- Решение: доменный слой работает через provider-интерфейсы, GitHub реализуется первым адаптером.
- Последствия: минимизируем vendor lock-in и повторную переработку домена.

## Контекст
- Проблема/драйвер: прямое зашивание GitHub логики затруднит добавление GitLab.
- Ограничения: не перегружать MVP лишними абстракциями.

## Decision Drivers (что важно)
- Расширяемость без rewrite.
- Чёткие доменные контракты.
- Контролируемая сложность MVP.

## Рассмотренные варианты
### Вариант A: GitHub-only без интерфейсов
- Плюсы: быстрый старт.
- Минусы: дорогой переход к GitLab.

### Вариант B: provider interface + GitHub adapter
- Плюсы: правильные границы и эволюционность.
- Минусы: небольшой upfront design cost.

### Вариант C: сразу GitHub+GitLab
- Плюсы: шире покрытие.
- Минусы: лишний объём для MVP.

## Решение
Мы выбираем: **Вариант B**.

## Обоснование (Rationale)
Интерфейсы и адаптеры дают нужный задел без значимого оверхеда.

## Последствия (Consequences)
### Позитивные
- GitLab можно добавить без изменения use-cases.
- Тестирование домена проще через mock providers.

### Негативные / компромиссы
- Появляется слой абстракций, который нужно держать минимальным.

### Технический долг
- Формализация provider capability matrix на этапе GitLab onboarding.

## План внедрения (минимально)
- Определить `RepositoryProvider` интерфейс и capability flags.
- Реализовать GitHub provider.
- Добавить encrypted token storage + rotation hooks.

## План отката/замены
- Условия отката: интерфейс становится слишком обобщённым и мешает.
- Как откатываем: делим интерфейс на меньшие специализированные контракты.

## Ссылки
- Constraints: `docs/product/constraints.md`
- Delivery Plan: `docs/delivery/delivery_plan.md`

## Апрув
- request_id: N/A
- Решение: approved
- Комментарий:
