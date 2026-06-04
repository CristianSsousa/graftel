# Graftel

[![Go Reference](https://pkg.go.dev/badge/github.com/CristianSsousa/graftel/v2.svg)](https://pkg.go.dev/github.com/CristianSsousa/graftel/v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/CristianSsousa/graftel/v2)](https://goreportcard.com/report/github.com/CristianSsousa/graftel/v2)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Graftel** é uma biblioteca Go que facilita o uso do OpenTelemetry, focada em **métricas, logs e traces**. Projetada para ser simples, intuitiva e seguir as melhores práticas da comunidade Go.

## Características

- Inicialização simplificada de métricas, logs **e traces** em uma única chamada
- Propagação automática de **trace context W3C** (`traceparent`/`tracestate`) entre serviços
- Suporte completo para métricas: Counter, Gauge, Histogram, UpDownCounter
- Logs estruturados com múltiplos níveis (Trace, Debug, Info, Warn, Error, Fatal)
- `Fatal`/`FatalWithFields` encerram o processo com `os.Exit(1)` após emitir o log
- Tracing distribuído com suporte completo a spans e traces via OTLP
- Middleware HTTP para Gin, Echo, Chi e net/http com observabilidade automática
- Helpers de contexto para propagação de tags e context logger
- Integração com Prometheus (opcional)
- Exportação via OTLP HTTP para qualquer backend compatível (Grafana, Jaeger, etc.)
- Processamento automático de URLs — aceita URLs completas com path
- Configuração via variáveis de ambiente `GRAFTEL_*`
- API fluente com pattern builder
- Interfaces bem definidas para testabilidade
- Resource sanitizado — remove campos sensíveis automaticamente

## Instalação

```bash
go get github.com/CristianSsousa/graftel/v2
```

## Uso Básico

### Inicialização

```go
package main

import (
    "context"
    "log"

    graftel "github.com/CristianSsousa/graftel/v2"
)

func main() {
    // Configurar com o pattern builder.
    // Prioridade: With*() > variáveis de ambiente GRAFTEL_* > valores padrão.
    config := graftel.NewConfig("meu-servico").
        WithServiceVersion("1.0.0").
        WithOTLPEndpoint("http://localhost:4318").
        WithInsecure(true) // somente para desenvolvimento local

    client, err := graftel.NewClient(config)
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    // Initialize configura métricas, logs e traces de uma vez.
    if err := client.Initialize(ctx); err != nil {
        log.Fatal(err)
    }
    defer client.Shutdown(ctx)
}
```

### Processamento de URLs

A biblioteca aceita diferentes formatos de URL para o endpoint OTLP:

```go
// URL completa com path (recomendado para Grafana Cloud)
WithOTLPEndpoint("https://otlp-gateway-prod-us-central-0.grafana.net/otlp")

// URL sem path (usa path padrão /v1/*)
WithOTLPEndpoint("http://localhost:4318")

// Apenas host:port
WithOTLPEndpoint("localhost:4318")

// Host:port com path customizado
WithOTLPEndpoint("localhost:4318/v1/custom")
```

## Métricas

### Counter

```go
metrics := client.NewMetricsHelper("meu-servico/metrics")

counter, err := metrics.NewCounter(
    "requests_total",
    "Total de requisições recebidas",
)
if err != nil {
    log.Fatal(err)
}

counter.Increment(ctx,
    attribute.String("method", "GET"),
    attribute.String("path", "/api/users"),
    attribute.Int("status", 200),
)

counter.Add(ctx, 5, attribute.String("method", "POST"))
```

### Histogram

```go
histogram, err := metrics.NewHistogram(
    "request_duration_seconds",
    "Duração das requisições em segundos",
)

start := time.Now()
// ... lógica ...
histogram.RecordDuration(ctx, time.Since(start),
    attribute.String("endpoint", "/api/users"),
)
```

### UpDownCounter

```go
connections, err := metrics.NewUpDownCounter(
    "active_connections",
    "Conexões ativas",
)

connections.Increment(ctx, attribute.String("type", "websocket"))
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
        observer.Observe(float64(m.Alloc), attribute.String("type", "heap"))
        return nil
    },
)
```

## Logs

```go
logs := client.NewLogsHelper("meu-servico/logs")

logs.Info(ctx, "Servidor iniciado",
    attribute.String("port", "8080"),
    attribute.String("environment", "production"),
)

logs.Error(ctx, "Falha ao processar requisição",
    attribute.String("error", "timeout"),
)

// Log com objeto de erro
err := fmt.Errorf("erro ao conectar ao banco")
logs.ErrorWithError(ctx, "Falha na conexão", err,
    attribute.String("database", "postgres"),
)

// Log com campos extras
logs.InfoWithFields(ctx, "Requisição processada",
    map[string]interface{}{
        "user_id":    12345,
        "request_id": "req-abc-123",
        "duration":   150.5,
    },
    attribute.String("method", "POST"),
)
```

> Todos os atributos customizados são automaticamente prefixados com `tags.` (ex: `"port"` vira `"tags.port"`) para melhor organização nos backends de observabilidade.

### Fatal

`Fatal` e `FatalWithFields` emitem o log com severity `FATAL` e encerram o processo com `os.Exit(1)`:

```go
logs.Fatal(ctx, "configuração inválida — encerrando",
    attribute.String("reason", "missing DATABASE_URL"),
)
// o processo termina aqui
```

## Tracing

### Spans Básicos

```go
tracing := client.NewTracingHelper("meu-servico")

ctx, span := tracing.StartSpanWithTags(ctx, "operacao",
    attribute.String("user_id", "123"),
)
defer span.End()

err := tracing.WithSpan(ctx, "processar-dados", func(ctx context.Context) error {
    // lógica
    return nil
}, attribute.String("tipo", "batch"))
```

### Spans com Retorno

```go
result, err := tracing.WithSpanAndReturn(ctx, "buscar-dados",
    func(ctx context.Context) (interface{}, error) {
        return dados, nil
    },
    attribute.String("tabela", "usuarios"),
)
```

### Gerenciamento de Erros e Status

```go
ctx, span := tracing.StartSpan(ctx, "operacao")
defer span.End()

if err != nil {
    tracing.SetSpanError(ctx, err, attribute.String("retry_count", "3"))
}

tracing.SetSpanStatus(ctx, codes.Ok, "Operação concluída")

traceID := tracing.GetTraceID(ctx)
spanID  := tracing.GetSpanID(ctx)
```

### Funções Globais Auxiliares

```go
err := graftel.WithSpanTiming(ctx, "operacao-lenta", func(ctx context.Context) error {
    return nil
}, attribute.String("tipo", "processamento"))
```

## Middleware HTTP

O middleware coleta automaticamente:

- **Métricas**: `http_requests_total`, `http_request_duration_seconds`, `http_request_size_bytes`, `http_response_size_bytes`
- **Traces**: span por requisição com atributos HTTP semconv
- **Logs**: log de entrada e saída de cada requisição
- **Propagação de trace**: extrai `traceparent`/`tracestate` dos headers de entrada, ligando spans de serviços distintos

### net/http

```go
import (
    "net/http"
    graftel "github.com/CristianSsousa/graftel/v2"
)

mux := http.NewServeMux()
mux.HandleFunc("/api/users", handler)

config := graftel.DefaultMiddlewareConfig("meu-servico")
http.ListenAndServe(":8080", graftel.HTTPMiddleware(client, config)(mux))
```

### Gin

```go
import (
    "github.com/gin-gonic/gin"
    graftel "github.com/CristianSsousa/graftel/v2"
)

router := gin.Default()
config := graftel.DefaultMiddlewareConfig("meu-servico")
router.Use(graftel.GinMiddleware(client, config))

router.GET("/api/users", func(c *gin.Context) {
    c.JSON(200, gin.H{"message": "OK"})
})

router.Run(":8080")
```

### Echo

```go
import (
    "github.com/labstack/echo/v4"
    graftel "github.com/CristianSsousa/graftel/v2"
)

e := echo.New()
config := graftel.DefaultMiddlewareConfig("meu-servico")
e.Use(graftel.EchoMiddleware(client, config))

e.GET("/api/users", func(c echo.Context) error {
    return c.JSON(200, map[string]string{"message": "OK"})
})

e.Start(":8080")
```

### Chi

```go
import (
    "github.com/go-chi/chi/v5"
    graftel "github.com/CristianSsousa/graftel/v2"
)

r := chi.NewRouter()
config := graftel.DefaultMiddlewareConfig("meu-servico")
r.Use(graftel.ChiMiddleware(client, config))

r.Get("/api/users", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("OK"))
})

http.ListenAndServe(":8080", r)
```

### Configuração do Middleware

```go
config := graftel.MiddlewareConfig{
    ServiceName:        "meu-servico",
    SkipPaths:          []string{"/health", "/metrics", "/ready"}, // padrão
    RecordRequestBody:  false,
    RecordResponseBody: false,
    MaxBodySize:        4096,
}

// Ou usar a configuração padrão
config := graftel.DefaultMiddlewareConfig("meu-servico")
```

### Propagação de Trace Distribuído

O middleware extrai automaticamente o contexto de trace dos headers W3C (`traceparent` / `tracestate`). Para ativar a propagação no servidor de origem, configure o propagador antes de fazer chamadas HTTP:

```go
// O Initialize() já registra o propagador globalmente:
//   otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
//       propagation.TraceContext{},
//       propagation.Baggage{},
//   ))

// Para injetar o contexto numa chamada HTTP sainte:
req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
resp, _ := http.DefaultClient.Do(req)
```

## Helpers de Contexto

```go
// Adicionar tags ao contexto
ctx = graftel.WithTags(ctx,
    attribute.String("user_id", "123"),
    attribute.String("request_id", "req-abc-123"),
)

// Obter e mesclar tags
tags := graftel.GetTagsFromContext(ctx)
ctx  = graftel.MergeContextTags(ctx, attribute.String("step", "validation"))
```

### Context Logger

Herda automaticamente as tags do contexto:

```go
ctx = graftel.WithTags(ctx,
    attribute.String("user_id", "123"),
    attribute.String("session_id", "sess-456"),
)

logger := graftel.NewContextLogger(logs, ctx)

// Todos os logs incluem user_id e session_id automaticamente
logger.Info("Operação iniciada")
logger.Error("Erro ao processar", attribute.String("error_code", "E001"))

// Tags adicionais apenas para este log
logger.WithTags(attribute.String("step", "validation")).Info("Validação concluída")
```

## Configuração

### Opções Disponíveis

| Método | Descrição | ENV | Padrão |
|---|---|---|---|
| `WithServiceVersion(v)` | Versão do serviço | `GRAFTEL_SERVICE_VERSION` | `""` |
| `WithOTLPEndpoint(url)` | Endpoint OTLP | `GRAFTEL_OTLP_ENDPOINT` | `http://localhost:4318` |
| `WithAPIKey(key)` | Chave de API (Bearer sem InstanceID, Basic com) | `GRAFTEL_API_KEY` | `""` |
| `WithInstanceID(id)` | ID da instância (`service.instance.id`) | `GRAFTEL_INSTANCE_ID` | `""` |
| `WithPrometheusEndpoint(addr)` | Habilita exporter Prometheus | `GRAFTEL_PROMETHEUS_ENDPOINT` | `""` |
| `WithResourceAttribute(k, v)` | Atributo extra no resource | — | — |
| `WithResourceAttributes(map)` | Múltiplos atributos no resource | — | — |
| `WithMetricExportInterval(d)` | Intervalo de exportação de métricas | `GRAFTEL_METRIC_EXPORT_INTERVAL` | `30s` |
| `WithLogExportInterval(d)` | Intervalo de exportação de logs | `GRAFTEL_LOG_EXPORT_INTERVAL` | `30s` |
| `WithExportTimeout(d)` | Timeout de exportação | `GRAFTEL_EXPORT_TIMEOUT` | `10s` |
| `WithInsecure(bool)` | Desabilita TLS | `GRAFTEL_INSECURE` | `false` |

### Autenticação

| Cenário | Header gerado |
|---|---|
| `WithInstanceID` + `WithAPIKey` | `Basic base64(instanceID:apiKey)` |
| Apenas `WithAPIKey` | `Bearer apiKey` |

### Configuração com Prometheus

Ao configurar `WithPrometheusEndpoint`, as métricas são expostas via Prometheus em vez de enviadas por OTLP. O servidor HTTP deve ser criado pela aplicação:

```go
config := graftel.NewConfig("meu-servico").
    WithPrometheusEndpoint(":8080")

client, _ := graftel.NewClient(config)
client.Initialize(ctx)
defer client.Shutdown(ctx)

exporter := client.GetPrometheusExporter()
http.Handle("/metrics", promhttp.HandlerFor(
    prometheus.DefaultGatherer,
    promhttp.HandlerOpts{},
))
http.ListenAndServe(":8080", nil)
```

> Quando `PrometheusEndpoint` está configurado, métricas vão **somente** para Prometheus (OTLP não é usado para métricas).

### Usando Apenas Variáveis de Ambiente

```bash
export GRAFTEL_SERVICE_NAME="meu-servico"
export GRAFTEL_OTLP_ENDPOINT="https://otlp.example.com/otlp"
export GRAFTEL_API_KEY="sua-chave"
export GRAFTEL_INSTANCE_ID="instance-123"
```

```go
config := graftel.NewConfig("") // ServiceName lido de GRAFTEL_SERVICE_NAME
client, _ := graftel.NewClient(config)
```

## Segurança — Resource Sanitizado

A biblioteca remove automaticamente campos sensíveis do Resource OpenTelemetry:

**Removidos:** `process.command_args`, `process.executable.path`, `process.executable.name`, `process.command`, `process.owner`

**Mantidos:** `process.pid`, `process.runtime.*`, `host.name`, `os.type`, `os.description`, `service.name`, `service.version`, `service.instance.id`

## Testando Localmente

### Testes unitários

```bash
go test ./... -v
go test ./... -cover
```

### End-to-end com Jaeger

Suba o Jaeger (aceita OTLP HTTP na porta 4318):

```bash
docker run -d --name jaeger \
  -p 4318:4318 \
  -p 16686:16686 \
  jaegertracing/all-in-one:latest
```

Execute um exemplo:

```bash
go run ./examples/middleware/main.go
curl http://localhost:8080/api/users
```

Acesse `http://localhost:16686` para visualizar os traces.

Para testar a propagação de trace entre serviços:

```bash
curl -H "traceparent: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01" \
     http://localhost:8080/api/users
```

O span criado pelo middleware aparecerá como filho do trace especificado no header.

## Exemplos

```bash
# Exemplo básico: métricas e logs
go run ./examples/basic/main.go

# Métricas via Prometheus
go run ./examples/prometheus/main.go
# Acesse http://localhost:8080/metrics

# Configuração via variáveis de ambiente (Grafana Cloud)
GRAFTEL_SERVICE_NAME="meu-servico" \
GRAFTEL_OTLP_ENDPOINT="https://otlp.grafana.net/otlp" \
GRAFTEL_API_KEY="sua-chave" \
go run ./examples/grafana-cloud/main.go

# Tracing com spans
go run ./examples/tracing/main.go

# Middleware HTTP com Gin
go run ./examples/middleware/main.go
# Acesse http://localhost:8080/api/users

# Context helpers e context logger
go run ./examples/context/main.go
```

## Estrutura do Projeto

```
.
├── client.go       # Cliente principal e inicialização
├── config.go       # Configuração com pattern builder
├── metrics.go      # Helpers para métricas
├── logs.go         # Helpers para logs
├── tracing.go      # Helpers para tracing
├── middleware.go   # Middlewares HTTP (Gin, Echo, Chi, net/http)
├── context.go      # Helpers de contexto e ContextLogger
├── errors.go       # Erros customizados
├── *_test.go       # Testes unitários
└── examples/       # Exemplos de uso
```

## Requisitos

- Go 1.23 ou superior
- OpenTelemetry SDK v1.38.0 ou superior

## Dependências Principais

- `go.opentelemetry.io/otel` — OpenTelemetry Go SDK
- `go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp` — Exportador OTLP para métricas
- `go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp` — Exportador OTLP para logs
- `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp` — Exportador OTLP para traces
- `go.opentelemetry.io/otel/exporters/prometheus` — Exportador Prometheus
- `github.com/gin-gonic/gin` — Framework Gin (opcional, para middleware)
- `github.com/labstack/echo/v4` — Framework Echo (opcional, para middleware)

## Contribuindo

1. Faça fork do projeto
2. Crie uma branch (`git checkout -b feature/minha-feature`)
3. Commit suas mudanças seguindo Conventional Commits (`feat:`, `fix:`, etc.)
4. Push para a branch (`git push origin feature/minha-feature`)
5. Abra um Pull Request

O repositório usa auto-tagging baseado em Conventional Commits para versionamento semântico automático.

## Licença

MIT — veja o arquivo [LICENSE](LICENSE) para detalhes.

## Autor

**Cristian S. Sousa** — [@CristianSsousa](https://github.com/CristianSsousa)
