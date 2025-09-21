package constants

// Status de entidades
const (
	StatusActive   = "active"
	StatusInactive = "inactive"
	StatusDeleted  = "deleted"
)

// Tipos de identidade
const (
	IdentityTypeCPF   = "cpf"
	IdentityTypeCNPJ  = "cnpj"
	IdentityTypeRG    = "rg"
	IdentityTypeOther = "other"
)

// Métodos de check-in/check-out
const (
	CheckMethodFacialRecognition = "facial_recognition"
	CheckMethodQRCode            = "qr_code"
	CheckMethodManual            = "manual"
)

// Tipos de QR Code
const (
	QRTypeCheckin  = "checkin"
	QRTypeCheckout = "checkout"
)

// Tipos de log
const (
	LogTypeSystem   = "system"
	LogTypeAudit    = "audit"
	LogTypeEvent    = "event"
	LogTypeSecurity = "security"
)

// Ações de auditoria
const (
	ActionCreate = "CREATE"
	ActionUpdate = "UPDATE"
	ActionDelete = "DELETE"
	ActionRead   = "READ"
)

// Tipos de entidade para logs
const (
	EntityTypeTenant   = "tenant"
	EntityTypeUser     = "user"
	EntityTypeEvent    = "event"
	EntityTypePartner  = "partner"
	EntityTypeEmployee = "employee"
	EntityTypeCheckin  = "checkin"
	EntityTypeCheckout = "checkout"
)

// Módulos do sistema
const (
	ModuleAuth      = "auth"
	ModuleEvents    = "events"
	ModulePartners  = "partners"
	ModuleEmployees = "employees"
	ModuleCheckins  = "checkins"
	ModuleReports   = "reports"
	ModuleAudit     = "audit"
	ModuleQRCode    = "qr_code"
	ModuleFacial    = "facial"
)

// Permissões básicas
const (
	PermissionRead   = "read"
	PermissionWrite  = "write"
	PermissionDelete = "delete"
	PermissionAdmin  = "admin"
)

// Roles padrão do sistema
const (
	RoleSuperAdmin = "SUPER_ADMIN"
	RoleAdmin      = "ADMIN"
	RoleManager    = "MANAGER"
	RoleOperator   = "OPERATOR"
	RoleViewer     = "VIEWER"
)

// Configurações de paginação
const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// Configurações de cache (em segundos)
const (
	CacheTTLShort  = 300  // 5 minutos
	CacheTTLMedium = 1800 // 30 minutos
	CacheTTLLong   = 3600 // 1 hora
)

// Configurações de QR Code
const (
	QRCodeValidityDuration = 60 // segundos
	QRCodeMaxUsage         = 1  // número máximo de usos
)
