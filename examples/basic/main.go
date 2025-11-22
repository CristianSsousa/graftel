package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/CristianSsousa/graftel/v2"
	"go.opentelemetry.io/otel/attribute"
)

func main() {
	// Configurar OpenTelemetry usando o pattern de builder
	config := graftel.NewConfig("meu-test-log").
		WithServiceVersion("1.0.0").
		WithOTLPEndpoint("http://localhost:4318").
		WithInsecure(true). // Para desenvolvimento local
		WithResourceAttribute("environment", "development")

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
		"requests_total",
		"Total de requisições recebidas",
	)
	if err != nil {
		log.Fatalf("Falha ao criar contador: %v", err)
	}

	requestDuration, err := metrics.NewHistogram(
		"request_duration_seconds",
		"Duração das requisições em segundos",
	)
	if err != nil {
		log.Fatalf("Falha ao criar histograma: %v", err)
	}

	// Simular algumas requisições
	for i := 0; i < 10; i++ {
		start := time.Now()

		// Simular processamento
		time.Sleep(100 * time.Millisecond)

		duration := time.Since(start)

		// Registrar métricas
		requestCounter.Increment(ctx,
			attribute.String("method", "GET"),
			attribute.String("path", "/api/users"),
			attribute.Int("status", 200),
		)

		requestDuration.RecordDuration(ctx, duration,
			attribute.String("method", "GET"),
			attribute.String("path", "/api/users"),
		)

		// Registrar log
		logs.Info(ctx, fmt.Sprintf("Requisição processada em %v", duration),
			attribute.String("method", "GET"),
			attribute.String("path", "/api/users"),
			attribute.Int("status", 200),
		)
	}

	fmt.Println("Exemplo concluído! Verifique o sistema de observabilidade para ver as métricas e logs.")
}
