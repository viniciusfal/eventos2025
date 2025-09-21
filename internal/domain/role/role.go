package role

import (
	"fmt"
	"strings"
	"time"

	"eventos-backend/internal/domain/shared/constants"
	"eventos-backend/internal/domain/shared/errors"
	"eventos-backend/internal/domain/shared/value_objects"
)

// Role representa um papel/função no sistema
type Role struct {
	ID          value_objects.UUID
	TenantID    value_objects.UUID
	Name        string
	DisplayName string
	Description string
	Level       int  // Nível hierárquico (1=mais alto, 999=mais baixo)
	IsSystem    bool // Roles do sistema não podem ser editados
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   *value_objects.UUID
	UpdatedBy   *value_objects.UUID
}

// NewRole cria uma nova role com validações
func NewRole(tenantID value_objects.UUID, name, displayName, description string, level int, createdBy value_objects.UUID) (*Role, error) {
	role := &Role{
		ID:          value_objects.NewUUID(),
		TenantID:    tenantID,
		Name:        strings.ToUpper(strings.TrimSpace(name)),
		DisplayName: strings.TrimSpace(displayName),
		Description: strings.TrimSpace(description),
		Level:       level,
		IsSystem:    false,
		Active:      true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		CreatedBy:   &createdBy,
		UpdatedBy:   &createdBy,
	}

	if err := role.Validate(); err != nil {
		return nil, err
	}

	return role, nil
}

// NewSystemRole cria uma role do sistema
func NewSystemRole(name, displayName, description string, level int) (*Role, error) {
	role := &Role{
		ID:          value_objects.NewUUID(),
		TenantID:    value_objects.UUID{}, // Roles do sistema não têm tenant
		Name:        strings.ToUpper(strings.TrimSpace(name)),
		DisplayName: strings.TrimSpace(displayName),
		Description: strings.TrimSpace(description),
		Level:       level,
		IsSystem:    true,
		Active:      true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := role.Validate(); err != nil {
		return nil, err
	}

	return role, nil
}

// Validate valida os dados da role
func (r *Role) Validate() error {
	if r.ID.IsZero() {
		return errors.NewValidationError("ID", "é obrigatório")
	}

	if !r.IsSystem && r.TenantID.IsZero() {
		return errors.NewValidationError("TenantID", "é obrigatório para roles não-sistema")
	}

	if err := r.validateName(); err != nil {
		return err
	}

	if err := r.validateDisplayName(); err != nil {
		return err
	}

	if err := r.validateLevel(); err != nil {
		return err
	}

	return nil
}

// validateName valida o nome da role
func (r *Role) validateName() error {
	if r.Name == "" {
		return errors.NewValidationError("Name", "é obrigatório")
	}

	if len(r.Name) < 2 {
		return errors.NewValidationError("Name", "deve ter pelo menos 2 caracteres")
	}

	if len(r.Name) > 50 {
		return errors.NewValidationError("Name", "deve ter no máximo 50 caracteres")
	}

	// Validar formato: apenas letras, números e underscore
	for _, char := range r.Name {
		if !((char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_') {
			return errors.NewValidationError("Name", "deve conter apenas letras maiúsculas, números e underscore")
		}
	}

	return nil
}

// validateDisplayName valida o nome de exibição
func (r *Role) validateDisplayName() error {
	if r.DisplayName == "" {
		return errors.NewValidationError("DisplayName", "é obrigatório")
	}

	if len(r.DisplayName) < 2 {
		return errors.NewValidationError("DisplayName", "deve ter pelo menos 2 caracteres")
	}

	if len(r.DisplayName) > 100 {
		return errors.NewValidationError("DisplayName", "deve ter no máximo 100 caracteres")
	}

	return nil
}

// validateLevel valida o nível hierárquico
func (r *Role) validateLevel() error {
	if r.Level < 1 {
		return errors.NewValidationError("Level", "deve ser maior que 0")
	}

	if r.Level > 999 {
		return errors.NewValidationError("Level", "deve ser menor que 1000")
	}

	return nil
}

// Update atualiza os dados da role
func (r *Role) Update(displayName, description string, level int, updatedBy value_objects.UUID) error {
	if r.IsSystem {
		return errors.NewForbiddenError("role", "alterar roles do sistema")
	}

	r.DisplayName = strings.TrimSpace(displayName)
	r.Description = strings.TrimSpace(description)
	r.Level = level
	r.UpdatedAt = time.Now()
	r.UpdatedBy = &updatedBy

	return r.Validate()
}

// Activate ativa a role
func (r *Role) Activate(updatedBy value_objects.UUID) error {
	if r.IsSystem {
		return errors.NewForbiddenError("role", "alterar status de roles do sistema")
	}

	r.Active = true
	r.UpdatedAt = time.Now()
	r.UpdatedBy = &updatedBy

	return nil
}

// Deactivate desativa a role
func (r *Role) Deactivate(updatedBy value_objects.UUID) error {
	if r.IsSystem {
		return errors.NewForbiddenError("role", "alterar status de roles do sistema")
	}

	r.Active = false
	r.UpdatedAt = time.Now()
	r.UpdatedBy = &updatedBy

	return nil
}

// CanManageRole verifica se esta role pode gerenciar outra role
func (r *Role) CanManageRole(targetRole *Role) bool {
	// Roles do sistema não podem ser gerenciadas
	if targetRole.IsSystem {
		return false
	}

	// Só pode gerenciar roles do mesmo tenant
	if !r.TenantID.Equals(targetRole.TenantID) {
		return false
	}

	// Só pode gerenciar roles de nível inferior
	return r.Level < targetRole.Level
}

// IsHigherThan verifica se esta role tem nível superior a outra
func (r *Role) IsHigherThan(other *Role) bool {
	return r.Level < other.Level
}

// IsLowerThan verifica se esta role tem nível inferior a outra
func (r *Role) IsLowerThan(other *Role) bool {
	return r.Level > other.Level
}

// IsSameLevel verifica se esta role tem o mesmo nível de outra
func (r *Role) IsSameLevel(other *Role) bool {
	return r.Level == other.Level
}

// String retorna uma representação string da role
func (r *Role) String() string {
	return fmt.Sprintf("Role{ID: %s, Name: %s, Level: %d}", r.ID.String(), r.Name, r.Level)
}

// GetSystemRoles retorna as roles padrão do sistema
func GetSystemRoles() []*Role {
	roles := []*Role{}

	// Super Admin - Nível mais alto
	superAdmin, _ := NewSystemRole(
		constants.RoleSuperAdmin,
		"Super Administrador",
		"Acesso total ao sistema, incluindo gerenciamento de tenants",
		1,
	)
	roles = append(roles, superAdmin)

	// Admin - Administrador do tenant
	admin, _ := NewSystemRole(
		constants.RoleAdmin,
		"Administrador",
		"Administrador completo do tenant, pode gerenciar usuários e configurações",
		10,
	)
	roles = append(roles, admin)

	// Manager - Gerente
	manager, _ := NewSystemRole(
		constants.RoleManager,
		"Gerente",
		"Pode gerenciar eventos, parceiros e funcionários",
		20,
	)
	roles = append(roles, manager)

	// Operator - Operador
	operator, _ := NewSystemRole(
		constants.RoleOperator,
		"Operador",
		"Pode realizar check-ins e operações básicas",
		30,
	)
	roles = append(roles, operator)

	// Viewer - Visualizador
	viewer, _ := NewSystemRole(
		constants.RoleViewer,
		"Visualizador",
		"Apenas visualização de dados, sem permissões de alteração",
		40,
	)
	roles = append(roles, viewer)

	return roles
}
