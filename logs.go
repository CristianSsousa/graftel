// Package graftel fornece helpers para trabalhar com logs OpenTelemetry.
package graftel

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	otellog "go.opentelemetry.io/otel/log"
)

// LogsHelper facilita a criação e uso de logs estruturados.
// Use NewLogsHelper para criar uma instância.
type LogsHelper interface {
	// Log envia um log com nível, mensagem e atributos.
	Log(ctx context.Context, level LogLevel, msg string, attrs ...attribute.KeyValue)

	// LogWithFields envia um log com campos extras formatados.
	LogWithFields(ctx context.Context, level LogLevel, msg string, fields map[string]interface{}, attrs ...attribute.KeyValue)

	// LogWithError envia um log de erro com uma mensagem de erro.
	LogWithError(ctx context.Context, level LogLevel, msg string, err error, attrs ...attribute.KeyValue)

	// Trace envia um log de nível trace.
	Trace(ctx context.Context, msg string, attrs ...attribute.KeyValue)

	// Debug envia um log de nível debug.
	Debug(ctx context.Context, msg string, attrs ...attribute.KeyValue)

	// Info envia um log de nível info.
	Info(ctx context.Context, msg string, attrs ...attribute.KeyValue)

	// Warn envia um log de nível warn.
	Warn(ctx context.Context, msg string, attrs ...attribute.KeyValue)

	// Error envia um log de nível error.
	Error(ctx context.Context, msg string, attrs ...attribute.KeyValue)

	// Fatal envia um log de nível fatal.
	Fatal(ctx context.Context, msg string, attrs ...attribute.KeyValue)

	// TraceWithFields envia um log trace com campos extras.
	TraceWithFields(ctx context.Context, msg string, fields map[string]interface{}, attrs ...attribute.KeyValue)

	// DebugWithFields envia um log debug com campos extras.
	DebugWithFields(ctx context.Context, msg string, fields map[string]interface{}, attrs ...attribute.KeyValue)

	// InfoWithFields envia um log info com campos extras.
	InfoWithFields(ctx context.Context, msg string, fields map[string]interface{}, attrs ...attribute.KeyValue)

	// WarnWithFields envia um log warn com campos extras.
	WarnWithFields(ctx context.Context, msg string, fields map[string]interface{}, attrs ...attribute.KeyValue)

	// ErrorWithFields envia um log error com campos extras.
	ErrorWithFields(ctx context.Context, msg string, fields map[string]interface{}, attrs ...attribute.KeyValue)

	// FatalWithFields envia um log fatal com campos extras.
	FatalWithFields(ctx context.Context, msg string, fields map[string]interface{}, attrs ...attribute.KeyValue)

	// ErrorWithError envia um log de erro com uma mensagem de erro.
	ErrorWithError(ctx context.Context, msg string, err error, attrs ...attribute.KeyValue)
}

// logsHelper é a implementação concreta do LogsHelper.
type logsHelper struct {
	logger otellog.Logger
}

// NewLogsHelper cria um novo helper de logs.
func NewLogsHelper(logger otellog.Logger) LogsHelper {
	return &logsHelper{
		logger: logger,
	}
}

// LogLevel representa o nível de log.
type LogLevel int

const (
	// LogLevelTrace representa o nível de log trace.
	LogLevelTrace LogLevel = iota
	// LogLevelDebug representa o nível de log debug.
	LogLevelDebug
	// LogLevelInfo representa o nível de log info.
	LogLevelInfo
	// LogLevelWarn representa o nível de log warn.
	LogLevelWarn
	// LogLevelError representa o nível de log error.
	LogLevelError
	// LogLevelFatal representa o nível de log fatal.
	LogLevelFatal
)

// String retorna a representação em string do nível de log.
func (l LogLevel) String() string {
	switch l {
	case LogLevelTrace:
		return "TRACE"
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	case LogLevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Log envia um log com nível, mensagem e atributos.
func (l *logsHelper) Log(ctx context.Context, level LogLevel, msg string, attrs ...attribute.KeyValue) {
	severity := otellog.SeverityInfo
	switch level {
	case LogLevelTrace:
		severity = otellog.SeverityTrace
	case LogLevelDebug:
		severity = otellog.SeverityDebug
	case LogLevelInfo:
		severity = otellog.SeverityInfo
	case LogLevelWarn:
		severity = otellog.SeverityWarn
	case LogLevelError:
		severity = otellog.SeverityError
	case LogLevelFatal:
		severity = otellog.SeverityFatal
	}

	// Criar record usando a API correta
	record := otellog.Record{}
	record.SetSeverity(severity)
	record.SetBody(otellog.StringValue(msg))
	if len(attrs) > 0 {
		// Converter attribute.KeyValue para otellog.KeyValue
		logAttrs := convertAttributes(attrs)
		record.AddAttributes(logAttrs...)
	}

	l.logger.Emit(ctx, record)
}

// LogWithFields envia um log com campos extras formatados.
func (l *logsHelper) LogWithFields(ctx context.Context, level LogLevel, msg string, fields map[string]interface{}, attrs ...attribute.KeyValue) {
	// Converter campos para atributos
	allAttrs := make([]attribute.KeyValue, 0, len(fields)+len(attrs))

	for k, v := range fields {
		switch val := v.(type) {
		case string:
			allAttrs = append(allAttrs, attribute.String(k, val))
		case int:
			allAttrs = append(allAttrs, attribute.Int(k, val))
		case int64:
			allAttrs = append(allAttrs, attribute.Int64(k, val))
		case float64:
			allAttrs = append(allAttrs, attribute.Float64(k, val))
		case bool:
			allAttrs = append(allAttrs, attribute.Bool(k, val))
		default:
			allAttrs = append(allAttrs, attribute.String(k, fmt.Sprintf("%v", val)))
		}
	}

	allAttrs = append(allAttrs, attrs...)
	l.Log(ctx, level, msg, allAttrs...)
}

// LogWithError envia um log de erro com uma mensagem de erro.
func (l *logsHelper) LogWithError(ctx context.Context, level LogLevel, msg string, err error, attrs ...attribute.KeyValue) {
	errorAttrs := []attribute.KeyValue{
		attribute.String("error", err.Error()),
	}
	errorAttrs = append(errorAttrs, attrs...)
	l.Log(ctx, level, msg, errorAttrs...)
}

// Trace envia um log de nível trace.
func (l *logsHelper) Trace(ctx context.Context, msg string, attrs ...attribute.KeyValue) {
	l.Log(ctx, LogLevelTrace, msg, attrs...)
}

// Debug envia um log de nível debug.
func (l *logsHelper) Debug(ctx context.Context, msg string, attrs ...attribute.KeyValue) {
	l.Log(ctx, LogLevelDebug, msg, attrs...)
}

// Info envia um log de nível info.
func (l *logsHelper) Info(ctx context.Context, msg string, attrs ...attribute.KeyValue) {
	l.Log(ctx, LogLevelInfo, msg, attrs...)
}

// Warn envia um log de nível warn.
func (l *logsHelper) Warn(ctx context.Context, msg string, attrs ...attribute.KeyValue) {
	l.Log(ctx, LogLevelWarn, msg, attrs...)
}

// Error envia um log de nível error.
func (l *logsHelper) Error(ctx context.Context, msg string, attrs ...attribute.KeyValue) {
	l.Log(ctx, LogLevelError, msg, attrs...)
}

// Fatal envia um log de nível fatal.
func (l *logsHelper) Fatal(ctx context.Context, msg string, attrs ...attribute.KeyValue) {
	l.Log(ctx, LogLevelFatal, msg, attrs...)
}

// TraceWithFields envia um log trace com campos extras.
func (l *logsHelper) TraceWithFields(ctx context.Context, msg string, fields map[string]interface{}, attrs ...attribute.KeyValue) {
	l.LogWithFields(ctx, LogLevelTrace, msg, fields, attrs...)
}

// DebugWithFields envia um log debug com campos extras.
func (l *logsHelper) DebugWithFields(ctx context.Context, msg string, fields map[string]interface{}, attrs ...attribute.KeyValue) {
	l.LogWithFields(ctx, LogLevelDebug, msg, fields, attrs...)
}

// InfoWithFields envia um log info com campos extras.
func (l *logsHelper) InfoWithFields(ctx context.Context, msg string, fields map[string]interface{}, attrs ...attribute.KeyValue) {
	l.LogWithFields(ctx, LogLevelInfo, msg, fields, attrs...)
}

// WarnWithFields envia um log warn com campos extras.
func (l *logsHelper) WarnWithFields(ctx context.Context, msg string, fields map[string]interface{}, attrs ...attribute.KeyValue) {
	l.LogWithFields(ctx, LogLevelWarn, msg, fields, attrs...)
}

// ErrorWithFields envia um log error com campos extras.
func (l *logsHelper) ErrorWithFields(ctx context.Context, msg string, fields map[string]interface{}, attrs ...attribute.KeyValue) {
	l.LogWithFields(ctx, LogLevelError, msg, fields, attrs...)
}

// FatalWithFields envia um log fatal com campos extras.
func (l *logsHelper) FatalWithFields(ctx context.Context, msg string, fields map[string]interface{}, attrs ...attribute.KeyValue) {
	l.LogWithFields(ctx, LogLevelFatal, msg, fields, attrs...)
}

// ErrorWithError envia um log de erro com uma mensagem de erro.
func (l *logsHelper) ErrorWithError(ctx context.Context, msg string, err error, attrs ...attribute.KeyValue) {
	l.LogWithError(ctx, LogLevelError, msg, err, attrs...)
}

// convertAttributes converte attribute.KeyValue para otellog.KeyValue,
// prefixando todos os atributos com "tags." para organização.
func convertAttributes(attrs []attribute.KeyValue) []otellog.KeyValue {
	logAttrs := make([]otellog.KeyValue, len(attrs))
	for i, attr := range attrs {
		// Obter a chave como string (Key é um tipo string)
		keyStr := string(attr.Key)

		// Prefixar com "tags." se ainda não tiver o prefixo
		if !strings.HasPrefix(keyStr, "tags.") {
			keyStr = "tags." + keyStr
		}

		// Criar novo atributo com a chave prefixada
		var newAttr attribute.KeyValue
		switch attr.Value.Type() {
		case attribute.STRING:
			newAttr = attribute.String(keyStr, attr.Value.AsString())
		case attribute.INT64:
			newAttr = attribute.Int64(keyStr, attr.Value.AsInt64())
		case attribute.FLOAT64:
			newAttr = attribute.Float64(keyStr, attr.Value.AsFloat64())
		case attribute.BOOL:
			newAttr = attribute.Bool(keyStr, attr.Value.AsBool())
		case attribute.STRINGSLICE:
			newAttr = attribute.StringSlice(keyStr, attr.Value.AsStringSlice())
		case attribute.INT64SLICE:
			newAttr = attribute.Int64Slice(keyStr, attr.Value.AsInt64Slice())
		case attribute.FLOAT64SLICE:
			newAttr = attribute.Float64Slice(keyStr, attr.Value.AsFloat64Slice())
		case attribute.BOOLSLICE:
			newAttr = attribute.BoolSlice(keyStr, attr.Value.AsBoolSlice())
		default:
			// Fallback para string
			newAttr = attribute.String(keyStr, attr.Value.AsString())
		}

		logAttrs[i] = otellog.KeyValueFromAttribute(newAttr)
	}
	return logAttrs
}
