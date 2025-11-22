package graftel

import (
	"os"
	"testing"
	"time"
)

func TestConfig_WithAPIKey(t *testing.T) {
	apiKey := "test-api-key"
	config := NewConfig("test-service").WithAPIKey(apiKey)
	if config.APIKey != apiKey {
		t.Errorf("esperado APIKey '%s', obtido '%s'", apiKey, config.APIKey)
	}
}

func TestConfig_WithInstanceID(t *testing.T) {
	instanceID := "instance-123"
	config := NewConfig("test-service").WithInstanceID(instanceID)
	if config.InstanceID != instanceID {
		t.Errorf("esperado InstanceID '%s', obtido '%s'", instanceID, config.InstanceID)
	}
}

func TestConfig_WithPrometheusEndpoint(t *testing.T) {
	endpoint := ":8080"
	config := NewConfig("test-service").WithPrometheusEndpoint(endpoint)
	if config.PrometheusEndpoint != endpoint {
		t.Errorf("esperado PrometheusEndpoint '%s', obtido '%s'", endpoint, config.PrometheusEndpoint)
	}
}

func TestConfig_WithResourceAttributes(t *testing.T) {
	attrs := map[string]string{
		"env":  "production",
		"team": "frontend",
	}
	config := NewConfig("test-service").WithResourceAttributes(attrs)

	if config.ResourceAttributes["env"] != "production" {
		t.Errorf("esperado atributo 'env'='production', obtido '%s'", config.ResourceAttributes["env"])
	}
	if config.ResourceAttributes["team"] != "frontend" {
		t.Errorf("esperado atributo 'team'='frontend', obtido '%s'", config.ResourceAttributes["team"])
	}
}

func TestConfig_WithMetricExportInterval(t *testing.T) {
	interval := 60 * time.Second
	config := NewConfig("test-service").WithMetricExportInterval(interval)
	if config.MetricExportInterval != interval {
		t.Errorf("esperado MetricExportInterval %v, obtido %v", interval, config.MetricExportInterval)
	}
}

func TestConfig_WithLogExportInterval(t *testing.T) {
	interval := 45 * time.Second
	config := NewConfig("test-service").WithLogExportInterval(interval)
	if config.LogExportInterval != interval {
		t.Errorf("esperado LogExportInterval %v, obtido %v", interval, config.LogExportInterval)
	}
}

func TestConfig_WithExportTimeout(t *testing.T) {
	timeout := 20 * time.Second
	config := NewConfig("test-service").WithExportTimeout(timeout)
	if config.ExportTimeout != timeout {
		t.Errorf("esperado ExportTimeout %v, obtido %v", timeout, config.ExportTimeout)
	}
}

func TestConfig_WithInsecure(t *testing.T) {
	config := NewConfig("test-service").WithInsecure(true)
	if !config.Insecure {
		t.Error("esperado Insecure=true, obtido false")
	}

	config = config.WithInsecure(false)
	if config.Insecure {
		t.Error("esperado Insecure=false, obtido true")
	}
}

func TestConfig_Validate_DefaultValues(t *testing.T) {
	config := Config{
		ServiceName: "test-service",
	}

	err := config.Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if config.OTLPEndpoint != "http://localhost:4318" {
		t.Errorf("esperado OTLPEndpoint padrão, obtido '%s'", config.OTLPEndpoint)
	}
	if config.MetricExportInterval != 30*time.Second {
		t.Errorf("esperado MetricExportInterval padrão, obtido %v", config.MetricExportInterval)
	}
	if config.LogExportInterval != 30*time.Second {
		t.Errorf("esperado LogExportInterval padrão, obtido %v", config.LogExportInterval)
	}
	if config.ExportTimeout != 10*time.Second {
		t.Errorf("esperado ExportTimeout padrão, obtido %v", config.ExportTimeout)
	}
}

func TestConfig_Chaining(t *testing.T) {
	config := NewConfig("test-service").
		WithServiceVersion("1.0.0").
		WithOTLPEndpoint("https://example.com").
		WithAPIKey("key123").
		WithInstanceID("inst456").
		WithPrometheusEndpoint(":8080").
		WithResourceAttribute("env", "test").
		WithMetricExportInterval(60 * time.Second).
		WithLogExportInterval(45 * time.Second).
		WithExportTimeout(20 * time.Second).
		WithInsecure(true)

	if config.ServiceName != "test-service" {
		t.Errorf("ServiceName = %v, esperado 'test-service'", config.ServiceName)
	}
	if config.ServiceVersion != "1.0.0" {
		t.Errorf("ServiceVersion = %v, esperado '1.0.0'", config.ServiceVersion)
	}
	if config.OTLPEndpoint != "https://example.com" {
		t.Errorf("OTLPEndpoint = %v, esperado 'https://example.com'", config.OTLPEndpoint)
	}
	if config.APIKey != "key123" {
		t.Errorf("APIKey = %v, esperado 'key123'", config.APIKey)
	}
	if config.InstanceID != "inst456" {
		t.Errorf("InstanceID = %v, esperado 'inst456'", config.InstanceID)
	}
	if config.PrometheusEndpoint != ":8080" {
		t.Errorf("PrometheusEndpoint = %v, esperado ':8080'", config.PrometheusEndpoint)
	}
	if config.ResourceAttributes["env"] != "test" {
		t.Errorf("ResourceAttributes['env'] = %v, esperado 'test'", config.ResourceAttributes["env"])
	}
	if config.MetricExportInterval != 60*time.Second {
		t.Errorf("MetricExportInterval = %v, esperado 60s", config.MetricExportInterval)
	}
	if config.LogExportInterval != 45*time.Second {
		t.Errorf("LogExportInterval = %v, esperado 45s", config.LogExportInterval)
	}
	if config.ExportTimeout != 20*time.Second {
		t.Errorf("ExportTimeout = %v, esperado 20s", config.ExportTimeout)
	}
	if !config.Insecure {
		t.Error("Insecure = false, esperado true")
	}
}

func TestConfig_LoadFromEnv(t *testing.T) {
	os.Setenv("GRAFTEL_SERVICE_NAME", "env-service")
	os.Setenv("GRAFTEL_SERVICE_VERSION", "2.0.0")
	os.Setenv("GRAFTEL_OTLP_ENDPOINT", "https://env-endpoint.com")
	os.Setenv("GRAFTEL_API_KEY", "env-key")
	os.Setenv("GRAFTEL_INSTANCE_ID", "env-instance")
	os.Setenv("GRAFTEL_PROMETHEUS_ENDPOINT", ":9090")
	os.Setenv("GRAFTEL_INSECURE", "true")
	os.Setenv("GRAFTEL_METRIC_EXPORT_INTERVAL", "60s")
	os.Setenv("GRAFTEL_LOG_EXPORT_INTERVAL", "45s")
	os.Setenv("GRAFTEL_EXPORT_TIMEOUT", "20s")

	defer func() {
		os.Unsetenv("GRAFTEL_SERVICE_NAME")
		os.Unsetenv("GRAFTEL_SERVICE_VERSION")
		os.Unsetenv("GRAFTEL_OTLP_ENDPOINT")
		os.Unsetenv("GRAFTEL_API_KEY")
		os.Unsetenv("GRAFTEL_INSTANCE_ID")
		os.Unsetenv("GRAFTEL_PROMETHEUS_ENDPOINT")
		os.Unsetenv("GRAFTEL_INSECURE")
		os.Unsetenv("GRAFTEL_METRIC_EXPORT_INTERVAL")
		os.Unsetenv("GRAFTEL_LOG_EXPORT_INTERVAL")
		os.Unsetenv("GRAFTEL_EXPORT_TIMEOUT")
	}()

	config := NewConfig("")
	if config.ServiceName != "env-service" {
		t.Errorf("ServiceName = %v, esperado 'env-service'", config.ServiceName)
	}
	if config.ServiceVersion != "2.0.0" {
		t.Errorf("ServiceVersion = %v, esperado '2.0.0'", config.ServiceVersion)
	}
	if config.OTLPEndpoint != "https://env-endpoint.com" {
		t.Errorf("OTLPEndpoint = %v, esperado 'https://env-endpoint.com'", config.OTLPEndpoint)
	}
	if config.APIKey != "env-key" {
		t.Errorf("APIKey = %v, esperado 'env-key'", config.APIKey)
	}
	if config.InstanceID != "env-instance" {
		t.Errorf("InstanceID = %v, esperado 'env-instance'", config.InstanceID)
	}
	if config.PrometheusEndpoint != ":9090" {
		t.Errorf("PrometheusEndpoint = %v, esperado ':9090'", config.PrometheusEndpoint)
	}
	if !config.Insecure {
		t.Error("Insecure = false, esperado true")
	}
	if config.MetricExportInterval != 60*time.Second {
		t.Errorf("MetricExportInterval = %v, esperado 60s", config.MetricExportInterval)
	}
	if config.LogExportInterval != 45*time.Second {
		t.Errorf("LogExportInterval = %v, esperado 45s", config.LogExportInterval)
	}
	if config.ExportTimeout != 20*time.Second {
		t.Errorf("ExportTimeout = %v, esperado 20s", config.ExportTimeout)
	}
}

func TestConfig_WithMethodsOverrideEnv(t *testing.T) {
	os.Setenv("GRAFTEL_SERVICE_VERSION", "env-version")
	defer os.Unsetenv("GRAFTEL_SERVICE_VERSION")

	config := NewConfig("test").WithServiceVersion("method-version")
	if config.ServiceVersion != "method-version" {
		t.Errorf("ServiceVersion = %v, esperado 'method-version' (With* tem prioridade)", config.ServiceVersion)
	}
}
