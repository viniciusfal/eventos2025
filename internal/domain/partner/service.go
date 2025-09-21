package partner

import (
	"context"

	"eventos-backend/internal/domain/shared/errors"
	"eventos-backend/internal/domain/shared/value_objects"

	"go.uber.org/zap"
)

// Service define os serviços de domínio para Partner
type Service interface {
	// CreatePartner cria um novo parceiro com validações de negócio
	CreatePartner(ctx context.Context, tenantID value_objects.UUID, name, email, email2, phone, phone2, identity, identityType, location, password string, createdBy value_objects.UUID) (*Partner, error)

	// UpdatePartner atualiza um parceiro existente
	UpdatePartner(ctx context.Context, id value_objects.UUID, name, email, email2, phone, phone2, identity, identityType, location string, updatedBy value_objects.UUID) (*Partner, error)

	// UpdatePartnerPassword atualiza a senha de um parceiro
	UpdatePartnerPassword(ctx context.Context, id value_objects.UUID, newPassword string, updatedBy value_objects.UUID) error

	// GetPartner busca um parceiro pelo ID
	GetPartner(ctx context.Context, id value_objects.UUID) (*Partner, error)

	// GetPartnerByTenant busca um parceiro pelo ID dentro de um tenant
	GetPartnerByTenant(ctx context.Context, id, tenantID value_objects.UUID) (*Partner, error)

	// GetPartnerByEmail busca um parceiro pelo email
	GetPartnerByEmail(ctx context.Context, email string) (*Partner, error)

	// GetPartnerByEmailAndTenant busca um parceiro pelo email dentro de um tenant
	GetPartnerByEmailAndTenant(ctx context.Context, email string, tenantID value_objects.UUID) (*Partner, error)

	// AuthenticatePartner autentica um parceiro com email e senha
	AuthenticatePartner(ctx context.Context, emailOrIdentity, password string, tenantID *value_objects.UUID) (*Partner, error)

	// ActivatePartner ativa um parceiro
	ActivatePartner(ctx context.Context, id value_objects.UUID, updatedBy value_objects.UUID) error

	// DeactivatePartner desativa um parceiro
	DeactivatePartner(ctx context.Context, id value_objects.UUID, updatedBy value_objects.UUID) error

	// UnlockPartner desbloqueia um parceiro
	UnlockPartner(ctx context.Context, id value_objects.UUID, updatedBy value_objects.UUID) error

	// ListPartners lista parceiros com filtros
	ListPartners(ctx context.Context, filters ListFilters) ([]*Partner, int, error)

	// ListPartnersByTenant lista parceiros de um tenant específico
	ListPartnersByTenant(ctx context.Context, tenantID value_objects.UUID, filters ListFilters) ([]*Partner, int, error)

	// ListPartnersByEvent lista parceiros associados a um evento
	ListPartnersByEvent(ctx context.Context, eventID value_objects.UUID, filters ListFilters) ([]*Partner, int, error)

	// DeletePartner remove um parceiro (soft delete)
	DeletePartner(ctx context.Context, id value_objects.UUID, deletedBy value_objects.UUID) error
}

// DomainService implementa os serviços de domínio para Partner
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

// CreatePartner cria um novo parceiro com validações de negócio
func (s *DomainService) CreatePartner(ctx context.Context, tenantID value_objects.UUID, name, email, email2, phone, phone2, identity, identityType, location, password string, createdBy value_objects.UUID) (*Partner, error) {
	s.logger.Debug("Creating new partner",
		zap.String("tenant_id", tenantID.String()),
		zap.String("name", name),
		zap.String("email", email),
		zap.String("identity", identity),
		zap.String("created_by", createdBy.String()),
	)

	// Verificar se já existe parceiro com o mesmo email no tenant (se email fornecido)
	if email != "" {
		exists, err := s.repository.ExistsByEmailInTenant(ctx, email, tenantID, nil)
		if err != nil {
			s.logger.Error("Failed to check email uniqueness in tenant", zap.Error(err))
			return nil, errors.NewInternalError("failed to validate email uniqueness", err)
		}
		if exists {
			return nil, errors.NewAlreadyExistsError("partner", "email", email)
		}
	}

	// Verificar se já existe parceiro com a mesma identidade no tenant (se identidade fornecida)
	if identity != "" {
		exists, err := s.repository.ExistsByIdentityInTenant(ctx, identity, tenantID, nil)
		if err != nil {
			s.logger.Error("Failed to check identity uniqueness in tenant", zap.Error(err))
			return nil, errors.NewInternalError("failed to validate identity uniqueness", err)
		}
		if exists {
			return nil, errors.NewAlreadyExistsError("partner", "identity", identity)
		}
	}

	// Criar nova instância do parceiro
	partner, err := NewPartner(tenantID, name, email, email2, phone, phone2, identity, identityType, location, password, createdBy)
	if err != nil {
		s.logger.Error("Failed to create partner instance", zap.Error(err))
		return nil, err
	}

	// Persistir no repositório
	if err := s.repository.Create(ctx, partner); err != nil {
		s.logger.Error("Failed to persist partner", zap.Error(err))
		return nil, errors.NewInternalError("failed to create partner", err)
	}

	s.logger.Info("Partner created successfully",
		zap.String("partner_id", partner.ID.String()),
		zap.String("name", partner.Name),
		zap.String("tenant_id", partner.TenantID.String()),
	)

	return partner, nil
}

// UpdatePartner atualiza um parceiro existente
func (s *DomainService) UpdatePartner(ctx context.Context, id value_objects.UUID, name, email, email2, phone, phone2, identity, identityType, location string, updatedBy value_objects.UUID) (*Partner, error) {
	s.logger.Debug("Updating partner",
		zap.String("partner_id", id.String()),
		zap.String("name", name),
		zap.String("updated_by", updatedBy.String()),
	)

	// Buscar parceiro existente
	partner, err := s.repository.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get partner for update", zap.Error(err))
		return nil, errors.NewInternalError("failed to get partner", err)
	}
	if partner == nil {
		return nil, errors.NewNotFoundError("partner", id.String())
	}

	// Verificar unicidade do email no tenant (se alterado)
	if email != "" && email != partner.Email {
		exists, err := s.repository.ExistsByEmailInTenant(ctx, email, partner.TenantID, &id)
		if err != nil {
			s.logger.Error("Failed to check email uniqueness in tenant", zap.Error(err))
			return nil, errors.NewInternalError("failed to validate email uniqueness", err)
		}
		if exists {
			return nil, errors.NewAlreadyExistsError("partner", "email", email)
		}
	}

	// Verificar unicidade da identidade no tenant (se alterada)
	if identity != "" && identity != partner.Identity {
		exists, err := s.repository.ExistsByIdentityInTenant(ctx, identity, partner.TenantID, &id)
		if err != nil {
			s.logger.Error("Failed to check identity uniqueness in tenant", zap.Error(err))
			return nil, errors.NewInternalError("failed to validate identity uniqueness", err)
		}
		if exists {
			return nil, errors.NewAlreadyExistsError("partner", "identity", identity)
		}
	}

	// Atualizar dados do parceiro
	if err := partner.Update(name, email, email2, phone, phone2, identity, identityType, location, updatedBy); err != nil {
		s.logger.Error("Failed to update partner data", zap.Error(err))
		return nil, err
	}

	// Persistir alterações
	if err := s.repository.Update(ctx, partner); err != nil {
		s.logger.Error("Failed to persist partner update", zap.Error(err))
		return nil, errors.NewInternalError("failed to update partner", err)
	}

	s.logger.Info("Partner updated successfully",
		zap.String("partner_id", partner.ID.String()),
	)

	return partner, nil
}

// UpdatePartnerPassword atualiza a senha de um parceiro
func (s *DomainService) UpdatePartnerPassword(ctx context.Context, id value_objects.UUID, newPassword string, updatedBy value_objects.UUID) error {
	s.logger.Debug("Updating partner password",
		zap.String("partner_id", id.String()),
		zap.String("updated_by", updatedBy.String()),
	)

	// Buscar parceiro existente
	partner, err := s.repository.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get partner for password update", zap.Error(err))
		return errors.NewInternalError("failed to get partner", err)
	}
	if partner == nil {
		return errors.NewNotFoundError("partner", id.String())
	}

	// Atualizar senha
	if err := partner.UpdatePassword(newPassword, updatedBy); err != nil {
		s.logger.Error("Failed to update partner password", zap.Error(err))
		return err
	}

	// Persistir alterações
	if err := s.repository.Update(ctx, partner); err != nil {
		s.logger.Error("Failed to persist partner password update", zap.Error(err))
		return errors.NewInternalError("failed to update partner password", err)
	}

	s.logger.Info("Partner password updated successfully",
		zap.String("partner_id", id.String()),
	)

	return nil
}

// GetPartner busca um parceiro pelo ID
func (s *DomainService) GetPartner(ctx context.Context, id value_objects.UUID) (*Partner, error) {
	partner, err := s.repository.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get partner", zap.Error(err))
		return nil, errors.NewInternalError("failed to get partner", err)
	}
	if partner == nil {
		return nil, errors.NewNotFoundError("partner", id.String())
	}

	return partner, nil
}

// GetPartnerByTenant busca um parceiro pelo ID dentro de um tenant
func (s *DomainService) GetPartnerByTenant(ctx context.Context, id, tenantID value_objects.UUID) (*Partner, error) {
	partner, err := s.repository.GetByIDAndTenant(ctx, id, tenantID)
	if err != nil {
		s.logger.Error("Failed to get partner by tenant", zap.Error(err))
		return nil, errors.NewInternalError("failed to get partner", err)
	}
	if partner == nil {
		return nil, errors.NewNotFoundError("partner", id.String())
	}

	return partner, nil
}

// GetPartnerByEmail busca um parceiro pelo email
func (s *DomainService) GetPartnerByEmail(ctx context.Context, email string) (*Partner, error) {
	partner, err := s.repository.GetByEmail(ctx, email)
	if err != nil {
		s.logger.Error("Failed to get partner by email", zap.Error(err))
		return nil, errors.NewInternalError("failed to get partner", err)
	}
	if partner == nil {
		return nil, errors.NewNotFoundError("partner", email)
	}

	return partner, nil
}

// GetPartnerByEmailAndTenant busca um parceiro pelo email dentro de um tenant
func (s *DomainService) GetPartnerByEmailAndTenant(ctx context.Context, email string, tenantID value_objects.UUID) (*Partner, error) {
	partner, err := s.repository.GetByEmailAndTenant(ctx, email, tenantID)
	if err != nil {
		s.logger.Error("Failed to get partner by email and tenant", zap.Error(err))
		return nil, errors.NewInternalError("failed to get partner", err)
	}
	if partner == nil {
		return nil, errors.NewNotFoundError("partner", email)
	}

	return partner, nil
}

// AuthenticatePartner autentica um parceiro com email e senha
func (s *DomainService) AuthenticatePartner(ctx context.Context, emailOrIdentity, password string, tenantID *value_objects.UUID) (*Partner, error) {
	s.logger.Debug("Authenticating partner",
		zap.String("email_or_identity", emailOrIdentity),
	)

	var partner *Partner
	var err error

	// Tentar buscar por email primeiro
	if tenantID != nil {
		partner, err = s.repository.GetByEmailAndTenant(ctx, emailOrIdentity, *tenantID)
	} else {
		partner, err = s.repository.GetByEmail(ctx, emailOrIdentity)
	}

	// Se não encontrou por email, tentar por identidade
	if err != nil || partner == nil {
		if tenantID != nil {
			partner, err = s.repository.GetByIdentityAndTenant(ctx, emailOrIdentity, *tenantID)
		} else {
			partner, err = s.repository.GetByIdentity(ctx, emailOrIdentity)
		}
	}

	if err != nil {
		s.logger.Error("Failed to get partner for authentication", zap.Error(err))
		return nil, errors.NewInternalError("authentication failed", err)
	}

	if partner == nil {
		return nil, errors.NewUnauthorizedError("invalid credentials")
	}

	// Verificar se o parceiro está ativo
	if !partner.IsActive() {
		return nil, errors.NewUnauthorizedError("partner is inactive")
	}

	// Verificar se o parceiro está bloqueado
	if partner.IsLocked() {
		return nil, errors.NewUnauthorizedError("partner account is locked")
	}

	// Verificar se o parceiro tem senha definida
	if !partner.HasPassword() {
		return nil, errors.NewUnauthorizedError("partner has no password set")
	}

	// Verificar senha
	if !partner.CheckPassword(password) {
		// Registrar tentativa falhada
		partner.RecordFailedLogin()
		if err := s.repository.Update(ctx, partner); err != nil {
			s.logger.Error("Failed to record failed login", zap.Error(err))
		}
		return nil, errors.NewUnauthorizedError("invalid credentials")
	}

	// Registrar login bem-sucedido
	partner.RecordSuccessfulLogin()
	if err := s.repository.Update(ctx, partner); err != nil {
		s.logger.Error("Failed to record successful login", zap.Error(err))
		// Não falhar a autenticação por causa disso
	}

	s.logger.Info("Partner authenticated successfully",
		zap.String("partner_id", partner.ID.String()),
		zap.String("name", partner.Name),
	)

	return partner, nil
}

// ActivatePartner ativa um parceiro
func (s *DomainService) ActivatePartner(ctx context.Context, id value_objects.UUID, updatedBy value_objects.UUID) error {
	partner, err := s.GetPartner(ctx, id)
	if err != nil {
		return err
	}

	if partner.IsActive() {
		return errors.NewValidationError("status", "partner is already active")
	}

	partner.Activate(updatedBy)

	if err := s.repository.Update(ctx, partner); err != nil {
		s.logger.Error("Failed to activate partner", zap.Error(err))
		return errors.NewInternalError("failed to activate partner", err)
	}

	s.logger.Info("Partner activated successfully",
		zap.String("partner_id", id.String()),
	)

	return nil
}

// DeactivatePartner desativa um parceiro
func (s *DomainService) DeactivatePartner(ctx context.Context, id value_objects.UUID, updatedBy value_objects.UUID) error {
	partner, err := s.GetPartner(ctx, id)
	if err != nil {
		return err
	}

	if !partner.IsActive() {
		return errors.NewValidationError("status", "partner is already inactive")
	}

	partner.Deactivate(updatedBy)

	if err := s.repository.Update(ctx, partner); err != nil {
		s.logger.Error("Failed to deactivate partner", zap.Error(err))
		return errors.NewInternalError("failed to deactivate partner", err)
	}

	s.logger.Info("Partner deactivated successfully",
		zap.String("partner_id", id.String()),
	)

	return nil
}

// UnlockPartner desbloqueia um parceiro
func (s *DomainService) UnlockPartner(ctx context.Context, id value_objects.UUID, updatedBy value_objects.UUID) error {
	partner, err := s.GetPartner(ctx, id)
	if err != nil {
		return err
	}

	if !partner.IsLocked() {
		return errors.NewValidationError("status", "partner is not locked")
	}

	partner.UnlockAccount(updatedBy)

	if err := s.repository.Update(ctx, partner); err != nil {
		s.logger.Error("Failed to unlock partner", zap.Error(err))
		return errors.NewInternalError("failed to unlock partner", err)
	}

	s.logger.Info("Partner unlocked successfully",
		zap.String("partner_id", id.String()),
	)

	return nil
}

// ListPartners lista parceiros com filtros
func (s *DomainService) ListPartners(ctx context.Context, filters ListFilters) ([]*Partner, int, error) {
	if err := filters.Validate(); err != nil {
		return nil, 0, err
	}

	partners, total, err := s.repository.List(ctx, filters)
	if err != nil {
		s.logger.Error("Failed to list partners", zap.Error(err))
		return nil, 0, errors.NewInternalError("failed to list partners", err)
	}

	return partners, total, nil
}

// ListPartnersByTenant lista parceiros de um tenant específico
func (s *DomainService) ListPartnersByTenant(ctx context.Context, tenantID value_objects.UUID, filters ListFilters) ([]*Partner, int, error) {
	if err := filters.Validate(); err != nil {
		return nil, 0, err
	}

	partners, total, err := s.repository.ListByTenant(ctx, tenantID, filters)
	if err != nil {
		s.logger.Error("Failed to list partners by tenant", zap.Error(err))
		return nil, 0, errors.NewInternalError("failed to list partners", err)
	}

	return partners, total, nil
}

// ListPartnersByEvent lista parceiros associados a um evento
func (s *DomainService) ListPartnersByEvent(ctx context.Context, eventID value_objects.UUID, filters ListFilters) ([]*Partner, int, error) {
	if err := filters.Validate(); err != nil {
		return nil, 0, err
	}

	partners, total, err := s.repository.ListByEvent(ctx, eventID, filters)
	if err != nil {
		s.logger.Error("Failed to list partners by event", zap.Error(err))
		return nil, 0, errors.NewInternalError("failed to list partners by event", err)
	}

	return partners, total, nil
}

// DeletePartner remove um parceiro (soft delete)
func (s *DomainService) DeletePartner(ctx context.Context, id value_objects.UUID, deletedBy value_objects.UUID) error {
	// Verificar se o parceiro existe
	partner, err := s.GetPartner(ctx, id)
	if err != nil {
		return err
	}

	// Verificar se o parceiro pode ser removido (regras de negócio)
	if partner.IsActive() {
		return errors.NewValidationError("status", "cannot delete active partner")
	}

	if err := s.repository.Delete(ctx, id, deletedBy); err != nil {
		s.logger.Error("Failed to delete partner", zap.Error(err))
		return errors.NewInternalError("failed to delete partner", err)
	}

	s.logger.Info("Partner deleted successfully",
		zap.String("partner_id", id.String()),
	)

	return nil
}
