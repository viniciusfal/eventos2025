package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"eventos-backend/internal/domain/checkout"
	"eventos-backend/internal/domain/shared/value_objects"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// CheckoutRepository implementa a interface de repositório para Checkout
type CheckoutRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// NewCheckoutRepository cria uma nova instância do repositório de checkout
func NewCheckoutRepository(db *sqlx.DB, logger *zap.Logger) checkout.Repository {
	return &CheckoutRepository{
		db:     db,
		logger: logger,
	}
}

// checkoutRow representa uma linha da tabela checkout no banco
type checkoutRow struct {
	ID                string         `db:"id_checkout"`
	TenantID          string         `db:"id_tenant"`
	EventID           string         `db:"id_event"`
	EmployeeID        string         `db:"id_employee"`
	PartnerID         string         `db:"id_partner"`
	CheckinID         string         `db:"id_checkin"`
	Method            string         `db:"method"`
	Latitude          float64        `db:"latitude"`
	Longitude         float64        `db:"longitude"`
	CheckoutTime      time.Time      `db:"checkout_time"`
	PhotoURL          sql.NullString `db:"photo_url"`
	Notes             sql.NullString `db:"notes"`
	WorkDurationSecs  int64          `db:"work_duration_seconds"`
	IsValid           bool           `db:"is_valid"`
	ValidationDetails sql.NullString `db:"validation_details"`
	CreatedAt         time.Time      `db:"created_at"`
	UpdatedAt         time.Time      `db:"updated_at"`
	CreatedBy         sql.NullString `db:"created_by"`
	UpdatedBy         sql.NullString `db:"updated_by"`
}

// workSessionRow representa uma sessão de trabalho completa (checkin + checkout)
type workSessionRow struct {
	CheckinID        string    `db:"checkin_id"`
	CheckoutID       string    `db:"checkout_id"`
	TenantID         string    `db:"id_tenant"`
	EventID          string    `db:"id_event"`
	EmployeeID       string    `db:"id_employee"`
	PartnerID        string    `db:"id_partner"`
	CheckinTime      time.Time `db:"checkin_time"`
	CheckoutTime     time.Time `db:"checkout_time"`
	WorkDurationSecs int64     `db:"work_duration_seconds"`
	IsValid          bool      `db:"is_valid"`
	IsComplete       bool      `db:"is_complete"`
}

// toEntity converte uma linha do banco para entidade de domínio
func (r *checkoutRow) toEntity() (*checkout.Checkout, error) {
	id, err := value_objects.ParseUUID(r.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid checkout ID: %w", err)
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

	checkinID, err := value_objects.ParseUUID(r.CheckinID)
	if err != nil {
		return nil, fmt.Errorf("invalid checkin ID: %w", err)
	}

	location, err := value_objects.NewLocation(r.Latitude, r.Longitude)
	if err != nil {
		return nil, fmt.Errorf("invalid location: %w", err)
	}

	checkoutEntity := &checkout.Checkout{
		ID:           id,
		TenantID:     tenantID,
		EventID:      eventID,
		EmployeeID:   employeeID,
		PartnerID:    partnerID,
		CheckinID:    checkinID,
		Method:       r.Method,
		Location:     location,
		CheckoutTime: r.CheckoutTime,
		WorkDuration: time.Duration(r.WorkDurationSecs) * time.Second,
		IsValid:      r.IsValid,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}

	// PhotoURL
	if r.PhotoURL.Valid {
		checkoutEntity.PhotoURL = r.PhotoURL.String
	}

	// Notes
	if r.Notes.Valid {
		checkoutEntity.Notes = r.Notes.String
	}

	// ValidationDetails
	if r.ValidationDetails.Valid && r.ValidationDetails.String != "" {
		var details map[string]interface{}
		if err := json.Unmarshal([]byte(r.ValidationDetails.String), &details); err == nil {
			checkoutEntity.ValidationDetails = details
		} else {
			checkoutEntity.ValidationDetails = make(map[string]interface{})
		}
	} else {
		checkoutEntity.ValidationDetails = make(map[string]interface{})
	}

	// CreatedBy
	if r.CreatedBy.Valid {
		createdBy, err := value_objects.ParseUUID(r.CreatedBy.String)
		if err == nil {
			checkoutEntity.CreatedBy = &createdBy
		}
	}

	// UpdatedBy
	if r.UpdatedBy.Valid {
		updatedBy, err := value_objects.ParseUUID(r.UpdatedBy.String)
		if err == nil {
			checkoutEntity.UpdatedBy = &updatedBy
		}
	}

	return checkoutEntity, nil
}

// toWorkSessionEntity converte uma linha de sessão de trabalho para entidade
func (r *workSessionRow) toWorkSessionEntity() (*checkout.WorkSession, error) {
	checkinID, err := value_objects.ParseUUID(r.CheckinID)
	if err != nil {
		return nil, fmt.Errorf("invalid checkin ID: %w", err)
	}

	checkoutID, err := value_objects.ParseUUID(r.CheckoutID)
	if err != nil {
		return nil, fmt.Errorf("invalid checkout ID: %w", err)
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

	return &checkout.WorkSession{
		CheckinID:    checkinID,
		CheckoutID:   checkoutID,
		EventID:      eventID,
		EmployeeID:   employeeID,
		PartnerID:    partnerID,
		CheckinTime:  r.CheckinTime,
		CheckoutTime: r.CheckoutTime,
		Duration:     time.Duration(r.WorkDurationSecs) * time.Second,
		IsValid:      r.IsValid,
		IsComplete:   r.IsComplete,
	}, nil
}

// fromEntity converte uma entidade de domínio para linha do banco
func (repo *CheckoutRepository) fromEntity(c *checkout.Checkout) *checkoutRow {
	row := &checkoutRow{
		ID:               c.ID.String(),
		TenantID:         c.TenantID.String(),
		EventID:          c.EventID.String(),
		EmployeeID:       c.EmployeeID.String(),
		PartnerID:        c.PartnerID.String(),
		CheckinID:        c.CheckinID.String(),
		Method:           c.Method,
		Latitude:         c.Location.Latitude,
		Longitude:        c.Location.Longitude,
		CheckoutTime:     c.CheckoutTime,
		WorkDurationSecs: int64(c.WorkDuration.Seconds()),
		IsValid:          c.IsValid,
		CreatedAt:        c.CreatedAt,
		UpdatedAt:        c.UpdatedAt,
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

// Create cria um novo checkout
func (repo *CheckoutRepository) Create(ctx context.Context, c *checkout.Checkout) error {
	row := repo.fromEntity(c)

	query := `
		INSERT INTO checkout (
			id_checkout, id_tenant, id_event, id_employee, id_partner, id_checkin,
			method, latitude, longitude, checkout_time, photo_url, notes,
			work_duration_seconds, is_valid, validation_details,
			created_at, updated_at, created_by, updated_by
		) VALUES (
			:id_checkout, :id_tenant, :id_event, :id_employee, :id_partner, :id_checkin,
			:method, :latitude, :longitude, :checkout_time, :photo_url, :notes,
			:work_duration_seconds, :is_valid, :validation_details,
			:created_at, :updated_at, :created_by, :updated_by
		)`

	_, err := repo.db.NamedExecContext(ctx, query, row)
	if err != nil {
		repo.logger.Error("Failed to create checkout", zap.Error(err), zap.String("checkout_id", c.ID.String()))
		return fmt.Errorf("failed to create checkout: %w", err)
	}

	repo.logger.Info("Checkout created successfully", zap.String("checkout_id", c.ID.String()))
	return nil
}

// GetByID busca um checkout por ID
func (repo *CheckoutRepository) GetByID(ctx context.Context, id value_objects.UUID) (*checkout.Checkout, error) {
	var row checkoutRow
	query := `
		SELECT id_checkout, id_tenant, id_event, id_employee, id_partner, id_checkin,
			   method, latitude, longitude, checkout_time, photo_url, notes,
			   work_duration_seconds, is_valid, validation_details,
			   created_at, updated_at, created_by, updated_by
		FROM checkout 
		WHERE id_checkout = $1`

	err := repo.db.GetContext(ctx, &row, query, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("checkout not found")
		}
		repo.logger.Error("Failed to get checkout by ID", zap.Error(err), zap.String("checkout_id", id.String()))
		return nil, fmt.Errorf("failed to get checkout: %w", err)
	}

	return row.toEntity()
}

// Update atualiza um checkout existente
func (repo *CheckoutRepository) Update(ctx context.Context, c *checkout.Checkout) error {
	row := repo.fromEntity(c)

	query := `
		UPDATE checkout SET
			photo_url = :photo_url,
			notes = :notes,
			work_duration_seconds = :work_duration_seconds,
			is_valid = :is_valid,
			validation_details = :validation_details,
			updated_at = :updated_at,
			updated_by = :updated_by
		WHERE id_checkout = :id_checkout`

	result, err := repo.db.NamedExecContext(ctx, query, row)
	if err != nil {
		repo.logger.Error("Failed to update checkout", zap.Error(err), zap.String("checkout_id", c.ID.String()))
		return fmt.Errorf("failed to update checkout: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("checkout not found")
	}

	repo.logger.Info("Checkout updated successfully", zap.String("checkout_id", c.ID.String()))
	return nil
}

// Delete remove um checkout (soft delete)
func (repo *CheckoutRepository) Delete(ctx context.Context, id value_objects.UUID, deletedBy value_objects.UUID) error {
	query := `
		UPDATE checkout SET
			is_valid = false,
			updated_at = NOW(),
			updated_by = $2
		WHERE id_checkout = $1`

	result, err := repo.db.ExecContext(ctx, query, id.String(), deletedBy.String())
	if err != nil {
		repo.logger.Error("Failed to delete checkout", zap.Error(err), zap.String("checkout_id", id.String()))
		return fmt.Errorf("failed to delete checkout: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("checkout not found")
	}

	repo.logger.Info("Checkout deleted successfully", zap.String("checkout_id", id.String()))
	return nil
}

// List lista checkouts com filtros e paginação
func (repo *CheckoutRepository) List(ctx context.Context, filters checkout.ListFilters) ([]*checkout.Checkout, int, error) {
	// Construir query base
	baseQuery := `
		FROM checkout c
		WHERE 1=1`

	var args []interface{}
	var conditions []string
	argCount := 0

	// Aplicar filtros
	if filters.TenantID != nil && !filters.TenantID.IsZero() {
		argCount++
		conditions = append(conditions, fmt.Sprintf("c.id_tenant = $%d", argCount))
		args = append(args, filters.TenantID.String())
	}

	if filters.EventID != nil && !filters.EventID.IsZero() {
		argCount++
		conditions = append(conditions, fmt.Sprintf("c.id_event = $%d", argCount))
		args = append(args, filters.EventID.String())
	}

	if filters.EmployeeID != nil && !filters.EmployeeID.IsZero() {
		argCount++
		conditions = append(conditions, fmt.Sprintf("c.id_employee = $%d", argCount))
		args = append(args, filters.EmployeeID.String())
	}

	if filters.PartnerID != nil && !filters.PartnerID.IsZero() {
		argCount++
		conditions = append(conditions, fmt.Sprintf("c.id_partner = $%d", argCount))
		args = append(args, filters.PartnerID.String())
	}

	if filters.CheckinID != nil && !filters.CheckinID.IsZero() {
		argCount++
		conditions = append(conditions, fmt.Sprintf("c.id_checkin = $%d", argCount))
		args = append(args, filters.CheckinID.String())
	}

	if filters.Method != nil && *filters.Method != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("c.method = $%d", argCount))
		args = append(args, *filters.Method)
	}

	if filters.IsValid != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("c.is_valid = $%d", argCount))
		args = append(args, *filters.IsValid)
	}

	if filters.HasPhoto != nil {
		if *filters.HasPhoto {
			conditions = append(conditions, "c.photo_url IS NOT NULL AND c.photo_url != ''")
		} else {
			conditions = append(conditions, "(c.photo_url IS NULL OR c.photo_url = '')")
		}
	}

	if filters.StartDate != nil || filters.EndDate != nil {
		if filters.StartDate != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("c.checkout_time >= $%d", argCount))
			args = append(args, *filters.StartDate)
		}
		if filters.EndDate != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("c.checkout_time <= $%d", argCount))
			args = append(args, *filters.EndDate)
		}
	}

	if filters.MinDurationHours != nil || filters.MaxDurationHours != nil {
		if filters.MinDurationHours != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("c.work_duration_seconds >= $%d", argCount))
			args = append(args, int64(*filters.MinDurationHours*3600))
		}
		if filters.MaxDurationHours != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("c.work_duration_seconds <= $%d", argCount))
			args = append(args, int64(*filters.MaxDurationHours*3600))
		}
	}

	if filters.Search != nil && *filters.Search != "" {
		searchTerm := "%" + strings.ToLower(strings.TrimSpace(*filters.Search)) + "%"
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
		repo.logger.Error("Failed to count checkouts", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count checkouts: %w", err)
	}

	// Query para buscar dados com paginação
	selectQuery := `
		SELECT c.id_checkout, c.id_tenant, c.id_event, c.id_employee, c.id_partner, c.id_checkin,
			   c.method, c.latitude, c.longitude, c.checkout_time, c.photo_url, c.notes,
			   c.work_duration_seconds, c.is_valid, c.validation_details,
			   c.created_at, c.updated_at, c.created_by, c.updated_by ` + baseQuery

	// Adicionar ordenação
	orderDirection := "ASC"
	if filters.OrderDesc {
		orderDirection = "DESC"
	}
	selectQuery += fmt.Sprintf(" ORDER BY c.%s %s", filters.OrderBy, orderDirection)

	// Adicionar paginação
	selectQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", filters.GetLimit(), filters.GetOffset())

	var rows []checkoutRow
	err = repo.db.SelectContext(ctx, &rows, selectQuery, args...)
	if err != nil {
		repo.logger.Error("Failed to list checkouts", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list checkouts: %w", err)
	}

	// Converter para entidades
	checkouts := make([]*checkout.Checkout, len(rows))
	for i, row := range rows {
		checkout, err := row.toEntity()
		if err != nil {
			repo.logger.Error("Failed to convert checkout row to entity", zap.Error(err))
			return nil, 0, fmt.Errorf("failed to convert checkout: %w", err)
		}
		checkouts[i] = checkout
	}

	return checkouts, total, nil
}

// ListByTenant lista checkouts de um tenant específico
func (repo *CheckoutRepository) ListByTenant(ctx context.Context, tenantID value_objects.UUID, filters checkout.ListFilters) ([]*checkout.Checkout, int, error) {
	filters.TenantID = &tenantID
	return repo.List(ctx, filters)
}

// GetByEmployee busca checkouts de um funcionário
func (repo *CheckoutRepository) GetByEmployee(ctx context.Context, employeeID value_objects.UUID, filters checkout.ListFilters) ([]*checkout.Checkout, int, error) {
	filters.EmployeeID = &employeeID
	return repo.List(ctx, filters)
}

// GetByEvent busca checkouts de um evento
func (repo *CheckoutRepository) GetByEvent(ctx context.Context, eventID value_objects.UUID, filters checkout.ListFilters) ([]*checkout.Checkout, int, error) {
	filters.EventID = &eventID
	return repo.List(ctx, filters)
}

// GetByPartner busca checkouts de um parceiro
func (repo *CheckoutRepository) GetByPartner(ctx context.Context, partnerID value_objects.UUID, filters checkout.ListFilters) ([]*checkout.Checkout, int, error) {
	filters.PartnerID = &partnerID
	return repo.List(ctx, filters)
}

// GetByCheckin busca checkout por checkin
func (repo *CheckoutRepository) GetByCheckin(ctx context.Context, checkinID value_objects.UUID) (*checkout.Checkout, error) {
	var row checkoutRow
	query := `
		SELECT id_checkout, id_tenant, id_event, id_employee, id_partner, id_checkin,
			   method, latitude, longitude, checkout_time, photo_url, notes,
			   work_duration_seconds, is_valid, validation_details,
			   created_at, updated_at, created_by, updated_by
		FROM checkout 
		WHERE id_checkin = $1`

	err := repo.db.GetContext(ctx, &row, query, checkinID.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("checkout not found")
		}
		repo.logger.Error("Failed to get checkout by checkin", zap.Error(err))
		return nil, fmt.Errorf("failed to get checkout: %w", err)
	}

	return row.toEntity()
}

// ExistsByCheckin verifica se já existe checkout para o checkin
func (repo *CheckoutRepository) ExistsByCheckin(ctx context.Context, checkinID value_objects.UUID) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*)
		FROM checkout 
		WHERE id_checkin = $1`

	err := repo.db.GetContext(ctx, &count, query, checkinID.String())
	if err != nil {
		repo.logger.Error("Failed to check checkout existence", zap.Error(err))
		return false, fmt.Errorf("failed to check checkout existence: %w", err)
	}

	return count > 0, nil
}

// GetByEmployeeAndEvent busca checkout específico de funcionário em evento
func (repo *CheckoutRepository) GetByEmployeeAndEvent(ctx context.Context, employeeID, eventID value_objects.UUID) (*checkout.Checkout, error) {
	var row checkoutRow
	query := `
		SELECT id_checkout, id_tenant, id_event, id_employee, id_partner, id_checkin,
			   method, latitude, longitude, checkout_time, photo_url, notes,
			   work_duration_seconds, is_valid, validation_details,
			   created_at, updated_at, created_by, updated_by
		FROM checkout 
		WHERE id_employee = $1 AND id_event = $2
		ORDER BY checkout_time DESC
		LIMIT 1`

	err := repo.db.GetContext(ctx, &row, query, employeeID.String(), eventID.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("checkout not found")
		}
		repo.logger.Error("Failed to get checkout by employee and event", zap.Error(err))
		return nil, fmt.Errorf("failed to get checkout: %w", err)
	}

	return row.toEntity()
}

// GetWorkSessions busca sessões de trabalho completas (checkin + checkout)
func (repo *CheckoutRepository) GetWorkSessions(ctx context.Context, tenantID value_objects.UUID, filters checkout.WorkSessionFilters) ([]*checkout.WorkSession, int, error) {
	// Construir query base
	baseQuery := `
		FROM checkin ci
		LEFT JOIN checkout co ON ci.id_checkin = co.id_checkin
		WHERE ci.id_tenant = $1`

	args := []interface{}{tenantID.String()}
	argCount := 1

	var conditions []string

	// Aplicar filtros
	if filters.EventID != nil && !filters.EventID.IsZero() {
		argCount++
		conditions = append(conditions, fmt.Sprintf("ci.id_event = $%d", argCount))
		args = append(args, filters.EventID.String())
	}

	if filters.EmployeeID != nil && !filters.EmployeeID.IsZero() {
		argCount++
		conditions = append(conditions, fmt.Sprintf("ci.id_employee = $%d", argCount))
		args = append(args, filters.EmployeeID.String())
	}

	if filters.PartnerID != nil && !filters.PartnerID.IsZero() {
		argCount++
		conditions = append(conditions, fmt.Sprintf("ci.id_partner = $%d", argCount))
		args = append(args, filters.PartnerID.String())
	}

	if filters.StartDate != nil || filters.EndDate != nil {
		if filters.StartDate != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("ci.checkin_time >= $%d", argCount))
			args = append(args, *filters.StartDate)
		}
		if filters.EndDate != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("ci.checkin_time <= $%d", argCount))
			args = append(args, *filters.EndDate)
		}
	}

	if filters.MinDurationHours != nil || filters.MaxDurationHours != nil {
		if filters.MinDurationHours != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("co.work_duration_seconds >= $%d", argCount))
			args = append(args, int64(*filters.MinDurationHours*3600))
		}
		if filters.MaxDurationHours != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("co.work_duration_seconds <= $%d", argCount))
			args = append(args, int64(*filters.MaxDurationHours*3600))
		}
	}

	if filters.IsValid != nil {
		if *filters.IsValid {
			conditions = append(conditions, "ci.is_valid = true AND co.is_valid = true")
		} else {
			conditions = append(conditions, "(ci.is_valid = false OR co.is_valid = false)")
		}
	}

	if filters.IsComplete != nil {
		if *filters.IsComplete {
			conditions = append(conditions, "co.id_checkout IS NOT NULL")
		} else {
			conditions = append(conditions, "co.id_checkout IS NULL")
		}
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
		repo.logger.Error("Failed to count work sessions", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count work sessions: %w", err)
	}

	// Query para buscar dados com paginação
	selectQuery := `
		SELECT ci.id_checkin as checkin_id,
			   co.id_checkout as checkout_id,
			   ci.id_tenant,
			   ci.id_event,
			   ci.id_employee,
			   ci.id_partner,
			   ci.checkin_time,
			   co.checkout_time,
			   co.work_duration_seconds,
			   (ci.is_valid AND COALESCE(co.is_valid, true)) as is_valid,
			   (co.id_checkout IS NOT NULL) as is_complete ` + baseQuery

	// Adicionar ordenação
	orderDirection := "ASC"
	if filters.OrderDesc {
		orderDirection = "DESC"
	}
	orderField := "ci.checkin_time"
	if filters.OrderBy == "duration" {
		orderField = "co.work_duration_seconds"
	} else if filters.OrderBy == "checkout_time" {
		orderField = "co.checkout_time"
	}
	selectQuery += fmt.Sprintf(" ORDER BY %s %s", orderField, orderDirection)

	// Adicionar paginação
	selectQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", filters.GetLimit(), filters.GetOffset())

	var rows []workSessionRow
	err = repo.db.SelectContext(ctx, &rows, selectQuery, args...)
	if err != nil {
		repo.logger.Error("Failed to list work sessions", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list work sessions: %w", err)
	}

	// Converter para entidades
	sessions := make([]*checkout.WorkSession, len(rows))
	for i, row := range rows {
		session, err := row.toWorkSessionEntity()
		if err != nil {
			repo.logger.Error("Failed to convert work session row to entity", zap.Error(err))
			return nil, 0, fmt.Errorf("failed to convert work session: %w", err)
		}
		sessions[i] = session
	}

	return sessions, total, nil
}

// GetByDateRange busca checkouts em um período
func (repo *CheckoutRepository) GetByDateRange(ctx context.Context, tenantID value_objects.UUID, startDate, endDate time.Time, filters checkout.ListFilters) ([]*checkout.Checkout, int, error) {
	filters.TenantID = &tenantID
	filters.StartDate = &startDate
	filters.EndDate = &endDate
	return repo.List(ctx, filters)
}

// GetByMethod busca checkouts por método
func (repo *CheckoutRepository) GetByMethod(ctx context.Context, tenantID value_objects.UUID, method string, filters checkout.ListFilters) ([]*checkout.Checkout, int, error) {
	filters.TenantID = &tenantID
	filters.Method = &method
	return repo.List(ctx, filters)
}

// GetValidCheckouts busca apenas checkouts válidos
func (repo *CheckoutRepository) GetValidCheckouts(ctx context.Context, tenantID value_objects.UUID, filters checkout.ListFilters) ([]*checkout.Checkout, int, error) {
	filters.TenantID = &tenantID
	valid := true
	filters.IsValid = &valid
	return repo.List(ctx, filters)
}

// GetInvalidCheckouts busca apenas checkouts inválidos
func (repo *CheckoutRepository) GetInvalidCheckouts(ctx context.Context, tenantID value_objects.UUID, filters checkout.ListFilters) ([]*checkout.Checkout, int, error) {
	filters.TenantID = &tenantID
	invalid := false
	filters.IsValid = &invalid
	return repo.List(ctx, filters)
}

// GetCheckoutsByLocation busca checkouts próximos a uma localização
func (repo *CheckoutRepository) GetCheckoutsByLocation(ctx context.Context, tenantID value_objects.UUID, location value_objects.Location, radiusKm float64, filters checkout.ListFilters) ([]*checkout.Checkout, int, error) {
	filters.TenantID = &tenantID
	filters.Location = &location
	filters.RadiusKm = &radiusKm

	// Construir query base com cálculo de distância usando PostGIS
	baseQuery := `
		FROM checkout c
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

	if filters.EventID != nil && !filters.EventID.IsZero() {
		argCount++
		conditions = append(conditions, fmt.Sprintf("c.id_event = $%d", argCount))
		args = append(args, filters.EventID.String())
	}

	if filters.EmployeeID != nil && !filters.EmployeeID.IsZero() {
		argCount++
		conditions = append(conditions, fmt.Sprintf("c.id_employee = $%d", argCount))
		args = append(args, filters.EmployeeID.String())
	}

	if filters.IsValid != nil {
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
		repo.logger.Error("Failed to count checkouts by location", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count checkouts: %w", err)
	}

	// Query para buscar dados com paginação
	selectQuery := `
		SELECT c.id_checkout, c.id_tenant, c.id_event, c.id_employee, c.id_partner, c.id_checkin,
			   c.method, c.latitude, c.longitude, c.checkout_time, c.photo_url, c.notes,
			   c.work_duration_seconds, c.is_valid, c.validation_details,
			   c.created_at, c.updated_at, c.created_by, c.updated_by,
			   ST_Distance(
				   ST_GeogFromText('POINT(' || c.longitude || ' ' || c.latitude || ')'),
				   ST_GeogFromText('POINT($3 $2)')
			   ) / 1000 as distance_km ` + baseQuery

	// Adicionar ordenação por distância
	selectQuery += " ORDER BY distance_km ASC"

	// Adicionar paginação
	selectQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", filters.GetLimit(), filters.GetOffset())

	var rows []checkoutRow
	err = repo.db.SelectContext(ctx, &rows, selectQuery, args...)
	if err != nil {
		repo.logger.Error("Failed to list checkouts by location", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list checkouts: %w", err)
	}

	// Converter para entidades
	checkouts := make([]*checkout.Checkout, len(rows))
	for i, row := range rows {
		checkout, err := row.toEntity()
		if err != nil {
			repo.logger.Error("Failed to convert checkout row to entity", zap.Error(err))
			return nil, 0, fmt.Errorf("failed to convert checkout: %w", err)
		}
		checkouts[i] = checkout
	}

	return checkouts, total, nil
}

// GetEmployeeWorkSessions busca sessões de trabalho de um funcionário
func (repo *CheckoutRepository) GetEmployeeWorkSessions(ctx context.Context, employeeID value_objects.UUID, filters checkout.WorkSessionFilters) ([]*checkout.WorkSession, int, error) {
	filters.EmployeeID = &employeeID
	// Usar um tenant vazio já que não temos como obter o tenantID aqui
	// Em uma implementação real, seria necessário buscar o tenantID do funcionário primeiro
	return nil, 0, fmt.Errorf("method not implemented - requires tenant information")
}

// GetEventWorkSessions busca sessões de trabalho de um evento
func (repo *CheckoutRepository) GetEventWorkSessions(ctx context.Context, eventID value_objects.UUID, filters checkout.WorkSessionFilters) ([]*checkout.WorkSession, int, error) {
	filters.EventID = &eventID
	// Usar um tenant vazio já que não temos como obter o tenantID aqui
	// Em uma implementação real, seria necessário buscar o tenantID do evento primeiro
	return nil, 0, fmt.Errorf("method not implemented - requires tenant information")
}

// GetRecentCheckouts busca checkouts recentes (últimas 24h)
func (repo *CheckoutRepository) GetRecentCheckouts(ctx context.Context, tenantID value_objects.UUID, limit int) ([]*checkout.Checkout, error) {
	query := `
		SELECT id_checkout, id_tenant, id_event, id_employee, id_partner, id_checkin,
			   method, latitude, longitude, checkout_time, photo_url, notes,
			   work_duration_seconds, is_valid, validation_details,
			   created_at, updated_at, created_by, updated_by
		FROM checkout 
		WHERE id_tenant = $1 AND checkout_time >= NOW() - INTERVAL '24 hours'
		ORDER BY checkout_time DESC
		LIMIT $2`

	var rows []checkoutRow
	err := repo.db.SelectContext(ctx, &rows, query, tenantID.String(), limit)
	if err != nil {
		repo.logger.Error("Failed to get recent checkouts", zap.Error(err))
		return nil, fmt.Errorf("failed to get recent checkouts: %w", err)
	}

	// Converter para entidades
	checkouts := make([]*checkout.Checkout, len(rows))
	for i, row := range rows {
		checkout, err := row.toEntity()
		if err != nil {
			repo.logger.Error("Failed to convert checkout row to entity", zap.Error(err))
			return nil, fmt.Errorf("failed to convert checkout: %w", err)
		}
		checkouts[i] = checkout
	}

	return checkouts, nil
}

// Implementar as demais funções seguindo o mesmo padrão...
// (GetByDateRange, GetByMethod, GetValidCheckouts, GetInvalidCheckouts, etc.)

// CountByTenant conta checkouts de um tenant
func (repo *CheckoutRepository) CountByTenant(ctx context.Context, tenantID value_objects.UUID) (int, error) {
	var count int
	query := `
		SELECT COUNT(*)
		FROM checkout 
		WHERE id_tenant = $1`

	err := repo.db.GetContext(ctx, &count, query, tenantID.String())
	if err != nil {
		repo.logger.Error("Failed to count checkouts by tenant", zap.Error(err))
		return 0, fmt.Errorf("failed to count checkouts: %w", err)
	}

	return count, nil
}

// CountByEvent conta checkouts de um evento
func (repo *CheckoutRepository) CountByEvent(ctx context.Context, eventID value_objects.UUID) (int, error) {
	var count int
	query := `
		SELECT COUNT(*)
		FROM checkout 
		WHERE id_event = $1`

	err := repo.db.GetContext(ctx, &count, query, eventID.String())
	if err != nil {
		repo.logger.Error("Failed to count checkouts by event", zap.Error(err))
		return 0, fmt.Errorf("failed to count checkouts: %w", err)
	}

	return count, nil
}

// CountByEmployee conta checkouts de um funcionário
func (repo *CheckoutRepository) CountByEmployee(ctx context.Context, employeeID value_objects.UUID) (int, error) {
	var count int
	query := `
		SELECT COUNT(*)
		FROM checkout 
		WHERE id_employee = $1`

	err := repo.db.GetContext(ctx, &count, query, employeeID.String())
	if err != nil {
		repo.logger.Error("Failed to count checkouts by employee", zap.Error(err))
		return 0, fmt.Errorf("failed to count checkouts: %w", err)
	}

	return count, nil
}

// Implementações restantes podem ser adicionadas seguindo o mesmo padrão...
