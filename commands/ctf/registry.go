package ctf

import (
	"blue/api"
	"blue/commands"
	"blue/config"
	"blue/database"
	"blue/models"
	"strings"
)

const (
	ctftimeEventsURL = "https://ctftime.org/api/v1/events/"
	callbackJoin     = "ctf_join:"
	reminderKeyStart = "start"
)

var (
	db  *database.DB
	cfg *config.Config
)

func init() {
	commands.Register("/livectfs", liveCTFsHandler)
	commands.Register("/ctfs", liveCTFsHandler)
	commands.Register("/upcomingctfs", upcomingCTFsHandler)
	commands.Register("/ctfadd", addCTFHandler)
	commands.Register("/ctfedit", editCTFHandler)
	commands.Register("/ctfrefresh", refreshCTFHandler)
	commands.Register("/ctftopic", topicCTFHandler)
	commands.Register("/ctfsync", syncCTFsHandler)
	commands.Register("/ctfhelp", helpCTFHandler)
	commands.Register("/imout", imOutHandler)
}

func SetServices(database *database.DB, config *config.Config) {
	db = database
	cfg = config
}

func HandleCallback(bot *api.Bot, query *models.CallbackQuery) bool {
	if !strings.HasPrefix(query.Data, callbackJoin) {
		return false
	}

	handleJoinCallback(bot, query)
	return true
}
