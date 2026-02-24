---
doc_id: EPC-CK8S-S1-D0
type: epic
title: "Epic Day 0: Bootstrap baseline on clean Ubuntu 24.04"
status: completed
owner_role: EM
created_at: 2026-02-06
updated_at: 2026-02-06
related_issues: [1]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-19-full-docset"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-19
---

# Epic Day 0: Bootstrap baseline on clean Ubuntu 24.04

## TL;DR
- Цель эпика: поднять production с нуля одним bootstrap-сценарием.
- Ключевая ценность: доказанная воспроизводимость развёртывания платформы.
- MVP-результат: сервер подготовлен, k3s/база/платформа/runner/ingress работают.

## Priority
- `P0` (critical baseline).

## Ожидаемые артефакты дня
- Рабочие host/remote bootstrap scripts в `bootstrap/host/*` и `bootstrap/remote/*`.
- Базовые deploy manifests и production routing в `deploy/**`.
- Подтвержденный отчёт о bootstrap с чистого Ubuntu 24.04 (production ready).
- Актуализированные delivery-документы с фиксацией статуса `Day 0 completed`.

## Контекст
- Почему эпик нужен: без подтверждённого `Day 0` ежедневные инкременты небезопасны.
- Связь с требованиями: FR-016, NFR-006, NFR-007, NFR-008.

## Scope
### In scope
- SSH bootstrap с хоста разработчика под `root`.
- Создание отдельного системного пользователя и базовый hardening.
- Установка k3s, ingress, cert-manager, production namespace.
- Разворачивание PostgreSQL и `codex-k8s`.
- Настройка production deploy pipeline.

### Out of scope
- Production hardening полного уровня.
- Multi-region disaster recovery.

## Декомпозиция (Stories/Tasks)
- Story-1: `bootstrap/host/*` launcher и env loading.
- Story-2: `bootstrap/remote/*` provisioning на Ubuntu 24.04.
- Story-3: DNS/TLS baseline и ingress routing.
- Story-4: production workflow + runner wiring.

## Data model impact (по шаблону data_model.md)
- Сущности: без изменения доменной модели.
- Связи/FK: без изменений.
- Индексы и запросы: без изменений.
- Миграции: не требуются.
- Retention/PII: без изменений.

## Критерии приемки эпика
- С чистого Ubuntu 24.04 production поднимается bootstrap-скриптом.
- Внешне доступны только `22/80/443`.
- `main`-push запускает деплой на production.

## Риски/зависимости
- Зависимости: DNS, GitHub token, OpenAI key.
- Риск: cloud-image различается по дефолтным сетевым настройкам.

## План релиза (верхний уровень)
- Зафиксировать как baseline для всех следующих дней спринта.

## Апрув
- request_id: owner-2026-02-06-day0
- Решение: approved
- Комментарий: Day 0 выполнен и принимается как базовый этап.
