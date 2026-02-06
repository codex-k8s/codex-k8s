# Лучшие практики из `book.md`

Каталог практик для подготовки рекомендаций по документам `docs/design-guidelines/*` и `docs/templates/*`.

## Метод извлечения
- Базовый чанк: главы и крупные разделы из `docs/source/book_anchors.md`.
- Опорные выдержки: разделы по качеству требований, приоритетам, проверке, изменениям, трассируемости и рискам.
- Практика считается применимой, если её можно превратить в проверяемый пункт документа или шаблона.

## Практики
- `BP_SCOPE`: Фиксировать концепцию, границы и ограничения документа/проекта в явном виде. Источник: [docs/source/book.md#L5401-L6092](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L5401-L6092) (Границы и концепция проекта).
- `BP_SCOPE_GOV`: Управлять изменениями границ через формальные решения и оценку влияния. Источник: [docs/source/book.md#L6532-L6610](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L6532-L6610) (Контроль границ и оценка изменений).
- `BP_STAKEHOLDER`: Явно определять стейкхолдеров и формат взаимодействия заказчиков/разработчиков. Источник: [docs/source/book.md#L2416-L2830](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L2416-L2830) (Сотрудничество клиентов и разработчиков).
- `BP_DECISION`: Назначать ответственных за принятие решений и эскалацию конфликтов. Источник: [docs/source/book.md#L2924-L2968](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L2924-L2968) (Ответственные за принятие решений).
- `BP_USER_CLASSES`: Выделять классы пользователей вместо усреднённого «пользователя». Источник: [docs/source/book.md#L6799-L7202](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L6799-L7202) (Классы пользователей).
- `BP_ELICIT_METHODS`: Комбинировать методы выявления требований в зависимости от источника информации. Источник: [docs/source/book.md#L8074-L8571](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L8074-L8571) (Методы выявления требований).
- `BP_ELICIT_PLAN`: Планировать выявление требований как отдельный управляемый процесс. Источник: [docs/source/book.md#L8572-L8660](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L8572-L8660) (Планирование выявления требований).
- `BP_POST_ELICIT`: После встреч фиксировать протоколы, открытые вопросы и последующие действия. Источник: [docs/source/book.md#L8965-L9060](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L8965-L9060) (Действия после выявления требований).
- `BP_SRS`: Структурировать спецификацию требований под разные аудитории и цели. Источник: [docs/source/book.md#L12254-L12670](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L12254-L12670) (Спецификация требований к ПО).
- `BP_REQ_QUALITY`: Проверять требования на полноту, корректность, однозначность и проверяемость. Источник: [docs/source/book.md#L13373-L13574](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L13373-L13574) (Характеристики превосходных требований).
- `BP_WRITING`: Использовать ясный стиль формулировок и устранять двусмысленность. Источник: [docs/source/book.md#L13575-L14319](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L13575-L14319) (Принципы создания требований).
- `BP_MODELING`: Использовать модели/прототипы как инструмент уточнения и проверки требований. Источник: [docs/source/book.md#L14580-L16545](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L14580-L16545) (Моделирование требований).
- `BP_DATA`: Вести словарь данных и модель данных как часть требований. Источник: [docs/source/book.md#L16584-L17165](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L16584-L17165) (Моделирование данных и словарь).
- `BP_NFR`: Описывать качества системы как измеримые и обоснованные требования. Источник: [docs/source/book.md#L18065-L19261](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L18065-L19261) (Требования к атрибутам качества).
- `BP_PRIORITY`: Приоритизировать требования по ценности, стоимости и риску. Источник: [docs/source/book.md#L21321-L21798](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L21321-L21798) (Приоритизация по ценности, стоимости и риску).
- `BP_REVIEW`: Проводить системное рецензирование требований с чек-листами дефектов. Источник: [docs/source/book.md#L21983-L22604](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L21983-L22604) (Рецензирование требований).
- `BP_TESTABILITY`: Проверять тестопригодность требований до реализации. Источник: [docs/source/book.md#L22637-L22945](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L22637-L22945) (Тестирование требований).
- `BP_ACCEPTANCE`: Фиксировать критерии приемки и приемочные тесты. Источник: [docs/source/book.md#L22946-L23122](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L22946-L23122) (Критерии приемки).
- `BP_REUSE`: Повторно использовать требования и шаблоны, когда это экономит усилия. Источник: [docs/source/book.md#L23123-L24148](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L23123-L24148) (Повторное использование требований).
- `BP_PLAN_TRACE`: Связывать требования с планированием и оценками проекта. Источник: [docs/source/book.md#L24411-L24599](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L24411-L24599) (От требований к планам проекта).
- `BP_DESIGN_TRACE`: Обеспечивать переход от требований к архитектуре и коду без потери смысла. Источник: [docs/source/book.md#L24600-L24943](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L24600-L24943) (От требований к дизайну и коду).
- `BP_TEST_TRACE`: Строить тестирование на основе требований, а не кода. Источник: [docs/source/book.md#L24944-L25062](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L24944-L25062) (От требований к тестированию).
- `BP_ATTR`: Вести атрибуты требований (источник, приоритет, версия, владелец и т.д.). Источник: [docs/source/book.md#L29827-L29918](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L29827-L29918) (Атрибуты требований).
- `BP_STATUS`: Отслеживать состояние требований в течение всего жизненного цикла. Источник: [docs/source/book.md#L29919-L30109](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L29919-L30109) (Отслеживание состояния требований).
- `BP_CHANGE_POLICY`: Устанавливать и соблюдать политику управления изменениями. Источник: [docs/source/book.md#L30558-L30594](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L30558-L30594) (Политика управления изменениями).
- `BP_CHANGE_PROCESS`: Описывать операционный процесс обработки change request. Источник: [docs/source/book.md#L30646-L31033](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L30646-L31033) (Процесс управления изменениями).
- `BP_TRACE_MATRIX`: Использовать матрицу трассируемости требований. Источник: [docs/source/book.md#L32044-L32311](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L32044-L32311) (Матрица отслеживаемости требований).
- `BP_PROCESS_ASSETS`: Поддерживать процессные артефакты: чек-листы, шаблоны, примеры, политики. Источник: [docs/source/book.md#L34311-L34510](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L34311-L34510) (Документы процесса разработки/управления требованиями).
- `BP_RISK`: Управлять рисками требований на всех этапах (выявление, анализ, спецификация, утверждение, изменение). Источник: [docs/source/book.md#L35136-L35355](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L35136-L35355) (Риски, связанные с требованиями).
