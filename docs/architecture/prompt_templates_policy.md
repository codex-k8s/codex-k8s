---
doc_id: ARC-PRM-CK8S-0001
type: prompt-policy
title: "kodex — Prompt Templates Policy"
status: active
owner_role: SA
created_at: 2026-02-11
updated_at: 2026-03-13
related_issues: [1, 19, 100, 247, 248, 249, 397]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-19-full-docset"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-19
---

# Prompt Templates Policy

## TL;DR
- В текущем MVP prompt templates берутся только из repo seeds, встроенных в `agent-runner`.
- Каноническая модель шаблонов role-specific: отдельный body для каждого `agent_key` и каждого `kind` (`work`/`revise`).
- DB overrides, prompt template lifecycle в БД и UI-редактор prompt templates в MVP отсутствуют.
- `services.yaml` задает docs context и role-aware refs к артефактам через:
  - `spec.projectDocs[]`
  - `spec.roleDocTemplates`
- Effective template source в runtime и audit фиксируется как `repo_seed`.
- Effective locale в текущем MVP берется из platform default `KODEX_AGENT_DEFAULT_LOCALE`, fallback `ru`; unsupported locale нормализуется к `en`.

## Классы шаблонов

| Kind | Назначение | Пример seed |
|---|---|---|
| `work` | Выполнение задачи | `services/jobs/agent-runner/internal/runner/promptseeds/dev-work.md` |
| `revise` | Устранение замечаний по существующему PR/артефакту | `services/jobs/agent-runner/internal/runner/promptseeds/dev-revise.md` |

Дополнительные prompt-блоки, которые рендерятся поверх task-body:
- role profile blocks:
  - `services/jobs/agent-runner/internal/runner/templates/prompt_blocks/role_profile_<locale>.tmpl`
- follow-up issue contract blocks:
  - `services/jobs/agent-runner/internal/runner/templates/prompt_blocks/issue_contract_<locale>.tmpl`
- PR/review/discussion contract blocks:
  - `services/jobs/agent-runner/internal/runner/templates/prompt_blocks/pr_contract_<kind>_<locale>.tmpl`
- title/body contract blocks выше обязаны оставаться синхронными с
  `docs/delivery/development_process_requirements.md`;
  при markdown-only trigger допустим только явный follow-up issue на закрытие drift в prompt files.

## Каноническая seed-матрица

- Для каждого `agent_key` должны существовать отдельные body-шаблоны:
  - `work`
  - `revise`
- Для системных ролей baseline локали: `ru`, `en`.
- Один общий body-шаблон для всех ролей не считается целевой моделью.

### Реализованный fallback order
1. `stage-role-kind_locale`
2. `stage-role-kind`
3. `role-role-kind_locale`
4. `role-role-kind`
5. `stage-kind_locale`
6. `stage-kind`
7. `dev-kind_locale`
8. `dev-kind`
9. `default-kind_locale`
10. `default-kind`
11. встроенные fallback templates runner-а

## Источник шаблонов

### Repo seeds
- Базовые шаблоны находятся в:
  `services/jobs/agent-runner/internal/runner/promptseeds/*.md`
- Они embed-ятся в `agent-runner` и не зависят от БД.
- В seed-файлах хранится только task-body, без runtime metadata и секретов.

### Что не поддерживается в MVP
- `project override` в БД
- `global override` в БД
- UI refresh/versioning/preview lifecycle
- selector `repo|db`

Если в документации или исторических epic-артефактах встречается модель `project override -> global override -> repo seed`, это post-MVP backlog, а не текущий runtime contract.

## Seed vs final prompt

Seed-файл не отправляется агенту напрямую. Final prompt собирается из:
1. system/runtime envelope;
2. runtime context;
3. MCP capabilities block;
4. issue/PR/run context;
5. task-body из repo seed;
6. output contract.

В output contract обязательно фиксируются:
- communication language текущего запуска;
- требования по tests/docs/PR flow;
- cadence progress-feedback через `run_status_report`.
- если заголовок артефакта явно содержит роль агента, optional-фрагмент `Sprint S<спринт> Day<день>`
  вставляется сразу после role token, но только когда sprint/day известны из stage context.

## Locale policy

### Effective locale
- Worker задает locale из `KODEX_AGENT_DEFAULT_LOCALE`.
- Если значение пустое, используется `ru`.
- В `agent-runner` locale нормализуется:
  - `ru*` -> `ru`
  - `en*` -> `en`
  - все остальное -> `en`

### User-facing communication
- PR title/body/comments
- issue replies
- feedback messages
- `run_status_report`

должны использовать communication language effective locale запуска.

## Контекстный рендер

Final prompt должен включать:
- run metadata (`run_id`, `issue`, `pr`, `branch`, `trigger`);
- runtime mode (`full-env` / `code-only`);
- available MCP tools и approval hints;
- короткие инструкции по built-in user interaction tools:
  - `user.notify` для кратких пользовательских уведомлений без wait-state;
  - `user.decision.request`, когда агенту нужно запросить у пользователя выбор или подтверждение;
- role profile block;
- issue/PR artifact contract blocks;
- role-aware docs refs из `services.yaml/spec.projectDocs[]`;
- role-aware artifact-template refs из `services.yaml/spec.roleDocTemplates`.

### Multi-repo docs federation
- Для multi-repo проекта docs refs должны указывать `repository` alias.
- Для monorepo `repository` может быть опущен.
- Для каждого docs source в prompt context должен фиксироваться resolved commit SHA.

## Policy для self-improve

- `run:self-improve` использует тот же repo-seed policy, что и обычные stage-run.
- Изменения prompt seeds вносятся только через обычный repo review-flow.
- Любая правка prompt seed должна быть traceable через:
  - `flow_events`
  - `agent_sessions`
  - PR/Issue evidence

## Безопасность и качество

- В seed-файлах запрещены секреты, токены и прямые credential-инструкции.
- Шаблон не может ослаблять RBAC/approval policy.
- Любая новая role/stage seed-модель должна сопровождаться тестом и doc update.

## Связанные документы
- `docs/product/agents_operating_model.md`
- `docs/product/requirements_machine_driven.md`
- `services/jobs/agent-runner/internal/runner/promptseeds/README.md`
- `services/jobs/agent-runner/internal/runner/promptseeds/*.md`
- `services/jobs/agent-runner/internal/runner/templates/prompt_envelope.tmpl`
- `services/jobs/agent-runner/internal/runner/templates/prompt_blocks/*.tmpl`
- `docs/architecture/data_model.md`
