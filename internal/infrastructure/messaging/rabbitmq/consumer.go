package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// MessageHandler define a interface para processamento de mensagens
type MessageHandler interface {
	Handle(ctx context.Context, message *Message) error
	CanHandle(messageType string) bool
	GetName() string
}

// Consumer gerencia o consumo de mensagens
type Consumer struct {
	client      *Client
	logger      *zap.Logger
	config      ConsumerConfig
	handlers    map[string]MessageHandler
	handlersMux sync.RWMutex
	running     bool
	runningMux  sync.RWMutex
	cancelFunc  context.CancelFunc
}

// ConsumerConfig configurações do consumer
type ConsumerConfig struct {
	QueueName           string
	ConsumerTag         string
	AutoAck             bool
	PrefetchCount       int
	PrefetchSize        int
	MaxRetries          int
	RetryDelay          time.Duration
	ProcessingTimeout   time.Duration
	ConcurrentConsumers int
}

// NewConsumer cria uma nova instância do consumer
func NewConsumer(client *Client, config ConsumerConfig, logger *zap.Logger) *Consumer {
	if config.ConsumerTag == "" {
		config.ConsumerTag = fmt.Sprintf("consumer-%d", time.Now().Unix())
	}
	if config.PrefetchCount == 0 {
		config.PrefetchCount = 10
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 5 * time.Second
	}
	if config.ProcessingTimeout == 0 {
		config.ProcessingTimeout = 30 * time.Second
	}
	if config.ConcurrentConsumers == 0 {
		config.ConcurrentConsumers = 1
	}

	return &Consumer{
		client:   client,
		logger:   logger,
		config:   config,
		handlers: make(map[string]MessageHandler),
	}
}

// RegisterHandler registra um handler para um tipo de mensagem
func (c *Consumer) RegisterHandler(messageType string, handler MessageHandler) {
	c.handlersMux.Lock()
	defer c.handlersMux.Unlock()

	c.handlers[messageType] = handler
	c.logger.Info("Message handler registered",
		zap.String("message_type", messageType),
		zap.String("handler", handler.GetName()),
	)
}

// UnregisterHandler remove um handler
func (c *Consumer) UnregisterHandler(messageType string) {
	c.handlersMux.Lock()
	defer c.handlersMux.Unlock()

	if handler, exists := c.handlers[messageType]; exists {
		delete(c.handlers, messageType)
		c.logger.Info("Message handler unregistered",
			zap.String("message_type", messageType),
			zap.String("handler", handler.GetName()),
		)
	}
}

// Start inicia o consumo de mensagens
func (c *Consumer) Start(ctx context.Context) error {
	c.runningMux.Lock()
	defer c.runningMux.Unlock()

	if c.running {
		return fmt.Errorf("consumer is already running")
	}

	// Configurar QoS
	if err := c.client.GetChannel().Qos(c.config.PrefetchCount, c.config.PrefetchSize, false); err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// Criar contexto cancelável
	ctx, c.cancelFunc = context.WithCancel(ctx)

	// Iniciar consumers concorrentes
	var wg sync.WaitGroup
	for i := 0; i < c.config.ConcurrentConsumers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			c.runConsumer(ctx, workerID)
		}(i)
	}

	c.running = true
	c.logger.Info("Consumer started",
		zap.String("queue", c.config.QueueName),
		zap.String("consumer_tag", c.config.ConsumerTag),
		zap.Int("concurrent_consumers", c.config.ConcurrentConsumers),
	)

	// Aguardar todos os workers terminarem
	go func() {
		wg.Wait()
		c.runningMux.Lock()
		c.running = false
		c.runningMux.Unlock()
		c.logger.Info("All consumer workers stopped")
	}()

	return nil
}

// Stop para o consumo de mensagens
func (c *Consumer) Stop() error {
	c.runningMux.Lock()
	defer c.runningMux.Unlock()

	if !c.running {
		return fmt.Errorf("consumer is not running")
	}

	if c.cancelFunc != nil {
		c.cancelFunc()
	}

	c.logger.Info("Consumer stop requested")
	return nil
}

// IsRunning verifica se o consumer está rodando
func (c *Consumer) IsRunning() bool {
	c.runningMux.RLock()
	defer c.runningMux.RUnlock()
	return c.running
}

// runConsumer executa um worker de consumo
func (c *Consumer) runConsumer(ctx context.Context, workerID int) {
	logger := c.logger.With(zap.Int("worker_id", workerID))
	logger.Info("Consumer worker started")

	for {
		select {
		case <-ctx.Done():
			logger.Info("Consumer worker stopped by context")
			return
		default:
			if err := c.consumeMessages(ctx, logger); err != nil {
				logger.Error("Error in consumer worker", zap.Error(err))

				// Aguardar antes de tentar novamente
				select {
				case <-time.After(c.config.RetryDelay):
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

// consumeMessages consome mensagens da fila
func (c *Consumer) consumeMessages(ctx context.Context, logger *zap.Logger) error {
	deliveries, err := c.client.Consume(c.config.QueueName, c.config.ConsumerTag, c.config.AutoAck)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case delivery, ok := <-deliveries:
			if !ok {
				logger.Warn("Delivery channel closed, reconnecting...")
				return fmt.Errorf("delivery channel closed")
			}

			c.processMessage(ctx, delivery, logger)
		}
	}
}

// processMessage processa uma mensagem recebida
func (c *Consumer) processMessage(ctx context.Context, delivery amqp091.Delivery, logger *zap.Logger) {
	// Criar contexto com timeout para processamento
	processCtx, cancel := context.WithTimeout(ctx, c.config.ProcessingTimeout)
	defer cancel()

	// Parse da mensagem
	message, err := c.parseMessage(delivery)
	if err != nil {
		logger.Error("Failed to parse message",
			zap.Error(err),
			zap.String("delivery_tag", fmt.Sprintf("%d", delivery.DeliveryTag)),
		)
		c.rejectMessage(delivery, false)
		return
	}

	logger = logger.With(
		zap.String("message_id", message.ID),
		zap.String("message_type", message.Type),
	)

	// Encontrar handler para o tipo de mensagem
	c.handlersMux.RLock()
	handler, exists := c.handlers[message.Type]
	c.handlersMux.RUnlock()

	if !exists {
		logger.Warn("No handler found for message type")
		c.rejectMessage(delivery, false)
		return
	}

	// Processar mensagem
	startTime := time.Now()
	err = handler.Handle(processCtx, message)
	processingTime := time.Since(startTime)

	if err != nil {
		logger.Error("Failed to process message",
			zap.Error(err),
			zap.Duration("processing_time", processingTime),
			zap.String("handler", handler.GetName()),
		)

		// Verificar se deve tentar novamente
		if c.shouldRetry(message, err) {
			c.requeueMessage(delivery, message, logger)
		} else {
			c.rejectMessage(delivery, false)
		}
		return
	}

	// Mensagem processada com sucesso
	logger.Debug("Message processed successfully",
		zap.Duration("processing_time", processingTime),
		zap.String("handler", handler.GetName()),
	)

	if !c.config.AutoAck {
		if err := delivery.Ack(false); err != nil {
			logger.Error("Failed to ack message", zap.Error(err))
		}
	}
}

// parseMessage converte delivery em Message
func (c *Consumer) parseMessage(delivery amqp091.Delivery) (*Message, error) {
	var message Message
	if err := json.Unmarshal(delivery.Body, &message); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	// Adicionar headers do delivery
	if message.Headers == nil {
		message.Headers = make(map[string]interface{})
	}

	for key, value := range delivery.Headers {
		message.Headers[key] = value
	}

	return &message, nil
}

// shouldRetry determina se uma mensagem deve ser reprocessada
func (c *Consumer) shouldRetry(message *Message, err error) bool {
	return message.Retry < c.config.MaxRetries
}

// requeueMessage recoloca uma mensagem na fila para retry
func (c *Consumer) requeueMessage(delivery amqp091.Delivery, message *Message, logger *zap.Logger) {
	message.IncrementRetry()

	// Por simplicidade, rejeitamos e deixamos o RabbitMQ recolocar na fila
	// Em produção, poderíamos usar dead letter exchanges ou delay exchanges
	if err := delivery.Reject(true); err != nil {
		logger.Error("Failed to requeue message", zap.Error(err))
	} else {
		logger.Info("Message requeued for retry", zap.Int("retry_count", message.Retry))
	}
}

// rejectMessage rejeita uma mensagem
func (c *Consumer) rejectMessage(delivery amqp091.Delivery, requeue bool) {
	if err := delivery.Reject(requeue); err != nil {
		c.logger.Error("Failed to reject message", zap.Error(err))
	}
}

// GetHandlerCount retorna o número de handlers registrados
func (c *Consumer) GetHandlerCount() int {
	c.handlersMux.RLock()
	defer c.handlersMux.RUnlock()
	return len(c.handlers)
}

// GetHandlerNames retorna os nomes dos handlers registrados
func (c *Consumer) GetHandlerNames() []string {
	c.handlersMux.RLock()
	defer c.handlersMux.RUnlock()

	names := make([]string, 0, len(c.handlers))
	for messageType, handler := range c.handlers {
		names = append(names, fmt.Sprintf("%s -> %s", messageType, handler.GetName()))
	}

	return names
}
