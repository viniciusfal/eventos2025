package handlers

import (
	"strconv"
	"time"

	"eventos-backend/internal/domain/role"
	"eventos-backend/internal/domain/shared/errors"
	"eventos-backend/internal/domain/shared/value_objects"
	jwtService "eventos-backend/internal/infrastructure/auth/jwt"
	httpResponses "eventos-backend/internal/interfaces/http/responses"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RoleHandler gerencia as operações de role
type RoleHandler struct {
	roleService role.Service
	logger      *zap.Logger
}

// NewRoleHandler cria uma nova instância do handler de role
func NewRoleHandler(roleService role.Service, logger *zap.Logger) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
		logger:      logger,
	}
}

// CreateRoleRequest representa uma requisição de criação de role
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	DisplayName string `json:"display_name" binding:"required"`
	Description string `json:"description"`
	Level       int    `json:"level" binding:"required,min=10,max=999"`
}

// UpdateRoleRequest representa uma requisição de atualização de role
type UpdateRoleRequest struct {
	DisplayName string `json:"display_name" binding:"required"`
	Description string `json:"description"`
	Level       int    `json:"level" binding:"required,min=10,max=999"`
}

// RoleResponse representa a resposta de uma role
type RoleResponse struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id,omitempty"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description"`
	Level       int       `json:"level"`
	IsSystem    bool      `json:"is_system"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedBy   *string   `json:"created_by,omitempty"`
	UpdatedBy   *string   `json:"updated_by,omitempty"`
}

// RoleListResponse representa a resposta de listagem de roles
type RoleListResponse struct {
	Roles      []RoleResponse           `json:"roles"`
	Pagination httpResponses.Pagination `json:"pagination"`
}

// Create cria uma nova role
func (h *RoleHandler) Create(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid create role request", zap.Error(err))
		httpResponses.BadRequest(c, "Invalid request data", map[string]interface{}{
			"validation_errors": err.Error(),
		})
		return
	}

	// Obter informações do usuário autenticado
	userClaims, exists := c.Get("user")
	if !exists {
		h.logger.Error("User claims not found in context")
		httpResponses.Unauthorized(c, "Authentication required")
		return
	}

	claims, ok := userClaims.(*jwtService.Claims)
	if !ok {
		h.logger.Error("Invalid user claims type")
		httpResponses.InternalServerError(c, "Authentication error")
		return
	}

	tenantID, err := value_objects.ParseUUID(claims.TenantID)
	if err != nil {
		h.logger.Error("Invalid tenant ID in claims", zap.Error(err))
		httpResponses.InternalServerError(c, "Invalid tenant ID")
		return
	}

	userID, err := value_objects.ParseUUID(claims.UserID)
	if err != nil {
		h.logger.Error("Invalid user ID in claims", zap.Error(err))
		httpResponses.InternalServerError(c, "Invalid user ID")
		return
	}

	// Criar a role
	newRole, err := h.roleService.CreateRole(
		c.Request.Context(),
		tenantID,
		req.Name,
		req.DisplayName,
		req.Description,
		req.Level,
		userID,
	)
	if err != nil {
		h.logger.Error("Failed to create role", zap.Error(err))
		if domainErr, ok := err.(*errors.DomainError); ok {
			switch domainErr.Type {
			case "ValidationError":
				httpResponses.BadRequest(c, domainErr.Message, domainErr.Context)
			case "AlreadyExistsError":
				httpResponses.Conflict(c, domainErr.Message, domainErr.Context)
			default:
				httpResponses.InternalServerError(c, "Failed to create role")
			}
		} else {
			httpResponses.InternalServerError(c, "Failed to create role")
		}
		return
	}

	h.logger.Info("Role created successfully", zap.String("role_id", newRole.ID.String()))
	httpResponses.Created(c, h.toRoleResponse(newRole), "Role created successfully")
}

// GetByID busca uma role por ID
func (h *RoleHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	roleID, err := value_objects.ParseUUID(idParam)
	if err != nil {
		h.logger.Warn("Invalid role ID", zap.String("id", idParam))
		httpResponses.BadRequest(c, "Invalid role ID", nil)
		return
	}

	role, err := h.roleService.GetRole(c.Request.Context(), roleID)
	if err != nil {
		h.logger.Error("Failed to get role", zap.Error(err), zap.String("role_id", idParam))
		if domainErr, ok := err.(*errors.DomainError); ok {
			switch domainErr.Type {
			case "NotFoundError":
				httpResponses.NotFound(c, "Role not found")
			default:
				httpResponses.InternalServerError(c, "Failed to get role")
			}
		} else {
			httpResponses.InternalServerError(c, "Failed to get role")
		}
		return
	}

	h.logger.Info("Role retrieved successfully", zap.String("role_id", role.ID.String()))
	httpResponses.Success(c, h.toRoleResponse(role), "Role retrieved successfully")
}

// Update atualiza uma role existente
func (h *RoleHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	roleID, err := value_objects.ParseUUID(idParam)
	if err != nil {
		h.logger.Warn("Invalid role ID", zap.String("id", idParam))
		httpResponses.BadRequest(c, "Invalid role ID", nil)
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid update role request", zap.Error(err))
		httpResponses.BadRequest(c, "Invalid request data", map[string]interface{}{
			"validation_errors": err.Error(),
		})
		return
	}

	// Obter informações do usuário autenticado
	userClaims, exists := c.Get("user")
	if !exists {
		h.logger.Error("User claims not found in context")
		httpResponses.Unauthorized(c, "Authentication required")
		return
	}

	claims, ok := userClaims.(*jwtService.Claims)
	if !ok {
		h.logger.Error("Invalid user claims type")
		httpResponses.InternalServerError(c, "Authentication error")
		return
	}

	userID, err := value_objects.ParseUUID(claims.UserID)
	if err != nil {
		h.logger.Error("Invalid user ID in claims", zap.Error(err))
		httpResponses.InternalServerError(c, "Invalid user ID")
		return
	}

	// Atualizar a role
	updatedRole, err := h.roleService.UpdateRole(
		c.Request.Context(),
		roleID,
		req.DisplayName,
		req.Description,
		req.Level,
		userID,
	)
	if err != nil {
		h.logger.Error("Failed to update role", zap.Error(err), zap.String("role_id", idParam))
		if domainErr, ok := err.(*errors.DomainError); ok {
			switch domainErr.Type {
			case "ValidationError":
				httpResponses.BadRequest(c, domainErr.Message, domainErr.Context)
			case "NotFoundError":
				httpResponses.NotFound(c, domainErr.Message)
			case "ForbiddenError":
				httpResponses.Forbidden(c, domainErr.Message)
			default:
				httpResponses.InternalServerError(c, "Failed to update role")
			}
		} else {
			httpResponses.InternalServerError(c, "Failed to update role")
		}
		return
	}

	h.logger.Info("Role updated successfully", zap.String("role_id", updatedRole.ID.String()))
	httpResponses.Success(c, h.toRoleResponse(updatedRole), "Role updated successfully")
}

// Delete remove uma role
func (h *RoleHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	roleID, err := value_objects.ParseUUID(idParam)
	if err != nil {
		h.logger.Warn("Invalid role ID", zap.String("id", idParam))
		httpResponses.BadRequest(c, "Invalid role ID", nil)
		return
	}

	// Obter informações do usuário autenticado
	userClaims, exists := c.Get("user")
	if !exists {
		h.logger.Error("User claims not found in context")
		httpResponses.Unauthorized(c, "Authentication required")
		return
	}

	claims, ok := userClaims.(*jwtService.Claims)
	if !ok {
		h.logger.Error("Invalid user claims type")
		httpResponses.InternalServerError(c, "Authentication error")
		return
	}

	userID, err := value_objects.ParseUUID(claims.UserID)
	if err != nil {
		h.logger.Error("Invalid user ID in claims", zap.Error(err))
		httpResponses.InternalServerError(c, "Invalid user ID")
		return
	}

	// Deletar a role
	err = h.roleService.DeleteRole(c.Request.Context(), roleID, userID)
	if err != nil {
		h.logger.Error("Failed to delete role", zap.Error(err), zap.String("role_id", idParam))
		if domainErr, ok := err.(*errors.DomainError); ok {
			switch domainErr.Type {
			case "NotFoundError":
				httpResponses.NotFound(c, domainErr.Message)
			case "ForbiddenError":
				httpResponses.Forbidden(c, domainErr.Message)
			default:
				httpResponses.InternalServerError(c, "Failed to delete role")
			}
		} else {
			httpResponses.InternalServerError(c, "Failed to delete role")
		}
		return
	}

	h.logger.Info("Role deleted successfully", zap.String("role_id", idParam))
	httpResponses.Success(c, nil, "Role deleted successfully")
}

// List lista roles com filtros e paginação
func (h *RoleHandler) List(c *gin.Context) {
	// Obter informações do usuário autenticado
	userClaims, exists := c.Get("user")
	if !exists {
		h.logger.Error("User claims not found in context")
		httpResponses.Unauthorized(c, "Authentication required")
		return
	}

	claims, ok := userClaims.(*jwtService.Claims)
	if !ok {
		h.logger.Error("Invalid user claims type")
		httpResponses.InternalServerError(c, "Authentication error")
		return
	}

	tenantID, err := value_objects.ParseUUID(claims.TenantID)
	if err != nil {
		h.logger.Error("Invalid tenant ID in claims", zap.Error(err))
		httpResponses.InternalServerError(c, "Invalid tenant ID")
		return
	}

	// Parse dos parâmetros de consulta
	filters := role.ListFilters{
		TenantID: &tenantID,
		Page:     1,
		PageSize: 20,
		OrderBy:  "level",
	}

	// Parse da página
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filters.Page = page
		}
	}

	// Parse do tamanho da página
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 && pageSize <= 100 {
			filters.PageSize = pageSize
		}
	}

	// Parse do filtro de ativo
	if activeStr := c.Query("active"); activeStr != "" {
		if active, err := strconv.ParseBool(activeStr); err == nil {
			filters.Active = &active
		}
	}

	// Parse do filtro de sistema
	if systemStr := c.Query("is_system"); systemStr != "" {
		if isSystem, err := strconv.ParseBool(systemStr); err == nil {
			filters.IsSystem = &isSystem
		}
	}

	// Parse do filtro de nível
	if levelStr := c.Query("level"); levelStr != "" {
		if level, err := strconv.Atoi(levelStr); err == nil && level >= 1 && level <= 999 {
			filters.Level = &level
		}
	}

	// Parse do filtro de nível mínimo
	if minLevelStr := c.Query("min_level"); minLevelStr != "" {
		if minLevel, err := strconv.Atoi(minLevelStr); err == nil && minLevel >= 1 && minLevel <= 999 {
			filters.MinLevel = &minLevel
		}
	}

	// Parse do filtro de nível máximo
	if maxLevelStr := c.Query("max_level"); maxLevelStr != "" {
		if maxLevel, err := strconv.Atoi(maxLevelStr); err == nil && maxLevel >= 1 && maxLevel <= 999 {
			filters.MaxLevel = &maxLevel
		}
	}

	// Parse da busca textual
	if search := c.Query("search"); search != "" {
		filters.Search = &search
	}

	// Parse da ordenação
	if orderBy := c.Query("order_by"); orderBy != "" {
		filters.OrderBy = orderBy
	}

	if orderDesc := c.Query("order_desc"); orderDesc == "true" {
		filters.OrderDesc = true
	}

	// Listar roles do tenant
	roles, total, err := h.roleService.ListTenantRoles(c.Request.Context(), tenantID, filters)
	if err != nil {
		h.logger.Error("Failed to list roles", zap.Error(err))
		httpResponses.InternalServerError(c, "Failed to list roles")
		return
	}

	// Converter para response
	roleResponses := make([]RoleResponse, len(roles))
	for i, role := range roles {
		roleResponses[i] = h.toRoleResponse(role)
	}

	// Calcular paginação
	totalPages := (total + filters.PageSize - 1) / filters.PageSize
	pagination := httpResponses.Pagination{
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}

	response := RoleListResponse{
		Roles:      roleResponses,
		Pagination: pagination,
	}

	h.logger.Info("Roles listed successfully",
		zap.Int("count", len(roles)),
		zap.Int("total", total),
		zap.Int("page", filters.Page),
	)
	httpResponses.Success(c, response, "Roles retrieved successfully")
}

// ListSystem lista roles do sistema
func (h *RoleHandler) ListSystem(c *gin.Context) {
	// Listar roles do sistema
	roles, err := h.roleService.ListSystemRoles(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list system roles", zap.Error(err))
		httpResponses.InternalServerError(c, "Failed to list system roles")
		return
	}

	// Converter para response
	roleResponses := make([]RoleResponse, len(roles))
	for i, role := range roles {
		roleResponses[i] = h.toRoleResponse(role)
	}

	h.logger.Info("System roles listed successfully", zap.Int("count", len(roles)))
	httpResponses.Success(c, roleResponses, "System roles retrieved successfully")
}

// Activate ativa uma role
func (h *RoleHandler) Activate(c *gin.Context) {
	idParam := c.Param("id")
	roleID, err := value_objects.ParseUUID(idParam)
	if err != nil {
		h.logger.Warn("Invalid role ID", zap.String("id", idParam))
		httpResponses.BadRequest(c, "Invalid role ID", nil)
		return
	}

	// Obter informações do usuário autenticado
	userClaims, exists := c.Get("user")
	if !exists {
		h.logger.Error("User claims not found in context")
		httpResponses.Unauthorized(c, "Authentication required")
		return
	}

	claims, ok := userClaims.(*jwtService.Claims)
	if !ok {
		h.logger.Error("Invalid user claims type")
		httpResponses.InternalServerError(c, "Authentication error")
		return
	}

	userID, err := value_objects.ParseUUID(claims.UserID)
	if err != nil {
		h.logger.Error("Invalid user ID in claims", zap.Error(err))
		httpResponses.InternalServerError(c, "Invalid user ID")
		return
	}

	// Ativar a role
	err = h.roleService.ActivateRole(c.Request.Context(), roleID, userID)
	if err != nil {
		h.logger.Error("Failed to activate role", zap.Error(err), zap.String("role_id", idParam))
		if domainErr, ok := err.(*errors.DomainError); ok {
			switch domainErr.Type {
			case "NotFoundError":
				httpResponses.NotFound(c, domainErr.Message)
			case "ForbiddenError":
				httpResponses.Forbidden(c, domainErr.Message)
			default:
				httpResponses.InternalServerError(c, "Failed to activate role")
			}
		} else {
			httpResponses.InternalServerError(c, "Failed to activate role")
		}
		return
	}

	h.logger.Info("Role activated successfully", zap.String("role_id", idParam))
	httpResponses.Success(c, nil, "Role activated successfully")
}

// Deactivate desativa uma role
func (h *RoleHandler) Deactivate(c *gin.Context) {
	idParam := c.Param("id")
	roleID, err := value_objects.ParseUUID(idParam)
	if err != nil {
		h.logger.Warn("Invalid role ID", zap.String("id", idParam))
		httpResponses.BadRequest(c, "Invalid role ID", nil)
		return
	}

	// Obter informações do usuário autenticado
	userClaims, exists := c.Get("user")
	if !exists {
		h.logger.Error("User claims not found in context")
		httpResponses.Unauthorized(c, "Authentication required")
		return
	}

	claims, ok := userClaims.(*jwtService.Claims)
	if !ok {
		h.logger.Error("Invalid user claims type")
		httpResponses.InternalServerError(c, "Authentication error")
		return
	}

	userID, err := value_objects.ParseUUID(claims.UserID)
	if err != nil {
		h.logger.Error("Invalid user ID in claims", zap.Error(err))
		httpResponses.InternalServerError(c, "Invalid user ID")
		return
	}

	// Desativar a role
	err = h.roleService.DeactivateRole(c.Request.Context(), roleID, userID)
	if err != nil {
		h.logger.Error("Failed to deactivate role", zap.Error(err), zap.String("role_id", idParam))
		if domainErr, ok := err.(*errors.DomainError); ok {
			switch domainErr.Type {
			case "NotFoundError":
				httpResponses.NotFound(c, domainErr.Message)
			case "ForbiddenError":
				httpResponses.Forbidden(c, domainErr.Message)
			default:
				httpResponses.InternalServerError(c, "Failed to deactivate role")
			}
		} else {
			httpResponses.InternalServerError(c, "Failed to deactivate role")
		}
		return
	}

	h.logger.Info("Role deactivated successfully", zap.String("role_id", idParam))
	httpResponses.Success(c, nil, "Role deactivated successfully")
}

// GetAvailableLevels retorna os níveis disponíveis para um tenant
func (h *RoleHandler) GetAvailableLevels(c *gin.Context) {
	// Obter informações do usuário autenticado
	userClaims, exists := c.Get("user")
	if !exists {
		h.logger.Error("User claims not found in context")
		httpResponses.Unauthorized(c, "Authentication required")
		return
	}

	claims, ok := userClaims.(*jwtService.Claims)
	if !ok {
		h.logger.Error("Invalid user claims type")
		httpResponses.InternalServerError(c, "Authentication error")
		return
	}

	tenantID, err := value_objects.ParseUUID(claims.TenantID)
	if err != nil {
		h.logger.Error("Invalid tenant ID in claims", zap.Error(err))
		httpResponses.InternalServerError(c, "Invalid tenant ID")
		return
	}

	// Buscar níveis disponíveis
	levels, err := h.roleService.GetAvailableLevels(c.Request.Context(), tenantID)
	if err != nil {
		h.logger.Error("Failed to get available levels", zap.Error(err))
		httpResponses.InternalServerError(c, "Failed to get available levels")
		return
	}

	h.logger.Info("Available levels retrieved successfully", zap.Int("count", len(levels)))
	httpResponses.Success(c, map[string]interface{}{
		"levels": levels,
	}, "Available levels retrieved successfully")
}

// SuggestLevel sugere um nível para uma nova role
func (h *RoleHandler) SuggestLevel(c *gin.Context) {
	// Obter informações do usuário autenticado
	userClaims, exists := c.Get("user")
	if !exists {
		h.logger.Error("User claims not found in context")
		httpResponses.Unauthorized(c, "Authentication required")
		return
	}

	claims, ok := userClaims.(*jwtService.Claims)
	if !ok {
		h.logger.Error("Invalid user claims type")
		httpResponses.InternalServerError(c, "Authentication error")
		return
	}

	tenantID, err := value_objects.ParseUUID(claims.TenantID)
	if err != nil {
		h.logger.Error("Invalid tenant ID in claims", zap.Error(err))
		httpResponses.InternalServerError(c, "Invalid tenant ID")
		return
	}

	// Sugerir nível
	level, err := h.roleService.SuggestLevel(c.Request.Context(), tenantID)
	if err != nil {
		h.logger.Error("Failed to suggest level", zap.Error(err))
		if domainErr, ok := err.(*errors.DomainError); ok {
			switch domainErr.Type {
			case "ValidationError":
				httpResponses.BadRequest(c, domainErr.Message, domainErr.Context)
			default:
				httpResponses.InternalServerError(c, "Failed to suggest level")
			}
		} else {
			httpResponses.InternalServerError(c, "Failed to suggest level")
		}
		return
	}

	h.logger.Info("Level suggested successfully", zap.Int("level", level))
	httpResponses.Success(c, map[string]interface{}{
		"suggested_level": level,
	}, "Level suggested successfully")
}

// toRoleResponse converte uma role para RoleResponse
func (h *RoleHandler) toRoleResponse(r *role.Role) RoleResponse {
	response := RoleResponse{
		ID:          r.ID.String(),
		Name:        r.Name,
		DisplayName: r.DisplayName,
		Description: r.Description,
		Level:       r.Level,
		IsSystem:    r.IsSystem,
		Active:      r.Active,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}

	// Adicionar TenantID apenas se não for role do sistema
	if !r.IsSystem && !r.TenantID.IsZero() {
		response.TenantID = r.TenantID.String()
	}

	// Adicionar CreatedBy se existir
	if r.CreatedBy != nil {
		createdBy := r.CreatedBy.String()
		response.CreatedBy = &createdBy
	}

	// Adicionar UpdatedBy se existir
	if r.UpdatedBy != nil {
		updatedBy := r.UpdatedBy.String()
		response.UpdatedBy = &updatedBy
	}

	return response
}
