package checkin

import (
	"fmt"
	"strings"
	"time"

	"eventos-backend/internal/domain/shared/constants"
	"eventos-backend/internal/domain/shared/errors"
	"eventos-backend/internal/domain/shared/value_objects"
)

// Checkin representa um check-in de funcionário em um evento
type Checkin struct {
	ID                value_objects.UUID
	TenantID          value_objects.UUID
	EventID           value_objects.UUID
	EmployeeID        value_objects.UUID
	PartnerID         value_objects.UUID
	Method            string // facial_recognition, qr_code, manual
	Location          value_objects.Location
	CheckinTime       time.Time
	PhotoURL          string                 // Foto capturada no momento do check-in
	Notes             string                 // Observações do check-in
	IsValid           bool                   // Se o check-in é válido (dentro da cerca, horário correto, etc.)
	ValidationDetails map[string]interface{} // Detalhes da validação (distância, similaridade facial, etc.)
	CreatedAt         time.Time
	UpdatedAt         time.Time
	CreatedBy         *value_objects.UUID
	UpdatedBy         *value_objects.UUID
}

// NewCheckin cria um novo check-in com validações
func NewCheckin(tenantID, eventID, employeeID, partnerID value_objects.UUID, method string, location value_objects.Location, photoURL, notes string, createdBy value_objects.UUID) (*Checkin, error) {
	checkin := &Checkin{
		ID:                value_objects.NewUUID(),
		TenantID:          tenantID,
		EventID:           eventID,
		EmployeeID:        employeeID,
		PartnerID:         partnerID,
		Method:            strings.ToLower(strings.TrimSpace(method)),
		Location:          location,
		CheckinTime:       time.Now(),
		PhotoURL:          strings.TrimSpace(photoURL),
		Notes:             strings.TrimSpace(notes),
		IsValid:           false, // Será validado posteriormente
		ValidationDetails: make(map[string]interface{}),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		CreatedBy:         &createdBy,
		UpdatedBy:         &createdBy,
	}

	if err := checkin.Validate(); err != nil {
		return nil, err
	}

	return checkin, nil
}

// Validate valida os dados do check-in
func (c *Checkin) Validate() error {
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

// validateMethod valida o método de check-in
func (c *Checkin) validateMethod() error {
	if c.Method == "" {
		return errors.NewValidationError("Method", "é obrigatório")
	}

	validMethods := map[string]bool{
		constants.CheckMethodFacialRecognition: true,
		constants.CheckMethodQRCode:            true,
		constants.CheckMethodManual:            true,
	}

	if !validMethods[c.Method] {
		return errors.NewValidationError("Method", "método de check-in não reconhecido")
	}

	return nil
}

// validateLocation valida a localização do check-in
func (c *Checkin) validateLocation() error {
	// A localização é validada pelo value object Location
	// Aqui podemos adicionar validações específicas do check-in se necessário
	return nil
}

// validateNotes valida as observações
func (c *Checkin) validateNotes() error {
	if len(c.Notes) > 1000 {
		return errors.NewValidationError("Notes", "deve ter no máximo 1000 caracteres")
	}

	return nil
}

// MarkAsValid marca o check-in como válido
func (c *Checkin) MarkAsValid(validationDetails map[string]interface{}, updatedBy value_objects.UUID) {
	c.IsValid = true
	c.ValidationDetails = validationDetails
	c.UpdatedAt = time.Now()
	c.UpdatedBy = &updatedBy
}

// MarkAsInvalid marca o check-in como inválido
func (c *Checkin) MarkAsInvalid(validationDetails map[string]interface{}, updatedBy value_objects.UUID) {
	c.IsValid = false
	c.ValidationDetails = validationDetails
	c.UpdatedAt = time.Now()
	c.UpdatedBy = &updatedBy
}

// AddNote adiciona uma observação ao check-in
func (c *Checkin) AddNote(note string, updatedBy value_objects.UUID) error {
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

// IsFacialRecognition verifica se o check-in foi feito por reconhecimento facial
func (c *Checkin) IsFacialRecognition() bool {
	return c.Method == constants.CheckMethodFacialRecognition
}

// IsQRCode verifica se o check-in foi feito por QR Code
func (c *Checkin) IsQRCode() bool {
	return c.Method == constants.CheckMethodQRCode
}

// IsManual verifica se o check-in foi feito manualmente
func (c *Checkin) IsManual() bool {
	return c.Method == constants.CheckMethodManual
}

// HasPhoto verifica se o check-in tem foto
func (c *Checkin) HasPhoto() bool {
	return c.PhotoURL != ""
}

// GetValidationDetail obtém um detalhe específico da validação
func (c *Checkin) GetValidationDetail(key string) (interface{}, bool) {
	value, exists := c.ValidationDetails[key]
	return value, exists
}

// SetValidationDetail define um detalhe da validação
func (c *Checkin) SetValidationDetail(key string, value interface{}) {
	if c.ValidationDetails == nil {
		c.ValidationDetails = make(map[string]interface{})
	}
	c.ValidationDetails[key] = value
}

// GetDistanceFromEvent retorna a distância do check-in em relação ao evento (se disponível)
func (c *Checkin) GetDistanceFromEvent() (float64, bool) {
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
func (c *Checkin) GetFacialSimilarity() (float64, bool) {
	similarity, exists := c.GetValidationDetail("facial_similarity")
	if !exists {
		return 0, false
	}

	if sim, ok := similarity.(float64); ok {
		return sim, true
	}

	return 0, false
}

// IsWithinEventBounds verifica se o check-in está dentro dos limites do evento
func (c *Checkin) IsWithinEventBounds() (bool, bool) {
	withinBounds, exists := c.GetValidationDetail("within_event_bounds")
	if !exists {
		return false, false
	}

	if bounds, ok := withinBounds.(bool); ok {
		return bounds, true
	}

	return false, false
}

// GetDuration retorna a duração desde o check-in
func (c *Checkin) GetDuration() time.Duration {
	return time.Since(c.CheckinTime)
}

// IsRecent verifica se o check-in foi feito recentemente (últimas 24 horas)
func (c *Checkin) IsRecent() bool {
	return c.GetDuration() <= 24*time.Hour
}

// String retorna uma representação string do check-in
func (c *Checkin) String() string {
	return fmt.Sprintf("Checkin{ID: %s, Employee: %s, Event: %s, Method: %s, Valid: %t}",
		c.ID.String(), c.EmployeeID.String(), c.EventID.String(), c.Method, c.IsValid)
}

// CheckinStatus representa os possíveis status de um check-in
type CheckinStatus string

const (
	CheckinStatusPending   CheckinStatus = "pending"   // Aguardando validação
	CheckinStatusValid     CheckinStatus = "valid"     // Válido
	CheckinStatusInvalid   CheckinStatus = "invalid"   // Inválido
	CheckinStatusCancelled CheckinStatus = "cancelled" // Cancelado
)

// GetStatus retorna o status atual do check-in
func (c *Checkin) GetStatus() CheckinStatus {
	if c.IsValid {
		return CheckinStatusValid
	}

	// Se tem detalhes de validação mas não é válido, é inválido
	if len(c.ValidationDetails) > 0 {
		return CheckinStatusInvalid
	}

	// Caso contrário, está pendente
	return CheckinStatusPending
}

// ValidationResult representa o resultado de uma validação de check-in
type ValidationResult struct {
	IsValid           bool
	Reason            string
	Details           map[string]interface{}
	DistanceFromEvent *float64
	FacialSimilarity  *float64
	WithinBounds      *bool
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

// AddDetail adiciona um detalhe customizado
func (vr *ValidationResult) AddDetail(key string, value interface{}) *ValidationResult {
	vr.Details[key] = value
	return vr
}
