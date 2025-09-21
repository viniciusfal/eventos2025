package tenant

import (
	"context"

	"eventos-backend/internal/domain/shared/value_objects"
)

// Repository define as operações de persistência para Tenant
type Repository interface {
	// Create cria um novo tenant
	Create(ctx context.Context, tenant *Tenant) error

	// GetByID busca um tenant pelo ID
	GetByID(ctx context.Context, id value_objects.UUID) (*Tenant, error)

	// GetByIdentity busca um tenant pela identidade (CPF/CNPJ)
	GetByIdentity(ctx context.Context, identity string) (*Tenant, error)

	// GetByEmail busca um tenant pelo email
	GetByEmail(ctx context.Context, email string) (*Tenant, error)

	// Update atualiza um tenant existente
	Update(ctx context.Context, tenant *Tenant) error

	// Delete remove um tenant (soft delete)
	Delete(ctx context.Context, id value_objects.UUID, deletedBy value_objects.UUID) error

	// List lista tenants com paginação e filtros
	List(ctx context.Context, filters ListFilters) ([]*Tenant, int, error)

	// ExistsByIdentity verifica se existe um tenant com a identidade informada
	ExistsByIdentity(ctx context.Context, identity string, excludeID *value_objects.UUID) (bool, error)

	// ExistsByEmail verifica se existe um tenant com o email informado
	ExistsByEmail(ctx context.Context, email string, excludeID *value_objects.UUID) (bool, error)
}

// ListFilters define os filtros para listagem de tenants
type ListFilters struct {
	// Filtros de busca
	Name         *string
	Identity     *string
	IdentityType *string
	Email        *string
	Active       *bool

	// Paginação
	Page     int
	PageSize int

	// Ordenação
	OrderBy   string
	OrderDesc bool
}

// Validate valida os filtros de listagem
func (f *ListFilters) Validate() error {
	if f.Page < 1 {
		f.Page = 1
	}

	if f.PageSize < 1 {
		f.PageSize = 20
	}

	if f.PageSize > 100 {
		f.PageSize = 100
	}

	if f.OrderBy == "" {
		f.OrderBy = "created_at"
	}

	validOrderFields := []string{"name", "identity", "email", "created_at", "updated_at"}
	isValidOrder := false
	for _, field := range validOrderFields {
		if f.OrderBy == field {
			isValidOrder = true
			break
		}
	}

	if !isValidOrder {
		f.OrderBy = "created_at"
	}

	return nil
}

// GetOffset calcula o offset para paginação
func (f *ListFilters) GetOffset() int {
	return (f.Page - 1) * f.PageSize
}
