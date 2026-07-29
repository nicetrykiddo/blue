package ctf

import (
	"blue/api"
	"blue/database"
	"blue/models"
	"hash/fnv"
	"strings"
	"sync"
)

var forumIcons = struct {
	sync.Mutex
	loaded   bool
	stickers []models.Sticker
}{}

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
	topic, err := bot.CreateForumTopic(chatID, topicName(event.Title), topicIconID(bot, event.Title))
	if err != nil {
		return false, err
	}

	if err := db.SetCTFTopic(event.ID, topic.MessageThreadID, 0); err != nil {
		_ = bot.CloseForumTopic(chatID, topic.MessageThreadID)
		return false, err
	}
	event.ForumTopicID = topic.MessageThreadID

	message, err := bot.SendHTMLMessageWithKeyboardToThread(
		chatID,
		topic.MessageThreadID,
		formatTopicInitialMessage(*event),
		singleJoinKeyboard(*event),
	)
	if err != nil {
		return false, err
	}

	if err := db.SetCTFInitialMessage(event.ID, message.MessageID); err != nil {
		return false, err
	}

	event.InitialMessageID = message.MessageID
	return true, nil
}

func topicIconID(bot *api.Bot, title string) string {
	forumIcons.Lock()
	defer forumIcons.Unlock()
	if !forumIcons.loaded {
		stickers, err := bot.GetForumTopicIconStickers()
		if err != nil {
			return ""
		}
		forumIcons.stickers = stickers
		forumIcons.loaded = true
	}
	return selectTopicIcon(title, forumIcons.stickers)
}

func selectTopicIcon(title string, stickers []models.Sticker) string {
	words := strings.FieldsFunc(strings.ToLower(title), func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9')
	})
	wanted := "🏆"
	for _, word := range words {
		switch word {
		case "web":
			wanted = "🌐"
		case "crypto", "cryptography":
			wanted = "🔐"
		case "pwn", "binary":
			wanted = "💻"
		case "reverse", "reversing", "rev":
			wanted = "🔍"
		case "forensic", "forensics":
			wanted = "🕵"
		case "osint":
			wanted = "👀"
		case "mobile", "android", "ios":
			wanted = "📱"
		case "cloud":
			wanted = "☁"
		case "hardware", "iot":
			wanted = "⚙"
		}
	}

	for _, sticker := range stickers {
		if cleanEmoji(sticker.Emoji) == cleanEmoji(wanted) {
			return sticker.CustomEmojiID
		}
	}
	if len(stickers) == 0 {
		return ""
	}
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(strings.ToLower(title)))
	return stickers[int(hash.Sum32())%len(stickers)].CustomEmojiID
}

func cleanEmoji(value string) string {
	return strings.ReplaceAll(value, "\ufe0f", "")
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
