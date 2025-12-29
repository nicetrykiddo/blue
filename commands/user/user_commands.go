package user

import (
	"blue/api"
	"blue/cache"
	"blue/commands"
	"blue/models"
	"fmt"
	"log"
	"strings"
)

func init() {
	commands.Register("/start", startHandler)
	commands.Register("/help", helpHandler)
	commands.Register("/echo", echoHandler)
}

func startHandler(bot *api.Bot, msg *models.Message, args []string, stickerCache *cache.StickerCache) {
	bot.SendChatAction(msg.Chat.ID, "typing")

	text := fmt.Sprintf("<code>[+] System Online\n| User: @%s\n| Status: Ready\n| Command: /start</code>", msg.From.Username)
	if _, err := bot.ReplyToMessage(msg.Chat.ID, msg.MessageID, text); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func helpHandler(bot *api.Bot, msg *models.Message, args []string, stickerCache *cache.StickerCache) {
	text := "<code>[*] Command List\n| /echo   - Echo Service\n| /info   - Group Info\n| /rules  - Group Rules</code>"
	if _, err := bot.ReplyToMessage(msg.Chat.ID, msg.MessageID, text); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func echoHandler(bot *api.Bot, msg *models.Message, args []string, stickerCache *cache.StickerCache) {
	if len(args) == 0 {
		bot.ReplyToMessage(msg.Chat.ID, msg.MessageID, "<code>[!] Usage: /echo <message></code>")
		return
	}

	response := fmt.Sprintf("<code>[+] Echo\n| Message: %s</code>", strings.Join(args, " "))
	if _, err := bot.ReplyToMessage(msg.Chat.ID, msg.MessageID, response); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}
