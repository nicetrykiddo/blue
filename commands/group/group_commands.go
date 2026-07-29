package group

import (
	"blue/api"
	"blue/commands"
	"blue/config"
	"blue/database"
	"blue/models"
	"fmt"
	"html"
	"log"
	"strconv"
	"strings"
)

var cfg *config.Config
var db *database.DB

func SetServices(config *config.Config, database *database.DB) {
	cfg = config
	db = database
}

func init() {
	commands.Register("/info", infoHandler)
	commands.Register("/id", idHandler)
}

func infoHandler(bot *api.Bot, msg *models.Message, args []string) {
	sendInfo(bot, msg, formatChat(msg.Chat, msg.MessageThreadID))
}

func idHandler(bot *api.Bot, msg *models.Message, args []string) {
	if msg.ReplyToMessage != nil && msg.ReplyToMessage.From != nil {
		sendInfo(bot, msg, formatUser(msg.ReplyToMessage.From))
		return
	}

	if len(args) == 0 || !isAdmin(msg.From) {
		sendInfo(bot, msg, formatChat(msg.Chat, msg.MessageThreadID))
		return
	}
	if len(args) != 1 {
		sendInfo(bot, msg, "use <code>/id &lt;chat_or_user_id&gt;</code>")
		return
	}

	id, err := strconv.ParseInt(strings.TrimSpace(args[0]), 10, 64)
	if err != nil {
		sendInfo(bot, msg, "that is not a valid numeric Telegram ID")
		return
	}

	chat, err := bot.GetChat(id)
	if err != nil {
		if id > 0 && db != nil {
			if user, dbErr := db.GetUser(id); dbErr == nil {
				sendInfo(bot, msg, formatStoredUser(user))
				return
			}
		}
		log.Printf("Error looking up Telegram ID %d: %v", id, err)
		sendInfo(bot, msg, "Telegram could not access that chat or user. They may not have interacted with this bot, or the bot may no longer be in that chat.")
		return
	}
	sendInfo(bot, msg, formatChat(chat, 0))
}

func formatStoredUser(user *database.UserInfo) string {
	name := strings.TrimSpace(strings.Join([]string{user.FirstName, user.LastName}, " "))
	if name == "" {
		name = "-"
	}
	username := "-"
	if user.Username != "" {
		username = "@" + user.Username
	}
	return fmt.Sprintf(
		"<b>user info</b>\nname: <b>%s</b>\nusername: <code>%s</code>\nuser id: <code>%d</code>\nbot: <code>unknown</code>",
		html.EscapeString(name),
		html.EscapeString(username),
		user.ID,
	)
}

func isAdmin(user *models.User) bool {
	return cfg != nil && user != nil && cfg.AdminUserIDs[user.ID]
}

func formatUser(user *models.User) string {
	name := strings.TrimSpace(strings.Join([]string{user.FirstName, user.LastName}, " "))
	if name == "" {
		name = "-"
	}
	username := "-"
	if user.Username != "" {
		username = "@" + user.Username
	}
	return fmt.Sprintf(
		"<b>user info</b>\nname: <b>%s</b>\nusername: <code>%s</code>\nuser id: <code>%d</code>\nbot: <code>%t</code>",
		html.EscapeString(name),
		html.EscapeString(username),
		user.ID,
		user.IsBot,
	)
}

func formatChat(chat *models.Chat, topicID int) string {
	if chat == nil {
		return "this update has no chat information"
	}
	name := strings.TrimSpace(strings.Join([]string{chat.FirstName, chat.LastName}, " "))
	if chat.Title != "" {
		name = chat.Title
	}
	if name == "" {
		name = "-"
	}
	username := ""
	if chat.Username != "" {
		username = fmt.Sprintf("\nusername: <code>@%s</code>", html.EscapeString(chat.Username))
	}
	topic := ""
	if topicID != 0 {
		topic = fmt.Sprintf("\ntopic id: <code>%d</code>", topicID)
	}
	return fmt.Sprintf(
		"<b>chat info</b>\nname: <b>%s</b>\ntype: <code>%s</code>\nchat id: <code>%d</code>%s%s",
		html.EscapeString(name),
		html.EscapeString(chat.Type),
		chat.ID,
		username,
		topic,
	)
}

func sendInfo(bot *api.Bot, msg *models.Message, text string) {
	_ = bot.SendChatAction(msg.Chat.ID, "typing")
	if _, err := bot.SendMessageWithOptions(api.SendMessageOptions{
		ChatID:                msg.Chat.ID,
		MessageThreadID:       msg.MessageThreadID,
		ReplyToMessageID:      msg.MessageID,
		Text:                  text,
		ParseMode:             "HTML",
		DisableWebPagePreview: true,
	}); err != nil {
		log.Printf("Error sending ID information: %v", err)
	}
}
