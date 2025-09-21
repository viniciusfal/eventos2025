package permission

import (
	"context"
	"strings"

	"eventos-backend/internal/domain/shared/errors"
	"eventos-backend/internal/domain/shared/value_objects"
)

// Repository define a interface para operações de persistência de permissões
type Repository interface {
	// Create cria uma nova permissão
	Create(ctx context.Context, permission *Permission) error

	// GetByID busca uma permissão por ID
	GetByID(ctx context.Context, id value_objects.UUID) (*Permission, error)

	// GetByName busca uma permissão por nome dentro de um tenant
	GetByName(ctx context.Context, tenantID value_objects.UUID, name string) (*Permission, error)

	// GetSystemPermissionByName busca uma permissão do sistema por nome
	GetSystemPermissionByName(ctx context.Context, name string) (*Permission, error)

	// Update atualiza uma permissão existente
	Update(ctx context.Context, permission *Permission) error

	// Delete remove uma permissão (soft delete)
	Delete(ctx context.Context, id value_objects.UUID, deletedBy value_objects.UUID) error

	// List lista permissões com filtros e paginação
	List(ctx context.Context, filters ListFilters) ([]*Permission, int, error)

	// ListByTenant lista permissões de um tenant específico
	ListByTenant(ctx context.Context, tenantID value_objects.UUID, filters ListFilters) ([]*Permission, int, error)

	// ListSystemPermissions lista todas as permissões do sistema
	ListSystemPermissions(ctx context.Context) ([]*Permission, error)

	// ExistsByName verifica se existe uma permissão com o nome especificado no tenant
	ExistsByName(ctx context.Context, tenantID value_objects.UUID, name string, excludeID *value_objects.UUID) (bool, error)

	// GetByModule busca permissões por módulo
	GetByModule(ctx context.Context, tenantID value_objects.UUID, module string) ([]*Permission, error)

	// GetByModuleAndAction busca permissões por módulo e ação
	GetByModuleAndAction(ctx context.Context, tenantID value_objects.UUID, module, action string) ([]*Permission, error)

	// GetByPattern busca permissões que correspondem a um padrão
	GetByPattern(ctx context.Context, tenantID value_objects.UUID, module, action, resource string) ([]*Permission, error)

	// CountByTenant conta o número de permissões de um tenant
	CountByTenant(ctx context.Context, tenantID value_objects.UUID) (int, error)

	// GetModules retorna todos os módulos disponíveis para um tenant
	GetModules(ctx context.Context, tenantID value_objects.UUID) ([]string, error)

	// GetActions retorna todas as ações disponíveis para um módulo
	GetActions(ctx context.Context, tenantID value_objects.UUID, module string) ([]string, error)
}

// ListFilters define os filtros para listagem de permissões
type ListFilters struct {
	// Filtros básicos
	TenantID *value_objects.UUID
	Active   *bool
	IsSystem *bool

	// Filtros específicos
	Module   *string
	Action   *string
	Resource *string
	Name     *string

	// Busca textual
	Search *string // Busca em name, display_name e description

	// Paginação
	Page     int
	PageSize int

	// Ordenação
	OrderBy   string // name, display_name, module, action, created_at, updated_at
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
		"module":       true,
		"action":       true,
		"created_at":   true,
		"updated_at":   true,
	}

	if f.OrderBy != "" && !validOrderFields[f.OrderBy] {
		f.OrderBy = "name" // Padrão: ordenar por nome
	}

	if f.OrderBy == "" {
		f.OrderBy = "name"
	}

	// Validar módulo
	if f.Module != nil && *f.Module != "" {
		*f.Module = strings.ToLower(strings.TrimSpace(*f.Module))
		if len(*f.Module) < 2 {
			return errors.NewValidationError("Module", "deve ter pelo menos 2 caracteres")
		}
	}

	// Validar ação
	if f.Action != nil && *f.Action != "" {
		*f.Action = strings.ToLower(strings.TrimSpace(*f.Action))
		if len(*f.Action) < 2 {
			return errors.NewValidationError("Action", "deve ter pelo menos 2 caracteres")
		}
	}

	// Validar recurso
	if f.Resource != nil && *f.Resource != "" {
		*f.Resource = strings.ToLower(strings.TrimSpace(*f.Resource))
	}

	// Validar nome
	if f.Name != nil && *f.Name != "" {
		*f.Name = strings.ToUpper(strings.TrimSpace(*f.Name))
		if len(*f.Name) < 2 {
			return errors.NewValidationError("Name", "deve ter pelo menos 2 caracteres")
		}
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

// HasSystemFilter verifica se há filtro por permissões do sistema
func (f *ListFilters) HasSystemFilter() bool {
	return f.IsSystem != nil
}

// HasModuleFilter verifica se há filtro por módulo
func (f *ListFilters) HasModuleFilter() bool {
	return f.Module != nil && *f.Module != ""
}

// HasActionFilter verifica se há filtro por ação
func (f *ListFilters) HasActionFilter() bool {
	return f.Action != nil && *f.Action != ""
}

// HasResourceFilter verifica se há filtro por recurso
func (f *ListFilters) HasResourceFilter() bool {
	return f.Resource != nil && *f.Resource != ""
}

// HasNameFilter verifica se há filtro por nome
func (f *ListFilters) HasNameFilter() bool {
	return f.Name != nil && *f.Name != ""
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

// GetModuleFilter retorna o filtro de módulo
func (f *ListFilters) GetModuleFilter() string {
	if f.Module == nil {
		return ""
	}
	return *f.Module
}

// GetActionFilter retorna o filtro de ação
func (f *ListFilters) GetActionFilter() string {
	if f.Action == nil {
		return ""
	}
	return *f.Action
}

// GetResourceFilter retorna o filtro de recurso
func (f *ListFilters) GetResourceFilter() string {
	if f.Resource == nil {
		return ""
	}
	return *f.Resource
}

// GetNameFilter retorna o filtro de nome
func (f *ListFilters) GetNameFilter() string {
	if f.Name == nil {
		return ""
	}
	return *f.Name
}
