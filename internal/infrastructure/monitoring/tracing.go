package monitoring

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// Tracer representa o tracer do OpenTelemetry
var Tracer trace.Tracer

// InitTracing inicializa o tracing com OpenTelemetry
func InitTracing(serviceName, serviceVersion string) (*sdktrace.TracerProvider, error) {
	// Criar resource com informações do serviço
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
			semconv.ServiceInstanceID("eventos-api-instance"),
			attribute.String("environment", getEnvOrDefault("ENV", "development")),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Configurar OTLP HTTP exporter
	otlpEndpoint := getEnvOrDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318/v1/traces")
	exporter, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint(otlpEndpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Criar tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exporter),
	)

	// Configurar como tracer provider global
	otel.SetTracerProvider(tp)

	// Criar tracer global
	Tracer = otel.Tracer(serviceName)

	return tp, nil
}

// ShutdownTracing finaliza o tracing
func ShutdownTracing(ctx context.Context, tp *sdktrace.TracerProvider) error {
	if tp != nil {
		return tp.Shutdown(ctx)
	}
	return nil
}

// TracingMiddleware cria um middleware para tracing HTTP
func TracingMiddleware() func(ctx context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		// Criar span para a requisição
		span := trace.SpanFromContext(ctx)
		if span == nil {
			_, span = Tracer.Start(ctx, "http_request")
			defer span.End()
		}
		return trace.ContextWithSpan(ctx, span)
	}
}

// StartSpan inicia um novo span
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return Tracer.Start(ctx, name, opts...)
}

// AddSpanAttributes adiciona atributos a um span
func AddSpanAttributes(span trace.Span, attrs ...attribute.KeyValue) {
	span.SetAttributes(attrs...)
}

// SetSpanError marca um span como erro
func SetSpanError(span trace.Span, err error) {
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// SetSpanStatus define o status de um span
func SetSpanStatus(span trace.Span, code codes.Code, description string) {
	span.SetStatus(code, description)
}

// AddSpanEvent adiciona um evento a um span
func AddSpanEvent(span trace.Span, name string, attrs ...attribute.KeyValue) {
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// getEnvOrDefault retorna o valor de uma variável de ambiente ou um valor padrão
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
