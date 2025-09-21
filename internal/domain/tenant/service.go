package tenant

import (
	"context"

	"eventos-backend/internal/domain/shared/errors"
	"eventos-backend/internal/domain/shared/value_objects"

	"go.uber.org/zap"
)

// Service define os serviços de domínio para Tenant
type Service interface {
	// CreateTenant cria um novo tenant com validações de negócio
	CreateTenant(ctx context.Context, name, identity, identityType, email, address string, createdBy value_objects.UUID) (*Tenant, error)

	// UpdateTenant atualiza um tenant existente
	UpdateTenant(ctx context.Context, id value_objects.UUID, name, identity, identityType, email, address string, updatedBy value_objects.UUID) (*Tenant, error)

	// GetTenant busca um tenant pelo ID
	GetTenant(ctx context.Context, id value_objects.UUID) (*Tenant, error)

	// ActivateTenant ativa um tenant
	ActivateTenant(ctx context.Context, id value_objects.UUID, updatedBy value_objects.UUID) error

	// DeactivateTenant desativa um tenant
	DeactivateTenant(ctx context.Context, id value_objects.UUID, updatedBy value_objects.UUID) error

	// ListTenants lista tenants com filtros
	ListTenants(ctx context.Context, filters ListFilters) ([]*Tenant, int, error)

	// DeleteTenant remove um tenant (soft delete)
	DeleteTenant(ctx context.Context, id value_objects.UUID, deletedBy value_objects.UUID) error
}

// DomainService implementa os serviços de domínio para Tenant
type DomainService struct {
	repository Repository
	logger     *zap.Logger
}

// NewDomainService cria uma nova instância do serviço de domínio
func NewDomainService(repository Repository, logger *zap.Logger) Service {
	return &DomainService{
		repository: repository,
		logger:     logger,
	}
}

// CreateTenant cria um novo tenant com validações de negócio
func (s *DomainService) CreateTenant(ctx context.Context, name, identity, identityType, email, address string, createdBy value_objects.UUID) (*Tenant, error) {
	s.logger.Debug("Creating new tenant",
		zap.String("name", name),
		zap.String("identity", identity),
		zap.String("identity_type", identityType),
		zap.String("email", email),
		zap.String("created_by", createdBy.String()),
	)

	// Verificar se já existe tenant com a mesma identidade
	if identity != "" {
		exists, err := s.repository.ExistsByIdentity(ctx, identity, nil)
		if err != nil {
			s.logger.Error("Failed to check identity uniqueness", zap.Error(err))
			return nil, errors.NewInternalError("failed to validate identity uniqueness", err)
		}
		if exists {
			return nil, errors.NewAlreadyExistsError("tenant", "identity", identity)
		}
	}

	// Verificar se já existe tenant com o mesmo email
	if email != "" {
		exists, err := s.repository.ExistsByEmail(ctx, email, nil)
		if err != nil {
			s.logger.Error("Failed to check email uniqueness", zap.Error(err))
			return nil, errors.NewInternalError("failed to validate email uniqueness", err)
		}
		if exists {
			return nil, errors.NewAlreadyExistsError("tenant", "email", email)
		}
	}

	// Criar nova instância do tenant
	tenant, err := NewTenant(name, identity, identityType, email, address, createdBy)
	if err != nil {
		s.logger.Error("Failed to create tenant instance", zap.Error(err))
		return nil, err
	}

	// Persistir no repositório
	if err := s.repository.Create(ctx, tenant); err != nil {
		s.logger.Error("Failed to persist tenant", zap.Error(err))
		return nil, errors.NewInternalError("failed to create tenant", err)
	}

	s.logger.Info("Tenant created successfully",
		zap.String("tenant_id", tenant.ID.String()),
		zap.String("name", tenant.Name),
	)

	return tenant, nil
}

// UpdateTenant atualiza um tenant existente
func (s *DomainService) UpdateTenant(ctx context.Context, id value_objects.UUID, name, identity, identityType, email, address string, updatedBy value_objects.UUID) (*Tenant, error) {
	s.logger.Debug("Updating tenant",
		zap.String("tenant_id", id.String()),
		zap.String("name", name),
		zap.String("updated_by", updatedBy.String()),
	)

	// Buscar tenant existente
	tenant, err := s.repository.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get tenant for update", zap.Error(err))
		return nil, errors.NewInternalError("failed to get tenant", err)
	}
	if tenant == nil {
		return nil, errors.NewNotFoundError("tenant", id.String())
	}

	// Verificar unicidade da identidade (se alterada)
	if identity != "" && identity != tenant.Identity {
		exists, err := s.repository.ExistsByIdentity(ctx, identity, &id)
		if err != nil {
			s.logger.Error("Failed to check identity uniqueness", zap.Error(err))
			return nil, errors.NewInternalError("failed to validate identity uniqueness", err)
		}
		if exists {
			return nil, errors.NewAlreadyExistsError("tenant", "identity", identity)
		}
	}

	// Verificar unicidade do email (se alterado)
	if email != "" && email != tenant.Email {
		exists, err := s.repository.ExistsByEmail(ctx, email, &id)
		if err != nil {
			s.logger.Error("Failed to check email uniqueness", zap.Error(err))
			return nil, errors.NewInternalError("failed to validate email uniqueness", err)
		}
		if exists {
			return nil, errors.NewAlreadyExistsError("tenant", "email", email)
		}
	}

	// Atualizar dados do tenant
	if err := tenant.Update(name, identity, identityType, email, address, updatedBy); err != nil {
		s.logger.Error("Failed to update tenant data", zap.Error(err))
		return nil, err
	}

	// Persistir alterações
	if err := s.repository.Update(ctx, tenant); err != nil {
		s.logger.Error("Failed to persist tenant update", zap.Error(err))
		return nil, errors.NewInternalError("failed to update tenant", err)
	}

	s.logger.Info("Tenant updated successfully",
		zap.String("tenant_id", tenant.ID.String()),
	)

	return tenant, nil
}

// GetTenant busca um tenant pelo ID
func (s *DomainService) GetTenant(ctx context.Context, id value_objects.UUID) (*Tenant, error) {
	tenant, err := s.repository.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get tenant", zap.Error(err))
		return nil, errors.NewInternalError("failed to get tenant", err)
	}
	if tenant == nil {
		return nil, errors.NewNotFoundError("tenant", id.String())
	}

	return tenant, nil
}

// ActivateTenant ativa um tenant
func (s *DomainService) ActivateTenant(ctx context.Context, id value_objects.UUID, updatedBy value_objects.UUID) error {
	tenant, err := s.GetTenant(ctx, id)
	if err != nil {
		return err
	}

	if tenant.IsActive() {
		return errors.NewValidationError("status", "tenant is already active")
	}

	tenant.Activate(updatedBy)

	if err := s.repository.Update(ctx, tenant); err != nil {
		s.logger.Error("Failed to activate tenant", zap.Error(err))
		return errors.NewInternalError("failed to activate tenant", err)
	}

	s.logger.Info("Tenant activated successfully",
		zap.String("tenant_id", id.String()),
	)

	return nil
}

// DeactivateTenant desativa um tenant
func (s *DomainService) DeactivateTenant(ctx context.Context, id value_objects.UUID, updatedBy value_objects.UUID) error {
	tenant, err := s.GetTenant(ctx, id)
	if err != nil {
		return err
	}

	if !tenant.IsActive() {
		return errors.NewValidationError("status", "tenant is already inactive")
	}

	tenant.Deactivate(updatedBy)

	if err := s.repository.Update(ctx, tenant); err != nil {
		s.logger.Error("Failed to deactivate tenant", zap.Error(err))
		return errors.NewInternalError("failed to deactivate tenant", err)
	}

	s.logger.Info("Tenant deactivated successfully",
		zap.String("tenant_id", id.String()),
	)

	return nil
}

// ListTenants lista tenants com filtros
func (s *DomainService) ListTenants(ctx context.Context, filters ListFilters) ([]*Tenant, int, error) {
	if err := filters.Validate(); err != nil {
		return nil, 0, err
	}

	tenants, total, err := s.repository.List(ctx, filters)
	if err != nil {
		s.logger.Error("Failed to list tenants", zap.Error(err))
		return nil, 0, errors.NewInternalError("failed to list tenants", err)
	}

	return tenants, total, nil
}

// DeleteTenant remove um tenant (soft delete)
func (s *DomainService) DeleteTenant(ctx context.Context, id value_objects.UUID, deletedBy value_objects.UUID) error {
	// Verificar se o tenant existe
	tenant, err := s.GetTenant(ctx, id)
	if err != nil {
		return err
	}

	// Verificar se o tenant pode ser removido (regras de negócio)
	if tenant.IsActive() {
		return errors.NewValidationError("status", "cannot delete active tenant")
	}

	if err := s.repository.Delete(ctx, id, deletedBy); err != nil {
		s.logger.Error("Failed to delete tenant", zap.Error(err))
		return errors.NewInternalError("failed to delete tenant", err)
	}

	s.logger.Info("Tenant deleted successfully",
		zap.String("tenant_id", id.String()),
	)

	return nil
}
