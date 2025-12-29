package group

import (
	"blue/api"
	"blue/cache"
	"blue/commands"
	"blue/models"
	"fmt"
	"log"
)

func init() {
	commands.Register("/info", infoHandler)
	commands.Register("/rules", rulesHandler)
}

func infoHandler(bot *api.Bot, msg *models.Message, args []string, stickerCache *cache.StickerCache) {
	if msg.Chat.Type != "group" && msg.Chat.Type != "supergroup" {
		return
	}
	
	text := fmt.Sprintf("<code>[*] Group Info\n| Title: %s\n| Type: %s\n| ID: %d</code>", msg.Chat.Title, msg.Chat.Type, msg.Chat.ID)
	if _, err := bot.ReplyToMessage(msg.Chat.ID, msg.MessageID, text); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func rulesHandler(bot *api.Bot, msg *models.Message, args []string, stickerCache *cache.StickerCache) {
	if msg.Chat.Type != "group" && msg.Chat.Type != "supergroup" {
		return
	}
	
	text := "<code>[*] Rules\n1. Be respectful\n2. No spam\n3. Stay on topic</code>"
	if _, err := bot.ReplyToMessage(msg.Chat.ID, msg.MessageID, text); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}
