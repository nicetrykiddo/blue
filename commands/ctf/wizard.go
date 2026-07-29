package ctf

import (
	"blue/api"
	"blue/database"
	"blue/models"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"
)

type wizardKey struct {
	chatID   int64
	threadID int
	userID   int64
}

type wizardState struct {
	kind            string
	stage           int
	eventID         int
	title           string
	start           time.Time
	finish          time.Time
	url             string
	promptMessageID int
	expires         time.Time
}

var wizardSessions = struct {
	sync.Mutex
	items map[wizardKey]wizardState
}{items: make(map[wizardKey]wizardState)}

// ponytail: forms live in memory; persist them only if surviving bot restarts becomes necessary.

func newCTFHandler(bot *api.Bot, msg *models.Message, args []string) {
	if strings.TrimSpace(commandPayload(msg)) != "" {
		addCTFHandler(bot, msg, args)
		return
	}
	if !canManageCTFs(msg.From) {
		replyHTML(bot, msg, voiceNoPermission, nil)
		return
	}
	startWizard(bot, msg, wizardState{kind: "new"}, "<b>new ctf · 1/5</b>\nWhat is the title?", "CTF title")
}

func smartEditCTFHandler(bot *api.Bot, msg *models.Message, args []string) {
	if len(args) > 0 {
		editCTFHandler(bot, msg, args)
		return
	}
	if !canManageCTFs(msg.From) {
		replyHTML(bot, msg, voiceNoPermission, nil)
		return
	}
	if event := eventForCurrentTopic(msg); event != nil {
		startEditWizard(bot, msg, event.ID)
		return
	}
	showEventPicker(bot, msg, callbackEdit, "Choose the CTF whose opening message you want to edit.")
}

func smartRefreshCTFHandler(bot *api.Bot, msg *models.Message, args []string) {
	if len(args) > 0 {
		refreshCTFHandler(bot, msg, args)
		return
	}
	if !canManageCTFs(msg.From) {
		replyHTML(bot, msg, voiceNoPermission, nil)
		return
	}
	if event := eventForCurrentTopic(msg); event != nil {
		refreshEvent(bot, msg, event)
		return
	}
	showEventPicker(bot, msg, callbackRefresh, "Choose the CTF to refresh.")
}

func smartTopicCTFHandler(bot *api.Bot, msg *models.Message, args []string) {
	if len(args) > 0 {
		topicCTFHandler(bot, msg, args)
		return
	}
	if !canManageCTFs(msg.From) {
		replyHTML(bot, msg, voiceNoPermission, nil)
		return
	}
	showEventPicker(bot, msg, callbackOpen, "Choose the CTF that needs a topic.")
}

func showEventPicker(bot *api.Bot, msg *models.Message, action, title string) {
	events, err := db.ListUpcomingCTFEvents(time.Now().UTC(), time.Now().AddDate(0, 0, 90).UTC(), 10)
	if err != nil {
		log.Printf("Error loading CTF picker: %v", err)
		replyHTML(bot, msg, "Could not load the CTF list. Try again shortly.", nil)
		return
	}
	if len(events) == 0 {
		replyHTML(bot, msg, "No live or upcoming CTFs are saved yet. Run <code>/upcomingctfs</code> once, then try again.", nil)
		return
	}

	rows := make([][]models.InlineKeyboardButton, 0, len(events))
	for _, event := range events {
		rows = append(rows, []models.InlineKeyboardButton{{
			Text:         compactCTFButtonName(event.Title),
			CallbackData: action + fmt.Sprint(event.ID),
		}})
	}
	replyHTML(bot, msg, title, &models.InlineKeyboardMarkup{InlineKeyboard: rows})
}

func handleManageCallback(bot *api.Bot, query *models.CallbackQuery, action string) {
	if query.From == nil || !canManageCTFs(query.From) {
		bot.AnswerCallbackQuery(query.ID, voiceNoPermission)
		return
	}
	if query.Message == nil || query.Message.Chat == nil {
		bot.AnswerCallbackQuery(query.ID, "That picker message is no longer available.")
		return
	}

	eventID, err := callbackEventID(query.Data, action)
	if err != nil {
		bot.AnswerCallbackQuery(query.ID, "That button is invalid. Run the command again.")
		return
	}
	event, err := db.GetCTFEvent(eventID)
	if err != nil {
		bot.AnswerCallbackQuery(query.ID, "That CTF is no longer available.")
		return
	}

	switch action {
	case callbackOpen:
		created, err := ensureCTFTopic(bot, manageCallbackChatID(query), event)
		if err != nil {
			log.Printf("Error creating selected CTF topic: %v", err)
			bot.AnswerCallbackQuery(query.ID, voiceTopicCreateFailed())
			return
		}
		if created {
			bot.AnswerCallbackQuery(query.ID, "Topic created.")
		} else {
			bot.AnswerCallbackQuery(query.ID, "That topic already exists.")
		}
	case callbackRefresh:
		if err := refreshSelectedEvent(bot, manageCallbackChatID(query), event); err != nil {
			log.Printf("Error refreshing selected CTF: %v", err)
			bot.AnswerCallbackQuery(query.ID, voiceRefreshFailed())
			return
		}
		bot.AnswerCallbackQuery(query.ID, "Opening message refreshed.")
	case callbackEdit:
		if event.ForumTopicID == 0 || event.InitialMessageID == 0 {
			bot.AnswerCallbackQuery(query.ID, voiceNoTopicShort())
			return
		}
		if !startCallbackEditWizard(bot, query, event.ID) {
			return
		}
		bot.AnswerCallbackQuery(query.ID, "Reply to the prompt.")
	}
}

func manageCallbackChatID(query *models.CallbackQuery) int64 {
	if cfg != nil && cfg.GroupID != 0 {
		return cfg.GroupID
	}
	return query.Message.Chat.ID
}

func callbackEventID(data, prefix string) (int, error) {
	return parseFirstID([]string{strings.TrimPrefix(data, prefix)})
}

func eventForCurrentTopic(msg *models.Message) *database.CTFEvent {
	if db == nil || msg.MessageThreadID == 0 {
		return nil
	}
	event, err := db.GetCTFEventByForumTopicID(msg.MessageThreadID)
	if err != nil {
		return nil
	}
	return event
}

func refreshEvent(bot *api.Bot, msg *models.Message, event *database.CTFEvent) {
	if err := refreshSelectedEvent(bot, groupOrMessageChatID(msg), event); err != nil {
		log.Printf("Error refreshing CTF initial message: %v", err)
		replyHTML(bot, msg, voiceRefreshFailed(), nil)
		return
	}
	replyHTML(bot, msg, voiceRefreshed(event.Title), nil)
}

func refreshSelectedEvent(bot *api.Bot, chatID int64, event *database.CTFEvent) error {
	if event.ForumTopicID == 0 || event.InitialMessageID == 0 {
		return fmt.Errorf("ctf has no topic")
	}
	_, err := bot.EditHTMLMessageWithKeyboard(chatID, event.InitialMessageID, formatTopicInitialMessage(*event), singleJoinKeyboard(*event))
	return err
}

func startEditWizard(bot *api.Bot, msg *models.Message, eventID int) {
	startWizard(bot, msg, wizardState{kind: "edit", eventID: eventID}, "<b>edit ctf</b>\nReply with the new opening message.", "New opening message")
}

func startWizard(bot *api.Bot, msg *models.Message, state wizardState, text, placeholder string) {
	if msg.From == nil {
		replyHTML(bot, msg, "I could not identify who started this form.", nil)
		return
	}
	prompt, err := sendWizardPrompt(bot, msg.Chat.ID, msg.MessageThreadID, msg.MessageID, text, placeholder)
	if err != nil {
		log.Printf("Error starting CTF wizard: %v", err)
		replyHTML(bot, msg, "Could not start the form. Try again.", nil)
		return
	}
	state.promptMessageID = prompt.MessageID
	state.expires = time.Now().Add(15 * time.Minute)
	setWizard(wizardKey{msg.Chat.ID, msg.MessageThreadID, msg.From.ID}, state)
}

func startCallbackEditWizard(bot *api.Bot, query *models.CallbackQuery, eventID int) bool {
	prompt, err := sendWizardPrompt(
		bot,
		query.Message.Chat.ID,
		query.Message.MessageThreadID,
		query.Message.MessageID,
		"<b>edit ctf</b>\nReply with the new opening message.",
		"New opening message",
	)
	if err != nil {
		log.Printf("Error starting edit wizard: %v", err)
		bot.AnswerCallbackQuery(query.ID, "Could not start the form.")
		return false
	}
	setWizard(
		wizardKey{query.Message.Chat.ID, query.Message.MessageThreadID, query.From.ID},
		wizardState{kind: "edit", eventID: eventID, promptMessageID: prompt.MessageID, expires: time.Now().Add(15 * time.Minute)},
	)
	return true
}

func HandleMessage(bot *api.Bot, msg *models.Message) bool {
	if msg.From == nil || msg.ReplyToMessage == nil {
		return false
	}
	key := wizardKey{msg.Chat.ID, msg.MessageThreadID, msg.From.ID}
	state, ok := getWizard(key)
	if !ok || state.promptMessageID != msg.ReplyToMessage.MessageID {
		return false
	}
	if time.Now().After(state.expires) {
		deleteWizard(key)
		replyHTML(bot, msg, "That form expired. Start it again.", nil)
		return true
	}

	if state.kind == "edit" {
		handleEditReply(bot, msg, key, state)
		return true
	}
	handleNewReply(bot, msg, key, state)
	return true
}

func handleEditReply(bot *api.Bot, msg *models.Message, key wizardKey, state wizardState) {
	body := strings.TrimSpace(msg.Text)
	if body == "" {
		reprompt(bot, msg, key, state, "The opening message cannot be empty.", "New opening message")
		return
	}
	event, err := db.GetCTFEvent(state.eventID)
	if err != nil {
		deleteWizard(key)
		replyHTML(bot, msg, voiceNotFound(), nil)
		return
	}
	if _, err := bot.EditHTMLMessageWithKeyboard(groupOrMessageChatID(msg), event.InitialMessageID, formatCustomInitialMessage(event.Title, body), singleJoinKeyboard(*event)); err != nil {
		log.Printf("Error editing CTF initial message: %v", err)
		replyHTML(bot, msg, voiceEditFailed(), nil)
		return
	}
	deleteWizard(key)
	replyHTML(bot, msg, voiceEdited(event.Title), nil)
}

func handleNewReply(bot *api.Bot, msg *models.Message, key wizardKey, state wizardState) {
	value := strings.TrimSpace(msg.Text)
	switch state.stage {
	case 0:
		if value == "" || len([]rune(value)) > 120 {
			reprompt(bot, msg, key, state, "Use a title between 1 and 120 characters.", "CTF title")
			return
		}
		state.title = value
		state.stage = 1
		reprompt(bot, msg, key, state, "<b>new ctf · 2/5</b>\nStart time? Example: <code>2026-08-01 18:00</code>", "YYYY-MM-DD HH:MM")
	case 1:
		start, err := parseCTFTime(value)
		if err != nil {
			reprompt(bot, msg, key, state, "I could not parse that start time. Use <code>YYYY-MM-DD HH:MM</code>.", "YYYY-MM-DD HH:MM")
			return
		}
		state.start = start
		state.stage = 2
		reprompt(bot, msg, key, state, "<b>new ctf · 3/5</b>\nFinish time?", "YYYY-MM-DD HH:MM")
	case 2:
		finish, err := parseCTFTime(value)
		if err != nil || !finish.After(state.start) {
			reprompt(bot, msg, key, state, "Use a valid finish time after the start time.", "YYYY-MM-DD HH:MM")
			return
		}
		state.finish = finish
		state.stage = 3
		reprompt(bot, msg, key, state, "<b>new ctf · 4/5</b>\nEvent URL? Reply with <code>-</code> to skip.", "https://… or -")
	case 3:
		if value != "-" {
			parsed, err := url.ParseRequestURI(value)
			if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
				reprompt(bot, msg, key, state, "Use a complete http(s) URL, or <code>-</code> to skip.", "https://… or -")
				return
			}
			state.url = value
		}
		state.stage = 4
		reprompt(bot, msg, key, state, "<b>new ctf · 5/5</b>\nShort description? Reply with <code>-</code> to skip.", "Description or -")
	case 4:
		description := value
		if description == "-" {
			description = ""
		}
		event, err := db.UpsertCTFEvent(&database.CTFEvent{
			Source:      "manual",
			Title:       state.title,
			Description: description,
			URL:         state.url,
			StartTime:   state.start,
			FinishTime:  state.finish,
		})
		if err != nil {
			log.Printf("Error saving CTF wizard result: %v", err)
			replyHTML(bot, msg, voiceSavedError(), nil)
			return
		}
		deleteWizard(key)
		targetChatID, targetThreadID := announcementTarget(msg)
		if targetChatID != 0 {
			if _, err := bot.SendHTMLMessageWithKeyboardToThread(targetChatID, targetThreadID, formatSingleAnnouncement(*event), singleJoinKeyboard(*event)); err != nil {
				log.Printf("Error sending manual CTF announcement: %v", err)
			}
		}
		replyHTML(bot, msg, voiceManualCreated(event.Title), nil)
	}
}

func reprompt(bot *api.Bot, msg *models.Message, key wizardKey, state wizardState, text, placeholder string) {
	prompt, err := sendWizardPrompt(bot, msg.Chat.ID, msg.MessageThreadID, msg.MessageID, text, placeholder)
	if err != nil {
		log.Printf("Error continuing CTF wizard: %v", err)
		replyHTML(bot, msg, "Could not continue the form. Try again.", nil)
		return
	}
	state.promptMessageID = prompt.MessageID
	state.expires = time.Now().Add(15 * time.Minute)
	setWizard(key, state)
}

func sendWizardPrompt(bot *api.Bot, chatID int64, threadID, replyTo int, text, placeholder string) (*models.Message, error) {
	return bot.SendMessageWithOptions(api.SendMessageOptions{
		ChatID:                chatID,
		MessageThreadID:       threadID,
		ReplyToMessageID:      replyTo,
		Text:                  text,
		ParseMode:             "HTML",
		ForceReplyPlaceholder: placeholder,
		DisableWebPagePreview: true,
	})
}

func cancelWizardHandler(bot *api.Bot, msg *models.Message, args []string) {
	if msg.From == nil {
		return
	}
	key := wizardKey{msg.Chat.ID, msg.MessageThreadID, msg.From.ID}
	if deleteWizard(key) {
		replyHTML(bot, msg, "Form cancelled.", nil)
		return
	}
	replyHTML(bot, msg, "You have no active form in this topic.", nil)
}

func setWizard(key wizardKey, state wizardState) {
	wizardSessions.Lock()
	now := time.Now()
	for existingKey, existing := range wizardSessions.items {
		if !existing.expires.IsZero() && now.After(existing.expires) {
			delete(wizardSessions.items, existingKey)
		}
	}
	wizardSessions.items[key] = state
	wizardSessions.Unlock()
}

func getWizard(key wizardKey) (wizardState, bool) {
	wizardSessions.Lock()
	state, ok := wizardSessions.items[key]
	wizardSessions.Unlock()
	return state, ok
}

func deleteWizard(key wizardKey) bool {
	wizardSessions.Lock()
	_, ok := wizardSessions.items[key]
	delete(wizardSessions.items, key)
	wizardSessions.Unlock()
	return ok
}
