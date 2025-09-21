package role

import (
	"context"
	"time"

	"eventos-backend/internal/domain/shared/value_objects"
)

// RolePermission representa o relacionamento Many-to-Many entre Role e Permission
type RolePermission struct {
	ID           value_objects.UUID
	RoleID       value_objects.UUID
	PermissionID value_objects.UUID
	TenantID     value_objects.UUID
	GrantedBy    value_objects.UUID
	GrantedAt    time.Time
	Active       bool
}

// NewRolePermission cria um novo relacionamento role-permission
func NewRolePermission(roleID, permissionID, tenantID, grantedBy value_objects.UUID) *RolePermission {
	return &RolePermission{
		ID:           value_objects.NewUUID(),
		RoleID:       roleID,
		PermissionID: permissionID,
		TenantID:     tenantID,
		GrantedBy:    grantedBy,
		GrantedAt:    time.Now(),
		Active:       true,
	}
}

// RolePermissionRepository define operações para o relacionamento role-permission
type RolePermissionRepository interface {
	// GrantPermission concede uma permissão a uma role
	GrantPermission(ctx context.Context, rolePermission *RolePermission) error

	// RevokePermission revoga uma permissão de uma role
	RevokePermission(ctx context.Context, roleID, permissionID value_objects.UUID) error

	// GetRolePermissions busca todas as permissões de uma role
	GetRolePermissions(ctx context.Context, roleID value_objects.UUID) ([]*RolePermission, error)

	// GetPermissionRoles busca todas as roles que têm uma permissão
	GetPermissionRoles(ctx context.Context, permissionID value_objects.UUID) ([]*RolePermission, error)

	// HasPermission verifica se uma role tem uma permissão específica
	HasPermission(ctx context.Context, roleID, permissionID value_objects.UUID) (bool, error)

	// ListRolePermissions lista relacionamentos com filtros
	ListRolePermissions(ctx context.Context, filters RolePermissionFilters) ([]*RolePermission, int, error)

	// BulkGrantPermissions concede múltiplas permissões a uma role
	BulkGrantPermissions(ctx context.Context, rolePermissions []*RolePermission) error

	// BulkRevokePermissions revoga múltiplas permissões de uma role
	BulkRevokePermissions(ctx context.Context, roleID value_objects.UUID, permissionIDs []value_objects.UUID) error

	// GetTenantRolePermissions busca relacionamentos por tenant
	GetTenantRolePermissions(ctx context.Context, tenantID value_objects.UUID) ([]*RolePermission, error)
}

// RolePermissionFilters define filtros para busca de relacionamentos
type RolePermissionFilters struct {
	TenantID     *value_objects.UUID
	RoleID       *value_objects.UUID
	PermissionID *value_objects.UUID
	Active       *bool
	GrantedBy    *value_objects.UUID

	// Paginação
	Page     int
	PageSize int

	// Ordenação
	OrderBy   string // granted_at, role_id, permission_id
	OrderDesc bool
}

// Validate valida os filtros
func (f *RolePermissionFilters) Validate() error {
	if f.Page < 1 {
		f.Page = 1
	}

	if f.PageSize < 1 {
		f.PageSize = 20
	}

	if f.PageSize > 100 {
		f.PageSize = 100
	}

	validOrderFields := map[string]bool{
		"granted_at":    true,
		"role_id":       true,
		"permission_id": true,
	}

	if f.OrderBy != "" && !validOrderFields[f.OrderBy] {
		f.OrderBy = "granted_at"
	}

	if f.OrderBy == "" {
		f.OrderBy = "granted_at"
	}

	return nil
}

// GetOffset calcula o offset para paginação
func (f *RolePermissionFilters) GetOffset() int {
	return (f.Page - 1) * f.PageSize
}

// GetLimit retorna o limite para paginação
func (f *RolePermissionFilters) GetLimit() int {
	return f.PageSize
}
