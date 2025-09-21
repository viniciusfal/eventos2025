package handlers

import (
	"strconv"

	"eventos-backend/internal/application/dto/responses"
	"eventos-backend/internal/domain/shared/value_objects"
	"eventos-backend/internal/domain/tenant"
	"eventos-backend/internal/interfaces/http/middleware"
	httpResponses "eventos-backend/internal/interfaces/http/responses"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// TenantHandler gerencia as operações de tenant
type TenantHandler struct {
	tenantService tenant.Service
	logger        *zap.Logger
}

// NewTenantHandler cria uma nova instância do handler de tenant
func NewTenantHandler(tenantService tenant.Service, logger *zap.Logger) *TenantHandler {
	return &TenantHandler{
		tenantService: tenantService,
		logger:        logger,
	}
}

// CreateTenantRequest representa uma requisição de criação de tenant
type CreateTenantRequest struct {
	Name         string `json:"name" binding:"required"`
	Identity     string `json:"identity,omitempty"`
	IdentityType string `json:"identity_type,omitempty"`
	Email        string `json:"email,omitempty"`
	Address      string `json:"address,omitempty"`
}

// UpdateTenantRequest representa uma requisição de atualização de tenant
type UpdateTenantRequest struct {
	Name         string `json:"name" binding:"required"`
	Identity     string `json:"identity,omitempty"`
	IdentityType string `json:"identity_type,omitempty"`
	Email        string `json:"email,omitempty"`
	Address      string `json:"address,omitempty"`
}

// Create cria um novo tenant
func (h *TenantHandler) Create(c *gin.Context) {
	var req CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid create tenant request", zap.Error(err))
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

	// Criar tenant
	newTenant, err := h.tenantService.CreateTenant(
		c.Request.Context(),
		req.Name,
		req.Identity,
		req.IdentityType,
		req.Email,
		req.Address,
		parsedUserID,
	)
	if err != nil {
		h.logger.Error("Failed to create tenant", zap.Error(err))
		httpResponses.DomainError(c, err)
		return
	}

	h.logger.Info("Tenant created successfully",
		zap.String("tenant_id", newTenant.ID.String()),
		zap.String("name", newTenant.Name),
		zap.String("created_by", userID),
	)

	// Retornar resposta
	response := responses.TenantResponse{
		ID:           newTenant.ID.String(),
		Name:         newTenant.Name,
		Identity:     newTenant.Identity,
		IdentityType: newTenant.IdentityType,
		Email:        newTenant.Email,
		Address:      newTenant.Address,
		Active:       newTenant.Active,
		CreatedAt:    newTenant.CreatedAt,
		UpdatedAt:    newTenant.UpdatedAt,
	}

	httpResponses.Created(c, response, "Tenant created successfully")
}

// GetByID busca um tenant pelo ID
func (h *TenantHandler) GetByID(c *gin.Context) {
	tenantID := c.Param("id")
	if tenantID == "" {
		httpResponses.BadRequest(c, "Tenant ID is required", nil)
		return
	}

	parsedTenantID, err := value_objects.ParseUUID(tenantID)
	if err != nil {
		httpResponses.BadRequest(c, "Invalid tenant ID format", nil)
		return
	}

	// Buscar tenant
	foundTenant, err := h.tenantService.GetTenant(c.Request.Context(), parsedTenantID)
	if err != nil {
		h.logger.Error("Failed to get tenant", zap.Error(err))
		httpResponses.DomainError(c, err)
		return
	}

	// Retornar resposta
	response := responses.TenantResponse{
		ID:           foundTenant.ID.String(),
		Name:         foundTenant.Name,
		Identity:     foundTenant.Identity,
		IdentityType: foundTenant.IdentityType,
		Email:        foundTenant.Email,
		Address:      foundTenant.Address,
		Active:       foundTenant.Active,
		CreatedAt:    foundTenant.CreatedAt,
		UpdatedAt:    foundTenant.UpdatedAt,
	}

	httpResponses.Success(c, response, "")
}

// Update atualiza um tenant
func (h *TenantHandler) Update(c *gin.Context) {
	tenantID := c.Param("id")
	if tenantID == "" {
		httpResponses.BadRequest(c, "Tenant ID is required", nil)
		return
	}

	parsedTenantID, err := value_objects.ParseUUID(tenantID)
	if err != nil {
		httpResponses.BadRequest(c, "Invalid tenant ID format", nil)
		return
	}

	var req UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid update tenant request", zap.Error(err))
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

	// Atualizar tenant
	updatedTenant, err := h.tenantService.UpdateTenant(
		c.Request.Context(),
		parsedTenantID,
		req.Name,
		req.Identity,
		req.IdentityType,
		req.Email,
		req.Address,
		parsedUserID,
	)
	if err != nil {
		h.logger.Error("Failed to update tenant", zap.Error(err))
		httpResponses.DomainError(c, err)
		return
	}

	h.logger.Info("Tenant updated successfully",
		zap.String("tenant_id", updatedTenant.ID.String()),
		zap.String("updated_by", userID),
	)

	// Retornar resposta
	response := responses.TenantResponse{
		ID:           updatedTenant.ID.String(),
		Name:         updatedTenant.Name,
		Identity:     updatedTenant.Identity,
		IdentityType: updatedTenant.IdentityType,
		Email:        updatedTenant.Email,
		Address:      updatedTenant.Address,
		Active:       updatedTenant.Active,
		CreatedAt:    updatedTenant.CreatedAt,
		UpdatedAt:    updatedTenant.UpdatedAt,
	}

	httpResponses.Success(c, response, "Tenant updated successfully")
}

// Delete desativa um tenant
func (h *TenantHandler) Delete(c *gin.Context) {
	tenantID := c.Param("id")
	if tenantID == "" {
		httpResponses.BadRequest(c, "Tenant ID is required", nil)
		return
	}

	parsedTenantID, err := value_objects.ParseUUID(tenantID)
	if err != nil {
		httpResponses.BadRequest(c, "Invalid tenant ID format", nil)
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

	// Desativar tenant
	err = h.tenantService.DeactivateTenant(c.Request.Context(), parsedTenantID, parsedUserID)
	if err != nil {
		h.logger.Error("Failed to deactivate tenant", zap.Error(err))
		httpResponses.DomainError(c, err)
		return
	}

	h.logger.Info("Tenant deactivated successfully",
		zap.String("tenant_id", tenantID),
		zap.String("deactivated_by", userID),
	)

	httpResponses.Success(c, nil, "Tenant deactivated successfully")
}

// List lista tenants com paginação
func (h *TenantHandler) List(c *gin.Context) {
	// Parâmetros de paginação
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// Filtros
	filters := tenant.ListFilters{
		Page:     page,
		PageSize: pageSize,
		OrderBy:  c.DefaultQuery("order_by", "created_at"),
	}

	if orderDesc := c.Query("order_desc"); orderDesc == "true" {
		filters.OrderDesc = true
	}

	if name := c.Query("name"); name != "" {
		filters.Name = &name
	}

	if identity := c.Query("identity"); identity != "" {
		filters.Identity = &identity
	}

	if email := c.Query("email"); email != "" {
		filters.Email = &email
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

	// Listar tenants
	tenants, total, err := h.tenantService.ListTenants(c.Request.Context(), filters)
	if err != nil {
		h.logger.Error("Failed to list tenants", zap.Error(err))
		httpResponses.DomainError(c, err)
		return
	}

	// Converter para response
	var tenantResponses []responses.TenantResponse
	for _, t := range tenants {
		tenantResponses = append(tenantResponses, responses.TenantResponse{
			ID:           t.ID.String(),
			Name:         t.Name,
			Identity:     t.Identity,
			IdentityType: t.IdentityType,
			Email:        t.Email,
			Address:      t.Address,
			Active:       t.Active,
			CreatedAt:    t.CreatedAt,
			UpdatedAt:    t.UpdatedAt,
		})
	}

	// Calcular paginação
	pagination := httpResponses.CalculatePagination(page, pageSize, total)

	httpResponses.Paginated(c, tenantResponses, pagination, "")
}
