# 📁 Diagrama de Estrutura de Pastas para o Backend em Go

## 🏗️ Estrutura Geral do Projeto

```
eventos-backend/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── domain/
│   │   ├── tenant/
│   │   │   ├── tenant.go
│   │   │   ├── repository.go
│   │   │   └── service.go
│   │   ├── user/
│   │   │   ├── user.go
│   │   │   ├── repository.go
│   │   │   └── service.go
│   │   ├── event/
│   │   │   ├── event.go
│   │   │   ├── repository.go
│   │   │   └── service.go
│   │   ├── partner/
│   │   │   ├── partner.go
│   │   │   ├── repository.go
│   │   │   └── service.go
│   │   ├── employee/
│   │   │   ├── employee.go
│   │   │   ├── repository.go
│   │   │   └── service.go
│   │   ├── checkin/
│   │   │   ├── checkin.go
│   │   │   ├── repository.go
│   │   │   └── service.go
│   │   ├── checkout/
│   │   │   ├── checkout.go
│   │   │   ├── repository.go
│   │   │   └── service.go
│   │   ├── role/
│   │   │   ├── role.go
│   │   │   ├── repository.go
│   │   │   └── service.go
│   │   ├── permission/
│   │   │   ├── permission.go
│   │   │   ├── repository.go
│   │   │   └── service.go
│   │   ├── module/
│   │   │   ├── module.go
│   │   │   ├── repository.go
│   │   │   └── service.go
│   │   ├── qr_code/
│   │   │   ├── qr_code.go
│   │   │   ├── repository.go
│   │   │   └── service.go
│   │   ├── log/
│   │   │   ├── event_log.go
│   │   │   ├── audit_log.go
│   │   │   ├── repository.go
│   │   │   └── service.go
│   │   └── shared/
│   │       ├── value_objects/
│   │       ├── errors/
│   │       └── constants/
│   ├── application/
│   │   ├── usecases/
│   │   │   ├── tenant/
│   │   │   ├── user/
│   │   │   ├── event/
│   │   │   ├── partner/
│   │   │   ├── employee/
│   │   │   ├── checkin/
│   │   │   ├── checkout/
│   │   │   ├── role/
│   │   │   ├── permission/
│   │   │   ├── module/
│   │   │   ├── qr_code/
│   │   │   └── log/
│   │   ├── dto/
│   │   │   ├── requests/
│   │   │   └── responses/
│   │   └── mappers/
│   ├── interfaces/
│   │   ├── http/
│   │   │   ├── handlers/
│   │   │   ├── middleware/
│   │   │   ├── validators/
│   │   │   └── presenters/
│   │   └── grpc/
│   ├── infrastructure/
│   │   ├── persistence/
│   │   │   ├── postgres/
│   │   │   │   ├── repositories/
│   │   │   │   └── connection.go
│   │   │   └── redis/
│   │   ├── messaging/
│   │   │   └── rabbitmq/
│   │   ├── cache/
│   │   │   └── redis/
│   │   ├── logging/
│   │   │   └── zap/
│   │   ├── monitoring/
│   │   │   ├── prometheus/
│   │   │   └── opentelemetry/
│   │   ├── auth/
│   │   │   └── jwt/
│   │   ├── qr/
│   │   │   └── generator/
│   │   ├── geolocation/
│   │   │   └── postgis/
│   │   └── config/
│   └── shared/
│       ├── utils/
│       ├── exceptions/
│       └── validation/
├── pkg/
│   ├── httpclient/
│   ├── logger/
│   └── config/
├── configs/
│   ├── app.yaml
│   ├── database.yaml
│   ├── redis.yaml
│   ├── rabbitmq.yaml
│   └── jwt.yaml
├── docs/
│   ├── api/
│   └── swagger/
├── migrations/
│   └── 001_create_database_schema.sql
├── scripts/
│   ├── start.sh
│   ├── build.sh
│   └── test.sh
├── tests/
│   ├── integration/
│   ├── e2e/
│   └── fixtures/
├── go.mod
├── go.sum
├── Dockerfile
├── docker-compose.yml
├── .dockerignore
├── .gitignore
├── README.md
└── Makefile
```

## 📖 Explicação da Estrutura

### 📁 `cmd/`
Ponto de entrada da aplicação. Cada subpasta representa um executável diferente.
- `api/` - Servidor HTTP principal

### 📁 `internal/`
Código principal da aplicação, seguindo a Clean Architecture:

#### 📁 `domain/`
Camada de domínio contendo as entidades de negócio e regras fundamentais:
- Cada domínio tem sua própria pasta (tenant, user, event, etc.)
- Cada domínio contém:
  - Entidade principal (`.go`)
  - Interface de repositório (`repository.go`)
  - Serviços de domínio (`service.go`)

#### 📁 `application/`
Camada de aplicação contendo os casos de uso:
- `usecases/` - Implementação dos casos de uso organizados por domínio
- `dto/` - Data Transfer Objects para entrada e saída de dados
- `mappers/` - Conversores entre entidades e DTOs

#### 📁 `interfaces/`
Camada de interfaces com o mundo externo:
- `http/` - Handlers HTTP, middleware, validadores
- `grpc/` - Interfaces gRPC (se necessário no futuro)

#### 📁 `infrastructure/`
Implementações concretas das interfaces definidas nas camadas internas:
- `persistence/` - Implementações de repositórios com PostgreSQL
- `messaging/` - Integração com RabbitMQ
- `cache/` - Implementações de cache com Redis
- `logging/` - Configuração e implementação do Zap
- `monitoring/` - Integração com Prometheus e OpenTelemetry
- `auth/` - Implementação de autenticação JWT
- `qr/` - Geração de QR Codes
- `geolocation/` - Integração com PostGIS
- `config/` - Carregamento de configurações

#### 📁 `shared/`
Componentes compartilhados entre as camadas internas:
- `utils/` - Funções utilitárias
- `exceptions/` - Erros customizados
- `validation/` - Funções de validação compartilhadas

### 📁 `pkg/`
Bibliotecas compartilhadas que podem ser usadas por outros projetos:
- Componentes reutilizáveis e bem definidos

### 📁 `configs/`
Arquivos de configuração da aplicação:
- Configurações por ambiente e componente

### 📁 `docs/`
Documentação do projeto:
- Documentação da API
- Arquivos Swagger/OpenAPI

### 📁 `migrations/`
Scripts de migração do banco de dados:
- Versionamento do schema do banco de dados

### 📁 `scripts/`
Scripts auxiliares para desenvolvimento:
- Scripts de inicialização, build e testes

### 📁 `tests/`
Testes da aplicação:
- `integration/` - Testes de integração
- `e2e/` - Testes end-to-end
- `fixtures/` - Dados de teste

### 📄 Arquivos Raiz
- `go.mod/go.sum` - Gerenciamento de dependências Go
- `Dockerfile/docker-compose.yml` - Configuração de containers
- `Makefile` - Comandos automatizados
- `README.md` - Documentação principal

## ✅ Benefícios da Estrutura

1. **Separação Clara de Responsabilidades**: Cada camada tem um propósito bem definido
2. **Facilidade de Manutenção**: Código bem organizado e fácil de encontrar
3. **Escalabilidade**: Estrutura suporta crescimento do projeto
4. **Testabilidade**: Fácil escrever testes unitários e de integração
5. **Independência de Frameworks**: Core da aplicação não depende de frameworks específicos
6. **Flexibilidade**: Fácil substituir componentes de infraestrutura