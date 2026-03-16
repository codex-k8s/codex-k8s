---
doc_id: ALT-0007
type: alternatives
title: "Quality Governance System — alternatives and trade-offs"
status: in-review
owner_role: SA
created_at: 2026-03-15
updated_at: 2026-03-15
related_issues: [466, 469, 470, 471, 476, 484, 488, 494]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-15-issue-484-arch"
---

# Alternatives & Trade-offs: Quality Governance System

## TL;DR
- Рассмотрели: GitHub-native/docs-only governance / dedicated governance service now / control-plane-owned aggregate with worker reconciliation.
- Рекомендуем: `control-plane` как owner canonical governance aggregate + `worker` для asynchronous reconciliation.
- Почему: лучший баланс архитектурной консистентности, auditability, proportionality и скорости handover в `run:design`.

## Контекст
- PRD Sprint S13 закрепил explicit risk tier, mandatory evidence package, verification minimum, waiver discipline, publication path `internal working draft -> semantic waves -> published waves` и boundary `Sprint S13 -> Sprint S14`.
- Нельзя нарушить:
  - thin-edge rule для `api-gateway` и `web-console`;
  - separate constructs `risk/evidence/verification/waiver/publication`;
  - no-silent-waiver policy для `high/critical`;
  - markdown-only scope на `run:arch`.

## Вариант A: GitHub-native/docs-only governance
- Описание:
  - считать canonical state комбинацией issue/PR comments, labels, markdown-артефактов и service-message narrative.
- Плюсы:
  - минимальный стартовый overhead;
  - не нужно вводить новый aggregate даже на уровне design.
- Минусы:
  - policy semantics расползаются между несколькими surfaces;
  - невозможно доказать separate constructs без дополнительного слоя интерпретации.
- Риски:
  - raw draft leakage в review stream;
  - inconsistent owner/operator visibility;
  - S14 начнёт переоткрывать semantics через UI/runtime path.
- Стоимость/сложность:
  - низкая на старте, высокая в стабилизации и аудите.

## Вариант B: Dedicated governance service уже сейчас
- Описание:
  - выделить новый runtime/service boundary и отдельный owner schema для `Quality Governance System` на Day4.
- Плюсы:
  - чистый отдельный bounded context;
  - прямой future scaling path.
- Минусы:
  - premature split до typed contracts и verified load signals;
  - дополнительный rollout, schema governance и integration burden в середине markdown-only stage.
- Риски:
  - задержка `run:design` и `run:plan`;
  - лишний coordination debt между `control-plane`, `worker` и новым сервисом.
- Стоимость/сложность:
  - высокая.

## Вариант C: Control-plane-owned aggregate + worker reconciliation (recommended)
- Описание:
  - `control-plane` владеет canonical governance semantics, publication gate и projections;
  - `worker` исполняет asynchronous sweeps, пишет reconciliation/evidence state и запрашивает policy-aware re-evaluation, а late reclassification / gap closure фиксирует `control-plane`;
  - `api-gateway`, `web-console`, `agent-runner` работают как typed adapters.
- Плюсы:
  - сохраняет текущие service boundaries;
  - даёт единый owner для policy semantics и audit;
  - готовит clean handover в `run:design` без преждевременного runtime split.
- Минусы:
  - увеличивает доменную ответственность `control-plane`;
  - требует аккуратной design-stage формализации aggregate/projections.
- Риски:
  - future extraction может всё равно понадобиться при росте throughput.
- Стоимость/сложность:
  - средняя.

## Сравнение
| Критерий | A | B | C |
|---|---:|---:|---:|
| Скорость Day4 handover | 5 | 1 | 4 |
| Архитектурная консистентность | 2 | 4 | 5 |
| Auditability | 2 | 4 | 5 |
| Риск premature lock-in | 3 | 1 | 4 |
| Простота перехода в `run:design` | 2 | 2 | 5 |
| Готовность к Sprint S14 boundary | 2 | 4 | 5 |

## Рекомендация
- Выбор: **Вариант C**.
- Обоснование:
  - сохраняет canonical owner для policy semantics без new-service overhead;
  - удерживает thin-edge и downstream boundary `S13 -> S14`;
  - делает design-stage focused на typed contracts/data model вместо переоткрытия ownership.
- Что теряем:
  - часть simplicity варианта A и потенциальный early extraction варианта B.
- Что выигрываем:
  - воспроизводимый governance state, меньший policy drift и быстреее продвижение в `run:design`.

## Нужен апрув от Owner
- [ ] Выбор варианта C.
- [ ] Разрешение компромисса: `control-plane` временно принимает ownership aggregate, а возможный dedicated service откладывается до доказанных scale-сигналов.
