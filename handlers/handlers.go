package handlers

import (
	"blue/api"
	"blue/commands"
	"blue/commands/admin"
	"blue/commands/ctf"
	_ "blue/commands/group"
	_ "blue/commands/user"
	"blue/config"
	"blue/database"
	"blue/models"
	"fmt"
	"log"
	"strings"
)

type Handler struct {
	bot *api.Bot
	db  *database.DB
	cfg *config.Config
}

func NewHandler(bot *api.Bot, db *database.DB, cfg *config.Config) *Handler {
	admin.SetDatabase(db)
	ctf.SetServices(db, cfg)
	return &Handler{
		bot: bot,
		db:  db,
		cfg: cfg,
	}
}

func (h *Handler) HandleUpdate(update models.Update) {
	if update.CallbackQuery != nil {
		if ctf.HandleCallback(h.bot, update.CallbackQuery) {
			return
		}
		admin.HandleCallback(h.bot, update.CallbackQuery)
		return
	}

	if update.Message == nil {
		return
	}

	msg := update.Message

	if msg.From != nil {
		h.db.UpsertUser(msg.From.ID, msg.From.Username, msg.From.FirstName, msg.From.LastName)
	}

	if msg.Chat.Type == "group" || msg.Chat.Type == "supergroup" {
		h.db.UpsertGroup(msg.Chat.ID, msg.Chat.Title, msg.Chat.Type)
	}

	if h.cfg.LogChatID != 0 && msg.Chat.ID == h.cfg.LogChatID && msg.Text != "" && msg.From != nil {
		h.db.LogMessage(msg.MessageID, msg.From.ID, msg.Chat.ID, msg.Text)
	}

	if msg.Document != nil {
		h.handleDocument(msg)
		return
	}

	if msg.Text == "" {
		return
	}

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
	if at := strings.Index(command, "@"); at >= 0 {
		command = command[:at]
	}
	args := parts[1:]

	if handler, exists := commands.Get(command); exists {
		handler(h.bot, msg, args)
	}
}

func (h *Handler) handleMessage(msg *models.Message) {
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

	response := fmt.Sprintf("```\nFile: %s\n\nLines: %d\nWords: %d\nSize: %s\n```",
		doc.FileName, lines, words, sizeStr)

	h.bot.ReplyToMessage(msg.Chat.ID, msg.MessageID, response)
}
