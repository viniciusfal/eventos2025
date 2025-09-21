package handlers

import (
	"strconv"
	"time"

	"eventos-backend/internal/domain/event"
	"eventos-backend/internal/domain/shared/errors"
	"eventos-backend/internal/domain/shared/value_objects"
	jwtService "eventos-backend/internal/infrastructure/auth/jwt"
	httpResponses "eventos-backend/internal/interfaces/http/responses"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// EventHandler gerencia as operações de evento
type EventHandler struct {
	eventService event.Service
	logger       *zap.Logger
}

// NewEventHandler cria uma nova instância do handler de evento
func NewEventHandler(eventService event.Service, logger *zap.Logger) *EventHandler {
	return &EventHandler{
		eventService: eventService,
		logger:       logger,
	}
}

// CreateEventRequest representa uma requisição de criação de evento
type CreateEventRequest struct {
	Name        string            `json:"name" binding:"required"`
	Location    string            `json:"location" binding:"required"`
	FenceEvent  []LocationRequest `json:"fence_event" binding:"required,min=3"`
	InitialDate string            `json:"initial_date" binding:"required"`
	FinalDate   string            `json:"final_date" binding:"required"`
}

// UpdateEventRequest representa uma requisição de atualização de evento
type UpdateEventRequest struct {
	Name        string            `json:"name" binding:"required"`
	Location    string            `json:"location" binding:"required"`
	FenceEvent  []LocationRequest `json:"fence_event" binding:"required,min=3"`
	InitialDate string            `json:"initial_date" binding:"required"`
	FinalDate   string            `json:"final_date" binding:"required"`
}

// LocationRequest representa uma coordenada geográfica
type LocationRequest struct {
	Latitude  float64 `json:"latitude" binding:"required,min=-90,max=90"`
	Longitude float64 `json:"longitude" binding:"required,min=-180,max=180"`
}

// EventResponse representa a resposta de um evento
type EventResponse struct {
	ID          string             `json:"id"`
	TenantID    string             `json:"tenant_id"`
	Name        string             `json:"name"`
	Location    string             `json:"location"`
	FenceEvent  []LocationResponse `json:"fence_event"`
	InitialDate string             `json:"initial_date"`
	FinalDate   string             `json:"final_date"`
	Status      string             `json:"status"`
	Active      bool               `json:"active"`
	CreatedAt   string             `json:"created_at"`
	UpdatedAt   string             `json:"updated_at"`
	CreatedBy   *string            `json:"created_by,omitempty"`
	UpdatedBy   *string            `json:"updated_by,omitempty"`
}

// LocationResponse representa uma coordenada geográfica na resposta
type LocationResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// EventListResponse representa a resposta de listagem de eventos
type EventListResponse struct {
	Events     []EventResponse          `json:"events"`
	Pagination httpResponses.Pagination `json:"pagination"`
}

// EventStatsResponse representa estatísticas de um evento
type EventStatsResponse struct {
	EventID        string `json:"event_id"`
	TotalCheckins  int    `json:"total_checkins"`
	TotalCheckouts int    `json:"total_checkouts"`
	ActiveSessions int    `json:"active_sessions"`
	TotalEmployees int    `json:"total_employees"`
}

// Create cria um novo evento
func (h *EventHandler) Create(c *gin.Context) {
	var req CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid create event request", zap.Error(err))
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

	// Converter datas
	initialDate, err := time.Parse("2006-01-02T15:04:05Z", req.InitialDate)
	if err != nil {
		h.logger.Warn("Invalid initial date format", zap.Error(err))
		httpResponses.BadRequest(c, "Invalid initial date format. Use ISO 8601 format", nil)
		return
	}

	finalDate, err := time.Parse("2006-01-02T15:04:05Z", req.FinalDate)
	if err != nil {
		h.logger.Warn("Invalid final date format", zap.Error(err))
		httpResponses.BadRequest(c, "Invalid final date format. Use ISO 8601 format", nil)
		return
	}

	// Converter fence event
	fenceEvent, err := h.convertLocationRequests(req.FenceEvent)
	if err != nil {
		h.logger.Warn("Invalid fence event coordinates", zap.Error(err))
		httpResponses.BadRequest(c, "Invalid fence event coordinates", nil)
		return
	}

	// Criar evento
	evt, err := h.eventService.CreateEvent(c.Request.Context(), tenantID, req.Name, req.Location, fenceEvent, initialDate, finalDate, userID)
	if err != nil {
		h.handleServiceError(c, err, "create event")
		return
	}

	response := h.convertToEventResponse(evt)
	h.logger.Info("Event created successfully", zap.String("event_id", evt.ID.String()))
	httpResponses.Created(c, response, "Event created successfully")
}

// GetByID busca um evento pelo ID
func (h *EventHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := value_objects.ParseUUID(idStr)
	if err != nil {
		h.logger.Warn("Invalid event ID", zap.String("id", idStr))
		httpResponses.BadRequest(c, "Invalid event ID format", nil)
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

	evt, err := h.eventService.GetEventByTenant(c.Request.Context(), id, tenantID)
	if err != nil {
		h.handleServiceError(c, err, "get event")
		return
	}

	response := h.convertToEventResponse(evt)
	httpResponses.Success(c, response, "Event retrieved successfully")
}

// Update atualiza um evento
func (h *EventHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := value_objects.ParseUUID(idStr)
	if err != nil {
		h.logger.Warn("Invalid event ID", zap.String("id", idStr))
		httpResponses.BadRequest(c, "Invalid event ID format", nil)
		return
	}

	var req UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid update event request", zap.Error(err))
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

	// Converter datas
	initialDate, err := time.Parse("2006-01-02T15:04:05Z", req.InitialDate)
	if err != nil {
		h.logger.Warn("Invalid initial date format", zap.Error(err))
		httpResponses.BadRequest(c, "Invalid initial date format. Use ISO 8601 format", nil)
		return
	}

	finalDate, err := time.Parse("2006-01-02T15:04:05Z", req.FinalDate)
	if err != nil {
		h.logger.Warn("Invalid final date format", zap.Error(err))
		httpResponses.BadRequest(c, "Invalid final date format. Use ISO 8601 format", nil)
		return
	}

	// Converter fence event
	fenceEvent, err := h.convertLocationRequests(req.FenceEvent)
	if err != nil {
		h.logger.Warn("Invalid fence event coordinates", zap.Error(err))
		httpResponses.BadRequest(c, "Invalid fence event coordinates", nil)
		return
	}

	evt, err := h.eventService.UpdateEvent(c.Request.Context(), id, req.Name, req.Location, fenceEvent, initialDate, finalDate, userID)
	if err != nil {
		h.handleServiceError(c, err, "update event")
		return
	}

	response := h.convertToEventResponse(evt)
	h.logger.Info("Event updated successfully", zap.String("event_id", evt.ID.String()))
	httpResponses.Success(c, response, "Event updated successfully")
}

// Delete remove um evento (soft delete)
func (h *EventHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := value_objects.ParseUUID(idStr)
	if err != nil {
		h.logger.Warn("Invalid event ID", zap.String("id", idStr))
		httpResponses.BadRequest(c, "Invalid event ID format", nil)
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

	err = h.eventService.DeleteEvent(c.Request.Context(), id, userID)
	if err != nil {
		h.handleServiceError(c, err, "delete event")
		return
	}

	h.logger.Info("Event deleted successfully", zap.String("event_id", id.String()))
	httpResponses.Success(c, nil, "Event deleted successfully")
}

// List lista eventos com paginação e filtros
func (h *EventHandler) List(c *gin.Context) {
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

	events, total, err := h.eventService.ListEvents(c.Request.Context(), filters)
	if err != nil {
		h.handleServiceError(c, err, "list events")
		return
	}

	// Converter para resposta
	eventResponses := make([]EventResponse, len(events))
	for i, evt := range events {
		eventResponses[i] = h.convertToEventResponse(evt)
	}

	response := EventListResponse{
		Events: eventResponses,
		Pagination: httpResponses.Pagination{
			Page:       filters.Page,
			PageSize:   filters.PageSize,
			Total:      total,
			TotalPages: (total + filters.PageSize - 1) / filters.PageSize,
		},
	}

	httpResponses.Success(c, response, "Events retrieved successfully")
}

// GetStats obtém estatísticas de um evento
func (h *EventHandler) GetStats(c *gin.Context) {
	idStr := c.Param("id")
	id, err := value_objects.ParseUUID(idStr)
	if err != nil {
		h.logger.Warn("Invalid event ID", zap.String("id", idStr))
		httpResponses.BadRequest(c, "Invalid event ID format", nil)
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

	// Verificar se o evento existe e pertence ao tenant
	_, err = h.eventService.GetEventByTenant(c.Request.Context(), id, tenantID)
	if err != nil {
		h.handleServiceError(c, err, "get event stats")
		return
	}

	// TODO: Implementar estatísticas reais quando os serviços de check-in estiverem prontos
	stats := EventStatsResponse{
		EventID:        id.String(),
		TotalCheckins:  0,
		TotalCheckouts: 0,
		ActiveSessions: 0,
		TotalEmployees: 0,
	}

	httpResponses.Success(c, stats, "Event statistics retrieved successfully")
}

// buildListFilters constrói os filtros de listagem a partir dos query parameters
func (h *EventHandler) buildListFilters(c *gin.Context) event.ListFilters {
	filters := event.ListFilters{
		Page:     1,
		PageSize: 20,
		OrderBy:  "initial_date",
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

	if location := c.Query("location"); location != "" {
		filters.Location = &location
	}

	if activeStr := c.Query("active"); activeStr != "" {
		if active, err := strconv.ParseBool(activeStr); err == nil {
			filters.Active = &active
		}
	}

	// Filtros de data
	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		if dateFrom, err := time.Parse("2006-01-02", dateFromStr); err == nil {
			filters.DateFrom = &dateFrom
		}
	}

	if dateToStr := c.Query("date_to"); dateToStr != "" {
		if dateTo, err := time.Parse("2006-01-02", dateToStr); err == nil {
			filters.DateTo = &dateTo
		}
	}

	// Filtro de status
	if statusStr := c.Query("status"); statusStr != "" {
		switch statusStr {
		case "ongoing", "upcoming", "finished":
			status := event.EventStatus(statusStr)
			filters.Status = &status
		}
	}

	// Ordenação
	if orderBy := c.Query("order_by"); orderBy != "" {
		validFields := []string{"name", "location", "initial_date", "final_date", "created_at", "updated_at"}
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

// convertLocationRequests converte LocationRequest para value_objects.Location
func (h *EventHandler) convertLocationRequests(locations []LocationRequest) ([]value_objects.Location, error) {
	result := make([]value_objects.Location, len(locations))
	for i, loc := range locations {
		location, err := value_objects.NewLocation(loc.Latitude, loc.Longitude)
		if err != nil {
			return nil, err
		}
		result[i] = location
	}
	return result, nil
}

// convertToEventResponse converte Event para EventResponse
func (h *EventHandler) convertToEventResponse(evt *event.Event) EventResponse {
	response := EventResponse{
		ID:          evt.ID.String(),
		TenantID:    evt.TenantID.String(),
		Name:        evt.Name,
		Location:    evt.Location,
		InitialDate: evt.InitialDate.Format(time.RFC3339),
		FinalDate:   evt.FinalDate.Format(time.RFC3339),
		Status:      h.getEventStatus(evt),
		Active:      evt.Active,
		CreatedAt:   evt.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   evt.UpdatedAt.Format(time.RFC3339),
	}

	// Converter fence event
	for _, location := range evt.FenceEvent {
		response.FenceEvent = append(response.FenceEvent, LocationResponse{
			Latitude:  location.Latitude,
			Longitude: location.Longitude,
		})
	}

	if evt.CreatedBy != nil {
		createdBy := evt.CreatedBy.String()
		response.CreatedBy = &createdBy
	}

	if evt.UpdatedBy != nil {
		updatedBy := evt.UpdatedBy.String()
		response.UpdatedBy = &updatedBy
	}

	return response
}

// getEventStatus determina o status atual do evento
func (h *EventHandler) getEventStatus(evt *event.Event) string {
	now := time.Now()

	if now.Before(evt.InitialDate) {
		return "upcoming"
	}

	if now.After(evt.FinalDate) {
		return "finished"
	}

	return "ongoing"
}

// handleServiceError trata erros do serviço de domínio
func (h *EventHandler) handleServiceError(c *gin.Context, err error, operation string) {
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
