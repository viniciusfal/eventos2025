package event

import (
	"time"

	"eventos-backend/internal/domain/shared/errors"
	"eventos-backend/internal/domain/shared/value_objects"
)

// Event representa um evento no sistema
type Event struct {
	ID          value_objects.UUID
	TenantID    value_objects.UUID
	Name        string
	Location    string
	FenceEvent  []value_objects.Location // Polígono que define a área do evento
	InitialDate time.Time
	FinalDate   time.Time
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   *value_objects.UUID
	UpdatedBy   *value_objects.UUID
}

// NewEvent cria uma nova instância de Event
func NewEvent(tenantID value_objects.UUID, name, location string, fenceEvent []value_objects.Location, initialDate, finalDate time.Time, createdBy value_objects.UUID) (*Event, error) {
	if err := validateEventData(name, location, fenceEvent, initialDate, finalDate); err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	return &Event{
		ID:          value_objects.NewUUID(),
		TenantID:    tenantID,
		Name:        name,
		Location:    location,
		FenceEvent:  fenceEvent,
		InitialDate: initialDate,
		FinalDate:   finalDate,
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
		CreatedBy:   &createdBy,
		UpdatedBy:   &createdBy,
	}, nil
}

// Update atualiza os dados do evento
func (e *Event) Update(name, location string, fenceEvent []value_objects.Location, initialDate, finalDate time.Time, updatedBy value_objects.UUID) error {
	if err := validateEventData(name, location, fenceEvent, initialDate, finalDate); err != nil {
		return err
	}

	e.Name = name
	e.Location = location
	e.FenceEvent = fenceEvent
	e.InitialDate = initialDate
	e.FinalDate = finalDate
	e.UpdatedAt = time.Now().UTC()
	e.UpdatedBy = &updatedBy

	return nil
}

// Activate ativa o evento
func (e *Event) Activate(updatedBy value_objects.UUID) {
	e.Active = true
	e.UpdatedAt = time.Now().UTC()
	e.UpdatedBy = &updatedBy
}

// Deactivate desativa o evento
func (e *Event) Deactivate(updatedBy value_objects.UUID) {
	e.Active = false
	e.UpdatedAt = time.Now().UTC()
	e.UpdatedBy = &updatedBy
}

// IsActive verifica se o evento está ativo
func (e *Event) IsActive() bool {
	return e.Active
}

// IsOngoing verifica se o evento está em andamento (dentro do período)
func (e *Event) IsOngoing() bool {
	now := time.Now().UTC()
	return e.Active && now.After(e.InitialDate) && now.Before(e.FinalDate)
}

// IsUpcoming verifica se o evento é futuro
func (e *Event) IsUpcoming() bool {
	now := time.Now().UTC()
	return e.Active && now.Before(e.InitialDate)
}

// IsFinished verifica se o evento já terminou
func (e *Event) IsFinished() bool {
	now := time.Now().UTC()
	return now.After(e.FinalDate)
}

// GetDuration retorna a duração do evento
func (e *Event) GetDuration() time.Duration {
	return e.FinalDate.Sub(e.InitialDate)
}

// BelongsToTenant verifica se o evento pertence ao tenant informado
func (e *Event) BelongsToTenant(tenantID value_objects.UUID) bool {
	return e.TenantID.Equals(tenantID)
}

// IsLocationWithinFence verifica se uma localização está dentro da cerca do evento
func (e *Event) IsLocationWithinFence(location value_objects.Location) bool {
	if len(e.FenceEvent) < 3 {
		// Se não há cerca definida ou é inválida, considerar que está dentro
		return true
	}

	return isPointInPolygon(location, e.FenceEvent)
}

// CanCheckIn verifica se é possível fazer check-in no evento
func (e *Event) CanCheckIn() error {
	if !e.IsActive() {
		return errors.NewValidationError("event", "event is not active")
	}

	if e.IsFinished() {
		return errors.NewValidationError("event", "event has already finished")
	}

	if !e.IsOngoing() && !e.IsUpcoming() {
		return errors.NewValidationError("event", "event is not available for check-in")
	}

	return nil
}

// CanCheckOut verifica se é possível fazer check-out no evento
func (e *Event) CanCheckOut() error {
	if !e.IsActive() {
		return errors.NewValidationError("event", "event is not active")
	}

	if !e.IsOngoing() {
		return errors.NewValidationError("event", "event is not ongoing")
	}

	return nil
}

// validateEventData valida os dados básicos do evento
func validateEventData(name, location string, fenceEvent []value_objects.Location, initialDate, finalDate time.Time) error {
	if name == "" {
		return errors.NewValidationError("name", "event name is required")
	}

	if len(name) < 3 || len(name) > 255 {
		return errors.NewValidationError("name", "event name must be between 3 and 255 characters")
	}

	if location == "" {
		return errors.NewValidationError("location", "event location is required")
	}

	if len(location) > 500 {
		return errors.NewValidationError("location", "event location must be at most 500 characters")
	}

	if initialDate.IsZero() {
		return errors.NewValidationError("initial_date", "initial date is required")
	}

	if finalDate.IsZero() {
		return errors.NewValidationError("final_date", "final date is required")
	}

	if finalDate.Before(initialDate) || finalDate.Equal(initialDate) {
		return errors.NewValidationError("final_date", "final date must be after initial date")
	}

	// Validar que o evento não seja muito longo (máximo 30 dias)
	maxDuration := 30 * 24 * time.Hour
	if finalDate.Sub(initialDate) > maxDuration {
		return errors.NewValidationError("duration", "event duration cannot exceed 30 days")
	}

	// Validar cerca do evento se fornecida
	if len(fenceEvent) > 0 {
		if len(fenceEvent) < 3 {
			return errors.NewValidationError("fence_event", "fence must have at least 3 points to form a polygon")
		}

		if len(fenceEvent) > 100 {
			return errors.NewValidationError("fence_event", "fence cannot have more than 100 points")
		}

		// Verificar se o primeiro e último ponto são iguais (polígono fechado)
		first := fenceEvent[0]
		last := fenceEvent[len(fenceEvent)-1]
		if first.Latitude != last.Latitude || first.Longitude != last.Longitude {
			// Adicionar o primeiro ponto no final para fechar o polígono
			fenceEvent = append(fenceEvent, first)
		}
	}

	return nil
}

// isPointInPolygon verifica se um ponto está dentro de um polígono usando o algoritmo Ray Casting
func isPointInPolygon(point value_objects.Location, polygon []value_objects.Location) bool {
	if len(polygon) < 3 {
		return false
	}

	inside := false
	j := len(polygon) - 1

	for i := 0; i < len(polygon); i++ {
		xi, yi := polygon[i].Longitude, polygon[i].Latitude
		xj, yj := polygon[j].Longitude, polygon[j].Latitude

		if ((yi > point.Latitude) != (yj > point.Latitude)) &&
			(point.Longitude < (xj-xi)*(point.Latitude-yi)/(yj-yi)+xi) {
			inside = !inside
		}
		j = i
	}

	return inside
}
