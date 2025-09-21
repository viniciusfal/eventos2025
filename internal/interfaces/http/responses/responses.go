package responses

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// APIResponse representa uma resposta de sucesso da API
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// APIError representa uma resposta de erro da API
type APIError struct {
	Success   bool                   `json:"success"`
	Error     string                 `json:"error"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Code      string                 `json:"code,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// PaginatedResponse representa uma resposta paginada
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
	Message    string      `json:"message,omitempty"`
	Timestamp  time.Time   `json:"timestamp"`
}

// Pagination contém informações de paginação
type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// Success envia uma resposta de sucesso
func Success(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusOK, APIResponse{
		Success:   true,
		Data:      data,
		Message:   message,
		Timestamp: time.Now().UTC(),
	})
}

// Created envia uma resposta de criação bem-sucedida
func Created(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusCreated, APIResponse{
		Success:   true,
		Data:      data,
		Message:   message,
		Timestamp: time.Now().UTC(),
	})
}

// NoContent envia uma resposta sem conteúdo
func NoContent(c *gin.Context) {
	c.JSON(http.StatusNoContent, APIResponse{
		Success:   true,
		Timestamp: time.Now().UTC(),
	})
}

// Error envia uma resposta de erro
func Error(c *gin.Context, statusCode int, message string, code string, details map[string]interface{}) {
	c.JSON(statusCode, APIError{
		Success:   false,
		Error:     message,
		Code:      code,
		Details:   details,
		Timestamp: time.Now().UTC(),
	})
}

// BadRequest envia uma resposta de erro 400
func BadRequest(c *gin.Context, message string, details map[string]interface{}) {
	Error(c, http.StatusBadRequest, message, "BAD_REQUEST", details)
}

// Unauthorized envia uma resposta de erro 401
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message, "UNAUTHORIZED", nil)
}

// Forbidden envia uma resposta de erro 403
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message, "FORBIDDEN", nil)
}

// NotFound envia uma resposta de erro 404
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message, "NOT_FOUND", nil)
}

// Conflict envia uma resposta de erro 409
func Conflict(c *gin.Context, message string, details map[string]interface{}) {
	Error(c, http.StatusConflict, message, "CONFLICT", details)
}

// UnprocessableEntity envia uma resposta de erro 422
func UnprocessableEntity(c *gin.Context, message string, details map[string]interface{}) {
	Error(c, http.StatusUnprocessableEntity, message, "UNPROCESSABLE_ENTITY", details)
}

// InternalServerError envia uma resposta de erro 500
func InternalServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message, "INTERNAL_SERVER_ERROR", nil)
}

// ServiceUnavailable envia uma resposta de erro 503
func ServiceUnavailable(c *gin.Context, message string) {
	Error(c, http.StatusServiceUnavailable, message, "SERVICE_UNAVAILABLE", nil)
}

// Paginated envia uma resposta paginada
func Paginated(c *gin.Context, data interface{}, pagination Pagination, message string) {
	c.JSON(http.StatusOK, PaginatedResponse{
		Success:    true,
		Data:       data,
		Pagination: pagination,
		Message:    message,
		Timestamp:  time.Now().UTC(),
	})
}

// ValidationError envia uma resposta de erro de validação
func ValidationError(c *gin.Context, errors map[string]string) {
	details := make(map[string]interface{})
	for field, message := range errors {
		details[field] = message
	}

	Error(c, http.StatusBadRequest, "Validation failed", "VALIDATION_ERROR", details)
}

// DomainError envia uma resposta baseada em erro de domínio
func DomainError(c *gin.Context, err error) {
	// TODO: Implementar mapeamento de erros de domínio para códigos HTTP
	// Por enquanto, retorna erro interno
	InternalServerError(c, err.Error())
}

// CalculatePagination calcula informações de paginação
func CalculatePagination(page, pageSize, total int) Pagination {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	totalPages := (total + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	return Pagination{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}
}
