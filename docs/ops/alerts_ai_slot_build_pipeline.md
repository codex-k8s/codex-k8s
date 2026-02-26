---
doc_id: ALT-CK8S-AISLOT-BUILD-0001
type: alerts
title: "AI Slot Build Pipeline — Alerting"
status: active
owner_role: SRE
created_at: 2026-02-26
updated_at: 2026-02-26
related_slo_docs: ["SLO-CK8S-AISLOT-BUILD-0001"]
related_issues: [205]
approvals:
  required: ["Owner"]
  status: pending
  request_id: "issue-205-ai-repair"
---

# Alerting: AI Slot Build Pipeline

## TL;DR
- Paging alerts:
  - `AI_SLOT_BUILD_FAILURE_BURST`;
  - `AI_SLOT_BUILD_MANIFEST_UNKNOWN_PERSISTENT`.
- Ticket alerts:
  - `AI_SLOT_BUILD_FAILURE_RATE_ELEVATED`;
  - `AI_SLOT_BUILD_MTTR_SLO_RISK`.
- Все алерты ссылаются на `docs/ops/runbook_ai_slot_build_failures.md`.

## Принципы
- Алёрты должны отражать пользовательский impact: невозможность собрать/запустить задачу в ai-слоте.
- Paging только при повторяемой деградации или отсутствии автоматического восстановления.
- Любой paging алерт обязан иметь быстрый mitigation (cache kill switch + verification).

## Каталог алертов
| Alert | Тип | Условие | Окно | Порог | Действие | Runbook |
|---|---|---|---|---:|---|---|
| `AI_SLOT_BUILD_FAILURE_BURST` | page | failed build/mirror jobs | 15m | `>=3` | on-call SRE, mitigation в течение 10m | `docs/ops/runbook_ai_slot_build_failures.md` |
| `AI_SLOT_BUILD_MANIFEST_UNKNOWN_PERSISTENT` | page | `MANIFEST_UNKNOWN` в control-plane logs | 15m | `>=2` после mitigation | on-call SRE + Owner escalation | `docs/ops/runbook_ai_slot_build_failures.md` |
| `AI_SLOT_BUILD_FAILURE_RATE_ELEVATED` | ticket | failure ratio build jobs | 1h | `>5%` | создать issue и разобрать trend | `docs/ops/runbook_ai_slot_build_failures.md` |
| `AI_SLOT_BUILD_MTTR_SLO_RISK` | ticket | recovery time превышает целевой | 24h | `>30m` | постмортем и корректировка порогов | `docs/ops/rollback_plan_ai_slot_build_pipeline.md` |

## SLO burn rate alerting
- Контрольный SLO: `SLO-CK8S-AISLOT-BUILD-0001`.
- Short window: 15m.
- Long window: 6h и 24h.
- Пороги:
  - fast burn: 15m + 6h;
  - slow burn: 1h + 24h.
- При burn-rate нарушении запускается rollback decision flow.

## Anti-noise меры
- Дедупликация page-alert по `namespace + error_signature`.
- Silence на maintenance window релиза (явно ограниченный период).
- Авто-закрытие ticket-alert при восстановлении в пределах SLO окна.

## Открытые вопросы
- Нужна финальная унификация имен метрик build/mirror jobs в Prometheus правилах.
- Нужна привязка алертов к staff alert-center карточкам (если включено в окружении).
