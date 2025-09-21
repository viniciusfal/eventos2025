package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	domainErrors "eventos-backend/internal/domain/shared/errors"
)

// DomainErrorsTestSuite é a suíte de testes para erros de domínio
type DomainErrorsTestSuite struct {
	suite.Suite
}

func TestDomainErrorsSuite(t *testing.T) {
	suite.Run(t, new(DomainErrorsTestSuite))
}

func (suite *DomainErrorsTestSuite) TestNewDomainError() {
	// Arrange
	errorType := "VALIDATION_ERROR"
	message := "Dados inválidos"
	cause := errors.New("invalid input")

	// Act
	domainError := domainErrors.NewDomainError(errorType, message, cause)

	// Assert
	assert.Equal(suite.T(), errorType, domainError.Type)
	assert.Equal(suite.T(), message, domainError.Message)
	assert.Equal(suite.T(), cause, domainError.Cause)
	assert.NotNil(suite.T(), domainError.Context)
}

func (suite *DomainErrorsTestSuite) TestDomainError_Error() {
	// Arrange
	domainError := domainErrors.NewDomainError("VALIDATION_ERROR", "Dados inválidos", nil)

	// Act
	errorString := domainError.Error()

	// Assert
	assert.Contains(suite.T(), errorString, "VALIDATION_ERROR")
	assert.Contains(suite.T(), errorString, "Dados inválidos")
}

func (suite *DomainErrorsTestSuite) TestDomainError_Error_WithCause() {
	// Arrange
	cause := errors.New("invalid input")
	domainError := domainErrors.NewDomainError("VALIDATION_ERROR", "Dados inválidos", cause)

	// Act
	errorString := domainError.Error()

	// Assert
	assert.Contains(suite.T(), errorString, "VALIDATION_ERROR")
	assert.Contains(suite.T(), errorString, "Dados inválidos")
	assert.Contains(suite.T(), errorString, "invalid input")
}

func (suite *DomainErrorsTestSuite) TestDomainError_Unwrap() {
	// Arrange
	cause := errors.New("root cause")
	domainError := domainErrors.NewDomainError("VALIDATION_ERROR", "Dados inválidos", cause)

	// Act
	unwrapped := domainError.Unwrap()

	// Assert
	assert.Equal(suite.T(), cause, unwrapped)
}

func (suite *DomainErrorsTestSuite) TestWithContext() {
	// Arrange
	domainError := domainErrors.NewDomainError("VALIDATION_ERROR", "Dados inválidos", nil)

	// Act
	domainError.WithContext("field", "email").WithContext("value", "invalid-email")

	// Assert
	assert.Equal(suite.T(), "email", domainError.Context["field"])
	assert.Equal(suite.T(), "invalid-email", domainError.Context["value"])
}

func (suite *DomainErrorsTestSuite) TestNewNotFoundError() {
	// Act
	domainError := domainErrors.NewNotFoundError("User", "123")

	// Assert
	assert.Equal(suite.T(), "NOT_FOUND", domainError.Type)
	assert.Contains(suite.T(), domainError.Message, "User not found")
	assert.Equal(suite.T(), "User", domainError.Context["resource"])
	assert.Equal(suite.T(), "123", domainError.Context["id"])
}

func (suite *DomainErrorsTestSuite) TestNewAlreadyExistsError() {
	// Act
	domainError := domainErrors.NewAlreadyExistsError("User", "email", "user@example.com")

	// Assert
	assert.Equal(suite.T(), "ALREADY_EXISTS", domainError.Type)
	assert.Contains(suite.T(), domainError.Message, "User with email already exists")
	assert.Equal(suite.T(), "User", domainError.Context["resource"])
	assert.Equal(suite.T(), "email", domainError.Context["field"])
	assert.Equal(suite.T(), "user@example.com", domainError.Context["value"])
}

func (suite *DomainErrorsTestSuite) TestNewValidationError() {
	// Act
	domainError := domainErrors.NewValidationError("email", "invalid format")

	// Assert
	assert.Equal(suite.T(), "VALIDATION_ERROR", domainError.Type)
	assert.Contains(suite.T(), domainError.Message, "validation failed for field 'email'")
	assert.Contains(suite.T(), domainError.Message, "invalid format")
	assert.Equal(suite.T(), "email", domainError.Context["field"])
}

func (suite *DomainErrorsTestSuite) TestNewUnauthorizedError() {
	// Act
	domainError := domainErrors.NewUnauthorizedError("Access denied")

	// Assert
	assert.Equal(suite.T(), "UNAUTHORIZED", domainError.Type)
	assert.Equal(suite.T(), "Access denied", domainError.Message)
}

func (suite *DomainErrorsTestSuite) TestNewForbiddenError() {
	// Act
	domainError := domainErrors.NewForbiddenError("User", "DELETE")

	// Assert
	assert.Equal(suite.T(), "FORBIDDEN", domainError.Type)
	assert.Contains(suite.T(), domainError.Message, "access denied to DELETE User")
	assert.Equal(suite.T(), "User", domainError.Context["resource"])
	assert.Equal(suite.T(), "DELETE", domainError.Context["action"])
}

func (suite *DomainErrorsTestSuite) TestNewInternalError() {
	// Arrange
	cause := errors.New("database connection failed")

	// Act
	domainError := domainErrors.NewInternalError("Failed to connect to database", cause)

	// Assert
	assert.Equal(suite.T(), "INTERNAL_ERROR", domainError.Type)
	assert.Equal(suite.T(), "Failed to connect to database", domainError.Message)
	assert.Equal(suite.T(), cause, domainError.Cause)
}

func (suite *DomainErrorsTestSuite) TestPredefinedErrors() {
	// Assert
	assert.Equal(suite.T(), "resource not found", domainErrors.ErrNotFound.Error())
	assert.Equal(suite.T(), "resource already exists", domainErrors.ErrAlreadyExists.Error())
	assert.Equal(suite.T(), "invalid input", domainErrors.ErrInvalidInput.Error())
	assert.Equal(suite.T(), "unauthorized", domainErrors.ErrUnauthorized.Error())
	assert.Equal(suite.T(), "forbidden", domainErrors.ErrForbidden.Error())
	assert.Equal(suite.T(), "internal error", domainErrors.ErrInternalError.Error())
	assert.Equal(suite.T(), "validation failed", domainErrors.ErrValidationFailed.Error())
	assert.Equal(suite.T(), "concurrency error", domainErrors.ErrConcurrencyError.Error())
}

func (suite *DomainErrorsTestSuite) TestDomainErrorChain() {
	// Arrange
	rootCause := errors.New("network error")
	internalError := domainErrors.NewInternalError("Database connection failed", rootCause)
	validationError := domainErrors.NewValidationError("email", "invalid format")
	validationError.Cause = internalError

	// Act
	errorString := validationError.Error()

	// Assert
	assert.Contains(suite.T(), errorString, "VALIDATION_ERROR")
	assert.Contains(suite.T(), errorString, "invalid format")
	assert.Contains(suite.T(), errorString, "Database connection failed")
	assert.Contains(suite.T(), errorString, "network error")

	// Test unwrapping
	unwrapped := validationError.Unwrap()
	assert.Equal(suite.T(), internalError, unwrapped)
}
