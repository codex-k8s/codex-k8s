---
doc_id: EPC-CK8S-S3-D9
type: epic
title: "Epic S3 Day 9: Declarative full-env deploy and runtime parity"
status: planned
owner_role: EM
created_at: 2026-02-13
updated_at: 2026-02-13
related_issues: [19]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
---

# Epic S3 Day 9: Declarative full-env deploy and runtime parity

## TL;DR
- Цель: реализовать детерминированное разворачивание полного окружения выполнения задач на основе декларативного `services.yaml`.
- MVP-результат: build/deploy orchestration переносится из shell-first подхода в YAML + бинарник `codex-k8s` с единым порядком зависимостей и runtime parity для non-prod.

## Priority
- `P0`.

## Scope
### In scope
- Декларативный inventory окружения на базе `services.yaml` (по референсам `codexctl` и `project-example`):
  - описание инфраструктуры, сервисов, образов, overlays и зависимостей в одном source of truth;
  - поддержка явного порядка развёртывания:
    `stateful dependencies -> migrations -> internal domain services -> edge services -> frontend`.
- Оркестрация через бинарник `codex-k8s`:
  - основной путь build/deploy/readiness переезжает из shell-скриптов в typed YAML-конфиг + Go-реализацию;
  - shell-скрипты остаются только как thin wrappers для вызова основного движка, и то в случае если без них не обойтись (кол-во скриптов должно быть сведено к минимуму).
- Runtime parity для non-prod (`dev/staging/ai-slot`):
  - frontend сервисы запускаются через hot-reload (`vite run dev`);
  - Go-сервисы запускаются через hot-reload (`CompileDaemon` или эквивалентный watcher);
  - Dockerfile/manifests фиксируют отдельные dev-target/runtime поведение.
- Shared workspace volume:
  - один и тот же PVC монтируется во все сервисы слота и в agent job для консистентного контекста исходников/артефактов;
  - read/write policy описывается явно в манифестах и проверяется на RBAC/namespace уровне.
- Reuse full-env namespace между итерациями ревью:
  - для `run:*:revise` namespace не пересоздаётся на каждую итерацию, если предыдущий full-env ещё активен;
  - вводится idle TTL для auto-cleanup warm namespace (default `8h`);
  - при истечении TTL следующий revise-trigger поднимает окружение заново и продолжает цикл.

### Out of scope
- Production-оптимизации (autoscaling tuning, cost-optimization).

## Критерии приемки
- Для минимум одного проекта full-env поднимается полностью из `services.yaml` без ручной правки shell-скриптов. С учетом dogfooding, это будет `codex-k8s`.
- Порядок deploy-этапов детерминированный и подтверждён audit/evidence.
- В non-prod окружениях подтверждён hot-reload для frontend и Go-сервисов.
- Shared PVC подключён ко всем сервисам слота и agent job без конфликтов прав доступа.
- Для revise-итераций подтверждён namespace reuse в пределах idle TTL и корректный auto-cleanup после TTL.

## Референсы
- [services.yaml reference](../project-example/services.yaml)
- [codexctl reference](../codexctl)
