package metrics

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

var (
	globalMeter   metric.Meter
	meterProvider *sdkmetric.MeterProvider
	initOnce      sync.Once
	initErr       error
)

// Init инициализирует глобальный MeterProvider с заданным именем сервиса.
// Если useOTLP true, экспорт идёт в OTLP HTTP (localhost:4318), иначе в Prometheus.
// Должна вызываться один раз при старте сервиса.
func Init(ctx context.Context, serviceName string, useOTLP bool) error {
	initOnce.Do(func() {
		res, err := resource.New(ctx,
			resource.WithAttributes(semconv.ServiceName(serviceName)),
		)
		if err != nil {
			initErr = fmt.Errorf("resource: %w", err)
			return
		}

		var reader sdkmetric.Reader
		if useOTLP {
			exp, err := otlpmetrichttp.New(ctx)
			if err != nil {
				initErr = fmt.Errorf("otlp exporter: %w", err)
				return
			}
			reader = sdkmetric.NewPeriodicReader(exp)
		} else {
			exp, err := prometheus.New()
			if err != nil {
				initErr = fmt.Errorf("prometheus exporter: %w", err)
				return
			}
			reader = exp
		}

		provider := sdkmetric.NewMeterProvider(
			sdkmetric.WithReader(reader),
			sdkmetric.WithResource(res),
		)
		otel.SetMeterProvider(provider)
		meterProvider = provider
		globalMeter = otel.Meter(serviceName)
	})
	return initErr
}

// Shutdown завершает работу провайдера метрик.
func Shutdown(ctx context.Context) {
	if meterProvider != nil {
		if err := meterProvider.Shutdown(ctx); err != nil {
			log.Printf("metrics shutdown: %v", err)
		}
	}
}

// Handler возвращает HTTP-обработчик для экспорта метрик Prometheus.
// Работает только если Init был вызван с useOTLP=false.
func Handler() http.Handler {
	return promhttp.Handler()
}

// Counter — счётчик с метками.
type Counter struct {
	name string
	keys []attribute.Key
	c    metric.Int64Counter
}

// NewCounter создаёт счётчик.
func NewCounter(name, help string, labels []string) *Counter {
	keys := make([]attribute.Key, len(labels))
	for i, l := range labels {
		keys[i] = attribute.Key(l)
	}
	c, err := globalMeter.Int64Counter(name, metric.WithDescription(help))
	if err != nil {
		log.Fatalf("metric counter %s: %v", name, err)
	}
	return &Counter{name: name, keys: keys, c: c}
}

// Inc увеличивает счётчик на 1. Значения меток должны идти в том же порядке, что и labels при создании.
func (c *Counter) Inc(labelVals ...string) {
	if len(labelVals) != len(c.keys) {
		log.Printf("counter %s: expected %d label values, got %d", c.name, len(c.keys), len(labelVals))
		return
	}
	attrs := make([]attribute.KeyValue, len(c.keys))
	for i, v := range labelVals {
		attrs[i] = c.keys[i].String(v)
	}
	c.c.Add(context.Background(), 1, metric.WithAttributes(attrs...))
}

// Histogram — гистограмма.
type Histogram struct {
	name string
	keys []attribute.Key
	h    metric.Float64Histogram
}

// NewHistogram создаёт гистограмму с явными границами buckets.
func NewHistogram(name, help string, buckets []float64, labels []string) *Histogram {
	keys := make([]attribute.Key, len(labels))
	for i, l := range labels {
		keys[i] = attribute.Key(l)
	}
	h, err := globalMeter.Float64Histogram(name,
		metric.WithDescription(help),
		metric.WithExplicitBucketBoundaries(buckets...),
	)
	if err != nil {
		log.Fatalf("metric histogram %s: %v", name, err)
	}
	return &Histogram{name: name, keys: keys, h: h}
}

// Observe записывает значение в гистограмму.
func (h *Histogram) Observe(val float64, labelVals ...string) {
	if len(labelVals) != len(h.keys) {
		log.Printf("histogram %s: expected %d label values, got %d", h.name, len(h.keys), len(labelVals))
		return
	}
	attrs := make([]attribute.KeyValue, len(h.keys))
	for i, v := range labelVals {
		attrs[i] = h.keys[i].String(v)
	}
	h.h.Record(context.Background(), val, metric.WithAttributes(attrs...))
}
