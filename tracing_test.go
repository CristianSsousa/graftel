package graftel

import (
	"context"
	"errors"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestNewTracingHelper(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")
	helper := NewTracingHelper(tracer)

	if helper == nil {
		t.Fatal("NewTracingHelper retornou nil")
	}
}

func TestTracingHelper_StartSpan(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")
	helper := NewTracingHelper(tracer)
	ctx := context.Background()

	ctx, span := helper.StartSpan(ctx, "test-span")
	if span == nil {
		t.Fatal("StartSpan retornou span nil")
	}
	defer span.End()
}

func TestTracingHelper_StartSpanWithTags(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")
	helper := NewTracingHelper(tracer)
	ctx := context.Background()

	ctx, span := helper.StartSpanWithTags(ctx, "test-span",
		attribute.String("key1", "value1"),
		attribute.String("key2", "value2"),
	)
	if span == nil {
		t.Fatal("StartSpanWithTags retornou span nil")
	}
	defer span.End()
}

func TestTracingHelper_WithSpan(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")
	helper := NewTracingHelper(tracer)
	ctx := context.Background()

	err := helper.WithSpan(ctx, "test-operation", func(ctx context.Context) error {
		return nil
	}, attribute.String("tag", "value"))

	if err != nil {
		t.Fatalf("WithSpan retornou erro inesperado: %v", err)
	}
}

func TestTracingHelper_WithSpan_Error(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")
	helper := NewTracingHelper(tracer)
	ctx := context.Background()

	expectedErr := errors.New("test error")
	err := helper.WithSpan(ctx, "test-operation", func(ctx context.Context) error {
		return expectedErr
	}, attribute.String("tag", "value"))

	if err != expectedErr {
		t.Fatalf("WithSpan retornou erro diferente: esperado %v, obtido %v", expectedErr, err)
	}
}

func TestTracingHelper_WithSpanAndReturn(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")
	helper := NewTracingHelper(tracer)
	ctx := context.Background()

	result, err := helper.WithSpanAndReturn(ctx, "test-operation", func(ctx context.Context) (interface{}, error) {
		return "result", nil
	}, attribute.String("tag", "value"))

	if err != nil {
		t.Fatalf("WithSpanAndReturn retornou erro: %v", err)
	}

	if result != "result" {
		t.Fatalf("WithSpanAndReturn retornou resultado diferente: esperado 'result', obtido %v", result)
	}
}

func TestTracingHelper_WithSpanAndReturn_Error(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")
	helper := NewTracingHelper(tracer)
	ctx := context.Background()

	expectedErr := errors.New("test error")
	result, err := helper.WithSpanAndReturn(ctx, "test-operation", func(ctx context.Context) (interface{}, error) {
		return nil, expectedErr
	}, attribute.String("tag", "value"))

	if err != expectedErr {
		t.Fatalf("WithSpanAndReturn retornou erro diferente: esperado %v, obtido %v", expectedErr, err)
	}

	if result != nil {
		t.Fatalf("WithSpanAndReturn retornou resultado não-nil em caso de erro: %v", result)
	}
}

func TestTracingHelper_AddSpanTags(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")
	helper := NewTracingHelper(tracer)
	ctx := context.Background()

	ctx, span := helper.StartSpan(ctx, "test-span")
	defer span.End()

	helper.AddSpanTags(ctx,
		attribute.String("key1", "value1"),
		attribute.String("key2", "value2"),
	)
}

func TestTracingHelper_SetSpanError(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")
	helper := NewTracingHelper(tracer)
	ctx := context.Background()

	ctx, span := helper.StartSpan(ctx, "test-span")
	defer span.End()

	err := errors.New("test error")
	helper.SetSpanError(ctx, err, attribute.String("retry", "true"))
}

func TestTracingHelper_SetSpanStatus(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")
	helper := NewTracingHelper(tracer)
	ctx := context.Background()

	ctx, span := helper.StartSpan(ctx, "test-span")
	defer span.End()

	helper.SetSpanStatus(ctx, codes.Ok, "Operação concluída")
	helper.SetSpanStatus(ctx, codes.Error, "Operação falhou")
}

func TestTracingHelper_GetSpan(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")
	helper := NewTracingHelper(tracer)
	ctx := context.Background()

	span := helper.GetSpan(ctx)
	if span == nil {
		t.Fatal("GetSpan retornou span nil")
	}
}

func TestTracingHelper_GetTraceID(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")
	helper := NewTracingHelper(tracer)
	ctx := context.Background()

	ctx, _ = helper.StartSpan(ctx, "test-span")
	traceID := helper.GetTraceID(ctx)

	if traceID == "" {
		t.Log("GetTraceID retornou string vazia (pode ser normal com noop tracer)")
	}
}

func TestTracingHelper_GetSpanID(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")
	helper := NewTracingHelper(tracer)
	ctx := context.Background()

	ctx, _ = helper.StartSpan(ctx, "test-span")
	spanID := helper.GetSpanID(ctx)

	if spanID == "" {
		t.Log("GetSpanID retornou string vazia (pode ser normal com noop tracer)")
	}
}

func TestStartSpan(t *testing.T) {
	ctx := context.Background()
	ctx, span := StartSpan(ctx, "test-span",
		attribute.String("key", "value"),
	)
	if span == nil {
		t.Fatal("StartSpan retornou span nil")
	}
	defer span.End()
}

func TestWithSpan(t *testing.T) {
	ctx := context.Background()

	err := WithSpan(ctx, "test-operation", func(ctx context.Context) error {
		return nil
	}, attribute.String("tag", "value"))

	if err != nil {
		t.Fatalf("WithSpan retornou erro: %v", err)
	}
}

func TestWithSpan_Error(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("test error")

	err := WithSpan(ctx, "test-operation", func(ctx context.Context) error {
		return expectedErr
	}, attribute.String("tag", "value"))

	if err != expectedErr {
		t.Fatalf("WithSpan retornou erro diferente: esperado %v, obtido %v", expectedErr, err)
	}
}

func TestWithSpanTiming(t *testing.T) {
	ctx := context.Background()

	err := WithSpanTiming(ctx, "test-operation", func(ctx context.Context) error {
		return nil
	}, attribute.String("tag", "value"))

	if err != nil {
		t.Fatalf("WithSpanTiming retornou erro: %v", err)
	}
}

func TestWithSpanTiming_Error(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("test error")

	err := WithSpanTiming(ctx, "test-operation", func(ctx context.Context) error {
		return expectedErr
	}, attribute.String("tag", "value"))

	if err != expectedErr {
		t.Fatalf("WithSpanTiming retornou erro diferente: esperado %v, obtido %v", expectedErr, err)
	}
}
