package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SlackNotifier sends alerts to a Slack webhook URL.
type SlackNotifier struct {
	webhookURL string
	client     *http.Client
}

type slackPayload struct {
	Text string `json:"text"`
}

// NewSlackNotifier creates a SlackNotifier with the given webhook URL.
func NewSlackNotifier(webhookURL string) *SlackNotifier {
	return &SlackNotifier{
		webhookURL: webhookURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Send delivers an alert message to Slack.
func (s *SlackNotifier) Send(a Alert) error {
	msg := fmt.Sprintf("[%s] %s — %s", a.Level, a.JobName, a.Message)
	payload := slackPayload{Text: msg}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("slack notifier: marshal payload: %w", err)
	}

	resp, err := s.client.Post(s.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("slack notifier: post request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("slack notifier: unexpected status %d", resp.StatusCode)
	}

	return nil
}
