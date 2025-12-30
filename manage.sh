#!/bin/bash

# Password Manager - Script de Gerenciamento
# Este script facilita a instalação, inicialização e gerenciamento dos serviços

set -e

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Diretório do projeto
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_FILE="password-manager.service"
SYSTEMD_DIR="/etc/systemd/system"

# Função para mostrar mensagens coloridas
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

# Verificar se Docker está instalado
check_docker() {
    if ! command -v docker &> /dev/null; then
        error "Docker não está instalado!"
        echo "Por favor, instale o Docker primeiro."
        exit 1
    fi

    if ! command -v docker compose &> /dev/null; then
        error "Docker Compose não está instalado!"
        echo "Por favor, instale o Docker Compose primeiro."
        exit 1
    fi

    success "Docker e Docker Compose estão instalados."
}

# Verificar arquivo .env
check_env() {
    if [ ! -f "$PROJECT_DIR/.env" ]; then
        warning "Arquivo .env não encontrado!"
        if [ -f "$PROJECT_DIR/.env.example" ]; then
            info "Copiando .env.example para .env..."
            cp "$PROJECT_DIR/.env.example" "$PROJECT_DIR/.env"
            warning "Por favor, edite o arquivo .env com suas configurações:"
            echo "  - TELEGRAM_BOT_TOKEN: Token do seu bot do Telegram"
            echo "  - ALLOWED_USER_IDS: IDs dos usuários autorizados (opcional)"
            echo ""
            echo "Execute: nano $PROJECT_DIR/.env"
            exit 1
        else
            error "Arquivo .env.example não encontrado!"
            exit 1
        fi
    fi
    success "Arquivo .env encontrado."
}

# Instalar serviço systemd
install_service() {
    info "Instalando serviço systemd..."
    check_docker
    check_env

    # Copiar arquivo de serviço
    sudo cp "$PROJECT_DIR/$SERVICE_FILE" "$SYSTEMD_DIR/"

    # Recarregar systemd
    sudo systemctl daemon-reload

    # Habilitar serviço
    sudo systemctl enable password-manager.service

    success "Serviço instalado e habilitado para iniciar automaticamente!"
    info "Use './manage.sh start' para iniciar os serviços agora."
}

# Desinstalar serviço systemd
uninstall_service() {
    info "Desinstalando serviço systemd..."

    # Parar serviço se estiver rodando
    if systemctl is-active --quiet password-manager.service; then
        sudo systemctl stop password-manager.service
    fi

    # Desabilitar serviço
    sudo systemctl disable password-manager.service 2>/dev/null || true

    # Remover arquivo de serviço
    sudo rm -f "$SYSTEMD_DIR/$SERVICE_FILE"

    # Recarregar systemd
    sudo systemctl daemon-reload

    success "Serviço desinstalado!"
}

# Iniciar serviços
start() {
    info "Iniciando serviços..."
    check_docker
    check_env

    if systemctl is-enabled --quiet password-manager.service 2>/dev/null; then
        info "Tentando iniciar via systemd..."
        if sudo systemctl start password-manager.service; then
            success "Serviços iniciados via systemd!"
        else
            error "Falha ao iniciar via systemd. Tentando docker compose direto..."
            cd "$PROJECT_DIR"
            docker compose up -d
            success "Serviços iniciados via docker compose!"
        fi
    else
        cd "$PROJECT_DIR"
        docker compose up -d
        success "Serviços iniciados via docker compose!"
    fi

    info "Aguardando serviços ficarem prontos..."
    sleep 5
    status
}

# Parar serviços
stop() {
    info "Parando serviços..."

    if systemctl is-enabled --quiet password-manager.service 2>/dev/null; then
        sudo systemctl stop password-manager.service
    else
        cd "$PROJECT_DIR"
        docker compose down
    fi

    success "Serviços parados!"
}

# Reiniciar serviços
restart() {
    info "Reiniciando serviços..."

    if systemctl is-enabled --quiet password-manager.service 2>/dev/null; then
        sudo systemctl restart password-manager.service
    else
        cd "$PROJECT_DIR"
        docker compose restart
    fi

    success "Serviços reiniciados!"
    sleep 3
    status
}

# Status dos serviços
status() {
    info "Status dos serviços:"
    echo ""

    if systemctl is-enabled --quiet password-manager.service 2>/dev/null; then
        sudo systemctl status password-manager.service --no-pager
        echo ""
    fi

    cd "$PROJECT_DIR"
    docker compose ps
}

# Ver logs
logs() {
    cd "$PROJECT_DIR"

    if [ -n "$1" ]; then
        # Log de um serviço específico
        docker compose logs -f "$1"
    else
        # Logs de todos os serviços
        docker compose logs -f
    fi
}

# Build das imagens
build() {
    info "Construindo imagens Docker..."
    check_docker
    cd "$PROJECT_DIR"
    docker compose build --no-cache
    success "Imagens construídas com sucesso!"
}

# Atualizar serviços
update() {
    info "Atualizando serviços..."
    build
    restart
    success "Serviços atualizados!"
}

# Limpar containers e volumes
clean() {
    warning "Isso irá remover todos os containers e imagens do projeto."
    read -p "Tem certeza? (s/N): " -n 1 -r
    echo

    if [[ $REPLY =~ ^[SsYy]$ ]]; then
        info "Limpando..."
        cd "$PROJECT_DIR"
        docker compose down -v --rmi all
        success "Limpeza concluída!"
    else
        info "Operação cancelada."
    fi
}

# Mostrar ajuda
show_help() {
    cat << EOF
Password Manager - Script de Gerenciamento

Uso: $0 [comando]

Comandos disponíveis:

  install       Instala e habilita o serviço systemd para iniciar automaticamente
  uninstall     Remove o serviço systemd

  start         Inicia todos os serviços (backend, telegram bot, frontend)
  stop          Para todos os serviços
  restart       Reinicia todos os serviços
  status        Mostra o status dos serviços

  logs [serv]   Exibe os logs (opcionalmente de um serviço específico)
                Serviços: backend, telegram-bot, frontend

  build         Reconstrói as imagens Docker
  update        Reconstrói e reinicia os serviços
  clean         Remove containers, imagens e volumes

  help          Mostra esta mensagem de ajuda

Exemplos:

  # Primeira instalação
  $0 install
  $0 start

  # Ver logs do telegram bot
  $0 logs telegram-bot

  # Atualizar após mudanças no código
  $0 update

  # Status dos serviços
  $0 status

EOF
}

# Processar comando
case "${1:-help}" in
    install)
        install_service
        ;;
    uninstall)
        uninstall_service
        ;;
    start)
        start
        ;;
    stop)
        stop
        ;;
    restart)
        restart
        ;;
    status)
        status
        ;;
    logs)
        logs "$2"
        ;;
    build)
        build
        ;;
    update)
        update
        ;;
    clean)
        clean
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        error "Comando desconhecido: $1"
        echo ""
        show_help
        exit 1
        ;;
esac
