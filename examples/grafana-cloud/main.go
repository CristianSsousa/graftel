package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/CristianSsousa/graftel"
	"go.opentelemetry.io/otel/attribute"
)

func main() {
	// As configurações podem ser fornecidas via variáveis de ambiente GRAFTEL_*
	// ou explicitamente via métodos With*. A ordem de prioridade é:
	// 1. Valores passados via With* (maior prioridade)
	// 2. Variáveis de ambiente GRAFTEL_*
	// 3. Valores padrão

	// Exemplo: usando variáveis de ambiente (recomendado)
	// Configure: GRAFTEL_SERVICE_NAME, GRAFTEL_OTLP_ENDPOINT, GRAFTEL_API_KEY, etc.
	config := graftel.NewConfig("meu-servico").
		WithServiceVersion("1.0.0").
		WithInsecure(false) // HTTPS por padrão

	// Ou fornecer explicitamente (sobrescreve ENV se existir)
	// config := graftel.NewConfig("meu-servico").
	// 	WithServiceVersion("1.0.0").
	// 	WithOTLPEndpoint("https://otlp-gateway-prod-us-central-0.grafana.net/otlp").
	// 	WithAPIKey(os.Getenv("GRAFTEL_API_KEY")).
	// 	WithInstanceID(os.Getenv("GRAFTEL_INSTANCE_ID")).
	// 	WithInsecure(false)

	config = config.
		WithResourceAttributes(map[string]string{
			"environment": "production",
			"team":        "backend",
		}).
		WithMetricExportInterval(30 * time.Second).
		WithLogExportInterval(30 * time.Second)

	client, err := graftel.NewClient(config)
	if err != nil {
		log.Fatalf("Falha ao criar cliente OpenTelemetry: %v", err)
	}

	ctx := context.Background()

	// Inicializar OpenTelemetry
	if err := client.Initialize(ctx); err != nil {
		log.Fatalf("Falha ao inicializar OpenTelemetry: %v", err)
	}
	defer client.Shutdown(ctx)

	// Criar helpers
	metrics := client.NewMetricsHelper("meu-servico/metrics")
	logs := client.NewLogsHelper("meu-servico/logs")

	// Criar métricas
	requestCounter, err := metrics.NewCounter(
		"http_requests_total",
		"Total de requisições HTTP recebidas",
	)
	if err != nil {
		log.Fatalf("Falha ao criar contador: %v", err)
	}

	requestDuration, err := metrics.NewHistogram(
		"http_request_duration_seconds",
		"Duração das requisições HTTP em segundos",
	)
	if err != nil {
		log.Fatalf("Falha ao criar histograma: %v", err)
	}

	activeConnections, err := metrics.NewUpDownCounter(
		"active_connections",
		"Número de conexões ativas",
	)
	if err != nil {
		log.Fatalf("Falha ao criar up-down counter: %v", err)
	}

	// Log inicial
	logs.Info(ctx, "Servidor iniciado e conectado",
		attribute.String("service", "meu-servico"),
		attribute.String("version", "1.0.0"),
	)

	// Simular atividade do servidor
	for i := 0; i < 20; i++ {
		start := time.Now()

		// Simular processamento
		time.Sleep(50 * time.Millisecond)

		duration := time.Since(start)

		// Incrementar conexões ativas
		activeConnections.Increment(ctx,
			attribute.String("type", "http"),
		)

		// Registrar métricas
		requestCounter.Increment(ctx,
			attribute.String("method", "GET"),
			attribute.String("path", "/api/data"),
			attribute.Int("status", 200),
		)

		requestDuration.RecordDuration(ctx, duration,
			attribute.String("method", "GET"),
			attribute.String("path", "/api/data"),
		)

		// Registrar log
		logs.Info(ctx, fmt.Sprintf("Requisição processada em %v", duration),
			attribute.String("method", "GET"),
			attribute.String("path", "/api/data"),
			attribute.Int("status", 200),
			attribute.Int("request_id", i),
		)

		// Decrementar conexões após um tempo
		if i%5 == 0 {
			activeConnections.Decrement(ctx,
				attribute.String("type", "http"),
			)
		}
	}

	logs.Info(ctx, "Simulação concluída. Verifique o sistema de observabilidade para ver as métricas e logs.")

	fmt.Println("✅ Exemplo concluído!")
	fmt.Println("📊 Verifique o sistema de observabilidade para ver as métricas e logs.")
}
