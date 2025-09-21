package responses

import (
	. "eventos-backend/internal/application/dto/responses"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// AuthResponsesTestSuite é a suíte de testes para DTOs de resposta de autenticação
type AuthResponsesTestSuite struct {
	suite.Suite
}

func TestAuthResponsesSuite(t *testing.T) {
	suite.Run(t, new(AuthResponsesTestSuite))
}

func (suite *AuthResponsesTestSuite) TestLoginResponse_CompleteData() {
	// Arrange
	now := time.Now()
	user := UserResponse{
		ID:        "550e8400-e29b-41d4-a716-446655440000",
		TenantID:  "550e8400-e29b-41d4-a716-446655440001",
		FullName:  "João Silva",
		Email:     "joao@example.com",
		Username:  "joao_silva",
		Phone:     "+5511999999999",
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Act
	response := LoginResponse{
		AccessToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
		RefreshToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		User:         user,
	}

	// Assert
	assert.Equal(suite.T(), "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...", response.AccessToken)
	assert.Equal(suite.T(), "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...", response.RefreshToken)
	assert.Equal(suite.T(), "Bearer", response.TokenType)
	assert.Equal(suite.T(), 3600, response.ExpiresIn)
	assert.Equal(suite.T(), user, response.User)
}

func (suite *AuthResponsesTestSuite) TestRefreshTokenResponse_CompleteData() {
	// Act
	response := RefreshTokenResponse{
		AccessToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
		RefreshToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
	}

	// Assert
	assert.Equal(suite.T(), "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...", response.AccessToken)
	assert.Equal(suite.T(), "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...", response.RefreshToken)
	assert.Equal(suite.T(), "Bearer", response.TokenType)
	assert.Equal(suite.T(), 3600, response.ExpiresIn)
}

func (suite *AuthResponsesTestSuite) TestUserResponse_CompleteData() {
	// Arrange
	now := time.Now()

	// Act
	user := UserResponse{
		ID:        "550e8400-e29b-41d4-a716-446655440000",
		TenantID:  "550e8400-e29b-41d4-a716-446655440001",
		FullName:  "João Silva",
		Email:     "joao@example.com",
		Username:  "joao_silva",
		Phone:     "+5511999999999",
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Assert
	assert.Equal(suite.T(), "550e8400-e29b-41d4-a716-446655440000", user.ID)
	assert.Equal(suite.T(), "550e8400-e29b-41d4-a716-446655440001", user.TenantID)
	assert.Equal(suite.T(), "João Silva", user.FullName)
	assert.Equal(suite.T(), "joao@example.com", user.Email)
	assert.Equal(suite.T(), "joao_silva", user.Username)
	assert.Equal(suite.T(), "+5511999999999", user.Phone)
	assert.True(suite.T(), user.Active)
	assert.Equal(suite.T(), now, user.CreatedAt)
	assert.Equal(suite.T(), now, user.UpdatedAt)
}

func (suite *AuthResponsesTestSuite) TestTenantResponse_CompleteData() {
	// Arrange
	now := time.Now()

	// Act
	tenant := TenantResponse{
		ID:           "550e8400-e29b-41d4-a716-446655440000",
		Name:         "Empresa Teste Ltda",
		Identity:     "12345678000123",
		IdentityType: "cnpj",
		Email:        "contato@empresa.com",
		Address:      "Rua das Flores, 123",
		Active:       true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Assert
	assert.Equal(suite.T(), "550e8400-e29b-41d4-a716-446655440000", tenant.ID)
	assert.Equal(suite.T(), "Empresa Teste Ltda", tenant.Name)
	assert.Equal(suite.T(), "12345678000123", tenant.Identity)
	assert.Equal(suite.T(), "cnpj", tenant.IdentityType)
	assert.Equal(suite.T(), "contato@empresa.com", tenant.Email)
	assert.Equal(suite.T(), "Rua das Flores, 123", tenant.Address)
	assert.True(suite.T(), tenant.Active)
	assert.Equal(suite.T(), now, tenant.CreatedAt)
	assert.Equal(suite.T(), now, tenant.UpdatedAt)
}

func (suite *AuthResponsesTestSuite) TestErrorResponse_CompleteData() {
	// Act
	errorResponse := ErrorResponse{
		Error:   "VALIDATION_ERROR",
		Message: "Dados inválidos fornecidos",
		Details: map[string]interface{}{
			"field": "email",
			"issue": "invalid format",
		},
	}

	// Assert
	assert.Equal(suite.T(), "VALIDATION_ERROR", errorResponse.Error)
	assert.Equal(suite.T(), "Dados inválidos fornecidos", errorResponse.Message)
	assert.Equal(suite.T(), "invalid format", errorResponse.Details["issue"])
}

func (suite *AuthResponsesTestSuite) TestSuccessResponse_CompleteData() {
	// Act
	successResponse := SuccessResponse{
		Message: "Operação realizada com sucesso",
		Data: map[string]interface{}{
			"id":    "550e8400-e29b-41d4-a716-446655440000",
			"email": "user@example.com",
		},
	}

	// Assert
	assert.Equal(suite.T(), "Operação realizada com sucesso", successResponse.Message)
	assert.Equal(suite.T(), "550e8400-e29b-41d4-a716-446655440000", successResponse.Data.(map[string]interface{})["id"])
}

func (suite *AuthResponsesTestSuite) TestPaginatedResponse_CompleteData() {
	// Arrange
	data := []map[string]interface{}{
		{"id": "1", "name": "Item 1"},
		{"id": "2", "name": "Item 2"},
	}

	pagination := PaginationInfo{
		Page:       1,
		PageSize:   20,
		Total:      50,
		TotalPages: 3,
	}

	// Act
	response := PaginatedResponse{
		Data:       data,
		Pagination: pagination,
	}

	// Assert
	assert.Equal(suite.T(), data, response.Data)
	assert.Equal(suite.T(), 1, response.Pagination.Page)
	assert.Equal(suite.T(), 20, response.Pagination.PageSize)
	assert.Equal(suite.T(), 50, response.Pagination.Total)
	assert.Equal(suite.T(), 3, response.Pagination.TotalPages)
}
