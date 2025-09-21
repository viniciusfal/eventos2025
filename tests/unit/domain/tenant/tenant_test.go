package tenant

import (
	. "eventos-backend/internal/domain/tenant"
	"testing"

	"eventos-backend/internal/domain/shared/value_objects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// TenantTestSuite é a suíte de testes para Tenant
type TenantTestSuite struct {
	suite.Suite
}

// SetupTest é executado antes de cada teste
func (suite *TenantTestSuite) SetupTest() {
	// Setup comum para todos os testes
}

// TearDownTest é executado após cada teste
func (suite *TenantTestSuite) TearDownTest() {
	// Cleanup comum para todos os testes
}

func TestTenantSuite(t *testing.T) {
	suite.Run(t, new(TenantTestSuite))
}

func (suite *TenantTestSuite) TestNewTenant_ValidData() {
	// Arrange
	tenantID := value_objects.NewUUID()
	name := "Empresa Teste"
	identity := "12345678000123"
	identityType := "cnpj"
	email := "teste@empresa.com"
	address := "Rua Teste, 123"

	// Act
	tenant, err := NewTenant(name, identity, identityType, email, address, tenantID)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), tenant)
	assert.NotEmpty(suite.T(), tenant.ID.String())
	assert.Equal(suite.T(), tenantID, *tenant.CreatedBy)
	assert.Equal(suite.T(), tenantID, *tenant.UpdatedBy)
	assert.Equal(suite.T(), name, tenant.Name)
	assert.Equal(suite.T(), identity, tenant.Identity)
	assert.Equal(suite.T(), identityType, tenant.IdentityType)
	assert.Equal(suite.T(), email, tenant.Email)
	assert.Equal(suite.T(), address, tenant.Address)
	assert.True(suite.T(), tenant.Active)
	assert.False(suite.T(), tenant.CreatedAt.IsZero())
	assert.False(suite.T(), tenant.UpdatedAt.IsZero())
	assert.Equal(suite.T(), tenant.CreatedAt, tenant.UpdatedAt)
}

func (suite *TenantTestSuite) TestNewTenant_InvalidData() {
	testCases := []struct {
		name         string
		identity     string
		identityType string
		email        string
		expectedErr  string
	}{
		{
			name:         "",
			identity:     "12345678000123",
			identityType: "cnpj",
			email:        "teste@empresa.com",
			expectedErr:  "name is required",
		},
		{
			name:         "A",
			identity:     "12345678000123",
			identityType: "cnpj",
			email:        "teste@empresa.com",
			expectedErr:  "name must be between 2 and 255 characters",
		},
		{
			name:         "Empresa Teste",
			identity:     "12",
			identityType: "cnpj",
			email:        "teste@empresa.com",
			expectedErr:  "identity must be between 3 and 50 characters",
		},
		{
			name:         "Empresa Teste",
			identity:     "12345678000123",
			identityType: "invalid_type",
			email:        "teste@empresa.com",
			expectedErr:  "invalid identity type",
		},
		{
			name:         "Empresa Teste",
			identity:     "12345678000123",
			identityType: "cnpj",
			email:        "test@invalid",
			expectedErr:  "invalid email format",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Arrange
			tenantID := value_objects.NewUUID()

			// Act
			tenant, err := NewTenant(tc.name, tc.identity, tc.identityType, tc.email, "", tenantID)

			// Assert
			assert.Error(suite.T(), err)
			assert.Nil(suite.T(), tenant)
			assert.Contains(suite.T(), err.Error(), tc.expectedErr)
		})
	}
}

func (suite *TenantTestSuite) TestTenantUpdate_ValidData() {
	// Arrange
	tenantID := value_objects.NewUUID()
	tenant, _ := NewTenant("Original", "12345678000123", "cnpj", "original@teste.com", "Address 1", tenantID)

	// Act
	updatedBy := value_objects.NewUUID()
	err := tenant.Update("Updated", "98765432000198", "cnpj", "updated@teste.com", "Address 2", updatedBy)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated", tenant.Name)
	assert.Equal(suite.T(), "98765432000198", tenant.Identity)
	assert.Equal(suite.T(), "cnpj", tenant.IdentityType)
	assert.Equal(suite.T(), "updated@teste.com", tenant.Email)
	assert.Equal(suite.T(), "Address 2", tenant.Address)
	assert.Equal(suite.T(), updatedBy, *tenant.UpdatedBy)
	assert.False(suite.T(), tenant.UpdatedAt.IsZero())
	assert.True(suite.T(), !tenant.UpdatedAt.Before(tenant.CreatedAt), "UpdatedAt should be after or equal to CreatedAt")
}

func (suite *TenantTestSuite) TestTenantUpdate_InvalidData() {
	// Arrange
	tenantID := value_objects.NewUUID()
	tenant, _ := NewTenant("Original", "12345678000123", "cnpj", "original@teste.com", "Address 1", tenantID)

	// Act
	updatedBy := value_objects.NewUUID()
	err := tenant.Update("", "123", "invalid_type", "invalid-email", "Address 2", updatedBy)

	// Assert
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "name is required")
}

func (suite *TenantTestSuite) TestTenantActivate() {
	// Arrange
	tenantID := value_objects.NewUUID()
	tenant, _ := NewTenant("Test", "12345678000123", "cnpj", "teste@teste.com", "Address", tenantID)
	tenant.Deactivate(tenantID) // Desativar primeiro

	// Act
	updatedBy := value_objects.NewUUID()
	tenant.Activate(updatedBy)

	// Assert
	assert.True(suite.T(), tenant.Active)
	assert.Equal(suite.T(), updatedBy, *tenant.UpdatedBy)
	assert.False(suite.T(), tenant.UpdatedAt.IsZero())
	assert.True(suite.T(), !tenant.UpdatedAt.Before(tenant.CreatedAt), "UpdatedAt should be after or equal to CreatedAt")
}

func (suite *TenantTestSuite) TestTenantDeactivate() {
	// Arrange
	tenantID := value_objects.NewUUID()
	tenant, _ := NewTenant("Test", "12345678000123", "cnpj", "teste@teste.com", "Address", tenantID)

	// Act
	updatedBy := value_objects.NewUUID()
	tenant.Deactivate(updatedBy)

	// Assert
	assert.False(suite.T(), tenant.Active)
	assert.Equal(suite.T(), updatedBy, *tenant.UpdatedBy)
	assert.False(suite.T(), tenant.UpdatedAt.IsZero())
	assert.True(suite.T(), !tenant.UpdatedAt.Before(tenant.CreatedAt), "UpdatedAt should be after or equal to CreatedAt")
}

func (suite *TenantTestSuite) TestTenantSetConfig() {
	// Arrange
	tenantID := value_objects.NewUUID()
	tenant, _ := NewTenant("Test", "12345678000123", "cnpj", "teste@teste.com", "Address", tenantID)

	// Act
	configID := value_objects.NewUUID()
	updatedBy := value_objects.NewUUID()
	tenant.SetConfig(configID, updatedBy)

	// Assert
	assert.Equal(suite.T(), configID, *tenant.ConfigID)
	assert.Equal(suite.T(), updatedBy, *tenant.UpdatedBy)
	assert.False(suite.T(), tenant.UpdatedAt.IsZero())
	assert.True(suite.T(), !tenant.UpdatedAt.Before(tenant.CreatedAt), "UpdatedAt should be after or equal to CreatedAt")
}

func (suite *TenantTestSuite) TestTenantIsActive() {
	// Arrange
	tenantID := value_objects.NewUUID()
	tenant, _ := NewTenant("Test", "12345678000123", "cnpj", "teste@teste.com", "Address", tenantID)

	// Assert
	assert.True(suite.T(), tenant.IsActive())
}

func (suite *TenantTestSuite) TestTenantHasConfig() {
	// Arrange
	tenantID := value_objects.NewUUID()
	tenant, _ := NewTenant("Test", "12345678000123", "cnpj", "teste@teste.com", "Address", tenantID)

	// Assert
	assert.False(suite.T(), tenant.HasConfig())

	// Act
	configID := value_objects.NewUUID()
	tenant.SetConfig(configID, tenantID)

	// Assert
	assert.True(suite.T(), tenant.HasConfig())
}

func (suite *TenantTestSuite) TestIsValidEmailImproved_ValidEmails() {
	validEmails := []string{
		"test@example.com",
		"teste@empresa.com.br",
		"user.name@domain.co.uk",
		"123@456.789",
	}

	for _, email := range validEmails {
		suite.T().Run(email, func(t *testing.T) {
			// Act
			result := IsValidEmailImproved(email)

			// Assert
			assert.True(suite.T(), result)
		})
	}
}

func (suite *TenantTestSuite) TestIsValidEmailImproved_InvalidEmails() {
	invalidEmails := []string{
		"invalidemail",     // sem @
		"@example.com",     // @ no início
		"test@",            // @ no final
		"test.example.com", // sem @ após test
		"",                 // vazio
	}

	for _, email := range invalidEmails {
		suite.T().Run(email, func(t *testing.T) {
			// Act
			result := IsValidEmailImproved(email)

			// Assert
			assert.False(suite.T(), result, "Email %s should be invalid", email)
		})
	}
}
