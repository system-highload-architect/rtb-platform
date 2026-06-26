# 🇬🇧 Documentation Navigation / 🇷🇺 Навигация по документации

This index helps you quickly find the information you need — from architecture overviews to detailed feature specifications.
Этот индекс поможет быстро найти нужную информацию — от обзора архитектуры до детальных требований к фичам.

---

## 🇬🇧 General / 🇷🇺 Общее

- [README.md](../README.md) – project overview, architecture diagram, tech stack
- [launch.md](../launch.md) – how to run the entire system and troubleshoot common issues

---

## 🇬🇧 Specifications / 🇷🇺 Спецификации

Detailed descriptions of services, protocols, and packages with diagrams.
Детальные описания сервисов, протоколов и пакетов с диаграммами.

| File | Description |
|------|-------------|
| [gateway.md](specification/gateway.md) | Gateway service: JSON‑RPC, REST, middleware, static serving, authentication |
| [auction.md](specification/auction.md) | Auction service: bidding engine, scoring, Radix Sort, fraud detection, Aerospike, async processing |
| [accounting.md](specification/accounting.md) | Accounting service: PostgreSQL balances, debit, idempotency |
| [analytics.md](specification/analytics.md) | Analytics service: ClickHouse, Kafka consumer, reports, forecasting, factor analysis |
| [auth.md](specification/auth.md) | Auth service: JWT, registration, login, token validation |
| [proto.md](specification/proto.md) | Protocol Buffer contracts: all `.proto` files and their usage |
| [pkg.md](specification/pkg.md) | Shared packages (`pkg/`): fixedpoint, statistics, regression, geospatial, etc. |

---

## 🇬🇧 Feature Specifications (SRS) / 🇷🇺 Технические задания (SRS)

Requirements and design rationale for each implemented feature.
Требования и обоснование дизайна для каждой реализованной фичи.

| File | Description |
|------|-------------|
| [infrastructure.md](srs/infrastructure.md) | Infrastructure choices: Docker, databases, message broker, network layout |
| [gateway.md](srs/gateway.md) | Gateway features: JSON‑RPC dispatcher, rate limiting, idempotency, CORS, static serving |
| [auction.md](srs/auction.md) | Auction features: bidding engine, scoring models, Radix Sort, fraud detection, A/B testing, async pipeline |
| [accounting.md](srs/accounting.md) | Accounting features: balance storage, debit with idempotency, PostgreSQL schema |
| [analytics.md](srs/analytics.md) | Analytics features: event ingestion from Kafka, ClickHouse storage, report generation, forecasting, PCA |
| [auth.md](srs/auth.md) | Auth features: registration, login, JWT issuance and validation |
| [pkg.md](srs/pkg.md) | Shared package requirements: API contracts, performance guarantees, usage examples |
| [dashboard.md](srs/dashboard.md) | Dashboard features: login, report table, forecast chart, factor analysis, Excel export, auction form |

---

## 🇬🇧 How to use this documentation / 🇷🇺 Как использовать документацию

- **To understand the system architecture**: start with [README.md](../README.md), then explore [specification/](specification/).
- **To learn why certain decisions were made**: read the corresponding file in [srs/](srs/).
- **To run or troubleshoot the system**: [launch.md](../launch.md) has you covered.