rtb-platform/                        # корень монорепо
├── go.work                          # Go workspace, объединяет модули
├── proto/                           # единый источник proto-контрактов
│   ├── auction/
│   │   └── v1/
│   │       └── auction.proto
│   ├── accounting/
│   │   └── v1/
│   │       └── accounting.proto
│   ├── analytics/
│   │   └── v1/
│   │       └── analytics.proto
│   └── common/
│       └── v1/
│           ├── money.proto
│           └── user.proto
├── pb/                              # сгенерированный код (go)
│   ├── go.mod                       # модуль github.com/rtb-platform/pb
│   └── ... (сгенерённые файлы)
├── pkg/                             # общие утилиты
│   ├── go.mod                       # модуль github.com/rtb-platform/pkg
│   ├── config/                      # базовая конфигурация (загрузка YAML/ENV)
│   ├── fixedpoint/                  # безвыделенная арифметика для денег
│   ├── zerocopy/                    # zero-allocation хелперы (JSON in-place)
│   └── backpressure/                # общий шаблон pipeline с backpressure
├── services/
│   ├── gateway/                     # Web-сервис (HTTP/JSON-RPC)
│   │   ├── go.mod                   # модуль github.com/rtb-platform/services/gateway
│   │   ├── cmd/
│   │   │   └── main.go
│   │   └── internal/
│   │       ├── ports/
│   │       ├── adapters/
│   │       └── domain/
│   ├── auction/                     # RTB-ядро аукциона
│   │   ├── go.mod
│   │   ├── cmd/
│   │   │   └── main.go
│   │   └── internal/
│   │       ├── domain/
│   │       │   ├── engine.go
│   │       │   └── sorter.go
│   │       ├── ports/
│   │       └── adapters/
│   │           ├── aerospike/
│   │           ├── mongodb/
│   │           └── fraud/
│   ├── accounting/                  # Аккаутинг (списание, бюджеты)
│   │   ├── go.mod
│   │   ├── cmd/
│   │   │   └── main.go
│   │   └── internal/...
│   └── analytics/                   # Аналитика (выгрузка, факторный анализ)
│       ├── go.mod
│       ├── cmd/
│       │   └── main.go
│       └── internal/...
└── docker-compose.yml               # локальная инфраструктура для разработки