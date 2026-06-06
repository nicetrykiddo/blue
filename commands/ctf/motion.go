package ctf

import (
	"blue/api"
	"blue/models"
	"log"
)

func startWorkingReply(bot *api.Bot, msg *models.Message, text string) *models.Message {
	_ = bot.SendChatAction(msg.Chat.ID, "typing")

	message, err := bot.SendMessageWithOptions(api.SendMessageOptions{
		ChatID:                msg.Chat.ID,
		MessageThreadID:       msg.MessageThreadID,
		ReplyToMessageID:      msg.MessageID,
		Text:                  trimTelegramMessage(text),
		ParseMode:             "HTML",
		DisableWebPagePreview: true,
	})
	if err != nil {
		log.Printf("Error sending CTF working reply: %v", err)
		return nil
	}

	return message
}

func finishWorkingReply(bot *api.Bot, msg *models.Message, working *models.Message, text string, keyboard *models.InlineKeyboardMarkup) {
	_ = bot.SendChatAction(msg.Chat.ID, "typing")

	if working != nil {
		if _, err := bot.EditHTMLMessageWithKeyboard(msg.Chat.ID, working.MessageID, trimTelegramMessage(text), keyboard); err != nil {
			log.Printf("Error editing CTF working reply: %v", err)
			replyHTML(bot, msg, text, keyboard)
		}
		return
	}

	replyHTML(bot, msg, text, keyboard)
}

func advanceWorkingReply(bot *api.Bot, msg *models.Message, working *models.Message, text string) {
	_ = bot.SendChatAction(msg.Chat.ID, "typing")
	if working == nil {
		return
	}

	if _, err := bot.EditHTMLMessage(msg.Chat.ID, working.MessageID, trimTelegramMessage(text)); err != nil {
		log.Printf("Error advancing CTF working reply: %v", err)
	}
}
