#!/bin/bash

# Script para executar migrações do banco de dados

set -e

# Configurações padrão
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_NAME=${DB_NAME:-eventos_db}
DB_USER=${DB_USER:-eventos_user}
DB_PASSWORD=${DB_PASSWORD:-eventos_password}

echo "Executando migrações do banco de dados..."
echo "Host: $DB_HOST:$DB_PORT"
echo "Database: $DB_NAME"
echo "User: $DB_USER"

# Verificar se o PostgreSQL está acessível
echo "Verificando conexão com o banco de dados..."
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "SELECT 1;" > /dev/null 2>&1

if [ $? -ne 0 ]; then
    echo "Erro: Não foi possível conectar ao PostgreSQL."
    echo "Verifique se o serviço está rodando e as credenciais estão corretas."
    exit 1
fi

# Verificar se o banco de dados existe, se não, criar
echo "Verificando se o banco de dados existe..."
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -tc "SELECT 1 FROM pg_database WHERE datname = '$DB_NAME';" | grep -q 1

if [ $? -ne 0 ]; then
    echo "Criando banco de dados $DB_NAME..."
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "CREATE DATABASE $DB_NAME;"
fi

# Executar migrações
echo "Executando migração: 001_create_database_schema.sql"
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f migrations/001_create_database_schema.sql

if [ $? -eq 0 ]; then
    echo "Migrações executadas com sucesso!"
else
    echo "Erro ao executar migrações!"
    exit 1
fi

echo "Verificando estrutura do banco..."
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "\dt"

echo "Migração concluída!"
