# 🇬🇧 Protocol Buffers Contracts / 🇷🇺 Proto‑контракты

## 🇬🇧 Overview / 🇷🇺 Обзор

All inter‑service communication uses Protocol Buffers (proto3). Contracts are defined in `proto/` and generated code lives in `pb/`. Every service has its own package (`auction.v1`, `accounting.v1`, `analytics.v1`, `auth.v1`) plus shared types in `common.v1`.
Всё межсервисное взаимодействие построено на Protocol Buffers (proto3). Контракты определены в `proto/`, сгенерированный код находится в `pb/`. Каждый сервис имеет свой пакет (`auction.v1`, `accounting.v1`, `analytics.v1`, `auth.v1`) плюс общие типы в `common.v1`.

## 🇬🇧 Shared Types / 🇷🇺 Общие типы (`common/v1/`)

### `money.proto`

```mermaid
classDiagram
    class Money {
        +int64 amount
        +int32 scale
    }
```

Represents a monetary value in minimal units (kopeks, cents). `amount` is the integer value, `scale` is the number of decimal places (usually 2). All financial operations use this type to avoid floating‑point errors.
Представляет денежную сумму в минимальных единицах (копейки, центы). `amount` — целое значение, `scale` — количество знаков после запятой (обычно 2). Все финансовые операции используют этот тип, чтобы избежать ошибок плавающей точки.

### `user.proto`

```mermaid
classDiagram
    class UserProfile {
        +string device_id
        +string ip
        +string user_agent
        +double lat
        +double lng
        +repeated double features
    }
```

A user profile used by Auction and Analytics. `features` is a vector of numerical attributes for LTV prediction and clustering.
Профиль пользователя, используемый Auction и Analytics. `features` — вектор числовых признаков для прогнозирования LTV и кластеризации.

---

## 🇬🇧 Auction (`auction/v1/auction.proto`)

```mermaid
classDiagram
    class BidRequest {
        +string device_id
        +string ip
        +string user_agent
        +double lat
        +double lng
        +repeated string campaign_ids
        +string idempotency_key
    }
    class BidResponse {
        +string campaign_id
        +Money bid_price
        +string creative_url
        +string error
    }
    BidRequest --> BidResponse : returns
```

- `BidRequest` contains device info, location, and an optional list of campaign IDs to restrict the auction.
- `BidResponse` returns the winner’s campaign ID, the bid price in Money, the creative URL, or an error message (`"accepted"` in async mode, `"fraud"`, `"no suitable campaign"`).

Service definition:

```protobuf
service AuctionService {
  rpc Bid (BidRequest) returns (BidResponse);
}
```

---

## 🇬🇧 Accounting (`accounting/v1/accounting.proto`)

```mermaid
classDiagram
    class DebitRequest {
        +string campaign_id
        +Money amount
        +string bid_id
    }
    class DebitResponse {
        +bool success
        +Money remaining_balance
        +string error
    }
    class GetBalanceRequest {
        +string campaign_id
    }
    class GetBalanceResponse {
        +Money balance
    }
    DebitRequest --> DebitResponse : returns
    GetBalanceRequest --> GetBalanceResponse : returns
```

- `DebitRequest.bid_id` is used for idempotency.
- `DebitResponse.error` may contain `"insufficient funds"`.
- `GetBalance` simply returns the current balance.

Service definition:

```protobuf
service AccountingService {
  rpc Debit (DebitRequest) returns (DebitResponse);
  rpc GetBalance (GetBalanceRequest) returns (GetBalanceResponse);
}
```

---

## 🇬🇧 Analytics (`analytics/v1/analytics.proto`)

```mermaid
classDiagram
    class ReportRequest {
        +string start_date
        +string end_date
        +repeated string dimensions
        +repeated string metrics
    }
    class ReportRow {
        +map~string, string~ dimension_values
        +map~string, double~ metric_values
    }
    class ForecastRequest {
        +repeated double history
        +int32 horizon
        +double alpha
        +double beta
        +double gamma
        +int32 period
    }
    class ForecastResponse {
        +repeated double forecast
    }
    class FactorRequest {
        +repeated UserProfile users
        +int32 n_components
    }
    class FactorResponse {
        +repeated double explained_variance_ratio
    }
```

- `GetReport` streams `ReportRow` messages, one per combination of dimensions.
- `Forecast` uses Holt‑Winters parameters; if not provided, sensible defaults are applied.
- `FactorAnalysis` accepts a set of `UserProfile` objects and returns the explained variance ratio for each principal component.

Service definition:

```protobuf
service AnalyticsService {
  rpc GetReport (ReportRequest) returns (stream ReportRow);
  rpc Forecast (ForecastRequest) returns (ForecastResponse);
  rpc FactorAnalysis (FactorRequest) returns (FactorResponse);
}
```

---

## 🇬🇧 Auth (`auth/v1/auth.proto`)

```mermaid
classDiagram
    class RegisterRequest {
        +string email
        +string password
        +string role
    }
    class RegisterResponse {
        +string user_id
        +string error
    }
    class LoginRequest {
        +string email
        +string password
    }
    class LoginResponse {
        +string access_token
        +string refresh_token
        +int64 expires_in
        +string error
    }
    class ValidateRequest {
        +string access_token
    }
    class ValidateResponse {
        +bool valid
        +string user_id
        +string role
        +google.protobuf.Timestamp expires_at
    }
    class RefreshRequest {
        +string refresh_token
    }
    class RefreshResponse {
        +string access_token
        +string refresh_token
        +int64 expires_in
        +string error
    }
```

- `Register` creates a new user.
- `Login` returns an access token (short‑lived) and a refresh token (long‑lived).
- `Validate` checks the access token’s validity and returns associated metadata.
- `Refresh` accepts a refresh token and returns a new pair (simplified implementation).

Service definition:

```protobuf
service AuthService {
  rpc Register (RegisterRequest) returns (RegisterResponse);
  rpc Login (LoginRequest) returns (LoginResponse);
  rpc Validate (ValidateRequest) returns (ValidateResponse);
  rpc Refresh (RefreshRequest) returns (RefreshResponse);
}
```

---

## 🇬🇧 How Contracts Are Used / 🇷🇺 Как используются контракты

```mermaid
graph TD
    Gateway -->|auction.Bid| Auction
    Gateway -->|accounting.Debit / GetBalance| Accounting
    Gateway -->|analytics.GetReport / Forecast / FactorAnalysis| Analytics
    Gateway -->|auth.Register / Login / Validate| Auth
    Auction -->|accounting.Debit| Accounting
```

- Gateway translates JSON‑RPC and REST calls into the corresponding gRPC methods.
- Auction calls `accounting.Debit` directly after a winner is chosen.
- All services use the generated Go code from `pb/` to ensure type safety.