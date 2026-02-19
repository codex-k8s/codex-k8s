# Prompt Seeds Catalog

Назначение:
- `prompt-seeds` — это базовые task-body шаблоны для runtime prompt.
- Финальный prompt формируется рантаймом поверх этих seed (envelope + context + policy).
- Каноническая модель шаблонов — role-specific (`agent_key + work/revise + locale`); этот каталог используется как bootstrap/fallback слой.

Нейминг:
- `<stage>-work.md` — шаблон выполнения этапа.
- `<stage>-revise.md` — шаблон ревизии этапа (`run:*:revise`).
- опционально поддерживается локализованный вариант: `<stage>-<kind>_<locale>.md`.
- role-aware варианты:
  - `<stage>-<agent_key>-<kind>_<locale>.md`;
  - `<stage>-<agent_key>-<kind>.md`;
  - `role-<agent_key>-<kind>_<locale>.md`;
  - `role-<agent_key>-<kind>.md`.

Порядок fallback (runtime):
1. stage+role+kind+locale;
2. stage+role+kind;
3. role+kind+locale;
4. role+kind;
5. stage+kind+locale;
6. stage+kind;
7. `dev`+kind+locale;
8. `dev`+kind;
9. `default`+kind+locale;
10. `default`+kind;
11. встроенные templates runner.

Текущий минимальный каталог:
- `intake-work.md`, `intake-revise.md`
- `vision-work.md`, `vision-revise.md`
- `prd-work.md`, `prd-revise.md`
- `arch-work.md`, `arch-revise.md`
- `design-work.md`, `design-revise.md`
- `plan-work.md`, `plan-revise.md`
- `dev-work.md`, `dev-revise.md`
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
- stage-specific seed-файлы не отменяют requirement на отдельные role-specific body-шаблоны `work/revise` в локалях минимум `ru` и `en`.
- role-specific baseline для поддержанных ролей:
  - `dev`, `pm`, `sa`, `em`, `reviewer`, `qa`, `sre`, `km` (каждая: `work/revise` и `ru/en`).
- backward compatibility:
  - runtime принимает legacy seed-файлы с `*-review*.md` как fallback для `kind=revise`.
- Для `self-improve-*` seed обязателен диагностический контур:
  - MCP `self_improve_runs_list` / `self_improve_run_lookup` / `self_improve_session_get`;
  - сохранение извлеченного `codex-cli` session JSON в `/tmp/codex-sessions/<run-id>`.
