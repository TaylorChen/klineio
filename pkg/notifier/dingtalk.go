package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"klineio/pkg/log"
)

// DingTalkNotifier sends messages to DingTalk.
type DingTalkNotifier struct {
	webhookURL string
	client     *http.Client
	logger     *log.Logger
}

// NewDingTalkNotifier creates a new DingTalkNotifier.
func NewDingTalkNotifier(webhookURL string, logger *log.Logger) *DingTalkNotifier {
	return &DingTalkNotifier{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 5 * time.Second},
		logger:     logger,
	}
}

// SendMarkdownMessage sends a markdown message to DingTalk.
func (d *DingTalkNotifier) SendMarkdownMessage(ctx context.Context, title, text string) error {
	msg := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"title": title,
			"text":  text,
		},
	}

	jsonBody, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal dingtalk message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", d.webhookURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create dingtalk request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send dingtalk request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("dingtalk API returned non-OK status: %s", res.Status)
	}

	// Optionally read and log response body for debugging
	// body, _ := ioutil.ReadAll(res.Body)
	// d.logger.Debug("DingTalk response", zap.ByteString("body", body))

	return nil
}

// SendTextMessage sends a plain text message to DingTalk.
func (d *DingTalkNotifier) SendTextMessage(ctx context.Context, text string) error {
	msg := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": text,
		},
	}

	jsonBody, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal dingtalk message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", d.webhookURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create dingtalk request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send dingtalk request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("dingtalk API returned non-OK status: %s", res.Status)
	}

	return nil
} 