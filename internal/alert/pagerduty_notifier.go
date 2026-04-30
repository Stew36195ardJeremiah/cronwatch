package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const pagerDutyEventURL = "https://events.pagerduty.com/v2/enqueue"

// PagerDutyNotifier sends alerts to PagerDuty via the Events API v2.
type PagerDutyNotifier struct {
	integrationKey string
	httpClient     *http.Client
	eventURL       string
}

type pdPayload struct {
	RoutingKey  string    `json:"routing_key"`
	EventAction string    `json:"event_action"`
	Payload     pdDetails `json:"payload"`
}

type pdDetails struct {
	Summary  string `json:"summary"`
	Severity string `json:"severity"`
	Source   string `json:"source"`
	Timestamp string `json:"timestamp"`
}

// NewPagerDutyNotifier creates a PagerDutyNotifier with the given integration key.
func NewPagerDutyNotifier(integrationKey string) *PagerDutyNotifier {
	return &PagerDutyNotifier{
		integrationKey: integrationKey,
		httpClient:     &http.Client{Timeout: 10 * time.Second},
		eventURL:       pagerDutyEventURL,
	}
}

// Send dispatches an alert to PagerDuty.
func (p *PagerDutyNotifier) Send(a Alert) error {
	severity := "warning"
	if a.Level == LevelError {
		severity = "critical"
	}

	body := pdPayload{
		RoutingKey:  p.integrationKey,
		EventAction: "trigger",
		Payload: pdDetails{
			Summary:   fmt.Sprintf("[cronwatch] %s: %s", a.Job, a.Message),
			Severity:  severity,
			Source:    "cronwatch",
			Timestamp: a.Time.UTC().Format(time.RFC3339),
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("pagerduty: marshal payload: %w", err)
	}

	resp, err := p.httpClient.Post(p.eventURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("pagerduty: http post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("pagerduty: unexpected status %d", resp.StatusCode)
	}
	return nil
}
