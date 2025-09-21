package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LoggingMiddleware configura middleware de logging estruturado
func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Não retornar string pois usamos logging estruturado
		return ""
	})
}

// StructuredLoggingMiddleware configura middleware de logging estruturado personalizado
func StructuredLoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Processar requisição
		c.Next()

		// Calcular latência
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		userAgent := c.Request.UserAgent()
		referer := c.Request.Referer()

		// Construir path completo
		if raw != "" {
			path = path + "?" + raw
		}

		// Determinar nível de log baseado no status code
		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.String("ip", clientIP),
			zap.Duration("latency", latency),
			zap.String("user_agent", userAgent),
			zap.String("referer", referer),
			zap.Int("response_size", c.Writer.Size()),
		}

		// Adicionar informações de erro se houver
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}

		// Log baseado no status code
		switch {
		case statusCode >= 500:
			logger.Error("HTTP Request - Server Error", fields...)
		case statusCode >= 400:
			logger.Warn("HTTP Request - Client Error", fields...)
		case statusCode >= 300:
			logger.Info("HTTP Request - Redirect", fields...)
		default:
			logger.Info("HTTP Request", fields...)
		}
	}
}

// RequestIDMiddleware adiciona um ID único para cada requisição
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Gerar ID da requisição (pode usar UUID se necessário)
		requestID := generateRequestID()

		// Adicionar ao contexto e header de resposta
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// LoggingWithRequestIDMiddleware combina logging com request ID
func LoggingWithRequestIDMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Processar requisição
		c.Next()

		// Calcular latência
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		userAgent := c.Request.UserAgent()
		referer := c.Request.Referer()

		// Construir path completo
		if raw != "" {
			path = path + "?" + raw
		}

		// Obter request ID do contexto
		requestID, exists := c.Get("request_id")
		if !exists {
			requestID = "unknown"
		}

		// Campos de log
		fields := []zap.Field{
			zap.String("request_id", requestID.(string)),
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.String("ip", clientIP),
			zap.Duration("latency", latency),
			zap.String("user_agent", userAgent),
			zap.String("referer", referer),
			zap.Int("response_size", c.Writer.Size()),
		}

		// Adicionar informações de erro se houver
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}

		// Log baseado no status code
		switch {
		case statusCode >= 500:
			logger.Error("HTTP Request - Server Error", fields...)
		case statusCode >= 400:
			logger.Warn("HTTP Request - Client Error", fields...)
		case statusCode >= 300:
			logger.Info("HTTP Request - Redirect", fields...)
		default:
			logger.Info("HTTP Request", fields...)
		}
	}
}

// generateRequestID gera um ID único para a requisição
func generateRequestID() string {
	// Implementação simples usando timestamp + random
	// Em produção, considere usar UUID
	return time.Now().Format("20060102150405") + "-" + randomString(6)
}

// randomString gera uma string aleatória
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
