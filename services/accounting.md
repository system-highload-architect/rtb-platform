# Accounting Service

## Назначение
Управление бюджетами рекламных кампаний. Предоставляет gRPC-методы для списания средств (`Debit`) и проверки баланса (`GetBalance`). Обеспечивает идемпотентность по `bid_id`.

## Структура
- `cmd/main.go` – инициализация in‑memory хранилища, тестовые балансы, gRPC-сервер
- `internal/domain/account.go` – модель `BalanceStore` (in‑memory реализация), методы `Get`, `Debit`, `Set`
- `internal/domain/service.go` – бизнес-логика с проверкой идемпотентности
- `internal/server/grpc.go` – gRPC‑сервер, реализующий `AccountingService`
- `configs/dev.yaml` – порт, настройки идемпотентности

## Используемые пакеты из pkg
`config`, `logger`, `metrics`, `shutdown`, `fixedpoint`, `idempotent`

## Текущее состояние
- Хранилище – in‑memory map, при перезапуске балансы сбрасываются.
- Тестовые кампании: "campaign-1" (10000 копеек), "1001" (100000), "1002" (50000).
- Идемпотентность реализована через `pkg/idempotent`.
- При списании проверяется достаточность средств.

## План развития
- Замена in‑memory на PostgreSQL с транзакциями.
- Добавление метрик и мониторинга.