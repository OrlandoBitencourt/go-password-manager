package telegram

import (
	"log"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// EphemeralMessage represents a message scheduled for deletion
type EphemeralMessage struct {
	ChatID    int64
	MessageID int
	DeleteAt  time.Time
}

// EphemeralMessageManager manages auto-deletion of messages containing secrets
type EphemeralMessageManager struct {
	bot      *tgbotapi.BotAPI
	messages map[string]*EphemeralMessage
	mu       sync.Mutex
	ttl      time.Duration
	ticker   *time.Ticker
	done     chan bool
}

// NewEphemeralMessageManager creates a new ephemeral message manager
func NewEphemeralMessageManager(bot *tgbotapi.BotAPI, ttl time.Duration) *EphemeralMessageManager {
	emm := &EphemeralMessageManager{
		bot:      bot,
		messages: make(map[string]*EphemeralMessage),
		ttl:      ttl,
		ticker:   time.NewTicker(5 * time.Second),
		done:     make(chan bool),
	}

	go emm.cleanupLoop()
	return emm
}

// ScheduleDelete schedules a message for automatic deletion
func (emm *EphemeralMessageManager) ScheduleDelete(chatID int64, messageID int) {
	emm.mu.Lock()
	defer emm.mu.Unlock()

	key := emm.getKey(chatID, messageID)
	emm.messages[key] = &EphemeralMessage{
		ChatID:    chatID,
		MessageID: messageID,
		DeleteAt:  time.Now().Add(emm.ttl),
	}

	log.Printf("Scheduled message %d in chat %d for deletion in %v", messageID, chatID, emm.ttl)
}

// cleanupLoop periodically checks and deletes expired messages
func (emm *EphemeralMessageManager) cleanupLoop() {
	for {
		select {
		case <-emm.ticker.C:
			emm.deleteExpiredMessages()
		case <-emm.done:
			return
		}
	}
}

// deleteExpiredMessages deletes all messages that have passed their TTL
func (emm *EphemeralMessageManager) deleteExpiredMessages() {
	emm.mu.Lock()
	now := time.Now()
	toDelete := make([]*EphemeralMessage, 0)

	for key, msg := range emm.messages {
		if now.After(msg.DeleteAt) {
			toDelete = append(toDelete, msg)
			delete(emm.messages, key)
		}
	}
	emm.mu.Unlock()

	// Delete messages outside the lock
	for _, msg := range toDelete {
		deleteMsg := tgbotapi.NewDeleteMessage(msg.ChatID, msg.MessageID)
		if _, err := emm.bot.Request(deleteMsg); err != nil {
			log.Printf("Failed to delete message %d in chat %d: %v", msg.MessageID, msg.ChatID, err)
		} else {
			log.Printf("Deleted ephemeral message %d in chat %d", msg.MessageID, msg.ChatID)
		}
	}
}

// getKey generates a unique key for a message
func (emm *EphemeralMessageManager) getKey(chatID int64, messageID int) string {
	return string(chatID) + ":" + string(messageID)
}

// Stop stops the cleanup loop
func (emm *EphemeralMessageManager) Stop() {
	emm.ticker.Stop()
	emm.done <- true
}
