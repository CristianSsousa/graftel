// Package graftel fornece helpers para trabalhar com métricas OpenTelemetry.
package graftel

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
)

// MetricsHelper facilita a criação e uso de métricas.
// Use NewMetricsHelper para criar uma instância.
type MetricsHelper interface {
	// NewCounter cria um novo contador de métricas.
	NewCounter(name, description string, opts ...otelmetric.Int64CounterOption) (*Counter, error)

	// NewUpDownCounter cria um novo contador que pode incrementar ou decrementar.
	NewUpDownCounter(name, description string, opts ...otelmetric.Int64UpDownCounterOption) (*UpDownCounter, error)

	// NewHistogram cria um novo histograma de métricas.
	NewHistogram(name, description string, opts ...otelmetric.Float64HistogramOption) (*Histogram, error)

	// NewGauge cria um novo gauge observável.
	NewGauge(name, description string, callback func(context.Context, otelmetric.Float64Observer) error, opts ...otelmetric.Float64ObservableGaugeOption) (*Gauge, error)
}

// metricsHelper é a implementação concreta do MetricsHelper.
type metricsHelper struct {
	meter otelmetric.Meter
}

// NewMetricsHelper cria um novo helper de métricas.
func NewMetricsHelper(meter otelmetric.Meter) MetricsHelper {
	return &metricsHelper{
		meter: meter,
	}
}

// Counter representa um contador de métricas.
// Use NewCounter para criar uma instância.
type Counter struct {
	counter otelmetric.Int64Counter
}

// NewCounter cria um novo contador.
func (m *metricsHelper) NewCounter(name, description string, opts ...otelmetric.Int64CounterOption) (*Counter, error) {
	counter, err := m.meter.Int64Counter(name, append([]otelmetric.Int64CounterOption{
		otelmetric.WithDescription(description),
	}, opts...)...)
	if err != nil {
		return nil, err
	}

	return &Counter{counter: counter}, nil
}

// Add incrementa o contador pelo valor especificado.
func (c *Counter) Add(ctx context.Context, value int64, attrs ...attribute.KeyValue) {
	c.counter.Add(ctx, value, otelmetric.WithAttributes(attrs...))
}

// Increment incrementa o contador em 1.
func (c *Counter) Increment(ctx context.Context, attrs ...attribute.KeyValue) {
	c.Add(ctx, 1, attrs...)
}

// Gauge representa um gauge de métricas observável.
type Gauge struct {
	gauge otelmetric.Float64ObservableGauge
}

// NewGauge cria um novo gauge observável.
func (m *metricsHelper) NewGauge(name, description string, callback func(context.Context, otelmetric.Float64Observer) error, opts ...otelmetric.Float64ObservableGaugeOption) (*Gauge, error) {
	gauge, err := m.meter.Float64ObservableGauge(name, append([]otelmetric.Float64ObservableGaugeOption{
		otelmetric.WithDescription(description),
		otelmetric.WithFloat64Callback(callback),
	}, opts...)...)
	if err != nil {
		return nil, err
	}

	return &Gauge{gauge: gauge}, nil
}

// UpDownCounter representa um contador que pode incrementar ou decrementar.
type UpDownCounter struct {
	counter otelmetric.Int64UpDownCounter
}

// NewUpDownCounter cria um novo up-down counter.
func (m *metricsHelper) NewUpDownCounter(name, description string, opts ...otelmetric.Int64UpDownCounterOption) (*UpDownCounter, error) {
	counter, err := m.meter.Int64UpDownCounter(name, append([]otelmetric.Int64UpDownCounterOption{
		otelmetric.WithDescription(description),
	}, opts...)...)
	if err != nil {
		return nil, err
	}

	return &UpDownCounter{counter: counter}, nil
}

// Add adiciona (ou subtrai) um valor ao contador.
func (u *UpDownCounter) Add(ctx context.Context, value int64, attrs ...attribute.KeyValue) {
	u.counter.Add(ctx, value, otelmetric.WithAttributes(attrs...))
}

// Increment incrementa o contador em 1.
func (u *UpDownCounter) Increment(ctx context.Context, attrs ...attribute.KeyValue) {
	u.Add(ctx, 1, attrs...)
}

// Decrement decrementa o contador em 1.
func (u *UpDownCounter) Decrement(ctx context.Context, attrs ...attribute.KeyValue) {
	u.Add(ctx, -1, attrs...)
}

// Histogram representa um histograma de métricas.
type Histogram struct {
	histogram otelmetric.Float64Histogram
}

// NewHistogram cria um novo histograma.
func (m *metricsHelper) NewHistogram(name, description string, opts ...otelmetric.Float64HistogramOption) (*Histogram, error) {
	histogram, err := m.meter.Float64Histogram(name, append([]otelmetric.Float64HistogramOption{
		otelmetric.WithDescription(description),
	}, opts...)...)
	if err != nil {
		return nil, err
	}

	return &Histogram{histogram: histogram}, nil
}

// Record registra um valor no histograma.
func (h *Histogram) Record(ctx context.Context, value float64, attrs ...attribute.KeyValue) {
	h.histogram.Record(ctx, value, otelmetric.WithAttributes(attrs...))
}

// RecordDuration registra uma duração no histograma (em segundos).
func (h *Histogram) RecordDuration(ctx context.Context, duration time.Duration, attrs ...attribute.KeyValue) {
	h.Record(ctx, duration.Seconds(), attrs...)
}

