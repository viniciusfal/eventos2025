package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"eventos-backend/internal/domain/checkin"
	"eventos-backend/internal/domain/shared/value_objects"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// CheckinRepository implementa a interface de repositório para Checkin
type CheckinRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// NewCheckinRepository cria uma nova instância do repositório de checkin
func NewCheckinRepository(db *sqlx.DB, logger *zap.Logger) checkin.Repository {
	return &CheckinRepository{
		db:     db,
		logger: logger,
	}
}

// checkinRow representa uma linha da tabela checkin no banco
type checkinRow struct {
	ID                string         `db:"id_checkin"`
	TenantID          string         `db:"id_tenant"`
	EventID           string         `db:"id_event"`
	EmployeeID        string         `db:"id_employee"`
	PartnerID         string         `db:"id_partner"`
	Method            string         `db:"method"`
	Latitude          float64        `db:"latitude"`
	Longitude         float64        `db:"longitude"`
	CheckinTime       time.Time      `db:"checkin_time"`
	PhotoURL          sql.NullString `db:"photo_url"`
	Notes             sql.NullString `db:"notes"`
	IsValid           bool           `db:"is_valid"`
	ValidationDetails sql.NullString `db:"validation_details"`
	CreatedAt         time.Time      `db:"created_at"`
	UpdatedAt         time.Time      `db:"updated_at"`
	CreatedBy         sql.NullString `db:"created_by"`
	UpdatedBy         sql.NullString `db:"updated_by"`
}

// toEntity converte uma linha do banco para entidade de domínio
func (r *checkinRow) toEntity() (*checkin.Checkin, error) {
	id, err := value_objects.ParseUUID(r.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid checkin ID: %w", err)
	}

	tenantID, err := value_objects.ParseUUID(r.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	eventID, err := value_objects.ParseUUID(r.EventID)
	if err != nil {
		return nil, fmt.Errorf("invalid event ID: %w", err)
	}

	employeeID, err := value_objects.ParseUUID(r.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("invalid employee ID: %w", err)
	}

	partnerID, err := value_objects.ParseUUID(r.PartnerID)
	if err != nil {
		return nil, fmt.Errorf("invalid partner ID: %w", err)
	}

	location, err := value_objects.NewLocation(r.Latitude, r.Longitude)
	if err != nil {
		return nil, fmt.Errorf("invalid location: %w", err)
	}

	checkinEntity := &checkin.Checkin{
		ID:          id,
		TenantID:    tenantID,
		EventID:     eventID,
		EmployeeID:  employeeID,
		PartnerID:   partnerID,
		Method:      r.Method,
		Location:    location,
		CheckinTime: r.CheckinTime,
		IsValid:     r.IsValid,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}

	// PhotoURL
	if r.PhotoURL.Valid {
		checkinEntity.PhotoURL = r.PhotoURL.String
	}

	// Notes
	if r.Notes.Valid {
		checkinEntity.Notes = r.Notes.String
	}

	// ValidationDetails
	if r.ValidationDetails.Valid && r.ValidationDetails.String != "" {
		var details map[string]interface{}
		if err := json.Unmarshal([]byte(r.ValidationDetails.String), &details); err == nil {
			checkinEntity.ValidationDetails = details
		} else {
			checkinEntity.ValidationDetails = make(map[string]interface{})
		}
	} else {
		checkinEntity.ValidationDetails = make(map[string]interface{})
	}

	// CreatedBy
	if r.CreatedBy.Valid {
		createdBy, err := value_objects.ParseUUID(r.CreatedBy.String)
		if err == nil {
			checkinEntity.CreatedBy = &createdBy
		}
	}

	// UpdatedBy
	if r.UpdatedBy.Valid {
		updatedBy, err := value_objects.ParseUUID(r.UpdatedBy.String)
		if err == nil {
			checkinEntity.UpdatedBy = &updatedBy
		}
	}

	return checkinEntity, nil
}

// fromEntity converte uma entidade de domínio para linha do banco
func (repo *CheckinRepository) fromEntity(c *checkin.Checkin) *checkinRow {
	row := &checkinRow{
		ID:          c.ID.String(),
		TenantID:    c.TenantID.String(),
		EventID:     c.EventID.String(),
		EmployeeID:  c.EmployeeID.String(),
		PartnerID:   c.PartnerID.String(),
		Method:      c.Method,
		Latitude:    c.Location.Latitude,
		Longitude:   c.Location.Longitude,
		CheckinTime: c.CheckinTime,
		IsValid:     c.IsValid,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}

	// PhotoURL
	if c.PhotoURL != "" {
		row.PhotoURL = sql.NullString{String: c.PhotoURL, Valid: true}
	}

	// Notes
	if c.Notes != "" {
		row.Notes = sql.NullString{String: c.Notes, Valid: true}
	}

	// ValidationDetails
	if len(c.ValidationDetails) > 0 {
		if detailsJSON, err := json.Marshal(c.ValidationDetails); err == nil {
			row.ValidationDetails = sql.NullString{String: string(detailsJSON), Valid: true}
		}
	}

	// CreatedBy
	if c.CreatedBy != nil {
		row.CreatedBy = sql.NullString{String: c.CreatedBy.String(), Valid: true}
	}

	// UpdatedBy
	if c.UpdatedBy != nil {
		row.UpdatedBy = sql.NullString{String: c.UpdatedBy.String(), Valid: true}
	}

	return row
}

// Create cria um novo checkin
func (repo *CheckinRepository) Create(ctx context.Context, c *checkin.Checkin) error {
	row := repo.fromEntity(c)

	query := `
		INSERT INTO checkin (
			id_checkin, id_tenant, id_event, id_employee, id_partner,
			method, latitude, longitude, checkin_time, photo_url, notes,
			is_valid, validation_details, created_at, updated_at, created_by, updated_by
		) VALUES (
			:id_checkin, :id_tenant, :id_event, :id_employee, :id_partner,
			:method, :latitude, :longitude, :checkin_time, :photo_url, :notes,
			:is_valid, :validation_details, :created_at, :updated_at, :created_by, :updated_by
		)`

	_, err := repo.db.NamedExecContext(ctx, query, row)
	if err != nil {
		repo.logger.Error("Failed to create checkin", zap.Error(err), zap.String("checkin_id", c.ID.String()))
		return fmt.Errorf("failed to create checkin: %w", err)
	}

	repo.logger.Info("Checkin created successfully", zap.String("checkin_id", c.ID.String()))
	return nil
}

// GetByID busca um checkin por ID
func (repo *CheckinRepository) GetByID(ctx context.Context, id value_objects.UUID) (*checkin.Checkin, error) {
	var row checkinRow
	query := `
		SELECT id_checkin, id_tenant, id_event, id_employee, id_partner,
			   method, latitude, longitude, checkin_time, photo_url, notes,
			   is_valid, validation_details, created_at, updated_at, created_by, updated_by
		FROM checkin 
		WHERE id_checkin = $1`

	err := repo.db.GetContext(ctx, &row, query, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("checkin not found")
		}
		repo.logger.Error("Failed to get checkin by ID", zap.Error(err), zap.String("checkin_id", id.String()))
		return nil, fmt.Errorf("failed to get checkin: %w", err)
	}

	return row.toEntity()
}

// Update atualiza um checkin existente
func (repo *CheckinRepository) Update(ctx context.Context, c *checkin.Checkin) error {
	row := repo.fromEntity(c)

	query := `
		UPDATE checkin SET
			photo_url = :photo_url,
			notes = :notes,
			is_valid = :is_valid,
			validation_details = :validation_details,
			updated_at = :updated_at,
			updated_by = :updated_by
		WHERE id_checkin = :id_checkin`

	result, err := repo.db.NamedExecContext(ctx, query, row)
	if err != nil {
		repo.logger.Error("Failed to update checkin", zap.Error(err), zap.String("checkin_id", c.ID.String()))
		return fmt.Errorf("failed to update checkin: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("checkin not found")
	}

	repo.logger.Info("Checkin updated successfully", zap.String("checkin_id", c.ID.String()))
	return nil
}

// Delete remove um checkin (soft delete)
func (repo *CheckinRepository) Delete(ctx context.Context, id value_objects.UUID, deletedBy value_objects.UUID) error {
	query := `
		UPDATE checkin SET
			is_valid = false,
			updated_at = NOW(),
			updated_by = $2
		WHERE id_checkin = $1`

	result, err := repo.db.ExecContext(ctx, query, id.String(), deletedBy.String())
	if err != nil {
		repo.logger.Error("Failed to delete checkin", zap.Error(err), zap.String("checkin_id", id.String()))
		return fmt.Errorf("failed to delete checkin: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("checkin not found")
	}

	repo.logger.Info("Checkin deleted successfully", zap.String("checkin_id", id.String()))
	return nil
}

// List lista checkins com filtros e paginação
func (repo *CheckinRepository) List(ctx context.Context, filters checkin.ListFilters) ([]*checkin.Checkin, int, error) {
	// Construir query base
	baseQuery := `
		FROM checkin c
		WHERE 1=1`

	var args []interface{}
	var conditions []string
	argCount := 0

	// Aplicar filtros
	if filters.HasTenantFilter() {
		argCount++
		conditions = append(conditions, fmt.Sprintf("c.id_tenant = $%d", argCount))
		args = append(args, filters.TenantID.String())
	}

	if filters.HasEventFilter() {
		argCount++
		conditions = append(conditions, fmt.Sprintf("c.id_event = $%d", argCount))
		args = append(args, filters.EventID.String())
	}

	if filters.HasEmployeeFilter() {
		argCount++
		conditions = append(conditions, fmt.Sprintf("c.id_employee = $%d", argCount))
		args = append(args, filters.EmployeeID.String())
	}

	if filters.HasPartnerFilter() {
		argCount++
		conditions = append(conditions, fmt.Sprintf("c.id_partner = $%d", argCount))
		args = append(args, filters.PartnerID.String())
	}

	if filters.HasMethodFilter() {
		argCount++
		conditions = append(conditions, fmt.Sprintf("c.method = $%d", argCount))
		args = append(args, filters.GetMethodFilter())
	}

	if filters.HasValidFilter() {
		argCount++
		conditions = append(conditions, fmt.Sprintf("c.is_valid = $%d", argCount))
		args = append(args, *filters.IsValid)
	}

	if filters.HasPhotoFilter() {
		if *filters.HasPhoto {
			conditions = append(conditions, "c.photo_url IS NOT NULL AND c.photo_url != ''")
		} else {
			conditions = append(conditions, "(c.photo_url IS NULL OR c.photo_url = '')")
		}
	}

	if filters.HasDateRangeFilter() {
		if filters.StartDate != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("c.checkin_time >= $%d", argCount))
			args = append(args, *filters.StartDate)
		}
		if filters.EndDate != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("c.checkin_time <= $%d", argCount))
			args = append(args, *filters.EndDate)
		}
	}

	if filters.HasSearchFilter() {
		searchTerm := "%" + strings.ToLower(filters.GetSearchTerm()) + "%"
		argCount++
		conditions = append(conditions, fmt.Sprintf("LOWER(c.notes) LIKE $%d", argCount))
		args = append(args, searchTerm)
	}

	// Adicionar condições à query
	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	// Query para contar total
	countQuery := "SELECT COUNT(*) " + baseQuery
	var total int
	err := repo.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		repo.logger.Error("Failed to count checkins", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count checkins: %w", err)
	}

	// Query para buscar dados com paginação
	selectQuery := `
		SELECT c.id_checkin, c.id_tenant, c.id_event, c.id_employee, c.id_partner,
			   c.method, c.latitude, c.longitude, c.checkin_time, c.photo_url, c.notes,
			   c.is_valid, c.validation_details, c.created_at, c.updated_at, c.created_by, c.updated_by ` + baseQuery

	// Adicionar ordenação
	orderDirection := "ASC"
	if filters.OrderDesc {
		orderDirection = "DESC"
	}
	selectQuery += fmt.Sprintf(" ORDER BY c.%s %s", filters.OrderBy, orderDirection)

	// Adicionar paginação
	selectQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", filters.GetLimit(), filters.GetOffset())

	var rows []checkinRow
	err = repo.db.SelectContext(ctx, &rows, selectQuery, args...)
	if err != nil {
		repo.logger.Error("Failed to list checkins", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list checkins: %w", err)
	}

	// Converter para entidades
	checkins := make([]*checkin.Checkin, len(rows))
	for i, row := range rows {
		checkin, err := row.toEntity()
		if err != nil {
			repo.logger.Error("Failed to convert checkin row to entity", zap.Error(err))
			return nil, 0, fmt.Errorf("failed to convert checkin: %w", err)
		}
		checkins[i] = checkin
	}

	return checkins, total, nil
}

// ListByTenant lista checkins de um tenant específico
func (repo *CheckinRepository) ListByTenant(ctx context.Context, tenantID value_objects.UUID, filters checkin.ListFilters) ([]*checkin.Checkin, int, error) {
	filters.TenantID = &tenantID
	return repo.List(ctx, filters)
}

// GetByEmployee busca checkins de um funcionário
func (repo *CheckinRepository) GetByEmployee(ctx context.Context, employeeID value_objects.UUID, filters checkin.ListFilters) ([]*checkin.Checkin, int, error) {
	filters.EmployeeID = &employeeID
	return repo.List(ctx, filters)
}

// GetByEvent busca checkins de um evento
func (repo *CheckinRepository) GetByEvent(ctx context.Context, eventID value_objects.UUID, filters checkin.ListFilters) ([]*checkin.Checkin, int, error) {
	filters.EventID = &eventID
	return repo.List(ctx, filters)
}

// GetByPartner busca checkins de um parceiro
func (repo *CheckinRepository) GetByPartner(ctx context.Context, partnerID value_objects.UUID, filters checkin.ListFilters) ([]*checkin.Checkin, int, error) {
	filters.PartnerID = &partnerID
	return repo.List(ctx, filters)
}

// GetByEmployeeAndEvent busca checkin específico de funcionário em evento
func (repo *CheckinRepository) GetByEmployeeAndEvent(ctx context.Context, employeeID, eventID value_objects.UUID) (*checkin.Checkin, error) {
	var row checkinRow
	query := `
		SELECT id_checkin, id_tenant, id_event, id_employee, id_partner,
			   method, latitude, longitude, checkin_time, photo_url, notes,
			   is_valid, validation_details, created_at, updated_at, created_by, updated_by
		FROM checkin 
		WHERE id_employee = $1 AND id_event = $2
		ORDER BY checkin_time DESC
		LIMIT 1`

	err := repo.db.GetContext(ctx, &row, query, employeeID.String(), eventID.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("checkin not found")
		}
		repo.logger.Error("Failed to get checkin by employee and event", zap.Error(err))
		return nil, fmt.Errorf("failed to get checkin: %w", err)
	}

	return row.toEntity()
}

// ExistsByEmployeeAndEvent verifica se já existe checkin do funcionário no evento
func (repo *CheckinRepository) ExistsByEmployeeAndEvent(ctx context.Context, employeeID, eventID value_objects.UUID) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*)
		FROM checkin 
		WHERE id_employee = $1 AND id_event = $2`

	err := repo.db.GetContext(ctx, &count, query, employeeID.String(), eventID.String())
	if err != nil {
		repo.logger.Error("Failed to check checkin existence", zap.Error(err))
		return false, fmt.Errorf("failed to check checkin existence: %w", err)
	}

	return count > 0, nil
}

// GetByDateRange busca checkins em um período
func (repo *CheckinRepository) GetByDateRange(ctx context.Context, tenantID value_objects.UUID, startDate, endDate time.Time, filters checkin.ListFilters) ([]*checkin.Checkin, int, error) {
	filters.TenantID = &tenantID
	filters.StartDate = &startDate
	filters.EndDate = &endDate
	return repo.List(ctx, filters)
}

// GetByMethod busca checkins por método
func (repo *CheckinRepository) GetByMethod(ctx context.Context, tenantID value_objects.UUID, method string, filters checkin.ListFilters) ([]*checkin.Checkin, int, error) {
	filters.TenantID = &tenantID
	filters.Method = &method
	return repo.List(ctx, filters)
}

// GetValidCheckins busca apenas checkins válidos
func (repo *CheckinRepository) GetValidCheckins(ctx context.Context, tenantID value_objects.UUID, filters checkin.ListFilters) ([]*checkin.Checkin, int, error) {
	filters.TenantID = &tenantID
	valid := true
	filters.IsValid = &valid
	return repo.List(ctx, filters)
}

// GetInvalidCheckins busca apenas checkins inválidos
func (repo *CheckinRepository) GetInvalidCheckins(ctx context.Context, tenantID value_objects.UUID, filters checkin.ListFilters) ([]*checkin.Checkin, int, error) {
	filters.TenantID = &tenantID
	invalid := false
	filters.IsValid = &invalid
	return repo.List(ctx, filters)
}

// CountByTenant conta checkins de um tenant
func (repo *CheckinRepository) CountByTenant(ctx context.Context, tenantID value_objects.UUID) (int, error) {
	var count int
	query := `
		SELECT COUNT(*)
		FROM checkin 
		WHERE id_tenant = $1`

	err := repo.db.GetContext(ctx, &count, query, tenantID.String())
	if err != nil {
		repo.logger.Error("Failed to count checkins by tenant", zap.Error(err))
		return 0, fmt.Errorf("failed to count checkins: %w", err)
	}

	return count, nil
}

// CountByEvent conta checkins de um evento
func (repo *CheckinRepository) CountByEvent(ctx context.Context, eventID value_objects.UUID) (int, error) {
	var count int
	query := `
		SELECT COUNT(*)
		FROM checkin 
		WHERE id_event = $1`

	err := repo.db.GetContext(ctx, &count, query, eventID.String())
	if err != nil {
		repo.logger.Error("Failed to count checkins by event", zap.Error(err))
		return 0, fmt.Errorf("failed to count checkins: %w", err)
	}

	return count, nil
}

// CountByEmployee conta checkins de um funcionário
func (repo *CheckinRepository) CountByEmployee(ctx context.Context, employeeID value_objects.UUID) (int, error) {
	var count int
	query := `
		SELECT COUNT(*)
		FROM checkin 
		WHERE id_employee = $1`

	err := repo.db.GetContext(ctx, &count, query, employeeID.String())
	if err != nil {
		repo.logger.Error("Failed to count checkins by employee", zap.Error(err))
		return 0, fmt.Errorf("failed to count checkins: %w", err)
	}

	return count, nil
}

// GetRecentCheckins busca checkins recentes (últimas 24h)
func (repo *CheckinRepository) GetRecentCheckins(ctx context.Context, tenantID value_objects.UUID, limit int) ([]*checkin.Checkin, error) {
	query := `
		SELECT id_checkin, id_tenant, id_event, id_employee, id_partner,
			   method, latitude, longitude, checkin_time, photo_url, notes,
			   is_valid, validation_details, created_at, updated_at, created_by, updated_by
		FROM checkin 
		WHERE id_tenant = $1 AND checkin_time >= NOW() - INTERVAL '24 hours'
		ORDER BY checkin_time DESC
		LIMIT $2`

	var rows []checkinRow
	err := repo.db.SelectContext(ctx, &rows, query, tenantID.String(), limit)
	if err != nil {
		repo.logger.Error("Failed to get recent checkins", zap.Error(err))
		return nil, fmt.Errorf("failed to get recent checkins: %w", err)
	}

	// Converter para entidades
	checkins := make([]*checkin.Checkin, len(rows))
	for i, row := range rows {
		checkin, err := row.toEntity()
		if err != nil {
			repo.logger.Error("Failed to convert checkin row to entity", zap.Error(err))
			return nil, fmt.Errorf("failed to convert checkin: %w", err)
		}
		checkins[i] = checkin
	}

	return checkins, nil
}

// GetCheckinsByLocation busca checkins próximos a uma localização
func (repo *CheckinRepository) GetCheckinsByLocation(ctx context.Context, tenantID value_objects.UUID, location value_objects.Location, radiusKm float64, filters checkin.ListFilters) ([]*checkin.Checkin, int, error) {
	filters.TenantID = &tenantID
	filters.Location = &location
	filters.RadiusKm = &radiusKm

	// Construir query base com cálculo de distância usando PostGIS
	baseQuery := `
		FROM checkin c
		WHERE c.id_tenant = $1
		AND ST_DWithin(
			ST_GeogFromText('POINT(' || c.longitude || ' ' || c.latitude || ')'),
			ST_GeogFromText('POINT($3 $2)'),
			$4 * 1000
		)`

	args := []interface{}{tenantID.String(), location.Latitude, location.Longitude, radiusKm}
	argCount := 4

	// Aplicar filtros adicionais
	var conditions []string

	if filters.HasEventFilter() {
		argCount++
		conditions = append(conditions, fmt.Sprintf("c.id_event = $%d", argCount))
		args = append(args, filters.EventID.String())
	}

	if filters.HasEmployeeFilter() {
		argCount++
		conditions = append(conditions, fmt.Sprintf("c.id_employee = $%d", argCount))
		args = append(args, filters.EmployeeID.String())
	}

	if filters.HasValidFilter() {
		argCount++
		conditions = append(conditions, fmt.Sprintf("c.is_valid = $%d", argCount))
		args = append(args, *filters.IsValid)
	}

	// Adicionar condições adicionais
	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	// Query para contar total
	countQuery := "SELECT COUNT(*) " + baseQuery
	var total int
	err := repo.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		repo.logger.Error("Failed to count checkins by location", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count checkins: %w", err)
	}

	// Query para buscar dados com paginação
	selectQuery := `
		SELECT c.id_checkin, c.id_tenant, c.id_event, c.id_employee, c.id_partner,
			   c.method, c.latitude, c.longitude, c.checkin_time, c.photo_url, c.notes,
			   c.is_valid, c.validation_details, c.created_at, c.updated_at, c.created_by, c.updated_by,
			   ST_Distance(
				   ST_GeogFromText('POINT(' || c.longitude || ' ' || c.latitude || ')'),
				   ST_GeogFromText('POINT($3 $2)')
			   ) / 1000 as distance_km ` + baseQuery

	// Adicionar ordenação por distância
	selectQuery += " ORDER BY distance_km ASC"

	// Adicionar paginação
	selectQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", filters.GetLimit(), filters.GetOffset())

	var rows []checkinRow
	err = repo.db.SelectContext(ctx, &rows, selectQuery, args...)
	if err != nil {
		repo.logger.Error("Failed to list checkins by location", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list checkins: %w", err)
	}

	// Converter para entidades
	checkins := make([]*checkin.Checkin, len(rows))
	for i, row := range rows {
		checkin, err := row.toEntity()
		if err != nil {
			repo.logger.Error("Failed to convert checkin row to entity", zap.Error(err))
			return nil, 0, fmt.Errorf("failed to convert checkin: %w", err)
		}
		checkins[i] = checkin
	}

	return checkins, total, nil
}
