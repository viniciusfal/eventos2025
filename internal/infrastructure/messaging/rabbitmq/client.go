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

// Config contém as configurações do RabbitMQ
type Config struct {
	Host              string
	Port              int
	Username          string
	Password          string
	VHost             string
	ConnectionName    string
	Heartbeat         time.Duration
	ConnectionTimeout time.Duration
	MaxRetries        int
	RetryDelay        time.Duration
}

// Client representa um cliente RabbitMQ
type Client struct {
	config     Config
	connection *amqp091.Connection
	channel    *amqp091.Channel
	logger     *zap.Logger
	mutex      sync.RWMutex
	closed     bool
}

// NewClient cria uma nova instância do cliente RabbitMQ
func NewClient(cfg Config, logger *zap.Logger) (*Client, error) {
	client := &Client{
		config: cfg,
		logger: logger,
	}

	if err := client.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	logger.Info("RabbitMQ client connected successfully",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("vhost", cfg.VHost),
	)

	return client, nil
}

// connect estabelece conexão com o RabbitMQ
func (c *Client) connect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.connection != nil && !c.connection.IsClosed() {
		return nil
	}

	// Construir URL de conexão
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		c.config.Username,
		c.config.Password,
		c.config.Host,
		c.config.Port,
		c.config.VHost,
	)

	// Configurar propriedades de conexão
	config := amqp091.Config{
		Heartbeat: c.config.Heartbeat,
		Properties: amqp091.Table{
			"connection_name": c.config.ConnectionName,
		},
		Dial: amqp091.DefaultDial(c.config.ConnectionTimeout),
	}

	// Estabelecer conexão
	conn, err := amqp091.DialConfig(url, config)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Criar canal
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to create RabbitMQ channel: %w", err)
	}

	c.connection = conn
	c.channel = ch
	c.closed = false

	// Configurar notificações de fechamento
	go c.handleConnectionClose()

	return nil
}

// handleConnectionClose lida com fechamento inesperado da conexão
func (c *Client) handleConnectionClose() {
	closeErr := <-c.connection.NotifyClose(make(chan *amqp091.Error))
	if closeErr != nil {
		c.logger.Error("RabbitMQ connection closed unexpectedly", zap.Error(closeErr))

		// Tentar reconectar
		c.reconnect()
	}
}

// reconnect tenta reconectar ao RabbitMQ
func (c *Client) reconnect() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.closed {
		return
	}

	retries := 0
	for retries < c.config.MaxRetries {
		c.logger.Info("Attempting to reconnect to RabbitMQ", zap.Int("attempt", retries+1))

		if err := c.connect(); err != nil {
			retries++
			c.logger.Error("Failed to reconnect to RabbitMQ",
				zap.Error(err),
				zap.Int("attempt", retries),
			)

			if retries < c.config.MaxRetries {
				time.Sleep(c.config.RetryDelay)
			}
		} else {
			c.logger.Info("Successfully reconnected to RabbitMQ")
			return
		}
	}

	c.logger.Error("Failed to reconnect to RabbitMQ after maximum retries")
}

// Close fecha a conexão com o RabbitMQ
func (c *Client) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.closed = true

	var lastErr error

	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			c.logger.Error("Failed to close RabbitMQ channel", zap.Error(err))
			lastErr = err
		}
	}

	if c.connection != nil {
		if err := c.connection.Close(); err != nil {
			c.logger.Error("Failed to close RabbitMQ connection", zap.Error(err))
			lastErr = err
		}
	}

	if lastErr == nil {
		c.logger.Info("RabbitMQ client closed successfully")
	}

	return lastErr
}

// IsConnected verifica se o cliente está conectado
func (c *Client) IsConnected() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.connection != nil && !c.connection.IsClosed() && c.channel != nil
}

// DeclareExchange declara um exchange
func (c *Client) DeclareExchange(name, kind string, durable, autoDelete bool, args amqp091.Table) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if !c.IsConnected() {
		return ErrNotConnected
	}

	err := c.channel.ExchangeDeclare(
		name,       // name
		kind,       // type
		durable,    // durable
		autoDelete, // auto-deleted
		false,      // internal
		false,      // no-wait
		args,       // arguments
	)

	if err != nil {
		c.logger.Error("Failed to declare exchange",
			zap.String("exchange", name),
			zap.String("type", kind),
			zap.Error(err),
		)
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	c.logger.Debug("Exchange declared successfully",
		zap.String("exchange", name),
		zap.String("type", kind),
		zap.Bool("durable", durable),
	)

	return nil
}

// DeclareQueue declara uma fila
func (c *Client) DeclareQueue(name string, durable, autoDelete, exclusive bool, args amqp091.Table) (amqp091.Queue, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if !c.IsConnected() {
		return amqp091.Queue{}, ErrNotConnected
	}

	queue, err := c.channel.QueueDeclare(
		name,       // name
		durable,    // durable
		autoDelete, // delete when unused
		exclusive,  // exclusive
		false,      // no-wait
		args,       // arguments
	)

	if err != nil {
		c.logger.Error("Failed to declare queue",
			zap.String("queue", name),
			zap.Error(err),
		)
		return amqp091.Queue{}, fmt.Errorf("failed to declare queue: %w", err)
	}

	c.logger.Debug("Queue declared successfully",
		zap.String("queue", queue.Name),
		zap.Bool("durable", durable),
		zap.Int("messages", queue.Messages),
		zap.Int("consumers", queue.Consumers),
	)

	return queue, nil
}

// BindQueue vincula uma fila a um exchange
func (c *Client) BindQueue(queueName, routingKey, exchangeName string, args amqp091.Table) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if !c.IsConnected() {
		return ErrNotConnected
	}

	err := c.channel.QueueBind(
		queueName,    // queue name
		routingKey,   // routing key
		exchangeName, // exchange
		false,        // no-wait
		args,         // args
	)

	if err != nil {
		c.logger.Error("Failed to bind queue",
			zap.String("queue", queueName),
			zap.String("exchange", exchangeName),
			zap.String("routing_key", routingKey),
			zap.Error(err),
		)
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	c.logger.Debug("Queue bound successfully",
		zap.String("queue", queueName),
		zap.String("exchange", exchangeName),
		zap.String("routing_key", routingKey),
	)

	return nil
}

// Publish publica uma mensagem
func (c *Client) Publish(ctx context.Context, exchange, routingKey string, message Message) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if !c.IsConnected() {
		return ErrNotConnected
	}

	// Serializar corpo da mensagem
	body, err := json.Marshal(message.Body)
	if err != nil {
		return fmt.Errorf("failed to marshal message body: %w", err)
	}

	// Preparar headers
	headers := make(amqp091.Table)
	for k, v := range message.Headers {
		headers[k] = v
	}

	// Adicionar metadados padrão
	headers["published_at"] = time.Now().UTC()
	headers["message_id"] = message.ID
	headers["message_type"] = message.Type

	// Publicar mensagem
	err = c.channel.PublishWithContext(
		ctx,
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp091.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp091.Persistent, // make message persistent
			MessageId:    message.ID,
			Type:         message.Type,
			Timestamp:    time.Now(),
			Headers:      headers,
			Body:         body,
		},
	)

	if err != nil {
		c.logger.Error("Failed to publish message",
			zap.String("exchange", exchange),
			zap.String("routing_key", routingKey),
			zap.String("message_id", message.ID),
			zap.String("message_type", message.Type),
			zap.Error(err),
		)
		return fmt.Errorf("failed to publish message: %w", err)
	}

	c.logger.Debug("Message published successfully",
		zap.String("exchange", exchange),
		zap.String("routing_key", routingKey),
		zap.String("message_id", message.ID),
		zap.String("message_type", message.Type),
	)

	return nil
}

// Consume inicia o consumo de mensagens de uma fila
func (c *Client) Consume(queueName, consumerTag string, autoAck bool) (<-chan amqp091.Delivery, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if !c.IsConnected() {
		return nil, ErrNotConnected
	}

	deliveries, err := c.channel.Consume(
		queueName,   // queue
		consumerTag, // consumer
		autoAck,     // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)

	if err != nil {
		c.logger.Error("Failed to start consuming",
			zap.String("queue", queueName),
			zap.String("consumer", consumerTag),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to start consuming: %w", err)
	}

	c.logger.Info("Started consuming messages",
		zap.String("queue", queueName),
		zap.String("consumer", consumerTag),
		zap.Bool("auto_ack", autoAck),
	)

	return deliveries, nil
}

// GetChannel retorna o canal RabbitMQ (para operações avançadas)
func (c *Client) GetChannel() *amqp091.Channel {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.channel
}
