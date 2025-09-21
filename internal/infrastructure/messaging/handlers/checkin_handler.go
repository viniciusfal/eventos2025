package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"eventos-backend/internal/infrastructure/cache"
	"eventos-backend/internal/infrastructure/messaging/rabbitmq"

	"go.uber.org/zap"
)

// CheckinEventHandler processa eventos relacionados a check-ins
type CheckinEventHandler struct {
	logger       *zap.Logger
	cacheService *cache.CacheService
	keyBuilder   cache.KeyBuilder
}

// NewCheckinEventHandler cria uma nova instância do handler
func NewCheckinEventHandler(logger *zap.Logger, cacheService *cache.CacheService, keyBuilder cache.KeyBuilder) *CheckinEventHandler {
	return &CheckinEventHandler{
		logger:       logger,
		cacheService: cacheService,
		keyBuilder:   keyBuilder,
	}
}

// Handle processa uma mensagem de evento de check-in
func (h *CheckinEventHandler) Handle(ctx context.Context, message *rabbitmq.Message) error {
	h.logger.Info("Processing checkin event",
		zap.String("message_id", message.ID),
		zap.String("message_type", message.Type),
	)

	switch message.Type {
	case rabbitmq.MessageTypeCheckinPerformed:
		return h.handleCheckinPerformed(ctx, message)
	case rabbitmq.MessageTypeCheckinValidated:
		return h.handleCheckinValidated(ctx, message)
	case rabbitmq.MessageTypeCheckinInvalid:
		return h.handleCheckinInvalid(ctx, message)
	default:
		return fmt.Errorf("unsupported message type: %s", message.Type)
	}
}

// CanHandle verifica se o handler pode processar o tipo de mensagem
func (h *CheckinEventHandler) CanHandle(messageType string) bool {
	supportedTypes := []string{
		rabbitmq.MessageTypeCheckinPerformed,
		rabbitmq.MessageTypeCheckinValidated,
		rabbitmq.MessageTypeCheckinInvalid,
	}

	for _, supportedType := range supportedTypes {
		if messageType == supportedType {
			return true
		}
	}

	return false
}

// GetName retorna o nome do handler
func (h *CheckinEventHandler) GetName() string {
	return "CheckinEventHandler"
}

// handleCheckinPerformed processa evento de check-in realizado
func (h *CheckinEventHandler) handleCheckinPerformed(ctx context.Context, message *rabbitmq.Message) error {
	// Parse do payload
	var payload rabbitmq.CheckinEventPayload
	payloadBytes, err := json.Marshal(message.Body)
	if err != nil {
		return fmt.Errorf("failed to marshal message body: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal checkin payload: %w", err)
	}

	h.logger.Info("Checkin performed event received",
		zap.String("checkin_id", payload.CheckinID),
		zap.String("employee_id", payload.EmployeeID),
		zap.String("event_id", payload.EventID),
		zap.String("method", payload.Method),
		zap.Bool("is_valid", payload.IsValid),
	)

	// Invalidar caches relacionados
	if h.cacheService != nil {
		if err := h.invalidateRelatedCaches(ctx, payload); err != nil {
			h.logger.Warn("Failed to invalidate caches", zap.Error(err))
			// Não falhar o processamento por causa do cache
		}
	}

	// Aqui poderíamos:
	// - Enviar notificações
	// - Atualizar estatísticas
	// - Integrar com sistemas externos
	// - Enviar emails/SMS
	// - Registrar auditoria

	h.logger.Info("Checkin performed event processed successfully",
		zap.String("checkin_id", payload.CheckinID),
	)

	return nil
}

// handleCheckinValidated processa evento de check-in validado
func (h *CheckinEventHandler) handleCheckinValidated(ctx context.Context, message *rabbitmq.Message) error {
	var payload rabbitmq.CheckinEventPayload
	payloadBytes, err := json.Marshal(message.Body)
	if err != nil {
		return fmt.Errorf("failed to marshal message body: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal checkin payload: %w", err)
	}

	h.logger.Info("Checkin validated event received",
		zap.String("checkin_id", payload.CheckinID),
		zap.String("employee_id", payload.EmployeeID),
		zap.Bool("is_valid", payload.IsValid),
	)

	// Invalidar caches de estatísticas
	if h.cacheService != nil {
		statsKey := h.keyBuilder.BuildKeyWithTenant(payload.TenantID, "stats", "checkin")
		if err := h.cacheService.InvalidatePattern(ctx, "default", statsKey+"*"); err != nil {
			h.logger.Warn("Failed to invalidate stats cache", zap.Error(err))
		}
	}

	// Processar validação
	// - Atualizar métricas de qualidade
	// - Notificar supervisores se necessário
	// - Atualizar dashboards em tempo real

	return nil
}

// handleCheckinInvalid processa evento de check-in inválido
func (h *CheckinEventHandler) handleCheckinInvalid(ctx context.Context, message *rabbitmq.Message) error {
	var payload rabbitmq.CheckinEventPayload
	payloadBytes, err := json.Marshal(message.Body)
	if err != nil {
		return fmt.Errorf("failed to marshal message body: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal checkin payload: %w", err)
	}

	h.logger.Warn("Invalid checkin event received",
		zap.String("checkin_id", payload.CheckinID),
		zap.String("employee_id", payload.EmployeeID),
		zap.String("event_id", payload.EventID),
		zap.String("method", payload.Method),
	)

	// Processar check-in inválido
	// - Notificar administradores
	// - Registrar tentativa suspeita
	// - Atualizar alertas de segurança
	// - Possivelmente bloquear funcionário temporariamente

	return nil
}

// invalidateRelatedCaches invalida caches relacionados ao check-in
func (h *CheckinEventHandler) invalidateRelatedCaches(ctx context.Context, payload rabbitmq.CheckinEventPayload) error {
	// Type assertion para acessar métodos específicos do DefaultKeyBuilder
	defaultKeyBuilder, ok := h.keyBuilder.(*cache.DefaultKeyBuilder)
	if !ok {
		// Fallback para usar métodos genéricos
		patterns := []string{
			// Cache do check-in específico
			h.keyBuilder.BuildKeyWithTenant(payload.TenantID, "checkin", payload.CheckinID),

			// Cache de listas de check-ins do funcionário
			h.keyBuilder.BuildKeyWithTenant(payload.TenantID, "list", "checkin", "employee", payload.EmployeeID) + "*",

			// Cache de listas de check-ins do evento
			h.keyBuilder.BuildKeyWithTenant(payload.TenantID, "list", "checkin", "event", payload.EventID) + "*",

			// Cache de estatísticas
			h.keyBuilder.BuildKeyWithTenant(payload.TenantID, "stats", "checkin") + "*",

			// Cache de check-ins recentes
			h.keyBuilder.BuildKeyWithTenant(payload.TenantID, "recent", "checkin") + "*",
		}

		for _, pattern := range patterns {
			if err := h.cacheService.InvalidatePattern(ctx, "default", pattern); err != nil {
				h.logger.Error("Failed to invalidate cache pattern",
					zap.String("pattern", pattern),
					zap.Error(err),
				)
				return err
			}
		}
	}

	patterns := []string{
		// Cache do check-in específico
		defaultKeyBuilder.CheckinKey(payload.TenantID, payload.CheckinID),

		// Cache de listas de check-ins do funcionário
		defaultKeyBuilder.BuildKeyWithTenant(payload.TenantID, "list", "checkin", "employee", payload.EmployeeID) + "*",

		// Cache de listas de check-ins do evento
		defaultKeyBuilder.BuildKeyWithTenant(payload.TenantID, "list", "checkin", "event", payload.EventID) + "*",

		// Cache de estatísticas
		defaultKeyBuilder.StatsKey(payload.TenantID, "checkin") + "*",

		// Cache de check-ins recentes
		defaultKeyBuilder.BuildKeyWithTenant(payload.TenantID, "recent", "checkin") + "*",
	}

	for _, pattern := range patterns {
		if err := h.cacheService.InvalidatePattern(ctx, "default", pattern); err != nil {
			h.logger.Error("Failed to invalidate cache pattern",
				zap.String("pattern", pattern),
				zap.Error(err),
			)
			return err
		}
	}

	h.logger.Debug("Related caches invalidated successfully",
		zap.String("checkin_id", payload.CheckinID),
		zap.Int("patterns", len(patterns)),
	)

	return nil
}
