package middleware

import (
	"net/http"
	"sync"
	"time"

	"eventos-backend/internal/interfaces/http/responses"

	"github.com/gin-gonic/gin"
)

// RateLimiter representa um limitador de taxa
type RateLimiter struct {
	visitors map[string]*Visitor
	mu       sync.RWMutex
	rate     int           // requests per minute
	burst    int           // burst capacity
	cleanup  time.Duration // cleanup interval
}

// Visitor representa um visitante com seu bucket de tokens
type Visitor struct {
	tokens   int
	lastSeen time.Time
	mu       sync.Mutex
}

// NewRateLimiter cria um novo rate limiter
func NewRateLimiter(rate, burst int, cleanup time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*Visitor),
		rate:     rate,
		burst:    burst,
		cleanup:  cleanup,
	}

	// Iniciar limpeza periódica
	go rl.cleanupVisitors()

	return rl
}

// Allow verifica se uma requisição é permitida para o IP dado
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	visitor, exists := rl.visitors[ip]
	if !exists {
		visitor = &Visitor{
			tokens:   rl.burst,
			lastSeen: time.Now(),
		}
		rl.visitors[ip] = visitor
	}

	visitor.mu.Lock()
	defer visitor.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(visitor.lastSeen)
	visitor.lastSeen = now

	// Adicionar tokens baseado no tempo decorrido
	tokensToAdd := int(elapsed.Minutes() * float64(rl.rate))
	visitor.tokens += tokensToAdd

	// Limitar ao burst máximo
	if visitor.tokens > rl.burst {
		visitor.tokens = rl.burst
	}

	// Verificar se há tokens disponíveis
	if visitor.tokens > 0 {
		visitor.tokens--
		return true
	}

	return false
}

// cleanupVisitors remove visitantes inativos
func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			cutoff := time.Now().Add(-rl.cleanup)

			for ip, visitor := range rl.visitors {
				visitor.mu.Lock()
				if visitor.lastSeen.Before(cutoff) {
					delete(rl.visitors, ip)
				}
				visitor.mu.Unlock()
			}
			rl.mu.Unlock()
		}
	}
}

// RateLimiterMiddleware cria middleware de rate limiting
func RateLimiterMiddleware() gin.HandlerFunc {
	// Configuração padrão: 100 requests/minuto, burst de 10, cleanup a cada 5 minutos
	limiter := NewRateLimiter(100, 10, 5*time.Minute)

	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !limiter.Allow(ip) {
			responses.Error(c, http.StatusTooManyRequests, "Rate limit exceeded", "RATE_LIMIT_EXCEEDED", map[string]interface{}{
				"ip":          ip,
				"retry_after": "60s",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimiterMiddlewareWithConfig cria middleware de rate limiting com configuração personalizada
func RateLimiterMiddlewareWithConfig(rate, burst int, cleanup time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, burst, cleanup)

	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !limiter.Allow(ip) {
			retryAfter := 60 / rate // segundos até próximo token
			responses.Error(c, http.StatusTooManyRequests, "Rate limit exceeded", "RATE_LIMIT_EXCEEDED", map[string]interface{}{
				"ip":          ip,
				"retry_after": retryAfter,
				"rate_limit": map[string]interface{}{
					"requests_per_minute": rate,
					"burst_capacity":      burst,
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// UserRateLimiterMiddleware rate limiting baseado em usuário autenticado
func UserRateLimiterMiddleware(rate, burst int) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, burst, 10*time.Minute)

	return func(c *gin.Context) {
		// Tentar obter ID do usuário do contexto (após autenticação)
		userID, exists := c.Get("user_id")
		if !exists {
			// Se não há usuário autenticado, usar IP
			userID = c.ClientIP()
		}

		key := userID.(string)

		if !limiter.Allow(key) {
			retryAfter := 60 / rate
			responses.Error(c, http.StatusTooManyRequests, "Rate limit exceeded", "RATE_LIMIT_EXCEEDED", map[string]interface{}{
				"user_id":     key,
				"retry_after": retryAfter,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// EndpointRateLimiterMiddleware rate limiting específico por endpoint
func EndpointRateLimiterMiddleware(endpointLimits map[string]RateLimit) gin.HandlerFunc {
	limiters := make(map[string]*RateLimiter)

	// Criar limitadores para cada endpoint
	for endpoint, limit := range endpointLimits {
		limiters[endpoint] = NewRateLimiter(limit.Rate, limit.Burst, 5*time.Minute)
	}

	return func(c *gin.Context) {
		endpoint := c.Request.Method + " " + c.FullPath()

		// Verificar se há limitador para este endpoint
		limiter, exists := limiters[endpoint]
		if !exists {
			// Se não há limitador específico, continuar
			c.Next()
			return
		}

		ip := c.ClientIP()
		key := endpoint + ":" + ip

		if !limiter.Allow(key) {
			limit := endpointLimits[endpoint]
			retryAfter := 60 / limit.Rate

			responses.Error(c, http.StatusTooManyRequests, "Endpoint rate limit exceeded", "ENDPOINT_RATE_LIMIT_EXCEEDED", map[string]interface{}{
				"endpoint":    endpoint,
				"ip":          ip,
				"retry_after": retryAfter,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimit representa configuração de rate limiting
type RateLimit struct {
	Rate  int // requests per minute
	Burst int // burst capacity
}
