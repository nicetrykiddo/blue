package handlers

import (
	"blue/api"
	"blue/cache"
	"blue/commands"
	_ "blue/commands/checker"
	_ "blue/commands/group"
	_ "blue/commands/user"
	"blue/models"
	"fmt"
	"log"
	"strings"
)

type Handler struct {
	bot          *api.Bot
	stickerCache *cache.StickerCache
}

func NewHandler(bot *api.Bot, stickerCache *cache.StickerCache) *Handler {
	return &Handler{
		bot:          bot,
		stickerCache: stickerCache,
	}
}

func (h *Handler) HandleUpdate(update models.Update) {
	if update.Message == nil {
		return
	}

	msg := update.Message

	if msg.Document != nil {
		h.handleDocument(msg)
		return
	}

	if msg.Text == "" {
		return
	}

	// log.Printf("Received message from %s: %s", msg.MessageID, msg.Text)

	if strings.HasPrefix(msg.Text, "/") {
		h.handleCommand(msg)
	} else {
		h.handleMessage(msg)
	}
}

func (h *Handler) handleCommand(msg *models.Message) {
	parts := strings.Fields(msg.Text)
	if len(parts) == 0 {
		return
	}

	command := strings.ToLower(parts[0])
	args := parts[1:]

	if handler, exists := commands.Get(command); exists {
		handler(h.bot, msg, args, h.stickerCache)
	}
}

func (h *Handler) handleMessage(msg *models.Message) {
	// Optional: Disable default echo to be more "professional" or keep it consistent
	// For now, let's just ignore non-commands or provide a minimal response if needed.
	// But the user code had an echo. Let's make it consistent.
	// response := fmt.Sprintf("<code>[+] Message Received\n| Len: %d</code>", len(msg.Text))
	// if _, err := h.bot.ReplyToMessage(msg.Chat.ID, msg.MessageID, response); err != nil {
	// 	log.Printf("Error sending message: %v", err)
	// }
}

func (h *Handler) handleDocument(msg *models.Message) {
	doc := msg.Document

	if doc.MimeType != "text/plain" && !strings.HasSuffix(doc.FileName, ".txt") {
		// h.bot.ReplyToMessage(msg.Chat.ID, msg.MessageID, "<code>[!] Error: Only .txt files supported.</code>")
		return
	}

	h.bot.SendChatAction(msg.Chat.ID, "typing")

	file, err := h.bot.GetFile(doc.FileID)
	if err != nil {
		log.Printf("Error getting file: %v", err)
		// h.bot.ReplyToMessage(msg.Chat.ID, msg.MessageID, "<code>[!] Error: File retrieval failed.</code>")
		return
	}

	content, err := h.bot.DownloadFile(file.FilePath)
	if err != nil {
		log.Printf("Error downloading file: %v", err)
		// h.bot.ReplyToMessage(msg.Chat.ID, msg.MessageID, "Download failed 😕")
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

	response := fmt.Sprintf("📄 %s\n\n📊 Stats:\n• Lines: %d\n• Words: %d\n• Size: %s",
		doc.FileName, lines, words, sizeStr)

	h.bot.ReplyToMessage(msg.Chat.ID, msg.MessageID, response)
}
