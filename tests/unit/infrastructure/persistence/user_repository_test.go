package repositories

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	. "eventos-backend/internal/infrastructure/persistence/postgres/repositories"

	"github.com/jmoiron/sqlx"
)

// UserRepositoryTestSuite é a suíte de testes para UserRepository
type UserRepositoryTestSuite struct {
	suite.Suite
	logger *zap.Logger
}

func TestUserRepositorySuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}

func (suite *UserRepositoryTestSuite) SetupTest() {
	suite.logger = zap.NewNop()
}

func (suite *UserRepositoryTestSuite) TestNewUserRepository() {
	// Arrange
	db := &sqlx.DB{} // Mock database
	logger := suite.logger

	// Act
	repository := NewUserRepository(db, logger)

	// Assert
	assert.NotNil(suite.T(), repository)
}

func (suite *UserRepositoryTestSuite) TestUserRepositoryInterface() {
	// Arrange
	db := &sqlx.DB{} // Mock database
	logger := suite.logger

	// Act
	repository := NewUserRepository(db, logger)

	// Assert - Verificar que implementa a interface correta
	assert.NotNil(suite.T(), repository)
	// O compilador já garante que implementa user.Repository
}
