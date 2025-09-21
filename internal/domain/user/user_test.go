package user

import (
	"testing"

	"eventos-backend/internal/domain/shared/value_objects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// UserTestSuite é a suíte de testes para User
type UserTestSuite struct {
	suite.Suite
}

func TestUserSuite(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}

func (suite *UserTestSuite) TestNewUser_ValidData() {
	// Arrange
	tenantID := value_objects.NewUUID()
	createdBy := value_objects.NewUUID()
	fullName := "João Silva"
	email := "joao.silva@example.com"
	phone := "+5511999999999"
	username := "joao_silva"
	password := "Teste123"

	// Act
	user, err := NewUser(tenantID, fullName, email, phone, username, password, createdBy)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)
	assert.NotEmpty(suite.T(), user.ID.String())
	assert.Equal(suite.T(), tenantID, user.TenantID)
	assert.Equal(suite.T(), fullName, user.FullName)
	assert.Equal(suite.T(), email, user.Email)
	assert.Equal(suite.T(), phone, user.Phone)
	assert.Equal(suite.T(), username, user.Username)
	assert.True(suite.T(), user.Active)
	assert.False(suite.T(), user.CreatedAt.IsZero())
	assert.False(suite.T(), user.UpdatedAt.IsZero())
	assert.Equal(suite.T(), user.CreatedAt, user.UpdatedAt)
	assert.NotNil(suite.T(), user.CreatedBy)
	assert.NotNil(suite.T(), user.UpdatedBy)
}

func (suite *UserTestSuite) TestNewUser_InvalidData() {
	testCases := []struct {
		name     string
		email    string
		username string
		password string
	}{
		{
			name:     "",
			email:    "test@example.com",
			username: "test_user",
			password: "Teste123",
		},
		{
			name:     "Test User",
			email:    "",
			username: "test_user",
			password: "Teste123",
		},
		{
			name:     "Test User",
			email:    "test@example.com",
			username: "",
			password: "Teste123",
		},
		{
			name:     "Test User",
			email:    "test@example.com",
			username: "test_user",
			password: "",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Arrange
			tenantID := value_objects.NewUUID()
			createdBy := value_objects.NewUUID()

			// Act
			user, err := NewUser(tenantID, tc.name, tc.email, "", tc.username, tc.password, createdBy)

			// Assert
			assert.Error(suite.T(), err)
			assert.Nil(suite.T(), user)
		})
	}
}

func (suite *UserTestSuite) TestUserUpdate_ValidData() {
	// Arrange
	tenantID := value_objects.NewUUID()
	createdBy := value_objects.NewUUID()
	user, _ := NewUser(tenantID, "João Silva", "joao@example.com", "+5511999999999", "joao_silva", "Teste123", createdBy)

	// Act
	updatedBy := value_objects.NewUUID()
	err := user.Update("João Silva Santos", "joao.santos@example.com", "+5511988888888", "joao.santos", updatedBy)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "João Silva Santos", user.FullName)
	assert.Equal(suite.T(), "joao.santos@example.com", user.Email)
	assert.Equal(suite.T(), "+5511988888888", user.Phone)
	assert.Equal(suite.T(), "joao.santos", user.Username)
	assert.NotNil(suite.T(), user.UpdatedBy)
	assert.False(suite.T(), user.UpdatedAt.IsZero())
	assert.True(suite.T(), user.UpdatedAt.After(user.CreatedAt))
}

func (suite *UserTestSuite) TestUserActivate() {
	// Arrange
	tenantID := value_objects.NewUUID()
	createdBy := value_objects.NewUUID()
	user, _ := NewUser(tenantID, "João Silva", "joao@example.com", "+5511999999999", "joao_silva", "Teste123", createdBy)
	user.Deactivate(createdBy) // Desativar primeiro

	// Act
	updatedBy := value_objects.NewUUID()
	user.Activate(updatedBy)

	// Assert
	assert.True(suite.T(), user.Active)
	assert.NotNil(suite.T(), user.UpdatedBy)
	assert.False(suite.T(), user.UpdatedAt.IsZero())
	assert.True(suite.T(), user.UpdatedAt.After(user.CreatedAt))
}

func (suite *UserTestSuite) TestUserDeactivate() {
	// Arrange
	tenantID := value_objects.NewUUID()
	createdBy := value_objects.NewUUID()
	user, _ := NewUser(tenantID, "João Silva", "joao@example.com", "+5511999999999", "joao_silva", "Teste123", createdBy)

	// Act
	updatedBy := value_objects.NewUUID()
	user.Deactivate(updatedBy)

	// Assert
	assert.False(suite.T(), user.Active)
	assert.NotNil(suite.T(), user.UpdatedBy)
	assert.False(suite.T(), user.UpdatedAt.IsZero())
	assert.True(suite.T(), user.UpdatedAt.After(user.CreatedAt))
}

func (suite *UserTestSuite) TestUserCheckPassword() {
	// Arrange
	tenantID := value_objects.NewUUID()
	createdBy := value_objects.NewUUID()
	password := "Teste123"
	user, _ := NewUser(tenantID, "João Silva", "joao@example.com", "+5511999999999", "joao.silva", password, createdBy)

	// Act & Assert
	assert.True(suite.T(), user.CheckPassword(password))
	assert.False(suite.T(), user.CheckPassword("WrongPassword"))
}

func (suite *UserTestSuite) TestUserBelongsToTenant() {
	// Arrange
	tenantID := value_objects.NewUUID()
	anotherTenantID := value_objects.NewUUID()
	createdBy := value_objects.NewUUID()
	user, _ := NewUser(tenantID, "João Silva", "joao@example.com", "+5511999999999", "joao_silva", "Teste123", createdBy)

	// Act & Assert
	assert.True(suite.T(), user.BelongsToTenant(tenantID))
	assert.False(suite.T(), user.BelongsToTenant(anotherTenantID))
}
