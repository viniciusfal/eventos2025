package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"eventos-backend/internal/domain/role"
	"eventos-backend/internal/domain/shared/value_objects"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// RoleRepository implementa a interface de repositório para Role
type RoleRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// NewRoleRepository cria uma nova instância do repositório de role
func NewRoleRepository(db *sqlx.DB, logger *zap.Logger) role.Repository {
	return &RoleRepository{
		db:     db,
		logger: logger,
	}
}

// roleRow representa uma linha da tabela role no banco
type roleRow struct {
	ID          string         `db:"id_role"`
	TenantID    sql.NullString `db:"id_tenant"`
	Name        string         `db:"name"`
	DisplayName string         `db:"display_name"`
	Description sql.NullString `db:"description"`
	Level       int            `db:"level"`
	IsSystem    bool           `db:"is_system"`
	Active      bool           `db:"active"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
	CreatedBy   sql.NullString `db:"created_by"`
	UpdatedBy   sql.NullString `db:"updated_by"`
}

// toEntity converte uma linha do banco para entidade de domínio
func (r *roleRow) toEntity() (*role.Role, error) {
	id, err := value_objects.ParseUUID(r.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid role ID: %w", err)
	}

	roleEntity := &role.Role{
		ID:          id,
		Name:        r.Name,
		DisplayName: r.DisplayName,
		Level:       r.Level,
		IsSystem:    r.IsSystem,
		Active:      r.Active,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}

	// TenantID (pode ser null para roles do sistema)
	if r.TenantID.Valid {
		tenantID, err := value_objects.ParseUUID(r.TenantID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid tenant ID: %w", err)
		}
		roleEntity.TenantID = tenantID
	}

	// Description
	if r.Description.Valid {
		roleEntity.Description = r.Description.String
	}

	// CreatedBy
	if r.CreatedBy.Valid {
		createdBy, err := value_objects.ParseUUID(r.CreatedBy.String)
		if err == nil {
			roleEntity.CreatedBy = &createdBy
		}
	}

	// UpdatedBy
	if r.UpdatedBy.Valid {
		updatedBy, err := value_objects.ParseUUID(r.UpdatedBy.String)
		if err == nil {
			roleEntity.UpdatedBy = &updatedBy
		}
	}

	return roleEntity, nil
}

// fromEntity converte uma entidade de domínio para linha do banco
func (repo *RoleRepository) fromEntity(r *role.Role) *roleRow {
	row := &roleRow{
		ID:          r.ID.String(),
		Name:        r.Name,
		DisplayName: r.DisplayName,
		Level:       r.Level,
		IsSystem:    r.IsSystem,
		Active:      r.Active,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}

	// TenantID (null para roles do sistema)
	if !r.IsSystem && !r.TenantID.IsZero() {
		row.TenantID = sql.NullString{String: r.TenantID.String(), Valid: true}
	}

	// Description
	if r.Description != "" {
		row.Description = sql.NullString{String: r.Description, Valid: true}
	}

	// CreatedBy
	if r.CreatedBy != nil {
		row.CreatedBy = sql.NullString{String: r.CreatedBy.String(), Valid: true}
	}

	// UpdatedBy
	if r.UpdatedBy != nil {
		row.UpdatedBy = sql.NullString{String: r.UpdatedBy.String(), Valid: true}
	}

	return row
}

// Create cria uma nova role
func (repo *RoleRepository) Create(ctx context.Context, r *role.Role) error {
	row := repo.fromEntity(r)

	query := `
		INSERT INTO role (
			id_role, id_tenant, name, display_name, description,
			level, is_system, active, created_at, updated_at,
			created_by, updated_by
		) VALUES (
			:id_role, :id_tenant, :name, :display_name, :description,
			:level, :is_system, :active, :created_at, :updated_at,
			:created_by, :updated_by
		)`

	_, err := repo.db.NamedExecContext(ctx, query, row)
	if err != nil {
		repo.logger.Error("Failed to create role", zap.Error(err), zap.String("role_id", r.ID.String()))
		return fmt.Errorf("failed to create role: %w", err)
	}

	repo.logger.Info("Role created successfully", zap.String("role_id", r.ID.String()))
	return nil
}

// GetByID busca uma role por ID
func (repo *RoleRepository) GetByID(ctx context.Context, id value_objects.UUID) (*role.Role, error) {
	var row roleRow
	query := `
		SELECT id_role, id_tenant, name, display_name, description,
			   level, is_system, active, created_at, updated_at,
			   created_by, updated_by
		FROM role 
		WHERE id_role = $1`

	err := repo.db.GetContext(ctx, &row, query, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("role not found")
		}
		repo.logger.Error("Failed to get role by ID", zap.Error(err), zap.String("role_id", id.String()))
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return row.toEntity()
}

// GetByName busca uma role por nome dentro de um tenant
func (repo *RoleRepository) GetByName(ctx context.Context, tenantID value_objects.UUID, name string) (*role.Role, error) {
	var row roleRow
	query := `
		SELECT id_role, id_tenant, name, display_name, description,
			   level, is_system, active, created_at, updated_at,
			   created_by, updated_by
		FROM role 
		WHERE id_tenant = $1 AND name = $2`

	err := repo.db.GetContext(ctx, &row, query, tenantID.String(), strings.ToUpper(name))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("role not found")
		}
		repo.logger.Error("Failed to get role by name", zap.Error(err), zap.String("tenant_id", tenantID.String()), zap.String("name", name))
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return row.toEntity()
}

// GetSystemRoleByName busca uma role do sistema por nome
func (repo *RoleRepository) GetSystemRoleByName(ctx context.Context, name string) (*role.Role, error) {
	var row roleRow
	query := `
		SELECT id_role, id_tenant, name, display_name, description,
			   level, is_system, active, created_at, updated_at,
			   created_by, updated_by
		FROM role 
		WHERE is_system = true AND name = $1`

	err := repo.db.GetContext(ctx, &row, query, strings.ToUpper(name))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("system role not found")
		}
		repo.logger.Error("Failed to get system role by name", zap.Error(err), zap.String("name", name))
		return nil, fmt.Errorf("failed to get system role: %w", err)
	}

	return row.toEntity()
}

// Update atualiza uma role existente
func (repo *RoleRepository) Update(ctx context.Context, r *role.Role) error {
	row := repo.fromEntity(r)

	query := `
		UPDATE role SET
			display_name = :display_name,
			description = :description,
			level = :level,
			active = :active,
			updated_at = :updated_at,
			updated_by = :updated_by
		WHERE id_role = :id_role`

	result, err := repo.db.NamedExecContext(ctx, query, row)
	if err != nil {
		repo.logger.Error("Failed to update role", zap.Error(err), zap.String("role_id", r.ID.String()))
		return fmt.Errorf("failed to update role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("role not found")
	}

	repo.logger.Info("Role updated successfully", zap.String("role_id", r.ID.String()))
	return nil
}

// Delete remove uma role (soft delete)
func (repo *RoleRepository) Delete(ctx context.Context, id value_objects.UUID, deletedBy value_objects.UUID) error {
	query := `
		UPDATE role SET
			active = false,
			updated_at = NOW(),
			updated_by = $2
		WHERE id_role = $1`

	result, err := repo.db.ExecContext(ctx, query, id.String(), deletedBy.String())
	if err != nil {
		repo.logger.Error("Failed to delete role", zap.Error(err), zap.String("role_id", id.String()))
		return fmt.Errorf("failed to delete role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("role not found")
	}

	repo.logger.Info("Role deleted successfully", zap.String("role_id", id.String()))
	return nil
}

// List lista roles com filtros e paginação
func (repo *RoleRepository) List(ctx context.Context, filters role.ListFilters) ([]*role.Role, int, error) {
	// Construir query base
	baseQuery := `
		FROM role r
		WHERE 1=1`

	var args []interface{}
	var conditions []string
	argCount := 0

	// Aplicar filtros
	if filters.HasTenantFilter() {
		argCount++
		conditions = append(conditions, fmt.Sprintf("r.id_tenant = $%d", argCount))
		args = append(args, filters.TenantID.String())
	}

	if filters.HasActiveFilter() {
		argCount++
		conditions = append(conditions, fmt.Sprintf("r.active = $%d", argCount))
		args = append(args, *filters.Active)
	}

	if filters.HasSystemFilter() {
		argCount++
		conditions = append(conditions, fmt.Sprintf("r.is_system = $%d", argCount))
		args = append(args, *filters.IsSystem)
	}

	if filters.HasLevelFilter() {
		argCount++
		conditions = append(conditions, fmt.Sprintf("r.level = $%d", argCount))
		args = append(args, *filters.Level)
	}

	if filters.HasLevelRangeFilter() {
		if filters.MinLevel != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("r.level >= $%d", argCount))
			args = append(args, *filters.MinLevel)
		}
		if filters.MaxLevel != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("r.level <= $%d", argCount))
			args = append(args, *filters.MaxLevel)
		}
	}

	if filters.HasSearchFilter() {
		searchTerm := "%" + strings.ToLower(filters.GetSearchTerm()) + "%"
		argCount++
		conditions = append(conditions, fmt.Sprintf("(LOWER(r.name) LIKE $%d OR LOWER(r.display_name) LIKE $%d OR LOWER(r.description) LIKE $%d)", argCount, argCount, argCount))
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
		repo.logger.Error("Failed to count roles", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count roles: %w", err)
	}

	// Query para buscar dados com paginação
	selectQuery := `
		SELECT r.id_role, r.id_tenant, r.name, r.display_name, r.description,
			   r.level, r.is_system, r.active, r.created_at, r.updated_at,
			   r.created_by, r.updated_by ` + baseQuery

	// Adicionar ordenação
	orderDirection := "ASC"
	if filters.OrderDesc {
		orderDirection = "DESC"
	}
	selectQuery += fmt.Sprintf(" ORDER BY r.%s %s", filters.OrderBy, orderDirection)

	// Adicionar paginação
	selectQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", filters.GetLimit(), filters.GetOffset())

	var rows []roleRow
	err = repo.db.SelectContext(ctx, &rows, selectQuery, args...)
	if err != nil {
		repo.logger.Error("Failed to list roles", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list roles: %w", err)
	}

	// Converter para entidades
	roles := make([]*role.Role, len(rows))
	for i, row := range rows {
		role, err := row.toEntity()
		if err != nil {
			repo.logger.Error("Failed to convert role row to entity", zap.Error(err))
			return nil, 0, fmt.Errorf("failed to convert role: %w", err)
		}
		roles[i] = role
	}

	return roles, total, nil
}

// ListByTenant lista roles de um tenant específico
func (repo *RoleRepository) ListByTenant(ctx context.Context, tenantID value_objects.UUID, filters role.ListFilters) ([]*role.Role, int, error) {
	// Definir tenant no filtro
	filters.TenantID = &tenantID
	return repo.List(ctx, filters)
}

// ListSystemRoles lista todas as roles do sistema
func (repo *RoleRepository) ListSystemRoles(ctx context.Context) ([]*role.Role, error) {
	query := `
		SELECT id_role, id_tenant, name, display_name, description,
			   level, is_system, active, created_at, updated_at,
			   created_by, updated_by
		FROM role 
		WHERE is_system = true
		ORDER BY level ASC`

	var rows []roleRow
	err := repo.db.SelectContext(ctx, &rows, query)
	if err != nil {
		repo.logger.Error("Failed to list system roles", zap.Error(err))
		return nil, fmt.Errorf("failed to list system roles: %w", err)
	}

	// Converter para entidades
	roles := make([]*role.Role, len(rows))
	for i, row := range rows {
		role, err := row.toEntity()
		if err != nil {
			repo.logger.Error("Failed to convert role row to entity", zap.Error(err))
			return nil, fmt.Errorf("failed to convert role: %w", err)
		}
		roles[i] = role
	}

	return roles, nil
}

// ExistsByName verifica se existe uma role com o nome especificado no tenant
func (repo *RoleRepository) ExistsByName(ctx context.Context, tenantID value_objects.UUID, name string, excludeID *value_objects.UUID) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM role 
		WHERE id_tenant = $1 AND name = $2`
	args := []interface{}{tenantID.String(), strings.ToUpper(name)}

	if excludeID != nil {
		query += " AND id_role != $3"
		args = append(args, excludeID.String())
	}

	var count int
	err := repo.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		repo.logger.Error("Failed to check role existence", zap.Error(err))
		return false, fmt.Errorf("failed to check role existence: %w", err)
	}

	return count > 0, nil
}

// GetRolesByLevel busca roles por nível hierárquico
func (repo *RoleRepository) GetRolesByLevel(ctx context.Context, tenantID value_objects.UUID, level int) ([]*role.Role, error) {
	query := `
		SELECT id_role, id_tenant, name, display_name, description,
			   level, is_system, active, created_at, updated_at,
			   created_by, updated_by
		FROM role 
		WHERE id_tenant = $1 AND level = $2
		ORDER BY name ASC`

	var rows []roleRow
	err := repo.db.SelectContext(ctx, &rows, query, tenantID.String(), level)
	if err != nil {
		repo.logger.Error("Failed to get roles by level", zap.Error(err))
		return nil, fmt.Errorf("failed to get roles by level: %w", err)
	}

	// Converter para entidades
	roles := make([]*role.Role, len(rows))
	for i, row := range rows {
		role, err := row.toEntity()
		if err != nil {
			repo.logger.Error("Failed to convert role row to entity", zap.Error(err))
			return nil, fmt.Errorf("failed to convert role: %w", err)
		}
		roles[i] = role
	}

	return roles, nil
}

// GetRolesByLevelRange busca roles dentro de um range de níveis
func (repo *RoleRepository) GetRolesByLevelRange(ctx context.Context, tenantID value_objects.UUID, minLevel, maxLevel int) ([]*role.Role, error) {
	query := `
		SELECT id_role, id_tenant, name, display_name, description,
			   level, is_system, active, created_at, updated_at,
			   created_by, updated_by
		FROM role 
		WHERE id_tenant = $1 AND level >= $2 AND level <= $3
		ORDER BY level ASC, name ASC`

	var rows []roleRow
	err := repo.db.SelectContext(ctx, &rows, query, tenantID.String(), minLevel, maxLevel)
	if err != nil {
		repo.logger.Error("Failed to get roles by level range", zap.Error(err))
		return nil, fmt.Errorf("failed to get roles by level range: %w", err)
	}

	// Converter para entidades
	roles := make([]*role.Role, len(rows))
	for i, row := range rows {
		role, err := row.toEntity()
		if err != nil {
			repo.logger.Error("Failed to convert role row to entity", zap.Error(err))
			return nil, fmt.Errorf("failed to convert role: %w", err)
		}
		roles[i] = role
	}

	return roles, nil
}

// CountByTenant conta o número de roles de um tenant
func (repo *RoleRepository) CountByTenant(ctx context.Context, tenantID value_objects.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM role 
		WHERE id_tenant = $1`

	var count int
	err := repo.db.GetContext(ctx, &count, query, tenantID.String())
	if err != nil {
		repo.logger.Error("Failed to count roles by tenant", zap.Error(err))
		return 0, fmt.Errorf("failed to count roles: %w", err)
	}

	return count, nil
}

// GetHighestLevelRole busca a role de nível mais alto de um tenant
func (repo *RoleRepository) GetHighestLevelRole(ctx context.Context, tenantID value_objects.UUID) (*role.Role, error) {
	var row roleRow
	query := `
		SELECT id_role, id_tenant, name, display_name, description,
			   level, is_system, active, created_at, updated_at,
			   created_by, updated_by
		FROM role 
		WHERE id_tenant = $1 AND active = true
		ORDER BY level ASC
		LIMIT 1`

	err := repo.db.GetContext(ctx, &row, query, tenantID.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no roles found for tenant")
		}
		repo.logger.Error("Failed to get highest level role", zap.Error(err))
		return nil, fmt.Errorf("failed to get highest level role: %w", err)
	}

	return row.toEntity()
}

// GetLowestLevelRole busca a role de nível mais baixo de um tenant
func (repo *RoleRepository) GetLowestLevelRole(ctx context.Context, tenantID value_objects.UUID) (*role.Role, error) {
	var row roleRow
	query := `
		SELECT id_role, id_tenant, name, display_name, description,
			   level, is_system, active, created_at, updated_at,
			   created_by, updated_by
		FROM role 
		WHERE id_tenant = $1 AND active = true
		ORDER BY level DESC
		LIMIT 1`

	err := repo.db.GetContext(ctx, &row, query, tenantID.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no roles found for tenant")
		}
		repo.logger.Error("Failed to get lowest level role", zap.Error(err))
		return nil, fmt.Errorf("failed to get lowest level role: %w", err)
	}

	return row.toEntity()
}
