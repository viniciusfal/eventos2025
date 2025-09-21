package monitoring

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// HealthStatus representa o status de saúde do sistema
type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Version   string            `json:"version"`
	Services  map[string]string `json:"services"`
}

// HealthCheckHandler implementa endpoints de health check
type HealthCheckHandler struct{}

// NewHealthCheckHandler cria uma nova instância do handler de health check
func NewHealthCheckHandler() *HealthCheckHandler {
	return &HealthCheckHandler{}
}

// HealthCheck endpoint básico de saúde
func (h *HealthCheckHandler) HealthCheck(c *gin.Context) {
	health := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0",
		Services: map[string]string{
			"database": "healthy",
			"redis":    "healthy",
			"rabbitmq": "healthy",
			"api":      "healthy",
		},
	}

	c.JSON(http.StatusOK, health)
}

// ReadinessCheck endpoint para verificar se o serviço está pronto para receber tráfego
func (h *HealthCheckHandler) ReadinessCheck(c *gin.Context) {
	// TODO: Implementar verificações reais de dependências
	// Por enquanto, sempre retorna healthy
	readiness := HealthStatus{
		Status:    "ready",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0",
		Services: map[string]string{
			"database": "ready",
			"redis":    "ready",
			"rabbitmq": "ready",
			"api":      "ready",
		},
	}

	c.JSON(http.StatusOK, readiness)
}

// LivenessCheck endpoint para verificar se o serviço está vivo
func (h *HealthCheckHandler) LivenessCheck(c *gin.Context) {
	liveness := map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"uptime":    "unknown", // TODO: implementar tracking de uptime
	}

	c.JSON(http.StatusOK, liveness)
}

// Metrics endpoint para expor métricas do Prometheus
func (h *HealthCheckHandler) Metrics(c *gin.Context) {
	// Atualizar métricas do sistema
	UpdateSystemMetrics()

	// Delegar para o handler padrão do Prometheus
	registry := GetMetricsRegistry()
	promHandler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	promHandler.ServeHTTP(c.Writer, c.Request)
}
