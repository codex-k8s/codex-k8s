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
- Когда откатываем: повторяемые падения build/mirror/codegen-check в ai-слоте после mitigation.
- Как откатываем:
  - сначала config rollback (отключение cache / нормализация `CODEXK8S_BUILD_REF` / возврат стабильных runtime env);
  - затем при необходимости image rollback control-plane/worker.
- Как проверяем: успешные ai-slot сборки + отсутствие `MANIFEST_UNKNOWN` и checkout/build-ref ошибок в окне 30 минут.

## Триггеры rollback
- `AI_SLOT_BUILD_FAILURE_BURST` page-alert не гасится после mitigation.
- `AI_SLOT_BUILD_MANIFEST_UNKNOWN_PERSISTENT` повторяется.
- `AI_SLOT_CODEGEN_CHECK_BUILD_REF_INVALID` повторяется после нормализации ref.
- Нарушается SLO восстановления (MTTR > 30m).

## Предусловия
- Доступ к `kubectl` в `codex-k8s-prod`.
- Зафиксирован последний стабильный релиз control-plane/worker.
- Подтверждено, что проблема в build pipeline, а не в секретах/сети конкретного проекта.
- Подтвержден текущий класс отказа: cache-signature или build-ref-signature.

## Варианты отката
### 1) Конфигурационный rollback (предпочтительный)
- Шаги:
  1. Зафиксировать `CODEXK8S_KANIKO_CACHE_ENABLED=false`.
  2. Нормализовать `CODEXK8S_BUILD_REF` (валидный git ref без CLI-флагов).
  3. Проверить `rollout status` control-plane.
  4. Повторить ai-slot run.
- Риски:
  - рост времени сборки из-за отключенного cache.
  - возможный откат на менее свежий ref, если выбран fallback (`main`/stable SHA).

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
2. Выполнить config rollback:
   - `CODEXK8S_KANIKO_CACHE_ENABLED=false`;
   - `CODEXK8S_BUILD_REF=<валидный_ref_без_флагов>`.
3. Проверить логи и jobs:

```bash
ns="codex-k8s-prod"
kubectl -n "$ns" rollout status deploy/codex-k8s-control-plane --timeout=180s
kubectl -n "$ns" get jobs --sort-by=.metadata.creationTimestamp | tail -n 30
kubectl -n "$ns" logs deploy/codex-k8s-control-plane --since=30m \
  | grep -E "MANIFEST_UNKNOWN|retrieving image from cache|checkout --detach|unknown switch|is not a commit" || true
kubectl -n "$ns" logs job/codex-k8s-codegen-check --tail=200 2>/dev/null \
  | grep -E "checkout --detach|unknown switch|is not a commit|CODEXK8S_BUILD_REF" || true
```

4. Если ошибка сохраняется, согласовать image rollback с Owner.
5. После восстановления создать/обновить RCA issue.

## Верификация после rollback
- Проверки функционала:
  - новый ai-slot run проходит build/stage/codegen-check без аварий.
- Метрики:
  - failure rate возвращается в SLO-коридор.
- Логи:
  - нет новых `MANIFEST_UNKNOWN` и checkout/build-ref ошибок в окне 30m.
- Smoke:
  - basic run lifecycle (`queued -> running -> completed`) в ai-слоте.

## Коммуникации
- Уведомляются:
  - SRE on-call;
  - Owner;
  - при необходимости PM/EM для обновления статуса issue.
- Шаблон:
  - `Инцидент AI-slot build: rollback <type> выполнен, статус <green/yellow>, следующее обновление через <N> мин.`
