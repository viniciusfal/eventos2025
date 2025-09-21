package value_objects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	. "eventos-backend/internal/domain/shared/value_objects"
)

// UUIDTestSuite é a suíte de testes para UUID
type UUIDTestSuite struct {
	suite.Suite
}

func TestUUIDSuite(t *testing.T) {
	suite.Run(t, new(UUIDTestSuite))
}

func (suite *UUIDTestSuite) TestNewUUID() {
	// Act
	uuid := NewUUID()

	// Assert
	assert.NotEmpty(suite.T(), uuid.String())
	assert.False(suite.T(), uuid.IsZero())
}

func (suite *UUIDTestSuite) TestParseUUID_ValidUUID() {
	// Arrange
	validUUIDString := "550e8400-e29b-41d4-a716-446655440000"

	// Act
	uuid, err := ParseUUID(validUUIDString)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), validUUIDString, uuid.String())
	assert.False(suite.T(), uuid.IsZero())
}

func (suite *UUIDTestSuite) TestParseUUID_InvalidUUID() {
	// Arrange
	invalidUUIDStrings := []string{
		"invalid-uuid",
		"550e8400-e29b-41d4-a716",               // UUID muito curto
		"550e8400-e29b-41d4-a716-4466554400000", // UUID muito longo
		"",
		"not-a-uuid-at-all",
	}

	for _, invalidUUID := range invalidUUIDStrings {
		suite.T().Run(invalidUUID, func(t *testing.T) {
			// Act
			uuid, err := ParseUUID(invalidUUID)

			// Assert
			assert.Error(suite.T(), err)
			assert.True(suite.T(), uuid.IsZero())
		})
	}
}

func (suite *UUIDTestSuite) TestMustParseUUID_ValidUUID() {
	// Arrange
	validUUIDString := "550e8400-e29b-41d4-a716-446655440000"

	// Act
	uuid := MustParseUUID(validUUIDString)

	// Assert
	assert.Equal(suite.T(), validUUIDString, uuid.String())
	assert.False(suite.T(), uuid.IsZero())
}

func (suite *UUIDTestSuite) TestMustParseUUID_InvalidUUID() {
	// Arrange
	invalidUUIDString := "invalid-uuid"

	// Act & Assert
	assert.Panics(suite.T(), func() {
		MustParseUUID(invalidUUIDString)
	})
}

func (suite *UUIDTestSuite) TestUUIDString() {
	// Arrange
	expectedUUID := "550e8400-e29b-41d4-a716-446655440000"
	uuid, _ := ParseUUID(expectedUUID)

	// Act
	result := uuid.String()

	// Assert
	assert.Equal(suite.T(), expectedUUID, result)
}

func (suite *UUIDTestSuite) TestUUIDIsZero() {
	// Arrange
	zeroUUID := UUID{}
	validUUID, _ := ParseUUID("550e8400-e29b-41d4-a716-446655440000")

	// Act & Assert
	assert.True(suite.T(), zeroUUID.IsZero())
	assert.False(suite.T(), validUUID.IsZero())
}

func (suite *UUIDTestSuite) TestUUIDEquals() {
	// Arrange
	uuid1, _ := ParseUUID("550e8400-e29b-41d4-a716-446655440000")
	uuid2, _ := ParseUUID("550e8400-e29b-41d4-a716-446655440000")
	uuid3, _ := ParseUUID("660e8400-e29b-41d4-a716-446655440000")

	// Act & Assert
	assert.True(suite.T(), uuid1.Equals(uuid2))
	assert.False(suite.T(), uuid1.Equals(uuid3))
	assert.False(suite.T(), uuid2.Equals(uuid3))
}

func (suite *UUIDTestSuite) TestUUIDValue() {
	// Arrange
	uuid, _ := ParseUUID("550e8400-e29b-41d4-a716-446655440000")
	zeroUUID := UUID{}

	// Act
	uuidValue, err := uuid.Value()
	zeroValue, zeroErr := zeroUUID.Value()

	// Assert
	assert.NoError(suite.T(), err)
	assert.NoError(suite.T(), zeroErr)
	assert.Equal(suite.T(), "550e8400-e29b-41d4-a716-446655440000", uuidValue)
	assert.Nil(suite.T(), zeroValue)
}

func (suite *UUIDTestSuite) TestUUIDScan() {
	// Arrange
	uuid := UUID{}
	validUUIDString := "550e8400-e29b-41d4-a716-446655440000"

	// Act
	err := uuid.Scan(validUUIDString)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), validUUIDString, uuid.String())
}

func (suite *UUIDTestSuite) TestUUIDScan_NilValue() {
	// Arrange
	uuid := UUID{}

	// Act
	err := uuid.Scan(nil)

	// Assert
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), uuid.IsZero())
}

func (suite *UUIDTestSuite) TestUUIDScan_InvalidValues() {
	// Arrange
	uuid := UUID{}
	invalidValues := []interface{}{
		123,
		[]byte("invalid-bytes"),
		make(chan int),
	}

	for _, invalidValue := range invalidValues {
		suite.T().Run("", func(t *testing.T) {
			// Act
			err := uuid.Scan(invalidValue)

			// Assert
			assert.Error(suite.T(), err)
		})
	}
}

func (suite *UUIDTestSuite) TestUUID_Uniqueness() {
	// Act
	uuid1 := NewUUID()
	uuid2 := NewUUID()

	// Assert
	assert.NotEqual(suite.T(), uuid1.String(), uuid2.String())
	assert.False(suite.T(), uuid1.Equals(uuid2))
}
