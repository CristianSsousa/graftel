package graftel

import (
	"fmt"
	"os"
	"strconv"
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
	// Pode ser configurado via GRAFTEL_OTLP_ENDPOINT ou WithOTLPEndpoint.
	// Padrão: http://localhost:4318
	OTLPEndpoint string

	// APIKey é a chave de API para autenticação (obrigatória se usar autenticação).
	// Pode ser configurada via GRAFTEL_API_KEY ou WithAPIKey.
	APIKey string

	// InstanceID é o ID da instância (opcional).
	// Se fornecido, será usado como service.instance.id no resource OpenTelemetry.
	// Pode ser configurado via GRAFTEL_INSTANCE_ID ou WithInstanceID.
	InstanceID string

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

	// ExportTimeout é o timeout para exportação de dados OTLP.
	// Padrão: 10 segundos
	ExportTimeout time.Duration

	// Insecure desabilita TLS (apenas para desenvolvimento local).
	Insecure bool
}

// NewConfig cria uma nova configuração com valores padrão.
// O serviceName é obrigatório, mas pode ser fornecido via GRAFTEL_SERVICE_NAME.
// Valores são carregados na seguinte ordem de prioridade:
// 1. Valores passados via métodos With* (maior prioridade)
// 2. Variáveis de ambiente GRAFTEL_*
// 3. Valores padrão
func NewConfig(serviceName string) Config {
	config := Config{
		ServiceName:          serviceName,
		OTLPEndpoint:         "http://localhost:4318",
		MetricExportInterval: 30 * time.Second,
		LogExportInterval:    30 * time.Second,
		ExportTimeout:        10 * time.Second,
		ResourceAttributes:   make(map[string]string),
	}

	// Carregar valores de variáveis de ambiente se não foram fornecidos
	config.loadFromEnv()

	return config
}

// loadFromEnv carrega valores de variáveis de ambiente se os campos estiverem vazios.
func (c *Config) loadFromEnv() {
	// ServiceName - se vazio, tenta ENV
	if c.ServiceName == "" {
		if val := os.Getenv("GRAFTEL_SERVICE_NAME"); val != "" {
			c.ServiceName = val
		}
	}

	// ServiceVersion - se vazio, tenta ENV
	if c.ServiceVersion == "" {
		if val := os.Getenv("GRAFTEL_SERVICE_VERSION"); val != "" {
			c.ServiceVersion = val
		}
	}

	// OTLPEndpoint - se vazio ou padrão, tenta ENV
	if c.OTLPEndpoint == "" || c.OTLPEndpoint == "http://localhost:4318" {
		if val := os.Getenv("GRAFTEL_OTLP_ENDPOINT"); val != "" {
			c.OTLPEndpoint = val
		} else if c.OTLPEndpoint == "" {
			c.OTLPEndpoint = "http://localhost:4318"
		}
	}

	// APIKey - se vazio, tenta ENV
	if c.APIKey == "" {
		if val := os.Getenv("GRAFTEL_API_KEY"); val != "" {
			c.APIKey = val
		}
	}

	// InstanceID - se vazio, tenta ENV
	if c.InstanceID == "" {
		if val := os.Getenv("GRAFTEL_INSTANCE_ID"); val != "" {
			c.InstanceID = val
		}
	}

	// PrometheusEndpoint - se vazio, tenta ENV
	if c.PrometheusEndpoint == "" {
		if val := os.Getenv("GRAFTEL_PROMETHEUS_ENDPOINT"); val != "" {
			c.PrometheusEndpoint = val
		}
	}

	// Insecure - se false (padrão), tenta ENV
	if !c.Insecure {
		if val := os.Getenv("GRAFTEL_INSECURE"); val != "" {
			if insecure, err := strconv.ParseBool(val); err == nil {
				c.Insecure = insecure
			}
		}
	}

	// MetricExportInterval - se zero ou padrão, tenta ENV
	if c.MetricExportInterval == 0 || c.MetricExportInterval == 30*time.Second {
		if val := os.Getenv("GRAFTEL_METRIC_EXPORT_INTERVAL"); val != "" {
			if duration, err := time.ParseDuration(val); err == nil {
				c.MetricExportInterval = duration
			} else if c.MetricExportInterval == 0 {
				c.MetricExportInterval = 30 * time.Second
			}
		} else if c.MetricExportInterval == 0 {
			c.MetricExportInterval = 30 * time.Second
		}
	}

	// LogExportInterval - se zero ou padrão, tenta ENV
	if c.LogExportInterval == 0 || c.LogExportInterval == 30*time.Second {
		if val := os.Getenv("GRAFTEL_LOG_EXPORT_INTERVAL"); val != "" {
			if duration, err := time.ParseDuration(val); err == nil {
				c.LogExportInterval = duration
			} else if c.LogExportInterval == 0 {
				c.LogExportInterval = 30 * time.Second
			}
		} else if c.LogExportInterval == 0 {
			c.LogExportInterval = 30 * time.Second
		}
	}

	// ExportTimeout - se zero ou padrão, tenta ENV
	if c.ExportTimeout == 0 || c.ExportTimeout == 10*time.Second {
		if val := os.Getenv("GRAFTEL_EXPORT_TIMEOUT"); val != "" {
			if duration, err := time.ParseDuration(val); err == nil {
				c.ExportTimeout = duration
			} else if c.ExportTimeout == 0 {
				c.ExportTimeout = 10 * time.Second
			}
		} else if c.ExportTimeout == 0 {
			c.ExportTimeout = 10 * time.Second
		}
	}
}

// Validate valida a configuração e retorna um erro se inválida.
// Define valores padrão se não foram configurados.
func (c *Config) Validate() error {
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

	if c.ExportTimeout == 0 {
		c.ExportTimeout = 10 * time.Second
	}

	return nil
}

// WithServiceVersion define a versão do serviço.
// Se não fornecido, será lido de GRAFTEL_SERVICE_VERSION.
func (c Config) WithServiceVersion(version string) Config {
	c.ServiceVersion = version
	return c
}

// WithOTLPEndpoint define o endpoint OTLP.
// Se não fornecido, será lido de GRAFTEL_OTLP_ENDPOINT.
func (c Config) WithOTLPEndpoint(endpoint string) Config {
	c.OTLPEndpoint = endpoint
	return c
}

// WithAPIKey define a chave de API para autenticação.
// Se não fornecido, será lido de GRAFTEL_API_KEY.
func (c Config) WithAPIKey(apiKey string) Config {
	c.APIKey = apiKey
	return c
}

// WithInstanceID define o ID da instância.
// Este ID será usado como service.instance.id no resource OpenTelemetry.
// Se não fornecido, será lido de GRAFTEL_INSTANCE_ID.
func (c Config) WithInstanceID(instanceID string) Config {
	c.InstanceID = instanceID
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

// WithExportTimeout define o timeout para exportação de dados OTLP.
func (c Config) WithExportTimeout(timeout time.Duration) Config {
	c.ExportTimeout = timeout
	return c
}

// WithInsecure desabilita TLS (apenas para desenvolvimento local).
func (c Config) WithInsecure(insecure bool) Config {
	c.Insecure = insecure
	return c
}
