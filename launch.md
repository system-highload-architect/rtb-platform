# Запуск RTB Platform

Этот документ описывает все способы запуска системы: от быстрого развёртывания в Docker до локальной разработки отдельных сервисов.

## Требования

- **Docker** и **Docker Compose** (версия 2 или новее)
- **Node.js 22+** (только если нужно запускать дашборд локально вне контейнера)
- Свободные порты: `8080`, `9001–9004`, `5432`, `9000`, `27017`, `3000–3002`, `2181`, `9092`

## Быстрый старт (Production‑like окружение)

1. Клонируйте репозиторий и перейдите в его корень:

   ```bash
   git clone <repo-url>
   cd rtb-platform
   ```

2. Запустите все сервисы и базы данных одной командой:

   ```bash
   docker compose up --build -d
   ```

   Первая сборка может занять несколько минут. Последующие запуски (`docker compose up -d`) будут мгновенными.

3. Откройте в браузере [http://localhost:8080](http://localhost:8080) – это дашборд.

4. Войдите, используя учётную запись администратора по умолчанию:

   - Email: `admin@rtb-platform.local`
   - Пароль: `Admin123!`

5. **Готово!** Можно проводить аукционы, просматривать отчёты, выгружать Excel.

## Конфигурация

Все сервисы настраиваются через переменные окружения с префиксом `RTB_`. Они автоматически переопределяют значения из YAML‑файлов. Основные параметры, которые могут понадобиться в production:

| Переменная               | Описание                                      | Значение по умолчанию в Docker |
|--------------------------|-----------------------------------------------|-------------------------------|
| `RTB_SERVER_PORT`        | Порт сервиса                                  | `8080` (Gateway), `9001‑9004` |
| `RTB_GRPC_AUCTION`       | Адрес Auction                                 | `auction:9001`                |
| `RTB_GRPC_ACCOUNTING`    | Адрес Accounting                              | `accounting:9002`             |
| `RTB_GRPC_ANALYTICS`     | Адрес Analytics                               | `analytics:9003`              |
| `RTB_GRPC_AUTH`          | Адрес Auth                                    | `auth:9004`                   |
| `RTB_DATABASE_DSN`       | DSN для PostgreSQL (Accounting)               | `postgres://rtb:rtbpass@postgres:5432/rtb?sslmode=disable` |
| `RTB_CLICKHOUSE_DSN`     | Адрес ClickHouse (Analytics)                  | `clickhouse:9000`             |
| `RTB_MONGO_URI`          | URI MongoDB (Auction)                         | `mongodb://mongodb:27017`     |
| `RTB_KAFKA_BROKERS`      | Список брокеров Kafka                         | `kafka:29092`                 |
| `RTB_KAFKA_TOPIC`        | Топик для событий                             | `bid_events`                  |
| `RTB_JWT_SECRET`         | Секретный ключ для JWT (Auth)                 | `super-secret-key-change-in-production` |
| `RTB_AEROSPIKE_HOSTS`    | Хосты Aerospike                               | `aerospike:3000`              |

Полные списки переменных смотрите в файлах `docker-compose.yml` и `services/*/configs/dev.yaml`.

## Локальная разработка (без Docker для Go)

Если вы хотите запускать сервисы нативно (Go), а базы данных – в Docker:

1. Поднимите инфраструктурные контейнеры:

   ```bash
   docker compose up -d postgres clickhouse mongodb aerospike kafka
   ```

2. В отдельных терминалах запустите сервисы:

   ```bash
   # Auth
   cd services/auth && go run ./cmd
   # Accounting
   cd services/accounting && go run ./cmd
   # Analytics
   cd services/analytics && go run ./cmd
   # Auction
   cd services/auction && go run ./cmd
   # Gateway
   cd services/gateway && go run ./cmd
   ```

   Сервисы прочитают конфигурацию из `configs/dev.yaml`, где адреса баз указаны как `localhost`. Убедитесь, что соответствующие порты проброшены в `docker-compose.yml`.

3. Запустите дашборд (опционально):

   ```bash
   cd services/dashboard
   npm install
   npm run dev
   ```

   Дашборд будет доступен на [http://localhost:3000](http://localhost:3000) и сам проксирует запросы к Gateway на порт 8080.

## Тестирование

Для проверки работоспособности используйте интеграционный скрипт:

```bash
./run_tests.sh
```

Он выполняет основные сценарии: валидация JSON‑RPC, фрод‑детекция, аукцион, списание бюджета, идемпотентность, аналитические эндпоинты. Все тесты должны показывать `PASS`. Если какой‑то тест падает, смотрите раздел «Устранение неполадок» ниже.

## Учётные данные по умолчанию

- **Администратор**: `admin@rtb-platform.local` / `Admin123!`
- **Тестовые кампании**: создаются автоматически при старте Auction (ID 1001, 1002).
- **Балансы**: для `campaign-1`, `1001`, `1002` устанавливаются начальные суммы.

## Устранение неполадок

### 1. Контейнеры не стартуют или падают с ошибкой

- Проверьте логи: `docker compose logs <service>`
- Убедитесь, что порты не заняты другими процессами.
- Для сервисов Go проверьте, что Dockerfile использует правильную версию Go (1.23 или 1.25) и все зависимости скачиваются (`go mod download`).

### 2. Gateway отвечает 404 на главной странице

- Статика дашборда должна быть скопирована в образ Gateway (см. `Dockerfile`).
- Проверьте, что в `cmd/main.go` Gateway передана опция `server.WithStaticDir("web/dist")`.
- Локально можно запустить дашборд через `npm run dev` (см. выше).

### 3. Ошибка «connection refused» между сервисами

- Убедитесь, что все сервисы используют правильные имена хостов в Docker‑сети (например, `auth:9004`, а не `localhost:9004`). Это настраивается через переменные окружения или в YAML‑конфигах.
- Проверьте, что все контейнеры находятся в одной сети (Docker Compose создаёт её автоматически).

### 4. Не работает аутентификация (Internal error при входе)

- Проверьте, что Auth‑сервис запущен и Gateway видит его по адресу `auth:9004`.
- Убедитесь, что в Gateway зарегистрированы обработчики `auth.login` и `auth.register` в `jsonrpc.go`.
- Попробуйте получить токен напрямую через `curl` (пример в тестах).

### 5. Тесты падают с кодом 409 (duplicate request)

- Идемпотентность работает по ключу. Перезапустите Gateway, чтобы сбросить кэш ключей, или дождитесь истечения TTL (по умолчанию 5 минут).

### 6. Асинхронный аукцион не списывает бюджет или не создаёт события

- Проверьте, что в `cmd/main.go` Auction создаётся `auctionCache` с финализатором и передаётся в `AuctionServer`.
- Убедитесь, что Kafka и ClickHouse доступны, и продюсер/консьюмер настроены правильно (логи сервисов помогут).

## Остановка системы

```bash
docker compose down
```

Чтобы также удалить все данные (volumes), добавьте флаг `-v`:

```bash
docker compose down -v
```

## Дополнительная информация

- Архитектура и описания сервисов: [docs/specification/](docs/specification/)
- Технические задания: [docs/srs/](docs/srs/)
- Конфигурация и переменные окружения: `docker-compose.yml` и `services/*/configs/`

Готово. Держи двуязычный `launch.md` — подробный, с учётом всех нюансов запуска и troubleshooting.

# 🇬🇧 Launch Guide / 🇷🇺 Инструкция по запуску

This document describes all ways to run the system: from quick Docker deployment to local development of individual services.
Этот документ описывает все способы запуска системы: от быстрого развёртывания в Docker до локальной разработки отдельных сервисов.

---

## 🇬🇧 Requirements / 🇷🇺 Требования

- **Docker** and **Docker Compose** (v2 or newer)
- **Node.js 22+** (only for running the dashboard locally outside a container)
- Free ports: `8080`, `9001–9004`, `5432`, `9000`, `27017`, `3000–3002`, `2181`, `9092`

---

## 🇬🇧 Quick Start (Production‑like environment) / 🇷🇺 Быстрый старт (Production‑like окружение)

1. Clone the repository and navigate to its root:
   ```bash
   git clone <repo-url>
   cd rtb-platform
   ```

2. Start all services and databases with a single command:
   ```bash
   docker compose up --build -d
   ```
   The first build may take a few minutes. Subsequent runs (`docker compose up -d`) will be instant.

3. Open [http://localhost:8080](http://localhost:8080) in your browser – this is the dashboard.

4. Sign in with the default administrator account:
   - Email: `admin@rtb-platform.local`
   - Password: `Admin123!`

5. **Done!** You can run auctions, view reports, and export Excel files.

---

## 🇬🇧 Configuration / 🇷🇺 Конфигурация

All services are configured via environment variables with the `RTB_` prefix. They automatically override YAML values. Below are the main parameters you may need in production:

| Variable                  | Description                                      | Default value in Docker |
|---------------------------|--------------------------------------------------|-------------------------|
| `RTB_SERVER_PORT`         | Service port                                     | `8080` (Gateway), `9001‑9004` |
| `RTB_GRPC_AUCTION`        | Auction address                                  | `auction:9001`          |
| `RTB_GRPC_ACCOUNTING`     | Accounting address                               | `accounting:9002`       |
| `RTB_GRPC_ANALYTICS`      | Analytics address                                | `analytics:9003`        |
| `RTB_GRPC_AUTH`           | Auth address                                     | `auth:9004`             |
| `RTB_DATABASE_DSN`        | PostgreSQL DSN (Accounting)                      | `postgres://rtb:rtbpass@postgres:5432/rtb?sslmode=disable` |
| `RTB_CLICKHOUSE_DSN`      | ClickHouse address (Analytics)                   | `clickhouse:9000`       |
| `RTB_MONGO_URI`           | MongoDB URI (Auction)                            | `mongodb://mongodb:27017` |
| `RTB_KAFKA_BROKERS`       | Kafka broker list                                | `kafka:29092`           |
| `RTB_KAFKA_TOPIC`         | Topic for auction events                         | `bid_events`            |
| `RTB_JWT_SECRET`          | JWT secret key (Auth)                            | `super-secret-key-change-in-production` |
| `RTB_AEROSPIKE_HOSTS`     | Aerospike hosts                                  | `aerospike:3000`        |

Full variable lists are available in `docker-compose.yml` and `services/*/configs/dev.yaml`.

---

## 🇬🇧 Local development (without Docker for Go) / 🇷🇺 Локальная разработка (без Docker для Go)

If you want to run services natively (Go) while databases run in Docker:

1. Start the infrastructure containers:
   ```bash
   docker compose up -d postgres clickhouse mongodb aerospike kafka
   ```

2. In separate terminals, launch the services:
   ```bash
   # Auth
   cd services/auth && go run ./cmd
   # Accounting
   cd services/accounting && go run ./cmd
   # Analytics
   cd services/analytics && go run ./cmd
   # Auction
   cd services/auction && go run ./cmd
   # Gateway
   cd services/gateway && go run ./cmd
   ```

   Services read configuration from `configs/dev.yaml`, where database addresses are `localhost`. Ensure the corresponding ports are exposed in `docker-compose.yml`.

3. Start the dashboard (optional):
   ```bash
   cd services/dashboard
   npm install
   npm run dev
   ```
   The dashboard will be available at [http://localhost:3000](http://localhost:3000) and will proxy API requests to Gateway on port 8080.

---

## 🇬🇧 Testing / 🇷🇺 Тестирование

Use the integration test script to verify the system:
```bash
./run_tests.sh
```
It runs core scenarios: JSON‑RPC validation, fraud detection, auction, budget debit, idempotency, analytics endpoints. All tests should show `PASS`. If a test fails, refer to the “Troubleshooting” section below.

---

## 🇬🇧 Default credentials / 🇷🇺 Учётные данные по умолчанию

- **Administrator**: `admin@rtb-platform.local` / `Admin123!`
- **Test campaigns**: automatically created on Auction startup (ID 1001, 1002).
- **Initial balances**: set for `campaign-1`, `1001`, `1002`.

---

## 🇬🇧 Troubleshooting / 🇷🇺 Устранение неполадок

### 1. Containers do not start or crash with an error
- Check logs: `docker compose logs <service>`
- Ensure ports are not occupied by other processes.
- For Go services, verify that the Dockerfile uses the correct Go version (1.23 or 1.25) and all dependencies download (`go mod download`).

### 2. Gateway returns 404 on the home page
- Dashboard static files must be copied into the Gateway image (see `Dockerfile`).
- Verify that `server.WithStaticDir("web/dist")` is passed in Gateway's `cmd/main.go`.
- Locally, you can run the dashboard via `npm run dev` (see above).

### 3. “connection refused” between services
- Ensure all services use correct Docker network hostnames (e.g., `auth:9004` instead of `localhost:9004`). This is configured via environment variables or YAML configs.
- Verify that all containers are on the same network (Docker Compose creates one automatically).

### 4. Authentication fails (Internal error on login)
- Check that the Auth service is running and Gateway can reach it at `auth:9004`.
- Ensure that `auth.login` and `auth.register` handlers are registered in Gateway's `jsonrpc.go`.
- Try obtaining a token directly via `curl` (example in tests).

### 5. Tests fail with 409 (duplicate request)
- Idempotency is key‑based. Restart Gateway to clear the key cache, or wait for the TTL to expire (default 5 minutes).

### 6. Asynchronous auction does not debit budget or produce events
- Verify that `auctionCache` is created in Auction's `cmd/main.go` with a finalizer and passed to `AuctionServer`.
- Ensure Kafka and ClickHouse are accessible, and the producer/consumer are correctly configured (service logs will help).

---

## 🇬🇧 Stopping the system / 🇷🇺 Остановка системы

```bash
docker compose down
```

To also remove all data (volumes), add the `-v` flag:
```bash
docker compose down -v
```

---

## 🇬🇧 Additional information / 🇷🇺 Дополнительная информация

- Architecture and service descriptions: [docs/specification/](docs/specification/)
- Feature specifications: [docs/srs/](docs/srs/)
- Configuration and environment variables: `docker-compose.yml` and `services/*/configs/`