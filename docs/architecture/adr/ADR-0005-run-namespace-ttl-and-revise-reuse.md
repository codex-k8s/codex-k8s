---
doc_id: ADR-0005
type: adr
title: "Run namespace TTL retention and revise lease extension"
status: accepted
owner_role: SA
created_at: 2026-02-20
updated_at: 2026-03-14
related_issues: [74, 461]
related_prs: []
supersedes: []
superseded_by: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-20-issue-74-namespace-ttl"
---

# ADR-0005: Run namespace TTL retention and revise lease extension

## TL;DR
- Контекст: run-namespace удаляется сразу после завершения run, что ломает ревью и диагностику.
- Решение: сохранять `full-env` run namespace по role-based TTL из `services.yaml` (default `24h`).
- Для `run:*:revise`: переиспользовать namespace текущей связки `(project, issue, agent_key)` и продлевать lease.
- Отдельный debug-label для manual-retention удаляется как избыточный: retention управляется TTL lease-политикой.

## Контекст
- Текущий baseline после S2 Day3:
  - namespace создаётся на run;
  - после run удаляется.
- Проблема из Issue #74:
  - после завершения run пропадает среда, невозможно пройти на домен/слот и проверить результат;
  - при revise нет гарантированного продолжения в том же runtime-контексте.
- Дополнительный драйвер:
  - требуется управляемая retention-policy, а не бесконечное накопление namespace.

## Decision Drivers (что важно)
- Сохраняемость среды на период ревью.
- Детерминированное поведение revise.
- Ограничение операционных затрат через TTL cleanup.
- Совместимость с существующей security/policy моделью (`managed namespace`, RBAC, audit).

## Рассмотренные варианты
### Вариант A: текущая модель (immediate cleanup без TTL lease)
- Плюсы: минимальные ресурсы.
- Минусы: неудобный review/revise, высокий ручной overhead.
- Риск: деградация воспроизводимости.

### Вариант B: хранить namespace бессрочно
- Плюсы: максимальная дебаг-удобность.
- Минусы: гарантированные утечки ресурсов.
- Риск: операционная деградация кластера.

### Вариант C: role-based TTL + lease extension на revise
- Плюсы: баланс review-удобства и контролируемого cleanup.
- Минусы: нужна lease-модель и sweep-cleanup.
- Риск: ошибки времени жизни при гонках/повторных запусках.

## Решение
Выбран **Вариант C**.

### Контракт `services.yaml` (добавление)
В `spec.webhookRuntime` вводится role-based TTL policy:

```yaml
webhookRuntime:
  defaultNamespaceTTL: 24h
  namespaceTTLByRole:
    pm: 24h
    sa: 24h
    em: 24h
    dev: 24h
    reviewer: 24h
    qa: 24h
    sre: 24h
    km: 24h
```

Правила:
- применяется только для `full-env` run;
- TTL роли имеет приоритет над `defaultNamespaceTTL`;
- если policy отсутствует, fallback = `24h`.

### Namespace identity и revise reuse
- Ключ reuse: `(project, issue_number, agent_key)`.
- На `run:*:revise`:
  - worker сначала валидирует managed namespace по persisted runtime fingerprint;
  - в fingerprint входят как минимум `project_id`, `issue_number`, `agent_key`, `runtime_mode`, `target_env`,
    `repository_full_name`, `services_yaml_path`, immutable `build_ref`, `deploy_only` и hash rendered manifests;
  - fast-path reuse разрешён только если fingerprint совпадает, namespace не `Terminating`
    и в нём нет активной `runtime_deploy_task`;
  - при положительной проверке build/apply path пропускается и reuse работает без нового runtime deploy task;
  - lease продлевается: `expires_at = now + role_ttl`;
  - при любой инвалидации (`fingerprint_missing|fingerprint_mismatch|repo_snapshot_stale|namespace_terminating|active_runtime_deploy_task|...`)
    worker фиксирует audit evidence и делает обычный runtime redeploy в тот же namespace;
  - если namespace отсутствует/неконсистентен -> создаётся новый namespace.

### Cleanup policy
- В Kubernetes нет built-in TTL для namespace (TTL-after-finished работает только для Job).
- Cleanup реализуется sweeper-контуром по managed namespace:
  - in-band sweep в worker reconcile tick;
  - отдельный production `CronJob` `codex-k8s-worker-namespace-cleanup` как backstop при сбоях/простоях worker;
  - отбор по `codex-k8s.dev/namespace-purpose=run`;
  - guardrails: ownership-label + allowlist platform runtime namespace names (issue-run prefix + slot namespaces `codex-k8s-dev-*`) + отсутствие non-terminal run в БД + отсутствие active workload в namespace, включая unsuspended `CronJob`;
  - удаление по достижении lease expiry;
  - write-audit на каждое действие.

### Retention без manual-retention label
- Отдельный debug-label для бессрочного удержания namespace не поддерживается.
- Единая модель retention:
  - `full-env` namespace живёт по TTL роли;
  - на `run:*:revise` lease продлевается.

## Обоснование (Rationale)
- Решение закрывает основной UX-gap Issue #74 без бесконтрольного роста ресурсов.
- Role-based policy не ломает текущую модель ролей и позволяет точечно менять TTL без изменения кода.
- Reuse на revise повышает воспроизводимость и уменьшает время прогонов.

## Последствия (Consequences)
### Позитивные
- Сохраняемая среда для review/debug минимум на `24h` (или больше по policy).
- Revise продолжает работу в том же namespace-контексте.
- Cleanup остаётся управляемым и аудируемым.

### Негативные / компромиссы
- Выше среднее потребление ресурсов кластера.
- Нужны дополнительные guardrails против reuse повреждённого namespace.

### Технический долг
- Добавить в staff UI явное отображение lease (`expires_at`, источник TTL, режим reuse).
- Ввести пер-role policy presets в UI/API управления агентами/проектом.

## Data/Audit изменения
- `flow_events` (новые события):
  - `run.namespace.reuse_fast_path`,
  - `run.namespace.reuse_fallback_redeploy`,
  - `run.namespace.ttl_scheduled`,
  - `run.namespace.ttl_extended`,
  - `run.namespace.cleaned` (reason=`ttl_expired`),
  - `run.namespace.cleanup_skipped`,
  - `run.namespace.cleanup_failed`.
- `run_payload`/runtime metadata (минимум):
  - `namespace_lease_ttl`,
  - `namespace_lease_expires_at`,
  - `namespace_reused` (bool).

## План внедрения
1. Расширить typed loader `services.yaml` новыми полями TTL policy.
2. Обновить worker namespace lifecycle:
   - schedule lease на create;
   - extend lease на revise reuse;
   - sweep-cleanup по expiry.
3. Обновить сообщения в issue/run status (TTL lease, `expires_at`, факт reuse при revise).
4. Обновить `services.yaml` проекта `codex-k8s`:
   - выставить `24h` для всех системных ролей.

## План миграции runtime
- Для существующих managed run namespace без lease:
  - проставить lease c fallback `24h` от момента миграции.
- Для незавершённых/текущих run:
  - lease вычисляется при следующем heartbeat/update события.

## План отката/замены
- Временный rollback:
  - установить минимальный TTL (например, `15m`) для ролей;
  - отключить revise reuse feature-flag и вернуться к create-per-run.
- Полный rollback:
  - возврат к immediate cleanup policy.

## Внешние ссылки
- Kubernetes TTL-after-finished (работает для Jobs, не для namespace):
  - https://kubernetes.io/docs/concepts/workloads/controllers/ttlafterfinished/
- Kubernetes namespace deletion behavior:
  - https://kubernetes.io/docs/tasks/administer-cluster/namespaces/
