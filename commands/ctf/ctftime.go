package ctf

import (
	"blue/database"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ctftimeEvent struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	URL          string `json:"url"`
	CTFTimeURL   string `json:"ctftime_url"`
	Format       string `json:"format"`
	Prizes       string `json:"prizes"`
	Restrictions string `json:"restrictions"`
	Location     string `json:"location"`
	Logo         string `json:"logo"`
	Onsite       bool   `json:"onsite"`
	Participants int    `json:"participants"`
	Start        string `json:"start"`
	Finish       string `json:"finish"`
}

var ctfSync = struct {
	sync.Mutex
	last map[string]time.Time
}{last: make(map[string]time.Time)}

func refreshUpcomingEvents(now time.Time, days int, limit int) ([]database.CTFEvent, error) {
	return refreshEvents(now.Add(-2*time.Hour), now.AddDate(0, 0, days), limit)
}

func refreshEventsIfStale(key string, start, finish time.Time, limit int) ([]database.CTFEvent, error) {
	ctfSync.Lock()
	defer ctfSync.Unlock()
	if time.Since(ctfSync.last[key]) < 15*time.Minute {
		return nil, nil
	}
	events, err := refreshEvents(start, finish, limit)
	if err == nil {
		ctfSync.last[key] = time.Now()
	}
	return events, err
}

func refreshUpcomingEventsIfStale(now time.Time, days, limit int) ([]database.CTFEvent, error) {
	return refreshEventsIfStale("upcoming", now.Add(-2*time.Hour), now.AddDate(0, 0, days), limit)
}

func refreshEvents(start, finish time.Time, limit int) ([]database.CTFEvent, error) {
	if db == nil {
		return nil, fmt.Errorf("ctf storage is not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	events, err := fetchCTFTimeEvents(ctx, start.UTC(), finish.UTC(), limit)
	if err != nil {
		return nil, err
	}

	saved := make([]database.CTFEvent, 0, len(events))
	for i := range events {
		event, err := db.UpsertCTFEvent(&events[i])
		if err != nil {
			log.Printf("Error saving CTFtime event %s: %v", events[i].Title, err)
			continue
		}
		saved = append(saved, *event)
	}

	return saved, nil
}

func fetchCTFTimeEvents(ctx context.Context, start, finish time.Time, limit int) ([]database.CTFEvent, error) {
	values := url.Values{}
	values.Set("limit", strconv.Itoa(limit))
	values.Set("start", strconv.FormatInt(start.Unix(), 10))
	values.Set("finish", strconv.FormatInt(finish.Unix(), 10))

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, ctftimeEventsURL+"?"+values.Encode(), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "maple/1.0")
	request.Header.Set("Accept", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("ctftime returned HTTP %d", response.StatusCode)
	}

	var remote []ctftimeEvent
	if err := json.NewDecoder(response.Body).Decode(&remote); err != nil {
		return nil, err
	}

	events := make([]database.CTFEvent, 0, len(remote))
	for _, item := range remote {
		if item.Onsite {
			continue
		}

		startTime, err := time.Parse(time.RFC3339, item.Start)
		if err != nil {
			log.Printf("Skipping CTFtime event %q with bad start time: %v", item.Title, err)
			continue
		}
		finishTime, err := time.Parse(time.RFC3339, item.Finish)
		if err != nil {
			log.Printf("Skipping CTFtime event %q with bad finish time: %v", item.Title, err)
			continue
		}

		events = append(events, database.CTFEvent{
			CTFTimeID:           sql.NullInt64{Int64: int64(item.ID), Valid: true},
			Source:              "ctftime",
			Title:               strings.TrimSpace(item.Title),
			Description:         strings.TrimSpace(item.Description),
			URL:                 strings.TrimSpace(item.URL),
			CTFTimeURL:          strings.TrimSpace(item.CTFTimeURL),
			Format:              strings.TrimSpace(item.Format),
			Prizes:              strings.TrimSpace(item.Prizes),
			Restrictions:        strings.TrimSpace(item.Restrictions),
			Location:            strings.TrimSpace(item.Location),
			Logo:                strings.TrimSpace(item.Logo),
			Onsite:              item.Onsite,
			CTFTimeParticipants: item.Participants,
			StartTime:           startTime,
			FinishTime:          finishTime,
		})
	}

	return events, nil
}
