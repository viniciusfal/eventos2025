package checkin

import (
	"context"
	"fmt"
	"time"

	"eventos-backend/internal/domain/shared/constants"
	"eventos-backend/internal/domain/shared/errors"
	"eventos-backend/internal/domain/shared/value_objects"
)

// Service define a interface para serviços de domínio de check-ins
type Service interface {
	// PerformCheckin realiza um check-in com validações completas
	PerformCheckin(ctx context.Context, request CheckinRequest) (*Checkin, *ValidationResult, error)

	// ValidateCheckin valida um check-in existente
	ValidateCheckin(ctx context.Context, checkinID value_objects.UUID, validationResult *ValidationResult, validatedBy value_objects.UUID) error

	// GetCheckin busca um check-in por ID
	GetCheckin(ctx context.Context, id value_objects.UUID) (*Checkin, error)

	// ListCheckins lista check-ins com filtros
	ListCheckins(ctx context.Context, filters ListFilters) ([]*Checkin, int, error)

	// GetEmployeeCheckins busca check-ins de um funcionário
	GetEmployeeCheckins(ctx context.Context, employeeID value_objects.UUID, filters ListFilters) ([]*Checkin, int, error)

	// GetEventCheckins busca check-ins de um evento
	GetEventCheckins(ctx context.Context, eventID value_objects.UUID, filters ListFilters) ([]*Checkin, int, error)

	// CanEmployeeCheckin verifica se funcionário pode fazer check-in no evento
	CanEmployeeCheckin(ctx context.Context, employeeID, eventID value_objects.UUID) (bool, string, error)

	// GetCheckinStats obtém estatísticas de check-ins
	GetCheckinStats(ctx context.Context, tenantID value_objects.UUID) (*CheckinStats, error)

	// AddCheckinNote adiciona observação a um check-in
	AddCheckinNote(ctx context.Context, checkinID value_objects.UUID, note string, updatedBy value_objects.UUID) error

	// GetRecentCheckins busca check-ins recentes
	GetRecentCheckins(ctx context.Context, tenantID value_objects.UUID, limit int) ([]*Checkin, error)

	// ValidateFacialRecognition valida check-in por reconhecimento facial
	ValidateFacialRecognition(ctx context.Context, checkin *Checkin, faceEmbedding []float32) (*ValidationResult, error)

	// ValidateGeolocation valida localização do check-in
	ValidateGeolocation(ctx context.Context, checkin *Checkin, eventLocation value_objects.Location, eventFence []value_objects.Location) (*ValidationResult, error)

	// ValidateEventTiming valida horário do check-in em relação ao evento
	ValidateEventTiming(ctx context.Context, checkin *Checkin, eventStartTime, eventEndTime time.Time) (*ValidationResult, error)
}

// CheckinRequest representa uma requisição de check-in
type CheckinRequest struct {
	TenantID      value_objects.UUID
	EventID       value_objects.UUID
	EmployeeID    value_objects.UUID
	PartnerID     value_objects.UUID
	Method        string
	Location      value_objects.Location
	PhotoURL      string
	Notes         string
	FaceEmbedding []float32 // Para reconhecimento facial
	QRCodeData    string    // Para check-in via QR Code
	CreatedBy     value_objects.UUID
}

// Validate valida a requisição de check-in
func (r *CheckinRequest) Validate() error {
	if r.TenantID.IsZero() {
		return errors.NewValidationError("TenantID", "é obrigatório")
	}

	if r.EventID.IsZero() {
		return errors.NewValidationError("EventID", "é obrigatório")
	}

	if r.EmployeeID.IsZero() {
		return errors.NewValidationError("EmployeeID", "é obrigatório")
	}

	if r.PartnerID.IsZero() {
		return errors.NewValidationError("PartnerID", "é obrigatório")
	}

	if r.CreatedBy.IsZero() {
		return errors.NewValidationError("CreatedBy", "é obrigatório")
	}

	validMethods := map[string]bool{
		constants.CheckMethodFacialRecognition: true,
		constants.CheckMethodQRCode:            true,
		constants.CheckMethodManual:            true,
	}

	if !validMethods[r.Method] {
		return errors.NewValidationError("Method", "método não reconhecido")
	}

	// Validações específicas por método
	if r.Method == constants.CheckMethodFacialRecognition {
		if len(r.FaceEmbedding) == 0 {
			return errors.NewValidationError("FaceEmbedding", "é obrigatório para reconhecimento facial")
		}

		if len(r.FaceEmbedding) != 512 {
			return errors.NewValidationError("FaceEmbedding", "deve ter exatamente 512 dimensões")
		}
	}

	if r.Method == constants.CheckMethodQRCode {
		if r.QRCodeData == "" {
			return errors.NewValidationError("QRCodeData", "é obrigatório para check-in via QR Code")
		}
	}

	return nil
}

// serviceImpl implementa a interface Service
type serviceImpl struct {
	repo      Repository
	statsRepo StatsRepository
}

// NewService cria uma nova instância do serviço
func NewService(repo Repository, statsRepo StatsRepository) Service {
	return &serviceImpl{
		repo:      repo,
		statsRepo: statsRepo,
	}
}

// PerformCheckin realiza um check-in com validações completas
func (s *serviceImpl) PerformCheckin(ctx context.Context, request CheckinRequest) (*Checkin, *ValidationResult, error) {
	// Validar requisição
	if err := request.Validate(); err != nil {
		return nil, nil, err
	}

	// Verificar se funcionário já fez check-in no evento
	exists, err := s.repo.ExistsByEmployeeAndEvent(ctx, request.EmployeeID, request.EventID)
	if err != nil {
		return nil, nil, errors.NewInternalError("Erro ao verificar check-in existente", err)
	}

	if exists {
		return nil, nil, errors.NewAlreadyExistsError("Checkin", "employee_event", fmt.Sprintf("%s-%s", request.EmployeeID.String(), request.EventID.String()))
	}

	// Verificar se funcionário pode fazer check-in
	canCheckin, reason, err := s.CanEmployeeCheckin(ctx, request.EmployeeID, request.EventID)
	if err != nil {
		return nil, nil, err
	}

	if !canCheckin {
		return nil, nil, errors.NewValidationError("Checkin", reason)
	}

	// Criar check-in
	checkin, err := NewCheckin(
		request.TenantID,
		request.EventID,
		request.EmployeeID,
		request.PartnerID,
		request.Method,
		request.Location,
		request.PhotoURL,
		request.Notes,
		request.CreatedBy,
	)
	if err != nil {
		return nil, nil, err
	}

	// Salvar check-in
	if err := s.repo.Create(ctx, checkin); err != nil {
		return nil, nil, errors.NewInternalError("Erro ao criar check-in", err)
	}

	// Iniciar validação assíncrona (por enquanto, validação básica)
	validationResult := s.performBasicValidation(checkin)

	// Atualizar check-in com resultado da validação
	if validationResult.IsValid {
		checkin.MarkAsValid(validationResult.Details, request.CreatedBy)
	} else {
		checkin.MarkAsInvalid(validationResult.Details, request.CreatedBy)
	}

	// Salvar check-in atualizado
	if err := s.repo.Update(ctx, checkin); err != nil {
		return nil, nil, errors.NewInternalError("Erro ao atualizar check-in", err)
	}

	return checkin, validationResult, nil
}

// performBasicValidation realiza validação básica do check-in
func (s *serviceImpl) performBasicValidation(checkin *Checkin) *ValidationResult {
	// Por enquanto, validação simples - todos os check-ins são considerados válidos
	// Em uma implementação completa, aqui seria feita a validação com:
	// - Dados do evento (localização, horário)
	// - Dados do funcionário (foto, embedding facial)
	// - Regras de negócio específicas

	result := NewValidationResult(true, "Check-in realizado com sucesso")
	result.AddDetail("validation_method", "basic")
	result.AddDetail("validation_timestamp", time.Now())

	return result
}

// ValidateCheckin valida um check-in existente
func (s *serviceImpl) ValidateCheckin(ctx context.Context, checkinID value_objects.UUID, validationResult *ValidationResult, validatedBy value_objects.UUID) error {
	checkin, err := s.repo.GetByID(ctx, checkinID)
	if err != nil {
		return errors.NewNotFoundError("Checkin não encontrado", err)
	}

	if validationResult.IsValid {
		checkin.MarkAsValid(validationResult.Details, validatedBy)
	} else {
		checkin.MarkAsInvalid(validationResult.Details, validatedBy)
	}

	if err := s.repo.Update(ctx, checkin); err != nil {
		return errors.NewInternalError("Erro ao atualizar check-in", err)
	}

	return nil
}

// GetCheckin busca um check-in por ID
func (s *serviceImpl) GetCheckin(ctx context.Context, id value_objects.UUID) (*Checkin, error) {
	checkin, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("Checkin não encontrado", err)
	}

	return checkin, nil
}

// ListCheckins lista check-ins com filtros
func (s *serviceImpl) ListCheckins(ctx context.Context, filters ListFilters) ([]*Checkin, int, error) {
	if err := filters.Validate(); err != nil {
		return nil, 0, err
	}

	checkins, total, err := s.repo.List(ctx, filters)
	if err != nil {
		return nil, 0, errors.NewInternalError("Erro ao listar check-ins", err)
	}

	return checkins, total, nil
}

// GetEmployeeCheckins busca check-ins de um funcionário
func (s *serviceImpl) GetEmployeeCheckins(ctx context.Context, employeeID value_objects.UUID, filters ListFilters) ([]*Checkin, int, error) {
	if err := filters.Validate(); err != nil {
		return nil, 0, err
	}

	checkins, total, err := s.repo.GetByEmployee(ctx, employeeID, filters)
	if err != nil {
		return nil, 0, errors.NewInternalError("Erro ao buscar check-ins do funcionário", err)
	}

	return checkins, total, nil
}

// GetEventCheckins busca check-ins de um evento
func (s *serviceImpl) GetEventCheckins(ctx context.Context, eventID value_objects.UUID, filters ListFilters) ([]*Checkin, int, error) {
	if err := filters.Validate(); err != nil {
		return nil, 0, err
	}

	checkins, total, err := s.repo.GetByEvent(ctx, eventID, filters)
	if err != nil {
		return nil, 0, errors.NewInternalError("Erro ao buscar check-ins do evento", err)
	}

	return checkins, total, nil
}

// CanEmployeeCheckin verifica se funcionário pode fazer check-in no evento
func (s *serviceImpl) CanEmployeeCheckin(ctx context.Context, employeeID, eventID value_objects.UUID) (bool, string, error) {
	// TODO: Implementar validações completas:
	// 1. Verificar se funcionário está ativo
	// 2. Verificar se evento está ativo e em andamento
	// 3. Verificar se funcionário está associado ao parceiro do evento
	// 4. Verificar regras de negócio específicas

	// Por enquanto, sempre permite check-in
	return true, "", nil
}

// GetCheckinStats obtém estatísticas de check-ins
func (s *serviceImpl) GetCheckinStats(ctx context.Context, tenantID value_objects.UUID) (*CheckinStats, error) {
	stats, err := s.statsRepo.GetTenantStats(ctx, tenantID)
	if err != nil {
		return nil, errors.NewInternalError("Erro ao obter estatísticas", err)
	}

	return stats, nil
}

// AddCheckinNote adiciona observação a um check-in
func (s *serviceImpl) AddCheckinNote(ctx context.Context, checkinID value_objects.UUID, note string, updatedBy value_objects.UUID) error {
	checkin, err := s.repo.GetByID(ctx, checkinID)
	if err != nil {
		return errors.NewNotFoundError("Checkin não encontrado", err)
	}

	if err := checkin.AddNote(note, updatedBy); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, checkin); err != nil {
		return errors.NewInternalError("Erro ao atualizar check-in", err)
	}

	return nil
}

// GetRecentCheckins busca check-ins recentes
func (s *serviceImpl) GetRecentCheckins(ctx context.Context, tenantID value_objects.UUID, limit int) ([]*Checkin, error) {
	if limit <= 0 || limit > 100 {
		limit = 20 // Padrão
	}

	checkins, err := s.repo.GetRecentCheckins(ctx, tenantID, limit)
	if err != nil {
		return nil, errors.NewInternalError("Erro ao buscar check-ins recentes", err)
	}

	return checkins, nil
}

// ValidateFacialRecognition valida check-in por reconhecimento facial
func (s *serviceImpl) ValidateFacialRecognition(ctx context.Context, checkin *Checkin, faceEmbedding []float32) (*ValidationResult, error) {
	// TODO: Implementar validação de reconhecimento facial
	// 1. Buscar embedding facial do funcionário
	// 2. Calcular similaridade coseno
	// 3. Verificar threshold de confiança
	// 4. Retornar resultado da validação

	result := NewValidationResult(true, "Reconhecimento facial validado")
	result.SetFacialSimilarity(0.95) // Simulado
	result.AddDetail("confidence_level", "high")

	return result, nil
}

// ValidateGeolocation valida localização do check-in
func (s *serviceImpl) ValidateGeolocation(ctx context.Context, checkin *Checkin, eventLocation value_objects.Location, eventFence []value_objects.Location) (*ValidationResult, error) {
	// TODO: Implementar validação geográfica
	// 1. Calcular distância do check-in ao evento
	// 2. Verificar se está dentro da cerca geográfica
	// 3. Aplicar tolerâncias configuráveis
	// 4. Retornar resultado da validação

	distance := checkin.Location.DistanceTo(eventLocation)
	withinBounds := distance <= 100 // 100 metros de tolerância

	result := NewValidationResult(withinBounds, "Localização validada")
	result.SetDistance(distance)
	result.SetWithinBounds(withinBounds)

	return result, nil
}

// ValidateEventTiming valida horário do check-in em relação ao evento
func (s *serviceImpl) ValidateEventTiming(ctx context.Context, checkin *Checkin, eventStartTime, eventEndTime time.Time) (*ValidationResult, error) {
	now := checkin.CheckinTime
	isWithinEventTime := now.After(eventStartTime) && now.Before(eventEndTime)

	var reason string
	if now.Before(eventStartTime) {
		reason = "Check-in realizado antes do início do evento"
	} else if now.After(eventEndTime) {
		reason = "Check-in realizado após o término do evento"
	} else {
		reason = "Check-in realizado no horário correto"
	}

	result := NewValidationResult(isWithinEventTime, reason)
	result.AddDetail("event_start", eventStartTime)
	result.AddDetail("event_end", eventEndTime)
	result.AddDetail("checkin_time", now)

	return result, nil
}
