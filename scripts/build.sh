#!/bin/bash

# Script para build da aplicação

echo "Fazendo build da aplicação..."

# Limpar build anterior
echo "Limpando build anterior..."
rm -rf build/

# Criar diretório de build
mkdir -p build/

# Verificar dependências
echo "Verificando dependências..."
go mod tidy

# Executar testes
echo "Executando testes..."
go test ./...

if [ $? -ne 0 ]; then
    echo "Erro: Testes falharam. Build cancelado."
    exit 1
fi

# Build da aplicação
echo "Compilando aplicação..."
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o build/main ./cmd/api

if [ $? -eq 0 ]; then
    echo "Build concluído com sucesso!"
    echo "Binário disponível em: build/main"
else
    echo "Erro no build!"
    exit 1
fi
