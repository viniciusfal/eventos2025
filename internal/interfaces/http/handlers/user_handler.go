package handlers

import (
	"strconv"

	"eventos-backend/internal/application/dto/requests"
	"eventos-backend/internal/application/dto/responses"
	"eventos-backend/internal/domain/shared/value_objects"
	"eventos-backend/internal/domain/user"
	"eventos-backend/internal/interfaces/http/middleware"
	httpResponses "eventos-backend/internal/interfaces/http/responses"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// UserHandler gerencia as operações de usuário
type UserHandler struct {
	userService user.Service
	logger      *zap.Logger
}

// NewUserHandler cria uma nova instância do handler de usuário
func NewUserHandler(userService user.Service, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

// CreateUserRequest representa uma requisição de criação de usuário
type CreateUserRequest struct {
	TenantID string `json:"tenant_id" binding:"required"`
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone,omitempty"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

// UpdateUserRequest representa uma requisição de atualização de usuário
type UpdateUserRequest struct {
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone,omitempty"`
	Username string `json:"username" binding:"required"`
}

// Create cria um novo usuário
func (h *UserHandler) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid create user request", zap.Error(err))
		httpResponses.BadRequest(c, "Invalid request format", map[string]interface{}{
			"validation_errors": err.Error(),
		})
		return
	}

	// Obter usuário autenticado
	userID, exists := middleware.GetUserID(c)
	if !exists {
		httpResponses.Unauthorized(c, "User not authenticated")
		return
	}

	parsedUserID, err := value_objects.ParseUUID(userID)
	if err != nil {
		h.logger.Error("Invalid user ID in token", zap.Error(err))
		httpResponses.InternalServerError(c, "Invalid user ID")
		return
	}

	// Parse tenant ID
	parsedTenantID, err := value_objects.ParseUUID(req.TenantID)
	if err != nil {
		httpResponses.BadRequest(c, "Invalid tenant ID format", nil)
		return
	}

	// Criar usuário
	newUser, err := h.userService.CreateUser(
		c.Request.Context(),
		parsedTenantID,
		req.FullName,
		req.Email,
		req.Phone,
		req.Username,
		req.Password,
		parsedUserID,
	)
	if err != nil {
		h.logger.Error("Failed to create user", zap.Error(err))
		httpResponses.DomainError(c, err)
		return
	}

	h.logger.Info("User created successfully",
		zap.String("user_id", newUser.ID.String()),
		zap.String("username", newUser.Username),
		zap.String("created_by", userID),
	)

	// Retornar resposta
	response := responses.UserResponse{
		ID:        newUser.ID.String(),
		TenantID:  newUser.TenantID.String(),
		FullName:  newUser.FullName,
		Email:     newUser.Email,
		Username:  newUser.Username,
		Phone:     newUser.Phone,
		Active:    newUser.Active,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
	}

	httpResponses.Created(c, response, "User created successfully")
}

// GetByID busca um usuário pelo ID
func (h *UserHandler) GetByID(c *gin.Context) {
	userIDParam := c.Param("id")
	if userIDParam == "" {
		httpResponses.BadRequest(c, "User ID is required", nil)
		return
	}

	parsedUserID, err := value_objects.ParseUUID(userIDParam)
	if err != nil {
		httpResponses.BadRequest(c, "Invalid user ID format", nil)
		return
	}

	// Buscar usuário
	foundUser, err := h.userService.GetUser(c.Request.Context(), parsedUserID)
	if err != nil {
		h.logger.Error("Failed to get user", zap.Error(err))
		httpResponses.DomainError(c, err)
		return
	}

	// Retornar resposta
	response := responses.UserResponse{
		ID:        foundUser.ID.String(),
		TenantID:  foundUser.TenantID.String(),
		FullName:  foundUser.FullName,
		Email:     foundUser.Email,
		Username:  foundUser.Username,
		Phone:     foundUser.Phone,
		Active:    foundUser.Active,
		CreatedAt: foundUser.CreatedAt,
		UpdatedAt: foundUser.UpdatedAt,
	}

	httpResponses.Success(c, response, "")
}

// Update atualiza um usuário
func (h *UserHandler) Update(c *gin.Context) {
	userIDParam := c.Param("id")
	if userIDParam == "" {
		httpResponses.BadRequest(c, "User ID is required", nil)
		return
	}

	parsedUserID, err := value_objects.ParseUUID(userIDParam)
	if err != nil {
		httpResponses.BadRequest(c, "Invalid user ID format", nil)
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid update user request", zap.Error(err))
		httpResponses.BadRequest(c, "Invalid request format", map[string]interface{}{
			"validation_errors": err.Error(),
		})
		return
	}

	// Obter usuário autenticado
	authenticatedUserID, exists := middleware.GetUserID(c)
	if !exists {
		httpResponses.Unauthorized(c, "User not authenticated")
		return
	}

	parsedAuthUserID, err := value_objects.ParseUUID(authenticatedUserID)
	if err != nil {
		h.logger.Error("Invalid authenticated user ID in token", zap.Error(err))
		httpResponses.InternalServerError(c, "Invalid user ID")
		return
	}

	// Atualizar usuário
	updatedUser, err := h.userService.UpdateUser(
		c.Request.Context(),
		parsedUserID,
		req.FullName,
		req.Email,
		req.Phone,
		req.Username,
		parsedAuthUserID,
	)
	if err != nil {
		h.logger.Error("Failed to update user", zap.Error(err))
		httpResponses.DomainError(c, err)
		return
	}

	h.logger.Info("User updated successfully",
		zap.String("user_id", updatedUser.ID.String()),
		zap.String("updated_by", authenticatedUserID),
	)

	// Retornar resposta
	response := responses.UserResponse{
		ID:        updatedUser.ID.String(),
		TenantID:  updatedUser.TenantID.String(),
		FullName:  updatedUser.FullName,
		Email:     updatedUser.Email,
		Username:  updatedUser.Username,
		Phone:     updatedUser.Phone,
		Active:    updatedUser.Active,
		CreatedAt: updatedUser.CreatedAt,
		UpdatedAt: updatedUser.UpdatedAt,
	}

	httpResponses.Success(c, response, "User updated successfully")
}

// Delete desativa um usuário
func (h *UserHandler) Delete(c *gin.Context) {
	userIDParam := c.Param("id")
	if userIDParam == "" {
		httpResponses.BadRequest(c, "User ID is required", nil)
		return
	}

	parsedUserID, err := value_objects.ParseUUID(userIDParam)
	if err != nil {
		httpResponses.BadRequest(c, "Invalid user ID format", nil)
		return
	}

	// Obter usuário autenticado
	authenticatedUserID, exists := middleware.GetUserID(c)
	if !exists {
		httpResponses.Unauthorized(c, "User not authenticated")
		return
	}

	parsedAuthUserID, err := value_objects.ParseUUID(authenticatedUserID)
	if err != nil {
		h.logger.Error("Invalid authenticated user ID in token", zap.Error(err))
		httpResponses.InternalServerError(c, "Invalid user ID")
		return
	}

	// Desativar usuário
	err = h.userService.DeactivateUser(c.Request.Context(), parsedUserID, parsedAuthUserID)
	if err != nil {
		h.logger.Error("Failed to deactivate user", zap.Error(err))
		httpResponses.DomainError(c, err)
		return
	}

	h.logger.Info("User deactivated successfully",
		zap.String("user_id", userIDParam),
		zap.String("deactivated_by", authenticatedUserID),
	)

	httpResponses.Success(c, nil, "User deactivated successfully")
}

// List lista usuários com paginação
func (h *UserHandler) List(c *gin.Context) {
	// Parâmetros de paginação
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// Filtros
	filters := user.ListFilters{
		Page:     page,
		PageSize: pageSize,
		OrderBy:  c.DefaultQuery("order_by", "created_at"),
	}

	if orderDesc := c.Query("order_desc"); orderDesc == "true" {
		filters.OrderDesc = true
	}

	if tenantID := c.Query("tenant_id"); tenantID != "" {
		parsedTenantID, err := value_objects.ParseUUID(tenantID)
		if err != nil {
			httpResponses.BadRequest(c, "Invalid tenant ID format", nil)
			return
		}
		filters.TenantID = &parsedTenantID
	}

	if fullName := c.Query("full_name"); fullName != "" {
		filters.FullName = &fullName
	}

	if email := c.Query("email"); email != "" {
		filters.Email = &email
	}

	if username := c.Query("username"); username != "" {
		filters.Username = &username
	}

	if active := c.Query("active"); active != "" {
		if active == "true" {
			activeValue := true
			filters.Active = &activeValue
		} else if active == "false" {
			activeValue := false
			filters.Active = &activeValue
		}
	}

	// Validar filtros
	if err := filters.Validate(); err != nil {
		httpResponses.BadRequest(c, "Invalid filters", map[string]interface{}{
			"validation_errors": err.Error(),
		})
		return
	}

	// Listar usuários
	users, total, err := h.userService.ListUsers(c.Request.Context(), filters)
	if err != nil {
		h.logger.Error("Failed to list users", zap.Error(err))
		httpResponses.DomainError(c, err)
		return
	}

	// Converter para response
	var userResponses []responses.UserResponse
	for _, u := range users {
		userResponses = append(userResponses, responses.UserResponse{
			ID:        u.ID.String(),
			TenantID:  u.TenantID.String(),
			FullName:  u.FullName,
			Email:     u.Email,
			Username:  u.Username,
			Phone:     u.Phone,
			Active:    u.Active,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
		})
	}

	// Calcular paginação
	pagination := httpResponses.CalculatePagination(page, pageSize, total)

	httpResponses.Paginated(c, userResponses, pagination, "")
}

// ChangePassword altera a senha de um usuário
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userIDParam := c.Param("id")
	if userIDParam == "" {
		httpResponses.BadRequest(c, "User ID is required", nil)
		return
	}

	parsedUserID, err := value_objects.ParseUUID(userIDParam)
	if err != nil {
		httpResponses.BadRequest(c, "Invalid user ID format", nil)
		return
	}

	var req requests.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid change password request", zap.Error(err))
		httpResponses.BadRequest(c, "Invalid request format", map[string]interface{}{
			"validation_errors": err.Error(),
		})
		return
	}

	// Validar request
	if err := req.Validate(); err != nil {
		httpResponses.BadRequest(c, err.Error(), nil)
		return
	}

	// Obter usuário autenticado
	authenticatedUserID, exists := middleware.GetUserID(c)
	if !exists {
		httpResponses.Unauthorized(c, "User not authenticated")
		return
	}

	parsedAuthUserID, err := value_objects.ParseUUID(authenticatedUserID)
	if err != nil {
		h.logger.Error("Invalid authenticated user ID in token", zap.Error(err))
		httpResponses.InternalServerError(c, "Invalid user ID")
		return
	}

	// Buscar usuário para validar senha atual
	user, err := h.userService.GetUser(c.Request.Context(), parsedUserID)
	if err != nil {
		h.logger.Error("Failed to get user for password change", zap.Error(err))
		httpResponses.DomainError(c, err)
		return
	}

	// Verificar senha atual
	if !user.CheckPassword(req.CurrentPassword) {
		httpResponses.BadRequest(c, "Current password is incorrect", nil)
		return
	}

	// Alterar senha
	err = h.userService.UpdateUserPassword(
		c.Request.Context(),
		parsedUserID,
		req.NewPassword,
		parsedAuthUserID,
	)
	if err != nil {
		h.logger.Error("Failed to change password", zap.Error(err))
		httpResponses.DomainError(c, err)
		return
	}

	h.logger.Info("Password changed successfully",
		zap.String("user_id", userIDParam),
		zap.String("changed_by", authenticatedUserID),
	)

	httpResponses.Success(c, nil, "Password changed successfully")
}
