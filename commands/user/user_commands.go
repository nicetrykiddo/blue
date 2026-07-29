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

	text := fmt.Sprintf("<b>awake.</b>\nyo <b>%s</b>.", html.EscapeString(username))
	replyHTML(bot, msg, text)
}

func helpHandler(bot *api.Bot, msg *models.Message, args []string) {
	text := `<b>stuff i can do rn</b>

<code>/id</code> - user, chat and topic ids
<code>/stats</code> - bot numbers
<code>/livectfs</code> - ctfs happening rn
<code>/upcomingctfs</code> - ctfs coming up
<code>/newctf</code> - guided new ctf form
<code>/editctf</code> - edit a ctf topic
<code>/openctf</code> - create a ctf topic`
	replyHTML(bot, msg, text)
}

func echoHandler(bot *api.Bot, msg *models.Message, args []string) {
	if len(args) == 0 {
		replyHTML(bot, msg, "echo what? use <code>/echo &lt;message&gt;</code>")
		return
	}

	response := strings.Join(args, " ")
	replyHTML(bot, msg, html.EscapeString(response))
}

func getUserHandler(bot *api.Bot, msg *models.Message, args []string) {
	if len(args) == 0 {
		replyHTML(bot, msg, "drop a user id first yk: <code>/getuser &lt;userid&gt;</code>")
		return
	}

	userID := args[0]
	text := fmt.Sprintf("user link: <a href=\"tg://user?id=%s\">%s</a>", html.EscapeString(userID), html.EscapeString(userID))
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
