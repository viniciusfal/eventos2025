package middleware

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"eventos-backend/internal/interfaces/http/responses"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ErrorHandlerMiddleware configura middleware de tratamento de erros e recovery
func ErrorHandlerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		// Log do panic
		if err, ok := recovered.(error); ok {
			logger.Error("Panic recovered",
				zap.Error(err),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("ip", c.ClientIP()),
				zap.String("user_agent", c.Request.UserAgent()),
				zap.String("stack", string(debug.Stack())),
			)
		} else {
			logger.Error("Panic recovered",
				zap.Any("recovered", recovered),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("ip", c.ClientIP()),
				zap.String("user_agent", c.Request.UserAgent()),
				zap.String("stack", string(debug.Stack())),
			)
		}

		// Responder com erro interno
		responses.InternalServerError(c, "An unexpected error occurred")
		c.Abort()
	})
}

// ErrorResponseMiddleware middleware para processar erros de handlers
func ErrorResponseMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Processar erros que foram adicionados ao contexto
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			// Log do erro
			logger.Error("Handler error",
				zap.Error(err.Err),
				zap.String("type", string(err.Type)),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("ip", c.ClientIP()),
			)

			// Se ainda não foi enviada uma resposta, enviar erro genérico
			if !c.Writer.Written() {
				responses.InternalServerError(c, "An error occurred while processing your request")
			}
		}
	}
}

// DomainErrorHandlerMiddleware middleware específico para erros de domínio
func DomainErrorHandlerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Verificar se há erros de domínio no contexto
		if domainErr, exists := c.Get("domain_error"); exists {
			if err, ok := domainErr.(error); ok {
				// Log do erro de domínio
				logger.Warn("Domain error",
					zap.Error(err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("ip", c.ClientIP()),
				)

				// Mapear erro de domínio para resposta HTTP
				mapDomainErrorToHTTP(c, err)
			}
		}
	}
}

// mapDomainErrorToHTTP mapeia erros de domínio para respostas HTTP apropriadas
func mapDomainErrorToHTTP(c *gin.Context, err error) {
	// TODO: Implementar mapeamento específico baseado nos tipos de erro de domínio
	// Por enquanto, usar mapeamento genérico baseado na mensagem

	errorMsg := err.Error()

	switch {
	case contains(errorMsg, "not found"):
		responses.NotFound(c, err.Error())
	case contains(errorMsg, "already exists"):
		responses.Conflict(c, err.Error(), nil)
	case contains(errorMsg, "invalid"):
		responses.BadRequest(c, err.Error(), nil)
	case contains(errorMsg, "unauthorized"):
		responses.Unauthorized(c, err.Error())
	case contains(errorMsg, "forbidden"):
		responses.Forbidden(c, err.Error())
	case contains(errorMsg, "validation"):
		responses.UnprocessableEntity(c, err.Error(), nil)
	default:
		responses.InternalServerError(c, "An error occurred while processing your request")
	}
}

// contains verifica se uma string contém uma substring (case insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			fmt.Sprintf("%s", s) != s ||
			containsIgnoreCase(s, substr))
}

// containsIgnoreCase verifica substring ignorando case
func containsIgnoreCase(s, substr string) bool {
	s = fmt.Sprintf("%s", s)
	substr = fmt.Sprintf("%s", substr)

	if len(s) < len(substr) {
		return false
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if toLower(s[i+j]) != toLower(substr[j]) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// toLower converte caractere para minúsculo
func toLower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + 32
	}
	return c
}

// TimeoutMiddleware middleware para timeout de requisições
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Criar contexto com timeout
		ctx := c.Request.Context()

		// Definir deadline
		if timeout > 0 {
			var cancel func()
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()

			// Atualizar contexto da requisição
			c.Request = c.Request.WithContext(ctx)
		}

		c.Next()
	}
}
