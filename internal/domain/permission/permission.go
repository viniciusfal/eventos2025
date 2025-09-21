package permission

import (
	"fmt"
	"strings"
	"time"

	"eventos-backend/internal/domain/shared/constants"
	"eventos-backend/internal/domain/shared/errors"
	"eventos-backend/internal/domain/shared/value_objects"
)

// Permission representa uma permissão no sistema
type Permission struct {
	ID          value_objects.UUID
	TenantID    value_objects.UUID
	Module      string // Módulo do sistema (auth, events, partners, etc.)
	Action      string // Ação (read, write, delete, admin)
	Resource    string // Recurso específico (opcional)
	Name        string // Nome único da permissão (MODULE_ACTION ou MODULE_ACTION_RESOURCE)
	DisplayName string // Nome para exibição
	Description string
	IsSystem    bool // Permissões do sistema não podem ser editadas
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   *value_objects.UUID
	UpdatedBy   *value_objects.UUID
}

// NewPermission cria uma nova permissão com validações
func NewPermission(tenantID value_objects.UUID, module, action, resource, displayName, description string, createdBy value_objects.UUID) (*Permission, error) {
	name := generatePermissionName(module, action, resource)

	permission := &Permission{
		ID:          value_objects.NewUUID(),
		TenantID:    tenantID,
		Module:      strings.ToLower(strings.TrimSpace(module)),
		Action:      strings.ToLower(strings.TrimSpace(action)),
		Resource:    strings.ToLower(strings.TrimSpace(resource)),
		Name:        name,
		DisplayName: strings.TrimSpace(displayName),
		Description: strings.TrimSpace(description),
		IsSystem:    false,
		Active:      true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		CreatedBy:   &createdBy,
		UpdatedBy:   &createdBy,
	}

	if err := permission.Validate(); err != nil {
		return nil, err
	}

	return permission, nil
}

// NewSystemPermission cria uma permissão do sistema
func NewSystemPermission(module, action, resource, displayName, description string) (*Permission, error) {
	name := generatePermissionName(module, action, resource)

	permission := &Permission{
		ID:          value_objects.NewUUID(),
		TenantID:    value_objects.UUID{}, // Permissões do sistema não têm tenant
		Module:      strings.ToLower(strings.TrimSpace(module)),
		Action:      strings.ToLower(strings.TrimSpace(action)),
		Resource:    strings.ToLower(strings.TrimSpace(resource)),
		Name:        name,
		DisplayName: strings.TrimSpace(displayName),
		Description: strings.TrimSpace(description),
		IsSystem:    true,
		Active:      true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := permission.Validate(); err != nil {
		return nil, err
	}

	return permission, nil
}

// generatePermissionName gera o nome único da permissão
func generatePermissionName(module, action, resource string) string {
	module = strings.ToUpper(strings.TrimSpace(module))
	action = strings.ToUpper(strings.TrimSpace(action))
	resource = strings.ToUpper(strings.TrimSpace(resource))

	if resource != "" {
		return fmt.Sprintf("%s_%s_%s", module, action, resource)
	}
	return fmt.Sprintf("%s_%s", module, action)
}

// Validate valida os dados da permissão
func (p *Permission) Validate() error {
	if p.ID.IsZero() {
		return errors.NewValidationError("ID", "é obrigatório")
	}

	if !p.IsSystem && p.TenantID.IsZero() {
		return errors.NewValidationError("TenantID", "é obrigatório para permissões não-sistema")
	}

	if err := p.validateModule(); err != nil {
		return err
	}

	if err := p.validateAction(); err != nil {
		return err
	}

	if err := p.validateName(); err != nil {
		return err
	}

	if err := p.validateDisplayName(); err != nil {
		return err
	}

	return nil
}

// validateModule valida o módulo
func (p *Permission) validateModule() error {
	if p.Module == "" {
		return errors.NewValidationError("Module", "é obrigatório")
	}

	if len(p.Module) < 2 {
		return errors.NewValidationError("Module", "deve ter pelo menos 2 caracteres")
	}

	if len(p.Module) > 50 {
		return errors.NewValidationError("Module", "deve ter no máximo 50 caracteres")
	}

	// Validar se é um módulo conhecido
	validModules := map[string]bool{
		constants.ModuleAuth:      true,
		constants.ModuleEvents:    true,
		constants.ModulePartners:  true,
		constants.ModuleEmployees: true,
		constants.ModuleCheckins:  true,
		constants.ModuleReports:   true,
		constants.ModuleAudit:     true,
		constants.ModuleQRCode:    true,
		constants.ModuleFacial:    true,
	}

	if !validModules[p.Module] {
		return errors.NewValidationError("Module", "módulo não reconhecido")
	}

	return nil
}

// validateAction valida a ação
func (p *Permission) validateAction() error {
	if p.Action == "" {
		return errors.NewValidationError("Action", "é obrigatório")
	}

	if len(p.Action) < 2 {
		return errors.NewValidationError("Action", "deve ter pelo menos 2 caracteres")
	}

	if len(p.Action) > 50 {
		return errors.NewValidationError("Action", "deve ter no máximo 50 caracteres")
	}

	// Validar se é uma ação conhecida
	validActions := map[string]bool{
		constants.PermissionRead:   true,
		constants.PermissionWrite:  true,
		constants.PermissionDelete: true,
		constants.PermissionAdmin:  true,
	}

	if !validActions[p.Action] {
		return errors.NewValidationError("Action", "ação não reconhecida")
	}

	return nil
}

// validateName valida o nome da permissão
func (p *Permission) validateName() error {
	if p.Name == "" {
		return errors.NewValidationError("Name", "é obrigatório")
	}

	if len(p.Name) < 3 {
		return errors.NewValidationError("Name", "deve ter pelo menos 3 caracteres")
	}

	if len(p.Name) > 100 {
		return errors.NewValidationError("Name", "deve ter no máximo 100 caracteres")
	}

	// Validar formato: apenas letras maiúsculas, números e underscore
	for _, char := range p.Name {
		if !((char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_') {
			return errors.NewValidationError("Name", "deve conter apenas letras maiúsculas, números e underscore")
		}
	}

	return nil
}

// validateDisplayName valida o nome de exibição
func (p *Permission) validateDisplayName() error {
	if p.DisplayName == "" {
		return errors.NewValidationError("DisplayName", "é obrigatório")
	}

	if len(p.DisplayName) < 2 {
		return errors.NewValidationError("DisplayName", "deve ter pelo menos 2 caracteres")
	}

	if len(p.DisplayName) > 100 {
		return errors.NewValidationError("DisplayName", "deve ter no máximo 100 caracteres")
	}

	return nil
}

// Update atualiza os dados da permissão
func (p *Permission) Update(displayName, description string, updatedBy value_objects.UUID) error {
	if p.IsSystem {
		return errors.NewForbiddenError("permission", "alterar permissões do sistema")
	}

	p.DisplayName = strings.TrimSpace(displayName)
	p.Description = strings.TrimSpace(description)
	p.UpdatedAt = time.Now()
	p.UpdatedBy = &updatedBy

	return p.Validate()
}

// Activate ativa a permissão
func (p *Permission) Activate(updatedBy value_objects.UUID) error {
	if p.IsSystem {
		return errors.NewForbiddenError("permission", "alterar status de permissões do sistema")
	}

	p.Active = true
	p.UpdatedAt = time.Now()
	p.UpdatedBy = &updatedBy

	return nil
}

// Deactivate desativa a permissão
func (p *Permission) Deactivate(updatedBy value_objects.UUID) error {
	if p.IsSystem {
		return errors.NewForbiddenError("permission", "alterar status de permissões do sistema")
	}

	p.Active = false
	p.UpdatedAt = time.Now()
	p.UpdatedBy = &updatedBy

	return nil
}

// IsReadPermission verifica se é uma permissão de leitura
func (p *Permission) IsReadPermission() bool {
	return p.Action == constants.PermissionRead
}

// IsWritePermission verifica se é uma permissão de escrita
func (p *Permission) IsWritePermission() bool {
	return p.Action == constants.PermissionWrite
}

// IsDeletePermission verifica se é uma permissão de exclusão
func (p *Permission) IsDeletePermission() bool {
	return p.Action == constants.PermissionDelete
}

// IsAdminPermission verifica se é uma permissão de administração
func (p *Permission) IsAdminPermission() bool {
	return p.Action == constants.PermissionAdmin
}

// HasResource verifica se a permissão tem um recurso específico
func (p *Permission) HasResource() bool {
	return p.Resource != ""
}

// MatchesPattern verifica se a permissão corresponde a um padrão
func (p *Permission) MatchesPattern(module, action, resource string) bool {
	if p.Module != strings.ToLower(module) {
		return false
	}

	if p.Action != strings.ToLower(action) {
		return false
	}

	// Se a permissão não tem recurso específico, corresponde a qualquer recurso do módulo/ação
	if p.Resource == "" {
		return true
	}

	// Se tem recurso específico, deve corresponder exatamente
	return p.Resource == strings.ToLower(resource)
}

// String retorna uma representação string da permissão
func (p *Permission) String() string {
	return fmt.Sprintf("Permission{ID: %s, Name: %s}", p.ID.String(), p.Name)
}

// GetSystemPermissions retorna as permissões padrão do sistema
func GetSystemPermissions() []*Permission {
	permissions := []*Permission{}

	// Permissões de autenticação
	authRead, _ := NewSystemPermission(constants.ModuleAuth, constants.PermissionRead, "", "Visualizar Autenticação", "Visualizar informações de autenticação")
	authWrite, _ := NewSystemPermission(constants.ModuleAuth, constants.PermissionWrite, "", "Gerenciar Autenticação", "Gerenciar configurações de autenticação")
	authAdmin, _ := NewSystemPermission(constants.ModuleAuth, constants.PermissionAdmin, "", "Administrar Autenticação", "Administração completa do sistema de autenticação")

	permissions = append(permissions, authRead, authWrite, authAdmin)

	// Permissões de eventos
	eventsRead, _ := NewSystemPermission(constants.ModuleEvents, constants.PermissionRead, "", "Visualizar Eventos", "Visualizar eventos e suas informações")
	eventsWrite, _ := NewSystemPermission(constants.ModuleEvents, constants.PermissionWrite, "", "Gerenciar Eventos", "Criar e editar eventos")
	eventsDelete, _ := NewSystemPermission(constants.ModuleEvents, constants.PermissionDelete, "", "Excluir Eventos", "Excluir eventos")
	eventsAdmin, _ := NewSystemPermission(constants.ModuleEvents, constants.PermissionAdmin, "", "Administrar Eventos", "Administração completa de eventos")

	permissions = append(permissions, eventsRead, eventsWrite, eventsDelete, eventsAdmin)

	// Permissões de parceiros
	partnersRead, _ := NewSystemPermission(constants.ModulePartners, constants.PermissionRead, "", "Visualizar Parceiros", "Visualizar parceiros e suas informações")
	partnersWrite, _ := NewSystemPermission(constants.ModulePartners, constants.PermissionWrite, "", "Gerenciar Parceiros", "Criar e editar parceiros")
	partnersDelete, _ := NewSystemPermission(constants.ModulePartners, constants.PermissionDelete, "", "Excluir Parceiros", "Excluir parceiros")
	partnersAdmin, _ := NewSystemPermission(constants.ModulePartners, constants.PermissionAdmin, "", "Administrar Parceiros", "Administração completa de parceiros")

	permissions = append(permissions, partnersRead, partnersWrite, partnersDelete, partnersAdmin)

	// Permissões de funcionários
	employeesRead, _ := NewSystemPermission(constants.ModuleEmployees, constants.PermissionRead, "", "Visualizar Funcionários", "Visualizar funcionários e suas informações")
	employeesWrite, _ := NewSystemPermission(constants.ModuleEmployees, constants.PermissionWrite, "", "Gerenciar Funcionários", "Criar e editar funcionários")
	employeesDelete, _ := NewSystemPermission(constants.ModuleEmployees, constants.PermissionDelete, "", "Excluir Funcionários", "Excluir funcionários")
	employeesAdmin, _ := NewSystemPermission(constants.ModuleEmployees, constants.PermissionAdmin, "", "Administrar Funcionários", "Administração completa de funcionários")

	permissions = append(permissions, employeesRead, employeesWrite, employeesDelete, employeesAdmin)

	// Permissões de check-ins
	checkinsRead, _ := NewSystemPermission(constants.ModuleCheckins, constants.PermissionRead, "", "Visualizar Check-ins", "Visualizar check-ins e relatórios")
	checkinsWrite, _ := NewSystemPermission(constants.ModuleCheckins, constants.PermissionWrite, "", "Realizar Check-ins", "Realizar check-ins e check-outs")
	checkinsAdmin, _ := NewSystemPermission(constants.ModuleCheckins, constants.PermissionAdmin, "", "Administrar Check-ins", "Administração completa de check-ins")

	permissions = append(permissions, checkinsRead, checkinsWrite, checkinsAdmin)

	// Permissões de relatórios
	reportsRead, _ := NewSystemPermission(constants.ModuleReports, constants.PermissionRead, "", "Visualizar Relatórios", "Visualizar relatórios e estatísticas")
	reportsWrite, _ := NewSystemPermission(constants.ModuleReports, constants.PermissionWrite, "", "Gerar Relatórios", "Gerar e exportar relatórios")
	reportsAdmin, _ := NewSystemPermission(constants.ModuleReports, constants.PermissionAdmin, "", "Administrar Relatórios", "Administração completa de relatórios")

	permissions = append(permissions, reportsRead, reportsWrite, reportsAdmin)

	// Permissões de auditoria
	auditRead, _ := NewSystemPermission(constants.ModuleAudit, constants.PermissionRead, "", "Visualizar Auditoria", "Visualizar logs de auditoria")
	auditAdmin, _ := NewSystemPermission(constants.ModuleAudit, constants.PermissionAdmin, "", "Administrar Auditoria", "Administração completa de auditoria")

	permissions = append(permissions, auditRead, auditAdmin)

	return permissions
}
