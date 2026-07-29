package database

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type DB struct {
	conn *sql.DB
}

func New(connectionString string) (*DB, error) {
	conn, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(4)
	conn.SetMaxIdleConns(1)
	conn.SetConnMaxIdleTime(2 * time.Minute)
	conn.SetConnMaxLifetime(30 * time.Minute)

	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, err
	}

	db := &DB{conn: conn}
	if err := db.createTables(); err != nil {
		conn.Close()
		return nil, err
	}

	log.Println("Database connected and tables ready")
	return db, nil
}

func (db *DB) createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		user_id BIGINT PRIMARY KEY,
		username TEXT,
		first_name TEXT,
		last_name TEXT,
		first_seen TIMESTAMP DEFAULT NOW(),
		last_seen TIMESTAMP DEFAULT NOW(),
		message_count INT DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS groups (
		group_id BIGINT PRIMARY KEY,
		title TEXT,
		type TEXT,
		first_seen TIMESTAMP DEFAULT NOW(),
		last_seen TIMESTAMP DEFAULT NOW(),
		message_count INT DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS ctf_events (
		id SERIAL PRIMARY KEY,
		ctftime_id INT UNIQUE,
		source TEXT NOT NULL DEFAULT 'ctftime',
		title TEXT NOT NULL,
		description TEXT DEFAULT '',
		url TEXT DEFAULT '',
		ctftime_url TEXT DEFAULT '',
		format TEXT DEFAULT '',
		prizes TEXT DEFAULT '',
		restrictions TEXT DEFAULT '',
		location TEXT DEFAULT '',
		logo TEXT DEFAULT '',
		onsite BOOLEAN DEFAULT FALSE,
		ctftime_participants INT DEFAULT 0,
		start_time TIMESTAMPTZ NOT NULL,
		finish_time TIMESTAMPTZ NOT NULL,
		forum_topic_id INT DEFAULT 0,
		initial_message_id INT DEFAULT 0,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_ctf_events_start_time ON ctf_events(start_time);
	CREATE INDEX IF NOT EXISTS idx_ctf_events_finish_time ON ctf_events(finish_time);

	CREATE TABLE IF NOT EXISTS ctf_participants (
		ctf_event_id INT REFERENCES ctf_events(id) ON DELETE CASCADE,
		user_id BIGINT,
		username TEXT,
		first_name TEXT,
		last_name TEXT,
		joined_at TIMESTAMPTZ DEFAULT NOW(),
		PRIMARY KEY (ctf_event_id, user_id)
	);

	CREATE TABLE IF NOT EXISTS ctf_daily_announcements (
		announcement_date DATE PRIMARY KEY,
		chat_id BIGINT NOT NULL,
		thread_id INT NOT NULL,
		message_id INT DEFAULT 0,
		sent_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS ctf_reminders (
		ctf_event_id INT REFERENCES ctf_events(id) ON DELETE CASCADE,
		reminder_key TEXT NOT NULL DEFAULT 'start',
		sent_at TIMESTAMPTZ DEFAULT NOW(),
		PRIMARY KEY (ctf_event_id, reminder_key)
	);

	CREATE TABLE IF NOT EXISTS reaction_emojis (
		custom_emoji_id TEXT PRIMARY KEY,
		added_by BIGINT NOT NULL,
		added_at TIMESTAMPTZ DEFAULT NOW()
	);
	`

	_, err := db.conn.Exec(schema)
	return err
}

func (db *DB) AddReactionEmoji(customEmojiID string, addedBy int64) error {
	_, err := db.conn.Exec(`
		INSERT INTO reaction_emojis (custom_emoji_id, added_by)
		VALUES ($1, $2)
		ON CONFLICT (custom_emoji_id) DO NOTHING
	`, customEmojiID, addedBy)
	return err
}

func (db *DB) RemoveReactionEmoji(customEmojiID string) error {
	_, err := db.conn.Exec("DELETE FROM reaction_emojis WHERE custom_emoji_id = $1", customEmojiID)
	return err
}

func (db *DB) ClearReactionEmojis() error {
	_, err := db.conn.Exec("DELETE FROM reaction_emojis")
	return err
}

func (db *DB) ListReactionEmojis() ([]string, error) {
	rows, err := db.conn.Query("SELECT custom_emoji_id FROM reaction_emojis ORDER BY added_at")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (db *DB) UpsertUser(userID int64, username, firstName, lastName string) error {
	_, err := db.conn.Exec(`
		INSERT INTO users (user_id, username, first_name, last_name, last_seen, message_count)
		VALUES ($1, $2, $3, $4, NOW(), 1)
		ON CONFLICT (user_id)
		DO UPDATE SET
			username = EXCLUDED.username,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			last_seen = NOW(),
			message_count = users.message_count + 1
	`, userID, username, firstName, lastName)
	return err
}

func (db *DB) UpsertGroup(groupID int64, title, groupType string) error {
	_, err := db.conn.Exec(`
		INSERT INTO groups (group_id, title, type, last_seen, message_count)
		VALUES ($1, $2, $3, NOW(), 1)
		ON CONFLICT (group_id)
		DO UPDATE SET
			title = EXCLUDED.title,
			type = EXCLUDED.type,
			last_seen = NOW(),
			message_count = groups.message_count + 1
	`, groupID, title, groupType)
	return err
}

func (db *DB) GetTotalUsers() (int, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

func (db *DB) GetTotalGroups() (int, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM groups").Scan(&count)
	return count, err
}

type GroupStats struct {
	GroupID      int64
	Title        string
	MessageCount int
}

type UserInfo struct {
	ID        int64
	Username  string
	FirstName string
	LastName  string
}

func (db *DB) GetUser(userID int64) (*UserInfo, error) {
	var user UserInfo
	err := db.conn.QueryRow(`
		SELECT user_id, COALESCE(username, ''), COALESCE(first_name, ''), COALESCE(last_name, '')
		FROM users
		WHERE user_id = $1
	`, userID).Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *DB) GetTop5Groups() ([]GroupStats, error) {
	rows, err := db.conn.Query(`
		SELECT group_id, title, message_count
		FROM groups
		ORDER BY message_count DESC
		LIMIT 5
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []GroupStats
	for rows.Next() {
		var g GroupStats
		if err := rows.Scan(&g.GroupID, &g.Title, &g.MessageCount); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}

	return groups, rows.Err()
}

func (db *DB) Close() error {
	return db.conn.Close()
}
