# Справочник пакетов `pkg/`

## 1. `pkg/config`
**Назначение:** загрузка конфигурации из YAML-файла с переопределением через переменные окружения.

**Файлы:** `config.go`, `options.go`, `env.go`

**Ключевые функции:**
- `Load(cfg interface{}, opts ...Option) error` — загружает конфигурацию в структуру `cfg` (ненулевой указатель). Приоритет: env > YAML.
- `WithPath(path string) Option` — путь к YAML (по умолчанию `config.yaml`).
- `WithEnvPrefix(prefix string) Option` — префикс для переменных окружения (по умолчанию `RTB_`).
- Теги структуры: `yaml:"field"`, `env:"FIELD"`.

---

## 2. `pkg/fixedpoint`
**Назначение:** точная денежная арифметика на базе `int64` (копейки/центы), без `float64`.

**Файлы:** `money.go`

**Тип:** `Money int64`

**Основные методы:**
- Конструкторы: `NewFromInt64`, `NewFromFloat64`, `ParseMoney`, `MustParseMoney`.
- Арифметика: `Add`, `Sub`, `Mul`, `Div`, `MulFloat`.
- Сравнение: `Cmp`, `IsZero`, `Sign`, `Abs`.
- Вывод: `String`, `Float64`, `MarshalText`, `UnmarshalText`.

---

## 3. `pkg/zerocopy`
**Назначение:** инструменты для избежания выделений памяти в горячем пути.

**Файлы:** `zerocopy.go`

**Функции:**
- Пул байтовых буферов: `GetBytes() *[]byte`, `PutBytes(buf *[]byte)`.
- Конвертация без копирования: `StringToBytes(s string) []byte`, `BytesToString(b []byte) string`.
- Инлайн-парсинг JSON: `GetJSONField(data []byte, field string) ([]byte, bool)`.
- Сборка JSON без аллокаций: `AppendJSONString`, `AppendJSONInt`.

---

## 4. `pkg/backpressure`
**Назначение:** многопоточный конвейер обработки с обратным давлением на дженериках.

**Файлы:** `pipeline.go`

**Типы и функции:**
- `type Stage[T any] func(ctx context.Context, item T) error`
- `NewPipeline[T](ctx, input <-chan T, stages []Stage[T], opts ...Option[T]) *Pipeline[T]`
- Опции: `WithWorkers[T](workers ...int)`, `WithBufferSize[T](size int)`.
- `(p *Pipeline[T]) Start() <-chan T` — запускает конвейер, возвращает выходной канал.
- `(p *Pipeline[T]) Wait()` — ожидает завершения.

---

## 5. `pkg/registry`
**Назначение:** типобезопасный реестр обработчиков (замена switch-case).

**Файлы:** `registry.go`

**Типы и методы:**
- `type Handler[Req, Resp any] func(ctx context.Context, req Req) (Resp, error)`
- `type Registry[K comparable, Req, Resp any]`
- `New[K, Req, Resp]() *Registry[K, Req, Resp]`
- `Register(key K, h Handler[Req, Resp])`, `Dispatch(ctx, key, req) (Resp, error)`, `Exists(key K) bool`.

---

## 6. `pkg/timedcache`
**Назначение:** потокобезопасный кэш с фиксированным TTL, упорядоченным списком и точным демоном очистки.

**Файлы:** `cache.go`, `daemon.go`, `options.go`

**Тип:** `Cache[K comparable, V any]`

**Методы:**
- `New[K, V](ttl time.Duration, opts ...Option[K, V]) *Cache[K, V]`
- Опции: `WithFinalizer(fn func(key K, value V))`, `WithFinalizerWorkers(n int)`, `WithFinalizerBuffer(size int)`, `WithNowFunc(fn func() time.Time)`.
- `Get(key K) (V, bool)`, `Set(key K, value V)`, `Extend(key K) bool`, `Delete(key K)`, `Stop()`, `Values() []V`.
- Демон засыпает точно до истечения хвоста; при `Get`/`Extend` элемент перемещается в голову.

---

## 7. `pkg/statistics`
**Назначение:** базовые статистические функции.

**Файлы:** `statistics.go`

**Функции:**
- `Mean`, `Variance`, `StdDev`, `Covariance`, `Correlation`, `Percentile`, `Median`.

---

## 8. `pkg/regression`
**Назначение:** линейная и логистическая регрессия.

**Файлы:** `linear.go`, `logistic.go`

**Типы:**
- `LinearModel` с методом `Predict(features []float64) float64`.
- `LogisticModel` с методом `PredictProb(features []float64) float64`.
- Обучение: `TrainLinear`, `TrainLogistic`.

---

## 9. `pkg/factoranalysis`
**Назначение:** метод главных компонент (PCA) и интерфейс факторного анализа.

**Файлы:** `pca.go`, `factoranalysis.go`

**Типы и методы:**
- `PCA` с полями `Components`, `ExplainedVariance`, `Mean`.
- `TrainPCA(X [][]float64, nComponents int) (*PCA, error)`.
- `(p *PCA) Transform(X) [][]float64`, `InverseTransform(Z) [][]float64`, `ExplainedVarianceRatio() []float64`.
- Интерфейс `FactorAnalysis`.

---

## 10. `pkg/timeseries`
**Назначение:** прогнозирование временных рядов (Хольт‑Уинтерс).

**Файлы:** `holt_winters.go`

**Тип:** `HoltWintersParams` (поля `Alpha`, `Beta`, `Gamma`, `Period`).

**Функция:** `HoltWintersForecast(data []float64, horizon int, params HoltWintersParams) ([]float64, error)`.

---

## 11. `pkg/geospatial`
**Назначение:** географические вычисления.

**Файлы:** `distance.go`, `geofence.go`

**Тип:** `Point{Lat, Lng float64}`

**Функции:**
- `HaversineDistance(a, b Point) float64` — расстояние в метрах.
- `PointInPolygon(pt Point, polygon []Point) bool`.

---

## 12. `pkg/valuation`
**Назначение:** комплексная оценка ценности показа и оптимальной ставки для RTB.

**Файлы:** `ltv.go`, `impression.go`, `geo_factor.go`, `bid_optimizer.go`, `valuation.go`

**Типы и их методы:**
- `LTVModel` — `NewLTVModel(coeff)`, `Predict(features) float64`.
- `ImpressionValue` — `NewImpressionValue(...)`, `Value(userFeats, adFeats) float64`.
- `GeoFactor` — `NewGeoFactor(decayRate, useRoad)`, `Factor(userPos, targetPos) float64`.
- `WinRateModel` — `NewWinRateModel(...)`, `OptimalBid(value Money) (Money, error)`.
- `Scorer` — `NewScorer(...)`, `Score(...) (score float64, bid Money, err error)` — агрегирует все факторы.

---

## 13. `pkg/ratelimit`
**Назначение:** защита от DDoS (lock‑free Token Bucket).

**Файлы:** `tokenbucket.go`, `limiter.go`

**Типы:**
- `TokenBucket` — `NewTokenBucket(rate, burst float64)`, `Allow() bool`.
- `Limiter` — `NewLimiter(rate, burst float64)`, `Allow(key string) bool`, `Stop()`.

---

## 14. `pkg/appsec`
**Назначение:** инструменты безопасности (OWASP): защита от XSS, валидация, безопасные редиректы, подпись URL.

**Файлы:** `url.go`, `sanitize.go`, `hmac.go`

**Функции:**
- `SafeRedirect(rawURL string, allowedHosts []string) (string, error)`
- `SanitizeHTML(input string) string`
- `ValidID(id string) bool`
- `ValidNumber(s string) bool`
- `SignURLParams(baseURL string, params map[string]string, secret []byte) (string, error)`
- `VerifyURLParams(fullURL string, secret []byte) error`

---

## 15. `pkg/idempotent`
**Назначение:** предотвращение дублирования операций.

**Файлы:** `idempotent.go`

**Тип:** `Store` (на основе `timedcache`)

**Метод:** `NewStore(ttl time.Duration) *Store`, `(s *Store) Check(key string) bool` — возвращает `true` для нового ключа, `false` для дубля.

---

## 16. `pkg/radixsort`
**Назначение:** поразрядная сортировка (LSD) для `int64`, в том числе с перестановкой индексов.

**Файлы:** `lsd.go`, `with_indices.go`

**Функции:**
- `SortInt64(data []int64)` — сортировка in‑place.
- `SortInt64WithIndices(data []int64, indices []int)` — синхронная перестановка индексов.

---

## 17. `pkg/geoip`
**Назначение:** in‑memory GeoIP (MaxMind GeoLite2).

**Файлы:** `geoip.go`

**Типы и методы:**
- `GeoDB` — `New(path string) (*GeoDB, error)`, `Lookup(ipStr string) (Result, error)`, `Close() error`.
- `Result` содержит `Country`, `City`, `Lat`, `Lng`.

---

## 18. `pkg/device`
**Назначение:** zero‑allocation парсинг User‑Agent.

**Файлы:** `device.go`

**Типы:** `DeviceInfo` с полями `Type`, `OS`, `Browser`.  
Константы типов: `UnknownDevice`, `Desktop`, `Mobile`, `Tablet`, `Bot`.

**Функция:** `Parse(ua string) DeviceInfo`.

---

## 19. `pkg/breaker`
**Назначение:** Circuit Breaker для внешних вызовов.

**Файлы:** `breaker.go`

**Тип:** `Breaker`

**Методы:**
- `New(name string, threshold int, timeout time.Duration) *Breaker`
- `State() State` (Closed/Open/HalfOpen)
- `Execute(ctx context.Context, fn func() error) error`

---

## 20. `pkg/sampler`
**Назначение:** вероятностный сэмплинг событий.

**Файлы:** `sampler.go`

**Тип:** `Sampler`

**Методы:** `NewSampler(rate float64) *Sampler`, `Sample() bool`.

---

## 21. `pkg/experiment`
**Назначение:** A/B‑флаги для раскатки новых алгоритмов.

**Файлы:** `experiment.go`

**Тип:** `Experiments`

**Методы:** `New(flags map[string]float64) *Experiments`, `IsInExperiment(userID, name string) bool` (детерминированный хэш).

---

## 22. `pkg/metrics`
**Назначение:** метрики на базе OpenTelemetry (счётчики, гистограммы).

**Файлы:** `metrics.go`

**Типы:** `Counter`, `Histogram`.

**Функции:**
- `Init(ctx, serviceName string, useOTLP bool) error`
- `Shutdown(ctx context.Context)`
- `Handler() http.Handler`
- `NewCounter(name, help string, labels []string) *Counter`, `(c *Counter) Inc(labelVals ...string)`
- `NewHistogram(name, help string, buckets []float64, labels []string) *Histogram`, `(h *Histogram) Observe(val float64, labelVals ...string)`

---

## 23. `pkg/logger`
**Назначение:** структурированный логгер на базе `log/slog`.

**Файлы:** `logger.go`

**Функция:** `New(level, format string, attrs ...slog.Attr) *slog.Logger`.

---

## 24. `pkg/shutdown`
**Назначение:** безопасное (graceful) завершение сервиса с приоритетами и таймаутами.

**Файлы:** `shutdown.go`

**Тип:** `Manager`

**Методы:**
- `NewManager(totalTimeout time.Duration) *Manager`
- `SetLogger(l Logger)`
- `Add(name string, priority int, fn Closer, timeout time.Duration)`
- `Shutdown(ctx context.Context) error`
- `Wait()` — блокируется до SIGINT/SIGTERM, затем вызывает все Closer по приоритету.

---