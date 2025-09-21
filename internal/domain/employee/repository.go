package employee

import (
	"context"

	"eventos-backend/internal/domain/shared/value_objects"
)

// Repository define as operações de persistência para Employee
type Repository interface {
	// Create cria um novo funcionário
	Create(ctx context.Context, employee *Employee) error

	// GetByID busca um funcionário pelo ID
	GetByID(ctx context.Context, id value_objects.UUID) (*Employee, error)

	// GetByIDAndTenant busca um funcionário pelo ID dentro de um tenant
	GetByIDAndTenant(ctx context.Context, id, tenantID value_objects.UUID) (*Employee, error)

	// GetByIdentity busca um funcionário pela identidade
	GetByIdentity(ctx context.Context, identity string) (*Employee, error)

	// GetByIdentityAndTenant busca um funcionário pela identidade dentro de um tenant
	GetByIdentityAndTenant(ctx context.Context, identity string, tenantID value_objects.UUID) (*Employee, error)

	// GetByEmail busca um funcionário pelo email
	GetByEmail(ctx context.Context, email string) (*Employee, error)

	// GetByEmailAndTenant busca um funcionário pelo email dentro de um tenant
	GetByEmailAndTenant(ctx context.Context, email string, tenantID value_objects.UUID) (*Employee, error)

	// Update atualiza um funcionário existente
	Update(ctx context.Context, employee *Employee) error

	// Delete remove um funcionário (soft delete)
	Delete(ctx context.Context, id value_objects.UUID, deletedBy value_objects.UUID) error

	// List lista funcionários com paginação e filtros
	List(ctx context.Context, filters ListFilters) ([]*Employee, int, error)

	// ListByTenant lista funcionários de um tenant específico
	ListByTenant(ctx context.Context, tenantID value_objects.UUID, filters ListFilters) ([]*Employee, int, error)

	// ListByPartner lista funcionários de um parceiro específico
	ListByPartner(ctx context.Context, partnerID value_objects.UUID, filters ListFilters) ([]*Employee, int, error)

	// ListByEvent lista funcionários associados a um evento (através de parceiros)
	ListByEvent(ctx context.Context, eventID value_objects.UUID, filters ListFilters) ([]*Employee, int, error)

	// ExistsByIdentity verifica se existe um funcionário com a identidade informada
	ExistsByIdentity(ctx context.Context, identity string, excludeID *value_objects.UUID) (bool, error)

	// ExistsByIdentityInTenant verifica se existe um funcionário com a identidade no tenant
	ExistsByIdentityInTenant(ctx context.Context, identity string, tenantID value_objects.UUID, excludeID *value_objects.UUID) (bool, error)

	// ExistsByEmail verifica se existe um funcionário com o email informado
	ExistsByEmail(ctx context.Context, email string, excludeID *value_objects.UUID) (bool, error)

	// ExistsByEmailInTenant verifica se existe um funcionário com o email no tenant
	ExistsByEmailInTenant(ctx context.Context, email string, tenantID value_objects.UUID, excludeID *value_objects.UUID) (bool, error)

	// FindByFaceEmbedding busca funcionários similares por embedding facial
	FindByFaceEmbedding(ctx context.Context, embedding []float32, tenantID *value_objects.UUID, threshold float32, limit int) ([]*Employee, []float32, error)

	// GetEmployeesWithFaceEmbedding busca funcionários que têm embedding facial
	GetEmployeesWithFaceEmbedding(ctx context.Context, tenantID *value_objects.UUID, filters ListFilters) ([]*Employee, int, error)
}

// ListFilters define os filtros para listagem de funcionários
type ListFilters struct {
	// Filtros de busca
	TenantID     *value_objects.UUID
	PartnerID    *value_objects.UUID
	EventID      *value_objects.UUID
	FullName     *string
	Identity     *string
	IdentityType *string
	Email        *string
	Phone        *string
	Active       *bool

	// Filtros específicos
	HasPhoto         *bool
	HasFaceEmbedding *bool
	IsMinor          *bool
	MinAge           *int
	MaxAge           *int

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
		f.OrderBy = "full_name"
	}

	validOrderFields := []string{"full_name", "identity", "email", "phone", "date_of_birth", "created_at", "updated_at"}
	isValidOrder := false
	for _, field := range validOrderFields {
		if f.OrderBy == field {
			isValidOrder = true
			break
		}
	}

	if !isValidOrder {
		f.OrderBy = "full_name"
	}

	// Validar filtros de idade
	if f.MinAge != nil && *f.MinAge < 0 {
		*f.MinAge = 0
	}

	if f.MaxAge != nil && *f.MaxAge > 120 {
		*f.MaxAge = 120
	}

	if f.MinAge != nil && f.MaxAge != nil && *f.MinAge > *f.MaxAge {
		*f.MinAge = *f.MaxAge
	}

	return nil
}

// GetOffset calcula o offset para paginação
func (f *ListFilters) GetOffset() int {
	return (f.Page - 1) * f.PageSize
}

// FaceRecognitionResult representa o resultado de uma busca por reconhecimento facial
type FaceRecognitionResult struct {
	Employee   *Employee
	Similarity float32
	Confidence string // "high", "medium", "low"
}

// GetConfidenceLevel retorna o nível de confiança baseado na similaridade
func (r *FaceRecognitionResult) GetConfidenceLevel() string {
	if r.Similarity >= 0.9 {
		return "high"
	} else if r.Similarity >= 0.75 {
		return "medium"
	}
	return "low"
}
