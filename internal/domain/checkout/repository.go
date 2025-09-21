package checkout

import (
	"context"
	"strings"
	"time"

	"eventos-backend/internal/domain/shared/errors"
	"eventos-backend/internal/domain/shared/value_objects"
)

// Repository define a interface para operações de persistência de check-outs
type Repository interface {
	// Create cria um novo check-out
	Create(ctx context.Context, checkout *Checkout) error

	// GetByID busca um check-out por ID
	GetByID(ctx context.Context, id value_objects.UUID) (*Checkout, error)

	// Update atualiza um check-out existente
	Update(ctx context.Context, checkout *Checkout) error

	// Delete remove um check-out (soft delete)
	Delete(ctx context.Context, id value_objects.UUID, deletedBy value_objects.UUID) error

	// List lista check-outs com filtros e paginação
	List(ctx context.Context, filters ListFilters) ([]*Checkout, int, error)

	// ListByTenant lista check-outs de um tenant específico
	ListByTenant(ctx context.Context, tenantID value_objects.UUID, filters ListFilters) ([]*Checkout, int, error)

	// GetByEmployee busca check-outs de um funcionário
	GetByEmployee(ctx context.Context, employeeID value_objects.UUID, filters ListFilters) ([]*Checkout, int, error)

	// GetByEvent busca check-outs de um evento
	GetByEvent(ctx context.Context, eventID value_objects.UUID, filters ListFilters) ([]*Checkout, int, error)

	// GetByPartner busca check-outs de um parceiro
	GetByPartner(ctx context.Context, partnerID value_objects.UUID, filters ListFilters) ([]*Checkout, int, error)

	// GetByCheckin busca check-out por check-in
	GetByCheckin(ctx context.Context, checkinID value_objects.UUID) (*Checkout, error)

	// ExistsByCheckin verifica se já existe check-out para o check-in
	ExistsByCheckin(ctx context.Context, checkinID value_objects.UUID) (bool, error)

	// GetByEmployeeAndEvent busca check-out específico de funcionário em evento
	GetByEmployeeAndEvent(ctx context.Context, employeeID, eventID value_objects.UUID) (*Checkout, error)

	// GetByDateRange busca check-outs em um período
	GetByDateRange(ctx context.Context, tenantID value_objects.UUID, startDate, endDate time.Time, filters ListFilters) ([]*Checkout, int, error)

	// GetByMethod busca check-outs por método
	GetByMethod(ctx context.Context, tenantID value_objects.UUID, method string, filters ListFilters) ([]*Checkout, int, error)

	// GetValidCheckouts busca apenas check-outs válidos
	GetValidCheckouts(ctx context.Context, tenantID value_objects.UUID, filters ListFilters) ([]*Checkout, int, error)

	// GetInvalidCheckouts busca apenas check-outs inválidos
	GetInvalidCheckouts(ctx context.Context, tenantID value_objects.UUID, filters ListFilters) ([]*Checkout, int, error)

	// CountByTenant conta check-outs de um tenant
	CountByTenant(ctx context.Context, tenantID value_objects.UUID) (int, error)

	// CountByEvent conta check-outs de um evento
	CountByEvent(ctx context.Context, eventID value_objects.UUID) (int, error)

	// CountByEmployee conta check-outs de um funcionário
	CountByEmployee(ctx context.Context, employeeID value_objects.UUID) (int, error)

	// GetRecentCheckouts busca check-outs recentes (últimas 24h)
	GetRecentCheckouts(ctx context.Context, tenantID value_objects.UUID, limit int) ([]*Checkout, error)

	// GetCheckoutsByLocation busca check-outs próximos a uma localização
	GetCheckoutsByLocation(ctx context.Context, tenantID value_objects.UUID, location value_objects.Location, radiusKm float64, filters ListFilters) ([]*Checkout, int, error)

	// GetWorkSessions busca sessões de trabalho completas (check-in + check-out)
	GetWorkSessions(ctx context.Context, tenantID value_objects.UUID, filters WorkSessionFilters) ([]*WorkSession, int, error)

	// GetEmployeeWorkSessions busca sessões de trabalho de um funcionário
	GetEmployeeWorkSessions(ctx context.Context, employeeID value_objects.UUID, filters WorkSessionFilters) ([]*WorkSession, int, error)

	// GetEventWorkSessions busca sessões de trabalho de um evento
	GetEventWorkSessions(ctx context.Context, eventID value_objects.UUID, filters WorkSessionFilters) ([]*WorkSession, int, error)
}

// ListFilters define os filtros para listagem de check-outs
type ListFilters struct {
	// Filtros básicos
	TenantID   *value_objects.UUID
	EventID    *value_objects.UUID
	EmployeeID *value_objects.UUID
	PartnerID  *value_objects.UUID
	CheckinID  *value_objects.UUID

	// Filtros específicos
	Method   *string
	IsValid  *bool
	HasPhoto *bool

	// Filtros temporais
	StartDate *time.Time
	EndDate   *time.Time

	// Filtros de duração
	MinDurationHours *float64
	MaxDurationHours *float64

	// Filtros geográficos
	Location *value_objects.Location
	RadiusKm *float64

	// Busca textual
	Search *string // Busca em notes

	// Paginação
	Page     int
	PageSize int

	// Ordenação
	OrderBy   string // checkout_time, work_duration, created_at, updated_at, method
	OrderDesc bool
}

// Validate valida os filtros de listagem
func (f *ListFilters) Validate() error {
	// Validar paginação
	if f.Page < 1 {
		f.Page = 1
	}

	if f.PageSize < 1 {
		f.PageSize = 20 // Padrão
	}

	if f.PageSize > 100 {
		f.PageSize = 100 // Máximo
	}

	// Validar ordenação
	validOrderFields := map[string]bool{
		"checkout_time": true,
		"work_duration": true,
		"created_at":    true,
		"updated_at":    true,
		"method":        true,
	}

	if f.OrderBy != "" && !validOrderFields[f.OrderBy] {
		f.OrderBy = "checkout_time" // Padrão: ordenar por horário do check-out
	}

	if f.OrderBy == "" {
		f.OrderBy = "checkout_time"
	}

	// Validar método
	if f.Method != nil && *f.Method != "" {
		*f.Method = strings.ToLower(strings.TrimSpace(*f.Method))
		validMethods := map[string]bool{
			"facial_recognition": true,
			"qr_code":            true,
			"manual":             true,
		}

		if !validMethods[*f.Method] {
			return errors.NewValidationError("Method", "método não reconhecido")
		}
	}

	// Validar datas
	if f.StartDate != nil && f.EndDate != nil {
		if f.StartDate.After(*f.EndDate) {
			return errors.NewValidationError("StartDate", "deve ser anterior à data final")
		}
	}

	// Validar duração
	if f.MinDurationHours != nil && *f.MinDurationHours < 0 {
		return errors.NewValidationError("MinDurationHours", "deve ser maior ou igual a zero")
	}

	if f.MaxDurationHours != nil && *f.MaxDurationHours < 0 {
		return errors.NewValidationError("MaxDurationHours", "deve ser maior ou igual a zero")
	}

	if f.MinDurationHours != nil && f.MaxDurationHours != nil {
		if *f.MinDurationHours > *f.MaxDurationHours {
			return errors.NewValidationError("MinDurationHours", "deve ser menor ou igual à duração máxima")
		}
	}

	// Validar raio geográfico
	if f.RadiusKm != nil {
		if *f.RadiusKm < 0 {
			return errors.NewValidationError("RadiusKm", "deve ser maior ou igual a zero")
		}

		if *f.RadiusKm > 100 {
			return errors.NewValidationError("RadiusKm", "deve ser menor ou igual a 100km")
		}
	}

	// Se tem raio, deve ter localização
	if f.RadiusKm != nil && f.Location == nil {
		return errors.NewValidationError("Location", "é obrigatória quando RadiusKm é especificado")
	}

	return nil
}

// WorkSessionFilters define filtros para sessões de trabalho
type WorkSessionFilters struct {
	// Filtros básicos
	TenantID   *value_objects.UUID
	EventID    *value_objects.UUID
	EmployeeID *value_objects.UUID
	PartnerID  *value_objects.UUID

	// Filtros temporais
	StartDate *time.Time
	EndDate   *time.Time

	// Filtros de duração
	MinDurationHours *float64
	MaxDurationHours *float64

	// Filtros de validade
	IsValid    *bool
	IsComplete *bool

	// Paginação
	Page     int
	PageSize int

	// Ordenação
	OrderBy   string // duration, checkin_time, checkout_time
	OrderDesc bool
}

// Validate valida os filtros de sessões de trabalho
func (f *WorkSessionFilters) Validate() error {
	// Validar paginação
	if f.Page < 1 {
		f.Page = 1
	}

	if f.PageSize < 1 {
		f.PageSize = 20 // Padrão
	}

	if f.PageSize > 100 {
		f.PageSize = 100 // Máximo
	}

	// Validar ordenação
	validOrderFields := map[string]bool{
		"duration":      true,
		"checkin_time":  true,
		"checkout_time": true,
	}

	if f.OrderBy != "" && !validOrderFields[f.OrderBy] {
		f.OrderBy = "duration" // Padrão: ordenar por duração
	}

	if f.OrderBy == "" {
		f.OrderBy = "duration"
	}

	// Validar datas
	if f.StartDate != nil && f.EndDate != nil {
		if f.StartDate.After(*f.EndDate) {
			return errors.NewValidationError("StartDate", "deve ser anterior à data final")
		}
	}

	// Validar duração
	if f.MinDurationHours != nil && *f.MinDurationHours < 0 {
		return errors.NewValidationError("MinDurationHours", "deve ser maior ou igual a zero")
	}

	if f.MaxDurationHours != nil && *f.MaxDurationHours < 0 {
		return errors.NewValidationError("MaxDurationHours", "deve ser maior ou igual a zero")
	}

	if f.MinDurationHours != nil && f.MaxDurationHours != nil {
		if *f.MinDurationHours > *f.MaxDurationHours {
			return errors.NewValidationError("MinDurationHours", "deve ser menor ou igual à duração máxima")
		}
	}

	return nil
}

// GetOffset calcula o offset para paginação
func (f *ListFilters) GetOffset() int {
	return (f.Page - 1) * f.PageSize
}

// GetLimit retorna o limite para paginação
func (f *ListFilters) GetLimit() int {
	return f.PageSize
}

// GetOffset calcula o offset para paginação de sessões
func (f *WorkSessionFilters) GetOffset() int {
	return (f.Page - 1) * f.PageSize
}

// GetLimit retorna o limite para paginação de sessões
func (f *WorkSessionFilters) GetLimit() int {
	return f.PageSize
}

// CheckoutStats representa estatísticas de check-outs
type CheckoutStats struct {
	TotalCheckouts     int
	ValidCheckouts     int
	InvalidCheckouts   int
	PendingCheckouts   int
	FacialCheckouts    int
	QRCodeCheckouts    int
	ManualCheckouts    int
	CheckoutsToday     int
	CheckoutsThisWeek  int
	CheckoutsThisMonth int
	AverageWorkHours   float64
	TotalWorkHours     float64
	ShortSessions      int // < 1 hora
	LongSessions       int // > 12 horas
	LastCheckoutTime   *time.Time
}

// WorkStats representa estatísticas de trabalho
type WorkStats struct {
	TotalSessions      int
	CompleteSessions   int
	IncompleteSessions int
	ValidSessions      int
	InvalidSessions    int
	TotalWorkHours     float64
	AverageWorkHours   float64
	MinWorkHours       float64
	MaxWorkHours       float64
	ShortSessions      int
	NormalSessions     int
	LongSessions       int
}

// StatsRepository define operações para estatísticas de check-outs
type StatsRepository interface {
	// GetTenantStats obtém estatísticas de um tenant
	GetTenantStats(ctx context.Context, tenantID value_objects.UUID) (*CheckoutStats, error)

	// GetEventStats obtém estatísticas de um evento
	GetEventStats(ctx context.Context, eventID value_objects.UUID) (*CheckoutStats, error)

	// GetEmployeeStats obtém estatísticas de um funcionário
	GetEmployeeStats(ctx context.Context, employeeID value_objects.UUID) (*CheckoutStats, error)

	// GetPartnerStats obtém estatísticas de um parceiro
	GetPartnerStats(ctx context.Context, partnerID value_objects.UUID) (*CheckoutStats, error)

	// GetDailyStats obtém estatísticas diárias
	GetDailyStats(ctx context.Context, tenantID value_objects.UUID, date time.Time) (*CheckoutStats, error)

	// GetWeeklyStats obtém estatísticas semanais
	GetWeeklyStats(ctx context.Context, tenantID value_objects.UUID, startDate time.Time) (*CheckoutStats, error)

	// GetMonthlyStats obtém estatísticas mensais
	GetMonthlyStats(ctx context.Context, tenantID value_objects.UUID, year int, month time.Month) (*CheckoutStats, error)

	// GetWorkStats obtém estatísticas de trabalho
	GetWorkStats(ctx context.Context, tenantID value_objects.UUID) (*WorkStats, error)

	// GetEmployeeWorkStats obtém estatísticas de trabalho de um funcionário
	GetEmployeeWorkStats(ctx context.Context, employeeID value_objects.UUID) (*WorkStats, error)

	// GetEventWorkStats obtém estatísticas de trabalho de um evento
	GetEventWorkStats(ctx context.Context, eventID value_objects.UUID) (*WorkStats, error)
}
