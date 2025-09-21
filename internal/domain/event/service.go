package event

import (
	"context"
	"time"

	"eventos-backend/internal/domain/shared/errors"
	"eventos-backend/internal/domain/shared/value_objects"

	"go.uber.org/zap"
)

// Service define os serviços de domínio para Event
type Service interface {
	// CreateEvent cria um novo evento com validações de negócio
	CreateEvent(ctx context.Context, tenantID value_objects.UUID, name, location string, fenceEvent []value_objects.Location, initialDate, finalDate time.Time, createdBy value_objects.UUID) (*Event, error)

	// UpdateEvent atualiza um evento existente
	UpdateEvent(ctx context.Context, id value_objects.UUID, name, location string, fenceEvent []value_objects.Location, initialDate, finalDate time.Time, updatedBy value_objects.UUID) (*Event, error)

	// GetEvent busca um evento pelo ID
	GetEvent(ctx context.Context, id value_objects.UUID) (*Event, error)

	// GetEventByTenant busca um evento pelo ID dentro de um tenant
	GetEventByTenant(ctx context.Context, id, tenantID value_objects.UUID) (*Event, error)

	// ActivateEvent ativa um evento
	ActivateEvent(ctx context.Context, id value_objects.UUID, updatedBy value_objects.UUID) error

	// DeactivateEvent desativa um evento
	DeactivateEvent(ctx context.Context, id value_objects.UUID, updatedBy value_objects.UUID) error

	// ListEvents lista eventos com filtros
	ListEvents(ctx context.Context, filters ListFilters) ([]*Event, int, error)

	// ListEventsByTenant lista eventos de um tenant específico
	ListEventsByTenant(ctx context.Context, tenantID value_objects.UUID, filters ListFilters) ([]*Event, int, error)

	// ListActiveEvents lista eventos ativos
	ListActiveEvents(ctx context.Context, tenantID *value_objects.UUID, filters ListFilters) ([]*Event, int, error)

	// ListOngoingEvents lista eventos em andamento
	ListOngoingEvents(ctx context.Context, tenantID *value_objects.UUID, filters ListFilters) ([]*Event, int, error)

	// ListUpcomingEvents lista eventos futuros
	ListUpcomingEvents(ctx context.Context, tenantID *value_objects.UUID, filters ListFilters) ([]*Event, int, error)

	// DeleteEvent remove um evento (soft delete)
	DeleteEvent(ctx context.Context, id value_objects.UUID, deletedBy value_objects.UUID) error

	// ValidateEventForCheckIn valida se um evento está disponível para check-in
	ValidateEventForCheckIn(ctx context.Context, eventID value_objects.UUID, location *value_objects.Location) (*Event, error)

	// ValidateEventForCheckOut valida se um evento está disponível para check-out
	ValidateEventForCheckOut(ctx context.Context, eventID value_objects.UUID, location *value_objects.Location) (*Event, error)

	// GetEventsInLocation busca eventos que contêm uma localização específica
	GetEventsInLocation(ctx context.Context, location value_objects.Location, tenantID *value_objects.UUID) ([]*Event, error)
}

// DomainService implementa os serviços de domínio para Event
type DomainService struct {
	repository Repository
	logger     *zap.Logger
}

// NewDomainService cria uma nova instância do serviço de domínio
func NewDomainService(repository Repository, logger *zap.Logger) Service {
	return &DomainService{
		repository: repository,
		logger:     logger,
	}
}

// CreateEvent cria um novo evento com validações de negócio
func (s *DomainService) CreateEvent(ctx context.Context, tenantID value_objects.UUID, name, location string, fenceEvent []value_objects.Location, initialDate, finalDate time.Time, createdBy value_objects.UUID) (*Event, error) {
	s.logger.Debug("Creating new event",
		zap.String("tenant_id", tenantID.String()),
		zap.String("name", name),
		zap.String("location", location),
		zap.Time("initial_date", initialDate),
		zap.Time("final_date", finalDate),
		zap.String("created_by", createdBy.String()),
	)

	// Verificar se já existe evento com o mesmo nome no tenant
	exists, err := s.repository.ExistsByNameInTenant(ctx, name, tenantID, nil)
	if err != nil {
		s.logger.Error("Failed to check event name uniqueness in tenant", zap.Error(err))
		return nil, errors.NewInternalError("failed to validate event name uniqueness", err)
	}
	if exists {
		return nil, errors.NewAlreadyExistsError("event", "name", name)
	}

	// Criar nova instância do evento
	event, err := NewEvent(tenantID, name, location, fenceEvent, initialDate, finalDate, createdBy)
	if err != nil {
		s.logger.Error("Failed to create event instance", zap.Error(err))
		return nil, err
	}

	// Persistir no repositório
	if err := s.repository.Create(ctx, event); err != nil {
		s.logger.Error("Failed to persist event", zap.Error(err))
		return nil, errors.NewInternalError("failed to create event", err)
	}

	s.logger.Info("Event created successfully",
		zap.String("event_id", event.ID.String()),
		zap.String("name", event.Name),
		zap.String("tenant_id", event.TenantID.String()),
	)

	return event, nil
}

// UpdateEvent atualiza um evento existente
func (s *DomainService) UpdateEvent(ctx context.Context, id value_objects.UUID, name, location string, fenceEvent []value_objects.Location, initialDate, finalDate time.Time, updatedBy value_objects.UUID) (*Event, error) {
	s.logger.Debug("Updating event",
		zap.String("event_id", id.String()),
		zap.String("name", name),
		zap.String("updated_by", updatedBy.String()),
	)

	// Buscar evento existente
	event, err := s.repository.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get event for update", zap.Error(err))
		return nil, errors.NewInternalError("failed to get event", err)
	}
	if event == nil {
		return nil, errors.NewNotFoundError("event", id.String())
	}

	// Verificar unicidade do nome no tenant (se alterado)
	if name != event.Name {
		exists, err := s.repository.ExistsByNameInTenant(ctx, name, event.TenantID, &id)
		if err != nil {
			s.logger.Error("Failed to check event name uniqueness in tenant", zap.Error(err))
			return nil, errors.NewInternalError("failed to validate event name uniqueness", err)
		}
		if exists {
			return nil, errors.NewAlreadyExistsError("event", "name", name)
		}
	}

	// Verificar se o evento pode ser atualizado (regras de negócio)
	if event.IsOngoing() {
		// Eventos em andamento têm restrições de atualização
		if !initialDate.Equal(event.InitialDate) {
			return nil, errors.NewValidationError("initial_date", "cannot change initial date of ongoing event")
		}

		// Só pode estender a data final, não reduzir
		if finalDate.Before(event.FinalDate) {
			return nil, errors.NewValidationError("final_date", "cannot reduce final date of ongoing event")
		}
	}

	// Atualizar dados do evento
	if err := event.Update(name, location, fenceEvent, initialDate, finalDate, updatedBy); err != nil {
		s.logger.Error("Failed to update event data", zap.Error(err))
		return nil, err
	}

	// Persistir alterações
	if err := s.repository.Update(ctx, event); err != nil {
		s.logger.Error("Failed to persist event update", zap.Error(err))
		return nil, errors.NewInternalError("failed to update event", err)
	}

	s.logger.Info("Event updated successfully",
		zap.String("event_id", event.ID.String()),
	)

	return event, nil
}

// GetEvent busca um evento pelo ID
func (s *DomainService) GetEvent(ctx context.Context, id value_objects.UUID) (*Event, error) {
	event, err := s.repository.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get event", zap.Error(err))
		return nil, errors.NewInternalError("failed to get event", err)
	}
	if event == nil {
		return nil, errors.NewNotFoundError("event", id.String())
	}

	return event, nil
}

// GetEventByTenant busca um evento pelo ID dentro de um tenant
func (s *DomainService) GetEventByTenant(ctx context.Context, id, tenantID value_objects.UUID) (*Event, error) {
	event, err := s.repository.GetByIDAndTenant(ctx, id, tenantID)
	if err != nil {
		s.logger.Error("Failed to get event by tenant", zap.Error(err))
		return nil, errors.NewInternalError("failed to get event", err)
	}
	if event == nil {
		return nil, errors.NewNotFoundError("event", id.String())
	}

	return event, nil
}

// ActivateEvent ativa um evento
func (s *DomainService) ActivateEvent(ctx context.Context, id value_objects.UUID, updatedBy value_objects.UUID) error {
	event, err := s.GetEvent(ctx, id)
	if err != nil {
		return err
	}

	if event.IsActive() {
		return errors.NewValidationError("status", "event is already active")
	}

	event.Activate(updatedBy)

	if err := s.repository.Update(ctx, event); err != nil {
		s.logger.Error("Failed to activate event", zap.Error(err))
		return errors.NewInternalError("failed to activate event", err)
	}

	s.logger.Info("Event activated successfully",
		zap.String("event_id", id.String()),
	)

	return nil
}

// DeactivateEvent desativa um evento
func (s *DomainService) DeactivateEvent(ctx context.Context, id value_objects.UUID, updatedBy value_objects.UUID) error {
	event, err := s.GetEvent(ctx, id)
	if err != nil {
		return err
	}

	if !event.IsActive() {
		return errors.NewValidationError("status", "event is already inactive")
	}

	// Verificar se o evento pode ser desativado
	if event.IsOngoing() {
		return errors.NewValidationError("status", "cannot deactivate ongoing event")
	}

	event.Deactivate(updatedBy)

	if err := s.repository.Update(ctx, event); err != nil {
		s.logger.Error("Failed to deactivate event", zap.Error(err))
		return errors.NewInternalError("failed to deactivate event", err)
	}

	s.logger.Info("Event deactivated successfully",
		zap.String("event_id", id.String()),
	)

	return nil
}

// ListEvents lista eventos com filtros
func (s *DomainService) ListEvents(ctx context.Context, filters ListFilters) ([]*Event, int, error) {
	if err := filters.Validate(); err != nil {
		return nil, 0, err
	}

	events, total, err := s.repository.List(ctx, filters)
	if err != nil {
		s.logger.Error("Failed to list events", zap.Error(err))
		return nil, 0, errors.NewInternalError("failed to list events", err)
	}

	return events, total, nil
}

// ListEventsByTenant lista eventos de um tenant específico
func (s *DomainService) ListEventsByTenant(ctx context.Context, tenantID value_objects.UUID, filters ListFilters) ([]*Event, int, error) {
	if err := filters.Validate(); err != nil {
		return nil, 0, err
	}

	events, total, err := s.repository.ListByTenant(ctx, tenantID, filters)
	if err != nil {
		s.logger.Error("Failed to list events by tenant", zap.Error(err))
		return nil, 0, errors.NewInternalError("failed to list events", err)
	}

	return events, total, nil
}

// ListActiveEvents lista eventos ativos
func (s *DomainService) ListActiveEvents(ctx context.Context, tenantID *value_objects.UUID, filters ListFilters) ([]*Event, int, error) {
	if err := filters.Validate(); err != nil {
		return nil, 0, err
	}

	events, total, err := s.repository.ListActiveEvents(ctx, tenantID, filters)
	if err != nil {
		s.logger.Error("Failed to list active events", zap.Error(err))
		return nil, 0, errors.NewInternalError("failed to list active events", err)
	}

	return events, total, nil
}

// ListOngoingEvents lista eventos em andamento
func (s *DomainService) ListOngoingEvents(ctx context.Context, tenantID *value_objects.UUID, filters ListFilters) ([]*Event, int, error) {
	if err := filters.Validate(); err != nil {
		return nil, 0, err
	}

	events, total, err := s.repository.ListOngoingEvents(ctx, tenantID, filters)
	if err != nil {
		s.logger.Error("Failed to list ongoing events", zap.Error(err))
		return nil, 0, errors.NewInternalError("failed to list ongoing events", err)
	}

	return events, total, nil
}

// ListUpcomingEvents lista eventos futuros
func (s *DomainService) ListUpcomingEvents(ctx context.Context, tenantID *value_objects.UUID, filters ListFilters) ([]*Event, int, error) {
	if err := filters.Validate(); err != nil {
		return nil, 0, err
	}

	events, total, err := s.repository.ListUpcomingEvents(ctx, tenantID, filters)
	if err != nil {
		s.logger.Error("Failed to list upcoming events", zap.Error(err))
		return nil, 0, errors.NewInternalError("failed to list upcoming events", err)
	}

	return events, total, nil
}

// DeleteEvent remove um evento (soft delete)
func (s *DomainService) DeleteEvent(ctx context.Context, id value_objects.UUID, deletedBy value_objects.UUID) error {
	// Verificar se o evento existe
	event, err := s.GetEvent(ctx, id)
	if err != nil {
		return err
	}

	// Verificar se o evento pode ser removido (regras de negócio)
	if event.IsOngoing() {
		return errors.NewValidationError("status", "cannot delete ongoing event")
	}

	if event.IsActive() {
		return errors.NewValidationError("status", "cannot delete active event")
	}

	if err := s.repository.Delete(ctx, id, deletedBy); err != nil {
		s.logger.Error("Failed to delete event", zap.Error(err))
		return errors.NewInternalError("failed to delete event", err)
	}

	s.logger.Info("Event deleted successfully",
		zap.String("event_id", id.String()),
	)

	return nil
}

// ValidateEventForCheckIn valida se um evento está disponível para check-in
func (s *DomainService) ValidateEventForCheckIn(ctx context.Context, eventID value_objects.UUID, location *value_objects.Location) (*Event, error) {
	event, err := s.GetEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}

	// Verificar se o evento permite check-in
	if err := event.CanCheckIn(); err != nil {
		return nil, err
	}

	// Verificar localização se fornecida
	if location != nil && !event.IsLocationWithinFence(*location) {
		return nil, errors.NewValidationError("location", "location is outside event fence")
	}

	return event, nil
}

// ValidateEventForCheckOut valida se um evento está disponível para check-out
func (s *DomainService) ValidateEventForCheckOut(ctx context.Context, eventID value_objects.UUID, location *value_objects.Location) (*Event, error) {
	event, err := s.GetEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}

	// Verificar se o evento permite check-out
	if err := event.CanCheckOut(); err != nil {
		return nil, err
	}

	// Verificar localização se fornecida
	if location != nil && !event.IsLocationWithinFence(*location) {
		return nil, errors.NewValidationError("location", "location is outside event fence")
	}

	return event, nil
}

// GetEventsInLocation busca eventos que contêm uma localização específica
func (s *DomainService) GetEventsInLocation(ctx context.Context, location value_objects.Location, tenantID *value_objects.UUID) ([]*Event, error) {
	events, err := s.repository.GetEventsInLocation(ctx, location, tenantID)
	if err != nil {
		s.logger.Error("Failed to get events in location", zap.Error(err))
		return nil, errors.NewInternalError("failed to get events in location", err)
	}

	return events, nil
}
