# Analytics Service

## Назначение
Сервис аналитики и прогнозирования. Предоставляет gRPC‑эндпоинты для получения отчётов, прогнозов временных рядов и факторного анализа. Данные хранятся во временном in‑memory хранилище (в будущем — ClickHouse).

## Структура
- `cmd/main.go` – инициализация хранилища, сервисов, gRPC‑сервер
- `internal/domain/event.go` – структура `Event`
- `internal/domain/report.go` – логика агрегации событий и формирования отчётов
- `internal/domain/forecast.go` – вызов `pkg/timeseries` (Хольт‑Уинтерс)
- `internal/domain/factor_analysis.go` – вызов `pkg/factoranalysis` (PCA)
- `internal/adapters/eventstore/memory.go` – in‑memory хранилище событий
- `internal/server/grpc.go` – gRPC‑сервер с методами `GetReport`, `Forecast`, `FactorAnalysis`
- `configs/dev.yaml` – порт, логирование, метрики

## Используемые пакеты из pkg
`config`, `logger`, `metrics`, `shutdown`, `timeseries`, `factoranalysis`, `statistics` (опосредованно), `regression` (может использоваться внутри factoranalysis)

## Примечания
- Для прогноза (`Forecast`) требуются дефолтные параметры Хольт‑Уинтерса (Alpha=0.5, Beta=0.3, Gamma=0.2, Period=4).
- `FactorAnalysis` пока возвращает только `ExplainedVarianceRatio`, не загружая матрицу нагрузок.
- Временное хранилище (`MemoryStore`) будет заменено на ClickHouse‑адаптер с подпиской на Kafka.