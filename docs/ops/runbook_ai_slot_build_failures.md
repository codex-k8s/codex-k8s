---
doc_id: RB-CK8S-AISLOT-BUILD-0001
type: runbook
title: "AI Slot Build Runbook: cache/MANIFEST_UNKNOWN и codegen-check build-ref errors"
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

# Runbook: AI slot build failures (cache + build-ref)

## TL;DR
- Симптом A: сборка run в ai-слоте падает в Kaniko c `MANIFEST_UNKNOWN` при `retrieving image from cache`.
- Симптом B: `codegen-check` job падает на `git checkout --detach` из-за некорректного `CODEXK8S_BUILD_REF` (например, ref с префиксом `-b`).
- Быстрая диагностика: проверить логи `codex-k8s-control-plane`/`codex-k8s-worker` и последние `build/mirror/codegen-check` jobs.
- Быстрое восстановление:
  - для cache-сигнатуры: `CODEXK8S_KANIKO_CACHE_ENABLED=false`;
  - для build-ref сигнатуры: нормализовать `CODEXK8S_BUILD_REF` до валидного git ref без CLI-флагов.

## Когда использовать
- Trigger: алерт из `docs/ops/alerts_ai_slot_build_pipeline.md`.
- Триггеры вручную:
  - в issue/run сообщается о повторяющемся падении сборок в ai-слоте;
  - в логах встречается `MANIFEST_UNKNOWN` или `retrieving image from cache`;
  - в `codegen-check` логах есть `unknown switch 'b'`, `is not a commit` или `checkout --detach`.
- Для инцидента Issue #205 подтверждена сигнатура build-ref: `codegen-check` падает на `git checkout --detach` при ref с префиксом `-b`.

## Предпосылки/доступы
- Нужен доступ `kubectl` в production namespace (`codex-k8s-prod`).
- В некоторых runtime-профилях `pods/exec` в БД-pod может быть закрыт RBAC; в этом случае диагностика ведётся по логам control-plane/worker и статусам jobs.

## Симптомы
- `build/mirror/codegen-check` job завершается `Failed`.
- В логах есть строки:
  - `Error while retrieving image from cache ... MANIFEST_UNKNOWN`;
  - `MANIFEST_UNKNOWN` при mirror/cache шагах.
  - `git checkout --detach ...`;
  - `unknown switch 'b'`;
  - `fatal: '-b ...' is not a commit`.
- Production сервисы могут оставаться healthy, но новые ai-slot сборки не стартуют или быстро падают на build/check стадии.

## Диагностика (пошагово)
1) Проверить общий статус runtime:

```bash
ns="codex-k8s-prod"
kubectl -n "$ns" get pods,job -o wide
```

2) Проверить cache/mirror ошибки в control-plane:

```bash
ns="codex-k8s-prod"
kubectl -n "$ns" logs deploy/codex-k8s-control-plane --since=6h \
  | grep -E "MANIFEST_UNKNOWN|retrieving image from cache|kaniko|mirror"
```

3) Проверить ошибки checkout/build-ref в codegen-check:

```bash
ns="codex-k8s-prod"
kubectl -n "$ns" logs job/codex-k8s-codegen-check --tail=200 2>/dev/null \
  | grep -E "checkout --detach|unknown switch|is not a commit|CODEXK8S_BUILD_REF" || true
kubectl -n "$ns" logs deploy/codex-k8s-control-plane --since=6h \
  | grep -E "codegen-check|CODEXK8S_BUILD_REF|checkout --detach|unknown switch|is not a commit" || true
```

4) Проверить worker логи на run-level симптомы:

```bash
ns="codex-k8s-prod"
kubectl -n "$ns" logs deploy/codex-k8s-worker --since=6h \
  | grep -E "MANIFEST_UNKNOWN|retrieving image from cache|checkout --detach|unknown switch|build failed|run_id"
```

5) Проверить неуспешные jobs:

```bash
ns="codex-k8s-prod"
kubectl -n "$ns" get jobs --sort-by=.metadata.creationTimestamp \
  | grep -E "kaniko|mirror|build|codegen-check|Failed|Error" || true
```

## Митигирование (восстановление)
### A) Cache-сигнатура (`MANIFEST_UNKNOWN`)
1) Выставить `CODEXK8S_KANIKO_CACHE_ENABLED=false` в runtime-конфигурации.
2) Убедиться, что rollout `codex-k8s-control-plane` завершён.
3) Повторить проблемную задачу/деплой в ai-слоте.

### B) Build-ref сигнатура (`git checkout --detach`)
1) Проверить `CODEXK8S_BUILD_REF` в runtime-конфигурации.
2) Нормализовать значение до валидного git ref:
   - допустимо: `main`, `feature/<branch>`, commit SHA;
   - запрещено: передавать shell/git CLI флаги (`-b`, `--...`) в значении ref.
3) Повторить codegen-check/runtime deploy и убедиться, что job завершается `Complete`.

Оперативная проверка:

```bash
ns="codex-k8s-prod"
kubectl -n "$ns" rollout status deploy/codex-k8s-control-plane --timeout=180s
kubectl -n "$ns" logs deploy/codex-k8s-control-plane --since=20m \
  | grep -E "MANIFEST_UNKNOWN|retrieving image from cache|checkout --detach|unknown switch|is not a commit" || true
kubectl -n "$ns" logs job/codex-k8s-codegen-check --tail=200 2>/dev/null \
  | grep -E "checkout --detach|unknown switch|is not a commit" || true
```

## Эскалация
- Эскалировать, если:
  - после отключения cache ошибка повторяется более 2 раз подряд;
  - после нормализации `CODEXK8S_BUILD_REF` ошибка checkout повторяется более 2 раз подряд;
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
- Новые ai-slot сборки и codegen-check завершаются успешно.
- В окне 30 минут отсутствуют новые `MANIFEST_UNKNOWN` и checkout/build-ref ошибки.
- В issue/инциденте подтверждён recovery и закрыты активные алерты.

## Пост-действия
- Обновить issue с root cause и таймлайном восстановления.
- Сверить алерты и пороги в `docs/ops/alerts_ai_slot_build_pipeline.md`.
- Сверить SLO-бюджет в `docs/ops/slo_ai_slot_build_pipeline.md`.
- Зафиксировать в RCA безопасный формат `CODEXK8S_BUILD_REF` (без CLI-флагов).
