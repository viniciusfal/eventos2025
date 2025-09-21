package middleware

import (
	"net/http"
	"time"

	"eventos-backend/internal/infrastructure/monitoring"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// TracingMiddleware adiciona tracing OpenTelemetry às requisições HTTP
func TracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Iniciar span para a requisição HTTP
		ctx, span := monitoring.StartSpan(
			c.Request.Context(),
			"http_request",
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		// Adicionar atributos básicos da requisição
		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.Path),
			attribute.String("http.user_agent", c.Request.UserAgent()),
			attribute.String("http.client_ip", c.ClientIP()),
		)

		// Adicionar ID do tenant se disponível
		if tenantID := c.GetHeader("X-Tenant-ID"); tenantID != "" {
			span.SetAttributes(attribute.String("tenant.id", tenantID))
		}

		// Adicionar ID do usuário se disponível
		if userID := c.GetHeader("X-User-ID"); userID != "" {
			span.SetAttributes(attribute.String("user.id", userID))
		}

		// Armazenar span no contexto da requisição
		c.Request = c.Request.WithContext(ctx)

		// Registrar tempo de início
		startTime := time.Now()

		// Continuar processamento da requisição
		c.Next()

		// Registrar duração da requisição
		duration := time.Since(startTime)
		span.SetAttributes(attribute.Int64("http.duration_ms", duration.Milliseconds()))

		// Registrar status da resposta
		statusCode := c.Writer.Status()
		span.SetAttributes(attribute.Int("http.status_code", statusCode))

		// Se houve erro, marcar span como erro
		if statusCode >= 400 {
			span.RecordError(nil)
			span.SetStatus(codes.Error, http.StatusText(statusCode))
		} else {
			span.SetStatus(codes.Ok, "OK")
		}

		// Adicionar informações de resposta
		span.SetAttributes(
			attribute.Int64("http.response_size", int64(c.Writer.Size())),
			attribute.String("http.status", http.StatusText(statusCode)),
		)

		// Registrar métricas HTTP
		monitoring.RecordHTTPRequest(
			c.Request.Method,
			c.Request.URL.Path,
			http.StatusText(statusCode),
			duration,
		)
	}
}
