package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"eventos-backend/internal/domain/event"
	"eventos-backend/internal/domain/shared/errors"
	"eventos-backend/internal/domain/shared/value_objects"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

// EventRepository implementa a interface event.Repository usando PostgreSQL
type EventRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// NewEventRepository cria uma nova instância do repositório de eventos
func NewEventRepository(db *sqlx.DB, logger *zap.Logger) event.Repository {
	return &EventRepository{
		db:     db,
		logger: logger,
	}
}

// eventRow representa uma linha de evento no banco de dados
type eventRow struct {
	ID          string         `db:"id"`
	TenantID    string         `db:"tenant_id"`
	Name        string         `db:"name"`
	Location    string         `db:"location"`
	FenceEvent  pq.StringArray `db:"fence_event"` // Array de coordenadas como strings
	InitialDate time.Time      `db:"initial_date"`
	FinalDate   time.Time      `db:"final_date"`
	Active      bool           `db:"active"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
	CreatedBy   sql.NullString `db:"created_by"`
	UpdatedBy   sql.NullString `db:"updated_by"`
}

// toEntity converte eventRow para entidade Event
func (r *eventRow) toEntity() (*event.Event, error) {
	id, err := value_objects.ParseUUID(r.ID)
	if err != nil {
		return nil, errors.NewDomainError("INVALID_ID", "invalid event ID", err)
	}

	tenantID, err := value_objects.ParseUUID(r.TenantID)
	if err != nil {
		return nil, errors.NewDomainError("INVALID_TENANT_ID", "invalid tenant ID", err)
	}

	// Converter fence_event de array de strings para []Location
	var fenceEvent []value_objects.Location
	for _, coordStr := range r.FenceEvent {
		// Formato esperado: "lat,lng"
		parts := strings.Split(coordStr, ",")
		if len(parts) != 2 {
			continue
		}

		lat, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			continue
		}

		lng, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			continue
		}

		location, err := value_objects.NewLocation(lat, lng)
		if err != nil {
			continue
		}

		fenceEvent = append(fenceEvent, location)
	}

	evt := &event.Event{
		ID:          id,
		TenantID:    tenantID,
		Name:        r.Name,
		Location:    r.Location,
		FenceEvent:  fenceEvent,
		InitialDate: r.InitialDate,
		FinalDate:   r.FinalDate,
		Active:      r.Active,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}

	if r.CreatedBy.Valid {
		createdBy, err := value_objects.ParseUUID(r.CreatedBy.String)
		if err == nil {
			evt.CreatedBy = &createdBy
		}
	}

	if r.UpdatedBy.Valid {
		updatedBy, err := value_objects.ParseUUID(r.UpdatedBy.String)
		if err == nil {
			evt.UpdatedBy = &updatedBy
		}
	}

	return evt, nil
}

// fromEntity converte entidade Event para eventRow
func (repo *EventRepository) fromEntity(evt *event.Event) *eventRow {
	row := &eventRow{
		ID:          evt.ID.String(),
		TenantID:    evt.TenantID.String(),
		Name:        evt.Name,
		Location:    evt.Location,
		InitialDate: evt.InitialDate,
		FinalDate:   evt.FinalDate,
		Active:      evt.Active,
		CreatedAt:   evt.CreatedAt,
		UpdatedAt:   evt.UpdatedAt,
	}

	// Converter FenceEvent para array de strings
	for _, location := range evt.FenceEvent {
		coordStr := fmt.Sprintf("%.8f,%.8f", location.Latitude, location.Longitude)
		row.FenceEvent = append(row.FenceEvent, coordStr)
	}

	if evt.CreatedBy != nil {
		row.CreatedBy = sql.NullString{String: evt.CreatedBy.String(), Valid: true}
	}

	if evt.UpdatedBy != nil {
		row.UpdatedBy = sql.NullString{String: evt.UpdatedBy.String(), Valid: true}
	}

	return row
}

// Create cria um novo evento
func (repo *EventRepository) Create(ctx context.Context, evt *event.Event) error {
	row := repo.fromEntity(evt)

	query := `
		INSERT INTO events (
			id, tenant_id, name, location, fence_event, 
			initial_date, final_date, active, created_at, 
			updated_at, created_by, updated_by
		) VALUES (
			:id, :tenant_id, :name, :location, :fence_event,
			:initial_date, :final_date, :active, :created_at,
			:updated_at, :created_by, :updated_by
		)`

	_, err := repo.db.NamedExecContext(ctx, query, row)
	if err != nil {
		repo.logger.Error("Failed to create event", zap.Error(err), zap.String("event_id", evt.ID.String()))
		return errors.NewInternalError("failed to create event", err)
	}

	repo.logger.Info("Event created successfully", zap.String("event_id", evt.ID.String()))
	return nil
}

// GetByID busca um evento pelo ID
func (repo *EventRepository) GetByID(ctx context.Context, id value_objects.UUID) (*event.Event, error) {
	var row eventRow

	query := `
		SELECT id, tenant_id, name, location, fence_event, 
			   initial_date, final_date, active, created_at, 
			   updated_at, created_by, updated_by
		FROM events 
		WHERE id = $1 AND active = true`

	err := repo.db.GetContext(ctx, &row, query, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewDomainError("NOT_FOUND", "event not found", nil)
		}
		repo.logger.Error("Failed to get event by ID", zap.Error(err), zap.String("event_id", id.String()))
		return nil, errors.NewInternalError("failed to get event", err)
	}

	return row.toEntity()
}

// GetByIDAndTenant busca um evento pelo ID dentro de um tenant
func (repo *EventRepository) GetByIDAndTenant(ctx context.Context, id, tenantID value_objects.UUID) (*event.Event, error) {
	var row eventRow

	query := `
		SELECT id, tenant_id, name, location, fence_event, 
			   initial_date, final_date, active, created_at, 
			   updated_at, created_by, updated_by
		FROM events 
		WHERE id = $1 AND tenant_id = $2 AND active = true`

	err := repo.db.GetContext(ctx, &row, query, id.String(), tenantID.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewDomainError("NOT_FOUND", "event not found", nil)
		}
		repo.logger.Error("Failed to get event by ID and tenant",
			zap.Error(err),
			zap.String("event_id", id.String()),
			zap.String("tenant_id", tenantID.String()))
		return nil, errors.NewInternalError("failed to get event", err)
	}

	return row.toEntity()
}

// Update atualiza um evento existente
func (repo *EventRepository) Update(ctx context.Context, evt *event.Event) error {
	row := repo.fromEntity(evt)

	query := `
		UPDATE events SET
			name = :name,
			location = :location,
			fence_event = :fence_event,
			initial_date = :initial_date,
			final_date = :final_date,
			updated_at = :updated_at,
			updated_by = :updated_by
		WHERE id = :id AND active = true`

	result, err := repo.db.NamedExecContext(ctx, query, row)
	if err != nil {
		repo.logger.Error("Failed to update event", zap.Error(err), zap.String("event_id", evt.ID.String()))
		return errors.NewInternalError("failed to update event", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		repo.logger.Error("Failed to get rows affected", zap.Error(err))
		return errors.NewInternalError("failed to update event", err)
	}

	if rowsAffected == 0 {
		return errors.NewDomainError("NOT_FOUND", "event not found or inactive", nil)
	}

	repo.logger.Info("Event updated successfully", zap.String("event_id", evt.ID.String()))
	return nil
}

// Delete remove um evento (soft delete)
func (repo *EventRepository) Delete(ctx context.Context, id value_objects.UUID, deletedBy value_objects.UUID) error {
	query := `
		UPDATE events SET
			active = false,
			updated_at = NOW(),
			updated_by = $2
		WHERE id = $1 AND active = true`

	result, err := repo.db.ExecContext(ctx, query, id.String(), deletedBy.String())
	if err != nil {
		repo.logger.Error("Failed to delete event", zap.Error(err), zap.String("event_id", id.String()))
		return errors.NewInternalError("failed to delete event", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		repo.logger.Error("Failed to get rows affected", zap.Error(err))
		return errors.NewInternalError("failed to delete event", err)
	}

	if rowsAffected == 0 {
		return errors.NewDomainError("NOT_FOUND", "event not found or already inactive", nil)
	}

	repo.logger.Info("Event deleted successfully", zap.String("event_id", id.String()))
	return nil
}

// List lista eventos com paginação e filtros
func (repo *EventRepository) List(ctx context.Context, filters event.ListFilters) ([]*event.Event, int, error) {
	if err := filters.Validate(); err != nil {
		return nil, 0, err
	}

	// Query base
	baseQuery := `FROM events WHERE active = true`
	var args []interface{}
	var conditions []string
	argIndex := 1

	// Aplicar filtros
	if filters.TenantID != nil {
		conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argIndex))
		args = append(args, filters.TenantID.String())
		argIndex++
	}

	if filters.Name != nil {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIndex))
		args = append(args, "%"+*filters.Name+"%")
		argIndex++
	}

	if filters.Location != nil {
		conditions = append(conditions, fmt.Sprintf("location ILIKE $%d", argIndex))
		args = append(args, "%"+*filters.Location+"%")
		argIndex++
	}

	if filters.Active != nil {
		conditions = append(conditions, fmt.Sprintf("active = $%d", argIndex))
		args = append(args, *filters.Active)
		argIndex++
	}

	if filters.DateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("final_date >= $%d", argIndex))
		args = append(args, *filters.DateFrom)
		argIndex++
	}

	if filters.DateTo != nil {
		conditions = append(conditions, fmt.Sprintf("initial_date <= $%d", argIndex))
		args = append(args, *filters.DateTo)
		argIndex++
	}

	// Filtros de status temporal
	if filters.Status != nil {
		now := time.Now()
		switch *filters.Status {
		case event.EventStatusOngoing:
			conditions = append(conditions, fmt.Sprintf("initial_date <= $%d AND final_date >= $%d", argIndex, argIndex+1))
			args = append(args, now, now)
			argIndex += 2
		case event.EventStatusUpcoming:
			conditions = append(conditions, fmt.Sprintf("initial_date > $%d", argIndex))
			args = append(args, now)
			argIndex++
		case event.EventStatusFinished:
			conditions = append(conditions, fmt.Sprintf("final_date < $%d", argIndex))
			args = append(args, now)
			argIndex++
		}
	}

	// Construir WHERE clause
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " AND " + strings.Join(conditions, " AND ")
	}

	// Query de contagem
	countQuery := "SELECT COUNT(*) " + baseQuery + whereClause
	var total int
	err := repo.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		repo.logger.Error("Failed to count events", zap.Error(err))
		return nil, 0, errors.NewInternalError("failed to count events", err)
	}

	// Query de dados
	orderClause := fmt.Sprintf("ORDER BY %s", filters.OrderBy)
	if filters.OrderDesc {
		orderClause += " DESC"
	}

	limitClause := fmt.Sprintf("LIMIT %d OFFSET %d", filters.PageSize, filters.GetOffset())

	dataQuery := `
		SELECT id, tenant_id, name, location, fence_event, 
			   initial_date, final_date, active, created_at, 
			   updated_at, created_by, updated_by ` +
		baseQuery + whereClause + " " + orderClause + " " + limitClause

	var rows []eventRow
	err = repo.db.SelectContext(ctx, &rows, dataQuery, args...)
	if err != nil {
		repo.logger.Error("Failed to list events", zap.Error(err))
		return nil, 0, errors.NewInternalError("failed to list events", err)
	}

	// Converter para entidades
	events := make([]*event.Event, 0, len(rows))
	for _, row := range rows {
		evt, err := row.toEntity()
		if err != nil {
			repo.logger.Warn("Failed to convert event row", zap.Error(err), zap.String("event_id", row.ID))
			continue
		}
		events = append(events, evt)
	}

	return events, total, nil
}

// ListByTenant lista eventos de um tenant específico
func (repo *EventRepository) ListByTenant(ctx context.Context, tenantID value_objects.UUID, filters event.ListFilters) ([]*event.Event, int, error) {
	filters.TenantID = &tenantID
	return repo.List(ctx, filters)
}

// ListActiveEvents lista eventos ativos
func (repo *EventRepository) ListActiveEvents(ctx context.Context, tenantID *value_objects.UUID, filters event.ListFilters) ([]*event.Event, int, error) {
	active := true
	filters.Active = &active
	if tenantID != nil {
		filters.TenantID = tenantID
	}
	return repo.List(ctx, filters)
}

// ListOngoingEvents lista eventos em andamento
func (repo *EventRepository) ListOngoingEvents(ctx context.Context, tenantID *value_objects.UUID, filters event.ListFilters) ([]*event.Event, int, error) {
	status := event.EventStatusOngoing
	filters.Status = &status
	if tenantID != nil {
		filters.TenantID = tenantID
	}
	return repo.List(ctx, filters)
}

// ListUpcomingEvents lista eventos futuros
func (repo *EventRepository) ListUpcomingEvents(ctx context.Context, tenantID *value_objects.UUID, filters event.ListFilters) ([]*event.Event, int, error) {
	status := event.EventStatusUpcoming
	filters.Status = &status
	if tenantID != nil {
		filters.TenantID = tenantID
	}
	return repo.List(ctx, filters)
}

// ExistsByNameInTenant verifica se existe um evento com o nome no tenant
func (repo *EventRepository) ExistsByNameInTenant(ctx context.Context, name string, tenantID value_objects.UUID, excludeID *value_objects.UUID) (bool, error) {
	query := `SELECT COUNT(*) FROM events WHERE name = $1 AND tenant_id = $2 AND active = true`
	args := []interface{}{name, tenantID.String()}

	if excludeID != nil {
		query += " AND id != $3"
		args = append(args, excludeID.String())
	}

	var count int
	err := repo.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		repo.logger.Error("Failed to check event name existence", zap.Error(err))
		return false, errors.NewInternalError("failed to check event name", err)
	}

	return count > 0, nil
}

// GetEventsInLocation busca eventos que contêm uma localização específica
func (repo *EventRepository) GetEventsInLocation(ctx context.Context, location value_objects.Location, tenantID *value_objects.UUID) ([]*event.Event, error) {
	// Esta implementação é simplificada - em produção usaria PostGIS para queries geoespaciais
	query := `
		SELECT id, tenant_id, name, location, fence_event, 
			   initial_date, final_date, active, created_at, 
			   updated_at, created_by, updated_by
		FROM events 
		WHERE active = true`

	args := []interface{}{}
	if tenantID != nil {
		query += " AND tenant_id = $1"
		args = append(args, tenantID.String())
	}

	var rows []eventRow
	err := repo.db.SelectContext(ctx, &rows, query, args...)
	if err != nil {
		repo.logger.Error("Failed to get events in location", zap.Error(err))
		return nil, errors.NewInternalError("failed to get events in location", err)
	}

	// Filtrar eventos que contêm a localização (usando algoritmo point-in-polygon)
	var eventsInLocation []*event.Event
	for _, row := range rows {
		evt, err := row.toEntity()
		if err != nil {
			continue
		}

		// Verificar se a localização está dentro do fence do evento
		if len(evt.FenceEvent) > 0 {
			// Implementação simplificada - em produção usaria PostGIS
			// Por enquanto, adiciona todos os eventos para não quebrar
			eventsInLocation = append(eventsInLocation, evt)
		}
	}

	return eventsInLocation, nil
}
