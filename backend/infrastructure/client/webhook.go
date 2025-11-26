package client

import (
	"bytes"
	"fmt"
	"io"
	"log"
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

	log.Printf("[Ntfy] Sending notification to: %s", url)

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
	log.Printf("[Ntfy] Response status: %d, body: %s", resp.StatusCode, string(respBody))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("ntfy returned non-2xx status: %d", resp.StatusCode)
	}

	log.Printf("[Ntfy] Successfully sent notification to: %s", url)
	return nil
}
