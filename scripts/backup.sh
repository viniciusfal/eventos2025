#!/bin/bash

# Script de Backup Automatizado
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
BACKUP_DIR="/var/backups/eventos"
LOG_FILE="/var/log/eventos-backup.log"
RETENTION_DAYS=${BACKUP_RETENTION_DAYS:-30}
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Configurações AWS S3 (opcional)
AWS_BACKUP_ENABLED=${AWS_BACKUP_ENABLED:-false}
S3_BUCKET=${BACKUP_S3_BUCKET:-""}

# Configurações de notificação (opcional)
SLACK_WEBHOOK=${SLACK_WEBHOOK_URL:-""}
DISCORD_WEBHOOK=${DISCORD_WEBHOOK_URL:-""}

# Funções de log
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}" | tee -a "$LOG_FILE"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}" | tee -a "$LOG_FILE"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}" | tee -a "$LOG_FILE"
    send_notification "❌ BACKUP FALHOU: $1" "error"
    exit 1
}

info() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}" | tee -a "$LOG_FILE"
}

# Função para enviar notificações
send_notification() {
    local message="$1"
    local type="${2:-info}"
    
    # Slack notification
    if [[ -n "$SLACK_WEBHOOK" ]]; then
        local color="good"
        [[ "$type" == "error" ]] && color="danger"
        [[ "$type" == "warning" ]] && color="warning"
        
        curl -X POST -H 'Content-type: application/json' \
            --data "{\"attachments\":[{\"color\":\"$color\",\"text\":\"$message\"}]}" \
            "$SLACK_WEBHOOK" >/dev/null 2>&1 || true
    fi
    
    # Discord notification
    if [[ -n "$DISCORD_WEBHOOK" ]]; then
        local embed_color=65280  # Green
        [[ "$type" == "error" ]] && embed_color=16711680    # Red
        [[ "$type" == "warning" ]] && embed_color=16776960  # Yellow
        
        curl -X POST -H 'Content-type: application/json' \
            --data "{\"embeds\":[{\"title\":\"Eventos Backup\",\"description\":\"$message\",\"color\":$embed_color}]}" \
            "$DISCORD_WEBHOOK" >/dev/null 2>&1 || true
    fi
}

# Verificar dependências
check_dependencies() {
    info "Verificando dependências para backup..."
    
    command -v docker >/dev/null 2>&1 || error "Docker não está instalado"
    command -v gzip >/dev/null 2>&1 || error "gzip não está instalado"
    
    # Verificar AWS CLI se backup S3 estiver habilitado
    if [[ "$AWS_BACKUP_ENABLED" == "true" ]]; then
        command -v aws >/dev/null 2>&1 || error "AWS CLI não está instalado"
        [[ -n "$S3_BUCKET" ]] || error "S3_BUCKET não está configurado"
    fi
    
    log "✅ Dependências verificadas"
}

# Criar diretórios necessários
setup_directories() {
    info "Configurando diretórios de backup..."
    
    mkdir -p "$BACKUP_DIR/database"
    mkdir -p "$BACKUP_DIR/redis"
    mkdir -p "$BACKUP_DIR/rabbitmq"
    mkdir -p "$BACKUP_DIR/configs"
    mkdir -p "$BACKUP_DIR/logs"
    
    log "✅ Diretórios configurados"
}

# Backup do banco de dados PostgreSQL
backup_database() {
    info "Iniciando backup do banco de dados..."
    
    local backup_file="$BACKUP_DIR/database/postgres_backup_$TIMESTAMP.sql"
    local compressed_file="$backup_file.gz"
    
    # Verificar se o container está rodando
    if ! docker ps | grep -q "eventos_postgres_prod"; then
        error "Container PostgreSQL não está rodando"
    fi
    
    # Fazer backup
    log "Exportando dados do PostgreSQL..."
    docker exec eventos_postgres_prod pg_dump \
        -U "${DB_USER:-eventos_user}" \
        -d "${DB_NAME:-eventos_db}" \
        --verbose \
        --no-password \
        --format=custom \
        --compress=9 \
        > "$backup_file" || error "Falha no backup do PostgreSQL"
    
    # Comprimir backup
    log "Comprimindo backup do banco de dados..."
    gzip "$backup_file" || error "Falha ao comprimir backup do banco"
    
    # Verificar integridade
    if [[ -f "$compressed_file" ]]; then
        local size=$(stat -f%z "$compressed_file" 2>/dev/null || stat -c%s "$compressed_file" 2>/dev/null)
        log "✅ Backup do banco criado: $compressed_file (${size} bytes)"
        echo "$compressed_file" > "$BACKUP_DIR/latest_db_backup.txt"
    else
        error "Arquivo de backup não foi criado"
    fi
}

# Backup do Redis
backup_redis() {
    info "Iniciando backup do Redis..."
    
    local backup_file="$BACKUP_DIR/redis/redis_backup_$TIMESTAMP.rdb"
    local compressed_file="$backup_file.gz"
    
    # Verificar se o container está rodando
    if ! docker ps | grep -q "eventos_redis_prod"; then
        warn "Container Redis não está rodando, pulando backup"
        return 0
    fi
    
    # Fazer snapshot do Redis
    log "Criando snapshot do Redis..."
    docker exec eventos_redis_prod redis-cli BGSAVE || warn "Falha ao criar snapshot do Redis"
    
    # Aguardar conclusão do snapshot
    sleep 5
    
    # Copiar arquivo RDB
    log "Copiando dados do Redis..."
    docker cp eventos_redis_prod:/data/dump.rdb "$backup_file" || warn "Falha ao copiar dados do Redis"
    
    if [[ -f "$backup_file" ]]; then
        # Comprimir backup
        gzip "$backup_file" || warn "Falha ao comprimir backup do Redis"
        
        local size=$(stat -f%z "$compressed_file" 2>/dev/null || stat -c%s "$compressed_file" 2>/dev/null)
        log "✅ Backup do Redis criado: $compressed_file (${size} bytes)"
        echo "$compressed_file" > "$BACKUP_DIR/latest_redis_backup.txt"
    else
        warn "⚠️  Backup do Redis não foi criado"
    fi
}

# Backup do RabbitMQ
backup_rabbitmq() {
    info "Iniciando backup do RabbitMQ..."
    
    local backup_file="$BACKUP_DIR/rabbitmq/rabbitmq_backup_$TIMESTAMP.json"
    local compressed_file="$backup_file.gz"
    
    # Verificar se o container está rodando
    if ! docker ps | grep -q "eventos_rabbitmq_prod"; then
        warn "Container RabbitMQ não está rodando, pulando backup"
        return 0
    fi
    
    # Exportar definições do RabbitMQ
    log "Exportando configurações do RabbitMQ..."
    docker exec eventos_rabbitmq_prod rabbitmqctl export_definitions /tmp/definitions.json || warn "Falha ao exportar definições do RabbitMQ"
    
    # Copiar arquivo de definições
    docker cp eventos_rabbitmq_prod:/tmp/definitions.json "$backup_file" || warn "Falha ao copiar definições do RabbitMQ"
    
    if [[ -f "$backup_file" ]]; then
        # Comprimir backup
        gzip "$backup_file" || warn "Falha ao comprimir backup do RabbitMQ"
        
        local size=$(stat -f%z "$compressed_file" 2>/dev/null || stat -c%s "$compressed_file" 2>/dev/null)
        log "✅ Backup do RabbitMQ criado: $compressed_file (${size} bytes)"
        echo "$compressed_file" > "$BACKUP_DIR/latest_rabbitmq_backup.txt"
    else
        warn "⚠️  Backup do RabbitMQ não foi criado"
    fi
}

# Backup das configurações
backup_configs() {
    info "Iniciando backup das configurações..."
    
    local backup_file="$BACKUP_DIR/configs/configs_backup_$TIMESTAMP.tar.gz"
    
    # Criar arquivo tar com configurações
    log "Comprimindo arquivos de configuração..."
    tar -czf "$backup_file" \
        -C "$PROJECT_DIR" \
        configs/ \
        docker-compose.production.yml \
        .env.production.example \
        scripts/ \
        2>/dev/null || warn "Alguns arquivos de configuração podem não existir"
    
    if [[ -f "$backup_file" ]]; then
        local size=$(stat -f%z "$backup_file" 2>/dev/null || stat -c%s "$backup_file" 2>/dev/null)
        log "✅ Backup das configurações criado: $backup_file (${size} bytes)"
        echo "$backup_file" > "$BACKUP_DIR/latest_config_backup.txt"
    else
        warn "⚠️  Backup das configurações não foi criado"
    fi
}

# Backup dos logs
backup_logs() {
    info "Iniciando backup dos logs..."
    
    local backup_file="$BACKUP_DIR/logs/logs_backup_$TIMESTAMP.tar.gz"
    local logs_dir="/var/log"
    
    # Criar backup dos logs se existirem
    if [[ -d "$logs_dir" ]]; then
        log "Comprimindo arquivos de log..."
        find "$logs_dir" -name "*eventos*" -type f -mtime -7 | \
            tar -czf "$backup_file" -T - 2>/dev/null || warn "Nenhum log encontrado para backup"
        
        if [[ -f "$backup_file" ]]; then
            local size=$(stat -f%z "$backup_file" 2>/dev/null || stat -c%s "$backup_file" 2>/dev/null)
            log "✅ Backup dos logs criado: $backup_file (${size} bytes)"
            echo "$backup_file" > "$BACKUP_DIR/latest_logs_backup.txt"
        fi
    else
        warn "⚠️  Diretório de logs não encontrado"
    fi
}

# Upload para S3 (se configurado)
upload_to_s3() {
    if [[ "$AWS_BACKUP_ENABLED" != "true" ]]; then
        info "Backup S3 não está habilitado, pulando upload"
        return 0
    fi
    
    info "Iniciando upload para S3..."
    
    local s3_prefix="eventos-backup/$(date +%Y/%m/%d)"
    
    # Upload dos backups
    find "$BACKUP_DIR" -name "*_$TIMESTAMP.*" -type f | while read -r file; do
        local filename=$(basename "$file")
        local s3_key="$s3_prefix/$filename"
        
        log "Uploading $filename para S3..."
        aws s3 cp "$file" "s3://$S3_BUCKET/$s3_key" || warn "Falha no upload de $filename"
    done
    
    log "✅ Upload para S3 concluído"
}

# Limpeza de backups antigos
cleanup_old_backups() {
    info "Limpando backups antigos (mais de $RETENTION_DAYS dias)..."
    
    # Limpeza local
    find "$BACKUP_DIR" -type f -mtime +$RETENTION_DAYS -delete 2>/dev/null || true
    
    # Limpeza no S3 (se configurado)
    if [[ "$AWS_BACKUP_ENABLED" == "true" ]]; then
        local cutoff_date=$(date -d "$RETENTION_DAYS days ago" +%Y-%m-%d)
        log "Limpando backups S3 anteriores a $cutoff_date..."
        
        aws s3api list-objects-v2 \
            --bucket "$S3_BUCKET" \
            --prefix "eventos-backup/" \
            --query "Contents[?LastModified<='$cutoff_date'].Key" \
            --output text | \
        while read -r key; do
            [[ -n "$key" && "$key" != "None" ]] && aws s3 rm "s3://$S3_BUCKET/$key" || true
        done 2>/dev/null || warn "Falha na limpeza do S3"
    fi
    
    log "✅ Limpeza concluída"
}

# Verificar integridade dos backups
verify_backups() {
    info "Verificando integridade dos backups..."
    
    local error_count=0
    
    # Verificar backup do banco
    if [[ -f "$BACKUP_DIR/latest_db_backup.txt" ]]; then
        local db_backup=$(cat "$BACKUP_DIR/latest_db_backup.txt")
        if [[ -f "$db_backup" ]]; then
            log "✅ Backup do banco verificado: $(basename "$db_backup")"
        else
            warn "⚠️  Backup do banco não encontrado"
            ((error_count++))
        fi
    fi
    
    # Verificar outros backups...
    for service in redis rabbitmq config logs; do
        local latest_file="$BACKUP_DIR/latest_${service}_backup.txt"
        if [[ -f "$latest_file" ]]; then
            local backup_file=$(cat "$latest_file")
            if [[ -f "$backup_file" ]]; then
                log "✅ Backup do $service verificado: $(basename "$backup_file")"
            else
                warn "⚠️  Backup do $service não encontrado"
                ((error_count++))
            fi
        fi
    done
    
    if [[ $error_count -gt 0 ]]; then
        warn "⚠️  $error_count backup(s) com problemas"
        return 1
    else
        log "✅ Todos os backups verificados com sucesso"
        return 0
    fi
}

# Relatório de backup
generate_report() {
    info "Gerando relatório de backup..."
    
    local report_file="$BACKUP_DIR/backup_report_$TIMESTAMP.txt"
    
    {
        echo "========================================="
        echo "RELATÓRIO DE BACKUP - $(date)"
        echo "========================================="
        echo
        echo "Timestamp: $TIMESTAMP"
        echo "Diretório: $BACKUP_DIR"
        echo "Retenção: $RETENTION_DAYS dias"
        echo "AWS S3: $AWS_BACKUP_ENABLED"
        echo
        echo "ARQUIVOS CRIADOS:"
        echo "-----------------"
        
        find "$BACKUP_DIR" -name "*_$TIMESTAMP.*" -type f | while read -r file; do
            local size=$(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null)
            echo "$(basename "$file"): ${size} bytes"
        done
        
        echo
        echo "ESPAÇO UTILIZADO:"
        echo "-----------------"
        du -sh "$BACKUP_DIR" 2>/dev/null || echo "Erro ao calcular espaço"
        
        echo
        echo "STATUS DOS SERVIÇOS:"
        echo "--------------------"
        docker ps --format "table {{.Names}}\t{{.Status}}" | grep eventos || echo "Nenhum container encontrado"
        
    } > "$report_file"
    
    log "✅ Relatório gerado: $report_file"
    
    # Enviar relatório por notificação
    local summary="🔄 Backup concluído em $(date). Verifique $report_file para detalhes."
    send_notification "$summary" "info"
}

# Função principal
main() {
    log "🔄 Iniciando processo de backup automatizado"
    log "📅 Data: $(date)"
    log "⏰ Timestamp: $TIMESTAMP"
    
    check_dependencies
    setup_directories
    
    # Executar backups
    backup_database
    backup_redis
    backup_rabbitmq
    backup_configs
    backup_logs
    
    # Upload e limpeza
    upload_to_s3
    cleanup_old_backups
    
    # Verificação e relatório
    if verify_backups; then
        generate_report
        log "🎉 Processo de backup concluído com sucesso!"
        send_notification "✅ Backup automatizado concluído com sucesso!" "info"
    else
        warn "⚠️  Backup concluído com alguns problemas. Verifique os logs."
        send_notification "⚠️ Backup concluído com problemas. Verifique os logs." "warning"
    fi
}

# Verificar argumentos
case "${1:-backup}" in
    "backup"|"")
        main
        ;;
    "restore")
        echo "Funcionalidade de restore ainda não implementada"
        echo "Para restaurar manualmente:"
        echo "1. Pare os serviços: docker-compose -f docker-compose.production.yml down"
        echo "2. Restaure o banco: docker exec -i postgres_container pg_restore -U user -d database < backup.sql"
        echo "3. Restaure outros serviços conforme necessário"
        echo "4. Reinicie os serviços: docker-compose -f docker-compose.production.yml up -d"
        ;;
    "verify")
        verify_backups
        ;;
    "cleanup")
        cleanup_old_backups
        ;;
    "report")
        generate_report
        ;;
    *)
        echo "Uso: $0 [backup|restore|verify|cleanup|report]"
        echo
        echo "Comandos disponíveis:"
        echo "  backup   - Executar backup completo (padrão)"
        echo "  restore  - Instruções para restaurar backup"
        echo "  verify   - Verificar integridade dos backups"
        echo "  cleanup  - Limpar backups antigos"
        echo "  report   - Gerar relatório dos backups"
        exit 1
        ;;
esac