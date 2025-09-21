package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"eventos-backend/internal/application/dto/requests"
	"eventos-backend/internal/application/dto/responses"
	"eventos-backend/internal/domain/shared/value_objects"
	"eventos-backend/internal/domain/user"
	jwtService "eventos-backend/internal/infrastructure/auth/jwt"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

// MockUserService é um mock do serviço de usuário
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, tenantID value_objects.UUID, fullName, email, phone, username, password string, createdBy value_objects.UUID) (*user.User, error) {
	args := m.Called(ctx, tenantID, fullName, email, phone, username, password, createdBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserService) GetUserByID(ctx context.Context, id value_objects.UUID) (*user.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserService) GetUserByUsernameOrEmail(ctx context.Context, usernameOrEmail string, tenantID *value_objects.UUID) (*user.User, error) {
	args := m.Called(ctx, usernameOrEmail, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserService) AuthenticateUser(ctx context.Context, usernameOrEmail, password string, tenantID *value_objects.UUID) (*user.User, error) {
	args := m.Called(ctx, usernameOrEmail, password, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, id value_objects.UUID, fullName, email, phone, username string, updatedBy value_objects.UUID) (*user.User, error) {
	args := m.Called(ctx, id, fullName, email, phone, username, updatedBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserService) DeleteUser(ctx context.Context, id value_objects.UUID, deletedBy value_objects.UUID) error {
	args := m.Called(ctx, id, deletedBy)
	return args.Error(0)
}

func (m *MockUserService) ActivateUser(ctx context.Context, id value_objects.UUID, updatedBy value_objects.UUID) error {
	args := m.Called(ctx, id, updatedBy)
	return args.Error(0)
}

func (m *MockUserService) DeactivateUser(ctx context.Context, id value_objects.UUID, updatedBy value_objects.UUID) error {
	args := m.Called(ctx, id, updatedBy)
	return args.Error(0)
}

func (m *MockUserService) GetUser(ctx context.Context, id value_objects.UUID) (*user.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserService) GetUserByUsername(ctx context.Context, username string) (*user.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserService) GetUserByUsernameAndTenant(ctx context.Context, username string, tenantID value_objects.UUID) (*user.User, error) {
	args := m.Called(ctx, username, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserService) UpdateUserPassword(ctx context.Context, id value_objects.UUID, newPassword string, updatedBy value_objects.UUID) error {
	args := m.Called(ctx, id, newPassword, updatedBy)
	return args.Error(0)
}

func (m *MockUserService) ListUsersByTenant(ctx context.Context, tenantID value_objects.UUID, filters user.ListFilters) ([]*user.User, int, error) {
	args := m.Called(ctx, tenantID, filters)
	return args.Get(0).([]*user.User), args.Int(1), args.Error(2)
}

func (m *MockUserService) ListUsers(ctx context.Context, filters user.ListFilters) ([]*user.User, int, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]*user.User), args.Int(1), args.Error(2)
}

// MockJWTService é um mock do serviço JWT
type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateToken(userID, tenantID value_objects.UUID, username, email string) (string, error) {
	args := m.Called(userID, tenantID, username, email)
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) GenerateRefreshToken(userID, tenantID value_objects.UUID, username, email string) (string, error) {
	args := m.Called(userID, tenantID, username, email)
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) RefreshToken(refreshTokenString string) (string, string, error) {
	args := m.Called(refreshTokenString)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockJWTService) GetUserIDFromToken(tokenString string) (value_objects.UUID, error) {
	args := m.Called(tokenString)
	return args.Get(0).(value_objects.UUID), args.Error(1)
}

func (m *MockJWTService) ValidateToken(tokenString string) (*jwtService.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwtService.Claims), args.Error(1)
}

func (m *MockJWTService) ValidateRefreshToken(tokenString string) (*jwtService.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwtService.Claims), args.Error(1)
}

// AuthHandlerTestSuite é a suíte de testes para AuthHandler
type AuthHandlerTestSuite struct {
	suite.Suite
	router         *gin.Engine
	mockUserSvc    *MockUserService
	mockJWTService *MockJWTService
	logger         *zap.Logger
	handler        *AuthHandler
}

func TestAuthHandlerSuite(t *testing.T) {
	suite.Run(t, new(AuthHandlerTestSuite))
}

func (suite *AuthHandlerTestSuite) SetupTest() {
	// Configurar Gin em modo teste
	gin.SetMode(gin.TestMode)

	// Criar mocks
	suite.mockUserSvc = new(MockUserService)
	suite.mockJWTService = new(MockJWTService)

	// Criar logger mock (simplificado para testes)
	suite.logger = zap.NewNop()

	// Criar handler
	suite.handler = NewAuthHandler(suite.mockUserSvc, suite.mockJWTService, suite.logger)

	// Configurar router
	suite.router = gin.New()
	suite.router.POST("/auth/login", suite.handler.Login)
	suite.router.POST("/auth/refresh", suite.handler.RefreshToken)
}

func (suite *AuthHandlerTestSuite) TearDownTest() {
	suite.mockUserSvc.AssertExpectations(suite.T())
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *AuthHandlerTestSuite) TestLogin_Success() {
	// Arrange
	tenantID := value_objects.NewUUID()
	userID := value_objects.NewUUID()
	username := "test_user"
	email := "test@example.com"
	accessToken := "access_token_123"
	refreshToken := "refresh_token_456"

	loginReq := requests.LoginRequest{
		UsernameOrEmail: email,
		Password:        "password123",
		TenantID:        tenantID.String(),
	}

	expectedUser := &user.User{
		ID:       userID,
		TenantID: tenantID,
		Username: username,
		Email:    email,
		FullName: "Test User",
		Phone:    "+5511999999999",
		Active:   true,
	}

	// Configurar mocks
	suite.mockUserSvc.On("AuthenticateUser",
		mock.Anything, // context
		email,
		"password123",
		&tenantID,
	).Return(expectedUser, nil)

	suite.mockJWTService.On("GenerateToken",
		userID,
		tenantID,
		username,
		email,
	).Return(accessToken, nil)

	suite.mockJWTService.On("GenerateRefreshToken",
		userID,
		tenantID,
		username,
		email,
	).Return(refreshToken, nil)

	// Act
	reqBody, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response responses.LoginResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), accessToken, response.AccessToken)
	assert.Equal(suite.T(), refreshToken, response.RefreshToken)
	assert.Equal(suite.T(), "Bearer", response.TokenType)
	assert.Equal(suite.T(), 3600, response.ExpiresIn)
	assert.Equal(suite.T(), userID.String(), response.User.ID)
	assert.Equal(suite.T(), email, response.User.Email)
}

func (suite *AuthHandlerTestSuite) TestLogin_InvalidJSON() {
	// Arrange
	invalidJSON := `{"invalid": json}`

	// Act
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Invalid request format", response["error"])
}

func (suite *AuthHandlerTestSuite) TestLogin_InvalidTenantID() {
	// Arrange
	loginReq := requests.LoginRequest{
		UsernameOrEmail: "test@example.com",
		Password:        "password123",
		TenantID:        "invalid-uuid",
	}

	// Act
	reqBody, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Invalid tenant ID format", response["error"])
}

func (suite *AuthHandlerTestSuite) TestLogin_AuthenticationFailed() {
	// Arrange
	tenantID := value_objects.NewUUID()

	loginReq := requests.LoginRequest{
		UsernameOrEmail: "test@example.com",
		Password:        "wrongpassword",
		TenantID:        tenantID.String(),
	}

	suite.mockUserSvc.On("AuthenticateUser",
		mock.Anything,
		"test@example.com",
		"wrongpassword",
		&tenantID,
	).Return(nil, errors.New("invalid credentials"))

	// Act
	reqBody, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Invalid credentials", response["error"])
}

func (suite *AuthHandlerTestSuite) TestLogin_TokenGenerationFailed() {
	// Arrange
	tenantID := value_objects.NewUUID()
	userID := value_objects.NewUUID()
	username := "test_user"
	email := "test@example.com"

	loginReq := requests.LoginRequest{
		UsernameOrEmail: email,
		Password:        "password123",
		TenantID:        tenantID.String(),
	}

	expectedUser := &user.User{
		ID:       userID,
		TenantID: tenantID,
		Username: username,
		Email:    email,
		FullName: "Test User",
		Active:   true,
	}

	// Configurar mocks
	suite.mockUserSvc.On("AuthenticateUser",
		mock.Anything,
		email,
		"password123",
		&tenantID,
	).Return(expectedUser, nil)

	suite.mockJWTService.On("GenerateToken",
		userID,
		tenantID,
		username,
		email,
	).Return("", errors.New("token generation failed"))

	// Act
	reqBody, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Failed to generate token", response["error"])
}
