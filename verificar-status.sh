#!/bin/bash

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${BLUE}    DIAGNÓSTICO COMPLETO - PASSWORD MANAGER${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo ""

echo -e "${YELLOW}1. Status dos Containers:${NC}"
docker ps -a --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
echo ""

echo -e "${YELLOW}2. Testando Conectividade:${NC}"
echo -n "Backend (8080): "
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo -e "${GREEN}✓ OK${NC}"
else
    echo -e "${RED}✗ FALHOU${NC}"
fi

echo -n "Frontend (3000): "
if curl -s http://localhost:3000 > /dev/null 2>&1; then
    echo -e "${GREEN}✓ OK${NC}"
else
    echo -e "${RED}✗ FALHOU${NC}"
fi
echo ""

echo -e "${YELLOW}3. Últimas 10 linhas dos logs de cada serviço:${NC}"
echo ""
echo -e "${BLUE}--- BACKEND ---${NC}"
docker logs password-manager-backend --tail 10 2>&1 || echo "Container não encontrado"
echo ""

echo -e "${BLUE}--- TELEGRAM BOT ---${NC}"
docker logs password-manager-telegram-bot --tail 10 2>&1 || echo "Container não encontrado"
echo ""

echo -e "${BLUE}--- FRONTEND ---${NC}"
docker logs password-manager-frontend --tail 10 2>&1 || echo "Container não encontrado"
echo ""

echo -e "${YELLOW}4. Portas em Uso:${NC}"
ss -tlnp 2>/dev/null | grep -E ':(8080|3000)' || netstat -tlnp 2>/dev/null | grep -E ':(8080|3000)' || echo "Nenhuma porta 8080 ou 3000 em uso"
echo ""

echo -e "${YELLOW}5. Recursos Docker:${NC}"
docker system df
echo ""

echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${GREEN}Para ver logs completos de um serviço específico:${NC}"
echo "  docker logs password-manager-backend -f"
echo "  docker logs password-manager-frontend -f"
echo "  docker logs password-manager-telegram-bot -f"
echo ""
echo -e "${GREEN}Para reconstruir tudo do zero:${NC}"
echo "  docker compose down -v --rmi all"
echo "  docker compose up -d --build"
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
