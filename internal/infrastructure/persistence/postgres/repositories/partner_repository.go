package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"eventos-backend/internal/domain/partner"
	"eventos-backend/internal/domain/shared/errors"
	"eventos-backend/internal/domain/shared/value_objects"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// PartnerRepository implementa a interface partner.Repository usando PostgreSQL
type PartnerRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// NewPartnerRepository cria uma nova instância do repositório de parceiros
func NewPartnerRepository(db *sqlx.DB, logger *zap.Logger) partner.Repository {
	return &PartnerRepository{
		db:     db,
		logger: logger,
	}
}

// partnerRow representa uma linha de parceiro no banco de dados
type partnerRow struct {
	ID                  string         `db:"id"`
	TenantID            string         `db:"tenant_id"`
	Name                string         `db:"name"`
	Email               string         `db:"email"`
	Email2              sql.NullString `db:"email2"`
	Phone               string         `db:"phone"`
	Phone2              sql.NullString `db:"phone2"`
	Identity            string         `db:"identity"`
	IdentityType        string         `db:"identity_type"`
	Location            string         `db:"location"`
	PasswordHash        string         `db:"password_hash"`
	LastLogin           sql.NullTime   `db:"last_login"`
	FailedLoginAttempts int            `db:"failed_login_attempts"`
	LockedUntil         sql.NullTime   `db:"locked_until"`
	Active              bool           `db:"active"`
	CreatedAt           time.Time      `db:"created_at"`
	UpdatedAt           time.Time      `db:"updated_at"`
	CreatedBy           sql.NullString `db:"created_by"`
	UpdatedBy           sql.NullString `db:"updated_by"`
}

// toEntity converte partnerRow para entidade Partner
func (r *partnerRow) toEntity() (*partner.Partner, error) {
	id, err := value_objects.ParseUUID(r.ID)
	if err != nil {
		return nil, errors.NewDomainError("INVALID_ID", "invalid partner ID", err)
	}

	tenantID, err := value_objects.ParseUUID(r.TenantID)
	if err != nil {
		return nil, errors.NewDomainError("INVALID_TENANT_ID", "invalid tenant ID", err)
	}

	p := &partner.Partner{
		ID:                  id,
		TenantID:            tenantID,
		Name:                r.Name,
		Email:               r.Email,
		Phone:               r.Phone,
		Identity:            r.Identity,
		IdentityType:        r.IdentityType,
		Location:            r.Location,
		PasswordHash:        r.PasswordHash,
		FailedLoginAttempts: r.FailedLoginAttempts,
		Active:              r.Active,
		CreatedAt:           r.CreatedAt,
		UpdatedAt:           r.UpdatedAt,
	}

	if r.Email2.Valid {
		p.Email2 = r.Email2.String
	}

	if r.Phone2.Valid {
		p.Phone2 = r.Phone2.String
	}

	if r.LastLogin.Valid {
		p.LastLogin = &r.LastLogin.Time
	}

	if r.LockedUntil.Valid {
		p.LockedUntil = &r.LockedUntil.Time
	}

	if r.CreatedBy.Valid {
		createdBy, err := value_objects.ParseUUID(r.CreatedBy.String)
		if err == nil {
			p.CreatedBy = &createdBy
		}
	}

	if r.UpdatedBy.Valid {
		updatedBy, err := value_objects.ParseUUID(r.UpdatedBy.String)
		if err == nil {
			p.UpdatedBy = &updatedBy
		}
	}

	return p, nil
}

// fromEntity converte entidade Partner para partnerRow
func (repo *PartnerRepository) fromEntity(p *partner.Partner) *partnerRow {
	row := &partnerRow{
		ID:                  p.ID.String(),
		TenantID:            p.TenantID.String(),
		Name:                p.Name,
		Email:               p.Email,
		Phone:               p.Phone,
		Identity:            p.Identity,
		IdentityType:        p.IdentityType,
		Location:            p.Location,
		PasswordHash:        p.PasswordHash,
		FailedLoginAttempts: p.FailedLoginAttempts,
		Active:              p.Active,
		CreatedAt:           p.CreatedAt,
		UpdatedAt:           p.UpdatedAt,
	}

	if p.Email2 != "" {
		row.Email2 = sql.NullString{String: p.Email2, Valid: true}
	}

	if p.Phone2 != "" {
		row.Phone2 = sql.NullString{String: p.Phone2, Valid: true}
	}

	if p.LastLogin != nil {
		row.LastLogin = sql.NullTime{Time: *p.LastLogin, Valid: true}
	}

	if p.LockedUntil != nil {
		row.LockedUntil = sql.NullTime{Time: *p.LockedUntil, Valid: true}
	}

	if p.CreatedBy != nil {
		row.CreatedBy = sql.NullString{String: p.CreatedBy.String(), Valid: true}
	}

	if p.UpdatedBy != nil {
		row.UpdatedBy = sql.NullString{String: p.UpdatedBy.String(), Valid: true}
	}

	return row
}

// Create cria um novo parceiro
func (repo *PartnerRepository) Create(ctx context.Context, p *partner.Partner) error {
	row := repo.fromEntity(p)

	query := `
		INSERT INTO partners (
			id, tenant_id, name, email, email2, phone, phone2,
			identity, identity_type, location, password_hash,
			last_login, failed_login_attempts, locked_until, active,
			created_at, updated_at, created_by, updated_by
		) VALUES (
			:id, :tenant_id, :name, :email, :email2, :phone, :phone2,
			:identity, :identity_type, :location, :password_hash,
			:last_login, :failed_login_attempts, :locked_until, :active,
			:created_at, :updated_at, :created_by, :updated_by
		)`

	_, err := repo.db.NamedExecContext(ctx, query, row)
	if err != nil {
		repo.logger.Error("Failed to create partner", zap.Error(err), zap.String("partner_id", p.ID.String()))
		return errors.NewInternalError("failed to create partner", err)
	}

	repo.logger.Info("Partner created successfully", zap.String("partner_id", p.ID.String()))
	return nil
}

// GetByID busca um parceiro pelo ID
func (repo *PartnerRepository) GetByID(ctx context.Context, id value_objects.UUID) (*partner.Partner, error) {
	var row partnerRow

	query := `
		SELECT id, tenant_id, name, email, email2, phone, phone2,
			   identity, identity_type, location, password_hash,
			   last_login, failed_login_attempts, locked_until, active,
			   created_at, updated_at, created_by, updated_by
		FROM partners 
		WHERE id = $1 AND active = true`

	err := repo.db.GetContext(ctx, &row, query, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewDomainError("NOT_FOUND", "partner not found", nil)
		}
		repo.logger.Error("Failed to get partner by ID", zap.Error(err), zap.String("partner_id", id.String()))
		return nil, errors.NewInternalError("failed to get partner", err)
	}

	return row.toEntity()
}

// GetByIDAndTenant busca um parceiro pelo ID dentro de um tenant
func (repo *PartnerRepository) GetByIDAndTenant(ctx context.Context, id, tenantID value_objects.UUID) (*partner.Partner, error) {
	var row partnerRow

	query := `
		SELECT id, tenant_id, name, email, email2, phone, phone2,
			   identity, identity_type, location, password_hash,
			   last_login, failed_login_attempts, locked_until, active,
			   created_at, updated_at, created_by, updated_by
		FROM partners 
		WHERE id = $1 AND tenant_id = $2 AND active = true`

	err := repo.db.GetContext(ctx, &row, query, id.String(), tenantID.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewDomainError("NOT_FOUND", "partner not found", nil)
		}
		repo.logger.Error("Failed to get partner by ID and tenant",
			zap.Error(err),
			zap.String("partner_id", id.String()),
			zap.String("tenant_id", tenantID.String()))
		return nil, errors.NewInternalError("failed to get partner", err)
	}

	return row.toEntity()
}

// GetByEmail busca um parceiro pelo email
func (repo *PartnerRepository) GetByEmail(ctx context.Context, email string) (*partner.Partner, error) {
	var row partnerRow

	query := `
		SELECT id, tenant_id, name, email, email2, phone, phone2,
			   identity, identity_type, location, password_hash,
			   last_login, failed_login_attempts, locked_until, active,
			   created_at, updated_at, created_by, updated_by
		FROM partners 
		WHERE email = $1 AND active = true`

	err := repo.db.GetContext(ctx, &row, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewDomainError("NOT_FOUND", "partner not found", nil)
		}
		repo.logger.Error("Failed to get partner by email", zap.Error(err), zap.String("email", email))
		return nil, errors.NewInternalError("failed to get partner", err)
	}

	return row.toEntity()
}

// GetByEmailAndTenant busca um parceiro pelo email dentro de um tenant
func (repo *PartnerRepository) GetByEmailAndTenant(ctx context.Context, email string, tenantID value_objects.UUID) (*partner.Partner, error) {
	var row partnerRow

	query := `
		SELECT id, tenant_id, name, email, email2, phone, phone2,
			   identity, identity_type, location, password_hash,
			   last_login, failed_login_attempts, locked_until, active,
			   created_at, updated_at, created_by, updated_by
		FROM partners 
		WHERE email = $1 AND tenant_id = $2 AND active = true`

	err := repo.db.GetContext(ctx, &row, query, email, tenantID.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewDomainError("NOT_FOUND", "partner not found", nil)
		}
		repo.logger.Error("Failed to get partner by email and tenant",
			zap.Error(err),
			zap.String("email", email),
			zap.String("tenant_id", tenantID.String()))
		return nil, errors.NewInternalError("failed to get partner", err)
	}

	return row.toEntity()
}

// GetByIdentity busca um parceiro pela identidade
func (repo *PartnerRepository) GetByIdentity(ctx context.Context, identity string) (*partner.Partner, error) {
	var row partnerRow

	query := `
		SELECT id, tenant_id, name, email, email2, phone, phone2,
			   identity, identity_type, location, password_hash,
			   last_login, failed_login_attempts, locked_until, active,
			   created_at, updated_at, created_by, updated_by
		FROM partners 
		WHERE identity = $1 AND active = true`

	err := repo.db.GetContext(ctx, &row, query, identity)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewDomainError("NOT_FOUND", "partner not found", nil)
		}
		repo.logger.Error("Failed to get partner by identity", zap.Error(err), zap.String("identity", identity))
		return nil, errors.NewInternalError("failed to get partner", err)
	}

	return row.toEntity()
}

// GetByIdentityAndTenant busca um parceiro pela identidade dentro de um tenant
func (repo *PartnerRepository) GetByIdentityAndTenant(ctx context.Context, identity string, tenantID value_objects.UUID) (*partner.Partner, error) {
	var row partnerRow

	query := `
		SELECT id, tenant_id, name, email, email2, phone, phone2,
			   identity, identity_type, location, password_hash,
			   last_login, failed_login_attempts, locked_until, active,
			   created_at, updated_at, created_by, updated_by
		FROM partners 
		WHERE identity = $1 AND tenant_id = $2 AND active = true`

	err := repo.db.GetContext(ctx, &row, query, identity, tenantID.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewDomainError("NOT_FOUND", "partner not found", nil)
		}
		repo.logger.Error("Failed to get partner by identity and tenant",
			zap.Error(err),
			zap.String("identity", identity),
			zap.String("tenant_id", tenantID.String()))
		return nil, errors.NewInternalError("failed to get partner", err)
	}

	return row.toEntity()
}

// Update atualiza um parceiro existente
func (repo *PartnerRepository) Update(ctx context.Context, p *partner.Partner) error {
	row := repo.fromEntity(p)

	query := `
		UPDATE partners SET
			name = :name,
			email = :email,
			email2 = :email2,
			phone = :phone,
			phone2 = :phone2,
			identity = :identity,
			identity_type = :identity_type,
			location = :location,
			password_hash = :password_hash,
			last_login = :last_login,
			failed_login_attempts = :failed_login_attempts,
			locked_until = :locked_until,
			updated_at = :updated_at,
			updated_by = :updated_by
		WHERE id = :id AND active = true`

	result, err := repo.db.NamedExecContext(ctx, query, row)
	if err != nil {
		repo.logger.Error("Failed to update partner", zap.Error(err), zap.String("partner_id", p.ID.String()))
		return errors.NewInternalError("failed to update partner", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		repo.logger.Error("Failed to get rows affected", zap.Error(err))
		return errors.NewInternalError("failed to update partner", err)
	}

	if rowsAffected == 0 {
		return errors.NewDomainError("NOT_FOUND", "partner not found or inactive", nil)
	}

	repo.logger.Info("Partner updated successfully", zap.String("partner_id", p.ID.String()))
	return nil
}

// Delete remove um parceiro (soft delete)
func (repo *PartnerRepository) Delete(ctx context.Context, id value_objects.UUID, deletedBy value_objects.UUID) error {
	query := `
		UPDATE partners SET
			active = false,
			updated_at = NOW(),
			updated_by = $2
		WHERE id = $1 AND active = true`

	result, err := repo.db.ExecContext(ctx, query, id.String(), deletedBy.String())
	if err != nil {
		repo.logger.Error("Failed to delete partner", zap.Error(err), zap.String("partner_id", id.String()))
		return errors.NewInternalError("failed to delete partner", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		repo.logger.Error("Failed to get rows affected", zap.Error(err))
		return errors.NewInternalError("failed to delete partner", err)
	}

	if rowsAffected == 0 {
		return errors.NewDomainError("NOT_FOUND", "partner not found or already inactive", nil)
	}

	repo.logger.Info("Partner deleted successfully", zap.String("partner_id", id.String()))
	return nil
}

// List lista parceiros com paginação e filtros
func (repo *PartnerRepository) List(ctx context.Context, filters partner.ListFilters) ([]*partner.Partner, int, error) {
	if err := filters.Validate(); err != nil {
		return nil, 0, err
	}

	// Query base
	baseQuery := `FROM partners WHERE active = true`
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

	if filters.Email != nil {
		conditions = append(conditions, fmt.Sprintf("email ILIKE $%d", argIndex))
		args = append(args, "%"+*filters.Email+"%")
		argIndex++
	}

	if filters.Identity != nil {
		conditions = append(conditions, fmt.Sprintf("identity = $%d", argIndex))
		args = append(args, *filters.Identity)
		argIndex++
	}

	if filters.IdentityType != nil {
		conditions = append(conditions, fmt.Sprintf("identity_type = $%d", argIndex))
		args = append(args, *filters.IdentityType)
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
		repo.logger.Error("Failed to count partners", zap.Error(err))
		return nil, 0, errors.NewInternalError("failed to count partners", err)
	}

	// Query de dados
	orderClause := fmt.Sprintf("ORDER BY %s", filters.OrderBy)
	if filters.OrderDesc {
		orderClause += " DESC"
	}

	limitClause := fmt.Sprintf("LIMIT %d OFFSET %d", filters.PageSize, filters.GetOffset())

	dataQuery := `
		SELECT id, tenant_id, name, email, email2, phone, phone2,
			   identity, identity_type, location, password_hash,
			   last_login, failed_login_attempts, locked_until, active,
			   created_at, updated_at, created_by, updated_by ` +
		baseQuery + whereClause + " " + orderClause + " " + limitClause

	var rows []partnerRow
	err = repo.db.SelectContext(ctx, &rows, dataQuery, args...)
	if err != nil {
		repo.logger.Error("Failed to list partners", zap.Error(err))
		return nil, 0, errors.NewInternalError("failed to list partners", err)
	}

	// Converter para entidades
	partners := make([]*partner.Partner, 0, len(rows))
	for _, row := range rows {
		p, err := row.toEntity()
		if err != nil {
			repo.logger.Warn("Failed to convert partner row", zap.Error(err), zap.String("partner_id", row.ID))
			continue
		}
		partners = append(partners, p)
	}

	return partners, total, nil
}

// ListByTenant lista parceiros de um tenant específico
func (repo *PartnerRepository) ListByTenant(ctx context.Context, tenantID value_objects.UUID, filters partner.ListFilters) ([]*partner.Partner, int, error) {
	filters.TenantID = &tenantID
	return repo.List(ctx, filters)
}

// ExistsByEmail verifica se existe um parceiro com o email informado
func (repo *PartnerRepository) ExistsByEmail(ctx context.Context, email string, excludeID *value_objects.UUID) (bool, error) {
	query := `SELECT COUNT(*) FROM partners WHERE email = $1 AND active = true`
	args := []interface{}{email}

	if excludeID != nil {
		query += " AND id != $2"
		args = append(args, excludeID.String())
	}

	var count int
	err := repo.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		repo.logger.Error("Failed to check partner email existence", zap.Error(err))
		return false, errors.NewInternalError("failed to check partner email", err)
	}

	return count > 0, nil
}

// ExistsByEmailInTenant verifica se existe um parceiro com o email no tenant
func (repo *PartnerRepository) ExistsByEmailInTenant(ctx context.Context, email string, tenantID value_objects.UUID, excludeID *value_objects.UUID) (bool, error) {
	query := `SELECT COUNT(*) FROM partners WHERE email = $1 AND tenant_id = $2 AND active = true`
	args := []interface{}{email, tenantID.String()}

	if excludeID != nil {
		query += " AND id != $3"
		args = append(args, excludeID.String())
	}

	var count int
	err := repo.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		repo.logger.Error("Failed to check partner email existence in tenant", zap.Error(err))
		return false, errors.NewInternalError("failed to check partner email", err)
	}

	return count > 0, nil
}

// ExistsByIdentity verifica se existe um parceiro com a identidade informada
func (repo *PartnerRepository) ExistsByIdentity(ctx context.Context, identity string, excludeID *value_objects.UUID) (bool, error) {
	query := `SELECT COUNT(*) FROM partners WHERE identity = $1 AND active = true`
	args := []interface{}{identity}

	if excludeID != nil {
		query += " AND id != $2"
		args = append(args, excludeID.String())
	}

	var count int
	err := repo.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		repo.logger.Error("Failed to check partner identity existence", zap.Error(err))
		return false, errors.NewInternalError("failed to check partner identity", err)
	}

	return count > 0, nil
}

// ExistsByIdentityInTenant verifica se existe um parceiro com a identidade no tenant
func (repo *PartnerRepository) ExistsByIdentityInTenant(ctx context.Context, identity string, tenantID value_objects.UUID, excludeID *value_objects.UUID) (bool, error) {
	query := `SELECT COUNT(*) FROM partners WHERE identity = $1 AND tenant_id = $2 AND active = true`
	args := []interface{}{identity, tenantID.String()}

	if excludeID != nil {
		query += " AND id != $3"
		args = append(args, excludeID.String())
	}

	var count int
	err := repo.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		repo.logger.Error("Failed to check partner identity existence in tenant", zap.Error(err))
		return false, errors.NewInternalError("failed to check partner identity", err)
	}

	return count > 0, nil
}

// ListByEvent lista parceiros associados a um evento específico
func (repo *PartnerRepository) ListByEvent(ctx context.Context, eventID value_objects.UUID, filters partner.ListFilters) ([]*partner.Partner, int, error) {
	// TODO: Implementar quando houver relacionamento Partner-Event
	// Por enquanto, retorna lista vazia
	return []*partner.Partner{}, 0, nil
}

// GetPartnersWithEmployees busca parceiros que têm funcionários
func (repo *PartnerRepository) GetPartnersWithEmployees(ctx context.Context, tenantID value_objects.UUID, filters partner.ListFilters) ([]*partner.Partner, int, error) {
	// TODO: Implementar quando houver relacionamento Partner-Employee
	// Por enquanto, retorna lista vazia
	return []*partner.Partner{}, 0, nil
}
