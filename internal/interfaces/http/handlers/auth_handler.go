package handlers

import (
	"net/http"

	"eventos-backend/internal/application/dto/requests"
	"eventos-backend/internal/application/dto/responses"
	"eventos-backend/internal/domain/shared/value_objects"
	"eventos-backend/internal/domain/user"
	jwtService "eventos-backend/internal/infrastructure/auth/jwt"
	"eventos-backend/internal/interfaces/http/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthHandler gerencia as operações de autenticação
type AuthHandler struct {
	userService user.Service
	jwtService  jwtService.Service
	logger      *zap.Logger
}

// NewAuthHandler cria uma nova instância do handler de autenticação
func NewAuthHandler(userService user.Service, jwtService jwtService.Service, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		jwtService:  jwtService,
		logger:      logger,
	}
}

// Login autentica um usuário e retorna tokens JWT
//
//	@Summary		Autenticar usuário
//	@Description	Autentica um usuário no sistema usando username/email e senha
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			loginRequest	body		requests.LoginRequest	true	"Dados de login"
//	@Success		200			{object}	responses.LoginResponse	"Login realizado com sucesso"
//	@Failure		400			{object}	map[string]string		"Requisição inválida"
//	@Failure		401			{object}	map[string]string		"Credenciais inválidas"
//	@Failure		500			{object}	map[string]string		"Erro interno do servidor"
//	@Router			/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req requests.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid login request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Validar request
	if err := req.Validate(); err != nil {
		h.logger.Warn("Login request validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Parse tenant ID se fornecido
	var tenantID *value_objects.UUID
	if req.TenantID != "" {
		parsed, err := value_objects.ParseUUID(req.TenantID)
		if err != nil {
			h.logger.Warn("Invalid tenant ID format", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid tenant ID format",
			})
			return
		}
		tenantID = &parsed
	}

	// Autenticar usuário
	authenticatedUser, err := h.userService.AuthenticateUser(
		c.Request.Context(),
		req.UsernameOrEmail,
		req.Password,
		tenantID,
	)
	if err != nil {
		h.logger.Warn("Authentication failed",
			zap.String("username_or_email", req.UsernameOrEmail),
			zap.Error(err),
		)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
		})
		return
	}

	// Gerar tokens
	accessToken, err := h.jwtService.GenerateToken(
		authenticatedUser.ID,
		authenticatedUser.TenantID,
		authenticatedUser.Username,
		authenticatedUser.Email,
	)
	if err != nil {
		h.logger.Error("Failed to generate access token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate token",
		})
		return
	}

	refreshToken, err := h.jwtService.GenerateRefreshToken(
		authenticatedUser.ID,
		authenticatedUser.TenantID,
		authenticatedUser.Username,
		authenticatedUser.Email,
	)
	if err != nil {
		h.logger.Error("Failed to generate refresh token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate refresh token",
		})
		return
	}

	h.logger.Info("User logged in successfully",
		zap.String("user_id", authenticatedUser.ID.String()),
		zap.String("username", authenticatedUser.Username),
		zap.String("tenant_id", authenticatedUser.TenantID.String()),
	)

	// Retornar resposta
	response := responses.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // 1 hora em segundos
		User: responses.UserResponse{
			ID:       authenticatedUser.ID.String(),
			TenantID: authenticatedUser.TenantID.String(),
			FullName: authenticatedUser.FullName,
			Email:    authenticatedUser.Email,
			Username: authenticatedUser.Username,
			Phone:    authenticatedUser.Phone,
			Active:   authenticatedUser.Active,
		},
	}

	c.JSON(http.StatusOK, response)
}

// RefreshToken gera novos tokens a partir de um refresh token válido
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req requests.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid refresh token request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Validar request
	if err := req.Validate(); err != nil {
		h.logger.Warn("Refresh token request validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Gerar novos tokens
	newAccessToken, newRefreshToken, err := h.jwtService.RefreshToken(req.RefreshToken)
	if err != nil {
		h.logger.Warn("Failed to refresh token", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid refresh token",
		})
		return
	}

	h.logger.Info("Token refreshed successfully")

	// Retornar resposta
	response := responses.RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // 1 hora em segundos
	}

	c.JSON(http.StatusOK, response)
}

// Logout invalida o token do usuário (placeholder - em produção seria necessário blacklist)
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if exists {
		h.logger.Info("User logged out",
			zap.String("user_id", userID),
		)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// Me retorna informações do usuário autenticado
func (h *AuthHandler) Me(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Parse user ID
	parsedUserID, err := value_objects.ParseUUID(userID)
	if err != nil {
		h.logger.Error("Invalid user ID in token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	// Buscar usuário
	authenticatedUser, err := h.userService.GetUser(c.Request.Context(), parsedUserID)
	if err != nil {
		h.logger.Error("Failed to get user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user information",
		})
		return
	}

	// Retornar informações do usuário
	response := responses.UserResponse{
		ID:       authenticatedUser.ID.String(),
		TenantID: authenticatedUser.TenantID.String(),
		FullName: authenticatedUser.FullName,
		Email:    authenticatedUser.Email,
		Username: authenticatedUser.Username,
		Phone:    authenticatedUser.Phone,
		Active:   authenticatedUser.Active,
	}

	c.JSON(http.StatusOK, response)
}
