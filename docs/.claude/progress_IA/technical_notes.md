# Notas Técnicas Importantes

## 🔧 Configurações Críticas

### 1. JWT Secret
**IMPORTANTE**: O JWT secret deve ser alterado em produção!
- **Desenvolvimento**: `desenvolvimento-jwt-secret-key-super-secreto-para-desenvolvimento-apenas`
- **Produção**: Deve ser definido via variável de ambiente `JWT_SECRET`
- **Validação**: Sistema rejeita o secret padrão em produção

### 2. Configuração de Banco
**Connection String**: 
```
host=localhost port=5432 user=eventos_user password=eventos_password dbname=eventos_db sslmode=disable
```

**Pool Settings**:
- Max Open Connections: 25
- Max Idle Connections: 5
- Connection Max Lifetime: 5 minutos

### 3. Extensões PostgreSQL Necessárias
```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "postgis";
```

---

## 🏗️ Padrões de Implementação

### 4. Estrutura de Domínio
Todos os domínios seguem o padrão:
```
internal/domain/{domain}/
├── {domain}.go          # Entidade + validações + regras de negócio
├── repository.go        # Interface do repositório
└── service.go          # Serviços de domínio + orquestração
```

### 5. Padrão de Validação
```go
// 1. Validação na criação da entidade
func NewEntity(...) (*Entity, error) {
    if err := validateEntityData(...); err != nil {
        return nil, err
    }
    // ...
}

// 2. Validação nos métodos de update
func (e *Entity) Update(...) error {
    if err := validateEntityData(...); err != nil {
        return err
    }
    // ...
}

// 3. Validação nos serviços de domínio
func (s *Service) CreateEntity(...) (*Entity, error) {
    // Validações de negócio (unicidade, etc.)
    // ...
}
```

### 6. Padrão de Erro
```go
// Erros de validação
return errors.NewValidationError("field", "message")

// Erros de negócio
return errors.NewAlreadyExistsError("resource", "field", value)

// Erros internos
return errors.NewInternalError("message", cause)
```

---

## 🔒 Segurança

### 7. Hash de Senhas
```go
// Sempre usar bcrypt.DefaultCost (10 rounds)
hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

// Verificação
err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
```

### 8. Validação de UUID
```go
// Sempre validar UUIDs de entrada
id, err := value_objects.ParseUUID(idString)
if err != nil {
    return errors.NewValidationError("id", "invalid UUID format")
}
```

### 9. Multi-tenancy
**CRÍTICO**: Todas as queries devem incluir tenant_id!
```go
// ❌ ERRADO - vazamento de dados
SELECT * FROM users WHERE id = $1

// ✅ CORRETO - isolamento por tenant
SELECT * FROM users WHERE id = $1 AND id_tenant = $2
```

---

## 📊 Banco de Dados

### 10. Convenções de Nomenclatura
- **Tabelas**: singular em inglês (`user`, `event`, `partner`)
- **Colunas**: snake_case (`id_tenant`, `created_at`, `full_name`)
- **PKs**: `id_{table}` (`id_user`, `id_event`)
- **FKs**: `id_{referenced_table}` (`id_tenant`, `id_user`)

### 11. Índices Importantes
```sql
-- Multi-tenancy (OBRIGATÓRIO em todas as tabelas)
CREATE INDEX idx_user_id_tenant ON "user"(id_tenant);
CREATE INDEX idx_event_id_tenant ON event(id_tenant);

-- Campos de busca
CREATE INDEX idx_user_username ON "user"(username);
CREATE INDEX idx_user_email ON "user"(email);

-- JSONB (para campos de configuração)
CREATE INDEX idx_config_tenant_modules ON config_tenant USING GIN (modules);
```

### 12. Triggers de Updated_at
```sql
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_user_updated_at 
    BEFORE UPDATE ON "user" 
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
```

---

## 🌍 Geolocalização

### 13. Formato de Coordenadas
```go
// Latitude: -90 a 90
// Longitude: -180 a 180
type Location struct {
    Latitude  float64  // Y coordinate
    Longitude float64  // X coordinate
}

// PostGIS format: "POINT(longitude latitude)"
func (l Location) String() string {
    return fmt.Sprintf("POINT(%f %f)", l.Longitude, l.Latitude)
}
```

### 14. Cálculo de Distância
```go
// Fórmula de Haversine para distância em metros
func (l Location) DistanceTo(other Location) float64 {
    const earthRadius = 6371000 // metros
    // ... implementação
}
```

### 15. Geofencing
```go
// Point-in-polygon usando ray casting
func isPointInPolygon(point Location, polygon []Location) bool {
    // Algoritmo ray casting
    // ... implementação
}
```

---

## 🤖 Reconhecimento Facial

### 16. Embeddings Faciais
```go
// Sempre 512 dimensões
type Employee struct {
    FaceEmbedding []float32 // len(FaceEmbedding) == 512
}

// Validação obrigatória
if len(embedding) != 512 {
    return errors.NewValidationError("face_embedding", "must have 512 dimensions")
}
```

### 17. Similaridade Coseno
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

### 18. Thresholds de Confiança
```go
const (
    HighConfidenceThreshold   = 0.9   // >= 90% similaridade
    MediumConfidenceThreshold = 0.75  // >= 75% similaridade
    DefaultThreshold          = 0.75  // Padrão do sistema
)
```

---

## 🔄 Paginação

### 19. Padrão de Filtros
```go
type ListFilters struct {
    // Filtros específicos
    TenantID *value_objects.UUID
    Active   *bool
    
    // Paginação (OBRIGATÓRIA)
    Page     int  // Mínimo 1
    PageSize int  // Padrão 20, máximo 100
    
    // Ordenação
    OrderBy   string // Campo válido
    OrderDesc bool   // Direção
}

func (f *ListFilters) Validate() error {
    if f.Page < 1 { f.Page = 1 }
    if f.PageSize < 1 { f.PageSize = 20 }
    if f.PageSize > 100 { f.PageSize = 100 }
    // ...
}
```

### 20. Cálculo de Offset
```go
func (f *ListFilters) GetOffset() int {
    return (f.Page - 1) * f.PageSize
}
```

---

## 🚨 Problemas Conhecidos e Soluções

### 21. Windows PowerShell
**Problema**: Comandos Unix não funcionam no PowerShell  
**Solução**: Usar comandos PowerShell equivalentes
```powershell
# ❌ mkdir -p
# ✅ New-Item -ItemType Directory -Force

# ❌ chmod +x
# ✅ Não necessário no Windows
```

### 22. Compilação Go
**Problema**: Imports não utilizados causam erro  
**Solução**: Sempre remover imports desnecessários
```bash
go build ./internal/domain/...  # Testar domínios
go build -o build/main ./cmd/api  # Testar aplicação completa
```

### 23. Configuração JWT
**Problema**: Erro "JWT secret must be set and changed from default"  
**Solução**: Configurar secret adequado para ambiente
```go
// Desenvolvimento: OK usar secret fixo
// Produção: DEVE usar variável de ambiente
```

---

## 📝 Convenções de Código

### 24. Nomenclatura
```go
// Interfaces: substantivo + "er" ou "Service"
type Repository interface {}
type Service interface {}

// Implementações: "Domain" + Interface
type DomainService struct {}

// Value Objects: substantivo simples
type UUID struct {}
type Location struct {}

// Entidades: substantivo singular
type User struct {}
type Event struct {}
```

### 25. Comentários
```go
// Funções públicas: sempre documentar
// CreateUser cria um novo usuário com validações de negócio
func (s *Service) CreateUser(...) (*User, error) {}

// Constantes: documentar propósito
const (
    DefaultPageSize = 20  // Tamanho padrão de página
    MaxPageSize     = 100 // Tamanho máximo permitido
)
```

### 26. Logs Estruturados
```go
// Sempre usar campos estruturados
s.logger.Info("User created successfully",
    zap.String("user_id", user.ID.String()),
    zap.String("username", user.Username),
    zap.String("tenant_id", user.TenantID.String()),
)

// Níveis apropriados
s.logger.Debug("...")  // Desenvolvimento
s.logger.Info("...")   // Operações normais
s.logger.Warn("...")   // Situações inesperadas
s.logger.Error("...")  // Erros que precisam atenção
```

---

## 🔧 Comandos Úteis

### 27. Desenvolvimento
```bash
# Compilar e testar
go build -o build/main ./cmd/api
./build/main.exe

# Executar com logs
go run ./cmd/api

# Testes
go test ./...
go test -v ./internal/domain/...

# Dependências
go mod tidy
go mod download
```

### 28. Docker
```bash
# Subir infraestrutura
docker-compose up -d postgres redis rabbitmq

# Subir monitoramento
docker-compose up -d prometheus grafana

# Ver logs
docker-compose logs -f postgres

# Parar tudo
docker-compose down
```

### 29. Banco de Dados
```bash
# Conectar ao banco
PGPASSWORD=eventos_password psql -h localhost -U eventos_user -d eventos_db

# Executar migrações
./scripts/migrate.sh

# Backup
PGPASSWORD=eventos_password pg_dump -h localhost -U eventos_user eventos_db > backup.sql
```

---

## ⚠️ Cuidados Importantes

### 30. Nunca Fazer
- ❌ Query sem tenant_id em produção
- ❌ Commit de senhas ou secrets
- ❌ Hard delete de dados importantes
- ❌ Quebrar interface de repositório
- ❌ Validação apenas no frontend

### 31. Sempre Fazer
- ✅ Testar compilação após mudanças
- ✅ Validar entrada em múltiplas camadas
- ✅ Usar transações para operações críticas
- ✅ Log de operações importantes
- ✅ Documentar decisões arquiteturais
