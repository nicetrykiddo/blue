package admin

import (
	"blue/api"
	"blue/commands"
	"blue/database"
	"blue/models"
	"fmt"
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
		bot.ReplyToMessage(msg.Chat.ID, msg.MessageID, "stats are still waking up bruh, poke me again in a sec")
		return
	}

	text, keyboard, err := statsHomeView()
	if err != nil {
		log.Printf("Error getting total users: %v", err)
		bot.ReplyToMessage(msg.Chat.ID, msg.MessageID, "stats page bonked itself. nothing deep, try again")
		return
	}

	if _, err := bot.SendMessageWithKeyboard(msg.Chat.ID, text, keyboard); err != nil {
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

	text := fmt.Sprintf("```\ngroup stats\n\ntotal groups: %d\n```", totalGroups)

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
	lines = append(lines, "```")
	lines = append(lines, "top 5 groups")
	lines = append(lines, "")

	for i, group := range top5 {
		lines = append(lines, fmt.Sprintf("%d. %s", i+1, group.Title))
		lines = append(lines, fmt.Sprintf("   messages: %d", group.MessageCount))
		if i < len(top5)-1 {
			lines = append(lines, "")
		}
	}

	lines = append(lines, "```")
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
		ParseMode:   "Markdown",
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

	text := fmt.Sprintf("```\nbot stats\n\ntotal users: %d\n\npick one, i'll edit this msg\n```", totalUsers)
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
