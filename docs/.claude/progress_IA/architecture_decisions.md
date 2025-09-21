# Decisões Arquiteturais - Registro de Decisões

## 🏗️ Decisões de Arquitetura Principal

### 1. Clean Architecture + DDD
**Decisão**: Usar Clean Architecture com inspiração Domain-Driven Design  
**Razão**: Separação clara de responsabilidades, testabilidade, manutenibilidade  
**Impacto**: Estrutura de pastas bem definida, dependências invertidas  
**Status**: ✅ Implementado

### 2. Multi-tenancy por Tenant ID
**Decisão**: Implementar multi-tenancy usando tenant_id em todas as entidades  
**Razão**: Isolamento de dados, escalabilidade, simplicidade  
**Alternativas Rejeitadas**: Schema por tenant, database por tenant  
**Status**: ✅ Implementado

### 3. PostgreSQL + PostGIS
**Decisão**: PostgreSQL como banco principal com extensão PostGIS  
**Razão**: Recursos geográficos nativos, robustez, ACID compliance  
**Uso**: Geolocalização de eventos, cálculos de distância, geofencing  
**Status**: ✅ Configurado

---

## 🔧 Decisões Técnicas

### 4. Go + Gin Framework
**Decisão**: Go como linguagem principal, Gin para HTTP  
**Razão**: Performance, simplicidade, ecosystem maduro  
**Alternativas**: Echo, Fiber, net/http puro  
**Status**: ✅ Implementado

### 5. JWT para Autenticação
**Decisão**: JWT com access token + refresh token  
**Razão**: Stateless, escalável, padrão da indústria  
**Configuração**: 
- Access token: 1 hora
- Refresh token: 7 dias
- HMAC-SHA256
**Status**: ✅ Implementado

### 6. bcrypt para Senhas
**Decisão**: bcrypt para hash de senhas  
**Razão**: Segurança comprovada, resistente a rainbow tables  
**Configuração**: DefaultCost (10 rounds)  
**Status**: ✅ Implementado

### 7. UUID como Identificadores
**Decisão**: UUID v4 para todos os IDs  
**Razão**: Únicos globalmente, não sequenciais, segurança  
**Implementação**: Value Object UUID com validações  
**Status**: ✅ Implementado

---

## 🎯 Decisões de Domínio

### 8. Reconhecimento Facial
**Decisão**: Embeddings de 512 dimensões + similaridade coseno  
**Razão**: Padrão da indústria, boa precisão, eficiência  
**Thresholds**:
- High confidence: ≥ 0.9
- Medium confidence: ≥ 0.75
- Low confidence: < 0.75
**Status**: ✅ Implementado

### 9. Geofencing com Polígonos
**Decisão**: Cerca de eventos como array de coordenadas (polígono)  
**Razão**: Flexibilidade para formas complexas  
**Algoritmo**: Point-in-polygon (ray casting)  
**Limitações**: Máximo 100 pontos por polígono  
**Status**: ✅ Implementado

### 10. Sistema de Bloqueio de Parceiros
**Decisão**: Bloqueio automático após 5 tentativas falhadas  
**Razão**: Segurança contra ataques de força bruta  
**Configuração**:
- 5 tentativas máximas
- Bloqueio por 30 minutos
- Desbloqueio automático ou manual
**Status**: ✅ Implementado

---

## 📊 Decisões de Dados

### 11. Soft Delete
**Decisão**: Soft delete para todas as entidades principais  
**Razão**: Auditoria, recuperação de dados, integridade referencial  
**Implementação**: Campo `active` boolean  
**Status**: ✅ Implementado

### 12. Timestamps Automáticos
**Decisão**: `created_at` e `updated_at` em todas as entidades  
**Razão**: Auditoria, debugging, ordenação  
**Implementação**: Triggers PostgreSQL + aplicação  
**Status**: ✅ Implementado

### 13. Campos de Auditoria
**Decisão**: `created_by` e `updated_by` em todas as entidades  
**Razão**: Rastreabilidade de mudanças  
**Tipo**: UUID referenciando usuário  
**Status**: ✅ Implementado

---

## 🔒 Decisões de Segurança

### 14. Validação em Múltiplas Camadas
**Decisão**: Validação no Domain + Application + Interface  
**Razão**: Defense in depth, consistência  
**Implementação**:
- Domain: Regras de negócio
- Application: DTOs e casos de uso
- Interface: Validação de entrada
**Status**: ✅ Implementado

### 15. Isolamento por Tenant
**Decisão**: Todas as queries incluem tenant_id  
**Razão**: Segurança, isolamento de dados  
**Implementação**: Filtros automáticos nos repositórios  
**Status**: ✅ Implementado

### 16. Rate Limiting
**Decisão**: Rate limiting por IP e usuário  
**Razão**: Proteção contra abuso  
**Configuração**: 100 requests/minuto, burst 10  
**Status**: 📋 Configurado (não implementado)

---

## 🚀 Decisões de Performance

### 17. Connection Pooling
**Decisão**: Pool de conexões PostgreSQL  
**Razão**: Performance, controle de recursos  
**Configuração**:
- Max open: 25
- Max idle: 5
- Max lifetime: 5 minutos
**Status**: ✅ Implementado

### 18. Índices de Banco
**Decisão**: Índices estratégicos para queries frequentes  
**Implementação**:
- tenant_id em todas as tabelas
- Campos de busca (email, identity, username)
- Campos de ordenação (created_at, updated_at)
- Índices GIN para JSONB
**Status**: ✅ Implementado

### 19. Paginação Padrão
**Decisão**: Paginação obrigatória em listagens  
**Razão**: Performance, UX  
**Configuração**:
- Página padrão: 20 itens
- Máximo: 100 itens
**Status**: ✅ Implementado

---

## 🔄 Decisões de Integração

### 20. Docker para Desenvolvimento
**Decisão**: Docker Compose para ambiente local  
**Razão**: Consistência, facilidade de setup  
**Serviços**: PostgreSQL, Redis, RabbitMQ, Prometheus, Grafana  
**Status**: ✅ Implementado

### 21. Structured Logging
**Decisão**: Zap para logging estruturado  
**Razão**: Performance, estruturação, integração com monitoring  
**Formato**: JSON em produção, console em desenvolvimento  
**Status**: ✅ Implementado

### 22. Monitoring Stack
**Decisão**: Prometheus + Grafana para monitoramento  
**Razão**: Padrão da indústria, flexibilidade  
**Métricas**: HTTP requests, database connections, business metrics  
**Status**: 📋 Configurado (métricas não implementadas)

---

## 🎨 Decisões de API Design

### 23. RESTful API
**Decisão**: API REST seguindo convenções HTTP  
**Razão**: Simplicidade, padrão amplamente conhecido  
**Estrutura**: `/api/v1/{resource}`  
**Status**: 📋 Planejado

### 24. JSON para Comunicação
**Decisão**: JSON para requests e responses  
**Razão**: Simplicidade, suporte universal  
**Alternativas Rejeitadas**: XML, Protocol Buffers  
**Status**: ✅ Implementado

### 25. Versionamento de API
**Decisão**: Versionamento via URL path (`/api/v1/`)  
**Razão**: Clareza, facilidade de implementação  
**Alternativas**: Headers, query params  
**Status**: 📋 Planejado

---

## 📝 Decisões de Documentação

### 26. Swagger/OpenAPI
**Decisão**: Swagger para documentação da API  
**Razão**: Padrão da indústria, geração automática  
**Ferramenta**: swaggo/swag  
**Status**: 📋 Planejado

### 27. Documentação de Progresso
**Decisão**: Documentação estruturada para continuidade  
**Razão**: Permitir que qualquer agente continue o trabalho  
**Localização**: `docs/.claude/progress_IA/`  
**Status**: ✅ Implementado

---

## 🔄 Decisões Reversíveis vs Irreversíveis

### Irreversíveis (Alto Custo de Mudança)
- ✅ PostgreSQL como banco principal
- ✅ Go como linguagem
- ✅ Multi-tenancy por tenant_id
- ✅ UUID como identificadores

### Reversíveis (Baixo Custo de Mudança)
- 📋 Gin Framework (pode trocar por Echo/Fiber)
- 📋 JWT (pode adicionar sessions)
- 📋 Zap Logger (pode trocar por logrus)
- 📋 Docker Compose (pode usar Kubernetes)

---

## 📊 Impacto das Decisões

### Positivos
- ✅ Arquitetura limpa e testável
- ✅ Segurança robusta
- ✅ Performance adequada
- ✅ Escalabilidade horizontal
- ✅ Manutenibilidade alta

### Trade-offs
- ⚠️ Complexidade inicial maior
- ⚠️ Mais código boilerplate
- ⚠️ Curva de aprendizado para novos devs
- ⚠️ Overhead de abstrações

### Riscos Mitigados
- ✅ Vazamento de dados entre tenants
- ✅ Ataques de força bruta
- ✅ SQL injection
- ✅ Perda de dados (soft delete)
- ✅ Performance degradation (índices, pooling)
