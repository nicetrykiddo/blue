package ctf

import (
	"blue/database"
	"blue/models"
	"fmt"
	"html"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func parseIDAndBody(input string) (int, string, error) {
	parts := strings.Fields(input)
	if len(parts) < 2 {
		return 0, "", fmt.Errorf("missing id or body")
	}

	id, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", err
	}

	body := strings.TrimSpace(strings.TrimPrefix(input, parts[0]))
	if body == "" {
		return 0, "", fmt.Errorf("missing body")
	}

	return id, body, nil
}

func parseFirstID(args []string) (int, error) {
	if len(args) == 0 {
		return 0, fmt.Errorf("missing id")
	}
	return strconv.Atoi(args[0])
}

func commandPayload(msg *models.Message) string {
	text := strings.TrimSpace(msg.Text)
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(text, parts[0]))
}

func announcementTarget(msg *models.Message) (int64, int) {
	if cfg != nil && cfg.GroupID != 0 {
		return cfg.GroupID, cfg.CTFTopicID
	}
	return msg.Chat.ID, msg.MessageThreadID
}

func groupOrMessageChatID(msg *models.Message) int64 {
	if cfg != nil && cfg.GroupID != 0 {
		return cfg.GroupID
	}
	return msg.Chat.ID
}

func canManageCTFs(user *models.User) bool {
	return cfg != nil && user != nil && cfg.AdminUserIDs[user.ID]
}

func defaultLookaheadDays() int {
	if cfg != nil && cfg.CTFDailyLookaheadDays > 0 {
		return cfg.CTFDailyLookaheadDays
	}
	return 14
}

func appLocation() *time.Location {
	if cfg != nil && cfg.Timezone != "" {
		if loc, err := time.LoadLocation(cfg.Timezone); err == nil {
			return loc
		}
	}
	return time.Local
}

func nextDailyRun(now time.Time, hour, minute int, loc *time.Location) time.Time {
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, loc)
	if !next.After(now) {
		next = next.AddDate(0, 0, 1)
	}
	return next
}

func dateOnly(t time.Time) time.Time {
	loc := appLocation()
	local := t.In(loc)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
}

func topicName(title string) string {
	name := strings.TrimSpace(title)
	if name == "" {
		name = "CTF"
	}
	if len([]rune(name)) > 120 {
		name = string([]rune(name)[:120])
	}
	return name
}

func formatDuration(start, finish time.Time) string {
	duration := finish.Sub(start)
	if duration <= 0 {
		return "unknown"
	}

	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	parts := []string{}
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 && days == 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}
	if len(parts) == 0 {
		return "less than 1m"
	}

	return strings.Join(parts, " ")
}

func eventLocation(event database.CTFEvent) string {
	if event.Location != "" {
		return event.Location
	}
	return "Online"
}

func participantMentions(participants []database.CTFParticipant, limit int) string {
	limit = min(len(participants), limit)
	mentions := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		mentions = append(mentions, participantMention(participants[i]))
	}
	if len(participants) > limit {
		mentions = append(mentions, "...")
	}
	return strings.Join(mentions, ", ")
}

func participantMention(participant database.CTFParticipant) string {
	name := strings.TrimSpace(strings.Join([]string{participant.FirstName, participant.LastName}, " "))
	if name == "" && participant.Username != "" {
		name = "@" + participant.Username
	}
	if name == "" {
		name = strconv.FormatInt(participant.UserID, 10)
	}
	return fmt.Sprintf("<a href=\"tg://user?id=%d\">%s</a>", participant.UserID, safe(name))
}

func userMention(user *models.User) string {
	name := strings.TrimSpace(strings.Join([]string{user.FirstName, user.LastName}, " "))
	if name == "" && user.Username != "" {
		name = "@" + user.Username
	}
	if name == "" {
		name = strconv.FormatInt(user.ID, 10)
	}
	return fmt.Sprintf("<a href=\"tg://user?id=%d\">%s</a>", user.ID, safe(name))
}

func safe(value string) string {
	return html.EscapeString(strings.TrimSpace(value))
}

func safeURL(value string) string {
	return html.EscapeString(httpURL(value))
}

func buttonURL(value string) string {
	return httpURL(value)
}

func httpURL(value string) string {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return ""
	}
	return parsed.String()
}

func emptyDash(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "-"
	}
	return safe(value)
}

func truncate(value string, max int) string {
	value = strings.Join(strings.Fields(value), " ")
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max-3]) + "..."
}

func trimTelegramMessage(value string) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= 3900 {
		return string(runes)
	}
	return string(runes[:3897]) + "..."
}

func parseBool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "onsite":
		return true
	default:
		return false
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
