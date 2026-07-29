package commands

import (
	"blue/api"
	"blue/models"
	"errors"
	"fmt"
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

func PublicMenu() []api.BotCommand {
	return []api.BotCommand{
		{Command: "id", Description: "show replied user or current chat/topic IDs"},
		{Command: "help", Description: "show available commands"},
		{Command: "livectfs", Description: "show live CTFs"},
		{Command: "upcomingctfs", Description: "show upcoming CTFs"},
		{Command: "imout", Description: "leave the current CTF roster"},
		{Command: "ctfhelp", Description: "show CTF commands"},
		{Command: "stats", Description: "show bot statistics"},
	}
}

func AdminMenu() []api.BotCommand {
	return append(PublicMenu(),
		api.BotCommand{Command: "newctf", Description: "guided form for a new CTF"},
		api.BotCommand{Command: "editctf", Description: "edit this CTF topic or choose one"},
		api.BotCommand{Command: "openctf", Description: "choose a CTF and create its topic"},
		api.BotCommand{Command: "refreshctf", Description: "refresh this CTF topic or choose one"},
		api.BotCommand{Command: "ctfsync", Description: "send the CTF digest now"},
		api.BotCommand{Command: "allowreaction", Description: "approve replied custom emojis"},
		api.BotCommand{Command: "removereaction", Description: "remove replied custom emojis"},
		api.BotCommand{Command: "clearreactions", Description: "clear approved reaction emojis"},
	)
}

func ConfigureMenu(bot *api.Bot, groupID int64, adminIDs map[int64]bool) error {
	if err := bot.SetCommands(PublicMenu()); err != nil {
		return err
	}
	if groupID == 0 {
		return nil
	}

	var errs []error
	for userID, allowed := range adminIDs {
		if !allowed {
			continue
		}
		if err := bot.SetCommandsForChatMember(AdminMenu(), groupID, userID); err != nil {
			errs = append(errs, fmt.Errorf("admin %d: %w", userID, err))
		}
	}
	return errors.Join(errs...)
}
