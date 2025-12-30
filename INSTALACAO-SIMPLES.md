# 🚀 Guia de Instalação Simplificado

## ⚠️ IMPORTANTE: Use o Terminal do Sistema

**NÃO use o terminal integrado do VSCode!** O VSCode está rodando via Flatpak e não tem acesso ao sistema.

## 📝 Passo a Passo

### 1. Abrir Terminal do Sistema

Há 3 formas de abrir um terminal real:

**Opção A:** Pressione `Ctrl + Alt + T`

**Opção B:** Clique no botão de aplicativos e procure por "Terminal"

**Opção C:** Clique com botão direito na área de trabalho → "Abrir Terminal Aqui"

### 2. Instalar Docker

No terminal que abriu, copie e cole estes comandos **um por vez**:

```bash
# Ir para o diretório do projeto
cd ~/Documentos/GitHub/go-password-manager

# Executar script de instalação
./install-docker.sh
```

O script vai pedir sua senha (`sudo`) e vai instalar tudo automaticamente.

### 3. Fazer Logout e Login

**Muito importante!** Após a instalação terminar:

1. Feche todas as janelas
2. Clique no seu nome no canto superior direito
3. Selecione "Sair" ou "Logout"
4. Faça login novamente

Ou simplesmente reinicie o computador.

### 4. Verificar se Funcionou

Abra um terminal novamente (`Ctrl + Alt + T`) e execute:

```bash
docker --version
```

Deve aparecer algo como: `Docker version 24.0.x`

### 5. Configurar o Projeto

```bash
# Ir para o projeto
cd ~/Documentos/GitHub/go-password-manager

# Editar arquivo de configuração
nano .env
```

No arquivo `.env`, configure pelo menos:
- `TELEGRAM_BOT_TOKEN=seu_token_aqui`

Salve com `Ctrl + O`, Enter, depois `Ctrl + X`

### 6. Instalar e Iniciar

```bash
# Instalar serviço (inicia automaticamente no boot)
./manage.sh install

# Iniciar agora
./manage.sh start
```

### 7. Verificar se Está Funcionando

```bash
# Ver status
./manage.sh status

# Ver containers rodando
docker ps
```

Deve mostrar 3 containers rodando:
- password-manager-backend
- password-manager-telegram-bot
- password-manager-frontend

### 8. Acessar os Serviços

- **Frontend (Web):** Abra o navegador em http://localhost:3000
- **API Backend:** http://localhost:8080
- **Telegram Bot:** Abra o Telegram e converse com seu bot

## ❓ Problemas Comuns

### "comando não encontrado" ou "sudo não encontrado"

Você ainda está no terminal do VSCode. Feche e use `Ctrl + Alt + T`.

### "Permission denied"

Execute com sudo:
```bash
sudo ./install-docker.sh
```

### Depois de instalar, docker não funciona

Você precisa fazer logout e login novamente para as permissões serem aplicadas.

### Containers não iniciam

Verifique os logs:
```bash
./manage.sh logs
```

## 🎯 Comandos Úteis

```bash
# Ver status
./manage.sh status

# Ver logs
./manage.sh logs

# Reiniciar tudo
./manage.sh restart

# Parar tudo
./manage.sh stop

# Desinstalar serviço
./manage.sh uninstall
```

## 💡 Dica

Salve este comando para facilitar:

```bash
# Criar alias permanente
echo 'alias pm="cd ~/Documentos/GitHub/go-password-manager && ./manage.sh"' >> ~/.bashrc
source ~/.bashrc

# Agora você pode usar de qualquer lugar:
pm status
pm logs
pm restart
```

---

**Precisa de ajuda?** Veja a documentação completa em [DOCKER_SETUP.md](DOCKER_SETUP.md)
