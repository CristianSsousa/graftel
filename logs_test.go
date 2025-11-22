package graftel

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

func createTestLogger() otellog.Logger {
	res, _ := resource.New(context.Background())
	provider := log.NewLoggerProvider(log.WithResource(res))
	return provider.Logger("test")
}

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{LogLevelTrace, "TRACE"},
		{LogLevelDebug, "DEBUG"},
		{LogLevelInfo, "INFO"},
		{LogLevelWarn, "WARN"},
		{LogLevelError, "ERROR"},
		{LogLevelFatal, "FATAL"},
		{LogLevel(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		if got := tt.level.String(); got != tt.expected {
			t.Errorf("LogLevel.String() = %v, want %v", got, tt.expected)
		}
	}
}

func TestNewLogsHelper(t *testing.T) {
	logger := createTestLogger()
	helper := NewLogsHelper(logger)

	if helper == nil {
		t.Fatal("NewLogsHelper() retornou nil")
	}
}

func TestLogsHelper_Log(t *testing.T) {
	logger := createTestLogger()
	helper := NewLogsHelper(logger).(*logsHelper)

	ctx := context.Background()
	helper.Log(ctx, LogLevelInfo, "test message",
		attribute.String("key1", "value1"),
		attribute.Int("key2", 42),
	)
}

func TestLogsHelper_LogLevels(t *testing.T) {
	logger := createTestLogger()
	helper := NewLogsHelper(logger).(*logsHelper)
	ctx := context.Background()

	levels := []struct {
		method func(context.Context, string, ...attribute.KeyValue)
		level  LogLevel
	}{
		{helper.Trace, LogLevelTrace},
		{helper.Debug, LogLevelDebug},
		{helper.Info, LogLevelInfo},
		{helper.Warn, LogLevelWarn},
		{helper.Error, LogLevelError},
		{helper.Fatal, LogLevelFatal},
	}

	for _, tt := range levels {
		tt.method(ctx, "test", attribute.String("tag", "value"))
	}
}

func TestLogsHelper_LogWithFields(t *testing.T) {
	logger := createTestLogger()
	helper := NewLogsHelper(logger).(*logsHelper)
	ctx := context.Background()

	fields := map[string]interface{}{
		"field1": "value1",
		"field2": 42,
		"field3": true,
	}

	helper.LogWithFields(ctx, LogLevelInfo, "test", fields,
		attribute.String("tag1", "value1"),
	)
}

func TestLogsHelper_LogWithError(t *testing.T) {
	logger := createTestLogger()
	helper := NewLogsHelper(logger).(*logsHelper)
	ctx := context.Background()

	err := fmt.Errorf("test error")
	helper.LogWithError(ctx, LogLevelError, "test message", err,
		attribute.String("tag", "value"),
	)
}

func TestFormatTags(t *testing.T) {
	tests := []struct {
		name     string
		tags     []attribute.KeyValue
		expected string
	}{
		{
			name:     "empty tags",
			tags:     []attribute.KeyValue{},
			expected: "",
		},
		{
			name: "single tag",
			tags: []attribute.KeyValue{
				attribute.String("key1", "value1"),
			},
			expected: "[key1:value1]",
		},
		{
			name: "multiple tags",
			tags: []attribute.KeyValue{
				attribute.String("key1", "value1"),
				attribute.Int("key2", 42),
				attribute.Bool("key3", true),
			},
			expected: "[key1:value1][key2:42][key3:true]",
		},
		{
			name: "tags with prefix",
			tags: []attribute.KeyValue{
				attribute.String("tags.key1", "value1"),
			},
			expected: "[key1:value1]",
		},
		{
			name: "float64 tag",
			tags: []attribute.KeyValue{
				attribute.Float64("key1", 3.14),
			},
			expected: "[key1:3.14]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTags(tt.tags)
			if result != tt.expected {
				t.Errorf("formatTags() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConvertAttributes(t *testing.T) {
	tags := []attribute.KeyValue{
		attribute.String("key1", "value1"),
		attribute.Int("key2", 42),
		attribute.Bool("key3", true),
	}

	result := convertAttributes(tags)

	if len(result) != len(tags) {
		t.Fatalf("esperado %d atributos, obtido %d", len(tags), len(result))
	}

	for i, attr := range result {
		key := string(attr.Key)
		if !strings.HasPrefix(key, "tags.") {
			t.Errorf("atributo %d não tem prefixo tags.: %s", i, key)
		}
	}
}

func TestConvertAttributes_WithPrefix(t *testing.T) {
	tags := []attribute.KeyValue{
		attribute.String("tags.key1", "value1"),
	}

	result := convertAttributes(tags)

	if len(result) != 1 {
		t.Fatalf("esperado 1 atributo, obtido %d", len(result))
	}

	key := string(result[0].Key)
	if key != "tags.key1" {
		t.Errorf("chave = %v, esperado tags.key1", key)
	}
}

func TestGetStackTrace(t *testing.T) {
	stack := getStackTrace()

	if stack == "" {
		t.Error("getStackTrace() retornou string vazia")
	}

	if !strings.Contains(stack, "goroutine") {
		t.Error("stacktrace não contém 'goroutine'")
	}
}

func TestLogsHelper_WithFieldsMethods(t *testing.T) {
	logger := createTestLogger()
	helper := NewLogsHelper(logger).(*logsHelper)
	ctx := context.Background()
	fields := map[string]interface{}{"field": "value"}

	methods := []func(context.Context, string, map[string]interface{}, ...attribute.KeyValue){
		helper.TraceWithFields,
		helper.DebugWithFields,
		helper.InfoWithFields,
		helper.WarnWithFields,
		helper.ErrorWithFields,
		helper.FatalWithFields,
	}

	for _, method := range methods {
		method(ctx, "test", fields)
	}
}

func TestLogsHelper_ErrorWithError(t *testing.T) {
	logger := createTestLogger()
	helper := NewLogsHelper(logger).(*logsHelper)
	ctx := context.Background()

	err := fmt.Errorf("test error")
	helper.ErrorWithError(ctx, "test", err, attribute.String("tag", "value"))
}

func TestFormatTags_AllTypes(t *testing.T) {
	tags := []attribute.KeyValue{
		attribute.String("str", "value"),
		attribute.Int64("int64", 42),
		attribute.Float64("float64", 3.14),
		attribute.Bool("bool", true),
		attribute.StringSlice("strslice", []string{"a", "b"}),
		attribute.Int64Slice("int64slice", []int64{1, 2}),
		attribute.Float64Slice("float64slice", []float64{1.1, 2.2}),
		attribute.BoolSlice("boolslice", []bool{true, false}),
	}

	result := formatTags(tags)

	if result == "" {
		t.Error("formatTags() retornou string vazia")
	}

	if !strings.Contains(result, "[str:value]") {
		t.Error("resultado não contém tag string")
	}

	if !strings.Contains(result, "[int64:42]") {
		t.Error("resultado não contém tag int64")
	}

	if !strings.Contains(result, "[bool:true]") {
		t.Error("resultado não contém tag bool")
	}
}

func TestLogsHelper_LogSeverityMapping(t *testing.T) {
	logger := createTestLogger()
	helper := NewLogsHelper(logger).(*logsHelper)
	ctx := context.Background()

	levels := []LogLevel{
		LogLevelTrace,
		LogLevelDebug,
		LogLevelInfo,
		LogLevelWarn,
		LogLevelError,
		LogLevelFatal,
	}

	for _, level := range levels {
		helper.Log(ctx, level, "test")
	}
}

func TestLogsHelper_LogBodyWithTags(t *testing.T) {
	logger := createTestLogger()
	helper := NewLogsHelper(logger).(*logsHelper)
	ctx := context.Background()

	helper.Log(ctx, LogLevelInfo, "test message",
		attribute.String("key1", "value1"),
		attribute.Int("key2", 42),
	)
}

func TestLogsHelper_LogWithoutTags(t *testing.T) {
	logger := createTestLogger()
	helper := NewLogsHelper(logger).(*logsHelper)
	ctx := context.Background()

	helper.Log(ctx, LogLevelInfo, "test message")
}
