package ctf

import (
	"blue/api"
	"blue/database"
	"strings"
)

func ensureCTFTopic(bot *api.Bot, chatID int64, event *database.CTFEvent) (bool, error) {
	if event.ForumTopicID != 0 {
		refreshed, err := refreshExistingCTFTopic(bot, chatID, event)
		if refreshed {
			return false, err
		}
	}

	return createCTFTopic(bot, chatID, event)
}

func createCTFTopic(bot *api.Bot, chatID int64, event *database.CTFEvent) (bool, error) {
	topic, err := bot.CreateForumTopic(chatID, topicName(event.Title))
	if err != nil {
		return false, err
	}

	message, err := bot.SendHTMLMessageWithKeyboardToThread(
		chatID,
		topic.MessageThreadID,
		formatTopicInitialMessage(*event),
		singleJoinKeyboard(*event),
	)
	if err != nil {
		return false, err
	}

	if err := db.SetCTFTopic(event.ID, topic.MessageThreadID, message.MessageID); err != nil {
		return false, err
	}

	event.ForumTopicID = topic.MessageThreadID
	event.InitialMessageID = message.MessageID
	return true, nil
}

func refreshExistingCTFTopic(bot *api.Bot, chatID int64, event *database.CTFEvent) (bool, error) {
	if event.InitialMessageID != 0 {
		_, err := bot.EditHTMLMessageWithKeyboard(chatID, event.InitialMessageID, formatTopicInitialMessage(*event), singleJoinKeyboard(*event))
		if err == nil || telegramMessageNotModified(err) {
			return true, nil
		}
	}

	message, err := bot.SendHTMLMessageWithKeyboardToThread(
		chatID,
		event.ForumTopicID,
		formatTopicInitialMessage(*event),
		singleJoinKeyboard(*event),
	)
	if err != nil {
		return false, err
	}

	if err := db.SetCTFInitialMessage(event.ID, message.MessageID); err != nil {
		return true, err
	}

	event.InitialMessageID = message.MessageID
	return true, nil
}

func telegramMessageNotModified(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "message is not modified")
}
