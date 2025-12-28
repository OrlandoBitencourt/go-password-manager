# Telegram Bot Quick Start Guide

This guide will help you set up and use the Telegram Bot frontend for the password manager.

## Prerequisites

- Go 1.25+ installed
- A Telegram account
- Access to create Telegram bots via BotFather

## Step 1: Create a Telegram Bot

1. Open Telegram and search for **@BotFather**
2. Start a conversation and send the command:
   ```
   /newbot
   ```
3. Follow the prompts:
   - Choose a name for your bot (e.g., "My Password Manager")
   - Choose a username (must end with 'bot', e.g., "mypasswordmanager_bot")
4. BotFather will provide you with a **bot token**. It looks like:
   ```
   123456789:ABCdefGHIjklMNOpqrsTUVwxyz
   ```
5. **Save this token securely** - you'll need it to run the bot

## Step 2: Get Your Telegram User ID (Optional but Recommended)

To restrict bot access to only yourself:

1. Search for **@userinfobot** on Telegram
2. Start the bot - it will automatically show your user ID
3. Copy your user ID (e.g., `123456789`)

## Step 3: Configure Environment

Create a `.env` file in the project root:

```bash
# Required
TELEGRAM_BOT_TOKEN=your_bot_token_here

# Optional - Restrict access to specific users (comma-separated)
ALLOWED_USER_IDS=123456789,987654321

# Optional - Custom settings
SESSION_TTL=5m
EPHEMERAL_MESSAGE_TTL=60s
VAULT_DIR=./vaults
```

## Step 4: Run the Bot

### Option A: Using Docker Compose (Recommended)

```bash
# Start both the backend and telegram bot
docker-compose up -d

# View logs
docker-compose logs -f telegram-bot

# Stop
docker-compose down
```

### Option B: Run Directly with Go

```bash
# Set environment variables
export TELEGRAM_BOT_TOKEN="your_token_here"
export VAULT_DIR="./vaults"

# Run the bot
go run cmd/telegram-bot/main.go
```

### Option C: Build and Run Binary

```bash
# Build
go build -o telegram-bot ./cmd/telegram-bot

# Run (Windows)
set TELEGRAM_BOT_TOKEN=your_token_here
telegram-bot.exe

# Run (Linux/Mac)
export TELEGRAM_BOT_TOKEN=your_token_here
./telegram-bot
```

## Step 5: Create a Vault (One-Time Setup)

Before using the Telegram bot, you need to create a vault. You can do this via:

### Option 1: Using the Web Interface

1. Start the HTTP server:
   ```bash
   go run cmd/server/main.go
   ```
2. Open `http://localhost:8080` in your browser
3. Click "Create Vault"
4. Enter a vault name (e.g., "personal") and a strong master password
5. Click Create

### Option 2: Using the HTTP API

```bash
curl -X POST http://localhost:8080/api/vaults/create \
  -H "Content-Type: application/json" \
  -d '{
    "name": "personal",
    "master_password": "YourStrongPassword123!"
  }'
```

## Step 6: Use the Telegram Bot

1. **Find your bot** on Telegram (search for the username you created)

2. **Start the bot**:
   ```
   /start
   ```
   You'll see helpful buttons: **🔑 Login**, **📋 List Vaults**, and **❓ Help**

3. **Login to your vault**:
   - Click the **🔑 Login** button
   - Enter vault name: `personal`
   - Enter master password: (your password)
   - After successful login, you'll see **📋 List Passwords** and **🚪 Logout** buttons

4. **List passwords**:
   - Click **📋 List Passwords** button
   - You'll see a clickable button for each password (e.g., **🔑 github**)

5. **Get a password** (auto-deletes after 60 seconds):
   - Click the password button (e.g., **🔑 github**), or
   - Type: `/get github`

6. **Add a password**:
   ```
   /add github myusername mypassword123
   ```

7. **Logout**:
   - Click **🚪 Logout** button, or
   - Type: `/logout`

## Available Commands

| Command | Description | Example |
|---------|-------------|---------|
| `/start` | Show welcome message | `/start` |
| `/help` | Show help | `/help` |
| `/login` | Login to a vault | `/login` |
| `/logout` | Logout from vault | `/logout` |
| `/vaults` | List available vaults | `/vaults` |
| `/list` | List password records | `/list` |
| `/get <name>` | Get password (ephemeral) | `/get github` |
| `/add <name> <user> <pass>` | Add password | `/add gitlab user pass` |

## Security Best Practices

### ✅ DO

- **Use strong master passwords** (12+ characters, mixed case, numbers, symbols)
- **Configure ALLOWED_USER_IDS** in production to restrict access
- **Use only in private chats** with the bot (never in groups)
- **Logout when finished** using `/logout`
- **Keep your bot token secret** - don't commit it to version control

### ❌ DON'T

- **Don't screenshot password messages** - defeats the ephemeral feature
- **Don't use in Telegram groups** - passwords could be visible to others
- **Don't share your bot token** - anyone with it can impersonate your bot
- **Don't store highly sensitive passwords** (banking, critical systems)
- **Don't share your master password** - the bot can't recover it

## How It Works

### Ephemeral Messages

When you retrieve a password using `/get`, the bot:
1. Sends the password in a message
2. Schedules the message for deletion in 60 seconds
3. Automatically deletes the message after the timeout

**Note**: This is a best-effort approach. Users can still:
- Screenshot the message before deletion
- See the message in notifications
- Find it in Telegram backups

### Master Password Protection

When you login with `/login`, the bot:
1. Asks for your vault name
2. Prompts for your master password
3. **Immediately deletes both your password message and the password prompt** after processing
4. This ensures your master password doesn't remain in chat history

**Important**: Even though messages are deleted, Telegram notifications and screenshots could still capture them. Always use the bot in a secure environment.

### Session Management

- Sessions expire after **5 minutes of inactivity**
- You must `/login` again after expiry
- Each session is tied to your Telegram user ID
- Only one vault can be unlocked per user at a time

### Rate Limiting

The bot implements two types of rate limiting:

1. **General Rate Limit**: Max 10 requests per minute
2. **Password Retrieval Limit**: Max 5 password retrievals per minute

This prevents brute-force attacks and abuse.

## Troubleshooting

### "You are not authorized to use this bot"

- Check that your user ID is in `ALLOWED_USER_IDS`
- Make sure the environment variable is set correctly

### "Vault not found"

- Make sure you created the vault first (see Step 5)
- Check that `VAULT_DIR` points to the correct directory
- Verify the vault file exists in the vaults directory

### "Invalid master password"

- Double-check your master password
- Passwords are case-sensitive
- If you forgot it, there's no recovery method (by design)

### Bot doesn't respond

- Check that the bot is running (`docker-compose ps` or check process)
- Verify your bot token is correct
- Check logs: `docker-compose logs telegram-bot`
- Ensure your Telegram account isn't banned/restricted

### Messages not being deleted

- This is a best-effort feature using Telegram's API
- Check bot logs for deletion errors
- Verify the bot has permission to delete messages
- Some Telegram clients may cache messages temporarily

## Configuration Reference

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `TELEGRAM_BOT_TOKEN` | *required* | Bot token from BotFather |
| `ALLOWED_USER_IDS` | (empty) | Comma-separated user IDs, empty = allow all |
| `VAULT_DIR` | `./vaults` | Directory for vault storage |
| `SESSION_TTL` | `5m` | Session expiry (e.g., `5m`, `1h`) |
| `EPHEMERAL_MESSAGE_TTL` | `60s` | Password message auto-delete time |
| `RATE_LIMIT_REQUESTS` | `10` | Max requests per window |
| `RATE_LIMIT_WINDOW` | `1m` | Rate limit time window |
| `PASSWORD_RETRIEVAL_MAX` | `5` | Max password retrievals per window |
| `PASSWORD_RETRIEVAL_WINDOW` | `1m` | Password retrieval window |

## Advanced Usage

### Using Multiple Vaults

You can create multiple vaults for different purposes:

```bash
# Create work vault
curl -X POST http://localhost:8080/api/vaults/create \
  -H "Content-Type: application/json" \
  -d '{"name":"work","master_password":"WorkPass123!"}'

# Create personal vault
curl -X POST http://localhost:8080/api/vaults/create \
  -H "Content-Type: application/json" \
  -d '{"name":"personal","master_password":"PersonalPass123!"}'
```

Then in Telegram:
```
/login
> work
> WorkPass123!

/logout
/login
> personal
> PersonalPass123!
```

### Allowing Multiple Users

To allow your team to use the bot:

1. Each team member gets their user ID from @userinfobot
2. Update `.env`:
   ```
   ALLOWED_USER_IDS=123456789,987654321,555555555
   ```
3. Restart the bot
4. Each user logs in with their own vault credentials

**Note**: Users cannot share vaults - each user needs their own vault.

## What's Next?

- Read [ADR-0002](ADR-0002-telegram-bot-frontend.md) for architectural details
- Check the main [README.md](README.md) for HTTP API documentation
- Explore the web interface at `http://localhost:8080`
- Consider implementing 2FA for additional security

## Support

If you encounter issues:

1. Check the logs: `docker-compose logs telegram-bot`
2. Review this guide's troubleshooting section
3. Check [GitHub Issues](https://github.com/yourusername/go-password-manager/issues)
4. Read the [ADR documents](./ADR-0002-telegram-bot-frontend.md)

---

**Remember**: This bot is designed for convenience with reasonable security. For critical passwords (banking, email recovery, etc.), consider using a dedicated password manager like Bitwarden or 1Password.
