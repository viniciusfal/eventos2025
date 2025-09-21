package permission

import (
	"context"
	"fmt"

	"eventos-backend/internal/domain/shared/errors"
	"eventos-backend/internal/domain/shared/value_objects"
)

// Service define a interface para serviços de domínio de permissões
type Service interface {
	// CreatePermission cria uma nova permissão
	CreatePermission(ctx context.Context, tenantID value_objects.UUID, module, action, resource, displayName, description string, createdBy value_objects.UUID) (*Permission, error)

	// UpdatePermission atualiza uma permissão existente
	UpdatePermission(ctx context.Context, id value_objects.UUID, displayName, description string, updatedBy value_objects.UUID) (*Permission, error)

	// GetPermission busca uma permissão por ID
	GetPermission(ctx context.Context, id value_objects.UUID) (*Permission, error)

	// GetPermissionByName busca uma permissão por nome
	GetPermissionByName(ctx context.Context, tenantID value_objects.UUID, name string) (*Permission, error)

	// DeletePermission remove uma permissão
	DeletePermission(ctx context.Context, id value_objects.UUID, deletedBy value_objects.UUID) error

	// ActivatePermission ativa uma permissão
	ActivatePermission(ctx context.Context, id value_objects.UUID, updatedBy value_objects.UUID) error

	// DeactivatePermission desativa uma permissão
	DeactivatePermission(ctx context.Context, id value_objects.UUID, updatedBy value_objects.UUID) error

	// ListPermissions lista permissões com filtros
	ListPermissions(ctx context.Context, filters ListFilters) ([]*Permission, int, error)

	// ListTenantPermissions lista permissões de um tenant
	ListTenantPermissions(ctx context.Context, tenantID value_objects.UUID, filters ListFilters) ([]*Permission, int, error)

	// ListSystemPermissions lista permissões do sistema
	ListSystemPermissions(ctx context.Context) ([]*Permission, error)

	// InitializeSystemPermissions inicializa as permissões padrão do sistema
	InitializeSystemPermissions(ctx context.Context) error

	// GetPermissionsByModule busca permissões por módulo
	GetPermissionsByModule(ctx context.Context, tenantID value_objects.UUID, module string) ([]*Permission, error)

	// GetPermissionsByPattern busca permissões que correspondem a um padrão
	GetPermissionsByPattern(ctx context.Context, tenantID value_objects.UUID, module, action, resource string) ([]*Permission, error)

	// ValidatePermissionAccess valida se uma permissão permite acesso a um recurso
	ValidatePermissionAccess(ctx context.Context, permissionID value_objects.UUID, module, action, resource string) (bool, error)

	// GetAvailableModules retorna os módulos disponíveis para um tenant
	GetAvailableModules(ctx context.Context, tenantID value_objects.UUID) ([]string, error)

	// GetAvailableActions retorna as ações disponíveis para um módulo
	GetAvailableActions(ctx context.Context, tenantID value_objects.UUID, module string) ([]string, error)

	// BulkCreatePermissions cria múltiplas permissões de uma vez
	BulkCreatePermissions(ctx context.Context, tenantID value_objects.UUID, permissions []CreatePermissionRequest, createdBy value_objects.UUID) ([]*Permission, error)
}

// CreatePermissionRequest representa uma requisição para criar permissão
type CreatePermissionRequest struct {
	Module      string
	Action      string
	Resource    string
	DisplayName string
	Description string
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

// CreatePermission cria uma nova permissão
func (s *serviceImpl) CreatePermission(ctx context.Context, tenantID value_objects.UUID, module, action, resource, displayName, description string, createdBy value_objects.UUID) (*Permission, error) {
	// Gerar o nome da permissão
	name := generatePermissionName(module, action, resource)

	// Verificar se já existe uma permissão com este nome no tenant
	exists, err := s.repo.ExistsByName(ctx, tenantID, name, nil)
	if err != nil {
		return nil, errors.NewInternalError("Erro ao verificar existência da permissão", err)
	}

	if exists {
		return nil, errors.NewAlreadyExistsError("Permission", "name", name)
	}

	// Criar a permissão
	permission, err := NewPermission(tenantID, module, action, resource, displayName, description, createdBy)
	if err != nil {
		return nil, err
	}

	// Salvar no repositório
	if err := s.repo.Create(ctx, permission); err != nil {
		return nil, errors.NewInternalError("Erro ao criar permissão", err)
	}

	return permission, nil
}

// UpdatePermission atualiza uma permissão existente
func (s *serviceImpl) UpdatePermission(ctx context.Context, id value_objects.UUID, displayName, description string, updatedBy value_objects.UUID) (*Permission, error) {
	// Buscar a permissão existente
	permission, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("Permission não encontrada", err)
	}

	// Atualizar a permissão
	if err := permission.Update(displayName, description, updatedBy); err != nil {
		return nil, err
	}

	// Salvar no repositório
	if err := s.repo.Update(ctx, permission); err != nil {
		return nil, errors.NewInternalError("Erro ao atualizar permissão", err)
	}

	return permission, nil
}

// GetPermission busca uma permissão por ID
func (s *serviceImpl) GetPermission(ctx context.Context, id value_objects.UUID) (*Permission, error) {
	permission, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("Permission não encontrada", err)
	}

	return permission, nil
}

// GetPermissionByName busca uma permissão por nome
func (s *serviceImpl) GetPermissionByName(ctx context.Context, tenantID value_objects.UUID, name string) (*Permission, error) {
	permission, err := s.repo.GetByName(ctx, tenantID, name)
	if err != nil {
		return nil, errors.NewNotFoundError("Permission não encontrada", err)
	}

	return permission, nil
}

// DeletePermission remove uma permissão
func (s *serviceImpl) DeletePermission(ctx context.Context, id value_objects.UUID, deletedBy value_objects.UUID) error {
	// Buscar a permissão
	permission, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.NewNotFoundError("Permission não encontrada", err)
	}

	// Verificar se é uma permissão do sistema
	if permission.IsSystem {
		return errors.NewForbiddenError("permission", "deletar permissões do sistema")
	}

	// TODO: Verificar se a permissão está sendo usada por roles
	// Esta verificação será implementada quando tivermos o relacionamento Role-Permission

	// Deletar a permissão
	if err := s.repo.Delete(ctx, id, deletedBy); err != nil {
		return errors.NewInternalError("Erro ao deletar permissão", err)
	}

	return nil
}

// ActivatePermission ativa uma permissão
func (s *serviceImpl) ActivatePermission(ctx context.Context, id value_objects.UUID, updatedBy value_objects.UUID) error {
	permission, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.NewNotFoundError("Permission não encontrada", err)
	}

	if err := permission.Activate(updatedBy); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, permission); err != nil {
		return errors.NewInternalError("Erro ao ativar permissão", err)
	}

	return nil
}

// DeactivatePermission desativa uma permissão
func (s *serviceImpl) DeactivatePermission(ctx context.Context, id value_objects.UUID, updatedBy value_objects.UUID) error {
	permission, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.NewNotFoundError("Permission não encontrada", err)
	}

	if err := permission.Deactivate(updatedBy); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, permission); err != nil {
		return errors.NewInternalError("Erro ao desativar permissão", err)
	}

	return nil
}

// ListPermissions lista permissões com filtros
func (s *serviceImpl) ListPermissions(ctx context.Context, filters ListFilters) ([]*Permission, int, error) {
	if err := filters.Validate(); err != nil {
		return nil, 0, err
	}

	permissions, total, err := s.repo.List(ctx, filters)
	if err != nil {
		return nil, 0, errors.NewInternalError("Erro ao listar permissões", err)
	}

	return permissions, total, nil
}

// ListTenantPermissions lista permissões de um tenant
func (s *serviceImpl) ListTenantPermissions(ctx context.Context, tenantID value_objects.UUID, filters ListFilters) ([]*Permission, int, error) {
	if err := filters.Validate(); err != nil {
		return nil, 0, err
	}

	permissions, total, err := s.repo.ListByTenant(ctx, tenantID, filters)
	if err != nil {
		return nil, 0, errors.NewInternalError("Erro ao listar permissões do tenant", err)
	}

	return permissions, total, nil
}

// ListSystemPermissions lista permissões do sistema
func (s *serviceImpl) ListSystemPermissions(ctx context.Context) ([]*Permission, error) {
	permissions, err := s.repo.ListSystemPermissions(ctx)
	if err != nil {
		return nil, errors.NewInternalError("Erro ao listar permissões do sistema", err)
	}

	return permissions, nil
}

// InitializeSystemPermissions inicializa as permissões padrão do sistema
func (s *serviceImpl) InitializeSystemPermissions(ctx context.Context) error {
	systemPermissions := GetSystemPermissions()

	for _, permission := range systemPermissions {
		// Verificar se a permissão já existe
		existing, err := s.repo.GetSystemPermissionByName(ctx, permission.Name)
		if err == nil && existing != nil {
			continue // Permissão já existe, pular
		}

		// Criar a permissão do sistema
		if err := s.repo.Create(ctx, permission); err != nil {
			return errors.NewInternalError(fmt.Sprintf("Erro ao criar permissão do sistema %s", permission.Name), err)
		}
	}

	return nil
}

// GetPermissionsByModule busca permissões por módulo
func (s *serviceImpl) GetPermissionsByModule(ctx context.Context, tenantID value_objects.UUID, module string) ([]*Permission, error) {
	permissions, err := s.repo.GetByModule(ctx, tenantID, module)
	if err != nil {
		return nil, errors.NewInternalError("Erro ao buscar permissões por módulo", err)
	}

	return permissions, nil
}

// GetPermissionsByPattern busca permissões que correspondem a um padrão
func (s *serviceImpl) GetPermissionsByPattern(ctx context.Context, tenantID value_objects.UUID, module, action, resource string) ([]*Permission, error) {
	permissions, err := s.repo.GetByPattern(ctx, tenantID, module, action, resource)
	if err != nil {
		return nil, errors.NewInternalError("Erro ao buscar permissões por padrão", err)
	}

	return permissions, nil
}

// ValidatePermissionAccess valida se uma permissão permite acesso a um recurso
func (s *serviceImpl) ValidatePermissionAccess(ctx context.Context, permissionID value_objects.UUID, module, action, resource string) (bool, error) {
	permission, err := s.repo.GetByID(ctx, permissionID)
	if err != nil {
		return false, errors.NewNotFoundError("Permission não encontrada", err)
	}

	if !permission.Active {
		return false, nil
	}

	return permission.MatchesPattern(module, action, resource), nil
}

// GetAvailableModules retorna os módulos disponíveis para um tenant
func (s *serviceImpl) GetAvailableModules(ctx context.Context, tenantID value_objects.UUID) ([]string, error) {
	modules, err := s.repo.GetModules(ctx, tenantID)
	if err != nil {
		return nil, errors.NewInternalError("Erro ao buscar módulos disponíveis", err)
	}

	return modules, nil
}

// GetAvailableActions retorna as ações disponíveis para um módulo
func (s *serviceImpl) GetAvailableActions(ctx context.Context, tenantID value_objects.UUID, module string) ([]string, error) {
	actions, err := s.repo.GetActions(ctx, tenantID, module)
	if err != nil {
		return nil, errors.NewInternalError("Erro ao buscar ações disponíveis", err)
	}

	return actions, nil
}

// BulkCreatePermissions cria múltiplas permissões de uma vez
func (s *serviceImpl) BulkCreatePermissions(ctx context.Context, tenantID value_objects.UUID, requests []CreatePermissionRequest, createdBy value_objects.UUID) ([]*Permission, error) {
	permissions := make([]*Permission, 0, len(requests))

	for _, req := range requests {
		permission, err := s.CreatePermission(ctx, tenantID, req.Module, req.Action, req.Resource, req.DisplayName, req.Description, createdBy)
		if err != nil {
			// Se uma permissão falhar, continuar com as outras
			// Em um cenário real, você pode querer fazer rollback
			continue
		}
		permissions = append(permissions, permission)
	}

	return permissions, nil
}
