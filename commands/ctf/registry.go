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
	callbackOpen     = "ctf_open:"
	callbackEdit     = "ctf_edit:"
	callbackRefresh  = "ctf_refresh:"
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
	commands.Register("/newctf", newCTFHandler)
	commands.Register("/ctfedit", smartEditCTFHandler)
	commands.Register("/editctf", smartEditCTFHandler)
	commands.Register("/ctfrefresh", smartRefreshCTFHandler)
	commands.Register("/refreshctf", smartRefreshCTFHandler)
	commands.Register("/ctftopic", smartTopicCTFHandler)
	commands.Register("/openctf", smartTopicCTFHandler)
	commands.Register("/ctfsync", syncCTFsHandler)
	commands.Register("/ctfhelp", helpCTFHandler)
	commands.Register("/imout", imOutHandler)
	commands.Register("/cancel", cancelWizardHandler)
}

func SetServices(database *database.DB, config *config.Config) {
	db = database
	cfg = config
}

func HandleCallback(bot *api.Bot, query *models.CallbackQuery) bool {
	switch {
	case strings.HasPrefix(query.Data, callbackJoin):
		handleJoinCallback(bot, query)
	case strings.HasPrefix(query.Data, callbackOpen):
		handleManageCallback(bot, query, callbackOpen)
	case strings.HasPrefix(query.Data, callbackEdit):
		handleManageCallback(bot, query, callbackEdit)
	case strings.HasPrefix(query.Data, callbackRefresh):
		handleManageCallback(bot, query, callbackRefresh)
	default:
		return false
	}
	return true
}
