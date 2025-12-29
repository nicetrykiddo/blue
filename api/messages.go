package api

import (
	"blue/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func (b *Bot) SendMessage(chatID int64, text string) (*models.Message, error) {
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "HTML",
	}

	return b.sendMessageRequest("/sendMessage", payload)
}

func (b *Bot) ReplyToMessage(chatID int64, messageID int, text string) (*models.Message, error) {
	payload := map[string]interface{}{
		"chat_id":             chatID,
		"text":                text,
		"reply_to_message_id": messageID,
		"parse_mode":          "HTML",
	}

	return b.sendMessageRequest("/sendMessage", payload)
}

func (b *Bot) EditMessage(chatID int64, messageID int, text string) (*models.Message, error) {
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"message_id": messageID,
		"text":       text,
		"parse_mode": "HTML",
	}

	return b.sendMessageRequest("/editMessageText", payload)
}

func (b *Bot) DeleteMessage(chatID int64, messageID int) error {
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"message_id": messageID,
	}

	return b.sendBoolRequest("/deleteMessage", payload)
}

func (b *Bot) SendPhoto(chatID int64, photoURL string, caption string) (*models.Message, error) {
	payload := map[string]interface{}{
		"chat_id": chatID,
		"photo":   photoURL,
	}
	
	if caption != "" {
		payload["caption"] = caption
	}

	return b.sendMessageRequest("/sendPhoto", payload)
}

func (b *Bot) SendDocument(chatID int64, documentURL string, caption string) (*models.Message, error) {
	payload := map[string]interface{}{
		"chat_id":  chatID,
		"document": documentURL,
	}
	
	if caption != "" {
		payload["caption"] = caption
	}

	return b.sendMessageRequest("/sendDocument", payload)
}

func (b *Bot) SendSticker(chatID int64, stickerID string) (*models.Message, error) {
	payload := map[string]interface{}{
		"chat_id": chatID,
		"sticker": stickerID,
	}

	return b.sendMessageRequest("/sendSticker", payload)
}

func (b *Bot) SendChatAction(chatID int64, action string) error {
	payload := map[string]interface{}{
		"chat_id": chatID,
		"action":  action,
	}

	return b.sendBoolRequest("/sendChatAction", payload)
}

func (b *Bot) sendMessageRequest(endpoint string, payload map[string]interface{}) (*models.Message, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := b.client.Post(
		b.apiURL+endpoint,
		"application/json",
		bytes.NewBuffer(data),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result models.SendMessageResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if !result.OK {
		return nil, fmt.Errorf("telegram API request failed: %s", result.Description)
	}

	return &result.Result, nil
}

func (b *Bot) sendBoolRequest(endpoint string, payload map[string]interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := b.client.Post(
		b.apiURL+endpoint,
		"application/json",
		bytes.NewBuffer(data),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result struct {
		OK          bool   `json:"ok"`
		Description string `json:"description,omitempty"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if !result.OK {
		return fmt.Errorf("telegram API request failed: %s", result.Description)
	}

	return nil
}
