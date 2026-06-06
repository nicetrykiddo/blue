package ctf

import (
	"blue/api"
	"blue/config"
	"blue/database"
	"fmt"
	"log"
	"time"
)

func StartDailyScheduler(bot *api.Bot, database *database.DB, config *config.Config) {
	SetServices(database, config)

	if config.GroupID == 0 || config.CTFTopicID == 0 {
		log.Println("CTF schedulers disabled: GROUP_ID or CTF_TOPIC_ID missing")
		return
	}

	startDailyDigestScheduler(bot)
	startReminderScheduler(bot)
}

func startDailyDigestScheduler(bot *api.Bot) {
	go func() {
		for {
			loc := appLocation()
			next := nextDailyRun(time.Now().In(loc), cfg.CTFDailyHour, cfg.CTFDailyMinute, loc)
			timer := time.NewTimer(time.Until(next))
			<-timer.C

			if err := SendDailyDigest(bot, false); err != nil {
				log.Printf("Error sending daily CTF digest: %v", err)
			}
		}
	}()
}

func startReminderScheduler(bot *api.Bot) {
	go func() {
		timer := time.NewTimer(15 * time.Second)
		for {
			<-timer.C

			if err := SendDueReminders(bot); err != nil {
				log.Printf("Error sending CTF reminders: %v", err)
			}

			timer.Reset(time.Minute)
		}
	}()
}

func SendDailyDigest(bot *api.Bot, force bool) error {
	if db == nil || cfg == nil {
		return fmt.Errorf("ctf services are not initialized")
	}
	if cfg.GroupID == 0 {
		return fmt.Errorf("GROUP_ID is required for CTF daily digest")
	}

	loc := appLocation()
	now := time.Now().In(loc)
	day := dateOnly(now)

	if !force {
		reserved, err := db.ReserveDailyCTFAnnouncement(day, cfg.GroupID, cfg.CTFTopicID)
		if err != nil {
			return err
		}
		if !reserved {
			return nil
		}
	}

	events, err := refreshUpcomingEvents(now, cfg.CTFDailyLookaheadDays, 80)
	if err != nil {
		log.Printf("Error refreshing CTFtime events, using cached events: %v", err)
	}

	if len(events) == 0 {
		events, err = db.ListUpcomingCTFEvents(now.UTC(), now.AddDate(0, 0, cfg.CTFDailyLookaheadDays).UTC(), 8)
		if err != nil {
			clearDailyReservation(force, day)
			return err
		}
	}

	message, err := bot.SendHTMLMessageWithKeyboardToThread(
		cfg.GroupID,
		cfg.CTFTopicID,
		formatDailyDigest(events, now),
		joinKeyboard(events),
	)
	if err != nil {
		clearDailyReservation(force, day)
		return err
	}

	if !force {
		return db.CompleteDailyCTFAnnouncement(day, message.MessageID)
	}

	return nil
}

func SendDueReminders(bot *api.Bot) error {
	if db == nil || cfg == nil || cfg.GroupID == 0 {
		return nil
	}

	reminderBefore := time.Duration(cfg.CTFReminderMinutes) * time.Minute
	if reminderBefore <= 0 {
		return nil
	}

	events, err := db.ListCTFEventsForReminder(time.Now().UTC(), reminderBefore, 20)
	if err != nil {
		return err
	}

	for i := range events {
		event := events[i]
		reserved, err := db.ReserveCTFReminder(event.ID, reminderKeyStart)
		if err != nil {
			log.Printf("Error reserving reminder for CTF %d: %v", event.ID, err)
			continue
		}
		if !reserved {
			continue
		}

		participants, err := db.ListCTFParticipants(event.ID)
		if err != nil {
			db.ClearCTFReminder(event.ID, reminderKeyStart)
			log.Printf("Error loading CTF participants for %d: %v", event.ID, err)
			continue
		}

		if _, err := bot.SendHTMLMessageWithKeyboardToThread(
			cfg.GroupID,
			event.ForumTopicID,
			formatStartReminder(event, participants),
			singleJoinKeyboard(event),
		); err != nil {
			db.ClearCTFReminder(event.ID, reminderKeyStart)
			log.Printf("Error sending CTF reminder for %d: %v", event.ID, err)
		}
	}

	return nil
}

func clearDailyReservation(force bool, day time.Time) {
	if !force {
		db.ClearDailyCTFAnnouncement(day)
	}
}
