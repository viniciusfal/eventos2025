#!/bin/bash

# Script de Deploy Automatizado para Produção
# Sistema de Check-in em Eventos
# Versão: 1.0

set -euo pipefail

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configurações
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
DOCKER_COMPOSE_FILE="docker-compose.production.yml"
ENV_FILE=".env.production"
BACKUP_DIR="/var/backups/eventos"
LOG_FILE="/var/log/eventos-deploy.log"

# Funções de log
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}" | tee -a "$LOG_FILE"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}" | tee -a "$LOG_FILE"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}" | tee -a "$LOG_FILE"
    exit 1
}

info() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}" | tee -a "$LOG_FILE"
}

# Verificar se está rodando como root ou com sudo
check_permissions() {
    if [[ $EUID -eq 0 ]]; then
        warn "Rodando como root. Considere usar um usuário específico para deploy."
    fi
}

# Verificar dependências
check_dependencies() {
    info "Verificando dependências..."
    
    command -v docker >/dev/null 2>&1 || error "Docker não está instalado"
    command -v docker-compose >/dev/null 2>&1 || error "Docker Compose não está instalado"
    command -v curl >/dev/null 2>&1 || error "curl não está instalado"
    
    # Verificar se Docker está rodando
    if ! docker info >/dev/null 2>&1; then
        error "Docker não está rodando"
    fi
    
    log "✅ Todas as dependências verificadas"
}

# Verificar arquivo de ambiente
check_env_file() {
    info "Verificando arquivo de ambiente..."
    
    if [[ ! -f "$PROJECT_DIR/$ENV_FILE" ]]; then
        error "Arquivo $ENV_FILE não encontrado. Crie o arquivo com as variáveis de produção."
    fi
    
    # Verificar variáveis obrigatórias
    required_vars=(
        "DB_PASSWORD"
        "REDIS_PASSWORD"
        "RABBITMQ_PASSWORD"
        "JWT_SECRET"
        "GRAFANA_PASSWORD"
    )
    
    source "$PROJECT_DIR/$ENV_FILE"
    
    for var in "${required_vars[@]}"; do
        if [[ -z "${!var:-}" ]]; then
            error "Variável $var não está definida no arquivo $ENV_FILE"
        fi
    done
    
    log "✅ Arquivo de ambiente verificado"
}

# Criar backup do banco de dados
create_backup() {
    info "Criando backup do banco de dados..."
    
    # Criar diretório de backup se não existir
    mkdir -p "$BACKUP_DIR"
    
    # Nome do backup com timestamp
    BACKUP_FILE="$BACKUP_DIR/eventos_backup_$(date +%Y%m%d_%H%M%S).sql"
    
    # Verificar se o container do PostgreSQL está rodando
    if docker ps | grep -q "eventos_postgres_prod"; then
        log "Fazendo backup do banco de dados..."
        
        docker exec eventos_postgres_prod pg_dump \
            -U "${DB_USER:-eventos_user}" \
            -d "${DB_NAME:-eventos_db}" \
            > "$BACKUP_FILE" || warn "Falha ao criar backup do banco"
        
        if [[ -f "$BACKUP_FILE" ]]; then
            log "✅ Backup criado: $BACKUP_FILE"
        else
            warn "⚠️  Backup não foi criado"
        fi
    else
        warn "⚠️  Container PostgreSQL não está rodando, pulando backup"
    fi
}

# Fazer pull da imagem mais recente
pull_latest_image() {
    info "Fazendo pull da imagem mais recente..."
    
    IMAGE_NAME="ghcr.io/viniciusfal/eventos2025:latest"
    
    docker pull "$IMAGE_NAME" || error "Falha ao fazer pull da imagem $IMAGE_NAME"
    
    log "✅ Imagem atualizada: $IMAGE_NAME"
}

# Deploy da aplicação
deploy_application() {
    info "Iniciando deploy da aplicação..."
    
    cd "$PROJECT_DIR"
    
    # Parar containers existentes
    log "Parando containers existentes..."
    docker-compose -f "$DOCKER_COMPOSE_FILE" --env-file "$ENV_FILE" down || warn "Falha ao parar containers"
    
    # Remover containers órfãos
    docker container prune -f || warn "Falha ao remover containers órfãos"
    
    # Subir aplicação
    log "Subindo aplicação..."
    docker-compose -f "$DOCKER_COMPOSE_FILE" --env-file "$ENV_FILE" up -d || error "Falha ao subir aplicação"
    
    log "✅ Aplicação deployada com sucesso"
}

# Verificar health da aplicação
check_health() {
    info "Verificando saúde da aplicação..."
    
    local max_attempts=30
    local attempt=1
    local health_url="http://localhost:8080/health"
    
    while [[ $attempt -le $max_attempts ]]; do
        info "Tentativa $attempt/$max_attempts - Verificando health endpoint..."
        
        if curl -f -s "$health_url" >/dev/null 2>&1; then
            log "✅ Aplicação está saudável!"
            return 0
        fi
        
        sleep 10
        ((attempt++))
    done
    
    error "❌ Aplicação não respondeu ao health check após $max_attempts tentativas"
}

# Executar smoke tests
run_smoke_tests() {
    info "Executando smoke tests..."
    
    local base_url="http://localhost:8080"
    local endpoints=(
        "/health"
        "/ready"
        "/live"
        "/metrics"
        "/ping"
    )
    
    for endpoint in "${endpoints[@]}"; do
        info "Testando endpoint: $endpoint"
        
        if curl -f -s "$base_url$endpoint" >/dev/null 2>&1; then
            log "✅ $endpoint - OK"
        else
            error "❌ $endpoint - FALHA"
        fi
    done
    
    log "✅ Todos os smoke tests passaram"
}

# Limpeza de imagens antigas
cleanup_old_images() {
    info "Limpando imagens Docker antigas..."
    
    # Remover imagens não utilizadas
    docker image prune -f || warn "Falha ao limpar imagens não utilizadas"
    
    # Remover volumes órfãos
    docker volume prune -f || warn "Falha ao limpar volumes órfãos"
    
    log "✅ Limpeza concluída"
}

# Mostrar status dos serviços
show_status() {
    info "Status dos serviços:"
    echo
    docker-compose -f "$PROJECT_DIR/$DOCKER_COMPOSE_FILE" --env-file "$PROJECT_DIR/$ENV_FILE" ps
    echo
    
    info "Logs recentes da aplicação:"
    docker-compose -f "$PROJECT_DIR/$DOCKER_COMPOSE_FILE" --env-file "$PROJECT_DIR/$ENV_FILE" logs --tail=10 api
}

# Rollback em caso de falha
rollback() {
    error "Deploy falhou! Iniciando rollback..."
    
    # Parar containers atuais
    docker-compose -f "$PROJECT_DIR/$DOCKER_COMPOSE_FILE" --env-file "$PROJECT_DIR/$ENV_FILE" down || true
    
    # Aqui você pode implementar lógica para voltar à versão anterior
    # Por exemplo, usar uma tag específica da imagem ou restaurar backup
    
    error "Rollback necessário. Verifique os logs e tente novamente."
}

# Função principal
main() {
    log "🚀 Iniciando deploy do Sistema de Check-in em Eventos"
    log "📅 Data: $(date)"
    log "👤 Usuário: $(whoami)"
    log "📂 Diretório: $PROJECT_DIR"
    
    # Trap para rollback em caso de erro
    trap rollback ERR
    
    check_permissions
    check_dependencies
    check_env_file
    create_backup
    pull_latest_image
    deploy_application
    check_health
    run_smoke_tests
    cleanup_old_images
    show_status
    
    log "🎉 Deploy concluído com sucesso!"
    log "🌐 Aplicação disponível em: http://localhost:8080"
    log "📊 Grafana disponível em: http://localhost:3000"
    log "📈 Prometheus disponível em: http://localhost:9090"
}

# Verificar argumentos
case "${1:-deploy}" in
    "deploy"|"")
        main
        ;;
    "rollback")
        rollback
        ;;
    "status")
        show_status
        ;;
    "backup")
        create_backup
        ;;
    "health")
        check_health
        ;;
    "cleanup")
        cleanup_old_images
        ;;
    *)
        echo "Uso: $0 [deploy|rollback|status|backup|health|cleanup]"
        echo
        echo "Comandos disponíveis:"
        echo "  deploy   - Deploy completo da aplicação (padrão)"
        echo "  rollback - Rollback da aplicação"
        echo "  status   - Mostrar status dos serviços"
        echo "  backup   - Criar backup do banco de dados"
        echo "  health   - Verificar saúde da aplicação"
        echo "  cleanup  - Limpar imagens e volumes não utilizados"
        exit 1
        ;;
esac