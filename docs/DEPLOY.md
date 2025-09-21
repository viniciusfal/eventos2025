# 🚀 Guia de Deploy e Produção

## Sistema de Check-in em Eventos - Fase 10 Implementada

Este documento contém todas as informações necessárias para fazer deploy do sistema em produção.

---

## 📋 **Pré-requisitos**

### **Servidor de Produção**
- **SO**: Ubuntu 20.04+ ou CentOS 8+
- **RAM**: Mínimo 4GB (Recomendado 8GB)
- **CPU**: Mínimo 2 cores (Recomendado 4 cores)
- **Disco**: Mínimo 50GB SSD
- **Rede**: Conexão estável com internet

### **Software Necessário**
```bash
# Docker & Docker Compose
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh
sudo usermod -aG docker $USER

# Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Utilitários
sudo apt update
sudo apt install -y curl wget git htop vim
```

---

## 🔧 **Configuração Inicial**

### **1. Clonar o Repositório**
```bash
git clone https://github.com/viniciusfal/eventos2025.git
cd eventos2025
```

### **2. Configurar Variáveis de Ambiente**
```bash
# Copiar arquivo de exemplo
cp .env.production.example .env.production

# Editar configurações
vim .env.production
```

**Variáveis Obrigatórias:**
```bash
# Banco de dados
DB_PASSWORD=sua-senha-super-segura-aqui
REDIS_PASSWORD=sua-senha-redis-aqui
RABBITMQ_PASSWORD=sua-senha-rabbitmq-aqui

# Segurança
JWT_SECRET=sua-chave-jwt-super-segura-de-pelo-menos-32-caracteres

# Monitoramento
GRAFANA_PASSWORD=sua-senha-grafana-aqui
```

### **3. Configurar Permissões**
```bash
# Tornar scripts executáveis
chmod +x scripts/*.sh

# Criar diretórios necessários
sudo mkdir -p /var/backups/eventos
sudo mkdir -p /var/log
sudo chown -R $USER:$USER /var/backups/eventos
```

---

## 🚀 **Deploy Automatizado**

### **Deploy Simples (Recomendado)**
```bash
# Executar script de deploy
sudo ./scripts/deploy.sh

# Verificar status
sudo ./scripts/deploy.sh status
```

### **Deploy Manual**
```bash
# 1. Fazer pull da imagem mais recente
docker pull ghcr.io/viniciusfal/eventos2025:latest

# 2. Parar containers existentes
docker-compose -f docker-compose.production.yml --env-file .env.production down

# 3. Subir aplicação
docker-compose -f docker-compose.production.yml --env-file .env.production up -d

# 4. Verificar saúde
curl http://localhost:8080/health
```

---

## 🔍 **Verificação de Deploy**

### **Health Checks**
```bash
# Health geral do sistema
curl http://localhost:8080/health

# Readiness check
curl http://localhost:8080/ready

# Liveness check
curl http://localhost:8080/live

# Métricas Prometheus
curl http://localhost:8080/metrics
```

### **Smoke Tests**
```bash
# Executar testes básicos
./scripts/deploy.sh health

# Verificar logs
docker-compose -f docker-compose.production.yml logs -f api
```

---

## 📊 **Monitoramento**

### **Interfaces Web**
- **API**: http://localhost:8080
- **Swagger**: http://localhost:8080/swagger/index.html
- **Grafana**: http://localhost:3000 (admin/sua-senha)
- **Prometheus**: http://localhost:9090
- **RabbitMQ Management**: http://localhost:15672

### **Métricas Disponíveis**
- **HTTP Requests**: Duração, status codes, throughput
- **Database**: Queries, conexões, performance
- **Cache**: Hit ratio, operações Redis
- **Business Logic**: Check-ins, usuários, eventos
- **Sistema**: CPU, memória, goroutines

---

## 💾 **Sistema de Backup**

### **Backup Manual**
```bash
# Backup completo
./scripts/backup.sh

# Verificar integridade
./scripts/backup.sh verify

# Gerar relatório
./scripts/backup.sh report
```

### **Backup Automatizado**
```bash
# Configurar cron jobs
sudo ./scripts/setup-cron.sh

# Verificar configuração
sudo ./scripts/setup-cron.sh show

# Testar configuração
sudo ./scripts/setup-cron.sh test
```

### **Agenda de Backup**
- **Backup Completo**: Diário às 2h da manhã
- **Verificação**: Semanal aos domingos às 3h
- **Limpeza**: Mensal no primeiro dia às 4h
- **Retenção**: 30 dias (configurável)

---

## 🔒 **Segurança**

### **Configurações de Segurança**
```bash
# Firewall básico
sudo ufw allow ssh
sudo ufw allow 80
sudo ufw allow 443
sudo ufw enable

# Limitar acesso aos serviços
# Grafana, Prometheus e RabbitMQ são acessíveis apenas via localhost
```

### **SSL/TLS (Recomendado)**
```bash
# Instalar certbot
sudo apt install certbot

# Obter certificado
sudo certbot certonly --standalone -d seu-dominio.com

# Configurar renovação automática
sudo crontab -e
# Adicionar: 0 12 * * * /usr/bin/certbot renew --quiet
```

### **Variáveis Sensíveis**
- ✅ Senhas geradas aleatoriamente
- ✅ JWT secrets com 32+ caracteres
- ✅ Containers rodando como usuário não-root
- ✅ Volumes com permissões restritivas

---

## 📈 **Escalabilidade**

### **Recursos por Container**
```yaml
# Configurado em docker-compose.production.yml
api:
  memory: 512M (limit) / 256M (reservation)
  cpu: 0.5 (limit) / 0.25 (reservation)

postgres:
  memory: 1G (limit) / 512M (reservation)
  cpu: 1.0 (limit) / 0.5 (reservation)
```

### **Otimizações de Performance**
- ✅ PostgreSQL com configurações otimizadas
- ✅ Redis com política LRU
- ✅ Conexões de banco com pool
- ✅ Compressão gzip habilitada
- ✅ Cache de queries implementado

---

## 🛠️ **Manutenção**

### **Comandos Úteis**
```bash
# Ver logs em tempo real
docker-compose -f docker-compose.production.yml logs -f

# Reiniciar serviço específico
docker-compose -f docker-compose.production.yml restart api

# Limpar recursos não utilizados
docker system prune -f

# Verificar uso de recursos
docker stats

# Backup de emergência
./scripts/backup.sh backup
```

### **Atualização do Sistema**
```bash
# 1. Fazer backup
./scripts/backup.sh

# 2. Fazer deploy da nova versão
./scripts/deploy.sh

# 3. Verificar saúde
./scripts/deploy.sh health
```

---

## 🚨 **Troubleshooting**

### **Problemas Comuns**

#### **Aplicação não inicia**
```bash
# Verificar logs
docker-compose -f docker-compose.production.yml logs api

# Verificar variáveis de ambiente
docker-compose -f docker-compose.production.yml config

# Verificar conectividade com banco
docker exec -it eventos_postgres_prod pg_isready
```

#### **Banco de dados não conecta**
```bash
# Verificar status do PostgreSQL
docker exec -it eventos_postgres_prod pg_isready

# Conectar manualmente
docker exec -it eventos_postgres_prod psql -U eventos_user -d eventos_db

# Verificar logs
docker-compose -f docker-compose.production.yml logs postgres
```

#### **Cache Redis não funciona**
```bash
# Verificar Redis
docker exec -it eventos_redis_prod redis-cli ping

# Verificar autenticação
docker exec -it eventos_redis_prod redis-cli -a sua-senha ping

# Limpar cache
docker exec -it eventos_redis_prod redis-cli -a sua-senha FLUSHALL
```

### **Rollback de Emergência**
```bash
# Parar versão atual
docker-compose -f docker-compose.production.yml down

# Restaurar backup (se necessário)
# Instruções detalhadas em ./scripts/backup.sh restore

# Subir versão anterior
# docker pull ghcr.io/viniciusfal/eventos2025:tag-anterior
# Editar docker-compose.production.yml com a tag anterior
# docker-compose -f docker-compose.production.yml up -d
```

---

## 📞 **Suporte e Contato**

### **Logs Importantes**
- **Aplicação**: `docker-compose logs api`
- **Deploy**: `/var/log/eventos-deploy.log`
- **Backup**: `/var/log/eventos-backup.log`
- **Health**: `/var/log/eventos-health.log`

### **Comandos de Diagnóstico**
```bash
# Status completo do sistema
./scripts/deploy.sh status

# Relatório de backup
./scripts/backup.sh report

# Health check completo
curl -s http://localhost:8080/health | jq

# Métricas de sistema
curl -s http://localhost:8080/metrics | grep -E "(http_requests|db_queries|cache_)"
```

---

## 🎯 **Próximos Passos Recomendados**

1. **✅ Sistema Funcional** - Deploy realizado com sucesso
2. **🔒 Configurar SSL** - Certificados HTTPS para produção
3. **📊 Configurar Alertas** - Slack/Discord para notificações
4. **🌐 Configurar Domínio** - DNS e proxy reverso
5. **📈 Monitorar Performance** - Dashboards personalizados no Grafana
6. **🔄 CI/CD Avançado** - Deploy automático via GitHub Actions
7. **🛡️ Auditoria de Segurança** - Testes de penetração
8. **📱 Integração Frontend** - APIs para interfaces de usuário

---

**🎉 Sistema de Check-in em Eventos - Fase 10 Completa!**

O sistema está pronto para produção com:
- ✅ CI/CD automatizado
- ✅ Deploy com Docker
- ✅ Backup automatizado
- ✅ Monitoramento completo
- ✅ Scripts de manutenção
- ✅ Documentação detalhada