package partner

import (
	"context"

	"eventos-backend/internal/domain/shared/value_objects"
)

// Repository define as operações de persistência para Partner
type Repository interface {
	// Create cria um novo parceiro
	Create(ctx context.Context, partner *Partner) error

	// GetByID busca um parceiro pelo ID
	GetByID(ctx context.Context, id value_objects.UUID) (*Partner, error)

	// GetByIDAndTenant busca um parceiro pelo ID dentro de um tenant
	GetByIDAndTenant(ctx context.Context, id, tenantID value_objects.UUID) (*Partner, error)

	// GetByEmail busca um parceiro pelo email
	GetByEmail(ctx context.Context, email string) (*Partner, error)

	// GetByEmailAndTenant busca um parceiro pelo email dentro de um tenant
	GetByEmailAndTenant(ctx context.Context, email string, tenantID value_objects.UUID) (*Partner, error)

	// GetByIdentity busca um parceiro pela identidade
	GetByIdentity(ctx context.Context, identity string) (*Partner, error)

	// GetByIdentityAndTenant busca um parceiro pela identidade dentro de um tenant
	GetByIdentityAndTenant(ctx context.Context, identity string, tenantID value_objects.UUID) (*Partner, error)

	// Update atualiza um parceiro existente
	Update(ctx context.Context, partner *Partner) error

	// Delete remove um parceiro (soft delete)
	Delete(ctx context.Context, id value_objects.UUID, deletedBy value_objects.UUID) error

	// List lista parceiros com paginação e filtros
	List(ctx context.Context, filters ListFilters) ([]*Partner, int, error)

	// ListByTenant lista parceiros de um tenant específico
	ListByTenant(ctx context.Context, tenantID value_objects.UUID, filters ListFilters) ([]*Partner, int, error)

	// ExistsByEmail verifica se existe um parceiro com o email informado
	ExistsByEmail(ctx context.Context, email string, excludeID *value_objects.UUID) (bool, error)

	// ExistsByEmailInTenant verifica se existe um parceiro com o email no tenant
	ExistsByEmailInTenant(ctx context.Context, email string, tenantID value_objects.UUID, excludeID *value_objects.UUID) (bool, error)

	// ExistsByIdentity verifica se existe um parceiro com a identidade informada
	ExistsByIdentity(ctx context.Context, identity string, excludeID *value_objects.UUID) (bool, error)

	// ExistsByIdentityInTenant verifica se existe um parceiro com a identidade no tenant
	ExistsByIdentityInTenant(ctx context.Context, identity string, tenantID value_objects.UUID, excludeID *value_objects.UUID) (bool, error)

	// ListByEvent lista parceiros associados a um evento específico
	ListByEvent(ctx context.Context, eventID value_objects.UUID, filters ListFilters) ([]*Partner, int, error)

	// GetPartnersWithEmployees busca parceiros que têm funcionários
	GetPartnersWithEmployees(ctx context.Context, tenantID value_objects.UUID, filters ListFilters) ([]*Partner, int, error)
}

// ListFilters define os filtros para listagem de parceiros
type ListFilters struct {
	// Filtros de busca
	TenantID     *value_objects.UUID
	Name         *string
	Email        *string
	Identity     *string
	IdentityType *string
	Location     *string
	Active       *bool
	HasPassword  *bool

	// Filtros de relacionamento
	EventID *value_objects.UUID // Parceiros associados a um evento

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
		f.OrderBy = "name"
	}

	validOrderFields := []string{"name", "email", "identity", "location", "created_at", "updated_at", "last_login"}
	isValidOrder := false
	for _, field := range validOrderFields {
		if f.OrderBy == field {
			isValidOrder = true
			break
		}
	}

	if !isValidOrder {
		f.OrderBy = "name"
	}

	return nil
}

// GetOffset calcula o offset para paginação
func (f *ListFilters) GetOffset() int {
	return (f.Page - 1) * f.PageSize
}
