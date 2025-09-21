package tenant

import (
	"time"

	"eventos-backend/internal/domain/shared/constants"
	"eventos-backend/internal/domain/shared/errors"
	"eventos-backend/internal/domain/shared/value_objects"
)

// Tenant representa uma organização no sistema multi-tenant
type Tenant struct {
	ID           value_objects.UUID
	ConfigID     *value_objects.UUID
	Name         string
	Identity     string
	IdentityType string
	Email        string
	Address      string
	Active       bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CreatedBy    *value_objects.UUID
	UpdatedBy    *value_objects.UUID
}

// NewTenant cria uma nova instância de Tenant
func NewTenant(name, identity, identityType, email, address string, createdBy value_objects.UUID) (*Tenant, error) {
	if err := validateTenantData(name, identity, identityType, email); err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	return &Tenant{
		ID:           value_objects.NewUUID(),
		Name:         name,
		Identity:     identity,
		IdentityType: identityType,
		Email:        email,
		Address:      address,
		Active:       true,
		CreatedAt:    now,
		UpdatedAt:    now,
		CreatedBy:    &createdBy,
		UpdatedBy:    &createdBy,
	}, nil
}

// Update atualiza os dados do tenant
func (t *Tenant) Update(name, identity, identityType, email, address string, updatedBy value_objects.UUID) error {
	if err := validateTenantData(name, identity, identityType, email); err != nil {
		return err
	}

	t.Name = name
	t.Identity = identity
	t.IdentityType = identityType
	t.Email = email
	t.Address = address
	t.UpdatedAt = time.Now().UTC()
	t.UpdatedBy = &updatedBy

	return nil
}

// Activate ativa o tenant
func (t *Tenant) Activate(updatedBy value_objects.UUID) {
	t.Active = true
	t.UpdatedAt = time.Now().UTC()
	t.UpdatedBy = &updatedBy
}

// Deactivate desativa o tenant
func (t *Tenant) Deactivate(updatedBy value_objects.UUID) {
	t.Active = false
	t.UpdatedAt = time.Now().UTC()
	t.UpdatedBy = &updatedBy
}

// SetConfig define a configuração do tenant
func (t *Tenant) SetConfig(configID value_objects.UUID, updatedBy value_objects.UUID) {
	t.ConfigID = &configID
	t.UpdatedAt = time.Now().UTC()
	t.UpdatedBy = &updatedBy
}

// IsActive verifica se o tenant está ativo
func (t *Tenant) IsActive() bool {
	return t.Active
}

// HasConfig verifica se o tenant tem configuração definida
func (t *Tenant) HasConfig() bool {
	return t.ConfigID != nil && !t.ConfigID.IsZero()
}

// validateTenantData valida os dados básicos do tenant
func validateTenantData(name, identity, identityType, email string) error {
	if name == "" {
		return errors.NewValidationError("name", "name is required")
	}

	if len(name) < 2 || len(name) > 255 {
		return errors.NewValidationError("name", "name must be between 2 and 255 characters")
	}

	if identity != "" {
		if len(identity) < 3 || len(identity) > 50 {
			return errors.NewValidationError("identity", "identity must be between 3 and 50 characters")
		}

		if !isValidIdentityType(identityType) {
			return errors.NewValidationError("identity_type", "invalid identity type")
		}
	}

	if email != "" {
		if !isValidEmail(email) {
			return errors.NewValidationError("email", "invalid email format")
		}
	}

	return nil
}

// isValidIdentityType verifica se o tipo de identidade é válido
func isValidIdentityType(identityType string) bool {
	validTypes := []string{
		constants.IdentityTypeCPF,
		constants.IdentityTypeCNPJ,
		constants.IdentityTypeRG,
		constants.IdentityTypeOther,
	}

	for _, validType := range validTypes {
		if identityType == validType {
			return true
		}
	}

	return false
}

// isValidEmail faz uma validação básica de email
func isValidEmail(email string) bool {
	// Validação básica - em produção usar uma biblioteca mais robusta
	if len(email) < 5 || len(email) > 255 {
		return false
	}

	// Deve conter @ e pelo menos um ponto após o @
	atCount := 0
	atIndex := -1
	for i, char := range email {
		if char == '@' {
			atCount++
			atIndex = i
		}
	}

	if atCount != 1 || atIndex == 0 || atIndex == len(email)-1 {
		return false
	}

	// Deve ter pelo menos um ponto após o @
	hasDotAfterAt := false
	for i := atIndex + 1; i < len(email); i++ {
		if email[i] == '.' && i < len(email)-1 {
			hasDotAfterAt = true
			break
		}
	}

	return hasDotAfterAt
}

// IsValidEmailImproved faz uma validação mais robusta de email
func IsValidEmailImproved(email string) bool {
	if len(email) < 5 || len(email) > 255 {
		return false
	}

	// Deve conter exatamente um @ e pelo menos um ponto após o @
	atIndex := -1

	for i, char := range email {
		if char == '@' {
			if atIndex != -1 {
				return false // Mais de um @
			}
			atIndex = i
		}
	}

	if atIndex == -1 || atIndex == 0 || atIndex == len(email)-1 {
		return false
	}

	// Deve ter pelo menos um ponto após o @
	hasDotAfterAt := false
	for i := atIndex + 1; i < len(email); i++ {
		if email[i] == '.' && i < len(email)-1 {
			// Deve haver pelo menos um caractere após o ponto
			hasDotAfterAt = true
			break
		}
	}

	return hasDotAfterAt
}
