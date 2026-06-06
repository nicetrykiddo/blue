package api

import (
	"blue/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

type SendMessageOptions struct {
	ChatID                int64
	Text                  string
	ParseMode             string
	MessageThreadID       int
	ReplyToMessageID      int
	ReplyMarkup           *models.InlineKeyboardMarkup
	DisableWebPagePreview bool
}

type EditMessageOptions struct {
	ChatID                int64
	MessageID             int
	Text                  string
	ParseMode             string
	ReplyMarkup           *models.InlineKeyboardMarkup
	DisableWebPagePreview bool
}

func (b *Bot) SendMessage(chatID int64, text string) (*models.Message, error) {
	return b.SendMessageWithOptions(SendMessageOptions{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "Markdown",
	})
}

func (b *Bot) ReplyToMessage(chatID int64, messageID int, text string) (*models.Message, error) {
	return b.SendMessageWithOptions(SendMessageOptions{
		ChatID:           chatID,
		Text:             text,
		ParseMode:        "Markdown",
		ReplyToMessageID: messageID,
	})
}

func (b *Bot) EditMessage(chatID int64, messageID int, text string) (*models.Message, error) {
	return b.EditMessageWithOptions(EditMessageOptions{
		ChatID:    chatID,
		MessageID: messageID,
		Text:      text,
		ParseMode: "Markdown",
	})
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

func (b *Bot) SendHTMLMessage(chatID int64, text string) (*models.Message, error) {
	return b.SendMessageWithOptions(SendMessageOptions{
		ChatID:                chatID,
		Text:                  text,
		ParseMode:             "HTML",
		DisableWebPagePreview: true,
	})
}

func (b *Bot) SendHTMLMessageToThread(chatID int64, threadID int, text string) (*models.Message, error) {
	return b.SendMessageWithOptions(SendMessageOptions{
		ChatID:                chatID,
		MessageThreadID:       threadID,
		Text:                  text,
		ParseMode:             "HTML",
		DisableWebPagePreview: true,
	})
}

func (b *Bot) ReplyHTMLToMessageInThread(chatID int64, threadID, messageID int, text string) (*models.Message, error) {
	return b.SendMessageWithOptions(SendMessageOptions{
		ChatID:                chatID,
		MessageThreadID:       threadID,
		ReplyToMessageID:      messageID,
		Text:                  text,
		ParseMode:             "HTML",
		DisableWebPagePreview: true,
	})
}

func (b *Bot) SendHTMLMessageWithKeyboard(chatID int64, text string, keyboard *models.InlineKeyboardMarkup) (*models.Message, error) {
	return b.SendMessageWithOptions(SendMessageOptions{
		ChatID:                chatID,
		Text:                  text,
		ParseMode:             "HTML",
		ReplyMarkup:           keyboard,
		DisableWebPagePreview: true,
	})
}

func (b *Bot) SendHTMLMessageWithKeyboardToThread(chatID int64, threadID int, text string, keyboard *models.InlineKeyboardMarkup) (*models.Message, error) {
	return b.SendMessageWithOptions(SendMessageOptions{
		ChatID:                chatID,
		MessageThreadID:       threadID,
		Text:                  text,
		ParseMode:             "HTML",
		ReplyMarkup:           keyboard,
		DisableWebPagePreview: true,
	})
}

func (b *Bot) EditHTMLMessage(chatID int64, messageID int, text string) (*models.Message, error) {
	return b.EditMessageWithOptions(EditMessageOptions{
		ChatID:                chatID,
		MessageID:             messageID,
		Text:                  text,
		ParseMode:             "HTML",
		DisableWebPagePreview: true,
	})
}

func (b *Bot) EditHTMLMessageWithKeyboard(chatID int64, messageID int, text string, keyboard *models.InlineKeyboardMarkup) (*models.Message, error) {
	return b.EditMessageWithOptions(EditMessageOptions{
		ChatID:                chatID,
		MessageID:             messageID,
		Text:                  text,
		ParseMode:             "HTML",
		ReplyMarkup:           keyboard,
		DisableWebPagePreview: true,
	})
}

func (b *Bot) SendMessageWithOptions(opts SendMessageOptions) (*models.Message, error) {
	payload := map[string]interface{}{
		"chat_id": opts.ChatID,
		"text":    opts.Text,
	}

	if opts.ParseMode != "" {
		payload["parse_mode"] = opts.ParseMode
	}
	if opts.MessageThreadID != 0 {
		payload["message_thread_id"] = opts.MessageThreadID
	}
	if opts.ReplyToMessageID != 0 {
		payload["reply_to_message_id"] = opts.ReplyToMessageID
	}
	if opts.ReplyMarkup != nil {
		payload["reply_markup"] = opts.ReplyMarkup
	}
	if opts.DisableWebPagePreview {
		payload["disable_web_page_preview"] = true
	}

	return b.sendMessageRequest("/sendMessage", payload)
}

func (b *Bot) EditMessageWithOptions(opts EditMessageOptions) (*models.Message, error) {
	payload := map[string]interface{}{
		"chat_id":    opts.ChatID,
		"message_id": opts.MessageID,
		"text":       opts.Text,
	}

	if opts.ParseMode != "" {
		payload["parse_mode"] = opts.ParseMode
	}
	if opts.ReplyMarkup != nil {
		payload["reply_markup"] = opts.ReplyMarkup
	}
	if opts.DisableWebPagePreview {
		payload["disable_web_page_preview"] = true
	}

	return b.sendMessageRequest("/editMessageText", payload)
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

	return parseBoolResponse(body)
}

func parseBoolResponse(body []byte) error {
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

func (b *Bot) SendMessageWithKeyboard(chatID int64, text string, keyboard *models.InlineKeyboardMarkup) (*models.Message, error) {
	return b.SendMessageWithOptions(SendMessageOptions{
		ChatID:      chatID,
		Text:        text,
		ParseMode:   "Markdown",
		ReplyMarkup: keyboard,
	})
}

func (b *Bot) AnswerCallbackQuery(callbackQueryID string, text string) error {
	payload := map[string]interface{}{
		"callback_query_id": callbackQueryID,
	}

	if text != "" {
		payload["text"] = text
	}

	return b.sendBoolRequest("/answerCallbackQuery", payload)
}

func (b *Bot) CreateForumTopic(chatID int64, name string) (*models.ForumTopic, error) {
	payload := map[string]interface{}{
		"chat_id": chatID,
		"name":    name,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := b.client.Post(
		b.apiURL+"/createForumTopic",
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

	var result models.CreateForumTopicResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if !result.OK {
		return nil, fmt.Errorf("telegram API request failed: %s", result.Description)
	}

	return &result.Result, nil
}
