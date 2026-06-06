package group

import (
	"blue/api"
	"blue/commands"
	"blue/models"
	"fmt"
	"html"
	"log"
)

func init() {
	commands.Register("/info", infoHandler)
}

func infoHandler(bot *api.Bot, msg *models.Message, args []string) {
	if msg.Chat.Type != "group" && msg.Chat.Type != "supergroup" {
		return
	}

	text := fmt.Sprintf("<b>chat paperwork</b>\ntitle: <b>%s</b>\ntype: <code>%s</code>\nchat id: <code>%d</code>\ntopic id: <code>%d</code>", html.EscapeString(msg.Chat.Title), msg.Chat.Type, msg.Chat.ID, msg.MessageThreadID)
	_ = bot.SendChatAction(msg.Chat.ID, "typing")
	if _, err := bot.SendMessageWithOptions(api.SendMessageOptions{
		ChatID:                msg.Chat.ID,
		MessageThreadID:       msg.MessageThreadID,
		ReplyToMessageID:      msg.MessageID,
		Text:                  text,
		ParseMode:             "HTML",
		DisableWebPagePreview: true,
	}); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}
