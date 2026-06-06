package api

import (
	"blue/models"
	"bytes"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type WebhookOptions struct {
	URL             string
	SecretToken     string
	CertificateFile string
}

func (b *Bot) SetWebhook(url, secretToken string) error {
	return b.SetWebhookWithOptions(WebhookOptions{
		URL:         url,
		SecretToken: secretToken,
	})
}

func (b *Bot) SetWebhookWithOptions(opts WebhookOptions) error {
	payload := map[string]interface{}{
		"url":             opts.URL,
		"secret_token":    opts.SecretToken,
		"allowed_updates": []string{"message", "callback_query"},
	}

	if opts.CertificateFile != "" {
		return b.setWebhookWithCertificate(opts)
	}

	return b.sendBoolRequest("/setWebhook", payload)
}

func (b *Bot) setWebhookWithCertificate(opts WebhookOptions) error {
	cert, err := os.Open(opts.CertificateFile)
	if err != nil {
		return fmt.Errorf("open webhook certificate: %w", err)
	}
	defer cert.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if err := writer.WriteField("url", opts.URL); err != nil {
		return err
	}
	if err := writer.WriteField("secret_token", opts.SecretToken); err != nil {
		return err
	}
	if err := writer.WriteField("allowed_updates", `["message","callback_query"]`); err != nil {
		return err
	}

	part, err := writer.CreateFormFile("certificate", filepath.Base(opts.CertificateFile))
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, cert); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	resp, err := b.client.Post(b.apiURL+"/setWebhook", writer.FormDataContentType(), &body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return parseBoolResponse(responseBody)
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
