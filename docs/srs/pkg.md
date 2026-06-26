# 🇬🇧 Shared Packages SRS / 🇷🇺 Техническое задание на общие пакеты (pkg)

## 🇬🇧 Overview / 🇷🇺 Обзор

The `pkg/` directory provides reusable, service‑agnostic libraries that implement cross‑cutting concerns for the RTB platform. This document specifies the requirements each package fulfills and the guarantees it offers.
Директория `pkg/` предоставляет переиспользуемые, независимые от сервисов библиотеки, реализующие сквозную функциональность RTB‑платформы. Этот документ определяет требования, которые выполняет каждый пакет, и гарантии, которые он предоставляет.

---

## 🇬🇧 General Requirements for All Packages / 🇷🇺 Общие требования ко всем пакетам

- **FR‑PKG‑001**: Every package MUST be importable independently and MUST NOT import any service‑specific code.  
  Каждый пакет ДОЛЖЕН импортироваться независимо и НЕ ДОЛЖЕН импортировать код, специфичный для какого‑либо сервиса.
- **FR‑PKG‑002**: Hot‑path functions MUST minimise or eliminate heap allocations.  
  Функции горячего пути ДОЛЖНЫ минимизировать или исключать выделения памяти в куче.
- **FR‑PKG‑003**: All exported APIs MUST be safe for concurrent use or explicitly documented as not safe.  
  Все экспортируемые API ДОЛЖНЫ быть безопасны для конкурентного использования или явно документированы как небезопасные.

---

## 🇬🇧 Package‑Specific Requirements / 🇷🇺 Требования к конкретным пакетам

### `config`
- **FR‑PKG‑CONF‑001**: MUST load configuration from a YAML file and then override fields with environment variables using a configurable prefix.  
  ДОЛЖЕН загружать конфигурацию из YAML‑файла, а затем переопределять поля переменными окружения с настраиваемым префиксом.
- **FR‑PKG‑CONF‑002**: MUST support nested structs, slices, `time.Duration`, `float64`, `int`, and `string` types.  
  ДОЛЖЕН поддерживать вложенные структуры, срезы, `time.Duration`, `float64`, `int` и `string`.

### `logger`
- **FR‑PKG‑LOG‑001**: MUST create a `*slog.Logger` with configurable level (debug, info, warn, error) and format (text, json).  
  ДОЛЖЕН создавать `*slog.Logger` с настраиваемым уровнем (debug, info, warn, error) и форматом (text, json).

### `metrics`
- **FR‑PKG‑MET‑001**: MUST initialise OpenTelemetry with OTLP or stdout exporter.  
  ДОЛЖЕН инициализировать OpenTelemetry с экспортёром OTLP или stdout.
- **FR‑PKG‑MET‑002**: MUST provide Prometheus‑compatible counters and histograms.  
  ДОЛЖЕН предоставлять счётчики и гистограммы, совместимые с Prometheus.

### `shutdown`
- **FR‑PKG‑SD‑001**: MUST manage graceful shutdown with prioritised closers and per‑closer timeouts.  
  ДОЛЖЕН управлять корректным завершением работы с приоритетами и таймаутами для каждой функции закрытия.

### `fixedpoint`
- **FR‑PKG‑FP‑001**: MUST represent monetary values as `int64` in minimal units. All arithmetic MUST be exact.  
  ДОЛЖЕН представлять денежные суммы как `int64` в минимальных единицах. Вся арифметика ДОЛЖНА быть точной.
- **FR‑PKG‑FP‑002**: MUST support conversion from `float64` and `string` with proper rounding.  
  ДОЛЖЕН поддерживать преобразование из `float64` и `string` с правильным округлением.

### `statistics`
- **FR‑PKG‑STAT‑001**: MUST compute mean, variance, stddev, covariance, correlation, percentiles, and median on `[]float64`.  
  ДОЛЖЕН вычислять среднее, дисперсию, стдоткл, ковариацию, корреляцию, перцентили и медиану на `[]float64`.

### `regression`
- **FR‑PKG‑REG‑001**: MUST train and predict linear and logistic regression models. Predictions MUST be allocation‑free.  
  ДОЛЖЕН обучать и предсказывать линейные и логистические регрессионные модели. Предсказания ДОЛЖНЫ быть без аллокаций.

### `factoranalysis`
- **FR‑PKG‑FA‑001**: MUST perform PCA with configurable number of components and return explained variance ratios.  
  ДОЛЖЕН выполнять PCA с настраиваемым количеством компонент и возвращать доли объяснённой дисперсии.

### `timeseries`
- **FR‑PKG‑TS‑001**: MUST implement Holt‑Winters exponential smoothing and return a forecast array.  
  ДОЛЖЕН реализовывать экспоненциальное сглаживание Хольт‑Уинтерса и возвращать массив прогнозов.

### `radixsort`
- **FR‑PKG‑RS‑001**: MUST sort `int64` slices in O(N) using LSD Radix Sort. MUST support synchronous index tracking.  
  ДОЛЖЕН сортировать срезы `int64` за O(N) с помощью LSD Radix Sort. ДОЛЖЕН поддерживать синхронное отслеживание индексов.

### `backpressure`
- **FR‑PKG‑BP‑001**: MUST provide a generic `Pipeline` with configurable worker count and buffer size, blocking when full.  
  ДОЛЖЕН предоставлять обобщённый `Pipeline` с настраиваемым количеством воркеров и размером буфера, блокирующий при заполнении.

### `breaker`
- **FR‑PKG‑BR‑001**: MUST implement the Circuit Breaker pattern with Closed, Open, and Half‑Open states.  
  ДОЛЖЕН реализовывать паттерн Circuit Breaker с состояниями Closed, Open и Half‑Open.

### `ratelimit`
- **FR‑PKG‑RL‑001**: MUST implement a per‑key token bucket rate limiter.  
  ДОЛЖЕН реализовывать ограничитель частоты на основе token bucket с поддержкой ключей.

### `timedcache`
- **FR‑PKG‑TC‑001**: MUST provide a thread‑safe cache with TTL, ordered eviction, a precise daemon, and configurable finalizers.  
  ДОЛЖЕН предоставлять потокобезопасный кэш с TTL, упорядоченным вытеснением, точным демоном и настраиваемыми финализаторами.

### `idempotent`
- **FR‑PKG‑ID‑001**: MUST prevent duplicate operations by key within a configurable TTL.  
  ДОЛЖЕН предотвращать дублирование операций по ключу в течение настраиваемого TTL.

### `registry`
- **FR‑PKG‑REG‑001**: MUST dispatch typed request/response handlers by key.  
  ДОЛЖЕН диспетчеризовать типизированные обработчики запросов/ответов по ключу.

### `appsec`
- **FR‑PKG‑AS‑001**: MUST validate redirect URLs against an allowlist. MUST sanitise HTML strings.  
  ДОЛЖЕН проверять URL редиректов по списку разрешённых. ДОЛЖЕН очищать HTML‑строки.

### `geoip`
- **FR‑PKG‑GI‑001**: MUST resolve IP addresses to geographic locations (country, city, coordinates) using MaxMind GeoLite2.  
  ДОЛЖЕН определять географическое положение IP‑адресов (страна, город, координаты) с помощью MaxMind GeoLite2.

### `device`
- **FR‑PKG‑DEV‑001**: MUST parse User‑Agent strings without allocations and return device type, OS, and browser.  
  ДОЛЖЕН разбирать строки User‑Agent без выделений памяти и возвращать тип устройства, ОС и браузер.

### `geospatial`
- **FR‑PKG‑GS‑001**: MUST compute distances (Haversine, Vincenty) and test point‑in‑polygon.  
  ДОЛЖЕН вычислять расстояния (Haversine, Vincenty) и проверять принадлежность точки полигону.

### `valuation`
- **FR‑PKG‑VAL‑001**: MUST combine LTV, impression value, geo factor, and win rate into a composite scorer for RTB.  
  ДОЛЖЕН объединять LTV, ценность показа, гео‑фактор и модель выигрыша в составной скорер для RTB.

### `experiment`
- **FR‑PKG‑EXP‑001**: MUST provide deterministic A/B experiment assignment based on user ID hash.  
  ДОЛЖЕН предоставлять детерминированное распределение A/B‑экспериментов на основе хэша ID пользователя.

### `sampler`
- **FR‑PKG‑SMP‑001**: MUST return `true` with a configurable probability.  
  ДОЛЖЕН возвращать `true` с настраиваемой вероятностью.

### `zerocopy`
- **FR‑PKG‑ZC‑001**: MUST provide a pool of byte buffers and zero‑alloc JSON helpers.  
  ДОЛЖЕН предоставлять пул байтовых буферов и хелперы для JSON без аллокаций.