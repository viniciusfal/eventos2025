package requests

import (
	. "eventos-backend/internal/application/dto/requests"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// AuthRequestsTestSuite é a suíte de testes para DTOs de autenticação
type AuthRequestsTestSuite struct {
	suite.Suite
}

func TestAuthRequestsSuite(t *testing.T) {
	suite.Run(t, new(AuthRequestsTestSuite))
}

func (suite *AuthRequestsTestSuite) TestLoginRequest_ValidData() {
	// Arrange
	request := &LoginRequest{
		UsernameOrEmail: "user@example.com",
		Password:        "SecurePass123",
		TenantID:        "550e8400-e29b-41d4-a716-446655440000",
	}

	// Act
	err := request.Validate()

	// Assert
	assert.NoError(suite.T(), err)
}

func (suite *AuthRequestsTestSuite) TestLoginRequest_InvalidData() {
	testCases := []struct {
		name            string
		usernameOrEmail string
		password        string
		expectedErr     string
	}{
		{
			name:            "EmptyUsernameOrEmail",
			usernameOrEmail: "",
			password:        "SecurePass123",
			expectedErr:     "username or email is required",
		},
		{
			name:            "UsernameTooShort",
			usernameOrEmail: "us",
			password:        "SecurePass123",
			expectedErr:     "username or email must be at least 3 characters",
		},
		{
			name:            "EmptyPassword",
			usernameOrEmail: "user@example.com",
			password:        "",
			expectedErr:     "password is required",
		},
		{
			name:            "PasswordTooShort",
			usernameOrEmail: "user@example.com",
			password:        "123",
			expectedErr:     "password must be at least 8 characters",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Arrange
			request := &LoginRequest{
				UsernameOrEmail: tc.usernameOrEmail,
				Password:        tc.password,
			}

			// Act
			err := request.Validate()

			// Assert
			assert.Error(suite.T(), err)
			assert.Contains(suite.T(), err.Error(), tc.expectedErr)
		})
	}
}

func (suite *AuthRequestsTestSuite) TestRefreshTokenRequest_ValidData() {
	// Arrange
	request := &RefreshTokenRequest{
		RefreshToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
	}

	// Act
	err := request.Validate()

	// Assert
	assert.NoError(suite.T(), err)
}

func (suite *AuthRequestsTestSuite) TestRefreshTokenRequest_InvalidData() {
	// Arrange
	request := &RefreshTokenRequest{
		RefreshToken: "",
	}

	// Act
	err := request.Validate()

	// Assert
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "refresh token is required")
}

func (suite *AuthRequestsTestSuite) TestChangePasswordRequest_ValidData() {
	// Arrange
	request := &ChangePasswordRequest{
		CurrentPassword: "OldPass123",
		NewPassword:     "NewPass456",
	}

	// Act
	err := request.Validate()

	// Assert
	assert.NoError(suite.T(), err)
}

func (suite *AuthRequestsTestSuite) TestChangePasswordRequest_InvalidData() {
	testCases := []struct {
		name            string
		currentPassword string
		newPassword     string
		expectedErr     string
	}{
		{
			name:            "EmptyCurrentPassword",
			currentPassword: "",
			newPassword:     "NewPass456",
			expectedErr:     "current password is required",
		},
		{
			name:            "EmptyNewPassword",
			currentPassword: "OldPass123",
			newPassword:     "",
			expectedErr:     "new password is required",
		},
		{
			name:            "NewPasswordTooShort",
			currentPassword: "OldPass123",
			newPassword:     "123",
			expectedErr:     "new password must be at least 8 characters",
		},
		{
			name:            "SamePasswords",
			currentPassword: "SamePass123",
			newPassword:     "SamePass123",
			expectedErr:     "new password must be different from current password",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Arrange
			request := &ChangePasswordRequest{
				CurrentPassword: tc.currentPassword,
				NewPassword:     tc.newPassword,
			}

			// Act
			err := request.Validate()

			// Assert
			assert.Error(suite.T(), err)
			assert.Contains(suite.T(), err.Error(), tc.expectedErr)
		})
	}
}
