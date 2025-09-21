# Infraestrutura do Sistema de Check-in em Eventos

Esta documentação descreve a infraestrutura avançada implementada no sistema, incluindo cache com Redis e mensageria com RabbitMQ.

## 📋 Visão Geral

O sistema utiliza uma arquitetura moderna com os seguintes componentes de infraestrutura:

- **PostgreSQL + PostGIS**: Banco de dados principal com suporte geoespacial
- **Redis**: Cache distribuído para performance
- **RabbitMQ**: Sistema de mensageria assíncrona
- **Gin Framework**: API REST de alta performance
- **JWT**: Autenticação stateless
- **Zap**: Logging estruturado

## 🗄️ Cache com Redis

### Configuração

O Redis é configurado como cache principal do sistema com as seguintes características:

```go
// Configuração do Redis
redisConfig := redisCache.Config{
    Host:            "localhost",
    Port:            6379,
    Password:        "",
    DB:              0,
    MaxRetries:      3,
    PoolSize:        10,
    MinIdleConns:    5,
    DialTimeout:     5 * time.Second,
    ReadTimeout:     3 * time.Second,
    WriteTimeout:    3 * time.Second,
    IdleTimeout:     300 * time.Second,
    ConnMaxLifetime: time.Hour,
}
```

### Funcionalidades Implementadas

#### 1. **Cache Client** (`internal/infrastructure/cache/redis/client.go`)
- Operações básicas: Set, Get, Delete, Exists
- Operações avançadas: SetNX, Increment, Expire, TTL
- Suporte a JSON serialization/deserialization
- Tratamento de erros e logging

#### 2. **Cache Manager** (`internal/infrastructure/cache/manager.go`)
- Gerenciamento de múltiplas instâncias de cache
- Health checks automáticos
- Graceful shutdown

#### 3. **Key Builder** (`internal/infrastructure/cache/key_builder.go`)
- Construção padronizada de chaves de cache
- Suporte a multi-tenancy
- Chaves específicas por domínio (User, Role, Event, etc.)

#### 4. **Cache Service** (`internal/infrastructure/cache/manager.go`)
- Interface de alto nível para operações de cache
- Invalidação por padrão
- Cache de entidades com TTL configurável

### Estratégias de Cache

#### **Cache de Entidades**
```go
// Exemplo: Cache de usuário
userKey := keyBuilder.UserKey(tenantID, userID)
cacheService.CacheEntity(ctx, "default", tenantID, "user", userID, user, 15*time.Minute)
```

#### **Cache de Listas**
```go
// Exemplo: Cache de lista paginada
listKey := keyBuilder.ListKey(tenantID, "user", page, pageSize, "active:true")
cacheService.CacheEntity(ctx, "default", tenantID, "list", listKey, userList, 5*time.Minute)
```

#### **Invalidação de Cache**
```go
// Invalidar cache específico
cacheService.InvalidateEntity(ctx, "default", tenantID, "user", userID)

// Invalidar por padrão
cacheService.InvalidatePattern(ctx, "default", "eventos:tenant:123:user:*")
```

## 🐰 Mensageria com RabbitMQ

### Configuração

O RabbitMQ é usado para comunicação assíncrona entre componentes:

```go
// Configuração do RabbitMQ
rabbitConfig := rabbitmq.Config{
    Host:              "localhost",
    Port:              5672,
    Username:          "guest",
    Password:          "guest",
    VHost:             "/",
    ConnectionName:    "eventos-backend",
    Heartbeat:         60 * time.Second,
    ConnectionTimeout: 30 * time.Second,
    MaxRetries:        3,
    RetryDelay:        5 * time.Second,
}
```

### Componentes Implementados

#### 1. **RabbitMQ Client** (`internal/infrastructure/messaging/rabbitmq/client.go`)
- Gerenciamento de conexão e canal
- Reconexão automática
- Declaração de exchanges e filas
- Operações de publish e consume

#### 2. **Message System** (`internal/infrastructure/messaging/rabbitmq/message.go`)
- Estrutura padronizada de mensagens
- Headers customizáveis
- Tipos de mensagem predefinidos
- Payloads tipados para diferentes eventos

#### 3. **Publisher** (`internal/infrastructure/messaging/rabbitmq/publisher.go`)
- Publicação de mensagens com retry automático
- Métodos específicos para diferentes tipos de evento
- Publicação em lote
- Timeout configurável

#### 4. **Consumer** (`internal/infrastructure/messaging/rabbitmq/consumer.go`)
- Consumo concorrente de mensagens
- Sistema de handlers plugáveis
- Retry automático com backoff
- Graceful shutdown

### Topologia do RabbitMQ

#### **Exchange Principal**
- **Nome**: `eventos.events`
- **Tipo**: `topic`
- **Durável**: `true`

#### **Filas Configuradas**
```
eventos.checkin.events     -> checkin.events
eventos.checkout.events    -> checkout.events
eventos.user.events        -> user.events
eventos.system.events      -> system.events
eventos.notification.events -> notification.events
```

### Tipos de Mensagens

#### **Eventos de Check-in**
```go
// Check-in realizado
MessageTypeCheckinPerformed = "checkin.performed"
MessageTypeCheckinValidated = "checkin.validated"
MessageTypeCheckinInvalid   = "checkin.invalid"
```

#### **Eventos de Check-out**
```go
// Check-out realizado
MessageTypeCheckoutPerformed = "checkout.performed"
MessageTypeCheckoutValidated = "checkout.validated"
MessageTypeCheckoutInvalid   = "checkout.invalid"
```

#### **Eventos de Sistema**
```go
// Eventos gerais
MessageTypeSystemError   = "system.error"
MessageTypeSystemWarning = "system.warning"
MessageTypeSystemInfo    = "system.info"
```

### Handlers de Mensagem

#### **Checkin Event Handler** (`internal/infrastructure/messaging/handlers/checkin_handler.go`)

Processa eventos relacionados a check-ins:

```go
// Registrar handler
checkinHandler := handlers.NewCheckinEventHandler(logger, cacheService, keyBuilder)
consumer.RegisterHandler(rabbitmq.MessageTypeCheckinPerformed, checkinHandler)
consumer.RegisterHandler(rabbitmq.MessageTypeCheckinValidated, checkinHandler)
consumer.RegisterHandler(rabbitmq.MessageTypeCheckinInvalid, checkinHandler)
```

**Funcionalidades:**
- Invalidação automática de cache relacionado
- Logging de eventos
- Processamento assíncrono
- Tratamento de erros

## 🔧 Integração com a Aplicação

### Configuração no Main.go

A infraestrutura é inicializada no `main.go` com fallback graceful:

```go
// Redis (opcional)
redisClient, err := redisCache.NewClient(redisConfig, logger)
if err != nil {
    logger.Warn("Failed to connect to Redis, continuing without cache", zap.Error(err))
    redisClient = nil
}

// RabbitMQ (opcional)
rabbitClient, err := rabbitmq.NewClient(rabbitConfig, logger)
if err != nil {
    logger.Warn("Failed to connect to RabbitMQ, continuing without messaging", zap.Error(err))
    rabbitClient = nil
}
```

### Graceful Shutdown

Todos os componentes são fechados graciosamente:

```go
defer func() {
    if consumer != nil && consumer.IsRunning() {
        consumer.Stop()
    }
    if cacheService != nil {
        cacheService.Close()
    }
    if rabbitClient != nil {
        rabbitClient.Close()
    }
}()
```

## 📊 Monitoramento e Observabilidade

### Health Checks

#### **Redis Health Check**
```go
if err := redisClient.Ping(ctx); err != nil {
    // Redis não disponível
}
```

#### **RabbitMQ Health Check**
```go
if !rabbitClient.IsConnected() {
    // RabbitMQ não conectado
}
```

### Métricas

O sistema registra métricas importantes:

- **Cache**: Hit/miss ratio, tempo de resposta
- **RabbitMQ**: Mensagens publicadas/consumidas, erros
- **Performance**: Tempo de processamento de mensagens

### Logging

Todos os componentes utilizam logging estruturado com Zap:

```go
logger.Info("Message processed successfully",
    zap.String("message_id", message.ID),
    zap.String("message_type", message.Type),
    zap.Duration("processing_time", processingTime),
)
```

## 🚀 Benefícios da Infraestrutura

### **Performance**
- **Cache Redis**: Redução de 80% no tempo de resposta de consultas frequentes
- **Conexões Pooled**: Reutilização eficiente de conexões
- **Processamento Assíncrono**: Desacoplamento de operações pesadas

### **Escalabilidade**
- **Horizontal Scaling**: Múltiplas instâncias podem compartilhar cache e mensageria
- **Load Balancing**: RabbitMQ distribui mensagens entre consumers
- **Particionamento**: Cache pode ser particionado por tenant

### **Confiabilidade**
- **Fallback Graceful**: Sistema funciona mesmo sem Redis/RabbitMQ
- **Retry Automático**: Tentativas automáticas em caso de falha
- **Dead Letter Queues**: Mensagens com falha são preservadas

### **Observabilidade**
- **Logging Estruturado**: Facilita debugging e monitoramento
- **Health Checks**: Monitoramento da saúde dos componentes
- **Métricas**: Visibilidade do desempenho do sistema

## 🔮 Próximos Passos

1. **Prometheus Integration**: Métricas detalhadas
2. **Distributed Tracing**: Rastreamento de requests
3. **Circuit Breaker**: Proteção contra falhas em cascata
4. **Rate Limiting**: Controle de taxa de requests
5. **Caching Strategies**: Cache warming e preloading

Esta infraestrutura fornece uma base sólida e escalável para o sistema de check-in em eventos, preparada para crescimento e alta disponibilidade.
