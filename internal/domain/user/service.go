package user

import (
	"context"

	"eventos-backend/internal/domain/shared/errors"
	"eventos-backend/internal/domain/shared/value_objects"

	"go.uber.org/zap"
)

// Service define os serviços de domínio para User
type Service interface {
	// CreateUser cria um novo usuário com validações de negócio
	CreateUser(ctx context.Context, tenantID value_objects.UUID, fullName, email, phone, username, password string, createdBy value_objects.UUID) (*User, error)

	// UpdateUser atualiza um usuário existente
	UpdateUser(ctx context.Context, id value_objects.UUID, fullName, email, phone, username string, updatedBy value_objects.UUID) (*User, error)

	// UpdateUserPassword atualiza a senha de um usuário
	UpdateUserPassword(ctx context.Context, id value_objects.UUID, newPassword string, updatedBy value_objects.UUID) error

	// GetUser busca um usuário pelo ID
	GetUser(ctx context.Context, id value_objects.UUID) (*User, error)

	// GetUserByUsername busca um usuário pelo username
	GetUserByUsername(ctx context.Context, username string) (*User, error)

	// GetUserByUsernameAndTenant busca um usuário pelo username dentro de um tenant
	GetUserByUsernameAndTenant(ctx context.Context, username string, tenantID value_objects.UUID) (*User, error)

	// AuthenticateUser autentica um usuário com username/email e senha
	AuthenticateUser(ctx context.Context, usernameOrEmail, password string, tenantID *value_objects.UUID) (*User, error)

	// ActivateUser ativa um usuário
	ActivateUser(ctx context.Context, id value_objects.UUID, updatedBy value_objects.UUID) error

	// DeactivateUser desativa um usuário
	DeactivateUser(ctx context.Context, id value_objects.UUID, updatedBy value_objects.UUID) error

	// ListUsers lista usuários com filtros
	ListUsers(ctx context.Context, filters ListFilters) ([]*User, int, error)

	// ListUsersByTenant lista usuários de um tenant específico
	ListUsersByTenant(ctx context.Context, tenantID value_objects.UUID, filters ListFilters) ([]*User, int, error)

	// DeleteUser remove um usuário (soft delete)
	DeleteUser(ctx context.Context, id value_objects.UUID, deletedBy value_objects.UUID) error
}

// DomainService implementa os serviços de domínio para User
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

// CreateUser cria um novo usuário com validações de negócio
func (s *DomainService) CreateUser(ctx context.Context, tenantID value_objects.UUID, fullName, email, phone, username, password string, createdBy value_objects.UUID) (*User, error) {
	s.logger.Debug("Creating new user",
		zap.String("tenant_id", tenantID.String()),
		zap.String("full_name", fullName),
		zap.String("email", email),
		zap.String("username", username),
		zap.String("created_by", createdBy.String()),
	)

	// Verificar se já existe usuário com o mesmo username no tenant
	exists, err := s.repository.ExistsByUsernameInTenant(ctx, username, tenantID, nil)
	if err != nil {
		s.logger.Error("Failed to check username uniqueness in tenant", zap.Error(err))
		return nil, errors.NewInternalError("failed to validate username uniqueness", err)
	}
	if exists {
		return nil, errors.NewAlreadyExistsError("user", "username", username)
	}

	// Verificar se já existe usuário com o mesmo email no tenant (se email fornecido)
	if email != "" {
		exists, err := s.repository.ExistsByEmailInTenant(ctx, email, tenantID, nil)
		if err != nil {
			s.logger.Error("Failed to check email uniqueness in tenant", zap.Error(err))
			return nil, errors.NewInternalError("failed to validate email uniqueness", err)
		}
		if exists {
			return nil, errors.NewAlreadyExistsError("user", "email", email)
		}
	}

	// Criar nova instância do usuário
	user, err := NewUser(tenantID, fullName, email, phone, username, password, createdBy)
	if err != nil {
		s.logger.Error("Failed to create user instance", zap.Error(err))
		return nil, err
	}

	// Persistir no repositório
	if err := s.repository.Create(ctx, user); err != nil {
		s.logger.Error("Failed to persist user", zap.Error(err))
		return nil, errors.NewInternalError("failed to create user", err)
	}

	s.logger.Info("User created successfully",
		zap.String("user_id", user.ID.String()),
		zap.String("username", user.Username),
		zap.String("tenant_id", user.TenantID.String()),
	)

	return user, nil
}

// UpdateUser atualiza um usuário existente
func (s *DomainService) UpdateUser(ctx context.Context, id value_objects.UUID, fullName, email, phone, username string, updatedBy value_objects.UUID) (*User, error) {
	s.logger.Debug("Updating user",
		zap.String("user_id", id.String()),
		zap.String("full_name", fullName),
		zap.String("username", username),
		zap.String("updated_by", updatedBy.String()),
	)

	// Buscar usuário existente
	user, err := s.repository.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get user for update", zap.Error(err))
		return nil, errors.NewInternalError("failed to get user", err)
	}
	if user == nil {
		return nil, errors.NewNotFoundError("user", id.String())
	}

	// Verificar unicidade do username no tenant (se alterado)
	if username != user.Username {
		exists, err := s.repository.ExistsByUsernameInTenant(ctx, username, user.TenantID, &id)
		if err != nil {
			s.logger.Error("Failed to check username uniqueness in tenant", zap.Error(err))
			return nil, errors.NewInternalError("failed to validate username uniqueness", err)
		}
		if exists {
			return nil, errors.NewAlreadyExistsError("user", "username", username)
		}
	}

	// Verificar unicidade do email no tenant (se alterado)
	if email != "" && email != user.Email {
		exists, err := s.repository.ExistsByEmailInTenant(ctx, email, user.TenantID, &id)
		if err != nil {
			s.logger.Error("Failed to check email uniqueness in tenant", zap.Error(err))
			return nil, errors.NewInternalError("failed to validate email uniqueness", err)
		}
		if exists {
			return nil, errors.NewAlreadyExistsError("user", "email", email)
		}
	}

	// Atualizar dados do usuário
	if err := user.Update(fullName, email, phone, username, updatedBy); err != nil {
		s.logger.Error("Failed to update user data", zap.Error(err))
		return nil, err
	}

	// Persistir alterações
	if err := s.repository.Update(ctx, user); err != nil {
		s.logger.Error("Failed to persist user update", zap.Error(err))
		return nil, errors.NewInternalError("failed to update user", err)
	}

	s.logger.Info("User updated successfully",
		zap.String("user_id", user.ID.String()),
	)

	return user, nil
}

// UpdateUserPassword atualiza a senha de um usuário
func (s *DomainService) UpdateUserPassword(ctx context.Context, id value_objects.UUID, newPassword string, updatedBy value_objects.UUID) error {
	s.logger.Debug("Updating user password",
		zap.String("user_id", id.String()),
		zap.String("updated_by", updatedBy.String()),
	)

	// Buscar usuário existente
	user, err := s.repository.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get user for password update", zap.Error(err))
		return errors.NewInternalError("failed to get user", err)
	}
	if user == nil {
		return errors.NewNotFoundError("user", id.String())
	}

	// Atualizar senha
	if err := user.UpdatePassword(newPassword, updatedBy); err != nil {
		s.logger.Error("Failed to update user password", zap.Error(err))
		return err
	}

	// Persistir alterações
	if err := s.repository.Update(ctx, user); err != nil {
		s.logger.Error("Failed to persist user password update", zap.Error(err))
		return errors.NewInternalError("failed to update user password", err)
	}

	s.logger.Info("User password updated successfully",
		zap.String("user_id", id.String()),
	)

	return nil
}

// GetUser busca um usuário pelo ID
func (s *DomainService) GetUser(ctx context.Context, id value_objects.UUID) (*User, error) {
	user, err := s.repository.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get user", zap.Error(err))
		return nil, errors.NewInternalError("failed to get user", err)
	}
	if user == nil {
		return nil, errors.NewNotFoundError("user", id.String())
	}

	return user, nil
}

// GetUserByUsername busca um usuário pelo username
func (s *DomainService) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	user, err := s.repository.GetByUsername(ctx, username)
	if err != nil {
		s.logger.Error("Failed to get user by username", zap.Error(err))
		return nil, errors.NewInternalError("failed to get user", err)
	}
	if user == nil {
		return nil, errors.NewNotFoundError("user", username)
	}

	return user, nil
}

// GetUserByUsernameAndTenant busca um usuário pelo username dentro de um tenant
func (s *DomainService) GetUserByUsernameAndTenant(ctx context.Context, username string, tenantID value_objects.UUID) (*User, error) {
	user, err := s.repository.GetByUsernameAndTenant(ctx, username, tenantID)
	if err != nil {
		s.logger.Error("Failed to get user by username and tenant", zap.Error(err))
		return nil, errors.NewInternalError("failed to get user", err)
	}
	if user == nil {
		return nil, errors.NewNotFoundError("user", username)
	}

	return user, nil
}

// AuthenticateUser autentica um usuário com username/email e senha
func (s *DomainService) AuthenticateUser(ctx context.Context, usernameOrEmail, password string, tenantID *value_objects.UUID) (*User, error) {
	s.logger.Debug("Authenticating user",
		zap.String("username_or_email", usernameOrEmail),
	)

	var user *User
	var err error

	// Tentar buscar por username primeiro
	if tenantID != nil {
		user, err = s.repository.GetByUsernameAndTenant(ctx, usernameOrEmail, *tenantID)
	} else {
		user, err = s.repository.GetByUsername(ctx, usernameOrEmail)
	}

	// Se não encontrou por username, tentar por email
	if err != nil || user == nil {
		if tenantID != nil {
			user, err = s.repository.GetByEmailAndTenant(ctx, usernameOrEmail, *tenantID)
		} else {
			user, err = s.repository.GetByEmail(ctx, usernameOrEmail)
		}
	}

	if err != nil {
		s.logger.Error("Failed to get user for authentication", zap.Error(err))
		return nil, errors.NewInternalError("authentication failed", err)
	}

	if user == nil {
		return nil, errors.NewUnauthorizedError("invalid credentials")
	}

	// Verificar se o usuário está ativo
	if !user.IsActive() {
		return nil, errors.NewUnauthorizedError("user is inactive")
	}

	// Verificar senha
	if !user.CheckPassword(password) {
		return nil, errors.NewUnauthorizedError("invalid credentials")
	}

	s.logger.Info("User authenticated successfully",
		zap.String("user_id", user.ID.String()),
		zap.String("username", user.Username),
	)

	return user, nil
}

// ActivateUser ativa um usuário
func (s *DomainService) ActivateUser(ctx context.Context, id value_objects.UUID, updatedBy value_objects.UUID) error {
	user, err := s.GetUser(ctx, id)
	if err != nil {
		return err
	}

	if user.IsActive() {
		return errors.NewValidationError("status", "user is already active")
	}

	user.Activate(updatedBy)

	if err := s.repository.Update(ctx, user); err != nil {
		s.logger.Error("Failed to activate user", zap.Error(err))
		return errors.NewInternalError("failed to activate user", err)
	}

	s.logger.Info("User activated successfully",
		zap.String("user_id", id.String()),
	)

	return nil
}

// DeactivateUser desativa um usuário
func (s *DomainService) DeactivateUser(ctx context.Context, id value_objects.UUID, updatedBy value_objects.UUID) error {
	user, err := s.GetUser(ctx, id)
	if err != nil {
		return err
	}

	if !user.IsActive() {
		return errors.NewValidationError("status", "user is already inactive")
	}

	user.Deactivate(updatedBy)

	if err := s.repository.Update(ctx, user); err != nil {
		s.logger.Error("Failed to deactivate user", zap.Error(err))
		return errors.NewInternalError("failed to deactivate user", err)
	}

	s.logger.Info("User deactivated successfully",
		zap.String("user_id", id.String()),
	)

	return nil
}

// ListUsers lista usuários com filtros
func (s *DomainService) ListUsers(ctx context.Context, filters ListFilters) ([]*User, int, error) {
	if err := filters.Validate(); err != nil {
		return nil, 0, err
	}

	users, total, err := s.repository.List(ctx, filters)
	if err != nil {
		s.logger.Error("Failed to list users", zap.Error(err))
		return nil, 0, errors.NewInternalError("failed to list users", err)
	}

	return users, total, nil
}

// ListUsersByTenant lista usuários de um tenant específico
func (s *DomainService) ListUsersByTenant(ctx context.Context, tenantID value_objects.UUID, filters ListFilters) ([]*User, int, error) {
	if err := filters.Validate(); err != nil {
		return nil, 0, err
	}

	users, total, err := s.repository.ListByTenant(ctx, tenantID, filters)
	if err != nil {
		s.logger.Error("Failed to list users by tenant", zap.Error(err))
		return nil, 0, errors.NewInternalError("failed to list users", err)
	}

	return users, total, nil
}

// DeleteUser remove um usuário (soft delete)
func (s *DomainService) DeleteUser(ctx context.Context, id value_objects.UUID, deletedBy value_objects.UUID) error {
	// Verificar se o usuário existe
	user, err := s.GetUser(ctx, id)
	if err != nil {
		return err
	}

	// Verificar se o usuário pode ser removido (regras de negócio)
	if user.IsActive() {
		return errors.NewValidationError("status", "cannot delete active user")
	}

	if err := s.repository.Delete(ctx, id, deletedBy); err != nil {
		s.logger.Error("Failed to delete user", zap.Error(err))
		return errors.NewInternalError("failed to delete user", err)
	}

	s.logger.Info("User deleted successfully",
		zap.String("user_id", id.String()),
	)

	return nil
}
