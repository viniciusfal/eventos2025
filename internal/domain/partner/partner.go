package partner

import (
	"time"

	"eventos-backend/internal/domain/shared/constants"
	"eventos-backend/internal/domain/shared/errors"
	"eventos-backend/internal/domain/shared/value_objects"

	"golang.org/x/crypto/bcrypt"
)

// Partner representa um parceiro no sistema
type Partner struct {
	ID                  value_objects.UUID
	TenantID            value_objects.UUID
	Name                string
	Email               string
	Email2              string
	Phone               string
	Phone2              string
	Identity            string
	IdentityType        string
	Location            string
	PasswordHash        string
	LastLogin           *time.Time
	FailedLoginAttempts int
	LockedUntil         *time.Time
	Active              bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
	CreatedBy           *value_objects.UUID
	UpdatedBy           *value_objects.UUID
}

// NewPartner cria uma nova instância de Partner
func NewPartner(tenantID value_objects.UUID, name, email, email2, phone, phone2, identity, identityType, location, password string, createdBy value_objects.UUID) (*Partner, error) {
	if err := validatePartnerData(name, email, identity, identityType, password); err != nil {
		return nil, err
	}

	// Hash da senha se fornecida
	var hashedPassword string
	if password != "" {
		hash, err := hashPassword(password)
		if err != nil {
			return nil, errors.NewInternalError("failed to hash password", err)
		}
		hashedPassword = hash
	}

	now := time.Now().UTC()

	return &Partner{
		ID:                  value_objects.NewUUID(),
		TenantID:            tenantID,
		Name:                name,
		Email:               email,
		Email2:              email2,
		Phone:               phone,
		Phone2:              phone2,
		Identity:            identity,
		IdentityType:        identityType,
		Location:            location,
		PasswordHash:        hashedPassword,
		FailedLoginAttempts: 0,
		Active:              true,
		CreatedAt:           now,
		UpdatedAt:           now,
		CreatedBy:           &createdBy,
		UpdatedBy:           &createdBy,
	}, nil
}

// Update atualiza os dados do parceiro
func (p *Partner) Update(name, email, email2, phone, phone2, identity, identityType, location string, updatedBy value_objects.UUID) error {
	if err := validatePartnerUpdateData(name, email, identity, identityType); err != nil {
		return err
	}

	p.Name = name
	p.Email = email
	p.Email2 = email2
	p.Phone = phone
	p.Phone2 = phone2
	p.Identity = identity
	p.IdentityType = identityType
	p.Location = location
	p.UpdatedAt = time.Now().UTC()
	p.UpdatedBy = &updatedBy

	return nil
}

// UpdatePassword atualiza a senha do parceiro
func (p *Partner) UpdatePassword(newPassword string, updatedBy value_objects.UUID) error {
	if err := validatePassword(newPassword); err != nil {
		return err
	}

	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		return errors.NewInternalError("failed to hash password", err)
	}

	p.PasswordHash = hashedPassword
	p.UpdatedAt = time.Now().UTC()
	p.UpdatedBy = &updatedBy

	return nil
}

// CheckPassword verifica se a senha fornecida está correta
func (p *Partner) CheckPassword(password string) bool {
	if p.PasswordHash == "" {
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(p.PasswordHash), []byte(password))
	return err == nil
}

// Activate ativa o parceiro
func (p *Partner) Activate(updatedBy value_objects.UUID) {
	p.Active = true
	p.UpdatedAt = time.Now().UTC()
	p.UpdatedBy = &updatedBy
}

// Deactivate desativa o parceiro
func (p *Partner) Deactivate(updatedBy value_objects.UUID) {
	p.Active = false
	p.UpdatedAt = time.Now().UTC()
	p.UpdatedBy = &updatedBy
}

// IsActive verifica se o parceiro está ativo
func (p *Partner) IsActive() bool {
	return p.Active
}

// IsLocked verifica se o parceiro está bloqueado
func (p *Partner) IsLocked() bool {
	if p.LockedUntil == nil {
		return false
	}
	return time.Now().UTC().Before(*p.LockedUntil)
}

// BelongsToTenant verifica se o parceiro pertence ao tenant informado
func (p *Partner) BelongsToTenant(tenantID value_objects.UUID) bool {
	return p.TenantID.Equals(tenantID)
}

// RecordFailedLogin registra uma tentativa de login falhada
func (p *Partner) RecordFailedLogin() {
	p.FailedLoginAttempts++

	// Bloquear após 5 tentativas falhadas
	if p.FailedLoginAttempts >= 5 {
		lockUntil := time.Now().UTC().Add(30 * time.Minute)
		p.LockedUntil = &lockUntil
	}

	p.UpdatedAt = time.Now().UTC()
}

// RecordSuccessfulLogin registra um login bem-sucedido
func (p *Partner) RecordSuccessfulLogin() {
	now := time.Now().UTC()
	p.LastLogin = &now
	p.FailedLoginAttempts = 0
	p.LockedUntil = nil
	p.UpdatedAt = now
}

// UnlockAccount desbloqueia a conta do parceiro
func (p *Partner) UnlockAccount(updatedBy value_objects.UUID) {
	p.FailedLoginAttempts = 0
	p.LockedUntil = nil
	p.UpdatedAt = time.Now().UTC()
	p.UpdatedBy = &updatedBy
}

// HasPassword verifica se o parceiro tem senha definida
func (p *Partner) HasPassword() bool {
	return p.PasswordHash != ""
}

// GetPrimaryEmail retorna o email principal do parceiro
func (p *Partner) GetPrimaryEmail() string {
	return p.Email
}

// GetSecondaryEmail retorna o email secundário do parceiro
func (p *Partner) GetSecondaryEmail() string {
	return p.Email2
}

// GetPrimaryPhone retorna o telefone principal do parceiro
func (p *Partner) GetPrimaryPhone() string {
	return p.Phone
}

// GetSecondaryPhone retorna o telefone secundário do parceiro
func (p *Partner) GetSecondaryPhone() string {
	return p.Phone2
}

// validatePartnerData valida os dados básicos do parceiro
func validatePartnerData(name, email, identity, identityType, password string) error {
	if err := validatePartnerUpdateData(name, email, identity, identityType); err != nil {
		return err
	}

	// Senha é opcional para parceiros
	if password != "" {
		return validatePassword(password)
	}

	return nil
}

// validatePartnerUpdateData valida os dados de atualização do parceiro
func validatePartnerUpdateData(name, email, identity, identityType string) error {
	if name == "" {
		return errors.NewValidationError("name", "partner name is required")
	}

	if len(name) < 2 || len(name) > 255 {
		return errors.NewValidationError("name", "partner name must be between 2 and 255 characters")
	}

	if email != "" {
		if !isValidEmail(email) {
			return errors.NewValidationError("email", "invalid email format")
		}
	}

	if identity != "" {
		if len(identity) < 3 || len(identity) > 50 {
			return errors.NewValidationError("identity", "identity must be between 3 and 50 characters")
		}

		if !isValidIdentityType(identityType) {
			return errors.NewValidationError("identity_type", "invalid identity type")
		}
	}

	return nil
}

// validatePassword valida a senha
func validatePassword(password string) error {
	if password == "" {
		return errors.NewValidationError("password", "password is required")
	}

	if len(password) < 8 {
		return errors.NewValidationError("password", "password must be at least 8 characters long")
	}

	if len(password) > 128 {
		return errors.NewValidationError("password", "password must be at most 128 characters long")
	}

	// Verificar se contém pelo menos uma letra e um número
	hasLetter := false
	hasNumber := false

	for _, char := range password {
		if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' {
			hasLetter = true
		}
		if char >= '0' && char <= '9' {
			hasNumber = true
		}
		if hasLetter && hasNumber {
			break
		}
	}

	if !hasLetter {
		return errors.NewValidationError("password", "password must contain at least one letter")
	}

	if !hasNumber {
		return errors.NewValidationError("password", "password must contain at least one number")
	}

	return nil
}

// hashPassword gera o hash da senha
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// isValidEmail faz uma validação básica de email
func isValidEmail(email string) bool {
	if len(email) < 5 || len(email) > 255 {
		return false
	}

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
