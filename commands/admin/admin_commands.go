package admin

import (
	"blue/api"
	"blue/commands"
	"blue/database"
	"blue/models"
	"fmt"
	"html"
	"log"
	"strings"
)

var db *database.DB

func SetDatabase(database *database.DB) {
	db = database
}

func init() {
	commands.Register("/stats", statsHandler)
}

func statsHandler(bot *api.Bot, msg *models.Message, args []string) {
	if db == nil {
		replyStatsHTML(bot, msg, "stats are still waking up bruh, poke me again in a sec")
		return
	}

	text, keyboard, err := statsHomeView()
	if err != nil {
		log.Printf("Error getting total users: %v", err)
		replyStatsHTML(bot, msg, "stats page bonked itself. nothing deep, try again")
		return
	}

	if _, err := bot.SendMessageWithOptions(api.SendMessageOptions{
		ChatID:                msg.Chat.ID,
		MessageThreadID:       msg.MessageThreadID,
		ReplyToMessageID:      msg.MessageID,
		Text:                  text,
		ParseMode:             "HTML",
		ReplyMarkup:           keyboard,
		DisableWebPagePreview: true,
	}); err != nil {
		log.Printf("Error sending stats: %v", err)
	}
}

func HandleCallback(bot *api.Bot, query *models.CallbackQuery) {
	if db == nil {
		return
	}
	if !strings.HasPrefix(query.Data, "stats_") {
		return
	}

	switch query.Data {
	case "stats_groups":
		handleGroupsCallback(bot, query)
	case "stats_top5":
		handleTop5Callback(bot, query)
	case "stats_back":
		handleStatsBackCallback(bot, query)
	default:
		bot.AnswerCallbackQuery(query.ID, "old button :P")
	}
}

func handleGroupsCallback(bot *api.Bot, query *models.CallbackQuery) {
	totalGroups, err := db.GetTotalGroups()
	if err != nil {
		log.Printf("Error getting total groups: %v", err)
		bot.AnswerCallbackQuery(query.ID, "stats blinked, tap again")
		return
	}

	text := fmt.Sprintf("<b>group stats</b>\n\n<pre>total groups: %d</pre>", totalGroups)

	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "top groups", CallbackData: "stats_top5"},
			},
			{
				{Text: "back", CallbackData: "stats_back"},
			},
		},
	}

	if !editStatsMessage(bot, query, text, keyboard, "group stats") {
		return
	}

	bot.AnswerCallbackQuery(query.ID, "")
}

func handleTop5Callback(bot *api.Bot, query *models.CallbackQuery) {
	top5, err := db.GetTop5Groups()
	if err != nil {
		log.Printf("Error getting top 5 groups: %v", err)
		bot.AnswerCallbackQuery(query.ID, "list blinked, try again")
		return
	}

	var lines []string
	lines = append(lines, "<b>top 5 groups</b>")
	lines = append(lines, "")

	for i, group := range top5 {
		lines = append(lines, fmt.Sprintf("<b>%d.</b> %s", i+1, html.EscapeString(group.Title)))
		lines = append(lines, fmt.Sprintf("<code>messages: %d</code>", group.MessageCount))
		if i < len(top5)-1 {
			lines = append(lines, "")
		}
	}

	text := strings.Join(lines, "\n")

	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "groups count", CallbackData: "stats_groups"},
			},
			{
				{Text: "back", CallbackData: "stats_back"},
			},
		},
	}

	if !editStatsMessage(bot, query, text, keyboard, "top 5 stats") {
		return
	}

	bot.AnswerCallbackQuery(query.ID, "")
}

func handleStatsBackCallback(bot *api.Bot, query *models.CallbackQuery) {
	text, keyboard, err := statsHomeView()
	if err != nil {
		log.Printf("Error getting total users: %v", err)
		bot.AnswerCallbackQuery(query.ID, "stats blinked, tap again")
		return
	}

	if !editStatsMessage(bot, query, text, keyboard, "stats home") {
		return
	}

	bot.AnswerCallbackQuery(query.ID, "")
}

func editStatsMessage(bot *api.Bot, query *models.CallbackQuery, text string, keyboard *models.InlineKeyboardMarkup, label string) bool {
	if query.Message == nil || query.Message.Chat == nil {
		bot.AnswerCallbackQuery(query.ID, "that message is gone gone")
		return false
	}

	if _, err := bot.EditMessageWithOptions(api.EditMessageOptions{
		ChatID:      query.Message.Chat.ID,
		MessageID:   query.Message.MessageID,
		Text:        text,
		ParseMode:   "HTML",
		ReplyMarkup: keyboard,
	}); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "message is not modified") {
			return true
		}
		log.Printf("Error editing %s: %v", label, err)
		bot.AnswerCallbackQuery(query.ID, "telegram refused to redraw it, classic")
		return false
	}

	return true
}

func statsHomeView() (string, *models.InlineKeyboardMarkup, error) {
	totalUsers, err := db.GetTotalUsers()
	if err != nil {
		return "", nil, err
	}

	text := fmt.Sprintf("<b>bot stats</b>\n\n<pre>total users: %d</pre>\n\npick one, i'll edit this msg", totalUsers)
	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "groups count", CallbackData: "stats_groups"},
			},
			{
				{Text: "top groups", CallbackData: "stats_top5"},
			},
		},
	}

	return text, keyboard, nil
}

func replyStatsHTML(bot *api.Bot, msg *models.Message, text string) {
	if _, err := bot.SendMessageWithOptions(api.SendMessageOptions{
		ChatID:                msg.Chat.ID,
		MessageThreadID:       msg.MessageThreadID,
		ReplyToMessageID:      msg.MessageID,
		Text:                  text,
		ParseMode:             "HTML",
		DisableWebPagePreview: true,
	}); err != nil {
		log.Printf("Error sending stats reply: %v", err)
	}
}
