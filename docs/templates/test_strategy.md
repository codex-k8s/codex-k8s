---
doc_id: TST-STR-XXXX
type: test-strategy
title: "<Система/Фича> — Test Strategy"
status: draft
owner_role: QA
created_at: YYYY-MM-DD
updated_at: YYYY-MM-DD
related_issues: []
related_prs: []
approvals:
  required: ["CTO"]
  status: approved
  request_id: "owner-2026-02-19-full-docset"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-19
---

# Test Strategy: <Система/Фича>

## TL;DR
- Подход:
- Уровни тестирования:
- Автоматизация:
- DNS-path coverage для новых/изменённых HTTP-ручек:
- Риски:

## Цели тестирования
- ...

## Область тестирования
### Что тестируем
- ...

### Что не тестируем
- ...

## Уровни и виды тестирования
- Unit:
- Integration:
- E2E:
- Manual acceptance: для новых/изменённых HTTP-ручек использовать Kubernetes service DNS path, а не только browser/OAuth flow:
- Performance:
- Security:
- Chaos (если нужно):

## Окружения
- Dev slot:
- Full-env namespace/service DNS:
- Production:
- Prod (только smoke/наблюдение):

## Тестовые данные
- Генерация:
- Маскирование PII:

## Инструменты и CI
- Линтеры:
- Тест раннеры:
- Отчеты:

## Критерии входа/выхода
### Entry criteria
- [ ] ...
- [ ] Для каждой новой/изменённой HTTP-ручки определён service DNS path и ожидаемый status code.

### Exit criteria
- [ ] ...
- [ ] Для применимых HTTP-ручек приложен DNS evidence bundle: namespace, FQDN, команда, status, excerpt ответа, timestamp, ссылка на issue/PR/checklist.

## Риски и меры
- ...

## Апрув
- request_id: ...
- Решение:
- Комментарий:
