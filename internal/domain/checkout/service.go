package checkout

import (
	"context"
	"time"

	"eventos-backend/internal/domain/shared/constants"
	"eventos-backend/internal/domain/shared/errors"
	"eventos-backend/internal/domain/shared/value_objects"
)

// Service define a interface para serviços de domínio de check-outs
type Service interface {
	// PerformCheckout realiza um check-out com validações completas
	PerformCheckout(ctx context.Context, request CheckoutRequest) (*Checkout, *ValidationResult, error)

	// ValidateCheckout valida um check-out existente
	ValidateCheckout(ctx context.Context, checkoutID value_objects.UUID, validationResult *ValidationResult, validatedBy value_objects.UUID) error

	// GetCheckout busca um check-out por ID
	GetCheckout(ctx context.Context, id value_objects.UUID) (*Checkout, error)

	// ListCheckouts lista check-outs com filtros
	ListCheckouts(ctx context.Context, filters ListFilters) ([]*Checkout, int, error)

	// GetEmployeeCheckouts busca check-outs de um funcionário
	GetEmployeeCheckouts(ctx context.Context, employeeID value_objects.UUID, filters ListFilters) ([]*Checkout, int, error)

	// GetEventCheckouts busca check-outs de um evento
	GetEventCheckouts(ctx context.Context, eventID value_objects.UUID, filters ListFilters) ([]*Checkout, int, error)

	// CanEmployeeCheckout verifica se funcionário pode fazer check-out
	CanEmployeeCheckout(ctx context.Context, employeeID, eventID value_objects.UUID) (bool, string, error)

	// GetCheckoutStats obtém estatísticas de check-outs
	GetCheckoutStats(ctx context.Context, tenantID value_objects.UUID) (*CheckoutStats, error)

	// GetWorkStats obtém estatísticas de trabalho
	GetWorkStats(ctx context.Context, tenantID value_objects.UUID) (*WorkStats, error)

	// AddCheckoutNote adiciona observação a um check-out
	AddCheckoutNote(ctx context.Context, checkoutID value_objects.UUID, note string, updatedBy value_objects.UUID) error

	// GetRecentCheckouts busca check-outs recentes
	GetRecentCheckouts(ctx context.Context, tenantID value_objects.UUID, limit int) ([]*Checkout, error)

	// GetWorkSessions busca sessões de trabalho completas
	GetWorkSessions(ctx context.Context, tenantID value_objects.UUID, filters WorkSessionFilters) ([]*WorkSession, int, error)

	// GetEmployeeWorkSessions busca sessões de trabalho de um funcionário
	GetEmployeeWorkSessions(ctx context.Context, employeeID value_objects.UUID, filters WorkSessionFilters) ([]*WorkSession, int, error)

	// ValidateFacialRecognition valida check-out por reconhecimento facial
	ValidateFacialRecognition(ctx context.Context, checkout *Checkout, faceEmbedding []float32) (*ValidationResult, error)

	// ValidateGeolocation valida localização do check-out
	ValidateGeolocation(ctx context.Context, checkout *Checkout, eventLocation value_objects.Location, eventFence []value_objects.Location) (*ValidationResult, error)

	// ValidateWorkDuration valida duração do trabalho
	ValidateWorkDuration(ctx context.Context, checkout *Checkout, checkinTime time.Time) (*ValidationResult, error)
}

// CheckoutRequest representa uma requisição de check-out
type CheckoutRequest struct {
	TenantID      value_objects.UUID
	EventID       value_objects.UUID
	EmployeeID    value_objects.UUID
	PartnerID     value_objects.UUID
	CheckinID     value_objects.UUID
	Method        string
	Location      value_objects.Location
	PhotoURL      string
	Notes         string
	FaceEmbedding []float32 // Para reconhecimento facial
	QRCodeData    string    // Para check-out via QR Code
	CreatedBy     value_objects.UUID
}

// Validate valida a requisição de check-out
func (r *CheckoutRequest) Validate() error {
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

	if r.CheckinID.IsZero() {
		return errors.NewValidationError("CheckinID", "é obrigatório")
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
			return errors.NewValidationError("QRCodeData", "é obrigatório para check-out via QR Code")
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

// PerformCheckout realiza um check-out com validações completas
func (s *serviceImpl) PerformCheckout(ctx context.Context, request CheckoutRequest) (*Checkout, *ValidationResult, error) {
	// Validar requisição
	if err := request.Validate(); err != nil {
		return nil, nil, err
	}

	// Verificar se já existe check-out para este check-in
	exists, err := s.repo.ExistsByCheckin(ctx, request.CheckinID)
	if err != nil {
		return nil, nil, errors.NewInternalError("Erro ao verificar check-out existente", err)
	}

	if exists {
		return nil, nil, errors.NewAlreadyExistsError("Checkout", "checkin_id", request.CheckinID.String())
	}

	// Verificar se funcionário pode fazer check-out
	canCheckout, reason, err := s.CanEmployeeCheckout(ctx, request.EmployeeID, request.EventID)
	if err != nil {
		return nil, nil, err
	}

	if !canCheckout {
		return nil, nil, errors.NewValidationError("Checkout", reason)
	}

	// Criar check-out
	checkout, err := NewCheckout(
		request.TenantID,
		request.EventID,
		request.EmployeeID,
		request.PartnerID,
		request.CheckinID,
		request.Method,
		request.Location,
		request.PhotoURL,
		request.Notes,
		request.CreatedBy,
	)
	if err != nil {
		return nil, nil, err
	}

	// Salvar check-out
	if err := s.repo.Create(ctx, checkout); err != nil {
		return nil, nil, errors.NewInternalError("Erro ao criar check-out", err)
	}

	// Realizar validação (por enquanto, validação básica)
	validationResult := s.performBasicValidation(checkout)

	// Atualizar check-out com resultado da validação
	if validationResult.IsValid {
		checkout.MarkAsValid(validationResult.Details, request.CreatedBy)
	} else {
		checkout.MarkAsInvalid(validationResult.Details, request.CreatedBy)
	}

	// Salvar check-out atualizado
	if err := s.repo.Update(ctx, checkout); err != nil {
		return nil, nil, errors.NewInternalError("Erro ao atualizar check-out", err)
	}

	return checkout, validationResult, nil
}

// performBasicValidation realiza validação básica do check-out
func (s *serviceImpl) performBasicValidation(checkout *Checkout) *ValidationResult {
	// Por enquanto, validação simples - todos os check-outs são considerados válidos
	// Em uma implementação completa, aqui seria feita a validação com:
	// - Dados do check-in correspondente
	// - Duração do trabalho
	// - Localização e regras de negócio

	result := NewValidationResult(true, "Check-out realizado com sucesso")
	result.AddDetail("validation_method", "basic")
	result.AddDetail("validation_timestamp", time.Now())

	return result
}

// ValidateCheckout valida um check-out existente
func (s *serviceImpl) ValidateCheckout(ctx context.Context, checkoutID value_objects.UUID, validationResult *ValidationResult, validatedBy value_objects.UUID) error {
	checkout, err := s.repo.GetByID(ctx, checkoutID)
	if err != nil {
		return errors.NewNotFoundError("Checkout não encontrado", err)
	}

	if validationResult.IsValid {
		checkout.MarkAsValid(validationResult.Details, validatedBy)
	} else {
		checkout.MarkAsInvalid(validationResult.Details, validatedBy)
	}

	if err := s.repo.Update(ctx, checkout); err != nil {
		return errors.NewInternalError("Erro ao atualizar check-out", err)
	}

	return nil
}

// GetCheckout busca um check-out por ID
func (s *serviceImpl) GetCheckout(ctx context.Context, id value_objects.UUID) (*Checkout, error) {
	checkout, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("Checkout não encontrado", err)
	}

	return checkout, nil
}

// ListCheckouts lista check-outs com filtros
func (s *serviceImpl) ListCheckouts(ctx context.Context, filters ListFilters) ([]*Checkout, int, error) {
	if err := filters.Validate(); err != nil {
		return nil, 0, err
	}

	checkouts, total, err := s.repo.List(ctx, filters)
	if err != nil {
		return nil, 0, errors.NewInternalError("Erro ao listar check-outs", err)
	}

	return checkouts, total, nil
}

// GetEmployeeCheckouts busca check-outs de um funcionário
func (s *serviceImpl) GetEmployeeCheckouts(ctx context.Context, employeeID value_objects.UUID, filters ListFilters) ([]*Checkout, int, error) {
	if err := filters.Validate(); err != nil {
		return nil, 0, err
	}

	checkouts, total, err := s.repo.GetByEmployee(ctx, employeeID, filters)
	if err != nil {
		return nil, 0, errors.NewInternalError("Erro ao buscar check-outs do funcionário", err)
	}

	return checkouts, total, nil
}

// GetEventCheckouts busca check-outs de um evento
func (s *serviceImpl) GetEventCheckouts(ctx context.Context, eventID value_objects.UUID, filters ListFilters) ([]*Checkout, int, error) {
	if err := filters.Validate(); err != nil {
		return nil, 0, err
	}

	checkouts, total, err := s.repo.GetByEvent(ctx, eventID, filters)
	if err != nil {
		return nil, 0, errors.NewInternalError("Erro ao buscar check-outs do evento", err)
	}

	return checkouts, total, nil
}

// CanEmployeeCheckout verifica se funcionário pode fazer check-out
func (s *serviceImpl) CanEmployeeCheckout(ctx context.Context, employeeID, eventID value_objects.UUID) (bool, string, error) {
	// TODO: Implementar validações completas:
	// 1. Verificar se funcionário tem check-in ativo no evento
	// 2. Verificar se não há check-out já realizado
	// 3. Verificar regras de negócio específicas

	// Por enquanto, sempre permite check-out
	return true, "", nil
}

// GetCheckoutStats obtém estatísticas de check-outs
func (s *serviceImpl) GetCheckoutStats(ctx context.Context, tenantID value_objects.UUID) (*CheckoutStats, error) {
	stats, err := s.statsRepo.GetTenantStats(ctx, tenantID)
	if err != nil {
		return nil, errors.NewInternalError("Erro ao obter estatísticas de check-out", err)
	}

	return stats, nil
}

// GetWorkStats obtém estatísticas de trabalho
func (s *serviceImpl) GetWorkStats(ctx context.Context, tenantID value_objects.UUID) (*WorkStats, error) {
	stats, err := s.statsRepo.GetWorkStats(ctx, tenantID)
	if err != nil {
		return nil, errors.NewInternalError("Erro ao obter estatísticas de trabalho", err)
	}

	return stats, nil
}

// AddCheckoutNote adiciona observação a um check-out
func (s *serviceImpl) AddCheckoutNote(ctx context.Context, checkoutID value_objects.UUID, note string, updatedBy value_objects.UUID) error {
	checkout, err := s.repo.GetByID(ctx, checkoutID)
	if err != nil {
		return errors.NewNotFoundError("Checkout não encontrado", err)
	}

	if err := checkout.AddNote(note, updatedBy); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, checkout); err != nil {
		return errors.NewInternalError("Erro ao atualizar check-out", err)
	}

	return nil
}

// GetRecentCheckouts busca check-outs recentes
func (s *serviceImpl) GetRecentCheckouts(ctx context.Context, tenantID value_objects.UUID, limit int) ([]*Checkout, error) {
	if limit <= 0 || limit > 100 {
		limit = 20 // Padrão
	}

	checkouts, err := s.repo.GetRecentCheckouts(ctx, tenantID, limit)
	if err != nil {
		return nil, errors.NewInternalError("Erro ao buscar check-outs recentes", err)
	}

	return checkouts, nil
}

// GetWorkSessions busca sessões de trabalho completas
func (s *serviceImpl) GetWorkSessions(ctx context.Context, tenantID value_objects.UUID, filters WorkSessionFilters) ([]*WorkSession, int, error) {
	if err := filters.Validate(); err != nil {
		return nil, 0, err
	}

	sessions, total, err := s.repo.GetWorkSessions(ctx, tenantID, filters)
	if err != nil {
		return nil, 0, errors.NewInternalError("Erro ao buscar sessões de trabalho", err)
	}

	return sessions, total, nil
}

// GetEmployeeWorkSessions busca sessões de trabalho de um funcionário
func (s *serviceImpl) GetEmployeeWorkSessions(ctx context.Context, employeeID value_objects.UUID, filters WorkSessionFilters) ([]*WorkSession, int, error) {
	if err := filters.Validate(); err != nil {
		return nil, 0, err
	}

	sessions, total, err := s.repo.GetEmployeeWorkSessions(ctx, employeeID, filters)
	if err != nil {
		return nil, 0, errors.NewInternalError("Erro ao buscar sessões de trabalho do funcionário", err)
	}

	return sessions, total, nil
}

// ValidateFacialRecognition valida check-out por reconhecimento facial
func (s *serviceImpl) ValidateFacialRecognition(ctx context.Context, checkout *Checkout, faceEmbedding []float32) (*ValidationResult, error) {
	// TODO: Implementar validação de reconhecimento facial
	// 1. Buscar embedding facial do funcionário
	// 2. Calcular similaridade coseno
	// 3. Verificar threshold de confiança
	// 4. Comparar com check-in (mesma pessoa?)

	result := NewValidationResult(true, "Reconhecimento facial validado")
	result.SetFacialSimilarity(0.93) // Simulado
	result.AddDetail("confidence_level", "high")

	return result, nil
}

// ValidateGeolocation valida localização do check-out
func (s *serviceImpl) ValidateGeolocation(ctx context.Context, checkout *Checkout, eventLocation value_objects.Location, eventFence []value_objects.Location) (*ValidationResult, error) {
	// TODO: Implementar validação geográfica
	// 1. Calcular distância do check-out ao evento
	// 2. Verificar se está dentro da cerca geográfica
	// 3. Aplicar tolerâncias configuráveis

	distance := checkout.Location.DistanceTo(eventLocation)
	withinBounds := distance <= 100 // 100 metros de tolerância

	result := NewValidationResult(withinBounds, "Localização validada")
	result.SetDistance(distance)
	result.SetWithinBounds(withinBounds)

	return result, nil
}

// ValidateWorkDuration valida duração do trabalho
func (s *serviceImpl) ValidateWorkDuration(ctx context.Context, checkout *Checkout, checkinTime time.Time) (*ValidationResult, error) {
	// Calcular duração do trabalho
	checkout.CalculateWorkDuration(checkinTime)

	// Validar duração
	isValid := true
	reason := "Duração de trabalho válida"

	if checkout.WorkDuration < 0 {
		isValid = false
		reason = "Check-out anterior ao check-in"
	} else if checkout.IsShortWork() {
		reason = "Trabalho muito curto (menos de 1 hora)"
		// Pode ser válido, mas com aviso
	} else if checkout.IsLongWork() {
		reason = "Trabalho muito longo (mais de 12 horas)"
		// Pode ser válido, mas com aviso
	}

	result := NewValidationResult(isValid, reason)
	result.SetWorkDuration(checkout.WorkDuration)
	result.AddDetail("checkin_time", checkinTime)
	result.AddDetail("checkout_time", checkout.CheckoutTime)
	result.AddDetail("is_short_work", checkout.IsShortWork())
	result.AddDetail("is_long_work", checkout.IsLongWork())

	return result, nil
}
