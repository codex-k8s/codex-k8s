# Prompt Seeds Catalog

Назначение:
- `prompt-seeds` — это базовые task-body шаблоны для runtime prompt.
- Финальный prompt формируется рантаймом поверх этих seed (envelope + context + policy).
- Role profile и контракты оформления Issue/PR/review/discussion берутся не из этого каталога,
  а из `services/jobs/agent-runner/internal/runner/templates/prompt_blocks/*.tmpl`.
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
- `ai-repair-work.md`
- `qa-work.md`
- `release-work.md`
- `postdeploy-work.md`
- `ops-work.md`
- `self-improve-work.md`
- `rethink-work.md`

Важно:
- шаблон должен описывать цель этапа, обязательные шаги, ожидаемые артефакты и критерий завершения;
- envelope-слой рантайма требует регулярный MCP feedback через `run_status_report` (каждые 3-4 инструментальных вызова, сразу после смены фазы и перед долгими операциями/сетевыми запросами/сборкой/ожиданием);
- для всех `*revise*` seed'ов обязателен pre-check: перед правками и перед commit/push повторно проверить новые комментарии и merge conflicts; при наличии — сначала разрешить их;
- для `*revise*` seed'ов перед финальным статусом/label transition обязателен preflight: проверить `mergeable` и отсутствие stage ambiguity по `run:*` label context (PR + Issue).
- секреты, токены и обход policy в шаблонах запрещены.
- для `run:intake|vision|prd|arch|design|plan|doc-audit|qa|release|postdeploy|ops|rethink` runtime policy ограничивает изменения только markdown-файлами (`*.md`);
- для роли `reviewer` runtime policy запрещает изменения репозитория (только комментарии в существующем PR);
- для `run:self-improve` runtime policy разрешает только: markdown-инструкции, prompt файлы и `services/jobs/agent-runner/Dockerfile`.
- stage-specific seed-файлы не отменяют requirement на отдельные role-specific body-шаблоны `work/revise` в локалях минимум `ru` и `en`.
- role-specific baseline для поддержанных ролей:
  - `dev`, `pm`, `sa`, `em`, `reviewer`, `qa`, `sre`, `km` (каждая: `work/revise` и `ru/en`).
- role-aware пути к шаблонам артефактов берутся из `services.yaml/spec.roleDocTemplates`;
  в seed-файлах указываются только имена шаблонов (без жестко заданных repository-relative путей).
- `services.yaml/spec.roleDocTemplates` и `services.yaml/spec.projectDocs` должны оставаться синхронными
  с `docs/delivery/development_process_requirements.md` (role-template matrix + doc IA).
- `templates/prompt_blocks/issue_contract_*.tmpl` и `pr_contract_*_*.tmpl` должны оставаться синхронными
  с разделом title/body contract в `docs/delivery/development_process_requirements.md`;
  для role-bearing title optional-фрагмент `Sprint S<спринт> Day<день>` вставляется сразу после role token,
  если sprint/day известны из текущего stage context.
  Исключение допускается только для markdown-only trigger, когда prompt files менять нельзя:
  тогда документационный PR обязан оставить явный follow-up issue на sync prompt contract blocks.
- для документационных stage seed'ов (`intake|vision|prd|arch|design|plan|doc-audit|qa|release|postdeploy|ops|rethink` и их `*:revise`):
  - обязательно использовать блок "Шаблоны артефактов по роли" из prompt envelope;
  - нельзя хардкодить пути к шаблонам артефактов в seed.
- non-doc seed'ы (`dev*`, `ai-repair`, `self-improve`) не должны получать лишние требования по документационным шаблонам.
- Для `self-improve-*` seed обязателен диагностический контур:
  - MCP `self_improve_runs_list` / `self_improve_run_lookup` / `self_improve_session_get`;
  - сохранение извлеченного `codex-cli` session JSON в `/tmp/codex-sessions/<run-id>`;
  - если `self_improve_session_get` возвращает `empty`/`not found`, run не прерывается:
    факт фиксируется в `tool_gaps` (с `run_id`), а анализ продолжается по issue/PR evidence.
