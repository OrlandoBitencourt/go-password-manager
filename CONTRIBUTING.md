# Contributing to Go Password Manager

Thank you for your interest in contributing! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Coding Standards](#coding-standards)

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow
- Maintain a professional environment

## Getting Started

1. **Fork the repository**
2. **Clone your fork**:
   ```bash
   git clone https://github.com/yourusername/go-password-manager.git
   cd go-password-manager
   ```
3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/orlan/go-password-manager.git
   ```
4. **Install dependencies**:
   ```bash
   go mod download
   ```

## Development Setup

### Prerequisites

- Go 1.25 or later
- Git
- Docker (optional, for testing containerization)
- A Telegram bot token (for testing bot features)

### Environment Setup

Create a `.env` file for local development:

```bash
cp .env.example .env
# Edit .env with your configuration
```

### Running Locally

```bash
# Run HTTP server
go run cmd/server/main.go

# Run Telegram bot
export TELEGRAM_BOT_TOKEN="your_token"
go run cmd/telegram-bot/main.go
```

## Project Structure

```
.
├── cmd/
│   ├── server/              # HTTP API server
│   └── telegram-bot/        # Telegram bot service
├── internal/
│   ├── application/         # Business logic
│   ├── crypto/              # Encryption services
│   ├── domain/              # Domain models & interfaces
│   ├── telegram/            # Telegram bot implementation
│   ├── transport/http/      # HTTP handlers
│   └── vault/               # Vault repository
├── web/                     # Web frontend
├── ADR-*.md                 # Architecture Decision Records
└── README.md
```

### Architecture Layers

1. **Domain Layer** (`internal/domain/`)
   - Core entities (Vault, PasswordRecord)
   - Repository interfaces
   - Domain errors

2. **Application Layer** (`internal/application/`)
   - VaultService (business logic)
   - Session management
   - Orchestration

3. **Infrastructure Layer** (`internal/crypto/`, `internal/vault/`)
   - Cryptography implementation
   - File-based repository
   - External dependencies

4. **Transport Layer** (`internal/transport/`, `internal/telegram/`)
   - HTTP handlers
   - Telegram bot handlers
   - Protocol adapters

## Making Changes

### Creating a Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

Branch naming:
- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation updates
- `refactor/` - Code refactoring
- `test/` - Test additions/updates

### Commit Messages

Follow conventional commits:

```
type(scope): subject

body (optional)

footer (optional)
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Formatting
- `refactor`: Code restructuring
- `test`: Tests
- `chore`: Maintenance

Examples:
```
feat(telegram): add password update command

fix(crypto): correct nonce size validation

docs(readme): update Telegram bot setup instructions
```

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/crypto/

# Run with verbose output
go test -v ./...
```

### Writing Tests

- Write tests for new features
- Maintain or improve coverage
- Test edge cases and error paths
- Use table-driven tests where appropriate

Example:

```go
func TestEncrypt(t *testing.T) {
    tests := []struct {
        name    string
        input   []byte
        wantErr bool
    }{
        {"valid input", []byte("test"), false},
        {"empty input", []byte(""), false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

## Submitting Changes

### Pull Request Process

1. **Update your fork**:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Push your branch**:
   ```bash
   git push origin feature/your-feature-name
   ```

3. **Create Pull Request**:
   - Go to GitHub and create a PR
   - Fill out the PR template
   - Link related issues
   - Request review

### PR Checklist

- [ ] Code follows project style guidelines
- [ ] Tests pass (`go test ./...`)
- [ ] New features have tests
- [ ] Documentation updated (if needed)
- [ ] No merge conflicts
- [ ] Commit messages are clear
- [ ] ADR created (for architectural changes)

### PR Description Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
How was this tested?

## Checklist
- [ ] Tests pass
- [ ] Documentation updated
- [ ] ADR created (if needed)
```

## Coding Standards

### Go Style

Follow official Go guidelines:
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

### Specific Guidelines

#### Naming

```go
// Good
func CreateVault(name string) error
type VaultRepository interface{}
var ErrVaultNotFound = errors.New("vault not found")

// Avoid
func create_vault(name string) error
type vaultRepo interface{}
var ERR_VAULT_NOT_FOUND = errors.New("vault not found")
```

#### Error Handling

```go
// Good
if err != nil {
    return fmt.Errorf("failed to create vault: %w", err)
}

// Avoid
if err != nil {
    return err // loses context
}
```

#### Comments

```go
// VaultService handles vault operations and session management.
// It provides thread-safe access to encrypted vaults.
type VaultService struct {
    // ...
}

// CreateVault creates a new encrypted vault with the given name
// and master password. It returns an error if the vault already exists.
func (s *VaultService) CreateVault(ctx context.Context, name, password string) error {
    // ...
}
```

#### Package Organization

- Keep packages focused and cohesive
- Avoid circular dependencies
- Use internal/ for non-exported packages
- Export only what's necessary

### Security Considerations

When contributing security-related code:

1. **Never log sensitive data**
   ```go
   // Bad
   log.Printf("Password: %s", password)

   // Good
   log.Printf("Processing password for vault: %s", vaultName)
   ```

2. **Use crypto/rand for randomness**
   ```go
   // Good
   salt := make([]byte, 32)
   _, err := rand.Read(salt)

   // Bad - Don't use math/rand for security
   ```

3. **Validate input**
   ```go
   if len(password) == 0 {
       return errors.New("password cannot be empty")
   }
   ```

4. **Handle errors properly**
   ```go
   // Check for specific errors
   if err == domain.ErrInvalidMasterPassword {
       // Handle authentication failure
   }
   ```

## Documentation

### When to Update Documentation

Update docs when:
- Adding new features
- Changing APIs
- Modifying configuration
- Updating dependencies
- Making architectural changes

### Documentation Files

- **README.md** - Main project documentation
- **TELEGRAM_BOT_GUIDE.md** - Telegram bot setup/usage
- **ADR-*.md** - Architectural decisions
- **CONTRIBUTING.md** - This file
- **Code comments** - Inline documentation

### ADR Process

For significant architectural changes, create an ADR:

1. Copy template from existing ADRs
2. Number sequentially (ADR-0003, etc.)
3. Include:
   - Status (Proposed, Accepted, Deprecated)
   - Context
   - Decision
   - Consequences
4. Submit with your PR

## Areas for Contribution

### High Priority

- [ ] Add unit tests (increase coverage)
- [ ] Implement password generator
- [ ] Add password strength meter
- [ ] Session timeout enforcement
- [ ] Import/export functionality

### Medium Priority

- [ ] Browser extension
- [ ] Mobile app
- [ ] Additional Telegram commands
- [ ] Web UI improvements
- [ ] Performance optimizations

### Documentation

- [ ] Video tutorials
- [ ] API examples
- [ ] Deployment guides
- [ ] Security audit documentation

## Getting Help

- **Questions?** Open a GitHub Discussion
- **Bugs?** Create an Issue
- **Features?** Start with a Discussion, then create an Issue
- **Security?** Email security@example.com (do not create public issue)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to Go Password Manager! 🎉
