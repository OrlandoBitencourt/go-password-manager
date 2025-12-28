# Go Password Manager

A secure, encrypted password manager with **multiple frontends** (HTTP API, Web UI, and Telegram Bot) built in Go.

**Challenge source:** https://codingchallenges.fyi/challenges/challenge-password-manager

## Features

### Core Features
- ✅ **Encrypted Vault Storage** - AES-256-GCM encryption with Argon2id key derivation
- ✅ **Multiple Vaults** - Support for multiple isolated password vaults
- ✅ **CRUD Operations** - Create, read, update, and delete password records
- ✅ **Session Management** - Secure session handling with auto-expiry
- ✅ **File-based Storage** - Encrypted vault files (`.vault` format)

### Frontends
1. **HTTP API** - RESTful API for programmatic access
2. **Web UI** - Browser-based interface for desktop use
3. **Telegram Bot** - Mobile-friendly bot with ephemeral password delivery

## Architecture

The project follows a clean architecture pattern with multiple frontends:

```
├── cmd/
│   ├── server/          # HTTP API & Web server
│   └── telegram-bot/    # Telegram bot service
├── internal/
│   ├── domain/          # Domain models and interfaces
│   ├── application/     # Business logic (VaultService)
│   ├── crypto/          # Encryption service (AES-256-GCM + Argon2id)
│   ├── vault/           # File repository implementation
│   ├── transport/http/  # HTTP handlers
│   └── telegram/        # Telegram bot implementation
└── web/                 # Web frontend static files
```

## Security

### Cryptography

- **Encryption Algorithm**: AES-256-GCM (Galois/Counter Mode)
  - 256-bit keys for maximum security
  - Authenticated encryption prevents tampering
  - Unique nonce for each encryption operation

- **Key Derivation Function**: Argon2id
  - Memory-hard algorithm resistant to GPU attacks
  - Parameters: 1 iteration, 64MB memory, 4 threads
  - Unique salt per vault (32 bytes)

### Security Features

- Master password never stored on disk
- Encryption keys held in memory only during active sessions
- Vault files are fully encrypted (only metadata is unencrypted)
- No sensitive data logged
- HTTPS recommended for production deployments

### Telegram Bot Security & Features
- **Ephemeral Messages**: Passwords auto-delete after 60 seconds
- **Master Password Protection**: Login credentials are immediately deleted from chat
- **Session Expiry**: Sessions expire after 5 minutes of inactivity
- **Rate Limiting**: Prevents brute-force attempts
- **User Allowlist**: Optional restriction to specific Telegram user IDs
- **Password Retrieval Limits**: Separate rate limit for password access
- **Inline Buttons**: Easy-to-use button interface for common actions

## Getting Started

### Prerequisites

- Go 1.23 or later
- Docker and Docker Compose (optional)

### Installation

1. Clone the repository:
```bash
git clone https://github.com/orlan/go-password-manager.git
cd go-password-manager
```

2. Install dependencies:
```bash
go mod download
```

### Running Locally

**Option 1: Direct Go execution**
```bash
go run cmd/server/main.go
```

**Option 2: Build and run**
```bash
go build -o password-manager cmd/server/main.go
./password-manager
```

The server will start on `http://localhost:8080`

### Running with Docker

**Option 1: Docker Compose (Recommended)**
```bash
docker-compose up --build
```

**Option 2: Docker only**
```bash
docker build -t password-manager .
docker run -p 8080:8080 -v $(pwd)/vaults:/root/vaults password-manager
```

### Configuration

Environment variables:

#### HTTP Server
- `PORT`: Server port (default: `8080`)
- `VAULT_DIR`: Directory for vault files (default: `./vaults`)
- `WEB_DIR`: Directory for web frontend (default: `./web`)

#### Telegram Bot
- `TELEGRAM_BOT_TOKEN`: Bot token from BotFather (required)
- `ALLOWED_USER_IDS`: Comma-separated Telegram user IDs (optional, empty = allow all)
- `SESSION_TTL`: Session expiry duration (default: `5m`)
- `EPHEMERAL_MESSAGE_TTL`: Auto-delete time for password messages (default: `60s`)
- `RATE_LIMIT_REQUESTS`: Max requests per window (default: `10`)
- `RATE_LIMIT_WINDOW`: Rate limit time window (default: `1m`)
- `PASSWORD_RETRIEVAL_MAX`: Max password retrievals per window (default: `5`)
- `PASSWORD_RETRIEVAL_WINDOW`: Password retrieval window (default: `1m`)

## Usage

### Web Interface

1. Open your browser to `http://localhost:8080`
2. Create a new vault with a name and master password
3. Unlock the vault with your master password
4. Add, view, and manage password records

### Telegram Bot

#### Setup

1. **Get a Telegram Bot Token**
   - Open Telegram and search for [@BotFather](https://t.me/botfather)
   - Send `/newbot` and follow the prompts
   - Copy the bot token provided
   - Add it to your `.env` file: `TELEGRAM_BOT_TOKEN=your_token_here`

2. **Optional: Restrict Access**
   - Search for [@userinfobot](https://t.me/userinfobot) on Telegram
   - Get your Telegram user ID
   - Add to `.env`: `ALLOWED_USER_IDS=your_user_id,another_user_id`

3. **Start the Bot**
   ```bash
   # Using Docker Compose (recommended)
   docker-compose up -d telegram-bot

   # Or run directly
   export TELEGRAM_BOT_TOKEN="your_token"
   go run cmd/telegram-bot/main.go
   ```

#### Using the Bot

1. **Start conversation**
   ```
   /start
   ```
   You'll see buttons for Login, List Vaults, and Help

2. **Login to vault**
   - Click the **🔑 Login** button (or use `/login`)
   - Bot will ask for vault name
   - Then for master password
   - After login, you'll see buttons for List Passwords and Logout

3. **List password records**
   - Click **📋 List Passwords** button (or use `/list`)
   - You'll see a button for each password record
   - Click any **🔑 Record Name** button to retrieve that password

4. **Retrieve a password** (auto-deletes after 60s)
   - Click the password button from the list, or
   - Use command: `/get github`

5. **Add a new password**
   ```
   /add github myusername mypassword123
   ```

6. **List available vaults**
   - Click **📋 List Vaults** button (or use `/vaults`)

7. **Logout**
   - Click **🚪 Logout** button (or use `/logout`)

#### Available Commands

| Command | Description |
|---------|-------------|
| `/start` | Welcome message and introduction |
| `/help` | Show available commands |
| `/login` | Authenticate with a vault |
| `/logout` | End your session |
| `/list` | List all password records (no passwords shown) |
| `/get <name>` | Retrieve password (ephemeral - auto-deletes in 60s) |
| `/add <name> <username> <password>` | Add new password record |
| `/vaults` | List all available vaults |

#### Security Notes

⚠️ **Important**:
- **Master passwords are immediately deleted** from chat after login
- Passwords sent via `/get` are automatically deleted after 60 seconds
- Password prompt messages are also deleted to prevent re-reading
- Users can still screenshot messages before deletion
- **Use only in private chats, never in groups**
- Always configure `ALLOWED_USER_IDS` in production
- Never share your master password

### API Endpoints

#### Vault Management

**List all vaults**
```bash
GET /api/vaults
```

**Create a new vault**
```bash
POST /api/vaults/create
Content-Type: application/json

{
  "name": "my-vault",
  "master_password": "your-secure-password"
}
```

**Unlock a vault**
```bash
POST /api/vaults/unlock
Content-Type: application/json

{
  "name": "my-vault",
  "master_password": "your-secure-password"
}
```

**Lock a vault**
```bash
POST /api/vaults/lock
Content-Type: application/json

{
  "name": "my-vault"
}
```

#### Password Record Management

**List all records in a vault**
```bash
GET /api/records?vault_name=my-vault
```

**Add a password record**
```bash
POST /api/records/add
Content-Type: application/json

{
  "vault_name": "my-vault",
  "name": "GitHub",
  "username": "john_doe",
  "password": "secret123"
}
```

**Get a specific password record**
```bash
GET /api/records/get?vault_name=my-vault&name=GitHub
```

**Update a password record**
```bash
PUT /api/records/update
Content-Type: application/json

{
  "vault_name": "my-vault",
  "name": "GitHub",
  "username": "new_username",
  "password": "new_password"
}
```

**Delete a password record**
```bash
DELETE /api/records/delete
Content-Type: application/json

{
  "vault_name": "my-vault",
  "name": "GitHub"
}
```

### Example: Using cURL

```bash
# Create a vault
curl -X POST http://localhost:8080/api/vaults/create \
  -H "Content-Type: application/json" \
  -d '{"name":"personal","master_password":"MySecurePass123!"}'

# Unlock the vault
curl -X POST http://localhost:8080/api/vaults/unlock \
  -H "Content-Type: application/json" \
  -d '{"name":"personal","master_password":"MySecurePass123!"}'

# Add a password
curl -X POST http://localhost:8080/api/records/add \
  -H "Content-Type: application/json" \
  -d '{"vault_name":"personal","name":"Gmail","username":"john@example.com","password":"gmail123"}'

# Retrieve a password
curl "http://localhost:8080/api/records/get?vault_name=personal&name=Gmail"

# List all passwords
curl "http://localhost:8080/api/records?vault_name=personal"
```

## Testing

Run the test suite:
```bash
go test ./...
```

Run with coverage:
```bash
go test -cover ./...
```

## Development

### Project Structure

- **Domain Layer** ([internal/domain/](internal/domain/)): Core entities, interfaces, and domain errors
- **Application Layer** ([internal/application/](internal/application/)): Use cases and business logic
- **Crypto Layer** ([internal/crypto/](internal/crypto/)): Encryption and key derivation
- **Vault Layer** ([internal/vault/](internal/vault/)): File-based vault persistence
- **Transport Layer** ([internal/transport/http/](internal/transport/http/)): HTTP handlers and routing
- **Web Frontend** ([web/](web/)): HTML/CSS/JavaScript web interface

### Adding New Features

The modular architecture makes it easy to extend:

1. Add new domain entities in `internal/domain/`
2. Implement business logic in `internal/application/`
3. Create HTTP endpoints in `internal/transport/http/`
4. Update web UI in `web/`

## Implementation Details

### Vault File Format

Each vault is stored as a `.vault` file containing JSON:

```json
{
  "version": "1.0",
  "salt": "<base64-encoded-salt>",
  "nonce": "<base64-encoded-nonce>",
  "encrypted": "<base64-encoded-ciphertext>"
}
```

The encrypted payload contains the actual vault data with all password records.

### Session Management

- Vaults must be explicitly unlocked before accessing records
- Unlocked vaults are held in memory with their encryption keys
- Call the lock endpoint to clear the vault from memory
- Future enhancement: auto-lock after timeout

## Limitations & Future Enhancements

### Current Limitations

- No cloud synchronization
- Single-user vaults only
- No password strength analysis
- No automatic session timeout

### Planned Features

- Password generator
- Password strength meter
- Import/export functionality (CSV, JSON)
- Browser extension
- Inline keyboard for Telegram bot
- Support for shared vaults
- Two-factor authentication (2FA)
- Biometric unlock for mobile
- Encrypted notes/files
- Password history tracking

## Security Considerations

1. **Master Password**: Choose a strong, unique master password
2. **HTTPS**: Use HTTPS in production to protect API traffic
3. **Backups**: Regularly backup your vault files
4. **Access Control**: Restrict filesystem access to vault directory
5. **Memory**: Vault data is unencrypted in memory while unlocked

## Architecture Decision Records

For detailed architectural decisions, see:
- [ADR-0001: Password Manager Core](ADR-0001-password-manager.md)
- [ADR-0002: Telegram Bot Frontend](ADR-0002-telegram-bot-frontend.md)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built as part of the [Coding Challenges](https://codingchallenges.fyi/) series
- Inspired by KeePass and 1Password
- Uses industry-standard cryptography (AES-256-GCM, Argon2id)
