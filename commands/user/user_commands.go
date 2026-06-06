package user

import (
	"blue/api"
	"blue/commands"
	"blue/models"
	"fmt"
	"html"
	"log"
	"strings"
)

func init() {
	commands.Register("/start", startHandler)
	commands.Register("/help", helpHandler)
	commands.Register("/echo", echoHandler)
	commands.Register("/getuser", getUserHandler)
}

func startHandler(bot *api.Bot, msg *models.Message, args []string) {
	username := msg.From.Username
	if username == "" {
		username = msg.From.FirstName
	}

	text := fmt.Sprintf("<b>awake.</b>\nI see you, <b>%s</b>. Commands are loaded; chaos is being kept within legal limits.", html.EscapeString(username))
	replyHTML(bot, msg, text)
}

func helpHandler(bot *api.Bot, msg *models.Message, args []string) {
	text := `<b>commands that currently have a pulse</b>

<code>/livectfs</code> - CTFs happening right now
<code>/upcomingctfs</code> - future scoreboard incidents
<code>/ctfhelp</code> - CTF controls
<code>/info</code> - this chat's ID paperwork
<code>/rules</code> - the boring wall, but useful
<code>/stats</code> - bot numbers
<code>/echo</code> - make me repeat your sentence for science`
	replyHTML(bot, msg, text)
}

func echoHandler(bot *api.Bot, msg *models.Message, args []string) {
	if len(args) == 0 {
		replyHTML(bot, msg, "Echo what, exactly? Give me some bytes: <code>/echo &lt;message&gt;</code>")
		return
	}

	response := strings.Join(args, " ")
	replyHTML(bot, msg, html.EscapeString(response))
}

func getUserHandler(bot *api.Bot, msg *models.Message, args []string) {
	if len(args) == 0 {
		replyHTML(bot, msg, "Give me a user ID first: <code>/getuser &lt;userid&gt;</code>")
		return
	}

	userID := args[0]
	text := fmt.Sprintf("Summoning link for <a href=\"tg://user?id=%s\">%s</a>.", html.EscapeString(userID), html.EscapeString(userID))
	replyHTML(bot, msg, text)
}

func replyHTML(bot *api.Bot, msg *models.Message, text string) {
	_ = bot.SendChatAction(msg.Chat.ID, "typing")

	if _, err := bot.SendMessageWithOptions(api.SendMessageOptions{
		ChatID:                msg.Chat.ID,
		MessageThreadID:       msg.MessageThreadID,
		ReplyToMessageID:      msg.MessageID,
		Text:                  text,
		ParseMode:             "HTML",
		DisableWebPagePreview: true,
	}); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}
