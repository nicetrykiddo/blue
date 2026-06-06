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

	lines := []string{
		fmt.Sprintf("<b>%d. %s</b> <code>#%d</code>", index, safe(event.Title), event.ID),
		fmt.Sprintf("Kickoff: <b>%s</b>", safe(event.StartTime.In(loc).Format("Mon, 02 Jan 15:04 MST"))),
		fmt.Sprintf("Wraps: %s", safe(event.FinishTime.In(loc).Format("Mon, 02 Jan 15:04 MST"))),
		fmt.Sprintf("Flavor: %s | %s", emptyDash(event.Format), safe(venue)),
	}

	if event.Prizes != "" {
		lines = append(lines, fmt.Sprintf("Prize noise: %s", safe(truncate(event.Prizes, 110))))
	}
	if event.VoteCount > 0 {
		lines = append(lines, fmt.Sprintf("Roster: <b>%d</b>", event.VoteCount))
	}

	return strings.Join(lines, "\n")
}

func formatSingleAnnouncement(event database.CTFEvent) string {
	view := newResponseView("Fresh CTF on the board", "Vote below if you are in. Fake confidence is allowed; fake attendance is not.")
	view.rawLine(formatEventCard(1, event))
	return view.text()
}

func formatTopicInitialMessage(event database.CTFEvent) string {
	loc := appLocation()
	view := newResponseView(event.Title, "War room open. Plans go here before they become screenshots.")
	view.field("Kickoff", safe(event.StartTime.In(loc).Format("Mon, 02 Jan 2006 15:04 MST")))
	view.field("Wraps", safe(event.FinishTime.In(loc).Format("Mon, 02 Jan 2006 15:04 MST")))
	view.field("Duration", safe(formatDuration(event.StartTime, event.FinishTime)))
	view.field("Flavor", emptyDash(event.Format))
	view.field("Rules wall", emptyDash(event.Restrictions))
	view.field("Where", safe(eventLocation(event)))

	if event.Prizes != "" {
		view.field("Prize noise", safe(truncate(event.Prizes, 350)))
	}

	if count, err := db.GetCTFParticipantCount(event.ID); err == nil {
		view.field("Roster", fmt.Sprintf("%d", count))
	}

	links := make([]string, 0, 2)
	if siteURL := safeURL(event.URL); siteURL != "" {
		links = append(links, fmt.Sprintf("<a href=\"%s\">Official site</a>", siteURL))
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
	view := newResponseView(title, "Custom opening note. Whoever wrote this now owns the consequences.")
	view.rawLine(safe(body))
	return view.text()
}

func formatStartReminder(event database.CTFEvent, participants []database.CTFParticipant) string {
	loc := appLocation()
	view := newResponseView(event.Title+" starts soon", "The scoreboard is stretching. Prep stops being theoretical now.")
	view.field("Kickoff", safe(event.StartTime.In(loc).Format("Mon, 02 Jan 2006 15:04 MST")))
	view.field("Roster", fmt.Sprintf("%d", len(participants)))
	view.blank()
	view.rawLine(participantMentions(participants, 40))
	view.blank()
	view.rawLine("Drop team/accounts/roles here now so nobody invents process at T-30 seconds.")
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
		linkRow = append(linkRow, models.InlineKeyboardButton{Text: "Official site", URL: siteURL})
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
