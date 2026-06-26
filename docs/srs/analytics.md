# 🇬🇧 Analytics SRS / 🇷🇺 Техническое задание на Analytics

## 🇬🇧 Overview / 🇷🇺 Обзор

Analytics is the data processing and business intelligence service. It ingests auction events, stores them for fast analytical queries, and provides reporting, forecasting, and factor analysis capabilities. This document defines the functional and non‑functional requirements implemented in Analytics.
Analytics — сервис обработки данных и бизнес‑аналитики. Он принимает события аукционов, сохраняет их для быстрых аналитических запросов и предоставляет возможности отчётности, прогнозирования и факторного анализа. Этот документ определяет реализованные функциональные и нефункциональные требования к Analytics.

---

## 🇬🇧 Functional Requirements / 🇷🇺 Функциональные требования

### FR‑ANL‑001: Event Ingestion
**🇬🇧** Analytics MUST accept auction events either directly (in‑memory) or via Kafka. When Kafka is configured, a consumer group MUST continuously read messages and persist them.  
**🇷🇺** Analytics ДОЛЖЕН принимать события аукционов напрямую (in‑memory) или через Kafka. Если Kafka настроен, потребительская группа ДОЛЖНА непрерывно читать сообщения и сохранять их.

### FR‑ANL‑002: Event Storage
**🇬🇧** Analytics MUST store events in a column‑oriented database (ClickHouse) for efficient aggregation. If ClickHouse is unavailable, an in‑memory store MUST be used as fallback.  
**🇷🇺** Analytics ДОЛЖЕН хранить события в колоночной базе данных (ClickHouse) для эффективной агрегации. Если ClickHouse недоступен, ДОЛЖНО использоваться in‑memory хранилище.

### FR‑ANL‑003: Report Generation
**🇬🇧** Analytics MUST provide a `GetReport` gRPC method that streams aggregated rows. The client MUST be able to specify a date range, dimensions (e.g., campaign, device), and metrics (impressions, clicks, spend).  
**🇷🇺** Analytics ДОЛЖЕН предоставлять gRPC‑метод `GetReport`, который стримит агрегированные строки. Клиент ДОЛЖЕН иметь возможность указать диапазон дат, измерения (кампания, устройство) и метрики (показы, клики, расходы).

### FR‑ANL‑004: Forecasting
**🇬🇧** Analytics MUST provide a `Forecast` gRPC method that applies Holt‑Winters exponential smoothing to a time series. Sensible default parameters MUST be used when not specified (α=0.5, β=0.3, γ=0.2, period=4).  
**🇷🇺** Analytics ДОЛЖЕН предоставлять gRPC‑метод `Forecast`, который применяет экспоненциальное сглаживание Хольт‑Уинтерса к временному ряду. Разумные параметры по умолчанию ДОЛЖНЫ использоваться, если они не указаны (α=0.5, β=0.3, γ=0.2, period=4).

### FR‑ANL‑005: Factor Analysis
**🇬🇧** Analytics MUST provide a `FactorAnalysis` gRPC method that performs Principal Component Analysis (PCA) on a set of user profiles. It MUST return the explained variance ratio for each principal component.  
**🇷🇺** Analytics ДОЛЖЕН предоставлять gRPC‑метод `FactorAnalysis`, который выполняет анализ главных компонент (PCA) на наборе профилей пользователей. Он ДОЛЖЕН возвращать долю объяснённой дисперсии для каждой главной компоненты.

### FR‑ANL‑006: Test Data Generation
**🇬🇧** At startup, Analytics MUST generate representative test events for the last 7 days for campaigns 1001 and 1002, to enable immediate dashboard visualisation.  
**🇷🇺** При запуске Analytics ДОЛЖЕН генерировать репрезентативные тестовые события за последние 7 дней для кампаний 1001 и 1002 для немедленной визуализации в дашборде.

---

## 🇬🇧 Non‑Functional Requirements / 🇷🇺 Нефункциональные требования

- **NFR‑ANL‑001**: Report queries MUST be executed directly in ClickHouse via aggregating SQL (`GROUP BY`) to minimise data transfer.  
  Запросы отчётов ДОЛЖНЫ выполняться непосредственно в ClickHouse через агрегирующий SQL (`GROUP BY`) для минимизации передачи данных.
- **NFR‑ANL‑002**: Kafka consumer MUST support at‑least‑once delivery semantics and automatic offset commit.  
  Kafka‑потребитель ДОЛЖЕН поддерживать семантику доставки at‑least‑once и автоматическую фиксацию смещения.
- **NFR‑ANL‑003**: Forecasting and factor analysis MUST use pure Go implementations (`pkg/timeseries`, `pkg/factoranalysis`) without external dependencies.  
  Прогнозирование и факторный анализ ДОЛЖНЫ использовать реализации на чистом Go (`pkg/timeseries`, `pkg/factoranalysis`) без внешних зависимостей.
- **NFR‑ANL‑004**: Analytics MUST support graceful shutdown, draining the Kafka consumer within 10 seconds.  
  Analytics ДОЛЖЕН поддерживать корректное завершение работы, завершая Kafka‑потребитель в течение 10 секунд.

---

## 🇬🇧 Data Flow / 🇷🇺 Поток данных

```mermaid
sequenceDiagram
    participant Kafka
    participant Analytics
    participant ClickHouse
    participant Gateway

    Kafka->>Analytics: BidEvent message
    Analytics->>ClickHouse: INSERT INTO events ...
    Gateway->>Analytics: GetReport(start, end)
    Analytics->>ClickHouse: SELECT ... GROUP BY ...
    ClickHouse-->>Analytics: aggregated rows
    Analytics-->>Gateway: stream of ReportRow