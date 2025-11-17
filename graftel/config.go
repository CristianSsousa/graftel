package graftel

import (
	"fmt"
	"time"
)

// Config contém as configurações para inicializar o OpenTelemetry.
// Use NewConfig para criar uma configuração com valores padrão.
type Config struct {
	// ServiceName é o nome do serviço (obrigatório).
	ServiceName string

	// ServiceVersion é a versão do serviço.
	ServiceVersion string

	// OTLPEndpoint é o endpoint OTLP para métricas e logs.
	// Para Grafana Cloud: https://otlp-gateway-prod-us-central-0.grafana.net/otlp
	// Para local: http://localhost:4318
	// Padrão: http://localhost:4318
	OTLPEndpoint string

	// GrafanaCloudAPIKey é a chave de API do Grafana Cloud (obrigatória se usar Grafana Cloud).
	GrafanaCloudAPIKey string

	// PrometheusEndpoint é o endpoint para expor métricas Prometheus (ex: :8080).
	// Se vazio, não expõe endpoint Prometheus.
	PrometheusEndpoint string

	// ResourceAttributes são atributos adicionais para o resource.
	ResourceAttributes map[string]string

	// MetricExportInterval é o intervalo de exportação de métricas.
	// Padrão: 30 segundos
	MetricExportInterval time.Duration

	// LogExportInterval é o intervalo de exportação de logs.
	// Padrão: 30 segundos
	LogExportInterval time.Duration

	// Insecure desabilita TLS (apenas para desenvolvimento local).
	Insecure bool
}

// NewConfig cria uma nova configuração com valores padrão.
// O serviceName é obrigatório.
func NewConfig(serviceName string) Config {
	return Config{
		ServiceName:         serviceName,
		OTLPEndpoint:        "http://localhost:4318",
		MetricExportInterval: 30 * time.Second,
		LogExportInterval:    30 * time.Second,
		ResourceAttributes:  make(map[string]string),
	}
}

// Validate valida a configuração e retorna um erro se inválida.
func (c Config) Validate() error {
	if c.ServiceName == "" {
		return fmt.Errorf("ServiceName é obrigatório")
	}

	if c.OTLPEndpoint == "" {
		c.OTLPEndpoint = "http://localhost:4318"
	}

	if c.MetricExportInterval == 0 {
		c.MetricExportInterval = 30 * time.Second
	}

	if c.LogExportInterval == 0 {
		c.LogExportInterval = 30 * time.Second
	}

	return nil
}

// WithServiceVersion define a versão do serviço.
func (c Config) WithServiceVersion(version string) Config {
	c.ServiceVersion = version
	return c
}

// WithOTLPEndpoint define o endpoint OTLP.
func (c Config) WithOTLPEndpoint(endpoint string) Config {
	c.OTLPEndpoint = endpoint
	return c
}

// WithGrafanaCloudAPIKey define a chave de API do Grafana Cloud.
func (c Config) WithGrafanaCloudAPIKey(apiKey string) Config {
	c.GrafanaCloudAPIKey = apiKey
	return c
}

// WithPrometheusEndpoint define o endpoint para expor métricas Prometheus.
func (c Config) WithPrometheusEndpoint(endpoint string) Config {
	c.PrometheusEndpoint = endpoint
	return c
}

// WithResourceAttribute adiciona um atributo ao resource.
func (c Config) WithResourceAttribute(key, value string) Config {
	if c.ResourceAttributes == nil {
		c.ResourceAttributes = make(map[string]string)
	}
	c.ResourceAttributes[key] = value
	return c
}

// WithResourceAttributes adiciona múltiplos atributos ao resource.
func (c Config) WithResourceAttributes(attrs map[string]string) Config {
	if c.ResourceAttributes == nil {
		c.ResourceAttributes = make(map[string]string)
	}
	for k, v := range attrs {
		c.ResourceAttributes[k] = v
	}
	return c
}

// WithMetricExportInterval define o intervalo de exportação de métricas.
func (c Config) WithMetricExportInterval(interval time.Duration) Config {
	c.MetricExportInterval = interval
	return c
}

// WithLogExportInterval define o intervalo de exportação de logs.
func (c Config) WithLogExportInterval(interval time.Duration) Config {
	c.LogExportInterval = interval
	return c
}

// WithInsecure desabilita TLS (apenas para desenvolvimento local).
func (c Config) WithInsecure(insecure bool) Config {
	c.Insecure = insecure
	return c
}

