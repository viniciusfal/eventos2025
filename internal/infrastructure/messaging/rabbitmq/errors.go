package rabbitmq

import "errors"

var (
	// ErrNotConnected indica que o cliente não está conectado
	ErrNotConnected = errors.New("rabbitmq client not connected")

	// ErrConnectionClosed indica que a conexão foi fechada
	ErrConnectionClosed = errors.New("rabbitmq connection closed")

	// ErrChannelClosed indica que o canal foi fechado
	ErrChannelClosed = errors.New("rabbitmq channel closed")

	// ErrPublishFailed indica falha na publicação de mensagem
	ErrPublishFailed = errors.New("failed to publish message")

	// ErrConsumeFailed indica falha no consumo de mensagens
	ErrConsumeFailed = errors.New("failed to consume messages")

	// ErrExchangeDeclarationFailed indica falha na declaração do exchange
	ErrExchangeDeclarationFailed = errors.New("failed to declare exchange")

	// ErrQueueDeclarationFailed indica falha na declaração da fila
	ErrQueueDeclarationFailed = errors.New("failed to declare queue")

	// ErrQueueBindFailed indica falha no binding da fila
	ErrQueueBindFailed = errors.New("failed to bind queue")

	// ErrMessageProcessingFailed indica falha no processamento da mensagem
	ErrMessageProcessingFailed = errors.New("failed to process message")

	// ErrInvalidMessage indica mensagem inválida
	ErrInvalidMessage = errors.New("invalid message")

	// ErrMaxRetriesExceeded indica que o número máximo de tentativas foi excedido
	ErrMaxRetriesExceeded = errors.New("maximum retries exceeded")
)
