#!/bin/bash

# Script para iniciar o ambiente de desenvolvimento

echo "Iniciando ambiente de desenvolvimento..."

# Verificar se Docker está rodando
if ! docker info > /dev/null 2>&1; then
    echo "Erro: Docker não está rodando. Por favor, inicie o Docker primeiro."
    exit 1
fi

# Subir serviços de infraestrutura
echo "Subindo serviços de infraestrutura..."
docker-compose up -d postgres redis rabbitmq

# Aguardar serviços ficarem prontos
echo "Aguardando serviços ficarem prontos..."
sleep 10

# Executar migrações
echo "Executando migrações do banco de dados..."
make migrate-up

# Subir serviços de monitoramento
echo "Subindo serviços de monitoramento..."
docker-compose up -d prometheus grafana

echo "Ambiente iniciado com sucesso!"
echo ""
echo "Serviços disponíveis:"
echo "- PostgreSQL: localhost:5432"
echo "- Redis: localhost:6379"
echo "- RabbitMQ Management: http://localhost:15672 (eventos_user/eventos_password)"
echo "- Prometheus: http://localhost:9090"
echo "- Grafana: http://localhost:3000 (admin/admin)"
echo ""
echo "Para iniciar a aplicação, execute: make run"
