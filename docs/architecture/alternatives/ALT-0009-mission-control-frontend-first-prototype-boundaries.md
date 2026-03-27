---
doc_id: ALT-0009
type: alternatives
title: "Mission Control frontend-first prototype — alternatives and trade-offs"
status: in-review
owner_role: SA
created_at: 2026-03-27
updated_at: 2026-03-27
related_issues: [480, 561, 562, 563, 565, 567, 571, 573]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-27-issue-571-arch"
---

# Alternatives & Trade-offs: Mission Control frontend-first prototype

## TL;DR
- Рассмотрели: backend-first rebuild now / prototype over current APIs and partial live sync / isolated `web-console` fake-data slice with explicit backend handover.
- Рекомендуем: isolated `web-console` fake-data slice + documented downstream seam к `#563`.
- Почему: лучший баланс архитектурной консистентности, Sprint S18 guardrails и скорости handover в `run:design`.

## Контекст
- Sprint S18 уже зафиксировал:
  - fullscreen свободный canvas;
  - taxonomy `Issue` / `PR` / `Run`;
  - compact nodes, explicit relations, drawer, toolbar и workflow preview на fake data;
  - platform-safe actions only;
  - repo-seed prompts как source of truth;
  - separate backend rebuild `#563`.
- Нельзя нарушить:
  - thin-edge rule для `api-gateway`;
  - отсутствие нового persisted source-of-truth path в Sprint S18;
  - запрет DB prompt editor и live provider mutation path;
  - continuity `arch -> design -> plan -> dev`.

## Вариант A: Backend-first rebuild now
- Описание:
  - немедленно перейти к provider mirror, persisted `Issue/PR/Run` truth и workflow definitions в рамках текущего Sprint S18.
- Плюсы:
  - быстрее начать долгий backend path.
- Минусы:
  - UX baseline снова становится функцией backend assumptions;
  - Sprint S18 смешивается с `#563`.
- Риски:
  - возврат к superseded S16 reasoning;
  - owner review теряет фокус на UX.
- Стоимость/сложность:
  - высокая.

## Вариант B: Prototype поверх current APIs и partial live sync
- Описание:
  - оставить fake-data UX, но постепенно подпитывать его текущими API или live provider reads.
- Плюсы:
  - кажется быстрым компромиссом.
- Минусы:
  - невозможно честно провести границу между prototype и canonical truth;
  - safe-action boundary начинает размываться.
- Риски:
  - implicit live mutation path;
  - drift в stale/freshness semantics и prompt-policy ownership.
- Стоимость/сложность:
  - средняя, но с высоким rework risk.

## Вариант C: Isolated `web-console` fake-data slice + explicit backend handover (recommended)
- Описание:
  - весь Wave 1 prototype живёт в `web-console`;
  - repo seeds дают policy wording;
  - backend rebuild остаётся отдельным flow `#563`.
- Плюсы:
  - сохраняет frontend-first sequencing;
  - не ломает service boundaries;
  - делает handover к `#563` explicit и audit-friendly.
- Минусы:
  - нужно отдельно документировать replacement seam от fake data к persisted truth.
- Риски:
  - если seam будет описан слишком слабо, `#563` получит неоднозначный handover.
- Стоимость/сложность:
  - средняя и контролируемая.

## Сравнение
| Критерий | A | B | C |
|---|---:|---:|---:|
| Сохранение frontend-first sequencing | 1 | 2 | 5 |
| Архитектурная консистентность | 2 | 2 | 5 |
| Риск скрытого backend lock-in | 1 | 2 | 5 |
| Ясность owner review | 2 | 3 | 5 |
| Простота handover в `run:design` | 2 | 3 | 5 |
| Соответствие prompt-policy guardrails | 2 | 2 | 5 |

## Рекомендация
- Выбор: **Вариант C**.
- Обоснование:
  - он сохраняет уже утверждённые Day1-Day3 решения и не даёт backend rebuild стать скрытым prerequisite;
  - он удерживает `api-gateway`, `control-plane` и `worker` в правильных границах;
  - он позволяет `run:design` сфокусироваться на explicit contracts и state slices вместо переоткрытия ownership.
- Что теряем:
  - раннюю проверку backend path кодом.
- Что выигрываем:
  - чистый UX baseline, прозрачный deferred scope и воспроизводимый handover в `#563`.

## Нужен апрув от Owner
- [ ] Выбор варианта C.
- [ ] Подтверждение компромисса: Sprint S18 `run:dev` остаётся isolated prototype без live provider path и без обязательного backend prerequisite.
