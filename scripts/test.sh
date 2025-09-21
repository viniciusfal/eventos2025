#!/bin/bash

# Script para executar testes

echo "Executando testes..."

# Executar testes unitários
echo "Executando testes unitários..."
go test -v ./internal/domain/... ./internal/application/...

# Executar testes de integração
echo "Executando testes de integração..."
go test -v ./tests/integration/...

# Executar testes com cobertura
echo "Executando testes com cobertura..."
go test -v -coverprofile=coverage.out ./...

# Gerar relatório de cobertura
echo "Gerando relatório de cobertura..."
go tool cover -html=coverage.out -o coverage.html

echo "Testes concluídos!"
echo "Relatório de cobertura disponível em: coverage.html"
