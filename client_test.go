package graftel

import (
	"context"
	"testing"
)

func TestNewConfig(t *testing.T) {
	config := NewConfig("test-service")
	if config.ServiceName != "test-service" {
		t.Errorf("esperado ServiceName 'test-service', obtido '%s'", config.ServiceName)
	}
	if config.OTLPEndpoint != "http://localhost:4318" {
		t.Errorf("esperado OTLPEndpoint padrão, obtido '%s'", config.OTLPEndpoint)
	}
}

func TestConfig_WithServiceVersion(t *testing.T) {
	config := NewConfig("test-service").WithServiceVersion("1.0.0")
	if config.ServiceVersion != "1.0.0" {
		t.Errorf("esperado ServiceVersion '1.0.0', obtido '%s'", config.ServiceVersion)
	}
}

func TestConfig_WithOTLPEndpoint(t *testing.T) {
	endpoint := "https://example.com/otlp"
	config := NewConfig("test-service").WithOTLPEndpoint(endpoint)
	if config.OTLPEndpoint != endpoint {
		t.Errorf("esperado OTLPEndpoint '%s', obtido '%s'", endpoint, config.OTLPEndpoint)
	}
}

func TestConfig_WithResourceAttribute(t *testing.T) {
	config := NewConfig("test-service").
		WithResourceAttribute("env", "test").
		WithResourceAttribute("team", "backend")

	if config.ResourceAttributes["env"] != "test" {
		t.Errorf("esperado atributo 'env'='test', obtido '%s'", config.ResourceAttributes["env"])
	}
	if config.ResourceAttributes["team"] != "backend" {
		t.Errorf("esperado atributo 'team'='backend', obtido '%s'", config.ResourceAttributes["team"])
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "config válida",
			config:  NewConfig("test-service"),
			wantErr: false,
		},
		{
			name:    "config sem ServiceName",
			config:  Config{},
			wantErr: true,
		},
		{
			name: "config com valores padrão aplicados",
			config: Config{
				ServiceName: "test-service",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	config := NewConfig("test-service")
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v, esperado nil", err)
	}
	if client == nil {
		t.Fatal("NewClient() retornou nil, esperado cliente")
	}
}

func TestNewClient_InvalidConfig(t *testing.T) {
	config := Config{} // ServiceName vazio
	_, err := NewClient(config)
	if err == nil {
		t.Error("NewClient() esperado erro com config inválida, obtido nil")
	}
}

// TestClient_Shutdown testa o shutdown sem inicializar (não deve causar panic)
func TestClient_Shutdown(t *testing.T) {
	config := NewConfig("test-service")
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	ctx := context.Background()
	// Shutdown sem inicializar não deve causar panic
	err = client.Shutdown(ctx)
	if err != nil {
		// Pode retornar erro, mas não deve causar panic
		t.Logf("Shutdown() retornou erro (esperado): %v", err)
	}
}

// TestClient_initializeTraces testa a inicialização de traces
func TestClient_initializeTraces(t *testing.T) {
	config := NewConfig("test-service").WithInsecure(true)
	cl, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// Fazer type assertion para acessar o método privado
	c, ok := cl.(*client)
	if !ok {
		t.Fatal("Falha ao fazer type assertion para *client")
	}

	ctx := context.Background()
	err = c.initializeTraces(ctx)
	if err != nil {
		t.Fatalf("initializeTraces() error = %v", err)
	}

	// Verificar se o traceProvider foi inicializado
	if c.traceProvider == nil {
		t.Error("traceProvider não foi inicializado")
	}

	// Verificar se GetTracer funciona após inicializar traces
	tracer := cl.GetTracer("test-tracer")
	if tracer == nil {
		t.Error("GetTracer() retornou nil após inicializar traces")
	}

	// Limpar
	defer func() {
		if err := cl.Shutdown(ctx); err != nil {
			t.Logf("Erro ao encerrar cliente: %v", err)
		}
	}()
}
