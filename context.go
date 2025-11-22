package graftel

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
)

type contextKey string

const (
	tagsContextKey contextKey = "graftel.tags"
)

func WithTags(ctx context.Context, tags ...attribute.KeyValue) context.Context {
	existingTags := GetTagsFromContext(ctx)
	allTags := append(existingTags, tags...)
	return context.WithValue(ctx, tagsContextKey, allTags)
}

func GetTagsFromContext(ctx context.Context) []attribute.KeyValue {
	if tags, ok := ctx.Value(tagsContextKey).([]attribute.KeyValue); ok {
		return tags
	}
	return nil
}

func MergeContextTags(ctx context.Context, additionalTags ...attribute.KeyValue) context.Context {
	existingTags := GetTagsFromContext(ctx)
	mergedTags := make([]attribute.KeyValue, 0, len(existingTags)+len(additionalTags))
	mergedTags = append(mergedTags, existingTags...)
	mergedTags = append(mergedTags, additionalTags...)
	return context.WithValue(ctx, tagsContextKey, mergedTags)
}

type ContextLogger struct {
	logger LogsHelper
	ctx    context.Context
}

func NewContextLogger(logger LogsHelper, ctx context.Context) *ContextLogger {
	return &ContextLogger{
		logger: logger,
		ctx:    ctx,
	}
}

func (cl *ContextLogger) WithTags(tags ...attribute.KeyValue) *ContextLogger {
	cl.ctx = WithTags(cl.ctx, tags...)
	return cl
}

func (cl *ContextLogger) Log(level LogLevel, msg string, tags ...attribute.KeyValue) {
	ctxTags := GetTagsFromContext(cl.ctx)
	allTags := append(ctxTags, tags...)
	cl.logger.Log(cl.ctx, level, msg, allTags...)
}

func (cl *ContextLogger) Trace(msg string, tags ...attribute.KeyValue) {
	cl.Log(LogLevelTrace, msg, tags...)
}

func (cl *ContextLogger) Debug(msg string, tags ...attribute.KeyValue) {
	cl.Log(LogLevelDebug, msg, tags...)
}

func (cl *ContextLogger) Info(msg string, tags ...attribute.KeyValue) {
	cl.Log(LogLevelInfo, msg, tags...)
}

func (cl *ContextLogger) Warn(msg string, tags ...attribute.KeyValue) {
	cl.Log(LogLevelWarn, msg, tags...)
}

func (cl *ContextLogger) Error(msg string, tags ...attribute.KeyValue) {
	cl.Log(LogLevelError, msg, tags...)
}

func (cl *ContextLogger) Fatal(msg string, tags ...attribute.KeyValue) {
	cl.Log(LogLevelFatal, msg, tags...)
}

func (cl *ContextLogger) LogWithFields(level LogLevel, msg string, fields map[string]interface{}, tags ...attribute.KeyValue) {
	ctxTags := GetTagsFromContext(cl.ctx)
	allTags := append(ctxTags, tags...)
	cl.logger.LogWithFields(cl.ctx, level, msg, fields, allTags...)
}

func (cl *ContextLogger) LogWithError(level LogLevel, msg string, err error, tags ...attribute.KeyValue) {
	ctxTags := GetTagsFromContext(cl.ctx)
	allTags := append(ctxTags, tags...)
	cl.logger.LogWithError(cl.ctx, level, msg, err, allTags...)
}

func (cl *ContextLogger) ErrorWithError(msg string, err error, tags ...attribute.KeyValue) {
	cl.LogWithError(LogLevelError, msg, err, tags...)
}
