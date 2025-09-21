package errors

import (
	"errors"
	"fmt"
)

// Erros de domínio comuns
var (
	ErrNotFound         = errors.New("resource not found")
	ErrAlreadyExists    = errors.New("resource already exists")
	ErrInvalidInput     = errors.New("invalid input")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrForbidden        = errors.New("forbidden")
	ErrInternalError    = errors.New("internal error")
	ErrValidationFailed = errors.New("validation failed")
	ErrConcurrencyError = errors.New("concurrency error")
)

// DomainError representa um erro de domínio com contexto adicional
type DomainError struct {
	Type    string
	Message string
	Cause   error
	Context map[string]interface{}
}

func (e *DomainError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *DomainError) Unwrap() error {
	return e.Cause
}

// NewDomainError cria um novo erro de domínio
func NewDomainError(errorType, message string, cause error) *DomainError {
	return &DomainError{
		Type:    errorType,
		Message: message,
		Cause:   cause,
		Context: make(map[string]interface{}),
	}
}

// WithContext adiciona contexto ao erro
func (e *DomainError) WithContext(key string, value interface{}) *DomainError {
	e.Context[key] = value
	return e
}

// Erros específicos de domínio

// NewNotFoundError cria um erro de recurso não encontrado
func NewNotFoundError(resource string, id interface{}) *DomainError {
	return NewDomainError(
		"NOT_FOUND",
		fmt.Sprintf("%s not found", resource),
		ErrNotFound,
	).WithContext("resource", resource).WithContext("id", id)
}

// NewAlreadyExistsError cria um erro de recurso já existente
func NewAlreadyExistsError(resource string, field string, value interface{}) *DomainError {
	return NewDomainError(
		"ALREADY_EXISTS",
		fmt.Sprintf("%s with %s already exists", resource, field),
		ErrAlreadyExists,
	).WithContext("resource", resource).WithContext("field", field).WithContext("value", value)
}

// NewValidationError cria um erro de validação
func NewValidationError(field string, message string) *DomainError {
	return NewDomainError(
		"VALIDATION_ERROR",
		fmt.Sprintf("validation failed for field '%s': %s", field, message),
		ErrValidationFailed,
	).WithContext("field", field)
}

// NewUnauthorizedError cria um erro de não autorizado
func NewUnauthorizedError(message string) *DomainError {
	return NewDomainError(
		"UNAUTHORIZED",
		message,
		ErrUnauthorized,
	)
}

// NewForbiddenError cria um erro de acesso negado
func NewForbiddenError(resource string, action string) *DomainError {
	return NewDomainError(
		"FORBIDDEN",
		fmt.Sprintf("access denied to %s %s", action, resource),
		ErrForbidden,
	).WithContext("resource", resource).WithContext("action", action)
}

// NewInternalError cria um erro interno
func NewInternalError(message string, cause error) *DomainError {
	return NewDomainError(
		"INTERNAL_ERROR",
		message,
		cause,
	)
}
