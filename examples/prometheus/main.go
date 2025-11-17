package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/CristianSsousa/graftel"
	"go.opentelemetry.io/otel/attribute"
)

func main() {
	// Configurar OpenTelemetry com Prometheus usando o pattern de builder
	config := graftel.NewConfig("meu-servico-prometheus").
		WithServiceVersion("1.0.0").
		WithPrometheusEndpoint(":8080"). // Expor métricas em http://localhost:8080/metrics
		WithResourceAttribute("environment", "production")

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
	activeConnections, err := metrics.NewUpDownCounter(
		"active_connections",
		"Número de conexões ativas",
	)
	if err != nil {
		log.Fatalf("Falha ao criar up-down counter: %v", err)
	}

	requestCounter, err := metrics.NewCounter(
		"http_requests_total",
		"Total de requisições HTTP",
	)
	if err != nil {
		log.Fatalf("Falha ao criar contador: %v", err)
	}

	// Expor endpoint Prometheus
	exporter := client.GetPrometheusExporter()
	if exporter != nil {
		http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
			// O exporter do Prometheus precisa ser usado com promhttp
			// Por enquanto, vamos apenas informar que está configurado
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("# Métricas Prometheus configuradas\n# Use o exporter diretamente com promhttp\n"))
		})
		log.Println("Métricas Prometheus disponíveis em http://localhost:8080/metrics")
	} else {
		log.Println("Exporter Prometheus não configurado")
		return
	}

	// Simular conexões
	go func() {
		for i := 0; i < 5; i++ {
			activeConnections.Increment(ctx,
				attribute.String("type", "websocket"),
			)
			time.Sleep(2 * time.Second)
		}
	}()

	// Simular requisições HTTP
	go func() {
		for i := 0; i < 20; i++ {
			requestCounter.Increment(ctx,
				attribute.String("method", "GET"),
				attribute.String("endpoint", "/api/data"),
				attribute.Int("status", 200),
			)
			time.Sleep(500 * time.Millisecond)
		}
	}()

	// Registrar logs
	logs.Info(ctx, "Servidor iniciado",
		attribute.String("port", "8080"),
		attribute.String("environment", "production"),
	)

	// Iniciar servidor HTTP
	fmt.Println("Servidor rodando em http://localhost:8080")
	fmt.Println("Acesse http://localhost:8080/metrics para ver as métricas Prometheus")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
