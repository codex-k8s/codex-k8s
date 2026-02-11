---
doc_id: EPC-CK8S-S1-D6
type: epic
title: "Epic Day 6: Staging hardening, network controls, observability"
status: completed
owner_role: EM
created_at: 2026-02-06
updated_at: 2026-02-11
related_issues: [1]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-06-day6"
---

# Epic Day 6: Staging hardening, network controls, observability

## TL;DR
- Цель эпика: минимально безопасный и наблюдаемый staging для ежедневных релизов.
- Ключевая ценность: снижение риска регрессий и открытых наружу сервисов.
- MVP-результат: внешне только `22/80/443`, базовые метрики/логи/алерты, DNS/TLS precheck.

## Priority
- `P1` (операционная надёжность staging).

## Ожидаемые артефакты дня
- DNS preflight и TLS issuance сценарии в `bootstrap/remote/*` и/или deploy scripts.
- ClusterIssuer/Ingress/Network baseline manifests в `deploy/**`.
- Проверка внешних портов (`22/80/443`) и фиксация результата.
- Базовые observability dashboards/метрики/логи для webhook-run pipeline.

## Контекст
- Почему эпик нужен: ежедневный deploy требует operational guardrails.
- Связь с требованиями: NFR-001, NFR-006, NFR-007.

## Scope
### In scope
- Проверка DNS резолва домена до deploy.
- TLS issuance через cert-manager ClusterIssuer (http-01) для staging domain.
- Сетевой baseline и firewall профиль.
- Базовая observability: health endpoints, structured logs, key metrics.

### Out of scope
- Полный production SOC/SIEM.
- Сложные multi-namespace network policies для всех будущих tenant namespaces.

## Декомпозиция (Stories/Tasks)
- Story-1: DNS preflight checks в bootstrap/deploy path.
- Story-2: cert-manager ClusterIssuer и проверка выпуска сертификата.
- Story-3: firewall rules и проверка открытых портов.
- Story-4: базовый dashboard/logging/metrics для webhook-run pipeline.

## Data model impact (по шаблону data_model.md)
- Сущности:
  - `flow_events`: расширение event типов для ops событий bootstrap/deploy/health.
- Связи/FK:
  - Без новых FK.
- Индексы и запросы:
  - Проверить индекс `flow_events(event_type, created_at)`.
- Миграции:
  - Добавить миграцию только если требуется новый справочник event_type.
- Retention/PII:
  - Ops-события без секретов, ключи/токены маскируются.

## Критерии приемки эпика
- До начала раскатки скрипт валидирует DNS и падает при ошибке.
- TLS сертификат staging домена успешно выпущен и применён в ingress.
- Наружу доступны только `22/80/443`.
- Изменения задеплоены и проверены на staging в день реализации.

## Риски/зависимости
- Зависимости: корректная DNS запись на staging IP.
- Риск: задержки обновления DNS/propagation.

## План релиза (верхний уровень)
- В день внедрения выполнить проверку портов и TLS из внешней сети.

## Апрув
- request_id: owner-2026-02-06-day6
- Решение: approved
- Комментарий: Day 6 scope принят.
