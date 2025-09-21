package checkin

import (
	"context"
	"strings"
	"time"

	"eventos-backend/internal/domain/shared/errors"
	"eventos-backend/internal/domain/shared/value_objects"
)

// Repository define a interface para operações de persistência de check-ins
type Repository interface {
	// Create cria um novo check-in
	Create(ctx context.Context, checkin *Checkin) error

	// GetByID busca um check-in por ID
	GetByID(ctx context.Context, id value_objects.UUID) (*Checkin, error)

	// Update atualiza um check-in existente
	Update(ctx context.Context, checkin *Checkin) error

	// Delete remove um check-in (soft delete)
	Delete(ctx context.Context, id value_objects.UUID, deletedBy value_objects.UUID) error

	// List lista check-ins com filtros e paginação
	List(ctx context.Context, filters ListFilters) ([]*Checkin, int, error)

	// ListByTenant lista check-ins de um tenant específico
	ListByTenant(ctx context.Context, tenantID value_objects.UUID, filters ListFilters) ([]*Checkin, int, error)

	// GetByEmployee busca check-ins de um funcionário
	GetByEmployee(ctx context.Context, employeeID value_objects.UUID, filters ListFilters) ([]*Checkin, int, error)

	// GetByEvent busca check-ins de um evento
	GetByEvent(ctx context.Context, eventID value_objects.UUID, filters ListFilters) ([]*Checkin, int, error)

	// GetByPartner busca check-ins de um parceiro
	GetByPartner(ctx context.Context, partnerID value_objects.UUID, filters ListFilters) ([]*Checkin, int, error)

	// GetByEmployeeAndEvent busca check-in específico de funcionário em evento
	GetByEmployeeAndEvent(ctx context.Context, employeeID, eventID value_objects.UUID) (*Checkin, error)

	// ExistsByEmployeeAndEvent verifica se já existe check-in do funcionário no evento
	ExistsByEmployeeAndEvent(ctx context.Context, employeeID, eventID value_objects.UUID) (bool, error)

	// GetByDateRange busca check-ins em um período
	GetByDateRange(ctx context.Context, tenantID value_objects.UUID, startDate, endDate time.Time, filters ListFilters) ([]*Checkin, int, error)

	// GetByMethod busca check-ins por método
	GetByMethod(ctx context.Context, tenantID value_objects.UUID, method string, filters ListFilters) ([]*Checkin, int, error)

	// GetValidCheckins busca apenas check-ins válidos
	GetValidCheckins(ctx context.Context, tenantID value_objects.UUID, filters ListFilters) ([]*Checkin, int, error)

	// GetInvalidCheckins busca apenas check-ins inválidos
	GetInvalidCheckins(ctx context.Context, tenantID value_objects.UUID, filters ListFilters) ([]*Checkin, int, error)

	// CountByTenant conta check-ins de um tenant
	CountByTenant(ctx context.Context, tenantID value_objects.UUID) (int, error)

	// CountByEvent conta check-ins de um evento
	CountByEvent(ctx context.Context, eventID value_objects.UUID) (int, error)

	// CountByEmployee conta check-ins de um funcionário
	CountByEmployee(ctx context.Context, employeeID value_objects.UUID) (int, error)

	// GetRecentCheckins busca check-ins recentes (últimas 24h)
	GetRecentCheckins(ctx context.Context, tenantID value_objects.UUID, limit int) ([]*Checkin, error)

	// GetCheckinsByLocation busca check-ins próximos a uma localização
	GetCheckinsByLocation(ctx context.Context, tenantID value_objects.UUID, location value_objects.Location, radiusKm float64, filters ListFilters) ([]*Checkin, int, error)
}

// ListFilters define os filtros para listagem de check-ins
type ListFilters struct {
	// Filtros básicos
	TenantID   *value_objects.UUID
	EventID    *value_objects.UUID
	EmployeeID *value_objects.UUID
	PartnerID  *value_objects.UUID

	// Filtros específicos
	Method   *string
	IsValid  *bool
	HasPhoto *bool

	// Filtros temporais
	StartDate *time.Time
	EndDate   *time.Time

	// Filtros geográficos
	Location *value_objects.Location
	RadiusKm *float64

	// Busca textual
	Search *string // Busca em notes

	// Paginação
	Page     int
	PageSize int

	// Ordenação
	OrderBy   string // checkin_time, created_at, updated_at, method
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
		"checkin_time": true,
		"created_at":   true,
		"updated_at":   true,
		"method":       true,
	}

	if f.OrderBy != "" && !validOrderFields[f.OrderBy] {
		f.OrderBy = "checkin_time" // Padrão: ordenar por horário do check-in
	}

	if f.OrderBy == "" {
		f.OrderBy = "checkin_time"
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

// GetOffset calcula o offset para paginação
func (f *ListFilters) GetOffset() int {
	return (f.Page - 1) * f.PageSize
}

// GetLimit retorna o limite para paginação
func (f *ListFilters) GetLimit() int {
	return f.PageSize
}

// HasTenantFilter verifica se há filtro por tenant
func (f *ListFilters) HasTenantFilter() bool {
	return f.TenantID != nil && !f.TenantID.IsZero()
}

// HasEventFilter verifica se há filtro por evento
func (f *ListFilters) HasEventFilter() bool {
	return f.EventID != nil && !f.EventID.IsZero()
}

// HasEmployeeFilter verifica se há filtro por funcionário
func (f *ListFilters) HasEmployeeFilter() bool {
	return f.EmployeeID != nil && !f.EmployeeID.IsZero()
}

// HasPartnerFilter verifica se há filtro por parceiro
func (f *ListFilters) HasPartnerFilter() bool {
	return f.PartnerID != nil && !f.PartnerID.IsZero()
}

// HasMethodFilter verifica se há filtro por método
func (f *ListFilters) HasMethodFilter() bool {
	return f.Method != nil && *f.Method != ""
}

// HasValidFilter verifica se há filtro por validade
func (f *ListFilters) HasValidFilter() bool {
	return f.IsValid != nil
}

// HasPhotoFilter verifica se há filtro por foto
func (f *ListFilters) HasPhotoFilter() bool {
	return f.HasPhoto != nil
}

// HasDateRangeFilter verifica se há filtro por período
func (f *ListFilters) HasDateRangeFilter() bool {
	return f.StartDate != nil || f.EndDate != nil
}

// HasLocationFilter verifica se há filtro por localização
func (f *ListFilters) HasLocationFilter() bool {
	return f.Location != nil && f.RadiusKm != nil
}

// HasSearchFilter verifica se há filtro de busca textual
func (f *ListFilters) HasSearchFilter() bool {
	return f.Search != nil && *f.Search != ""
}

// GetSearchTerm retorna o termo de busca limpo
func (f *ListFilters) GetSearchTerm() string {
	if f.Search == nil {
		return ""
	}
	return strings.TrimSpace(*f.Search)
}

// GetMethodFilter retorna o filtro de método
func (f *ListFilters) GetMethodFilter() string {
	if f.Method == nil {
		return ""
	}
	return *f.Method
}

// GetStartDate retorna a data inicial do filtro
func (f *ListFilters) GetStartDate() *time.Time {
	return f.StartDate
}

// GetEndDate retorna a data final do filtro
func (f *ListFilters) GetEndDate() *time.Time {
	return f.EndDate
}

// GetLocation retorna a localização do filtro
func (f *ListFilters) GetLocation() *value_objects.Location {
	return f.Location
}

// GetRadiusKm retorna o raio em km do filtro
func (f *ListFilters) GetRadiusKm() *float64 {
	return f.RadiusKm
}

// CheckinStats representa estatísticas de check-ins
type CheckinStats struct {
	TotalCheckins     int
	ValidCheckins     int
	InvalidCheckins   int
	PendingCheckins   int
	FacialCheckins    int
	QRCodeCheckins    int
	ManualCheckins    int
	CheckinsToday     int
	CheckinsThisWeek  int
	CheckinsThisMonth int
	AveragePerDay     float64
	LastCheckinTime   *time.Time
}

// StatsRepository define operações para estatísticas de check-ins
type StatsRepository interface {
	// GetTenantStats obtém estatísticas de um tenant
	GetTenantStats(ctx context.Context, tenantID value_objects.UUID) (*CheckinStats, error)

	// GetEventStats obtém estatísticas de um evento
	GetEventStats(ctx context.Context, eventID value_objects.UUID) (*CheckinStats, error)

	// GetEmployeeStats obtém estatísticas de um funcionário
	GetEmployeeStats(ctx context.Context, employeeID value_objects.UUID) (*CheckinStats, error)

	// GetPartnerStats obtém estatísticas de um parceiro
	GetPartnerStats(ctx context.Context, partnerID value_objects.UUID) (*CheckinStats, error)

	// GetDailyStats obtém estatísticas diárias
	GetDailyStats(ctx context.Context, tenantID value_objects.UUID, date time.Time) (*CheckinStats, error)

	// GetWeeklyStats obtém estatísticas semanais
	GetWeeklyStats(ctx context.Context, tenantID value_objects.UUID, startDate time.Time) (*CheckinStats, error)

	// GetMonthlyStats obtém estatísticas mensais
	GetMonthlyStats(ctx context.Context, tenantID value_objects.UUID, year int, month time.Month) (*CheckinStats, error)
}
