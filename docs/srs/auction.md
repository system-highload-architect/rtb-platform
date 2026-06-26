# 🇬🇧 Auction SRS / 🇷🇺 Техническое задание на Auction

## 🇬🇧 Overview / 🇷🇺 Обзор

Auction is the core Real‑Time Bidding service. It receives bid requests, evaluates campaigns, determines winners, debits budgets, and publishes events. This document specifies the functional and non‑functional requirements implemented in Auction.
Auction — основной сервис Real‑Time Bidding. Он принимает запросы на ставку, оценивает кампании, определяет победителей, списывает бюджеты и публикует события. Этот документ определяет реализованные функциональные и нефункциональные требования к Auction.

---

## 🇬🇧 Functional Requirements / 🇷🇺 Функциональные требования

### FR‑AUC‑001: Bid Request Handling
**🇬🇧** Auction MUST accept gRPC `BidRequest` messages and return `BidResponse`. The request MUST contain `device_id`, `ip`, `user_agent`, optional location (`lat`, `lng`), and an idempotency key.  
**🇷🇺** Auction ДОЛЖЕН принимать gRPC‑сообщения `BidRequest` и возвращать `BidResponse`. Запрос ДОЛЖЕН содержать `device_id`, `ip`, `user_agent`, опциональные координаты (`lat`, `lng`) и ключ идемпотентности.

### FR‑AUC‑002: Idempotency
**🇬🇧** Auction MUST ensure that bids with the same `idempotency_key` are processed only once. Duplicate requests MUST return an ALREADY_EXISTS gRPC error.  
**🇷🇺** Auction ДОЛЖЕН гарантировать, что ставки с одинаковым `idempotency_key` обрабатываются только один раз. Повторные запросы ДОЛЖНЫ возвращать gRPC‑ошибку ALREADY_EXISTS.

### FR‑AUC‑003: Fraud Detection
**🇬🇧** Auction MUST check the request IP and device ID against an in‑memory blacklist. If either is present, the request MUST be rejected immediately with `"fraud"` error.  
**🇷🇺** Auction ДОЛЖЕН проверять IP и device ID запроса по in‑memory чёрному списку. Если любой из них присутствует, запрос ДОЛЖЕН быть немедленно отклонён с ошибкой `"fraud"`.

### FR‑AUC‑004: User Profile Enrichment
**🇬🇧** Auction MUST fetch the user profile from Aerospike using the `device_id`. If the profile is unavailable, the request MUST be rejected.  
**🇷🇺** Auction ДОЛЖЕН загружать профиль пользователя из Aerospike по `device_id`. Если профиль недоступен, запрос ДОЛЖЕН быть отклонён.

### FR‑AUC‑005: GeoIP Fallback
**🇬🇧** If the user profile does not contain valid coordinates, Auction MUST attempt to resolve the IP via MaxMind GeoLite2 and use the resulting coordinates.  
**🇷🇺** Если профиль пользователя не содержит корректных координат, Auction ДОЛЖЕН попытаться разрешить IP через MaxMind GeoLite2 и использовать полученные координаты.

### FR‑AUC‑006: Device Parsing
**🇬🇧** Auction MUST parse the `user_agent` string without allocations to determine the device type (desktop, mobile, tablet, bot). The result MUST be available for scoring.  
**🇷🇺** Auction ДОЛЖЕН разбирать строку `user_agent` без аллокаций для определения типа устройства (desktop, mobile, tablet, bot). Результат ДОЛЖЕН быть доступен для скоринга.

### FR‑AUC‑007: Campaign Retrieval
**🇬🇧** Auction MUST retrieve active campaigns from a cache backed by MongoDB. The cache MUST be refreshed periodically. If no campaigns are available, the auction MUST return `"no suitable campaign"`.  
**🇷🇺** Auction ДОЛЖЕН получать активные кампании из кэша, поддерживаемого MongoDB. Кэш ДОЛЖЕН периодически обновляться. Если кампании отсутствуют, аукцион ДОЛЖЕН вернуть `"no suitable campaign"`.

### FR‑AUC‑008: Scoring
**🇬🇧** Auction MUST compute a score and optimal bid for each campaign using:
- **LTV model** – predicts user lifetime value from profile features.
- **Impression value** – estimates the value of showing the ad.
- **Geo factor** – adjusts value based on proximity to the billboard.
The composite scorer from `pkg/valuation` MUST be used.  
**🇷🇺** Auction ДОЛЖЕН вычислять скор и оптимальную ставку для каждой кампании, используя:
- **LTV‑модель** – прогнозирует пожизненную ценность пользователя по признакам профиля.
- **Ценность показа** – оценивает ценность показа рекламы.
- **Гео‑фактор** – корректирует ценность в зависимости от близости к билборду.
ДОЛЖЕН использоваться составной скорер из `pkg/valuation`.

### FR‑AUC‑009: Winner Selection
**🇬🇧** Auction MUST sort campaigns by effective bid in descending order using LSD Radix Sort (O(N)). The campaign with the highest effective bid MUST be selected as the winner.  
**🇷🇺** Auction ДОЛЖЕН сортировать кампании по эффективной ставке по убыванию, используя LSD Radix Sort (O(N)). Кампания с максимальной эффективной ставкой ДОЛЖНА быть выбрана победителем.

### FR‑AUC‑010: Budget Debit
**🇬🇧** After a winner is selected, Auction MUST call Accounting’s `Debit` gRPC with the campaign ID, the original bid price, and the auction’s idempotency key. If the debit fails, the error MUST be logged.  
**🇷🇺** После выбора победителя Auction ДОЛЖЕН вызвать gRPC‑метод `Debit` Accounting с ID кампании, исходной ставкой и ключом идемпотентности аукциона. Если списание не удалось, ошибка ДОЛЖНА быть залогирована.

### FR‑AUC‑011: Event Publishing
**🇬🇧** Auction MUST publish a `BidEvent` to Kafka for every processed bid (subject to sampling). The event MUST contain the bid ID, campaign ID, device ID, price, and scoring components.  
**🇷🇺** Auction ДОЛЖЕН публиковать `BidEvent` в Kafka для каждой обработанной ставки (с учётом сэмплирования). Событие ДОЛЖНО содержать ID ставки, ID кампании, ID устройства, цену и компоненты скоринга.

### FR‑AUC‑012: A/B Experimentation
**🇬🇧** Auction MUST support A/B experiments using `pkg/experiment`. If a user falls into the `"new_scoring_v2"` experiment, an alternative scorer MUST be used.  
**🇷🇺** Auction ДОЛЖЕН поддерживать A/B‑эксперименты с использованием `pkg/experiment`. Если пользователь попадает в эксперимент `"new_scoring_v2"`, ДОЛЖЕН использоваться альтернативный скорер.

### FR‑AUC‑013: Asynchronous Processing
**🇬🇧** When an auction cache is configured (`timedcache` with TTL), Auction MUST enqueue bid requests and immediately return `"accepted"`. The cache finalizer MUST process all queued bids after TTL expires, running the full scoring, debit, and publishing flow.  
**🇷🇺** Если настроен кэш аукционов (`timedcache` с TTL), Auction ДОЛЖЕН помещать запросы в очередь и немедленно возвращать `"accepted"`. Финализатор кэша ДОЛЖЕН обрабатывать все накопленные заявки после истечения TTL, выполняя полный цикл скоринга, списания и публикации.

---

## 🇬🇧 Non‑Functional Requirements / 🇷🇺 Нефункциональные требования

- **NFR‑AUC‑001**: Auction MUST process synchronous bids in under 50 ms (p95) under normal load.  
  Auction ДОЛЖЕН обрабатывать синхронные ставки менее чем за 50 мс (p95) при нормальной нагрузке.
- **NFR‑AUC‑002**: Radix Sort MUST be allocation‑free (pre‑allocated buffers).  
  Radix Sort ДОЛЖЕН работать без выделений памяти (заранее выделенные буферы).
- **NFR‑AUC‑003**: All external calls (Aerospike, MongoDB, Accounting, Kafka) MUST be protected by circuit breakers.  
  Все внешние вызовы (Aerospike, MongoDB, Accounting, Kafka) ДОЛЖНЫ быть защищены circuit breaker’ами.
- **NFR‑AUC‑004**: Auction MUST support graceful shutdown, completing in‑flight asynchronous auctions within 10 seconds.  
  Auction ДОЛЖЕН поддерживать корректное завершение работы, завершая выполняющиеся асинхронные аукционы в течение 10 секунд.
- **NFR‑AUC‑005**: Sampling rate for Kafka events MUST be configurable.  
  Частота сэмплирования событий Kafka ДОЛЖНА быть настраиваемой.

---

## 🇬🇧 Data Flow / 🇷🇺 Поток данных

```mermaid
sequenceDiagram
    participant Gateway
    participant Auction
    participant Aerospike
    participant CampaignCache[MongoDB Cache]
    participant Accounting
    participant Kafka

    Gateway->>Auction: BidRequest
    Auction->>Auction: idempotency + fraud check
    Auction->>Aerospike: Get profile
    Auction->>CampaignCache: Get campaigns
    Auction->>Auction: Score & sort (Radix LSD)
    Auction->>Accounting: Debit
    Auction->>Kafka: Publish BidEvent
    Auction-->>Gateway: BidResponse