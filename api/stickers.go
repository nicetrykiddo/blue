package api

import (
	"blue/models"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

func (b *Bot) GetStickerSet(name string) (*models.StickerSet, error) {
	params := url.Values{}
	params.Add("name", name)

	resp, err := b.client.Get(b.apiURL + "/getStickerSet?" + params.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result models.GetStickerSetResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if !result.OK {
		return nil, fmt.Errorf("failed to get sticker set: %s", result.Description)
	}

	return &result.Result, nil
}
