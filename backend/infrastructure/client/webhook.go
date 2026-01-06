package client

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// WebhookClient handles sending ntfy notifications
type WebhookClient struct {
	httpClient *http.Client
}

// NewWebhookClient creates a new webhook client
func NewWebhookClient() *WebhookClient {
	return &WebhookClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Alarm sends a urgent notification to the specified ntfy topic URL with a message
func (c *WebhookClient) Alarm(url string, message string) error {
	if url == "" {
		return fmt.Errorf("ntfy URL is empty")
	}

	if message == "" {
		message = "Work time notification"
	}

	slog.Info("[Ntfy] Sending notification", "url", url)

	req, err := http.NewRequest("POST", url, bytes.NewBufferString(message))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set ntfy headers
	req.Header.Set("Title", "Offline Me Alarm")
	req.Header.Set("Priority", "urgent")
	req.Header.Set("Tags", "warning,skull")
	req.Header.Set("Content-Type", "text/plain")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	slog.Info("[Ntfy] Response", "status", resp.StatusCode, "body", string(respBody))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("ntfy returned non-2xx status: %d", resp.StatusCode)
	}

	slog.Info("[Ntfy] Successfully sent notification", "url", url)
	return nil
}
