package role

import (
	"context"

	"eventos-backend/internal/domain/shared/errors"
	"eventos-backend/internal/domain/shared/value_objects"
)

// RolePermissionService define operações para gerenciar relacionamentos role-permission
type RolePermissionService interface {
	// GrantPermissionToRole concede uma permissão a uma role
	GrantPermissionToRole(ctx context.Context, roleID, permissionID, grantedBy value_objects.UUID) error

	// RevokePermissionFromRole revoga uma permissão de uma role
	RevokePermissionFromRole(ctx context.Context, roleID, permissionID value_objects.UUID) error

	// GetRolePermissions busca todas as permissões de uma role
	GetRolePermissions(ctx context.Context, roleID value_objects.UUID) ([]*RolePermission, error)

	// GetPermissionRoles busca todas as roles que têm uma permissão
	GetPermissionRoles(ctx context.Context, permissionID value_objects.UUID) ([]*RolePermission, error)

	// HasPermission verifica se uma role tem uma permissão específica
	HasPermission(ctx context.Context, roleID, permissionID value_objects.UUID) (bool, error)

	// BulkGrantPermissions concede múltiplas permissões a uma role
	BulkGrantPermissions(ctx context.Context, roleID value_objects.UUID, permissionIDs []value_objects.UUID, grantedBy value_objects.UUID) error

	// BulkRevokePermissions revoga múltiplas permissões de uma role
	BulkRevokePermissions(ctx context.Context, roleID value_objects.UUID, permissionIDs []value_objects.UUID) error

	// SyncRolePermissions sincroniza as permissões de uma role (remove antigas e adiciona novas)
	SyncRolePermissions(ctx context.Context, roleID value_objects.UUID, permissionIDs []value_objects.UUID, grantedBy value_objects.UUID) error

	// ValidateRolePermissionAccess valida se uma role tem acesso a um recurso
	ValidateRolePermissionAccess(ctx context.Context, roleID value_objects.UUID, module, action, resource string) (bool, error)

	// GetEffectivePermissions retorna todas as permissões efetivas de uma role (incluindo hierarquia)
	GetEffectivePermissions(ctx context.Context, roleID value_objects.UUID) ([]*RolePermission, error)

	// ListRolePermissions lista relacionamentos com filtros
	ListRolePermissions(ctx context.Context, filters RolePermissionFilters) ([]*RolePermission, int, error)
}

// rolePermissionServiceImpl implementa RolePermissionService
type rolePermissionServiceImpl struct {
	roleRepo           Repository
	rolePermissionRepo RolePermissionRepository
}

// NewRolePermissionService cria uma nova instância do serviço
func NewRolePermissionService(roleRepo Repository, rolePermissionRepo RolePermissionRepository) RolePermissionService {
	return &rolePermissionServiceImpl{
		roleRepo:           roleRepo,
		rolePermissionRepo: rolePermissionRepo,
	}
}

// GrantPermissionToRole concede uma permissão a uma role
func (s *rolePermissionServiceImpl) GrantPermissionToRole(ctx context.Context, roleID, permissionID, grantedBy value_objects.UUID) error {
	// Verificar se a role existe
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return errors.NewNotFoundError("Role não encontrada", err)
	}

	// Verificar se já tem a permissão
	hasPermission, err := s.rolePermissionRepo.HasPermission(ctx, roleID, permissionID)
	if err != nil {
		return errors.NewInternalError("Erro ao verificar permissão existente", err)
	}

	if hasPermission {
		return errors.NewAlreadyExistsError("RolePermission", "permission", permissionID.String())
	}

	// Criar o relacionamento
	rolePermission := NewRolePermission(roleID, permissionID, role.TenantID, grantedBy)

	// Salvar no repositório
	if err := s.rolePermissionRepo.GrantPermission(ctx, rolePermission); err != nil {
		return errors.NewInternalError("Erro ao conceder permissão", err)
	}

	return nil
}

// RevokePermissionFromRole revoga uma permissão de uma role
func (s *rolePermissionServiceImpl) RevokePermissionFromRole(ctx context.Context, roleID, permissionID value_objects.UUID) error {
	// Verificar se a role existe
	_, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return errors.NewNotFoundError("Role não encontrada", err)
	}

	// Verificar se tem a permissão
	hasPermission, err := s.rolePermissionRepo.HasPermission(ctx, roleID, permissionID)
	if err != nil {
		return errors.NewInternalError("Erro ao verificar permissão", err)
	}

	if !hasPermission {
		return errors.NewNotFoundError("RolePermission não encontrada", nil)
	}

	// Revogar a permissão
	if err := s.rolePermissionRepo.RevokePermission(ctx, roleID, permissionID); err != nil {
		return errors.NewInternalError("Erro ao revogar permissão", err)
	}

	return nil
}

// GetRolePermissions busca todas as permissões de uma role
func (s *rolePermissionServiceImpl) GetRolePermissions(ctx context.Context, roleID value_objects.UUID) ([]*RolePermission, error) {
	// Verificar se a role existe
	_, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return nil, errors.NewNotFoundError("Role não encontrada", err)
	}

	permissions, err := s.rolePermissionRepo.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, errors.NewInternalError("Erro ao buscar permissões da role", err)
	}

	return permissions, nil
}

// GetPermissionRoles busca todas as roles que têm uma permissão
func (s *rolePermissionServiceImpl) GetPermissionRoles(ctx context.Context, permissionID value_objects.UUID) ([]*RolePermission, error) {
	roles, err := s.rolePermissionRepo.GetPermissionRoles(ctx, permissionID)
	if err != nil {
		return nil, errors.NewInternalError("Erro ao buscar roles da permissão", err)
	}

	return roles, nil
}

// HasPermission verifica se uma role tem uma permissão específica
func (s *rolePermissionServiceImpl) HasPermission(ctx context.Context, roleID, permissionID value_objects.UUID) (bool, error) {
	hasPermission, err := s.rolePermissionRepo.HasPermission(ctx, roleID, permissionID)
	if err != nil {
		return false, errors.NewInternalError("Erro ao verificar permissão", err)
	}

	return hasPermission, nil
}

// BulkGrantPermissions concede múltiplas permissões a uma role
func (s *rolePermissionServiceImpl) BulkGrantPermissions(ctx context.Context, roleID value_objects.UUID, permissionIDs []value_objects.UUID, grantedBy value_objects.UUID) error {
	// Verificar se a role existe
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return errors.NewNotFoundError("Role não encontrada", err)
	}

	// Criar relacionamentos
	rolePermissions := make([]*RolePermission, 0, len(permissionIDs))
	for _, permissionID := range permissionIDs {
		// Verificar se já tem a permissão
		hasPermission, err := s.rolePermissionRepo.HasPermission(ctx, roleID, permissionID)
		if err != nil {
			continue // Pular em caso de erro
		}

		if hasPermission {
			continue // Já tem a permissão, pular
		}

		rolePermission := NewRolePermission(roleID, permissionID, role.TenantID, grantedBy)
		rolePermissions = append(rolePermissions, rolePermission)
	}

	// Salvar em lote
	if len(rolePermissions) > 0 {
		if err := s.rolePermissionRepo.BulkGrantPermissions(ctx, rolePermissions); err != nil {
			return errors.NewInternalError("Erro ao conceder permissões em lote", err)
		}
	}

	return nil
}

// BulkRevokePermissions revoga múltiplas permissões de uma role
func (s *rolePermissionServiceImpl) BulkRevokePermissions(ctx context.Context, roleID value_objects.UUID, permissionIDs []value_objects.UUID) error {
	// Verificar se a role existe
	_, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return errors.NewNotFoundError("Role não encontrada", err)
	}

	// Revogar permissões em lote
	if err := s.rolePermissionRepo.BulkRevokePermissions(ctx, roleID, permissionIDs); err != nil {
		return errors.NewInternalError("Erro ao revogar permissões em lote", err)
	}

	return nil
}

// SyncRolePermissions sincroniza as permissões de uma role
func (s *rolePermissionServiceImpl) SyncRolePermissions(ctx context.Context, roleID value_objects.UUID, permissionIDs []value_objects.UUID, grantedBy value_objects.UUID) error {
	// Buscar permissões atuais
	currentPermissions, err := s.rolePermissionRepo.GetRolePermissions(ctx, roleID)
	if err != nil {
		return errors.NewInternalError("Erro ao buscar permissões atuais", err)
	}

	// Mapear permissões atuais
	currentMap := make(map[string]bool)
	currentIDs := make([]value_objects.UUID, 0, len(currentPermissions))
	for _, rp := range currentPermissions {
		currentMap[rp.PermissionID.String()] = true
		currentIDs = append(currentIDs, rp.PermissionID)
	}

	// Mapear novas permissões
	newMap := make(map[string]bool)
	for _, permissionID := range permissionIDs {
		newMap[permissionID.String()] = true
	}

	// Identificar permissões a remover
	toRemove := make([]value_objects.UUID, 0)
	for _, permissionID := range currentIDs {
		if !newMap[permissionID.String()] {
			toRemove = append(toRemove, permissionID)
		}
	}

	// Identificar permissões a adicionar
	toAdd := make([]value_objects.UUID, 0)
	for _, permissionID := range permissionIDs {
		if !currentMap[permissionID.String()] {
			toAdd = append(toAdd, permissionID)
		}
	}

	// Remover permissões
	if len(toRemove) > 0 {
		if err := s.BulkRevokePermissions(ctx, roleID, toRemove); err != nil {
			return err
		}
	}

	// Adicionar permissões
	if len(toAdd) > 0 {
		if err := s.BulkGrantPermissions(ctx, roleID, toAdd, grantedBy); err != nil {
			return err
		}
	}

	return nil
}

// ValidateRolePermissionAccess valida se uma role tem acesso a um recurso
func (s *rolePermissionServiceImpl) ValidateRolePermissionAccess(ctx context.Context, roleID value_objects.UUID, module, action, resource string) (bool, error) {
	// Buscar permissões da role
	rolePermissions, err := s.rolePermissionRepo.GetRolePermissions(ctx, roleID)
	if err != nil {
		return false, errors.NewInternalError("Erro ao buscar permissões da role", err)
	}

	// TODO: Aqui precisaríamos buscar as permissões completas e verificar se alguma corresponde ao padrão
	// Por enquanto, retornamos true se a role tem pelo menos uma permissão ativa
	for _, rp := range rolePermissions {
		if rp.Active {
			return true, nil
		}
	}

	return false, nil
}

// GetEffectivePermissions retorna todas as permissões efetivas de uma role
func (s *rolePermissionServiceImpl) GetEffectivePermissions(ctx context.Context, roleID value_objects.UUID) ([]*RolePermission, error) {
	// Buscar role
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return nil, errors.NewNotFoundError("Role não encontrada", err)
	}

	// Buscar permissões diretas
	permissions, err := s.rolePermissionRepo.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, errors.NewInternalError("Erro ao buscar permissões da role", err)
	}

	// TODO: Implementar herança de permissões baseada na hierarquia de roles
	// Por enquanto, retornamos apenas as permissões diretas
	effectivePermissions := make([]*RolePermission, 0)
	for _, rp := range permissions {
		if rp.Active {
			effectivePermissions = append(effectivePermissions, rp)
		}
	}

	// Se a role tem nível superior, pode herdar permissões de roles de nível inferior
	if role.Level < 50 { // Roles de nível superior (administrativas)
		// TODO: Buscar e incluir permissões herdadas
	}

	return effectivePermissions, nil
}

// ListRolePermissions lista relacionamentos com filtros
func (s *rolePermissionServiceImpl) ListRolePermissions(ctx context.Context, filters RolePermissionFilters) ([]*RolePermission, int, error) {
	if err := filters.Validate(); err != nil {
		return nil, 0, err
	}

	rolePermissions, total, err := s.rolePermissionRepo.ListRolePermissions(ctx, filters)
	if err != nil {
		return nil, 0, errors.NewInternalError("Erro ao listar relacionamentos role-permission", err)
	}

	return rolePermissions, total, nil
}
