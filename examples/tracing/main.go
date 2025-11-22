package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/CristianSsousa/graftel/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	config := graftel.NewConfig("tracing-example").
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
	defer func(client graftel.Client, ctx context.Context) {
		err := client.Shutdown(ctx)
		if err != nil {
			log.Fatalf("Falha ao finalizar: %v", err)
		}
	}(client, ctx)

	tracing := client.NewTracingHelper("tracing-example")
	logs := client.NewLogsHelper("tracing-example/logs")

	ctx, span := tracing.StartSpanWithTags(ctx, "main-operation",
		attribute.String("operation", "exemplo-tracing"),
		attribute.String("version", "1.0.0"),
	)
	defer span.End()

	logs.Info(ctx, "Iniciando exemplo de tracing")

	err = processarDados(ctx, tracing, logs)
	if err != nil {
		tracing.SetSpanError(ctx, err, attribute.String("retry", "false"))
		logs.Error(ctx, "Erro ao processar dados", attribute.String("error", err.Error()))
	}

	tracing.AddSpanTags(ctx, attribute.String("status", "concluido"))
	tracing.SetSpanStatus(ctx, codes.Ok, "Operação concluída com sucesso")

	traceID := tracing.GetTraceID(ctx)
	spanID := tracing.GetSpanID(ctx)
	fmt.Printf("\nTrace ID: %s\nSpan ID: %s\n", traceID, spanID)

	fmt.Println("\nExemplo de tracing concluído!")
}

func processarDados(ctx context.Context, tracing graftel.TracingHelper, logs graftel.LogsHelper) error {
	ctx, span := tracing.StartSpan(ctx, "processar-dados",
		trace.WithAttributes(
			attribute.String("tipo", "batch"),
			attribute.Int("items", 10),
		),
	)
	defer span.End()

	logs.Info(ctx, "Processando dados", attribute.Int("total", 10))

	for i := 0; i < 3; i++ {
		err := processarItem(ctx, tracing, logs, i)
		if err != nil {
			return err
		}
	}

	tracing.AddSpanTags(ctx, attribute.String("processed", "3"))
	return nil
}

func processarItem(ctx context.Context, tracing graftel.TracingHelper, logs graftel.LogsHelper, index int) error {
	return tracing.WithSpan(ctx, fmt.Sprintf("processar-item-%d", index), func(ctx context.Context) error {
		logs.Info(ctx, "Processando item",
			attribute.Int("index", index),
			attribute.String("status", "em-processamento"),
		)

		time.Sleep(50 * time.Millisecond)

		if index == 1 {
			return errors.New("erro simulado no item 1")
		}

		logs.Info(ctx, "Item processado",
			attribute.Int("index", index),
			attribute.String("status", "concluido"),
		)

		return nil
	}, attribute.Int("item_index", index))
}
