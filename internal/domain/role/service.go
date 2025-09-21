package role

import (
	"context"
	"fmt"

	"eventos-backend/internal/domain/shared/errors"
	"eventos-backend/internal/domain/shared/value_objects"
)

// Service define a interface para serviços de domínio de roles
type Service interface {
	// CreateRole cria uma nova role
	CreateRole(ctx context.Context, tenantID value_objects.UUID, name, displayName, description string, level int, createdBy value_objects.UUID) (*Role, error)

	// UpdateRole atualiza uma role existente
	UpdateRole(ctx context.Context, id value_objects.UUID, displayName, description string, level int, updatedBy value_objects.UUID) (*Role, error)

	// GetRole busca uma role por ID
	GetRole(ctx context.Context, id value_objects.UUID) (*Role, error)

	// GetRoleByName busca uma role por nome
	GetRoleByName(ctx context.Context, tenantID value_objects.UUID, name string) (*Role, error)

	// DeleteRole remove uma role
	DeleteRole(ctx context.Context, id value_objects.UUID, deletedBy value_objects.UUID) error

	// ActivateRole ativa uma role
	ActivateRole(ctx context.Context, id value_objects.UUID, updatedBy value_objects.UUID) error

	// DeactivateRole desativa uma role
	DeactivateRole(ctx context.Context, id value_objects.UUID, updatedBy value_objects.UUID) error

	// ListRoles lista roles com filtros
	ListRoles(ctx context.Context, filters ListFilters) ([]*Role, int, error)

	// ListTenantRoles lista roles de um tenant
	ListTenantRoles(ctx context.Context, tenantID value_objects.UUID, filters ListFilters) ([]*Role, int, error)

	// ListSystemRoles lista roles do sistema
	ListSystemRoles(ctx context.Context) ([]*Role, error)

	// InitializeSystemRoles inicializa as roles padrão do sistema
	InitializeSystemRoles(ctx context.Context) error

	// ValidateRoleHierarchy valida se uma role pode gerenciar outra
	ValidateRoleHierarchy(ctx context.Context, managerRoleID, targetRoleID value_objects.UUID) error

	// GetAvailableLevels retorna os níveis disponíveis para um tenant
	GetAvailableLevels(ctx context.Context, tenantID value_objects.UUID) ([]int, error)

	// SuggestLevel sugere um nível para uma nova role
	SuggestLevel(ctx context.Context, tenantID value_objects.UUID) (int, error)
}

// serviceImpl implementa a interface Service
type serviceImpl struct {
	repo Repository
}

// NewService cria uma nova instância do serviço
func NewService(repo Repository) Service {
	return &serviceImpl{
		repo: repo,
	}
}

// CreateRole cria uma nova role
func (s *serviceImpl) CreateRole(ctx context.Context, tenantID value_objects.UUID, name, displayName, description string, level int, createdBy value_objects.UUID) (*Role, error) {
	// Verificar se já existe uma role com este nome no tenant
	exists, err := s.repo.ExistsByName(ctx, tenantID, name, nil)
	if err != nil {
		return nil, errors.NewInternalError("Erro ao verificar existência da role", err)
	}

	if exists {
		return nil, errors.NewAlreadyExistsError("Role", "name", name)
	}

	// Verificar se o nível está disponível
	existingRoles, err := s.repo.GetRolesByLevel(ctx, tenantID, level)
	if err != nil {
		return nil, errors.NewInternalError("Erro ao verificar nível da role", err)
	}

	if len(existingRoles) > 0 {
		return nil, errors.NewValidationError("Level", fmt.Sprintf("já existe uma role com nível %d neste tenant", level))
	}

	// Criar a role
	role, err := NewRole(tenantID, name, displayName, description, level, createdBy)
	if err != nil {
		return nil, err
	}

	// Salvar no repositório
	if err := s.repo.Create(ctx, role); err != nil {
		return nil, errors.NewInternalError("Erro ao criar role", err)
	}

	return role, nil
}

// UpdateRole atualiza uma role existente
func (s *serviceImpl) UpdateRole(ctx context.Context, id value_objects.UUID, displayName, description string, level int, updatedBy value_objects.UUID) (*Role, error) {
	// Buscar a role existente
	role, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("Role não encontrada", err)
	}

	// Verificar se o nível está disponível (excluindo a própria role)
	if role.Level != level {
		existingRoles, err := s.repo.GetRolesByLevel(ctx, role.TenantID, level)
		if err != nil {
			return nil, errors.NewInternalError("Erro ao verificar nível da role", err)
		}

		for _, existingRole := range existingRoles {
			if !existingRole.ID.Equals(role.ID) {
				return nil, errors.NewValidationError("Level", fmt.Sprintf("já existe uma role com nível %d neste tenant", level))
			}
		}
	}

	// Atualizar a role
	if err := role.Update(displayName, description, level, updatedBy); err != nil {
		return nil, err
	}

	// Salvar no repositório
	if err := s.repo.Update(ctx, role); err != nil {
		return nil, errors.NewInternalError("Erro ao atualizar role", err)
	}

	return role, nil
}

// GetRole busca uma role por ID
func (s *serviceImpl) GetRole(ctx context.Context, id value_objects.UUID) (*Role, error) {
	role, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("Role não encontrada", err)
	}

	return role, nil
}

// GetRoleByName busca uma role por nome
func (s *serviceImpl) GetRoleByName(ctx context.Context, tenantID value_objects.UUID, name string) (*Role, error) {
	role, err := s.repo.GetByName(ctx, tenantID, name)
	if err != nil {
		return nil, errors.NewNotFoundError("Role não encontrada", err)
	}

	return role, nil
}

// DeleteRole remove uma role
func (s *serviceImpl) DeleteRole(ctx context.Context, id value_objects.UUID, deletedBy value_objects.UUID) error {
	// Buscar a role
	role, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.NewNotFoundError("Role não encontrada", err)
	}

	// Verificar se é uma role do sistema
	if role.IsSystem {
		return errors.NewForbiddenError("role", "deletar roles do sistema")
	}

	// TODO: Verificar se a role está sendo usada por usuários
	// Esta verificação será implementada quando tivermos o relacionamento User-Role

	// Deletar a role
	if err := s.repo.Delete(ctx, id, deletedBy); err != nil {
		return errors.NewInternalError("Erro ao deletar role", err)
	}

	return nil
}

// ActivateRole ativa uma role
func (s *serviceImpl) ActivateRole(ctx context.Context, id value_objects.UUID, updatedBy value_objects.UUID) error {
	role, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.NewNotFoundError("Role não encontrada", err)
	}

	if err := role.Activate(updatedBy); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, role); err != nil {
		return errors.NewInternalError("Erro ao ativar role", err)
	}

	return nil
}

// DeactivateRole desativa uma role
func (s *serviceImpl) DeactivateRole(ctx context.Context, id value_objects.UUID, updatedBy value_objects.UUID) error {
	role, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.NewNotFoundError("Role não encontrada", err)
	}

	if err := role.Deactivate(updatedBy); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, role); err != nil {
		return errors.NewInternalError("Erro ao desativar role", err)
	}

	return nil
}

// ListRoles lista roles com filtros
func (s *serviceImpl) ListRoles(ctx context.Context, filters ListFilters) ([]*Role, int, error) {
	if err := filters.Validate(); err != nil {
		return nil, 0, err
	}

	roles, total, err := s.repo.List(ctx, filters)
	if err != nil {
		return nil, 0, errors.NewInternalError("Erro ao listar roles", err)
	}

	return roles, total, nil
}

// ListTenantRoles lista roles de um tenant
func (s *serviceImpl) ListTenantRoles(ctx context.Context, tenantID value_objects.UUID, filters ListFilters) ([]*Role, int, error) {
	if err := filters.Validate(); err != nil {
		return nil, 0, err
	}

	roles, total, err := s.repo.ListByTenant(ctx, tenantID, filters)
	if err != nil {
		return nil, 0, errors.NewInternalError("Erro ao listar roles do tenant", err)
	}

	return roles, total, nil
}

// ListSystemRoles lista roles do sistema
func (s *serviceImpl) ListSystemRoles(ctx context.Context) ([]*Role, error) {
	roles, err := s.repo.ListSystemRoles(ctx)
	if err != nil {
		return nil, errors.NewInternalError("Erro ao listar roles do sistema", err)
	}

	return roles, nil
}

// InitializeSystemRoles inicializa as roles padrão do sistema
func (s *serviceImpl) InitializeSystemRoles(ctx context.Context) error {
	systemRoles := GetSystemRoles()

	for _, role := range systemRoles {
		// Verificar se a role já existe
		existing, err := s.repo.GetSystemRoleByName(ctx, role.Name)
		if err == nil && existing != nil {
			continue // Role já existe, pular
		}

		// Criar a role do sistema
		if err := s.repo.Create(ctx, role); err != nil {
			return errors.NewInternalError(fmt.Sprintf("Erro ao criar role do sistema %s", role.Name), err)
		}
	}

	return nil
}

// ValidateRoleHierarchy valida se uma role pode gerenciar outra
func (s *serviceImpl) ValidateRoleHierarchy(ctx context.Context, managerRoleID, targetRoleID value_objects.UUID) error {
	managerRole, err := s.repo.GetByID(ctx, managerRoleID)
	if err != nil {
		return errors.NewNotFoundError("Role do gerenciador não encontrada", err)
	}

	targetRole, err := s.repo.GetByID(ctx, targetRoleID)
	if err != nil {
		return errors.NewNotFoundError("Role alvo não encontrada", err)
	}

	if !managerRole.CanManageRole(targetRole) {
		return errors.NewForbiddenError("role", "gerenciar a role alvo")
	}

	return nil
}

// GetAvailableLevels retorna os níveis disponíveis para um tenant
func (s *serviceImpl) GetAvailableLevels(ctx context.Context, tenantID value_objects.UUID) ([]int, error) {
	// Buscar todas as roles do tenant
	filters := ListFilters{
		TenantID: &tenantID,
		Active:   boolPtr(true),
		PageSize: 1000, // Buscar todas
	}

	roles, _, err := s.repo.ListByTenant(ctx, tenantID, filters)
	if err != nil {
		return nil, errors.NewInternalError("Erro ao buscar roles do tenant", err)
	}

	// Mapear níveis ocupados
	occupiedLevels := make(map[int]bool)
	for _, role := range roles {
		occupiedLevels[role.Level] = true
	}

	// Gerar lista de níveis disponíveis (10-999, excluindo 1-9 reservados para sistema)
	availableLevels := []int{}
	for level := 10; level <= 999; level++ {
		if !occupiedLevels[level] {
			availableLevels = append(availableLevels, level)
		}
	}

	return availableLevels, nil
}

// SuggestLevel sugere um nível para uma nova role
func (s *serviceImpl) SuggestLevel(ctx context.Context, tenantID value_objects.UUID) (int, error) {
	availableLevels, err := s.GetAvailableLevels(ctx, tenantID)
	if err != nil {
		return 0, err
	}

	if len(availableLevels) == 0 {
		return 0, errors.NewValidationError("Level", "não há níveis disponíveis para este tenant")
	}

	// Retornar o menor nível disponível
	return availableLevels[0], nil
}

// boolPtr retorna um ponteiro para bool
func boolPtr(b bool) *bool {
	return &b
}
