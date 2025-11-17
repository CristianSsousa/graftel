# Graftel

[![Go Reference](https://pkg.go.dev/badge/github.com/CristianSsousa/graftel.svg)](https://pkg.go.dev/github.com/CristianSsousa/graftel)
[![Go Report Card](https://goreportcard.com/badge/github.com/CristianSsousa/graftel)](https://goreportcard.com/report/github.com/CristianSsousa/graftel)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Graftel** é uma biblioteca Go que facilita o uso do OpenTelemetry, focada em **métricas e logs**. Projetada para ser simples, intuitiva e seguir as melhores práticas da comunidade Go.

## 🚀 Características

-   ✅ **Inicialização simplificada** do OpenTelemetry
-   ✅ **Suporte completo para métricas**: Counter, Gauge, Histogram, UpDownCounter
-   ✅ **Logs estruturados** com múltiplos níveis (Trace, Debug, Info, Warn, Error, Fatal)
-   ✅ **Integração com Prometheus** (opcional)
-   ✅ **Exportação via OTLP HTTP** para sistemas de observabilidade
-   ✅ **Processamento automático de URLs** - aceita URLs completas com path
-   ✅ **Configuração via variáveis de ambiente** - suporte completo a ENVs
-   ✅ **API fluente** com pattern builder
-   ✅ **Interfaces bem definidas** para testabilidade
-   ✅ **Documentação completa** com exemplos práticos
-   ✅ **Atributos de log organizados** - prefixo automático `tags.` para melhor estruturação
-   ✅ **Resource sanitizado** - remove campos sensíveis automaticamente

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
    // As configurações podem ser fornecidas via variáveis de ambiente GRAFTEL_*
    // ou explicitamente via métodos With*. A ordem de prioridade é:
    // 1. Valores passados via With* (maior prioridade)
    // 2. Variáveis de ambiente GRAFTEL_*
    // 3. Valores padrão

    config := graftel.NewConfig("meu-servico").
        WithServiceVersion("1.0.0").
        WithOTLPEndpoint("http://localhost:4318"). // Aceita URLs completas com path
        WithInsecure(true) // Para desenvolvimento local (HTTP sem TLS)

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

### Processamento de URLs

A biblioteca processa automaticamente diferentes formatos de URL:

-   **URLs completas**: `https://example.com:4318/v1/traces` → extrai host:port e path
-   **URLs sem path**: `http://localhost:4318` → usa path padrão
-   **Host:port simples**: `localhost:4318` → funciona normalmente
-   **Host:port com path**: `localhost:4318/otlp` → extrai path corretamente

O processamento é feito automaticamente, então você pode usar qualquer formato que preferir.

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
// Nota: Todos os atributos customizados são automaticamente prefixados com "tags."
// para melhor organização (ex: "port" vira "tags.port")
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

### Formatos de URL Suportados

A biblioteca aceita diferentes formatos de URL para o endpoint OTLP:

```go
// URL completa com protocolo e path (recomendado)
config := graftel.NewConfig("meu-servico").
    WithOTLPEndpoint("https://otlp-gateway-prod-us-central-0.grafana.net/otlp")

// URL completa sem path (usa path padrão)
config := graftel.NewConfig("meu-servico").
    WithOTLPEndpoint("http://localhost:4318")

// Apenas host:port (sem protocolo)
config := graftel.NewConfig("meu-servico").
    WithOTLPEndpoint("localhost:4318")

// Host:port com path
config := graftel.NewConfig("meu-servico").
    WithOTLPEndpoint("localhost:4318/v1/traces")
```

**Nota:** A biblioteca processa automaticamente a URL, extraindo o host:port e o path quando necessário. URLs completas com `http://` ou `https://` são automaticamente parseadas.

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
    WithLogExportInterval(30 * time.Second).
    WithInsecure(true) // Para desenvolvimento local (HTTP sem TLS)
```

### Opções de Configuração Disponíveis

| Método                               | Descrição                                                 | ENV                              | Padrão                    |
| ------------------------------------ | --------------------------------------------------------- | -------------------------------- | ------------------------- |
| `WithServiceVersion(version)`        | Define a versão do serviço                                | `GRAFTEL_SERVICE_VERSION`        | `""`                      |
| `WithOTLPEndpoint(endpoint)`         | Define o endpoint OTLP (aceita URLs completas)            | `GRAFTEL_OTLP_ENDPOINT`          | `"http://localhost:4318"` |
| `WithAPIKey(key)`                    | Define a chave de API para autenticação                   | `GRAFTEL_API_KEY`                | `""`                      |
| `WithInstanceID(id)`                 | Define o ID da instância (usado como service.instance.id) | `GRAFTEL_INSTANCE_ID`            | `""`                      |
| `WithPrometheusEndpoint(endpoint)`   | Define o endpoint para expor métricas Prometheus          | `GRAFTEL_PROMETHEUS_ENDPOINT`    | `""`                      |
| `WithResourceAttribute(key, value)`  | Adiciona um atributo ao resource                          | -                                | `{}`                      |
| `WithResourceAttributes(attrs)`      | Adiciona múltiplos atributos ao resource                  | -                                | `{}`                      |
| `WithMetricExportInterval(interval)` | Define o intervalo de exportação de métricas              | `GRAFTEL_METRIC_EXPORT_INTERVAL` | `30s`                     |
| `WithLogExportInterval(interval)`    | Define o intervalo de exportação de logs                  | `GRAFTEL_LOG_EXPORT_INTERVAL`    | `30s`                     |
| `WithExportTimeout(timeout)`         | Define o timeout para exportação                          | `GRAFTEL_EXPORT_TIMEOUT`         | `10s`                     |
| `WithInsecure(insecure)`             | Desabilita TLS (apenas para desenvolvimento)              | `GRAFTEL_INSECURE`               | `false`                   |

## 🔧 Configuração via Variáveis de Ambiente

A biblioteca suporta configuração completa via variáveis de ambiente, facilitando o deploy em diferentes ambientes sem alterar código.

### Ordem de Prioridade

As configurações são carregadas na seguinte ordem (maior para menor prioridade):

1. **Valores passados via métodos `With*`** (maior prioridade)
2. **Variáveis de ambiente `GRAFTEL_*`**
3. **Valores padrão**

### Variáveis de Ambiente Disponíveis

| Variável                         | Descrição                           | Exemplo                         |
| -------------------------------- | ----------------------------------- | ------------------------------- |
| `GRAFTEL_SERVICE_NAME`           | Nome do serviço                     | `meu-servico`                   |
| `GRAFTEL_SERVICE_VERSION`        | Versão do serviço                   | `1.0.0`                         |
| `GRAFTEL_OTLP_ENDPOINT`          | Endpoint OTLP                       | `https://otlp.example.com/otlp` |
| `GRAFTEL_API_KEY`                | Chave de API para autenticação      | `sua-chave-api`                 |
| `GRAFTEL_INSTANCE_ID`            | ID da instância                     | `instance-123`                  |
| `GRAFTEL_PROMETHEUS_ENDPOINT`    | Endpoint Prometheus                 | `:8080`                         |
| `GRAFTEL_INSECURE`               | Desabilitar TLS                     | `true` ou `false`               |
| `GRAFTEL_METRIC_EXPORT_INTERVAL` | Intervalo de exportação de métricas | `30s`                           |
| `GRAFTEL_LOG_EXPORT_INTERVAL`    | Intervalo de exportação de logs     | `30s`                           |
| `GRAFTEL_EXPORT_TIMEOUT`         | Timeout para exportação             | `10s`                           |

### Exemplo: Usando Apenas Variáveis de Ambiente

```go
package main

import (
    "context"
    "log"

    "github.com/CristianSsousa/graftel"
)

func main() {
    // Todas as configurações vêm das variáveis de ambiente GRAFTEL_*
    // Configure-as antes de executar:
    // export GRAFTEL_SERVICE_NAME="meu-servico"
    // export GRAFTEL_OTLP_ENDPOINT="https://otlp.example.com/otlp"
    // export GRAFTEL_API_KEY="sua-chave"

    config := graftel.NewConfig("") // ServiceName será lido de GRAFTEL_SERVICE_NAME

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

### Exemplo: Misturando ENV e With\*

```go
// Valores passados via With* têm prioridade sobre ENV
config := graftel.NewConfig("meu-servico"). // ServiceName explícito
    WithServiceVersion("1.0.0").            // Version explícita
    // OTLPEndpoint será lido de GRAFTEL_OTLP_ENDPOINT se não fornecido
    // APIKey será lido de GRAFTEL_API_KEY se não fornecido
```

## 🏷️ Atributos de Log Organizados

Todos os atributos customizados adicionados aos logs são automaticamente prefixados com `tags.` para melhor organização e estruturação dos metadados.

```go
logs.Info(ctx, "Requisição processada",
    attribute.String("method", "GET"),      // Vira "tags.method"
    attribute.String("path", "/api/users"), // Vira "tags.path"
    attribute.Int("status", 200),           // Vira "tags.status"
)
```

Os atributos do Resource do OpenTelemetry (como `process.pid`, `host.name`, `os.type`, etc.) não são prefixados, mantendo a compatibilidade com os padrões do OpenTelemetry.

## 🛡️ Resource Sanitizado

A biblioteca automaticamente remove campos sensíveis ou desnecessários do Resource OpenTelemetry:

**Campos removidos:**

-   `process.command_args` - Argumentos de linha de comando
-   `process.executable.path` - Caminho completo do executável
-   `process.executable.name` - Nome do executável
-   `process.command` - Comando completo
-   `process.owner` - Proprietário do processo

**Campos mantidos:**

-   `process.pid` - ID do processo (útil para debugging)
-   `process.runtime.*` - Informações sobre o runtime (Go version, etc.)
-   `host.name` - Nome do host
-   `os.type`, `os.description` - Informações do sistema operacional
-   `service.name`, `service.version` - Informações do serviço
-   `service.instance.id` - ID da instância (se configurado)

Isso reduz o volume de dados enviados e remove informações sensíveis dos logs.

## 📚 Exemplos

A biblioteca inclui exemplos completos na pasta `examples/`:

-   **`examples/basic/`** - Exemplo básico com métricas e logs usando endpoint local
-   **`examples/prometheus/`** - Exemplo com Prometheus para expor métricas
-   **`examples/grafana-cloud/`** - Exemplo usando variáveis de ambiente e autenticação

Para executar um exemplo:

```bash
# Exemplo básico (endpoint local)
cd examples/basic
go run main.go

# Exemplo com Prometheus
cd examples/prometheus
go run main.go
# Acesse http://localhost:8080/metrics

# Exemplo com configuração via variáveis de ambiente
cd examples/grafana-cloud
export GRAFTEL_SERVICE_NAME="meu-servico"
export GRAFTEL_OTLP_ENDPOINT="https://otlp.example.com/otlp"
export GRAFTEL_API_KEY="sua-chave-aqui"
export GRAFTEL_INSTANCE_ID="seu-instance-id"  # Opcional
go run main.go
```

### Exemplo: Uso com Diferentes Formatos de URL

```go
// Exemplo 1: URL completa com path e autenticação
config1 := graftel.NewConfig("servico-1").
    WithOTLPEndpoint("https://otlp.example.com/otlp").
    WithAPIKey("sua-chave").
    WithInstanceID("instance-123").
    WithInsecure(false)

// Exemplo 2: URL local sem path
config2 := graftel.NewConfig("servico-2").
    WithOTLPEndpoint("http://localhost:4318").
    WithInsecure(true)

// Exemplo 3: Apenas host:port
config3 := graftel.NewConfig("servico-3").
    WithOTLPEndpoint("localhost:4318").
    WithInsecure(true)

// Exemplo 4: Host:port com path customizado
config4 := graftel.NewConfig("servico-4").
    WithOTLPEndpoint("localhost:4318/v1/custom").
    WithInsecure(true)
```

## 🏗️ Estrutura do Projeto

```
.
├── client.go             # Cliente principal e inicialização
├── config.go             # Configuração com pattern builder
├── metrics.go            # Helpers para métricas
├── logs.go               # Helpers para logs
├── errors.go             # Erros customizados
├── client_test.go        # Testes unitários
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

-   Go 1.23 ou superior
-   OpenTelemetry SDK v1.38.0 ou superior

## 🔗 Dependências Principais

-   `go.opentelemetry.io/otel` - OpenTelemetry Go SDK
-   `go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp` - Exportador OTLP para métricas
-   `go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp` - Exportador OTLP para logs
-   `go.opentelemetry.io/otel/exporters/prometheus` - Exportador Prometheus

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

-   GitHub: [@CristianSsousa](https://github.com/CristianSsousa)
-   Repositório: [github.com/CristianSsousa/graftel](https://github.com/CristianSsousa/graftel)

## 🙏 Agradecimentos

-   [OpenTelemetry](https://opentelemetry.io/) pela excelente especificação e SDK
-   Comunidade Go por todas as ferramentas e bibliotecas incríveis

## 📖 Documentação Adicional

-   [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
-   [Go Documentation](https://go.dev/doc/)

---

⭐ Se este projeto foi útil para você, considere dar uma estrela no repositório!
