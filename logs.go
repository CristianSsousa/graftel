// Package graftel fornece helpers para trabalhar com logs OpenTelemetry.
package graftel

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	otellog "go.opentelemetry.io/otel/log"
)

// LogsHelper facilita a criação e uso de logs estruturados.
// Use NewLogsHelper para criar uma instância.
type LogsHelper interface {
	// Log envia um log com nível, mensagem e tags.
	Log(ctx context.Context, level LogLevel, msg string, tags ...attribute.KeyValue)

	// LogWithFields envia um log com campos extras formatados.
	LogWithFields(ctx context.Context, level LogLevel, msg string, fields map[string]interface{}, tags ...attribute.KeyValue)

	// LogWithError envia um log de erro com uma mensagem de erro.
	LogWithError(ctx context.Context, level LogLevel, msg string, err error, tags ...attribute.KeyValue)

	// Trace envia um log de nível trace.
	Trace(ctx context.Context, msg string, tags ...attribute.KeyValue)

	// Debug envia um log de nível debug.
	Debug(ctx context.Context, msg string, tags ...attribute.KeyValue)

	// Info envia um log de nível info.
	Info(ctx context.Context, msg string, tags ...attribute.KeyValue)

	// Warn envia um log de nível warn.
	Warn(ctx context.Context, msg string, tags ...attribute.KeyValue)

	// Error envia um log de nível error.
	Error(ctx context.Context, msg string, tags ...attribute.KeyValue)

	// Fatal envia um log de nível fatal.
	Fatal(ctx context.Context, msg string, tags ...attribute.KeyValue)

	// TraceWithFields envia um log trace com campos extras.
	TraceWithFields(ctx context.Context, msg string, fields map[string]interface{}, tags ...attribute.KeyValue)

	// DebugWithFields envia um log debug com campos extras.
	DebugWithFields(ctx context.Context, msg string, fields map[string]interface{}, tags ...attribute.KeyValue)

	// InfoWithFields envia um log info com campos extras.
	InfoWithFields(ctx context.Context, msg string, fields map[string]interface{}, tags ...attribute.KeyValue)

	// WarnWithFields envia um log warn com campos extras.
	WarnWithFields(ctx context.Context, msg string, fields map[string]interface{}, tags ...attribute.KeyValue)

	// ErrorWithFields envia um log error com campos extras.
	ErrorWithFields(ctx context.Context, msg string, fields map[string]interface{}, tags ...attribute.KeyValue)

	// FatalWithFields envia um log fatal com campos extras.
	FatalWithFields(ctx context.Context, msg string, fields map[string]interface{}, tags ...attribute.KeyValue)

	// ErrorWithError envia um log de erro com uma mensagem de erro.
	ErrorWithError(ctx context.Context, msg string, err error, tags ...attribute.KeyValue)
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

// Log envia um log com nível, mensagem e tags.
func (l *logsHelper) Log(ctx context.Context, level LogLevel, msg string, tags ...attribute.KeyValue) {
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

	// Incluir tags no body da mensagem para garantir que apareçam no Loki
	var bodyMsg string
	if len(tags) > 0 {
		// Formatar tags para incluir na mensagem
		tagsStr := formatTags(tags)
		bodyMsg = msg
		if tagsStr != "" {
			bodyMsg += " " + tagsStr
		}
		record.SetBody(otellog.StringValue(bodyMsg))

		// Também adicionar como atributos para estruturação
		logAttrs := convertAttributes(tags)
		record.AddAttributes(logAttrs...)
	} else {
		record.SetBody(otellog.StringValue(msg))
	}

	// Imprimir no console no formato solicitado (antes de enviar para OpenTelemetry)
	l.printFormattedLog(level, msg, tags)

	// Emitir log via OpenTelemetry (sem impressão automática)
	l.logger.Emit(ctx, record)
}

// LogWithFields envia um log com campos extras formatados.
func (l *logsHelper) LogWithFields(ctx context.Context, level LogLevel, msg string, fields map[string]interface{}, tags ...attribute.KeyValue) {
	// Converter campos para atributos
	allTags := make([]attribute.KeyValue, 0, len(fields)+len(tags))

	for k, v := range fields {
		switch val := v.(type) {
		case string:
			allTags = append(allTags, attribute.String(k, val))
		case int:
			allTags = append(allTags, attribute.Int(k, val))
		case int64:
			allTags = append(allTags, attribute.Int64(k, val))
		case float64:
			allTags = append(allTags, attribute.Float64(k, val))
		case bool:
			allTags = append(allTags, attribute.Bool(k, val))
		default:
			allTags = append(allTags, attribute.String(k, fmt.Sprintf("%v", val)))
		}
	}

	allTags = append(allTags, tags...)
	l.Log(ctx, level, msg, allTags...)
}

// LogWithError envia um log de erro com uma mensagem de erro.
func (l *logsHelper) LogWithError(ctx context.Context, level LogLevel, msg string, err error, tags ...attribute.KeyValue) {
	errorTags := []attribute.KeyValue{
		attribute.String("error", err.Error()),
	}
	errorTags = append(errorTags, tags...)
	l.Log(ctx, level, msg, errorTags...)
}

// Trace envia um log de nível trace.
func (l *logsHelper) Trace(ctx context.Context, msg string, tags ...attribute.KeyValue) {
	l.Log(ctx, LogLevelTrace, msg, tags...)
}

// Debug envia um log de nível debug.
func (l *logsHelper) Debug(ctx context.Context, msg string, tags ...attribute.KeyValue) {
	l.Log(ctx, LogLevelDebug, msg, tags...)
}

// Info envia um log de nível info.
func (l *logsHelper) Info(ctx context.Context, msg string, tags ...attribute.KeyValue) {
	l.Log(ctx, LogLevelInfo, msg, tags...)
}

// Warn envia um log de nível warn.
func (l *logsHelper) Warn(ctx context.Context, msg string, tags ...attribute.KeyValue) {
	l.Log(ctx, LogLevelWarn, msg, tags...)
}

// Error envia um log de nível error.
func (l *logsHelper) Error(ctx context.Context, msg string, tags ...attribute.KeyValue) {
	l.Log(ctx, LogLevelError, msg, tags...)
}

// Fatal envia um log de nível fatal.
func (l *logsHelper) Fatal(ctx context.Context, msg string, tags ...attribute.KeyValue) {
	l.Log(ctx, LogLevelFatal, msg, tags...)
}

// TraceWithFields envia um log trace com campos extras.
func (l *logsHelper) TraceWithFields(ctx context.Context, msg string, fields map[string]interface{}, tags ...attribute.KeyValue) {
	l.LogWithFields(ctx, LogLevelTrace, msg, fields, tags...)
}

// DebugWithFields envia um log debug com campos extras.
func (l *logsHelper) DebugWithFields(ctx context.Context, msg string, fields map[string]interface{}, tags ...attribute.KeyValue) {
	l.LogWithFields(ctx, LogLevelDebug, msg, fields, tags...)
}

// InfoWithFields envia um log info com campos extras.
func (l *logsHelper) InfoWithFields(ctx context.Context, msg string, fields map[string]interface{}, tags ...attribute.KeyValue) {
	l.LogWithFields(ctx, LogLevelInfo, msg, fields, tags...)
}

// WarnWithFields envia um log warn com campos extras.
func (l *logsHelper) WarnWithFields(ctx context.Context, msg string, fields map[string]interface{}, tags ...attribute.KeyValue) {
	l.LogWithFields(ctx, LogLevelWarn, msg, fields, tags...)
}

// ErrorWithFields envia um log error com campos extras.
func (l *logsHelper) ErrorWithFields(ctx context.Context, msg string, fields map[string]interface{}, tags ...attribute.KeyValue) {
	l.LogWithFields(ctx, LogLevelError, msg, fields, tags...)
}

// FatalWithFields envia um log fatal com campos extras.
func (l *logsHelper) FatalWithFields(ctx context.Context, msg string, fields map[string]interface{}, tags ...attribute.KeyValue) {
	l.LogWithFields(ctx, LogLevelFatal, msg, fields, tags...)
}

// ErrorWithError envia um log de erro com uma mensagem de erro.
func (l *logsHelper) ErrorWithError(ctx context.Context, msg string, err error, tags ...attribute.KeyValue) {
	l.LogWithError(ctx, LogLevelError, msg, err, tags...)
}

// convertAttributes converte attribute.KeyValue para otellog.KeyValue,
// prefixando todos os atributos com "tags." para organização.
func convertAttributes(tags []attribute.KeyValue) []otellog.KeyValue {
	logAttrs := make([]otellog.KeyValue, len(tags))
	for i, attr := range tags {
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

func (l *logsHelper) printFormattedLog(level LogLevel, msg string, tags []attribute.KeyValue) {

	tagsStr := formatTags(tags)

	stacktrace := ""
	if level == LogLevelError || level == LogLevelFatal {
		stacktrace = getStackTrace()
	}

	logLine := msg
	if tagsStr != "" {
		logLine += " " + tagsStr
	}
	if stacktrace != "" {
		logLine += " " + stacktrace
	}

	fmt.Fprintln(os.Stderr, logLine)
}

func formatTags(tags []attribute.KeyValue) string {
	if len(tags) == 0 {
		return ""
	}

	var parts []string
	for _, tag := range tags {
		key := string(tag.Key)
		key = strings.TrimPrefix(key, "tags.")

		var value string
		switch tag.Value.Type() {
		case attribute.STRING:
			value = tag.Value.AsString()
		case attribute.INT64:
			value = fmt.Sprintf("%d", tag.Value.AsInt64())
		case attribute.FLOAT64:
			value = fmt.Sprintf("%g", tag.Value.AsFloat64())
		case attribute.BOOL:
			value = fmt.Sprintf("%t", tag.Value.AsBool())
		case attribute.STRINGSLICE:
			value = fmt.Sprintf("%v", tag.Value.AsStringSlice())
		case attribute.INT64SLICE:
			value = fmt.Sprintf("%v", tag.Value.AsInt64Slice())
		case attribute.FLOAT64SLICE:
			value = fmt.Sprintf("%v", tag.Value.AsFloat64Slice())
		case attribute.BOOLSLICE:
			value = fmt.Sprintf("%v", tag.Value.AsBoolSlice())
		default:
			value = tag.Value.AsString()
		}

		parts = append(parts, fmt.Sprintf("[%s:%s]", key, value))
	}

	return strings.Join(parts, "")
}

// getStackTrace retorna o stacktrace formatado
func getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	if n > 0 {
		// Pegar apenas as primeiras linhas do stacktrace (últimas 10 linhas)
		lines := strings.Split(string(buf[:n]), "\n")
		if len(lines) > 10 {
			lines = lines[len(lines)-10:]
		}
		return strings.Join(lines, "\n")
	}
	return ""
}
