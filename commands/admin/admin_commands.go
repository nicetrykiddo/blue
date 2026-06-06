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
		bot.ReplyToMessage(msg.Chat.ID, msg.MessageID, "Database not initialized")
		return
	}

	totalUsers, err := db.GetTotalUsers()
	if err != nil {
		log.Printf("Error getting total users: %v", err)
		bot.ReplyToMessage(msg.Chat.ID, msg.MessageID, "Error fetching stats")
		return
	}

	text := fmt.Sprintf("```\nBot Statistics\n\nTotal Users: %d\n\nSelect an option below\n```", totalUsers)

	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Total Groups", CallbackData: "stats_groups"},
			},
			{
				{Text: "Top 5 Groups", CallbackData: "stats_top5"},
			},
		},
	}

	if _, err := bot.SendMessageWithKeyboard(msg.Chat.ID, text, keyboard); err != nil {
		log.Printf("Error sending stats: %v", err)
	}
}

func HandleCallback(bot *api.Bot, query *models.CallbackQuery) {
	if db == nil {
		return
	}

	switch query.Data {
	case "stats_groups":
		handleGroupsCallback(bot, query)
	case "stats_top5":
		handleTop5Callback(bot, query)
	}
}

func handleGroupsCallback(bot *api.Bot, query *models.CallbackQuery) {
	totalGroups, err := db.GetTotalGroups()
	if err != nil {
		log.Printf("Error getting total groups: %v", err)
		bot.AnswerCallbackQuery(query.ID, "Error fetching groups")
		return
	}

	text := fmt.Sprintf("```\nGroup Stats\n\nTotal Groups: %d\n```", totalGroups)

	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Top 5 Groups", CallbackData: "stats_top5"},
			},
			{
				{Text: "Back", CallbackData: "stats_back"},
			},
		},
	}

	if query.Message != nil {
		bot.EditMessage(query.Message.Chat.ID, query.Message.MessageID, text)
		bot.SendMessageWithKeyboard(query.Message.Chat.ID, text, keyboard)
	}

	bot.AnswerCallbackQuery(query.ID, "")
}

func handleTop5Callback(bot *api.Bot, query *models.CallbackQuery) {
	top5, err := db.GetTop5Groups()
	if err != nil {
		log.Printf("Error getting top 5 groups: %v", err)
		bot.AnswerCallbackQuery(query.ID, "Error fetching top groups")
		return
	}

	var lines []string
	lines = append(lines, "```")
	lines = append(lines, "Top 5 Groups")
	lines = append(lines, "")

	for i, group := range top5 {
		lines = append(lines, fmt.Sprintf("%d. %s", i+1, group.Title))
		lines = append(lines, fmt.Sprintf("   Messages: %d", group.MessageCount))
		if i < len(top5)-1 {
			lines = append(lines, "")
		}
	}

	lines = append(lines, "```")
	text := strings.Join(lines, "\n")

	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Total Groups", CallbackData: "stats_groups"},
			},
			{
				{Text: "Back", CallbackData: "stats_back"},
			},
		},
	}

	if query.Message != nil {
		bot.EditMessage(query.Message.Chat.ID, query.Message.MessageID, text)
		bot.SendMessageWithKeyboard(query.Message.Chat.ID, text, keyboard)
	}

	bot.AnswerCallbackQuery(query.ID, "")
}
