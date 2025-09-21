package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"eventos-backend/internal/domain/shared/value_objects"
	"eventos-backend/internal/domain/tenant"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// TenantRepository implementa a interface de repositório para Tenant
type TenantRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// NewTenantRepository cria uma nova instância do repositório de tenant
func NewTenantRepository(db *sqlx.DB, logger *zap.Logger) tenant.Repository {
	return &TenantRepository{
		db:     db,
		logger: logger,
	}
}

// tenantRow representa uma linha da tabela tenant no banco
type tenantRow struct {
	ID           string         `db:"id_tenant"`
	ConfigID     sql.NullString `db:"id_config_tenant"`
	Name         string         `db:"name"`
	Identity     sql.NullString `db:"identity"`
	IdentityType sql.NullString `db:"type_identity"`
	Email        sql.NullString `db:"email"`
	Address      sql.NullString `db:"address"`
	Active       bool           `db:"active"`
	CreatedAt    time.Time      `db:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at"`
	CreatedBy    sql.NullString `db:"created_by"`
	UpdatedBy    sql.NullString `db:"updated_by"`
}

// toEntity converte uma linha do banco para entidade de domínio
func (r *tenantRow) toEntity() (*tenant.Tenant, error) {
	id, err := value_objects.ParseUUID(r.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	t := &tenant.Tenant{
		ID:        id,
		Name:      r.Name,
		Active:    r.Active,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}

	// Campos opcionais
	if r.ConfigID.Valid {
		configID, err := value_objects.ParseUUID(r.ConfigID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid config ID: %w", err)
		}
		t.ConfigID = &configID
	}

	if r.Identity.Valid {
		t.Identity = r.Identity.String
	}

	if r.IdentityType.Valid {
		t.IdentityType = r.IdentityType.String
	}

	if r.Email.Valid {
		t.Email = r.Email.String
	}

	if r.Address.Valid {
		t.Address = r.Address.String
	}

	if r.CreatedBy.Valid {
		createdBy, err := value_objects.ParseUUID(r.CreatedBy.String)
		if err != nil {
			return nil, fmt.Errorf("invalid created_by ID: %w", err)
		}
		t.CreatedBy = &createdBy
	}

	if r.UpdatedBy.Valid {
		updatedBy, err := value_objects.ParseUUID(r.UpdatedBy.String)
		if err != nil {
			return nil, fmt.Errorf("invalid updated_by ID: %w", err)
		}
		t.UpdatedBy = &updatedBy
	}

	return t, nil
}

// fromEntity converte uma entidade de domínio para linha do banco
func (repo *TenantRepository) fromEntity(t *tenant.Tenant) *tenantRow {
	row := &tenantRow{
		ID:        t.ID.String(),
		Name:      t.Name,
		Active:    t.Active,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}

	// Campos opcionais
	if t.ConfigID != nil && !t.ConfigID.IsZero() {
		row.ConfigID = sql.NullString{String: t.ConfigID.String(), Valid: true}
	}

	if t.Identity != "" {
		row.Identity = sql.NullString{String: t.Identity, Valid: true}
	}

	if t.IdentityType != "" {
		row.IdentityType = sql.NullString{String: t.IdentityType, Valid: true}
	}

	if t.Email != "" {
		row.Email = sql.NullString{String: t.Email, Valid: true}
	}

	if t.Address != "" {
		row.Address = sql.NullString{String: t.Address, Valid: true}
	}

	if t.CreatedBy != nil && !t.CreatedBy.IsZero() {
		row.CreatedBy = sql.NullString{String: t.CreatedBy.String(), Valid: true}
	}

	if t.UpdatedBy != nil && !t.UpdatedBy.IsZero() {
		row.UpdatedBy = sql.NullString{String: t.UpdatedBy.String(), Valid: true}
	}

	return row
}

// Create cria um novo tenant
func (repo *TenantRepository) Create(ctx context.Context, t *tenant.Tenant) error {
	query := `
		INSERT INTO tenant (
			id_tenant, id_config_tenant, name, identity, type_identity, 
			email, address, active, created_at, updated_at, created_by, updated_by
		) VALUES (
			:id_tenant, :id_config_tenant, :name, :identity, :type_identity,
			:email, :address, :active, :created_at, :updated_at, :created_by, :updated_by
		)`

	row := repo.fromEntity(t)

	_, err := repo.db.NamedExecContext(ctx, query, row)
	if err != nil {
		repo.logger.Error("Failed to create tenant", zap.Error(err))
		return fmt.Errorf("failed to create tenant: %w", err)
	}

	return nil
}

// GetByID busca um tenant pelo ID
func (repo *TenantRepository) GetByID(ctx context.Context, id value_objects.UUID) (*tenant.Tenant, error) {
	query := `
		SELECT id_tenant, id_config_tenant, name, identity, type_identity,
		       email, address, active, created_at, updated_at, created_by, updated_by
		FROM tenant 
		WHERE id_tenant = $1`

	var row tenantRow
	err := repo.db.GetContext(ctx, &row, query, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		repo.logger.Error("Failed to get tenant by ID", zap.Error(err))
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return row.toEntity()
}

// GetByIdentity busca um tenant pela identidade
func (repo *TenantRepository) GetByIdentity(ctx context.Context, identity string) (*tenant.Tenant, error) {
	query := `
		SELECT id_tenant, id_config_tenant, name, identity, type_identity,
		       email, address, active, created_at, updated_at, created_by, updated_by
		FROM tenant 
		WHERE identity = $1`

	var row tenantRow
	err := repo.db.GetContext(ctx, &row, query, identity)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		repo.logger.Error("Failed to get tenant by identity", zap.Error(err))
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return row.toEntity()
}

// GetByEmail busca um tenant pelo email
func (repo *TenantRepository) GetByEmail(ctx context.Context, email string) (*tenant.Tenant, error) {
	query := `
		SELECT id_tenant, id_config_tenant, name, identity, type_identity,
		       email, address, active, created_at, updated_at, created_by, updated_by
		FROM tenant 
		WHERE email = $1`

	var row tenantRow
	err := repo.db.GetContext(ctx, &row, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		repo.logger.Error("Failed to get tenant by email", zap.Error(err))
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return row.toEntity()
}

// Update atualiza um tenant existente
func (repo *TenantRepository) Update(ctx context.Context, t *tenant.Tenant) error {
	query := `
		UPDATE tenant SET
			id_config_tenant = :id_config_tenant,
			name = :name,
			identity = :identity,
			type_identity = :type_identity,
			email = :email,
			address = :address,
			active = :active,
			updated_at = :updated_at,
			updated_by = :updated_by
		WHERE id_tenant = :id_tenant`

	row := repo.fromEntity(t)

	result, err := repo.db.NamedExecContext(ctx, query, row)
	if err != nil {
		repo.logger.Error("Failed to update tenant", zap.Error(err))
		return fmt.Errorf("failed to update tenant: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tenant not found")
	}

	return nil
}

// Delete remove um tenant (soft delete)
func (repo *TenantRepository) Delete(ctx context.Context, id value_objects.UUID, deletedBy value_objects.UUID) error {
	query := `
		UPDATE tenant SET
			active = false,
			updated_at = CURRENT_TIMESTAMP,
			updated_by = $2
		WHERE id_tenant = $1`

	result, err := repo.db.ExecContext(ctx, query, id.String(), deletedBy.String())
	if err != nil {
		repo.logger.Error("Failed to delete tenant", zap.Error(err))
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tenant not found")
	}

	return nil
}

// List lista tenants com paginação e filtros
func (repo *TenantRepository) List(ctx context.Context, filters tenant.ListFilters) ([]*tenant.Tenant, int, error) {
	// Construir query base
	baseQuery := `
		SELECT id_tenant, id_config_tenant, name, identity, type_identity,
		       email, address, active, created_at, updated_at, created_by, updated_by
		FROM tenant`

	countQuery := "SELECT COUNT(*) FROM tenant"

	// Construir condições WHERE
	var conditions []string
	var args []interface{}
	argIndex := 1

	if filters.Name != nil {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIndex))
		args = append(args, "%"+*filters.Name+"%")
		argIndex++
	}

	if filters.Identity != nil {
		conditions = append(conditions, fmt.Sprintf("identity = $%d", argIndex))
		args = append(args, *filters.Identity)
		argIndex++
	}

	if filters.IdentityType != nil {
		conditions = append(conditions, fmt.Sprintf("type_identity = $%d", argIndex))
		args = append(args, *filters.IdentityType)
		argIndex++
	}

	if filters.Email != nil {
		conditions = append(conditions, fmt.Sprintf("email ILIKE $%d", argIndex))
		args = append(args, "%"+*filters.Email+"%")
		argIndex++
	}

	if filters.Active != nil {
		conditions = append(conditions, fmt.Sprintf("active = $%d", argIndex))
		args = append(args, *filters.Active)
		argIndex++
	}

	// Adicionar WHERE se houver condições
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// Contar total de registros
	var total int
	err := repo.db.GetContext(ctx, &total, countQuery+whereClause, args...)
	if err != nil {
		repo.logger.Error("Failed to count tenants", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count tenants: %w", err)
	}

	// Construir query com ordenação e paginação
	orderClause := fmt.Sprintf(" ORDER BY %s", filters.OrderBy)
	if filters.OrderDesc {
		orderClause += " DESC"
	}

	limitClause := fmt.Sprintf(" LIMIT %d OFFSET %d", filters.PageSize, filters.GetOffset())

	finalQuery := baseQuery + whereClause + orderClause + limitClause

	// Executar query
	var rows []tenantRow
	err = repo.db.SelectContext(ctx, &rows, finalQuery, args...)
	if err != nil {
		repo.logger.Error("Failed to list tenants", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list tenants: %w", err)
	}

	// Converter para entidades
	tenants := make([]*tenant.Tenant, len(rows))
	for i, row := range rows {
		t, err := row.toEntity()
		if err != nil {
			repo.logger.Error("Failed to convert tenant row to entity", zap.Error(err))
			return nil, 0, fmt.Errorf("failed to convert tenant: %w", err)
		}
		tenants[i] = t
	}

	return tenants, total, nil
}

// ExistsByIdentity verifica se existe um tenant com a identidade informada
func (repo *TenantRepository) ExistsByIdentity(ctx context.Context, identity string, excludeID *value_objects.UUID) (bool, error) {
	query := "SELECT COUNT(*) FROM tenant WHERE identity = $1"
	args := []interface{}{identity}

	if excludeID != nil {
		query += " AND id_tenant != $2"
		args = append(args, excludeID.String())
	}

	var count int
	err := repo.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		repo.logger.Error("Failed to check identity existence", zap.Error(err))
		return false, fmt.Errorf("failed to check identity existence: %w", err)
	}

	return count > 0, nil
}

// ExistsByEmail verifica se existe um tenant com o email informado
func (repo *TenantRepository) ExistsByEmail(ctx context.Context, email string, excludeID *value_objects.UUID) (bool, error) {
	query := "SELECT COUNT(*) FROM tenant WHERE email = $1"
	args := []interface{}{email}

	if excludeID != nil {
		query += " AND id_tenant != $2"
		args = append(args, excludeID.String())
	}

	var count int
	err := repo.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		repo.logger.Error("Failed to check email existence", zap.Error(err))
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return count > 0, nil
}
