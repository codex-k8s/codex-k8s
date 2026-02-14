# codex-k8s

`codex-k8s` — webhook-driven платформа управления AI-агентами в Kubernetes.

```text
codex-k8s/                                                монорепозиторий платформы; с этой карты удобно начинать навигацию
├── AGENTS.md                                             обязательные правила для всех агентов; ОБЯЗАТЕЛЬНО читать первым перед любыми правками
├── README.md                                             карта структуры репозитория; смотреть при онбординге и перед поиском нужного контекста
├── services.yaml                                         инвентарь deployable-сервисов и окружений; ОБЯЗАТЕЛЬНО смотреть при изменениях состава сервисов
├── bin/codex-bootstrap/                                  install/validate/reconcile CLI для операторского bootstrap/deploy (first-install/reconcile)
├── services/                                             сервисный код по архитектурным зонам
│   ├── dev/webhook-simulator/                            dev-only симулятор webhook; смотреть при локальной/стендовой отладке входящих событий
│   ├── external/api-gateway/                             внешний ingress и edge-слой; смотреть при изменениях OpenAPI/webhook/auth
│   ├── internal/control-plane/                           доменная логика и владелец БД-схемы; смотреть при изменениях use-case/миграций/policy
│   ├── jobs/agent-runner/                                runtime запуска агентных pod; смотреть при изменениях выполнения run и toolchain
│   ├── jobs/worker/                                      фоновые jobs/reconciliation; смотреть при изменениях очередей, ретраев и cleanup
│   └── staff/web-console/                                staff UI на Vue; смотреть при изменениях интерфейса и staff API-клиента
├── libs/                                                 общий переиспользуемый код; ОБЯЗАТЕЛЬНО смотреть перед выносом helper-кода
│   └── go/servicescfg/                                   typed services.yaml + template partials рендерер для bootstrap/deploy
├── proto/                                                gRPC-контракты internal сервисов; ОБЯЗАТЕЛЬНО смотреть при изменениях внутренних API
├── deploy/                                               Kubernetes-манифесты и deploy-скрипты; смотреть при изменениях staging/prod выкладки
├── bootstrap/                                            bootstrap и CI/CD подготовка окружений; смотреть при изменениях env/vars/secrets/процесса выкладки
├── tools/                                                codegen и сервисные утилиты; смотреть при изменениях генерации контрактов и инфраструктурных утилит
└── docs/                                                 source of truth по продукту, архитектуре и процессу
    ├── architecture/                                     архитектурные решения и контракты; ОБЯЗАТЕЛЬНО смотреть при изменении архитектуры/API/данных
    │   ├── c4_context.md                                 системный контекст и внешние зависимости платформы
    │   ├── c4_container.md                               контейнерная декомпозиция сервисов и потоков взаимодействия
    │   ├── api_contract.md                               обзор API-контрактов (OpenAPI/gRPC/AsyncAPI) и правил их эволюции
    │   ├── data_model.md                                 каноническая модель данных и инварианты сущностей
    │   ├── agent_runtime_rbac.md                         модель runtime-доступа агентов и RBAC-ограничений
    │   ├── mcp_approval_and_audit_flow.md                policy для MCP-апрувов, audit events и governance-переходов
    │   ├── prompt_templates_policy.md                    политика шаблонов промптов и runtime-рендера контекста
    │   └── adr/                                          архитектурные ADR; смотреть при изменениях фундаментальных решений
    │       ├── ADR-0001-kubernetes-only.md               базовый ADR: поддерживается только Kubernetes
    │       ├── ...                                       промежуточные ADR
    │       └── ADR-0004-repository-provider-interface.md ADR по интерфейсу провайдеров репозиториев
    ├── delivery/                                         планирование и трассируемость поставки; ОБЯЗАТЕЛЬНО смотреть при планировании и закрытии задач
    │   ├── development_process_requirements.md           обязательный процесс разработки и doc-governance
    │   ├── delivery_plan.md                              сквозной план поставки MVP и этапов
    │   ├── issue_map.md                                  связность issue ↔ документы ↔ этапы
    │   ├── requirements_traceability.md                  матрица трассируемости требований к реализации
    │   ├── roadmap.md                                    этапы развития продукта после MVP
    │   ├── regression_s1_gate.md                         регрессионные критерии и чекпойнты качества
    │   ├── epic_s1.md                                    каталог эпиков спринта S1
    │   ├── ...                                           серийные каталоги эпиков по спринтам
    │   ├── epic_s3.md                                    каталог эпиков спринта S3
    │   ├── sprint_s1_mvp_vertical_slice.md               план и рамки спринта S1
    │   ├── ...                                           серийные планы спринтов
    │   ├── sprint_s3_mvp_completion.md                   план и рамки спринта S3
    │   └── epics/                                        подневные эпики; смотреть при детализации работ конкретного дня
    │       ├── epic-s1-day0-bootstrap-baseline.md        первый эпик дневной декомпозиции
    │       ├── ...                                       остальные day-эпики
    │       └── epic-s3-day12-mvp-closeout-and-handover.md последний эпик дневной декомпозиции MVP
    ├── design-guidelines/                                инженерные стандарты; ОБЯЗАТЕЛЬНО смотреть перед кодом и перед PR
    │   ├── AGENTS.md                                     индекс обязательных гайдов по backend/frontend/common
    │   ├── common/                                       общие правила проектирования и PR self-check
    │   │   ├── AGENTS.md                                 обзор common-гайдов и обязательных ссылок
    │   │   ├── design_principles.md                      DDD/SOLID/DRY/KISS/Clean Architecture для всех языков
    │   │   ├── project_architecture.md                   целевая структура репозитория и зоны сервисов
    │   │   ├── libraries_reusable_code_requirements.md   правила выноса общего кода в libs
    │   │   ├── external_dependencies_catalog.md          каталог разрешённых внешних зависимостей
    │   │   └── check_list.md                             общий чек-лист перед PR
    │   ├── go/                                           backend-стандарты; ОБЯЗАТЕЛЬНО смотреть при Go-изменениях
    │   │   ├── AGENTS.md                                 индекс Go-правил
    │   │   ├── services_design_requirements.md           правила слоёв и размещения backend-кода
    │   │   ├── code_generation.md                        правила OpenAPI/gRPC codegen для backend
    │   │   └── check_list.md                             Go-чек-лист перед PR
    │   ├── vue/                                          frontend-стандарты; ОБЯЗАТЕЛЬНО смотреть при Vue/TS-изменениях
    │   │   ├── AGENTS.md                                 индекс frontend-правил
    │   │   ├── frontend_architecture.md                  архитектура frontend-приложений
    │   │   ├── frontend_code_rules.md                    правила размещения и стиля кода Vue/TS
    │   │   ├── frontend_data_and_state.md                правила работы с API-данными и состоянием
    │   │   └── check_list.md                             Vue-чек-лист перед PR
    │   └── visual/                                       визуальные и UX-стандарты frontend
    │       ├── AGENTS.md                                 индекс визуальных правил
    │       ├── visual_character.md                       стиль и визуальный язык интерфейса
    │       └── check_list.md                             visual-чек-лист перед PR
    ├── product/                                          продуктовые требования и операционная модель; ОБЯЗАТЕЛЬНО смотреть при изменениях процесса/лейблов/ролей
    │   ├── requirements_machine_driven.md                канонический baseline требований платформы
    │   ├── agents_operating_model.md                     модель ролей агентов, режимов исполнения и ответственности
    │   ├── labels_and_trigger_policy.md                  политика лейблов (`run:*`, `state:*`, `need:*`) и trigger-flow
    │   ├── stage_process_model.md                        stage-модель процесса (`intake -> ... -> ops`)
    │   ├── brief.md                                      продуктовый обзор и контекст инициативы
    │   ├── constraints.md                                зафиксированные продуктовые ограничения
    │   └── prompt-seeds/                                 seed-шаблоны агентных промптов
    │       ├── dev-work.md                               базовый шаблон выполнения задач
    │       └── dev-review.md                             базовый шаблон ревью и аудита
    ├── ops/                                              runbook и операционные проверки; смотреть при staging/debug работах
    │   └── staging_runbook.md                            обязательные проверки staging и типовые диагностические команды
    ├── research/                                         исследовательские материалы и исходные бизнес-идеи
    │   └── src_idea-machine_driven_company_requirements.md первоисточник бизнес-идеи (с учётом текущей архитектуры платформы)
    └── templates/                                        шаблоны проектной документации; смотреть при создании новых продуктовых/архитектурных артефактов
        ├── adr.md                                        шаблон архитектурного решения (ADR)
        ├── ...                                           остальные шаблоны документов
        └── user_story.md                                 шаблон пользовательской истории
```
