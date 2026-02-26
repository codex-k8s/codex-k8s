---
doc_id: RLB-CK8S-AISLOT-BUILD-0001
type: rollback-plan
title: "AI Slot Build Pipeline — Rollback Plan"
status: active
owner_role: SRE
created_at: 2026-02-26
updated_at: 2026-02-26
related_issues: [205]
related_runbooks: ["RB-CK8S-AISLOT-BUILD-0001"]
approvals:
  required: ["Owner"]
  status: pending
  request_id: "issue-205-ai-repair"
---

# Rollback Plan: AI Slot Build Pipeline

## TL;DR
- Когда откатываем: повторяемые падения build/mirror в ai-слоте после mitigation.
- Как откатываем:
  - сначала config rollback (отключение cache / возврат стабильных runtime env);
  - затем при необходимости image rollback control-plane/worker.
- Как проверяем: успешные ai-slot сборки + отсутствие `MANIFEST_UNKNOWN` в окне 30 минут.

## Триггеры rollback
- `AI_SLOT_BUILD_FAILURE_BURST` page-alert не гасится после mitigation.
- `AI_SLOT_BUILD_MANIFEST_UNKNOWN_PERSISTENT` повторяется.
- Нарушается SLO восстановления (MTTR > 30m).

## Предусловия
- Доступ к `kubectl` в `codex-k8s-prod`.
- Зафиксирован последний стабильный релиз control-plane/worker.
- Подтверждено, что проблема в build pipeline, а не в секретах/сети конкретного проекта.

## Варианты отката
### 1) Конфигурационный rollback (предпочтительный)
- Шаги:
  1. Зафиксировать `CODEXK8S_KANIKO_CACHE_ENABLED=false`.
  2. Проверить `rollout status` control-plane.
  3. Повторить ai-slot run.
- Риски:
  - рост времени сборки из-за отключенного cache.

### 2) Rollback runtime image (control-plane/worker)
- Шаги:
  1. Откатить image tag control-plane и worker к последнему stable.
  2. Дождаться rollout complete.
  3. Перепроверить build/mirror jobs.
- Ограничения/риски:
  - возможная потеря новых, но не критичных фич релиза;
  - требуется синхронная коммуникация с Owner.

### 3) Kill switch only (кратковременная стабилизация)
- Шаги:
  1. Временно запретить cache/mirror path.
  2. Держать режим до завершения RCA.

## Пошаговая инструкция rollback
1. Подтвердить инцидент по runbook (`RB-CK8S-AISLOT-BUILD-0001`).
2. Выполнить config rollback (`CODEXK8S_KANIKO_CACHE_ENABLED=false`).
3. Проверить логи и jobs:

```bash
ns="codex-k8s-prod"
kubectl -n "$ns" rollout status deploy/codex-k8s-control-plane --timeout=180s
kubectl -n "$ns" get jobs --sort-by=.metadata.creationTimestamp | tail -n 30
kubectl -n "$ns" logs deploy/codex-k8s-control-plane --since=30m \
  | grep -E "MANIFEST_UNKNOWN|retrieving image from cache" || true
```

4. Если ошибка сохраняется, согласовать image rollback с Owner.
5. После восстановления создать/обновить RCA issue.

## Верификация после rollback
- Проверки функционала:
  - новый ai-slot run проходит build/stage без аварий.
- Метрики:
  - failure rate возвращается в SLO-коридор.
- Логи:
  - нет новых `MANIFEST_UNKNOWN` в окне 30m.
- Smoke:
  - basic run lifecycle (`queued -> running -> completed`) в ai-слоте.

## Коммуникации
- Уведомляются:
  - SRE on-call;
  - Owner;
  - при необходимости PM/EM для обновления статуса issue.
- Шаблон:
  - `Инцидент AI-slot build: rollback <type> выполнен, статус <green/yellow>, следующее обновление через <N> мин.`
