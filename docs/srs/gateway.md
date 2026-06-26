# 🇬🇧 Gateway SRS / 🇷🇺 Техническое задание на Gateway

## 🇬🇧 Overview / 🇷🇺 Обзор

Gateway is the single public-facing service. It translates external HTTP requests into internal gRPC calls and serves the web dashboard. This document defines the functional and non‑functional requirements implemented in the Gateway.
Gateway — единственный сервис, обращённый к внешнему миру. Он преобразует внешние HTTP‑запросы во внутренние gRPC‑вызовы и раздаёт веб‑дашборд. Этот документ определяет реализованные функциональные и нефункциональные требования к Gateway.

---

## 🇬🇧 Functional Requirements / 🇷🇺 Функциональные требования

### FR‑GW‑001: JSON‑RPC 2.0 Dispatcher
**🇬🇧** The Gateway MUST accept HTTP POST requests to `/rpc` with a valid JSON‑RPC 2.0 body. It MUST parse the `method` field without heap allocations and route it to the appropriate backend via gRPC.  
**🇷🇺** Gateway ДОЛЖЕН принимать HTTP POST‑запросы на `/rpc` с корректным телом JSON‑RPC 2.0. Он ДОЛЖЕН разбирать поле `method` без выделений памяти и маршрутизировать его к нужному бэкенду через gRPC.

**Supported methods / Поддерживаемые методы:**
- `auction.bid` → Auction
- `accounting.debit` → Accounting
- `accounting.getBalance` → Accounting
- `auth.register` → Auth
- `auth.login` → Auth

**🇬🇧 Implementation**: Uses `pkg/registry` for type‑safe dispatch and `pkg/zerocopy` for zero‑alloc JSON parsing.  
**🇷🇺 Реализация**: Использует `pkg/registry` для типобезопасной диспетчеризации и `pkg/zerocopy` для разбора JSON без аллокаций.

### FR‑GW‑002: REST API for Analytics
**🇬🇧** The Gateway MUST provide protected REST endpoints that proxy to the Analytics gRPC service. Endpoints: `GET /api/report`, `GET /api/forecast`, `GET /api/factor-analysis`.  
**🇷🇺** Gateway ДОЛЖЕН предоставлять защищённые REST‑эндпоинты, проксирующие gRPC‑сервис Analytics. Эндпоинты: `GET /api/report`, `GET /api/forecast`, `GET /api/factor-analysis`.

### FR‑GW‑003: Excel Export
**🇬🇧** The Gateway MUST expose `GET /export/report` that generates and returns an `.xlsx` file built from Analytics data.  
**🇷🇺** Gateway ДОЛЖЕН предоставлять `GET /export/report`, который формирует и возвращает `.xlsx` файл на основе данных Analytics.

### FR‑GW‑004: Static File Serving
**🇬🇧** When configured with a static directory, the Gateway MUST serve the React dashboard and support SPA fallback: any non‑API request that does not match an existing file MUST return `index.html`.  
**🇷🇺** При настроенной статической директории Gateway ДОЛЖЕН раздавать React‑дашборд и поддерживать SPA‑фолбек: любой не‑API запрос, не соответствующий существующему файлу, ДОЛЖЕН возвращать `index.html`.

### FR‑GW‑005: Rate Limiting
**🇬🇧** The Gateway MUST apply per‑IP rate limiting using a token bucket algorithm. The rate and burst MUST be configurable.  
**🇷🇺** Gateway ДОЛЖЕН применять ограничение частоты запросов с одного IP на основе алгоритма token bucket. Частота и burst ДОЛЖНЫ быть настраиваемыми.

### FR‑GW‑006: Idempotency
**🇬🇧** The Gateway MUST support the `Idempotency-Key` header for POST requests. Duplicate keys within the TTL MUST return HTTP 409 Conflict.  
**🇷🇺** Gateway ДОЛЖЕН поддерживать заголовок `Idempotency-Key` для POST‑запросов. Повторяющиеся ключи в течение TTL ДОЛЖНЫ возвращать HTTP 409 Conflict.

### FR‑GW‑007: CORS
**🇬🇧** The Gateway MUST include permissive CORS headers to allow browser‑based dashboard access.  
**🇷🇺** Gateway ДОЛЖЕН включать разрешающие CORS‑заголовки для доступа браузерного дашборда.

### FR‑GW‑008: Authentication
**🇬🇧** The Gateway MUST support JWT authentication for protected endpoints. It MUST validate tokens by calling the Auth gRPC service. Authentication MAY be disabled in development.  
**🇷🇺** Gateway ДОЛЖЕН поддерживать JWT‑аутентификацию для защищённых эндпоинтов. Он ДОЛЖЕН проверять токены, вызывая gRPC‑сервис Auth. Аутентификация МОЖЕТ быть отключена при разработке.

### FR‑GW‑009: AppSec
**🇬🇧** The Gateway MUST validate the `Host` header against a configured allowlist to prevent DNS rebinding attacks.  
**🇷🇺** Gateway ДОЛЖЕН проверять заголовок `Host` по настроенному списку разрешённых для предотвращения атак DNS rebinding.

---

## 🇬🇧 Non‑Functional Requirements / 🇷🇺 Нефункциональные требования

- **NFR‑GW‑001**: JSON‑RPC request body parsing MUST be zero‑allocation in the hot path.  
  Разбор тела JSON‑RPC запроса ДОЛЖЕН быть без аллокаций на горячем пути.
- **NFR‑GW‑002**: All backend calls MUST have configurable timeouts and MUST NOT block indefinitely.  
  Все вызовы к бэкендам ДОЛЖНЫ иметь настраиваемые таймауты и НЕ ДОЛЖНЫ блокироваться навсегда.
- **NFR‑GW‑003**: The service MUST support graceful shutdown with a maximum total timeout of 30 seconds.  
  Сервис ДОЛЖЕН поддерживать корректное завершение работы с общим таймаутом 30 секунд.
- **NFR‑GW‑004**: Metrics MUST be exported via Prometheus endpoint at `/metrics`.  
  Метрики ДОЛЖНЫ экспортироваться через Prometheus эндпоинт `/metrics`.

---

## 🇬🇧 Middleware Chain / 🇷🇺 Цепочка промежуточных слоёв

Requests pass through a chain of middleware in the following order:
Запросы проходят через цепочку промежуточных слоёв в следующем порядке:

1. **CORS** – adds permissive headers.
2. **AppSec** – validates `Host` header.
3. **Rate Limit** – per‑IP token bucket.
4. **Idempotency** – checks `Idempotency-Key`.
5. **Auth** – validates JWT (only for protected routes).

This ordering ensures that security and rate limiting are applied before any business logic.
Такой порядок гарантирует, что безопасность и ограничение частоты применяются до любой бизнес‑логики.