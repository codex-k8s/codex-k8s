---
doc_id: EPC-CK8S-0001
type: epic
title: "Epic: Staging bootstrap and deploy loop"
status: draft
owner_role: EM
created_at: 2026-02-06
updated_at: 2026-02-06
related_issues: [1]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-06-mvp"
---

# Epic: Staging bootstrap and deploy loop

## TL;DR
- Цель эпика: получить воспроизводимый запуск staging и автоматический deploy из `main`.
- Ключевая ценность: минимальный операционный порог для начала ручных тестов.
- MVP-результат: bootstrap скрипт + self-hosted runner + deploy workflow + smoke checks.

## Контекст
- Почему эпик нужен: без staging невозможно проверять жизненный цикл webhook-driven платформы.
- Связь с PRD: базовая платформа должна быть deployable и testable ранним этапом.

## Scope
### In scope
- SSH bootstrap launcher с хоста разработчика.
- Создание отдельного пользователя на сервере и базовый hardening.
- Установка k3s/сети/зависимостей.
- Установка PostgreSQL и `codex-k8s`.
- Настройка GitHub runner в k8s для staging (local=1 persistent; server+domain=autoscaled set).
- Настройка deploy workflow и секретов.

### Out of scope
- Полная production hardening программа.
- DR multi-region.

## Декомпозиция (Stories/Tasks)
- Story-1: host-side bootstrap launcher (`bootstrap/host/bootstrap_remote_staging.sh`).
- Story-2: remote provisioning scripts (`bootstrap/remote/*.sh`) для Ubuntu 24.04.
- Story-3: k8s dependencies and base manifests.
- Story-4: ARC runner install + GitHub repo wiring.
- Story-5: staging deploy workflow + smoke test job.
- Story-6: runbook ручного восстановления.

## Критерии приемки эпика
- Один запуск скрипта с host машины разворачивает staging end-to-end.
- Runner онлайн, deploy workflow выполняется успешно.
- После push в `main` staging обновляется автоматически.

## Риски/зависимости
- Зависимости: GitHub fine-grained token с правами на repo/actions/secrets/variables, доступ root SSH.
- Риски: различия cloud image Ubuntu 24.04, сетевые ограничения провайдера.

## План релиза (верхний уровень)
- Выпустить как milestone "staging-ready".
- После стабилизации включить в обязательный gate перед расширением функционала.

## Апрув
- request_id: owner-2026-02-06-mvp
- Решение: approved
- Комментарий: Эпик staging bootstrap/deploy loop утверждён.
