#!/bin/bash

# Script para configurar Cron Jobs de Backup
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
CRON_USER=${CRON_USER:-root}
BACKUP_SCHEDULE=${BACKUP_SCHEDULE:-"0 2 * * *"}  # Todo dia às 2h da manhã

# Funções de log
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
    exit 1
}

info() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

# Verificar permissões
check_permissions() {
    if [[ $EUID -ne 0 ]]; then
        error "Este script deve ser executado como root para configurar cron jobs"
    fi
}

# Verificar se os scripts existem
check_scripts() {
    info "Verificando scripts de backup..."
    
    if [[ ! -f "$SCRIPT_DIR/backup.sh" ]]; then
        error "Script de backup não encontrado: $SCRIPT_DIR/backup.sh"
    fi
    
    if [[ ! -x "$SCRIPT_DIR/backup.sh" ]]; then
        log "Tornando script de backup executável..."
        chmod +x "$SCRIPT_DIR/backup.sh"
    fi
    
    log "✅ Scripts verificados"
}

# Configurar logrotate para logs de backup
setup_logrotate() {
    info "Configurando logrotate para logs de backup..."
    
    local logrotate_config="/etc/logrotate.d/eventos-backup"
    
    cat > "$logrotate_config" << EOF
/var/log/eventos-backup.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 644 root root
    postrotate
        # Restart rsyslog if needed
        /bin/kill -HUP \$(cat /var/run/rsyslogd.pid 2> /dev/null) 2> /dev/null || true
    endscript
}

/var/log/eventos-deploy.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 644 root root
    postrotate
        /bin/kill -HUP \$(cat /var/run/rsyslogd.pid 2> /dev/null) 2> /dev/null || true
    endscript
}
EOF
    
    log "✅ Logrotate configurado: $logrotate_config"
}

# Configurar cron job para backup
setup_backup_cron() {
    info "Configurando cron job para backup automatizado..."
    
    local cron_file="/etc/cron.d/eventos-backup"
    local backup_script="$SCRIPT_DIR/backup.sh"
    
    # Criar arquivo de cron
    cat > "$cron_file" << EOF
# Backup automatizado do Sistema de Check-in em Eventos
# Executa backup completo diariamente às 2h da manhã
SHELL=/bin/bash
PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin

# Variáveis de ambiente para backup
BACKUP_RETENTION_DAYS=30
AWS_BACKUP_ENABLED=false
LOG_FILE=/var/log/eventos-backup.log

# Backup diário
$BACKUP_SCHEDULE $CRON_USER cd $PROJECT_DIR && $backup_script backup >> /var/log/eventos-backup.log 2>&1

# Verificação semanal da integridade dos backups (domingos às 3h)
0 3 * * 0 $CRON_USER cd $PROJECT_DIR && $backup_script verify >> /var/log/eventos-backup.log 2>&1

# Limpeza mensal de backups antigos (primeiro dia do mês às 4h)
0 4 1 * * $CRON_USER cd $PROJECT_DIR && $backup_script cleanup >> /var/log/eventos-backup.log 2>&1
EOF
    
    # Definir permissões corretas
    chmod 644 "$cron_file"
    chown root:root "$cron_file"
    
    log "✅ Cron job configurado: $cron_file"
    log "📅 Agenda de backup: $BACKUP_SCHEDULE"
}

# Configurar cron job para monitoramento
setup_monitoring_cron() {
    info "Configurando cron job para monitoramento..."
    
    local cron_file="/etc/cron.d/eventos-monitoring"
    local health_check_script="$SCRIPT_DIR/health-check.sh"
    
    # Criar script de health check se não existir
    if [[ ! -f "$health_check_script" ]]; then
        info "Criando script de health check..."
        
        cat > "$health_check_script" << 'EOF'
#!/bin/bash
# Health check automatizado

LOG_FILE="/var/log/eventos-health.log"
HEALTH_URL="http://localhost:8080/health"

log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1" >> "$LOG_FILE"
}

# Verificar se a aplicação está respondendo
if curl -f -s "$HEALTH_URL" >/dev/null 2>&1; then
    log "✅ Aplicação saudável"
    exit 0
else
    log "❌ Aplicação não está respondendo"
    
    # Tentar reiniciar se necessário (opcional)
    # docker-compose -f docker-compose.production.yml restart api
    
    exit 1
fi
EOF
        
        chmod +x "$health_check_script"
    fi
    
    # Criar arquivo de cron para monitoramento
    cat > "$cron_file" << EOF
# Monitoramento automatizado do Sistema de Check-in em Eventos
SHELL=/bin/bash
PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin

# Health check a cada 5 minutos
*/5 * * * * $CRON_USER $health_check_script

# Verificação de espaço em disco a cada hora
0 * * * * $CRON_USER df -h / /var >> /var/log/eventos-disk.log 2>&1
EOF
    
    chmod 644 "$cron_file"
    chown root:root "$cron_file"
    
    log "✅ Monitoramento configurado: $cron_file"
}

# Configurar alertas por email (opcional)
setup_email_alerts() {
    info "Configurando alertas por email..."
    
    # Verificar se postfix ou sendmail está instalado
    if command -v sendmail >/dev/null 2>&1 || command -v postfix >/dev/null 2>&1; then
        log "✅ Sistema de email detectado"
        
        # Criar script de alerta
        local alert_script="$SCRIPT_DIR/send-alert.sh"
        
        cat > "$alert_script" << 'EOF'
#!/bin/bash
# Script para envio de alertas por email

ADMIN_EMAIL=${ADMIN_EMAIL:-"admin@localhost"}
SUBJECT="[Eventos] Alerta do Sistema"

# Enviar email
echo "$2" | mail -s "$SUBJECT: $1" "$ADMIN_EMAIL"
EOF
        
        chmod +x "$alert_script"
        log "✅ Script de alertas criado: $alert_script"
    else
        warn "⚠️  Sistema de email não detectado. Instale postfix ou sendmail para alertas por email."
    fi
}

# Testar configuração
test_configuration() {
    info "Testando configuração do cron..."
    
    # Verificar sintaxe dos arquivos de cron
    for cron_file in /etc/cron.d/eventos-*; do
        if [[ -f "$cron_file" ]]; then
            log "Verificando $cron_file..."
            # Verificação básica de sintaxe
            if grep -q "^[0-9*]" "$cron_file"; then
                log "✅ $cron_file parece válido"
            else
                warn "⚠️  $cron_file pode ter problemas de sintaxe"
            fi
        fi
    done
    
    # Reiniciar cron para aplicar mudanças
    log "Reiniciando serviço cron..."
    systemctl restart cron || systemctl restart crond || warn "Falha ao reiniciar cron"
    
    # Verificar status do cron
    if systemctl is-active --quiet cron || systemctl is-active --quiet crond; then
        log "✅ Serviço cron está ativo"
    else
        error "❌ Serviço cron não está ativo"
    fi
}

# Mostrar configuração atual
show_configuration() {
    info "Configuração atual do cron:"
    echo
    
    echo "📋 Arquivos de cron criados:"
    ls -la /etc/cron.d/eventos-* 2>/dev/null || echo "Nenhum arquivo encontrado"
    echo
    
    echo "📅 Jobs agendados:"
    crontab -l -u "$CRON_USER" 2>/dev/null || echo "Nenhum crontab pessoal encontrado"
    echo
    
    echo "🗂️  Conteúdo dos arquivos de cron:"
    for file in /etc/cron.d/eventos-*; do
        if [[ -f "$file" ]]; then
            echo "--- $file ---"
            cat "$file"
            echo
        fi
    done
}

# Remover configuração
remove_configuration() {
    info "Removendo configuração do cron..."
    
    # Remover arquivos de cron
    rm -f /etc/cron.d/eventos-*
    
    # Remover configuração do logrotate
    rm -f /etc/logrotate.d/eventos-backup
    
    # Reiniciar cron
    systemctl restart cron || systemctl restart crond || warn "Falha ao reiniciar cron"
    
    log "✅ Configuração removida"
}

# Função principal
main() {
    log "🔧 Configurando cron jobs para o Sistema de Check-in em Eventos"
    log "📅 Agenda de backup: $BACKUP_SCHEDULE"
    log "👤 Usuário do cron: $CRON_USER"
    
    check_permissions
    check_scripts
    setup_logrotate
    setup_backup_cron
    setup_monitoring_cron
    setup_email_alerts
    test_configuration
    
    log "🎉 Configuração do cron concluída com sucesso!"
    echo
    show_configuration
    
    echo
    info "💡 Próximos passos:"
    echo "1. Verifique os logs em /var/log/eventos-*.log"
    echo "2. Configure as variáveis de ambiente no arquivo .env.production"
    echo "3. Teste o backup manualmente: $SCRIPT_DIR/backup.sh"
    echo "4. Configure notificações (Slack, Discord) se necessário"
}

# Verificar argumentos
case "${1:-setup}" in
    "setup"|"")
        main
        ;;
    "remove")
        remove_configuration
        ;;
    "test")
        test_configuration
        ;;
    "show")
        show_configuration
        ;;
    *)
        echo "Uso: $0 [setup|remove|test|show]"
        echo
        echo "Comandos disponíveis:"
        echo "  setup  - Configurar cron jobs (padrão)"
        echo "  remove - Remover configuração do cron"
        echo "  test   - Testar configuração atual"
        echo "  show   - Mostrar configuração atual"
        exit 1
        ;;
esac