package api

import (
	"blue/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func (b *Bot) GetFile(fileID string) (*models.File, error) {
	payload := map[string]interface{}{
		"file_id": fileID,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := b.client.Post(
		b.apiURL+"/getFile",
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

	var result struct {
		OK          bool        `json:"ok"`
		Result      models.File `json:"result"`
		Description string      `json:"description,omitempty"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if !result.OK {
		return nil, fmt.Errorf("failed to get file: %s", result.Description)
	}

	return &result.Result, nil
}

func (b *Bot) DownloadFile(filePath string) ([]byte, error) {
	url := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", b.token, filePath)

	resp, err := b.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
