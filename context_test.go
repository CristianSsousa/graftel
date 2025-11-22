package graftel

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/log/noop"
)

func TestWithTags(t *testing.T) {
	ctx := context.Background()
	ctx = WithTags(ctx,
		attribute.String("key1", "value1"),
		attribute.String("key2", "value2"),
	)

	tags := GetTagsFromContext(ctx)
	if len(tags) != 2 {
		t.Fatalf("Esperado 2 tags, obtido %d", len(tags))
	}
}

func TestGetTagsFromContext(t *testing.T) {
	ctx := context.Background()
	tags := GetTagsFromContext(ctx)
	if tags != nil && len(tags) != 0 {
		t.Fatalf("Esperado tags vazias, obtido %d tags", len(tags))
	}

	ctx = WithTags(ctx, attribute.String("key", "value"))
	tags = GetTagsFromContext(ctx)
	if len(tags) != 1 {
		t.Fatalf("Esperado 1 tag, obtido %d", len(tags))
	}
}

func TestMergeContextTags(t *testing.T) {
	ctx := context.Background()
	ctx = WithTags(ctx, attribute.String("key1", "value1"))
	ctx = MergeContextTags(ctx, attribute.String("key2", "value2"))

	tags := GetTagsFromContext(ctx)
	if len(tags) != 2 {
		t.Fatalf("Esperado 2 tags após merge, obtido %d", len(tags))
	}
}

func TestNewContextLogger(t *testing.T) {
	logger := noop.NewLoggerProvider().Logger("test")
	ctx := context.Background()
	ctx = WithTags(ctx, attribute.String("user_id", "123"))

	ctxLogger := NewContextLogger(NewLogsHelper(logger), ctx)
	if ctxLogger == nil {
		t.Fatal("NewContextLogger retornou nil")
	}
}

func TestContextLogger_WithTags(t *testing.T) {
	logger := noop.NewLoggerProvider().Logger("test")
	ctx := context.Background()
	ctxLogger := NewContextLogger(NewLogsHelper(logger), ctx)

	newLogger := ctxLogger.WithTags(attribute.String("key", "value"))
	if newLogger == nil {
		t.Fatal("WithTags retornou nil")
	}

	tags := GetTagsFromContext(newLogger.ctx)
	if len(tags) == 0 {
		t.Fatal("WithTags deve adicionar tags ao contexto")
	}
}

func TestContextLogger_Log(t *testing.T) {
	logger := noop.NewLoggerProvider().Logger("test")
	ctx := context.Background()
	ctx = WithTags(ctx, attribute.String("user_id", "123"))

	ctxLogger := NewContextLogger(NewLogsHelper(logger), ctx)
	ctxLogger.Log(LogLevelInfo, "test message", attribute.String("additional", "tag"))
}

func TestContextLogger_Trace(t *testing.T) {
	logger := noop.NewLoggerProvider().Logger("test")
	ctx := context.Background()
	ctxLogger := NewContextLogger(NewLogsHelper(logger), ctx)
	ctxLogger.Trace("test message")
}

func TestContextLogger_Debug(t *testing.T) {
	logger := noop.NewLoggerProvider().Logger("test")
	ctx := context.Background()
	ctxLogger := NewContextLogger(NewLogsHelper(logger), ctx)
	ctxLogger.Debug("test message")
}

func TestContextLogger_Info(t *testing.T) {
	logger := noop.NewLoggerProvider().Logger("test")
	ctx := context.Background()
	ctxLogger := NewContextLogger(NewLogsHelper(logger), ctx)
	ctxLogger.Info("test message")
}

func TestContextLogger_Warn(t *testing.T) {
	logger := noop.NewLoggerProvider().Logger("test")
	ctx := context.Background()
	ctxLogger := NewContextLogger(NewLogsHelper(logger), ctx)
	ctxLogger.Warn("test message")
}

func TestContextLogger_Error(t *testing.T) {
	logger := noop.NewLoggerProvider().Logger("test")
	ctx := context.Background()
	ctxLogger := NewContextLogger(NewLogsHelper(logger), ctx)
	ctxLogger.Error("test message")
}

func TestContextLogger_Fatal(t *testing.T) {
	logger := noop.NewLoggerProvider().Logger("test")
	ctx := context.Background()
	ctxLogger := NewContextLogger(NewLogsHelper(logger), ctx)
	ctxLogger.Fatal("test message")
}

func TestContextLogger_LogWithFields(t *testing.T) {
	logger := noop.NewLoggerProvider().Logger("test")
	ctx := context.Background()
	ctxLogger := NewContextLogger(NewLogsHelper(logger), ctx)

	fields := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}
	ctxLogger.LogWithFields(LogLevelInfo, "test message", fields, attribute.String("tag", "value"))
}

func TestContextLogger_LogWithError(t *testing.T) {
	logger := noop.NewLoggerProvider().Logger("test")
	ctx := context.Background()
	ctxLogger := NewContextLogger(NewLogsHelper(logger), ctx)

	err := context.DeadlineExceeded
	ctxLogger.LogWithError(LogLevelError, "test error", err, attribute.String("tag", "value"))
}

func TestContextLogger_ErrorWithError(t *testing.T) {
	logger := noop.NewLoggerProvider().Logger("test")
	ctx := context.Background()
	ctxLogger := NewContextLogger(NewLogsHelper(logger), ctx)

	err := context.DeadlineExceeded
	ctxLogger.ErrorWithError("test error", err, attribute.String("tag", "value"))
}

func TestContextLogger_InheritsTags(t *testing.T) {
	logger := noop.NewLoggerProvider().Logger("test")
	ctx := context.Background()
	ctx = WithTags(ctx,
		attribute.String("user_id", "123"),
		attribute.String("session_id", "sess-456"),
	)

	ctxLogger := NewContextLogger(NewLogsHelper(logger), ctx)

	tags := GetTagsFromContext(ctxLogger.ctx)
	if len(tags) != 2 {
		t.Fatalf("ContextLogger deve herdar tags do contexto: esperado 2, obtido %d", len(tags))
	}

	ctxLogger.Info("test message")
}
