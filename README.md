---

# 📋 RTB-платформа — Сводка проектных решений

## 1. Общая архитектура
- **Монорепозиторий** с Go workspace (`go.work`)
- **Чистая архитектура (Hexagonal/Clean)**: разделение на `domain`, `ports`, `adapters`
- **Структура каталогов** зафиксирована (см. ниже)
- **Коммуникации**: gRPC для синхронных вызовов, Kafka/Redpanda для асинхронных событий
- **Сервисы**: Gateway, Auction, Accounting, Analytics (+ Dashboard для фронтенда)

## 2. Сервисы и их зоны ответственности
### Gateway
- Принимает внешние HTTP/JSON-RPC от SSP
- Проксирует в Auction через gRPC
- REST-эндпоинты для дашборда и Excel-выгрузки (библиотека `excelize`)
- Может включать rate limiter и быстрый fraud quick-check

### Auction Core
- Обрабатывает BidRequest
- Профиль пользователя из Aerospike (< 1.5 мс)
- Fraud-чек в памяти (IP/Device blacklist)
- Скоринг и аукцион (Radix Sort LSD O(N))
- Два режима аукционов:
  - **RTB (мгновенный)** — синхронная обработка
  - **Длительные аукционы** — сессии управляются `timedcache`
- При выигрыше отправляет события в Kafka `bid_events`
- Вызывает Accounting gRPC (дебетование/проверка бюджета)

### Accounting
- Управление бюджетами рекламодателей
- gRPC: `Debit`, `CheckBalance`
- Консистентность через PostgreSQL
- Идемпотентность по `bid_id`
- Генерирует `balance_updates` в Kafka

### Analytics
- Консьюмит `bid_events` → ClickHouse (батчами)
- Предоставляет API для отчётов
- Использует материализованные представления и UDF в ClickHouse
- Выполняет факторный анализ (PCA), прогнозирование

### Dashboard
- Vite + React + TypeScript + Tailwind CSS + shadcn/ui
- Графики через Recharts
- Выгрузка Excel через Gateway

## 3. Общие библиотеки (`pkg/`)
- `fixedpoint` — безвыделенная арифметика для денег
- `zerocopy` — zero-allocation парсеры JSON
- `backpressure` — шаблон pipeline с каналами
- `config` — загрузка YAML/ENV
- `statistics` — средние, корреляции, t-тест
- `regression` — линейная, логистическая регрессии, оценка качества
- `factoranalysis` — PCA, Varimax
- `geospatial` — расстояния (Haversine), маршруты, quadkey/geohash
- `valuation` — LTV, impression value, bid optimizer
- `timeseries` — Holt-Winters, ARIMA (опционально)
- `timedcache` — специализированный кэш для сессий с TTL

## 4. `pkg/timedcache` — дизайн
- **Назначение:** хранение метаданных длительных аукционов с фиксированным TTL
- **Упорядоченность:** по времени последнего доступа (LRU-подход), т.к. TTL одинаков для всех сессий в экземпляре
- **Операции:**
  - `Set` (O1)
  - `Get` (O1)
  - `Extend` (O1) — перемещает в голову, обновляет `expiresAt`
- **Параллельность:** `sync.Mutex` защищает список и мапу
- **Демон:**
  1. Смотрит на хвост (ближайший к истечению)
  2. Спит до `expiresAt` хвоста
  3. При пробуждении удаляет просроченные элементы из кэша
  4. Отправляет метаданные в буферизированный канал `finalizeCh`
- **Метаданные сессии:**
  - `auctionID`
  - `expiresAt`
  - `auctionPtr unsafe.Pointer` (ссылка на объект аукциона)
  - `stopCh chan struct{}` (опционально, для сигнала)
- **Финализация:**
  - Пул воркеров (4–5 горутин) читают из `finalizeCh`
  - Каждый воркер вызывает `auction.Finalize()` (захватывая внутренний мьютекс аукциона)
  - Продления больше невозможны, т.к. элемент уже удалён из кэша
- **Экземпляры:** по одному на каждый тип длинного аукциона (фиксированный TTL)
- **RTB аукционы:** не используют timedcache, обрабатываются синхронно

## 5. ClickHouse — соглашения
- Материализованные представления для агрегатов
- Параметризованные представления (ClickHouse ≥23.4) для передачи параметров из бэка
- Пользовательские функции (UDF) для сложной аналитики внутри БД
- Бэкенд не генерирует сложные SQL, а вызывает готовые объекты

## 6. Поток длинного аукциона
1. Auction получает запрос на создание аукциона с длинным TTL
2. Создаёт объект аукциона, регистрирует в `timedcache` нужного типа
3. При каждой ставке вызывает `Extend` (O1), который только двигает элемент в голове
4. Демон `timedcache` отслеживает хвост, по истечении отправляет метаданные в `finalizeCh`
5. Воркер финализации определяет победителя, списывает бюджет (Accounting gRPC), пишет лог

## 7. Текущая структура каталогов
```
rtb-platform/
├── pkg/
│   ├── go.mod
│   ├── config/
│   ├── fixedpoint/
│   ├── zerocopy/
│   ├── backpressure/
│   ├── statistics/                  # Статистические функции
│   │   ├── summary.go              # Среднее, медиана, перцентили
│   │   ├── correlation.go          # Корреляция Пирсона, Спирмена
│   │   ├── hypothesis.go           # t-тест, хи-квадрат
│   │   └── bootstrap.go            # Бутстрэп для доверительных интервалов
│   ├── regression/                 # Регрессионный анализ
│   │   ├── linear.go               # МНК, Ridge, Lasso (с использованием gonum)
│   │   ├── logistic.go             # Логистическая регрессия
│   │   └── evaluation.go           # R², RMSE, MAE, кросс-валидация
│   ├── factoranalysis/             # Факторный анализ / PCA
│   │   ├── pca.go                  # Метод главных компонент
│   │   └── varimax.go             # Вращение факторов
│   ├── geospatial/                 # Геопространственные вычисления
│   │   ├── distance.go            # Хаверсинус, Винсенти, сферическое расстояние
│   │   ├── route.go               # Приближённый маршрут по прямой с дорожным графом (если загружен)
│   │   ├── geofence.go            # Попадание в полигон (геозоны)
│   │   └── quadkey.go             # Работа с тайлами (Quadkey/Geohash) для быстрых запросов
│   ├── valuation/                   # Оценка ценности рекламного места / кампании
│   │   ├── lifetime_value.go      # Прогноз LTV пользователя (на основе регрессий)
│   │   ├── impression_value.go    # Оценка показа: вероятность конверсии × ценность
│   │   └── bid_optimizer.go       # Вычисление оптимальной ставки (например, по модели Катца)
│   └── timeseries/                  # Временные ряды (для прогноза спроса)
│       ├── holt_winters.go
│       └── arima.go               # (опционально, можно обёртку к внешней либе)
├── services/
│   ├── dashboard/                 # Новый веб-фронтенд
│   │   ├── package.json
│   │   ├── vite.config.ts
│   │   ├── tailwind.config.js
│   │   └── src/
│   │       ├── components/        # UI-компоненты (графики, таблицы)
│   │       ├── pages/             # Страницы: дашборд, аналитика, аккаунт
│   │       └── api/               # Клиент для общения с Gateway (fetch)
│   ├── gateway/                   # Добавить REST-ручки:
│   │   └── internal/
│   │       └── adapters/
│   │           └── http/
│   │               ├── analytics_handler.go  # Эндпоинты отчётов
│   │               └── excel_handler.go      # Генерация и отдача Excel
│   ├── gateway/                     # Web-сервис (HTTP/JSON-RPC)
│   │   ├── go.mod                   # модуль github.com/rtb-platform/services/gateway
│   │   ├── cmd/
│   │   │   └── main.go
│   │   └── internal/
│   │       ├── ports/
│   │       ├── adapters/
│   │       └── domain/
│   ├── auction/
│   │   ├── internal/
│   │   │   ├── domain/
│   │   │   │   ├── engine.go
│   │   │   │   ├── sorter.go
│   │   │   │   └── scoring/           # Скоринг ставок с учётом ценности
│   │   │   │       ├── scorer.go          # Интерфейс Scorer
│   │   │   │       ├── predictive.go      # Реализация: вызывает pkg/valuation, pkg/regression
│   │   │   │       └── factors.go         # Факторы: гео-близость, история пользователя
│   │   │   ├── ports/
│   │   │   └── adapters/
│   │   │       ├── aerospike/
│   │   │       ├── mongodb/
│   │   │       ├── fraud/
│   │   │       └── geodata/             # Адаптер к данным о местоположении (OSM, карты)
│   ├── accounting/                  # Аккаутинг (списание, бюджеты)
│   │   ├── go.mod
│   │   ├── cmd/
│   │   │   └── main.go
│   │   └── internal/...
│   └── analytics/                   # Аналитика (выгрузка, факторный анализ)
│   │   ├── cmd/
│   │   ├── internal/
│   │   │   ├── domain/
│   │   │   │   ├── reporter.go          # Подготовка отчётов
│   │   │   │   ├── factor_analyzer.go   # Логика факторного анализа для UI
│   │   │   │   └── forecasting.go       # Прогнозирование прибыли от рекламной акции
│   │   │   ├── ports/
│   │   │   └── adapters/
│   │   │       └── clickhouse/          # Запросы в ClickHouse для обучения/статистики
└── docker-compose.yml               # локальная инфраструктура для разработки
```

## 8. Ключевые принятые решения
- LRU-подобный кэш с TTL, упорядоченный по последнему доступу, а не min-heap
- Для разных TTL — отдельные экземпляры кэша
- Финализация через общий канал + пул воркеров (не по горутине на аукцион)
- Отказ от `runtime.GC()` в janitor'е (антипаттерн)
- Использование `fixedpoint` для денег, никаких float64
- В Go-коде zero-allocation на горячем пути, переиспользование буферов через `sync.Pool`
- Безопасные утверждения типов с проверкой `ok`
- Для Mermaid диаграмм — заключение выражений со скобками/спецсимволами в кавычки

### Позже добавим quadkey.go и route.go в аналитический сервис, который будет кормить valuation через ClickHouse.

Мы завершили фундаментальный слой `pkg/`. Вот полное описание каждого пакета и его публичных API, чтобы в любой момент восстановить контекст.

---

### 1. `pkg/config`
**Назначение:** загрузка конфигурации из YAML с переопределением через env-переменные.  
**Файлы:** `config.go`, `options.go`, `env.go`  
**Методы:**
```go
func Load(cfg interface{}, opts ...Option) error
func WithPath(path string) Option
func WithEnvPrefix(prefix string) Option
```
- `cfg` — ненулевой указатель на структуру с тегами `yaml` и (опционально) `env`.
- По умолчанию читает `config.yaml`, префикс env — `RTB_`.
- Приоритет: env > YAML.

---

### 2. `pkg/fixedpoint`
**Назначение:** точная денежная арифметика в копейках/центах (int64).  
**Файлы:** `money.go`  
**Тип:** `type Money int64`  
**Методы:**
```go
NewFromInt64(units int64) Money
NewFromFloat64(amount float64) (Money, error)
ParseMoney(s string) (Money, error)
MustParseMoney(s string) Money
(m Money) Add(other Money) (Money, error)
(m Money) Sub(other Money) (Money, error)
(m Money) Mul(factor int64) (Money, error)
(m Money) Div(divisor int64) (Money, error)
(m Money) MulFloat(coeff float64) (Money, error)
(m Money) Cmp(other Money) int
(m Money) IsZero() bool
(m Money) Abs() Money
(m Money) Sign() int
(m Money) Float64() float64
(m Money) String() string
(m Money) MarshalText() ([]byte, error)
(*Money) UnmarshalText(text []byte) error
```

---

### 3. `pkg/zerocopy`
**Назначение:** zero-allocation утилиты для горячего пути.  
**Файлы:** `zerocopy.go`  
**Методы:**
```go
func GetBytes() *[]byte
func PutBytes(buf *[]byte)
func StringToBytes(s string) []byte
func BytesToString(b []byte) string
func GetJSONField(data []byte, field string) ([]byte, bool)
func AppendJSONString(dst []byte, s string) []byte
func AppendJSONInt(dst []byte, n int64) []byte
```

---

### 4. `pkg/backpressure`
**Назначение:** многопоточный конвейер обработки с обратным давлением.
**Файлы:** `pipeline.go`  
**Типы:**
```go
type Stage[T any] func(ctx context.Context, item T) error
type Pipeline[T any] struct{...}
```
**Методы:**
```go
func NewPipeline[T any](ctx context.Context, input <-chan T, stages []Stage[T], opts ...Option[T]) *Pipeline[T]
func WithWorkers[T any](workers ...int) Option[T]
func WithBufferSize[T any](size int) Option[T]
func (p *Pipeline[T]) Start() <-chan T
func (p *Pipeline[T]) Wait()
```

---

### 5. `pkg/registry`
**Назначение:** типобезопасный реестр обработчиков (замена switch-case).  
**Файлы:** `registry.go`  
**Типы:**
```go
type Handler[Req, Resp any] func(ctx context.Context, req Req) (Resp, error)
type Registry[K comparable, Req, Resp any] struct{...}
```
**Методы:**
```go
func New[K comparable, Req, Resp any]() *Registry[K, Req, Resp]
func (r *Registry[K, Req, Resp]) Register(key K, h Handler[Req, Resp])
func (r *Registry[K, Req, Resp]) Dispatch(ctx context.Context, key K, req Req) (Resp, error)
func (r *Registry[K, Req, Resp]) Exists(key K) bool
```

---

### 6. `pkg/timedcache`
**Назначение:** потокобезопасный кэш с фиксированным TTL, упорядоченным списком и точным демоном очистки.  
**Файлы:** `cache.go`, `daemon.go`, `options.go`  
**Тип:** `type Cache[K comparable, V any] struct{...}`  
**Методы:**
```go
func New[K comparable, V any](ttl time.Duration, opts ...Option[K, V]) *Cache[K, V]
func WithFinalizer[K comparable, V any](fn func(key K, value V)) Option[K, V]
func WithFinalizerWorkers[K comparable, V any](n int) Option[K, V]
func WithFinalizerBuffer[K comparable, V any](size int) Option[K, V]
func WithNowFunc[K comparable, V any](fn func() time.Time) Option[K, V]
func (c *Cache[K, V]) Get(key K) (V, bool)
func (c *Cache[K, V]) Set(key K, value V)
func (c *Cache[K, V]) Extend(key K) bool
func (c *Cache[K, V]) Delete(key K)
func (c *Cache[K, V]) Stop()
```
- Демон спит точно до истечения хвоста, после чего каскадно удаляет просроченные элементы.
- При `Get` / `Extend` элемент обновляет TTL и перемещается в голову списка.

---

### 7. `pkg/statistics`
**Назначение:** базовые статистические функции (без аллокаций в горячем пути).  
**Файлы:** `statistics.go`  
**Методы:**
```go
func Mean(values []float64) float64
func Variance(values []float64) float64
func StdDev(values []float64) float64
func Covariance(x, y []float64) float64
func Correlation(x, y []float64) float64
func Percentile(values []float64, p float64) float64
func Median(values []float64) float64
```

---

### 8. `pkg/regression`
**Назначение:** линейная и логистическая регрессия.  
**Файлы:** `linear.go`, `logistic.go`  
**Типы:**
```go
type LinearModel struct{ Coefficients []float64 }
type LogisticModel struct{ Coefficients []float64 }
```
**Методы:**
```go
func TrainLinear(x [][]float64, y []float64) (*LinearModel, error)
func (m *LinearModel) Predict(features []float64) float64
func TrainLogistic(x [][]float64, y []float64, learningRate float64, epochs int) (*LogisticModel, error)
func (m *LogisticModel) PredictProb(features []float64) float64
```

---

### 9. `pkg/factoranalysis`
**Назначение:** метод главных компонент (PCA).  
**Файлы:** `pca.go`, `factoranalysis.go`  
**Типы:**
```go
type PCA struct{ Components [][]float64; ExplainedVariance []float64; Mean []float64 }
type FactorAnalysis interface{ ... } // Transform / InverseTransform
```
**Методы:**
```go
func TrainPCA(X [][]float64, nComponents int) (*PCA, error)
func (p *PCA) Transform(X [][]float64) [][]float64
func (p *PCA) InverseTransform(Z [][]float64) [][]float64
func (p *PCA) ExplainedVarianceRatio() []float64
```

---

### 10. `pkg/timeseries`
**Назначение:** прогнозирование временных рядов (Хольт-Уинтерс).  
**Файлы:** `holt_winters.go`  
**Тип:** `type HoltWintersParams struct{ Alpha, Beta, Gamma float64; Period int }`  
**Метод:**
```go
func HoltWintersForecast(data []float64, horizon int, params HoltWintersParams) ([]float64, error)
```

---

### 11. `pkg/geospatial`
**Назначение:** географические расчёты.  
**Файлы:** `distance.go`, `geofence.go`  
**Тип:** `type Point struct{ Lat, Lng float64 }`  
**Методы:**
```go
func HaversineDistance(a, b Point) float64
func PointInPolygon(pt Point, polygon []Point) bool
```

---

### 12. `pkg/valuation`
**Назначение:** комплексная оценка ценности показа и оптимальной ставки для RTB.  
**Файлы:** `ltv.go`, `impression.go`, `geo_factor.go`, `bid_optimizer.go`, `valuation.go`  
**Типы и методы:**
```go
func NewLTVModel(coefficients []float64) *LTVModel
func (m *LTVModel) Predict(features []float64) float64

func NewImpressionValue(pCTRCoeff, pCVRCoeff []float64, convValue float64) *ImpressionValue
func (iv *ImpressionValue) Value(userFeatures, adFeatures []float64) float64

func NewGeoFactor(decayRate float64, useRoadMultiplier bool) *GeoFactor
func (g *GeoFactor) Factor(userPos, targetPos geospatial.Point) float64

func NewWinRateModel(mid, scale float64, maxBid fixedpoint.Money) *WinRateModel
func (w *WinRateModel) OptimalBid(value fixedpoint.Money) (fixedpoint.Money, error)

func NewScorer(ltv *LTVModel, imp *ImpressionValue, geo *GeoFactor, winRate *WinRateModel) *Scorer
func (s *Scorer) Score(userFeatures, adFeatures []float64, userPos, targetPos geospatial.Point) (score float64, bid fixedpoint.Money, err error)
```

---

### 13. `pkg/ratelimit`
**Назначение:** защита от DDoS (lock‑free Token Bucket).  
**Файлы:** `tokenbucket.go`, `limiter.go`  
**Типы и методы:**
```go
func NewTokenBucket(rate, burst float64) *TokenBucket
func (tb *TokenBucket) Allow() bool

func NewLimiter(rate, burst float64) *Limiter
func (l *Limiter) Allow(key string) bool
func (l *Limiter) Stop()
```

---

### 14. `pkg/appsec`
**Назначение:** OWASP-инструменты (XSS, редиректы, валидация, HMAC).  
**Файлы:** `url.go`, `sanitize.go`, `hmac.go`  
**Методы:**
```go
func SafeRedirect(rawURL string, allowedHosts []string) (string, error)
func SanitizeHTML(input string) string
func ValidID(id string) bool
func ValidNumber(s string) bool
func SignURLParams(baseURL string, params map[string]string, secret []byte) (string, error)
func VerifyURLParams(fullURL string, secret []byte) error
```

---

### 15. `pkg/idempotent`
**Назначение:** предотвращение дублирования операций.  
**Файлы:** `idempotent.go`  
**Тип и методы:**
```go
func NewStore(ttl time.Duration) *Store
func (s *Store) Check(key string) bool  // true — ключ новый, false — дубль
```
Использует `timedcache` под капотом.

---
