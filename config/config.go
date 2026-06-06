package config

import (
	"bufio"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	BotToken              string
	GroupID               int64
	DatabaseURL           string
	WebhookURL            string
	WebhookSecret         string
	WebhookPath           string
	ListenAddr            string
	LogChatID             int64
	CTFTopicID            int
	CTFDailyHour          int
	CTFDailyMinute        int
	CTFDailyLookaheadDays int
	CTFTopicVoteThreshold int
	CTFReminderMinutes    int
	Timezone              string
	AdminUserIDs          map[int64]bool
}

func Load() *Config {
	loadEnvFile(".env")

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN not found in .env file or environment variables")
	}

	var groupID int64
	if idStr := os.Getenv("GROUP_ID"); idStr != "" {
		if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
			groupID = id
		}
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL not found in .env file or environment variables")
	}

	webhookURL := os.Getenv("WEBHOOK_URL")
	if webhookURL == "" {
		log.Fatal("WEBHOOK_URL is required. This bot is webhook-only.")
	}
	parsedWebhookURL, err := url.Parse(webhookURL)
	if err != nil || parsedWebhookURL.Scheme != "https" || parsedWebhookURL.Host == "" {
		log.Fatal("WEBHOOK_URL must be a public HTTPS URL")
	}

	webhookSecret := os.Getenv("WEBHOOK_SECRET")
	if !validWebhookSecret(webhookSecret) {
		log.Fatal("WEBHOOK_SECRET must be 32-256 chars and contain only A-Z, a-z, 0-9, _ or -")
	}

	webhookPath := os.Getenv("WEBHOOK_PATH")
	if webhookPath == "" {
		webhookPath = parsedWebhookURL.Path
	}
	if webhookPath == "" || webhookPath == "/" || !strings.HasPrefix(webhookPath, "/") {
		log.Fatal("WEBHOOK_PATH must start with / and should not be /")
	}
	if parsedWebhookURL.Path != webhookPath {
		log.Fatal("WEBHOOK_PATH must match the path portion of WEBHOOK_URL")
	}

	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = "127.0.0.1:8080"
	}

	timezone := os.Getenv("TIMEZONE")
	if timezone == "" {
		timezone = "Asia/Kolkata"
	}

	return &Config{
		BotToken:              token,
		GroupID:               groupID,
		DatabaseURL:           databaseURL,
		WebhookURL:            webhookURL,
		WebhookSecret:         webhookSecret,
		WebhookPath:           webhookPath,
		ListenAddr:            listenAddr,
		LogChatID:             getEnvInt64("LOG_CHAT_ID", 0),
		CTFTopicID:            getEnvInt("CTF_TOPIC_ID", 0),
		CTFDailyHour:          getEnvInt("CTF_DAILY_HOUR", 0),
		CTFDailyMinute:        getEnvInt("CTF_DAILY_MINUTE", 0),
		CTFDailyLookaheadDays: getEnvInt("CTF_DAILY_LOOKAHEAD_DAYS", 14),
		CTFTopicVoteThreshold: getEnvInt("CTF_TOPIC_VOTE_THRESHOLD", 1),
		CTFReminderMinutes:    getEnvInt("CTF_REMINDER_MINUTES_BEFORE", 60),
		Timezone:              timezone,
		AdminUserIDs:          parseIDSet(os.Getenv("ADMIN_USER_IDS")),
	}
}

func validWebhookSecret(secret string) bool {
	if len(secret) < 32 || len(secret) > 256 {
		return false
	}

	for _, char := range secret {
		if char >= 'a' && char <= 'z' {
			continue
		}
		if char >= 'A' && char <= 'Z' {
			continue
		}
		if char >= '0' && char <= '9' {
			continue
		}
		if char == '_' || char == '-' {
			continue
		}
		return false
	}

	return true
}

func getEnvInt64(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func parseIDSet(value string) map[int64]bool {
	ids := make(map[int64]bool)
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		id, err := strconv.ParseInt(part, 10, 64)
		if err == nil {
			ids[id] = true
		}
	}

	return ids
}

func loadEnvFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		value = strings.Trim(value, `"'`)

		os.Setenv(key, value)
	}
}
