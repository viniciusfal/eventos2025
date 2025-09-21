package user

import (
	"time"

	"eventos-backend/internal/domain/shared/errors"
	"eventos-backend/internal/domain/shared/value_objects"

	"golang.org/x/crypto/bcrypt"
)

// User representa um usuário do sistema
type User struct {
	ID        value_objects.UUID
	TenantID  value_objects.UUID
	FullName  string
	Email     string
	Phone     string
	Username  string
	Password  string // Hash da senha
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy *value_objects.UUID
	UpdatedBy *value_objects.UUID
}

// NewUser cria uma nova instância de User
func NewUser(tenantID value_objects.UUID, fullName, email, phone, username, password string, createdBy value_objects.UUID) (*User, error) {
	if err := validateUserData(fullName, email, username, password); err != nil {
		return nil, err
	}

	// Hash da senha
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return nil, errors.NewInternalError("failed to hash password", err)
	}

	now := time.Now().UTC()

	return &User{
		ID:        value_objects.NewUUID(),
		TenantID:  tenantID,
		FullName:  fullName,
		Email:     email,
		Phone:     phone,
		Username:  username,
		Password:  hashedPassword,
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
		CreatedBy: &createdBy,
		UpdatedBy: &createdBy,
	}, nil
}

// Update atualiza os dados do usuário
func (u *User) Update(fullName, email, phone, username string, updatedBy value_objects.UUID) error {
	if err := validateUserUpdateData(fullName, email, username); err != nil {
		return err
	}

	u.FullName = fullName
	u.Email = email
	u.Phone = phone
	u.Username = username
	u.UpdatedAt = time.Now().UTC()
	u.UpdatedBy = &updatedBy

	return nil
}

// UpdatePassword atualiza a senha do usuário
func (u *User) UpdatePassword(newPassword string, updatedBy value_objects.UUID) error {
	if err := validatePassword(newPassword); err != nil {
		return err
	}

	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		return errors.NewInternalError("failed to hash password", err)
	}

	u.Password = hashedPassword
	u.UpdatedAt = time.Now().UTC()
	u.UpdatedBy = &updatedBy

	return nil
}

// CheckPassword verifica se a senha fornecida está correta
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// Activate ativa o usuário
func (u *User) Activate(updatedBy value_objects.UUID) {
	u.Active = true
	u.UpdatedAt = time.Now().UTC()
	u.UpdatedBy = &updatedBy
}

// Deactivate desativa o usuário
func (u *User) Deactivate(updatedBy value_objects.UUID) {
	u.Active = false
	u.UpdatedAt = time.Now().UTC()
	u.UpdatedBy = &updatedBy
}

// IsActive verifica se o usuário está ativo
func (u *User) IsActive() bool {
	return u.Active
}

// BelongsToTenant verifica se o usuário pertence ao tenant informado
func (u *User) BelongsToTenant(tenantID value_objects.UUID) bool {
	return u.TenantID.Equals(tenantID)
}

// validateUserData valida os dados básicos do usuário
func validateUserData(fullName, email, username, password string) error {
	if err := validateUserUpdateData(fullName, email, username); err != nil {
		return err
	}

	return validatePassword(password)
}

// validateUserUpdateData valida os dados de atualização do usuário
func validateUserUpdateData(fullName, email, username string) error {
	if fullName == "" {
		return errors.NewValidationError("full_name", "full name is required")
	}

	if len(fullName) < 2 || len(fullName) > 255 {
		return errors.NewValidationError("full_name", "full name must be between 2 and 255 characters")
	}

	if email != "" {
		if !isValidEmail(email) {
			return errors.NewValidationError("email", "invalid email format")
		}
	}

	if username == "" {
		return errors.NewValidationError("username", "username is required")
	}

	if len(username) < 3 || len(username) > 50 {
		return errors.NewValidationError("username", "username must be between 3 and 50 characters")
	}

	// Validar formato do username (apenas letras, números e underscore)
	if !isValidUsername(username) {
		return errors.NewValidationError("username", "username can only contain letters, numbers and underscore")
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

// isValidUsername valida o formato do username
func isValidUsername(username string) bool {
	for _, char := range username {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_') {
			return false
		}
	}
	return true
}
