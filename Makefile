# Makefile para o projeto eventos-backend

# Variáveis
APP_NAME=eventos-backend
DOCKER_COMPOSE=docker-compose
GO_CMD=go
BINARY_NAME=main
BUILD_DIR=build

# Comandos principais
.PHONY: help setup build run test clean docker-up docker-down migrate-up migrate-down

help: ## Mostrar ajuda
	@echo "Comandos disponíveis:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

setup: ## Configurar ambiente de desenvolvimento
	@echo "Configurando ambiente..."
	$(GO_CMD) mod download
	$(GO_CMD) mod tidy
	@echo "Ambiente configurado!"

build: ## Compilar a aplicação
	@echo "Compilando aplicação..."
	mkdir -p $(BUILD_DIR)
	$(GO_CMD) build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/api
	@echo "Aplicação compilada em $(BUILD_DIR)/$(BINARY_NAME)"

run: build ## Executar a aplicação localmente
	@echo "Executando aplicação..."
	./$(BUILD_DIR)/$(BINARY_NAME)

test: ## Executar testes
	@echo "Executando testes..."
	$(GO_CMD) test -v ./...

test-coverage: ## Executar testes com cobertura
	@echo "Executando testes com cobertura..."
	$(GO_CMD) test -v -coverprofile=coverage.out ./...
	$(GO_CMD) tool cover -html=coverage.out -o coverage.html
	@echo "Relatório de cobertura gerado em coverage.html"

lint: ## Executar linter
	@echo "Executando linter..."
	golangci-lint run

format: ## Formatar código
	@echo "Formatando código..."
	$(GO_CMD) fmt ./...
	goimports -w .

clean: ## Limpar arquivos gerados
	@echo "Limpando arquivos..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	$(GO_CMD) clean

# Comandos Docker
docker-up: ## Subir ambiente Docker
	@echo "Subindo ambiente Docker..."
	$(DOCKER_COMPOSE) up -d

docker-down: ## Parar ambiente Docker
	@echo "Parando ambiente Docker..."
	$(DOCKER_COMPOSE) down

docker-logs: ## Ver logs do Docker
	$(DOCKER_COMPOSE) logs -f

docker-build: ## Build da imagem Docker da aplicação
	@echo "Fazendo build da imagem Docker..."
	$(DOCKER_COMPOSE) build api

docker-run-app: ## Executar aplicação com Docker
	@echo "Executando aplicação com Docker..."
	$(DOCKER_COMPOSE) --profile app up -d

# Comandos de migração
migrate-up: ## Executar migrações do banco de dados
	@echo "Executando migrações..."
	@if [ ! -f ./scripts/migrate.sh ]; then \
		echo "Script de migração não encontrado. Criando..."; \
		mkdir -p scripts; \
		echo '#!/bin/bash' > ./scripts/migrate.sh; \
		echo 'PGPASSWORD=eventos_password psql -h localhost -U eventos_user -d eventos_db -f migrations/001_create_database_schema.sql' >> ./scripts/migrate.sh; \
		chmod +x ./scripts/migrate.sh; \
	fi
	./scripts/migrate.sh

migrate-down: ## Reverter migrações (cuidado!)
	@echo "ATENÇÃO: Isso irá apagar todos os dados!"
	@read -p "Tem certeza? [y/N] " confirm && [ "$$confirm" = "y" ]
	@echo "Revertendo migrações..."
	PGPASSWORD=eventos_password psql -h localhost -U eventos_user -d eventos_db -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"

# Comandos de desenvolvimento
dev-setup: docker-up migrate-up ## Configurar ambiente completo de desenvolvimento
	@echo "Ambiente de desenvolvimento configurado!"

dev-reset: docker-down clean docker-up migrate-up ## Resetar ambiente de desenvolvimento
	@echo "Ambiente resetado!"

# Comandos de qualidade
quality: format lint test ## Executar verificações de qualidade

# Comandos de produção
prod-build: ## Build para produção
	@echo "Build para produção..."
	CGO_ENABLED=0 GOOS=linux $(GO_CMD) build -a -installsuffix cgo -ldflags '-w -s' -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/api

# Comandos de monitoramento
monitor-up: ## Subir apenas serviços de monitoramento
	$(DOCKER_COMPOSE) up -d prometheus grafana

monitor-down: ## Parar serviços de monitoramento
	$(DOCKER_COMPOSE) stop prometheus grafana

# Comandos de banco de dados
db-connect: ## Conectar ao banco de dados
	PGPASSWORD=eventos_password psql -h localhost -U eventos_user -d eventos_db

db-backup: ## Fazer backup do banco de dados
	@echo "Fazendo backup do banco..."
	PGPASSWORD=eventos_password pg_dump -h localhost -U eventos_user eventos_db > backup_$(shell date +%Y%m%d_%H%M%S).sql

# Comandos de cache
redis-cli: ## Conectar ao Redis CLI
	docker exec -it eventos_redis redis-cli

redis-flush: ## Limpar cache Redis
	docker exec -it eventos_redis redis-cli FLUSHALL

# Comandos de mensageria
rabbitmq-management: ## Abrir interface de gerenciamento do RabbitMQ
	@echo "Interface do RabbitMQ disponível em: http://localhost:15672"
	@echo "Usuário: eventos_user"
	@echo "Senha: eventos_password"
