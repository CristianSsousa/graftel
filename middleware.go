package graftel

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

type MiddlewareConfig struct {
	ServiceName        string
	SkipPaths          []string
	RecordRequestBody  bool
	RecordResponseBody bool
	MaxBodySize        int64
}

func DefaultMiddlewareConfig(serviceName string) MiddlewareConfig {
	return MiddlewareConfig{
		ServiceName:        serviceName,
		SkipPaths:          []string{"/health", "/metrics", "/ready"},
		RecordRequestBody:  false,
		RecordResponseBody: false,
		MaxBodySize:        4096,
	}
}

func shouldSkip(path string, skipPaths []string) bool {
	for _, skip := range skipPaths {
		if path == skip {
			return true
		}
	}
	return false
}

func HTTPMiddleware(client Client, config MiddlewareConfig) func(http.Handler) http.Handler {
	tracing := client.NewTracingHelper(config.ServiceName)
	metrics := client.NewMetricsHelper(config.ServiceName + "/http")
	logs := client.NewLogsHelper(config.ServiceName + "/http")

	requestCounter, _ := metrics.NewCounter("http_requests_total", "Total de requisições HTTP")
	requestDuration, _ := metrics.NewHistogram("http_request_duration_seconds", "Duração das requisições HTTP em segundos")
	requestSize, _ := metrics.NewHistogram("http_request_size_bytes", "Tamanho das requisições HTTP em bytes")
	responseSize, _ := metrics.NewHistogram("http_response_size_bytes", "Tamanho das respostas HTTP em bytes")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if shouldSkip(r.URL.Path, config.SkipPaths) {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			ctx := r.Context()

			ctx, span := tracing.StartSpan(ctx, "http.request",
				trace.WithAttributes(
					semconv.HTTPMethodKey.String(r.Method),
					semconv.HTTPURLKey.String(r.URL.String()),
					semconv.HTTPRouteKey.String(r.URL.Path),
					attribute.String("http.user_agent", r.UserAgent()),
					attribute.String("http.client_ip", r.RemoteAddr),
				),
			)

			ctx = WithTags(ctx,
				attribute.String("http.method", r.Method),
				attribute.String("http.path", r.URL.Path),
				attribute.String("http.user_agent", r.UserAgent()),
				attribute.String("http.remote_addr", r.RemoteAddr),
			)

			requestSizeBytes := r.ContentLength
			if requestSizeBytes > 0 {
				requestSize.Record(ctx, float64(requestSizeBytes),
					attribute.String("method", r.Method),
					attribute.String("path", r.URL.Path),
				)
			}

			ctxLogger := NewContextLogger(logs, ctx)

			ctxLogger.Info("Requisição recebida",
				attribute.String("method", r.Method),
				attribute.String("path", r.URL.Path),
			)

			ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			r = r.WithContext(ctx)

			next.ServeHTTP(ww, r)

			duration := time.Since(start)
			statusCode := ww.statusCode

			span.SetAttributes(
				semconv.HTTPStatusCodeKey.Int(statusCode),
				attribute.Int64("http.duration_ms", duration.Milliseconds()),
			)

			if statusCode >= 400 {
				span.SetStatus(codes.Error, "HTTP "+strconv.Itoa(statusCode))
			} else {
				span.SetStatus(codes.Ok, "")
			}

			responseSizeBytes := int64(ww.responseSize)
			if responseSizeBytes > 0 {
				responseSize.Record(ctx, float64(responseSizeBytes),
					attribute.String("method", r.Method),
					attribute.String("path", r.URL.Path),
					attribute.Int("status", statusCode),
				)
			}

			requestCounter.Increment(ctx,
				attribute.String("method", r.Method),
				attribute.String("path", r.URL.Path),
				attribute.Int("status", statusCode),
			)

			requestDuration.RecordDuration(ctx, duration,
				attribute.String("method", r.Method),
				attribute.String("path", r.URL.Path),
				attribute.Int("status", statusCode),
			)

			ctxLogger.Info("Requisição finalizada",
				attribute.String("method", r.Method),
				attribute.String("path", r.URL.Path),
				attribute.Int("status", statusCode),
				attribute.Int64("duration_ms", duration.Milliseconds()),
			)

			span.End()
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	responseSize int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.responseSize += len(b)
	return rw.ResponseWriter.Write(b)
}

func GinMiddleware(client Client, config MiddlewareConfig) gin.HandlerFunc {
	tracing := client.NewTracingHelper(config.ServiceName)
	metrics := client.NewMetricsHelper(config.ServiceName + "/http")
	logs := client.NewLogsHelper(config.ServiceName + "/http")

	requestCounter, _ := metrics.NewCounter("http_requests_total", "Total de requisições HTTP")
	requestDuration, _ := metrics.NewHistogram("http_request_duration_seconds", "Duração das requisições HTTP")
	requestSize, _ := metrics.NewHistogram("http_request_size_bytes", "Tamanho das requisições HTTP")
	responseSize, _ := metrics.NewHistogram("http_response_size_bytes", "Tamanho das respostas HTTP")

	return func(c *gin.Context) {
		if shouldSkip(c.Request.URL.Path, config.SkipPaths) {
			c.Next()
			return
		}

		start := time.Now()
		ctx := c.Request.Context()

		ctx, span := tracing.StartSpan(ctx, "http.request",
			trace.WithAttributes(
				semconv.HTTPMethodKey.String(c.Request.Method),
				semconv.HTTPURLKey.String(c.Request.URL.String()),
				semconv.HTTPRouteKey.String(c.FullPath()),
				attribute.String("http.user_agent", c.Request.UserAgent()),
				attribute.String("http.client_ip", c.ClientIP()),
			),
		)

		ctx = WithTags(ctx,
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.path", c.FullPath()),
			attribute.String("http.user_agent", c.Request.UserAgent()),
			attribute.String("http.client_ip", c.ClientIP()),
		)

		requestSizeBytes := c.Request.ContentLength
		if requestSizeBytes > 0 {
			requestSize.Record(ctx, float64(requestSizeBytes),
				attribute.String("method", c.Request.Method),
				attribute.String("path", c.FullPath()),
			)
		}

		ctxLogger := NewContextLogger(logs, ctx)
		ctxLogger.Info("Requisição recebida",
			attribute.String("method", c.Request.Method),
			attribute.String("path", c.FullPath()),
		)

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()

		span.SetAttributes(
			semconv.HTTPStatusCodeKey.Int(statusCode),
			attribute.Int64("http.duration_ms", duration.Milliseconds()),
		)

		if statusCode >= 400 {
			span.SetStatus(codes.Error, "HTTP "+strconv.Itoa(statusCode))
		} else {
			span.SetStatus(codes.Ok, "")
		}

		responseSizeBytes := int64(c.Writer.Size())
		if responseSizeBytes > 0 {
			responseSize.Record(ctx, float64(responseSizeBytes),
				attribute.String("method", c.Request.Method),
				attribute.String("path", c.FullPath()),
				attribute.Int("status", statusCode),
			)
		}

		requestCounter.Increment(ctx,
			attribute.String("method", c.Request.Method),
			attribute.String("path", c.FullPath()),
			attribute.Int("status", statusCode),
		)

		requestDuration.RecordDuration(ctx, duration,
			attribute.String("method", c.Request.Method),
			attribute.String("path", c.FullPath()),
			attribute.Int("status", statusCode),
		)

		ctxLogger.Info("Requisição finalizada",
			attribute.String("method", c.Request.Method),
			attribute.String("path", c.FullPath()),
			attribute.Int("status", statusCode),
			attribute.Int64("duration_ms", duration.Milliseconds()),
		)

		span.End()
	}
}

func EchoMiddleware(client Client, config MiddlewareConfig) echo.MiddlewareFunc {
	tracing := client.NewTracingHelper(config.ServiceName)
	metrics := client.NewMetricsHelper(config.ServiceName + "/http")
	logs := client.NewLogsHelper(config.ServiceName + "/http")

	requestCounter, _ := metrics.NewCounter("http_requests_total", "Total de requisições HTTP")
	requestDuration, _ := metrics.NewHistogram("http_request_duration_seconds", "Duração das requisições HTTP")
	requestSize, _ := metrics.NewHistogram("http_request_size_bytes", "Tamanho das requisições HTTP")
	responseSize, _ := metrics.NewHistogram("http_response_size_bytes", "Tamanho das respostas HTTP")

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if shouldSkip(c.Path(), config.SkipPaths) {
				return next(c)
			}

			start := time.Now()
			ctx := c.Request().Context()

			ctx, span := tracing.StartSpan(ctx, "http.request",
				trace.WithAttributes(
					semconv.HTTPMethodKey.String(c.Request().Method),
					semconv.HTTPURLKey.String(c.Request().URL.String()),
					semconv.HTTPRouteKey.String(c.Path()),
					attribute.String("http.user_agent", c.Request().UserAgent()),
					attribute.String("http.client_ip", c.RealIP()),
				),
			)

			ctx = WithTags(ctx,
				attribute.String("http.method", c.Request().Method),
				attribute.String("http.path", c.Path()),
				attribute.String("http.user_agent", c.Request().UserAgent()),
				attribute.String("http.client_ip", c.RealIP()),
			)

			requestSizeBytes := c.Request().ContentLength
			if requestSizeBytes > 0 {
				requestSize.Record(ctx, float64(requestSizeBytes),
					attribute.String("method", c.Request().Method),
					attribute.String("path", c.Path()),
				)
			}

			ctxLogger := NewContextLogger(logs, ctx)
			ctxLogger.Info("Requisição recebida",
				attribute.String("method", c.Request().Method),
				attribute.String("path", c.Path()),
			)

			c.SetRequest(c.Request().WithContext(ctx))

			err := next(c)

			duration := time.Since(start)
			statusCode := c.Response().Status

			span.SetAttributes(
				semconv.HTTPStatusCodeKey.Int(statusCode),
				attribute.Int64("http.duration_ms", duration.Milliseconds()),
			)

			if statusCode >= 400 || err != nil {
				span.SetStatus(codes.Error, "HTTP "+strconv.Itoa(statusCode))
				if err != nil {
					span.RecordError(err)
				}
			} else {
				span.SetStatus(codes.Ok, "")
			}

			responseSizeBytes := int64(c.Response().Size)
			if responseSizeBytes > 0 {
				responseSize.Record(ctx, float64(responseSizeBytes),
					attribute.String("method", c.Request().Method),
					attribute.String("path", c.Path()),
					attribute.Int("status", statusCode),
				)
			}

			requestCounter.Increment(ctx,
				attribute.String("method", c.Request().Method),
				attribute.String("path", c.Path()),
				attribute.Int("status", statusCode),
			)

			requestDuration.RecordDuration(ctx, duration,
				attribute.String("method", c.Request().Method),
				attribute.String("path", c.Path()),
				attribute.Int("status", statusCode),
			)

			ctxLogger.Info("Requisição finalizada",
				attribute.String("method", c.Request().Method),
				attribute.String("path", c.Path()),
				attribute.Int("status", statusCode),
				attribute.Int64("duration_ms", duration.Milliseconds()),
			)

			span.End()
			return err
		}
	}
}

func ChiMiddleware(client Client, config MiddlewareConfig) func(http.Handler) http.Handler {
	return HTTPMiddleware(client, config)
}
