# Gateway Service

## Назначение
Единая точка входа для внешних клиентов (SSP, веб-дашборд). Принимает JSON-RPC запросы к аукциону и аккаутингу, а также REST-запросы для аналитики и экспорта Excel. Обеспечивает безопасность, rate limiting, идемпотентность, CORS.

## Структура
- `cmd/main.go` – инициализация, gRPC-клиенты, middleware, HTTP-сервер
- `internal/domain/jsonrpc.go` – диспетчеризация JSON-RPC методов через реестр (`pkg/registry`)
- `internal/ports/interfaces.go` – интерфейсы `AuctionPort`, `AccountingPort`, `AnalyticsPort`, `AuthPort`
- `internal/adapters/grpcclient/` – реализации портов через gRPC (проксирование)
- `internal/handler/` – REST-обработчики для аналитики и экспорта Excel
- `internal/middleware/` – auth (пока отключён), appsec, ratelimit, idempotent
- `internal/server/http.go` – настройка маршрутов и цепочки middleware
- `configs/dev.yaml` – конфигурация (порт, адреса gRPC, security)

## Ключевые API
- `/rpc` — JSON-RPC (POST)
- `/api/report`, `/api/forecast`, `/api/factor-analysis` — REST для дашборда (требуют JWT, но сейчас auth отключён)
- `/export/report` — выгрузка Excel

## Используемые пакеты из pkg
`config`, `logger`, `metrics`, `shutdown`, `zerocopy`, `registry`, `ratelimit`, `appsec`, `idempotent`

## Аутентификация
На данный момент отключена (authMiddleware = nil). В будущем будет проверять JWT через сервис auth.

## Особенности
- Zero‑copy парсинг JSON для JSON-RPC
- Поддержка идемпотентности через заголовок Idempotency-Key
- Rate limiting по IP
- CORS для браузерных запросов