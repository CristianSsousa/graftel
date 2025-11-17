package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/CristianSsousa/graftel"
	"go.opentelemetry.io/otel/attribute"
)

func main() {
	// Obter chave de API do Grafana Cloud (configure via variável de ambiente)
	apiKey := os.Getenv("GRAFANA_CLOUD_API_KEY")
	if apiKey == "" {
		log.Fatal("GRAFANA_CLOUD_API_KEY não configurada. Configure a variável de ambiente com sua chave de API do Grafana Cloud.")
	}

	// Obter endpoint OTLP do Grafana Cloud (configure via variável de ambiente)
	otlpEndpoint := os.Getenv("OTLP_ENDPOINT")
	if otlpEndpoint == "" {
		// Endpoint padrão do Grafana Cloud (ajuste conforme sua região)
		otlpEndpoint = "https://otlp-gateway-prod-us-central-0.grafana.net/otlp"
	}

	// Obter Instance ID do Grafana Cloud (opcional, mas recomendado)
	instanceID := os.Getenv("GRAFANA_CLOUD_INSTANCE_ID")

	// Configurar OpenTelemetry para Grafana Cloud usando o pattern de builder
	config := graftel.NewConfig("meu-servico-grafana").
		WithServiceVersion("1.0.0").
		WithOTLPEndpoint(otlpEndpoint).
		WithGrafanaCloudAPIKey(apiKey).
		WithInsecure(false) // Grafana Cloud usa HTTPS

	// Adicionar Instance ID se fornecido
	if instanceID != "" {
		config = config.WithGrafanaCloudInstanceID(instanceID)
	}

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
	logs.Info(ctx, "Servidor iniciado e conectado ao Grafana Cloud",
		attribute.String("service", config.ServiceName),
		attribute.String("version", config.ServiceVersion),
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

	logs.Info(ctx, "Simulação concluída. Verifique o Grafana Cloud para ver as métricas e logs.")

	fmt.Println("✅ Exemplo concluído!")
	fmt.Println("📊 Verifique o Grafana Cloud para ver as métricas e logs.")
	fmt.Println("🔗 Acesse: https://grafana.com/orgs/<seu-org>/")
}
