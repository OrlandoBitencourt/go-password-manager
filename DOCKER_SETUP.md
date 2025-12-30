# Docker Setup - Password Manager

Este guia mostra como executar o Password Manager (backend, telegram bot e frontend) usando Docker, com inicialização automática no boot do sistema.

## 📋 Pré-requisitos

- Docker instalado e funcionando
- Docker Compose instalado
- Token do bot do Telegram (se for usar o bot)

## 🚀 Instalação Rápida

### 1. Configurar variáveis de ambiente

Primeiro, configure suas variáveis de ambiente:

```bash
# Copie o arquivo de exemplo (se ainda não tiver o .env)
cp .env.example .env

# Edite o arquivo .env com suas configurações
nano .env
```

**Importante:** Configure pelo menos o `TELEGRAM_BOT_TOKEN` no arquivo `.env`.

### 2. Instalar e iniciar automaticamente

Use o script de gerenciamento para instalar o serviço systemd:

```bash
# Dar permissão de execução ao script (apenas uma vez)
chmod +x manage.sh

# Instalar o serviço para iniciar automaticamente com o sistema
./manage.sh install

# Iniciar os serviços agora
./manage.sh start
```

Pronto! Os serviços estão rodando e iniciarão automaticamente quando o sistema reiniciar.

## 🎯 Acessando os Serviços

Após iniciar, você pode acessar:

- **Frontend (Web UI):** http://localhost:3000
- **Backend (API):** http://localhost:8080
- **Telegram Bot:** Use o bot no Telegram (@seu_bot)

## 🔧 Comandos do Script de Gerenciamento

O script `manage.sh` facilita todas as operações:

### Gerenciamento do Serviço Systemd

```bash
# Instalar serviço para iniciar automaticamente
./manage.sh install

# Remover serviço systemd
./manage.sh uninstall
```

### Controle dos Serviços

```bash
# Iniciar todos os serviços
./manage.sh start

# Parar todos os serviços
./manage.sh stop

# Reiniciar todos os serviços
./manage.sh restart

# Ver status dos serviços
./manage.sh status
```

### Logs e Monitoramento

```bash
# Ver logs de todos os serviços
./manage.sh logs

# Ver logs apenas do backend
./manage.sh logs backend

# Ver logs apenas do telegram bot
./manage.sh logs telegram-bot

# Ver logs apenas do frontend
./manage.sh logs frontend
```

### Manutenção

```bash
# Reconstruir as imagens Docker (após mudanças no código)
./manage.sh build

# Atualizar: rebuild + restart
./manage.sh update

# Limpar tudo (containers, imagens, volumes)
./manage.sh clean
```

## 📦 O que foi configurado?

### Docker Compose

Os três serviços foram configurados em [docker-compose.yml](docker-compose.yml):

1. **Backend** (porta 8080)
   - API REST para gerenciamento de senhas
   - Health check configurado
   - Restart automático

2. **Telegram Bot**
   - Integração com Telegram
   - Compartilha o mesmo diretório de vaults com o backend
   - Restart automático
   - Aguarda backend estar saudável antes de iniciar

3. **Frontend** (porta 3000)
   - Interface web Next.js
   - Conecta automaticamente ao backend
   - Restart automático

### Serviço Systemd

O arquivo [password-manager.service](password-manager.service) foi criado para:

- ✅ Iniciar automaticamente no boot do sistema
- ✅ Reiniciar em caso de falha
- ✅ Aguardar o Docker estar pronto
- ✅ Gerenciar todos os 3 serviços juntos

## 🔍 Verificando o Status

### Verificar se está rodando

```bash
# Ver status via systemd
sudo systemctl status password-manager

# Ver containers do Docker
docker ps

# Ver status detalhado
./manage.sh status
```

### Verificar logs

```bash
# Logs do systemd
sudo journalctl -u password-manager -f

# Logs dos containers
./manage.sh logs
```

## 🛠️ Comandos Docker Manuais

Se preferir usar Docker Compose diretamente:

```bash
# Iniciar em segundo plano
docker compose up -d

# Parar
docker compose down

# Ver logs
docker compose logs -f

# Ver status
docker compose ps

# Reconstruir
docker compose build

# Reiniciar um serviço específico
docker compose restart backend
```

## 🔄 Processo de Inicialização

Quando o sistema inicia:

1. Systemd aguarda o Docker estar pronto
2. Systemd executa `docker compose up -d`
3. Docker inicia o **backend** primeiro
4. Backend passa pelo health check
5. **Telegram bot** e **frontend** iniciam após backend estar saudável
6. Todos os serviços ficam rodando em segundo plano

## ⚙️ Configurações Avançadas

### Alterar Portas

Edite o [docker-compose.yml](docker-compose.yml) e mude as portas:

```yaml
ports:
  - "8080:8080"  # Mudar primeira porta: "PORTA_HOST:PORTA_CONTAINER"
```

### Variáveis de Ambiente

Todas as configurações estão no arquivo `.env`:

```bash
# Telegram
TELEGRAM_BOT_TOKEN=seu_token_aqui
ALLOWED_USER_IDS=123456789,987654321

# Sessões
SESSION_TTL=5m
EPHEMERAL_MESSAGE_TTL=60s

# Rate limiting
RATE_LIMIT_REQUESTS=10
RATE_LIMIT_WINDOW=1m
```

### Persistência de Dados

Os dados das senhas são armazenados em:
- **Host:** `./vaults` (no diretório do projeto)
- **Container:** `/root/vaults`

Os dados persistem mesmo quando os containers são recriados.

## 🐛 Troubleshooting

### Serviço não inicia

```bash
# Verificar status do serviço
sudo systemctl status password-manager

# Ver logs do systemd
sudo journalctl -u password-manager -n 50

# Verificar se Docker está rodando
sudo systemctl status docker
```

### Containers com erro

```bash
# Ver logs detalhados
./manage.sh logs

# Verificar containers
docker ps -a

# Recriar containers
./manage.sh stop
./manage.sh build
./manage.sh start
```

### Porta já em uso

Se as portas 8080 ou 3000 já estiverem em uso:

1. Edite [docker-compose.yml](docker-compose.yml)
2. Altere as portas externas (primeira porta no mapeamento)
3. Reconstrua: `./manage.sh update`

### Arquivo .env não encontrado

```bash
# Copiar exemplo
cp .env.example .env

# Editar com suas configurações
nano .env
```

## 🔐 Segurança

### Recomendações

1. ✅ Configure `ALLOWED_USER_IDS` para restringir acesso ao bot
2. ✅ Use HTTPS em produção (configure um reverse proxy como nginx)
3. ✅ Faça backup regular do diretório `./vaults`
4. ✅ Não exponha as portas diretamente na internet sem firewall
5. ✅ Mantenha o token do bot seguro (não commite o arquivo `.env`)

### Backup dos Vaults

```bash
# Criar backup
tar -czf vaults-backup-$(date +%Y%m%d).tar.gz vaults/

# Restaurar backup
tar -xzf vaults-backup-YYYYMMDD.tar.gz
```

## 📚 Próximos Passos

Depois de instalar:

1. 📖 Leia [TELEGRAM_BOT_GUIDE.md](TELEGRAM_BOT_GUIDE.md) para configurar o bot
2. 🌐 Acesse http://localhost:3000 para usar a interface web
3. 🤖 Abra o Telegram e inicie conversa com seu bot
4. 📝 Configure usuários autorizados no `.env`

## 💡 Dicas

- Use `./manage.sh status` regularmente para monitorar
- Configure alertas de monitoramento para produção
- Considere usar Docker secrets para o token do bot em produção
- Documente qualquer customização que fizer

---

**Precisa de ajuda?** Consulte a [documentação principal](README.md) ou abra uma issue no repositório.
