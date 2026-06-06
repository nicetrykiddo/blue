package api

import (
	"blue/models"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (b *Bot) SetWebhook(url, secretToken string) error {
	payload := map[string]interface{}{
		"url":             url,
		"secret_token":    secretToken,
		"allowed_updates": []string{"message", "callback_query"},
	}

	return b.sendBoolRequest("/setWebhook", payload)
}

func ParseWebhookUpdate(r *http.Request, secretToken string) (*models.Update, error) {
	if secretToken == "" {
		return nil, fmt.Errorf("missing configured secret token")
	}

	gotSecret := r.Header.Get("X-Telegram-Bot-Api-Secret-Token")
	if subtle.ConstantTimeCompare([]byte(gotSecret), []byte(secretToken)) != 1 {
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
