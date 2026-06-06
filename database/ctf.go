package database

import (
	"database/sql"
	"time"
)

type CTFEvent struct {
	ID                  int
	CTFTimeID           sql.NullInt64
	Source              string
	Title               string
	Description         string
	URL                 string
	CTFTimeURL          string
	Format              string
	Prizes              string
	Restrictions        string
	Location            string
	Logo                string
	Onsite              bool
	CTFTimeParticipants int
	StartTime           time.Time
	FinishTime          time.Time
	ForumTopicID        int
	InitialMessageID    int
	VoteCount           int
}

type CTFParticipant struct {
	UserID    int64
	Username  string
	FirstName string
	LastName  string
}

func (db *DB) UpsertCTFEvent(event *CTFEvent) (*CTFEvent, error) {
	if event.Source == "" {
		event.Source = "ctftime"
	}

	if event.CTFTimeID.Valid {
		var id int
		err := db.conn.QueryRow(`
			INSERT INTO ctf_events (
				ctftime_id, source, title, description, url, ctftime_url, format,
				prizes, restrictions, location, logo, onsite, ctftime_participants,
				start_time, finish_time
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
			ON CONFLICT (ctftime_id)
			DO UPDATE SET
				source = EXCLUDED.source,
				title = EXCLUDED.title,
				description = EXCLUDED.description,
				url = EXCLUDED.url,
				ctftime_url = EXCLUDED.ctftime_url,
				format = EXCLUDED.format,
				prizes = EXCLUDED.prizes,
				restrictions = EXCLUDED.restrictions,
				location = EXCLUDED.location,
				logo = EXCLUDED.logo,
				onsite = EXCLUDED.onsite,
				ctftime_participants = EXCLUDED.ctftime_participants,
				start_time = EXCLUDED.start_time,
				finish_time = EXCLUDED.finish_time,
				updated_at = NOW()
			RETURNING id
		`,
			event.CTFTimeID.Int64,
			event.Source,
			event.Title,
			event.Description,
			event.URL,
			event.CTFTimeURL,
			event.Format,
			event.Prizes,
			event.Restrictions,
			event.Location,
			event.Logo,
			event.Onsite,
			event.CTFTimeParticipants,
			event.StartTime,
			event.FinishTime,
		).Scan(&id)
		if err != nil {
			return nil, err
		}

		return db.GetCTFEvent(id)
	}

	var id int
	err := db.conn.QueryRow(`
		INSERT INTO ctf_events (
			source, title, description, url, ctftime_url, format, prizes,
			restrictions, location, logo, onsite, ctftime_participants,
			start_time, finish_time
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id
	`,
		event.Source,
		event.Title,
		event.Description,
		event.URL,
		event.CTFTimeURL,
		event.Format,
		event.Prizes,
		event.Restrictions,
		event.Location,
		event.Logo,
		event.Onsite,
		event.CTFTimeParticipants,
		event.StartTime,
		event.FinishTime,
	).Scan(&id)
	if err != nil {
		return nil, err
	}

	return db.GetCTFEvent(id)
}

func (db *DB) GetCTFEvent(id int) (*CTFEvent, error) {
	row := db.conn.QueryRow(`
		SELECT
			e.id, e.ctftime_id, e.source, e.title, e.description, e.url, e.ctftime_url,
			e.format, e.prizes, e.restrictions, e.location, e.logo, e.onsite,
			e.ctftime_participants, e.start_time, e.finish_time, e.forum_topic_id,
			e.initial_message_id, COUNT(p.user_id) AS vote_count
		FROM ctf_events e
		LEFT JOIN ctf_participants p ON p.ctf_event_id = e.id
		WHERE e.id = $1
		GROUP BY e.id
	`, id)

	event, err := scanCTFEvent(row)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (db *DB) GetCTFEventByForumTopicID(forumTopicID int) (*CTFEvent, error) {
	row := db.conn.QueryRow(`
		SELECT
			e.id, e.ctftime_id, e.source, e.title, e.description, e.url, e.ctftime_url,
			e.format, e.prizes, e.restrictions, e.location, e.logo, e.onsite,
			e.ctftime_participants, e.start_time, e.finish_time, e.forum_topic_id,
			e.initial_message_id, COUNT(p.user_id) AS vote_count
		FROM ctf_events e
		LEFT JOIN ctf_participants p ON p.ctf_event_id = e.id
		WHERE e.forum_topic_id = $1
		GROUP BY e.id
	`, forumTopicID)

	event, err := scanCTFEvent(row)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (db *DB) ListLiveCTFEvents(now time.Time, limit int) ([]CTFEvent, error) {
	rows, err := db.conn.Query(`
		SELECT
			e.id, e.ctftime_id, e.source, e.title, e.description, e.url, e.ctftime_url,
			e.format, e.prizes, e.restrictions, e.location, e.logo, e.onsite,
			e.ctftime_participants, e.start_time, e.finish_time, e.forum_topic_id,
			e.initial_message_id, COUNT(p.user_id) AS vote_count
		FROM ctf_events e
		LEFT JOIN ctf_participants p ON p.ctf_event_id = e.id
		WHERE e.start_time <= $1 AND e.finish_time > $1 AND e.onsite = FALSE
		GROUP BY e.id
		ORDER BY e.finish_time ASC
		LIMIT $2
	`, now, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCTFEvents(rows)
}

func (db *DB) ListUpcomingCTFEvents(now, until time.Time, limit int) ([]CTFEvent, error) {
	rows, err := db.conn.Query(`
		SELECT
			e.id, e.ctftime_id, e.source, e.title, e.description, e.url, e.ctftime_url,
			e.format, e.prizes, e.restrictions, e.location, e.logo, e.onsite,
			e.ctftime_participants, e.start_time, e.finish_time, e.forum_topic_id,
			e.initial_message_id, COUNT(p.user_id) AS vote_count
		FROM ctf_events e
		LEFT JOIN ctf_participants p ON p.ctf_event_id = e.id
		WHERE e.finish_time > $1 AND e.start_time <= $2 AND e.onsite = FALSE
		GROUP BY e.id
		ORDER BY e.start_time ASC
		LIMIT $3
	`, now, until, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCTFEvents(rows)
}

func (db *DB) ListCTFEventsForReminder(now time.Time, reminderBefore time.Duration, limit int) ([]CTFEvent, error) {
	rows, err := db.conn.Query(`
		SELECT
			e.id, e.ctftime_id, e.source, e.title, e.description, e.url, e.ctftime_url,
			e.format, e.prizes, e.restrictions, e.location, e.logo, e.onsite,
			e.ctftime_participants, e.start_time, e.finish_time, e.forum_topic_id,
			e.initial_message_id, COUNT(p.user_id) AS vote_count
		FROM ctf_events e
		LEFT JOIN ctf_participants p ON p.ctf_event_id = e.id
		WHERE e.start_time > $1
			AND e.start_time <= $2
			AND e.onsite = FALSE
			AND e.forum_topic_id <> 0
			AND NOT EXISTS (
				SELECT 1
				FROM ctf_reminders r
				WHERE r.ctf_event_id = e.id AND r.reminder_key = 'start'
			)
		GROUP BY e.id
		HAVING COUNT(p.user_id) > 0
		ORDER BY e.start_time ASC
		LIMIT $3
	`, now, now.Add(reminderBefore), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCTFEvents(rows)
}

func (db *DB) AddCTFParticipant(eventID int, userID int64, username, firstName, lastName string) (bool, int, error) {
	result, err := db.conn.Exec(`
		INSERT INTO ctf_participants (ctf_event_id, user_id, username, first_name, last_name)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (ctf_event_id, user_id) DO NOTHING
	`, eventID, userID, username, firstName, lastName)
	if err != nil {
		return false, 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, 0, err
	}

	count, err := db.GetCTFParticipantCount(eventID)
	if err != nil {
		return rowsAffected > 0, 0, err
	}

	return rowsAffected > 0, count, nil
}

func (db *DB) RemoveCTFParticipant(eventID int, userID int64) (bool, int, error) {
	result, err := db.conn.Exec(`
		DELETE FROM ctf_participants
		WHERE ctf_event_id = $1 AND user_id = $2
	`, eventID, userID)
	if err != nil {
		return false, 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, 0, err
	}

	count, err := db.GetCTFParticipantCount(eventID)
	if err != nil {
		return rowsAffected > 0, 0, err
	}

	return rowsAffected > 0, count, nil
}

func (db *DB) GetCTFParticipantCount(eventID int) (int, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM ctf_participants WHERE ctf_event_id = $1", eventID).Scan(&count)
	return count, err
}

func (db *DB) ListCTFParticipants(eventID int) ([]CTFParticipant, error) {
	rows, err := db.conn.Query(`
		SELECT user_id, username, first_name, last_name
		FROM ctf_participants
		WHERE ctf_event_id = $1
		ORDER BY joined_at ASC
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []CTFParticipant
	for rows.Next() {
		var participant CTFParticipant
		if err := rows.Scan(
			&participant.UserID,
			&participant.Username,
			&participant.FirstName,
			&participant.LastName,
		); err != nil {
			return nil, err
		}
		participants = append(participants, participant)
	}

	return participants, rows.Err()
}

func (db *DB) SetCTFTopic(eventID int, forumTopicID, initialMessageID int) error {
	_, err := db.conn.Exec(`
		UPDATE ctf_events
		SET forum_topic_id = $2, initial_message_id = $3, updated_at = NOW()
		WHERE id = $1
	`, eventID, forumTopicID, initialMessageID)
	return err
}

func (db *DB) SetCTFInitialMessage(eventID int, initialMessageID int) error {
	_, err := db.conn.Exec(`
		UPDATE ctf_events
		SET initial_message_id = $2, updated_at = NOW()
		WHERE id = $1
	`, eventID, initialMessageID)
	return err
}

func (db *DB) ClearCTFTopic(eventID int) error {
	_, err := db.conn.Exec(`
		UPDATE ctf_events
		SET forum_topic_id = 0, initial_message_id = 0, updated_at = NOW()
		WHERE id = $1
	`, eventID)
	return err
}

func (db *DB) ReserveDailyCTFAnnouncement(day time.Time, chatID int64, threadID int) (bool, error) {
	result, err := db.conn.Exec(`
		INSERT INTO ctf_daily_announcements (announcement_date, chat_id, thread_id)
		VALUES ($1::date, $2, $3)
		ON CONFLICT (announcement_date) DO NOTHING
	`, day, chatID, threadID)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

func (db *DB) CompleteDailyCTFAnnouncement(day time.Time, messageID int) error {
	_, err := db.conn.Exec(`
		UPDATE ctf_daily_announcements
		SET message_id = $2, sent_at = NOW()
		WHERE announcement_date = $1::date
	`, day, messageID)
	return err
}

func (db *DB) ClearDailyCTFAnnouncement(day time.Time) error {
	_, err := db.conn.Exec("DELETE FROM ctf_daily_announcements WHERE announcement_date = $1::date", day)
	return err
}

func (db *DB) ReserveCTFReminder(eventID int, reminderKey string) (bool, error) {
	result, err := db.conn.Exec(`
		INSERT INTO ctf_reminders (ctf_event_id, reminder_key)
		VALUES ($1, $2)
		ON CONFLICT (ctf_event_id, reminder_key) DO NOTHING
	`, eventID, reminderKey)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

func (db *DB) ClearCTFReminder(eventID int, reminderKey string) error {
	_, err := db.conn.Exec(`
		DELETE FROM ctf_reminders
		WHERE ctf_event_id = $1 AND reminder_key = $2
	`, eventID, reminderKey)
	return err
}

type ctfScanner interface {
	Scan(dest ...interface{}) error
}

func scanCTFEvent(scanner ctfScanner) (*CTFEvent, error) {
	var event CTFEvent
	err := scanner.Scan(
		&event.ID,
		&event.CTFTimeID,
		&event.Source,
		&event.Title,
		&event.Description,
		&event.URL,
		&event.CTFTimeURL,
		&event.Format,
		&event.Prizes,
		&event.Restrictions,
		&event.Location,
		&event.Logo,
		&event.Onsite,
		&event.CTFTimeParticipants,
		&event.StartTime,
		&event.FinishTime,
		&event.ForumTopicID,
		&event.InitialMessageID,
		&event.VoteCount,
	)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

func scanCTFEvents(rows *sql.Rows) ([]CTFEvent, error) {
	var events []CTFEvent
	for rows.Next() {
		event, err := scanCTFEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, *event)
	}

	return events, rows.Err()
}
