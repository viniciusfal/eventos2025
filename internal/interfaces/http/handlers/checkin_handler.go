package handlers

import (
	"strconv"
	"time"

	"eventos-backend/internal/domain/checkin"
	"eventos-backend/internal/domain/shared/errors"
	"eventos-backend/internal/domain/shared/value_objects"
	jwtService "eventos-backend/internal/infrastructure/auth/jwt"
	httpResponses "eventos-backend/internal/interfaces/http/responses"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CheckinHandler gerencia as operações de check-in
type CheckinHandler struct {
	checkinService checkin.Service
	logger         *zap.Logger
}

// NewCheckinHandler cria uma nova instância do handler de check-in
func NewCheckinHandler(checkinService checkin.Service, logger *zap.Logger) *CheckinHandler {
	return &CheckinHandler{
		checkinService: checkinService,
		logger:         logger,
	}
}

// PerformCheckinRequest representa uma requisição de check-in
type PerformCheckinRequest struct {
	EventID       string    `json:"event_id" binding:"required"`
	EmployeeID    string    `json:"employee_id" binding:"required"`
	PartnerID     string    `json:"partner_id" binding:"required"`
	Method        string    `json:"method" binding:"required"`
	Latitude      float64   `json:"latitude" binding:"required"`
	Longitude     float64   `json:"longitude" binding:"required"`
	PhotoURL      string    `json:"photo_url"`
	Notes         string    `json:"notes"`
	FaceEmbedding []float32 `json:"face_embedding"`
	QRCodeData    string    `json:"qr_code_data"`
}

// CheckinResponse representa a resposta de um check-in
type CheckinResponse struct {
	ID                string                 `json:"id"`
	TenantID          string                 `json:"tenant_id"`
	EventID           string                 `json:"event_id"`
	EmployeeID        string                 `json:"employee_id"`
	PartnerID         string                 `json:"partner_id"`
	Method            string                 `json:"method"`
	Location          LocationResponse       `json:"location"`
	CheckinTime       time.Time              `json:"checkin_time"`
	PhotoURL          string                 `json:"photo_url,omitempty"`
	Notes             string                 `json:"notes,omitempty"`
	IsValid           bool                   `json:"is_valid"`
	ValidationDetails map[string]interface{} `json:"validation_details,omitempty"`
	Status            string                 `json:"status"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	CreatedBy         *string                `json:"created_by,omitempty"`
	UpdatedBy         *string                `json:"updated_by,omitempty"`
}

// ValidationResultResponse representa o resultado de validação
type ValidationResultResponse struct {
	IsValid           bool                   `json:"is_valid"`
	Reason            string                 `json:"reason"`
	Details           map[string]interface{} `json:"details"`
	DistanceFromEvent *float64               `json:"distance_from_event,omitempty"`
	FacialSimilarity  *float64               `json:"facial_similarity,omitempty"`
	WithinBounds      *bool                  `json:"within_bounds,omitempty"`
	Timestamp         time.Time              `json:"timestamp"`
}

// CheckinListResponse representa a resposta de listagem de check-ins
type CheckinListResponse struct {
	Checkins   []CheckinResponse        `json:"checkins"`
	Pagination httpResponses.Pagination `json:"pagination"`
}

// CheckinStatsResponse representa estatísticas de check-ins
type CheckinStatsResponse struct {
	TotalCheckins     int        `json:"total_checkins"`
	ValidCheckins     int        `json:"valid_checkins"`
	InvalidCheckins   int        `json:"invalid_checkins"`
	PendingCheckins   int        `json:"pending_checkins"`
	FacialCheckins    int        `json:"facial_checkins"`
	QRCodeCheckins    int        `json:"qr_code_checkins"`
	ManualCheckins    int        `json:"manual_checkins"`
	CheckinsToday     int        `json:"checkins_today"`
	CheckinsThisWeek  int        `json:"checkins_this_week"`
	CheckinsThisMonth int        `json:"checkins_this_month"`
	AveragePerDay     float64    `json:"average_per_day"`
	LastCheckinTime   *time.Time `json:"last_checkin_time,omitempty"`
}

// PerformCheckin realiza um check-in
func (h *CheckinHandler) PerformCheckin(c *gin.Context) {
	var req PerformCheckinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid perform checkin request", zap.Error(err))
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

	// Parse dos IDs
	eventID, err := value_objects.ParseUUID(req.EventID)
	if err != nil {
		h.logger.Warn("Invalid event ID", zap.String("event_id", req.EventID))
		httpResponses.BadRequest(c, "Invalid event ID", nil)
		return
	}

	employeeID, err := value_objects.ParseUUID(req.EmployeeID)
	if err != nil {
		h.logger.Warn("Invalid employee ID", zap.String("employee_id", req.EmployeeID))
		httpResponses.BadRequest(c, "Invalid employee ID", nil)
		return
	}

	partnerID, err := value_objects.ParseUUID(req.PartnerID)
	if err != nil {
		h.logger.Warn("Invalid partner ID", zap.String("partner_id", req.PartnerID))
		httpResponses.BadRequest(c, "Invalid partner ID", nil)
		return
	}

	// Criar localização
	location, err := value_objects.NewLocation(req.Latitude, req.Longitude)
	if err != nil {
		h.logger.Warn("Invalid location", zap.Float64("latitude", req.Latitude), zap.Float64("longitude", req.Longitude))
		httpResponses.BadRequest(c, "Invalid location coordinates", nil)
		return
	}

	// Criar requisição de check-in
	checkinRequest := checkin.CheckinRequest{
		TenantID:      tenantID,
		EventID:       eventID,
		EmployeeID:    employeeID,
		PartnerID:     partnerID,
		Method:        req.Method,
		Location:      location,
		PhotoURL:      req.PhotoURL,
		Notes:         req.Notes,
		FaceEmbedding: req.FaceEmbedding,
		QRCodeData:    req.QRCodeData,
		CreatedBy:     userID,
	}

	// Realizar check-in
	checkinResult, validationResult, err := h.checkinService.PerformCheckin(c.Request.Context(), checkinRequest)
	if err != nil {
		h.handleServiceError(c, err, "perform checkin")
		return
	}

	// Converter para response
	response := h.toCheckinResponse(checkinResult)
	validationResponse := h.toValidationResultResponse(validationResult)

	h.logger.Info("Checkin performed successfully",
		zap.String("checkin_id", checkinResult.ID.String()),
		zap.String("employee_id", employeeID.String()),
		zap.String("event_id", eventID.String()),
		zap.Bool("is_valid", validationResult.IsValid),
	)

	httpResponses.Created(c, map[string]interface{}{
		"checkin":    response,
		"validation": validationResponse,
	}, "Check-in realizado com sucesso")
}

// GetByID busca um check-in por ID
func (h *CheckinHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	checkinID, err := value_objects.ParseUUID(idParam)
	if err != nil {
		h.logger.Warn("Invalid checkin ID", zap.String("id", idParam))
		httpResponses.BadRequest(c, "Invalid checkin ID", nil)
		return
	}

	checkin, err := h.checkinService.GetCheckin(c.Request.Context(), checkinID)
	if err != nil {
		h.handleServiceError(c, err, "get checkin")
		return
	}

	h.logger.Info("Checkin retrieved successfully", zap.String("checkin_id", checkin.ID.String()))
	httpResponses.Success(c, h.toCheckinResponse(checkin), "Check-in recuperado com sucesso")
}

// List lista check-ins com filtros e paginação
func (h *CheckinHandler) List(c *gin.Context) {
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
	filters := checkin.ListFilters{
		TenantID: &tenantID,
		Page:     1,
		PageSize: 20,
		OrderBy:  "checkin_time",
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

	// Parse do filtro de evento
	if eventIDStr := c.Query("event_id"); eventIDStr != "" {
		if eventID, err := value_objects.ParseUUID(eventIDStr); err == nil {
			filters.EventID = &eventID
		}
	}

	// Parse do filtro de funcionário
	if employeeIDStr := c.Query("employee_id"); employeeIDStr != "" {
		if employeeID, err := value_objects.ParseUUID(employeeIDStr); err == nil {
			filters.EmployeeID = &employeeID
		}
	}

	// Parse do filtro de parceiro
	if partnerIDStr := c.Query("partner_id"); partnerIDStr != "" {
		if partnerID, err := value_objects.ParseUUID(partnerIDStr); err == nil {
			filters.PartnerID = &partnerID
		}
	}

	// Parse do filtro de método
	if method := c.Query("method"); method != "" {
		filters.Method = &method
	}

	// Parse do filtro de válido
	if validStr := c.Query("is_valid"); validStr != "" {
		if valid, err := strconv.ParseBool(validStr); err == nil {
			filters.IsValid = &valid
		}
	}

	// Parse do filtro de foto
	if hasPhotoStr := c.Query("has_photo"); hasPhotoStr != "" {
		if hasPhoto, err := strconv.ParseBool(hasPhotoStr); err == nil {
			filters.HasPhoto = &hasPhoto
		}
	}

	// Parse do filtro de data inicial
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filters.StartDate = &startDate
		}
	}

	// Parse do filtro de data final
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			filters.EndDate = &endDate
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

	// Listar check-ins
	checkins, total, err := h.checkinService.ListCheckins(c.Request.Context(), filters)
	if err != nil {
		h.logger.Error("Failed to list checkins", zap.Error(err))
		httpResponses.InternalServerError(c, "Failed to list checkins")
		return
	}

	// Converter para response
	checkinResponses := make([]CheckinResponse, len(checkins))
	for i, checkin := range checkins {
		checkinResponses[i] = h.toCheckinResponse(checkin)
	}

	// Calcular paginação
	totalPages := (total + filters.PageSize - 1) / filters.PageSize
	pagination := httpResponses.Pagination{
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}

	response := CheckinListResponse{
		Checkins:   checkinResponses,
		Pagination: pagination,
	}

	h.logger.Info("Checkins listed successfully",
		zap.Int("count", len(checkins)),
		zap.Int("total", total),
		zap.Int("page", filters.Page),
	)
	httpResponses.Success(c, response, "Check-ins recuperados com sucesso")
}

// GetByEmployee busca check-ins de um funcionário
func (h *CheckinHandler) GetByEmployee(c *gin.Context) {
	employeeIDStr := c.Param("employee_id")
	employeeID, err := value_objects.ParseUUID(employeeIDStr)
	if err != nil {
		h.logger.Warn("Invalid employee ID", zap.String("employee_id", employeeIDStr))
		httpResponses.BadRequest(c, "Invalid employee ID", nil)
		return
	}

	// Parse dos filtros básicos
	filters := checkin.ListFilters{
		Page:     1,
		PageSize: 20,
		OrderBy:  "checkin_time",
	}

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

	checkins, total, err := h.checkinService.GetEmployeeCheckins(c.Request.Context(), employeeID, filters)
	if err != nil {
		h.handleServiceError(c, err, "get employee checkins")
		return
	}

	// Converter para response
	checkinResponses := make([]CheckinResponse, len(checkins))
	for i, checkin := range checkins {
		checkinResponses[i] = h.toCheckinResponse(checkin)
	}

	// Calcular paginação
	totalPages := (total + filters.PageSize - 1) / filters.PageSize
	pagination := httpResponses.Pagination{
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}

	response := CheckinListResponse{
		Checkins:   checkinResponses,
		Pagination: pagination,
	}

	h.logger.Info("Employee checkins retrieved successfully",
		zap.String("employee_id", employeeID.String()),
		zap.Int("count", len(checkins)),
	)
	httpResponses.Success(c, response, "Check-ins do funcionário recuperados com sucesso")
}

// GetByEvent busca check-ins de um evento
func (h *CheckinHandler) GetByEvent(c *gin.Context) {
	eventIDStr := c.Param("event_id")
	eventID, err := value_objects.ParseUUID(eventIDStr)
	if err != nil {
		h.logger.Warn("Invalid event ID", zap.String("event_id", eventIDStr))
		httpResponses.BadRequest(c, "Invalid event ID", nil)
		return
	}

	// Parse dos filtros básicos
	filters := checkin.ListFilters{
		Page:     1,
		PageSize: 20,
		OrderBy:  "checkin_time",
	}

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

	checkins, total, err := h.checkinService.GetEventCheckins(c.Request.Context(), eventID, filters)
	if err != nil {
		h.handleServiceError(c, err, "get event checkins")
		return
	}

	// Converter para response
	checkinResponses := make([]CheckinResponse, len(checkins))
	for i, checkin := range checkins {
		checkinResponses[i] = h.toCheckinResponse(checkin)
	}

	// Calcular paginação
	totalPages := (total + filters.PageSize - 1) / filters.PageSize
	pagination := httpResponses.Pagination{
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}

	response := CheckinListResponse{
		Checkins:   checkinResponses,
		Pagination: pagination,
	}

	h.logger.Info("Event checkins retrieved successfully",
		zap.String("event_id", eventID.String()),
		zap.Int("count", len(checkins)),
	)
	httpResponses.Success(c, response, "Check-ins do evento recuperados com sucesso")
}

// AddNote adiciona uma observação a um check-in
func (h *CheckinHandler) AddNote(c *gin.Context) {
	idParam := c.Param("id")
	checkinID, err := value_objects.ParseUUID(idParam)
	if err != nil {
		h.logger.Warn("Invalid checkin ID", zap.String("id", idParam))
		httpResponses.BadRequest(c, "Invalid checkin ID", nil)
		return
	}

	var req struct {
		Note string `json:"note" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid add note request", zap.Error(err))
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

	err = h.checkinService.AddCheckinNote(c.Request.Context(), checkinID, req.Note, userID)
	if err != nil {
		h.handleServiceError(c, err, "add checkin note")
		return
	}

	h.logger.Info("Note added to checkin successfully", zap.String("checkin_id", checkinID.String()))
	httpResponses.Success(c, nil, "Observação adicionada com sucesso")
}

// GetStats obtém estatísticas de check-ins
func (h *CheckinHandler) GetStats(c *gin.Context) {
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

	stats, err := h.checkinService.GetCheckinStats(c.Request.Context(), tenantID)
	if err != nil {
		h.handleServiceError(c, err, "get checkin stats")
		return
	}

	response := CheckinStatsResponse{
		TotalCheckins:     stats.TotalCheckins,
		ValidCheckins:     stats.ValidCheckins,
		InvalidCheckins:   stats.InvalidCheckins,
		PendingCheckins:   stats.PendingCheckins,
		FacialCheckins:    stats.FacialCheckins,
		QRCodeCheckins:    stats.QRCodeCheckins,
		ManualCheckins:    stats.ManualCheckins,
		CheckinsToday:     stats.CheckinsToday,
		CheckinsThisWeek:  stats.CheckinsThisWeek,
		CheckinsThisMonth: stats.CheckinsThisMonth,
		AveragePerDay:     stats.AveragePerDay,
		LastCheckinTime:   stats.LastCheckinTime,
	}

	h.logger.Info("Checkin stats retrieved successfully", zap.String("tenant_id", tenantID.String()))
	httpResponses.Success(c, response, "Estatísticas de check-in recuperadas com sucesso")
}

// GetRecent busca check-ins recentes
func (h *CheckinHandler) GetRecent(c *gin.Context) {
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

	// Parse do limite
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	checkins, err := h.checkinService.GetRecentCheckins(c.Request.Context(), tenantID, limit)
	if err != nil {
		h.handleServiceError(c, err, "get recent checkins")
		return
	}

	// Converter para response
	checkinResponses := make([]CheckinResponse, len(checkins))
	for i, checkin := range checkins {
		checkinResponses[i] = h.toCheckinResponse(checkin)
	}

	h.logger.Info("Recent checkins retrieved successfully",
		zap.String("tenant_id", tenantID.String()),
		zap.Int("count", len(checkins)),
	)
	httpResponses.Success(c, checkinResponses, "Check-ins recentes recuperados com sucesso")
}

// toCheckinResponse converte um check-in para CheckinResponse
func (h *CheckinHandler) toCheckinResponse(c *checkin.Checkin) CheckinResponse {
	response := CheckinResponse{
		ID:         c.ID.String(),
		TenantID:   c.TenantID.String(),
		EventID:    c.EventID.String(),
		EmployeeID: c.EmployeeID.String(),
		PartnerID:  c.PartnerID.String(),
		Method:     c.Method,
		Location: LocationResponse{
			Latitude:  c.Location.Latitude,
			Longitude: c.Location.Longitude,
		},
		CheckinTime:       c.CheckinTime,
		PhotoURL:          c.PhotoURL,
		Notes:             c.Notes,
		IsValid:           c.IsValid,
		ValidationDetails: c.ValidationDetails,
		Status:            h.getCheckinStatus(c),
		CreatedAt:         c.CreatedAt,
		UpdatedAt:         c.UpdatedAt,
	}

	// Adicionar CreatedBy se existir
	if c.CreatedBy != nil {
		createdBy := c.CreatedBy.String()
		response.CreatedBy = &createdBy
	}

	// Adicionar UpdatedBy se existir
	if c.UpdatedBy != nil {
		updatedBy := c.UpdatedBy.String()
		response.UpdatedBy = &updatedBy
	}

	return response
}

// toValidationResultResponse converte um ValidationResult para ValidationResultResponse
func (h *CheckinHandler) toValidationResultResponse(vr *checkin.ValidationResult) ValidationResultResponse {
	return ValidationResultResponse{
		IsValid:           vr.IsValid,
		Reason:            vr.Reason,
		Details:           vr.Details,
		DistanceFromEvent: vr.DistanceFromEvent,
		FacialSimilarity:  vr.FacialSimilarity,
		WithinBounds:      vr.WithinBounds,
		Timestamp:         vr.Timestamp,
	}
}

// getCheckinStatus determina o status do check-in
func (h *CheckinHandler) getCheckinStatus(c *checkin.Checkin) string {
	if c.IsValid {
		return "valid"
	}

	// Se tem detalhes de validação mas não é válido, é inválido
	if len(c.ValidationDetails) > 0 {
		return "invalid"
	}

	// Caso contrário, está pendente
	return "pending"
}

// handleServiceError trata erros do serviço de domínio
func (h *CheckinHandler) handleServiceError(c *gin.Context, err error, operation string) {
	h.logger.Error("Checkin service error", zap.Error(err), zap.String("operation", operation))

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
