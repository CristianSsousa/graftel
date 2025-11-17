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
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
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
		// Parse da URL para extrair endpoint e path
		endpoint, urlPath, err := parseOTLPEndpoint(c.config.OTLPEndpoint)
		if err != nil {
			return fmt.Errorf("falha ao processar endpoint OTLP: %w", err)
		}

		opts := []otlpmetrichttp.Option{
			otlpmetrichttp.WithEndpoint(endpoint),
		}

		// Configurar path
		// Para Grafana Cloud com /otlp, usar /otlp/v1/metrics
		// Para outros casos, usar o path fornecido ou deixar padrão
		if urlPath == "/otlp" {
			opts = append(opts, otlpmetrichttp.WithURLPath("/otlp/v1/metrics"))
		} else if urlPath != "" && urlPath != "/" {
			opts = append(opts, otlpmetrichttp.WithURLPath(urlPath))
		}

		// Configurar TLS
		if c.config.Insecure {
			opts = append(opts, otlpmetrichttp.WithInsecure())
		}

		// Configurar timeout
		opts = append(opts, otlpmetrichttp.WithTimeout(c.config.ExportTimeout))

		// Configurar autenticação para Grafana Cloud
		if c.config.GrafanaCloudAPIKey != "" {
			authHeader := buildGrafanaCloudAuthHeader(c.config.GrafanaCloudInstanceID, c.config.GrafanaCloudAPIKey)
			opts = append(opts, otlpmetrichttp.WithHeaders(map[string]string{
				"Authorization": authHeader,
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
	// Parse da URL para extrair endpoint e path
	endpoint, urlPath, err := parseOTLPEndpoint(c.config.OTLPEndpoint)
	if err != nil {
		return fmt.Errorf("falha ao processar endpoint OTLP: %w", err)
	}

	opts := []otlploghttp.Option{
		otlploghttp.WithEndpoint(endpoint),
	}

	// Configurar path
	// Para Grafana Cloud com /otlp, usar /otlp/v1/logs
	// Para outros casos, usar o path fornecido ou deixar padrão
	if urlPath == "/otlp" {
		opts = append(opts, otlploghttp.WithURLPath("/otlp/v1/logs"))
	} else if urlPath != "" && urlPath != "/" {
		opts = append(opts, otlploghttp.WithURLPath(urlPath))
	}

	// Configurar TLS
	if c.config.Insecure {
		opts = append(opts, otlploghttp.WithInsecure())
	}

	// Configurar timeout
	opts = append(opts, otlploghttp.WithTimeout(c.config.ExportTimeout))

	// Configurar autenticação para Grafana Cloud
	if c.config.GrafanaCloudAPIKey != "" {
		authHeader := buildGrafanaCloudAuthHeader(c.config.GrafanaCloudInstanceID, c.config.GrafanaCloudAPIKey)
		opts = append(opts, otlploghttp.WithHeaders(map[string]string{
			"Authorization": authHeader,
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

// buildGrafanaCloudAuthHeader constrói o header de autenticação para Grafana Cloud.
// O formato é: Basic <base64(instance_id:api_key)>
func buildGrafanaCloudAuthHeader(instanceID, apiKey string) string {
	if instanceID != "" {
		// Formato: instance_id:api_key codificado em base64
		credentials := instanceID + ":" + apiKey
		encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
		return "Basic " + encoded
	}
	// Se não tiver instance ID, usar apenas a API key (formato alternativo)
	return "Basic " + apiKey
}

// parseOTLPEndpoint extrai o host:port e o path de uma URL OTLP.
// Retorna o endpoint (host:port) e o path (se houver).
// Para Grafana Cloud, o path /otlp é removido pois o OpenTelemetry adiciona /v1/metrics ou /v1/logs automaticamente.
func parseOTLPEndpoint(endpointURL string) (endpoint, urlPath string, err error) {
	// Se não começar com http:// ou https://, assumir que é apenas host:port
	if !strings.HasPrefix(endpointURL, "http://") && !strings.HasPrefix(endpointURL, "https://") {
		// Se contém /, pode ser host:port/path
		if idx := strings.Index(endpointURL, "/"); idx != -1 {
			return endpointURL[:idx], endpointURL[idx:], nil
		}
		return endpointURL, "", nil
	}

	// Parse da URL completa
	parsedURL, err := url.Parse(endpointURL)
	if err != nil {
		return "", "", fmt.Errorf("falha ao fazer parse da URL: %w", err)
	}

	// Montar endpoint como host:port (o OpenTelemetry lida com portas padrão automaticamente)
	endpoint = parsedURL.Host

	// Extrair path
	urlPath = parsedURL.Path

	// Para Grafana Cloud com path /otlp, manter o path pois o OpenTelemetry
	// adiciona automaticamente /v1/metrics ou /v1/logs ao path base
	// O endpoint final será: https://otlp-gateway-prod-sa-east-1.grafana.net/otlp/v1/metrics
	// Se não houver path, usar path padrão
	if urlPath == "" {
		urlPath = "/"
	}

	return endpoint, urlPath, nil
}

// createResource cria o resource OpenTelemetry.
func createResource(config Config) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{
		semconv.ServiceNameKey.String(config.ServiceName),
	}

	if config.ServiceVersion != "" {
		attrs = append(attrs, semconv.ServiceVersionKey.String(config.ServiceVersion))
	}

	// Adicionar Instance ID se fornecido
	if config.GrafanaCloudInstanceID != "" {
		attrs = append(attrs, semconv.ServiceInstanceIDKey.String(config.GrafanaCloudInstanceID))
	}

	// Adicionar atributos customizados
	for k, v := range config.ResourceAttributes {
		attrs = append(attrs, attribute.String(k, v))
	}

	// Criar resource com todos os atributos automáticos
	res, err := resource.New(context.Background(),
		resource.WithAttributes(attrs...),
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithContainer(),
		resource.WithHost(),
	)
	if err != nil {
		return nil, err
	}

	// Sanitizar o resource removendo campos desnecessários
	return sanitizeResource(res), nil
}

// sanitizeResource remove campos desnecessários ou sensíveis do Resource.
// Remove: process.command_args, process.executable.path, process.executable.name, process.owner
func sanitizeResource(res *resource.Resource) *resource.Resource {
	// Lista de chaves a serem removidas
	keysToRemove := map[string]bool{
		"process.command_args":    true, // Argumentos de linha de comando
		"process.executable.path": true, // Caminho completo do executável
		"process.executable.name": true, // Nome do executável
		"process.command":         true, // Comando completo
		"process.owner":           true, // Proprietário do processo (pode ser sensível)
	}

	// Obter todos os atributos do resource
	attrs := res.Attributes()
	filteredAttrs := make([]attribute.KeyValue, 0, len(attrs))

	// Filtrar atributos
	for _, attr := range attrs {
		keyStr := string(attr.Key)
		// Manter apenas atributos que não estão na lista de remoção
		if !keysToRemove[keyStr] {
			filteredAttrs = append(filteredAttrs, attr)
		}
	}

	// Criar novo resource apenas com os atributos filtrados
	// Se não houver atributos após filtragem, retornar o resource original
	if len(filteredAttrs) == 0 {
		return res
	}

	filteredRes, err := resource.New(context.Background(),
		resource.WithAttributes(filteredAttrs...),
	)
	// Se houver erro ao criar resource filtrado, retornar o original
	if err != nil {
		return res
	}

	return filteredRes
}
