package monitor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookPayload is the JSON body sent to a generic webhook endpoint.
type WebhookPayload struct {
	Job       string    `json:"job"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// WebhookNotifier sends job status events to a configurable HTTP endpoint.
type WebhookNotifier struct {
	URL    string
	Client *http.Client
}

// NewWebhookNotifier creates a WebhookNotifier with a default HTTP client.
func NewWebhookNotifier(url string) *WebhookNotifier {
	return &WebhookNotifier{
		URL:    url,
		Client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Notify sends a webhook POST with job status details.
func (w *WebhookNotifier) Notify(job, status, message string) error {
	payload := WebhookPayload{
		Job:       job,
		Status:    status,
		Message:   message,
		Timestamp: time.Now().UTC(),
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook: marshal payload: %w", err)
	}
	resp, err := w.Client.Post(w.URL, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("webhook: post request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: unexpected status %d", resp.StatusCode)
	}
	return nil
}
