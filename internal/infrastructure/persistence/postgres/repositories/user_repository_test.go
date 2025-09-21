package repositories

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"eventos-backend/internal/domain/shared/value_objects"
	"eventos-backend/internal/domain/user"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

// MockDB é um mock do banco de dados
type MockDB struct {
	mock.Mock
}

func (m *MockDB) Get(dest interface{}, query string, args ...interface{}) error {
	arguments := m.Called(dest, query, args)
	return arguments.Error(0)
}

func (m *MockDB) Select(dest interface{}, query string, args ...interface{}) error {
	arguments := m.Called(dest, query, args)
	return arguments.Error(0)
}

func (m *MockDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	arguments := m.Called(query, args)
	return arguments.Get(0).(sql.Result), arguments.Error(1)
}

// MockSQLResult é um mock do resultado SQL
type MockSQLResult struct {
	mock.Mock
}

func (m *MockSQLResult) LastInsertId() (int64, error) {
	arguments := m.Called()
	return arguments.Get(0).(int64), arguments.Error(1)
}

func (m *MockSQLResult) RowsAffected() (int64, error) {
	arguments := m.Called()
	return arguments.Get(0).(int64), arguments.Error(1)
}

// UserRepositoryTestSuite é a suíte de testes para UserRepository
type UserRepositoryTestSuite struct {
	suite.Suite
	mockDB     *MockDB
	mockResult *MockSQLResult
	repository user.Repository
	logger     *zap.Logger
}

func TestUserRepositorySuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}

func (suite *UserRepositoryTestSuite) SetupTest() {
	suite.mockDB = new(MockDB)
	suite.mockResult = new(MockSQLResult)

	// Configurar logger mock (simplificado)
	suite.logger = zap.NewNop()

	// Criar repositório com mock
	suite.repository = NewUserRepository(&sqlx.DB{}, suite.logger)

	// Usar reflection para injetar o mock DB
	// Nota: Em um cenário real, você poderia usar dependency injection
}

func (suite *UserRepositoryTestSuite) TestCreate_Success() {
	// Arrange
	ctx := context.Background()
	tenantID := value_objects.NewUUID()
	createdBy := value_objects.NewUUID()
	user := &user.User{
		ID:        value_objects.NewUUID(),
		TenantID:  tenantID,
		FullName:  "João Silva",
		Email:     "joao@example.com",
		Phone:     "+5511999999999",
		Username:  "joao_silva",
		Password:  "$2a$10$...",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: &createdBy,
		UpdatedBy: &createdBy,
	}

	expectedQuery := `INSERT INTO user \(id_user, id_tenant, full_name, email, phone, username, password, active, created_at, updated_at, created_by, updated_by\) VALUES \(\$1, \$2, \$3, \$4, \$5, \$6, \$7, \$8, \$9, \$10, \$11, \$12\)`

	suite.mockDB.On("Exec", expectedQuery,
		user.ID.String(), user.TenantID.String(), user.FullName, user.Email,
		sql.NullString{String: user.Phone, Valid: true}, user.Username,
		user.Password, user.Active, user.CreatedAt, user.UpdatedAt,
		user.CreatedBy.String(), user.UpdatedBy.String()).
		Return(suite.mockResult, nil)

	suite.mockResult.On("RowsAffected").Return(int64(1), nil)

	// Act
	err := suite.repository.Create(ctx, user)

	// Assert
	assert.NoError(suite.T(), err)
	suite.mockDB.AssertExpectations(suite.T())
	suite.mockResult.AssertExpectations(suite.T())
}

func (suite *UserRepositoryTestSuite) TestGetByID_Success() {
	// Arrange
	ctx := context.Background()
	userID := value_objects.NewUUID()
	expectedUser := &user.User{
		ID:        userID,
		TenantID:  value_objects.NewUUID(),
		FullName:  "João Silva",
		Email:     "joao@example.com",
		Username:  "joao_silva",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	expectedQuery := `SELECT id_user, id_tenant, full_name, email, phone, username, password, active, created_at, updated_at, created_by, updated_by FROM user WHERE id_user = \$1 AND active = true`

	suite.mockDB.On("Get", mock.Anything, expectedQuery, userID.String()).
		Run(func(args mock.Arguments) {
			// Simular o resultado populando o destino
			dest := args.Get(0).(*userRow)
			dest.ID = expectedUser.ID.String()
			dest.TenantID = expectedUser.TenantID.String()
			dest.FullName = expectedUser.FullName
			dest.Email = expectedUser.Email
			dest.Username = expectedUser.Username
			dest.Active = expectedUser.Active
		}).
		Return(nil)

	// Act
	result, err := suite.repository.GetByID(ctx, userID)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), expectedUser.ID, result.ID)
	assert.Equal(suite.T(), expectedUser.FullName, result.FullName)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *UserRepositoryTestSuite) TestGetByID_NotFound() {
	// Arrange
	ctx := context.Background()
	userID := value_objects.NewUUID()
	expectedQuery := `SELECT id_user, id_tenant, full_name, email, phone, username, password, active, created_at, updated_at, created_by, updated_by FROM user WHERE id_user = \$1 AND active = true`

	suite.mockDB.On("Get", mock.Anything, expectedQuery, userID.String()).
		Return(sql.ErrNoRows)

	// Act
	result, err := suite.repository.GetByID(ctx, userID)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "user not found")
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *UserRepositoryTestSuite) TestGetByUsername_Success() {
	// Arrange
	ctx := context.Background()
	username := "joao_silva"
	expectedUser := &user.User{
		ID:       value_objects.NewUUID(),
		FullName: "João Silva",
		Email:    "joao@example.com",
		Username: username,
		Active:   true,
	}

	expectedQuery := `SELECT id_user, id_tenant, full_name, email, phone, username, password, active, created_at, updated_at, created_by, updated_by FROM user WHERE username = \$1 AND active = true`

	suite.mockDB.On("Get", mock.Anything, expectedQuery, username).
		Run(func(args mock.Arguments) {
			dest := args.Get(0).(*userRow)
			dest.ID = expectedUser.ID.String()
			dest.FullName = expectedUser.FullName
			dest.Email = expectedUser.Email
			dest.Username = expectedUser.Username
			dest.Active = expectedUser.Active
		}).
		Return(nil)

	// Act
	result, err := suite.repository.GetByUsername(ctx, username)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), expectedUser.Username, result.Username)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *UserRepositoryTestSuite) TestExistsByUsername_Success() {
	// Arrange
	ctx := context.Background()
	username := "joao_silva"
	expectedQuery := `SELECT EXISTS\(SELECT 1 FROM user WHERE username = \$1 AND active = true\)`

	suite.mockDB.On("Get", mock.Anything, expectedQuery, username).
		Run(func(args mock.Arguments) {
			dest := args.Get(0).(*bool)
			*dest = true
		}).
		Return(nil)

	// Act
	exists, err := suite.repository.(interface {
		ExistsByUsername(context.Context, string, *value_objects.UUID) (bool, error)
	}).ExistsByUsername(ctx, username, nil)

	// Assert
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *UserRepositoryTestSuite) TestUpdate_Success() {
	// Arrange
	ctx := context.Background()
	tenantID := value_objects.NewUUID()
	updatedBy := value_objects.NewUUID()
	user := &user.User{
		ID:        value_objects.NewUUID(),
		TenantID:  tenantID,
		FullName:  "João Silva Santos",
		Email:     "joao.santos@example.com",
		Phone:     "+5511988888888",
		Username:  "joao_santos",
		UpdatedAt: time.Now(),
		UpdatedBy: &updatedBy,
	}

	expectedQuery := `UPDATE user SET full_name = \$1, email = \$2, phone = \$3, username = \$4, updated_at = \$5, updated_by = \$6 WHERE id_user = \$7`

	suite.mockDB.On("Exec", expectedQuery,
		user.FullName, user.Email,
		sql.NullString{String: user.Phone, Valid: true},
		user.Username, user.UpdatedAt, user.UpdatedBy.String(), user.ID.String()).
		Return(suite.mockResult, nil)

	suite.mockResult.On("RowsAffected").Return(int64(1), nil)

	// Act
	err := suite.repository.Update(ctx, user)

	// Assert
	assert.NoError(suite.T(), err)
	suite.mockDB.AssertExpectations(suite.T())
	suite.mockResult.AssertExpectations(suite.T())
}

func (suite *UserRepositoryTestSuite) TestDelete_Success() {
	// Arrange
	ctx := context.Background()
	userID := value_objects.NewUUID()
	deletedBy := value_objects.NewUUID()
	expectedQuery := `UPDATE user SET active = false, updated_at = \$1, updated_by = \$2 WHERE id_user = \$3`

	suite.mockDB.On("Exec", expectedQuery, mock.Anything, deletedBy.String(), userID.String()).
		Return(suite.mockResult, nil)

	suite.mockResult.On("RowsAffected").Return(int64(1), nil)

	// Act
	err := suite.repository.Delete(ctx, userID, deletedBy)

	// Assert
	assert.NoError(suite.T(), err)
	suite.mockDB.AssertExpectations(suite.T())
	suite.mockResult.AssertExpectations(suite.T())
}

func (suite *UserRepositoryTestSuite) TestList_Success() {
	// Arrange
	ctx := context.Background()
	filters := user.ListFilters{
		Page:     1,
		PageSize: 10,
		OrderBy:  "created_at",
	}

	var users []*user.User
	expectedQuery := `SELECT id_user, id_tenant, full_name, email, phone, username, password, active, created_at, updated_at, created_by, updated_by FROM user WHERE active = true ORDER BY created_at ASC LIMIT \$1 OFFSET \$2`

	suite.mockDB.On("Select", &users, expectedQuery, 10, 0).
		Run(func(args mock.Arguments) {
			// Simular alguns usuários
			users = append(users, &user.User{
				ID:       value_objects.NewUUID(),
				FullName: "João Silva",
				Email:    "joao@example.com",
				Username: "joao_silva",
				Active:   true,
			})
		}).
		Return(nil)

	// Act
	result, total, err := suite.repository.List(ctx, filters)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 1)
	assert.Equal(suite.T(), 1, total)
	suite.mockDB.AssertExpectations(suite.T())
}
