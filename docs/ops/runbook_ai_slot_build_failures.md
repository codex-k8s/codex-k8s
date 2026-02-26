---
doc_id: RB-CK8S-AISLOT-BUILD-0001
type: runbook
title: "AI Slot Build Runbook: MANIFEST_UNKNOWN / cache retrieval errors"
status: active
owner_role: SRE
created_at: 2026-02-26
updated_at: 2026-02-26
related_issues: [205]
related_alerts: ["ALT-CK8S-AISLOT-BUILD-0001"]
approvals:
  required: ["Owner"]
  status: pending
  request_id: "issue-205-ai-repair"
---

# Runbook: AI slot build failures (`MANIFEST_UNKNOWN`)

## TL;DR
- Симптом: сборка run в ai-слоте падает в Kaniko c `MANIFEST_UNKNOWN` при `retrieving image from cache`.
- Быстрая диагностика: проверить `codex-k8s-control-plane`/`codex-k8s-worker` логи и неуспешные build/mirror jobs.
- Быстрое восстановление: отключить cache (`CODEXK8S_KANIKO_CACHE_ENABLED=false`), перезапустить control-plane и повторить deploy/run.

## Когда использовать
- Trigger: алерт из `docs/ops/alerts_ai_slot_build_pipeline.md`.
- Триггеры вручную:
  - в issue/run сообщается о повторяющемся падении сборок в ai-слоте;
  - в логах встречается `MANIFEST_UNKNOWN` или `retrieving image from cache`.

## Предпосылки/доступы
- Нужен доступ `kubectl` в production namespace (`codex-k8s-prod`).
- В некоторых runtime-профилях `pods/exec` в БД-pod может быть закрыт RBAC; в этом случае диагностика ведётся по логам control-plane/worker и статусам jobs.

## Симптомы
- build/deploy job завершается `Failed`.
- в логах есть строки:
  - `Error while retrieving image from cache ... MANIFEST_UNKNOWN`;
  - `MANIFEST_UNKNOWN` при mirror/cache шагах.
- production сервисы остаются healthy, но новые ai-slot сборки не стартуют или быстро падают.

## Диагностика (пошагово)
1) Проверить общий статус runtime:

```bash
ns="codex-k8s-prod"
kubectl -n "$ns" get pods,job -o wide
```

2) Проверить ошибки cache/mirror в control-plane:

```bash
ns="codex-k8s-prod"
kubectl -n "$ns" logs deploy/codex-k8s-control-plane --since=6h \
  | grep -E "MANIFEST_UNKNOWN|retrieving image from cache|kaniko|mirror"
```

3) Проверить worker логи на run-level симптомы:

```bash
ns="codex-k8s-prod"
kubectl -n "$ns" logs deploy/codex-k8s-worker --since=6h \
  | grep -E "MANIFEST_UNKNOWN|retrieving image from cache|build failed|run_id"
```

4) Проверить неуспешные jobs (если есть):

```bash
ns="codex-k8s-prod"
kubectl -n "$ns" get jobs --sort-by=.metadata.creationTimestamp \
  | grep -E "kaniko|mirror|build|Failed|Error" || true
```

## Митигирование (восстановление)
1) Применить kill switch cache:
   - выставить `CODEXK8S_KANIKO_CACHE_ENABLED=false` в runtime-конфигурации;
   - убедиться, что rollout `codex-k8s-control-plane` завершён.
2) Повторить проблемную задачу/деплой в ai-слоте.
3) Проверить, что новые build/mirror jobs завершаются `Complete`.

Оперативная проверка:

```bash
ns="codex-k8s-prod"
kubectl -n "$ns" rollout status deploy/codex-k8s-control-plane --timeout=180s
kubectl -n "$ns" logs deploy/codex-k8s-control-plane --since=20m \
  | grep -E "MANIFEST_UNKNOWN|retrieving image from cache" || true
```

## Эскалация
- Эскалировать, если:
  - после отключения cache ошибка повторяется более 2 раз подряд;
  - одновременно деградируют mirror jobs и registry availability;
  - время восстановления превышает 30 минут.
- Кому:
  - SRE on-call;
  - Owner (если нужен rollback релиза платформы).
- Что приложить:
  - run id/issue id;
  - фрагменты логов control-plane/worker;
  - список failed jobs и время первой/последней ошибки.

## План отката
- Использовать `docs/ops/rollback_plan_ai_slot_build_pipeline.md`.

## Проверка результата
- Новые ai-slot сборки завершаются успешно.
- В окне 30 минут отсутствуют новые `MANIFEST_UNKNOWN` в control-plane.
- В issue/инциденте подтверждён recovery и закрыты активные алерты.

## Пост-действия
- Обновить issue с root cause и таймлайном восстановления.
- Сверить алерты и пороги в `docs/ops/alerts_ai_slot_build_pipeline.md`.
- Сверить SLO-бюджет в `docs/ops/slo_ai_slot_build_pipeline.md`.
