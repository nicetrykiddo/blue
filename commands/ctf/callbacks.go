package ctf

import (
	"blue/api"
	"blue/models"
	"log"
	"strconv"
	"strings"
)

func handleJoinCallback(bot *api.Bot, query *models.CallbackQuery) {
	if db == nil || cfg == nil {
		bot.AnswerCallbackQuery(query.ID, voiceCallbackBooting())
		return
	}
	if query.From == nil {
		bot.AnswerCallbackQuery(query.ID, voiceCallbackNoUser())
		return
	}

	eventID, err := strconv.Atoi(strings.TrimPrefix(query.Data, callbackJoin))
	if err != nil {
		bot.AnswerCallbackQuery(query.ID, voiceCallbackBadButton())
		return
	}

	event, err := db.GetCTFEvent(eventID)
	if err != nil {
		log.Printf("Error loading CTF for vote: %v", err)
		bot.AnswerCallbackQuery(query.ID, voiceCallbackMissingCTF())
		return
	}

	joined, count, err := db.AddCTFParticipant(event.ID, query.From.ID, query.From.Username, query.From.FirstName, query.From.LastName)
	if err != nil {
		log.Printf("Error saving CTF participant: %v", err)
		bot.AnswerCallbackQuery(query.ID, voiceCallbackVoteFailed())
		return
	}

	chatID := callbackChatID(query)
	createdTopic := false
	if chatID != 0 && (event.ForumTopicID != 0 || count >= cfg.CTFTopicVoteThreshold) {
		createdTopic, err = ensureCTFTopic(bot, chatID, event)
		if err != nil {
			log.Printf("Error repairing CTF topic after vote: %v", err)
		}
	}

	event, _ = db.GetCTFEvent(event.ID)
	if joined && event != nil && event.ForumTopicID != 0 {
		if _, err := bot.SendHTMLMessageToThread(chatID, event.ForumTopicID, voiceRosterJoined(query.From)); err != nil {
			log.Printf("Error announcing CTF join: %v", err)
		}
	}
	bot.AnswerCallbackQuery(query.ID, voiceCallbackJoined(createdTopic, joined, count))
}

func callbackChatID(query *models.CallbackQuery) int64 {
	if cfg != nil && cfg.GroupID != 0 {
		return cfg.GroupID
	}
	if query.Message != nil && query.Message.Chat != nil {
		return query.Message.Chat.ID
	}
	return 0
}
