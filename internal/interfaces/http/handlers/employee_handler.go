package handlers

import (
	"strconv"
	"time"

	"eventos-backend/internal/domain/employee"
	"eventos-backend/internal/domain/shared/errors"
	"eventos-backend/internal/domain/shared/value_objects"
	jwtService "eventos-backend/internal/infrastructure/auth/jwt"
	httpResponses "eventos-backend/internal/interfaces/http/responses"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// EmployeeHandler gerencia as operações de funcionário
type EmployeeHandler struct {
	employeeService employee.Service
	logger          *zap.Logger
}

// NewEmployeeHandler cria uma nova instância do handler de funcionário
func NewEmployeeHandler(employeeService employee.Service, logger *zap.Logger) *EmployeeHandler {
	return &EmployeeHandler{
		employeeService: employeeService,
		logger:          logger,
	}
}

// CreateEmployeeRequest representa uma requisição de criação de funcionário
type CreateEmployeeRequest struct {
	FullName     string `json:"full_name" binding:"required"`
	Identity     string `json:"identity" binding:"required"`
	IdentityType string `json:"identity_type" binding:"required"`
	DateOfBirth  string `json:"date_of_birth,omitempty"` // Format: 2006-01-02
	Phone        string `json:"phone" binding:"required"`
	Email        string `json:"email" binding:"required,email"`
}

// UpdateEmployeeRequest representa uma requisição de atualização de funcionário
type UpdateEmployeeRequest struct {
	FullName     string `json:"full_name" binding:"required"`
	Identity     string `json:"identity" binding:"required"`
	IdentityType string `json:"identity_type" binding:"required"`
	DateOfBirth  string `json:"date_of_birth,omitempty"` // Format: 2006-01-02
	Phone        string `json:"phone" binding:"required"`
	Email        string `json:"email" binding:"required,email"`
}

// UploadPhotoRequest representa uma requisição de upload de foto
type UploadPhotoRequest struct {
	PhotoURL string `json:"photo_url" binding:"required"`
}

// FaceRecognitionRequest representa uma requisição de reconhecimento facial
type FaceRecognitionRequest struct {
	FaceEmbedding []float32 `json:"face_embedding" binding:"required"`
	Threshold     float64   `json:"threshold,omitempty"`
}

// EmployeeResponse representa a resposta de um funcionário
type EmployeeResponse struct {
	ID               string  `json:"id"`
	TenantID         string  `json:"tenant_id"`
	FullName         string  `json:"full_name"`
	Identity         string  `json:"identity"`
	IdentityType     string  `json:"identity_type"`
	DateOfBirth      *string `json:"date_of_birth,omitempty"`
	PhotoURL         string  `json:"photo_url,omitempty"`
	HasFaceEmbedding bool    `json:"has_face_embedding"`
	Phone            string  `json:"phone"`
	Email            string  `json:"email"`
	Active           bool    `json:"active"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
	CreatedBy        *string `json:"created_by,omitempty"`
	UpdatedBy        *string `json:"updated_by,omitempty"`
}

// EmployeeListResponse representa a resposta de listagem de funcionários
type EmployeeListResponse struct {
	Employees  []EmployeeResponse       `json:"employees"`
	Pagination httpResponses.Pagination `json:"pagination"`
}

// FaceRecognitionResponse representa a resposta de reconhecimento facial
type FaceRecognitionResponse struct {
	Matches []FaceMatch `json:"matches"`
}

// FaceMatch representa um match de reconhecimento facial
type FaceMatch struct {
	Employee   EmployeeResponse `json:"employee"`
	Similarity float64          `json:"similarity"`
	Confidence string           `json:"confidence"` // high, medium, low
}

// Create cria um novo funcionário
func (h *EmployeeHandler) Create(c *gin.Context) {
	var req CreateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid create employee request", zap.Error(err))
		httpResponses.BadRequest(c, "Invalid request data", map[string]interface{}{
			"validation_errors": err.Error(),
		})
		return
	}

	// Obter dados do usuário autenticado
	userClaims, exists := c.Get("claims")
	if !exists {
		h.logger.Error("User claims not found in context")
		httpResponses.Unauthorized(c, "Authentication required")
		return
	}

	claims := userClaims.(*jwtService.Claims)
	tenantID, err := value_objects.ParseUUID(claims.TenantID)
	if err != nil {
		h.logger.Error("Invalid tenant ID in claims", zap.Error(err))
		httpResponses.InternalServerError(c, "Invalid authentication data")
		return
	}

	userID, err := value_objects.ParseUUID(claims.UserID)
	if err != nil {
		h.logger.Error("Invalid user ID in claims", zap.Error(err))
		httpResponses.InternalServerError(c, "Invalid authentication data")
		return
	}

	// Converter data de nascimento
	var dateOfBirth *time.Time
	if req.DateOfBirth != "" {
		dob, err := time.Parse("2006-01-02", req.DateOfBirth)
		if err != nil {
			h.logger.Warn("Invalid date of birth format", zap.Error(err))
			httpResponses.BadRequest(c, "Invalid date of birth format. Use YYYY-MM-DD", nil)
			return
		}
		dateOfBirth = &dob
	}

	// Criar funcionário
	emp, err := h.employeeService.CreateEmployee(
		c.Request.Context(),
		tenantID,
		req.FullName,
		req.Identity,
		req.IdentityType,
		req.Phone,
		req.Email,
		dateOfBirth,
		userID,
	)
	if err != nil {
		h.handleServiceError(c, err, "create employee")
		return
	}

	response := h.convertToEmployeeResponse(emp)
	h.logger.Info("Employee created successfully", zap.String("employee_id", emp.ID.String()))
	httpResponses.Created(c, response, "Employee created successfully")
}

// GetByID busca um funcionário pelo ID
func (h *EmployeeHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := value_objects.ParseUUID(idStr)
	if err != nil {
		h.logger.Warn("Invalid employee ID", zap.String("id", idStr))
		httpResponses.BadRequest(c, "Invalid employee ID format", nil)
		return
	}

	// Obter tenant do usuário autenticado
	userClaims, exists := c.Get("claims")
	if !exists {
		httpResponses.Unauthorized(c, "Authentication required")
		return
	}

	claims := userClaims.(*jwtService.Claims)
	tenantID, err := value_objects.ParseUUID(claims.TenantID)
	if err != nil {
		h.logger.Error("Invalid tenant ID in claims", zap.Error(err))
		httpResponses.InternalServerError(c, "Invalid authentication data")
		return
	}

	emp, err := h.employeeService.GetEmployeeByTenant(c.Request.Context(), id, tenantID)
	if err != nil {
		h.handleServiceError(c, err, "get employee")
		return
	}

	response := h.convertToEmployeeResponse(emp)
	httpResponses.Success(c, response, "Employee retrieved successfully")
}

// Update atualiza um funcionário
func (h *EmployeeHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := value_objects.ParseUUID(idStr)
	if err != nil {
		h.logger.Warn("Invalid employee ID", zap.String("id", idStr))
		httpResponses.BadRequest(c, "Invalid employee ID format", nil)
		return
	}

	var req UpdateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid update employee request", zap.Error(err))
		httpResponses.BadRequest(c, "Invalid request data", map[string]interface{}{
			"validation_errors": err.Error(),
		})
		return
	}

	// Obter dados do usuário autenticado
	userClaims, exists := c.Get("claims")
	if !exists {
		httpResponses.Unauthorized(c, "Authentication required")
		return
	}

	claims := userClaims.(*jwtService.Claims)
	userID, err := value_objects.ParseUUID(claims.UserID)
	if err != nil {
		h.logger.Error("Invalid user ID in claims", zap.Error(err))
		httpResponses.InternalServerError(c, "Invalid authentication data")
		return
	}

	// Converter data de nascimento
	var dateOfBirth *time.Time
	if req.DateOfBirth != "" {
		dob, err := time.Parse("2006-01-02", req.DateOfBirth)
		if err != nil {
			h.logger.Warn("Invalid date of birth format", zap.Error(err))
			httpResponses.BadRequest(c, "Invalid date of birth format. Use YYYY-MM-DD", nil)
			return
		}
		dateOfBirth = &dob
	}

	emp, err := h.employeeService.UpdateEmployee(
		c.Request.Context(),
		id,
		req.FullName,
		req.Identity,
		req.IdentityType,
		req.Phone,
		req.Email,
		dateOfBirth,
		userID,
	)
	if err != nil {
		h.handleServiceError(c, err, "update employee")
		return
	}

	response := h.convertToEmployeeResponse(emp)
	h.logger.Info("Employee updated successfully", zap.String("employee_id", emp.ID.String()))
	httpResponses.Success(c, response, "Employee updated successfully")
}

// Delete remove um funcionário (soft delete)
func (h *EmployeeHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := value_objects.ParseUUID(idStr)
	if err != nil {
		h.logger.Warn("Invalid employee ID", zap.String("id", idStr))
		httpResponses.BadRequest(c, "Invalid employee ID format", nil)
		return
	}

	// Obter dados do usuário autenticado
	userClaims, exists := c.Get("claims")
	if !exists {
		httpResponses.Unauthorized(c, "Authentication required")
		return
	}

	claims := userClaims.(*jwtService.Claims)
	userID, err := value_objects.ParseUUID(claims.UserID)
	if err != nil {
		h.logger.Error("Invalid user ID in claims", zap.Error(err))
		httpResponses.InternalServerError(c, "Invalid authentication data")
		return
	}

	err = h.employeeService.DeleteEmployee(c.Request.Context(), id, userID)
	if err != nil {
		h.handleServiceError(c, err, "delete employee")
		return
	}

	h.logger.Info("Employee deleted successfully", zap.String("employee_id", id.String()))
	httpResponses.Success(c, nil, "Employee deleted successfully")
}

// List lista funcionários com paginação e filtros
func (h *EmployeeHandler) List(c *gin.Context) {
	// Obter tenant do usuário autenticado
	userClaims, exists := c.Get("claims")
	if !exists {
		httpResponses.Unauthorized(c, "Authentication required")
		return
	}

	claims := userClaims.(*jwtService.Claims)
	tenantID, err := value_objects.ParseUUID(claims.TenantID)
	if err != nil {
		h.logger.Error("Invalid tenant ID in claims", zap.Error(err))
		httpResponses.InternalServerError(c, "Invalid authentication data")
		return
	}

	// Construir filtros
	filters := h.buildListFilters(c)
	filters.TenantID = &tenantID

	employees, total, err := h.employeeService.ListEmployees(c.Request.Context(), filters)
	if err != nil {
		h.handleServiceError(c, err, "list employees")
		return
	}

	// Converter para resposta
	employeeResponses := make([]EmployeeResponse, len(employees))
	for i, emp := range employees {
		employeeResponses[i] = h.convertToEmployeeResponse(emp)
	}

	response := EmployeeListResponse{
		Employees: employeeResponses,
		Pagination: httpResponses.Pagination{
			Page:       filters.Page,
			PageSize:   filters.PageSize,
			Total:      total,
			TotalPages: (total + filters.PageSize - 1) / filters.PageSize,
		},
	}

	httpResponses.Success(c, response, "Employees retrieved successfully")
}

// UploadPhoto faz upload da foto de um funcionário
func (h *EmployeeHandler) UploadPhoto(c *gin.Context) {
	idStr := c.Param("id")
	id, err := value_objects.ParseUUID(idStr)
	if err != nil {
		h.logger.Warn("Invalid employee ID", zap.String("id", idStr))
		httpResponses.BadRequest(c, "Invalid employee ID format", nil)
		return
	}

	var req UploadPhotoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid upload photo request", zap.Error(err))
		httpResponses.BadRequest(c, "Invalid request data", map[string]interface{}{
			"validation_errors": err.Error(),
		})
		return
	}

	// Obter dados do usuário autenticado
	userClaims, exists := c.Get("claims")
	if !exists {
		httpResponses.Unauthorized(c, "Authentication required")
		return
	}

	claims := userClaims.(*jwtService.Claims)
	userID, err := value_objects.ParseUUID(claims.UserID)
	if err != nil {
		h.logger.Error("Invalid user ID in claims", zap.Error(err))
		httpResponses.InternalServerError(c, "Invalid authentication data")
		return
	}

	err = h.employeeService.UpdateEmployeePhoto(c.Request.Context(), id, req.PhotoURL, userID)
	if err != nil {
		h.handleServiceError(c, err, "upload employee photo")
		return
	}

	h.logger.Info("Employee photo uploaded successfully", zap.String("employee_id", id.String()))
	httpResponses.Success(c, nil, "Photo uploaded successfully")
}

// UpdateFaceEmbedding atualiza o embedding facial de um funcionário
func (h *EmployeeHandler) UpdateFaceEmbedding(c *gin.Context) {
	idStr := c.Param("id")
	id, err := value_objects.ParseUUID(idStr)
	if err != nil {
		h.logger.Warn("Invalid employee ID", zap.String("id", idStr))
		httpResponses.BadRequest(c, "Invalid employee ID format", nil)
		return
	}

	var req FaceRecognitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid face embedding request", zap.Error(err))
		httpResponses.BadRequest(c, "Invalid request data", map[string]interface{}{
			"validation_errors": err.Error(),
		})
		return
	}

	// Validar tamanho do embedding
	if len(req.FaceEmbedding) != 512 {
		h.logger.Warn("Invalid face embedding size", zap.Int("size", len(req.FaceEmbedding)))
		httpResponses.BadRequest(c, "Face embedding must have exactly 512 dimensions", nil)
		return
	}

	// Obter dados do usuário autenticado
	userClaims, exists := c.Get("claims")
	if !exists {
		httpResponses.Unauthorized(c, "Authentication required")
		return
	}

	claims := userClaims.(*jwtService.Claims)
	userID, err := value_objects.ParseUUID(claims.UserID)
	if err != nil {
		h.logger.Error("Invalid user ID in claims", zap.Error(err))
		httpResponses.InternalServerError(c, "Invalid authentication data")
		return
	}

	err = h.employeeService.UpdateEmployeeFaceEmbedding(c.Request.Context(), id, req.FaceEmbedding, userID)
	if err != nil {
		h.handleServiceError(c, err, "update employee face embedding")
		return
	}

	h.logger.Info("Employee face embedding updated successfully", zap.String("employee_id", id.String()))
	httpResponses.Success(c, nil, "Face embedding updated successfully")
}

// RecognizeFace busca funcionários por similaridade facial
func (h *EmployeeHandler) RecognizeFace(c *gin.Context) {
	var req FaceRecognitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid face recognition request", zap.Error(err))
		httpResponses.BadRequest(c, "Invalid request data", map[string]interface{}{
			"validation_errors": err.Error(),
		})
		return
	}

	// Validar tamanho do embedding
	if len(req.FaceEmbedding) != 512 {
		h.logger.Warn("Invalid face embedding size", zap.Int("size", len(req.FaceEmbedding)))
		httpResponses.BadRequest(c, "Face embedding must have exactly 512 dimensions", nil)
		return
	}

	// Obter tenant do usuário autenticado
	userClaims, exists := c.Get("claims")
	if !exists {
		httpResponses.Unauthorized(c, "Authentication required")
		return
	}

	claims := userClaims.(*jwtService.Claims)
	tenantID, err := value_objects.ParseUUID(claims.TenantID)
	if err != nil {
		h.logger.Error("Invalid tenant ID in claims", zap.Error(err))
		httpResponses.InternalServerError(c, "Invalid authentication data")
		return
	}

	// Definir threshold padrão se não fornecido
	threshold := req.Threshold
	if threshold <= 0 {
		threshold = 0.8 // 80% de similaridade
	}

	results, err := h.employeeService.RecognizeFace(
		c.Request.Context(),
		req.FaceEmbedding,
		&tenantID,
		float32(threshold),
	)
	if err != nil {
		h.handleServiceError(c, err, "face recognition")
		return
	}

	// Converter para resposta
	matches := make([]FaceMatch, len(results))
	for i, result := range results {
		confidence := h.getConfidenceLevel(float64(result.Similarity))
		matches[i] = FaceMatch{
			Employee:   h.convertToEmployeeResponse(result.Employee),
			Similarity: float64(result.Similarity),
			Confidence: confidence,
		}
	}

	response := FaceRecognitionResponse{
		Matches: matches,
	}

	httpResponses.Success(c, response, "Face recognition completed successfully")
}

// buildListFilters constrói os filtros de listagem a partir dos query parameters
func (h *EmployeeHandler) buildListFilters(c *gin.Context) employee.ListFilters {
	filters := employee.ListFilters{
		Page:     1,
		PageSize: 20,
		OrderBy:  "full_name",
	}

	// Paginação
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filters.Page = page
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 && pageSize <= 100 {
			filters.PageSize = pageSize
		}
	}

	// Filtros de busca
	if fullName := c.Query("full_name"); fullName != "" {
		filters.FullName = &fullName
	}

	if identity := c.Query("identity"); identity != "" {
		filters.Identity = &identity
	}

	if identityType := c.Query("identity_type"); identityType != "" {
		filters.IdentityType = &identityType
	}

	if email := c.Query("email"); email != "" {
		filters.Email = &email
	}

	if phone := c.Query("phone"); phone != "" {
		filters.Phone = &phone
	}

	if hasPhotoStr := c.Query("has_photo"); hasPhotoStr != "" {
		if hasPhoto, err := strconv.ParseBool(hasPhotoStr); err == nil {
			filters.HasPhoto = &hasPhoto
		}
	}

	if hasFaceStr := c.Query("has_face_embedding"); hasFaceStr != "" {
		if hasFace, err := strconv.ParseBool(hasFaceStr); err == nil {
			filters.HasFaceEmbedding = &hasFace
		}
	}

	if activeStr := c.Query("active"); activeStr != "" {
		if active, err := strconv.ParseBool(activeStr); err == nil {
			filters.Active = &active
		}
	}

	// Ordenação
	if orderBy := c.Query("order_by"); orderBy != "" {
		validFields := []string{"full_name", "identity", "email", "created_at", "updated_at"}
		for _, field := range validFields {
			if orderBy == field {
				filters.OrderBy = orderBy
				break
			}
		}
	}

	if orderDesc := c.Query("order_desc"); orderDesc == "true" {
		filters.OrderDesc = true
	}

	return filters
}

// convertToEmployeeResponse converte Employee para EmployeeResponse
func (h *EmployeeHandler) convertToEmployeeResponse(emp *employee.Employee) EmployeeResponse {
	response := EmployeeResponse{
		ID:               emp.ID.String(),
		TenantID:         emp.TenantID.String(),
		FullName:         emp.FullName,
		Identity:         emp.Identity,
		IdentityType:     emp.IdentityType,
		PhotoURL:         emp.PhotoURL,
		HasFaceEmbedding: len(emp.FaceEmbedding) > 0,
		Phone:            emp.Phone,
		Email:            emp.Email,
		Active:           emp.Active,
		CreatedAt:        emp.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        emp.UpdatedAt.Format(time.RFC3339),
	}

	if emp.DateOfBirth != nil {
		dateOfBirth := emp.DateOfBirth.Format("2006-01-02")
		response.DateOfBirth = &dateOfBirth
	}

	if emp.CreatedBy != nil {
		createdBy := emp.CreatedBy.String()
		response.CreatedBy = &createdBy
	}

	if emp.UpdatedBy != nil {
		updatedBy := emp.UpdatedBy.String()
		response.UpdatedBy = &updatedBy
	}

	return response
}

// getConfidenceLevel determina o nível de confiança baseado na similaridade
func (h *EmployeeHandler) getConfidenceLevel(similarity float64) string {
	if similarity >= 0.95 {
		return "high"
	} else if similarity >= 0.85 {
		return "medium"
	}
	return "low"
}

// handleServiceError trata erros do serviço de domínio
func (h *EmployeeHandler) handleServiceError(c *gin.Context, err error, operation string) {
	switch e := err.(type) {
	case *errors.DomainError:
		switch e.Type {
		case "VALIDATION_ERROR":
			h.logger.Warn("Validation error in "+operation, zap.Error(err))
			httpResponses.BadRequest(c, e.Message, e.Context)
		case "NOT_FOUND":
			h.logger.Warn("Resource not found in "+operation, zap.Error(err))
			httpResponses.NotFound(c, e.Message)
		case "CONFLICT":
			h.logger.Warn("Conflict error in "+operation, zap.Error(err))
			httpResponses.Conflict(c, e.Message, e.Context)
		default:
			h.logger.Error("Domain error in "+operation, zap.Error(err))
			httpResponses.InternalServerError(c, "An internal error occurred")
		}
	default:
		h.logger.Error("Internal error in "+operation, zap.Error(err))
		httpResponses.InternalServerError(c, "An internal error occurred")
	}
}
