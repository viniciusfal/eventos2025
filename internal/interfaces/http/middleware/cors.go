package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORSMiddleware configura middleware de CORS
func CORSMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Configurar headers CORS
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS,PATCH")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization,X-Requested-With,Accept,Origin,Cache-Control,X-File-Name")
		c.Header("Access-Control-Expose-Headers", "Content-Length,Access-Control-Allow-Origin,Access-Control-Allow-Headers,Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		// Responder a requisições OPTIONS (preflight)
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
}

// CORSMiddlewareWithConfig configura middleware de CORS com configurações específicas
func CORSMiddlewareWithConfig(allowedOrigins []string, allowedMethods []string, allowedHeaders []string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Verificar se a origem é permitida
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		// Configurar métodos permitidos
		methods := "GET,POST,PUT,DELETE,OPTIONS,PATCH"
		if len(allowedMethods) > 0 {
			methods = ""
			for i, method := range allowedMethods {
				if i > 0 {
					methods += ","
				}
				methods += method
			}
		}
		c.Header("Access-Control-Allow-Methods", methods)

		// Configurar headers permitidos
		headers := "Content-Type,Authorization,X-Requested-With,Accept,Origin,Cache-Control,X-File-Name"
		if len(allowedHeaders) > 0 {
			headers = ""
			for i, header := range allowedHeaders {
				if i > 0 {
					headers += ","
				}
				headers += header
			}
		}
		c.Header("Access-Control-Allow-Headers", headers)

		c.Header("Access-Control-Expose-Headers", "Content-Length,Access-Control-Allow-Origin,Access-Control-Allow-Headers,Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		// Responder a requisições OPTIONS (preflight)
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
}
