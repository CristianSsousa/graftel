package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/CristianSsousa/graftel/v2"
	"go.opentelemetry.io/otel/attribute"
)

func main() {
	config := graftel.NewConfig("meu-test-log").
		WithServiceVersion("1.0.0").
		WithOTLPEndpoint("http://localhost:4318").
		WithInsecure(true).
		WithResourceAttribute("environment", "development")

	client, err := graftel.NewClient(config)
	if err != nil {
		log.Fatalf("Falha ao criar cliente: %v", err)
	}

	ctx := context.Background()
	if err := client.Initialize(ctx); err != nil {
		log.Fatalf("Falha ao inicializar: %v", err)
	}
	defer func() {
		if err := client.Shutdown(ctx); err != nil {
			log.Printf("Erro ao encerrar cliente: %v", err)
		}
	}()

	logs := client.NewLogsHelper("context-example/logs")

	ctx = graftel.WithTags(ctx,
		attribute.String("user_id", "12345"),
		attribute.String("session_id", "sess-abc-123"),
		attribute.String("request_id", "req-xyz-789"),
	)

	logger := graftel.NewContextLogger(logs, ctx)
	logger.Info("Requisição recebida")

	processarRequisicao(ctx, logs)

	logger.Info("Requisição finalizada",
		attribute.String("status", "sucesso"),
	)

	fmt.Println("\nExemplo de context helpers concluído!")
}

func processarRequisicao(ctx context.Context, logs graftel.LogsHelper) {
	ctx = graftel.MergeContextTags(ctx,
		attribute.String("function", "processarRequisicao"),
		attribute.String("step", "inicio"),
	)

	logger := graftel.NewContextLogger(logs, ctx)
	logger.Info("Processando requisição")

	validarDados(ctx, logs)
	processarDados(ctx, logs)
	enviarResposta(ctx, logs)
}

func validarDados(ctx context.Context, logs graftel.LogsHelper) {
	ctx = graftel.WithTags(ctx, attribute.String("step", "validacao"))
	logger := graftel.NewContextLogger(logs, ctx).
		WithTags(attribute.String("validation_type", "schema"))

	logger.Info("Validando dados")

	time.Sleep(10 * time.Millisecond)

	logger.Info("Validação concluída",
		attribute.Bool("valid", true),
	)
}

func processarDados(ctx context.Context, logs graftel.LogsHelper) {
	ctx = graftel.WithTags(ctx, attribute.String("step", "processamento"))
	logger := graftel.NewContextLogger(logs, ctx)

	logger.Info("Processando dados")

	time.Sleep(20 * time.Millisecond)

	tags := graftel.GetTagsFromContext(ctx)
	fmt.Printf("\nTags no contexto: %d tags\n", len(tags))
	for _, tag := range tags {
		fmt.Printf("  - %s: %v\n", string(tag.Key), tag.Value.AsString())
	}

	logger.Info("Processamento concluído",
		attribute.Int("items_processed", 10),
	)
}

func enviarResposta(ctx context.Context, logs graftel.LogsHelper) {
	ctx = graftel.WithTags(ctx, attribute.String("step", "resposta"))
	logger := graftel.NewContextLogger(logs, ctx)

	logger.Info("Enviando resposta",
		attribute.Int("status_code", 200),
		attribute.String("content_type", "application/json"),
	)
}
