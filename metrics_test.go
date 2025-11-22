package graftel

import (
	"context"
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

func TestNewMetricsHelper(t *testing.T) {
	reader := metric.NewManualReader()
	meterProvider := metric.NewMeterProvider(
		metric.WithReader(reader),
		metric.WithResource(resource.Empty()),
	)
	meter := meterProvider.Meter("test")

	helper := NewMetricsHelper(meter)

	if helper == nil {
		t.Fatal("NewMetricsHelper() retornou nil")
	}
}

func TestMetricsHelper_NewCounter(t *testing.T) {
	reader := metric.NewManualReader()
	meterProvider := metric.NewMeterProvider(
		metric.WithReader(reader),
		metric.WithResource(resource.Empty()),
	)
	meter := meterProvider.Meter("test")
	helper := NewMetricsHelper(meter)

	counter, err := helper.NewCounter("test_counter", "Test counter description")
	if err != nil {
		t.Fatalf("NewCounter() erro = %v", err)
	}

	if counter == nil {
		t.Fatal("NewCounter() retornou nil")
	}
}

func TestCounter_Add(t *testing.T) {
	reader := metric.NewManualReader()
	meterProvider := metric.NewMeterProvider(
		metric.WithReader(reader),
		metric.WithResource(resource.Empty()),
	)
	meter := meterProvider.Meter("test")
	helper := NewMetricsHelper(meter)

	counter, err := helper.NewCounter("test_counter", "Test counter")
	if err != nil {
		t.Fatalf("NewCounter() erro = %v", err)
	}

	ctx := context.Background()
	counter.Add(ctx, 10, attribute.String("key", "value"))
	counter.Add(ctx, 5, attribute.Int("num", 42))
}

func TestCounter_Increment(t *testing.T) {
	reader := metric.NewManualReader()
	meterProvider := metric.NewMeterProvider(
		metric.WithReader(reader),
		metric.WithResource(resource.Empty()),
	)
	meter := meterProvider.Meter("test")
	helper := NewMetricsHelper(meter)

	counter, err := helper.NewCounter("test_counter", "Test counter")
	if err != nil {
		t.Fatalf("NewCounter() erro = %v", err)
	}

	ctx := context.Background()
	counter.Increment(ctx, attribute.String("key", "value"))
	counter.Increment(ctx)
}

func TestMetricsHelper_NewUpDownCounter(t *testing.T) {
	reader := metric.NewManualReader()
	meterProvider := metric.NewMeterProvider(
		metric.WithReader(reader),
		metric.WithResource(resource.Empty()),
	)
	meter := meterProvider.Meter("test")
	helper := NewMetricsHelper(meter)

	counter, err := helper.NewUpDownCounter("test_updown", "Test up-down counter")
	if err != nil {
		t.Fatalf("NewUpDownCounter() erro = %v", err)
	}

	if counter == nil {
		t.Fatal("NewUpDownCounter() retornou nil")
	}
}

func TestUpDownCounter_Add(t *testing.T) {
	reader := metric.NewManualReader()
	meterProvider := metric.NewMeterProvider(
		metric.WithReader(reader),
		metric.WithResource(resource.Empty()),
	)
	meter := meterProvider.Meter("test")
	helper := NewMetricsHelper(meter)

	counter, err := helper.NewUpDownCounter("test_updown", "Test")
	if err != nil {
		t.Fatalf("NewUpDownCounter() erro = %v", err)
	}

	ctx := context.Background()
	counter.Add(ctx, 10)
	counter.Add(ctx, -5)
}

func TestUpDownCounter_Increment(t *testing.T) {
	reader := metric.NewManualReader()
	meterProvider := metric.NewMeterProvider(
		metric.WithReader(reader),
		metric.WithResource(resource.Empty()),
	)
	meter := meterProvider.Meter("test")
	helper := NewMetricsHelper(meter)

	counter, err := helper.NewUpDownCounter("test_updown", "Test")
	if err != nil {
		t.Fatalf("NewUpDownCounter() erro = %v", err)
	}

	ctx := context.Background()
	counter.Increment(ctx)
}

func TestUpDownCounter_Decrement(t *testing.T) {
	reader := metric.NewManualReader()
	meterProvider := metric.NewMeterProvider(
		metric.WithReader(reader),
		metric.WithResource(resource.Empty()),
	)
	meter := meterProvider.Meter("test")
	helper := NewMetricsHelper(meter)

	counter, err := helper.NewUpDownCounter("test_updown", "Test")
	if err != nil {
		t.Fatalf("NewUpDownCounter() erro = %v", err)
	}

	ctx := context.Background()
	counter.Decrement(ctx)
}

func TestMetricsHelper_NewHistogram(t *testing.T) {
	reader := metric.NewManualReader()
	meterProvider := metric.NewMeterProvider(
		metric.WithReader(reader),
		metric.WithResource(resource.Empty()),
	)
	meter := meterProvider.Meter("test")
	helper := NewMetricsHelper(meter)

	histogram, err := helper.NewHistogram("test_histogram", "Test histogram")
	if err != nil {
		t.Fatalf("NewHistogram() erro = %v", err)
	}

	if histogram == nil {
		t.Fatal("NewHistogram() retornou nil")
	}
}

func TestHistogram_Record(t *testing.T) {
	reader := metric.NewManualReader()
	meterProvider := metric.NewMeterProvider(
		metric.WithReader(reader),
		metric.WithResource(resource.Empty()),
	)
	meter := meterProvider.Meter("test")
	helper := NewMetricsHelper(meter)

	histogram, err := helper.NewHistogram("test_histogram", "Test")
	if err != nil {
		t.Fatalf("NewHistogram() erro = %v", err)
	}

	ctx := context.Background()
	histogram.Record(ctx, 42.5, attribute.String("key", "value"))
	histogram.Record(ctx, 100.0)
}

func TestHistogram_RecordDuration(t *testing.T) {
	reader := metric.NewManualReader()
	meterProvider := metric.NewMeterProvider(
		metric.WithReader(reader),
		metric.WithResource(resource.Empty()),
	)
	meter := meterProvider.Meter("test")
	helper := NewMetricsHelper(meter)

	histogram, err := helper.NewHistogram("test_histogram", "Test")
	if err != nil {
		t.Fatalf("NewHistogram() erro = %v", err)
	}

	ctx := context.Background()
	duration := 100 * time.Millisecond
	histogram.RecordDuration(ctx, duration, attribute.String("key", "value"))
	histogram.RecordDuration(ctx, duration)
}

func TestMetricsHelper_NewGauge(t *testing.T) {
	reader := metric.NewManualReader()
	meterProvider := metric.NewMeterProvider(
		metric.WithReader(reader),
		metric.WithResource(resource.Empty()),
	)
	meter := meterProvider.Meter("test")
	helper := NewMetricsHelper(meter)

	callback := func(ctx context.Context, obs otelmetric.Float64Observer) error {
		obs.Observe(42.0)
		return nil
	}

	gauge, err := helper.NewGauge("test_gauge", "Test gauge", callback)
	if err != nil {
		t.Fatalf("NewGauge() erro = %v", err)
	}

	if gauge == nil {
		t.Fatal("NewGauge() retornou nil")
	}
}

func TestCounter_WithAttributes(t *testing.T) {
	reader := metric.NewManualReader()
	meterProvider := metric.NewMeterProvider(
		metric.WithReader(reader),
		metric.WithResource(resource.Empty()),
	)
	meter := meterProvider.Meter("test")
	helper := NewMetricsHelper(meter)

	counter, err := helper.NewCounter("test_counter", "Test")
	if err != nil {
		t.Fatalf("NewCounter() erro = %v", err)
	}

	ctx := context.Background()
	attrs := []attribute.KeyValue{
		attribute.String("method", "GET"),
		attribute.String("path", "/api"),
		attribute.Int("status", 200),
	}

	counter.Add(ctx, 1, attrs...)
	counter.Increment(ctx, attrs...)
}

func TestHistogram_WithAttributes(t *testing.T) {
	reader := metric.NewManualReader()
	meterProvider := metric.NewMeterProvider(
		metric.WithReader(reader),
		metric.WithResource(resource.Empty()),
	)
	meter := meterProvider.Meter("test")
	helper := NewMetricsHelper(meter)

	histogram, err := helper.NewHistogram("test_histogram", "Test")
	if err != nil {
		t.Fatalf("NewHistogram() erro = %v", err)
	}

	ctx := context.Background()
	attrs := []attribute.KeyValue{
		attribute.String("method", "GET"),
		attribute.String("path", "/api"),
	}

	histogram.Record(ctx, 42.5, attrs...)
	histogram.RecordDuration(ctx, 100*time.Millisecond, attrs...)
}

func TestUpDownCounter_WithAttributes(t *testing.T) {
	reader := metric.NewManualReader()
	meterProvider := metric.NewMeterProvider(
		metric.WithReader(reader),
		metric.WithResource(resource.Empty()),
	)
	meter := meterProvider.Meter("test")
	helper := NewMetricsHelper(meter)

	counter, err := helper.NewUpDownCounter("test_updown", "Test")
	if err != nil {
		t.Fatalf("NewUpDownCounter() erro = %v", err)
	}

	ctx := context.Background()
	attrs := []attribute.KeyValue{
		attribute.String("key", "value"),
	}

	counter.Add(ctx, 10, attrs...)
	counter.Increment(ctx, attrs...)
	counter.Decrement(ctx, attrs...)
}
