# 🇬🇧 Infrastructure SRS / 🇷🇺 Техническое задание на инфраструктуру

## 🇬🇧 Overview / 🇷🇺 Обзор

This document describes the infrastructure choices and requirements for the RTB Platform. It covers containerisation, databases, message broker, networking, and configuration management.
Этот документ описывает инфраструктурные решения и требования для RTB‑платформы. Он охватывает контейнеризацию, базы данных, брокер сообщений, сетевые взаимодействия и управление конфигурацией.

## 🇬🇧 Functional Requirements / 🇷🇺 Функциональные требования

### FR‑INF‑001: Containerisation
**🇬🇧** All services and third‑party components MUST run in Docker containers orchestrated by Docker Compose.  
**🇷🇺** Все сервисы и сторонние компоненты ДОЛЖНЫ работать в Docker‑контейнерах, управляемых Docker Compose.

### FR‑INF‑002: Service Isolation
**🇬🇧** Each microservice MUST be built from its own Dockerfile and be independently deployable.  
**🇷🇺** Каждый микросервис ДОЛЖЕН собираться из собственного Dockerfile и быть независимо развёртываемым.

### FR‑INF‑003: Database Persistence
**🇬🇧** Stateful services (PostgreSQL, ClickHouse, MongoDB, Aerospike) MUST persist data in named Docker volumes to survive container restarts.  
**🇷🇺** Сервисы с состоянием (PostgreSQL, ClickHouse, MongoDB, Aerospike) ДОЛЖНЫ сохранять данные в именованных томах Docker, чтобы переживать перезапуски контейнеров.

### FR‑INF‑004: Health Checks
**🇬🇧** All infrastructure containers MUST implement health checks. Dependent services MUST wait for healthy dependencies before starting.  
**🇷🇺** Все инфраструктурные контейнеры ДОЛЖНЫ реализовывать проверки здоровья. Зависимые сервисы ДОЛЖНЫ ожидать готовности зависимостей перед стартом.

### FR‑INF‑005: Inter‑Service Communication
**🇬🇧** Synchronous communication between services MUST use gRPC with Protocol Buffers. Asynchronous communication MUST use Kafka.  
**🇷🇺** Синхронное взаимодействие между сервисами ДОЛЖНО использовать gRPC с Protocol Buffers. Асинхронное взаимодействие ДОЛЖНО использовать Kafka.

### FR‑INF‑006: Configuration Management
**🇬🇧** Each service MUST support configuration via YAML files and environment variables with the `RTB_` prefix. Environment variables MUST override YAML values.  
**🇷🇺** Каждый сервис ДОЛЖЕН поддерживать конфигурацию через YAML‑файлы и переменные окружения с префиксом `RTB_`. Переменные окружения ДОЛЖНЫ переопределять значения из YAML.

### FR‑INF‑007: Network Security
**🇬🇧** Internal services MUST NOT expose ports to the host unless required for development. Gateway is the only service that MUST be accessible externally.  
**🇷🇺** Внутренние сервисы НЕ ДОЛЖНЫ открывать порты на хост, если это не требуется для разработки. Gateway — единственный сервис, который ДОЛЖЕН быть доступен извне.

### FR‑INF‑008: Resource Limits
**🇬🇧** Infrastructure containers SHOULD have reasonable resource limits to prevent resource starvation.  
**🇷🇺** Инфраструктурные контейнеры ДОЛЖНЫ иметь разумные ограничения ресурсов для предотвращения голодания ресурсов.

## 🇬🇧 Infrastructure Components / 🇷🇺 Компоненты инфраструктуры

### PostgreSQL
- **🇬🇧 Purpose**: Stores campaign balances for Accounting.  
  **🇷🇺 Назначение**: Хранит балансы кампаний для Accounting.
- **🇬🇧 Version**: 16‑alpine  
  **🇷🇺 Версия**: 16‑alpine
- **🇬🇧 Requirements**: ACID transactions, row‑level locking (`SELECT ... FOR UPDATE`), stored functions.  
  **🇷🇺 Требования**: ACID‑транзакции, блокировка строк (`SELECT ... FOR UPDATE`), хранимые функции.

### ClickHouse
- **🇬🇧 Purpose**: Stores auction events for Analytics.  
  **🇷🇺 Назначение**: Хранит события аукционов для Analytics.
- **🇬🇧 Version**: latest  
  **🇷🇺 Версия**: latest
- **🇬🇧 Requirements**: Column‑oriented storage, MergeTree engine, efficient aggregation by `(campaign_id, timestamp)`.  
  **🇷🇺 Требования**: Колоночное хранение, движок MergeTree, эффективная агрегация по `(campaign_id, timestamp)`.

### MongoDB
- **🇬🇧 Purpose**: Stores advertising campaigns for Auction.  
  **🇷🇺 Назначение**: Хранит рекламные кампании для Auction.
- **🇬🇧 Version**: 7  
  **🇷🇺 Версия**: 7
- **🇬🇧 Requirements**: Document‑oriented, flexible schema, upsert support.  
  **🇷🇺 Требования**: Документо‑ориентированная, гибкая схема, поддержка upsert.

### Aerospike
- **🇬🇧 Purpose**: Stores user profiles for ultra‑low latency access (<1.5 ms).  
  **🇷🇺 Назначение**: Хранит профили пользователей для доступа со сверхнизкой задержкой (<1.5 мс).
- **🇬🇧 Version**: Enterprise 7.0 (community edition acceptable for development).  
  **🇷🇺 Версия**: Enterprise 7.0 (community‑версия допустима для разработки).
- **🇬🇧 Requirements**: In‑memory storage, sub‑millisecond latency, TTL support.  
  **🇷🇺 Требования**: In‑memory хранение, субмиллисекундная задержка, поддержка TTL.

### Kafka
- **🇬🇧 Purpose**: Asynchronous event bus between Auction and Analytics.  
  **🇷🇺 Назначение**: Асинхронная шина событий между Auction и Analytics.
- **🇬🇧 Version**: 7.5.0 (Confluent)  
  **🇷🇺 Версия**: 7.5.0 (Confluent)
- **🇬🇧 Requirements**: At‑least‑once delivery, consumer groups, partition tolerance.  
  **🇷🇺 Требования**: Доставка at‑least‑once, потребительские группы, устойчивость к разделению.

### Zookeeper (Kafka dependency)
- **🇬🇧 Purpose**: Cluster coordination for Kafka.  
  **🇷🇺 Назначение**: Координация кластера для Kafka.
- **🇬🇧 Version**: 7.5.0 (Confluent)  
  **🇷🇺 Версия**: 7.5.0 (Confluent)

## 🇬🇧 Networking / 🇷🇺 Сетевое взаимодействие

All containers communicate over the default bridge network created by Docker Compose. Service discovery uses container names as hostnames.
Все контейнеры взаимодействуют через сеть bridge по умолчанию, создаваемую Docker Compose. Service discovery использует имена контейнеров как имена хостов.

### Port Map

| Service | Internal Port | External Port (Dev) |
|---------|:------------:|:-------------------:|
| Gateway | 8080 | 8080 |
| Auction | 9001 | 9001 |
| Accounting | 9002 | 9002 |
| Analytics | 9003 | 9003 |
| Auth | 9004 | 9004 |
| PostgreSQL | 5432 | 5432 |
| ClickHouse | 9000 / 8123 | 9000 / 8123 |
| MongoDB | 27017 | 27017 |
| Aerospike | 3000‑3002 | 3000‑3002 |
| Kafka | 9092 / 29092 | 9092 |
| Zookeeper | 2181 | 2181 |

In production, only Gateway should be exposed.
В production только Gateway должен быть открыт наружу.

## 🇬🇧 Non‑Functional Requirements / 🇷🇺 Нефункциональные требования

- **NFR‑INF‑001**: Services must start in the correct order (databases → services → gateway).  
  Сервисы должны запускаться в правильном порядке (базы данных → сервисы → gateway).
- **NFR‑INF‑002**: Graceful shutdown must be implemented (30s timeout, priority‑based closing).  
  Должно быть реализовано корректное завершение работы (таймаут 30с, закрытие по приоритетам).
- **NFR‑INF‑003**: Logs must be structured (JSON) and include service name and timestamp.  
  Логи должны быть структурированными (JSON) и содержать имя сервиса и временную метку.
- **NFR‑INF‑004**: Metrics must be exportable via Prometheus‑compatible endpoint (`/metrics`).  
  Метрики должны экспортироваться через Prometheus‑совместимый эндпоинт (`/metrics`).