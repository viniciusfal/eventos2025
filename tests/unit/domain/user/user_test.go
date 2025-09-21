package user

import (
	. "eventos-backend/internal/domain/user"
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
	email := "joao@example.com"
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
	assert.NotEmpty(suite.T(), user.Password)           // Deve ter hash da senha
	assert.NotEqual(suite.T(), password, user.Password) // Hash deve ser diferente da senha original
	assert.True(suite.T(), user.Active)
	assert.False(suite.T(), user.CreatedAt.IsZero())
	assert.False(suite.T(), user.UpdatedAt.IsZero())
	assert.Equal(suite.T(), user.CreatedAt, user.UpdatedAt)
	assert.NotNil(suite.T(), user.CreatedBy)
	assert.NotNil(suite.T(), user.UpdatedBy)
}

func (suite *UserTestSuite) TestNewUser_InvalidData() {
	// Arrange
	tenantID := value_objects.NewUUID()
	createdBy := value_objects.NewUUID()

	testCases := []struct {
		name     string
		fullName string
		email    string
		username string
		password string
	}{
		{"empty name", "", "test@example.com", "test_user", "Teste123"},
		{"invalid email", "Test User", "invalid-email", "test_user", "Teste123"},
		{"invalid username", "Test User", "test@example.com", "test.user", "Teste123"}, // ponto não é permitido
		{"weak password", "Test User", "test@example.com", "test_user", "123"},         // muito curta
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Act
			user, err := NewUser(tenantID, tc.fullName, tc.email, "", tc.username, tc.password, createdBy)

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
	err := user.Update("João Silva Santos", "joao_santos@example.com", "+5511988888888", "joao_santos", updatedBy)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "João Silva Santos", user.FullName)
	assert.Equal(suite.T(), "joao_santos@example.com", user.Email)
	assert.Equal(suite.T(), "+5511988888888", user.Phone)
	assert.Equal(suite.T(), "joao_santos", user.Username)
	assert.NotNil(suite.T(), user.UpdatedBy)
	assert.False(suite.T(), user.UpdatedAt.IsZero())
	assert.True(suite.T(), user.UpdatedAt.After(user.CreatedAt) || user.UpdatedAt.Equal(user.CreatedAt))
}

func (suite *UserTestSuite) TestUserActivate() {
	// Arrange
	tenantID := value_objects.NewUUID()
	createdBy := value_objects.NewUUID()
	user, _ := NewUser(tenantID, "Test User", "test@example.com", "", "test_user", "Teste123", createdBy)

	// Desativar primeiro
	user.Deactivate(createdBy)
	assert.False(suite.T(), user.Active)

	// Act
	updatedBy := value_objects.NewUUID()
	user.Activate(updatedBy)

	// Assert
	assert.True(suite.T(), user.IsActive())
	assert.NotNil(suite.T(), user.UpdatedBy)
	assert.False(suite.T(), user.UpdatedAt.IsZero())
	assert.True(suite.T(), user.UpdatedAt.After(user.CreatedAt) || user.UpdatedAt.Equal(user.CreatedAt))
}

func (suite *UserTestSuite) TestUserDeactivate() {
	// Arrange
	tenantID := value_objects.NewUUID()
	createdBy := value_objects.NewUUID()
	user, _ := NewUser(tenantID, "Test User", "test@example.com", "", "test_user", "Teste123", createdBy)

	// Verificar estado inicial
	assert.True(suite.T(), user.Active)

	// Act
	updatedBy := value_objects.NewUUID()
	user.Deactivate(updatedBy)

	// Assert
	assert.False(suite.T(), user.IsActive())
	assert.NotNil(suite.T(), user.UpdatedBy)
	assert.False(suite.T(), user.UpdatedAt.IsZero())
	assert.True(suite.T(), user.UpdatedAt.After(user.CreatedAt) || user.UpdatedAt.Equal(user.CreatedAt))
}

func (suite *UserTestSuite) TestUserCheckPassword() {
	// Arrange
	tenantID := value_objects.NewUUID()
	createdBy := value_objects.NewUUID()
	password := "Teste123"
	user, _ := NewUser(tenantID, "Test User", "test@example.com", "", "test_user", password, createdBy)

	// Act & Assert
	assert.True(suite.T(), user.CheckPassword(password))
	assert.False(suite.T(), user.CheckPassword("wrong_password"))
}

func (suite *UserTestSuite) TestUserIsActive() {
	// Arrange
	tenantID := value_objects.NewUUID()
	createdBy := value_objects.NewUUID()
	user, _ := NewUser(tenantID, "Test User", "test@example.com", "", "test_user", "Teste123", createdBy)

	// Assert
	assert.True(suite.T(), user.IsActive())

	// Desativar e verificar novamente
	user.Deactivate(createdBy)
	assert.False(suite.T(), user.IsActive())
}

func (suite *UserTestSuite) TestUserBelongsToTenant() {
	// Arrange
	tenantID := value_objects.NewUUID()
	otherTenantID := value_objects.NewUUID()
	createdBy := value_objects.NewUUID()
	user, _ := NewUser(tenantID, "Test User", "test@example.com", "", "test_user", "Teste123", createdBy)

	// Act & Assert
	assert.True(suite.T(), user.BelongsToTenant(tenantID))
	assert.False(suite.T(), user.BelongsToTenant(otherTenantID))
}

func (suite *UserTestSuite) TestUserUpdatePassword() {
	// Arrange
	tenantID := value_objects.NewUUID()
	createdBy := value_objects.NewUUID()
	originalPassword := "Teste123"
	newPassword := "NovaSenh4"
	user, _ := NewUser(tenantID, "Test User", "test@example.com", "", "test_user", originalPassword, createdBy)
	originalPasswordHash := user.Password

	// Act
	updatedBy := value_objects.NewUUID()
	err := user.UpdatePassword(newPassword, updatedBy)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotEqual(suite.T(), originalPasswordHash, user.Password) // Hash deve ter mudado
	assert.True(suite.T(), user.CheckPassword(newPassword))
	assert.False(suite.T(), user.CheckPassword(originalPassword))
	assert.NotNil(suite.T(), user.UpdatedBy)
	assert.False(suite.T(), user.UpdatedAt.IsZero())
	assert.True(suite.T(), user.UpdatedAt.After(user.CreatedAt) || user.UpdatedAt.Equal(user.CreatedAt))
}
