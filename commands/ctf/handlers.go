package ctf

import (
	"blue/api"
	"blue/models"
	"log"
	"strconv"
	"time"
)

func liveCTFsHandler(bot *api.Bot, msg *models.Message, args []string) {
	if db == nil {
		replyHTML(bot, msg, voiceStorageMissing, nil)
		return
	}

	working := startWorkingReply(bot, msg, voiceLoadingLive)
	now := time.Now()
	if _, err := refreshEvents(now.AddDate(0, 0, -7), now.AddDate(0, 0, 1), 100); err != nil {
		log.Printf("Error refreshing live CTFs: %v", err)
	}
	advanceWorkingReply(bot, msg, working, voiceLoadingLiveDone)

	events, err := db.ListLiveCTFEvents(now.UTC(), 10)
	if err != nil {
		log.Printf("Error listing live CTFs: %v", err)
		finishWorkingReply(bot, msg, working, voiceLiveError, nil)
		return
	}

	finishWorkingReply(bot, msg, working, formatLiveList(events, time.Now()), joinKeyboard(events))
}

func upcomingCTFsHandler(bot *api.Bot, msg *models.Message, args []string) {
	if db == nil {
		replyHTML(bot, msg, voiceStorageMissing, nil)
		return
	}

	working := startWorkingReply(bot, msg, voiceLoadingUpcoming)
	days := defaultLookaheadDays()
	if len(args) > 0 {
		if parsed, err := strconv.Atoi(args[0]); err == nil && parsed > 0 && parsed <= 90 {
			days = parsed
		}
	}

	now := time.Now()
	if _, err := refreshUpcomingEvents(now, days, 100); err != nil {
		log.Printf("Error refreshing upcoming CTFs: %v", err)
	}
	advanceWorkingReply(bot, msg, working, voiceLoadingFutureDone)

	events, err := db.ListUpcomingCTFEvents(now.UTC(), now.AddDate(0, 0, days).UTC(), 10)
	if err != nil {
		log.Printf("Error listing upcoming CTFs: %v", err)
		finishWorkingReply(bot, msg, working, voiceUpcomingError, nil)
		return
	}

	finishWorkingReply(bot, msg, working, formatUpcomingList(events, days, time.Now()), joinKeyboard(events))
}

func addCTFHandler(bot *api.Bot, msg *models.Message, args []string) {
	if !canManageCTFs(msg.From) {
		replyHTML(bot, msg, voiceNoPermission, nil)
		return
	}

	event, err := parseManualCTF(commandPayload(msg))
	if err != nil {
		replyHTML(bot, msg, addUsage(err), nil)
		return
	}

	saved, err := db.UpsertCTFEvent(event)
	if err != nil {
		log.Printf("Error creating manual CTF: %v", err)
		replyHTML(bot, msg, voiceSavedError(), nil)
		return
	}

	targetChatID, targetThreadID := announcementTarget(msg)
	if targetChatID != 0 {
		if _, err := bot.SendHTMLMessageWithKeyboardToThread(targetChatID, targetThreadID, formatSingleAnnouncement(*saved), singleJoinKeyboard(*saved)); err != nil {
			log.Printf("Error sending manual CTF announcement: %v", err)
		}
	}

	replyHTML(bot, msg, voiceManualCreated(saved.Title), nil)
}

func editCTFHandler(bot *api.Bot, msg *models.Message, args []string) {
	if !canManageCTFs(msg.From) {
		replyHTML(bot, msg, voiceNoPermission, nil)
		return
	}

	eventID, body, err := parseIDAndBody(commandPayload(msg))
	if err != nil {
		replyHTML(bot, msg, voiceUsageEdit(), nil)
		return
	}

	event, err := db.GetCTFEvent(eventID)
	if err != nil {
		replyHTML(bot, msg, voiceNotFound(), nil)
		return
	}
	if event.ForumTopicID == 0 || event.InitialMessageID == 0 {
		replyHTML(bot, msg, voiceNoTopic(event.ID), nil)
		return
	}

	chatID := groupOrMessageChatID(msg)
	if _, err := bot.EditHTMLMessageWithKeyboard(chatID, event.InitialMessageID, formatCustomInitialMessage(event.Title, body), singleJoinKeyboard(*event)); err != nil {
		log.Printf("Error editing CTF initial message: %v", err)
		replyHTML(bot, msg, voiceEditFailed(), nil)
		return
	}

	replyHTML(bot, msg, voiceEdited(event.Title), nil)
}

func refreshCTFHandler(bot *api.Bot, msg *models.Message, args []string) {
	if !canManageCTFs(msg.From) {
		replyHTML(bot, msg, voiceNoPermission, nil)
		return
	}

	eventID, err := parseFirstID(args)
	if err != nil {
		replyHTML(bot, msg, voiceUsageRefresh(), nil)
		return
	}

	event, err := db.GetCTFEvent(eventID)
	if err != nil {
		replyHTML(bot, msg, voiceNotFound(), nil)
		return
	}
	if event.ForumTopicID == 0 || event.InitialMessageID == 0 {
		replyHTML(bot, msg, voiceNoTopicShort(), nil)
		return
	}

	chatID := groupOrMessageChatID(msg)
	if _, err := bot.EditHTMLMessageWithKeyboard(chatID, event.InitialMessageID, formatTopicInitialMessage(*event), singleJoinKeyboard(*event)); err != nil {
		log.Printf("Error refreshing CTF initial message: %v", err)
		replyHTML(bot, msg, voiceRefreshFailed(), nil)
		return
	}

	replyHTML(bot, msg, voiceRefreshed(event.Title), nil)
}

func topicCTFHandler(bot *api.Bot, msg *models.Message, args []string) {
	if !canManageCTFs(msg.From) {
		replyHTML(bot, msg, voiceNoPermission, nil)
		return
	}

	eventID, err := parseFirstID(args)
	if err != nil {
		replyHTML(bot, msg, voiceUsageTopic(), nil)
		return
	}

	event, err := db.GetCTFEvent(eventID)
	if err != nil {
		replyHTML(bot, msg, voiceNotFound(), nil)
		return
	}

	topicCreated, err := ensureCTFTopic(bot, groupOrMessageChatID(msg), event)
	if err != nil {
		log.Printf("Error creating CTF topic: %v", err)
		replyHTML(bot, msg, voiceTopicCreateFailed(), nil)
		return
	}

	if topicCreated {
		replyHTML(bot, msg, voiceTopicCreated(event.Title), nil)
	} else {
		replyHTML(bot, msg, voiceTopicExists(event.Title), nil)
	}
}

func syncCTFsHandler(bot *api.Bot, msg *models.Message, args []string) {
	if !canManageCTFs(msg.From) {
		replyHTML(bot, msg, voiceNoPermission, nil)
		return
	}

	if err := SendDailyDigest(bot, true); err != nil {
		log.Printf("Error force-sending CTF digest: %v", err)
		replyHTML(bot, msg, voiceDigestFailed(), nil)
		return
	}

	replyHTML(bot, msg, voiceDigestSent(), nil)
}

func helpCTFHandler(bot *api.Bot, msg *models.Message, args []string) {
	text := `<b>ctf commands</b>

<code>/livectfs</code> - what is burning rn
<code>/upcomingctfs [days]</code> - upcoming ctfs
<code>/ctfsync</code> - send today's digest rn
<code>/ctfadd</code> - add one by hand
<code>/ctftopic &lt;id&gt;</code> - open the topic
<code>/ctfedit &lt;id&gt; &lt;text&gt;</code> - rewrite the opening msg
<code>/ctfrefresh &lt;id&gt;</code> - refresh the opening msg`
	replyHTML(bot, msg, text, nil)
}
