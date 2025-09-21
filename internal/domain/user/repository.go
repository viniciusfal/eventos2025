package user

import (
	"context"

	"eventos-backend/internal/domain/shared/value_objects"
)

// Repository define as operações de persistência para User
type Repository interface {
	// Create cria um novo usuário
	Create(ctx context.Context, user *User) error

	// GetByID busca um usuário pelo ID
	GetByID(ctx context.Context, id value_objects.UUID) (*User, error)

	// GetByUsername busca um usuário pelo username
	GetByUsername(ctx context.Context, username string) (*User, error)

	// GetByEmail busca um usuário pelo email
	GetByEmail(ctx context.Context, email string) (*User, error)

	// GetByUsernameAndTenant busca um usuário pelo username dentro de um tenant
	GetByUsernameAndTenant(ctx context.Context, username string, tenantID value_objects.UUID) (*User, error)

	// GetByEmailAndTenant busca um usuário pelo email dentro de um tenant
	GetByEmailAndTenant(ctx context.Context, email string, tenantID value_objects.UUID) (*User, error)

	// Update atualiza um usuário existente
	Update(ctx context.Context, user *User) error

	// Delete remove um usuário (soft delete)
	Delete(ctx context.Context, id value_objects.UUID, deletedBy value_objects.UUID) error

	// List lista usuários com paginação e filtros
	List(ctx context.Context, filters ListFilters) ([]*User, int, error)

	// ExistsByUsername verifica se existe um usuário com o username informado
	ExistsByUsername(ctx context.Context, username string, excludeID *value_objects.UUID) (bool, error)

	// ExistsByEmail verifica se existe um usuário com o email informado
	ExistsByEmail(ctx context.Context, email string, excludeID *value_objects.UUID) (bool, error)

	// ExistsByUsernameInTenant verifica se existe um usuário com o username no tenant
	ExistsByUsernameInTenant(ctx context.Context, username string, tenantID value_objects.UUID, excludeID *value_objects.UUID) (bool, error)

	// ExistsByEmailInTenant verifica se existe um usuário com o email no tenant
	ExistsByEmailInTenant(ctx context.Context, email string, tenantID value_objects.UUID, excludeID *value_objects.UUID) (bool, error)

	// ListByTenant lista usuários de um tenant específico
	ListByTenant(ctx context.Context, tenantID value_objects.UUID, filters ListFilters) ([]*User, int, error)
}

// ListFilters define os filtros para listagem de usuários
type ListFilters struct {
	// Filtros de busca
	TenantID *value_objects.UUID
	FullName *string
	Email    *string
	Username *string
	Active   *bool

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

	validOrderFields := []string{"full_name", "email", "username", "created_at", "updated_at"}
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
