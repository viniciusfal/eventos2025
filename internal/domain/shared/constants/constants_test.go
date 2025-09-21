package constants

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// ConstantsTestSuite é a suíte de testes para constantes do sistema
type ConstantsTestSuite struct {
	suite.Suite
}

func TestConstantsSuite(t *testing.T) {
	suite.Run(t, new(ConstantsTestSuite))
}

func (suite *ConstantsTestSuite) TestStatusConstants() {
	// Assert
	assert.Equal(suite.T(), "active", StatusActive)
	assert.Equal(suite.T(), "inactive", StatusInactive)
	assert.Equal(suite.T(), "deleted", StatusDeleted)
}

func (suite *ConstantsTestSuite) TestIdentityTypeConstants() {
	// Assert
	assert.Equal(suite.T(), "cpf", IdentityTypeCPF)
	assert.Equal(suite.T(), "cnpj", IdentityTypeCNPJ)
	assert.Equal(suite.T(), "rg", IdentityTypeRG)
	assert.Equal(suite.T(), "other", IdentityTypeOther)
}

func (suite *ConstantsTestSuite) TestCheckMethodConstants() {
	// Assert
	assert.Equal(suite.T(), "facial_recognition", CheckMethodFacialRecognition)
	assert.Equal(suite.T(), "qr_code", CheckMethodQRCode)
	assert.Equal(suite.T(), "manual", CheckMethodManual)
}

func (suite *ConstantsTestSuite) TestQRTypeConstants() {
	// Assert
	assert.Equal(suite.T(), "checkin", QRTypeCheckin)
	assert.Equal(suite.T(), "checkout", QRTypeCheckout)
}

func (suite *ConstantsTestSuite) TestLogTypeConstants() {
	// Assert
	assert.Equal(suite.T(), "system", LogTypeSystem)
	assert.Equal(suite.T(), "audit", LogTypeAudit)
	assert.Equal(suite.T(), "event", LogTypeEvent)
	assert.Equal(suite.T(), "security", LogTypeSecurity)
}

func (suite *ConstantsTestSuite) TestActionConstants() {
	// Assert
	assert.Equal(suite.T(), "CREATE", ActionCreate)
	assert.Equal(suite.T(), "UPDATE", ActionUpdate)
	assert.Equal(suite.T(), "DELETE", ActionDelete)
	assert.Equal(suite.T(), "READ", ActionRead)
}

func (suite *ConstantsTestSuite) TestEntityTypeConstants() {
	// Assert
	assert.Equal(suite.T(), "tenant", EntityTypeTenant)
	assert.Equal(suite.T(), "user", EntityTypeUser)
	assert.Equal(suite.T(), "event", EntityTypeEvent)
	assert.Equal(suite.T(), "partner", EntityTypePartner)
	assert.Equal(suite.T(), "employee", EntityTypeEmployee)
	assert.Equal(suite.T(), "checkin", EntityTypeCheckin)
	assert.Equal(suite.T(), "checkout", EntityTypeCheckout)
}

func (suite *ConstantsTestSuite) TestModuleConstants() {
	// Assert
	assert.Equal(suite.T(), "auth", ModuleAuth)
	assert.Equal(suite.T(), "events", ModuleEvents)
	assert.Equal(suite.T(), "partners", ModulePartners)
	assert.Equal(suite.T(), "employees", ModuleEmployees)
	assert.Equal(suite.T(), "checkins", ModuleCheckins)
	assert.Equal(suite.T(), "reports", ModuleReports)
	assert.Equal(suite.T(), "audit", ModuleAudit)
	assert.Equal(suite.T(), "qr_code", ModuleQRCode)
	assert.Equal(suite.T(), "facial", ModuleFacial)
}

func (suite *ConstantsTestSuite) TestPermissionConstants() {
	// Assert
	assert.Equal(suite.T(), "read", PermissionRead)
	assert.Equal(suite.T(), "write", PermissionWrite)
	assert.Equal(suite.T(), "delete", PermissionDelete)
	assert.Equal(suite.T(), "admin", PermissionAdmin)
}

func (suite *ConstantsTestSuite) TestRoleConstants() {
	// Assert
	assert.Equal(suite.T(), "SUPER_ADMIN", RoleSuperAdmin)
	assert.Equal(suite.T(), "ADMIN", RoleAdmin)
	assert.Equal(suite.T(), "MANAGER", RoleManager)
	assert.Equal(suite.T(), "OPERATOR", RoleOperator)
	assert.Equal(suite.T(), "VIEWER", RoleViewer)
}

func (suite *ConstantsTestSuite) TestPaginationConstants() {
	// Assert
	assert.Equal(suite.T(), 20, DefaultPageSize)
	assert.Equal(suite.T(), 100, MaxPageSize)
}

func (suite *ConstantsTestSuite) TestCacheTTLConstants() {
	// Assert
	assert.Equal(suite.T(), 300, CacheTTLShort)   // 5 minutos
	assert.Equal(suite.T(), 1800, CacheTTLMedium) // 30 minutos
	assert.Equal(suite.T(), 3600, CacheTTLLong)   // 1 hora

	// Verify timing relationships
	assert.True(suite.T(), CacheTTLShort < CacheTTLMedium)
	assert.True(suite.T(), CacheTTLMedium < CacheTTLLong)
}

func (suite *ConstantsTestSuite) TestQRCodeConstants() {
	// Assert
	assert.Equal(suite.T(), 60, QRCodeValidityDuration) // 60 segundos
	assert.Equal(suite.T(), 1, QRCodeMaxUsage)          // 1 uso

	// Verify that validity duration is reasonable
	assert.True(suite.T(), QRCodeValidityDuration > 0)
	assert.True(suite.T(), QRCodeValidityDuration <= 300) // máximo 5 minutos
}

func (suite *ConstantsTestSuite) TestConstantsUniqueness() {
	// Test that important constants have unique values
	constantsMap := map[string]string{
		"StatusActive":   StatusActive,
		"StatusInactive": StatusInactive,
		"StatusDeleted":  StatusDeleted,
	}

	// Assert uniqueness
	assert.Len(suite.T(), constantsMap, 3)
	for key1, value1 := range constantsMap {
		for key2, value2 := range constantsMap {
			if key1 != key2 {
				assert.NotEqual(suite.T(), value1, value2, "Constants %s and %s should have different values", key1, key2)
			}
		}
	}
}

func (suite *ConstantsTestSuite) TestConstantsReasonableValues() {
	// Test that constants have reasonable values for their purpose

	// Pagination should be reasonable
	assert.True(suite.T(), DefaultPageSize > 0)
	assert.True(suite.T(), MaxPageSize > DefaultPageSize)
	assert.True(suite.T(), MaxPageSize <= 1000) // Should not be too large

	// Cache TTL should be reasonable
	assert.True(suite.T(), CacheTTLShort > 0)
	assert.True(suite.T(), CacheTTLMedium > CacheTTLShort)
	assert.True(suite.T(), CacheTTLLong > CacheTTLMedium)

	// QR Code settings should be reasonable
	assert.True(suite.T(), QRCodeValidityDuration > 10)  // At least 10 seconds
	assert.True(suite.T(), QRCodeValidityDuration < 600) // Less than 10 minutes
	assert.True(suite.T(), QRCodeMaxUsage >= 1)
}

func (suite *ConstantsTestSuite) TestConstantsImmutability() {
	// This test ensures that constants don't change during runtime
	// While we can't test this directly, we can test that they have expected values

	expectedValues := map[string]string{
		"StatusActive":   "active",
		"RoleAdmin":      "ADMIN",
		"PermissionRead": "read",
	}

	for constantName, expectedValue := range expectedValues {
		var actualValue string
		switch constantName {
		case "StatusActive":
			actualValue = StatusActive
		case "RoleAdmin":
			actualValue = RoleAdmin
		case "PermissionRead":
			actualValue = PermissionRead
		}

		assert.Equal(suite.T(), expectedValue, actualValue, "Constant %s should not change", constantName)
	}
}
