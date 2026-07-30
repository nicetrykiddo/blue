package handlers

import (
	"blue/api"
	"blue/commands"
	"blue/commands/admin"
	"blue/commands/ctf"
	"blue/commands/group"
	"blue/commands/user"
	"blue/config"
	"blue/database"
	"blue/models"
	"fmt"
	"html"
	"log"
	"math/rand"
	"strings"
	"sync"
)

type Handler struct {
	bot                  *api.Bot
	db                   *database.DB
	cfg                  *config.Config
	reactionMu           sync.Mutex
	customReactionEmojis []string
	botUserID            int64
	botUsername          string
}

func NewHandler(bot *api.Bot, db *database.DB, cfg *config.Config) *Handler {
	admin.SetDatabase(db)
	ctf.SetServices(db, cfg)
	group.SetServices(cfg, db)
	user.SetConfig(cfg)
	handler := &Handler{
		bot: bot,
		db:  db,
		cfg: cfg,
	}
	if ids, err := db.ListReactionEmojis(); err != nil {
		log.Printf("Error loading approved reaction emojis: %v", err)
	} else {
		handler.customReactionEmojis = ids
	}
	if identity, err := bot.GetMe(); err != nil {
		log.Printf("Could not load bot identity for mention reactions: %v", err)
	} else {
		handler.botUserID = identity.ID
		handler.botUsername = identity.Username
	}
	commands.Register("/allowreaction", handler.allowReactionHandler)
	commands.Register("/removereaction", handler.removeReactionHandler)
	commands.Register("/clearreactions", handler.clearReactionsHandler)
	return handler
}

func (h *Handler) HandleUpdate(update models.Update) {
	if update.CallbackQuery != nil {
		if ctf.HandleCallback(h.bot, update.CallbackQuery) {
			return
		}
		admin.HandleCallback(h.bot, update.CallbackQuery)
		return
	}

	if update.ChannelPost != nil {
		if update.ChannelPost.Text != "" && !strings.HasPrefix(update.ChannelPost.Text, "/") {
			h.maybeReactToAdminMessage(update.ChannelPost)
		}
		return
	}

	if update.Message == nil {
		return
	}

	msg := update.Message

	if msg.Document != nil {
		h.trackActivity(msg)
		h.handleDocument(msg)
		return
	}

	if msg.Text == "" {
		return
	}

	if strings.HasPrefix(msg.Text, "/") {
		h.trackActivity(msg)
		h.handleCommand(msg)
		return
	}

	if ctf.HandleMessage(h.bot, msg) {
		h.trackActivity(msg)
		return
	}

	h.maybeReactToAdminMessage(msg)

	if h.cfg.LogChatID != 0 && msg.Chat.ID == h.cfg.LogChatID {
		h.trackActivity(msg)
	}
}

func (h *Handler) maybeReactToAdminMessage(msg *models.Message) {
	if !shouldReactToAdmin(msg, h.cfg, h.botUserID, h.botUsername) {
		return
	}
	customEmojiID := h.customReactionEmoji(msg)
	if customEmojiID == "" {
		return
	}
	if err := h.bot.SetCustomMessageReaction(msg.Chat.ID, msg.MessageID, customEmojiID); err != nil {
		log.Printf("Error reacting to admin message: %v", err)
	}
}

func (h *Handler) customReactionEmoji(msg *models.Message) string {
	h.reactionMu.Lock()
	defer h.reactionMu.Unlock()
	if len(h.customReactionEmojis) == 0 {
		return ""
	}
	value := uint64(msg.MessageID) ^ uint64(msg.Chat.ID)
	if msg.From != nil {
		value ^= uint64(msg.From.ID)
	}
	return h.customReactionEmojis[(value/8)%uint64(len(h.customReactionEmojis))]
}

func (h *Handler) allowReactionHandler(bot *api.Bot, msg *models.Message, args []string) {
	if !h.isConfiguredAdmin(msg.From) {
		h.replyCommand(msg, "Only a configured admin can approve reactions.")
		return
	}
	ids := repliedCustomEmojiIDs(msg)
	if len(ids) == 0 {
		h.replyCommand(msg, "Reply to a message containing the custom emoji you want to approve.")
		return
	}
	for _, id := range ids {
		if err := h.db.AddReactionEmoji(id, msg.From.ID); err != nil {
			log.Printf("Error approving reaction emoji: %v", err)
			h.replyCommand(msg, "Could not save that reaction emoji.")
			return
		}
		h.addReactionEmoji(id)
	}
	h.replyCommand(msg, fmt.Sprintf("Approved %d custom reaction emoji(s).", len(ids)))
}

func (h *Handler) removeReactionHandler(bot *api.Bot, msg *models.Message, args []string) {
	if !h.isConfiguredAdmin(msg.From) {
		h.replyCommand(msg, "Only a configured admin can remove reactions.")
		return
	}
	ids := repliedCustomEmojiIDs(msg)
	if len(ids) == 0 {
		h.replyCommand(msg, "Reply to a message containing the custom emoji you want to remove.")
		return
	}
	for _, id := range ids {
		if err := h.db.RemoveReactionEmoji(id); err != nil {
			log.Printf("Error removing reaction emoji: %v", err)
			h.replyCommand(msg, "Could not remove that reaction emoji.")
			return
		}
		h.removeReactionEmoji(id)
	}
	h.replyCommand(msg, fmt.Sprintf("Removed %d custom reaction emoji(s).", len(ids)))
}

func (h *Handler) clearReactionsHandler(bot *api.Bot, msg *models.Message, args []string) {
	if !h.isConfiguredAdmin(msg.From) {
		h.replyCommand(msg, "Only a configured admin can clear reactions.")
		return
	}
	if err := h.db.ClearReactionEmojis(); err != nil {
		log.Printf("Error clearing reaction emojis: %v", err)
		h.replyCommand(msg, "Could not clear reaction emojis.")
		return
	}
	h.reactionMu.Lock()
	h.customReactionEmojis = nil
	h.reactionMu.Unlock()
	h.replyCommand(msg, "Cleared all approved reaction emojis.")
}

func (h *Handler) isConfiguredAdmin(user *models.User) bool {
	return h.cfg != nil && user != nil && h.cfg.AdminUserIDs[user.ID]
}

func (h *Handler) addReactionEmoji(id string) {
	h.reactionMu.Lock()
	defer h.reactionMu.Unlock()
	for _, existing := range h.customReactionEmojis {
		if existing == id {
			return
		}
	}
	h.customReactionEmojis = append(h.customReactionEmojis, id)
}

func (h *Handler) removeReactionEmoji(id string) {
	h.reactionMu.Lock()
	defer h.reactionMu.Unlock()
	for i, existing := range h.customReactionEmojis {
		if existing == id {
			h.customReactionEmojis = append(h.customReactionEmojis[:i], h.customReactionEmojis[i+1:]...)
			return
		}
	}
}

func repliedCustomEmojiIDs(msg *models.Message) []string {
	if msg == nil || msg.ReplyToMessage == nil {
		return nil
	}
	seen := make(map[string]bool)
	var ids []string
	for _, entity := range msg.ReplyToMessage.Entities {
		if entity.Type == "custom_emoji" && entity.CustomEmojiID != "" && !seen[entity.CustomEmojiID] {
			seen[entity.CustomEmojiID] = true
			ids = append(ids, entity.CustomEmojiID)
		}
	}
	return ids
}

func (h *Handler) replyCommand(msg *models.Message, text string) {
	if _, err := h.bot.SendMessageWithOptions(api.SendMessageOptions{
		ChatID:           msg.Chat.ID,
		MessageThreadID:  msg.MessageThreadID,
		ReplyToMessageID: msg.MessageID,
		Text:             text,
	}); err != nil {
		log.Printf("Error sending reaction command reply: %v", err)
	}
}

func shouldReactToAdmin(msg *models.Message, cfg *config.Config, botUserID int64, botUsername string) bool {
	if !eligibleForAdminReaction(msg, cfg) {
		return false
	}
	return rand.Intn(reactionChanceDenominator(msg, botUserID, botUsername)) == 0
}

func eligibleForAdminReaction(msg *models.Message, cfg *config.Config) bool {
	if cfg == nil || msg == nil || msg.Chat == nil {
		return false
	}
	if msg.Text == "" || strings.HasPrefix(msg.Text, "/") {
		return false
	}
	if msg.Chat.Type == "channel" {
		return true
	}
	return msg.From != nil && cfg.AdminUserIDs[msg.From.ID]
}

func reactionChanceDenominator(msg *models.Message, botUserID int64, botUsername string) int {
	if talkingToBot(msg, botUserID, botUsername) {
		return 2
	}
	return 20
}

func talkingToBot(msg *models.Message, botUserID int64, botUsername string) bool {
	if botUserID != 0 && msg.ReplyToMessage != nil && msg.ReplyToMessage.From != nil &&
		msg.ReplyToMessage.From.ID == botUserID {
		return true
	}
	if botUsername == "" {
		return false
	}

	text := strings.ToLower(msg.Text)
	mention := "@" + strings.ToLower(botUsername)
	for start := 0; ; {
		index := strings.Index(text[start:], mention)
		if index < 0 {
			return false
		}
		end := start + index + len(mention)
		if end == len(text) || !isUsernameCharacter(text[end]) {
			return true
		}
		start = end
	}
}

func isUsernameCharacter(char byte) bool {
	return char >= 'a' && char <= 'z' || char >= '0' && char <= '9' || char == '_'
}

func (h *Handler) trackActivity(msg *models.Message) {
	if msg.From == nil {
		return
	}
	if err := h.db.UpsertUser(msg.From.ID, msg.From.Username, msg.From.FirstName, msg.From.LastName); err != nil {
		log.Printf("Error updating user activity: %v", err)
	}
	if msg.Chat.Type == "group" || msg.Chat.Type == "supergroup" {
		if err := h.db.UpsertGroup(msg.Chat.ID, msg.Chat.Title, msg.Chat.Type); err != nil {
			log.Printf("Error updating group activity: %v", err)
		}
	}
}

func (h *Handler) handleCommand(msg *models.Message) {
	parts := strings.Fields(msg.Text)
	if len(parts) == 0 {
		return
	}

	command := strings.ToLower(parts[0])
	if at := strings.Index(command, "@"); at >= 0 {
		command = command[:at]
	}
	args := parts[1:]

	if handler, exists := commands.Get(command); exists {
		handler(h.bot, msg, args)
	}
}

func (h *Handler) handleDocument(msg *models.Message) {
	doc := msg.Document

	if doc.MimeType != "text/plain" && !strings.HasSuffix(doc.FileName, ".txt") {
		return
	}

	h.bot.SendChatAction(msg.Chat.ID, "typing")

	file, err := h.bot.GetFile(doc.FileID)
	if err != nil {
		log.Printf("Error getting file: %v", err)
		return
	}

	content, err := h.bot.DownloadFile(file.FilePath)
	if err != nil {
		log.Printf("Error downloading file: %v", err)
		return
	}

	lines := strings.Count(string(content), "\n") + 1
	words := len(strings.Fields(string(content)))
	size := float64(len(content))

	var sizeStr string
	if size < 1024 {
		sizeStr = fmt.Sprintf("%.0f B", size)
	} else if size < 1024*1024 {
		sizeStr = fmt.Sprintf("%.2f KB", size/1024)
	} else {
		sizeStr = fmt.Sprintf("%.2f MB", size/(1024*1024))
	}

	response := fmt.Sprintf("<b>file</b>\n<pre>name: %s\nlines: %d\nwords: %d\nsize: %s</pre>",
		html.EscapeString(doc.FileName), lines, words, sizeStr)

	if _, err := h.bot.SendMessageWithOptions(api.SendMessageOptions{
		ChatID:                msg.Chat.ID,
		MessageThreadID:       msg.MessageThreadID,
		ReplyToMessageID:      msg.MessageID,
		Text:                  response,
		ParseMode:             "HTML",
		DisableWebPagePreview: true,
	}); err != nil {
		log.Printf("Error sending file summary: %v", err)
	}
}
