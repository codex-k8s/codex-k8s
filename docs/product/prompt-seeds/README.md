# Prompt Seeds Catalog

Назначение:
- `prompt-seeds` — это базовые task-body шаблоны для runtime prompt.
- Финальный prompt формируется рантаймом поверх этих seed (envelope + context + policy).
- Каноническая модель шаблонов — role-specific (`agent_key + work/review + locale`); этот каталог используется как bootstrap/fallback слой.

Нейминг:
- `<stage>-work.md` — шаблон выполнения этапа.
- `<stage>-review.md` — шаблон ревизии этапа (`run:*:revise`).
- опционально поддерживается локализованный вариант: `<stage>-<kind>_<locale>.md`.

Текущий минимальный каталог:
- `intake-work.md`, `intake-review.md`
- `vision-work.md`, `vision-review.md`
- `prd-work.md`, `prd-review.md`
- `arch-work.md`, `arch-review.md`
- `design-work.md`, `design-review.md`
- `plan-work.md`, `plan-review.md`
- `dev-work.md`, `dev-review.md`
- `doc-audit-work.md`
- `qa-work.md`
- `release-work.md`
- `postdeploy-work.md`
- `ops-work.md`
- `self-improve-work.md`
- `rethink-work.md`

Важно:
- шаблон должен описывать цель этапа, обязательные шаги, ожидаемые артефакты и критерий завершения;
- секреты, токены и обход policy в шаблонах запрещены.
- stage-specific seed-файлы не отменяют requirement на отдельные role-specific body-шаблоны `work/review` в локалях минимум `ru` и `en`.
