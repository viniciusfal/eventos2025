package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"eventos-backend/internal/domain/permission"
	"eventos-backend/internal/domain/shared/value_objects"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// PermissionRepository implementa a interface de repositório para Permission
type PermissionRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// NewPermissionRepository cria uma nova instância do repositório de permission
func NewPermissionRepository(db *sqlx.DB, logger *zap.Logger) permission.Repository {
	return &PermissionRepository{
		db:     db,
		logger: logger,
	}
}

// permissionRow representa uma linha da tabela permission no banco
type permissionRow struct {
	ID          string         `db:"id_permission"`
	TenantID    sql.NullString `db:"id_tenant"`
	Module      string         `db:"module"`
	Action      string         `db:"action"`
	Resource    sql.NullString `db:"resource"`
	Name        string         `db:"name"`
	DisplayName string         `db:"display_name"`
	Description sql.NullString `db:"description"`
	IsSystem    bool           `db:"is_system"`
	Active      bool           `db:"active"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
	CreatedBy   sql.NullString `db:"created_by"`
	UpdatedBy   sql.NullString `db:"updated_by"`
}

// toEntity converte uma linha do banco para entidade de domínio
func (r *permissionRow) toEntity() (*permission.Permission, error) {
	id, err := value_objects.ParseUUID(r.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid permission ID: %w", err)
	}

	permissionEntity := &permission.Permission{
		ID:          id,
		Module:      r.Module,
		Action:      r.Action,
		Name:        r.Name,
		DisplayName: r.DisplayName,
		IsSystem:    r.IsSystem,
		Active:      r.Active,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}

	// TenantID (pode ser null para permissions do sistema)
	if r.TenantID.Valid {
		tenantID, err := value_objects.ParseUUID(r.TenantID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid tenant ID: %w", err)
		}
		permissionEntity.TenantID = tenantID
	}

	// Resource
	if r.Resource.Valid {
		permissionEntity.Resource = r.Resource.String
	}

	// Description
	if r.Description.Valid {
		permissionEntity.Description = r.Description.String
	}

	// CreatedBy
	if r.CreatedBy.Valid {
		createdBy, err := value_objects.ParseUUID(r.CreatedBy.String)
		if err == nil {
			permissionEntity.CreatedBy = &createdBy
		}
	}

	// UpdatedBy
	if r.UpdatedBy.Valid {
		updatedBy, err := value_objects.ParseUUID(r.UpdatedBy.String)
		if err == nil {
			permissionEntity.UpdatedBy = &updatedBy
		}
	}

	return permissionEntity, nil
}

// fromEntity converte uma entidade de domínio para linha do banco
func (repo *PermissionRepository) fromEntity(p *permission.Permission) *permissionRow {
	row := &permissionRow{
		ID:          p.ID.String(),
		Module:      p.Module,
		Action:      p.Action,
		Name:        p.Name,
		DisplayName: p.DisplayName,
		IsSystem:    p.IsSystem,
		Active:      p.Active,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}

	// TenantID (null para permissions do sistema)
	if !p.IsSystem && !p.TenantID.IsZero() {
		row.TenantID = sql.NullString{String: p.TenantID.String(), Valid: true}
	}

	// Resource
	if p.Resource != "" {
		row.Resource = sql.NullString{String: p.Resource, Valid: true}
	}

	// Description
	if p.Description != "" {
		row.Description = sql.NullString{String: p.Description, Valid: true}
	}

	// CreatedBy
	if p.CreatedBy != nil {
		row.CreatedBy = sql.NullString{String: p.CreatedBy.String(), Valid: true}
	}

	// UpdatedBy
	if p.UpdatedBy != nil {
		row.UpdatedBy = sql.NullString{String: p.UpdatedBy.String(), Valid: true}
	}

	return row
}

// Create cria uma nova permission
func (repo *PermissionRepository) Create(ctx context.Context, p *permission.Permission) error {
	row := repo.fromEntity(p)

	query := `
		INSERT INTO permission (
			id_permission, id_tenant, module, action, resource,
			name, display_name, description, is_system, active,
			created_at, updated_at, created_by, updated_by
		) VALUES (
			:id_permission, :id_tenant, :module, :action, :resource,
			:name, :display_name, :description, :is_system, :active,
			:created_at, :updated_at, :created_by, :updated_by
		)`

	_, err := repo.db.NamedExecContext(ctx, query, row)
	if err != nil {
		repo.logger.Error("Failed to create permission", zap.Error(err), zap.String("permission_id", p.ID.String()))
		return fmt.Errorf("failed to create permission: %w", err)
	}

	repo.logger.Info("Permission created successfully", zap.String("permission_id", p.ID.String()))
	return nil
}

// GetByID busca uma permission por ID
func (repo *PermissionRepository) GetByID(ctx context.Context, id value_objects.UUID) (*permission.Permission, error) {
	var row permissionRow
	query := `
		SELECT id_permission, id_tenant, module, action, resource,
			   name, display_name, description, is_system, active,
			   created_at, updated_at, created_by, updated_by
		FROM permission 
		WHERE id_permission = $1`

	err := repo.db.GetContext(ctx, &row, query, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("permission not found")
		}
		repo.logger.Error("Failed to get permission by ID", zap.Error(err), zap.String("permission_id", id.String()))
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	return row.toEntity()
}

// GetByName busca uma permission por nome dentro de um tenant
func (repo *PermissionRepository) GetByName(ctx context.Context, tenantID value_objects.UUID, name string) (*permission.Permission, error) {
	var row permissionRow
	query := `
		SELECT id_permission, id_tenant, module, action, resource,
			   name, display_name, description, is_system, active,
			   created_at, updated_at, created_by, updated_by
		FROM permission 
		WHERE id_tenant = $1 AND name = $2`

	err := repo.db.GetContext(ctx, &row, query, tenantID.String(), strings.ToUpper(name))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("permission not found")
		}
		repo.logger.Error("Failed to get permission by name", zap.Error(err), zap.String("tenant_id", tenantID.String()), zap.String("name", name))
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	return row.toEntity()
}

// GetSystemPermissionByName busca uma permission do sistema por nome
func (repo *PermissionRepository) GetSystemPermissionByName(ctx context.Context, name string) (*permission.Permission, error) {
	var row permissionRow
	query := `
		SELECT id_permission, id_tenant, module, action, resource,
			   name, display_name, description, is_system, active,
			   created_at, updated_at, created_by, updated_by
		FROM permission 
		WHERE is_system = true AND name = $1`

	err := repo.db.GetContext(ctx, &row, query, strings.ToUpper(name))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("system permission not found")
		}
		repo.logger.Error("Failed to get system permission by name", zap.Error(err), zap.String("name", name))
		return nil, fmt.Errorf("failed to get system permission: %w", err)
	}

	return row.toEntity()
}

// Update atualiza uma permission existente
func (repo *PermissionRepository) Update(ctx context.Context, p *permission.Permission) error {
	row := repo.fromEntity(p)

	query := `
		UPDATE permission SET
			display_name = :display_name,
			description = :description,
			active = :active,
			updated_at = :updated_at,
			updated_by = :updated_by
		WHERE id_permission = :id_permission`

	result, err := repo.db.NamedExecContext(ctx, query, row)
	if err != nil {
		repo.logger.Error("Failed to update permission", zap.Error(err), zap.String("permission_id", p.ID.String()))
		return fmt.Errorf("failed to update permission: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("permission not found")
	}

	repo.logger.Info("Permission updated successfully", zap.String("permission_id", p.ID.String()))
	return nil
}

// Delete remove uma permission (soft delete)
func (repo *PermissionRepository) Delete(ctx context.Context, id value_objects.UUID, deletedBy value_objects.UUID) error {
	query := `
		UPDATE permission SET
			active = false,
			updated_at = NOW(),
			updated_by = $2
		WHERE id_permission = $1`

	result, err := repo.db.ExecContext(ctx, query, id.String(), deletedBy.String())
	if err != nil {
		repo.logger.Error("Failed to delete permission", zap.Error(err), zap.String("permission_id", id.String()))
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("permission not found")
	}

	repo.logger.Info("Permission deleted successfully", zap.String("permission_id", id.String()))
	return nil
}

// List lista permissions com filtros e paginação
func (repo *PermissionRepository) List(ctx context.Context, filters permission.ListFilters) ([]*permission.Permission, int, error) {
	// Construir query base
	baseQuery := `
		FROM permission p
		WHERE 1=1`

	var args []interface{}
	var conditions []string
	argCount := 0

	// Aplicar filtros
	if filters.HasTenantFilter() {
		argCount++
		conditions = append(conditions, fmt.Sprintf("p.id_tenant = $%d", argCount))
		args = append(args, filters.TenantID.String())
	}

	if filters.HasActiveFilter() {
		argCount++
		conditions = append(conditions, fmt.Sprintf("p.active = $%d", argCount))
		args = append(args, *filters.Active)
	}

	if filters.HasSystemFilter() {
		argCount++
		conditions = append(conditions, fmt.Sprintf("p.is_system = $%d", argCount))
		args = append(args, *filters.IsSystem)
	}

	if filters.HasModuleFilter() {
		argCount++
		conditions = append(conditions, fmt.Sprintf("p.module = $%d", argCount))
		args = append(args, filters.GetModuleFilter())
	}

	if filters.HasActionFilter() {
		argCount++
		conditions = append(conditions, fmt.Sprintf("p.action = $%d", argCount))
		args = append(args, filters.GetActionFilter())
	}

	if filters.HasResourceFilter() {
		argCount++
		conditions = append(conditions, fmt.Sprintf("p.resource = $%d", argCount))
		args = append(args, filters.GetResourceFilter())
	}

	if filters.HasNameFilter() {
		argCount++
		conditions = append(conditions, fmt.Sprintf("p.name = $%d", argCount))
		args = append(args, filters.GetNameFilter())
	}

	if filters.HasSearchFilter() {
		searchTerm := "%" + strings.ToLower(filters.GetSearchTerm()) + "%"
		argCount++
		conditions = append(conditions, fmt.Sprintf("(LOWER(p.name) LIKE $%d OR LOWER(p.display_name) LIKE $%d OR LOWER(p.description) LIKE $%d)", argCount, argCount, argCount))
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
		repo.logger.Error("Failed to count permissions", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count permissions: %w", err)
	}

	// Query para buscar dados com paginação
	selectQuery := `
		SELECT p.id_permission, p.id_tenant, p.module, p.action, p.resource,
			   p.name, p.display_name, p.description, p.is_system, p.active,
			   p.created_at, p.updated_at, p.created_by, p.updated_by ` + baseQuery

	// Adicionar ordenação
	orderDirection := "ASC"
	if filters.OrderDesc {
		orderDirection = "DESC"
	}
	selectQuery += fmt.Sprintf(" ORDER BY p.%s %s", filters.OrderBy, orderDirection)

	// Adicionar paginação
	selectQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", filters.GetLimit(), filters.GetOffset())

	var rows []permissionRow
	err = repo.db.SelectContext(ctx, &rows, selectQuery, args...)
	if err != nil {
		repo.logger.Error("Failed to list permissions", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list permissions: %w", err)
	}

	// Converter para entidades
	permissions := make([]*permission.Permission, len(rows))
	for i, row := range rows {
		permission, err := row.toEntity()
		if err != nil {
			repo.logger.Error("Failed to convert permission row to entity", zap.Error(err))
			return nil, 0, fmt.Errorf("failed to convert permission: %w", err)
		}
		permissions[i] = permission
	}

	return permissions, total, nil
}

// ListByTenant lista permissions de um tenant específico
func (repo *PermissionRepository) ListByTenant(ctx context.Context, tenantID value_objects.UUID, filters permission.ListFilters) ([]*permission.Permission, int, error) {
	// Definir tenant no filtro
	filters.TenantID = &tenantID
	return repo.List(ctx, filters)
}

// ListSystemPermissions lista todas as permissions do sistema
func (repo *PermissionRepository) ListSystemPermissions(ctx context.Context) ([]*permission.Permission, error) {
	query := `
		SELECT id_permission, id_tenant, module, action, resource,
			   name, display_name, description, is_system, active,
			   created_at, updated_at, created_by, updated_by
		FROM permission 
		WHERE is_system = true
		ORDER BY module ASC, action ASC, resource ASC`

	var rows []permissionRow
	err := repo.db.SelectContext(ctx, &rows, query)
	if err != nil {
		repo.logger.Error("Failed to list system permissions", zap.Error(err))
		return nil, fmt.Errorf("failed to list system permissions: %w", err)
	}

	// Converter para entidades
	permissions := make([]*permission.Permission, len(rows))
	for i, row := range rows {
		permission, err := row.toEntity()
		if err != nil {
			repo.logger.Error("Failed to convert permission row to entity", zap.Error(err))
			return nil, fmt.Errorf("failed to convert permission: %w", err)
		}
		permissions[i] = permission
	}

	return permissions, nil
}

// ExistsByName verifica se existe uma permission com o nome especificado no tenant
func (repo *PermissionRepository) ExistsByName(ctx context.Context, tenantID value_objects.UUID, name string, excludeID *value_objects.UUID) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM permission 
		WHERE id_tenant = $1 AND name = $2`
	args := []interface{}{tenantID.String(), strings.ToUpper(name)}

	if excludeID != nil {
		query += " AND id_permission != $3"
		args = append(args, excludeID.String())
	}

	var count int
	err := repo.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		repo.logger.Error("Failed to check permission existence", zap.Error(err))
		return false, fmt.Errorf("failed to check permission existence: %w", err)
	}

	return count > 0, nil
}

// GetByModule busca permissions por módulo
func (repo *PermissionRepository) GetByModule(ctx context.Context, tenantID value_objects.UUID, module string) ([]*permission.Permission, error) {
	query := `
		SELECT id_permission, id_tenant, module, action, resource,
			   name, display_name, description, is_system, active,
			   created_at, updated_at, created_by, updated_by
		FROM permission 
		WHERE id_tenant = $1 AND module = $2 AND active = true
		ORDER BY action ASC, resource ASC`

	var rows []permissionRow
	err := repo.db.SelectContext(ctx, &rows, query, tenantID.String(), strings.ToLower(module))
	if err != nil {
		repo.logger.Error("Failed to get permissions by module", zap.Error(err))
		return nil, fmt.Errorf("failed to get permissions by module: %w", err)
	}

	// Converter para entidades
	permissions := make([]*permission.Permission, len(rows))
	for i, row := range rows {
		permission, err := row.toEntity()
		if err != nil {
			repo.logger.Error("Failed to convert permission row to entity", zap.Error(err))
			return nil, fmt.Errorf("failed to convert permission: %w", err)
		}
		permissions[i] = permission
	}

	return permissions, nil
}

// GetByModuleAndAction busca permissions por módulo e ação
func (repo *PermissionRepository) GetByModuleAndAction(ctx context.Context, tenantID value_objects.UUID, module, action string) ([]*permission.Permission, error) {
	query := `
		SELECT id_permission, id_tenant, module, action, resource,
			   name, display_name, description, is_system, active,
			   created_at, updated_at, created_by, updated_by
		FROM permission 
		WHERE id_tenant = $1 AND module = $2 AND action = $3 AND active = true
		ORDER BY resource ASC`

	var rows []permissionRow
	err := repo.db.SelectContext(ctx, &rows, query, tenantID.String(), strings.ToLower(module), strings.ToLower(action))
	if err != nil {
		repo.logger.Error("Failed to get permissions by module and action", zap.Error(err))
		return nil, fmt.Errorf("failed to get permissions by module and action: %w", err)
	}

	// Converter para entidades
	permissions := make([]*permission.Permission, len(rows))
	for i, row := range rows {
		permission, err := row.toEntity()
		if err != nil {
			repo.logger.Error("Failed to convert permission row to entity", zap.Error(err))
			return nil, fmt.Errorf("failed to convert permission: %w", err)
		}
		permissions[i] = permission
	}

	return permissions, nil
}

// GetByPattern busca permissions que correspondem a um padrão
func (repo *PermissionRepository) GetByPattern(ctx context.Context, tenantID value_objects.UUID, module, action, resource string) ([]*permission.Permission, error) {
	query := `
		SELECT id_permission, id_tenant, module, action, resource,
			   name, display_name, description, is_system, active,
			   created_at, updated_at, created_by, updated_by
		FROM permission 
		WHERE id_tenant = $1 AND module = $2 AND action = $3 AND active = true
		AND (resource IS NULL OR resource = '' OR resource = $4)
		ORDER BY resource ASC`

	var rows []permissionRow
	err := repo.db.SelectContext(ctx, &rows, query, tenantID.String(), strings.ToLower(module), strings.ToLower(action), strings.ToLower(resource))
	if err != nil {
		repo.logger.Error("Failed to get permissions by pattern", zap.Error(err))
		return nil, fmt.Errorf("failed to get permissions by pattern: %w", err)
	}

	// Converter para entidades
	permissions := make([]*permission.Permission, len(rows))
	for i, row := range rows {
		permission, err := row.toEntity()
		if err != nil {
			repo.logger.Error("Failed to convert permission row to entity", zap.Error(err))
			return nil, fmt.Errorf("failed to convert permission: %w", err)
		}
		permissions[i] = permission
	}

	return permissions, nil
}

// CountByTenant conta o número de permissions de um tenant
func (repo *PermissionRepository) CountByTenant(ctx context.Context, tenantID value_objects.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM permission 
		WHERE id_tenant = $1`

	var count int
	err := repo.db.GetContext(ctx, &count, query, tenantID.String())
	if err != nil {
		repo.logger.Error("Failed to count permissions by tenant", zap.Error(err))
		return 0, fmt.Errorf("failed to count permissions: %w", err)
	}

	return count, nil
}

// GetModules retorna todos os módulos disponíveis para um tenant
func (repo *PermissionRepository) GetModules(ctx context.Context, tenantID value_objects.UUID) ([]string, error) {
	query := `
		SELECT DISTINCT module
		FROM permission 
		WHERE id_tenant = $1 AND active = true
		ORDER BY module ASC`

	var modules []string
	err := repo.db.SelectContext(ctx, &modules, query, tenantID.String())
	if err != nil {
		repo.logger.Error("Failed to get modules", zap.Error(err))
		return nil, fmt.Errorf("failed to get modules: %w", err)
	}

	return modules, nil
}

// GetActions retorna todas as ações disponíveis para um módulo
func (repo *PermissionRepository) GetActions(ctx context.Context, tenantID value_objects.UUID, module string) ([]string, error) {
	query := `
		SELECT DISTINCT action
		FROM permission 
		WHERE id_tenant = $1 AND module = $2 AND active = true
		ORDER BY action ASC`

	var actions []string
	err := repo.db.SelectContext(ctx, &actions, query, tenantID.String(), strings.ToLower(module))
	if err != nil {
		repo.logger.Error("Failed to get actions", zap.Error(err))
		return nil, fmt.Errorf("failed to get actions: %w", err)
	}

	return actions, nil
}
