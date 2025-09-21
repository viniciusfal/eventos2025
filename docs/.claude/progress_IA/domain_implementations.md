# Implementações dos Domínios - Detalhes Técnicos

## 🏗️ Arquitetura dos Domínios

Todos os domínios seguem o padrão Clean Architecture com DDD:

```
internal/domain/{domain}/
├── {domain}.go          # Entidade principal
├── repository.go        # Interface do repositório
└── service.go          # Serviços de domínio
```

---

## 🏢 Domínio Tenant

### Entidade Principal
```go
type Tenant struct {
    ID           value_objects.UUID
    ConfigID     *value_objects.UUID
    Name         string
    Identity     string      // CPF/CNPJ
    IdentityType string
    Email        string
    Address      string
    Active       bool
    CreatedAt    time.Time
    UpdatedAt    time.Time
    CreatedBy    *value_objects.UUID
    UpdatedBy    *value_objects.UUID
}
```

### Funcionalidades
- ✅ Validação de nome (2-255 chars)
- ✅ Validação de identidade (CPF/CNPJ)
- ✅ Validação de email
- ✅ Ativação/desativação
- ✅ Configuração de módulos (JSON)

### Regras de Negócio
- Nome obrigatório
- Identidade única por tenant
- Email único por tenant
- Não pode deletar tenant ativo

---

## 👤 Domínio User

### Entidade Principal
```go
type User struct {
    ID        value_objects.UUID
    TenantID  value_objects.UUID
    FullName  string
    Email     string
    Phone     string
    Username  string
    Password  string // Hash bcrypt
    Active    bool
    CreatedAt time.Time
    UpdatedAt time.Time
    CreatedBy *value_objects.UUID
    UpdatedBy *value_objects.UUID
}
```

### Funcionalidades
- ✅ Hash de senhas com bcrypt
- ✅ Validação de senha (8+ chars, letra + número)
- ✅ Username único por tenant
- ✅ Email único por tenant
- ✅ Autenticação com verificação de senha

### Regras de Negócio
- Username obrigatório (3-50 chars)
- Senha forte obrigatória
- Não pode deletar usuário ativo
- Multi-tenancy isolado

---

## 🎪 Domínio Event

### Entidade Principal
```go
type Event struct {
    ID          value_objects.UUID
    TenantID    value_objects.UUID
    Name        string
    Location    string
    FenceEvent  []value_objects.Location // Polígono
    InitialDate time.Time
    FinalDate   time.Time
    Active      bool
    CreatedAt   time.Time
    UpdatedAt   time.Time
    CreatedBy   *value_objects.UUID
    UpdatedBy   *value_objects.UUID
}
```

### Funcionalidades Geográficas
- ✅ **Geofencing**: Verificação point-in-polygon
- ✅ **Cálculo de Distância**: Fórmula de Haversine
- ✅ **Validação de Coordenadas**: Lat (-90,90), Lng (-180,180)
- ✅ **PostGIS Integration**: Preparado para consultas espaciais

### Value Object Location
```go
type Location struct {
    Latitude  float64
    Longitude float64
}

// Métodos principais
func (l Location) DistanceTo(other Location) float64
func (l Location) String() string // "POINT(lng lat)"
func (l Location) Value() (driver.Value, error) // PostGIS
```

### Funcionalidades Temporais
- ✅ **Status Temporal**: IsOngoing(), IsUpcoming(), IsFinished()
- ✅ **Validações**: Data final > inicial, duração máx 30 dias
- ✅ **Check-in/Check-out**: Validações por período

### Regras de Negócio
- Nome único por tenant
- Data final deve ser após inicial
- Duração máxima de 30 dias
- Não pode alterar data inicial de evento em andamento
- Não pode desativar evento em andamento

---

## 🤝 Domínio Partner

### Entidade Principal
```go
type Partner struct {
    ID                   value_objects.UUID
    TenantID             value_objects.UUID
    Name                 string
    Email                string
    Email2               string // Email secundário
    Phone                string
    Phone2               string // Telefone secundário
    Identity             string
    IdentityType         string
    Location             string
    PasswordHash         string
    LastLogin            *time.Time
    FailedLoginAttempts  int
    LockedUntil          *time.Time
    Active               bool
    CreatedAt            time.Time
    UpdatedAt            time.Time
    CreatedBy            *value_objects.UUID
    UpdatedBy            *value_objects.UUID
}
```

### Sistema de Autenticação
- ✅ **Hash de Senhas**: bcrypt
- ✅ **Controle de Tentativas**: 5 tentativas máx
- ✅ **Bloqueio Automático**: 30 minutos após 5 falhas
- ✅ **Desbloqueio**: Manual ou automático
- ✅ **Último Login**: Registro de acesso

### Funcionalidades
- ✅ Login por email ou identidade
- ✅ Campos redundantes (email2, phone2)
- ✅ Validações de identidade e email
- ✅ Relacionamento com eventos

### Regras de Negócio
- Nome obrigatório (2-255 chars)
- Email único por tenant (se fornecido)
- Identidade única por tenant (se fornecida)
- Senha opcional (pode ser definida depois)
- Não pode deletar parceiro ativo

---

## 👷 Domínio Employee

### Entidade Principal
```go
type Employee struct {
    ID            value_objects.UUID
    TenantID      value_objects.UUID
    FullName      string
    Identity      string
    IdentityType  string
    DateOfBirth   *time.Time
    PhotoURL      string
    FaceEmbedding []float32 // 512 dimensões
    Phone         string
    Email         string
    Active        bool
    CreatedAt     time.Time
    UpdatedAt     time.Time
    CreatedBy     *value_objects.UUID
    UpdatedBy     *value_objects.UUID
}
```

### Sistema de Reconhecimento Facial
- ✅ **Embeddings**: Vetores de 512 dimensões
- ✅ **Similaridade Coseno**: Algoritmo de comparação
- ✅ **Thresholds**: Configuráveis (0.5-1.0)
- ✅ **Níveis de Confiança**: High (≥0.9), Medium (≥0.75), Low (<0.75)
- ✅ **Validação**: Valores entre -1.0 e 1.0

### Algoritmo de Similaridade
```go
func cosineSimilarity(a, b []float32) float32 {
    var dotProduct, normA, normB float32
    for i := 0; i < len(a); i++ {
        dotProduct += a[i] * b[i]
        normA += a[i] * a[i]
        normB += b[i] * b[i]
    }
    return dotProduct / (sqrt32(normA) * sqrt32(normB))
}
```

### Validações de Idade
- ✅ **Idade Mínima**: 14 anos (legislação trabalhista)
- ✅ **Idade Máxima**: 120 anos (validação realista)
- ✅ **Data Futura**: Não permite nascimento futuro
- ✅ **Cálculo de Idade**: Considera ano e dia do ano

### Funcionalidades
- ✅ Gestão de fotos (URL)
- ✅ Reconhecimento facial
- ✅ Validações de dados pessoais
- ✅ Relacionamento com parceiros
- ✅ Filtros por idade, foto, embedding

### Regras de Negócio
- Nome obrigatório (2-255 chars)
- Identidade única por tenant (se fornecida)
- Email único por tenant (se fornecido)
- Idade mínima 14 anos
- Embedding deve ter exatamente 512 dimensões
- Não pode deletar funcionário ativo

---

## 🔧 Value Objects Compartilhados

### UUID
```go
type UUID struct {
    value uuid.UUID
}
// Métodos: NewUUID(), ParseUUID(), String(), IsZero(), Equals()
// Implementa: driver.Valuer, sql.Scanner
```

### Location
```go
type Location struct {
    Latitude  float64
    Longitude float64
}
// Métodos: DistanceTo(), String(), Value(), Scan()
// Validações: Lat (-90,90), Lng (-180,180)
```

---

## 🚨 Sistema de Erros

### Domain Errors
```go
type DomainError struct {
    Type    string
    Message string
    Cause   error
    Context map[string]interface{}
}
```

### Tipos de Erro
- ✅ **NotFoundError**: Recurso não encontrado
- ✅ **AlreadyExistsError**: Recurso já existe
- ✅ **ValidationError**: Validação falhou
- ✅ **UnauthorizedError**: Não autorizado
- ✅ **ForbiddenError**: Acesso negado
- ✅ **InternalError**: Erro interno

---

## 📊 Constantes do Sistema

### Status
- `StatusActive`, `StatusInactive`, `StatusDeleted`

### Tipos de Identidade
- `IdentityTypeCPF`, `IdentityTypeCNPJ`, `IdentityTypeRG`, `IdentityTypeOther`

### Métodos de Check-in
- `CheckMethodFacialRecognition`, `CheckMethodQRCode`, `CheckMethodManual`

### Módulos
- `ModuleAuth`, `ModuleEvents`, `ModulePartners`, `ModuleEmployees`, etc.

---

## 🔄 Padrões de Implementação

### Repository Pattern
```go
type Repository interface {
    Create(ctx context.Context, entity *Entity) error
    GetByID(ctx context.Context, id UUID) (*Entity, error)
    Update(ctx context.Context, entity *Entity) error
    Delete(ctx context.Context, id UUID, deletedBy UUID) error
    List(ctx context.Context, filters ListFilters) ([]*Entity, int, error)
}
```

### Service Pattern
```go
type Service interface {
    CreateEntity(ctx context.Context, ...) (*Entity, error)
    UpdateEntity(ctx context.Context, ...) (*Entity, error)
    GetEntity(ctx context.Context, id UUID) (*Entity, error)
    // ... outros métodos
}
```

### Filtros de Listagem
```go
type ListFilters struct {
    // Filtros específicos
    TenantID *UUID
    Active   *bool
    
    // Paginação
    Page     int
    PageSize int
    
    // Ordenação
    OrderBy   string
    OrderDesc bool
}
```

---

## 🎯 Domínio Role

### Entidade Principal
```go
type Role struct {
    ID          value_objects.UUID
    TenantID    value_objects.UUID
    Name        string
    DisplayName string
    Description string
    Level       int  // Nível hierárquico (1=mais alto, 999=mais baixo)
    IsSystem    bool // Roles do sistema não podem ser editados
    Active      bool
    CreatedAt   time.Time
    UpdatedAt   time.Time
    CreatedBy   *value_objects.UUID
    UpdatedBy   *value_objects.UUID
}
```

### Funcionalidades
- ✅ Hierarquia de níveis (1-999)
- ✅ Roles do sistema pré-definidas
- ✅ Validações de nome e nível
- ✅ Gerenciamento hierárquico
- ✅ Multi-tenancy isolado

### Regras de Negócio
- Nome único por tenant (formato UPPER_CASE)
- Nível único por tenant
- Roles superiores podem gerenciar inferiores
- Roles do sistema (1-9) não podem ser alteradas
- Roles customizadas (10-999) podem ser gerenciadas

---

## 🔐 Domínio Permission

### Entidade Principal
```go
type Permission struct {
    ID          value_objects.UUID
    TenantID    value_objects.UUID
    Module      string // auth, events, partners, etc.
    Action      string // read, write, delete, admin
    Resource    string // Recurso específico (opcional)
    Name        string // MODULE_ACTION ou MODULE_ACTION_RESOURCE
    DisplayName string
    Description string
    IsSystem    bool
    Active      bool
    CreatedAt   time.Time
    UpdatedAt   time.Time
    CreatedBy   *value_objects.UUID
    UpdatedBy   *value_objects.UUID
}
```

### Funcionalidades
- ✅ Permissões granulares por módulo/ação/recurso
- ✅ Geração automática de nomes
- ✅ Pattern matching para verificação
- ✅ Permissões do sistema pré-definidas
- ✅ Bulk operations

### Regras de Negócio
- Nome único por tenant
- Módulo deve ser reconhecido pelo sistema
- Ação deve ser válida (read/write/delete/admin)
- Permissões do sistema não podem ser alteradas
- Pattern matching flexível para verificação de acesso

---

## 🔗 Relacionamento Role-Permission

### Entidade Principal
```go
type RolePermission struct {
    ID           value_objects.UUID
    RoleID       value_objects.UUID
    PermissionID value_objects.UUID
    TenantID     value_objects.UUID
    GrantedBy    value_objects.UUID
    GrantedAt    time.Time
    Active       bool
}
```

### Funcionalidades
- ✅ Relacionamento Many-to-Many
- ✅ Grant/Revoke individual e em lote
- ✅ Sincronização completa
- ✅ Auditoria de concessões
- ✅ Validação de hierarquia

### Regras de Negócio
- Uma role pode ter múltiplas permissões
- Uma permissão pode ser concedida a múltiplas roles
- Auditoria completa de quem concedeu e quando
- Respeita hierarquia de roles
- Sincronização remove antigas e adiciona novas

---

## ✅ Domínio Checkin

### Entidade Principal
```go
type Checkin struct {
    ID                value_objects.UUID
    TenantID          value_objects.UUID
    EventID           value_objects.UUID
    EmployeeID        value_objects.UUID
    PartnerID         value_objects.UUID
    Method            string // facial_recognition, qr_code, manual
    Location          value_objects.Location
    CheckinTime       time.Time
    PhotoURL          string
    Notes             string
    IsValid           bool
    ValidationDetails map[string]interface{}
    CreatedAt         time.Time
    UpdatedAt         time.Time
    CreatedBy         *value_objects.UUID
    UpdatedBy         *value_objects.UUID
}
```

### Funcionalidades
- ✅ Múltiplos métodos de check-in
- ✅ Validações geográficas e temporais
- ✅ Reconhecimento facial com embeddings
- ✅ Sistema de validação detalhado
- ✅ Estatísticas completas

### Regras de Negócio
- Um funcionário pode ter apenas um check-in por evento
- Check-in deve estar dentro da cerca do evento
- Validação facial com threshold configurável
- Auditoria completa de tentativas
- Estatísticas por método, validade e período

---

## ⏰ Domínio Checkout

### Entidade Principal
```go
type Checkout struct {
    ID                value_objects.UUID
    TenantID          value_objects.UUID
    EventID           value_objects.UUID
    EmployeeID        value_objects.UUID
    PartnerID         value_objects.UUID
    CheckinID         value_objects.UUID // Referência ao check-in
    Method            string
    Location          value_objects.Location
    CheckoutTime      time.Time
    PhotoURL          string
    Notes             string
    WorkDuration      time.Duration
    IsValid           bool
    ValidationDetails map[string]interface{}
    CreatedAt         time.Time
    UpdatedAt         time.Time
    CreatedBy         *value_objects.UUID
    UpdatedBy         *value_objects.UUID
}
```

### Funcionalidades
- ✅ Checkout vinculado ao check-in
- ✅ Cálculo automático de duração
- ✅ Validações de duração de trabalho
- ✅ Sessões de trabalho completas
- ✅ Estatísticas de trabalho

### Regras de Negócio
- Deve existir check-in correspondente
- Duração calculada automaticamente
- Validação de trabalho curto/longo
- Sessões completas para análise
- Estatísticas de horas trabalhadas

---

## 💼 WorkSession (Sessão de Trabalho)

### Entidade Principal
```go
type WorkSession struct {
    CheckinID    value_objects.UUID
    CheckoutID   value_objects.UUID
    EmployeeID   value_objects.UUID
    EventID      value_objects.UUID
    PartnerID    value_objects.UUID
    CheckinTime  time.Time
    CheckoutTime time.Time
    Duration     time.Duration
    IsComplete   bool
    IsValid      bool
}
```

### Funcionalidades
- ✅ Sessão completa de trabalho
- ✅ Cálculo de duração total
- ✅ Validações de sessão
- ✅ Análise de produtividade
- ✅ Relatórios de trabalho

Todos os domínios seguem esses padrões consistentemente, garantindo manutenibilidade e extensibilidade.
