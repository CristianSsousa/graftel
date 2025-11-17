# Graftel

[![Go Reference](https://pkg.go.dev/badge/github.com/CristianSsousa/graftel.svg)](https://pkg.go.dev/github.com/CristianSsousa/graftel)
[![Go Report Card](https://goreportcard.com/badge/github.com/CristianSsousa/graftel)](https://goreportcard.com/report/github.com/CristianSsousa/graftel)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Graftel** é uma biblioteca Go que facilita o uso do OpenTelemetry com Grafana, focada em **métricas e logs**. Projetada para ser simples, intuitiva e seguir as melhores práticas da comunidade Go.

## 🚀 Características

- ✅ **Inicialização simplificada** do OpenTelemetry
- ✅ **Suporte completo para métricas**: Counter, Gauge, Histogram, UpDownCounter
- ✅ **Logs estruturados** com múltiplos níveis (Trace, Debug, Info, Warn, Error, Fatal)
- ✅ **Integração com Prometheus** (opcional)
- ✅ **Exportação via OTLP HTTP** para Grafana
- ✅ **API fluente** com pattern builder
- ✅ **Interfaces bem definidas** para testabilidade
- ✅ **Documentação completa** com exemplos práticos
- ✅ **Compatível com Grafana Cloud**

## 📦 Instalação

```bash
go get github.com/CristianSsousa/graftel
```

## 🎯 Uso Básico

### Inicialização

```go
package main

import (
    "context"
    "log"

    "github.com/CristianSsousa/graftel"
)

func main() {
    // Configurar usando o pattern de builder
    config := graftel.NewConfig("meu-servico").
        WithServiceVersion("1.0.0").
        WithOTLPEndpoint("http://localhost:4318").
        WithInsecure(true) // Para desenvolvimento local

    client, err := graftel.NewClient(config)
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    if err := client.Initialize(ctx); err != nil {
        log.Fatal(err)
    }
    defer client.Shutdown(ctx)

    // Usar métricas e logs...
}
```

## 📊 Métricas

### Counter (Contador)

```go
metrics := client.NewMetricsHelper("meu-servico/metrics")

counter, err := metrics.NewCounter(
    "requests_total",
    "Total de requisições recebidas",
)
if err != nil {
    log.Fatal(err)
}

// Incrementar contador
counter.Increment(ctx,
    attribute.String("method", "GET"),
    attribute.String("path", "/api/users"),
    attribute.Int("status", 200),
)

// Adicionar valor específico
counter.Add(ctx, 5, attribute.String("method", "POST"))
```

### Histogram

```go
histogram, err := metrics.NewHistogram(
    "request_duration_seconds",
    "Duração das requisições em segundos",
)
if err != nil {
    log.Fatal(err)
}

// Registrar duração
start := time.Now()
// ... fazer algo ...
duration := time.Since(start)
histogram.RecordDuration(ctx, duration,
    attribute.String("endpoint", "/api/users"),
)
```

### UpDownCounter

```go
connections, err := metrics.NewUpDownCounter(
    "active_connections",
    "Número de conexões ativas",
)
if err != nil {
    log.Fatal(err)
}

// Incrementar
connections.Increment(ctx, attribute.String("type", "websocket"))

// Decrementar
connections.Decrement(ctx, attribute.String("type", "websocket"))
```

### Gauge (Observable)

```go
gauge, err := metrics.NewGauge(
    "memory_usage_bytes",
    "Uso de memória em bytes",
    func(ctx context.Context, observer metric.Float64Observer) error {
        var m runtime.MemStats
        runtime.ReadMemStats(&m)
        observer.Observe(float64(m.Alloc),
            attribute.String("type", "heap"))
        return nil
    },
)
```

## 📝 Logs

### Logs Simples

```go
logs := client.NewLogsHelper("meu-servico/logs")

// Logs simples
logs.Info(ctx, "Servidor iniciado",
    attribute.String("port", "8080"),
    attribute.String("environment", "production"),
)

logs.Debug(ctx, "Processando requisição",
    attribute.String("method", "GET"),
    attribute.String("path", "/api/users"),
)

logs.Warn(ctx, "Tentativa de acesso não autorizado",
    attribute.String("ip", "192.168.1.1"),
)

logs.Error(ctx, "Falha ao processar requisição",
    attribute.String("error", "timeout"),
)

// Log com erro
err := fmt.Errorf("erro ao conectar ao banco")
logs.ErrorWithError(ctx, "Falha na conexão", err,
    attribute.String("database", "postgres"),
)
```

### Logs com Campos Extras

```go
// Logs com campos extras usando map
logs.InfoWithFields(ctx, "Processando requisição",
    map[string]interface{}{
        "user_id":    12345,
        "request_id": "req-abc-123",
        "ip":         "192.168.1.1",
        "duration":   150.5,
        "success":    true,
    },
    attribute.String("method", "POST"),
    attribute.String("path", "/api/users"),
)

// Log de erro com campos extras
err := fmt.Errorf("falha na conexão")
logs.ErrorWithFields(ctx, "Erro ao processar",
    map[string]interface{}{
        "error_code": "DB_CONNECTION_FAILED",
        "retry_count": 3,
    },
)
logs.ErrorWithError(ctx, "Erro ao processar", err,
    attribute.String("error_code", "DB_CONNECTION_FAILED"),
)
```

## ⚙️ Configuração

### Configuração com Prometheus

Para expor métricas via Prometheus (útil para Grafana):

```go
config := graftel.NewConfig("meu-servico").
    WithPrometheusEndpoint(":8080") // Expor em http://localhost:8080/metrics

client, err := graftel.NewClient(config)
if err != nil {
    log.Fatal(err)
}

ctx := context.Background()
if err := client.Initialize(ctx); err != nil {
    log.Fatal(err)
}
defer client.Shutdown(ctx)

// Obter exporter Prometheus
exporter := client.GetPrometheusExporter()
if exporter != nil {
    http.Handle("/metrics", exporter)
    http.ListenAndServe(":8080", nil)
}
```

### Configuração Avançada

```go
config := graftel.NewConfig("meu-servico").
    WithServiceVersion("1.0.0").
    WithOTLPEndpoint("http://localhost:4318").
    WithResourceAttributes(map[string]string{
        "environment": "production",
        "team":        "backend",
    }).
    WithMetricExportInterval(30 * time.Second).
    WithLogExportInterval(30 * time.Second)
```

## ☁️ Integração com Grafana Cloud

### Configuração Básica

```go
config := graftel.NewConfig("meu-servico").
    WithServiceVersion("1.0.0").
    WithOTLPEndpoint("https://otlp-gateway-prod-us-central-0.grafana.net/otlp").
    WithGrafanaCloudAPIKey("sua-chave-api-aqui").
    WithInsecure(false) // Grafana Cloud usa HTTPS
```

### Obter Chave de API do Grafana Cloud

1. Acesse o [Grafana Cloud](https://grafana.com)
2. Vá em **Connections** > **Add new connection**
3. Selecione **OpenTelemetry**
4. Copie a chave de API fornecida
5. Configure a variável de ambiente `GRAFANA_CLOUD_API_KEY` ou passe diretamente na configuração

### Exemplo Completo

Veja `examples/grafana-cloud/main.go` para um exemplo completo de integração.

## 📚 Exemplos

A biblioteca inclui exemplos completos na pasta `examples/`:

- **`examples/basic/`** - Exemplo básico com métricas e logs
- **`examples/prometheus/`** - Exemplo com Prometheus
- **`examples/grafana-cloud/`** - Exemplo de integração com Grafana Cloud

Para executar um exemplo:

```bash
cd examples/basic
go run main.go
```

## 🏗️ Estrutura do Projeto

```
.
├── graftel/              # Pacote principal da biblioteca
│   ├── client.go         # Cliente principal e inicialização
│   ├── config.go         # Configuração com pattern builder
│   ├── metrics.go        # Helpers para métricas
│   ├── logs.go           # Helpers para logs
│   ├── errors.go         # Erros customizados
│   └── client_test.go    # Testes unitários
├── examples/             # Exemplos de uso
│   ├── basic/            # Exemplo básico
│   ├── prometheus/       # Exemplo com Prometheus
│   └── grafana-cloud/    # Exemplo com Grafana Cloud
├── go.mod                # Dependências do módulo
├── go.sum                # Checksums das dependências
├── .gitignore            # Arquivos ignorados pelo Git
└── README.md             # Esta documentação
```

## 🧪 Testes

Execute os testes:

```bash
go test ./graftel/... -v
```

## 📋 Requisitos

- Go 1.23 ou superior
- OpenTelemetry SDK v1.38.0 ou superior

## 🔗 Dependências Principais

- `go.opentelemetry.io/otel` - OpenTelemetry Go SDK
- `go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp` - Exportador OTLP para métricas
- `go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp` - Exportador OTLP para logs
- `go.opentelemetry.io/otel/exporters/prometheus` - Exportador Prometheus

## 🤝 Contribuindo

Contribuições são bem-vindas! Sinta-se à vontade para:

1. Fazer fork do projeto
2. Criar uma branch para sua feature (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abrir um Pull Request

## 📄 Licença

Este projeto está licenciado sob a Licença MIT - veja o arquivo [LICENSE](LICENSE) para detalhes.

## 👤 Autor

**Cristian S. Sousa**

- GitHub: [@CristianSsousa](https://github.com/CristianSsousa)
- Repositório: [github.com/CristianSsousa/graftel](https://github.com/CristianSsousa/graftel)

## 🙏 Agradecimentos

- [OpenTelemetry](https://opentelemetry.io/) pela excelente especificação e SDK
- [Grafana](https://grafana.com/) pela plataforma de observabilidade
- Comunidade Go por todas as ferramentas e bibliotecas incríveis

## 📖 Documentação Adicional

- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [Grafana Cloud Documentation](https://grafana.com/docs/grafana-cloud/)
- [Go Documentation](https://go.dev/doc/)

---

⭐ Se este projeto foi útil para você, considere dar uma estrela no repositório!
