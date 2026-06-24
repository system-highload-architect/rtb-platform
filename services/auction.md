# Auction Service

## Назначение
Ядро RTB-платформы. Выполняет аукцион в реальном времени: получает запрос через gRPC, обогащает данные пользователя, проверяет фрод, проводит скоринг и сортировку ставок (Radix Sort LSD), выбирает победителя. Взаимодействует с Accounting для списания бюджета.

## Структура
- `cmd/main.go` – инициализация адаптеров (Aerospike, MongoDB, fraud, geo, Kafka), gRPC-сервера, тестовые кампании
- `internal/domain/engine.go` – логика аукциона: перебор кампаний, вычисление эффективной ставки, сортировка, выбор победителя
- `internal/domain/sorter.go` – обёртка над `pkg/radixsort`
- `internal/domain/types.go` – структуры `Campaign`, `BidEvent`, `BidResponse`, интерфейс `GeoResolver`
- `internal/domain/scoring/` – интерфейс `Scorer`, реализация `predictiveScorer` на основе `pkg/valuation`
- `internal/ports/interfaces.go` – интерфейсы репозиториев и сервисов
- `internal/adapters/` – Aerospike, MongoDB (c `timedcache`), fraud, geodata, kafka, grpcclient (Accounting)
- `internal/server/grpc.go` – gRPC-сервер, оркестрация вызовов, проверка идемпотентности, fraud, GeoIP, эксперименты, списание бюджета
- `configs/dev.yaml` – порт, параметры БД, коэффициенты скоринга, списки блокировки

## Используемые пакеты из pkg
Все, кроме `zerocopy`, `backpressure`, `registry`, `ratelimit`, `appsec`, `factoranalysis`, `timeseries`. Активно используются `valuation`, `geospatial`, `radixsort`, `timedcache`, `breaker`, `device`, `geoip`, `sampler`, `experiment`, `statistics`.

## Примечания
- Тестовые кампании загружаются при старте (ID 1001, 1002) с TTL 24h.
- GeoIP опционален, при ошибке загрузки сервис продолжает работу.
- Кэш кампаний использует `timedcache` с TTL.
- Accounting‑клиент вызывает `Debit` после успешного аукциона (идемпотентный ключ `bid_id`).
- A/B‑тесты реализованы через `experiment` и альтернативный скорер.
- События аукциона сэмплируются перед отправкой (если настроен sampler).