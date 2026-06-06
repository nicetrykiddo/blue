package ctf

import (
	"blue/api"
	"blue/database"
)

func ensureCTFTopic(bot *api.Bot, chatID int64, event *database.CTFEvent) (bool, error) {
	if event.ForumTopicID != 0 {
		return false, nil
	}

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
