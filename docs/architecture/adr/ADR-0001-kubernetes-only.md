---
doc_id: ADR-0001
type: adr
title: "Kubernetes-only orchestration"
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

# ADR-0001: Kubernetes-only orchestration

## TL;DR
- Контекст: платформа управляет агентами и окружениями.
- Решение: поддерживать только Kubernetes.
- Последствия: быстрее MVP, меньше вариативности, ниже стоимость поддержки.

## Контекст
- Проблема/драйвер: multi-orchestrator резко увеличивает сложность.
- Ограничения: нужно быстро получить рабочий production и стабильные процессы.
- Связанные требования: fixed focus на Kubernetes SDK.
- Что “ломается” без решения: scope расползается на поддержание нескольких рантаймов.

## Decision Drivers (что важно)
- Скорость поставки, эксплуатационная предсказуемость, надёжность.

## Рассмотренные варианты
### Вариант A: Kubernetes-only
- Плюсы: минимальный scope, понятные API и RBAC, проще bootstrap.
- Минусы: нет portability на ECS/Nomad.
- Риски: vendor lock-in на Kubernetes.

### Вариант B: абстрактный multi-orchestrator слой сразу
- Плюсы: потенциальная portability.
- Минусы: большая стоимость и задержка MVP.
- Риски: абстракции низкой ценности на раннем этапе.

## Решение
Мы выбираем: **Вариант A (Kubernetes-only)**.

## Обоснование (Rationale)
Ограничение scope даёт самый короткий путь к рабочему production и снижает операционную неопределённость.

## Последствия (Consequences)
### Позитивные
- Единый runtime/набор инструментов.
- Простые и прозрачные инструкции bootstrap/deploy.

### Негативные / компромиссы
- Без поддержки альтернативных оркестраторов.

### Технический долг
- Потенциальный abstraction слой для оркестраторов откладывается.

## План внедрения (минимально)
- Изменения в коде: `client-go` adapters + orchestration interfaces.
- Изменения в инфраструктуре: только k8s manifests/bootstrap.
- Миграции/совместимость: не требуется.
- Наблюдаемость: k8s actions в audit.

## План отката/замены
- Условия отката: стратегическое решение о поддержке альтернативного оркестратора.
- Как откатываем: новый ADR с отдельным интерфейсом оркестратора.

## Ссылки
- Brief: `docs/product/brief.md`
- Constraints: `docs/product/constraints.md`

## Апрув
- request_id: N/A
- Решение: approved
- Комментарий:
