package handlers

import (
	"strconv"
	"time"

	"eventos-backend/internal/domain/partner"
	"eventos-backend/internal/domain/shared/errors"
	"eventos-backend/internal/domain/shared/value_objects"
	jwtService "eventos-backend/internal/infrastructure/auth/jwt"
	httpResponses "eventos-backend/internal/interfaces/http/responses"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PartnerHandler gerencia as operações de parceiro
type PartnerHandler struct {
	partnerService partner.Service
	logger         *zap.Logger
}

// NewPartnerHandler cria uma nova instância do handler de parceiro
func NewPartnerHandler(partnerService partner.Service, logger *zap.Logger) *PartnerHandler {
	return &PartnerHandler{
		partnerService: partnerService,
		logger:         logger,
	}
}

// CreatePartnerRequest representa uma requisição de criação de parceiro
type CreatePartnerRequest struct {
	Name         string `json:"name" binding:"required"`
	Email        string `json:"email" binding:"required,email"`
	Email2       string `json:"email2,omitempty"`
	Phone        string `json:"phone" binding:"required"`
	Phone2       string `json:"phone2,omitempty"`
	Identity     string `json:"identity" binding:"required"`
	IdentityType string `json:"identity_type" binding:"required"`
	Location     string `json:"location" binding:"required"`
	Password     string `json:"password" binding:"required,min=8"`
}

// UpdatePartnerRequest representa uma requisição de atualização de parceiro
type UpdatePartnerRequest struct {
	Name         string `json:"name" binding:"required"`
	Email        string `json:"email" binding:"required,email"`
	Email2       string `json:"email2,omitempty"`
	Phone        string `json:"phone" binding:"required"`
	Phone2       string `json:"phone2,omitempty"`
	Identity     string `json:"identity" binding:"required"`
	IdentityType string `json:"identity_type" binding:"required"`
	Location     string `json:"location" binding:"required"`
}

// ChangePasswordRequest representa uma requisição de alteração de senha
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// PartnerLoginRequest representa uma requisição de login de parceiro
type PartnerLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// PartnerResponse representa a resposta de um parceiro
type PartnerResponse struct {
	ID                  string  `json:"id"`
	TenantID            string  `json:"tenant_id"`
	Name                string  `json:"name"`
	Email               string  `json:"email"`
	Email2              string  `json:"email2,omitempty"`
	Phone               string  `json:"phone"`
	Phone2              string  `json:"phone2,omitempty"`
	Identity            string  `json:"identity"`
	IdentityType        string  `json:"identity_type"`
	Location            string  `json:"location"`
	LastLogin           *string `json:"last_login,omitempty"`
	FailedLoginAttempts int     `json:"failed_login_attempts"`
	IsLocked            bool    `json:"is_locked"`
	Active              bool    `json:"active"`
	CreatedAt           string  `json:"created_at"`
	UpdatedAt           string  `json:"updated_at"`
	CreatedBy           *string `json:"created_by,omitempty"`
	UpdatedBy           *string `json:"updated_by,omitempty"`
}

// PartnerListResponse representa a resposta de listagem de parceiros
type PartnerListResponse struct {
	Partners   []PartnerResponse        `json:"partners"`
	Pagination httpResponses.Pagination `json:"pagination"`
}

// PartnerLoginResponse representa a resposta de login de parceiro
type PartnerLoginResponse struct {
	AccessToken  string          `json:"access_token"`
	RefreshToken string          `json:"refresh_token"`
	ExpiresIn    int             `json:"expires_in"`
	Partner      PartnerResponse `json:"partner"`
}

// Create cria um novo parceiro
func (h *PartnerHandler) Create(c *gin.Context) {
	var req CreatePartnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid create partner request", zap.Error(err))
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

	// Criar parceiro
	p, err := h.partnerService.CreatePartner(
		c.Request.Context(),
		tenantID,
		req.Name,
		req.Email,
		req.Email2,
		req.Phone,
		req.Phone2,
		req.Identity,
		req.IdentityType,
		req.Location,
		req.Password,
		userID,
	)
	if err != nil {
		h.handleServiceError(c, err, "create partner")
		return
	}

	response := h.convertToPartnerResponse(p)
	h.logger.Info("Partner created successfully", zap.String("partner_id", p.ID.String()))
	httpResponses.Created(c, response, "Partner created successfully")
}

// GetByID busca um parceiro pelo ID
func (h *PartnerHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := value_objects.ParseUUID(idStr)
	if err != nil {
		h.logger.Warn("Invalid partner ID", zap.String("id", idStr))
		httpResponses.BadRequest(c, "Invalid partner ID format", nil)
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

	p, err := h.partnerService.GetPartnerByTenant(c.Request.Context(), id, tenantID)
	if err != nil {
		h.handleServiceError(c, err, "get partner")
		return
	}

	response := h.convertToPartnerResponse(p)
	httpResponses.Success(c, response, "Partner retrieved successfully")
}

// Update atualiza um parceiro
func (h *PartnerHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := value_objects.ParseUUID(idStr)
	if err != nil {
		h.logger.Warn("Invalid partner ID", zap.String("id", idStr))
		httpResponses.BadRequest(c, "Invalid partner ID format", nil)
		return
	}

	var req UpdatePartnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid update partner request", zap.Error(err))
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

	p, err := h.partnerService.UpdatePartner(
		c.Request.Context(),
		id,
		req.Name,
		req.Email,
		req.Email2,
		req.Phone,
		req.Phone2,
		req.Identity,
		req.IdentityType,
		req.Location,
		userID,
	)
	if err != nil {
		h.handleServiceError(c, err, "update partner")
		return
	}

	response := h.convertToPartnerResponse(p)
	h.logger.Info("Partner updated successfully", zap.String("partner_id", p.ID.String()))
	httpResponses.Success(c, response, "Partner updated successfully")
}

// Delete remove um parceiro (soft delete)
func (h *PartnerHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := value_objects.ParseUUID(idStr)
	if err != nil {
		h.logger.Warn("Invalid partner ID", zap.String("id", idStr))
		httpResponses.BadRequest(c, "Invalid partner ID format", nil)
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

	err = h.partnerService.DeletePartner(c.Request.Context(), id, userID)
	if err != nil {
		h.handleServiceError(c, err, "delete partner")
		return
	}

	h.logger.Info("Partner deleted successfully", zap.String("partner_id", id.String()))
	httpResponses.Success(c, nil, "Partner deleted successfully")
}

// List lista parceiros com paginação e filtros
func (h *PartnerHandler) List(c *gin.Context) {
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

	partners, total, err := h.partnerService.ListPartners(c.Request.Context(), filters)
	if err != nil {
		h.handleServiceError(c, err, "list partners")
		return
	}

	// Converter para resposta
	partnerResponses := make([]PartnerResponse, len(partners))
	for i, p := range partners {
		partnerResponses[i] = h.convertToPartnerResponse(p)
	}

	response := PartnerListResponse{
		Partners: partnerResponses,
		Pagination: httpResponses.Pagination{
			Page:       filters.Page,
			PageSize:   filters.PageSize,
			Total:      total,
			TotalPages: (total + filters.PageSize - 1) / filters.PageSize,
		},
	}

	httpResponses.Success(c, response, "Partners retrieved successfully")
}

// ChangePassword altera a senha de um parceiro
func (h *PartnerHandler) ChangePassword(c *gin.Context) {
	idStr := c.Param("id")
	id, err := value_objects.ParseUUID(idStr)
	if err != nil {
		h.logger.Warn("Invalid partner ID", zap.String("id", idStr))
		httpResponses.BadRequest(c, "Invalid partner ID format", nil)
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid change password request", zap.Error(err))
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

	err = h.partnerService.UpdatePartnerPassword(c.Request.Context(), id, req.NewPassword, userID)
	if err != nil {
		h.handleServiceError(c, err, "change partner password")
		return
	}

	h.logger.Info("Partner password changed successfully", zap.String("partner_id", id.String()))
	httpResponses.Success(c, nil, "Password changed successfully")
}

// Login realiza o login de um parceiro
func (h *PartnerHandler) Login(c *gin.Context) {
	var req PartnerLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid partner login request", zap.Error(err))
		httpResponses.BadRequest(c, "Invalid request data", map[string]interface{}{
			"validation_errors": err.Error(),
		})
		return
	}

	// Realizar login
	p, err := h.partnerService.AuthenticatePartner(c.Request.Context(), req.Email, req.Password, nil)
	if err != nil {
		h.handleServiceError(c, err, "partner login")
		return
	}

	// TODO: Implementar geração de tokens JWT para parceiros
	response := PartnerLoginResponse{
		AccessToken:  "TODO_IMPLEMENT_JWT",
		RefreshToken: "TODO_IMPLEMENT_JWT",
		ExpiresIn:    3600, // 1 hora
		Partner:      h.convertToPartnerResponse(p),
	}

	h.logger.Info("Partner logged in successfully",
		zap.String("partner_id", p.ID.String()),
		zap.String("email", p.Email))
	httpResponses.Success(c, response, "Login successful")
}

// buildListFilters constrói os filtros de listagem a partir dos query parameters
func (h *PartnerHandler) buildListFilters(c *gin.Context) partner.ListFilters {
	filters := partner.ListFilters{
		Page:     1,
		PageSize: 20,
		OrderBy:  "name",
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
	if name := c.Query("name"); name != "" {
		filters.Name = &name
	}

	if email := c.Query("email"); email != "" {
		filters.Email = &email
	}

	if identity := c.Query("identity"); identity != "" {
		filters.Identity = &identity
	}

	if identityType := c.Query("identity_type"); identityType != "" {
		filters.IdentityType = &identityType
	}

	if location := c.Query("location"); location != "" {
		filters.Location = &location
	}

	if activeStr := c.Query("active"); activeStr != "" {
		if active, err := strconv.ParseBool(activeStr); err == nil {
			filters.Active = &active
		}
	}

	// Ordenação
	if orderBy := c.Query("order_by"); orderBy != "" {
		validFields := []string{"name", "email", "identity", "location", "created_at", "updated_at"}
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

// convertToPartnerResponse converte Partner para PartnerResponse
func (h *PartnerHandler) convertToPartnerResponse(p *partner.Partner) PartnerResponse {
	response := PartnerResponse{
		ID:                  p.ID.String(),
		TenantID:            p.TenantID.String(),
		Name:                p.Name,
		Email:               p.Email,
		Email2:              p.Email2,
		Phone:               p.Phone,
		Phone2:              p.Phone2,
		Identity:            p.Identity,
		IdentityType:        p.IdentityType,
		Location:            p.Location,
		FailedLoginAttempts: p.FailedLoginAttempts,
		IsLocked:            p.IsLocked(),
		Active:              p.Active,
		CreatedAt:           p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:           p.UpdatedAt.Format(time.RFC3339),
	}

	if p.LastLogin != nil {
		lastLogin := p.LastLogin.Format(time.RFC3339)
		response.LastLogin = &lastLogin
	}

	if p.CreatedBy != nil {
		createdBy := p.CreatedBy.String()
		response.CreatedBy = &createdBy
	}

	if p.UpdatedBy != nil {
		updatedBy := p.UpdatedBy.String()
		response.UpdatedBy = &updatedBy
	}

	return response
}

// handleServiceError trata erros do serviço de domínio
func (h *PartnerHandler) handleServiceError(c *gin.Context, err error, operation string) {
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
		case "UNAUTHORIZED":
			h.logger.Warn("Unauthorized error in "+operation, zap.Error(err))
			httpResponses.Unauthorized(c, e.Message)
		default:
			h.logger.Error("Domain error in "+operation, zap.Error(err))
			httpResponses.InternalServerError(c, "An internal error occurred")
		}
	default:
		h.logger.Error("Internal error in "+operation, zap.Error(err))
		httpResponses.InternalServerError(c, "An internal error occurred")
	}
}
