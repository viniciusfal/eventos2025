package rabbitmq

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// Publisher gerencia a publicação de mensagens
type Publisher struct {
	client *Client
	logger *zap.Logger
	config PublisherConfig
}

// PublisherConfig configurações do publisher
type PublisherConfig struct {
	DefaultExchange string
	DefaultTimeout  time.Duration
	MaxRetries      int
	RetryDelay      time.Duration
}

// NewPublisher cria uma nova instância do publisher
func NewPublisher(client *Client, config PublisherConfig, logger *zap.Logger) *Publisher {
	if config.DefaultTimeout == 0 {
		config.DefaultTimeout = 30 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 1 * time.Second
	}

	return &Publisher{
		client: client,
		logger: logger,
		config: config,
	}
}

// Publish publica uma mensagem com retry automático
func (p *Publisher) Publish(ctx context.Context, exchange, routingKey string, message *Message) error {
	if exchange == "" {
		exchange = p.config.DefaultExchange
	}

	// Contexto com timeout
	ctx, cancel := context.WithTimeout(ctx, p.config.DefaultTimeout)
	defer cancel()

	var lastErr error
	for attempt := 0; attempt <= p.config.MaxRetries; attempt++ {
		if attempt > 0 {
			p.logger.Warn("Retrying message publication",
				zap.String("message_id", message.ID),
				zap.String("exchange", exchange),
				zap.String("routing_key", routingKey),
				zap.Int("attempt", attempt),
			)

			select {
			case <-time.After(p.config.RetryDelay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		err := p.client.Publish(ctx, exchange, routingKey, *message)
		if err == nil {
			if attempt > 0 {
				p.logger.Info("Message published successfully after retry",
					zap.String("message_id", message.ID),
					zap.Int("attempts", attempt+1),
				)
			}
			return nil
		}

		lastErr = err
		p.logger.Error("Failed to publish message",
			zap.String("message_id", message.ID),
			zap.String("exchange", exchange),
			zap.String("routing_key", routingKey),
			zap.Int("attempt", attempt+1),
			zap.Error(err),
		)
	}

	return fmt.Errorf("failed to publish message after %d attempts: %w", p.config.MaxRetries+1, lastErr)
}

// PublishToDefault publica uma mensagem no exchange padrão
func (p *Publisher) PublishToDefault(ctx context.Context, routingKey string, message *Message) error {
	return p.Publish(ctx, p.config.DefaultExchange, routingKey, message)
}

// PublishUserEvent publica eventos relacionados a usuários
func (p *Publisher) PublishUserEvent(ctx context.Context, eventType string, payload UserEventPayload) error {
	message := NewMessage(eventType, payload)
	message.SetTenantID(payload.TenantID)

	return p.PublishToDefault(ctx, "user.events", message)
}

// PublishTenantEvent publica eventos relacionados a tenants
func (p *Publisher) PublishTenantEvent(ctx context.Context, eventType string, payload TenantEventPayload) error {
	message := NewMessage(eventType, payload)

	return p.PublishToDefault(ctx, "tenant.events", message)
}

// PublishEventEvent publica eventos relacionados a eventos
func (p *Publisher) PublishEventEvent(ctx context.Context, eventType string, payload EventEventPayload) error {
	message := NewMessage(eventType, payload)
	message.SetTenantID(payload.TenantID)

	return p.PublishToDefault(ctx, "event.events", message)
}

// PublishEmployeeEvent publica eventos relacionados a funcionários
func (p *Publisher) PublishEmployeeEvent(ctx context.Context, eventType string, payload EmployeeEventPayload) error {
	message := NewMessage(eventType, payload)
	message.SetTenantID(payload.TenantID)

	return p.PublishToDefault(ctx, "employee.events", message)
}

// PublishCheckinEvent publica eventos relacionados a check-ins
func (p *Publisher) PublishCheckinEvent(ctx context.Context, eventType string, payload CheckinEventPayload) error {
	message := NewMessage(eventType, payload)
	message.SetTenantID(payload.TenantID)

	return p.PublishToDefault(ctx, "checkin.events", message)
}

// PublishCheckoutEvent publica eventos relacionados a check-outs
func (p *Publisher) PublishCheckoutEvent(ctx context.Context, eventType string, payload CheckoutEventPayload) error {
	message := NewMessage(eventType, payload)
	message.SetTenantID(payload.TenantID)

	return p.PublishToDefault(ctx, "checkout.events", message)
}

// PublishSystemEvent publica eventos de sistema
func (p *Publisher) PublishSystemEvent(ctx context.Context, eventType string, payload SystemEventPayload) error {
	message := NewMessage(eventType, payload)

	return p.PublishToDefault(ctx, "system.events", message)
}

// PublishNotificationEvent publica eventos de notificação
func (p *Publisher) PublishNotificationEvent(ctx context.Context, eventType string, payload NotificationEventPayload) error {
	message := NewMessage(eventType, payload)
	if payload.TenantID != "" {
		message.SetTenantID(payload.TenantID)
	}

	return p.PublishToDefault(ctx, "notification.events", message)
}

// PublishDelayedMessage publica uma mensagem com delay
func (p *Publisher) PublishDelayedMessage(ctx context.Context, exchange, routingKey string, message *Message, delay time.Duration) error {
	// Adicionar header de delay
	message.SetHeader("x-delay", delay.Milliseconds())

	return p.Publish(ctx, exchange, routingKey, message)
}

// PublishBatch publica múltiplas mensagens em lote
func (p *Publisher) PublishBatch(ctx context.Context, exchange string, messages map[string]*Message) error {
	if len(messages) == 0 {
		return nil
	}

	var errors []error
	for routingKey, message := range messages {
		if err := p.Publish(ctx, exchange, routingKey, message); err != nil {
			p.logger.Error("Failed to publish message in batch",
				zap.String("message_id", message.ID),
				zap.String("routing_key", routingKey),
				zap.Error(err),
			)
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to publish %d messages in batch", len(errors))
	}

	p.logger.Info("Batch messages published successfully", zap.Int("count", len(messages)))
	return nil
}

// PublishWithConfirmation publica uma mensagem com confirmação
func (p *Publisher) PublishWithConfirmation(ctx context.Context, exchange, routingKey string, message *Message) error {
	// Esta funcionalidade requer publisher confirms do RabbitMQ
	// Por simplicidade, usamos a publicação normal por enquanto
	return p.Publish(ctx, exchange, routingKey, message)
}
