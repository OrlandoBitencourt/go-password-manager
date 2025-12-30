#!/bin/bash

# Script para verificar e instalar Docker no Ubuntu

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Verificar se já está instalado
if command -v docker &> /dev/null; then
    success "Docker já está instalado!"
    docker --version

    if command -v docker compose &> /dev/null; then
        success "Docker Compose já está instalado!"
        docker compose version
    else
        warning "Docker Compose não encontrado. Instalando..."
    fi
else
    info "Docker não encontrado. Iniciando instalação..."

    # Atualizar repositórios
    info "Atualizando repositórios..."
    sudo apt update

    # Instalar dependências
    info "Instalando dependências..."
    sudo apt install -y ca-certificates curl gnupg lsb-release

    # Adicionar chave GPG do Docker
    info "Adicionando chave GPG do Docker..."
    sudo install -m 0755 -d /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
    sudo chmod a+r /etc/apt/keyrings/docker.gpg

    # Adicionar repositório
    info "Adicionando repositório do Docker..."
    echo \
      "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
      $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
      sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

    # Atualizar novamente
    sudo apt update

    # Instalar Docker
    info "Instalando Docker Engine..."
    sudo apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

    success "Docker instalado com sucesso!"
fi

# Verificar se o serviço está rodando
if ! sudo systemctl is-active --quiet docker; then
    info "Iniciando serviço Docker..."
    sudo systemctl enable docker
    sudo systemctl start docker
    success "Serviço Docker iniciado!"
else
    success "Serviço Docker já está rodando!"
fi

# Verificar se o usuário está no grupo docker
if ! groups | grep -q docker; then
    warning "Adicionando usuário $USER ao grupo docker..."
    sudo usermod -aG docker $USER
    warning "IMPORTANTE: Você precisa fazer logout e login novamente (ou reiniciar)"
    warning "para que as mudanças de grupo tenham efeito!"
    echo ""
    warning "Após fazer logout/login, execute novamente: ./manage.sh install"
else
    success "Usuário já está no grupo docker!"
fi

# Teste final
echo ""
info "Testando instalação..."
sudo docker run --rm hello-world

echo ""
success "Instalação concluída!"
echo ""
info "Próximos passos:"
echo "  1. Se você acabou de ser adicionado ao grupo docker, faça logout e login"
echo "  2. Configure o arquivo .env com suas credenciais"
echo "  3. Execute: ./manage.sh install"
echo "  4. Execute: ./manage.sh start"
