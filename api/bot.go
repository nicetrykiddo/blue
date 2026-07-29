package api

import (
	"net/http"
	"time"
)

const baseURL = "https://api.telegram.org/bot"

type Bot struct {
	token  string
	apiURL string
	client *http.Client
}

func NewBot(token string) *Bot {
	return &Bot{
		token:  token,
		apiURL: baseURL + token,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}
