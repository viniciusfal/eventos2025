package requests

import (
	"eventos-backend/internal/domain/shared/errors"
)

// LoginRequest representa uma requisição de login
type LoginRequest struct {
	UsernameOrEmail string `json:"username_or_email" binding:"required"`
	Password        string `json:"password" binding:"required"`
	TenantID        string `json:"tenant_id,omitempty"`
}

// Validate valida os dados da requisição de login
func (r *LoginRequest) Validate() error {
	if r.UsernameOrEmail == "" {
		return errors.NewValidationError("username_or_email", "username or email is required")
	}

	if len(r.UsernameOrEmail) < 3 {
		return errors.NewValidationError("username_or_email", "username or email must be at least 3 characters")
	}

	if r.Password == "" {
		return errors.NewValidationError("password", "password is required")
	}

	if len(r.Password) < 8 {
		return errors.NewValidationError("password", "password must be at least 8 characters")
	}

	return nil
}

// RefreshTokenRequest representa uma requisição de refresh token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Validate valida os dados da requisição de refresh token
func (r *RefreshTokenRequest) Validate() error {
	if r.RefreshToken == "" {
		return errors.NewValidationError("refresh_token", "refresh token is required")
	}

	return nil
}

// ChangePasswordRequest representa uma requisição de alteração de senha
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
}

// Validate valida os dados da requisição de alteração de senha
func (r *ChangePasswordRequest) Validate() error {
	if r.CurrentPassword == "" {
		return errors.NewValidationError("current_password", "current password is required")
	}

	if r.NewPassword == "" {
		return errors.NewValidationError("new_password", "new password is required")
	}

	if len(r.NewPassword) < 8 {
		return errors.NewValidationError("new_password", "new password must be at least 8 characters")
	}

	if r.CurrentPassword == r.NewPassword {
		return errors.NewValidationError("new_password", "new password must be different from current password")
	}

	return nil
}
