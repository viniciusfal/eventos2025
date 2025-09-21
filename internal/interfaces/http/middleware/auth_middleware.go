package middleware

import (
	"net/http"
	"strings"

	jwtService "eventos-backend/internal/infrastructure/auth/jwt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthMiddleware representa o middleware de autenticação
type AuthMiddleware struct {
	jwtService jwtService.Service
	logger     *zap.Logger
}

// NewAuthMiddleware cria uma nova instância do middleware de autenticação
func NewAuthMiddleware(jwtService jwtService.Service, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
		logger:     logger,
	}
}

// RequireAuth middleware que exige autenticação
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extrair token do header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			m.logger.Warn("Missing authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		// Verificar formato Bearer token
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			m.logger.Warn("Invalid authorization header format")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		// Validar token
		claims, err := m.jwtService.ValidateToken(tokenString)
		if err != nil {
			m.logger.Warn("Invalid token", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			c.Abort()
			return
		}

		// Adicionar informações do usuário no contexto
		c.Set("user_id", claims.UserID)
		c.Set("tenant_id", claims.TenantID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("claims", claims)

		m.logger.Debug("User authenticated",
			zap.String("user_id", claims.UserID),
			zap.String("tenant_id", claims.TenantID),
			zap.String("username", claims.Username),
		)

		c.Next()
	}
}

// OptionalAuth middleware que permite autenticação opcional
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extrair token do header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// Sem token, continuar sem autenticação
			c.Next()
			return
		}

		// Verificar formato Bearer token
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			// Token mal formatado, continuar sem autenticação
			c.Next()
			return
		}

		tokenString := tokenParts[1]

		// Validar token
		claims, err := m.jwtService.ValidateToken(tokenString)
		if err != nil {
			// Token inválido, continuar sem autenticação
			m.logger.Debug("Invalid optional token", zap.Error(err))
			c.Next()
			return
		}

		// Adicionar informações do usuário no contexto
		c.Set("user_id", claims.UserID)
		c.Set("tenant_id", claims.TenantID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("claims", claims)

		m.logger.Debug("User optionally authenticated",
			zap.String("user_id", claims.UserID),
			zap.String("tenant_id", claims.TenantID),
			zap.String("username", claims.Username),
		)

		c.Next()
	}
}

// RequireTenant middleware que exige que o usuário pertença a um tenant específico
func (m *AuthMiddleware) RequireTenant(tenantID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userTenantID, exists := c.Get("tenant_id")
		if !exists {
			m.logger.Warn("No tenant information in context")
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied",
			})
			c.Abort()
			return
		}

		if userTenantID != tenantID {
			m.logger.Warn("Tenant mismatch",
				zap.String("required_tenant", tenantID),
				zap.String("user_tenant", userTenantID.(string)),
			)
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied to this tenant",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserID extrai o ID do usuário do contexto
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}
	return userID.(string), true
}

// GetTenantID extrai o ID do tenant do contexto
func GetTenantID(c *gin.Context) (string, bool) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		return "", false
	}
	return tenantID.(string), true
}

// GetUsername extrai o username do contexto
func GetUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}
	return username.(string), true
}

// GetEmail extrai o email do contexto
func GetEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get("email")
	if !exists {
		return "", false
	}
	return email.(string), true
}

// GetClaims extrai as claims completas do contexto
func GetClaims(c *gin.Context) (*jwtService.Claims, bool) {
	claims, exists := c.Get("claims")
	if !exists {
		return nil, false
	}
	return claims.(*jwtService.Claims), true
}
