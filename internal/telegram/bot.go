package telegram

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/orlan/go-password-manager/internal/application"
)

// Bot represents the Telegram bot service
type Bot struct {
	api               *tgbotapi.BotAPI
	vaultService      *application.VaultService
	sessionManager    *SessionManager
	ephemeralManager  *EphemeralMessageManager
	rateLimiter       *RateLimiter
	passwordRetrieval *RateLimiter
	config            *Config
}

// Config holds bot configuration
type Config struct {
	BotToken              string
	SessionTTL            time.Duration
	EphemeralMessageTTL   time.Duration
	RateLimitRequests     int
	RateLimitWindow       time.Duration
	PasswordRetrievalMax  int
	PasswordRetrievalWin  time.Duration
	AllowedUserIDs        []int64
}

// NewBot creates a new Telegram bot instance
func NewBot(config *Config, vaultService *application.VaultService) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(config.BotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	api.Debug = false
	log.Printf("Authorized on Telegram bot account: %s", api.Self.UserName)

	bot := &Bot{
		api:               api,
		vaultService:      vaultService,
		sessionManager:    NewSessionManager(config.SessionTTL),
		ephemeralManager:  NewEphemeralMessageManager(api, config.EphemeralMessageTTL),
		rateLimiter:       NewRateLimiter(config.RateLimitRequests, config.RateLimitWindow),
		passwordRetrieval: NewRateLimiter(config.PasswordRetrievalMax, config.PasswordRetrievalWin),
		config:            config,
	}

	return bot, nil
}

// Start begins processing updates from Telegram
func (b *Bot) Start(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	log.Println("Telegram bot started. Waiting for messages...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down Telegram bot...")
			b.Stop()
			return nil
		case update := <-updates:
			// Handle callback queries (button clicks)
			if update.CallbackQuery != nil {
				go b.handleCallbackQuery(update.CallbackQuery)
				continue
			}

			// Handle regular messages
			if update.Message == nil {
				continue
			}

			go b.handleUpdate(update)
		}
	}
}

// handleUpdate processes incoming messages
func (b *Bot) handleUpdate(update tgbotapi.Update) {
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	// Check if user is allowed (if allowlist is configured)
	if len(b.config.AllowedUserIDs) > 0 {
		allowed := false
		for _, id := range b.config.AllowedUserIDs {
			if id == userID {
				allowed = true
				break
			}
		}
		if !allowed {
			b.sendMessage(chatID, "⛔ You are not authorized to use this bot.")
			return
		}
	}

	// Rate limiting
	if !b.rateLimiter.Allow(userID) {
		b.sendMessage(chatID, "⏱️ Too many requests. Please slow down.")
		return
	}

	// Check for login flow
	state, pendingVault := b.sessionManager.GetLoginState(userID)

	if state == StateAwaitingVaultName {
		b.handleVaultNameInput(userID, chatID, update.Message.Text)
		return
	}

	if state == StateAwaitingMasterPassword {
		// Delete the user's password message immediately
		deleteMsg := tgbotapi.NewDeleteMessage(chatID, update.Message.MessageID)
		b.api.Request(deleteMsg)

		b.handleMasterPasswordInput(userID, chatID, pendingVault, update.Message.Text)
		return
	}

	// Handle commands
	if update.Message.IsCommand() {
		b.handleCommand(userID, chatID, update.Message)
	} else {
		b.sendMessage(chatID, "Please use /help to see available commands.")
	}
}

// handleCommand routes commands to appropriate handlers
func (b *Bot) handleCommand(userID, chatID int64, message *tgbotapi.Message) {
	command := message.Command()
	args := message.CommandArguments()

	switch command {
	case "start":
		b.handleStart(chatID)
	case "help":
		b.handleHelp(chatID)
	case "login":
		b.handleLogin(userID, chatID)
	case "logout":
		b.handleLogout(userID, chatID)
	case "list":
		b.handleList(userID, chatID)
	case "get":
		b.handleGet(userID, chatID, args)
	case "add":
		b.handleAdd(userID, chatID, args)
	case "vaults":
		b.handleVaults(chatID)
	default:
		b.sendMessage(chatID, "Unknown command. Use /help to see available commands.")
	}
}

// handleStart welcomes the user
func (b *Bot) handleStart(chatID int64) {
	message := `🔐 *Password Manager Bot*

Welcome! I'm a secure password manager bot that helps you manage your encrypted password vaults.

*Getting Started:*
1️⃣ Use Login button to authenticate with your vault
2️⃣ Use List to see your passwords
3️⃣ Use Logout when you're done

⚠️ *Security Notice:*
• Passwords are automatically deleted from chat after 60 seconds
• Sessions expire after 5 minutes of inactivity
• Never share your master password
• Use this bot in private chats only`

	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"

	// Add inline keyboard with main actions
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔑 Login", "cmd_login"),
			tgbotapi.NewInlineKeyboardButtonData("📋 List Vaults", "cmd_vaults"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❓ Help", "cmd_help"),
		),
	)
	msg.ReplyMarkup = keyboard

	b.api.Send(msg)
}

// handleHelp shows available commands
func (b *Bot) handleHelp(chatID int64) {
	message := `📚 *Available Commands:*

*Authentication:*
/login - Sign into your vault
/logout - Sign out of your vault

*Password Management:*
/get <name> - Retrieve a password (auto-deletes)
/list - List all password records
/add <name> <username> <password> - Add new password

*Other:*
/vaults - List available vaults
/help - Show this help message

💡 *Tips:*
• Passwords sent via /get are automatically deleted after 60 seconds
• Your session expires after 5 minutes of inactivity
• Always /logout when you're done`

	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"

	// Add quick action buttons
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔑 Login", "cmd_login"),
			tgbotapi.NewInlineKeyboardButtonData("📋 List Vaults", "cmd_vaults"),
		),
	)
	msg.ReplyMarkup = keyboard

	b.api.Send(msg)
}

// handleLogin starts the login flow
func (b *Bot) handleLogin(userID, chatID int64) {
	if b.sessionManager.IsAuthenticated(userID) {
		session, _ := b.sessionManager.GetSession(userID)
		b.sendMessage(chatID, fmt.Sprintf("✅ You're already logged into vault: *%s*\n\nUse /logout to sign out first.", session.VaultName))
		return
	}

	b.sessionManager.SetLoginState(userID, StateAwaitingVaultName, "")
	b.sendMessage(chatID, "🔑 Please enter the vault name:")
}

// handleVaultNameInput processes vault name during login
func (b *Bot) handleVaultNameInput(userID, chatID int64, vaultName string) {
	vaultName = strings.TrimSpace(vaultName)

	if vaultName == "" {
		b.sendMessage(chatID, "❌ Vault name cannot be empty. Please try again:")
		return
	}

	// Check if vault exists
	ctx := context.Background()
	vaults, err := b.vaultService.ListVaults(ctx)
	if err != nil {
		b.sendMessage(chatID, "❌ Error checking vaults. Please try /login again.")
		b.sessionManager.SetLoginState(userID, StateIdle, "")
		return
	}

	found := false
	for _, v := range vaults {
		if v == vaultName {
			found = true
			break
		}
	}

	if !found {
		b.sendMessage(chatID, fmt.Sprintf("❌ Vault '%s' not found. Please try /login again.", vaultName))
		b.sessionManager.SetLoginState(userID, StateIdle, "")
		return
	}

	b.sessionManager.SetLoginState(userID, StateAwaitingMasterPassword, vaultName)

	// Send password prompt and store its message ID for later deletion
	msg := tgbotapi.NewMessage(chatID, "🔐 Please enter your master password:")
	msg.ParseMode = "Markdown"
	sent, err := b.api.Send(msg)
	if err == nil {
		b.sessionManager.SetPasswordPromptMsgID(userID, sent.MessageID)
	}
}

// handleMasterPasswordInput processes master password during login
func (b *Bot) handleMasterPasswordInput(userID, chatID int64, vaultName, masterPassword string) {
	// Delete the password prompt message
	promptMsgID := b.sessionManager.GetAndClearPasswordPromptMsgID(userID)
	if promptMsgID != 0 {
		deletePrompt := tgbotapi.NewDeleteMessage(chatID, promptMsgID)
		b.api.Request(deletePrompt)
	}

	ctx := context.Background()
	err := b.vaultService.UnlockVault(ctx, vaultName, masterPassword)

	if err != nil {
		b.sendMessage(chatID, "❌ Invalid master password or vault error. Please try /login again.")
		b.sessionManager.SetLoginState(userID, StateIdle, "")
		return
	}

	// Create session
	b.sessionManager.CreateSession(userID, vaultName)
	b.sessionManager.SetLoginState(userID, StateIdle, "")

	// Send success message with action buttons
	message := fmt.Sprintf("✅ Successfully logged into vault: *%s*\n\nWhat would you like to do?", vaultName)
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"

	// Add quick action buttons
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 List Passwords", "cmd_list"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🚪 Logout", "cmd_logout"),
		),
	)
	msg.ReplyMarkup = keyboard

	b.api.Send(msg)
}

// handleLogout signs out the user
func (b *Bot) handleLogout(userID, chatID int64) {
	if !b.sessionManager.IsAuthenticated(userID) {
		b.sendMessage(chatID, "ℹ️ You're not logged in.")
		return
	}

	session, _ := b.sessionManager.GetSession(userID)
	vaultName := session.VaultName

	// Lock vault in backend
	ctx := context.Background()
	b.vaultService.LockVault(ctx, vaultName)

	// Delete session
	b.sessionManager.DeleteSession(userID)

	// Send logout confirmation with login options
	message := "👋 You've been logged out successfully."
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"

	// Add login button for easy re-login
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔑 Login Again", "cmd_login"),
			tgbotapi.NewInlineKeyboardButtonData("📋 List Vaults", "cmd_vaults"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❓ Help", "cmd_help"),
		),
	)
	msg.ReplyMarkup = keyboard

	b.api.Send(msg)
}

// handleList lists all password records
func (b *Bot) handleList(userID, chatID int64) {
	if !b.sessionManager.IsAuthenticated(userID) {
		// Send message with login button
		message := "🔒 You need to login first to view passwords."
		msg := tgbotapi.NewMessage(chatID, message)
		msg.ParseMode = "Markdown"

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔑 Login", "cmd_login"),
				tgbotapi.NewInlineKeyboardButtonData("📋 List Vaults", "cmd_vaults"),
			),
		)
		msg.ReplyMarkup = keyboard
		b.api.Send(msg)
		return
	}

	session, _ := b.sessionManager.GetSession(userID)
	b.sessionManager.UpdateActivity(userID)

	ctx := context.Background()
	records, err := b.vaultService.ListPasswordRecords(ctx, session.VaultName)

	if err != nil {
		b.sendMessage(chatID, "❌ Error retrieving records.")
		return
	}

	if len(records) == 0 {
		b.sendMessage(chatID, "📭 No password records found in this vault.")
		return
	}

	message := "📋 *Password Records:*\n\n"
	for i, record := range records {
		message += fmt.Sprintf("%d. *%s*\n   └ Username: `%s`\n", i+1, record.Name, record.Username)
	}
	message += "\n💡 Click a button below or use `/get <name>`"

	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"

	// Create inline keyboard with buttons for each record
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, record := range records {
		button := tgbotapi.NewInlineKeyboardButtonData(
			"🔑 "+record.Name,
			"get_"+record.Name,
		)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))
	}

	// Add logout button at the end
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🚪 Logout", "cmd_logout"),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg.ReplyMarkup = keyboard

	b.api.Send(msg)
}

// handleGet retrieves a password (ephemeral)
func (b *Bot) handleGet(userID, chatID int64, args string) {
	if !b.sessionManager.IsAuthenticated(userID) {
		// Send message with login button
		message := "🔒 You need to login first to retrieve passwords."
		msg := tgbotapi.NewMessage(chatID, message)
		msg.ParseMode = "Markdown"

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔑 Login", "cmd_login"),
			),
		)
		msg.ReplyMarkup = keyboard
		b.api.Send(msg)
		return
	}

	// Rate limit password retrievals
	if !b.passwordRetrieval.Allow(userID) {
		b.sendMessage(chatID, "⏱️ Too many password retrievals. Please wait before trying again.")
		return
	}

	recordName := strings.TrimSpace(args)
	if recordName == "" {
		b.sendMessage(chatID, "❌ Usage: /get <record_name>")
		return
	}

	session, _ := b.sessionManager.GetSession(userID)
	b.sessionManager.UpdateActivity(userID)

	ctx := context.Background()
	record, err := b.vaultService.GetPasswordRecord(ctx, session.VaultName, recordName)

	if err != nil {
		// Send error with helpful action button
		message := fmt.Sprintf("❌ Password record '%s' not found.\n\nWould you like to see all available passwords?", recordName)
		msg := tgbotapi.NewMessage(chatID, message)
		msg.ParseMode = "Markdown"

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📋 List All Passwords", "cmd_list"),
			),
		)
		msg.ReplyMarkup = keyboard
		b.api.Send(msg)
		return
	}

	// Send ephemeral password message
	message := fmt.Sprintf("🔑 *Password for: %s*\n\n"+
		"Username: `%s`\n"+
		"Password: `%s`\n\n"+
		"⚠️ This message will be deleted in 60 seconds.",
		record.Name, record.Username, record.Password)

	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"

	sent, err := b.api.Send(msg)
	if err != nil {
		log.Printf("Failed to send password message: %v", err)
		return
	}

	// Schedule for deletion
	b.ephemeralManager.ScheduleDelete(chatID, sent.MessageID)

	// Send follow-up menu with next actions
	b.sendActionMenu(chatID, "What would you like to do next?")
}

// handleAdd adds a new password record
func (b *Bot) handleAdd(userID, chatID int64, args string) {
	if !b.sessionManager.IsAuthenticated(userID) {
		b.sendMessage(chatID, "🔒 Please /login first.")
		return
	}

	parts := strings.Fields(args)
	if len(parts) < 3 {
		b.sendMessage(chatID, "❌ Usage: /add <name> <username> <password>")
		return
	}

	name := parts[0]
	username := parts[1]
	password := parts[2]

	session, _ := b.sessionManager.GetSession(userID)
	b.sessionManager.UpdateActivity(userID)

	ctx := context.Background()
	err := b.vaultService.AddPasswordRecord(ctx, session.VaultName, name, username, password)

	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("❌ Failed to add password: %s", err.Error()))
		return
	}

	b.sendMessage(chatID, fmt.Sprintf("✅ Password record '%s' added successfully!", name))

	// Send action menu
	b.sendActionMenu(chatID, "What would you like to do next?")
}

// handleVaults lists all available vaults
func (b *Bot) handleVaults(chatID int64) {
	ctx := context.Background()
	vaults, err := b.vaultService.ListVaults(ctx)

	if err != nil {
		b.sendMessage(chatID, "❌ Error retrieving vaults.")
		return
	}

	if len(vaults) == 0 {
		b.sendMessage(chatID, "📭 No vaults available.")
		return
	}

	message := "🗄️ *Available Vaults:*\n\n"
	for i, vault := range vaults {
		message += fmt.Sprintf("%d. `%s`\n", i+1, vault)
	}
	message += "\n💡 Click the button below to login."

	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"

	// Add login button
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔑 Login to Vault", "cmd_login"),
		),
	)
	msg.ReplyMarkup = keyboard

	b.api.Send(msg)
}

// handleCallbackQuery processes button clicks
func (b *Bot) handleCallbackQuery(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID
	data := query.Data

	// Answer the callback query immediately (removes loading state)
	callback := tgbotapi.NewCallback(query.ID, "")
	b.api.Request(callback)

	// Check if user is allowed (if allowlist is configured)
	if len(b.config.AllowedUserIDs) > 0 {
		allowed := false
		for _, id := range b.config.AllowedUserIDs {
			if id == userID {
				allowed = true
				break
			}
		}
		if !allowed {
			b.sendMessage(chatID, "⛔ You are not authorized to use this bot.")
			return
		}
	}

	// Rate limiting
	if !b.rateLimiter.Allow(userID) {
		b.sendMessage(chatID, "⏱️ Too many requests. Please slow down.")
		return
	}

	// Handle different callback commands
	switch {
	case data == "cmd_login":
		b.handleLogin(userID, chatID)
	case data == "cmd_logout":
		b.handleLogout(userID, chatID)
	case data == "cmd_help":
		b.handleHelp(chatID)
	case data == "cmd_vaults":
		b.handleVaults(chatID)
	case data == "cmd_list":
		b.handleList(userID, chatID)
	case strings.HasPrefix(data, "get_"):
		// Extract record name from callback data
		recordName := strings.TrimPrefix(data, "get_")
		b.handleGet(userID, chatID, recordName)
	default:
		b.sendMessage(chatID, "❌ Unknown action.")
	}
}

// sendMessage sends a text message
func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	b.api.Send(msg)
}

// sendActionMenu sends a menu with common action buttons
func (b *Bot) sendActionMenu(chatID int64, promptText string) {
	msg := tgbotapi.NewMessage(chatID, promptText)
	msg.ParseMode = "Markdown"

	// Create action menu with common options
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 List Passwords", "cmd_list"),
			tgbotapi.NewInlineKeyboardButtonData("🗄️ Vaults", "cmd_vaults"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❓ Help", "cmd_help"),
			tgbotapi.NewInlineKeyboardButtonData("🚪 Logout", "cmd_logout"),
		),
	)
	msg.ReplyMarkup = keyboard

	b.api.Send(msg)
}

// Stop gracefully stops the bot
func (b *Bot) Stop() {
	b.sessionManager.Stop()
	b.ephemeralManager.Stop()
	b.api.StopReceivingUpdates()
}
