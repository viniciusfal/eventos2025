package rabbitmq

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Message representa uma mensagem do sistema
type Message struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Body      interface{}            `json:"body"`
	Headers   map[string]interface{} `json:"headers,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Retry     int                    `json:"retry,omitempty"`
}

// NewMessage cria uma nova mensagem
func NewMessage(messageType string, body interface{}) *Message {
	return &Message{
		ID:        uuid.New().String(),
		Type:      messageType,
		Body:      body,
		Headers:   make(map[string]interface{}),
		Timestamp: time.Now().UTC(),
		Retry:     0,
	}
}

// SetHeader define um header na mensagem
func (m *Message) SetHeader(key string, value interface{}) *Message {
	if m.Headers == nil {
		m.Headers = make(map[string]interface{})
	}
	m.Headers[key] = value
	return m
}

// GetHeader retorna um header da mensagem
func (m *Message) GetHeader(key string) (interface{}, bool) {
	if m.Headers == nil {
		return nil, false
	}
	value, exists := m.Headers[key]
	return value, exists
}

// SetTenantID define o tenant ID nos headers
func (m *Message) SetTenantID(tenantID string) *Message {
	return m.SetHeader("tenant_id", tenantID)
}

// GetTenantID retorna o tenant ID dos headers
func (m *Message) GetTenantID() (string, bool) {
	if value, exists := m.GetHeader("tenant_id"); exists {
		if tenantID, ok := value.(string); ok {
			return tenantID, true
		}
	}
	return "", false
}

// SetUserID define o user ID nos headers
func (m *Message) SetUserID(userID string) *Message {
	return m.SetHeader("user_id", userID)
}

// GetUserID retorna o user ID dos headers
func (m *Message) GetUserID() (string, bool) {
	if value, exists := m.GetHeader("user_id"); exists {
		if userID, ok := value.(string); ok {
			return userID, true
		}
	}
	return "", false
}

// SetCorrelationID define um correlation ID para rastreamento
func (m *Message) SetCorrelationID(correlationID string) *Message {
	return m.SetHeader("correlation_id", correlationID)
}

// GetCorrelationID retorna o correlation ID
func (m *Message) GetCorrelationID() (string, bool) {
	if value, exists := m.GetHeader("correlation_id"); exists {
		if correlationID, ok := value.(string); ok {
			return correlationID, true
		}
	}
	return "", false
}

// IncrementRetry incrementa o contador de tentativas
func (m *Message) IncrementRetry() *Message {
	m.Retry++
	return m
}

// ToJSON converte a mensagem para JSON
func (m *Message) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// FromJSON cria uma mensagem a partir de JSON
func FromJSON(data []byte) (*Message, error) {
	var message Message
	err := json.Unmarshal(data, &message)
	if err != nil {
		return nil, err
	}
	return &message, nil
}

// Tipos de mensagens predefinidos
const (
	// Eventos de usuário
	MessageTypeUserCreated   = "user.created"
	MessageTypeUserUpdated   = "user.updated"
	MessageTypeUserDeleted   = "user.deleted"
	MessageTypeUserLoggedIn  = "user.logged_in"
	MessageTypeUserLoggedOut = "user.logged_out"

	// Eventos de tenant
	MessageTypeTenantCreated = "tenant.created"
	MessageTypeTenantUpdated = "tenant.updated"
	MessageTypeTenantDeleted = "tenant.deleted"

	// Eventos de evento
	MessageTypeEventCreated = "event.created"
	MessageTypeEventUpdated = "event.updated"
	MessageTypeEventDeleted = "event.deleted"
	MessageTypeEventStarted = "event.started"
	MessageTypeEventEnded   = "event.ended"

	// Eventos de funcionário
	MessageTypeEmployeeCreated = "employee.created"
	MessageTypeEmployeeUpdated = "employee.updated"
	MessageTypeEmployeeDeleted = "employee.deleted"

	// Eventos de parceiro
	MessageTypePartnerCreated = "partner.created"
	MessageTypePartnerUpdated = "partner.updated"
	MessageTypePartnerDeleted = "partner.deleted"

	// Eventos de role
	MessageTypeRoleCreated = "role.created"
	MessageTypeRoleUpdated = "role.updated"
	MessageTypeRoleDeleted = "role.deleted"

	// Eventos de permissão
	MessageTypePermissionCreated = "permission.created"
	MessageTypePermissionUpdated = "permission.updated"
	MessageTypePermissionDeleted = "permission.deleted"

	// Eventos de check-in
	MessageTypeCheckinPerformed = "checkin.performed"
	MessageTypeCheckinValidated = "checkin.validated"
	MessageTypeCheckinInvalid   = "checkin.invalid"

	// Eventos de check-out
	MessageTypeCheckoutPerformed = "checkout.performed"
	MessageTypeCheckoutValidated = "checkout.validated"
	MessageTypeCheckoutInvalid   = "checkout.invalid"

	// Eventos de sessão de trabalho
	MessageTypeWorkSessionCompleted = "work_session.completed"
	MessageTypeWorkSessionInvalid   = "work_session.invalid"

	// Eventos de sistema
	MessageTypeSystemError   = "system.error"
	MessageTypeSystemWarning = "system.warning"
	MessageTypeSystemInfo    = "system.info"

	// Eventos de cache
	MessageTypeCacheInvalidated = "cache.invalidated"

	// Eventos de notificação
	MessageTypeNotificationSent = "notification.sent"
	MessageTypeEmailSent        = "email.sent"
	MessageTypeSMSSent          = "sms.sent"
)

// EventPayloads define as estruturas de payload para diferentes tipos de eventos

// UserEventPayload payload para eventos de usuário
type UserEventPayload struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	Email    string `json:"email,omitempty"`
	Name     string `json:"name,omitempty"`
}

// TenantEventPayload payload para eventos de tenant
type TenantEventPayload struct {
	TenantID string `json:"tenant_id"`
	Name     string `json:"name,omitempty"`
	Domain   string `json:"domain,omitempty"`
}

// EventEventPayload payload para eventos de evento
type EventEventPayload struct {
	EventID   string    `json:"event_id"`
	TenantID  string    `json:"tenant_id"`
	Name      string    `json:"name,omitempty"`
	StartTime time.Time `json:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"`
}

// EmployeeEventPayload payload para eventos de funcionário
type EmployeeEventPayload struct {
	EmployeeID string `json:"employee_id"`
	TenantID   string `json:"tenant_id"`
	Name       string `json:"name,omitempty"`
	Email      string `json:"email,omitempty"`
	PartnerID  string `json:"partner_id,omitempty"`
}

// CheckinEventPayload payload para eventos de check-in
type CheckinEventPayload struct {
	CheckinID   string    `json:"checkin_id"`
	TenantID    string    `json:"tenant_id"`
	EventID     string    `json:"event_id"`
	EmployeeID  string    `json:"employee_id"`
	PartnerID   string    `json:"partner_id"`
	Method      string    `json:"method"`
	IsValid     bool      `json:"is_valid"`
	CheckinTime time.Time `json:"checkin_time"`
}

// CheckoutEventPayload payload para eventos de check-out
type CheckoutEventPayload struct {
	CheckoutID   string        `json:"checkout_id"`
	CheckinID    string        `json:"checkin_id"`
	TenantID     string        `json:"tenant_id"`
	EventID      string        `json:"event_id"`
	EmployeeID   string        `json:"employee_id"`
	PartnerID    string        `json:"partner_id"`
	Method       string        `json:"method"`
	IsValid      bool          `json:"is_valid"`
	CheckoutTime time.Time     `json:"checkout_time"`
	WorkDuration time.Duration `json:"work_duration"`
}

// SystemEventPayload payload para eventos de sistema
type SystemEventPayload struct {
	Level     string                 `json:"level"` // error, warning, info
	Message   string                 `json:"message"`
	Component string                 `json:"component,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// NotificationEventPayload payload para eventos de notificação
type NotificationEventPayload struct {
	NotificationID string                 `json:"notification_id"`
	TenantID       string                 `json:"tenant_id,omitempty"`
	UserID         string                 `json:"user_id,omitempty"`
	Type           string                 `json:"type"` // email, sms, push
	Subject        string                 `json:"subject,omitempty"`
	Body           string                 `json:"body"`
	Recipients     []string               `json:"recipients"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}
