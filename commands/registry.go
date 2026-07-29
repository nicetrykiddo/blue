package commands

import (
	"blue/api"
	"blue/models"
)

type CommandFunc func(*api.Bot, *models.Message, []string)

var registry = make(map[string]CommandFunc)

func Register(name string, handler CommandFunc) {
	registry[name] = handler
}

func Get(name string) (CommandFunc, bool) {
	handler, exists := registry[name]
	return handler, exists
}

func Menu() []api.BotCommand {
	return []api.BotCommand{
		{Command: "id", Description: "show replied user or current chat/topic IDs"},
		{Command: "newctf", Description: "guided form for a new CTF"},
		{Command: "editctf", Description: "edit this CTF topic or choose one"},
		{Command: "openctf", Description: "choose a CTF and create its topic"},
		{Command: "refreshctf", Description: "refresh this CTF topic or choose one"},
		{Command: "livectfs", Description: "show live CTFs"},
		{Command: "upcomingctfs", Description: "show upcoming CTFs"},
		{Command: "imout", Description: "leave the current CTF roster"},
		{Command: "ctfhelp", Description: "show CTF commands"},
		{Command: "allowreaction", Description: "approve replied custom emojis"},
		{Command: "removereaction", Description: "remove replied custom emojis"},
	}
}
