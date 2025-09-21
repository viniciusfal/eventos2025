package checkout

import (
	"fmt"
	"strings"
	"time"

	"eventos-backend/internal/domain/shared/constants"
	"eventos-backend/internal/domain/shared/errors"
	"eventos-backend/internal/domain/shared/value_objects"
)

// Checkout representa um check-out de funcionário de um evento
type Checkout struct {
	ID                value_objects.UUID
	TenantID          value_objects.UUID
	EventID           value_objects.UUID
	EmployeeID        value_objects.UUID
	PartnerID         value_objects.UUID
	CheckinID         value_objects.UUID // Referência ao check-in correspondente
	Method            string             // facial_recognition, qr_code, manual
	Location          value_objects.Location
	CheckoutTime      time.Time
	PhotoURL          string                 // Foto capturada no momento do check-out
	Notes             string                 // Observações do check-out
	WorkDuration      time.Duration          // Duração calculada entre check-in e check-out
	IsValid           bool                   // Se o check-out é válido
	ValidationDetails map[string]interface{} // Detalhes da validação
	CreatedAt         time.Time
	UpdatedAt         time.Time
	CreatedBy         *value_objects.UUID
	UpdatedBy         *value_objects.UUID
}

// NewCheckout cria um novo check-out com validações
func NewCheckout(tenantID, eventID, employeeID, partnerID, checkinID value_objects.UUID, method string, location value_objects.Location, photoURL, notes string, createdBy value_objects.UUID) (*Checkout, error) {
	checkout := &Checkout{
		ID:                value_objects.NewUUID(),
		TenantID:          tenantID,
		EventID:           eventID,
		EmployeeID:        employeeID,
		PartnerID:         partnerID,
		CheckinID:         checkinID,
		Method:            strings.ToLower(strings.TrimSpace(method)),
		Location:          location,
		CheckoutTime:      time.Now(),
		PhotoURL:          strings.TrimSpace(photoURL),
		Notes:             strings.TrimSpace(notes),
		WorkDuration:      0,     // Será calculado posteriormente
		IsValid:           false, // Será validado posteriormente
		ValidationDetails: make(map[string]interface{}),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		CreatedBy:         &createdBy,
		UpdatedBy:         &createdBy,
	}

	if err := checkout.Validate(); err != nil {
		return nil, err
	}

	return checkout, nil
}

// Validate valida os dados do check-out
func (c *Checkout) Validate() error {
	if c.ID.IsZero() {
		return errors.NewValidationError("ID", "é obrigatório")
	}

	if c.TenantID.IsZero() {
		return errors.NewValidationError("TenantID", "é obrigatório")
	}

	if c.EventID.IsZero() {
		return errors.NewValidationError("EventID", "é obrigatório")
	}

	if c.EmployeeID.IsZero() {
		return errors.NewValidationError("EmployeeID", "é obrigatório")
	}

	if c.PartnerID.IsZero() {
		return errors.NewValidationError("PartnerID", "é obrigatório")
	}

	if c.CheckinID.IsZero() {
		return errors.NewValidationError("CheckinID", "é obrigatório")
	}

	if err := c.validateMethod(); err != nil {
		return err
	}

	if err := c.validateLocation(); err != nil {
		return err
	}

	if err := c.validateNotes(); err != nil {
		return err
	}

	return nil
}

// validateMethod valida o método de check-out
func (c *Checkout) validateMethod() error {
	if c.Method == "" {
		return errors.NewValidationError("Method", "é obrigatório")
	}

	validMethods := map[string]bool{
		constants.CheckMethodFacialRecognition: true,
		constants.CheckMethodQRCode:            true,
		constants.CheckMethodManual:            true,
	}

	if !validMethods[c.Method] {
		return errors.NewValidationError("Method", "método de check-out não reconhecido")
	}

	return nil
}

// validateLocation valida a localização do check-out
func (c *Checkout) validateLocation() error {
	// A localização é validada pelo value object Location
	// Aqui podemos adicionar validações específicas do check-out se necessário
	return nil
}

// validateNotes valida as observações
func (c *Checkout) validateNotes() error {
	if len(c.Notes) > 1000 {
		return errors.NewValidationError("Notes", "deve ter no máximo 1000 caracteres")
	}

	return nil
}

// CalculateWorkDuration calcula a duração do trabalho baseada no check-in
func (c *Checkout) CalculateWorkDuration(checkinTime time.Time) {
	c.WorkDuration = c.CheckoutTime.Sub(checkinTime)
	c.SetValidationDetail("work_duration_minutes", c.WorkDuration.Minutes())
}

// MarkAsValid marca o check-out como válido
func (c *Checkout) MarkAsValid(validationDetails map[string]interface{}, updatedBy value_objects.UUID) {
	c.IsValid = true
	c.ValidationDetails = validationDetails
	c.UpdatedAt = time.Now()
	c.UpdatedBy = &updatedBy
}

// MarkAsInvalid marca o check-out como inválido
func (c *Checkout) MarkAsInvalid(validationDetails map[string]interface{}, updatedBy value_objects.UUID) {
	c.IsValid = false
	c.ValidationDetails = validationDetails
	c.UpdatedAt = time.Now()
	c.UpdatedBy = &updatedBy
}

// AddNote adiciona uma observação ao check-out
func (c *Checkout) AddNote(note string, updatedBy value_objects.UUID) error {
	note = strings.TrimSpace(note)
	if note == "" {
		return errors.NewValidationError("Note", "não pode estar vazia")
	}

	if c.Notes == "" {
		c.Notes = note
	} else {
		c.Notes = fmt.Sprintf("%s\n%s", c.Notes, note)
	}

	if len(c.Notes) > 1000 {
		return errors.NewValidationError("Notes", "deve ter no máximo 1000 caracteres")
	}

	c.UpdatedAt = time.Now()
	c.UpdatedBy = &updatedBy

	return nil
}

// IsFacialRecognition verifica se o check-out foi feito por reconhecimento facial
func (c *Checkout) IsFacialRecognition() bool {
	return c.Method == constants.CheckMethodFacialRecognition
}

// IsQRCode verifica se o check-out foi feito por QR Code
func (c *Checkout) IsQRCode() bool {
	return c.Method == constants.CheckMethodQRCode
}

// IsManual verifica se o check-out foi feito manualmente
func (c *Checkout) IsManual() bool {
	return c.Method == constants.CheckMethodManual
}

// HasPhoto verifica se o check-out tem foto
func (c *Checkout) HasPhoto() bool {
	return c.PhotoURL != ""
}

// GetValidationDetail obtém um detalhe específico da validação
func (c *Checkout) GetValidationDetail(key string) (interface{}, bool) {
	value, exists := c.ValidationDetails[key]
	return value, exists
}

// SetValidationDetail define um detalhe da validação
func (c *Checkout) SetValidationDetail(key string, value interface{}) {
	if c.ValidationDetails == nil {
		c.ValidationDetails = make(map[string]interface{})
	}
	c.ValidationDetails[key] = value
}

// GetDistanceFromEvent retorna a distância do check-out em relação ao evento (se disponível)
func (c *Checkout) GetDistanceFromEvent() (float64, bool) {
	distance, exists := c.GetValidationDetail("distance_from_event")
	if !exists {
		return 0, false
	}

	if dist, ok := distance.(float64); ok {
		return dist, true
	}

	return 0, false
}

// GetFacialSimilarity retorna a similaridade facial (se disponível)
func (c *Checkout) GetFacialSimilarity() (float64, bool) {
	similarity, exists := c.GetValidationDetail("facial_similarity")
	if !exists {
		return 0, false
	}

	if sim, ok := similarity.(float64); ok {
		return sim, true
	}

	return 0, false
}

// IsWithinEventBounds verifica se o check-out está dentro dos limites do evento
func (c *Checkout) IsWithinEventBounds() (bool, bool) {
	withinBounds, exists := c.GetValidationDetail("within_event_bounds")
	if !exists {
		return false, false
	}

	if bounds, ok := withinBounds.(bool); ok {
		return bounds, true
	}

	return false, false
}

// GetWorkDurationHours retorna a duração do trabalho em horas
func (c *Checkout) GetWorkDurationHours() float64 {
	return c.WorkDuration.Hours()
}

// GetWorkDurationMinutes retorna a duração do trabalho em minutos
func (c *Checkout) GetWorkDurationMinutes() float64 {
	return c.WorkDuration.Minutes()
}

// IsShortWork verifica se o trabalho foi muito curto (menos de 1 hora)
func (c *Checkout) IsShortWork() bool {
	return c.WorkDuration < time.Hour
}

// IsLongWork verifica se o trabalho foi muito longo (mais de 12 horas)
func (c *Checkout) IsLongWork() bool {
	return c.WorkDuration > 12*time.Hour
}

// GetTimeSinceCheckout retorna o tempo desde o check-out
func (c *Checkout) GetTimeSinceCheckout() time.Duration {
	return time.Since(c.CheckoutTime)
}

// IsRecent verifica se o check-out foi feito recentemente (últimas 24 horas)
func (c *Checkout) IsRecent() bool {
	return c.GetTimeSinceCheckout() <= 24*time.Hour
}

// String retorna uma representação string do check-out
func (c *Checkout) String() string {
	return fmt.Sprintf("Checkout{ID: %s, Employee: %s, Event: %s, Method: %s, Duration: %s, Valid: %t}",
		c.ID.String(), c.EmployeeID.String(), c.EventID.String(), c.Method, c.WorkDuration.String(), c.IsValid)
}

// CheckoutStatus representa os possíveis status de um check-out
type CheckoutStatus string

const (
	CheckoutStatusPending   CheckoutStatus = "pending"   // Aguardando validação
	CheckoutStatusValid     CheckoutStatus = "valid"     // Válido
	CheckoutStatusInvalid   CheckoutStatus = "invalid"   // Inválido
	CheckoutStatusCancelled CheckoutStatus = "cancelled" // Cancelado
)

// GetStatus retorna o status atual do check-out
func (c *Checkout) GetStatus() CheckoutStatus {
	if c.IsValid {
		return CheckoutStatusValid
	}

	// Se tem detalhes de validação mas não é válido, é inválido
	if len(c.ValidationDetails) > 0 {
		return CheckoutStatusInvalid
	}

	// Caso contrário, está pendente
	return CheckoutStatusPending
}

// ValidationResult representa o resultado de uma validação de check-out
type ValidationResult struct {
	IsValid           bool
	Reason            string
	Details           map[string]interface{}
	DistanceFromEvent *float64
	FacialSimilarity  *float64
	WithinBounds      *bool
	WorkDuration      *time.Duration
	Timestamp         time.Time
}

// NewValidationResult cria um novo resultado de validação
func NewValidationResult(isValid bool, reason string) *ValidationResult {
	return &ValidationResult{
		IsValid:   isValid,
		Reason:    reason,
		Details:   make(map[string]interface{}),
		Timestamp: time.Now(),
	}
}

// SetDistance define a distância do evento
func (vr *ValidationResult) SetDistance(distance float64) *ValidationResult {
	vr.DistanceFromEvent = &distance
	vr.Details["distance_from_event"] = distance
	return vr
}

// SetFacialSimilarity define a similaridade facial
func (vr *ValidationResult) SetFacialSimilarity(similarity float64) *ValidationResult {
	vr.FacialSimilarity = &similarity
	vr.Details["facial_similarity"] = similarity
	return vr
}

// SetWithinBounds define se está dentro dos limites
func (vr *ValidationResult) SetWithinBounds(withinBounds bool) *ValidationResult {
	vr.WithinBounds = &withinBounds
	vr.Details["within_event_bounds"] = withinBounds
	return vr
}

// SetWorkDuration define a duração do trabalho
func (vr *ValidationResult) SetWorkDuration(duration time.Duration) *ValidationResult {
	vr.WorkDuration = &duration
	vr.Details["work_duration"] = duration
	vr.Details["work_duration_hours"] = duration.Hours()
	vr.Details["work_duration_minutes"] = duration.Minutes()
	return vr
}

// AddDetail adiciona um detalhe customizado
func (vr *ValidationResult) AddDetail(key string, value interface{}) *ValidationResult {
	vr.Details[key] = value
	return vr
}

// WorkSession representa uma sessão de trabalho (check-in + check-out)
type WorkSession struct {
	CheckinID    value_objects.UUID
	CheckoutID   value_objects.UUID
	EmployeeID   value_objects.UUID
	EventID      value_objects.UUID
	PartnerID    value_objects.UUID
	CheckinTime  time.Time
	CheckoutTime time.Time
	Duration     time.Duration
	IsComplete   bool
	IsValid      bool
}

// NewWorkSession cria uma nova sessão de trabalho
func NewWorkSession(checkinID, checkoutID, employeeID, eventID, partnerID value_objects.UUID, checkinTime, checkoutTime time.Time) *WorkSession {
	duration := checkoutTime.Sub(checkinTime)

	return &WorkSession{
		CheckinID:    checkinID,
		CheckoutID:   checkoutID,
		EmployeeID:   employeeID,
		EventID:      eventID,
		PartnerID:    partnerID,
		CheckinTime:  checkinTime,
		CheckoutTime: checkoutTime,
		Duration:     duration,
		IsComplete:   true,
		IsValid:      duration > 0, // Básico: duração deve ser positiva
	}
}

// GetDurationHours retorna a duração em horas
func (ws *WorkSession) GetDurationHours() float64 {
	return ws.Duration.Hours()
}

// GetDurationMinutes retorna a duração em minutos
func (ws *WorkSession) GetDurationMinutes() float64 {
	return ws.Duration.Minutes()
}

// IsShortSession verifica se a sessão foi muito curta
func (ws *WorkSession) IsShortSession() bool {
	return ws.Duration < time.Hour
}

// IsLongSession verifica se a sessão foi muito longa
func (ws *WorkSession) IsLongSession() bool {
	return ws.Duration > 12*time.Hour
}

// String retorna uma representação string da sessão
func (ws *WorkSession) String() string {
	return fmt.Sprintf("WorkSession{Employee: %s, Event: %s, Duration: %s, Valid: %t}",
		ws.EmployeeID.String(), ws.EventID.String(), ws.Duration.String(), ws.IsValid)
}
