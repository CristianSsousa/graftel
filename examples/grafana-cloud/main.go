// Smoke test da lib graftel contra o Grafana Cloud.
// Envia métricas, logs e traces via OTLP HTTP e aguarda o flush completo.
//
// Como obter as credenciais no Grafana Cloud:
//  1. Acesse grafana.com → faça login → clique no seu stack
//  2. Vá em "Connections" → "Open Telemetry" (ou busque por "OTLP")
//  3. Copie:
//     - OTLP HTTP Endpoint  → GRAFTEL_OTLP_ENDPOINT
//     - Instance ID (número) → GRAFTEL_INSTANCE_ID
//     - API Token com scopes MetricsPublisher + LogsPublisher + TracesPublisher → GRAFTEL_API_KEY
//
// Execução:
//
//	export GRAFTEL_OTLP_ENDPOINT="https://otlp-gateway-prod-us-central-0.grafana.net/otlp"
//	export GRAFTEL_INSTANCE_ID="123456"
//	export GRAFTEL_API_KEY="glc_eyJ..."
//	go run ./examples/grafana-cloud/main.go
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	graftel "github.com/CristianSsousa/graftel/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func main() {
	config := graftel.NewConfig("graftel-smoke-test").
		WithServiceVersion("2.0.0").
		WithResourceAttributes(map[string]string{
			"environment": "local",
			"team":        "backend",
		}).
		WithMetricExportInterval(15 * time.Second).
		WithLogExportInterval(5 * time.Second).
		WithExportTimeout(15 * time.Second)
	// Endpoint, InstanceID e APIKey são lidos das variáveis de ambiente:
	//   GRAFTEL_OTLP_ENDPOINT, GRAFTEL_INSTANCE_ID, GRAFTEL_API_KEY

	client, err := graftel.NewClient(config)
	if err != nil {
		log.Fatalf("erro ao criar cliente: %v", err)
	}

	ctx := context.Background()
	if err := client.Initialize(ctx); err != nil {
		log.Fatalf("erro ao inicializar: %v", err)
	}
	defer func() {
		fmt.Println("→ Enviando dados restantes (flush)…")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		if err := client.Shutdown(shutdownCtx); err != nil {
			log.Printf("aviso no shutdown: %v", err)
		}
		fmt.Println("→ Flush concluído.")
	}()

	metrics := client.NewMetricsHelper("graftel-smoke-test/metrics")
	logs := client.NewLogsHelper("graftel-smoke-test/logs")
	tracing := client.NewTracingHelper("graftel-smoke-test")

	// ── Instrumentos de métricas ──────────────────────────────────────────────
	requestCounter, _ := metrics.NewCounter(
		"smoke_requests_total",
		"Requisições simuladas pelo smoke test",
	)
	requestDuration, _ := metrics.NewHistogram(
		"smoke_request_duration_seconds",
		"Duração simulada das requisições",
	)
	errorCounter, _ := metrics.NewCounter(
		"smoke_errors_total",
		"Erros simulados pelo smoke test",
	)

	// ── Início ────────────────────────────────────────────────────────────────
	logs.Info(ctx, "smoke test iniciado",
		attribute.String("service", "graftel-smoke-test"),
		attribute.String("version", "2.0.0"),
	)

	// ── Simula 10 requisições com traces + métricas + logs ───────────────────
	for i := range 10 {
		reqCtx, rootSpan := tracing.StartSpanWithTags(ctx, "handle_request",
			attribute.Int("request.index", i),
			attribute.String("http.method", "GET"),
			attribute.String("http.route", "/api/items"),
		)

		start := time.Now()
		statusCode, reqErr := simulateRequest(reqCtx, tracing, logs, i)
		duration := time.Since(start)

		requestCounter.Increment(reqCtx,
			attribute.String("method", "GET"),
			attribute.String("path", "/api/items"),
			attribute.Int("status", statusCode),
		)
		requestDuration.RecordDuration(reqCtx, duration,
			attribute.String("path", "/api/items"),
		)

		if reqErr != nil {
			errorCounter.Increment(reqCtx, attribute.String("type", "simulated"))
			rootSpan.RecordError(reqErr)
			rootSpan.SetStatus(codes.Error, reqErr.Error())
			logs.Error(reqCtx, "requisição falhou",
				attribute.Int("request.index", i),
				attribute.String("error", reqErr.Error()),
			)
		} else {
			rootSpan.SetStatus(codes.Ok, "")
			logs.Info(reqCtx, "requisição concluída",
				attribute.Int("request.index", i),
				attribute.Int("status", statusCode),
				attribute.Int64("duration_ms", duration.Milliseconds()),
			)
		}

		rootSpan.End()
		fmt.Printf("  [%02d] status=%d  duration=%v\n", i, statusCode, duration.Round(time.Millisecond))
		time.Sleep(200 * time.Millisecond)
	}

	logs.Info(ctx, "smoke test concluído — aguardando flush para o Grafana Cloud")
	fmt.Println("\nDados gerados. Aguardando flush…")
	fmt.Println("Acesse seu Grafana Cloud para verificar:")
	fmt.Println("  • Explore → Tempo   → service.name = graftel-smoke-test  (traces)")
	fmt.Println("  • Explore → Loki    → {service_name=\"graftel-smoke-test\"} (logs)")
	fmt.Println("  • Explore → Metrics → smoke_requests_total                (métricas)")
}

// simulateRequest simula o processamento de uma requisição com spans filhos.
func simulateRequest(
	ctx context.Context,
	tracing graftel.TracingHelper,
	logs graftel.LogsHelper,
	index int,
) (int, error) {
	// Span de validação
	ctx, validateSpan := tracing.StartSpanWithTags(ctx, "validate_request",
		attribute.Int("request.index", index),
	)
	time.Sleep(time.Duration(rand.Intn(20)+5) * time.Millisecond)
	validateSpan.SetStatus(codes.Ok, "")
	validateSpan.End()

	// Span de busca no "banco"
	dbErr := tracing.WithSpan(ctx, "db.query", func(dbCtx context.Context) error {
		time.Sleep(time.Duration(rand.Intn(40)+10) * time.Millisecond)
		if index == 7 {
			return errors.New("query timeout: conexão encerrada pelo servidor")
		}
		logs.Debug(dbCtx, "query executada",
			attribute.String("table", "items"),
			attribute.Int("rows", rand.Intn(50)+1),
		)
		return nil
	}, attribute.String("db.system", "postgresql"))

	if dbErr != nil {
		return 500, dbErr
	}

	if index == 3 {
		logs.Warn(ctx, "item não encontrado, retornando 404",
			attribute.Int("request.index", index),
		)
		return 404, nil
	}

	return 200, nil
}
