---
doc_id: EPC-CK8S-S2-D7
type: epic
title: "Epic S2 Day 7: Dogfooding regression gate for MVP readiness"
status: completed
owner_role: EM
created_at: 2026-02-10
updated_at: 2026-02-13
related_issues: [19]
related_prs: [20, 22, 23]
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-13-s2-day7"
---

# Epic S2 Day 7: Dogfooding regression gate for MVP readiness

## TL;DR
- Цель эпика: подтвердить, что S2 baseline + Day6 hardening готовы к расширению MVP в Sprint S3.
- Ключевая ценность: снимаем риски регрессий перед включением полного набора stage-flow и `run:self-improve`.
- MVP-результат: воспроизводимый regression bundle, формальный go/no-go и зафиксированный backlog Sprint S3 Day1..Day12.

## Priority
- `P0`.

## Scope
### In scope
- Regression matrix по текущему контуру:
  - `run:dev` -> run -> job -> PR;
  - `run:dev:revise` -> changes -> update PR;
  - отказ при конфликтных `ai-model` / `ai-reasoning` labels;
  - отказ privileged MCP операций без required approval.
- Операционная модель выполнения:
  - процесс ведут агенты end-to-end;
  - Owner участвует в контрольных точках и принимает `approve/deny` решения по запросам агента.
- Regression matrix по Day6 control tools:
  - deterministic secret sync (GitHub + Kubernetes) с проверкой idempotency;
  - database create/delete по окружению;
  - owner feedback request (варианты + custom input).
- Проверка staff observability:
  - список running jobs;
  - исторические логи и flow events;
  - wait queue (`waiting_mcp`, `waiting_owner_review`) и причина ожидания.
- Проверка runtime hygiene:
  - отсутствие утечек namespace/job/slot после успешных/ошибочных прогонов;
  - поведение `run:debug` (cleanup skip + audit evidence).
- Документационный gate:
  - синхронизация product/architecture/delivery docs с расширенным MVP scope;
  - готовый Sprint S3 plan с 7-12 эпиками (в этом пакете: Day1..Day12).

### Out of scope
- Полный e2e regression по всем `run:*` стадиям до их реализации в Sprint S3.
- Production rollout; проверка ограничивается staging/dev dogfooding средой.

## Критерии приемки эпика
- Regression matrix и evidence опубликованы и воспроизводимы на staging.
- Нет открытых `P0` блокеров для старта Sprint S3.
- Зафиксирован go/no-go протокол и список рисков/долгов с owner decision.

## Фактический результат (выполнено)
- Подготовлен и опубликован regression bundle:
  - `docs/delivery/regression_s2_gate.md`.
- Подтверждена рабочая матрица dogfooding baseline:
  - по данным `agent_runs`: `run:dev` (`succeeded=9`, `failed=17`) и `run:dev:revise` (`succeeded=3`);
  - по данным `flow_events`: `run.pr.created=5`, `run.pr.updated=3`.
- Подтверждена observability и runtime hygiene:
  - staging deploy на `main` успешен (workflow run `21985095587`);
  - на момент gate отсутствуют активные run namespaces (`codex-k8s.dev/namespace-purpose=run`);
  - `run:debug` режим фиксируется аудитом (`run.namespace.cleanup_skipped=5`).
- Подтверждена регрессия label-конфликтов и приоритета config labels тестами worker-домена.
- Approval queue консистентна: зависших `mcp_action_requests` в `requested` состоянии нет.

## Go/No-Go (2026-02-13)
- Решение: **Go** для старта Sprint S3.
- Обоснование: P0 блокеры в S2 контуре не выявлены, Day6 hardening и Day7 regression gate завершены.
