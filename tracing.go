package graftel

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type TracingHelper interface {
	StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span)
	StartSpanWithTags(ctx context.Context, name string, tags ...attribute.KeyValue) (context.Context, trace.Span)
	WithSpan(ctx context.Context, name string, fn func(context.Context) error, tags ...attribute.KeyValue) error
	WithSpanAndReturn(ctx context.Context, name string, fn func(context.Context) (interface{}, error), tags ...attribute.KeyValue) (interface{}, error)
	AddSpanTags(ctx context.Context, tags ...attribute.KeyValue)
	SetSpanError(ctx context.Context, err error, tags ...attribute.KeyValue)
	SetSpanStatus(ctx context.Context, code codes.Code, description string)
	GetSpan(ctx context.Context) trace.Span
	GetTraceID(ctx context.Context) string
	GetSpanID(ctx context.Context) string
}

type tracingHelper struct {
	tracer trace.Tracer
}

func NewTracingHelper(tracer trace.Tracer) TracingHelper {
	return &tracingHelper{
		tracer: tracer,
	}
}

func (t *tracingHelper) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, name, opts...)
}

func (t *tracingHelper) StartSpanWithTags(ctx context.Context, name string, tags ...attribute.KeyValue) (context.Context, trace.Span) {
	ctx, span := t.tracer.Start(ctx, name)
	if len(tags) > 0 {
		span.SetAttributes(tags...)
	}
	return ctx, span
}

func (t *tracingHelper) WithSpan(ctx context.Context, name string, fn func(context.Context) error, tags ...attribute.KeyValue) error {
	ctx, span := t.StartSpanWithTags(ctx, name, tags...)
	defer span.End()

	err := fn(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (t *tracingHelper) WithSpanAndReturn(ctx context.Context, name string, fn func(context.Context) (interface{}, error), tags ...attribute.KeyValue) (interface{}, error) {
	ctx, span := t.StartSpanWithTags(ctx, name, tags...)
	defer span.End()

	result, err := fn(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(codes.Ok, "")
	return result, nil
}

func (t *tracingHelper) AddSpanTags(ctx context.Context, tags ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(tags...)
	}
}

func (t *tracingHelper) SetSpanError(ctx context.Context, err error, tags ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if len(tags) > 0 {
			span.SetAttributes(tags...)
		}
	}
}

func (t *tracingHelper) SetSpanStatus(ctx context.Context, code codes.Code, description string) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetStatus(code, description)
	}
}

func (t *tracingHelper) GetSpan(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

func (t *tracingHelper) GetTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

func (t *tracingHelper) GetSpanID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}

func StartSpan(ctx context.Context, name string, tags ...attribute.KeyValue) (context.Context, trace.Span) {
	tracerProvider := trace.SpanFromContext(ctx).TracerProvider()
	if tracerProvider == nil {
		tracerProvider = otel.GetTracerProvider()
	}
	ctx, newSpan := tracerProvider.Tracer("graftel").Start(ctx, name)
	if len(tags) > 0 {
		newSpan.SetAttributes(tags...)
	}
	return ctx, newSpan
}

func WithSpan(ctx context.Context, name string, fn func(context.Context) error, tags ...attribute.KeyValue) error {
	ctx, span := StartSpan(ctx, name, tags...)
	defer span.End()

	err := fn(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func WithSpanTiming(ctx context.Context, name string, fn func(context.Context) error, tags ...attribute.KeyValue) error {
	ctx, span := StartSpan(ctx, name, tags...)
	defer span.End()

	start := time.Now()
	err := fn(ctx)
	duration := time.Since(start)

	span.SetAttributes(attribute.String("duration", duration.String()))
	span.SetAttributes(attribute.Int64("duration_ms", duration.Milliseconds()))

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}
