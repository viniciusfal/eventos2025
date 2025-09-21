# Próximos Passos - Roadmap Detalhado

**Última Atualização**: 21/09/2025 23:50  
**Status Atual**: Fase 6.3 80% Completa | Role Handler Implementado  
**Próxima Tarefa**: Corrigir Permission Handler ou avançar para Fase 6.4

## 🎯 Próximas Tarefas Imediatas

### ✅ COMPLETADO - Fases 1-4 + 6.1-6.2 + 6.3 Parcial
- **Fase 1**: Configuração Inicial e Infraestrutura (100%)
- **Fase 2**: Core Domain (Tenant, User, Auth JWT) (100%)
- **Fase 3**: Domínios Principais (100%)
  - Event, Partner, Employee, Role, Permission, Role-Permission
- **Fase 4**: Check-in/Check-out (100%)
  - Checkin, Checkout, WorkSession
- **Fase 6.1**: Configuração do Gin Framework (100%)
  - Router, Middleware, Responses, Error Handling
- **Fase 6.2**: Handlers de Tenant e User (100%)
  - Auth, Tenant, User handlers + User Repository + Testes
- **Fase 6.3**: Handlers de Domínios de Negócio (80%)
  - ✅ Event, Partner, Employee, Role handlers implementados
  - 🔄 Permission handler removido temporariamente por erros

### 🔄 OPÇÃO 1 - Completar Fase 6.3: Permission Handler

#### 6.3.1 Corrigir Permission Handler (PRÓXIMO IMEDIATO)

**Objetivo**: Corrigir e reimplementar o Permission Handler que foi removido temporariamente.

**Problema**: O Permission Handler original tinha erros de sintaxe nas chamadas de resposta HTTP.

**Arquivos a corrigir/recriar**:
- `internal/interfaces/http/handlers/permission_handler.go` - Recriar com sintaxe correta
- `internal/interfaces/http/router/router.go` - Reativar setupPermissionRoutes
- `cmd/api/main.go` - Reativar RolePermissionService

**Funcionalidades a implementar**:
```go
// Endpoints Permission
POST   /api/v1/permissions           - Criar permissão
GET    /api/v1/permissions/:id       - Buscar permissão
PUT    /api/v1/permissions/:id       - Atualizar permissão
DELETE /api/v1/permissions/:id       - Deletar permissão
GET    /api/v1/permissions           - Listar permissões
GET    /api/v1/permissions/system    - Listar permissões do sistema

// Endpoints Role-Permission Management
POST   /api/v1/roles/:role_id/permissions                - Conceder permissão
DELETE /api/v1/roles/:role_id/permissions/:permission_id - Revogar permissão
GET    /api/v1/roles/:role_id/permissions                - Listar permissões da role
PUT    /api/v1/roles/:role_id/permissions                - Sincronizar permissões
POST   /api/v1/roles/:role_id/permissions/bulk           - Conceder múltiplas permissões
```

**Correções necessárias**:
1. Usar `httpResponses.InternalServerError` em vez de `httpResponses.InternalError`
2. Usar `httpResponses.Unauthorized(c, "message")` em vez de `httpResponses.Unauthorized(c, "message", nil)`
3. Usar `httpResponses.NotFound(c, "message")` em vez de `httpResponses.NotFound(c, "message", nil)`
4. Usar `httpResponses.Forbidden(c, "message")` em vez de `httpResponses.Forbidden(c, "message", nil)`
5. Usar `httpResponses.Success(c, data, "message")` com parâmetros na ordem correta
6. Usar strings literais para tipos de erro: `"ValidationError"` em vez de `errors.ValidationError`
7. Usar `domainErr.Context` em vez de `domainErr.Details`

---

### 🔄 OPÇÃO 2 - Avançar para Fase 6.4: Handlers Check-in/Check-out

#### 6.4 Handlers de Check-in/Check-out (ALTERNATIVA)

**Objetivo**: Implementar handlers para Checkin, Checkout e WorkSessions.

**Arquivos a criar**:
- `internal/interfaces/http/handlers/checkin_handler.go`
- `internal/interfaces/http/handlers/checkout_handler.go`
- `internal/interfaces/http/handlers/work_session_handler.go`
- `internal/infrastructure/persistence/postgres/repositories/checkin_repository.go`
- `internal/infrastructure/persistence/postgres/repositories/checkout_repository.go`

**Endpoints Checkin**:
```go
POST   /api/v1/checkins         - Realizar check-in
GET    /api/v1/checkins/:id     - Buscar check-in
PUT    /api/v1/checkins/:id     - Atualizar check-in
GET    /api/v1/checkins         - Listar check-ins
POST   /api/v1/checkins/:id/validate - Validar check-in
POST   /api/v1/checkins/:id/notes - Adicionar nota
GET    /api/v1/checkins/stats   - Estatísticas de check-ins
```

**Endpoints Checkout**:
```go
POST   /api/v1/checkouts        - Realizar check-out
GET    /api/v1/checkouts/:id    - Buscar check-out
PUT    /api/v1/checkouts/:id    - Atualizar check-out
GET    /api/v1/checkouts        - Listar check-outs
POST   /api/v1/checkouts/:id/validate - Validar check-out
POST   /api/v1/checkouts/:id/notes - Adicionar nota
GET    /api/v1/checkouts/stats  - Estatísticas de check-outs
```

**Endpoints WorkSession**:
```go
GET    /api/v1/work-sessions    - Listar sessões de trabalho
GET    /api/v1/work-sessions/:employeeId - Sessões por funcionário
GET    /api/v1/work-sessions/:eventId - Sessões por evento
GET    /api/v1/work-sessions/stats - Estatísticas de trabalho
```

---

## 📋 Estrutura de Resposta Padronizada (Já Implementada)

### Resposta de Sucesso:
```go
httpResponses.Success(c, data, "message")
```

### Resposta de Erro:
```go
httpResponses.BadRequest(c, "message", details)
httpResponses.Unauthorized(c, "message")
httpResponses.Forbidden(c, "message")
httpResponses.NotFound(c, "message")
httpResponses.InternalServerError(c, "message")
```

### Resposta de Criação:
```go
httpResponses.Created(c, data, "message")
```

---

## 🔧 Padrão dos Handlers (Já Estabelecido)

### 1. Estrutura Básica:
```go
type Handler struct {
    service Service
    logger  *zap.Logger
}

func NewHandler(service Service, logger *zap.Logger) *Handler {
    return &Handler{service: service, logger: logger}
}
```

### 2. Tratamento de Erros:
```go
if domainErr, ok := err.(*errors.DomainError); ok {
    switch domainErr.Type {
    case "ValidationError":
        httpResponses.BadRequest(c, domainErr.Message, domainErr.Context)
    case "NotFoundError":
        httpResponses.NotFound(c, domainErr.Message)
    case "ForbiddenError":
        httpResponses.Forbidden(c, domainErr.Message)
    default:
        httpResponses.InternalServerError(c, "Operation failed")
    }
} else {
    httpResponses.InternalServerError(c, "Operation failed")
}
```

### 3. Autenticação:
```go
userClaims, exists := c.Get("user")
if !exists {
    httpResponses.Unauthorized(c, "Authentication required")
    return
}

claims, ok := userClaims.(*jwtService.Claims)
if !ok {
    httpResponses.InternalServerError(c, "Authentication error")
    return
}
```

---

## 🚀 Próximas Fases (Após Fase 6)

### Fase 7: Infraestrutura Avançada
- RolePermission Repository PostgreSQL concreto
- Cache com Redis
- Mensageria com RabbitMQ

### Fase 8: Testes
- Testes unitários estratégicos
- Testes de integração
- Testes E2E

### Fase 9: Monitoramento
- Métricas com Prometheus
- Tracing com OpenTelemetry
- Documentação Swagger

### Fase 10: Deployment
- CI/CD Pipeline
- Docker otimizado
- Preparação para produção

---

## 📝 Instruções para Próximo Agente

### ✅ O que está PRONTO:
- **10 domínios completos** implementados
- **7 handlers HTTP** funcionando (Auth, Tenant, User, Event, Partner, Employee, Role)
- **7 repositories PostgreSQL** implementados
- **Arquitetura Clean** estabelecida
- **Validações robustas** em todos os níveis
- **Multi-tenancy** funcionando
- **Sistema de permissões** hierárquico
- **API REST** funcionando para 7 domínios
- **Aplicação compilando e rodando** perfeitamente

### 🔄 O que fazer AGORA:
**OPÇÃO 1 (Recomendada)**: Corrigir Permission Handler
1. Recriar `permission_handler.go` com sintaxe correta
2. Reativar rotas no router
3. Testar endpoints de Permission

**OPÇÃO 2**: Avançar para Fase 6.4
1. Implementar Checkin Handler + Repository
2. Implementar Checkout Handler + Repository
3. Implementar WorkSession Handler

### 📋 Checklist Fase 6.3 Restante:
- [x] Event Handler + Repository ✅
- [x] Partner Handler + Repository ✅  
- [x] Employee Handler + Repository ✅
- [x] Role Handler + Repository ✅
- [ ] Permission Handler + Repository 🔄 (removido temporariamente)

### 📋 Checklist Fase 6.4:
- [ ] Checkin Handler + Repository 📋
- [ ] Checkout Handler + Repository 📋
- [ ] WorkSession Handler 📋
- [ ] Atualizar main.go e router com novos serviços 📋

### 🎯 Objetivo Final Fase 6:
Handlers completos para todos os domínios com funcionalidades avançadas (hierarquia de roles, role-permission management, reconhecimento facial, geolocalização, check-in/check-out inteligente).

**Status**: ✅ **7 handlers funcionando** | ✅ **Fase 6.3 80% completa** | 🔄 **Role Handler implementado** | 📋 **Permission Handler pendente** | 📋 **Seguindo plano rigorosamente**