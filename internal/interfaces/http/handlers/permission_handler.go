package handlers

import (
	"strconv"
	"time"

	"eventos-backend/internal/domain/permission"
	"eventos-backend/internal/domain/shared/errors"
	"eventos-backend/internal/domain/shared/value_objects"
	jwtService "eventos-backend/internal/infrastructure/auth/jwt"
	httpResponses "eventos-backend/internal/interfaces/http/responses"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PermissionHandler gerencia as operações de permission
type PermissionHandler struct {
	permissionService permission.Service
	logger            *zap.Logger
}

// NewPermissionHandler cria uma nova instância do handler de permission
func NewPermissionHandler(permissionService permission.Service, logger *zap.Logger) *PermissionHandler {
	return &PermissionHandler{
		permissionService: permissionService,
		logger:            logger,
	}
}

// CreatePermissionRequest representa uma requisição de criação de permission
type CreatePermissionRequest struct {
	Module      string `json:"module" binding:"required"`
	Action      string `json:"action" binding:"required"`
	Resource    string `json:"resource"`
	DisplayName string `json:"display_name" binding:"required"`
	Description string `json:"description"`
}

// UpdatePermissionRequest representa uma requisição de atualização de permission
type UpdatePermissionRequest struct {
	DisplayName string `json:"display_name" binding:"required"`
	Description string `json:"description"`
}

// BulkCreatePermissionRequest representa uma requisição de criação em lote de permissions
type BulkCreatePermissionRequest struct {
	Permissions []CreatePermissionRequest `json:"permissions" binding:"required,min=1"`
}

// PermissionResponse representa a resposta de uma permission
type PermissionResponse struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id,omitempty"`
	Module      string    `json:"module"`
	Action      string    `json:"action"`
	Resource    string    `json:"resource,omitempty"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description,omitempty"`
	IsSystem    bool      `json:"is_system"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedBy   *string   `json:"created_by,omitempty"`
	UpdatedBy   *string   `json:"updated_by,omitempty"`
}

// PermissionListResponse representa a resposta de listagem de permissions
type PermissionListResponse struct {
	Permissions []PermissionResponse     `json:"permissions"`
	Pagination  httpResponses.Pagination `json:"pagination"`
}

// Create cria uma nova permission
func (h *PermissionHandler) Create(c *gin.Context) {
	var req CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid create permission request", zap.Error(err))
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

	// Criar a permission
	newPermission, err := h.permissionService.CreatePermission(
		c.Request.Context(),
		tenantID,
		req.Module,
		req.Action,
		req.Resource,
		req.DisplayName,
		req.Description,
		userID,
	)
	if err != nil {
		h.handleServiceError(c, err, "create permission")
		return
	}

	h.logger.Info("Permission created successfully", zap.String("permission_id", newPermission.ID.String()))
	httpResponses.Created(c, h.toPermissionResponse(newPermission), "Permission created successfully")
}

// GetByID busca uma permission por ID
func (h *PermissionHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	permissionID, err := value_objects.ParseUUID(idParam)
	if err != nil {
		h.logger.Warn("Invalid permission ID", zap.String("id", idParam))
		httpResponses.BadRequest(c, "Invalid permission ID", nil)
		return
	}

	permission, err := h.permissionService.GetPermission(c.Request.Context(), permissionID)
	if err != nil {
		h.handleServiceError(c, err, "get permission")
		return
	}

	h.logger.Info("Permission retrieved successfully", zap.String("permission_id", permission.ID.String()))
	httpResponses.Success(c, h.toPermissionResponse(permission), "Permission retrieved successfully")
}

// Update atualiza uma permission existente
func (h *PermissionHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	permissionID, err := value_objects.ParseUUID(idParam)
	if err != nil {
		h.logger.Warn("Invalid permission ID", zap.String("id", idParam))
		httpResponses.BadRequest(c, "Invalid permission ID", nil)
		return
	}

	var req UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid update permission request", zap.Error(err))
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

	// Atualizar a permission
	updatedPermission, err := h.permissionService.UpdatePermission(
		c.Request.Context(),
		permissionID,
		req.DisplayName,
		req.Description,
		userID,
	)
	if err != nil {
		h.handleServiceError(c, err, "update permission")
		return
	}

	h.logger.Info("Permission updated successfully", zap.String("permission_id", updatedPermission.ID.String()))
	httpResponses.Success(c, h.toPermissionResponse(updatedPermission), "Permission updated successfully")
}

// Delete remove uma permission
func (h *PermissionHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	permissionID, err := value_objects.ParseUUID(idParam)
	if err != nil {
		h.logger.Warn("Invalid permission ID", zap.String("id", idParam))
		httpResponses.BadRequest(c, "Invalid permission ID", nil)
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

	// Deletar a permission
	err = h.permissionService.DeletePermission(c.Request.Context(), permissionID, userID)
	if err != nil {
		h.handleServiceError(c, err, "delete permission")
		return
	}

	h.logger.Info("Permission deleted successfully", zap.String("permission_id", idParam))
	httpResponses.Success(c, nil, "Permission deleted successfully")
}

// List lista permissions com filtros e paginação
func (h *PermissionHandler) List(c *gin.Context) {
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
	filters := permission.ListFilters{
		TenantID: &tenantID,
		Page:     1,
		PageSize: 20,
		OrderBy:  "name",
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

	// Parse do filtro de módulo
	if module := c.Query("module"); module != "" {
		filters.Module = &module
	}

	// Parse do filtro de ação
	if action := c.Query("action"); action != "" {
		filters.Action = &action
	}

	// Parse do filtro de recurso
	if resource := c.Query("resource"); resource != "" {
		filters.Resource = &resource
	}

	// Parse do filtro de nome
	if name := c.Query("name"); name != "" {
		filters.Name = &name
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

	// Listar permissions do tenant
	permissions, total, err := h.permissionService.ListTenantPermissions(c.Request.Context(), tenantID, filters)
	if err != nil {
		h.logger.Error("Failed to list permissions", zap.Error(err))
		httpResponses.InternalServerError(c, "Failed to list permissions")
		return
	}

	// Converter para response
	permissionResponses := make([]PermissionResponse, len(permissions))
	for i, permission := range permissions {
		permissionResponses[i] = h.toPermissionResponse(permission)
	}

	// Calcular paginação
	totalPages := (total + filters.PageSize - 1) / filters.PageSize
	pagination := httpResponses.Pagination{
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}

	response := PermissionListResponse{
		Permissions: permissionResponses,
		Pagination:  pagination,
	}

	h.logger.Info("Permissions listed successfully",
		zap.Int("count", len(permissions)),
		zap.Int("total", total),
		zap.Int("page", filters.Page),
	)
	httpResponses.Success(c, response, "Permissions retrieved successfully")
}

// ListSystem lista permissions do sistema
func (h *PermissionHandler) ListSystem(c *gin.Context) {
	// Listar permissions do sistema
	permissions, err := h.permissionService.ListSystemPermissions(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list system permissions", zap.Error(err))
		httpResponses.InternalServerError(c, "Failed to list system permissions")
		return
	}

	// Converter para response
	permissionResponses := make([]PermissionResponse, len(permissions))
	for i, permission := range permissions {
		permissionResponses[i] = h.toPermissionResponse(permission)
	}

	h.logger.Info("System permissions listed successfully", zap.Int("count", len(permissions)))
	httpResponses.Success(c, permissionResponses, "System permissions retrieved successfully")
}

// Activate ativa uma permission
func (h *PermissionHandler) Activate(c *gin.Context) {
	idParam := c.Param("id")
	permissionID, err := value_objects.ParseUUID(idParam)
	if err != nil {
		h.logger.Warn("Invalid permission ID", zap.String("id", idParam))
		httpResponses.BadRequest(c, "Invalid permission ID", nil)
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

	// Ativar a permission
	err = h.permissionService.ActivatePermission(c.Request.Context(), permissionID, userID)
	if err != nil {
		h.handleServiceError(c, err, "activate permission")
		return
	}

	h.logger.Info("Permission activated successfully", zap.String("permission_id", idParam))
	httpResponses.Success(c, nil, "Permission activated successfully")
}

// Deactivate desativa uma permission
func (h *PermissionHandler) Deactivate(c *gin.Context) {
	idParam := c.Param("id")
	permissionID, err := value_objects.ParseUUID(idParam)
	if err != nil {
		h.logger.Warn("Invalid permission ID", zap.String("id", idParam))
		httpResponses.BadRequest(c, "Invalid permission ID", nil)
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

	// Desativar a permission
	err = h.permissionService.DeactivatePermission(c.Request.Context(), permissionID, userID)
	if err != nil {
		h.handleServiceError(c, err, "deactivate permission")
		return
	}

	h.logger.Info("Permission deactivated successfully", zap.String("permission_id", idParam))
	httpResponses.Success(c, nil, "Permission deactivated successfully")
}

// GetModules retorna os módulos disponíveis para um tenant
func (h *PermissionHandler) GetModules(c *gin.Context) {
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

	// Buscar módulos disponíveis
	modules, err := h.permissionService.GetAvailableModules(c.Request.Context(), tenantID)
	if err != nil {
		h.logger.Error("Failed to get available modules", zap.Error(err))
		httpResponses.InternalServerError(c, "Failed to get available modules")
		return
	}

	h.logger.Info("Available modules retrieved successfully", zap.Int("count", len(modules)))
	httpResponses.Success(c, map[string]interface{}{
		"modules": modules,
	}, "Available modules retrieved successfully")
}

// GetActions retorna as ações disponíveis para um módulo
func (h *PermissionHandler) GetActions(c *gin.Context) {
	module := c.Param("module")
	if module == "" {
		h.logger.Warn("Module parameter is required")
		httpResponses.BadRequest(c, "Module parameter is required", nil)
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

	// Buscar ações disponíveis
	actions, err := h.permissionService.GetAvailableActions(c.Request.Context(), tenantID, module)
	if err != nil {
		h.logger.Error("Failed to get available actions", zap.Error(err))
		httpResponses.InternalServerError(c, "Failed to get available actions")
		return
	}

	h.logger.Info("Available actions retrieved successfully", zap.Int("count", len(actions)))
	httpResponses.Success(c, map[string]interface{}{
		"actions": actions,
	}, "Available actions retrieved successfully")
}

// BulkCreate cria múltiplas permissions de uma vez
func (h *PermissionHandler) BulkCreate(c *gin.Context) {
	var req BulkCreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid bulk create permission request", zap.Error(err))
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

	// Converter requests
	requests := make([]permission.CreatePermissionRequest, len(req.Permissions))
	for i, p := range req.Permissions {
		requests[i] = permission.CreatePermissionRequest{
			Module:      p.Module,
			Action:      p.Action,
			Resource:    p.Resource,
			DisplayName: p.DisplayName,
			Description: p.Description,
		}
	}

	// Criar permissions em lote
	permissions, err := h.permissionService.BulkCreatePermissions(c.Request.Context(), tenantID, requests, userID)
	if err != nil {
		h.handleServiceError(c, err, "bulk create permissions")
		return
	}

	// Converter para response
	permissionResponses := make([]PermissionResponse, len(permissions))
	for i, permission := range permissions {
		permissionResponses[i] = h.toPermissionResponse(permission)
	}

	h.logger.Info("Permissions created in bulk successfully", zap.Int("count", len(permissions)))
	httpResponses.Created(c, permissionResponses, "Permissions created successfully")
}

// InitializeSystemPermissions inicializa as permissions padrão do sistema
func (h *PermissionHandler) InitializeSystemPermissions(c *gin.Context) {
	// Inicializar permissions do sistema
	err := h.permissionService.InitializeSystemPermissions(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to initialize system permissions", zap.Error(err))
		httpResponses.InternalServerError(c, "Failed to initialize system permissions")
		return
	}

	h.logger.Info("System permissions initialized successfully")
	httpResponses.Success(c, nil, "System permissions initialized successfully")
}

// toPermissionResponse converte uma permission para PermissionResponse
func (h *PermissionHandler) toPermissionResponse(p *permission.Permission) PermissionResponse {
	response := PermissionResponse{
		ID:          p.ID.String(),
		Module:      p.Module,
		Action:      p.Action,
		Resource:    p.Resource,
		Name:        p.Name,
		DisplayName: p.DisplayName,
		Description: p.Description,
		IsSystem:    p.IsSystem,
		Active:      p.Active,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}

	// Adicionar TenantID apenas se não for permission do sistema
	if !p.IsSystem && !p.TenantID.IsZero() {
		response.TenantID = p.TenantID.String()
	}

	// Adicionar CreatedBy se existir
	if p.CreatedBy != nil {
		createdBy := p.CreatedBy.String()
		response.CreatedBy = &createdBy
	}

	// Adicionar UpdatedBy se existir
	if p.UpdatedBy != nil {
		updatedBy := p.UpdatedBy.String()
		response.UpdatedBy = &updatedBy
	}

	return response
}

// handleServiceError trata erros do serviço de domínio
func (h *PermissionHandler) handleServiceError(c *gin.Context, err error, operation string) {
	h.logger.Error("Permission service error", zap.Error(err), zap.String("operation", operation))

	if domainErr, ok := err.(*errors.DomainError); ok {
		switch domainErr.Type {
		case "ValidationError":
			httpResponses.BadRequest(c, domainErr.Message, domainErr.Context)
		case "AlreadyExistsError":
			httpResponses.Conflict(c, domainErr.Message, domainErr.Context)
		case "NotFoundError", "NOT_FOUND":
			httpResponses.NotFound(c, domainErr.Message)
		case "ForbiddenError":
			httpResponses.Forbidden(c, domainErr.Message)
		default:
			httpResponses.InternalServerError(c, "Failed to "+operation)
		}
	} else {
		httpResponses.InternalServerError(c, "Failed to "+operation)
	}
}
