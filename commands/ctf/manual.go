package ctf

import (
	"blue/database"
	"fmt"
	"strings"
	"time"
)

func parseManualCTF(input string) (*database.CTFEvent, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("missing CTF details")
	}

	fields := parseManualFields(input)
	title := firstNonEmpty(fields["title"], fields["name"])
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}

	startTime, err := parseCTFTime(fields["start"])
	if err != nil {
		return nil, fmt.Errorf("start time is invalid")
	}
	finishTime, err := parseCTFTime(firstNonEmpty(fields["finish"], fields["end"]))
	if err != nil {
		return nil, fmt.Errorf("finish time is invalid")
	}
	if !finishTime.After(startTime) {
		return nil, fmt.Errorf("finish time must be after start time")
	}

	return &database.CTFEvent{
		Source:       "manual",
		Title:        title,
		Description:  firstNonEmpty(fields["description"], fields["info"]),
		URL:          fields["url"],
		Format:       fields["format"],
		Prizes:       fields["prizes"],
		Restrictions: fields["restrictions"],
		Location:     fields["location"],
		Onsite:       false,
		StartTime:    startTime,
		FinishTime:   finishTime,
	}, nil
}

func parseManualFields(input string) map[string]string {
	fields := map[string]string{}
	if strings.Contains(input, "|") {
		parts := strings.Split(input, "|")
		if len(parts) > 0 {
			fields["title"] = strings.TrimSpace(parts[0])
		}
		if len(parts) > 1 {
			fields["start"] = strings.TrimSpace(parts[1])
		}
		if len(parts) > 2 {
			fields["finish"] = strings.TrimSpace(parts[2])
		}
		if len(parts) > 3 {
			fields["url"] = strings.TrimSpace(parts[3])
		}
		if len(parts) > 4 {
			fields["format"] = strings.TrimSpace(parts[4])
		}
		if len(parts) > 5 {
			fields["prizes"] = strings.TrimSpace(parts[5])
		}
		if len(parts) > 6 {
			fields["description"] = strings.TrimSpace(strings.Join(parts[6:], "|"))
		}
		return fields
	}

	for _, line := range strings.Split(input, "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		fields[strings.ToLower(strings.TrimSpace(key))] = strings.TrimSpace(value)
	}

	return fields
}

func parseCTFTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, fmt.Errorf("empty time")
	}

	formatsWithZone := []string{
		time.RFC3339,
		"2006-01-02 15:04 MST",
		"2006-01-02 15:04 -0700",
		"2006-01-02 15:04:05 MST",
		"2006-01-02 15:04:05 -0700",
	}
	for _, layout := range formatsWithZone {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, nil
		}
	}

	formatsLocal := []string{
		"2006-01-02 15:04",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	loc := appLocation()
	for _, layout := range formatsLocal {
		if parsed, err := time.ParseInLocation(layout, value, loc); err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported time format")
}

func addUsage(reason error) string {
	return fmt.Sprintf(`Couldn't put that CTF on the board: %s

Use either:
<code>/ctfadd Title | 2026-06-20 18:00 | 2026-06-21 18:00 | https://example.com | Jeopardy | Prizes | Description</code>

Or:
<code>/ctfadd
Title: Example CTF
Start: 2026-06-20 18:00
Finish: 2026-06-21 18:00
URL: https://example.com
Format: Jeopardy
Prizes: Swag
Description: Short info</code>`, safe(reason.Error()))
}
