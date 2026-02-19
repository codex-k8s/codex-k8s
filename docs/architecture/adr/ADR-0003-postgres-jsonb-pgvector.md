---
doc_id: ADR-0003
type: adr
title: "PostgreSQL as unified state backend with JSONB and pgvector"
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

# ADR-0003: PostgreSQL as unified state backend with JSONB and pgvector

## TL;DR
- Контекст: нужно единое хранилище состояния, аудита и документной памяти.
- Решение: использовать PostgreSQL с `JSONB` и `pgvector`.
- Последствия: простая синхронизация multi-pod и единый стек хранения.

## Контекст
- Проблема/драйвер: раздельные хранилища усложняют консистентность и эксплуатацию.
- Ограничения: нужен быстрый MVP и понятная операционная модель.

## Decision Drivers (что важно)
- Консистентность процессов.
- Простота эксплуатации.
- Гибкость схемы + векторный поиск.

## Рассмотренные варианты
### Вариант A: PostgreSQL + JSONB + pgvector
- Плюсы: единый backend, транзакции, зрелый tooling.
- Минусы: нагрузка разных профилей в одной БД.

### Вариант B: PostgreSQL + отдельный vector DB
- Плюсы: изоляция vector workload.
- Минусы: сложнее ops и консистентность.

### Вариант C: event store + отдельные read models
- Плюсы: богатая event-driven архитектура.
- Минусы: слишком дорого для MVP.

## Решение
Мы выбираем: **Вариант A**.

## Обоснование (Rationale)
Один backend даёт предсказуемость и короткий путь к продакшн-пригодному MVP.

## Последствия (Consequences)
### Позитивные
- Простая синхронизация pod'ов через БД.
- Единая модель миграций, backup, мониторинга.

### Негативные / компромиссы
- Нужна аккуратная настройка индексов и retention.

### Технический долг
- Возможное выделение vector workload в отдельный контур позже.

## План внедрения (минимально)
- Миграции таблиц state/audit/docs/chunks.
- Подключение `pgvector` extension и индексов.
- Базовые retention/archiving jobs.

## План отката/замены
- Условия отката: неприемлемая деградация по latency/стоимости.
- Как откатываем: вынос vector в отдельное хранилище через новый ADR.

## Ссылки
- Data Model: `docs/architecture/data_model.md`

## Апрув
- request_id: N/A
- Решение: approved
- Комментарий:
