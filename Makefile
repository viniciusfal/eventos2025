# Makefile para o projeto eventos-backend
# Versão: 2.0 - Com suporte a produção e deploy automatizado

# Variáveis
APP_NAME=eventos-backend
DOCKER_COMPOSE=docker-compose
DOCKER_COMPOSE_PROD=docker-compose -f docker-compose.production.yml
GO_CMD=go
BINARY_NAME=main
BUILD_DIR=build
IMAGE_NAME=ghcr.io/viniciusfal/eventos2025
VERSION=$(shell git rev-parse --short HEAD)

# Comandos principais
.PHONY: help setup build run test clean docker-up docker-down migrate-up migrate-down

help: ## Mostrar ajuda
	@echo "🚀 Sistema de Check-in em Eventos - Comandos Disponíveis"
	@echo "=========================================================="
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}'
	@echo
	@echo "📋 Exemplos de uso:"
	@echo "  make dev-setup     # Configurar ambiente completo"
	@echo "  make prod-deploy   # Deploy para produção"
	@echo "  make backup        # Fazer backup do sistema"

# ========================================
# COMANDOS DE DESENVOLVIMENTO
# ========================================

setup: ## Configurar ambiente de desenvolvimento
	@echo "🔧 Configurando ambiente..."
	$(GO_CMD) mod download
	$(GO_CMD) mod tidy
	@echo "✅ Ambiente configurado!"

build: ## Compilar a aplicação
	@echo "🔨 Compilando aplicação..."
	mkdir -p $(BUILD_DIR)
	$(GO_CMD) build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/api
	@echo "✅ Aplicação compilada em $(BUILD_DIR)/$(BINARY_NAME)"

build-prod: ## Compilar para produção
	@echo "🔨 Compilando para produção..."
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO_CMD) build \
		-a -installsuffix cgo \
		-ldflags '-w -s -X main.version=$(VERSION)' \
		-o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/api
	@echo "✅ Build de produção criado: $(BUILD_DIR)/$(BINARY_NAME)"

run: build ## Executar a aplicação localmente
	@echo "🚀 Executando aplicação..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# ========================================
# COMANDOS DE TESTE
# ========================================

test: ## Executar testes
	@echo "🧪 Executando testes..."
	$(GO_CMD) test -v ./...

test-coverage: ## Executar testes com cobertura
	@echo "📊 Executando testes com cobertura..."
	$(GO_CMD) test -v -coverprofile=coverage.out ./...
	$(GO_CMD) tool cover -html=coverage.out -o coverage.html
	@echo "✅ Relatório de cobertura: coverage.html"

test-race: ## Executar testes com detecção de race conditions
	@echo "🏃 Executando testes com race detection..."
	$(GO_CMD) test -race -v ./...

benchmark: ## Executar benchmarks
	@echo "⚡ Executando benchmarks..."
	$(GO_CMD) test -bench=. -benchmem ./...

# ========================================
# COMANDOS DE QUALIDADE
# ========================================

lint: ## Executar linter
	@echo "🔍 Executando linter..."
	golangci-lint run

format: ## Formatar código
	@echo "✨ Formatando código..."
	$(GO_CMD) fmt ./...
	goimports -w .

vet: ## Executar go vet
	@echo "🔎 Executando go vet..."
	$(GO_CMD) vet ./...

quality: format lint vet test ## Executar verificações de qualidade completas

# ========================================
# COMANDOS DOCKER - DESENVOLVIMENTO
# ========================================

docker-up: ## Subir ambiente Docker para desenvolvimento
	@echo "🐳 Subindo ambiente Docker..."
	$(DOCKER_COMPOSE) up -d

docker-down: ## Parar ambiente Docker
	@echo "🛑 Parando ambiente Docker..."
	$(DOCKER_COMPOSE) down

docker-logs: ## Ver logs do Docker
	$(DOCKER_COMPOSE) logs -f

docker-build: ## Build da imagem Docker da aplicação
	@echo "🏗️  Fazendo build da imagem Docker..."
	$(DOCKER_COMPOSE) build api

docker-run-app: ## Executar aplicação com Docker
	@echo "🚀 Executando aplicação com Docker..."
	$(DOCKER_COMPOSE) --profile app up -d

docker-restart: ## Reiniciar containers Docker
	@echo "🔄 Reiniciando containers..."
	$(DOCKER_COMPOSE) restart

# ========================================
# COMANDOS DE MIGRAÇÃO
# ========================================

migrate-up: ## Executar migrações do banco de dados
	@echo "📊 Executando migrações..."
	@if [ ! -f ./scripts/migrate.sh ]; then \
		echo "Criando script de migração..."; \
		mkdir -p scripts; \
		echo '#!/bin/bash' > ./scripts/migrate.sh; \
		echo 'PGPASSWORD=eventos_password psql -h localhost -U eventos_user -d eventos_db -f migrations/001_create_database_schema.sql' >> ./scripts/migrate.sh; \
		chmod +x ./scripts/migrate.sh; \
	fi
	./scripts/migrate.sh

migrate-down: ## Reverter migrações (cuidado!)
	@echo "⚠️  ATENÇÃO: Isso irá apagar todos os dados!"
	@read -p "Tem certeza? [y/N] " confirm && [ "$$confirm" = "y" ]
	@echo "🗑️  Revertendo migrações..."
	PGPASSWORD=eventos_password psql -h localhost -U eventos_user -d eventos_db -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"

# ========================================
# COMANDOS DE PRODUÇÃO
# ========================================

prod-build-image: ## Build da imagem Docker para produção
	@echo "🏭 Fazendo build da imagem de produção..."
	docker build -f Dockerfile.production -t $(IMAGE_NAME):$(VERSION) -t $(IMAGE_NAME):latest .

prod-push-image: prod-build-image ## Push da imagem para registry
	@echo "📤 Fazendo push da imagem..."
	docker push $(IMAGE_NAME):$(VERSION)
	docker push $(IMAGE_NAME):latest

prod-deploy: ## Deploy para produção
	@echo "🚀 Iniciando deploy para produção..."
	@if [ ! -f .env.production ]; then \
		echo "❌ Arquivo .env.production não encontrado!"; \
		echo "Copie .env.production.example e configure as variáveis."; \
		exit 1; \
	fi
	chmod +x scripts/deploy.sh
	./scripts/deploy.sh

prod-up: ## Subir ambiente de produção
	@echo "🏭 Subindo ambiente de produção..."
	$(DOCKER_COMPOSE_PROD) --env-file .env.production up -d

prod-down: ## Parar ambiente de produção
	@echo "🛑 Parando ambiente de produção..."
	$(DOCKER_COMPOSE_PROD) --env-file .env.production down

prod-logs: ## Ver logs de produção
	$(DOCKER_COMPOSE_PROD) --env-file .env.production logs -f

prod-status: ## Verificar status da produção
	@echo "📊 Status do ambiente de produção:"
	$(DOCKER_COMPOSE_PROD) --env-file .env.production ps
	@echo
	@echo "🏥 Health check:"
	@curl -s http://localhost:8080/health | jq '.' || echo "Health check falhou"

prod-restart: ## Reiniciar ambiente de produção
	@echo "🔄 Reiniciando produção..."
	$(DOCKER_COMPOSE_PROD) --env-file .env.production restart

# ========================================
# COMANDOS DE BACKUP
# ========================================

backup: ## Fazer backup completo do sistema
	@echo "💾 Iniciando backup..."
	chmod +x scripts/backup.sh
	./scripts/backup.sh

backup-verify: ## Verificar integridade dos backups
	@echo "🔍 Verificando backups..."
	./scripts/backup.sh verify

backup-report: ## Gerar relatório de backups
	@echo "📋 Gerando relatório..."
	./scripts/backup.sh report

backup-setup-cron: ## Configurar backup automatizado
	@echo "⏰ Configurando backup automático..."
	chmod +x scripts/setup-cron.sh
	sudo ./scripts/setup-cron.sh

# ========================================
# COMANDOS DE MONITORAMENTO
# ========================================

monitor-up: ## Subir apenas serviços de monitoramento
	$(DOCKER_COMPOSE) up -d prometheus grafana

monitor-down: ## Parar serviços de monitoramento
	$(DOCKER_COMPOSE) stop prometheus grafana

health-check: ## Verificar saúde da aplicação
	@echo "🏥 Verificando saúde da aplicação..."
	@curl -s http://localhost:8080/health | jq '.' || echo "❌ Aplicação não está respondendo"
	@curl -s http://localhost:8080/ready | jq '.' || echo "❌ Aplicação não está pronta"
	@curl -s http://localhost:8080/live | jq '.' || echo "❌ Aplicação não está viva"

metrics: ## Mostrar métricas básicas
	@echo "📈 Métricas da aplicação:"
	@curl -s http://localhost:8080/metrics | grep -E "(http_requests|db_queries|cache_)" | head -10

# ========================================
# COMANDOS DE BANCO DE DADOS
# ========================================

db-connect: ## Conectar ao banco de dados
	PGPASSWORD=eventos_password psql -h localhost -U eventos_user -d eventos_db

db-backup: ## Fazer backup do banco de dados
	@echo "💾 Fazendo backup do banco..."
	PGPASSWORD=eventos_password pg_dump -h localhost -U eventos_user eventos_db > backup_$(shell date +%Y%m%d_%H%M%S).sql

db-restore: ## Restaurar backup do banco (especificar arquivo)
	@echo "📥 Restaurando backup..."
	@if [ -z "$(FILE)" ]; then echo "Use: make db-restore FILE=backup.sql"; exit 1; fi
	PGPASSWORD=eventos_password psql -h localhost -U eventos_user -d eventos_db < $(FILE)

# ========================================
# COMANDOS DE CACHE E MENSAGERIA
# ========================================

redis-cli: ## Conectar ao Redis CLI
	docker exec -it eventos_redis redis-cli

redis-flush: ## Limpar cache Redis
	docker exec -it eventos_redis redis-cli FLUSHALL

rabbitmq-management: ## Abrir interface de gerenciamento do RabbitMQ
	@echo "🐰 Interface do RabbitMQ: http://localhost:15672"
	@echo "👤 Usuário: eventos_user"
	@echo "🔑 Senha: eventos_password"

# ========================================
# COMANDOS DE LIMPEZA
# ========================================

clean: ## Limpar arquivos gerados
	@echo "🧹 Limpando arquivos..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	$(GO_CMD) clean

clean-docker: ## Limpar recursos Docker não utilizados
	@echo "🧹 Limpando Docker..."
	docker system prune -f
	docker volume prune -f

clean-all: clean clean-docker ## Limpeza completa

# ========================================
# COMANDOS DE DESENVOLVIMENTO RÁPIDO
# ========================================

dev-setup: docker-up migrate-up ## Configurar ambiente completo de desenvolvimento
	@echo "🎉 Ambiente de desenvolvimento configurado!"
	@echo "🌐 Aplicação: http://localhost:8080"
	@echo "📊 Grafana: http://localhost:3000"
	@echo "📈 Prometheus: http://localhost:9090"

dev-reset: docker-down clean docker-up migrate-up ## Resetar ambiente de desenvolvimento
	@echo "🔄 Ambiente resetado!"

dev-test: docker-up test ## Executar testes com ambiente Docker

# ========================================
# COMANDOS DE CI/CD
# ========================================

ci-setup: setup ## Configurar para CI/CD
	@echo "🤖 Configurando para CI/CD..."

ci-test: test-coverage lint ## Executar testes para CI/CD
	@echo "✅ Testes CI/CD concluídos"

ci-build: build-prod ## Build para CI/CD
	@echo "✅ Build CI/CD concluído"

# ========================================
# COMANDOS DE SEGURANÇA
# ========================================

security-scan: ## Executar scan de segurança
	@echo "🔒 Executando scan de segurança..."
	@command -v gosec >/dev/null 2>&1 || { echo "Instalando gosec..."; go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; }
	gosec ./...

vulnerability-check: ## Verificar vulnerabilidades nas dependências
	@echo "🛡️  Verificando vulnerabilidades..."
	$(GO_CMD) list -json -m all | nancy sleuth

# ========================================
# COMANDOS DE INFORMAÇÃO
# ========================================

info: ## Mostrar informações do sistema
	@echo "📋 Informações do Sistema de Check-in em Eventos"
	@echo "================================================="
	@echo "🏷️  Versão: $(VERSION)"
	@echo "📂 Diretório: $(PWD)"
	@echo "🐹 Go Version: $(shell $(GO_CMD) version)"
	@echo "🐳 Docker: $(shell docker --version 2>/dev/null || echo 'Não instalado')"
	@echo "📦 Docker Compose: $(shell docker-compose --version 2>/dev/null || echo 'Não instalado')"
	@echo
	@echo "🔗 Links Úteis:"
	@echo "   📖 Documentação: docs/DEPLOY.md"
	@echo "   🐙 Repositório: https://github.com/viniciusfal/eventos2025"
	@echo "   🌐 API Local: http://localhost:8080"
	@echo "   📊 Swagger: http://localhost:8080/swagger/index.html"

status: ## Mostrar status completo do sistema
	@echo "📊 Status do Sistema"
	@echo "==================="
	@echo
	@echo "🐳 Containers Docker:"
	@docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep eventos || echo "Nenhum container encontrado"
	@echo
	@echo "💾 Uso de disco:"
	@df -h . | tail -1
	@echo
	@echo "🏥 Health checks:"
	@make health-check 2>/dev/null || echo "Aplicação não está rodando"

version: ## Mostrar versão
	@echo "$(VERSION)"