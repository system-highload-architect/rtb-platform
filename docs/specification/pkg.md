# 🇬🇧 Shared Packages (pkg) / 🇷🇺 Общие пакеты (pkg)

## 🇬🇧 Overview / 🇷🇺 Обзор

The `pkg/` directory contains reusable, service‑agnostic libraries that form the foundation of the RTB platform. They cover monetary arithmetic, mathematics and statistics, networking, security, geospatial calculations, data structures, and observability. Every package is designed for zero or minimal allocations in the hot path and is independent of any particular service.
Директория `pkg/` содержит переиспользуемые библиотеки, не привязанные к конкретному сервису, которые составляют фундамент RTB‑платформы. Они покрывают денежную арифметику, математику и статистику, сетевые взаимодействия, безопасность, гео‑вычисления, структуры данных и наблюдаемость. Каждый пакет спроектирован для нулевых или минимальных аллокаций на горячем пути и не зависит от какого‑либо конкретного сервиса.

---

## 🇬🇧 Package Index / 🇷🇺 Индекс пакетов

### 💰 Monetary
| Package | Purpose |
|---------|---------|
| `fixedpoint` | Exact monetary arithmetic on `int64` (kopeks / cents). Constructors from `int64`, `float64`, string. Methods for add, sub, mul, div, comparison, rounding. |

### 📊 Mathematics & Statistics
| Package | Purpose |
|---------|---------|
| `statistics` | Descriptive statistics: mean, variance, stddev, covariance, correlation, percentiles, median. |
| `regression` | Linear and logistic regression with training and prediction. |
| `factoranalysis` | Principal Component Analysis (PCA) with variance explained and varimax rotation. |
| `timeseries` | Holt‑Winters forecasting and optional ARIMA wrapper. |
| `radixsort` | LSD Radix Sort for `int64`, in‑place and with index tracking. O(N). |

### 🌐 Networking & Resilience
| Package | Purpose |
|---------|---------|
| `backpressure` | Generic pipeline with configurable workers and buffer size, providing back‑pressure. |
| `breaker` | Circuit Breaker pattern for external calls (Aerospike, MongoDB, etc.). |
| `ratelimit` | Token bucket rate limiter (per‑key and global). |

### 🗄️ Data Structures & Caching
| Package | Purpose |
|---------|---------|
| `timedcache` | Thread‑safe TTL cache with ordered eviction, configurable finalizer and daemon. |
| `idempotent` | Idempotency key store backed by `timedcache`. |
| `registry` | Type‑safe handler registry (used for JSON‑RPC dispatching). |

### 🔒 Security
| Package | Purpose |
|---------|---------|
| `appsec` | URL safe redirect, HTML sanitisation, HMAC‑signed URLs. |
| `geoip` | In‑memory MaxMind GeoLite2 lookup with city and coordinates. |
| `device` | Zero‑allocation User‑Agent parser (desktop, mobile, tablet, bot). |

### 🗺️ Geospatial
| Package | Purpose |
|---------|---------|
| `geospatial` | Haversine and Vincenty distances, point‑in‑polygon, Quadkey / Geohash tiles. |

### 📈 Valuation (RTB‑specific maths)
| Package | Purpose |
|---------|---------|
| `valuation` | LTV prediction, impression value calculation, geo‑factor, win‑rate model, composite scorer. |

### 🧪 Experimentation & Sampling
| Package | Purpose |
|---------|---------|
| `experiment` | Deterministic A/B experiment flags based on user ID hash. |
| `sampler` | Probabilistic event sampler. |

### 🛠️ Observability & Foundation
| Package | Purpose |
|---------|---------|
| `config` | YAML + environment variable configuration with custom prefix and automatic mapping. |
| `logger` | Structured `slog` logger with level and format (text / json) helpers. |
| `metrics` | OpenTelemetry counters and histograms, OTLP or stdout exporter. |
| `shutdown` | Graceful shutdown manager with priorities and per‑closer timeouts. |
| `zerocopy` | Pooled byte buffers, unsafe string ↔ bytes conversion, zero‑alloc JSON helpers. |

---

## 🇬🇧 Package Details / 🇷🇺 Описание пакетов

### `fixedpoint`
Represents money as `int64` (kopeks). All operations are exact and panic‑free. Supports JSON / text marshaling.
Хранит деньги как `int64` (копейки). Все операции точны и не вызывают паник. Поддерживает JSON / текстовое представление.

### `statistics`
Pure functions that operate on `[]float64`. Useful for monitoring (median of bids, percentiles of latency).
Чистые функции, работающие с `[]float64`. Полезны для мониторинга (медиана ставок, перцентили задержек).

### `regression`
`LinearModel` and `LogisticModel` with `Predict`. Training is done offline or at startup, predictions are allocation‑free.
`LinearModel` и `LogisticModel` с методом `Predict`. Обучение выполняется офлайн или при старте, предсказания без аллокаций.

### `factoranalysis`
`PCA` with `Train`, `Transform`, `InverseTransform`, `ExplainedVarianceRatio`. Used in Analytics to discover hidden factors.
`PCA` с методами `Train`, `Transform`, `InverseTransform`, `ExplainedVarianceRatio`. Используется в Analytics для поиска скрытых факторов.

### `timeseries`
`HoltWintersForecast` applies Holt‑Winters exponential smoothing. The Analytics service uses it with default parameters.
`HoltWintersForecast` применяет экспоненциальное сглаживание Хольт‑Уинтерса. Сервис Analytics использует его с параметрами по умолчанию.

### `radixsort`
`SortInt64` sorts a slice of int64 in ascending order in O(N). `SortInt64WithIndices` also reorders an index array synchronously. Used in Auction to pick the winner.
`SortInt64` сортирует срез int64 по возрастанию за O(N). `SortInt64WithIndices` синхронно переставляет массив индексов. Используется в Auction для определения победителя.

### `backpressure`
`Pipeline[T]` connects stages with buffered channels and limits concurrency. Provides back‑pressure when the pipeline is full.
`Pipeline[T]` соединяет стадии буферизированными каналами и ограничивает конкурентность. Обеспечивает обратное давление при заполнении.

### `breaker`
`Breaker` with Closed / Open / Half‑Open states. Protects calls to Aerospike, MongoDB, and other external services.
`Breaker` с состояниями Closed / Open / Half‑Open. Защищает вызовы к Aerospike, MongoDB и другим внешним сервисам.

### `ratelimit`
`TokenBucket` is a classic token bucket. `Limiter` wraps it with per‑key rate limiting, used in Gateway middleware.
`TokenBucket` — классический алгоритм token bucket. `Limiter` оборачивает его ограничением по ключу, используется в middleware Gateway.

### `timedcache`
`Cache[K, V]` with TTL, ordered eviction list, daemon that sleeps until the next expiry. Configurable finalizer (called synchronously or via worker pool). Used for campaign cache and async auction cache.
`Cache[K, V]` с TTL, упорядоченным списком вытеснения, демоном, который засыпает до ближайшего истечения. Настраиваемый финализатор (вызывается синхронно или через пул воркеров). Используется для кэша кампаний и асинхронного кэша аукционов.

### `idempotent`
`Store` backed by `timedcache`. `Check(key)` returns `true` for a new key and `false` for a duplicate. Used in Gateway and Accounting.
`Store` на базе `timedcache`. `Check(key)` возвращает `true` для нового ключа и `false` для дубликата. Используется в Gateway и Accounting.

### `registry`
`Registry[K, Req, Resp]` stores handlers and dispatches by key. Gateway uses it to route JSON‑RPC methods to gRPC ports.
`Registry[K, Req, Resp]` хранит обработчики и диспетчеризует по ключу. Gateway использует его для маршрутизации JSON‑RPC методов к gRPC‑портам.

### `appsec`
`SafeRedirect` validates redirect URLs against an allowed host list. `SanitizeHTML` strips dangerous tags. `SignURLParams` / `VerifyURLParams` protect URL parameters with HMAC.
`SafeRedirect` проверяет URL редиректа по списку разрешённых хостов. `SanitizeHTML` удаляет опасные теги. `SignURLParams` / `VerifyURLParams` защищают параметры URL с помощью HMAC.

### `geoip`
`GeoDB` wraps the MaxMind GeoLite2 reader. `Lookup(ip)` returns country, city, lat, lng. Used by Auction to enrich user location.
`GeoDB` оборачивает читатель MaxMind GeoLite2. `Lookup(ip)` возвращает страну, город, широту, долготу. Используется Auction для обогащения местоположения пользователя.

### `device`
`Parse(userAgent)` returns `DeviceInfo` with type (desktop, mobile, tablet, bot), OS, and browser. Zero allocations.
`Parse(userAgent)` возвращает `DeviceInfo` с типом (desktop, mobile, tablet, bot), ОС и браузером. Без аллокаций.

### `geospatial`
`HaversineDistance` and `VincentyDistance` compute distances in metres. `PointInPolygon` checks inclusion. `Quadkey` / `Geohash` support tile indexing.
`HaversineDistance` и `VincentyDistance` вычисляют расстояние в метрах. `PointInPolygon` проверяет принадлежность точки полигону. `Quadkey` / `Geohash` поддерживают индексацию тайлов.

### `valuation`
`LTVModel`, `ImpressionValue`, `GeoFactor`, `WinRateModel`, and a composite `Scorer`. All models are used in Auction to compute the optimal bid.
`LTVModel`, `ImpressionValue`, `GeoFactor`, `WinRateModel` и составной `Scorer`. Все модели используются в Auction для вычисления оптимальной ставки.

### `experiment`
`Experiments` holds named flags with percentages. `IsInExperiment(userID, name)` uses a deterministic hash to decide membership. Used in Auction for A/B testing.
`Experiments` хранит именованные флаги с процентами. `IsInExperiment(userID, name)` использует детерминированный хэш для определения принадлежности. Используется в Auction для A/B‑тестирования.

### `sampler`
`Sampler` with a configured rate. `Sample()` returns `true` with the given probability. Used in Auction to downsample events sent to Kafka.
`Sampler` с настроенной вероятностью. `Sample()` возвращает `true` с заданной вероятностью. Используется в Auction для снижения частоты событий, отправляемых в Kafka.

### `config`
`Load` reads a YAML file and then overrides fields with environment variables (prefix `RTB_`). Supports `time.Duration`, `float64`, `int`, `string`, slices, and nested structs.
`Load` читает YAML‑файл, затем переопределяет поля переменными окружения (префикс `RTB_`). Поддерживает `time.Duration`, `float64`, `int`, `string`, срезы и вложенные структуры.

### `logger`
`New` creates a `*slog.Logger` with the chosen level (debug, info, warn, error) and format (text, json). Adds service name as a default attribute.
`New` создаёт `*slog.Logger` с выбранным уровнем (debug, info, warn, error) и форматом (text, json). Добавляет имя сервиса как атрибут по умолчанию.

### `metrics`
`Init` starts an OpenTelemetry exporter (OTLP or stdout). `NewCounter` / `NewHistogram` create Prometheus‑compatible instruments. `Handler()` returns an HTTP handler for scraping.
`Init` запускает экспортёр OpenTelemetry (OTLP или stdout). `NewCounter` / `NewHistogram` создают инструменты, совместимые с Prometheus. `Handler()` возвращает HTTP‑обработчик для сбора метрик.

### `shutdown`
`Manager` allows registering named closers with priorities and timeouts. `Wait()` blocks until SIGINT/SIGTERM and then calls all closers in order.
`Manager` позволяет регистрировать именованные функции закрытия с приоритетами и таймаутами. `Wait()` блокируется до SIGINT/SIGTERM и затем вызывает все функции закрытия по порядку.

### `zerocopy`
`GetBytes` / `PutBytes` manage a pool of byte slices. `StringToBytes` / `BytesToString` perform unsafe conversions (read‑only). `GetJSONField` extracts a field without parsing the whole document. `AppendJSONString` / `AppendJSONInt` build JSON directly into a byte slice.
`GetBytes` / `PutBytes` управляют пулом байтовых срезов. `StringToBytes` / `BytesToString` выполняют небезопасные преобразования (только для чтения). `GetJSONField` извлекает поле без разбора всего документа. `AppendJSONString` / `AppendJSONInt` формируют JSON прямо в байтовый срез.

---

## 🇬🇧 Usage Map / 🇷🇺 Карта использования

| Package | Gateway | Auction | Accounting | Analytics | Auth |
|---------|:-------:|:-------:|:----------:|:---------:|:----:|
| `config` | ✓ | ✓ | ✓ | ✓ | ✓ |
| `logger` | ✓ | ✓ | ✓ | ✓ | ✓ |
| `metrics` | ✓ | ✓ | ✓ | ✓ | ✓ |
| `shutdown` | ✓ | ✓ | ✓ | ✓ | ✓ |
| `fixedpoint` | | ✓ | ✓ | | |
| `statistics` | | ✓ | | | |
| `regression` | | | | (indirect) | |
| `factoranalysis` | | | | ✓ | |
| `timeseries` | | | | ✓ | |
| `radixsort` | | ✓ | | | |
| `backpressure` | (optional) | | | | |
| `breaker` | | ✓ | | | |
| `ratelimit` | ✓ | | | | |
| `timedcache` | | ✓ | | | |
| `idempotent` | ✓ | ✓ | ✓ | | |
| `registry` | ✓ | | | | |
| `appsec` | ✓ | | | | |
| `geoip` | | ✓ | | | |
| `device` | | ✓ | | | |
| `geospatial` | | ✓ | | | |
| `valuation` | | ✓ | | | |
| `experiment` | | ✓ | | | |
| `sampler` | | ✓ | | | |
| `zerocopy` | ✓ | | | | |