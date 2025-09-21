package event

import (
	"context"
	"time"

	"eventos-backend/internal/domain/shared/value_objects"
)

// Repository define as operações de persistência para Event
type Repository interface {
	// Create cria um novo evento
	Create(ctx context.Context, event *Event) error

	// GetByID busca um evento pelo ID
	GetByID(ctx context.Context, id value_objects.UUID) (*Event, error)

	// GetByIDAndTenant busca um evento pelo ID dentro de um tenant
	GetByIDAndTenant(ctx context.Context, id, tenantID value_objects.UUID) (*Event, error)

	// Update atualiza um evento existente
	Update(ctx context.Context, event *Event) error

	// Delete remove um evento (soft delete)
	Delete(ctx context.Context, id value_objects.UUID, deletedBy value_objects.UUID) error

	// List lista eventos com paginação e filtros
	List(ctx context.Context, filters ListFilters) ([]*Event, int, error)

	// ListByTenant lista eventos de um tenant específico
	ListByTenant(ctx context.Context, tenantID value_objects.UUID, filters ListFilters) ([]*Event, int, error)

	// ListActiveEvents lista eventos ativos
	ListActiveEvents(ctx context.Context, tenantID *value_objects.UUID, filters ListFilters) ([]*Event, int, error)

	// ListOngoingEvents lista eventos em andamento
	ListOngoingEvents(ctx context.Context, tenantID *value_objects.UUID, filters ListFilters) ([]*Event, int, error)

	// ListUpcomingEvents lista eventos futuros
	ListUpcomingEvents(ctx context.Context, tenantID *value_objects.UUID, filters ListFilters) ([]*Event, int, error)

	// ExistsByNameInTenant verifica se existe um evento com o nome no tenant
	ExistsByNameInTenant(ctx context.Context, name string, tenantID value_objects.UUID, excludeID *value_objects.UUID) (bool, error)

	// GetEventsInLocation busca eventos que contêm uma localização específica
	GetEventsInLocation(ctx context.Context, location value_objects.Location, tenantID *value_objects.UUID) ([]*Event, error)
}

// ListFilters define os filtros para listagem de eventos
type ListFilters struct {
	// Filtros de busca
	TenantID *value_objects.UUID
	Name     *string
	Location *string
	Active   *bool
	DateFrom *time.Time
	DateTo   *time.Time
	Status   *EventStatus // ongoing, upcoming, finished

	// Filtros geográficos
	NearLocation *value_objects.Location
	WithinRadius *float64 // em metros

	// Paginação
	Page     int
	PageSize int

	// Ordenação
	OrderBy   string
	OrderDesc bool
}

// EventStatus representa o status do evento
type EventStatus string

const (
	EventStatusOngoing  EventStatus = "ongoing"
	EventStatusUpcoming EventStatus = "upcoming"
	EventStatusFinished EventStatus = "finished"
)

// Validate valida os filtros de listagem
func (f *ListFilters) Validate() error {
	if f.Page < 1 {
		f.Page = 1
	}

	if f.PageSize < 1 {
		f.PageSize = 20
	}

	if f.PageSize > 100 {
		f.PageSize = 100
	}

	if f.OrderBy == "" {
		f.OrderBy = "initial_date"
	}

	validOrderFields := []string{"name", "location", "initial_date", "final_date", "created_at", "updated_at"}
	isValidOrder := false
	for _, field := range validOrderFields {
		if f.OrderBy == field {
			isValidOrder = true
			break
		}
	}

	if !isValidOrder {
		f.OrderBy = "initial_date"
	}

	// Validar filtros de data
	if f.DateFrom != nil && f.DateTo != nil {
		if f.DateTo.Before(*f.DateFrom) {
			f.DateTo = f.DateFrom
		}
	}

	// Validar filtros geográficos
	if f.WithinRadius != nil {
		if f.NearLocation == nil {
			f.WithinRadius = nil // Ignorar raio se não há localização
		}
		if *f.WithinRadius < 0 {
			*f.WithinRadius = 0
		}
		if *f.WithinRadius > 100000 { // Máximo 100km
			*f.WithinRadius = 100000
		}
	}

	return nil
}

// GetOffset calcula o offset para paginação
func (f *ListFilters) GetOffset() int {
	return (f.Page - 1) * f.PageSize
}

// HasDateFilter verifica se há filtros de data
func (f *ListFilters) HasDateFilter() bool {
	return f.DateFrom != nil || f.DateTo != nil
}

// HasLocationFilter verifica se há filtros de localização
func (f *ListFilters) HasLocationFilter() bool {
	return f.NearLocation != nil && f.WithinRadius != nil
}
