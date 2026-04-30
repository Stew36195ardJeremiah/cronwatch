package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// JobConfig holds the configuration for a single monitored cron job.
type JobConfig struct {
	Name           string        `yaml:"name"`
	Schedule       string        `yaml:"schedule"`
	Timeout        time.Duration `yaml:"timeout"`
	DriftThreshold time.Duration `yaml:"drift_threshold"`
}

// SlackConfig holds Slack notifier settings.
type SlackConfig struct {
	WebhookURL string `yaml:"webhook_url"`
}

// PagerDutyConfig holds PagerDuty notifier settings.
type PagerDutyConfig struct {
	IntegrationKey string `yaml:"integration_key"`
}

// AlertsConfig groups all notifier configurations.
type AlertsConfig struct {
	Slack      *SlackConfig      `yaml:"slack"`
	PagerDuty  *PagerDutyConfig  `yaml:"pagerduty"`
}

// Config is the top-level cronwatch configuration.
type Config struct {
	Alerts AlertsConfig `yaml:"alerts"`
	Jobs   []JobConfig  `yaml:"jobs"`
}

// Load reads and validates a cronwatch YAML config file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse yaml: %w", err)
	}

	if len(cfg.Jobs) == 0 {
		return nil, fmt.Errorf("config: no jobs defined")
	}

	for i, job := range cfg.Jobs {
		if job.Schedule == "" {
			return nil, fmt.Errorf("config: job[%d] %q missing schedule", i, job.Name)
		}
	}

	return &cfg, nil
}
