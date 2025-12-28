# 🚀 Quick Start Guide

Guia rápido para executar o Password Manager com todas as interfaces (Web, API e Telegram Bot).

## 📋 Pré-requisitos

- Go 1.23+
- Node.js 18+
- Docker & Docker Compose (opcional)
- Telegram Bot Token (para o bot)

## 🎯 Opção 1: Executar Tudo Localmente (Desenvolvimento)

### 1. Backend (Go API Server)

```bash
# Terminal 1
go run cmd/server/main.go
```

✅ Backend rodando em: http://localhost:8080

### 2. Frontend (Next.js Web UI)

```bash
# Terminal 2
cd frontend
npm install
npm run dev
```

✅ Frontend rodando em: http://localhost:3001

### 3. Telegram Bot (Opcional)

```bash
# Terminal 3
# Primeiro, configure seu .env com o token do bot
go run cmd/telegram-bot/main.go
```

✅ Bot Telegram ativo e esperando mensagens

## 🐳 Opção 2: Docker Compose (Produção)

### Passo 1: Configure o .env

Copie o arquivo de exemplo:
```bash
cp .env.example .env
```

Edite o `.env` e adicione seu token do Telegram:
```bash
TELEGRAM_BOT_TOKEN=your_actual_bot_token_here
```

### Passo 2: Execute com Docker Compose

**Todos os serviços:**
```bash
docker-compose up -d
```

**Ou serviços individuais:**
```bash
# Apenas backend
docker-compose up -d backend

# Apenas frontend
docker-compose up -d frontend

# Apenas Telegram bot
docker-compose up -d telegram-bot
```

### Passo 3: Acesse

- 🌐 **Web UI**: http://localhost:3000
- 🔌 **API**: http://localhost:8080
- 💬 **Telegram**: Busque seu bot no Telegram

## 📱 Configurando o Telegram Bot

### 1. Criar o Bot

1. Abra o Telegram e busque por `@BotFather`
2. Envie `/newbot`
3. Siga as instruções:
   - Nome do bot: `My Password Manager`
   - Username: `my_password_manager_bot` (deve terminar com `_bot`)
4. Copie o token fornecido

### 2. Adicionar o Token

Edite o arquivo `.env`:
```bash
TELEGRAM_BOT_TOKEN=1234567890:ABCdefGHIjklMNOpqrsTUVwxyz
```

### 3. (Opcional) Restringir Acesso

Para permitir apenas usuários específicos:

1. Descubra seu Telegram User ID:
   - Busque `@userinfobot` no Telegram
   - Envie qualquer mensagem
   - Copie o ID fornecido

2. Adicione no `.env`:
```bash
ALLOWED_USER_IDS=123456789,987654321
```

### 4. Iniciar o Bot

```bash
# Se rodando localmente
go run cmd/telegram-bot/main.go

# Se usando Docker
docker-compose up -d telegram-bot
```

### 5. Usar o Bot

1. Busque seu bot no Telegram pelo username
2. Envie `/start`
3. Use os comandos:

```
/login          - Fazer login em um vault
/add            - Adicionar nova senha
/get <nome>     - Buscar uma senha
/list           - Listar todas as senhas
/vaults         - Ver vaults disponíveis
/logout         - Sair do vault
/help           - Ajuda
```

## 🎮 Exemplos de Uso

### Web UI

1. Acesse http://localhost:3000 (ou 3001 em dev)
2. Crie um vault ou faça unlock
3. Adicione senhas
4. Use o gerador de senhas
5. Busque e copie senhas

### API (cURL)

```bash
# Criar vault
curl -X POST http://localhost:8080/api/vaults/create \
  -H "Content-Type: application/json" \
  -d '{"name":"personal","master_password":"SecurePass123!"}'

# Unlock vault
curl -X POST http://localhost:8080/api/vaults/unlock \
  -H "Content-Type: application/json" \
  -d '{"name":"personal","master_password":"SecurePass123!"}'

# Adicionar senha
curl -X POST http://localhost:8080/api/records/add \
  -H "Content-Type: application/json" \
  -d '{
    "vault_name":"personal",
    "name":"GitHub",
    "username":"myusername",
    "password":"mypassword123"
  }'

# Buscar senha
curl "http://localhost:8080/api/records/get?vault_name=personal&name=GitHub"
```

### Telegram Bot

```
Você: /start
Bot: 👋 Welcome to Password Manager!

Você: /login
Bot: 🔐 Please enter your vault name:
Você: personal
Bot: 🔑 Please enter your master password:
Você: [sua senha]
Bot: ✅ Successfully unlocked vault "personal"

Você: /add github myusername mypassword123
Bot: ✅ Password "github" added successfully!

Você: /get github
Bot: 🔐 Password for "github":
     Username: myusername
     Password: mypassword123
     [Esta mensagem será deletada em 60 segundos]

Você: /list
Bot: 📋 Your passwords:
     • github
     • gmail
     • twitter
```

## 🔧 Troubleshooting

### Backend não inicia

```bash
# Verifique se a porta 8080 está livre
netstat -ano | findstr :8080

# Ou use outra porta
PORT=8081 go run cmd/server/main.go
```

### Frontend não conecta ao backend

Verifique o arquivo `frontend/.env.local`:
```bash
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### Telegram Bot não responde

1. Verifique se o token está correto no `.env`
2. Verifique os logs do bot:
```bash
docker-compose logs telegram-bot
```

3. Teste se o bot está ativo:
```bash
curl https://api.telegram.org/bot<YOUR_TOKEN>/getMe
```

### Erros de CORS

O backend já tem CORS habilitado. Se ainda tiver problemas:
- Verifique se o backend está rodando
- Teste diretamente a API com cURL
- Verifique o console do navegador

## 📊 Logs

### Ver logs de todos os serviços (Docker)
```bash
docker-compose logs -f
```

### Ver logs de um serviço específico
```bash
docker-compose logs -f backend
docker-compose logs -f frontend
docker-compose logs -f telegram-bot
```

## 🛑 Parar os Serviços

### Docker Compose
```bash
# Parar todos
docker-compose down

# Parar e remover volumes
docker-compose down -v
```

### Localmente
Pressione `Ctrl+C` em cada terminal

## 🔒 Segurança

### Produção

1. **Use HTTPS**: Configure um reverse proxy (nginx, Caddy)
2. **Senha forte**: Master password com 16+ caracteres
3. **Backup**: Faça backup da pasta `vaults/`
4. **Restrinja Telegram**: Use `ALLOWED_USER_IDS`
5. **Firewall**: Proteja as portas 8080 e 3000

### Backup de Vaults

```bash
# Criar backup
tar -czf vaults-backup-$(date +%Y%m%d).tar.gz vaults/

# Restaurar backup
tar -xzf vaults-backup-20231227.tar.gz
```

## 📚 Recursos Adicionais

- [README.md](README.md) - Documentação completa
- [ADR-0001](ADR-0001-password-manager.md) - Decisões de arquitetura
- [ADR-0002](ADR-0002-telegram-bot-frontend.md) - Telegram bot design
- [FRONTEND_UPGRADE.md](FRONTEND_UPGRADE.md) - Melhorias do frontend
- [TELEGRAM_BOT_GUIDE.md](TELEGRAM_BOT_GUIDE.md) - Guia detalhado do bot

## 🎉 Tudo Pronto!

Agora você tem um password manager completo com:
- ✅ Backend seguro em Go
- ✅ Frontend moderno em Next.js
- ✅ Telegram Bot para acesso móvel
- ✅ Docker para deploy fácil
- ✅ Dark mode
- ✅ Gerador de senhas
- ✅ Busca em tempo real

Aproveite! 🔐
