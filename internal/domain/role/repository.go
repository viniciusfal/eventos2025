package role

import (
	"context"
	"strings"

	"eventos-backend/internal/domain/shared/errors"
	"eventos-backend/internal/domain/shared/value_objects"
)

// Repository define a interface para operações de persistência de roles
type Repository interface {
	// Create cria uma nova role
	Create(ctx context.Context, role *Role) error

	// GetByID busca uma role por ID
	GetByID(ctx context.Context, id value_objects.UUID) (*Role, error)

	// GetByName busca uma role por nome dentro de um tenant
	GetByName(ctx context.Context, tenantID value_objects.UUID, name string) (*Role, error)

	// GetSystemRoleByName busca uma role do sistema por nome
	GetSystemRoleByName(ctx context.Context, name string) (*Role, error)

	// Update atualiza uma role existente
	Update(ctx context.Context, role *Role) error

	// Delete remove uma role (soft delete)
	Delete(ctx context.Context, id value_objects.UUID, deletedBy value_objects.UUID) error

	// List lista roles com filtros e paginação
	List(ctx context.Context, filters ListFilters) ([]*Role, int, error)

	// ListByTenant lista roles de um tenant específico
	ListByTenant(ctx context.Context, tenantID value_objects.UUID, filters ListFilters) ([]*Role, int, error)

	// ListSystemRoles lista todas as roles do sistema
	ListSystemRoles(ctx context.Context) ([]*Role, error)

	// ExistsByName verifica se existe uma role com o nome especificado no tenant
	ExistsByName(ctx context.Context, tenantID value_objects.UUID, name string, excludeID *value_objects.UUID) (bool, error)

	// GetRolesByLevel busca roles por nível hierárquico
	GetRolesByLevel(ctx context.Context, tenantID value_objects.UUID, level int) ([]*Role, error)

	// GetRolesByLevelRange busca roles dentro de um range de níveis
	GetRolesByLevelRange(ctx context.Context, tenantID value_objects.UUID, minLevel, maxLevel int) ([]*Role, error)

	// CountByTenant conta o número de roles de um tenant
	CountByTenant(ctx context.Context, tenantID value_objects.UUID) (int, error)

	// GetHighestLevelRole busca a role de nível mais alto de um tenant
	GetHighestLevelRole(ctx context.Context, tenantID value_objects.UUID) (*Role, error)

	// GetLowestLevelRole busca a role de nível mais baixo de um tenant
	GetLowestLevelRole(ctx context.Context, tenantID value_objects.UUID) (*Role, error)
}

// ListFilters define os filtros para listagem de roles
type ListFilters struct {
	// Filtros básicos
	TenantID *value_objects.UUID
	Active   *bool
	IsSystem *bool

	// Filtros específicos
	Name        *string
	DisplayName *string
	Level       *int
	MinLevel    *int
	MaxLevel    *int

	// Busca textual
	Search *string // Busca em name, display_name e description

	// Paginação
	Page     int
	PageSize int

	// Ordenação
	OrderBy   string // name, display_name, level, created_at, updated_at
	OrderDesc bool
}

// Validate valida os filtros de listagem
func (f *ListFilters) Validate() error {
	// Validar paginação
	if f.Page < 1 {
		f.Page = 1
	}

	if f.PageSize < 1 {
		f.PageSize = 20 // Padrão
	}

	if f.PageSize > 100 {
		f.PageSize = 100 // Máximo
	}

	// Validar ordenação
	validOrderFields := map[string]bool{
		"name":         true,
		"display_name": true,
		"level":        true,
		"created_at":   true,
		"updated_at":   true,
	}

	if f.OrderBy != "" && !validOrderFields[f.OrderBy] {
		f.OrderBy = "level" // Padrão: ordenar por nível
	}

	if f.OrderBy == "" {
		f.OrderBy = "level"
	}

	// Validar níveis
	if f.Level != nil && (*f.Level < 1 || *f.Level > 999) {
		return errors.NewValidationError("Level", "deve estar entre 1 e 999")
	}

	if f.MinLevel != nil && (*f.MinLevel < 1 || *f.MinLevel > 999) {
		return errors.NewValidationError("MinLevel", "deve estar entre 1 e 999")
	}

	if f.MaxLevel != nil && (*f.MaxLevel < 1 || *f.MaxLevel > 999) {
		return errors.NewValidationError("MaxLevel", "deve estar entre 1 e 999")
	}

	if f.MinLevel != nil && f.MaxLevel != nil && *f.MinLevel > *f.MaxLevel {
		return errors.NewValidationError("MinLevel", "deve ser menor ou igual ao máximo")
	}

	return nil
}

// GetOffset calcula o offset para paginação
func (f *ListFilters) GetOffset() int {
	return (f.Page - 1) * f.PageSize
}

// GetLimit retorna o limite para paginação
func (f *ListFilters) GetLimit() int {
	return f.PageSize
}

// HasTenantFilter verifica se há filtro por tenant
func (f *ListFilters) HasTenantFilter() bool {
	return f.TenantID != nil && !f.TenantID.IsZero()
}

// HasActiveFilter verifica se há filtro por status ativo
func (f *ListFilters) HasActiveFilter() bool {
	return f.Active != nil
}

// HasSystemFilter verifica se há filtro por roles do sistema
func (f *ListFilters) HasSystemFilter() bool {
	return f.IsSystem != nil
}

// HasLevelFilter verifica se há filtro por nível
func (f *ListFilters) HasLevelFilter() bool {
	return f.Level != nil
}

// HasLevelRangeFilter verifica se há filtro por range de níveis
func (f *ListFilters) HasLevelRangeFilter() bool {
	return f.MinLevel != nil || f.MaxLevel != nil
}

// HasSearchFilter verifica se há filtro de busca textual
func (f *ListFilters) HasSearchFilter() bool {
	return f.Search != nil && *f.Search != ""
}

// GetSearchTerm retorna o termo de busca limpo
func (f *ListFilters) GetSearchTerm() string {
	if f.Search == nil {
		return ""
	}
	return strings.TrimSpace(*f.Search)
}
