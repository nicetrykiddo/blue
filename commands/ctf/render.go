package ctf

import (
	"blue/api"
	"blue/database"
	"blue/models"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

func formatDailyDigest(events []database.CTFEvent, now time.Time) string {
	return renderEventList(voiceDailyList(now), events, 6)
}

func formatLiveList(events []database.CTFEvent, now time.Time) string {
	return renderEventList(voiceLiveList(now), events, 10)
}

func formatUpcomingList(events []database.CTFEvent, days int, now time.Time) string {
	return renderEventList(voiceUpcomingList(now, days), events, 10)
}

func formatEventCard(index int, event database.CTFEvent) string {
	loc := appLocation()
	venue := "Online"
	if event.Location != "" {
		venue += " | " + event.Location
	}

	title := fmt.Sprintf("<b>%d. %s</b> <code>#%d</code>", index, safe(event.Title), event.ID)
	details := []string{
		fmt.Sprintf("starts: <b>%s</b>", safe(event.StartTime.In(loc).Format("Mon, 02 Jan 15:04 MST"))),
		fmt.Sprintf("ends: %s", safe(event.FinishTime.In(loc).Format("Mon, 02 Jan 15:04 MST"))),
		fmt.Sprintf("type: %s | %s", emptyDash(event.Format), safe(venue)),
	}

	if event.Prizes != "" {
		details = append(details, fmt.Sprintf("prizes: %s", safe(truncate(event.Prizes, 180))))
	}
	if event.VoteCount > 0 {
		details = append(details, fmt.Sprintf("roster: <b>%d</b>", event.VoteCount))
	}

	return title + "\n" + expandableQuote(details)
}

func formatSingleAnnouncement(event database.CTFEvent) string {
	view := newResponseView("ctf dropped", "vote if you're playing.")
	view.rawLine(formatEventCard(1, event))
	return view.text()
}

func formatTopicInitialMessage(event database.CTFEvent) string {
	loc := appLocation()
	view := newResponseView(event.Title, "topic open. drop plans here.")
	view.field("starts", safe(event.StartTime.In(loc).Format("Mon, 02 Jan 2006 15:04 MST")))
	view.field("ends", safe(event.FinishTime.In(loc).Format("Mon, 02 Jan 2006 15:04 MST")))
	view.field("length", safe(formatDuration(event.StartTime, event.FinishTime)))
	view.field("type", emptyDash(event.Format))
	view.field("rules", emptyDash(event.Restrictions))
	view.field("where", safe(eventLocation(event)))

	if event.Prizes != "" {
		view.field("prizes", safe(truncate(event.Prizes, 350)))
	}

	if count, err := db.GetCTFParticipantCount(event.ID); err == nil {
		view.field("roster", fmt.Sprintf("%d", count))
	}

	links := make([]string, 0, 2)
	if siteURL := safeURL(event.URL); siteURL != "" {
		links = append(links, fmt.Sprintf("<a href=\"%s\">site</a>", siteURL))
	}
	if ctftimeURL := safeURL(event.CTFTimeURL); ctftimeURL != "" {
		links = append(links, fmt.Sprintf("<a href=\"%s\">CTFtime</a>", ctftimeURL))
	}
	if len(links) > 0 {
		view.blank()
		view.rawLine(strings.Join(links, " | "))
	}

	if event.Description != "" {
		view.blank()
		view.rawLine(safe(truncate(event.Description, 700)))
	}

	return view.text()
}

func formatCustomInitialMessage(title, body string) string {
	view := newResponseView(title, "custom opening msg.")
	view.rawLine(safe(body))
	return view.text()
}

func formatStartReminder(event database.CTFEvent, participants []database.CTFParticipant) string {
	loc := appLocation()
	view := newResponseView(event.Title+" starts soon", "time to get ready.")
	view.field("starts", safe(event.StartTime.In(loc).Format("Mon, 02 Jan 2006 15:04 MST")))
	view.field("roster", fmt.Sprintf("%d", len(participants)))
	view.blank()
	view.rawLine(participantMentions(participants, 40))
	view.blank()
	view.rawLine("drop team/accounts/roles here.")
	return view.text()
}

func joinKeyboard(events []database.CTFEvent) *models.InlineKeyboardMarkup {
	if len(events) == 0 {
		return nil
	}

	limit := min(len(events), 10)
	rows := make([][]models.InlineKeyboardButton, 0, limit)
	for i := 0; i < limit; i++ {
		event := events[i]
		label := fmt.Sprintf("I'm in #%d", event.ID)
		if event.VoteCount > 0 {
			label = fmt.Sprintf("I'm in #%d (%d)", event.ID, event.VoteCount)
		}
		rows = append(rows, []models.InlineKeyboardButton{{Text: label, CallbackData: callbackJoin + strconv.Itoa(event.ID)}})
	}

	return &models.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func singleJoinKeyboard(event database.CTFEvent) *models.InlineKeyboardMarkup {
	rows := [][]models.InlineKeyboardButton{
		{{Text: fmt.Sprintf("I'm in #%d", event.ID), CallbackData: callbackJoin + strconv.Itoa(event.ID)}},
	}

	linkRow := []models.InlineKeyboardButton{}
	if siteURL := buttonURL(event.URL); siteURL != "" {
		linkRow = append(linkRow, models.InlineKeyboardButton{Text: "site", URL: siteURL})
	}
	if ctftimeURL := buttonURL(event.CTFTimeURL); ctftimeURL != "" {
		linkRow = append(linkRow, models.InlineKeyboardButton{Text: "CTFtime", URL: ctftimeURL})
	}
	if len(linkRow) > 0 {
		rows = append(rows, linkRow)
	}

	return &models.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func replyHTML(bot *api.Bot, msg *models.Message, text string, keyboard *models.InlineKeyboardMarkup) {
	_ = bot.SendChatAction(msg.Chat.ID, "typing")

	opts := api.SendMessageOptions{
		ChatID:                msg.Chat.ID,
		MessageThreadID:       msg.MessageThreadID,
		ReplyToMessageID:      msg.MessageID,
		Text:                  trimTelegramMessage(text),
		ParseMode:             "HTML",
		ReplyMarkup:           keyboard,
		DisableWebPagePreview: true,
	}
	if _, err := bot.SendMessageWithOptions(opts); err != nil {
		log.Printf("Error sending CTF reply: %v", err)
	}
}

func expandableQuote(lines []string) string {
	return "<blockquote expandable>" + strings.Join(lines, "\n") + "</blockquote>"
}
