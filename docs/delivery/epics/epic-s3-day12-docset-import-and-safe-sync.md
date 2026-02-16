---
doc_id: EPC-CK8S-S3-D12
type: epic
title: "Epic S3 Day 12: Docset import + safe sync (agent-knowledge-base -> projects)"
status: completed
owner_role: EM
created_at: 2026-02-16
updated_at: 2026-02-16
related_issues: [19]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-16-s3-day12"
---

# Epic S3 Day 12: Docset import + safe sync (agent-knowledge-base -> projects)

## TL;DR
- Цель: дать платформе способ импортировать и безопасно синхронизировать переносимый доксет документации из `agent-knowledge-base` в новые и существующие проекты.
- Ключевая ценность: быстрый bootstrap проектной документации + безопасные обновления доксета без риска затереть локальные правки проекта.
- MVP-результат: import/sync через PR + `docs/.docset-lock.json` + UI/API для выбора групп и локали.

## Priority
- `P0`.

## Контекст
- Источник доксета (dev): `../agent-knowledge-base`.
- Источник доксета (prod): `https://github.com/codex-k8s/agent-knowledge-base.git`.
- Индекс: `docset.manifest.json` (format `manifest_version=1`).
- Спецификация import/sync и пример lock-файла: `agent-knowledge-base/docs/codex-k8s/docset_import_and_sync_<locale>.md`.
- В доксете файлы лежат как `*_ru.md` и `*_en.md`, но `import_path` в манифесте уже без суффикса локали (например `docs/templates/adr.md`).
- В проекте целевые пути должны быть без `_ru/_en`, потому что ссылки внутри markdown уже без суффиксов локали.

## Scope
### In scope (MVP)
- Manifest parser `manifest_version=1` и модели данных:
  - `groups[]`: `id`, `title.ru/en`, `description.ru/en`, `default_selected`, `items[]`.
  - `items[]`: `import_path`, `source_paths.ru/en`, `sha256.ru/en`, `category`, `common`, `stack_tags`, `title.ru/en`, `description.ru/en`, `frontmatter` (subset).
  - Валидация: неизвестная `manifest_version` должна возвращать ошибку.
- UX/интеграция:
  - предоставить способ получить список доступных `groups` из manifest, чтобы UI/CLI мог дать пользователю выбор;
  - `default_selected=true` должны предлагаться как дефолтный набор;
  - группа `examples` по умолчанию не должна импортироваться.
- Импорт (import):
  - вход: целевой репозиторий проекта (`owner/name`), `locale` (`ru|en`), список `group IDs`, `docset ref` (commit SHA/tag/branch);
  - действие: получить доксет по `ref`, прочитать manifest, собрать `items` выбранных групп, скопировать `source_paths[locale] -> import_path`;
  - результат: PR в проектный репозиторий с изменениями + lock-файл `docs/.docset-lock.json`.
- Синхронизация (sync), safe-by-default:
  - вход: проектный репозиторий, новый `docset ref`;
  - действие: прочитать `docs/.docset-lock.json`, взять новый manifest, для каждого файла из lock определить `update|drift` по sha256 и обновить только safe-файлы;
  - результат: PR с обновлениями + summary (сколько обновлено, список drift).
- Lock-файл (минимум):
  - `docs/.docset-lock.json`;
  - `lock_version`, `docset.id`, `docset.ref`, `docset.locale`, `docset.selected_groups`, `files[{path, sha256, source_path}]`.
- Тесты (unit):
  - парсинг manifest и выборка файлов по группам;
  - построение import plan (`src -> dst`) для locale;
  - drift detection (`cur_sha != old_sha`);
  - формирование и обновление lock-файла.

### Out of scope
- Изменения в репозитории `agent-knowledge-base` (это внешний источник).
- Автоматический смысловой merge конфликтов (только safe update или ручное решение).
- “Force overwrite” по умолчанию (force допускается только отдельным флагом и с явным предупреждением; в MVP можно оставить как planned).

## Алгоритм import (MVP)
1. Получить доксет по `ref` (commit SHA/tag/branch).
2. Прочитать `docset.manifest.json`, провалиться с ошибкой при `manifest_version != 1`.
3. Получить список `items` как объединение `groups[*].items` по выбранным group IDs (dedupe, стабильный порядок).
4. Для каждого item:
   - `src = item.source_paths[locale]`
   - `dst = item.import_path`
   - скопировать `src -> dst` (в целевых путях суффиксы `_ru/_en` не используются).
5. Сформировать и записать `docs/.docset-lock.json`.
6. Создать PR в репозиторий проекта с summary: docset id/ref, locale, выбранные группы, список файлов.

## Алгоритм sync (MVP), safe-by-default
1. Прочитать `docs/.docset-lock.json` и получить: `old_ref`, `locale`, `selected_groups`, `files[]`.
2. Получить новый `docset.manifest.json` по новому `ref`.
3. Для каждого файла из lock:
   - `old_sha = lock.files[].sha256`
   - `cur_sha = sha256(файла в проекте сейчас)`
   - `new_sha = manifest_item.sha256[locale]` (по `import_path == lock.files[].path`)
4. Применить правила:
   - если `cur_sha == old_sha` и `new_sha != old_sha`, обновить файл на новую версию;
   - если `cur_sha != old_sha`, не перезаписывать и пометить как `drift/conflict`;
   - если файл отсутствует или не найден в новом manifest, пометить как `drift/conflict`.
5. Обновить lock:
   - `docset.ref = <new ref>`;
   - обновить sha256 и source_path только для обновлённых файлов.
6. Создать PR с summary:
   - `updated_count`;
   - список `drift` файлов (пути + причина).

## Декомпозиция (Stories/Tasks)
- Story-1: Manifest v1 models + parser + validation.
- Story-2: Selection engine: group IDs -> item list (dedupe, стабильный порядок, ошибки на неизвестные item IDs).
- Story-3: Import plan builder:
  - вход `(groups, locale)` и выход `{src_path, dst_path, expected_sha256}`;
  - создание директорий, нормализация путей, запрет path traversal.
- Story-4: Lock-file IO:
  - чтение/валидация `lock_version`;
  - запись lock с нормализованными путями и sha256.
- Story-5: Sync engine (safe-by-default):
  - вычисление `old_sha` (из lock), `cur_sha` (по файлу в repo), `new_sha` (из manifest);
  - правила обновления и drift-репорт.
- Story-6: GitHub integration (staff management path):
  - fetch доксета по `ref` (локальный путь в dev, git URL в prod);
  - создание ветки, коммит, PR, описание PR с summary и выбранными группами.
- Story-7: Staff API + staff UI:
  - список групп (id/title/description/default_selected/stack_tags);
  - форма import (locale, ref, selected groups, preview списка файлов);
  - форма sync (new ref, preview обновлений и drift).
- Story-8: Тестовый набор и фиксация поведения:
  - unit tests на core-алгоритмы;
  - минимальный e2e smoke на одном тест-репо (PR создаётся, lock добавляется, sync даёт drift при локальных правках).

## Критерии приемки
- При `manifest_version != 1` операции fail с понятной ошибкой.
- Import создаёт PR, который:
  - добавляет/обновляет файлы по правилам `source_paths[locale] -> import_path`;
  - добавляет `docs/.docset-lock.json` в заданном формате.
- Sync создаёт PR, который:
  - обновляет только файлы без локальных правок (по sha256);
  - формирует summary: `updated_count` и список `drift` файлов;
  - обновляет lock: `docset.ref` и sha256 обновлённых файлов.
- По умолчанию нет force-режима, который перезаписывает локальные правки без явного действия пользователя.

## Риски/зависимости
- Требуется чёткое разделение “management path” (staff UI/API) и runtime path, чтобы не смешивать токены/права.
- Важно не допустить path traversal через `import_path` и `source_paths`.
- Нужен предсказуемый порядок файлов в PR и lock (чтобы не было шумных diff при одинаковом выборе групп).

## Фактический результат (выполнено)
- Реализован parser `manifest_version=1` и модели manifest/lock:
  - manifest: группы, items, локализованные поля;
  - lock: `docs/.docset-lock.json` (`lock_version=1`).
- Реализован import plan builder:
  - выбор групп по умолчанию через `default_selected=true`;
  - группа `examples` по умолчанию исключена;
  - dedupe по `import_path`;
  - валидация путей (нормализация + запрет path traversal).
- Реализован safe sync engine:
  - обновление только файлов без локальных изменений (по sha256);
  - drift detection с причинами.
- Staff API/UI:
  - staff UI получает список `groups` по ref+locale;
  - import/sync выполняются через создание PR в репозитории проекта;
  - результат содержит summary и сырой JSON-отчёт.

## Data model impact
- Изменений схемы БД не требовалось.
- Состояние синхронизации фиксируется через `docs/.docset-lock.json` в целевом репозитории.

## Проверки
- `go test ./...` — passed.
