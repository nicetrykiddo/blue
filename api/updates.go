package api

import (
	"blue/models"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
)

func (b *Bot) GetUpdates(offset, timeout int) ([]models.Update, error) {
	params := url.Values{}
	params.Add("offset", strconv.Itoa(offset))
	params.Add("timeout", strconv.Itoa(timeout))

	resp, err := b.client.Get(b.apiURL + "/getUpdates?" + params.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result models.GetUpdatesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if !result.OK {
		return nil, fmt.Errorf("telegram API error: %s", result.Description)
	}

	return result.Result, nil
}
