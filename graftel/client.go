// Package graftel fornece uma interface simplificada para trabalhar com OpenTelemetry
// em aplicações Go, focando em métricas e logs.
//
// Exemplo básico:
//
//	config := graftel.NewConfig("meu-servico").
//		WithServiceVersion("1.0.0").
//		WithOTLPEndpoint("http://localhost:4318").
//		WithInsecure(true)
//
//	client, err := graftel.NewClient(config)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	ctx := context.Background()
//	if err := client.Initialize(ctx); err != nil {
//		log.Fatal(err)
//	}
//	defer client.Shutdown(ctx)
package graftel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	otellog "go.opentelemetry.io/otel/log"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// Client gerencia a inicialização e uso do OpenTelemetry.
// É a interface principal para trabalhar com métricas e logs.
type Client interface {
	// Initialize inicializa o OpenTelemetry com métricas e logs.
	// Deve ser chamado antes de usar qualquer funcionalidade.
	Initialize(ctx context.Context) error

	// Shutdown encerra o cliente OpenTelemetry de forma segura.
	// Deve ser chamado ao finalizar a aplicação.
	Shutdown(ctx context.Context) error

	// GetMeter retorna um Meter para criar métricas.
	GetMeter(name string, opts ...otelmetric.MeterOption) otelmetric.Meter

	// GetLogger retorna um Logger para criar logs.
	GetLogger(name string) otellog.Logger

	// GetPrometheusExporter retorna o exporter Prometheus, se configurado.
	// Retorna nil se Prometheus não estiver habilitado.
	GetPrometheusExporter() *prometheus.Exporter

	// NewMetricsHelper cria um helper para facilitar o uso de métricas.
	NewMetricsHelper(name string) MetricsHelper

	// NewLogsHelper cria um helper para facilitar o uso de logs.
	NewLogsHelper(name string) LogsHelper
}

// client é a implementação concreta do Client.
type client struct {
	config             Config
	meterProvider      *sdkmetric.MeterProvider
	loggerProvider     *log.LoggerProvider
	prometheusExporter *prometheus.Exporter
	resource           *resource.Resource
}

// NewClient cria uma nova instância do cliente OpenTelemetry.
// A configuração é validada antes de criar o cliente.
func NewClient(config Config) (Client, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuração inválida: %w", err)
	}

	// Criar resource
	res, err := createResource(config)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar resource: %w", err)
	}

	return &client{
		config:   config,
		resource: res,
	}, nil
}

// Initialize inicializa o OpenTelemetry com métricas e logs.
func (c *client) Initialize(ctx context.Context) error {
	// Inicializar métricas
	if err := c.initializeMetrics(ctx); err != nil {
		return fmt.Errorf("falha ao inicializar métricas: %w", err)
	}

	// Inicializar logs
	if err := c.initializeLogs(ctx); err != nil {
		return fmt.Errorf("falha ao inicializar logs: %w", err)
	}

	return nil
}

// initializeMetrics configura o provider de métricas.
func (c *client) initializeMetrics(ctx context.Context) error {
	var reader sdkmetric.Reader

	// Se PrometheusEndpoint estiver configurado, criar exporter Prometheus
	if c.config.PrometheusEndpoint != "" {
		exporter, err := prometheus.New()
		if err != nil {
			return fmt.Errorf("falha ao criar exporter Prometheus: %w", err)
		}
		c.prometheusExporter = exporter
		reader = exporter
	} else {
		// Caso contrário, usar OTLP HTTP
		opts := []otlpmetrichttp.Option{
			otlpmetrichttp.WithEndpoint(c.config.OTLPEndpoint),
		}

		// Configurar TLS
		if c.config.Insecure {
			opts = append(opts, otlpmetrichttp.WithInsecure())
		}

		// Configurar autenticação para Grafana Cloud
		if c.config.GrafanaCloudAPIKey != "" {
			opts = append(opts, otlpmetrichttp.WithHeaders(map[string]string{
				"Authorization": "Basic " + c.config.GrafanaCloudAPIKey,
			}))
		}

		exporter, err := otlpmetrichttp.New(ctx, opts...)
		if err != nil {
			return fmt.Errorf("falha ao criar exporter OTLP: %w", err)
		}

		reader = sdkmetric.NewPeriodicReader(exporter,
			sdkmetric.WithInterval(c.config.MetricExportInterval),
		)
	}

	// Criar MeterProvider
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(c.resource),
		sdkmetric.WithReader(reader),
	)

	c.meterProvider = meterProvider
	otel.SetMeterProvider(meterProvider)

	return nil
}

// initializeLogs configura o provider de logs.
func (c *client) initializeLogs(ctx context.Context) error {
	opts := []otlploghttp.Option{
		otlploghttp.WithEndpoint(c.config.OTLPEndpoint),
	}

	// Configurar TLS
	if c.config.Insecure {
		opts = append(opts, otlploghttp.WithInsecure())
	}

	// Configurar autenticação para Grafana Cloud
	if c.config.GrafanaCloudAPIKey != "" {
		opts = append(opts, otlploghttp.WithHeaders(map[string]string{
			"Authorization": "Basic " + c.config.GrafanaCloudAPIKey,
		}))
	}

	exporter, err := otlploghttp.New(ctx, opts...)
	if err != nil {
		return fmt.Errorf("falha ao criar exporter de logs OTLP: %w", err)
	}

	// Criar LoggerProvider
	loggerProvider := log.NewLoggerProvider(
		log.WithResource(c.resource),
		log.WithProcessor(log.NewBatchProcessor(exporter)),
	)

	c.loggerProvider = loggerProvider

	return nil
}

// GetMeter retorna um Meter para criar métricas.
func (c *client) GetMeter(name string, opts ...otelmetric.MeterOption) otelmetric.Meter {
	if c.meterProvider == nil {
		// Retornar meter do provider global se ainda não inicializado
		return otel.Meter(name, opts...)
	}
	return c.meterProvider.Meter(name, opts...)
}

// GetLogger retorna um Logger para criar logs.
func (c *client) GetLogger(name string) otellog.Logger {
	if c.loggerProvider == nil {
		// Retornar um logger básico se ainda não inicializado
		// Isso não deve acontecer se Initialize() foi chamado corretamente
		// Criar um logger provider temporário com resource básico
		basicResource, _ := resource.New(context.Background())
		tempProvider := log.NewLoggerProvider(
			log.WithResource(basicResource),
		)
		return tempProvider.Logger(name)
	}
	return c.loggerProvider.Logger(name)
}

// GetPrometheusExporter retorna o exporter Prometheus, se configurado.
func (c *client) GetPrometheusExporter() *prometheus.Exporter {
	return c.prometheusExporter
}

// NewMetricsHelper cria um helper para facilitar o uso de métricas.
func (c *client) NewMetricsHelper(name string) MetricsHelper {
	return NewMetricsHelper(c.GetMeter(name))
}

// NewLogsHelper cria um helper para facilitar o uso de logs.
func (c *client) NewLogsHelper(name string) LogsHelper {
	return NewLogsHelper(c.GetLogger(name))
}

// Shutdown encerra o cliente OpenTelemetry de forma segura.
func (c *client) Shutdown(ctx context.Context) error {
	var errs []error

	if c.meterProvider != nil {
		if err := c.meterProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("erro ao encerrar meter provider: %w", err))
		}
	}

	if c.loggerProvider != nil {
		if err := c.loggerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("erro ao encerrar logger provider: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("erros ao encerrar: %v", errs)
	}

	return nil
}

// createResource cria o resource OpenTelemetry.
func createResource(config Config) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{
		semconv.ServiceNameKey.String(config.ServiceName),
	}

	if config.ServiceVersion != "" {
		attrs = append(attrs, semconv.ServiceVersionKey.String(config.ServiceVersion))
	}

	// Adicionar atributos customizados
	for k, v := range config.ResourceAttributes {
		attrs = append(attrs, attribute.String(k, v))
	}

	return resource.New(context.Background(),
		resource.WithAttributes(attrs...),
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithContainer(),
		resource.WithHost(),
	)
}

