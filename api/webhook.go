package api

import (
	"blue/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (b *Bot) SetWebhook(url, secretToken string) error {
	payload := map[string]interface{}{
		"url": url,
	}

	if secretToken != "" {
		payload["secret_token"] = secretToken
	}

	return b.sendBoolRequest("/setWebhook", payload)
}

func (b *Bot) DeleteWebhook() error {
	return b.sendBoolRequest("/deleteWebhook", map[string]interface{}{})
}

func ParseWebhookUpdate(r *http.Request, secretToken string) (*models.Update, error) {
	if secretToken != "" && r.Header.Get("X-Telegram-Bot-Api-Secret-Token") != secretToken {
		return nil, fmt.Errorf("invalid secret token")
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var update models.Update
	if err := json.Unmarshal(body, &update); err != nil {
		return nil, err
	}

	return &update, nil
}
