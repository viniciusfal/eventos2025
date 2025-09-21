package handlers

import (
	"strconv"
	"time"

	"eventos-backend/internal/domain/checkout"
	"eventos-backend/internal/domain/shared/errors"
	"eventos-backend/internal/domain/shared/value_objects"
	jwtService "eventos-backend/internal/infrastructure/auth/jwt"
	httpResponses "eventos-backend/internal/interfaces/http/responses"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CheckoutHandler gerencia as operações de check-out
type CheckoutHandler struct {
	checkoutService checkout.Service
	logger          *zap.Logger
}

// NewCheckoutHandler cria uma nova instância do handler de check-out
func NewCheckoutHandler(checkoutService checkout.Service, logger *zap.Logger) *CheckoutHandler {
	return &CheckoutHandler{
		checkoutService: checkoutService,
		logger:          logger,
	}
}

// PerformCheckoutRequest representa uma requisição de check-out
type PerformCheckoutRequest struct {
	CheckinID     string    `json:"checkin_id" binding:"required"`
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

// CheckoutResponse representa a resposta de um check-out
type CheckoutResponse struct {
	ID                string                 `json:"id"`
	TenantID          string                 `json:"tenant_id"`
	EventID           string                 `json:"event_id"`
	EmployeeID        string                 `json:"employee_id"`
	PartnerID         string                 `json:"partner_id"`
	CheckinID         string                 `json:"checkin_id"`
	Method            string                 `json:"method"`
	Location          LocationResponse       `json:"location"`
	CheckoutTime      time.Time              `json:"checkout_time"`
	PhotoURL          string                 `json:"photo_url,omitempty"`
	Notes             string                 `json:"notes,omitempty"`
	WorkDuration      string                 `json:"work_duration"` // Formato "2h30m"
	WorkDurationHours float64                `json:"work_duration_hours"`
	IsValid           bool                   `json:"is_valid"`
	ValidationDetails map[string]interface{} `json:"validation_details,omitempty"`
	Status            string                 `json:"status"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	CreatedBy         *string                `json:"created_by,omitempty"`
	UpdatedBy         *string                `json:"updated_by,omitempty"`
}

// CheckoutValidationResult representa o resultado de validação do check-out
type CheckoutValidationResult struct {
	IsValid           bool                   `json:"is_valid"`
	Reason            string                 `json:"reason"`
	Details           map[string]interface{} `json:"details"`
	DistanceFromEvent *float64               `json:"distance_from_event,omitempty"`
	FacialSimilarity  *float64               `json:"facial_similarity,omitempty"`
	WithinBounds      *bool                  `json:"within_bounds,omitempty"`
	WorkDuration      *time.Duration         `json:"work_duration,omitempty"`
	Timestamp         time.Time              `json:"timestamp"`
}

// CheckoutListResponse representa a resposta de listagem de check-outs
type CheckoutListResponse struct {
	Checkouts  []CheckoutResponse       `json:"checkouts"`
	Pagination httpResponses.Pagination `json:"pagination"`
}

// WorkSessionResponse representa uma sessão de trabalho
type WorkSessionResponse struct {
	CheckinID         string    `json:"checkin_id"`
	CheckoutID        string    `json:"checkout_id"`
	EmployeeID        string    `json:"employee_id"`
	EventID           string    `json:"event_id"`
	PartnerID         string    `json:"partner_id"`
	CheckinTime       time.Time `json:"checkin_time"`
	CheckoutTime      time.Time `json:"checkout_time"`
	Duration          string    `json:"duration"` // Formato "2h30m"
	DurationHours     float64   `json:"duration_hours"`
	DurationMinutes   float64   `json:"duration_minutes"`
	IsComplete        bool      `json:"is_complete"`
	IsValid           bool      `json:"is_valid"`
	IsShortSession    bool      `json:"is_short_session"`
	IsOvertimeSession bool      `json:"is_overtime_session"`
}

// WorkSessionListResponse representa a resposta de listagem de sessões de trabalho
type WorkSessionListResponse struct {
	WorkSessions []WorkSessionResponse    `json:"work_sessions"`
	Pagination   httpResponses.Pagination `json:"pagination"`
}

// CheckoutStatsResponse representa estatísticas de check-outs
type CheckoutStatsResponse struct {
	TotalCheckouts     int        `json:"total_checkouts"`
	ValidCheckouts     int        `json:"valid_checkouts"`
	InvalidCheckouts   int        `json:"invalid_checkouts"`
	PendingCheckouts   int        `json:"pending_checkouts"`
	FacialCheckouts    int        `json:"facial_checkouts"`
	QRCodeCheckouts    int        `json:"qr_code_checkouts"`
	ManualCheckouts    int        `json:"manual_checkouts"`
	CheckoutsToday     int        `json:"checkouts_today"`
	CheckoutsThisWeek  int        `json:"checkouts_this_week"`
	CheckoutsThisMonth int        `json:"checkouts_this_month"`
	AveragePerDay      float64    `json:"average_per_day"`
	LastCheckoutTime   *time.Time `json:"last_checkout_time,omitempty"`
}

// WorkStatsResponse representa estatísticas de trabalho
type WorkStatsResponse struct {
	TotalWorkSessions      int        `json:"total_work_sessions"`
	CompletedSessions      int        `json:"completed_sessions"`
	IncompleteSessions     int        `json:"incomplete_sessions"`
	ValidSessions          int        `json:"valid_sessions"`
	InvalidSessions        int        `json:"invalid_sessions"`
	ShortSessions          int        `json:"short_sessions"`
	OvertimeSessions       int        `json:"overtime_sessions"`
	TotalWorkTime          string     `json:"total_work_time"`
	TotalWorkTimeHours     float64    `json:"total_work_time_hours"`
	AverageSessionTime     string     `json:"average_session_time"`
	AverageSessionHours    float64    `json:"average_session_hours"`
	MaxSessionTime         string     `json:"max_session_time"`
	MaxSessionHours        float64    `json:"max_session_hours"`
	MinSessionTime         string     `json:"min_session_time"`
	MinSessionHours        float64    `json:"min_session_hours"`
	WorkTimeToday          string     `json:"work_time_today"`
	WorkTimeTodayHours     float64    `json:"work_time_today_hours"`
	WorkTimeThisWeek       string     `json:"work_time_this_week"`
	WorkTimeThisWeekHours  float64    `json:"work_time_this_week_hours"`
	WorkTimeThisMonth      string     `json:"work_time_this_month"`
	WorkTimeThisMonthHours float64    `json:"work_time_this_month_hours"`
	LastWorkSession        *time.Time `json:"last_work_session,omitempty"`
}

// PerformCheckout realiza um check-out
func (h *CheckoutHandler) PerformCheckout(c *gin.Context) {
	var req PerformCheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid perform checkout request", zap.Error(err))
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
	checkinID, err := value_objects.ParseUUID(req.CheckinID)
	if err != nil {
		h.logger.Warn("Invalid checkin ID", zap.String("checkin_id", req.CheckinID))
		httpResponses.BadRequest(c, "Invalid checkin ID", nil)
		return
	}

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

	// Criar requisição de check-out
	checkoutRequest := checkout.CheckoutRequest{
		TenantID:      tenantID,
		EventID:       eventID,
		EmployeeID:    employeeID,
		PartnerID:     partnerID,
		CheckinID:     checkinID,
		Method:        req.Method,
		Location:      location,
		PhotoURL:      req.PhotoURL,
		Notes:         req.Notes,
		FaceEmbedding: req.FaceEmbedding,
		QRCodeData:    req.QRCodeData,
		CreatedBy:     userID,
	}

	// Realizar check-out
	checkoutResult, validationResult, err := h.checkoutService.PerformCheckout(c.Request.Context(), checkoutRequest)
	if err != nil {
		h.handleServiceError(c, err, "perform checkout")
		return
	}

	// Converter para response
	response := h.toCheckoutResponse(checkoutResult)
	validationResponse := h.toCheckoutValidationResultResponse(validationResult)

	h.logger.Info("Checkout performed successfully",
		zap.String("checkout_id", checkoutResult.ID.String()),
		zap.String("checkin_id", checkinID.String()),
		zap.String("employee_id", employeeID.String()),
		zap.String("event_id", eventID.String()),
		zap.Bool("is_valid", validationResult.IsValid),
		zap.Duration("work_duration", checkoutResult.WorkDuration),
	)

	httpResponses.Created(c, map[string]interface{}{
		"checkout":   response,
		"validation": validationResponse,
	}, "Check-out realizado com sucesso")
}

// GetByID busca um check-out por ID
func (h *CheckoutHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	checkoutID, err := value_objects.ParseUUID(idParam)
	if err != nil {
		h.logger.Warn("Invalid checkout ID", zap.String("id", idParam))
		httpResponses.BadRequest(c, "Invalid checkout ID", nil)
		return
	}

	checkout, err := h.checkoutService.GetCheckout(c.Request.Context(), checkoutID)
	if err != nil {
		h.handleServiceError(c, err, "get checkout")
		return
	}

	h.logger.Info("Checkout retrieved successfully", zap.String("checkout_id", checkout.ID.String()))
	httpResponses.Success(c, h.toCheckoutResponse(checkout), "Check-out recuperado com sucesso")
}

// List lista check-outs com filtros e paginação
func (h *CheckoutHandler) List(c *gin.Context) {
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
	filters := checkout.ListFilters{
		TenantID: &tenantID,
		Page:     1,
		PageSize: 20,
		OrderBy:  "checkout_time",
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

	// Parse do filtro de check-in
	if checkinIDStr := c.Query("checkin_id"); checkinIDStr != "" {
		if checkinID, err := value_objects.ParseUUID(checkinIDStr); err == nil {
			filters.CheckinID = &checkinID
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

	// Parse do filtro de duração mínima
	if minDurationStr := c.Query("min_duration_hours"); minDurationStr != "" {
		if minDuration, err := strconv.ParseFloat(minDurationStr, 64); err == nil {
			filters.MinDurationHours = &minDuration
		}
	}

	// Parse do filtro de duração máxima
	if maxDurationStr := c.Query("max_duration_hours"); maxDurationStr != "" {
		if maxDuration, err := strconv.ParseFloat(maxDurationStr, 64); err == nil {
			filters.MaxDurationHours = &maxDuration
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

	// Listar check-outs
	checkouts, total, err := h.checkoutService.ListCheckouts(c.Request.Context(), filters)
	if err != nil {
		h.logger.Error("Failed to list checkouts", zap.Error(err))
		httpResponses.InternalServerError(c, "Failed to list checkouts")
		return
	}

	// Converter para response
	checkoutResponses := make([]CheckoutResponse, len(checkouts))
	for i, checkout := range checkouts {
		checkoutResponses[i] = h.toCheckoutResponse(checkout)
	}

	// Calcular paginação
	totalPages := (total + filters.PageSize - 1) / filters.PageSize
	pagination := httpResponses.Pagination{
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}

	response := CheckoutListResponse{
		Checkouts:  checkoutResponses,
		Pagination: pagination,
	}

	h.logger.Info("Checkouts listed successfully",
		zap.Int("count", len(checkouts)),
		zap.Int("total", total),
		zap.Int("page", filters.Page),
	)
	httpResponses.Success(c, response, "Check-outs recuperados com sucesso")
}

// GetByEmployee busca check-outs de um funcionário
func (h *CheckoutHandler) GetByEmployee(c *gin.Context) {
	employeeIDStr := c.Param("employee_id")
	employeeID, err := value_objects.ParseUUID(employeeIDStr)
	if err != nil {
		h.logger.Warn("Invalid employee ID", zap.String("employee_id", employeeIDStr))
		httpResponses.BadRequest(c, "Invalid employee ID", nil)
		return
	}

	// Parse dos filtros básicos
	filters := checkout.ListFilters{
		Page:     1,
		PageSize: 20,
		OrderBy:  "checkout_time",
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

	checkouts, total, err := h.checkoutService.GetEmployeeCheckouts(c.Request.Context(), employeeID, filters)
	if err != nil {
		h.handleServiceError(c, err, "get employee checkouts")
		return
	}

	// Converter para response
	checkoutResponses := make([]CheckoutResponse, len(checkouts))
	for i, checkout := range checkouts {
		checkoutResponses[i] = h.toCheckoutResponse(checkout)
	}

	// Calcular paginação
	totalPages := (total + filters.PageSize - 1) / filters.PageSize
	pagination := httpResponses.Pagination{
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}

	response := CheckoutListResponse{
		Checkouts:  checkoutResponses,
		Pagination: pagination,
	}

	h.logger.Info("Employee checkouts retrieved successfully",
		zap.String("employee_id", employeeID.String()),
		zap.Int("count", len(checkouts)),
	)
	httpResponses.Success(c, response, "Check-outs do funcionário recuperados com sucesso")
}

// GetByEvent busca check-outs de um evento
func (h *CheckoutHandler) GetByEvent(c *gin.Context) {
	eventIDStr := c.Param("event_id")
	eventID, err := value_objects.ParseUUID(eventIDStr)
	if err != nil {
		h.logger.Warn("Invalid event ID", zap.String("event_id", eventIDStr))
		httpResponses.BadRequest(c, "Invalid event ID", nil)
		return
	}

	// Parse dos filtros básicos
	filters := checkout.ListFilters{
		Page:     1,
		PageSize: 20,
		OrderBy:  "checkout_time",
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

	checkouts, total, err := h.checkoutService.GetEventCheckouts(c.Request.Context(), eventID, filters)
	if err != nil {
		h.handleServiceError(c, err, "get event checkouts")
		return
	}

	// Converter para response
	checkoutResponses := make([]CheckoutResponse, len(checkouts))
	for i, checkout := range checkouts {
		checkoutResponses[i] = h.toCheckoutResponse(checkout)
	}

	// Calcular paginação
	totalPages := (total + filters.PageSize - 1) / filters.PageSize
	pagination := httpResponses.Pagination{
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}

	response := CheckoutListResponse{
		Checkouts:  checkoutResponses,
		Pagination: pagination,
	}

	h.logger.Info("Event checkouts retrieved successfully",
		zap.String("event_id", eventID.String()),
		zap.Int("count", len(checkouts)),
	)
	httpResponses.Success(c, response, "Check-outs do evento recuperados com sucesso")
}

// GetWorkSessions busca sessões de trabalho
func (h *CheckoutHandler) GetWorkSessions(c *gin.Context) {
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
	filters := checkout.WorkSessionFilters{
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

	// Parse do filtro de duração mínima
	if minDurationStr := c.Query("min_duration_hours"); minDurationStr != "" {
		if minDuration, err := strconv.ParseFloat(minDurationStr, 64); err == nil {
			filters.MinDurationHours = &minDuration
		}
	}

	// Parse do filtro de duração máxima
	if maxDurationStr := c.Query("max_duration_hours"); maxDurationStr != "" {
		if maxDuration, err := strconv.ParseFloat(maxDurationStr, 64); err == nil {
			filters.MaxDurationHours = &maxDuration
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

	// Parse da ordenação
	if orderBy := c.Query("order_by"); orderBy != "" {
		filters.OrderBy = orderBy
	}

	if orderDesc := c.Query("order_desc"); orderDesc == "true" {
		filters.OrderDesc = true
	}

	// Buscar sessões de trabalho
	workSessions, total, err := h.checkoutService.GetWorkSessions(c.Request.Context(), tenantID, filters)
	if err != nil {
		h.logger.Error("Failed to get work sessions", zap.Error(err))
		httpResponses.InternalServerError(c, "Failed to get work sessions")
		return
	}

	// Converter para response
	workSessionResponses := make([]WorkSessionResponse, len(workSessions))
	for i, ws := range workSessions {
		workSessionResponses[i] = h.toWorkSessionResponse(ws)
	}

	// Calcular paginação
	totalPages := (total + filters.PageSize - 1) / filters.PageSize
	pagination := httpResponses.Pagination{
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}

	response := WorkSessionListResponse{
		WorkSessions: workSessionResponses,
		Pagination:   pagination,
	}

	h.logger.Info("Work sessions retrieved successfully",
		zap.Int("count", len(workSessions)),
		zap.Int("total", total),
		zap.Int("page", filters.Page),
	)
	httpResponses.Success(c, response, "Sessões de trabalho recuperadas com sucesso")
}

// GetEmployeeWorkSessions busca sessões de trabalho de um funcionário
func (h *CheckoutHandler) GetEmployeeWorkSessions(c *gin.Context) {
	employeeIDStr := c.Param("employee_id")
	employeeID, err := value_objects.ParseUUID(employeeIDStr)
	if err != nil {
		h.logger.Warn("Invalid employee ID", zap.String("employee_id", employeeIDStr))
		httpResponses.BadRequest(c, "Invalid employee ID", nil)
		return
	}

	// Parse dos filtros básicos
	filters := checkout.WorkSessionFilters{
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

	workSessions, total, err := h.checkoutService.GetEmployeeWorkSessions(c.Request.Context(), employeeID, filters)
	if err != nil {
		h.handleServiceError(c, err, "get employee work sessions")
		return
	}

	// Converter para response
	workSessionResponses := make([]WorkSessionResponse, len(workSessions))
	for i, ws := range workSessions {
		workSessionResponses[i] = h.toWorkSessionResponse(ws)
	}

	// Calcular paginação
	totalPages := (total + filters.PageSize - 1) / filters.PageSize
	pagination := httpResponses.Pagination{
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}

	response := WorkSessionListResponse{
		WorkSessions: workSessionResponses,
		Pagination:   pagination,
	}

	h.logger.Info("Employee work sessions retrieved successfully",
		zap.String("employee_id", employeeID.String()),
		zap.Int("count", len(workSessions)),
	)
	httpResponses.Success(c, response, "Sessões de trabalho do funcionário recuperadas com sucesso")
}

// AddNote adiciona uma observação a um check-out
func (h *CheckoutHandler) AddNote(c *gin.Context) {
	idParam := c.Param("id")
	checkoutID, err := value_objects.ParseUUID(idParam)
	if err != nil {
		h.logger.Warn("Invalid checkout ID", zap.String("id", idParam))
		httpResponses.BadRequest(c, "Invalid checkout ID", nil)
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

	err = h.checkoutService.AddCheckoutNote(c.Request.Context(), checkoutID, req.Note, userID)
	if err != nil {
		h.handleServiceError(c, err, "add checkout note")
		return
	}

	h.logger.Info("Note added to checkout successfully", zap.String("checkout_id", checkoutID.String()))
	httpResponses.Success(c, nil, "Observação adicionada com sucesso")
}

// GetStats obtém estatísticas de check-outs
func (h *CheckoutHandler) GetStats(c *gin.Context) {
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

	stats, err := h.checkoutService.GetCheckoutStats(c.Request.Context(), tenantID)
	if err != nil {
		h.handleServiceError(c, err, "get checkout stats")
		return
	}

	response := CheckoutStatsResponse{
		TotalCheckouts:     stats.TotalCheckouts,
		ValidCheckouts:     stats.ValidCheckouts,
		InvalidCheckouts:   stats.InvalidCheckouts,
		PendingCheckouts:   stats.PendingCheckouts,
		FacialCheckouts:    stats.FacialCheckouts,
		QRCodeCheckouts:    stats.QRCodeCheckouts,
		ManualCheckouts:    stats.ManualCheckouts,
		CheckoutsToday:     stats.CheckoutsToday,
		CheckoutsThisWeek:  stats.CheckoutsThisWeek,
		CheckoutsThisMonth: stats.CheckoutsThisMonth,
		AveragePerDay:      0, // Campo não disponível nas estatísticas atuais
		LastCheckoutTime:   stats.LastCheckoutTime,
	}

	h.logger.Info("Checkout stats retrieved successfully", zap.String("tenant_id", tenantID.String()))
	httpResponses.Success(c, response, "Estatísticas de check-out recuperadas com sucesso")
}

// GetWorkStats obtém estatísticas de trabalho
func (h *CheckoutHandler) GetWorkStats(c *gin.Context) {
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

	stats, err := h.checkoutService.GetWorkStats(c.Request.Context(), tenantID)
	if err != nil {
		h.handleServiceError(c, err, "get work stats")
		return
	}

	response := WorkStatsResponse{
		TotalWorkSessions:      stats.TotalSessions,
		CompletedSessions:      stats.CompleteSessions,
		IncompleteSessions:     stats.IncompleteSessions,
		ValidSessions:          stats.ValidSessions,
		InvalidSessions:        stats.InvalidSessions,
		ShortSessions:          stats.ShortSessions,
		OvertimeSessions:       stats.LongSessions, // Usando LongSessions como proxy para overtime
		TotalWorkTime:          time.Duration(stats.TotalWorkHours * float64(time.Hour)).String(),
		TotalWorkTimeHours:     stats.TotalWorkHours,
		AverageSessionTime:     time.Duration(stats.AverageWorkHours * float64(time.Hour)).String(),
		AverageSessionHours:    stats.AverageWorkHours,
		MaxSessionTime:         time.Duration(stats.MaxWorkHours * float64(time.Hour)).String(),
		MaxSessionHours:        stats.MaxWorkHours,
		MinSessionTime:         time.Duration(stats.MinWorkHours * float64(time.Hour)).String(),
		MinSessionHours:        stats.MinWorkHours,
		WorkTimeToday:          "0h", // Campo não disponível nas estatísticas atuais
		WorkTimeTodayHours:     0,
		WorkTimeThisWeek:       "0h", // Campo não disponível nas estatísticas atuais
		WorkTimeThisWeekHours:  0,
		WorkTimeThisMonth:      "0h", // Campo não disponível nas estatísticas atuais
		WorkTimeThisMonthHours: 0,
		LastWorkSession:        nil, // Campo não disponível nas estatísticas atuais
	}

	h.logger.Info("Work stats retrieved successfully", zap.String("tenant_id", tenantID.String()))
	httpResponses.Success(c, response, "Estatísticas de trabalho recuperadas com sucesso")
}

// GetRecent busca check-outs recentes
func (h *CheckoutHandler) GetRecent(c *gin.Context) {
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

	checkouts, err := h.checkoutService.GetRecentCheckouts(c.Request.Context(), tenantID, limit)
	if err != nil {
		h.handleServiceError(c, err, "get recent checkouts")
		return
	}

	// Converter para response
	checkoutResponses := make([]CheckoutResponse, len(checkouts))
	for i, checkout := range checkouts {
		checkoutResponses[i] = h.toCheckoutResponse(checkout)
	}

	h.logger.Info("Recent checkouts retrieved successfully",
		zap.String("tenant_id", tenantID.String()),
		zap.Int("count", len(checkouts)),
	)
	httpResponses.Success(c, checkoutResponses, "Check-outs recentes recuperados com sucesso")
}

// toCheckoutResponse converte um check-out para CheckoutResponse
func (h *CheckoutHandler) toCheckoutResponse(c *checkout.Checkout) CheckoutResponse {
	response := CheckoutResponse{
		ID:         c.ID.String(),
		TenantID:   c.TenantID.String(),
		EventID:    c.EventID.String(),
		EmployeeID: c.EmployeeID.String(),
		PartnerID:  c.PartnerID.String(),
		CheckinID:  c.CheckinID.String(),
		Method:     c.Method,
		Location: LocationResponse{
			Latitude:  c.Location.Latitude,
			Longitude: c.Location.Longitude,
		},
		CheckoutTime:      c.CheckoutTime,
		PhotoURL:          c.PhotoURL,
		Notes:             c.Notes,
		WorkDuration:      c.WorkDuration.String(),
		WorkDurationHours: c.WorkDuration.Hours(),
		IsValid:           c.IsValid,
		ValidationDetails: c.ValidationDetails,
		Status:            h.getCheckoutStatus(c),
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

// toCheckoutValidationResultResponse converte um ValidationResult para CheckoutValidationResult
func (h *CheckoutHandler) toCheckoutValidationResultResponse(vr *checkout.ValidationResult) CheckoutValidationResult {
	return CheckoutValidationResult{
		IsValid:           vr.IsValid,
		Reason:            vr.Reason,
		Details:           vr.Details,
		DistanceFromEvent: vr.DistanceFromEvent,
		FacialSimilarity:  vr.FacialSimilarity,
		WithinBounds:      vr.WithinBounds,
		WorkDuration:      vr.WorkDuration,
		Timestamp:         vr.Timestamp,
	}
}

// toWorkSessionResponse converte uma WorkSession para WorkSessionResponse
func (h *CheckoutHandler) toWorkSessionResponse(ws *checkout.WorkSession) WorkSessionResponse {
	return WorkSessionResponse{
		CheckinID:         ws.CheckinID.String(),
		CheckoutID:        ws.CheckoutID.String(),
		EmployeeID:        ws.EmployeeID.String(),
		EventID:           ws.EventID.String(),
		PartnerID:         ws.PartnerID.String(),
		CheckinTime:       ws.CheckinTime,
		CheckoutTime:      ws.CheckoutTime,
		Duration:          ws.Duration.String(),
		DurationHours:     ws.GetDurationHours(),
		DurationMinutes:   ws.GetDurationMinutes(),
		IsComplete:        ws.IsComplete,
		IsValid:           ws.IsValid,
		IsShortSession:    ws.IsShortSession(),
		IsOvertimeSession: ws.IsLongSession(),
	}
}

// getCheckoutStatus determina o status do check-out
func (h *CheckoutHandler) getCheckoutStatus(c *checkout.Checkout) string {
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
func (h *CheckoutHandler) handleServiceError(c *gin.Context, err error, operation string) {
	h.logger.Error("Checkout service error", zap.Error(err), zap.String("operation", operation))

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
